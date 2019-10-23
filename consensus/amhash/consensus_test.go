// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package amhash

import (
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/params"
	"math/big"
	"testing"
)

type testChain struct {
	headers map[common.Hash]*types.Header
}

func (self *testChain) Config() *params.ChainConfig {
	return nil
}

func (self *testChain) CurrentHeader() *types.Header {
	return nil
}

func (self *testChain) GetHeader(hash common.Hash, number uint64) *types.Header {
	return nil
}
func (self *testChain) GetHeaderByNumber(number uint64) *types.Header {
	for _, header := range self.headers {
		if header.Number.Uint64() == number {
			return header
		}
	}
	return nil
}
func (self *testChain) GetHeaderByHash(hash common.Hash) *types.Header {
	header, exist := self.headers[hash]
	if exist {
		return header
	}
	return nil
}
func (self *testChain) GetBlock(hash common.Hash, number uint64) *types.Block {
	return nil
}
func (self *testChain) GetGraphByHash(hash common.Hash) (*mc.TopologyGraph, *mc.ElectGraph, error) {
	return nil, nil, nil
}
func (self *testChain) GetInnerMinerAccounts(blockHash common.Hash) ([]common.Address, error) {
	return nil, nil
}
func (self *testChain) GetMinDifficulty(blockHash common.Hash) (*big.Int, error) {
	return nil, nil
}
func (self *testChain) GetMaxDifficulty(blockHash common.Hash) (*big.Int, error) {
	return nil, nil
}
func (self *testChain) GetReelectionDifficulty(blockHash common.Hash) (*big.Int, error) {
	return nil, nil
}
func (self *testChain) GetBroadcastIntervalByHash(blockHash common.Hash) (*mc.BCIntervalInfo, error) {
	return nil, nil
}
func (self *testChain) GetAncestorHash(sonHash common.Hash, ancestorNumber uint64) (common.Hash, error) {
	return common.Hash{}, nil
}
func (self *testChain) GetBlockDurationStatus(blockHash common.Hash) (*mc.BlockDurationStatus, error) {
	return nil, nil
}

func Test_GetAvgHeaders(t *testing.T) {
	//log.InitLog(5)
	chain := &testChain{
		headers: make(map[common.Hash]*types.Header),
	}
	genensHeader := &types.Header{
		Number:     big.NewInt(int64(0)),
		ParentHash: common.HexToHash("0xffffffffffff"),
		Time:       big.NewInt(100),
		Difficulty: big.NewInt(0),
	}
	chain.headers[genensHeader.Hash()] = genensHeader
	for i := uint64(1); i < 400; i++ {
		header := &types.Header{
			Number:     big.NewInt(int64(i)),
			ParentHash: chain.GetHeaderByNumber(i - 1).Hash(),
			Time:       big.NewInt(int64(i)), //big.NewInt(0).Add(chain.GetHeaderByNumber(i-1).Time, big.NewInt(int64(i))),
			Difficulty: big.NewInt(int64(i)),
		}
		chain.headers[header.Hash()] = header
	}

	bcInterval := &mc.BCIntervalInfo{
		LastReelectNumber: 0,
		BCInterval:        100,
	}
	amhash := &Amhash{}

	curNumber := uint64(297)
	if curNumber >= uint64(300) {
		bcInterval.LastReelectNumber = 300
	}
	curheader := chain.GetHeaderByNumber(curNumber)
	parentHeader := chain.GetHeaderByNumber(curNumber - 1)

	results, err := amhash.getDifficultyInfors(parentHeader.Hash(), curNumber, curheader.Time.Uint64(), bcInterval, chain)
	if err != nil {
		t.Fatalf("err := %v", err)
	}

	t.Logf("result size := %d", len(results))
	for i, item := range results {
		t.Logf("index := %d difficulty := %d, time := %d", i, item.difficulty.Uint64(), item.Duration)
	}

}
