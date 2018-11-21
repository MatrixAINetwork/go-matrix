/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

//This package is a wrapper of verifiable random function using curve secp256r1.
package vrf

import (
	"crypto"
	"crypto/elliptic"
	"errors"
	"hash"

	"crypto/ecdsa"
	"github.com/matrix/go-matrix/log"
)

var (
	ErrKeyNotSupported = errors.New("only support ECC key")
	ErrEvalVRF         = errors.New("failed to evaluate vrf")
)

//Vrf returns the verifiable random function evaluated m and a NIZK proof
func Vrf(pri *ecdsa.PrivateKey, msg []byte) (vrf, nizk []byte, err error) {
	h := getHash(pri.Curve)
	if h == nil {
		return nil, nil, ErrKeyNotSupported
	}
	byteLen := (pri.Params().BitSize + 7) >> 3
	_, proof := Evaluate(pri, h, msg)
	if proof == nil {
		return nil, nil, ErrEvalVRF
	}

	nizk = proof[0 : 2*byteLen]
	vrf = proof[2*byteLen : 2*byteLen+2*byteLen+1]
	err = nil
	return
}

//Verify returns true if vrf and nizk is correct for msg
func Verify(pub *ecdsa.PublicKey, msg, vrf, nizk []byte) (bool, error) {
	h := getHash(pub.Curve)
	if h == nil {
		return false, ErrKeyNotSupported
	}
	byteLen := (pub.Params().BitSize + 7) >> 3
	if len(vrf) != byteLen*2+1 || len(nizk) != byteLen*2 {
		return false, nil
	}
	proof := append(nizk, vrf...)
	_, err := ProofToHash(pub, h, msg, proof)
	if err != nil {
		log.Error("verifying VRF failed: %v", err)
		return false, nil
	}
	return true, nil
}


func getHash(curve elliptic.Curve) hash.Hash {
	bitSize := curve.Params().BitSize
	switch bitSize {
	case 224:
		return crypto.SHA224.New()
	case 256:
		return crypto.SHA256.New()
	case 384:
		return crypto.SHA384.New()
	default:
		return nil
	}
}
