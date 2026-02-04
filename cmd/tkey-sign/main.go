// SPDX-FileCopyrightText: 2022 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha512"
	_ "embed"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"strings"
	"syscall"

	"github.com/spf13/pflag"
	"github.com/tillitis/tkey-sign-cli/signify"
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
	cmdDump
)

type devArgs struct {
	Path  string
	Speed int
}

type USSArgs struct {
	FileName string
	Request  bool
}

// nolint:typecheck // Avoid lint error when the embedding file is missing.
// Build copies the built signer here
//
//go:embed signer.bin-v1.0.0
var signerBinary []byte

const appName string = "tkey-device-signer 1.0.0"

// Use when printing err/diag msgs
var le = log.New(os.Stderr, "", 0)

var (
	version string
	verbose = false
)

// May be set to non-empty at build time to indicate that the signer
// app has been compiled with touch requirement removed.
var signerAppNoTouch string

// GetEmbeddedAppName returns the name of the embedded device app.
func GetEmbeddedAppName() string {
	return appName
}

// GetEmbeddedAppDigest returns a string of the SHA512 digest for the embedded
// device app
func GetEmbeddedAppDigest() string {
	digest := sha512.Sum512(signerBinary)
	return hex.EncodeToString(digest[:])
}

// signMessage uses the connection to the signer and produces an Ed25519
// signature over the file in fileName. It automatically verifies the
// signature against the provided pubkey.
//
// It returns the Ed25519 signature on success or an error.
func signMessage(signer tkeysign.Signer, pubkey []byte, message string) (*signify.Signature, error) {
	if signerAppNoTouch != "" {
		le.Printf("WARNING! This tkey-sign and signer app is built with the touch requirement removed")
	}

	sig, err := signer.Sign([]byte(message))
	if err != nil {
		return nil, fmt.Errorf("signing failed: %w", err)
	}

	s, err := signify.NewSignature(signify.Ed, sig)
	if err != nil {
		return nil, fmt.Errorf("couldn't convert to signify signature")
	}

	if verbose {
		le.Printf("signature: %x", sig)
	}

	if !ed25519.Verify(pubkey, []byte(message), sig) {
		return nil, fmt.Errorf("signature FAILED verification")
	}

	return &s, nil
}

