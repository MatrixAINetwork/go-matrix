package mineroutreward

import (
	"github.com/pkg/errors"
	"math/big"

	"github.com/matrix/go-matrix/params/manparams"

	"github.com/matrix/go-matrix/core/matrixstate"
	"github.com/matrix/go-matrix/mc"

	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/core/types"
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
}

func (mr *MinerOutReward) GetPreMinerReward(state util.StateDB, rewardType uint8) (*big.Int, error) {
	var key string
	if util.TxsReward == rewardType {
		key = mc.MSKeyPreMinerTxsReward
	} else {
		key = mc.MSKeyPreMinerBlkReward
	}
	preMiner, err := matrixstate.GetDataByState(key, state)
	if nil != err {
		log.Error(PackageName, "获取矿工奖励金额错误", err, "key", key)
		return nil, errors.New("获取矿工金额错误")
	}
	if preMiner == nil {
		log.Error(PackageName, "反射失败", err, "key", key)
		return nil, errors.New("反射失败")
	}
	currentReward, ok := preMiner.(*mc.MinerOutReward)
	if !ok {
		log.Error(PackageName, "类型转换失败", err, "key", key)
		return nil, errors.New("类型转换失败")
	}
	log.INFO(PackageName, "获取前一个矿工奖励值为", currentReward.Reward, "key", key)
	return &currentReward.Reward, nil

}

func (mr *MinerOutReward) SetPreMinerReward(state util.StateDB, reward *big.Int, rewardType uint8) {
	var key string
	if util.TxsReward == rewardType {
		key = mc.MSKeyPreMinerTxsReward
	} else {
		key = mc.MSKeyPreMinerBlkReward
	}
	log.INFO(PackageName, "设置前矿工奖励值为", reward, "key", key)
	minerOutReward := &mc.MinerOutReward{Reward: *reward}
	matrixstate.SetDataToState(key, minerOutReward, state)
	return

}

func (mr *MinerOutReward) SetMinerOutRewards(curReward *big.Int, state util.StateDB, num uint64, parentHash common.Hash, reader util.ChainReader, innerMiners []common.Address, rewardType uint8) map[common.Address]*big.Int {
	//后一块给前一块的矿工发钱，广播区块不发钱， 广播区块下一块给广播区块前一块发钱

	bcInterval, err := manparams.NewBCIntervalByHash(parentHash)
	if err != nil {
		log.Error(PackageName, "获取广播周期失败", err)
		return nil
	}
	if bcInterval.IsBroadcastNumber(num) {
		log.WARN(PackageName, "广播区块不发钱：", num)
		return nil
	}

	preReward, err := mr.GetPreMinerReward(state, rewardType)
	mr.SetPreMinerReward(state, curReward, rewardType)
	if nil != err {
		return nil
	}

	coinBase, err := mr.canSetMinerOutRewards(num, preReward, reader, bcInterval, parentHash, innerMiners)
	if nil != err {
		return nil
	}

	rewards := make(map[common.Address]*big.Int)
	util.SetAccountRewards(rewards, coinBase, preReward)
	log.Info(PackageName, "出块矿工账户：", coinBase.String(), "发放奖励高度", num, "奖励金额", preReward)
	return rewards
}

func (mr *MinerOutReward) canSetMinerOutRewards(num uint64, reward *big.Int, reader util.ChainReader, bcInterval *manparams.BCInterval, parentHash common.Hash, innerMiners []common.Address) (common.Address, error) {
	if num < 2 {
		log.Debug(PackageName, "高度为小于2 不发放奖励：", "")
		return common.Address{}, errors.New("高度为小于2 不发放奖励：")
	}

	if reward.Cmp(big.NewInt(0)) <= 0 {
		log.WARN(PackageName, "奖励金额不合法", reward)
		return common.Address{}, errors.New("奖励金额不合法")
	}

	var header *types.Header
	var originNum uint64
	originNum = num - 100
	if num < 101 {
		originNum = 0
	}
	preHash := parentHash
	for i := num - 1; i > originNum; i-- {
		header = reader.GetHeaderByHash(preHash)
		preHash = header.ParentHash
		if bcInterval.IsBroadcastNumber(i) {
			continue
		}
		if nil == header {
			log.ERROR(PackageName, "获取区块头错误，高度为", i)
			break
		}
		if !header.IsSuperHeader() {
			break
		}
	}
	if nil == header {
		log.ERROR(PackageName, "无法获取区块头错误", num)
		return common.Address{}, errors.New("无法获取区块头错误")
	}
	coinbase := header.Coinbase
	if coinbase.Equal(common.Address{}) {
		log.ERROR(PackageName, "矿工奖励的地址非法", coinbase.Hex())
		return common.Address{}, errors.New("矿工奖励的地址非法")
	}
	for _, v := range innerMiners {
		if coinbase.Equal(v) {
			log.Warn(PackageName, "基金会矿工不发钱，账户为", coinbase)
			return common.Address{}, errors.New("基金会矿工")
		}
	}
	return coinbase, nil
}
