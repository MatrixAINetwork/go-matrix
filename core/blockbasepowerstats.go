// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package core

import (
	"github.com/MatrixAINetwork/go-matrix/ca"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/matrixstate"
	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/params/manparams"
	"github.com/MatrixAINetwork/go-matrix/params/manversion"
	"github.com/pkg/errors"
)

func (bc *BlockChain) basePowerSlashCfgProc(state *state.StateDBManage, num uint64) (*mc.BasePowerSlashCfg, bool, error) {
	if nil == state {
		return nil, false, ErrStatePtrIsNil
	}
	slashCfg, err := bc.BasePowerGetSlashCfg(state)
	if nil != err {
		log.Crit(ModuleName, "Get BasePower Slash Cfg", ErrGetSlashCfgStateErr)
		return nil, false, ErrGetSlashCfgStateErr
	} else {
		return slashCfg, true, nil
	}
}

func (bc *BlockChain) BasePowerGProduceSlash(version string, state *state.StateDBManage, header *types.Header) error {
	if manversion.VersionCmp(version, manversion.VersionAIMine) < 0 {
		return nil
	}

	if nil == state {
		return ErrStatePtrIsNil
	}
	if nil == header {
		return ErrHeaderPtrIsNil
	}

	bcInterval, err := matrixstate.GetBroadcastInterval(state)
	if err != nil {
		return err
	}

	parentHeader := bc.GetHeaderByHash(header.ParentHash)
	if nil == parentHeader {
		return errors.Errorf("获取区块错误，区块hash=%v", header.ParentHash)
	}

	if bcInterval.IsBroadcastNumber(parentHeader.Number.Uint64()) {
		log.Debug("BasePowerGProduceSlash", "区块是广播区块", "跳过处理")
		return nil
	}

	slashCfg, _, err := bc.basePowerSlashCfgProc(state, parentHeader.Number.Uint64())
	if nil == slashCfg {
		return err
	}

	if status, err := bc.shouldBasePowerStatsStart(state, parentHeader.ParentHash, slashCfg); status {
		log.Trace(ModuleName, "执行初始化统计列表,高度", parentHeader.Number.Uint64(), "Err", err)
		BasePowerinitStatsList(state, parentHeader.Number.Uint64())
	}

	if statsList, ok := bc.BasePowerShouldAddRecorder(state, slashCfg); ok {
		log.Trace(ModuleName, "增加出块统计,高度", parentHeader.Number.Uint64(), "账户", parentHeader.Coinbase.String(), "高度", parentHeader.Number.Uint64())
		BasePowerStatsListAddRecorder(state, statsList, parentHeader.BasePowers)
		basePowerstatsListPrint(statsList)
		if ok := BasePowerShouldBasePowerSlash(state, parentHeader, slashCfg); ok {

			bc.basePowerStatsListToBlackList(state, statsList, slashCfg, parentHeader)

		}
	}

	return nil
}

func (bc *BlockChain) basePowerStatsListToBlackList(state *state.StateDBManage, statsList *mc.BasePowerStats, slashCfg *mc.BasePowerSlashCfg, header *types.Header) {
	preBlackList := bc.BasePowerGetBlackList(state)
	basePowerBlackListPrint(preBlackList)
	//新增退选后从黑名单移除
	var handleBlackList = BasePowerNewBlackListMaintain(header.ParentHash, preBlackList.BlackList)
	handleBlackList.BasePowerAddBlackList(statsList.StatsList, slashCfg)
	log.Trace(ModuleName, "黑名单更新后状态，高度", header.Number.Uint64())
	basePowerBlackListPrint(&mc.BasePowerSlashBlackList{BlackList: handleBlackList.blacklist})
	if err := matrixstate.SetBasePowerBlackList(state, &mc.BasePowerSlashBlackList{BlackList: handleBlackList.blacklist}); err != nil {
		log.Crit(ModuleName, "State Write Err : ", err)
	}
}
func basePowerstatsListPrint(stats *mc.BasePowerStats) {
	for _, v := range stats.StatsList {
		log.Debug(ModuleName, "Address", v.Address.String(), "Produce Block Num", v.ProduceNum)
	}
}
func basePowerBlackListPrint(blackList *mc.BasePowerSlashBlackList) {
	for _, v := range blackList.BlackList {
		log.Debug(ModuleName, "Address", v.Address.String(), "Ban", v.ProhibitCycleCounter)
	}
}

