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
package verifier

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/params/man"
	"github.com/pkg/errors"
	"time"
)

type msgPool struct {
	parentHeader     *types.Header
	posNotifyCache   []*mc.BlockPOSFinishedNotify
	inquiryReqCache  map[common.Address]*mc.HD_ReelectInquiryReqMsg
	rlConsensusCache map[uint32]*mc.HD_ReelectLeaderConsensus
}

func newMsgPool() *msgPool {
	return &msgPool{
		parentHeader:     nil,
		posNotifyCache:   make([]*mc.BlockPOSFinishedNotify, 0),
		inquiryReqCache:  make(map[common.Address]*mc.HD_ReelectInquiryReqMsg),
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

func (mp *msgPool) SaveInquiryReqMsg(msg *mc.HD_ReelectInquiryReqMsg) {
	if nil == msg || (msg.From == common.Address{}) {
		return
	}

	old, exist := mp.inquiryReqCache[msg.From]
	if exist && old.TimeStamp > msg.TimeStamp {
		return
	}
	mp.inquiryReqCache[msg.From] = msg
}

func (mp *msgPool) GetInquiryReqMsg(leader common.Address) (*mc.HD_ReelectInquiryReqMsg, error) {
	msg, OK := mp.inquiryReqCache[leader]
	if !OK {
		return nil, ErrNoMsgInCache
	}

	passTime := time.Now().Unix() - msg.TimeStamp
	if passTime > man.LRSReelectInterval {
		delete(mp.inquiryReqCache, leader)
		return nil, errors.Errorf("消息已过期, timestamp=%d, passTime=%d", msg.TimeStamp, passTime)
	}
	return msg, nil
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
