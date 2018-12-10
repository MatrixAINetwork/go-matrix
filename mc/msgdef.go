// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package mc

import (
	"math/big"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/core/vm"
	"github.com/matrix/go-matrix/p2p/discover"
)

//common

type HdRev struct {
	FromNodeId string
	Input      interface{}
}

type BlockData struct {
	Header *types.Header
	Txs    types.SelfTransactions
}

//Miner Module
type HD_MiningReqMsg struct {
	From   common.Address
	Header *types.Header
}

type HD_MiningRspMsg struct {
	From       common.Address
	Number     uint64
	Blockhash  common.Hash
	Difficulty *big.Int
	Nonce      types.BlockNonce
	Coinbase   common.Address
	MixDigest  common.Hash
	Signatures []common.Signature
}

type BlockGenor_BroadcastMiningReqMsg struct {
	BlockMainData *BlockData
}

type HD_BroadcastMiningRspMsg struct {
	From          common.Address
	BlockMainData *BlockData
}

//拓扑生成模块
type DepositDetail struct {
	Address    common.Address
	NodeID     discover.NodeID
	Deposit    *big.Int
	WithdrawH  *big.Int
	OnlineTime *big.Int
}
type TopologyNodeInfo struct {
	Account  common.Address
	Position uint16
	Type     common.RoleType
	Stock    uint16
	//	OnlineState bool
}
type Alternative struct {
	A        common.Address
	B        common.Address
	Position uint16
}
type TopologyGraph struct {
	Number   *big.Int
	NodeList []TopologyNodeInfo
}

//矿工主节点生成请求
type MasterMinerReElectionReqMsg struct {
	SeqNum    uint64
	RandSeed  *big.Int
	MinerList []vm.DepositDetail
}

//验证者主节点生成请求
type MasterValidatorReElectionReqMsg struct {
	SeqNum                  uint64
	RandSeed                *big.Int
	ValidatorList           []vm.DepositDetail
	FoundationValidatoeList []vm.DepositDetail
}

//矿工主节点生成响应
type MasterMinerReElectionRsp struct {
	SeqNum      uint64
	MasterMiner []TopologyNodeInfo
	BackUpMiner []TopologyNodeInfo
}

//验证者主节点生成响应
type MasterValidatorReElectionRsq struct {
	SeqNum             uint64
	MasterValidator    []TopologyNodeInfo
	BackUpValidator    []TopologyNodeInfo
	CandidateValidator []TopologyNodeInfo
}
type RoleUpdatedMsg struct {
	Role     common.RoleType
	BlockNum uint64
	Leader   common.Address
}

type NetDataMsg struct {
	From common.Address
	Data interface{}
}

type LeaderChangeNotify struct {
	ConsensusState bool //共识结果
	Leader         common.Address
	NextLeader     common.Address
	Number         uint64
	ReelectTurn    uint8
}

//block verify server
type HD_BlkConsensusReqMsg struct {
	From    common.Address
	Header  *types.Header
	TxsCode []uint32
}

type LocalBlockVerifyConsensusReq struct {
	BlkVerifyConsensusReq *HD_BlkConsensusReqMsg
	Txs                   types.SelfTransactions // 交易列表
	Receipts              []*types.Receipt   // 收据
	State                 *state.StateDB     // apply state changes here 状态数据库
}

type BlockVerifyStateNotify struct {
	Leader common.Address
	Number uint64
	State  bool // True: begin verify, False: end verify
}

type BlockVerifyConsensusOK struct {
	Header    *types.Header // 包含签名列表的header
	BlockHash common.Hash
	Txs       types.SelfTransactions // 交易列表
	Receipts  []*types.Receipt   // 收据
	State     *state.StateDB     // apply state changes here 状态数据库
}

//BolckGenor
type HeaderGenerateReq struct {
	Height uint64
}
type HD_BlockInsertNotify struct {
	From   common.Address
	Header *types.Header
}
type HeaderGenNotify struct {
	Leader common.Address
	Height uint64
}

type NewBlockReady struct {
	Leader     common.Address
	Number     uint64
	Validators *TopologyGraph
}

type PreBlockBroadcastFinished struct {
	BlockHash common.Hash
	Number    uint64
}

//随机数生成请求
type RandomRequest struct {
	MinHash    common.Hash
	PrivateMap map[common.Address][]byte
	PublicMap  map[common.Address][]byte
}

//随机数生成响应
type ElectionEvent struct {
	Seed *big.Int
}

//在线状态共识请求
type OnlineConsensusReq struct {
	Leader      common.Address //leader地址
	Seq         uint64         //共识轮次
	Node        common.Address // node 地址
	OnlineState int            //在线状态
}

//在线状态共识请求消息
type HD_OnlineConsensusReqs struct {
	From    common.Address
	ReqList []*OnlineConsensusReq //请求结构
}

//共识投票消息
type HD_ConsensusVote struct {
	SignHash common.Hash
	Round    uint64
	Sign     common.Signature
	From     common.Address
}

type HD_OnlineConsensusVotes struct {
	Votes []HD_ConsensusVote
}

//共识结果
type HD_OnlineConsensusVoteResultMsg struct {
	Req      *OnlineConsensusReq //请求结构
	SignList []common.Signature  //签名列表
}

type ReelectLeaderDoneMsg struct {
	Leader      common.Address
	ReelectTurn uint8
}

//重选leader 主节点会话启动或完成消息
type LeaderReelectMsg struct {
	Leader      common.Address
	Number      uint64
	ReelectTurn uint8
}

//重选leader从节点会话启动或完成消息
type FollowerReelectMsg struct {
	Leader      common.Address
	Number      uint64
	ReelectTurn uint8
}

type HD_LeaderReelectVoteReqMsg struct {
	Leader      common.Address
	Height      uint64
	ReelectTurn uint8
	TimeStamp   uint64
}

type HD_LeaderReelectConsensusBroadcastMsg struct {
	Req        HD_LeaderReelectVoteReqMsg
	Signatures []common.Signature
}

//重选leader成功消息
type ReelectLeaderSuccMsg struct {
	Height      uint64
	Leader      common.Address
	ReelectTurn uint8
}

//特殊交易
type BroadCastEvent struct {
	Txtyps string
	Height *big.Int
	Data   []byte
}
