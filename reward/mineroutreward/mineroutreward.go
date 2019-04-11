package mineroutreward

import (
	"math/big"

	"github.com/MatrixAINetwork/go-matrix/params/manparams"

	"github.com/pkg/errors"

	"github.com/MatrixAINetwork/go-matrix/core/matrixstate"
	"github.com/MatrixAINetwork/go-matrix/mc"

	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/params"

	"github.com/MatrixAINetwork/go-matrix/reward/util"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/log"
)

type MinerOutReward struct {
	PreReward   []mc.MultiCoinMinerOutReward
	InnerMiners []common.Address
	RewardType  uint8
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
	GetHeader(hash []common.CoinRoot, number uint64) *types.Header

	// GetHeaderByNumber retrieves a block header from the database by number.
	GetHeaderByNumber(number uint64) *types.Header

	// GetHeaderByHash retrieves a block header from the database by its hash.
	GetHeaderByHash(hash []common.CoinRoot) *types.Header

	GetBlockByNumber(number uint64) *types.Block

	// GetBlock retrieves a block sfrom the database by hash and number.
	GetBlock(hash []common.CoinRoot, number uint64) *types.Block
	StateAt(root []common.CoinRoot) (*state.StateDBManage, error)
	State() (*state.StateDBManage, error)
}

func SetPreMinerReward(state util.StateDB, reward *big.Int, rewardType uint8, coinType string) {
	//log.INFO(PackageName, "设置前矿工奖励值为", reward, "type", rewardType)
	minerOutReward := &mc.MinerOutReward{Reward: *reward}
	var err error
	if util.TxsReward == rewardType {
		version := matrixstate.GetVersionInfo(state)
		if version == manparams.VersionAlpha {
			err = matrixstate.SetPreMinerTxsReward(state, minerOutReward)
		} else if version == manparams.VersionBeta {
			multiCoinMinerOut, err := matrixstate.GetPreMinerMultiCoinTxsReward(state)
			if err != nil {
				log.Error(PackageName, "获取前矿工奖励值错误", err)
			}
			multiCoinMinerOut = addMultiCoinMinerOutReward(coinType, reward, multiCoinMinerOut)
			err = matrixstate.SetPreMinerMultiCoinTxsReward(state, multiCoinMinerOut)
		} else {
			log.Error(PackageName, "设置前矿工奖励值版本号错误", version)
			return
		}
	} else {
		err = matrixstate.SetPreMinerBlkReward(state, minerOutReward)
	}
	if err != nil {
		log.Error(PackageName, "设置前矿工奖励值错误", err)
		return
	}
	log.INFO(PackageName, "设置前一个矿工奖励值为", reward, "type", rewardType, "币种", coinType)
	return
}

func findMultiCoinMinerOutReward(cointype string, multiCoinMinerOut []mc.MultiCoinMinerOutReward) *big.Int {
	for _, v := range multiCoinMinerOut {
		if v.CoinType == cointype {
			return &v.Reward
		}
	}
	return nil
}

func addMultiCoinMinerOutReward(cointype string, reward *big.Int, multiCoinMinerOut []mc.MultiCoinMinerOutReward) []mc.MultiCoinMinerOutReward {
	if multiCoinMinerOut == nil {
		multiCoinMinerOut = make([]mc.MultiCoinMinerOutReward, 0)
	}
	minerOutReward := mc.MultiCoinMinerOutReward{CoinType: cointype, Reward: *reward}
	findFlag := false
	for i, v := range multiCoinMinerOut {
		if v.CoinType == cointype {
			multiCoinMinerOut[i] = minerOutReward
			findFlag = true
		}
	}
	if findFlag == false {
		multiCoinMinerOut = append(multiCoinMinerOut, minerOutReward)
	}
	return multiCoinMinerOut
}

func (mr *MinerOutReward) SetMinerOutRewards(curReward *big.Int, state util.StateDB, num uint64, parentHash common.Hash, reader util.ChainReader, coinType string) map[common.Address]*big.Int {
	//后一块给前一块的矿工发钱，广播区块不发钱， 广播区块下一块给广播区块前一块发钱
	bcInterval, err := matrixstate.GetBroadcastInterval(state)
	if err != nil {
		log.Error(PackageName, "获取广播周期失败", err)
		return nil
	}
	if bcInterval.IsBroadcastNumber(num) {
		log.WARN(PackageName, "广播区块不发钱：", num)
		return nil
	}

	//preReward, err := mr.GetPreMinerReward(state, rewardType, coinType)
	reward := findMultiCoinMinerOutReward(coinType, mr.PreReward)
	SetPreMinerReward(state, curReward, mr.RewardType, coinType)
	if nil == reward {
		log.Error(coinType+PackageName, "无法获取对应币种奖励", "")
		return nil
	}

	coinBase, err := mr.canSetMinerOutRewards(num, reward, reader, bcInterval, parentHash, mr.InnerMiners)
	if nil != err {
		return nil
	}

	rewards := make(map[common.Address]*big.Int)
	util.SetAccountRewards(rewards, coinBase, reward)
	log.Debug(PackageName, "出块矿工账户：", coinBase.String(), "发放奖励高度", num, "奖励金额", reward)
	return rewards
}

func (mr *MinerOutReward) canSetMinerOutRewards(num uint64, reward *big.Int, reader util.ChainReader, bcInterval *mc.BCIntervalInfo, parentHash common.Hash, innerMiners []common.Address) (common.Address, error) {
	if num < 2 {
		log.Debug(PackageName, "高度为小于2 不发放奖励：", "")
		return common.Address{}, errors.New("高度为小于2 不发放奖励：")
	}

	if reward.Cmp(big.NewInt(0)) <= 0 {
		//log.WARN(PackageName, "奖励金额不合法", reward)
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
