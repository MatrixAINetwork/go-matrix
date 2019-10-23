// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package types

import (
	"math/big"

	"github.com/MatrixAINetwork/go-matrix/common"
)

type HeaderV1 struct {
	ParentHash  common.Hash        `json:"parentHash"       gencodec:"required"`
	UncleHash   common.Hash        `json:"sha3Uncles"       gencodec:"required"`
	Leader      common.Address     `json:"leader"            gencodec:"required"`
	Coinbase    common.Address     `json:"miner"            gencodec:"required"`
	Roots       []common.CoinRoot  `json:"stateRoot"        gencodec:"required"`
	Sharding    []common.Coinbyte  `json:"sharding"        gencodec:"required"`
	Difficulty  *big.Int           `json:"difficulty"       gencodec:"required"`
	Number      *big.Int           `json:"number"           gencodec:"required"`
	GasLimit    uint64             `json:"gasLimit"         gencodec:"required"`
	GasUsed     uint64             `json:"gasUsed"          gencodec:"required"`
	Time        *big.Int           `json:"timestamp"        gencodec:"required"`
	Elect       []common.Elect     `json:"elect"        gencodec:"required"`
	NetTopology common.NetTopology `json:"nettopology"        gencodec:"required"`
	Signatures  []common.Signature `json:"signatures"        gencodec:"required"`

	Extra             []byte             `json:"extraData"        gencodec:"required"`
	MixDigest         common.Hash        `json:"mixHash"          gencodec:"required"`
	Nonce             BlockNonce         `json:"nonce"            gencodec:"required"`
	Version           []byte             `json:"version"              gencodec:"required"`
	VersionSignatures []common.Signature `json:"versionSignatures"              gencodec:"required"`
	VrfValue          []byte             `json:"vrfvalue"        gencodec:"required"`
}

func (oh *HeaderV1) TransferHeader() *Header {
	var h Header
	h.ParentHash = oh.ParentHash
	h.UncleHash = oh.UncleHash
	h.Leader = oh.Leader
	h.Coinbase = oh.Coinbase
	h.Roots = oh.Roots
	h.Sharding = oh.Sharding
	h.Difficulty = oh.Difficulty
	h.Number = oh.Number
	h.GasLimit = oh.GasLimit
	h.GasUsed = oh.GasUsed
	h.Time = oh.Time
	h.Elect = oh.Elect
	h.NetTopology = oh.NetTopology
	h.Signatures = oh.Signatures
	h.AIHash = common.Hash{}
	h.Extra = oh.Extra
	h.MixDigest = oh.MixDigest
	h.Nonce = oh.Nonce
	h.Version = oh.Version
	h.VersionSignatures = oh.VersionSignatures
	h.VrfValue = oh.VrfValue
	return &h
}
