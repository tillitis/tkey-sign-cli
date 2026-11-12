// SPDX-FileCopyrightText: 2026 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

package distfile

import (
	"encoding/json"
	"fmt"
	"os"

	"golang.org/x/crypto/blake2s"
)

type AlgType int

const (
	Ed AlgType = iota
	B2sEd
)

const nameLenMax = 16

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

func (m *Message) ToBuffer() ([]byte, error) {
	rawName := [nameLenMax]byte{}
	if len(m.Name) > len(rawName) {
		return []byte{}, fmt.Errorf("app name too long")
	}
	copy(rawName[:], m.Name)

	flags := [2]byte{}
	if m.Installable {
		flags[0] |= 1
	}

	d, err := blake2s.New256(nil)
	if err != nil {
		return []byte{}, fmt.Errorf("blake2s init failed: %w", err)
	}
	d.Write(m.AppDigest[:])
	d.Write(rawName[:])
	d.Write(flags[:])

	return d.Sum([]byte{}), nil
}

func (m *Message) FromBuffer(b []byte) error {
	err := json.Unmarshal(b, m)
	if err != nil {
		return fmt.Errorf("unmarshalling failed: %w", err)
	}

	if len(m.Name) > nameLenMax {
		return fmt.Errorf("app name too long")
	}

	return nil
}

func (m *Message) FromFile(fileName string) error {
	input, err := os.ReadFile(fileName)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return m.FromBuffer(input)
}

func (d *DistInfo) FromBuffer(b []byte) error {
	err := json.Unmarshal(b, d)
	if err != nil {
		return fmt.Errorf("JSON unmarshalling failed: %w", err)
	}

	return nil
}

func (d *DistInfo) FromFile(fileName string) error {
	input, err := os.ReadFile(fileName)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return d.FromBuffer(input)
}

func (d *DistInfo) ToBuffer() ([]byte, error) {
	j, err := json.Marshal(d)
	if err != nil {
		return []byte{}, fmt.Errorf("JSON marshalling failed: %w", err)
	}

	return j, nil
}
