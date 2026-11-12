// SPDX-FileCopyrightText: 2026 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

package distfile

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/blake2s"
)

type AlgType int

const (
	Ed AlgType = iota
	B2sEd
)

const nameLenMax = 16

const distInfoVersion = 1

type Message struct {
	Name        string   `json:"name"`
	AppDigest   [32]byte `json:"appDigest"`
	Installable bool     `json:"installable"`
}

type PubKey struct {
	Alg AlgType  `json:"alg"`
	Key [32]byte `json:"key"`
}

type Signature struct {
	Alg AlgType  `json:"alg"`
	Sig [64]byte `json:"signature"`
}

type DistInfo struct {
	Version   int       `json:"version"`
	Message   Message   `json:"message"`
	Signature Signature `json:"signature"`
	PubKey    PubKey    `json:"pubkey"`
}

// NewSignature instantiates a Signature from a byte slice.
//
// It expects an algorithm type and an Ed25519 signature slice.
//
// It returns a Signature and any error.
func NewSignature(alg AlgType, srcSig []byte) (Signature, error) {
	var sig Signature

	if alg != B2sEd {
		return sig, fmt.Errorf("unsupported algorithm")
	}

	if len(srcSig) != 64 {
		return sig, fmt.Errorf("signature wrong size")
	}

	sig.Alg = alg
	copy(sig.Sig[:], srcSig)

	return sig, nil
}

// NewPubKey instantiates a public key (PubKey) from a byte
// slice.
//
// srcKey is expected to be an Ed25519 public key of exactly 32 bytes.
func NewPubKey(srcKey []byte) (PubKey, error) {
	var key PubKey

	if len(srcKey) != 32 {
		return key, fmt.Errorf("key size wrong")
	}

	key.Alg = Ed
	copy(key.Key[:], srcKey)

	return key, nil
}

// SigningDigest validates the message, and returns a byte slice of m
// prepared to be signed.
func (m *Message) SigningDigest() ([]byte, error) {
	err := m.Validate()
	if err != nil {
		return nil, fmt.Errorf("invalid message: %w", err)
	}

	rawName := [nameLenMax]byte{}
	copy(rawName[:], m.Name)

	flags := [2]byte{}
	if m.Installable {
		flags[0] |= 1
	}

	d, err := blake2s.New256(nil)
	if err != nil {
		return nil, fmt.Errorf("blake2s init failed: %w", err)
	}
	d.Write(m.AppDigest[:])
	d.Write(rawName[:])
	d.Write(flags[:])

	return d.Sum([]byte{}), nil
}

// FromBuffer unmarshals m from JSON bytes and validates the result.
func (m *Message) FromBuffer(b []byte) error {
	var temp Message
	err := json.Unmarshal(b, &temp)
	if err != nil {
		return fmt.Errorf("unmarshalling failed: %w", err)
	}

	err = temp.Validate()
	if err != nil {
		return fmt.Errorf("invalid message: %w", err)
	}

	*m = temp

	return nil
}

// FromFile reads a JSON-encoded Message from the named file and
// unmarshals it into m.
func (m *Message) FromFile(fileName string) error {
	input, err := os.ReadFile(fileName)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return m.FromBuffer(input)
}

// ToBuffer returns the JSON encoding of m.
func (m *Message) ToBuffer() ([]byte, error) {
	j, err := json.Marshal(m)
	if err != nil {
		return []byte{}, fmt.Errorf("JSON marshalling failed: %w", err)
	}

	return j, nil
}

// Validate checks that the message fields conform to the required constraints.
func (m *Message) Validate() error {
	if len(m.Name) > nameLenMax {
		return fmt.Errorf("app name exceeds %d bytes", nameLenMax)
	}

	if strings.HasSuffix(m.Name, "\x00") {
		return fmt.Errorf("app name ends with a NUL byte")
	}

	for i := 0; i < len(m.Name); i++ {
		if m.Name[i] > 0x7f {
			return fmt.Errorf("app name contains a non-ASCII byte at offset %d", i)
		}
	}
	return nil
}

// FromBuffer unmarshals d from JSON bytes and validates the result.
func (d *DistInfo) FromBuffer(b []byte) error {
	var temp DistInfo
	err := json.Unmarshal(b, &temp)
	if err != nil {
		return fmt.Errorf("JSON unmarshalling failed: %w", err)
	}

	err = temp.Validate()
	if err != nil {
		return fmt.Errorf("invalid distinfo: %w", err)
	}

	*d = temp

	return nil
}

// FromFile reads a JSON-encoded DistInfo from the named file and
// unmarshals it into d.
func (d *DistInfo) FromFile(fileName string) error {
	input, err := os.ReadFile(fileName)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return d.FromBuffer(input)
}

// ToBuffer returns the JSON encoding of d.
func (d *DistInfo) ToBuffer() ([]byte, error) {
	j, err := json.Marshal(d)
	if err != nil {
		return []byte{}, fmt.Errorf("JSON marshalling failed: %w", err)
	}

	return j, nil
}

// Validate checks the version and algorithm fields for supported
// values, and validates the message.
func (d *DistInfo) Validate() error {
	if d.Version != distInfoVersion {
		return fmt.Errorf("unsupported dist file version %d", d.Version)
	}
	if d.PubKey.Alg != Ed {
		return fmt.Errorf("unsupported public key algorithm %d", d.PubKey.Alg)
	}
	if d.Signature.Alg != B2sEd {
		return fmt.Errorf("unsupported signature algorithm %d", d.Signature.Alg)
	}
	if err := d.Message.Validate(); err != nil {
		return fmt.Errorf("invalid message: %w", err)
	}
	return nil
}
