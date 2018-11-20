// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php

package core

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/rawdb"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/mandb"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/params"
	"github.com/hashicorp/golang-lru"
	"github.com/pkg/errors"
)

const (
	graphCacheLimit = 512
)

var (
	errHeaderNotExit    = errors.New("header not exit in chain")
	errGraphNotExitInDb = errors.New("topology graph not exit in db")
)

type chainReader interface {
	GetHeaderByNumber(number uint64) *types.Header
	CurrentHeader() *types.Header
}

type TopologyStore struct {
	chainDb    mandb.Database
	reader     chainReader
	graphCache *lru.Cache
}

func NewTopologyStore(reader chainReader, chainDb mandb.Database) *TopologyStore {
	graphCache, _ := lru.New(graphCacheLimit)

	return &TopologyStore{
		chainDb:    chainDb,
		reader:     reader,
		graphCache: graphCache,
	}
}

func (ts *TopologyStore) HasTopologyGraph(blockHash common.Hash, number uint64) bool {
	if ts.graphCache.Contains(blockHash) {
		return true
	}
	return rawdb.HasTopologyGraph(ts.chainDb, blockHash, number)
}

func (ts *TopologyStore) GetTopologyGraphByNumber(number uint64) (*mc.TopologyGraph, error) {
	header := ts.reader.GetHeaderByNumber(number)
	if header == nil {
		return nil, errHeaderNotExit
	}

	hash := header.Hash()
	if graph, ok := ts.graphCache.Get(hash); ok {
		return graph.(*mc.TopologyGraph), nil
	}

	graph := rawdb.ReadTopologyGraph(ts.chainDb, hash, header.Number.Uint64())
	if graph == nil {
		return nil, errGraphNotExitInDb
	}
	ts.graphCache.Add(hash, graph)
	return graph, nil
}

func (ts *TopologyStore) WriteTopologyGraph(header *types.Header) error {
	number := header.Number.Uint64()
	hash := header.Hash()
	log.Info("拓扑缓存", "开始缓存，高度", number, "hash", hash.Hex())
	defer log.Info("拓扑缓存", "结束缓存，高度", number, "hash", hash.Hex())
	if ts.HasTopologyGraph(hash, number) {
		return nil
	}

	newGraph, err := ts.NewTopologyGraph(header)
	if err != nil {
		return err
	}

	rawdb.WriteTopologyGraph(ts.chainDb, hash, number, newGraph)
	ts.graphCache.Add(hash, newGraph)
	log.Info("拓扑缓存", "完成缓存，高度", number, "hash", hash.Hex())
	return nil
}

func (ts *TopologyStore) NewTopologyGraph(header *types.Header) (*mc.TopologyGraph, error) {
	number := header.Number.Uint64()
	if number == 0 {
		return mc.NewGenesisTopologyGraph(header)
	}

	preGraph, err := ts.GetTopologyGraphByNumber(number - 1)
	if err != nil {
		return nil, errors.Errorf("获取父拓扑图失败:%v", err)
	}
	electList, err := ts.GetOriginalElect(number)
	if err != nil {
		return nil, errors.Errorf("获取选举信息失败:%v", err)
	}
	newGraph, err := preGraph.Transfer2NextGraph(number, &header.NetTopology, electList)
	if err != nil {
		return nil, err
	}
	return newGraph, nil
}

//todo 按高度拿选举使用的是主链上的块，以后加入硬分叉后，可能导致同一高度区块不一致，需要改进为根据hash拿块的elect
func (ts *TopologyStore) GetOriginalElect(number uint64) ([]common.Elect, error) {
	lastElectNumber := common.GetLastReElectionNumber(number + 1)
	if lastElectNumber == 0 {
		header := ts.reader.GetHeaderByNumber(0)
		return header.Elect, nil
	}

	minerElectNumber := lastElectNumber - params.MinerNetChangeUpTime
	validatorElectNumber := lastElectNumber - params.VerifyNetChangeUpTime
	if minerElectNumber != validatorElectNumber {
		vHeader := ts.reader.GetHeaderByNumber(validatorElectNumber)
		if vHeader == nil {
			return nil, errors.Errorf("get validator header err, header = nil. validator number = %d, input number = %d", validatorElectNumber, number)
		}
		mHeader := ts.reader.GetHeaderByNumber(minerElectNumber)
		if mHeader == nil {
			return nil, errors.Errorf("get miner header err, header = nil. miner number = %d, input number = %d", minerElectNumber, number)
		}

		return append(vHeader.Elect, mHeader.Elect...), nil
	} else {
		header := ts.reader.GetHeaderByNumber(minerElectNumber)
		if header == nil {
			return nil, errors.Errorf("get elect header err, header = nil. elect number = %d, input number = %d", minerElectNumber, number)
		}
		return header.Elect, nil
	}
}

func (ts *TopologyStore) GetNextElect(number uint64) ([]common.Elect, error) {
	NextElectNumber := common.GetLastReElectionNumber(number+1) + common.GetReElectionInterval()
	minerElectNumber := NextElectNumber - params.MinerNetChangeUpTime
	validatorElectNumber := NextElectNumber - params.VerifyNetChangeUpTime
	curNumber := ts.reader.CurrentHeader().Number.Uint64()

	nextElect := make([]common.Elect, 0)
	if curNumber >= validatorElectNumber {
		vHeader := ts.reader.GetHeaderByNumber(validatorElectNumber)
		if vHeader == nil {
			return nil, errors.Errorf("get validator header err, header = nil. validator number(%d), input number(%d), current number(%d)", validatorElectNumber, number, curNumber)
		}
		nextElect = append(nextElect, vHeader.Elect...)
	}

	if curNumber >= minerElectNumber {
		mHeader := ts.reader.GetHeaderByNumber(minerElectNumber)
		if mHeader == nil {
			return nil, errors.Errorf("get miner header err, header = nil. miner number(%d), input number(%d), current number(%d)", minerElectNumber, number, curNumber)
		}
		nextElect = append(nextElect, mHeader.Elect...)
	}

	return nextElect, nil
}
