// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package reelection

import (
	"encoding/json"
	"errors"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/matrixstate"
	"github.com/matrix/go-matrix/election/support"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
)

func GetAllNativeDataForUpdate(electstate mc.ElectGraph, electonline mc.ElectOnlineStatus, top *mc.TopologyGraph) support.AllNative {
	mapTopStatus := make(map[common.Address]common.RoleType, 0)
	for _, v := range top.NodeList {
		mapTopStatus[v.Account] = v.Type
	}
	native := support.AllNative{}
	mapELectStatus := make(map[common.Address]common.RoleType, 0)
	for _, v := range electstate.ElectList {
		mapELectStatus[v.Account] = v.Type
		switch v.Type {
		case common.RoleValidator:
			native.Master = append(native.Master, v)
		case common.RoleBackupValidator:
			native.BackUp = append(native.BackUp, v)
		case common.RoleCandidateValidator:
			native.Candidate = append(native.Candidate, v)
		}
	}
	for _, v := range electonline.ElectOnline {
		if v.Position != common.PosOnline { //过滤在线的
			continue
		}
		if _, ok := mapTopStatus[v.Account]; ok == true { //过滤当前不在拓扑图中的
			continue
		}
		if _, ok := mapELectStatus[v.Account]; ok == true { //在初选列表中的
			switch mapELectStatus[v.Account] {
			case common.RoleValidator:
				native.MasterQ = append(native.MasterQ, v.Account)
			case common.RoleBackupValidator:
				native.BackUpQ = append(native.BackUpQ, v.Account)
			case common.RoleCandidateValidator:
				native.CandidateQ = append(native.CandidateQ, v.Account)
			}
		}
	}
	return native
}
func GetOnlineAlter(offline []common.Address, online []common.Address, electonline mc.ElectOnlineStatus) []mc.Alternative {
	ans := []mc.Alternative{}
	mappOnlineStatus := make(map[common.Address]uint16)
	for _, v := range electonline.ElectOnline {
		mappOnlineStatus[v.Account] = v.Position
	}
	for _, v := range offline {
		if _, ok := mappOnlineStatus[v]; ok == false {
			log.ERROR(Module, "计算下线节点的alter时 下线节点不在初选列表中 账户", v.String())
			continue
		}
		if mappOnlineStatus[v] == common.PosOffline {
			log.ERROR(Module, "该节点已处于下线阶段 不需要上块 账户", v.String())
			continue
		}
		temp := mc.Alternative{
			A:        v,
			Position: common.PosOffline,
		}
		ans = append(ans, temp)
	}

	for _, v := range online {
		if _, ok := mappOnlineStatus[v]; ok == false {
			log.ERROR(Module, "计算上线节点的alter时 上线节点不在初选列表中 账户", v.String())
			continue
		}
		if mappOnlineStatus[v] == common.PosOnline {
			log.ERROR(Module, "该节点已处于上线阶段，不需要上块 账户", v.String())
			continue
		}
		temp := mc.Alternative{
			A:        v,
			Position: common.PosOnline,
		}
		ans = append(ans, temp)
	}
	log.INFO(Module, "计算上下线节点结果 online", online, "offline", offline, "ans", ans)
	return ans
}

func (self *ReElection) TopoUpdate(allNative support.AllNative, top *mc.TopologyGraph, height uint64) ([]mc.Alternative, error) {
	elect, err := self.GetElectPlug(height)
	if err != nil {
		log.ERROR(Module, "获取选举插件")
		return []mc.Alternative{}, err
	}

	data, err := self.bc.GetMatrixStateDataByNumber(mc.MSKeyElectConfigInfo, height)
	if err != nil {
		log.ERROR("GetElectInfo", "获取选举基础信息失败 err", err)
		return nil, err
	}
	electInfo, OK := data.(*mc.ElectConfigInfo)
	if OK == false || electInfo == nil {
		log.ERROR("ElectConfigInfo", "ElectConfigInfo ", "反射失败", "高度", height)
		return nil, errors.New("反射失败")
	}
	allNative.ElectInfo = electInfo
	return elect.ToPoUpdate(allNative, top), nil
}

func (self *ReElection) LastMinerGenTimeStamp(height uint64, types common.RoleType, hash common.Hash) (uint64, error) {

	data, err := self.GetElectGenTimes(height)
	if err != nil {
		log.ERROR(Module, "获取配置文件失败 err", err)
		return 0, err
	}
	minerGenTime := uint64(data.MinerNetChange)
	validatorGenTime := uint64(data.ValidatorNetChange)

	bcInterval, err := self.GetBroadcastIntervalByHash(hash)
	if err != nil {
		log.ERROR(Module, "根据hash获取广播周期信息 err", err)
		return 0, err
	}
	switch types {
	case common.RoleMiner:
		return bcInterval.GetNextReElectionNumber(height) - minerGenTime, nil
	default:
		return bcInterval.GetNextReElectionNumber(height) - validatorGenTime, nil
	}

}

func (self *ReElection) GetTopNodeInfo(hash common.Hash, types common.RoleType) ([]mc.ElectNodeInfo, []mc.ElectNodeInfo, []mc.ElectNodeInfo, error) {
	height, err := self.GetNumberByHash(hash)
	if err != nil {
		log.ERROR(Module, "根据hash获取高度失败 err", err)
		return []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, err
	}
	heightPos, err := self.LastMinerGenTimeStamp(height, types, hash)
	if err != nil {
		log.ERROR(Module, "根据生成点高度失败", height, "types", types)
		return []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, err
	}

	hashPos, err := self.GetHeaderHashByNumber(hash, heightPos)
	log.INFO(Module, "GetTopNodeInfo pos", heightPos)
	if err != nil {
		log.ERROR(Module, "根据hash算父header失败 hash", hashPos)
		return []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, err
	}
	headerPos := self.bc.GetHeaderByHash(hashPos)
	stateDB, err := self.bc.StateAt(headerPos.Root)
	ElectGraphBytes := stateDB.GetMatrixData(matrixstate.GetKeyHash(mc.MSKeyElectGraph))
	var electState mc.ElectGraph
	if err := json.Unmarshal(ElectGraphBytes, &electState); err != nil {
		log.ERROR(Module, "GetElection Unmarshal err", err)
		return []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, err
	}
	master := []mc.ElectNodeInfo{}
	backup := []mc.ElectNodeInfo{}
	cand := []mc.ElectNodeInfo{}

	switch types {
	case common.RoleMiner:
		for _, v := range electState.NextMinerElect {
			switch v.Type {
			case common.RoleMiner:
				master = append(master, v)
			}
		}
	case common.RoleValidator:
		for _, v := range electState.NextValidatorElect {
			switch v.Type {
			case common.RoleValidator:
				master = append(master, v)
			case common.RoleBackupValidator:
				backup = append(backup, v)
			case common.RoleCandidateValidator:
				cand = append(cand, v)

			}
		}
	}
	return master, backup, cand, nil
}
