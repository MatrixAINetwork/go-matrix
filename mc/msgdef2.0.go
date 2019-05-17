// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package mc

import (
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/types"
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
