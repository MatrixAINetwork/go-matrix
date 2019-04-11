// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package blkverify

import (
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/pkg/errors"

	"github.com/MatrixAINetwork/go-matrix/core"
	"github.com/MatrixAINetwork/go-matrix/log"
	"sort"
	"sync"
)

const reqCountMax = 5
const fromCountLimit = 3

var (
	paramErr    = errors.New("param error")
	reqExistErr = errors.New("req already exist")
	cantFindErr = errors.New("can't find req in cache")
)

type reqType uint8

const (
	reqTypeLocalReq   reqType = iota // 本地的请求
	reqTypeLeaderReq                 // leader = from 的请求
	reqTypeOtherReq                  // leader != from 的请求
	reqTypeUnknownReq                // 尚未验证from的req
	reqTypeFromBadReq                // 无法获取A0账户的req
)

type reqData struct {
	reqType           reqType
	req               *mc.HD_BlkConsensusReqMsg
	hash              common.Hash
	originalTxs       []types.CoinSelfTransaction
	finalTxs          []types.CoinSelfTransaction
	receipts          []types.CoinReceipts
	stateDB           *state.StateDBManage
	localVerifyResult verifyResult
	posFinished       bool
	votes             []*common.VerifiedSign
}

func newReqData(req *mc.HD_BlkConsensusReqMsg, isDBRecovery bool, reqType reqType) *reqData {
	data := &reqData{
		reqType:           reqType,
		req:               req,
		hash:              req.Header.HashNoSignsAndNonce(),
		originalTxs:       nil,
		finalTxs:          nil,
		receipts:          nil,
		stateDB:           nil,
		localVerifyResult: localVerifyResultProcessing,
		posFinished:       false,
		votes:             make([]*common.VerifiedSign, 0),
	}
	if isDBRecovery {
		data.localVerifyResult = localVerifyResultDBRecovery
		data.reqType = reqTypeLeaderReq
	}

	return data
}

func newReqDataByLocalReq(localReq *mc.LocalBlockVerifyConsensusReq) *reqData {
	return &reqData{
		reqType:           reqTypeLocalReq,
		req:               localReq.BlkVerifyConsensusReq,
		hash:              localReq.BlkVerifyConsensusReq.Header.HashNoSignsAndNonce(),
		originalTxs:       localReq.OriginalTxs,
		finalTxs:          localReq.FinalTxs,
		receipts:          localReq.Receipts,
		stateDB:           localReq.State,
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
	mu            sync.RWMutex
	curTurn       mc.ConsensusTurnInfo
	reqCache      []*reqData
	reqCountLimit int
	fromLimit     int
	blkChain      *core.BlockChain
}

func newReqCache(chain *core.BlockChain) *reqCache {
	return &reqCache{
		curTurn:       mc.ConsensusTurnInfo{PreConsensusTurn: 0, UsedReelectTurn: 0},
		reqCache:      make([]*reqData, 0),
		reqCountLimit: reqCountMax,
		fromLimit:     fromCountLimit,
		blkChain:      chain,
	}
}

func (rc *reqCache) AddReq(req *mc.HD_BlkConsensusReqMsg, isDBRecovery bool) (*reqData, error) {
	if nil == req || nil == req.Header {
		return nil, paramErr
	}

	if req.From == (common.Address{}) {
		log.Error("blk consensus req cache", "req from err", "is empty address")
		return nil, paramErr
	}

	rc.mu.Lock()
	defer rc.mu.Unlock()

	if req.ConsensusTurn.Cmp(rc.curTurn) < 0 {
		return nil, errors.Errorf("区块请求消息的轮次高低,消息轮次(%s) < 本地轮次(%s)", req.ConsensusTurn.String(), rc.curTurn.String())
	}

	count := len(rc.reqCache)
	fromSize := 0
	for i := 0; i < count; i++ {
		if rc.reqCache[i].req.From == req.From &&
			rc.reqCache[i].req.Header.Leader == req.Header.Leader &&
			rc.reqCache[i].req.ConsensusTurn == req.ConsensusTurn {
			return nil, reqExistErr
		}
		if rc.reqCache[i].req.From == req.From {
			fromSize++
		}
	}
	if fromSize >= rc.fromLimit {
		return nil, errors.Errorf("req from[%s] is too many(%d)", req.From.Hex(), fromSize)
	}

	reqType := reqTypeUnknownReq
	if preBlk := rc.blkChain.GetBlockByHash(req.Header.ParentHash); preBlk != nil {
		a0Account, _, err := rc.blkChain.GetA0AccountFromAnyAccount(req.From, req.Header.ParentHash)
		if err != nil {
			return nil, errors.Errorf("req from[%s] find a0 account err: %v", req.From.Hex(), err)
		}
		if a0Account == req.Header.Leader {
			reqType = reqTypeLeaderReq
		} else {
			reqType = reqTypeOtherReq
		}
	}

	reqData := newReqData(req, isDBRecovery, reqType)
	if count >= rc.reqCountLimit {
		rc.reqCache = append(rc.reqCache[:rc.reqCountLimit-1], reqData)
	} else {
		rc.reqCache = append(rc.reqCache, reqData)
	}
	delBadReqAndSort(rc.reqCache, false)
	return reqData, nil
}

func (rc *reqCache) AddLocalReq(req *mc.LocalBlockVerifyConsensusReq) (*reqData, error) {
	if nil == req {
		return nil, paramErr
	}

	rc.mu.Lock()
	defer rc.mu.Unlock()
	reqData := newReqDataByLocalReq(req)
	rc.reqCache = append(rc.reqCache, reqData)
	delBadReqAndSort(rc.reqCache, false)
	return reqData, nil
}

func (rc *reqCache) SetCurTurn(consensusTurn mc.ConsensusTurnInfo) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	if rc.curTurn.Cmp(consensusTurn) >= 0 {
		return
	}

	rc.curTurn = consensusTurn
	//fix req cache
	newCache := make([]*reqData, 0)
	count := len(rc.reqCache)
	for i := 0; i < count; i++ {
		if rc.reqCache[i].req.ConsensusTurn.Cmp(rc.curTurn) >= 0 {
			newCache = append(newCache, rc.reqCache[i])
		}
	}
	rc.reqCache = newCache
}

