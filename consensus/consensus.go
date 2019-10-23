// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

// Package consensus implements different Matrix consensus engines.
package consensus

import (
	"math/big"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/params"
	"github.com/MatrixAINetwork/go-matrix/rpc"
)

// ChainReader defines a small collection of methods needed to access the local
// blockchain during header and/or uncle verification.
type ChainReader interface {
	// Config retrieves the blockchain's chain configuration.
	Config() *params.ChainConfig

	// CurrentHeader retrieves the current header from the local chain.
	CurrentHeader() *types.Header

	// GetHeader retrieves a block header from the database by hash and number.
	GetHeader(hash common.Hash, number uint64) *types.Header

	// GetHeaderByNumber retrieves a block header from the database by number.
	GetHeaderByNumber(number uint64) *types.Header

	// GetHeaderByHash retrieves a block header from the database by its hash.
	GetHeaderByHash(hash common.Hash) *types.Header

	// GetBlock retrieves a block from the database by hash and number.
	GetBlock(hash common.Hash, number uint64) *types.Block

	GetGraphByHash(hash common.Hash) (*mc.TopologyGraph, *mc.ElectGraph, error)

	GetInnerMinerAccounts(blockHash common.Hash) ([]common.Address, error)

	GetMinDifficulty(blockHash common.Hash) (*big.Int, error)
	GetMaxDifficulty(blockHash common.Hash) (*big.Int, error)
	GetReelectionDifficulty(blockHash common.Hash) (*big.Int, error)
	GetBroadcastIntervalByHash(blockHash common.Hash) (*mc.BCIntervalInfo, error)

	GetAncestorHash(sonHash common.Hash, ancestorNumber uint64) (common.Hash, error)
	GetBlockDurationStatus(blockHash common.Hash) (*mc.BlockDurationStatus, error)
}

// Engine is an algorithm agnostic consensus engine.
type Engine interface {
	// Author retrieves the Matrix address of the account that minted the given
	// block, which may be different from the header's coinbase if a consensus
	// engine is based on signatures.
	Author(header *types.Header) (common.Address, error)

	// VerifyHeader checks whether a header conforms to the consensus rules of a
	// given engine. Verifying the seal may be done optionally here, or explicitly
	// via the VerifySeal method.
	VerifyHeader(chain ChainReader, header *types.Header, seal bool, ai bool) error

	// VerifyHeaders is similar to VerifyHeader, but verifies a batch of headers
	// concurrently. The method returns a quit channel to abort the operations and
	// a results channel to retrieve the async verifications (the order is that of
	// the input slice).
	VerifyHeaders(chain ChainReader, headers []*types.Header, seals []bool, ais []bool) (chan<- struct{}, <-chan error)

	// VerifyUncles verifies that the given block's uncles conform to the consensus
	// rules of a given engine.
	VerifyUncles(chain ChainReader, block *types.Block) error

	// VerifySeal checks whether the crypto seal on a header is valid according to
	// the consensus rules of the given engine.
	VerifySeal(chain ChainReader, header *types.Header) error

	VerifyAISeal(chain ChainReader, header *types.Header) error

	VerifyBasePow(chain ChainReader, header *types.Header, basePower types.BasePowers) error

	// Prepare initializes the consensus fields of a block header according to the
	// rules of a particular engine. The changes are executed inline.
	Prepare(chain ChainReader, header *types.Header) error

	// Finalize runs any post-transaction state modifications (e.g. block rewards)
	// and assembles the final block.
	// Note: The block header and state database might be updated to reflect any
	// consensus rules that happen at finalization (e.g. block rewards).
	Finalize(chain ChainReader, header *types.Header, state *state.StateDBManage,
		uncles []*types.Header, currencyBlock []types.CurrencyBlock) (*types.Block, error)

	// Finalize runs any post-transaction state modifications (e.g. block rewards)
	// and assembles the final block.
	// Note: The block header and state database might be updated to reflect any
	// consensus rules that happen at finalization (e.g. block rewards).
	GenOtherCurrencyBlock(chain ChainReader, header *types.Header, state *state.StateDBManage,
		uncles []*types.Header, currencyBlock []types.CurrencyBlock) (*types.Block, error)

	// Finalize runs any post-transaction state modifications (e.g. block rewards)
	// and assembles the final block.
	// Note: The block header and state database might be updated to reflect any
	// consensus rules that happen at finalization (e.g. block rewards).
	GenManBlock(chain ChainReader, header *types.Header, state *state.StateDBManage,
		uncles []*types.Header, currencyBlock []types.CurrencyBlock) (*types.Block, error)

	// Seal generates a new block for the given input block with the local miner's
	// seal place on top.
	SealAI(chain ChainReader, header *types.Header, stop <-chan struct{}) (*types.Header, error)
	SealPow(chain ChainReader, header *types.Header, stop <-chan struct{}, resultchan chan<- *types.Header, isBroadcastNode bool) (*types.Header, error)

	// CalcDifficulty is the difficulty adjustment algorithm. It returns the difficulty
	// that a new block should have.
	CalcDifficulty(chain ChainReader, version string, time uint64, parent *types.Header) (*big.Int, error)

	// APIs returns the RPC APIs this consensus engine provides.
	APIs(chain ChainReader) []rpc.API
}

// PoW is a consensus engine based on proof-of-work.
type PoW interface {
	Engine

	// Hashrate returns the current mining hashrate of a PoW consensus engine.
	Hashrate() float64
}

type StateReader interface {
	GetCurrentHash() common.Hash
	GetGraphByHash(hash common.Hash) (*mc.TopologyGraph, *mc.ElectGraph, error)
	GetBroadcastAccounts(blockHash common.Hash) ([]common.Address, error)
	GetVersionSuperAccounts(blockHash common.Hash) ([]common.Address, error)
	GetBlockSuperAccounts(blockHash common.Hash) ([]common.Address, error)
	GetBroadcastIntervalByHash(blockHash common.Hash) (*mc.BCIntervalInfo, error)
	GetA0AccountFromAnyAccount(account common.Address, blockHash common.Hash) (common.Address, common.Address, error)
}

type DPOSEngine interface {
	VerifyVersionSigns(reader StateReader, header *types.Header) error

	CheckSuperBlock(reader StateReader, header *types.Header) error

	VerifyBlock(reader StateReader, header *types.Header) error

	//verify hash in current block
	VerifyHash(reader StateReader, signHash common.Hash, signs []common.Signature) ([]common.Signature, error)

	//verify hash in given block
	VerifyHashWithBlock(reader StateReader, signHash common.Hash, signs []common.Signature, blockHash common.Hash) ([]common.Signature, error)

	VerifyHashWithVerifiedSigns(reader StateReader, signs []*common.VerifiedSign) ([]common.Signature, error)

	VerifyHashWithVerifiedSignsAndBlock(reader StateReader, signs []*common.VerifiedSign, blockHash common.Hash) ([]common.Signature, error)
}
