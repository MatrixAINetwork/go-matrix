// Copyright 2018 The MATRIX Authors as well as Copyright 2014-2017 The go-ethereum Authors
// This file is consisted of the MATRIX library and part of the go-2018/10/29ereum library.
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
package blkverify

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/mc"
	"github.com/pkg/errors"

	"sync"
)

const fromLimitCount uint32 = 3

var (
	paramErr           = errors.New("param error")
	countOutOfLimitErr = errors.New("the req count from the account is out of limit")
	leaderReqExistErr  = errors.New("req from this leader already exist")
	cantFindErr        = errors.New("can't find req in cache")
)

type reqData struct {
	req               *mc.HD_BlkConsensusReqMsg
	hash              common.Hash
	txs               types.Transactions
	receipts          []*types.Receipt
	stateDB           *state.StateDB
	localReq          bool
	localVerifyResult uint8
}

func newReqData(req *mc.HD_BlkConsensusReqMsg) *reqData {
	return &reqData{
		req:               req,
		hash:              common.Hash{},
		txs:               nil,
		receipts:          nil,
		stateDB:           nil,
		localReq:          false,
		localVerifyResult: localVerifyResultProcessing,
	}
}

func newReqDataByLocalReq(localReq *mc.LocalBlockVerifyConsensusReq) *reqData {
	return &reqData{
		req:               localReq.BlkVerifyConsensusReq,
		hash:              common.Hash{},
		txs:               localReq.Txs,
		receipts:          localReq.Receipts,
		stateDB:           localReq.State,
		localReq:          true,
		localVerifyResult: localVerifyResultProcessing,
	}
}

type reqCache struct {
	mu             sync.RWMutex
	leaderReqCache map[common.Address]*reqData //from = leader 的req
	otherReqCache  []*reqData                  //from != leader 的req
	countMap       map[common.Address]uint32
	countLimit     uint32
}

func newReqCache() *reqCache {
	return &reqCache{
		countMap:       make(map[common.Address]uint32),
		countLimit:     fromLimitCount,
		otherReqCache:  make([]*reqData, 0),
		leaderReqCache: make(map[common.Address]*reqData, 0),
	}
}

func (rc *reqCache) AddReq(req *mc.HD_BlkConsensusReqMsg) error {
	if nil == req {
		return paramErr
	}

	rc.mu.Lock()
	defer rc.mu.Unlock()
	if req.Header.Leader == req.From {
		_, exit := rc.leaderReqCache[req.From]
		if exit {
			return leaderReqExistErr
		}
		rc.leaderReqCache[req.From] = newReqData(req)
		return nil
	}

	//other req
	count := rc.getCount(req.From)
	if count >= rc.countLimit {
		return countOutOfLimitErr
	}

	rc.otherReqCache = append(rc.otherReqCache, newReqData(req))
	rc.setCount(req.From, count+1)
	return nil
}

func (rc *reqCache) AddLocalReq(req *mc.LocalBlockVerifyConsensusReq) error {
	if nil == req {
		return paramErr
	}

	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.leaderReqCache[req.BlkVerifyConsensusReq.Header.Leader] = newReqDataByLocalReq(req)
	return nil
}

func (rc *reqCache) GetLeaderReq(leader common.Address) (*reqData, error) {
	if (leader == common.Address{}) {
		return nil, paramErr
	}

	rc.mu.RLock()
	defer rc.mu.RUnlock()
	req, OK := rc.leaderReqCache[leader]
	if !OK {
		return nil, cantFindErr
	}
	return req, nil
}

func (rc *reqCache) GetAllReq() []*reqData {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	result := make([]*reqData, 0, len(rc.leaderReqCache)+cap(rc.otherReqCache))
	for _, req := range rc.leaderReqCache {
		result = append(result, req)
	}
	result = append(result, rc.otherReqCache...)
	return result
}

func (rc *reqCache) getCount(from common.Address) uint32 {
	count, OK := rc.countMap[from]
	if OK {
		return count
	} else {
		return 0
	}
}

func (rc *reqCache) setCount(from common.Address, count uint32) {
	rc.countMap[from] = count
}
