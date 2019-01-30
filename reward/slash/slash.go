package slash

import (
	"math/big"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/matrixstate"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/depoistInfo"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/reward/util"
)

const PackageName = "惩罚"

type BlockSlash struct {
	chain            util.ChainReader
	eleMaxOnlineTime uint64
	SlashRate        uint64
	bcInterval       *mc.BCIntervalInfo
	preElectRoot     common.Hash
	preElectList     []mc.ElectNodeInfo
}

func New(chain util.ChainReader, st util.StateDB) *BlockSlash {

	data, err := matrixstate.GetSlashCalc(st)
	if nil != err {
		log.ERROR(PackageName, "获取状态树配置错误")
		return nil
	}

	if data == util.Stop {
		log.ERROR(PackageName, "停止发放区块奖励", "")
		return nil
	}

	SC, err := matrixstate.GetSlashCfg(st)
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

	bcInterval, err := matrixstate.GetBroadcastInterval(st)
	if err != nil {
		log.ERROR(PackageName, "获取广播周期数据结构失败", err)
		return nil
	}
	return &BlockSlash{chain: chain, eleMaxOnlineTime: bcInterval.GetBroadcastInterval() - 3, SlashRate: SlashRate, bcInterval: bcInterval} //todo 周期固定3倍关系
}

func (bp *BlockSlash) CalcSlash(currentState *state.StateDB, num uint64, upTimeMap map[common.Address]uint64, interestCalcMap map[common.Address]*big.Int) {

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
	st, err := bp.chain.StateAtNumber(eleNum)
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
