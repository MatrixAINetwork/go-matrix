// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package interest

import (
	"math"
	"math/big"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/matrixstate"
	"github.com/MatrixAINetwork/go-matrix/reward/util"

	"github.com/MatrixAINetwork/go-matrix/mc"

	"github.com/MatrixAINetwork/go-matrix/ca"
	"github.com/MatrixAINetwork/go-matrix/core/vm"
	"github.com/MatrixAINetwork/go-matrix/depoistInfo"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/params"
)

const (
	PackageName = "利息奖励"
	Denominator = 1000000000
	POWER       = float64(1)
)

type interest struct {
	VIPConfig      []mc.VIPConfig
	InterestConfig *mc.InterestCfg
	Calc           string
}

type DepositInterestRate struct {
	Deposit  *big.Int
	Interest uint64
}

type DepositInterestRateList []*DepositInterestRate

func (p DepositInterestRateList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p DepositInterestRateList) Len() int           { return len(p) }
func (p DepositInterestRateList) Less(i, j int) bool { return p[i].Deposit.Cmp(p[j].Deposit) < 0 }

func New(st util.StateDB, preSt util.StateDB) *interest {

	calc, err := matrixstate.GetInterestCalc(preSt)
	if nil != err {
		log.ERROR(PackageName, "获取状态树配置错误")
		return nil
	}

	if calc == util.Stop {
		log.ERROR(PackageName, "停止发放区块奖励", "")
		return nil
	}

	_, err = matrixstate.GetBroadcastInterval(preSt)
	if err != nil {
		log.ERROR(PackageName, "获取广播周期数据结构失败", err)
		return nil
	}
	IC, err := matrixstate.GetInterestCfg(preSt)
	if nil != err {
		log.ERROR(PackageName, "获取利息状态树配置错误", "")
		return nil
	}
	if IC == nil {
		log.ERROR(PackageName, "利息配置", "配置为nil")
		return nil
	}

	if IC.PayInterval == 0 {
		log.ERROR(PackageName, "利息周期配置错误，支付周期", IC.PayInterval)
		return nil
	}

	VipCfg, err := matrixstate.GetVIPConfig(preSt)
	if nil != err {
		log.ERROR(PackageName, "获取VIP状态树配置错误", "")
		return nil
	}
	if 0 == len(VipCfg) {
		log.ERROR(PackageName, "利率表为空", "")
		return nil
	}

	return &interest{VIPConfig: VipCfg, InterestConfig: IC, Calc: calc}
}
func (ic *interest) calcNodeInterest(deposit *big.Int, depositInterestRate []*DepositInterestRate, denominator uint64) *big.Int {

	if deposit.Cmp(big.NewInt(0)) <= 0 {
		log.ERROR(PackageName, "抵押获取错误", deposit)
		return big.NewInt(0)
	}
	var blockInterest uint64
	for i, depositInterest := range depositInterestRate {
		if deposit.Cmp(big.NewInt(0)) <= 0 {
			log.ERROR(PackageName, "抵押获取错误", deposit)
			return big.NewInt(0)
		}
		if deposit.Cmp(depositInterest.Deposit) < 0 {
			blockInterest = depositInterestRate[i-1].Interest
			break
		}
	}
	if blockInterest == 0 {
		blockInterest = depositInterestRate[len(depositInterestRate)-1].Interest
	}
	originResult := new(big.Int).Mul(deposit, new(big.Int).SetUint64(blockInterest))
	finalResult := new(big.Int).Div(originResult, new(big.Int).SetUint64(denominator))
	return finalResult
}

func (ic *interest) calcNodeInterestB(deposit *big.Int) *big.Int {

	if deposit.Cmp(big.NewInt(0)) <= 0 {
		log.ERROR(PackageName, "抵押获取错误", deposit)
		return big.NewInt(0)
	}

	depositMan := new(big.Int).Div(deposit, util.ManPrice).Uint64()
	originResult := math.Pow(float64(depositMan), POWER)
	return new(big.Int).SetUint64(uint64(originResult))
}

func (ic *interest) PayInterest(state vm.StateDBManager, num uint64, time uint64) map[common.Address]*big.Int {
	if !ic.canPayInterest(state, num, ic.InterestConfig.PayInterval) {
		return nil
	}

	//1.获取所有利息转到抵押账户 2.清除所有利息
	log.Debug(PackageName, "发放利息,高度", num)

	AllInterestMap := depoistInfo.GetAllInterest(state)
	allInterest := big.NewInt(0)

	for account, originInterest := range AllInterestMap {
		if originInterest.Cmp(big.NewInt(0)) <= 0 {
			log.ERROR(PackageName, "获取的利息非法", originInterest)
			continue
		}
		slash, _ := depoistInfo.GetSlash(state, account)
		if slash.Cmp(big.NewInt(0)) < 0 {
			log.ERROR(PackageName, "获取的惩罚非法", originInterest)
			continue
		}

		finalInterest := new(big.Int).Sub(originInterest, slash)
		if finalInterest.Cmp(big.NewInt(0)) <= 0 {
			log.ERROR(PackageName, "支付的的利息非法", finalInterest)
			continue
		}
		if slash.Cmp(big.NewInt(0)) > 0 {
			log.Debug(PackageName, "账户", account, "原始利息", originInterest.String(), "惩罚利息", slash.String(), "剩余利息", finalInterest.String())
		}
		AllInterestMap[account] = finalInterest
		allInterest = new(big.Int).Add(allInterest, finalInterest)
		depoistInfo.ResetSlash(state, account)
	}
	balance := state.GetBalance(params.MAN_COIN, common.InterestRewardAddress)
	if balance[common.MainAccount].Balance.Cmp(allInterest) < 0 {
		log.ERROR(PackageName, "利息账户余额不足，余额为", balance[common.MainAccount].Balance.String())
		return nil
	}
	AllInterestMap[common.ContractAddress] = allInterest
	return AllInterestMap
}

