// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
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
