// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package blkgenorV2

import (
	"sort"
	"sync"
	"time"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/params/manparams"
	"github.com/pkg/errors"
)

type aiResultInfo struct {
	aiMsg     *mc.HD_V2_AIMiningRspMsg
	localTime int64
	verified  bool
	legal     bool
}

///////////////////////////////////////////////////////////////////////////////////////////
// 协程安全AI挖矿结果池
type AIResultPool struct {
	// 缓存结构为：map <parentHash, map <from address, *data> >
	aiMap      map[common.Hash]map[common.Address]*aiResultInfo
	countMap   map[common.Address]int
	countLimit int
	logInfo    string
	mu         sync.RWMutex
}

func NewAIResultPool(logInfo string) *AIResultPool {
	return &AIResultPool{
		aiMap:      make(map[common.Hash]map[common.Address]*aiResultInfo),
		countMap:   make(map[common.Address]int),
		countLimit: manparams.AIResultCountLimit,
		logInfo:    logInfo,
	}
}

func (self *AIResultPool) AddAIResult(aiResult *mc.HD_V2_AIMiningRspMsg) error {
	if nil == aiResult {
		return errors.Errorf("AI挖矿结果是空")
	}
	if common.EmptyHash(aiResult.BlockHash) {
		return errors.Errorf("父区块hash是空")
	}

	self.mu.Lock()
	defer self.mu.Unlock()

	if count := self.getFromCount(aiResult.From); count >= self.countLimit {
		return errors.Errorf("该账户发送AI挖矿结果超过存储最大的数目")
	}

	fromMap, OK := self.aiMap[aiResult.BlockHash]
	if !OK {
		fromMap = make(map[common.Address]*aiResultInfo)
		self.aiMap[aiResult.BlockHash] = fromMap
	}

	_, exist := fromMap[aiResult.From]
	if exist {
		//log.Warn(self.logInfo, "AddAIResult", "AI结果已存在", "from", aiResult.From.Hex(), "parent hash", aiResult.BlockHash.TerminalString())
		return errors.Errorf("矿工AI挖矿结果已经存在")
	}
	fromMap[aiResult.From] = &aiResultInfo{aiMsg: aiResult, verified: false, legal: false, localTime: time.Now().UnixNano()}

	self.plusFromCount(aiResult.From)
	return nil
}

func (self *AIResultPool) GetAIResults(parentHash common.Hash) ([]*aiResultInfo, error) {
	if common.EmptyHash(parentHash) {
		return nil, errors.Errorf("父区块Hash是空")
	}

	self.mu.RLock()
	defer self.mu.RUnlock()

	fromMap, OK := self.aiMap[parentHash]
	if !OK || len(fromMap) == 0 {
		return nil, nil
	}

	list := make([]*aiResultInfo, 0)
	for _, result := range fromMap {
		list = append(list, result)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].localTime < list[j].localTime
	})
	return list, nil
}

func (self *AIResultPool) getFromCount(address common.Address) int {
	if count, OK := self.countMap[address]; OK {
		return count
	}
	return 0
}

func (self *AIResultPool) plusFromCount(address common.Address) {
	count, OK := self.countMap[address]
	if !OK {
		self.countMap[address] = 1
	} else {
		self.countMap[address] = count + 1
	}
}

func (self *AIResultPool) minusFromCount(address common.Address) {
	count, OK := self.countMap[address]
	if OK {
		if count > 0 {
			self.countMap[address] = count - 1
		} else {
			self.countMap[address] = 0
		}
	}
}
