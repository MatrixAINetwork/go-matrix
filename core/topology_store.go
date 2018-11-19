// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php

package core

import (
	"github.com/hashicorp/golang-lru"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/rawdb"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mandb"
	"github.com/matrix/go-matrix/mc"
	"github.com/pkg/errors"
	"github.com/matrix/go-matrix/params/manparams"
)

const (
	graphCacheLimit      = 512
	electIndexCacheLimit = 512
)

var (
	errHeaderNotExit        = errors.New("header not exit in chain")
	errGraphNotExitInDb     = errors.New("topology graph not exit in db")
	errElectIndexCantCreate = errors.New("elect index data  can not create, chain data err")
	errReflectElectIndex    = errors.New("reflect elect index data failed")
	errElectIndexIsNil      = errors.New("elect index data is nil")
)

type chainReader interface {
	GetHeaderByNumber(number uint64) *types.Header
	GetHeaderByHash(hash common.Hash) *types.Header
	CurrentHeader() *types.Header
	GetAncestorHash(sonHash common.Hash, ancestorNumber uint64) (common.Hash, error)
}

type TopologyStore struct {
	chainDb         mandb.Database
	reader          chainReader
	graphCache      *lru.Cache
	electIndexCache *lru.Cache
}

func NewTopologyStore(reader chainReader, chainDb mandb.Database) *TopologyStore {
	graphCache, _ := lru.New(graphCacheLimit)
	electIndexCache, _ := lru.New(electIndexCacheLimit)

	return &TopologyStore{
		chainDb:         chainDb,
		reader:          reader,
		graphCache:      graphCache,
		electIndexCache: electIndexCache,
	}
}

func (ts *TopologyStore) GetHashByNumber(number uint64) common.Hash {
	header := ts.reader.GetHeaderByNumber(number)
	if header == nil {
		return common.Hash{}
	}
	return header.Hash()
}

func (ts *TopologyStore) HasTopologyGraph(blockHash common.Hash, number uint64) bool {
	if ts.graphCache.Contains(blockHash) {
		return true
	}
	return rawdb.HasTopologyGraph(ts.chainDb, blockHash, number)
}

func (ts *TopologyStore) GetTopologyGraphByHash(blockHash common.Hash) (*mc.TopologyGraph, error) {
	header := ts.reader.GetHeaderByHash(blockHash)
	if header == nil {
		return nil, errHeaderNotExit
	}
	if graph, ok := ts.graphCache.Get(blockHash); ok {
		return graph.(*mc.TopologyGraph), nil
	}
	graph := rawdb.ReadTopologyGraph(ts.chainDb, blockHash, header.Number.Uint64())
	if graph == nil {
		return nil, errGraphNotExitInDb
	}
	ts.graphCache.Add(blockHash, graph)
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

	preGraph, err := ts.GetTopologyGraphByHash(header.ParentHash)
	if err != nil {
		return nil, errors.Errorf("获取父拓扑图失败:%v", err)
	}

	electList, err := ts.getOriginalElectByHeader(header)
	if err != nil {
		return nil, errors.Errorf("获取选举信息失败:%v", err)
	}
	newGraph, err := preGraph.Transfer2NextGraph(number, &header.NetTopology, electList)
	if err != nil {
		return nil, err
	}
	return newGraph, nil
}

func (ts *TopologyStore) GetOriginalElectByHash(blockHash common.Hash) ([]common.Elect, error) {
	header := ts.reader.GetHeaderByHash(blockHash)
	if header == nil {
		return nil, errHeaderNotExit
	}
	electIndex, err := ts.getElectIndex(header)
	if err != nil {
		return nil, err
	}
	return ts.transferElectIndex2Elect(electIndex)
}

func (ts *TopologyStore) getOriginalElectByHeader(header *types.Header) ([]common.Elect, error) {
	electIndex, err := ts.getElectIndex(header)
	if err != nil {
		return nil, err
	}
	return ts.transferElectIndex2Elect(electIndex)
}

func (ts *TopologyStore) GetNextElectByHash(blockHash common.Hash) ([]common.Elect, error) {
	header := ts.reader.GetHeaderByHash(blockHash)
	if header == nil {
		return nil, errHeaderNotExit
	}
	number := header.Number.Uint64()
	NextElectNumber := common.GetLastReElectionNumber(number+1) + common.GetReElectionInterval()
	minerElectNumber := NextElectNumber - manparams.MinerNetChangeUpTime
	validatorElectNumber := NextElectNumber - manparams.VerifyNetChangeUpTime
	curNumber := ts.reader.CurrentHeader().Number.Uint64()

	nextElect := make([]common.Elect, 0)
	if curNumber >= validatorElectNumber {
		vHeaderHash := blockHash
		if curNumber > validatorElectNumber {
			var err error
			vHeaderHash, err = ts.reader.GetAncestorHash(blockHash, validatorElectNumber)
			if err != nil {
				return nil, errors.Errorf("get validator header hash err(%v). validator number(%d), input number(%d), current number(%d)", err, validatorElectNumber, number, curNumber)
			}
		}
		vHeader := ts.reader.GetHeaderByHash(vHeaderHash)
		if vHeader == nil {
			return nil, errors.Errorf("get validator header err(header is nil). validator number(%d), input number(%d), current number(%d)", validatorElectNumber, number, curNumber)
		}
		nextElect = append(nextElect, vHeader.Elect...)
	}

	if curNumber >= minerElectNumber {
		mHeaderHash := blockHash
		if curNumber > minerElectNumber {
			var err error
			mHeaderHash, err = ts.reader.GetAncestorHash(blockHash, minerElectNumber)
			if err != nil {
				return nil, errors.Errorf("get miner header hash err(%v). validator number(%d), input number(%d), current number(%d)", err, minerElectNumber, number, curNumber)
			}
		}
		mHeader := ts.reader.GetHeaderByHash(mHeaderHash)
		if mHeader == nil {
			return nil, errors.Errorf("get miner header err(header is nil). miner number(%d), input number(%d), current number(%d)", minerElectNumber, number, curNumber)
		}
		nextElect = append(nextElect, mHeader.Elect...)
	}

	return nextElect, nil
}