func (rc *reqCache) CheckUnknownReq() {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	count := len(rc.reqCache)
	update := false
	for i := 0; i < count; i++ {
		if rc.reqCache[i].reqType != reqTypeUnknownReq {
			continue
		}

		req := rc.reqCache[i].req
		block := rc.blkChain.GetBlockByHash(req.Header.ParentHash)
		if block == nil {
			// 还没有父区块，无法验证
			continue
		}

		a0Account, _, err := rc.blkChain.GetA0AccountFromAnyAccount(req.From, req.Header.ParentHash)
		if err != nil {
			log.Debug("blk consensus req cache", "获取from的抵押账户失败", err, "from", req.From.Hex())
			rc.reqCache[i].reqType = reqTypeFromBadReq
		} else {
			if a0Account == req.Header.Leader {
				rc.reqCache[i].reqType = reqTypeLeaderReq
			} else {
				rc.reqCache[i].reqType = reqTypeOtherReq
			}
		}
		update = true
	}
	if update {
		delBadReqAndSort(rc.reqCache, true)
	}
}

func (rc *reqCache) GetLeaderReq(leader common.Address, consensusTurn mc.ConsensusTurnInfo) (*reqData, error) {
	if (leader == common.Address{}) {
		return nil, paramErr
	}

	rc.mu.RLock()
	defer rc.mu.RUnlock()
	count := len(rc.reqCache)
	for i := 0; i < count; i++ {
		req := rc.reqCache[i]
		if req.reqType != reqTypeLeaderReq && req.reqType != reqTypeLocalReq {
			// 请求不是leader的请求,忽略
			continue
		}
		if req.req.Header.Leader == leader &&
			req.req.ConsensusTurn == consensusTurn {
			return req, nil
		}
	}
	return nil, cantFindErr
}

func (rc *reqCache) GetLeaderReqByHash(hash common.Hash) (*reqData, error) {
	if (hash == common.Hash{}) {
		return nil, paramErr
	}

	rc.mu.RLock()
	defer rc.mu.RUnlock()
	count := len(rc.reqCache)
	for i := 0; i < count; i++ {
		if rc.reqCache[i].hash == hash {
			return rc.reqCache[i], nil
		}
	}
	return nil, cantFindErr
}

func (rc *reqCache) GetAllReq() []*reqData {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	return rc.reqCache[:]
}

func delBadReqAndSort(cache []*reqData, del bool) []*reqData {
	if len(cache) == 0 {
		return make([]*reqData, 0)
	}
	sort.Slice(cache, func(i, j int) bool {
		if cache[i].reqType == cache[j].reqType {
			return cache[i].req.Header.Time.Cmp(cache[j].req.Header.Time) > 0
		} else {
			return cache[i].reqType < cache[j].reqType
		}
	})

	if del == false {
		return cache
	}

	if cache[0].reqType == reqTypeFromBadReq {
		return make([]*reqData, 0)
	}

	count := len(cache)
	pos := count - 1
	for ; pos > 0; pos-- {
		if cache[pos].reqType != reqTypeFromBadReq {
			break
		}
	}
	return cache[:pos+1]
}
