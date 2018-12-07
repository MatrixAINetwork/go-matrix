// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php

package miner

import (
	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/consensus"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/log"
	"github.com/pkg/errors"
	"math/big"
	"sync"
	"github.com/matrix/go-matrix/params/manparams"
)

type mineReqData struct {
	mu                 sync.Mutex
	mined              bool
	headerHash         common.Hash
	header             *types.Header
	isBroadcastReq     bool
	txs                types.SelfTransactions
	mineDiff           *big.Int
	mineResultSendTime int64
}

func newMineReqData(headerHash common.Hash, header *types.Header, txs types.SelfTransactions, isBroadcastReq bool) *mineReqData {
	return &mineReqData{
		mined:              false,
		headerHash:         headerHash,
		header:             header,
		isBroadcastReq:     isBroadcastReq,
		txs:                txs,
		mineDiff:           nil,
		mineResultSendTime: 0,
	}
}

func (self *mineReqData) ResendMineResult(curTime int64) error {
	if false == self.mined {
		return errors.New("尚未挖矿完成")
	}

	self.mu.Lock()
	defer self.mu.Unlock()
	if curTime-self.mineResultSendTime < manparams.MinerResultSendInterval {
		return errors.Errorf("挖矿发送间隔尚未到, 上次发送时间(%d), 当前时间(%d)", self.mineResultSendTime, curTime)
	}
	self.mineResultSendTime = curTime
	return nil
}

type mineReqCtrl struct {
	curNumber       uint64
	currentMineReq  *mineReqData
	role            common.RoleType
	posEngine       consensus.DPOSEngine
	validatorReader consensus.ValidatorReader
	reqCache        map[common.Hash]*mineReqData
	futureReq       map[uint64][]*mineReqData //todo 考虑作恶，可以加入限长
}

func newMinReqCtrl(posEngine consensus.DPOSEngine, validatorReader consensus.ValidatorReader) *mineReqCtrl {
	return &mineReqCtrl{
		curNumber:       0,
		currentMineReq:  nil,
		role:            common.RoleNil,
		validatorReader: validatorReader,
		posEngine:       posEngine,
		reqCache:        make(map[common.Hash]*mineReqData),
		futureReq:       make(map[uint64][]*mineReqData),
	}
}

func (ctrl *mineReqCtrl) SetNewNumber(number uint64, role common.RoleType) {
	if ctrl.curNumber > number {
		return
	} else if ctrl.curNumber == number {
		ctrl.role = role
		return
	} else {
		ctrl.curNumber = number
		ctrl.role = role
		ctrl.fixMap()
		return
	}
}

func (ctrl *mineReqCtrl) AddMineReq(header *types.Header, txs types.SelfTransactions, isBroadcastReq bool) (*mineReqData, error) {
	if nil == header {
		return nil, errors.New("header为nil")
	}

	reqNumber := header.Number.Uint64()
	headerHash := header.HashNoSignsAndNonce()
	if reqNumber > ctrl.curNumber {
		list, exist := ctrl.futureReq[reqNumber]
		reqData := newMineReqData(headerHash, header, txs, isBroadcastReq)
		if exist {
			ctrl.futureReq[reqNumber] = append(list, reqData)
		} else {
			ctrl.futureReq[reqNumber] = []*mineReqData{reqData}
		}
		return nil, nil
	} else if reqNumber < ctrl.curNumber {
		return nil, errors.Errorf("挖矿请求消息高度(%d) 小于 当前高度(%d)", reqNumber, ctrl.curNumber)
	} else {
		data, exist := ctrl.reqCache[headerHash]
		if exist {
			return data, nil
		}

		if err := ctrl.checkMineReq(header); err != nil {
			return nil, err
		}

		reqData := newMineReqData(headerHash, header, txs, isBroadcastReq)
		ctrl.reqCache[headerHash] = reqData
		return reqData, nil
	}
}

