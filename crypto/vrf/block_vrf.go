// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package vrf

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/btcec"

	"github.com/MatrixAINetwork/go-matrix/baseinterface"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/crypto"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
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

func (self *vrfWithHash) verifyVrf(pk *ecdsa.PublicKey, prevVrf, newVrf, proof []byte) error {
	result, err := Verify(pk, prevVrf, newVrf, proof)
	if err != nil {
		return fmt.Errorf("verifyVrf failed: %s", err)
	}
	if !result {
		return fmt.Errorf("verifyVrf failed")
	}
	return nil
}

func (self *vrfWithHash) DecodeVrf(header *types.Header, preHeader *types.Header) (common.Address, error) {
	log.INFO("vrf", "len header.VrfValue", len(header.VrfValue), "data", header.VrfValue, "高度", header.Number.Uint64())
	account, _, _ := self.GetVrfInfoFromHeader(header.VrfValue)

	log.Error("vrf", "从区块头重算出账户户", account, "高度", header.Number.Uint64())

	public := account
	curve := btcec.S256()
	pk1, err := btcec.ParsePubKey(public, curve)
	if err != nil {
		log.Error("vrf转换失败", "err", err, "account", account, "len", len(account))
		return common.Address{}, err
	}

	pk1_1 := (*ecdsa.PublicKey)(pk1)
	_, vrfValue, vrfProof := baseinterface.NewVrf().GetVrfInfoFromHeader(header.VrfValue)

	_, preVrfValue, preVrfProof := self.GetVrfInfoFromHeader(preHeader.VrfValue)

	preMsg := mc.VrfMsg{
		VrfValue: preVrfValue,
		VrfProof: preVrfProof,
		Hash:     header.ParentHash,
	}

	preVrfMsg, err := json.Marshal(preMsg)
	if err != nil {
		log.Error("vrf", "生成vefmsg出错", err, "parentMsg", preVrfMsg)
		return common.Address{}, errors.New("生成vrfmsg出错")
	} else {
		log.Error("生成vrfmsg成功")
	}
	//log.Info("msgggggvrf_verify","preVrfMsg",preVrfMsg,"高度",header.Number.Uint64(),"VrfProof",preMsg.VrfProof,"VrfValue",preMsg.VrfValue,"Hash",preMsg.Hash)
	if err := self.verifyVrf(pk1_1, preVrfMsg, vrfValue, vrfProof); err != nil {
		log.Error("vrf verify ", "err", err)
		return common.Address{}, err
	}

	ans := crypto.PubkeyToAddress(*pk1_1)
	return ans, nil
}

func (self *vrfWithHash) GetHeaderVrf(account []byte, vrfvalue []byte, vrfproof []byte) []byte {
	var buf bytes.Buffer
	buf.Write(account)
	buf.Write(vrfvalue)
	buf.Write(vrfproof)

	return buf.Bytes()

}

func (self *vrfWithHash) GetVrfInfoFromHeader(headerVrf []byte) ([]byte, []byte, []byte) {
	var account, vrfvalue, vrfproof []byte
	if len(headerVrf) >= 33 {
		account = headerVrf[0:33]
	}
	if len(headerVrf) >= 33+65 {
		vrfvalue = headerVrf[33 : 33+65]
	}
	if len(headerVrf) >= 33+65+64 {
		vrfproof = headerVrf[33+65 : 33+65+64]
	}

	return account, vrfvalue, vrfproof
}
