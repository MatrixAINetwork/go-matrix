// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package mc

import (
	"math/big"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/core/vm"
	"github.com/MatrixAINetwork/go-matrix/p2p/discover"
)

//common

type HdRev struct {
	FromNodeId string
	Input      interface{}
}

type BlockData struct {
	Header *types.Header
	Txs    []types.CoinSelfTransaction
}

//Miner Module
type HD_MiningReqMsg struct {
	From   common.Address
	Header *types.Header
}

type HD_MiningRspMsg struct {
	From       common.Address
	Number     uint64
	BlockHash  common.Hash
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

type Alternative struct {
	A        common.Address
	Position uint16
}

type TopologyNodeInfo struct {
	Account    common.Address
	Position   uint16
	Type       common.RoleType
	NodeNumber uint8 //0-99
}

type TopologyGraph struct {
	NodeList      []TopologyNodeInfo
	CurNodeNumber uint8
}

type ElectNodeInfo struct {
	Account  common.Address
	Position uint16
	Stock    uint16
	VIPLevel common.VIPRoleType
	Type     common.RoleType
}

type ElectGraph struct {
	Number             uint64
	ElectList          []ElectNodeInfo
	NextMinerElect     []ElectNodeInfo
	NextValidatorElect []ElectNodeInfo
}

type ElectOnlineStatus struct {
	Number      uint64
	ElectOnline []ElectNodeInfo
}

//矿工主节点生成请求
type MasterMinerReElectionReqMsg struct {
	SeqNum      uint64
	RandSeed    *big.Int
	MinerList   []vm.DepositDetail
	ElectConfig ElectConfigInfo_All
}

//验证者主节点生成请求
type MasterValidatorReElectionReqMsg struct {
	SeqNum                  uint64
	RandSeed                *big.Int
	ValidatorList           []vm.DepositDetail
	FoundationValidatorList []vm.DepositDetail
	ElectConfig             ElectConfigInfo_All
	VIPList                 []VIPConfig
	BlockProduceBlackList   BlockProduceSlashBlackList
}

//矿工主节点生成响应
type MasterMinerReElectionRsp struct {
	SeqNum      uint64
	MasterMiner []ElectNodeInfo
}

//验证者主节点生成响应
type MasterValidatorReElectionRsq struct {
	SeqNum             uint64
	MasterValidator    []ElectNodeInfo
	BackUpValidator    []ElectNodeInfo
	CandidateValidator []ElectNodeInfo
}

type RoleUpdatedMsg struct {
	Role      common.RoleType
	BlockNum  uint64
	BlockHash common.Hash
	Leader    common.Address
	SuperSeq  uint64
}

type LeaderChangeNotify struct {
	ConsensusState bool //共识结果
	PreLeader      common.Address
	Leader         common.Address
	NextLeader     common.Address
	Number         uint64
	ConsensusTurn  ConsensusTurnInfo
	ReelectTurn    uint32
	TurnBeginTime  int64
	TurnEndTime    int64
}

//block verify server
type HD_BlkConsensusReqMsg struct {
	From                   common.Address
	Header                 *types.Header
	ConsensusTurn          ConsensusTurnInfo
	TxsCode                []*common.RetCallTxN
	OnlineConsensusResults []*HD_OnlineConsensusVoteResultMsg
}

type LocalBlockVerifyConsensusReq struct {
	BlkVerifyConsensusReq *HD_BlkConsensusReqMsg
	OriginalTxs           []types.CoinSelfTransaction // 原始交易列表
	FinalTxs              []types.CoinSelfTransaction // 最终交易列表(含奖励交易)
	Receipts              []types.CoinReceipts        // 收据
	State                 *state.StateDBManage        // apply state changes here 状态数据库
}

type BlockPOSFinishedNotify struct {
	Number        uint64
	Header        *types.Header // 包含签名列表的header
	ConsensusTurn ConsensusTurnInfo
	TxsCode       []*common.RetCallTxN
}

type BlockLocalVerifyOK struct {
	Header      *types.Header // 包含签名列表的header
	BlockHash   common.Hash
	OriginalTxs []types.CoinSelfTransaction // 原始交易列表
	FinalTxs    []types.CoinSelfTransaction // 最终交易列表(含奖励交易)
	Receipts    []types.CoinReceipts        // 收据
	State       *state.StateDBManage        // apply state changes here 状态数据库
}

//BolckGenor
type HD_BlockInsertNotify struct {
	From   common.Address
	Header *types.Header
}

type NewBlockReadyMsg struct {
	Header *types.Header
	State  *state.StateDBManage
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

type OnlineState uint8

const (
	OnLine OnlineState = iota + 1
	OffLine
)

//在线状态共识请求
type OnlineConsensusReq struct {
	Number      uint64         // 高度
	LeaderTurn  uint32         // leader轮次
	Leader      common.Address // leader地址
	Node        common.Address // node 地址
	OnlineState OnlineState    //在线状态
}

//在线状态共识请求消息
type HD_OnlineConsensusReqs struct {
	From    common.Address
	ReqList []*OnlineConsensusReq //请求结构
}

//共识投票消息
type HD_ConsensusVote struct {
	SignHash common.Hash
	Number   uint64
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
	From     common.Address
}

//特殊交易
type BroadCastEvent struct {
	Txtyps string
	Height *big.Int
	Data   []byte
}

//
type ConsensusTurnInfo struct {
	PreConsensusTurn uint32 // 前一次共识轮次
	UsedReelectTurn  uint32 // 完成共识花费的重选轮次
}

type HD_ReelectInquiryReqMsg struct {
	Number        uint64
	ConsensusTurn ConsensusTurnInfo
	ReelectTurn   uint32
	TimeStamp     uint64
	Master        common.Address
	From          common.Address
}

type ReelectRSPType uint8

const (
	ReelectRSPTypeNone ReelectRSPType = iota
	ReelectRSPTypePOS
	ReelectRSPTypeAlreadyRL
	ReelectRSPTypeAgree
	ReelectRSPTypeNewBlockReady
)

type HD_ReelectInquiryRspMsg struct {
	Number    uint64
	ReqHash   common.Hash
	Type      ReelectRSPType
	AgreeSign common.Signature
	POSResult *HD_BlkConsensusReqMsg
	RLResult  *HD_ReelectLeaderConsensus
	NewBlock  *types.Header
	From      common.Address
}

type HD_ReelectLeaderReqMsg struct {
	InquiryReq *HD_ReelectInquiryReqMsg
	AgreeSigns []common.Signature
	TimeStamp  uint64
}

type HD_ReelectLeaderConsensus struct {
	Req   *HD_ReelectLeaderReqMsg
	Votes []common.Signature
}

type HD_ReelectBroadcastMsg struct {
	Number    uint64
	Type      ReelectRSPType
	POSResult *HD_BlkConsensusReqMsg
	RLResult  *HD_ReelectLeaderConsensus
	TimeStamp uint64
	From      common.Address
}

type HD_ReelectBroadcastRspMsg struct {
	Number     uint64
	ResultHash common.Hash
	Sign       common.Signature
	From       common.Address
}

type RecoveryType uint8

const (
	RecoveryTypePOS RecoveryType = iota
	RecoveryTypeFullHeader
)

type RecoveryStateMsg struct {
	Type        RecoveryType
	IsBroadcast bool
	Header      *types.Header
	From        common.Address
}

type HD_FullBlockReqMsg struct {
	HeaderHash common.Hash
	Number     uint64
	From       common.Address
}

type HD_FullBlockRspMsg struct {
	Header *types.Header
	Txs    []types.CoinSelfTransaction
	From   common.Address
}

type EveryBlockSeedRspMsg struct {
	PublicKey []byte
	Private   []byte
}
type VrfMsg struct {
	VrfValue []byte
	VrfProof []byte
	Hash     common.Hash
}
type EntrustInfo struct {
	Address  string
	Password string
}

type BlockInfo struct {
	Hash   common.Hash
	Number uint64
}

type BlockInsertedMsg struct {
	Block      BlockInfo
	InsertTime uint64
	CanonState bool
}
