package mineroutreward

import (
	"math/big"

	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/params"

	"github.com/matrix/go-matrix/reward/util"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/log"
)

type MinerOutReward struct {
}

const (
	PackageName = "矿工挖矿奖励"
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

func (mr *MinerOutReward) SetMinerOutRewards(reward *big.Int, chain ChainReader, num *big.Int, rewards map[common.Address]*big.Int) {
	//后一块给前一块的矿工发钱，广播区块下一块的不用给广播节点发钱
	if num.Uint64() < uint64(2) || common.IsBroadcastNumber(num.Uint64()-1) {
		log.WARN(PackageName, "miner out height is wrong,height", num.Uint64())
		return
	}
	coinBase := chain.GetHeaderByNumber(num.Uint64() - 1).Coinbase
	util.SetAccountRewards(rewards, coinBase, reward)
	log.Info(PackageName, "miner out coinbase", coinBase.String(), "height", num.Uint64()-1, "reward", reward)
}
