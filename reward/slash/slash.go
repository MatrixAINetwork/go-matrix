// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package slash

import (
	"math/big"

	"github.com/MatrixAINetwork/go-matrix/mc"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/common/readstatedb"
	"github.com/MatrixAINetwork/go-matrix/core/matrixstate"
	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/core/vm"
	"github.com/MatrixAINetwork/go-matrix/depoistInfo"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/reward/util"
)

const PackageName = "惩罚"

type BlockSlash struct {
	chain            util.ChainReader
	eleMaxOnlineTime uint64
	SlashRate        uint64
	bcInterval       *mc.BCIntervalInfo
	preBroadcastRoot *mc.PreBroadStateRoot
}

func New(chain util.ChainReader, st util.StateDB, preSt util.StateDB) *BlockSlash {

	data, err := matrixstate.GetSlashCalc(preSt)
	if nil != err {
		log.ERROR(PackageName, "获取状态树配置错误")
		return nil
	}

	if data == util.Stop {
		log.ERROR(PackageName, "停止发放区块奖励", "")
		return nil
	}

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
	return &BlockSlash{chain: chain, eleMaxOnlineTime: bcInterval.GetBroadcastInterval() - 3, SlashRate: SlashRate, bcInterval: bcInterval} // 周期固定3倍关系
}
func (bp *BlockSlash) GetCurrentInterest(preState *state.StateDBManage, currentState vm.StateDBManager, num uint64) map[common.Address]*big.Int {
	allInterest := depoistInfo.GetAllInterest(currentState)
	interestMap := make(map[common.Address]*big.Int)
	latestNum, err := matrixstate.GetInterestPayNum(currentState)
	if nil != err {
		log.ERROR(PackageName, "状态树获取前一计算利息高度错误", err)
		return nil
	}
	//前一个广播周期支付利息，利息会清空，直接用当前值,其它时间点用差值
	if num-bp.bcInterval.BCInterval <= latestNum && num > latestNum {
		interestMap = allInterest
	} else {
		for account, currentInterest := range allInterest {
			preInterest, _ := depoistInfo.GetInterest(preState, account)
			interestMap[account] = new(big.Int).Sub(currentInterest, preInterest)
		}
	}
	//for account, interest := range interestMap {
	//	log.Debug(PackageName, "账户", account, "利息", interest)
	//}

	return interestMap
}
func (bp *BlockSlash) CalcSlash(currentState *state.StateDBManage, num uint64, upTimeMap map[common.Address]uint64, parentHash common.Hash, time uint64) {
	if bp.bcInterval.IsBroadcastNumber(num) {
		log.WARN(PackageName, "广播周期不处理", "")
		return
	}
	//选举周期的开始分配
	latestNum, err := matrixstate.GetSlashNum(currentState)
	if nil != err {
		log.ERROR(PackageName, "状态树获取前一发放惩罚高度错误", err)
		return
	}
	if latestNum > bp.bcInterval.GetLastBroadcastNumber() {
		//log.Debug(PackageName, "当前惩罚已处理无须再处理", "")
		return
	}

	if err := matrixstate.SetSlashNum(currentState, num); err != nil {
		log.Error(PackageName, "设置惩罚状态失败", err)
	}
	preBroadcastRoot, err := readstatedb.GetPreBroadcastRoot(bp.chain, parentHash)
	if err != nil {
		log.Error(PackageName, "获取之前广播区块的root值失败 err", err)
		return
	}
	beforeLastStateRoot := preBroadcastRoot.LastStateRoot
	preState, err := bp.chain.StateAt(beforeLastStateRoot)
	if err != nil {
		log.Error("GetBroadcastTxMap StateAt err")
		return
	}
	interestCalcMap := bp.GetCurrentInterest(preState, currentState, num)
	if 0 == len(interestCalcMap) {
		log.WARN(PackageName, "获取到利息为空", "")
		return
	}
	if 0 == len(upTimeMap) {
		log.WARN(PackageName, "获取到uptime为空", "")
		return
	}
	//计算选举的拓扑图的高度
	eleNum := bp.bcInterval.GetLastBroadcastNumber() - 2
	ancetorHash, err := bp.chain.GetAncestorHash(parentHash, eleNum)
	if err != nil {
		log.Error(PackageName, "获取制动高度hash失败", err, "eleNum", eleNum)
		return
	}
	st, err := bp.chain.StateAtBlockHash(ancetorHash)
	if err != nil {
		log.Error(PackageName, "获取选举高度的状态树失败", err, "eleNum", eleNum)
		return
	}
	electGraph, err := matrixstate.GetElectGraph(st)
	if err != nil {
		log.Error(PackageName, "获取拓扑图错误", err)
		return
	}
	if electGraph == nil {
		log.Error(PackageName, "获取拓扑图错误", "is nil")
		return
	}
	if 0 == len(electGraph.ElectList) {
		log.Error(PackageName, "get获取初选列表为空", "")
		return
	}

	for _, v := range electGraph.ElectList {
		if v.Type == common.RoleValidator || v.Type == common.RoleBackupValidator {
			interest, ok := interestCalcMap[v.Account]
			if !ok {
				log.WARN(PackageName, "无法获取利息，账户", v.Account)
				continue
			}
			if interest.Cmp(new(big.Int).SetUint64(0)) <= 0 {
				log.WARN(PackageName, "获取利息非法，账户", v.Account)
				continue
			}

			upTime, ok := upTimeMap[v.Account]
			if !ok {
				log.WARN(PackageName, "获取uptime错误，账户", v.Account)
				continue
			}

			slash := bp.getSlash(upTime, interest)
			if slash.Cmp(big.NewInt(0)) < 0 {
				log.ERROR(PackageName, "惩罚比例为负数", "")
				continue
			}
			if slash.Cmp(big.NewInt(0)) > 0 {
				log.Debug(PackageName, "惩罚账户", v.Account, "惩罚金额", slash)
			}
			depoistInfo.AddSlash(currentState, v.Account, slash)
		}

	}
}

func (bp *BlockSlash) getSlash(upTime uint64, accountReward *big.Int) *big.Int {
	rate := uint64((bp.eleMaxOnlineTime - upTime) * util.RewardFullRate / (bp.eleMaxOnlineTime))

	if rate >= bp.SlashRate {
		rate = bp.SlashRate
	}
	tmp := new(big.Int).Mul(accountReward, new(big.Int).SetUint64(rate))

	slash := new(big.Int).Div(tmp, new(big.Int).SetUint64(util.RewardFullRate))
	return slash
}
