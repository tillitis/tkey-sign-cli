// Copyright (C) 2022, 2023 - Tillitis AB
// SPDX-License-Identifier: GPL-2.0-only

package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha512"
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"strings"
	"syscall"

	"github.com/spf13/pflag"
	"github.com/tillitis/tkeyclient"
	"github.com/tillitis/tkeysign"
	"github.com/tillitis/tkeyutil"
)

type command int

const (
	cmdUnknown = iota
	cmdGetKey
	cmdSign
	cmdVerify
)

// nolint:typecheck // Avoid lint error when the embedding file is missing.
// Build copies the built signer here
//
//go:embed signer.bin
var signerBinary []byte

// Use when printing err/diag msgs
var le = log.New(os.Stderr, "", 0)

var (
	version string
	verbose = false
)

type pubKey struct {
	Alg    [2]byte
	KeyNum [8]byte
	Key    [32]byte
}

type signature struct {
	Alg    [2]byte
	KeyNum [8]byte
	Sig    [64]byte
}

// May be set to non-empty at build time to indicate that the signer
// app has been compiled with touch requirement removed.
var signerAppNoTouch string

// signFile uses the connection to the signer and produces an Ed25519
// signature over the file in fileName. It automatically verifies the
// signature against the provided pubkey.
//
// It returns the Ed25519 signature on success or an error.
func signFile(signer tkeysign.Signer, pubkey []byte, fileName string) (*signature, error) {
	message, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("ReadFile: %w", err)
	}

	fileDigest := sha512.Sum512(message)
	fileDigestHex := fmt.Sprintf("%x  %s\n", fileDigest, fileName)
	if verbose {
		le.Printf("SHA512 hash: %x", fileDigest)
		le.Printf("SHA512 hash: %v", fileDigest)
	}

	if signerAppNoTouch != "" {
		le.Printf("WARNING! This tkey-sign and signer app is built with the touch requirement removed")
	}

	sig, err := signer.Sign([]byte(fileDigestHex))
	if err != nil {
		return nil, fmt.Errorf("signing failed: %w", err)
	}

	if verbose {
		le.Printf("signature: %x", sig)
	}

	if !ed25519.Verify(pubkey, []byte(fileDigestHex), sig) {
		return nil, fmt.Errorf("signature FAILED verification")
	}

	s := signature{
		Alg:    [2]byte{'E', 'd'},
		KeyNum: [8]byte{1, 7},
		Sig:    [64]byte{},
	}

	copy(s.Sig[:], sig)

	return &s, nil
}

// verifySignature verifies a Ed25519 signature stored in sigFile over
// messageFile with public key in pubkeyFile
func verifySignature(messageFile string, sigFile string, pubkeyFile string) error {
	signature, err := readSig(sigFile)
	if err != nil {
		if errors.Is(errors.Unwrap(err), fs.ErrNotExist) {
			return fmt.Errorf("Signature file %v not found, specify with '-x sigfile'", sigFile)
		}

		return fmt.Errorf("%w", err)
	}

	if len(signature.Sig) != 64 {
		return fmt.Errorf("invalid length of signature. Expected 64 bytes, got %d bytes", len(signature.Sig))
	}

	pubkey, err := readKey(pubkeyFile)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	if len(pubkey.Key) != 32 {
		return fmt.Errorf("invalid length of public key. Expected 32 bytes, got %d bytes", len(pubkey.Key))
	}

	message, err := os.ReadFile(messageFile)
	if err != nil {
		return fmt.Errorf("could not read %s: %w", messageFile, err)
	}

	digest := sha512.Sum512(message)
	digestHex := fmt.Sprintf("%x  %s\n", digest, messageFile)
	if verbose {
		le.Printf("SHA512 hash: %x", digest)
	}

	if !ed25519.Verify(pubkey.Key[:], []byte(digestHex), signature.Sig[:]) {
		return fmt.Errorf("signature not valid")
	}

	return nil
}

