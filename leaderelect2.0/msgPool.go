// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package leaderelect2

import (
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/mc"
)

type msgPool struct {
	parentHeader     *types.Header
	posNotifyCache   []*mc.BlockPOSFinishedNotify
	rlConsensusCache map[mc.ConsensusTurnInfo]*mc.HD_V2_ReelectLeaderConsensus
}

func newMsgPool() *msgPool {
	return &msgPool{
		parentHeader:     nil,
		posNotifyCache:   make([]*mc.BlockPOSFinishedNotify, 0),
		rlConsensusCache: make(map[mc.ConsensusTurnInfo]*mc.HD_V2_ReelectLeaderConsensus),
	}
}

func (mp *msgPool) SavePOSNotifyMsg(msg *mc.BlockPOSFinishedNotify) error {
	if nil == msg || nil == msg.Header || (msg.Header.Leader == common.Address{}) {
		return ErrParamsIsNil
	}

	//存在检查
	for _, oldMsg := range mp.posNotifyCache {
		if oldMsg.ConsensusTurn == msg.ConsensusTurn && oldMsg.Header.Leader == msg.Header.Leader {
			return ErrMsgExistInCache
		}
	}

	mp.posNotifyCache = append(mp.posNotifyCache, msg)
	return nil
}

func (mp *msgPool) GetPOSNotifyMsg(leader common.Address, consensusTurn mc.ConsensusTurnInfo) (*mc.BlockPOSFinishedNotify, error) {
	for _, msg := range mp.posNotifyCache {
		if msg.ConsensusTurn == consensusTurn && msg.Header.Leader == leader {
			return msg, nil
		}
	}
	return nil, ErrNoMsgInCache
}

func (mp *msgPool) SaveRLConsensusMsg(msg *mc.HD_V2_ReelectLeaderConsensus) {
	if nil == msg || nil == msg.Req || nil == msg.Req.InquiryReq || (msg.Req.InquiryReq.Master == common.Address{}) {
		return
	}
	consensusTurn := calcNextConsensusTurn(msg.Req.InquiryReq.ConsensusTurn, msg.Req.InquiryReq.ReelectTurn)
	mp.rlConsensusCache[consensusTurn] = msg
}

func (mp *msgPool) GetRLConsensusMsg(consensusTurn mc.ConsensusTurnInfo) (*mc.HD_V2_ReelectLeaderConsensus, error) {
	msg, OK := mp.rlConsensusCache[consensusTurn]
	if !OK {
		return nil, ErrNoMsgInCache
	}
	return msg, nil
}

func (mp *msgPool) SaveParentHeader(header *types.Header) {
	if nil == header {
		return
	}
	mp.parentHeader = header
}

func (mp *msgPool) GetParentHeader() *types.Header {
	return mp.parentHeader
}
