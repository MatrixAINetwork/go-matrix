// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package mc

import (
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"math/big"
)

type HD_V2_ReelectInquiryReqMsg struct {
	Number        uint64
	HeaderTime    uint64
	ConsensusTurn ConsensusTurnInfo
	ReelectTurn   uint32
	TimeStamp     uint64
	Master        common.Address
	From          common.Address
}

type HD_V2_ReelectInquiryRspMsg struct {
	Number    uint64
	ReqHash   common.Hash
	Type      ReelectRSPType
	AgreeSign common.Signature
	POSResult *HD_BlkConsensusReqMsg
	RLResult  *HD_V2_ReelectLeaderConsensus
	NewBlock  *types.Header
	From      common.Address
}

type HD_V2_ReelectLeaderReqMsg struct {
	InquiryReq *HD_V2_ReelectInquiryReqMsg
	AgreeSigns []common.Signature
	TimeStamp  uint64
}

//共识投票消息
type HD_V2_ConsensusVote struct {
	SignHash common.Hash
	Number   uint64
	Sign     common.Signature
	From     common.Address
}

type HD_V2_ReelectLeaderConsensus struct {
	Req   *HD_V2_ReelectLeaderReqMsg
	Votes []common.Signature
}

type HD_V2_ReelectBroadcastMsg struct {
	Number    uint64
	Type      ReelectRSPType
	POSResult *HD_BlkConsensusReqMsg
	RLResult  *HD_V2_ReelectLeaderConsensus
	TimeStamp uint64
	From      common.Address
}

type HD_V2_ReelectBroadcastRspMsg struct {
	Number     uint64
	ResultHash common.Hash
	Sign       common.Signature
	From       common.Address
}

type HD_V2_MiningReqMsg struct {
	From   common.Address
	Header *types.Header
}

type HD_V2_AIMiningRspMsg struct {
	From       common.Address
	Number     uint64
	BlockHash  common.Hash
	AIHash     common.Hash
	AICoinbase common.Address
}

type HD_V2_PowMiningRspMsg struct {
	From       common.Address
	Number     uint64
	BlockHash  common.Hash
	Difficulty *big.Int
	Nonce      types.BlockNonce
	Coinbase   common.Address
	MixDigest  common.Hash
	Sm3Nonce   types.BlockNonce
}

type BlockPOSFinishedV2 struct {
	Header      *types.Header // 包含签名列表的header
	BlockHash   common.Hash
	OriginalTxs []types.CoinSelfTransaction // 原始交易列表
	FinalTxs    []types.CoinSelfTransaction // 最终交易列表(含奖励交易)
	Receipts    []types.CoinReceipts        // 收据
	State       *state.StateDBManage        // apply state changes here 状态数据库
}

type HD_V2_FullBlockReqMsg struct {
	HeaderHash common.Hash
	Number     uint64
	From       common.Address
}

type HD_V2_FullBlockRspMsg struct {
	Header *types.Header
	Txs    []types.CoinSelfTransaction
	From   common.Address
}

type HD_BasePowerDifficulty HD_V2_PowMiningRspMsg
