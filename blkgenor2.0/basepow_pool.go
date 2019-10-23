// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package blkgenorV2

import (
	"math/big"

	"sync"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/params/manparams"
	"github.com/pkg/errors"
)

type basepowInfo struct {
	powMsg   *mc.HD_BasePowerDifficulty
	verified bool
	legal    bool
}

type blockBasePowCache struct {
	// 缓存结构为：map <difficulty, map <from address, *data> >
	resultMap map[common.Hash]map[common.Address]*basepowInfo
	blockHash common.Hash
	logInfo   string
}

func newBlockBasePowCache(blockHash common.Hash, logInfo string) *blockBasePowCache {
	return &blockBasePowCache{
		resultMap: make(map[common.Hash]map[common.Address]*basepowInfo),
		blockHash: blockHash,
		logInfo:   logInfo,
	}
}

func (bpc *blockBasePowCache) addBasePow(diff *big.Int, basPowResult *mc.HD_BasePowerDifficulty) error {
	diffHash := common.BytesToHash(diff.Bytes())
	fromMap, OK := bpc.resultMap[diffHash]
	if !OK {
		fromMap = make(map[common.Address]*basepowInfo)
		bpc.resultMap[diffHash] = fromMap
	}

	_, exist := fromMap[basPowResult.From]
	if exist {
		log.Warn(bpc.logInfo, "添加算力检测结果结果池,已存在的算力检测结果结果from", basPowResult.From.Hex(), "diff", diff, "block hash", bpc.blockHash.TerminalString())
		return errors.Errorf("矿工算力检测结果结果已经存在")
	}
	fromMap[basPowResult.From] = &basepowInfo{basPowResult, false, false}
	return nil
}

func (bpc *blockBasePowCache) delPow(diff *big.Int, from common.Address) bool {
	diffHash := common.BytesToHash(diff.Bytes())
	fromMap, OK := bpc.resultMap[diffHash]
	if !OK {
		return false
	}

	beforeLen := len(fromMap)
	delete(fromMap, from)
	afterLen := len(fromMap)

	return beforeLen != afterLen
}

func (bpc *blockBasePowCache) getPow(diff *big.Int) ([]*basepowInfo, error) {
	diffHash := common.BytesToHash(diff.Bytes())
	fromMap, OK := bpc.resultMap[diffHash]
	if !OK || len(fromMap) == 0 {
		return nil, errors.New("通过难度获取算力检测结果结果失败")
	}

	list := make([]*basepowInfo, 0)
	for _, result := range fromMap {
		list = append(list, result)
	}
	return list, nil
}

///////////////////////////////////////////////////////////////////////////////////////////
// 协程安全投算力检测结果结果池
type BasePowPool struct {
	// 缓存结构为：map <blockHash, *cache>
	powMap     map[common.Hash]*blockBasePowCache
	countMap   map[common.Address]int
	countLimit int
	logInfo    string
	mu         sync.RWMutex
}

func NewBasePowPool(logInfo string) *BasePowPool {
	return &BasePowPool{
		powMap:     make(map[common.Hash]*blockBasePowCache),
		countMap:   make(map[common.Address]int),
		countLimit: manparams.BasePowerCountLimit,
		logInfo:    logInfo,
	}
}

func (self *BasePowPool) AddBasePowResult(blockHash common.Hash, diff *big.Int, basPowResult *mc.HD_BasePowerDifficulty) error {
	self.mu.Lock()
	defer self.mu.Unlock()

	if common.EmptyHash(blockHash) {
		return errors.Errorf("区块hash是空")
	}

	if nil == diff || diff.Cmp(big.NewInt(0)) <= 0 {
		return errors.Errorf("难度不合法")
	}

	if nil == basPowResult {
		return errors.Errorf("矿工算力检测结果结果是空")
	}

	if count := self.getFromCount(basPowResult.From); count >= self.countLimit {
		return errors.Errorf("该账户发送矿工算力检测结果超过存储最大的数目")
	}

	blockCache, OK := self.powMap[blockHash]
	if !OK {
		blockCache = newBlockBasePowCache(blockHash, self.logInfo)
		self.powMap[blockHash] = blockCache
	}

	err := blockCache.addBasePow(diff, basPowResult)
	if err != nil {
		return err
	}
	self.plusFromCount(basPowResult.From)
	log.Info(self.logInfo, "加入算力检测结果结果池成功 账户", basPowResult.From.Hex(), "难度", diff, "区块 hash", blockHash.TerminalString())
	return nil
}

func (self *BasePowPool) DelOneResult(blockHash common.Hash, diff *big.Int, from common.Address) error {
	self.mu.Lock()
	defer self.mu.Unlock()

	if common.EmptyHash(blockHash) {
		return errors.Errorf("区块哈希是空")
	}

	if nil == diff || diff.Cmp(big.NewInt(0)) <= 0 {
		return errors.Errorf("难度不合法")
	}

	if (from == common.Address{}) {
		return errors.Errorf("账户地址是空")
	}
	blockCache, OK := self.powMap[blockHash]
	if !OK {
		return errors.Errorf("没有该数据,删除失败")
	}

	success := blockCache.delPow(diff, from)
	if success {
		count := self.getFromCount(from)
		log.Info(self.logInfo, "删除算力检测结果结果成功, 账户", from.Hex(), "原结果总数", count)
		self.minusFromCount(from)
	}
	return nil
}

func (self *BasePowPool) GetbasPowResults(blockHash common.Hash, diff *big.Int) ([]*basepowInfo, error) {
	self.mu.RLock()
	defer self.mu.RUnlock()

	if common.EmptyHash(blockHash) {
		return nil, errors.Errorf("区块哈希是空")
	}

	if nil == diff || diff.Cmp(big.NewInt(0)) <= 0 {
		return nil, errors.Errorf("难度不合法")
	}

	blockCache, OK := self.powMap[blockHash]
	if !OK {
		return nil, errors.New("没有对应区块hash的数据")
	}

	return blockCache.getPow(diff)
}

func (self *BasePowPool) getFromCount(address common.Address) int {
	if count, OK := self.countMap[address]; OK {
		return count
	}
	return 0
}

func (self *BasePowPool) plusFromCount(address common.Address) {
	count, OK := self.countMap[address]
	if !OK {
		self.countMap[address] = 1
	} else {
		self.countMap[address] = count + 1
	}
}

func (self *BasePowPool) minusFromCount(address common.Address) {
	count, OK := self.countMap[address]
	if OK {
		if count > 0 {
			self.countMap[address] = count - 1
		} else {
			self.countMap[address] = 0
		}
	}
}
