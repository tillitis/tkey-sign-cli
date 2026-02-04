// SPDX-FileCopyrightText: 2026 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

// Package signify implements types and methods to interact with data
// compatible with the files the signify command use.
//
// NB! We only support Ed25519 signatures and public keys.
package signify

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

// The PubKey and Signature types adher to this interface.
//
// FromFile parses a public key or signature from the Signify file
// format.
//
// ToFile exports a public key or signature to the Signify file format.
//
// FromBuffer parses from a Signify buffer.
//
// ToBuffer export to a Signify buffer.
type Data interface {
	FromFile(fileName string) error
	ToFile(fileName string, comment string, overwrite bool) error
	FromBuffer(buf []byte) error
	ToBuffer(comment string) ([]byte, error)
}

type AlgType int

const (
	Ed AlgType = iota
	B2sEd
)

// A signify-compatible Ed25519 public key. Instantiate directly or
// using NewPubKey if you have a slice.
type PubKey [ed25519.PublicKeySize]byte

type signifyPubKey struct {
	Alg    [2]byte
	KeyNum [8]byte
	Key    [ed25519.PublicKeySize]byte
}

// A signify-compatible Ed25519 signature. Instantiate directly or
// using NewSignature if you have a slice.
type Signature struct {
	Alg AlgType
	Sig [ed25519.SignatureSize]byte
}

type signifySignature struct {
	Alg    [2]byte
	KeyNum [8]byte
	Sig    [ed25519.SignatureSize]byte
}

// NewPubKey instantiates a signify PubKey from a byte slice.
func NewPubKey(srcKey []byte) (PubKey, error) {
	var key PubKey

	if len(srcKey) != ed25519.PublicKeySize {
		return key, fmt.Errorf("key too large")
	}

	copy(key[:], srcKey)

	return key, nil
}

func (p *PubKey) FromFile(fileName string) error {
	input, err := os.ReadFile(fileName)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return p.FromBuffer(input)
}

func (p *PubKey) FromBuffer(b []byte) error {
	var pubKey signifyPubKey

	buf, err := fromSlice(b)
	if err != nil {
		return fmt.Errorf("could not decode: %w", err)
	}

	r := bytes.NewReader(buf)
	if err := binary.Read(r, binary.BigEndian, &pubKey); err != nil {
		return fmt.Errorf("%w", err)
	}

	if pubKey.Alg != [2]byte{'E', 'd'} || pubKey.KeyNum != [8]byte{1, 7} {
		return fmt.Errorf("incompatible key")
	}

	copy(p[:], pubKey.Key[:])

	return nil
}

func (p *PubKey) ToBuffer(comment string) ([]byte, error) {
	signifyKey := signifyPubKey{
		Alg:    [2]uint8{'E', 'd'},
		KeyNum: [8]uint8{0x1, 0x7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		Key:    *p,
	}

	return toSlice(signifyKey, comment)
}

func (p *PubKey) ToFile(fileName string, comment string, overwrite bool) error {
	buf, err := p.ToBuffer(comment)
	if err != nil {
		return err
	}

	return writeFile(fileName, buf, overwrite)
}

// NewSignature instantiates a signify Signature from a byte slice.
func NewSignature(t AlgType, srcSig []byte) (Signature, error) {
	var sig Signature

	if t != Ed && t != B2sEd {
		return sig, fmt.Errorf("unknown algorithm")
	}

	if len(srcSig) != ed25519.SignatureSize {
		return sig, fmt.Errorf("signature too large")
	}

	sig.Alg = t
	copy(sig.Sig[:], srcSig)

	return sig, nil
}

func (s *Signature) FromFile(fileName string) error {
	input, err := os.ReadFile(fileName)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return s.FromBuffer(input)
}

func (s *Signature) FromBuffer(b []byte) error {
	var sig signifySignature

	buf, err := fromSlice(b)
	if err != nil {
		return fmt.Errorf("could not decode: %w", err)
	}

	r := bytes.NewReader(buf)
	if err := binary.Read(r, binary.BigEndian, &sig); err != nil {
		return fmt.Errorf("binary read: %w", err)
	}

	switch sig.Alg {
	case [2]byte{'E', 'd'}:
		s.Alg = Ed

	case [2]byte{'E', 'b'}:
		s.Alg = B2sEd

	default:
		return fmt.Errorf("unknown signature algorithm")
	}

	if sig.KeyNum != [8]byte{1, 7} {
		return fmt.Errorf("incompatible signature")
	}

	copy(s.Sig[:], sig.Sig[:])

	return nil
}

func (s *Signature) ToBuffer(comment string) ([]byte, error) {
	signifySig := signifySignature{
		KeyNum: [8]uint8{1, 7},
		Sig:    s.Sig,
	}

	switch s.Alg {
	case Ed:
		signifySig.Alg = [2]uint8{'E', 'd'}

	case B2sEd:
		signifySig.Alg = [2]uint8{'E', 'b'}

	default:
		return nil, fmt.Errorf("unknown sig algorithm")

	}

	return toSlice(signifySig, comment)
}

func (s *Signature) ToFile(fileName string, comment string, overwrite bool) error {
	buf, err := s.ToBuffer(comment)
	if err != nil {
		return err
	}

	return writeFile(fileName, buf, overwrite)
}

func fromSlice(input []byte) ([]byte, error) {
	lines := strings.Split(string(input), "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("too few lines")
	}

	data, err := base64.StdEncoding.DecodeString(lines[1])
	if err != nil {
		return nil, fmt.Errorf("could not decode: %w", err)
	}

	return data, nil
}

func toSlice(data any, comment string) ([]byte, error) {
	var binBuf bytes.Buffer

	err := binary.Write(&binBuf, binary.BigEndian, data)
	if err != nil {
		return []byte{}, fmt.Errorf("%w", err)
	}

	b64 := base64.StdEncoding.EncodeToString(binBuf.Bytes())
	b64 += "\n"

	var buf bytes.Buffer
	if _, err := buf.WriteString(fmt.Sprintf("untrusted comment: %s\n%s", comment, b64)); err != nil {
		return []byte{}, fmt.Errorf("%w", err)
	}

	return buf.Bytes(), nil
}

// writeFile writes buf into file filename. If overwrite is true it
// overwrites any existing file, otherwise it returns an error.
func writeFile(filename string, buf []byte, overwrite bool) error {
	var f *os.File

	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0o666)
	if err != nil {
		if os.IsExist(err) && overwrite {
			f, err = os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0o666)
			if err != nil {
				return fmt.Errorf("%w", err)
			}
		} else {
			return fmt.Errorf("%w", err)
		}
	}

	defer f.Close()

	if _, err := f.Write(buf); err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}
