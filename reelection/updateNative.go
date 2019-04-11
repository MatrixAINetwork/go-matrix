// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package reelection

import (
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/matrixstate"
	"github.com/MatrixAINetwork/go-matrix/election/support"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/pkg/errors"
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

func (self *ReElection) TopoUpdate(allNative support.AllNative, top *mc.TopologyGraph, hash common.Hash) ([]mc.Alternative, error) {
	elect, err := self.GetElectPlug(hash)
	if err != nil {
		log.ERROR(Module, "获取选举插件")
		return []mc.Alternative{}, err
	}

	st, err := self.bc.StateAtBlockHash(hash)
	if err != nil {
		log.Error(Module, "get state by height err", err, "hash", hash)
		return nil, err
	}
	electInfo, err := matrixstate.GetElectConfigInfo(st)
	if err != nil || electInfo == nil {
		log.ERROR("GetElectInfo", "获取选举基础信息失败 err", err)
		return nil, err
	}
	allNative.ElectInfo = electInfo
	return elect.ToPoUpdate(allNative, top), nil
}

func (self *ReElection) LastMinerGenTimeStamp(height uint64, types common.RoleType, hash common.Hash) (uint64, error) {

	data, err := self.GetElectGenTimes(hash)
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

func (self *ReElection) GetNextElectNodeInfo(electGraph *mc.ElectGraph, types common.RoleType) (master []mc.ElectNodeInfo, backup []mc.ElectNodeInfo, cand []mc.ElectNodeInfo, err error) {
	if electGraph == nil {
		err = errors.New("param elect graph is nil")
		return
	}

	master = []mc.ElectNodeInfo{}
	backup = []mc.ElectNodeInfo{}
	cand = []mc.ElectNodeInfo{}

	switch types {
	case common.RoleMiner:
		size := len(electGraph.NextMinerElect)
		if size != 0 {
			master = make([]mc.ElectNodeInfo, size, size)
			if copy(master, electGraph.NextMinerElect) != size {
				err = errors.New("copy next miner graph err")
				return
			}
		}
	case common.RoleValidator:
		for _, v := range electGraph.NextValidatorElect {
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
