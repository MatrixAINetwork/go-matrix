// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

// +build !nacl,!js,!nocgo

package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"errors"
	"fmt"

	"math/big"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/common/math"
	"github.com/MatrixAINetwork/go-matrix/crypto/secp256k1"
)

// Ecrecover returns the uncompressed public key that created the given signature.
func Ecrecover(hash, sig []byte) ([]byte, error) {
	return secp256k1.RecoverPubkey(hash, sig)
}

// SigToPub returns the public key that created the given signature.
func SigToPub(hash, sig []byte) (*ecdsa.PublicKey, error) {
	s, err := Ecrecover(hash, sig)
	if err != nil {
		return nil, err
	}

	x, y := elliptic.Unmarshal(S256(), s)
	return &ecdsa.PublicKey{Curve: S256(), X: x, Y: y}, nil
}

// Sign calculates an ECDSA signature.
//
// This function is susceptible to chosen plaintext attacks that can leak
// information about the private key that is used for signing. Callers must
// be aware that the given hash cannot be chosen by an adversery. Common
// solution is to hash any input before calculating the signature.
//
// The produced signature is in the [R || S || V] format where V is 0 or 1.
func Sign(hash []byte, prv *ecdsa.PrivateKey) (sig []byte, err error) {
	if len(hash) != 32 {
		return nil, fmt.Errorf("hash is required to be exactly 32 bytes (%d)", len(hash))
	}
	seckey := math.PaddedBigBytes(prv.D, prv.Params().BitSize/8)
	defer zeroBytes(seckey)
	return secp256k1.Sign(hash, seckey)
}

// VerifySignature checks that the given public key created signature over hash.
// The public key should be in compressed (33 bytes) or uncompressed (65 bytes) format.
// The signature should have the 64 byte [R || S] format.
func VerifySignature(pubkey, hash, signature []byte) bool {
	return secp256k1.VerifySignature(pubkey, hash, signature)
}

// DecompressPubkey parses a public key in the 33-byte compressed format.
func DecompressPubkey(pubkey []byte) (*ecdsa.PublicKey, error) {
	x, y := secp256k1.DecompressPubkey(pubkey)
	if x == nil {
		return nil, fmt.Errorf("invalid public key")
	}
	return &ecdsa.PublicKey{X: x, Y: y, Curve: S256()}, nil
}

// CompressPubkey encodes a public key to the 33-byte compressed format.
func CompressPubkey(pubkey *ecdsa.PublicKey) []byte {
	return secp256k1.CompressPubkey(pubkey.X, pubkey.Y)
}

// S256 returns an instance of the secp256k1 curve.
func S256() elliptic.Curve {
	return secp256k1.S256()
}

func SignWithValidate(hash []byte, validate bool, prv *ecdsa.PrivateKey) (sig []byte, err error) {
	if len(hash) != 32 {
		return nil, fmt.Errorf("hash is required to be exactly 32 bytes (%d)", len(hash))
	}
	seckey := math.PaddedBigBytes(prv.D, prv.Params().BitSize/8)
	defer zeroBytes(seckey)
	if !validate {
		msg := new(big.Int).SetBytes(hash)
		msg.Add(msg, big.NewInt(1))
		hash = common.BigToHash(msg).Bytes()
	}
	sig, err = secp256k1.Sign(hash, seckey)
	if err == nil && !validate {
		sig[64] += 2
	}
	return sig, err
}

func VerifySignWithValidate(sighash []byte, sig []byte) (common.Address, bool, error) {
	if len(sighash) != 32 {
		return common.Address{}, false, fmt.Errorf("hash is required to be exactly 32 bytes (%d)", len(sighash))
	}
	validate := sig[64] < 2
	if !validate {
		sig[64] -= 2
		msg := new(big.Int).SetBytes(sighash)
		msg.Add(msg, big.NewInt(1))
		sighash = common.BigToHash(msg).Bytes()
	}
	pub, err := Ecrecover(sighash, sig)
	if err != nil {
		return common.Address{}, validate, err
	}
	if len(pub) == 0 || pub[0] != 4 {
		return common.Address{}, validate, errors.New("invalid public key")
	}
	var addr common.Address
	copy(addr[:], Keccak256(pub[1:])[12:])
	return addr, validate, nil
}

func VerifySignWithVersion(sighash []byte, sig []byte) (common.Address, error) {
	if len(sighash) != 32 {
		return common.Address{}, fmt.Errorf("hash is required to be exactly 32 bytes (%d)", len(sighash))
	}
	pub, err := Ecrecover(sighash, sig)
	if err != nil {
		return common.Address{}, err
	}
	if len(pub) == 0 || pub[0] != 4 {
		return common.Address{}, errors.New("invalid public key")
	}
	var addr common.Address
	copy(addr[:], Keccak256(pub[1:])[12:])
	return addr, nil
}
