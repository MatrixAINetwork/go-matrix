// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package slash

import (
	"errors"
	"math/big"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/common/readstatedb"
	"github.com/MatrixAINetwork/go-matrix/core/matrixstate"
	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/core/vm"
	"github.com/MatrixAINetwork/go-matrix/depoistInfo"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/reward/util"
)

type SlashDelta struct {
	chain            util.ChainReader
	eleMaxOnlineTime uint64
	SlashRate        uint64
	bcInterval       *mc.BCIntervalInfo
	preBroadcastRoot *mc.PreBroadStateRoot
	preSt            *state.StateDBManage
}

func DeltaNew(chain util.ChainReader, st util.StateDB, preSt *state.StateDBManage) *SlashDelta {

	SC, err := matrixstate.GetSlashCfg(preSt)
	if nil != err || nil == SC {
		log.ERROR(PackageName, "获取状态树配置错误", "")
		return nil
	}
	var SlashRate uint64

	if SC.SlashRate > util.RewardFullRate {
		SlashRate = util.RewardFullRate
	} else {
		SlashRate = SC.SlashRate
	}

	bcInterval, err := matrixstate.GetBroadcastInterval(preSt)
	if err != nil {
		log.ERROR(PackageName, "获取广播周期数据结构失败", err)
		return nil
	}
	return &SlashDelta{chain: chain, eleMaxOnlineTime: bcInterval.GetBroadcastInterval() - 3, SlashRate: SlashRate, bcInterval: bcInterval, preSt: preSt} // 周期固定3倍关系
}

func (bp *SlashDelta) compareVersion(preState *state.StateDBManage, currentState vm.StateDBManager) bool {
	preVersion, err := matrixstate.GetSlashCalc(preState)
	if nil != err {
		log.Crit(PackageName, "获取版本1错误", "")
		return false
	}
	curVersion, err := matrixstate.GetSlashCalc(currentState)
	if nil != err {
		log.Crit(PackageName, "获取版本2错误", "")
		return false
	}
	if preVersion != curVersion {
		log.Info(PackageName, "版本切换，前一个版本", preVersion, "当前版本", curVersion)
		return true
	}
	return false

}
func (bp *SlashDelta) getCurrentInterest(preBcState *state.StateDBManage, currentState vm.StateDBManager, num uint64, time uint64) map[common.Address][]common.OperationalInterestSlash {

	allAccountInterest := depoistInfo.GetAllInterest_v2(currentState, time)
	allInterest := make(map[common.Address][]common.OperationalInterestSlash)
	for account, accountInterest := range allAccountInterest {

		allInterest[account] = accountInterest.CalcDeposit
	}

	latestNum, err := matrixstate.GetInterestPayNum(currentState)
	if nil != err {
		log.ERROR(PackageName, "状态树获取前一计算利息高度错误", err)
		return nil
	}
	//前一个广播周期支付利息，利息会清空，直接用当前值,其它时间点用差值
	if num-bp.bcInterval.BCInterval <= latestNum && num > latestNum {
		log.Trace(PackageName, "支付利息的下一个广播周期，使用当前累加利息", "")
		return allInterest
	} else if bp.compareVersion(preBcState, currentState) {
		return allInterest
	} else {
		allBcInterest := make(map[common.Address][]common.OperationalInterestSlash)
		preAccountInterestMap := depoistInfo.GetAllInterest_v2(preBcState, 0)
		for account, currentInterestData := range allInterest {

			if preAccountInterestData, exit := preAccountInterestMap[account]; exit {
				bcInterestData := make([]common.OperationalInterestSlash, 0)
				for _, interest := range currentInterestData {
					bcGenInterest := interest.OperAmount
					if interest.Position == 0 && 0 == bcGenInterest.Cmp(big.NewInt(0)) {
						log.Trace(PackageName, "活期账户已退选", bcGenInterest, "仓位", interest.Position, "利息", bcGenInterest)
						continue
					}
					for _, preInterest := range preAccountInterestData.CalcDeposit {

						if interest.Position == preInterest.Position {
							bcGenInterest = new(big.Int).Sub(interest.OperAmount, preInterest.OperAmount)

						}
					}
					if bcGenInterest.Cmp(new(big.Int).SetUint64(0)) < 0 {
						log.Error(PackageName, "利息为负值", bcGenInterest, "仓位", interest.Position, "利息", bcGenInterest)
						continue
					}
					//util.LogExtraDebug(PackageName, "账户", account, "仓位", interest.Position, "利息", bcGenInterest)
					bcInterestData = append(bcInterestData, common.OperationalInterestSlash{Position: interest.Position, OperAmount: bcGenInterest})
				}
				allBcInterest[account] = bcInterestData
			} else {
				//util.LogExtraDebug(PackageName, "账户前一周期无累加利息", account)
				allBcInterest[account] = currentInterestData
			}

		}
		return allBcInterest
	}
	return allInterest
}
func (bp *SlashDelta) GetDeposit(parentHash common.Hash) []common.DepositBase {
	depositNodes, err := depoistInfo.GetAllDepositByHash_v2(parentHash)
	if nil != err {
		log.ERROR(PackageName, "获取的抵押列表错误", err)
		return nil
	}
	return depositNodes
}
func (bp *SlashDelta) CalcSlash(currentState *state.StateDBManage, num uint64, upTimeMap map[common.Address]uint64, parentHash common.Hash, time uint64) {
	if bp.bcInterval.IsBroadcastNumber(num) {
		log.WARN(PackageName, "广播周期不处理", "")
		return
	}
	//选举周期的开始分配
	if !bp.canSlash(currentState) {
		return
	}

	interestCalcMap, electGraph, err := bp.GetElectAndInterest(currentState, num, parentHash, upTimeMap, time)
	if nil != err {
		return
	}

	bp.SetSlash(electGraph, upTimeMap, interestCalcMap, currentState)
}

