package util

import (
	"errors"
	"math/big"
	"sort"

	"github.com/MatrixAINetwork/go-matrix/core/matrixstate"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/params/manparams"

	"github.com/MatrixAINetwork/go-matrix/log"

	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/params"

	"github.com/MatrixAINetwork/go-matrix/common"
)

const (
	PackageName = "奖励util"
)
const (
	RewardFullRate = uint64(10000)
	Stop           = "0"
	TxsReward      = 0
	BlkReward      = 1
)

var (
	//ValidatorBlockReward  *big.Int = new(big.Int).Exp(big.NewInt(10), big.NewInt(18), big.NewInt(0)) // Block reward in wei for successfully mining a block
	MultilCoinBlockReward *big.Int = new(big.Int).Exp(big.NewInt(10), big.NewInt(18), big.NewInt(0)) // Block reward in wei for successfully mining a block upward from Byzantium
	//分母10000
	ByzantiumTxsRewardDen *big.Int = big.NewInt(1000000000) // Block reward in wei for successfully mining a block upward from Byzantium
	ValidatorsBlockReward *big.Int = big.NewInt(5e+18)
	MinersBlockReward     *big.Int = big.NewInt(5e+18)

	ManPrice *big.Int = big.NewInt(1e18)

	Precision *big.Int = big.NewInt(1)
)

type ChainReader interface {
	// Config retrieves the blockchain's chain configuration.
	//Config() *params.ChainConfig

	// CurrentHeader retrieves the current header from the local chain.

	// GetHeader retrieves a block header from the database by hash and number.
	//GetHeader(hash common.Hash, number uint64) *types.Header

	// GetHeaderByNumber retrieves a block header from the database by number.

	// GetHeaderByHash retrieves a block header from the database by its hash.
	GetHeaderByHash(hash common.Hash) *types.Header

	// GetBlock retrieves a block sfrom the database by hash and number.
	StateAt(root []common.CoinRoot) (*state.StateDBManage, error)
	State() (*state.StateDBManage, error)
	GetGraphByState(state matrixstate.StateDB) (*mc.TopologyGraph, *mc.ElectGraph, error)
	StateAtBlockHash(hash common.Hash) (*state.StateDBManage, error)
	GetAncestorHash(sonHash common.Hash, ancestorNumber uint64) (common.Hash, error)
}

type StateDB interface {
	GetBalance(typ string, addr common.Address) common.BalanceType
	GetMatrixData(hash common.Hash) (val []byte)
	SetMatrixData(hash common.Hash, val []byte)
}

type DepositInfo struct {
	Deposit  *big.Int
	FixStock uint64
}