func (ts *TopologyStore) HasElectIndex(blockHash common.Hash, number uint64) bool {
	if ts.electIndexCache.Contains(blockHash) {
		return true
	}
	return rawdb.HasElectIndex(ts.chainDb, blockHash, number)
}

func (ts *TopologyStore) WriteElectIndex(header *types.Header) error {
	number := header.Number.Uint64()
	hash := header.Hash()
	log.Info("选举索引缓存", "开始缓存，高度", number, "hash", hash.Hex())
	defer log.Info("选举索引缓存", "结束缓存，高度", number, "hash", hash.Hex())
	if ts.HasTopologyGraph(hash, number) {
		return nil
	}

	newGraph, err := ts.NewTopologyGraph(header)
	if err != nil {
		return err
	}

	rawdb.WriteTopologyGraph(ts.chainDb, hash, number, newGraph)
	ts.graphCache.Add(hash, newGraph)
	log.Info("选举索引缓存", "完成缓存，高度", number, "hash", hash.Hex())
	return nil
}

func (ts *TopologyStore) getElectIndex(header *types.Header) (*rawdb.ElectIndexData, error) {
	hash := header.Hash()
	if index, ok := ts.electIndexCache.Get(hash); ok {
		indexData, reflectOK := index.(*rawdb.ElectIndexData)
		if !reflectOK {
			return nil, errReflectElectIndex
		}
		return indexData, nil
	}
	number := header.Number.Uint64()
	index := rawdb.ReadElectIndex(ts.chainDb, hash, number)
	if index == nil {
		if index = ts.newElectIndex(header); index == nil {
			return nil, errElectIndexCantCreate
		}
		rawdb.WriteElectIndex(ts.chainDb, hash, number, index)
	}
	ts.electIndexCache.Add(hash, index)
	return index, nil
}

func (ts *TopologyStore) newElectIndex(header *types.Header) *rawdb.ElectIndexData {
	number := header.Number.Uint64()
	if number < common.GetReElectionInterval() {
		genesisBlockHash := ts.reader.GetHeaderByNumber(0).Hash()
		return &rawdb.ElectIndexData{
			VElectBlock: genesisBlockHash,
			MElectBlock: genesisBlockHash,
		}
	}
	if common.IsReElectionNumber(number + 1) {
		sonHash := header.Hash()
		minerElectNumber := number + 1 - manparams.MinerNetChangeUpTime
		validatorElectNumber := number + 1 - manparams.VerifyNetChangeUpTime

		MElectHash, err := ts.reader.GetAncestorHash(sonHash, minerElectNumber)
		if err != nil {
			log.ERROR("创建选举索引", "寻找矿工选举区块hash错误", err, "number", minerElectNumber, "curNumber", number, "curHash", sonHash.TerminalString())
			return nil
		}
		VElectHash, err := ts.reader.GetAncestorHash(sonHash, validatorElectNumber)
		if err != nil {
			log.ERROR("创建选举索引", "寻找验证者选举区块hash错误", err, "number", validatorElectNumber, "curNumber", number, "curHash", sonHash.TerminalString())
			return nil
		}
		return &rawdb.ElectIndexData{
			VElectBlock: VElectHash,
			MElectBlock: MElectHash,
		}
	} else {
		parentHeader := ts.reader.GetHeaderByHash(header.ParentHash)
		if parentHeader == nil {
			log.ERROR("创建选举索引", "获取父节区块错误", header.ParentHash.TerminalString(), "number", header.Number.Uint64()-1)
			return nil
		}
		electIndex, err := ts.getElectIndex(parentHeader)
		if err != nil {
			log.ERROR("创建选举索引", "获取父节点选举索引异常", err)
			return nil
		}
		return electIndex
	}
}

func (ts *TopologyStore) transferElectIndex2Elect(index *rawdb.ElectIndexData) ([]common.Elect, error) {
	if index == nil {
		return nil, errElectIndexIsNil
	}
	vHeader := ts.reader.GetHeaderByHash(index.VElectBlock)
	if vHeader == nil {
		return nil, errors.Errorf("get validator header err, header = nil. header hash(%s)", index.VElectBlock.Hex())
	}
	mHeader := ts.reader.GetHeaderByHash(index.MElectBlock)
	if mHeader == nil {
		return nil, errors.Errorf("get miner header err, header = nil. header hash(%s)", index.MElectBlock.Hex())
	}
	elect := make([]common.Elect, 0)
	elect = append(elect, vHeader.Elect...)
	elect = append(elect, mHeader.Elect...)
	return elect, nil
}
