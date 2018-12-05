package util

import (
	"github.com/matrix/go-matrix/log"
	"math/big"
	"sort"

	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/mc"

	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/params"

	"github.com/matrix/go-matrix/common"
)

const (
	PackageName = "奖励util"
)
const (
	RewardFullRate = uint64(10000)
)

var (
	//ValidatorBlockReward  *big.Int = new(big.Int).Exp(big.NewInt(10), big.NewInt(18), big.NewInt(0)) // Block reward in wei for successfully mining a block
	MultilCoinBlockReward *big.Int =  new(big.Int).Exp(big.NewInt(10), big.NewInt(18), big.NewInt(0)) // Block reward in wei for successfully mining a block upward from Byzantium
	//分母10000
	ByzantiumTxsRewardDen *big.Int = big.NewInt(10000) // Block reward in wei for successfully mining a block upward from Byzantium
	ValidatorsBlockReward  *big.Int = big.NewInt(5e+18)
	MinersBlockReward *big.Int = big.NewInt(5e+18)
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
	Genesis() *types.Block
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


func CalcDepositRate(reward *big.Int, depositNodes map  [common.Address]*big.Int,rewards map[common.Address ]*big.Int )  {

	totalDeposit := new(big.Int)
	if 0==len(depositNodes){
		log.ERROR(PackageName,"抵押列表为nil","")
		return
	}
	depositNodesFix := make(map[common.Address ]*big.Int)
	for k, v := range depositNodes {
		depositTemp := new(big.Int).Div(v,big.NewInt(1e18))
		depositNodesFix[k] = depositTemp
		totalDeposit.Add(totalDeposit, depositTemp)
	}
	if 0==totalDeposit.Cmp(big.NewInt(0)){
		log.ERROR(PackageName,"定点化抵押值为0","")
		return
	}
	log.INFO(PackageName,"计算抵押总额,账户总抵押",totalDeposit,"定点化抵押", totalDeposit)

	rewardFixed :=new(big.Int).Div(reward,big.NewInt(1e8))


	sorted_keys := make([]string, 0)

	for k, _ := range depositNodesFix {
		sorted_keys = append(sorted_keys, k.String())
	}
	sort.Strings(sorted_keys)
	for _,k:=range sorted_keys{
		rateTemp := new(big.Int).Mul(depositNodesFix[common.HexToAddress(k)],big.NewInt(1e10))
		rate:=new(big.Int).Div(rateTemp,totalDeposit)
		log.INFO(PackageName,"计算比例,账户",k,"定点化比例", rate)

		rewardTemp:= new(big.Int).Mul(rewardFixed,rate)
		rewardTemp1:= new(big.Int).Div(rewardTemp,big.NewInt(1e10))
		oneNodeReward :=  new(big.Int).Mul(rewardTemp1,big.NewInt(1e8))
		SetAccountRewards(rewards, common.HexToAddress(k),oneNodeReward)
		log.INFO(PackageName,"计算奖励金额,账户",k,"定点化金额", rewards[common.HexToAddress(k)] )
		log.INFO(PackageName,"","" )
	}
}

//func calcTxGas(chain ChainReader, tx types.SelfTransaction, BlockNumber *big.Int) *big.Int {
//	gas, err := core.IntrinsicGas(tx.Data())
//	if err != nil {
//		return new(big.Int).SetUint64(0)
//	}
//	//YY
//	tmpExtra := tx.GetMatrix_EX()
//	if len(tmpExtra) != 0 {
//		if uint64(len(tmpExtra[0].ExtraTo)) > params.TxCount-1 { //减1是为了和txpool中的验证统一，因为还要算上外层的那笔交易
//			return new(big.Int).SetUint64(0)
//		}
//		for _, ex := range tmpExtra[0].ExtraTo {
//			tmpgas, tmperr := core.IntrinsicGas(ex.Payload)
//			if tmperr != nil {
//				return new(big.Int).SetUint64(0)
//			}
//			//0.7+0.3*pow(0.9,(num-1))
//			gas += tmpgas
//		}
//	}
//	return new(big.Int).Mul(tx.GasPrice(), new(big.Int).SetUint64(gas))
//}
//func calcTxsFees(chain ChainReader, num *big.Int, txs types.SelfTransactions) *big.Int {
//	var transactionsRewards *big.Int
//	transactionsRewards = big.NewInt(int64(0))
//	for _, tx := range txs {
//		txFees := calcTxGas(chain, tx, num)
//		transactionsRewards.Add(transactionsRewards, txFees)
//	}
//	log.INFO(PackageName, "all txs Reward", transactionsRewards)
//	return transactionsRewards
//}