/*确定是否执行统计初始化：
配置不存在，或惩罚关闭，不执行
*/
func (bc *BlockChain) shouldBasePowerStatsStart(currentState *state.StateDBManage, parentHash common.Hash, slashCfg *mc.BasePowerSlashCfg) (bool, error) {
	if nil == currentState {
		return false, ErrStatePtrIsNil
	}
	if nil == slashCfg {
		return false, ErrSlashCfgPtrIsNil
	}
	//如果惩罚关闭，不执行
	if !slashCfg.Switcher {
		return false, nil
	}

	//如果初始化高度为空，需要执行
	latestUpdateTime, err := basePowerGetLatestInitStatsNum(currentState)
	if 0 == latestUpdateTime {
		return true, nil
	}
	//如果初始化高度是上一个周期，需要执行
	hasInit, err := basePowerHasStatsInit(parentHash, latestUpdateTime)
	if nil != err {
		//状态数读取错误，不执行初始化。不会照成错误的黑名单
		return false, err
	} else {
		return !hasInit, nil
	}

}

func (bc *BlockChain) BasePowerGetSlashCfg(state *state.StateDBManage) (*mc.BasePowerSlashCfg, error) {
	slashCfg, err := matrixstate.GetBasePowerSlashCfg(state)
	if err != nil {
		log.Error(ModuleName, "获取区块生产惩罚配置失败", err)
		return &mc.BasePowerSlashCfg{}, err
	}

	return slashCfg, nil
}

func basePowerGetSlashStatsList(state *state.StateDBManage) (*mc.BasePowerStats, error) {
	if nil == state {
		return nil, ErrStatePtrIsNil
	}

	statsInfo, err := matrixstate.GetBasePowerStats(state)
	if nil != err {
		log.Error(ModuleName, "获取区块惩罚统计信息错误", err)
		return &mc.BasePowerStats{}, err
	}

	return statsInfo, nil
}
func basePowerGetLatestInitStatsNum(state *state.StateDBManage) (uint64, error) {
	if nil == state {
		return 0, ErrStatePtrIsNil
	}
	updateInfo, err := matrixstate.GetBasePowerStatsStatus(state)
	if nil != err {
		log.Debug(ModuleName, "获取区块生产统计错误", err)
		return 0, err
	}

	return updateInfo.Number, nil
}
func basePowerHasStatsInit(parentHash common.Hash, latestUpdateTime uint64) (bool, error) {
	bcInterval, err := manparams.GetBCIntervalInfoByHash(parentHash)
	if err != nil {
		log.Error(ModuleName, "获取广播周期失败", err)
		return false, err
	}

	if latestUpdateTime < bcInterval.GetLastReElectionNumber()+1 {
		return false, nil
	} else {
		return true, nil
	}

}
func BasePowerinitStatsList(state *state.StateDBManage, updateNumber uint64) {
	if nil == state {
		log.Error(ModuleName, "Input state ptr ", nil)
		return
	}
	//获取当前的初选主节点
	currElectInfo, err := matrixstate.GetElectGraph(state)
	statsList := mc.BasePowerStats{}

	if nil != err {
		log.Error(ModuleName, "获取状态树选举信息错误", err)
		matrixstate.SetBasePowerStats(state, &statsList)
		matrixstate.SetBasePowerStatsStatus(state, &mc.BasePowerSlashStatsStatus{updateNumber})
		return
	}

	for _, v := range currElectInfo.ElectList {
		if v.Type == common.RoleMiner {
			statsList.StatsList = append(statsList.StatsList, mc.BasePowerNum{Address: v.Account, ProduceNum: 0})
		}
	}
	matrixstate.SetBasePowerStats(state, &statsList)
	matrixstate.SetBasePowerStatsStatus(state, &mc.BasePowerSlashStatsStatus{updateNumber})
}
func (bc *BlockChain) BasePowerShouldAddRecorder(state *state.StateDBManage, slashCfg *mc.BasePowerSlashCfg) (*mc.BasePowerStats, bool) {
	if !slashCfg.Switcher {
		return nil, false
	}
	if statsList, err := basePowerGetSlashStatsList(state); nil != err {
		return statsList, false
	} else {
		return statsList, true
	}
}

