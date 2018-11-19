// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package vrf

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/matrix/go-matrix/baseinterface"
)

type vrfWithHash struct {
}

func newVrfWithHash() baseinterface.VrfInterface {
	return &vrfWithHash{}
}
func init() {
	baseinterface.RegVrf("withHash", newVrfWithHash)
}
func (self *vrfWithHash) ComputeVrf(sk *ecdsa.PrivateKey, prevVrf []byte) ([]byte, []byte, error) {
	return Vrf(sk, prevVrf)
}

func (self *vrfWithHash) VerifyVrf(pk *ecdsa.PublicKey, prevVrf, newVrf, proof []byte) error {
	result, err := Verify(pk, prevVrf, newVrf, proof)
	if err != nil {
		return fmt.Errorf("verifyVrf failed: %s", err)
	}
	if !result {
		return fmt.Errorf("verifyVrf failed")
	}
	return nil
}
