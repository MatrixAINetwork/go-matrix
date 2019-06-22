// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package interest

import (
	"math/big"
	"os"
	"strconv"

	"github.com/MatrixAINetwork/go-matrix/reward/depositcfg"

	"github.com/pkg/errors"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/matrixstate"
	"github.com/MatrixAINetwork/go-matrix/core/vm"
	"github.com/MatrixAINetwork/go-matrix/depoistInfo"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/params"
	"github.com/MatrixAINetwork/go-matrix/reward/util"
)

type interestDelta struct {
	interestConfig *mc.InterestCfg
	depositCfg     depositcfg.DepositCfgInterface
}

const INTERESTDIR = "./interestdir"

func init() {
	_, e := os.Stat(INTERESTDIR)
	if e != nil {
		os.Mkdir(INTERESTDIR, os.ModePerm)
	}
}

func DeltaNew(st util.StateDB, preSt util.StateDB, depositCfgVersion string) *interestDelta {

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

	return &interestDelta{interestConfig: IC, depositCfg: depositcfg.GetDepositCfg(depositCfgVersion)}
}

func (ic *interestDelta) calcWeightDeposit(deposit *big.Int, blockInterest uint64) *big.Int {

	if deposit.Cmp(big.NewInt(0)) <= 0 {
		log.ERROR(PackageName, "抵押获取错误", deposit)
		return big.NewInt(0)
	}

	originResult := new(big.Int).Mul(deposit, new(big.Int).SetUint64(blockInterest))
	return originResult
}

func (ic *interestDelta) PayInterest(state vm.StateDBManager, num uint64, time uint64) map[common.Address]*big.Int {
	if !ic.canPayInterest(state, num, ic.interestConfig.PayInterval) {
		return nil
	}

	snapshot := state.Snapshot(params.MAN_COIN)
	if err := ic.payAllInterest(num, state, time); nil != err {
		state.RevertToSnapshot(params.MAN_COIN, snapshot)
	}

	return nil
}

func (ic *interestDelta) payAllInterest(num uint64, state vm.StateDBManager, time uint64) error {
	//1.获取所有利息转到抵押账户 2.清除所有利息
	log.Debug(PackageName, "发放利息,高度", num)
	AllInterestMap := depoistInfo.GetAllInterest_v2(state, time)
	allInterest := big.NewInt(0)
	outputPayInterest := make(map[common.Address][]common.OperationalInterestSlash, 0)
	outputSlash := make(map[common.Address][]common.OperationalInterestSlash, 0)
	for account, originAccountInterest := range AllInterestMap {
		accountSlash, _ := depoistInfo.GetSlash_v2(state, account)
		finalInterestData := make([]common.OperationalInterestSlash, 0)
		for _, originInterest := range originAccountInterest.CalcDeposit {
			positionSlash := util.GetDataByPosition(accountSlash.CalcDeposit, originInterest.Position)
			if nil == positionSlash || positionSlash.OperAmount.Cmp(big.NewInt(0)) < 0 {
				log.Warn(PackageName, "获取惩罚仓位已退选,账户", account.String(), "仓位", originInterest.Position)
				continue
			}

			if originInterest.OperAmount.Cmp(big.NewInt(0)) <= 0 {
				log.ERROR(PackageName, "获取的利息非法", originInterest, "账户", account.String())
				continue
			}
			finalInterest := originInterest
			finalInterest.OperAmount = new(big.Int).Sub(originInterest.OperAmount, positionSlash.OperAmount)
			if finalInterest.OperAmount.Cmp(big.NewInt(0)) <= 0 {
				log.ERROR(PackageName, "支付的的利息非法", finalInterest, "账户", account.String())
				continue
			}
			if positionSlash.OperAmount.Cmp(big.NewInt(0)) > 0 {
				log.Trace(PackageName, "账户", account, "仓位", originInterest.Position, "原始利息", originInterest.OperAmount.String(), "惩罚利息", positionSlash.OperAmount.String(), "剩余利息", finalInterest.OperAmount.String())
			}
			balance := state.GetBalance(params.MAN_COIN, common.InterestRewardAddress)
			allInterest = new(big.Int).Add(allInterest, finalInterest.OperAmount)
			if balance[common.MainAccount].Balance.Cmp(allInterest) < 0 {
				log.ERROR(PackageName, "利息账户余额不足，余额为", balance[common.MainAccount].Balance.String())

				return errors.New("余额不足")
			}
			depoistInfo.PayInterest(state, time, account, originInterest.Position, finalInterest.OperAmount)
			finalInterestData = append(finalInterestData, finalInterest)
		}
		outputPayInterest[account] = finalInterestData
		outputSlash[account] = accountSlash.CalcDeposit

	}
	util.PrintLog2File(INTERESTDIR+"/payinterest_"+strconv.FormatUint(num, 10)+".json", outputPayInterest)
	util.PrintLog2File(INTERESTDIR+"/slash_"+strconv.FormatUint(num, 10)+".json", outputSlash)
	state.SubBalance(params.MAN_COIN, common.MainAccount, common.InterestRewardAddress, allInterest)
	state.AddBalance(params.MAN_COIN, common.MainAccount, common.ContractAddress, allInterest)
	return nil
}