//todo:根据头中算力检测写入
func BasePowerStatsListAddRecorder(state *state.StateDBManage, list *mc.BasePowerStats, basePowers []types.BasePowers) {
	for k, v := range list.StatsList {
		for _, bs := range basePowers {
			if v.Address.Equal(bs.Miner) {
				list.StatsList[k].ProduceNum = list.StatsList[k].ProduceNum + 1
			}
		}

	}
	err := matrixstate.SetBasePowerStats(state, list)
	if err != nil {
		log.Error(ModuleName, "写区块生成错误", err)
	}
}

func BasePowerShouldBasePowerSlash(state *state.StateDBManage, header *types.Header, slashCfg *mc.BasePowerSlashCfg) bool {
	if !slashCfg.Switcher {
		log.Debug(ModuleName, "不执行惩罚，原因", "配置关闭")
		return false
	}
	if isNextBlockMinerGenTiming(state, header.Number.Uint64(), header.ParentHash) {
		log.Trace(ModuleName, "执行惩罚，原因高度", header.Number.Uint64())
		return true
	}
	return false
}

func isNextBlockMinerGenTiming(state *state.StateDBManage, currNum uint64, parentHash common.Hash) bool {

	bcInterval, err := manparams.GetBCIntervalInfoByHash(parentHash)
	if nil != err {
		log.Error(ModuleName, "获取广播配置 err", err)
		return false
	}

	electTiming, err := getElectTimingCfg(state)
	if nil != err {
		log.Error(ModuleName, "获取选举时序错误", err)
		return false
	}

	return bcInterval.IsReElectionNumber(currNum + 1 + uint64(electTiming.MinerGen))
}

func (bc *BlockChain) BasePowerGetBlackList(state *state.StateDBManage) *mc.BasePowerSlashBlackList {
	blackList, err := matrixstate.GetBasePowerBlackList(state)
	if nil != err {
		log.Error(ModuleName, "Get Block Produce BlackList State Err", err)
		return &mc.BasePowerSlashBlackList{}
	}
	return blackList
}

type basepowerBlacklistMaintain struct {
	blacklist []mc.BasePowerSlash
}

func BasePowerNewBlackListMaintain(parentHash common.Hash, list []mc.BasePowerSlash) *basepowerBlacklistMaintain {
	var bl = new(basepowerBlacklistMaintain)
	depoistlist, err := ca.GetElectedByHeightAndRoleByHash(parentHash, common.RoleMiner)
	if nil != err {
		log.Crit(ModuleName, "读取验证者抵押列表错误", err)
		return nil
	}
	for _, v := range list {
		if v.ProhibitCycleCounter > 0 && searchExistDeposit(depoistlist, v.Address) {
			bl.blacklist = append(bl.blacklist, v)
		}
	}
	return bl
}

func (bl *blacklistMaintain) BasePowerCounterMaintain() {
	for i := 0; i < len(bl.blacklist); i++ {
		if 0 != bl.blacklist[i].ProhibitCycleCounter {
			bl.blacklist[i].ProhibitCycleCounter--
		}
	}
}

func basePowerSearchExistAddress(statsList []mc.BasePowerSlash, target common.Address) (int, bool) {

	for k, v := range statsList {
		if v.Address.Equal(target) {
			return k, true
		}
	}

	return 0, false
}

func (bl *basepowerBlacklistMaintain) BasePowerAddBlackList(statsList []mc.BasePowerNum, slashCfg *mc.BasePowerSlashCfg) {
	if nil == bl.blacklist {
		log.Warn(ModuleName, "blacklist Err", "Is Nil")
		bl.blacklist = make([]mc.BasePowerSlash, 0)
	}

	if nil == slashCfg {
		log.Warn(ModuleName, "惩罚配置错误", "未配置")
		return
	}

	if false == slashCfg.Switcher {
		bl.blacklist = make([]mc.BasePowerSlash, 0)
		return
	}

	if nil == statsList {
		log.Error(ModuleName, "统计列表错误", "未配置")
		return
	}

	if 0 == slashCfg.ProhibitCycleNum {
		log.Warn(ModuleName, "禁止周期为", "0", "不加入黑名单")
		return
	}
	for _, v := range statsList {
		if v.ProduceNum >= slashCfg.LowTHR {
			continue
		}
		if position, exist := basePowerSearchExistAddress(bl.blacklist, v.Address); exist {
			bl.blacklist[position].ProhibitCycleCounter = slashCfg.ProhibitCycleNum
		} else {
			bl.blacklist = append(bl.blacklist, mc.BasePowerSlash{Address: v.Address, ProhibitCycleCounter: slashCfg.ProhibitCycleNum})
		}
	}
}
