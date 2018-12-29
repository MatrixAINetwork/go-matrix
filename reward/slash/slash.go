package slash

import (
	"github.com/matrix/go-matrix/depoistInfo"
	"math/big"

	"github.com/matrix/go-matrix/params/manparams"

	"github.com/matrix/go-matrix/reward/util"

	"github.com/matrix/go-matrix/core/matrixstate"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"

	"github.com/matrix/go-matrix/common"
)

const PackageName = "惩罚"

type BlockSlash struct {
	chain            util.ChainReader
	eleMaxOnlineTime uint64
	SlashRate        uint64
	bcInterval       *manparams.BCInterval
	preElectRoot     common.Hash
	preElectList     []mc.ElectNodeInfo
}

func New(chain util.ChainReader, st util.StateDB) *BlockSlash {
	StateCfg, err := matrixstate.GetDataByState(mc.MSKeySlashCfg, st)
	if nil != err {
		log.ERROR(PackageName, "获取状态树配置错误", "")
		return nil
	}
	SC, ok := StateCfg.(*mc.SlashCfgStruct)
	if !ok {
		log.ERROR(PackageName, "反射失败", "")
		return nil
	}
	if SC.SlashCalc == util.Stop {
		log.ERROR(PackageName, "停止", PackageName)
		return nil
	}

	var SlashRate uint64

	if SC.SlashRate > util.RewardFullRate {
		SlashRate = util.RewardFullRate
	} else {
		SlashRate = SC.SlashRate
	}

	intervalData, err := matrixstate.GetDataByState(mc.MSKeyBroadcastInterval, st)
	if err != nil {
		log.ERROR(PackageName, "获取广播周期失败", err)
		return nil
	}
	bcInterval, err := manparams.NewBCIntervalWithInterval(intervalData)
	if err != nil {
		log.ERROR(PackageName, "创建广播周期数据结构失败", err)
		return nil
	}
	return &BlockSlash{chain: chain, eleMaxOnlineTime: bcInterval.GetBroadcastInterval() - 3, SlashRate: SlashRate, bcInterval: bcInterval} //todo 周期固定3倍关系
}

func (bp *BlockSlash) CalcSlash(currentState *state.StateDB, num uint64, upTimeMap map[common.Address]uint64, interestCalcMap map[common.Address]*big.Int) {
	var eleNum uint64

	if bp.bcInterval.IsBroadcastNumber(num) {
		log.WARN(PackageName, "广播周期不处理", "")
		return
	}
	//选举周期的开始分配
	latestNum, err := matrixstate.GetNumByState(mc.MSKeySlashNum, currentState)
	if nil != err {
		log.ERROR(PackageName, "状态树获取前一发放惩罚高度错误", err)
		return
	}
	if latestNum > bp.bcInterval.GetLastBroadcastNumber() {
		log.Debug(PackageName, "当前惩罚已处理无须再处理", "")
		return
	}

	matrixstate.SetNumByState(mc.MSKeySlashNum, currentState, num)

	if 0 == len(interestCalcMap) {
		log.WARN(PackageName, "获取到利息为空", "")
		return
	}
	if 0 == len(upTimeMap) {
		log.WARN(PackageName, "获取到uptime为空", "")
		return
	}
	//计算选举的拓扑图的高度
	if num < bp.bcInterval.GetReElectionInterval()+2 {
		eleNum = 1
	} else {
		// 下一个选举+1
		eleNum = num - bp.bcInterval.GetBroadcastInterval()
	}

	electGraph, err := bp.chain.GetMatrixStateDataByNumber(mc.MSKeyElectGraph, eleNum)
	if err != nil {
		log.Error(PackageName, "获取拓扑图错误", err)
		return
	}
	if electGraph == nil {
		log.Error(PackageName, "获取拓扑图反射错误")
		return
	}
	originElectNodes := electGraph.(*mc.ElectGraph)
	if 0 == len(originElectNodes.ElectList) {
		log.Error(PackageName, "get获取初选列表为空", "")
		return
	}

	for _, v := range originElectNodes.ElectList {

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
			log.Debug(PackageName, "惩罚账户", v.Account, "惩罚金额", slash)
			depoistInfo.AddSlash(currentState, v.Account, slash)
		}

	}
}

func (bp *BlockSlash) getSlash(upTime uint64, accountReward *big.Int) *big.Int {
	rate := 1 - float64(upTime)/float64(bp.eleMaxOnlineTime)
	maxRate := float64(bp.SlashRate) / float64(util.RewardFullRate)
	if rate >= maxRate {
		rate = maxRate
	}
	slash := new(big.Int).SetUint64(uint64(float64(accountReward.Uint64()) * rate))
	return slash
}
