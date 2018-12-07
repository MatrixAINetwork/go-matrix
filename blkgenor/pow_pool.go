// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package blkgenor

import (
	"math/big"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/pkg/errors"

	"sync"
	"github.com/matrix/go-matrix/params/manparams"
)

type blockPowCache struct {
	// 缓存结构为：map <difficulty, map <from address, *data> >
	resultMap map[common.Hash]map[common.Address]*mc.HD_MiningRspMsg
	blockHash common.Hash
	powPool   *PowPool
}

func newBlockPowCache(blockHash common.Hash, powPool *PowPool) *blockPowCache {
	return &blockPowCache{
		resultMap: make(map[common.Hash]map[common.Address]*mc.HD_MiningRspMsg),
		blockHash: blockHash,
		powPool:   powPool,
	}
}

func (bpc *blockPowCache) addPow(diff *big.Int, minerResult *mc.HD_MiningRspMsg) error {
	diffHash := common.BytesToHash(diff.Bytes())
	fromMap, OK := bpc.resultMap[diffHash]
	if !OK {
		fromMap = make(map[common.Address]*mc.HD_MiningRspMsg)
		bpc.resultMap[diffHash] = fromMap
	}

	_, exist := fromMap[minerResult.From]
	if exist {
		log.ERROR(bpc.powPool.logInfo, "添加挖矿结果池,已存在的挖矿结果from", minerResult.From.Hex(), "diff", diff, "block hash", bpc.blockHash.TerminalString())
		return errors.Errorf("pow is already exist")
	}
	fromMap[minerResult.From] = minerResult
	return nil
}

func (bpc *blockPowCache) delPow(diff *big.Int, from common.Address) bool {
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

func (bpc *blockPowCache) getPow(diff *big.Int) ([]*mc.HD_MiningRspMsg, error) {
	diffHash := common.BytesToHash(diff.Bytes())
	fromMap, OK := bpc.resultMap[diffHash]
	if !OK || len(fromMap) == 0 {
		return nil, errors.New("not result in pool, by diff")
	}

	list := make([]*mc.HD_MiningRspMsg, 0)
	for _, result := range fromMap {
		list = append(list, result)
	}
	return list, nil
}

///////////////////////////////////////////////////////////////////////////////////////////
// 协程安全投挖矿结果池
type PowPool struct {
	// 缓存结构为：map <blockHash, *cache>
	powMap     map[common.Hash]*blockPowCache
	countMap   map[common.Address]int
	countLimit int
	logInfo    string
	mu         sync.RWMutex
}

func NewPowPool(logInfo string) *PowPool {
	return &PowPool{
		powMap:     make(map[common.Hash]*blockPowCache),
		countMap:   make(map[common.Address]int),
		countLimit: manparams.VotePoolCountLimit,
		logInfo:    logInfo,
	}
}

func (self *PowPool) AddMinerResult(blockHash common.Hash, diff *big.Int, minerResult *mc.HD_MiningRspMsg) error {
	self.mu.Lock()
	defer self.mu.Unlock()

	if common.EmptyHash(blockHash) {
		return errors.Errorf("block hash is null")
	}

	if nil == diff || 0 == diff.Uint64() {
		return errors.Errorf("diff is  illegal")
	}

	if nil == minerResult {
		return errors.Errorf("minerResult is  null")
	}

	if count := self.getFromCount(minerResult.From); count >= self.countLimit {
		return errors.Errorf("from account had send too much mining result!")
	}

	blockCache, OK := self.powMap[blockHash]
	if !OK {
		blockCache = newBlockPowCache(blockHash, self)
		self.powMap[blockHash] = blockCache
	}

	err := blockCache.addPow(diff, minerResult)
	if err != nil {
		return err
	}
	self.plusFromCount(minerResult.From)
	log.INFO(self.logInfo, "加入挖矿结果池成功 from", minerResult.From.Hex(), "diff", diff, "block hash", blockHash.TerminalString())
	return nil
}

func (self *PowPool) DelOneResult(blockHash common.Hash, diff *big.Int, from common.Address) error {
	self.mu.Lock()
	defer self.mu.Unlock()

	if common.EmptyHash(blockHash) {
		return errors.Errorf("block hash is null")
	}

	if nil == diff || 0 == diff.Uint64() {
		return errors.Errorf("diff is  illegal")
	}

	if (from == common.Address{}) {
		return errors.Errorf("block hash is 0")
	}
	blockCache, OK := self.powMap[blockHash]
	if !OK {
		return errors.Errorf("dont have data,delete fail")
	}

	success := blockCache.delPow(diff, from)
	if success {
		count := self.getFromCount(from)
		log.INFO(self.logInfo, "删除挖矿结果成功, from", from.Hex(), "原结果总数", count)
		self.minusFromCount(from)
	}
	return nil
}

func (self *PowPool) GetMinerResults(blockHash common.Hash, diff *big.Int) ([]*mc.HD_MiningRspMsg, error) {
	self.mu.Lock()
	defer self.mu.Unlock()

	if common.EmptyHash(blockHash) {
		return nil, errors.Errorf("block hash is null")
	}

	if nil == diff || 0 == diff.Uint64() {
		return nil, errors.Errorf("diff is  illegal")
	}

	blockCache, OK := self.powMap[blockHash]
	if !OK {
		return nil, errors.New("not result in pool, by block hash")
	}

	return blockCache.getPow(diff)
}

func (self *PowPool) getFromCount(address common.Address) int {
	if count, OK := self.countMap[address]; OK {
		return count
	}
	return 0
}

func (self *PowPool) plusFromCount(address common.Address) {
	count, OK := self.countMap[address]
	if !OK {
		self.countMap[address] = 1
	} else {
		self.countMap[address] = count + 1
	}
}

func (self *PowPool) minusFromCount(address common.Address) {
	count, OK := self.countMap[address]
	if OK {
		if count > 0 {
			self.countMap[address] = count - 1
		} else {
			self.countMap[address] = 0
		}
	}
}
