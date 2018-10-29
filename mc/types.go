// Copyright 2018 The MATRIX Authors as well as Copyright 2014-2017 The go-ethereum Authors
// This file is consisted of the MATRIX library and part of the go-ethereum library.
//
// The MATRIX-ethereum library is free software: you can redistribute it and/or modify it under the terms of the MIT License.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, 
//and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject tothe following conditions:
//
//The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
//
//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, 
//WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISINGFROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE
//OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
package mc

import (
	"math/big"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/p2p/discover"
)

//by hezi //YY 2018-08-18由tx_pool.go转移到此
const (
	Heartbeat   = "Heartbeat"      // 心跳交易（广播区块Hash对99取余）
	Publickey   = "SeedPublicKey"  // 公钥交易
	Privatekey  = "SeedPrivateKey" // 私钥交易
	CallTheRoll = "CallTheRoll"    //点名交易  （广播节点随机连接1000个点）
)

type BlockToBucket struct {
	Ms     []discover.NodeID
	Height *big.Int
	Role   common.RoleType
}

type BlockToLinker struct {
	Height *big.Int
	Role   common.RoleType
}

//Election Module
type MasterMinerReElectionRspMsg struct {
	SeqNum uint64
	//MasterMinerList []election.ElectionResultInfo
	//BackUpMinerList []election.ElectionResultInfo
}
type MasterValidatorReElectionRspMsg struct {
	SeqNum uint64
	//MasterValidatorList          []election.ElectionResultInfo
	//BackUpMasterValidatorList    []election.ElectionResultInfo
	//CandidateMasterValidatorList []election.ElectionResultInfo
}

type LeaderStateMsg struct {
	Leader      common.Address
	Number      big.Int
	ReelectTurn uint8
}

// type BlockVerifyReqMsg struct {
// 	Header  types.Header
// 	TxsCode []uint32
// }

// type BlockVerifyResultMsg struct {
// 	Header  *types.Header // 包含签名列表的header
// 	TxsCode []uint32      // 交易列表
// }

// type VoteMsg struct {
// 	SignHash common.Hash
// 	//Sign        common.Signature
// 	FromAccount common.Address
// }

// // ReElection
// type RandomSeedReq struct {
// 	Height *big.Int
// }
// type ElectionSeedRsp struct {
// 	Seed common.Hash
// }