func (ctrl *mineReqCtrl) CanMining() bool {
	return ctrl.roleCanMine(ctrl.role, ctrl.curNumber)
}

func (ctrl *mineReqCtrl) GetMineReqData(headerHash common.Hash) (*mineReqData, error) {
	reqData, exist := ctrl.reqCache[headerHash]
	if !exist {
		return nil, errors.New("请求消息未找到")
	}
	if reqData == nil {
		return nil, errors.New("请求消息找到，但是为nil")
	}
	return reqData, nil
}

func (ctrl *mineReqCtrl) GetUnMinedReq() *mineReqData {
	//todo 获取时间戳最大的

	for hash, req := range ctrl.reqCache {
		if req == nil {
			log.ERROR(ModuleMiner, "GetUnMinedReq", "有reqData为nil", "hash", hash.TerminalString())
			continue
		}
		if req.mined {
			continue
		}
		return req
	}
	return nil
}

func (ctrl *mineReqCtrl) SetCurrentMineReq(headerHash common.Hash) error {
	if ctrl.currentMineReq != nil && ctrl.currentMineReq.headerHash == headerHash {
		return nil
	}
	req, err := ctrl.GetMineReqData(headerHash)
	if err != nil {
		return err
	}
	if req.mined {
		return errors.Errorf("请求(%s)已挖矿完成", headerHash.TerminalString())
	}
	ctrl.currentMineReq = req
	return nil
}

func (ctrl *mineReqCtrl) GetCurrentMineReq() *mineReqData {
	return ctrl.currentMineReq
}

func (ctrl *mineReqCtrl) SetMiningResult(result *types.Header) (*mineReqData, error) {
	if nil == result {
		return nil, errors.New("消息为nil")
	}
	headerHash := result.HashNoSignsAndNonce()
	req, err := ctrl.GetMineReqData(headerHash)
	if err != nil {
		return nil, err
	}

	if req.mined {
		return nil, errors.Errorf("请求(%s)已挖矿完成", headerHash.TerminalString())
	}

	req.mineDiff = result.Difficulty

	if req.isBroadcastReq {
		req.header.Coinbase = ca.GetAddress()
	} else {
		req.header.Nonce = result.Nonce
		req.header.Coinbase = result.Coinbase
		req.header.MixDigest = result.MixDigest
		req.header.Signatures = result.Signatures
	}

	req.mined = true

	if ctrl.currentMineReq != nil && ctrl.currentMineReq.headerHash == headerHash {
		ctrl.currentMineReq = nil
	}
	return req, nil
}

func (ctrl *mineReqCtrl) checkMineReq(header *types.Header) error {
	if header.Difficulty.Uint64() == 0 {
		return difficultyIsZero
	}
	err := ctrl.posEngine.VerifyBlock(ctrl.validatorReader, header)
	if err != nil {
		return errors.Errorf("挖矿请求POS验证失败(%v)", err)
	}
	return nil
}

func (ctrl *mineReqCtrl) roleCanMine(role common.RoleType, number uint64) bool {
	if common.IsBroadcastNumber(number) {
		return role == common.RoleBroadcast
	} else {
		return role == common.RoleMiner || role == common.RoleInnerMiner
	}
}

func (ctrl *mineReqCtrl) fixMap() {
	ctrl.reqCache = make(map[common.Hash]*mineReqData)
	reqList, exist := ctrl.futureReq[ctrl.curNumber]
	if !exist {
		return
	}

	for i := 0; i < len(reqList); i++ {
		reqData := reqList[i]
		_, exist := ctrl.reqCache[reqData.headerHash]
		if exist {
			continue
		}
		err := ctrl.checkMineReq(reqData.header)
		if err != nil {
			log.WARN(ModuleMiner, "fixMap", "检测请求时，验证失败", err, "高度", ctrl.curNumber)
			continue
		}
		ctrl.reqCache[reqData.headerHash] = reqData
	}

	delete(ctrl.futureReq, ctrl.curNumber)
}
