// Copyright 2018 The MATRIX Authors as well as Copyright 2014-2017 The go-ethereum Authors
// This file is consisted of the MATRIX library and part of the go-ethereum library.
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
package blockgenor

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/mc"
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
