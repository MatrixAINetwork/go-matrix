// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package core

import (
	"math/big"

	"github.com/MatrixAINetwork/go-matrix/baseinterface"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/core/vm"
)

// Validator is an interface which defines the standard for block validation. It
// is only responsible for validating block contents, as the header validation is
// done by the specific consensus engines.
//
type Validator interface {
	// ValidateBody validates the given block's content.
	ValidateBody(block *types.Block) error
	// ValidateBody validates the given block's content.
	ValidateHeader(header *types.Header) error
	// ValidateState validates the given statedb and optionally the receipts and
	// gas used.
	//ValidateState(block, parent *types.Block, state *state.StateDBManage, receipts types.Receipts, usedGas uint64) error
	ValidateState(block, parent *types.Block, statedb *state.StateDBManage, usedGas uint64) error
}

// Processor is an interface for processing blocks using a given initial state.
//
// Process takes the block to be processed and the statedb upon which the
// initial state is based. It should return the receipts generated, amount
// of gas used in the process and return an error if any of the internal rules
// failed.
type Processor interface {
	ProcessSuperBlk(block *types.Block, statedb *state.StateDBManage) error
	ProcessTxs(block *types.Block, statedb *state.StateDBManage, cfg vm.Config, upTime map[common.Address]uint64) ([]types.CoinLogs, uint64, error)
	Process(block *types.Block, parent *types.Block, statedb *state.StateDBManage, cfg vm.Config) ([]types.CoinReceipts, []types.CoinLogs, uint64, error)
	SetRandom(random *baseinterface.Random)
	ProcessReward(state *state.StateDBManage, header *types.Header, upTime map[common.Address]uint64, account map[string][]common.Address, usedGas map[string]*big.Int) []common.RewarTx
}
