// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package vrf

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"hash"
	"math/big"

	"github.com/MatrixAINetwork/go-matrix/crypto/vrf/ec"
)

var (
	ErrInvalidVRF  = errors.New("invalid VRF proof")
	ErrInvalidHash = errors.New("hash function does not match elliptic curve bitsize")
)

// hashToCurve hashes to a point on elliptic curve
func hashToCurve(curve elliptic.Curve, h hash.Hash, m []byte) (x, y *big.Int) {
	var i uint32
	params := curve.Params()
	byteLen := (params.BitSize + 7) >> 3
	for x == nil && i < 100 {
		// TODO: Use a NIST specified DRBG.
		h.Reset()
		binary.Write(h, binary.BigEndian, i)
		h.Write(m)
		r := []byte{2} // Set point encoding to "compressed", y=0.
		r = h.Sum(r)
		p, err := ec.DecodePublicKey(r[:byteLen+1], curve)
		if err != nil {
			x, y = nil, nil
		} else {
			x, y = p.X, p.Y
		}
		i++
	}
	return
}

var one = big.NewInt(1)

// hashToInt hashes to an integer [1,N-1]
func hashToInt(curve elliptic.Curve, h hash.Hash, m []byte) *big.Int {
	// NIST SP 800-90A § A.5.1: Simple discard method.
	params := curve.Params()
	byteLen := (params.BitSize + 7) >> 3
	for i := uint32(0); ; i++ {
		// TODO: Use a NIST specified DRBG.
		h.Reset()
		binary.Write(h, binary.BigEndian, i)
		h.Write(m)
		b := h.Sum(nil)
		k := new(big.Int).SetBytes(b[:byteLen])
		if k.Cmp(new(big.Int).Sub(params.N, one)) == -1 {
			return k.Add(k, one)
		}
	}
}

// Evaluate returns the verifiable unpredictable(random) function evaluated at m
func Evaluate(pri *ecdsa.PrivateKey, h hash.Hash, m []byte) (index [32]byte, proof []byte) {
	curve := pri.Curve
	params := curve.Params()
	nilIndex := [32]byte{}

	byteLen := (params.BitSize + 7) >> 3
	if byteLen != h.Size() {
		return nilIndex, nil
	}
	// Prover chooses r <-- [1,N-1]
	r, _, _, err := elliptic.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nilIndex, nil
	}
	ri := new(big.Int).SetBytes(r)

	// H = hashToCurve(pk || m)
	var buf bytes.Buffer
	buf.Write(elliptic.Marshal(curve, pri.PublicKey.X, pri.PublicKey.Y))
	buf.Write(m)
	Hx, Hy := hashToCurve(curve, h, buf.Bytes())

	// VRF_pri(m) = [pri]H
	sHx, sHy := curve.ScalarMult(Hx, Hy, pri.D.Bytes())
	vrf := elliptic.Marshal(curve, sHx, sHy) // 2*byteLen+1 bytes.

	// G is the base point
	// s = hashToInt(G, H, [pri]G, VRF, [r]G, [r]H)
	rGx, rGy := curve.ScalarBaseMult(r)
	rHx, rHy := curve.ScalarMult(Hx, Hy, r)
	var b bytes.Buffer
	b.Write(elliptic.Marshal(curve, params.Gx, params.Gy))
	b.Write(elliptic.Marshal(curve, Hx, Hy))
	b.Write(elliptic.Marshal(curve, pri.PublicKey.X, pri.PublicKey.Y))
	b.Write(vrf)
	b.Write(elliptic.Marshal(curve, rGx, rGy))
	b.Write(elliptic.Marshal(curve, rHx, rHy))
	s := hashToInt(curve, h, b.Bytes())

	// t = r−s*pri mod N
	t := new(big.Int).Sub(ri, new(big.Int).Mul(s, pri.D))
	t.Mod(t, params.N)

	// Index = SHA256(vrf)
	index = sha256.Sum256(vrf)

	// Write s, t, and vrf to a proof blob. Also write leading zeros before s and t
	// if needed.
	buf.Reset()
	buf.Write(make([]byte, byteLen-len(s.Bytes())))
	buf.Write(s.Bytes())
	buf.Write(make([]byte, byteLen-len(t.Bytes())))
	buf.Write(t.Bytes())
	buf.Write(vrf) //byteLen*2 + byteLen*2 + 1

	return index, buf.Bytes()
}

// ProofToHash asserts that proof is correct for m and outputs index.
func ProofToHash(pk *ecdsa.PublicKey, h hash.Hash, m, proof []byte) (index [32]byte, err error) {
	nilIndex := [32]byte{}
	curve := pk.Curve
	params := curve.Params()
	byteLen := (params.BitSize + 7) >> 3
	if byteLen != h.Size() {
		return nilIndex, ErrInvalidHash
	}

	// verifier checks that s == hashToInt(m, [t]G + [s]([pri]G), [t]hashToCurve(pk, m) + [s]VRF_pri(m))
	if got, want := len(proof), (2*byteLen)+(2*byteLen+1); got != want {
		return nilIndex, ErrInvalidVRF
	}

	// Parse proof into s, t, and vrf.
	s := proof[0:byteLen]
	t := proof[byteLen : 2*byteLen]
	vrf := proof[2*byteLen : 2*byteLen+2*byteLen+1]

	uHx, uHy := elliptic.Unmarshal(curve, vrf)
	if uHx == nil {
		return nilIndex, ErrInvalidVRF
	}

	// [t]G + [s]([pri]G) = [t+pri*s]G
	tGx, tGy := curve.ScalarBaseMult(t)
	ksGx, ksGy := curve.ScalarMult(pk.X, pk.Y, s)
	tksGx, tksGy := curve.Add(tGx, tGy, ksGx, ksGy)

	// H = hashToCurve(pk || m)
	// [t]H + [s]VRF = [t+pri*s]H
	buf := new(bytes.Buffer)
	buf.Write(elliptic.Marshal(curve, pk.X, pk.Y))
	buf.Write(m)
	Hx, Hy := hashToCurve(pk, h, buf.Bytes())
	tHx, tHy := curve.ScalarMult(Hx, Hy, t)
	sHx, sHy := curve.ScalarMult(uHx, uHy, s)
	tksHx, tksHy := curve.Add(tHx, tHy, sHx, sHy)

	//   hashToInt(G, H, [pri]G, VRF, [t]G + [s]([pri]G), [t]H + [s]VRF)
	// = hashToInt(G, H, [pri]G, VRF, [t+pri*s]G, [t+pri*s]H)
	// = hashToInt(G, H, [pri]G, VRF, [r]G, [r]H)
	var b bytes.Buffer
	b.Write(elliptic.Marshal(curve, params.Gx, params.Gy))
	b.Write(elliptic.Marshal(curve, Hx, Hy))
	b.Write(elliptic.Marshal(curve, pk.X, pk.Y))
	b.Write(vrf)
	b.Write(elliptic.Marshal(curve, tksGx, tksGy))
	b.Write(elliptic.Marshal(curve, tksHx, tksHy))
	h2 := hashToInt(curve, h, b.Bytes())

	// Left pad h2 with zeros if needed. This will ensure that h2 is padded
	// the same way s is.
	buf.Reset()
	buf.Write(make([]byte, byteLen-len(h2.Bytes())))
	buf.Write(h2.Bytes())

	if !hmac.Equal(s, buf.Bytes()) {
		return nilIndex, ErrInvalidVRF
	}
	return sha256.Sum256(vrf), nil
}
