package util

import (
	"math/big"

	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/mc"

	"github.com/matrix/go-matrix/core"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/params"

	"github.com/matrix/go-matrix/common"
)

const (
	PackageName = "奖励"
)
const (
	RewardFullRate = uint64(10000)
)

var (
	FrontierBlockReward  *big.Int = big.NewInt(5e+18) // Block reward in wei for successfully mining a block
	ByzantiumBlockReward *big.Int = big.NewInt(3e+18) // Block reward in wei for successfully mining a block upward from Byzantium
	//分母10000
	ByzantiumTxsRewardDen *big.Int = big.NewInt(10000) // Block reward in wei for successfully mining a block upward from Byzantium
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

func SetAccountRewards(rewards map[common.Address]*big.Int, account common.Address, reward *big.Int) {
	if _, ok := rewards[account]; ok {
		rewards[account] = rewards[account].Add(rewards[account], reward)
	} else {
		rewards[account] = reward
	}
}

func CalcRateReward(rewardAmount *big.Int, rate uint64) *big.Int {
	temp := new(big.Int).Mul(rewardAmount, new(big.Int).SetUint64(rate))
	return new(big.Int).Div(temp, new(big.Int).SetUint64(RewardFullRate))
}

func calcTxGas(chain ChainReader, tx types.SelfTransaction, BlockNumber *big.Int) *big.Int {
	gas, err := core.IntrinsicGas(tx.Data())
	if err != nil {
		return new(big.Int).SetUint64(0)
	}
	//YY
	tmpExtra := tx.GetMatrix_EX()
	if len(tmpExtra) != 0 {
		if uint64(len(tmpExtra[0].ExtraTo)) > params.TxCount-1 { //减1是为了和txpool中的验证统一，因为还要算上外层的那笔交易
			return new(big.Int).SetUint64(0)
		}
		for _, ex := range tmpExtra[0].ExtraTo {
			tmpgas, tmperr := core.IntrinsicGas(ex.Payload)
			if tmperr != nil {
				return new(big.Int).SetUint64(0)
			}
			//0.7+0.3*pow(0.9,(num-1))
			gas += tmpgas
		}
	}
	return new(big.Int).Mul(tx.GasPrice(), new(big.Int).SetUint64(gas))
}
func calcTxsFees(chain ChainReader, num *big.Int, txs types.SelfTransactions) *big.Int {
	var transactionsRewards *big.Int
	transactionsRewards = big.NewInt(int64(0))
	for _, tx := range txs {
		txFees := calcTxGas(chain, tx, num)
		transactionsRewards.Add(transactionsRewards, txFees)
	}
	log.INFO(PackageName, "all txs Reward", transactionsRewards)
	return transactionsRewards
}
