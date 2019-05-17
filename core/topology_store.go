// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package core

import (
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/matrixstate"
	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/pkg/errors"
)

type TopologyStore struct {
	bc *BlockChain
}

func NewTopologyStore(bc *BlockChain) *TopologyStore {
	return &TopologyStore{
		bc: bc,
	}
}

func (ts TopologyStore) ProduceTopologyStateData(block *types.Block, state *state.StateDBManage, readFn PreStateReadFn) (interface{}, error) {
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

func (ts *TopologyStore) GetHashByNumber(number uint64) common.Hash {
	return ts.bc.GetHashByNumber(number)
}

func (ts *TopologyStore) GetCurrentHash() common.Hash {
	return ts.bc.GetCurrentHash()
}

func (ts *TopologyStore) GetTopologyGraphByHash(blockHash common.Hash) (*mc.TopologyGraph, error) {
	st, err := ts.bc.StateAtBlockHash(blockHash)
	if err != nil {
		return nil, err
	}
	return matrixstate.GetTopologyGraph(st)
}

func (ts *TopologyStore) GetElectGraphByHash(blockHash common.Hash) (*mc.ElectGraph, error) {
	st, err := ts.bc.StateAtBlockHash(blockHash)
	if err != nil {
		return nil, err
	}
	return matrixstate.GetElectGraph(st)
}

func (ts *TopologyStore) GetOriginalElectByHash(blockHash common.Hash) ([]common.Elect, error) {
	elect, err := ts.GetElectGraphByHash(blockHash)
	if err != nil {
		return nil, err
	}
	if elect == nil {
		return nil, errors.New("elect data is illegal")
	}
	return elect.TransferElect2CommonElect(), nil
}

func (ts *TopologyStore) GetNextElectByHash(blockHash common.Hash) ([]common.Elect, error) {
	elect, err := ts.GetElectGraphByHash(blockHash)
	if err != nil {
		return nil, err
	}
	if elect == nil {
		return nil, errors.New("elect data is illegal")
	}
	return elect.TransferNextElect2CommonElect(), nil
}

func (ts *TopologyStore) GetElectOnlineStateByHash(blockHash common.Hash) (*mc.ElectOnlineStatus, error) {
	st, err := ts.bc.StateAtBlockHash(blockHash)
	if err != nil {
		return nil, err
	}
	return matrixstate.GetElectOnlineState(st)
}

func (ts *TopologyStore) GetBroadcastAccounts(blockHash common.Hash) ([]common.Address, error) {
	st, err := ts.bc.StateAtBlockHash(blockHash)
	if err != nil {
		return nil, err
	}
	return matrixstate.GetBroadcastAccounts(st)
}

func (ts *TopologyStore) GetInnerMinersAccount(blockHash common.Hash) ([]common.Address, error) {
	st, err := ts.bc.StateAtBlockHash(blockHash)
	if err != nil {
		return nil, err
	}
	return matrixstate.GetInnerMinerAccounts(st)
}

func (ts *TopologyStore) GetSuperSeq(blockHash common.Hash) (uint64, error) {
	st, err := ts.bc.StateAtBlockHash(blockHash)
	if err != nil {
		return 0, err
	}
	supBlkState, err := matrixstate.GetSuperBlockCfg(st)
	if err != nil {
		return 0, err
	}
	return supBlkState.Seq, nil
}