func (bp *SlashDelta) SetSlash(electGraph *mc.ElectGraph, upTimeMap map[common.Address]uint64, allAccountInterest map[common.Address][]common.OperationalInterestSlash, currentState *state.StateDBManage) {
	for _, v := range electGraph.ElectList {
		if v.Type == common.RoleValidator || v.Type == common.RoleBackupValidator {

			upTime, ok := upTimeMap[v.Account]
			if !ok {
				log.WARN(PackageName, "获取uptime错误，账户", v.Account)
				continue
			}
			bcInterest, ok := allAccountInterest[v.Account]
			if !ok {
				log.WARN(PackageName, "账户退选", v.Account)
				continue
			}
			slashRate := bp.getSlashRate(upTime)

			bp.addSlash(v.Account, bcInterest, currentState, slashRate)

		}

	}
}

func (bp *SlashDelta) addSlash(account common.Address, accountInterest []common.OperationalInterestSlash, currentState *state.StateDBManage, rate uint64) {

	accountSlash, _ := depoistInfo.GetSlash_v2(currentState, account)
	newSlashData := make([]common.OperationalInterestSlash, 0)
	for _, bcInterest := range accountInterest {
		slash := bp.getSlash(rate, bcInterest.OperAmount)
		for _, slashData := range accountSlash.CalcDeposit {
			if bcInterest.Position == slashData.Position {
				slash = slashData.OperAmount.Add(slashData.OperAmount, slash)
			}
		}
		if slash.Cmp(big.NewInt(0)) > 0 {
			log.Trace(PackageName, "账户", account, "仓位", bcInterest.Position, "累加惩罚金额", slash)
		}
		newSlashData = append(newSlashData, common.OperationalInterestSlash{Position: bcInterest.Position, OperAmount: slash})
	}
	accountSlash.CalcDeposit = newSlashData
	depoistInfo.AddSlash_v2(currentState, account, accountSlash)
}

func (bp *SlashDelta) GetElectAndInterest(currentState *state.StateDBManage, num uint64, parentHash common.Hash, upTimeMap map[common.Address]uint64, time uint64) (map[common.Address][]common.OperationalInterestSlash, *mc.ElectGraph, error) {
	if err := matrixstate.SetSlashNum(currentState, num); err != nil {
		log.Error(PackageName, "设置惩罚状态失败", err)
		return nil, nil, err
	}
	preBroadcastRoot, err := readstatedb.GetPreBroadcastRoot(bp.chain, parentHash)
	if err != nil {
		log.Error(PackageName, "获取之前广播区块的root值失败 err", err)
		return nil, nil, err
	}
	beforeLastStateRoot := preBroadcastRoot.LastStateRoot
	preBCState, err := bp.chain.StateAt(beforeLastStateRoot)
	if err != nil {
		log.Error("GetBroadcastTxMap StateAt err")
		return nil, nil, err
	}
	interestCalcMap := bp.getCurrentInterest(preBCState, bp.preSt, num, time)
	if 0 == len(interestCalcMap) {
		log.Error(PackageName, "获取到利息为空", "")
		return nil, nil, errors.New("获取到利息为空")
	}
	if 0 == len(upTimeMap) {
		log.WARN(PackageName, "获取到uptime为空", "")
		return nil, nil, errors.New("获取到uptime为空")
	}
	//计算选举的拓扑图的高度
	eleNum := bp.bcInterval.GetLastBroadcastNumber() - 2
	ancetorHash, err := bp.chain.GetAncestorHash(parentHash, eleNum)
	if err != nil {
		log.Error(PackageName, "获取选举高度hash失败", err, "eleNum", eleNum)
		return nil, nil, err
	}
	st, err := bp.chain.StateAtBlockHash(ancetorHash)
	if err != nil {
		log.Error(PackageName, "获取选举高度的状态树失败", err, "eleNum", eleNum)
		return nil, nil, err
	}
	electGraph, err := matrixstate.GetElectGraph(st)
	if err != nil {
		log.Error(PackageName, "获取拓扑图错误", err)
		return nil, nil, err
	}
	if electGraph == nil {
		log.Error(PackageName, "获取拓扑图错误", "is nil")
		return nil, nil, errors.New("获取拓扑图错误")
	}
	if 0 == len(electGraph.ElectList) {
		log.Error(PackageName, "get获取初选列表为空", "")
		return nil, nil, errors.New("get获取初选列表为空")
	}
	return interestCalcMap, electGraph, nil
}

func (bp *SlashDelta) canSlash(currentState *state.StateDBManage) bool {
	latestNum, err := matrixstate.GetSlashNum(currentState)
	if nil != err {
		log.ERROR(PackageName, "状态树获取前一发放惩罚高度错误", err)
		return false
	}
	if latestNum > bp.bcInterval.GetLastBroadcastNumber() {
		//log.Debug(PackageName, "当前惩罚已处理无须再处理", "")
		return false
	}
	return true
}

func (bp *SlashDelta) getSlashRate(upTime uint64) uint64 {
	rate := uint64((bp.eleMaxOnlineTime - upTime) * util.RewardFullRate / (bp.eleMaxOnlineTime))

	if rate >= bp.SlashRate {
		rate = bp.SlashRate
	}

	return rate
}

func (bp *SlashDelta) getSlash(rate uint64, accountReward *big.Int) *big.Int {

	tmp := new(big.Int).Mul(accountReward, new(big.Int).SetUint64(rate))

	slash := new(big.Int).Div(tmp, new(big.Int).SetUint64(util.RewardFullRate))
	return slash
}
