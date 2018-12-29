// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package blkverify

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/mc"
	"github.com/pkg/errors"

	"sort"
	"sync"
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
	originalTxs       types.SelfTransactions
	finalTxs          types.SelfTransactions
	receipts          []*types.Receipt
	stateDB           *state.StateDB
	localReq          bool
	localVerifyResult verifyResult
	posFinished       bool
	votes             []*common.VerifiedSign
}

func newReqData(req *mc.HD_BlkConsensusReqMsg, isDBRecovery bool) *reqData {
	data := &reqData{
		req:               req,
		hash:              req.Header.HashNoSignsAndNonce(),
		originalTxs:       nil,
		finalTxs:          nil,
		receipts:          nil,
		stateDB:           nil,
		localReq:          false,
		localVerifyResult: localVerifyResultProcessing,
		posFinished:       false,
		votes:             make([]*common.VerifiedSign, 0),
	}
	if isDBRecovery {
		data.localVerifyResult = localVerifyResultDBRecovery
	}
	return data
}

func newReqDataByLocalReq(localReq *mc.LocalBlockVerifyConsensusReq) *reqData {
	return &reqData{
		req:               localReq.BlkVerifyConsensusReq,
		hash:              localReq.BlkVerifyConsensusReq.Header.HashNoSignsAndNonce(),
		originalTxs:       localReq.OriginalTxs,
		finalTxs:          localReq.FinalTxs,
		receipts:          localReq.Receipts,
		stateDB:           localReq.State,
		localReq:          true,
		localVerifyResult: localVerifyResultProcessing,
		posFinished:       false,
		votes:             make([]*common.VerifiedSign, 0),
	}
}

func (rd *reqData) isAccountExistVote(account common.Address) bool {
	if (account == common.Address{}) {
		return true
	}

	for _, item := range rd.votes {
		if item.Account == account {
			return true
		}
	}
	return false
}

func (rd *reqData) addVote(vote *common.VerifiedSign) error {
	if vote == nil || (vote.Account == common.Address{}) {
		return ErrParamIsNil
	}

	for _, item := range rd.votes {
		if item.Account == vote.Account {
			return ErrExistVote
		}
	}
	rd.votes = append(rd.votes, vote)
	return nil
}

func (rd *reqData) getVotes() []*common.VerifiedSign {
	return rd.votes[:]
}

func (rd *reqData) clearVotes() {
	rd.votes = make([]*common.VerifiedSign, 0)
}

type reqCache struct {
	mu             sync.RWMutex
	curTurn        mc.ConsensusTurnInfo
	leaderReqCache map[common.Address]*reqData //from = leader 的req
	otherReqCache  []*reqData                  //from != leader 的req
	otherReqLimit  int
}

func newReqCache() *reqCache {
	return &reqCache{
		curTurn:        mc.ConsensusTurnInfo{0, 0},
		leaderReqCache: make(map[common.Address]*reqData),
		otherReqCache:  make([]*reqData, 0),
		otherReqLimit:  otherReqCountMax,
	}
}

func (rc *reqCache) AddReq(req *mc.HD_BlkConsensusReqMsg, isDBRecovery bool) (*reqData, error) {
	if nil == req {
		return nil, paramErr
	}

	rc.mu.Lock()
	defer rc.mu.Unlock()

	if req.ConsensusTurn.Cmp(rc.curTurn) < 0 {
		return nil, errors.Errorf("区块请求消息的轮次高低,消息轮次(%s) < 本地轮次(%s)", req.ConsensusTurn.String(), rc.curTurn.String())
	}

	if req.Header.Leader == req.From {
		oldReq, exit := rc.leaderReqCache[req.From]
		if exit && oldReq.req.ConsensusTurn.Cmp(req.ConsensusTurn) >= 0 {
			return nil, leaderReqExistErr
		}
		reqData := newReqData(req, isDBRecovery)
		rc.leaderReqCache[req.From] = reqData
		return reqData, nil
	}

	//other req
	reqData := newReqData(req, isDBRecovery)
	count := len(rc.otherReqCache)
	if count >= rc.otherReqLimit {
		rc.otherReqCache = append(rc.otherReqCache[1:], reqData)
	} else {
		rc.otherReqCache = append(rc.otherReqCache, reqData)
	}
	return reqData, nil
}

func (rc *reqCache) AddLocalReq(req *mc.LocalBlockVerifyConsensusReq) (*reqData, error) {
	if nil == req {
		return nil, paramErr
	}

	rc.mu.Lock()
	defer rc.mu.Unlock()
	reqData := newReqDataByLocalReq(req)
	rc.leaderReqCache[req.BlkVerifyConsensusReq.Header.Leader] = reqData
	return reqData, nil
}

func (rc *reqCache) SetCurTurn(consensusTurn mc.ConsensusTurnInfo) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	if rc.curTurn.Cmp(consensusTurn) >= 0 {
		return
	}

	rc.curTurn = consensusTurn
	//fix leader req cache
	deleteList := make([]common.Address, 0)
	for key, req := range rc.leaderReqCache {
		if req.req.ConsensusTurn.Cmp(rc.curTurn) < 0 {
			deleteList = append(deleteList, key)
		}
	}
	for _, delKey := range deleteList {
		delete(rc.leaderReqCache, delKey)
	}

	//fix other req cache
	newCache := make([]*reqData, 0)
	for _, req := range rc.otherReqCache {
		if req.req.ConsensusTurn.Cmp(rc.curTurn) >= 0 {
			newCache = append(newCache, req)
		}
	}
	rc.otherReqCache = newCache
}

func (rc *reqCache) GetLeaderReq(leader common.Address, consensusTurn mc.ConsensusTurnInfo) (*reqData, error) {
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
		return nil, errors.Errorf("请求轮次不匹配,缓存(%s) != 目标(%s)", req.req.ConsensusTurn.String(), consensusTurn.String())
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