// loadSigner loads the signer device app into the TKey at device
// devPath with speed b/s, possibly using a User Supplied Secret
// either in fileUSS or prompting for the USS interactively if
// enterUSS is true.
//
// It then connects to the running signer and returns an interface to
// the Signer, the public key, and a possible error.
func loadSigner(devPath string, speed int, fileUSS string, enterUSS bool) (*tkeysign.Signer, []byte, error) {
	if !verbose {
		tkeyclient.SilenceLogging()
	}

	if devPath == "" {
		var err error
		devPath, err = tkeyclient.DetectSerialPort(false)
		if err != nil {
			return nil, nil, fmt.Errorf("DetectSerialPort: %w", err)
		}
	}

	tk := tkeyclient.New()
	if verbose {
		le.Printf("Connecting to TKey on serial port %s ...", devPath)
	}
	if err := tk.Connect(devPath, tkeyclient.WithSpeed(speed)); err != nil {
		return nil, nil, fmt.Errorf("could not open %s: %w", devPath, err)
	}

	if isFirmwareMode(tk) {
		var secret []byte
		var err error

		if enterUSS {
			secret, err = tkeyutil.InputUSS()
			if err != nil {
				tk.Close()
				return nil, nil, fmt.Errorf("InputUSS: %w", err)
			}
		}
		if fileUSS != "" {
			secret, err = tkeyutil.ReadUSS(fileUSS)
			if err != nil {
				tk.Close()
				return nil, nil, fmt.Errorf("ReadUSS: %w", err)
			}
		}

		if err := tk.LoadApp(signerBinary, secret); err != nil {
			tk.Close()
			return nil, nil, fmt.Errorf("couldn't load signer: %w", err)
		}

		if verbose {
			le.Printf("Signer app loaded.")
		}
	} else {
		if enterUSS || fileUSS != "" {
			le.Printf("WARNING: App already loaded, your USS won't be used.")
		} else {
			le.Printf("WARNING: App already loaded.")
		}
	}

	signer := tkeysign.New(tk)

	handleSignals(func() { os.Exit(1) }, os.Interrupt, syscall.SIGTERM)

	if !isWantedApp(signer) {
		signer.Close()
		return nil, nil, fmt.Errorf("no TKey on the serial port, or it's running wrong app (and is not in firmware mode)")
	}

	pubkey, err := signer.GetPubkey()
	if err != nil {
		signer.Close()
		return nil, nil, fmt.Errorf("GetPubKey failed: %w", err)
	}

	return &signer, pubkey, nil
}

func usage() {
	desc := fmt.Sprintf(`Usage:

%[1]s -h/--help

%[1]s -G/--getkey -p/--public pubkey [-d/--port device] [-f/--force] [-s/--speed speed] [--uss] [--uss-file ussfile] [--verbose] 

%[1]s -S/--sign -m message -p/--public pubkey [-d/--port device] [-f/--force] [-s speed] [--uss] [--uss-file ussfile] [--verbose] [-x sigfile]

%[1]s -V/--verify -m message -p/--public pubkey [-x sigfile]

%[1]s --version

%[1]s signs (-S) or verifies (-V) the signature of a message in a
file. The message will be hashed with SHA-512 and either signed using
the TKey's private key or verified given the public key. The signing
algorithm is Ed25519.

Exit status code is 0 if everything went well or non-zero if unsuccessful.

Alternatively, -G/--getkey can be used to receive the public key of
the signer app on the TKey. Specify where to store it with -p key.pub`,
		os.Args[0])

	le.Printf("%s\n\n%s", desc,
		pflag.CommandLine.FlagUsagesWrapped(86))
}