func (ic *interest) canPayInterest(state vm.StateDBManager, num uint64, payInterestPeriod uint64) bool {
	latestNum, err := matrixstate.GetInterestPayNum(state)
	if nil != err {
		log.ERROR(PackageName, "状态树获取前一计算利息高度错误", err)
		return false
	}
	if latestNum >= ic.getLastInterestNumber(num-1, payInterestPeriod)+1 {
		//log.Debug(PackageName, "当前周期利息已支付无须再处理", "")
		return false
	}
	matrixstate.SetInterestPayNum(state, num)
	return true
}

func (ic *interest) getLastInterestNumber(number uint64, InterestInterval uint64) uint64 {
	if number%InterestInterval == 0 {
		return number
	}
	ans := (number / InterestInterval) * InterestInterval
	return ans
}
func (ic *interest) GetReward(state vm.StateDBManager, num uint64, parentHash common.Hash) map[common.Address]*big.Int {
	RewardMan := new(big.Int).Mul(new(big.Int).SetUint64(ic.InterestConfig.RewardMount), util.GetPrice(ic.Calc))
	blockReward := util.CalcRewardMountByNumber(state, RewardMan, num-1, ic.InterestConfig.AttenuationPeriod, common.InterestRewardAddress, ic.InterestConfig.AttenuationRate)
	if blockReward.Uint64() == 0 {
		log.Error(PackageName, "账户余额为0，不发放利息奖励", "")
		return nil
	}
	InterestMap := ic.GetInterest(num, parentHash)
	RewardMap := util.CalcInterestReward(blockReward, InterestMap)
	return RewardMap
}

func (ic *interest) CalcReward(state vm.StateDBManager, num uint64, parentHash common.Hash) {
	RewardMap := ic.GetReward(state, num, parentHash)
	ic.SetReward(RewardMap, state)
}

func (ic *interest) SetReward(InterestMap map[common.Address]*big.Int, state vm.StateDBManager) {
	for k, v := range InterestMap {
		depoistInfo.AddInterest(state, k, v)
	}
}

func (ic *interest) GetInterest(num uint64, parentHash common.Hash) map[common.Address]*big.Int {
	depositInterestRateList := make(DepositInterestRateList, 0)
	for _, v := range ic.VIPConfig {
		if v.MinMoney < 0 {
			log.ERROR(PackageName, "最小金额设置非法", "")
			return nil
		}
		deposit := new(big.Int).Mul(new(big.Int).SetUint64(v.MinMoney), util.ManPrice)
		depositInterestRateList = append(depositInterestRateList, &DepositInterestRate{deposit, v.InterestRate})
	}
	//sort.Sort(depositInterestRateList
	depositNodes, err := ca.GetElectedByHeightByHash(parentHash)
	if nil != err {
		log.ERROR(PackageName, "获取的抵押列表错误", err)
		return nil
	}
	if 0 == len(depositNodes) {
		log.ERROR(PackageName, "获取的抵押列表为空", "")
		return nil
	}

	log.Debug(PackageName, "计算利息,高度", num)
	InterestMap := make(map[common.Address]*big.Int)
	if ic.Calc >= util.CalcGamma {
		for _, dv := range depositNodes {

			result := ic.calcNodeInterestB(dv.Deposit)
			if result.Cmp(big.NewInt(0)) <= 0 {
				log.ERROR(PackageName, "计算的利息非法", result)
				continue
			}
			InterestMap[dv.Address] = result
			//log.Debug(PackageName, "账户", dv.Address.String(), "deposit", dv.Deposit.String(), "利息", result.String())
		}
	} else {
		for _, dv := range depositNodes {

			result := ic.calcNodeInterest(dv.Deposit, depositInterestRateList, Denominator)
			if result.Cmp(big.NewInt(0)) <= 0 {
				log.ERROR(PackageName, "计算的利息非法", result)
				continue
			}
			InterestMap[dv.Address] = result
			//log.Debug(PackageName, "账户", dv.Address.String(), "deposit", dv.Deposit.String(), "利息", result.String())
		}
	}

	return InterestMap
}

func (ic *interest) canCalcInterest(state vm.StateDBManager, num uint64, calcInterestInterval uint64) bool {
	latestNum, err := matrixstate.GetInterestCalcNum(state)
	if nil != err {
		log.ERROR(PackageName, "状态树获取前一计算利息高度错误", err)
		return false
	}
	if latestNum >= ic.getLastInterestNumber(num-1, calcInterestInterval)+1 {
		//log.Info(PackageName, "当前利息已计算无须再处理", "")
		return false
	}
	matrixstate.SetInterestCalcNum(state, num)
	return true
}