func SetAccountRewards(rewards map[common.Address]*big.Int, account common.Address, reward *big.Int) {

	if 0 == reward.Cmp(big.NewInt(0)) {
		return
	}
	if nil == rewards {
		return
	}
	if account.Equal(common.Address{}) {
		log.ERROR(PackageName, "奖励的地址非法", account.Hex())
		return
	}
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

func CalcStockRate(reward *big.Int, depositNodes map[common.Address]DepositInfo) map[common.Address]*big.Int {

	if 0 == len(depositNodes) {
		log.ERROR(PackageName, "抵押列表为空", "")
		return nil
	}
	totalStock := uint64(0)

	for _, v := range depositNodes {

		totalStock = v.FixStock + totalStock
	}

	//log.INFO(PackageName, "计算抵押总额,账户股权", totalStock)

	sortedKeys := make([]string, 0)

	for k := range depositNodes {
		sortedKeys = append(sortedKeys, k.String())
	}
	sort.Strings(sortedKeys)
	rewards := make(map[common.Address]*big.Int)
	for _, k := range sortedKeys {
		temp := new(big.Int).Mul(reward, new(big.Int).SetUint64(uint64(depositNodes[common.HexToAddress(k)].FixStock)))
		oneNodeReward := new(big.Int).Div(temp, new(big.Int).SetUint64(uint64(totalStock)))
		rewards[common.HexToAddress(k)] = oneNodeReward
		//log.Debug(PackageName, "计算奖励金额,账户", k, "奖励金额", oneNodeReward)
	}
	return rewards
}

func CalcInterestReward(reward *big.Int, interest map[common.Address]*big.Int) map[common.Address]*big.Int {

	if 0 == len(interest) {
		log.ERROR(PackageName, "利息列表为空", "")
		return nil
	}
	totalInterest := new(big.Int)

	for _, v := range interest {

		totalInterest.Add(totalInterest, v)
	}
	if totalInterest.Cmp(big.NewInt(0)) <= 0 {
		log.ERROR(PackageName, "计算的总利息值非法", totalInterest)
		return nil
	}
	log.Trace(PackageName, "计算的总利息值", totalInterest)

	if 0 == reward.Cmp(big.NewInt(0)) {
		log.ERROR(PackageName, "定点化奖励金额为0", "")
		return nil
	}

	rewards := make(map[common.Address]*big.Int)
	for k, v := range interest {
		temp := new(big.Int).Mul(reward, v)
		rewards[k] = new(big.Int).Div(temp, totalInterest)
		log.Trace(PackageName, "计算奖励金额,账户", k, "金额", rewards[k])
	}
	return rewards
}
func MergeReward(dst map[common.Address]*big.Int, src map[common.Address]*big.Int) {
	if 0 == len(src) {
		return
	}
	if nil == dst {
		log.ERROR(PackageName, "dst is nil", dst)
		return
	}
	for account, reward := range src {

		SetAccountRewards(dst, account, reward)
	}

}
func CalcN(halfNum uint64, num uint64) uint64 {
	n := uint64(0)
	if 0 != halfNum {
		n = num / halfNum
	}
	return n
}

func CalcRewardMount(blockReward *big.Int, n uint64, x uint16) *big.Int {
	var reward *big.Int
	if 0 == n {
		reward = blockReward
	} else {
		rate := new(big.Int).Exp(new(big.Int).SetUint64(uint64(x)), new(big.Int).SetUint64(n), big.NewInt(0))
		tmp := new(big.Int).Mul(blockReward, rate)
		base := new(big.Int).Exp(new(big.Int).SetUint64(mc.RewardFullRate), new(big.Int).SetUint64(n), big.NewInt(0))
		reward := new(big.Int).Div(tmp, base)
		return reward
	}
	return reward
}

func CalcRewardMountByNumber(st StateDB, blockReward *big.Int, num uint64, halfNum uint64, address common.Address, attenuationRate uint16) *big.Int {

	if blockReward.Cmp(big.NewInt(0)) < 0 {
		log.WARN(PackageName, "折半计算的奖励金额不合法", blockReward)
		return big.NewInt(0)
	}

	balance, err := getBalance(st, address)
	if nil != err {
		log.ERROR(PackageName, "账户余额获取错误，账户为", address.Hex())
		return big.NewInt(0)
	}

	n := CalcN(halfNum, num)

	reward := CalcRewardMount(blockReward, n, attenuationRate)
	log.Debug(PackageName, "计算衰减奖励金额:", reward.String())
	if balance[common.MainAccount].Balance.Cmp(reward) < 0 {
		log.ERROR(PackageName, "账户余额不足，余额为", balance[common.MainAccount].Balance.String())
		return big.NewInt(0)
	} else {
		return reward
	}

}

func getBalance(st StateDB, address common.Address) (common.BalanceType, error) {

	balance := st.GetBalance(params.MAN_COIN, address)
	if len(balance) == 0 {
		log.ERROR(PackageName, "账户余额获取不到", "")
		return nil, errors.New("账户余额获取不到")
	}
	if balance[common.MainAccount].Balance.Cmp(big.NewInt(0)) < 0 {
		log.WARN(PackageName, "发送账户余额不合法，地址", address.Hex(), "余额", balance[common.MainAccount].Balance)
		return nil, errors.New("发送账户余额不合法")
	}
	return balance, nil
}
func getBalanceByCoinType(st StateDB, address common.Address, cointype string) (common.BalanceType, error) {

	balance := st.GetBalance(cointype, address)
	if len(balance) == 0 {
		log.ERROR(PackageName, "账户余额获取不到", "")
		return nil, errors.New("账户余额获取不到")
	}
	if balance[common.MainAccount].Balance.Cmp(big.NewInt(0)) < 0 {
		log.WARN(PackageName, "发送账户余额不合法，地址", address.Hex(), "余额", balance[common.MainAccount].Balance)
		return nil, errors.New("发送账户余额不合法")
	}
	return balance, nil
}

func getRewardSum(reardMap map[common.Address]*big.Int) *big.Int {
	sum := big.NewInt(0)
	for _, v := range reardMap {
		sum.Add(sum, v)
	}
	return sum
}

func CointypeCheck(st StateDB, rewardIn []common.RewarTx) []common.RewarTx {
	rewardMap := make(map[string]*big.Int)

	for _, v := range rewardIn {
		if v.RewardTyp == common.RewardTxsType {
			if _, ok := rewardMap[v.CoinType]; ok {
				rewardMap[v.CoinType] = new(big.Int).Add(rewardMap[v.CoinType], getRewardSum(v.To_Amont))
			} else {
				rewardMap[v.CoinType] = getRewardSum(v.To_Amont)
			}
		}
	}

	for coinType, all := range rewardMap {
		balance, err := getBalanceByCoinType(st, common.TxGasRewardAddress, coinType)
		log.Info(PackageName, "发放币种", coinType, "计算的奖励总额为", all, "账户余额为", balance[common.MainAccount].Balance)
		if nil != err {
			continue
		}
		if all.Cmp(balance[common.MainAccount].Balance) > 0 {
			log.Crit(PackageName, "发放币种", coinType, "交易费奖励余额不足，计算的奖励总额为", all, "账户余额为", balance[common.MainAccount].Balance)
			return nil
		}
	}
	return rewardIn
}

func AccumulatorCheck(st StateDB, rewardIn []common.RewarTx) []common.RewarTx {

	ValidatorBalance, _ := getBalance(st, common.BlkMinerRewardAddress)
	minerBalance, _ := getBalance(st, common.BlkValidatorRewardAddress)
	interestBalance, _ := getBalance(st, common.InterestRewardAddress)
	lotteryBalance, _ := getBalance(st, common.LotteryRewardAddress)
	allValidator := new(big.Int).SetUint64(0)
	allMiner := new(big.Int).SetUint64(0)
	allInterest := new(big.Int).SetUint64(0)
	allLottery := new(big.Int).SetUint64(0)
	for _, v := range rewardIn {
		if v.Fromaddr == common.BlkMinerRewardAddress {
			for account, Amount := range v.To_Amont {
				if !account.Equal(common.ContractAddress) {
					allMiner = new(big.Int).Add(allMiner, Amount)
				}
			}
		}
		if v.Fromaddr == common.BlkValidatorRewardAddress {
			for account, Amount := range v.To_Amont {
				if !account.Equal(common.ContractAddress) {
					allValidator = new(big.Int).Add(allValidator, Amount)
				}
			}
		}

		if v.Fromaddr == common.InterestRewardAddress {
			for account, Amount := range v.To_Amont {
				if !account.Equal(common.ContractAddress) {
					allInterest = new(big.Int).Add(allInterest, Amount)
				}

			}
		}
		if v.Fromaddr == common.LotteryRewardAddress {
			for account, Amount := range v.To_Amont {
				if !account.Equal(common.ContractAddress) {
					allLottery = new(big.Int).Add(allLottery, Amount)
				}
			}
		}
	}

	rewardOut := make([]common.RewarTx, 0)

	if allMiner.Cmp(minerBalance[common.MainAccount].Balance) <= 0 {
		for _, v := range rewardIn {
			if v.RewardTyp == common.RewardMinerType {
				rewardOut = append(rewardOut, v)
			}
		}

	} else {
		log.Error(PackageName, "矿工账户余额不足,余额", allMiner.String())
	}
	if allValidator.Cmp(ValidatorBalance[common.MainAccount].Balance) <= 0 {
		for _, v := range rewardIn {
			if v.RewardTyp == common.RewardValidatorType {
				rewardOut = append(rewardOut, v)
			}
		}
	} else {
		log.Error(PackageName, "验证者账户余额不足", allValidator.String())
	}

	for _, v := range rewardIn {
		if v.RewardTyp == common.RewardTxsType {
			rewardOut = append(rewardOut, v)
		}
	}

	if allInterest.Cmp(interestBalance[common.MainAccount].Balance) <= 0 {
		for _, v := range rewardIn {
			if v.RewardTyp == common.RewardInterestType {
				rewardOut = append(rewardOut, v)
			}
		}
	} else {
		log.Error(PackageName, "利息账户余额不足", allInterest.String())
	}

	if allLottery.Cmp(lotteryBalance[common.MainAccount].Balance) <= 0 {
		for _, v := range rewardIn {
			//通过类型判断
			if v.RewardTyp == common.RewardLotteryType {
				rewardOut = append(rewardOut, v)
			}
		}
	} else {
		log.Error(PackageName, "彩票账户余额不足", allLottery.String())
	}

	return rewardOut
}

func GetPreMinerReward(state StateDB, rewardType uint8) ([]mc.MultiCoinMinerOutReward, error) {
	var currentReward *mc.MinerOutReward
	var err error
	if TxsReward == rewardType {
		version := matrixstate.GetVersionInfo(state)
		if version == manparams.VersionAlpha {
			currentReward, err = matrixstate.GetPreMinerTxsReward(state)
			if err != nil {
				log.Error(PackageName, "获取矿工交易奖励金额错误", err)
				return nil, errors.New("获取矿工交易金额错误")
			}
		} else if version == manparams.VersionBeta {
			multiCoin, err := matrixstate.GetPreMinerMultiCoinTxsReward(state)
			if err != nil {
				log.Error(PackageName, "获取矿工交易奖励金额错误", err)
				return make([]mc.MultiCoinMinerOutReward, 0), errors.New("获取矿工交易金额错误")
			}
			for _, v := range multiCoin {
				log.Trace(PackageName, "获取前一个矿工奖励值为", v.Reward, "type", v.CoinType)
			}

			return multiCoin, nil
		} else {
			log.Error(PackageName, "获取前矿工奖励值版本号错误", version)
		}
	} else {
		currentReward, err = matrixstate.GetPreMinerBlkReward(state)
		if err != nil {
			log.Error(PackageName, "获取矿工区块奖励金额错误", err)
			return nil, errors.New("获取矿工区块金额错误")
		}
	}
	multiCoinMinerOut := make([]mc.MultiCoinMinerOutReward, 0)
	minerOutReward := mc.MultiCoinMinerOutReward{CoinType: params.MAN_COIN, Reward: currentReward.Reward}
	log.INFO(PackageName, "获取前一个矿工奖励值为", currentReward.Reward, "type", rewardType)
	multiCoinMinerOut = append(multiCoinMinerOut, minerOutReward)
	return multiCoinMinerOut, nil

}
