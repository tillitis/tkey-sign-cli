// SPDX-FileCopyrightText: 2026 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

package signify

import (
	"crypto/ed25519"
	"fmt"
	"testing"
)

// Full signature file
const WorkingSigString = `untrusted comment: app_a.bin
RWQBBwAAAAAAALE/DLVQ8RU5OA11qzhxDZ5nDOgbVGhxNwWlylI2YdPHIBVH/Q+HnhWfO5CxgPUb6EOCxG8ZzPVy+lQt4atUHAE=
`

// Fully parsed
var WorkingSig = Signature{
	Alg: Ed,
	Sig: [ed25519.SignatureSize]byte{0xb1, 0x3f, 0xc, 0xb5, 0x50, 0xf1, 0x15, 0x39, 0x38, 0xd, 0x75, 0xab, 0x38, 0x71, 0xd, 0x9e, 0x67, 0xc, 0xe8, 0x1b, 0x54, 0x68, 0x71, 0x37, 0x5, 0xa5, 0xca, 0x52, 0x36, 0x61, 0xd3, 0xc7, 0x20, 0x15, 0x47, 0xfd, 0xf, 0x87, 0x9e, 0x15, 0x9f, 0x3b, 0x90, 0xb1, 0x80, 0xf5, 0x1b, 0xe8, 0x43, 0x82, 0xc4, 0x6f, 0x19, 0xcc, 0xf5, 0x72, 0xfa, 0x54, 0x2d, 0xe1, 0xab, 0x54, 0x1c, 0x1},
}

// Full public key file
const WorkingKeyString = `untrusted comment: 
RWQBBwAAAAAAAJtidzMj70GhGDSCQZTlUWTTJeuc3MEN3afRCt5PvY9t
`

var WorkingKey = PubKey{0x9b, 0x62, 0x77, 0x33, 0x23, 0xef, 0x41, 0xa1, 0x18, 0x34, 0x82, 0x41, 0x94, 0xe5, 0x51, 0x64, 0xd3, 0x25, 0xeb, 0x9c, 0xdc, 0xc1, 0xd, 0xdd, 0xa7, 0xd1, 0xa, 0xde, 0x4f, 0xbd, 0x8f, 0x6d}

// TestSig generates a Signify format buffer, then parses the same
// buffer back into a signature.
func TestSig(t *testing.T) {
	var s Signature

	t.Parallel()
	buf, err := WorkingSig.ToBuffer("")
	if err != nil {
		t.Fatalf("Couldn't generate signify format: %v", err)
	}

	if err := s.FromBuffer(buf); err != nil {
		t.Fatalf("Couldn't read or parse signify format: %v", err)
	}

	if s != WorkingSig {
		t.Fatal("generating and parsing back signature was different")
	}

}

func TestParseSig(t *testing.T) {
	var sig Signature

	t.Parallel()
	if err := sig.FromBuffer([]byte(WorkingSigString)); err != nil {
		t.Fatalf("%v\n", err)
	}

	if sig != WorkingSig {
		t.Fatal("Parsed signature not correct\n")
	}
}

// TestKey generates a signify format buffer for a public key, then
// parses the same buffer back into a key.
func TestKey(t *testing.T) {
	var p PubKey

	t.Parallel()
	buf, err := WorkingKey.ToBuffer("")
	if err != nil {
		t.Fatalf("Couldn't generate signify format: %v", err)
	}

	if err := p.FromBuffer(buf); err != nil {
		t.Fatalf("Couldn't read or parse signify format: %v", err)
	}

	if p != WorkingKey {
		t.Fatal("generating and parsing back key was different")
	}

}

func TestParsePubKey(t *testing.T) {
	var key PubKey

	t.Parallel()
	if err := key.FromBuffer([]byte(WorkingKeyString)); err != nil {
		t.Fatalf("%v\n", err)
	}

	if key != WorkingKey {
		t.Fatal("Parsed public key not correct\n")
	}

}

func TestGenSigB64(t *testing.T) {
	t.Parallel()

	b64, err := WorkingSig.ToBuffer("app_a.bin")
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	if string(b64) != WorkingSigString {
		fmt.Printf("Expected (%v bytes): %v\n", len(WorkingSigString), WorkingSigString)
		fmt.Printf("Got      (%v bytes): %v\n", len(string(b64)), string(b64))

		t.Fatal("Known good sig generates no good string")
	}
}

func TestGenPubKeyB64(t *testing.T) {
	t.Parallel()

	b64, err := WorkingKey.ToBuffer("")
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	if string(b64) != WorkingKeyString {
		fmt.Printf("Expected (%v bytes): %v\n", len(WorkingKeyString), WorkingKeyString)
		fmt.Printf("Got      (%v bytes): %v\n", len(string(b64)), string(b64))

		t.Fatal("Known good key generates no good string")
	}
}

func TestSigFile(t *testing.T) {
	t.Parallel()

	var s Signature

	if err := WorkingSig.ToFile("test.sig", "a comment", true); err != nil {
		t.Fatalf("Couldn't generate or write file test.sig: %v", err)
	}

	if err := s.FromFile("test.sig"); err != nil {
		t.Fatalf("Couldn't read or parse test.sig: %v", err)
	}

	if s != WorkingSig {
		t.Fatal("Signature read from file not same as written")
	}
}

func TestPubFile(t *testing.T) {
	t.Parallel()

	var p PubKey

	if err := WorkingKey.ToFile("test.pub", "a comment", true); err != nil {
		t.Fatalf("Couldn't generate or write file test.pub: %v", err)
	}

	if err := p.FromFile("test.pub"); err != nil {
		t.Fatalf("Couldn't read or parse test.pub: %v", err)
	}

	if p != WorkingKey {
		t.Fatal("Key read from file not same as written")
	}
}