func main() {
	var cmd command
	var cmdArgs int
	getKey := pflag.BoolP("getkey", "G", false, "Get public key.")
	sign := pflag.BoolP("sign", "S", false, "Sign the message.")
	verify := pflag.BoolP("verify", "V", false, "Verify signature of the message.")
	force := pflag.BoolP("force", "f", false, "Force writing of signature and pubkey files, overwriting any existing files.")
	keyFile := pflag.StringP("public", "p", "", "Public key `pubkey`.")
	sigFile := pflag.StringP("sig", "x", "", "Signature `sigfile`.")
	messageFile := pflag.StringP("message", "m", "", "Specify file containing `message`.")
	devPath := pflag.StringP("port", "d", "",
		"Set serial port `device`. If this is not used, auto-detection will be attempted.")
	speed := pflag.IntP("speed", "s", tkeyclient.SerialSpeed,
		"Set serial port `speed` in bits per second.")
	enterUss := pflag.Bool("uss", false,
		"Enable typing of a phrase to be hashed as the User Supplied Secret. The USS is loaded onto the TKey along with the app itself. A different USS results in different public/private keys.")
	ussFile := pflag.String("uss-file", "",
		"Read `ussfile` and hash its contents as the USS. Use '-' (dash) to read from stdin. The full contents are hashed unmodified (e.g. newlines are not stripped).")
	versionOnly := pflag.BoolP("version", "v", false, "Output version information.")
	helpOnly := pflag.BoolP("help", "h", false, "Output this help.")

	if version == "" {
		version = readBuildInfo()
	}

	pflag.BoolVar(&verbose, "verbose", false, "Enable verbose output.")
	pflag.Usage = usage
	pflag.Parse()

	if pflag.NArg() > 0 {
		le.Printf("Unexpected argument: %s\n\n", strings.Join(pflag.Args(), " "))
		pflag.Usage()
		os.Exit(2)
	}

	if *versionOnly {
		le.Printf("tkey-sign %s", version)
		os.Exit(0)
	}

	if *helpOnly {
		pflag.Usage()
		os.Exit(0)

	}

	if *getKey {
		cmd = cmdGetKey
		cmdArgs++
	}

	if *sign {
		cmd = cmdSign
		cmdArgs++
	}

	if *verify {
		cmd = cmdVerify
		cmdArgs++
	}

	if cmdArgs > 1 {
		pflag.Usage()
		os.Exit(1)
	}

	switch cmd {
	case cmdGetKey:
		if *keyFile == "" {
			le.Printf("Provide public key file with -p pubkey")
			os.Exit(1)
		}

		signer, pub, err := loadSigner(*devPath, *speed, *ussFile, *enterUss)
		if err != nil {
			le.Printf("Couldn't load signer: %v", err)
			os.Exit(1)
		}

		pubkey := pubKey{
			Alg:    [2]byte{'E', 'd'},
			KeyNum: [8]byte{1, 7},
			Key:    [32]byte{},
		}

		copy(pubkey.Key[:], pub)

		comment := "tkey public key"
		if *force {
			err = writeBase64(*keyFile, pubkey, comment, true)
		} else {
			err = writeRetry(*keyFile, pubkey, comment)
		}

		if err != nil {
			le.Printf("%v", err)
			signer.Close()
			os.Exit(1)
		}

		signer.Close()

	case cmdSign:
		if *messageFile == "" {
			le.Printf("Provide message file with -m message")
			os.Exit(1)
		}

		if *keyFile == "" {
			le.Printf("Provide public key file with -p pubkey")
			os.Exit(1)
		}

		if *sigFile == "" {
			*sigFile = *messageFile + ".sig"
		}

		pubkey, err := readKey(*keyFile)
		if err != nil {
			le.Printf("Couldn't read pubkey file: %v", err)
			os.Exit(1)
		}

		signer, pub, err := loadSigner(*devPath, *speed, *ussFile, *enterUss)
		if err != nil {
			le.Printf("Couldn't load signer: %v", err)
			os.Exit(1)
		}

		if !bytes.Equal(pub, pubkey.Key[:]) {
			le.Printf("Public key from file %v not equal to loaded app's", *keyFile)
			os.Exit(1)
		}

		sig, err := signFile(*signer, pub, *messageFile)
		if err != nil {
			le.Printf("signing failed: %v", err)
			signer.Close()
			os.Exit(1)
		}

		comment := fmt.Sprintf("verify with %v", *keyFile)
		if *force {
			err = writeBase64(*sigFile, sig, comment, true)
		} else {
			err = writeRetry(*sigFile, sig, comment)
		}

		if err != nil {
			le.Printf("Couldn't store signature: %v", err)
			signer.Close()
			os.Exit(1)
		}

		signer.Close()

	case cmdVerify:
		if *messageFile == "" {
			le.Printf("Provide message file with -m message")
			os.Exit(1)
		}

		if *keyFile == "" {
			le.Printf("Provide public key file path with -p pubkey")
			os.Exit(1)
		}

		if *sigFile == "" {
			*sigFile = *messageFile + ".sig"
		}

		err := verifySignature(*messageFile, *sigFile, *keyFile)
		if err != nil {
			le.Printf("Error verifying: %v", err)
			os.Exit(1)
		}
		le.Printf("Signature verified")

	default:
		pflag.Usage()
		os.Exit(2)
	}

	// Success
	os.Exit(0)
}
