//1542490604.0397766
//1542489831.5459044
//1542489137.9864962
//1542488455.458757
//1542487691.8165412
//1542487056.3058524
// Copyright (c) 2018?The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package verifier

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/mc"
)

type msgPool struct {
	parentHeader     *types.Header
	posNotifyCache   []*mc.BlockPOSFinishedNotify
	rlConsensusCache map[uint32]*mc.HD_ReelectLeaderConsensus
}

func newMsgPool() *msgPool {
	return &msgPool{
		parentHeader:     nil,
		posNotifyCache:   make([]*mc.BlockPOSFinishedNotify, 0),
		rlConsensusCache: make(map[uint32]*mc.HD_ReelectLeaderConsensus),
	}
}

func (mp *msgPool) SavePOSNotifyMsg(msg *mc.BlockPOSFinishedNotify) error {
	if nil == msg || (msg.Header.Leader == common.Address{}) {
		return ErrMsgIsNil
	}

	for _, oldMsg := range mp.posNotifyCache {
		if oldMsg.ConsensusTurn == msg.ConsensusTurn && oldMsg.Header.Leader == msg.Header.Leader {
			return ErrMsgExistInCache
		}
	}

	mp.posNotifyCache = append(mp.posNotifyCache, msg)
	return nil
}

func (mp *msgPool) GetPOSNotifyMsg(leader common.Address, consensusTurn uint32) (*mc.BlockPOSFinishedNotify, error) {
	for _, msg := range mp.posNotifyCache {
		if msg.ConsensusTurn == msg.ConsensusTurn && msg.Header.Leader == msg.Header.Leader {
			return msg, nil
		}
	}
	return nil, ErrNoMsgInCache
}

func (mp *msgPool) SaveRLConsensusMsg(msg *mc.HD_ReelectLeaderConsensus) {
	if nil == msg || (msg.Req.InquiryReq.Master == common.Address{}) {
		return
	}
	consensusTurn := msg.Req.InquiryReq.ConsensusTurn + msg.Req.InquiryReq.ReelectTurn
	mp.rlConsensusCache[consensusTurn] = msg
}

func (mp *msgPool) GetRLConsensusMsg(consensusTurn uint32) (*mc.HD_ReelectLeaderConsensus, error) {
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
