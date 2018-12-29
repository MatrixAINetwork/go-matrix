// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package matrixstate

import (
	"github.com/hashicorp/golang-lru"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/mc"
	"github.com/pkg/errors"
)

const (
	topologyCacheLimit = 512
	electCacheLimit    = 512
)

type GraphStore struct {
	stateKey      string
	reader        stateReader
	topologyCache *lru.Cache
	electCache    *lru.Cache
}

func NewGraphStore(reader stateReader) *GraphStore {
	topologyCache, _ := lru.New(topologyCacheLimit)
	electCache, _ := lru.New(electCacheLimit)

	return &GraphStore{
		reader:        reader,
		topologyCache: topologyCache,
		electCache:    electCache,
	}
}

func (gs GraphStore) ProduceTopologyStateData(block *types.Block, readFn PreStateReadFn) (interface{}, error) {
	header := block.Header()
	number := header.Number.Uint64()

	preData, err := readFn(mc.MSKeyTopologyGraph)
	if err != nil {
		return nil, errors.Errorf("read pre data err(%v)", err)
	}
	preGraph, OK := preData.(*mc.TopologyGraph)
	if OK == false || preGraph == nil {
		return nil, errors.Errorf("Invalid preGraph(number = %d)", number-1)
	}
	newGraph, err := preGraph.Transfer2NextGraph(number, &header.NetTopology)
	if err != nil {
		return nil, err
	}
	return newGraph, nil
}

func (gs GraphStore) GetHashByNumber(number uint64) common.Hash {
	return gs.reader.GetHashByNumber(number)
}

func (gs GraphStore) GetCurrentHash() common.Hash {
	return gs.reader.GetCurrentHash()
}

func (gs GraphStore) GetTopologyGraphByHash(blockHash common.Hash) (*mc.TopologyGraph, error) {
	if graph, ok := gs.topologyCache.Get(blockHash); ok {
		return graph.(*mc.TopologyGraph), nil
	}

	graphData, err := gs.reader.GetMatrixStateDataByHash(mc.MSKeyTopologyGraph, blockHash)
	if err != nil {
		return nil, err
	}
	graph, OK := graphData.(*mc.TopologyGraph)
	if OK == false || graph == nil {
		return nil, errors.New("topology data reflect err, msg is nil")
	}
	gs.topologyCache.Add(blockHash, graph)
	return graph, nil
}

func (gs GraphStore) GetElectGraphByHash(blockHash common.Hash) (*mc.ElectGraph, error) {
	if elect, ok := gs.electCache.Get(blockHash); ok {
		return elect.(*mc.ElectGraph), nil
	}

	electData, err := gs.reader.GetMatrixStateDataByHash(mc.MSKeyElectGraph, blockHash)
	if err != nil {
		return nil, err
	}
	elect, OK := electData.(*mc.ElectGraph)
	if OK == false || elect == nil {
		return nil, errors.New("elect data reflect err, msg is nil")
	}
	gs.electCache.Add(blockHash, elect)
	return elect, nil
}

func (gs *GraphStore) GetOriginalElectByHash(blockHash common.Hash) ([]common.Elect, error) {
	elect, err := gs.GetElectGraphByHash(blockHash)
	if err != nil {
		return nil, err
	}

	return elect.TransferElect2CommonElect(), nil
}

func (gs *GraphStore) GetNextElectByHash(blockHash common.Hash) ([]common.Elect, error) {
	elect, err := gs.GetElectGraphByHash(blockHash)
	if err != nil {
		return nil, err
	}

	if elect == nil {
		return nil, errors.New("elect data is illegal")
	}

	return elect.TransferNextElect2CommonElect(), nil
}

func (gs *GraphStore) GetBroadcastAccount(blockHash common.Hash) (common.Address, error) {
	data, err := gs.reader.GetMatrixStateDataByHash(mc.MSKeyAccountBroadcast, blockHash)
	if err != nil {
		return common.Address{}, err
	}
	account, OK := data.(common.Address)
	if OK == false || account == (common.Address{}) {
		return common.Address{}, errors.New("broadcast account reflect err")
	}

	return account, nil
}

func (gs *GraphStore) GetInnerMinersAccount(blockHash common.Hash) ([]common.Address, error) {
	data, err := gs.reader.GetMatrixStateDataByHash(mc.MSKeyAccountInnerMiners, blockHash)
	if err != nil {
		return nil, err
	}
	accounts, OK := data.([]common.Address)
	if OK == false || accounts == nil {
		return nil, errors.New("inner miner accounts reflect err")
	}

	return accounts, nil
}
