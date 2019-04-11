// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package enr

import (
	"crypto/ecdsa"
	"fmt"
	"sync"

	"github.com/MatrixAINetwork/go-matrix/common/math"
	"github.com/MatrixAINetwork/go-matrix/crypto"
	"github.com/MatrixAINetwork/go-matrix/crypto/sha3"
	"github.com/MatrixAINetwork/go-matrix/rlp"
)

// Registry of known identity schemes.
var schemes sync.Map

// An IdentityScheme is capable of verifying record signatures and
// deriving node addresses.
type IdentityScheme interface {
	Verify(r *Record, sig []byte) error
	NodeAddr(r *Record) []byte
}

// RegisterIdentityScheme adds an identity scheme to the global registry.
func RegisterIdentityScheme(name string, scheme IdentityScheme) {
	if _, loaded := schemes.LoadOrStore(name, scheme); loaded {
		panic("identity scheme " + name + " already registered")
	}
}

// FindIdentityScheme resolves name to an identity scheme in the global registry.
func FindIdentityScheme(name string) IdentityScheme {
	s, ok := schemes.Load(name)
	if !ok {
		return nil
	}
	return s.(IdentityScheme)
}

// v4ID is the "v4" identity scheme.
type v4ID struct{}

func init() {
	RegisterIdentityScheme("v4", v4ID{})
}

// SignV4 signs a record using the v4 scheme.
func SignV4(r *Record, privkey *ecdsa.PrivateKey) error {
	// Copy r to avoid modifying it if signing fails.
	cpy := *r
	cpy.Set(ID("v4"))
	cpy.Set(Secp256k1(privkey.PublicKey))

	h := sha3.NewKeccak256()
	rlp.Encode(h, cpy.AppendElements(nil))
	sig, err := crypto.Sign(h.Sum(nil), privkey)
	if err != nil {
		return err
	}
	sig = sig[:len(sig)-1] // remove v
	if err = cpy.SetSig("v4", sig); err == nil {
		*r = cpy
	}
	return err
}

// s256raw is an unparsed secp256k1 public key entry.
type s256raw []byte

func (s256raw) ENRKey() string { return "secp256k1" }

func (v4ID) Verify(r *Record, sig []byte) error {
	var entry s256raw
	if err := r.Load(&entry); err != nil {
		return err
	} else if len(entry) != 33 {
		return fmt.Errorf("invalid public key")
	}

	h := sha3.NewKeccak256()
	rlp.Encode(h, r.AppendElements(nil))
	if !crypto.VerifySignature(entry, h.Sum(nil), sig) {
		return errInvalidSig
	}
	return nil
}

func (v4ID) NodeAddr(r *Record) []byte {
	var pubkey Secp256k1
	err := r.Load(&pubkey)
	if err != nil {
		return nil
	}
	buf := make([]byte, 64)
	math.ReadBits(pubkey.X, buf[:32])
	math.ReadBits(pubkey.Y, buf[32:])
	return crypto.Keccak256(buf)
}
