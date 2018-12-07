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
	"sort"
)

const otherReqCountMax = 20

var (
	paramErr          = errors.New("param error")
	leaderReqExistErr = errors.New("req from this leader already exist")
	cantFindErr       = errors.New("can't find req in cache")
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
		hash:              req.Header.HashNoSignsAndNonce(),
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
		hash:              localReq.BlkVerifyConsensusReq.Header.HashNoSignsAndNonce(),
		txs:               localReq.Txs,
		receipts:          localReq.Receipts,
		stateDB:           localReq.State,
		localReq:          true,
		localVerifyResult: localVerifyResultProcessing,
	}
}

type reqCache struct {
	mu             sync.RWMutex
	curTurn        uint32
	leaderReqCache map[common.Address]*reqData //from = leader 的req
	otherReqCache  []*reqData                  //from != leader 的req
	otherReqLimit  int
}

func newReqCache() *reqCache {
	return &reqCache{
		curTurn:        0,
		leaderReqCache: make(map[common.Address]*reqData),
		otherReqCache:  make([]*reqData, 0),
		otherReqLimit:  otherReqCountMax,
	}
}

func (rc *reqCache) AddReq(req *mc.HD_BlkConsensusReqMsg) error {
	if nil == req {
		return paramErr
	}

	rc.mu.Lock()
	defer rc.mu.Unlock()

	if req.ConsensusTurn < rc.curTurn {
		return errors.Errorf("区块请求消息的轮次高低,消息轮次(%d) < 本地轮次(%d)", req.ConsensusTurn, rc.curTurn)
	}

	if req.Header.Leader == req.From {
		oldReq, exit := rc.leaderReqCache[req.From]
		if exit && oldReq.req.ConsensusTurn >= req.ConsensusTurn {
			return leaderReqExistErr
		}
		rc.leaderReqCache[req.From] = newReqData(req)
		return nil
	}

	//other req
	count := len(rc.otherReqCache)
	if count >= rc.otherReqLimit {
		rc.otherReqCache = append(rc.otherReqCache[1:], newReqData(req))
	} else {
		rc.otherReqCache = append(rc.otherReqCache, newReqData(req))
	}
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

func (rc *reqCache) SetCurTurn(consensusTurn uint32) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	if rc.curTurn >= consensusTurn {
		return
	}

	rc.curTurn = consensusTurn
	//fix leader req cache
	deleteList := make([]common.Address, 0)
	for key, req := range rc.leaderReqCache {
		if req.req.ConsensusTurn < rc.curTurn {
			deleteList = append(deleteList, key)
		}
	}
	for _, delKey := range deleteList {
		delete(rc.leaderReqCache, delKey)
	}

	//fix other req cache
	newCache := make([]*reqData, 0)
	for _, req := range rc.otherReqCache {
		if req.req.ConsensusTurn >= rc.curTurn {
			newCache = append(newCache, req)
		}
	}
	rc.otherReqCache = newCache
}

func (rc *reqCache) GetLeaderReq(leader common.Address, consensusTurn uint32) (*reqData, error) {
	if (leader == common.Address{}) {
		return nil, paramErr
	}

	rc.mu.RLock()
	defer rc.mu.RUnlock()
	req, OK := rc.leaderReqCache[leader]
	if !OK {
		return nil, cantFindErr
	}

	if req.req.ConsensusTurn != consensusTurn {
		return nil, errors.Errorf("请求轮次不匹配,缓存(%d) != 目标(%d)", req.req.ConsensusTurn, consensusTurn)
	}

	return req, nil
}

func (rc *reqCache) GetLeaderReqByHash(hash common.Hash) (*reqData, error) {
	if (hash == common.Hash{}) {
		return nil, paramErr
	}

	rc.mu.RLock()
	defer rc.mu.RUnlock()
	for _, req := range rc.leaderReqCache {
		if req.hash == hash {
			return req, nil
		}
	}
	for _, req := range rc.otherReqCache {
		if req.hash == hash {
			return req, nil
		}
	}
	return nil, cantFindErr
}

func (rc *reqCache) GetAllReq() []*reqData {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	result := make([]*reqData, 0, len(rc.leaderReqCache)+cap(rc.otherReqCache))
	for _, req := range rc.leaderReqCache {
		result = append(result, req)
	}
	result = append(result, rc.otherReqCache...)

	sort.Slice(result, func(i, j int) bool {
		return result[i].req.Header.Time.Cmp(result[j].req.Header.Time) > 0
	})

	return result
}