// verifySignature verifies a Ed25519 signature stored in sigFile over
// messageFile with public key in pubkeyFile
func verifySignature(message string, sigFile string, pubkeyFile string) error {
	var signature signify.Signature

	if err := signature.FromFile(sigFile); err != nil {
		if errors.Is(errors.Unwrap(err), fs.ErrNotExist) {
			return fmt.Errorf("signature file %v not found, specify with '-x sigfile'", sigFile)
		}

		return fmt.Errorf("%w", err)
	}

	var pubKey signify.PubKey

	if err := pubKey.FromFile(pubkeyFile); err != nil {
		return fmt.Errorf("%w", err)
	}

	if !ed25519.Verify(pubKey[:], []byte(message), signature.Sig[:]) {
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

func dumpFiles(sigFn string, keyFn string) error {
	var sig signify.Signature
	var key signify.PubKey

	if err := sig.FromFile(sigFn); err != nil {
		return fmt.Errorf("%w", err)
	}

	fmt.Printf("Signature\n  Alg: ")
	switch sig.Alg {
	case signify.Ed:
		fmt.Printf("Ed\n")

	case signify.B2sEd:
		fmt.Printf("B2sEd\n")

	default:
		fmt.Printf(" <unknown>: %v\n", sig.Alg)
	}

	fmt.Printf("  Sig: %x\n", sig.Sig)

	if err := key.FromFile(keyFn); err != nil {
		return fmt.Errorf("%w", err)
	}

	fmt.Printf("Key: %x\n", key)

	return nil
}

func GetKey(keyFn string, overwrite bool, dev devArgs, uss USSArgs) error {
	if keyFn == "" {
		return errors.New("please provide -p pubkey")
	}

	signer, pub, err := loadSigner(dev.Path, dev.Speed, uss.FileName, uss.Request)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	defer signer.Close()
	pubkey, err := signify.NewPubKey(pub)
	if err != nil {
		le.Printf("Couldn't convert public key from signer to Signify key\n")
	}

	comment := "tkey public key"
	if overwrite {
		err = pubkey.ToFile(keyFn, comment, true)
	} else {
		err = writeRetry(keyFn, &pubkey, comment)
	}

	if err != nil {
		signer.Close()
		return fmt.Errorf("%w", err)
	}

	return nil
}

func Sign(msg string, keyFn string, sigFn string, overwrite bool, dev devArgs, uss USSArgs) error {
	var pubKey signify.PubKey

	if keyFn == "" {
		return errors.New("provide -p pubkey")
	}

	if err := pubKey.FromFile(keyFn); err != nil {
		return fmt.Errorf("couldn't read pubkey file: %w", err)
	}

	signer, pub, err := loadSigner(dev.Path, dev.Speed, uss.FileName, uss.Request)
	if err != nil {
		return fmt.Errorf("couldn't load signer: %w", err)
	}

	defer signer.Close()

	if !bytes.Equal(pub, pubKey[:]) {
		return fmt.Errorf("key from file %v not equal to loaded app's", keyFn)
	}

	sig, err := signMessage(*signer, pub, msg)
	if err != nil {
		return fmt.Errorf("signing failed: %w", err)
	}

	comment := fmt.Sprintf("verify with %v", keyFn)
	if overwrite {
		err = sig.ToFile(sigFn, comment, true)
	} else {
		err = writeRetry(sigFn, sig, comment)
	}

	if err != nil {
		return fmt.Errorf("couldn't store signature: %w", err)
	}

	return nil
}

func Verify(msg string, keyFn string, sigFn string) error {
	if keyFn == "" {
		return errors.New("provide public key file path with -p pubkey")
	}

	if err := verifySignature(msg, sigFn, keyFn); err != nil {
		return fmt.Errorf("verifying failed: %w", err)
	}

	return nil
}

func Dump(keyFn string, sigFn string) error {
	if keyFn == "" {
		return errors.New("provide public key file path with -p pubkey")
	}

	err := dumpFiles(sigFn, keyFn)
	if err != nil {
		return fmt.Errorf("error dumping data: %w", err)
	}

	return nil
}

// getMessage returns the message to sign or verify.
func getMessage(msgFn string) (string, error) {
	if msgFn == "" {
		return "", errors.New("provide -m messagefile")
	}

	file, err := os.ReadFile(msgFn)
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}

	fileDigest := sha512.Sum512(file)
	fileDigestHex := fmt.Sprintf("%x  %s\n", fileDigest, msgFn)
	if verbose {
		le.Printf("SHA512 hash: %x", fileDigest)
		le.Printf("SHA512 hash: %v", fileDigest)
	}

	return fileDigestHex, nil
}

func usage() {
	desc := fmt.Sprintf(`Usage:

%[1]s -h/--help

%[1]s -D/--dump -p/--public pubkey -m message  [-x sigfile]
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
the signer app on the TKey. Specify where to store it with -p key.pub.

Use -D/--dump to get more information about signature and public key
files.`, os.Args[0])

	le.Printf("%s\n\n%s", desc,
		pflag.CommandLine.FlagUsagesWrapped(86))
}

func main() {
	var cmd command
	var cmdArgs int
	dump := pflag.BoolP("dump", "D", false, "Dump data about files.")
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
		le.Printf("tkey-sign %s\n\n", version)
		le.Printf("Embedded device app:\n%s\nSHA512: %s\n", GetEmbeddedAppName(), GetEmbeddedAppDigest())
		os.Exit(0)
	}

	if *helpOnly {
		pflag.Usage()
		os.Exit(0)

	}

	if *dump {
		cmd = cmdDump
		cmdArgs++
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

	dev := devArgs{
		Path:  *devPath,
		Speed: *speed,
	}

	uss := USSArgs{
		FileName: *ussFile,
		Request:  *enterUss,
	}

	var msg string

	if cmd == cmdSign || cmd == cmdVerify {
		var err error

		msg, err = getMessage(*messageFile)
		if err != nil {
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}
	}

	if *sigFile == "" {
		*sigFile = *messageFile + ".sig"
	}

	switch cmd {
	case cmdGetKey:
		if err := GetKey(*keyFile, *force, dev, uss); err != nil {
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}

	case cmdSign:
		if err := Sign(msg, *keyFile, *sigFile, *force, dev, uss); err != nil {
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}

	case cmdVerify:
		if err := Verify(msg, *keyFile, *sigFile); err != nil {
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}

		le.Printf("Signature verified")

	case cmdDump:
		if err := Dump(*keyFile, *sigFile); err != nil {
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}

	default:
		pflag.Usage()
		os.Exit(2)
	}

	// Success
	os.Exit(0)
}
