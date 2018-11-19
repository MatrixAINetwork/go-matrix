package blkreward

import (
	"math/big"

	"github.com/matrix/go-matrix/reward/rewardexec"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/params"
	"github.com/matrix/go-matrix/reward"
	"github.com/matrix/go-matrix/reward/cfg"
)

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

	GetBlockByNumber(number uint64) *types.Block

	// GetBlock retrieves a block sfrom the database by hash and number.
	GetBlock(hash common.Hash, number uint64) *types.Block
	StateAt(root common.Hash) (*state.StateDB, error)
	State() (*state.StateDB, error)
	NewTopologyGraph(header *types.Header) (*mc.TopologyGraph, error)
}

type blkreward struct {
	blockReward *rewardexec.BlockReward
}

func New(chain ChainReader) reward.Reward {

	rewardCfg := cfg.New(nil, nil)
	return rewardexec.New(chain, rewardCfg)
}

func (tr *blkreward) CalcBlockRewards(blockReward *big.Int, Leader common.Address, header *types.Header) map[common.Address]*big.Int {
	return tr.blockReward.CalcBlockRewards(blockReward, Leader, header)
}