func (ic *interestDelta) canPayInterest(state vm.StateDBManager, num uint64, payInterestPeriod uint64) bool {
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

func (ic *interestDelta) getLastInterestNumber(number uint64, InterestInterval uint64) uint64 {
	if number%InterestInterval == 0 {
		return number
	}
	ans := (number / InterestInterval) * InterestInterval
	return ans
}
func (ic *interestDelta) GetReward(blockReward *big.Int, depositNodes []common.DepositBase) map[common.Address][]common.OperationalInterestSlash {

	weightDepositMap, totalWeightDeposit := ic.GetWeightDeposit(depositNodes)
	RewardMap := ic.CalcInterestReward(totalWeightDeposit, blockReward, weightDepositMap)
	return RewardMap
}

func (ic *interestDelta) CalcInterestReward(totalWeightDeposit, reward *big.Int, weightDepositMap map[common.Address][]common.OperationalInterestSlash) map[common.Address][]common.OperationalInterestSlash {

	if 0 == len(weightDepositMap) {
		log.ERROR(PackageName, "利息列表为空", "")
		return nil
	}

	if totalWeightDeposit.Cmp(big.NewInt(0)) <= 0 {
		log.ERROR(PackageName, "计算的总利息值非法", totalWeightDeposit)
		return nil
	}
	log.Trace(PackageName, "计算的总抵押值", totalWeightDeposit)

	if 0 == reward.Cmp(big.NewInt(0)) {
		log.ERROR(PackageName, "定点化奖励金额为0", "")
		return nil
	}

	rewards := make(map[common.Address][]common.OperationalInterestSlash)
	for k, node := range weightDepositMap {
		nodeReward := make([]common.OperationalInterestSlash, 0)
		for _, v := range node {
			temp := new(big.Int).Mul(reward, v.DepositAmount)
			blockAmount := new(big.Int).Div(temp, totalWeightDeposit)
			v.OperAmount.Add(v.OperAmount, blockAmount)
			nodeReward = append(nodeReward, v)
			//util.LogExtraDebug(PackageName, "账户", k.String(), "仓位", v.Position, "加权抵押结果", v.DepositAmount.String(), "奖励金额", blockAmount.String(), "累计的抵押奖励金额", v.OperAmount)
		}
		rewards[k] = nodeReward

	}
	return rewards
}

func (ic *interestDelta) CalcReward(state vm.StateDBManager, num uint64, parentHash common.Hash) {
	RewardMan := new(big.Int).Mul(new(big.Int).SetUint64(ic.interestConfig.RewardMount), util.ThousandthManPrice)
	blockReward := util.CalcRewardMountByNumber(state, RewardMan, num-1, ic.interestConfig.AttenuationPeriod, common.InterestRewardAddress, ic.interestConfig.AttenuationRate)
	if blockReward.Uint64() == 0 {
		log.Error(PackageName, "账户余额为0，不发放利息奖励", "")
		return
	}
	log.Debug(PackageName, "计算加权抵押,高度", num)
	depositNodes := ic.GetDeposit(parentHash)
	util.PrintLog2File(INTERESTDIR+"/deposit_"+strconv.FormatUint(num, 10)+".json", depositNodes)
	RewardMap := ic.GetReward(blockReward, depositNodes)
	util.PrintLog2File(INTERESTDIR+"/interest_"+strconv.FormatUint(num, 10)+".json", RewardMap)
	ic.SetReward(RewardMap, state)
	return
}

func (ic *interestDelta) SetReward(InterestMap map[common.Address][]common.OperationalInterestSlash, state vm.StateDBManager) {

	for k, v := range InterestMap {
		//log.Debug(PackageName, "账户", k, "数据", v)
		depoistInfo.AddInterest_v2(state, k, common.CalculateDeposit{AddressA0: k, CalcDeposit: v})
	}
}

func (ic *interestDelta) GetDeposit(parentHash common.Hash) []common.DepositBase {
	depositNodes, err := depoistInfo.GetAllDepositByHash_v2(parentHash)
	if nil != err {
		log.ERROR(PackageName, "获取的抵押列表错误", err)
		return nil
	}
	log.Debug(PackageName, "抵押账户数目", len(depositNodes), "hash", parentHash.String())
	return depositNodes
}
func (ic *interestDelta) GetWeightDeposit(depositNodes []common.DepositBase) (map[common.Address][]common.OperationalInterestSlash, *big.Int) {

	InterestMap := make(map[common.Address][]common.OperationalInterestSlash)
	totalDeposit := new(big.Int)
	for _, node := range depositNodes {
		nodeSlice := make([]common.OperationalInterestSlash, len(node.Dpstmsg))
		for i, dv := range node.Dpstmsg {
			cfg := ic.depositCfg.GetDepositPositionCfg(dv.DepositType)
			if cfg == nil {
				log.ERROR(PackageName, "获取利率配置错误，抵押类型为", dv.DepositType)
				continue
			}
			rate := cfg.GetRate()
			if 0 == rate {
				log.ERROR(PackageName, "获取利率错误，抵押类型为", dv.DepositType)
				continue
			}
			weightDeposit := ic.calcWeightDeposit(dv.DepositAmount, rate)
			if weightDeposit.Cmp(big.NewInt(0)) <= 0 && dv.Position != 0 {
				log.ERROR(PackageName, "计算的利息非法", weightDeposit)
				continue
			}
			//util.LogExtraDebug(PackageName, "账户", node.AddressA0.String(), "仓位", dv.Position, "原始抵押", dv.DepositAmount, "抵押类型", dv.DepositType, "利率", rate)
			totalDeposit.Add(totalDeposit, weightDeposit)
			nodeSlice[i] = common.OperationalInterestSlash{DepositAmount: weightDeposit, Position: dv.Position, OperAmount: dv.Interest, DepositType: dv.DepositType}
		}
		InterestMap[node.AddressA0] = nodeSlice

	}

	return InterestMap, totalDeposit
}

func (ic *interestDelta) canCalcInterest(state vm.StateDBManager, num uint64, calcInterestInterval uint64) bool {
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
