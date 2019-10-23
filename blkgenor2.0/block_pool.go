// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package blkgenorV2

import (
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/mc"
)

type blockState uint8

const (
	blockStatePosFinished blockState = iota
	blockStateComplete
)

type blockInfoData struct {
	block *mc.BlockPOSFinishedV2
	state blockState
}

type BlockPool struct {
	cache map[common.Address]*blockInfoData
}

func NewBlockPool() *BlockPool {
	return &BlockPool{
		cache: make(map[common.Address]*blockInfoData),
	}
}

func (self *BlockPool) SavePosBlock(block *mc.BlockPOSFinishedV2) {
	if _, exist := self.cache[block.Header.Leader]; exist {
		return
	}

	self.cache[block.Header.Leader] = &blockInfoData{block, blockStatePosFinished}
}
func (self *BlockPool) SaveCompleteBlock(block *mc.BlockPOSFinishedV2) *blockInfoData {
	if blockData, exist := self.cache[block.Header.Leader]; exist {
		if blockData.state != blockStateComplete {
			blockData.block = block
			blockData.state = blockStateComplete
		}
		return blockData
	}

	data := &blockInfoData{block, blockStateComplete}
	self.cache[block.Header.Leader] = data
	return data
}

func (self *BlockPool) GetBlockData(leader common.Address) *blockInfoData {
	blockData, OK := self.cache[leader]
	if !OK {
		return nil
	}
	return blockData
}

func (self *BlockPool) GetLastBlockData() (blockData *blockInfoData) {
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

func (self *BlockPool) GetBlockDataByBlockHash(hash common.Hash) *blockInfoData {
	for _, value := range self.cache {
		if hash == value.block.BlockHash {
			return value
		}
	}
	return nil
}
