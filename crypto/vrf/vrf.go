// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package vrf

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"errors"
	"hash"

	"github.com/MatrixAINetwork/go-matrix/log"
)

var (
	ErrKeyNotSupported = errors.New("only support ECC key")
	ErrEvalVRF         = errors.New("failed to evaluate vrf")
)

func checkPri(pri *ecdsa.PrivateKey) error {
	if pri == nil {
		return errors.New("私钥指针为空")
	}
	if pri.Curve == nil {
		return errors.New("私钥的Curve为空")
	}
	if pri.Params() == nil {
		return errors.New("私钥的Params为空")
	}
	if pri.PublicKey.X == nil {
		return errors.New("私钥对应的Publickey.X为空")
	}
	if pri.PublicKey.Y == nil {
		return errors.New("私钥对应的Publickey.Y为空")
	}
	if pri.D == nil {
		return errors.New("私钥对应的D为空")
	}
	return nil

}
func checkPub(pub *ecdsa.PublicKey) error {
	if pub == nil {
		return errors.New("公钥指针为空")
	}
	if pub.Curve == nil {
		return errors.New("公钥的Curve为空")
	}
	if pub.Params() == nil {
		return errors.New("公钥的Params为空")
	}

	if pub.X == nil {
		return errors.New("公钥对应的Publickey.X为空")
	}
	if pub.Y == nil {
		return errors.New("公钥对应的Publickey.Y为空")
	}

	return nil
}

//Vrf returns the verifiable random function evaluated m and a NIZK proof
func Vrf(pri *ecdsa.PrivateKey, msg []byte) (vrf, nizk []byte, err error) {
	if err := checkPri(pri); err != nil {
		return nil, nil, errors.New("私钥不合法")
	}
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
	if err := checkPub(pub); err != nil {
		return false, errors.New("公钥不合法")
	}
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
		log.Error("verifying VRF failed: %v", "err", err)
		return false, nil
	}
	return true, nil
}

func getHash(curve elliptic.Curve) hash.Hash {

	if curve == nil || curve.Params() == nil {
		return nil
	}
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
