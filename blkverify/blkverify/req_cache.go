// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
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
	fromAddressErr     = errors.New("the address where req from is illegal")
	cantFindErr        = errors.New("can't find req in cache")
)

type reqData struct {
	req               *mc.HD_BlkConsensusReqMsg
	hash              common.Hash
	txs               types.SelfTransactions
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
	mu         sync.RWMutex
	countMap   map[common.Address]uint32
	countLimit uint32
	cache      []*reqData
}

func newReqCache() *reqCache {
	return &reqCache{
		countMap:   make(map[common.Address]uint32),
		countLimit: fromLimitCount,
		cache:      make([]*reqData, 0),
	}
}

func (rc *reqCache) AddReq(req *mc.HD_BlkConsensusReqMsg) error {
	if nil == req {
		return paramErr
	}

	rc.mu.Lock()
	defer rc.mu.Unlock()

	//todo 判断合法from
	count := rc.getCount(req.From)
	if count >= rc.countLimit {
		return countOutOfLimitErr
	}

	rc.cache = append(rc.cache, newReqData(req))
	rc.setCount(req.From, count+1)
	return nil
}

func (rc *reqCache) AddLocalReq(req *mc.LocalBlockVerifyConsensusReq) error {
	if nil == req {
		return paramErr
	}

	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.cache = append(rc.cache, newReqDataByLocalReq(req))
	return nil
}

func (rc *reqCache) GetLeaderReq(leader common.Address) (*reqData, error) {
	if (leader == common.Address{}) {
		return nil, paramErr
	}

	rc.mu.RLock()
	defer rc.mu.RUnlock()

	for _, req := range rc.cache {
		if req.req.Header.Leader == leader && req.req.From == leader {
			return req, nil
		}
	}
	return nil, cantFindErr
}

func (rc *reqCache) GetAllReq() []*reqData {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	result := make([]*reqData, len(rc.cache))
	copy(result, rc.cache)
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
