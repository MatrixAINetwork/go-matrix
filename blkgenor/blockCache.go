// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package blkgenor

import (
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/mc"
)

type blockState uint8

const (
	blockStateLocalVerified blockState = iota
	blockStateReady
)

type blockCacheData struct {
	block *mc.BlockLocalVerifyOK
	state blockState
}

type blockCache struct {
	cache map[common.Address]*blockCacheData
}

func newBlockCache() *blockCache {
	return &blockCache{
		cache: make(map[common.Address]*blockCacheData),
	}
}

func (self *blockCache) SaveVerifiedBlock(block *mc.BlockLocalVerifyOK) {
	if _, exist := self.cache[block.Header.Leader]; exist {
		return
	}

	self.cache[block.Header.Leader] = &blockCacheData{block: block, state: blockStateLocalVerified}
}

func (self *blockCache) SaveReadyBlock(block *mc.BlockLocalVerifyOK) {
	self.cache[block.Header.Leader] = &blockCacheData{block: block, state: blockStateReady}
}

func (self *blockCache) GetBlockData(leader common.Address) *blockCacheData {
	blockData, OK := self.cache[leader]
	if !OK {
		return nil
	}
	return blockData
}

func (self *blockCache) GetLastBlockData() (blockData *blockCacheData) {
	blockData = nil
	lastTime := int64(0)
	for _, value := range self.cache {
		blockTime := value.block.Header.Time.Int64()
		if blockTime > lastTime {
			blockData = value
			lastTime = blockTime
		}
	}
	return blockData
}

func (self *blockCache) GetBlockDataByBlockHash(hash common.Hash) *blockCacheData {
	for _, value := range self.cache {
		if hash == value.block.BlockHash {
			return value
		}
	}
	return nil
}
