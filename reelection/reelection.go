// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package reelection

import (
	"sync"
	"time"

	"github.com/MatrixAINetwork/go-matrix/baseinterface"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core"
	"github.com/MatrixAINetwork/go-matrix/core/matrixstate"
	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/pkg/errors"
)

var (
	Time_Out_Limit = 2 * time.Second
	ChanSize       = 10
)

const (
	Module = "换届服务"
)

type ElectReturnInfo struct {
	MasterMiner     []mc.ElectNodeInfo
	BackUpMiner     []mc.ElectNodeInfo
	MasterValidator []mc.ElectNodeInfo
	BackUpValidator []mc.ElectNodeInfo
}
type ReElection struct {
	bc      *core.BlockChain
	topNode TopNodeService
	random  *baseinterface.Random
	lock    sync.Mutex
}

type TopNodeService interface {
	GetConsensusOnlineResults() []*mc.HD_OnlineConsensusVoteResultMsg
}

func New(bc *core.BlockChain, random *baseinterface.Random, topNode TopNodeService) (*ReElection, error) {
	reelection := &ReElection{
		bc:      bc,
		random:  random,
		topNode: topNode,
	}
	return reelection, nil
}

func (self *ReElection) GetElection(state *state.StateDBManage, hash common.Hash) (*ElectReturnInfo, error) {
	log.Trace(Module, "GetElection", "start", "hash", hash)
	defer log.Trace(Module, "GetElection", "end", "hash", hash)
	preElectGraph, err := matrixstate.GetElectGraph(state)
	if err != nil || nil == preElectGraph {
		log.ERROR(Module, "GetElection err", err)
		return nil, err
	}

	//log.INFO(Module, "开始获取选举信息 hash", hash.String())
	height, err := self.GetNumberByHash(hash)
	//log.INFO(Module, "preElectGraph", preElectGraph, "高度", height, "err", err)
	if err != nil {
		log.Error(Module, "GetElection", "获取hash的高度失败")
		return nil, err
	}
	data := &ElectReturnInfo{}

	if self.IsMinerTopGenTiming(hash) {
		log.Trace(Module, "GetElection", "IsMinerTopGenTiming", "高度", height)
		for _, v := range preElectGraph.NextMinerElect {
			switch v.Type {
			case common.RoleMiner:
				data.MasterMiner = append(data.MasterMiner, v)

			}
		}
	}
	if self.IsValidatorTopGenTiming(hash) {
		log.Trace(Module, "GetElection", "IsValidatorTopGenTiming", "高度", height)
		for _, v := range preElectGraph.NextValidatorElect {
			switch v.Type {
			case common.RoleValidator:
				data.MasterValidator = append(data.MasterValidator, v)
			case common.RoleBackupValidator:
				data.BackUpValidator = append(data.BackUpValidator, v)
			}
		}
	}

	//log.INFO(Module, "不是任何网络切换时间点 height", height)

	return data, nil
}
func (self *ReElection) GetTopoChange(hash common.Hash, offline []common.Address, online []common.Address) ([]mc.Alternative, error) {
	log.Trace(Module, "GetTopoChange", "start", "hash", hash, "online", online, "offline", offline)
	defer log.Trace(Module, "GetTopoChange", "end", "hash", hash, "online", online, "offline", offline)
	height, err := self.GetNumberByHash(hash)
	if err != nil {
		log.ERROR(Module, "根据hash获取高度失败 err", err)
		return []mc.Alternative{}, err
	}
	bcInterval, err := self.GetBroadcastIntervalByHash(hash)
	if err != nil {
		log.ERROR(Module, "根据hash获取广播周期信息 err", err)
		return []mc.Alternative{}, err
	}
	if bcInterval.IsReElectionNumber(height + 1) {
		log.ERROR(Module, "当前是广播区块 无差值", "height", height+1)
		return []mc.Alternative{}, err
	}
	lastHash, err := self.GetHeaderHashByNumber(hash, height)
	if err != nil {
		log.ERROR(Module, "根据hash找高度失败 hash ", hash, "高度", height-1)
		return []mc.Alternative{}, err
	}

	headerPos := self.bc.GetHeaderByHash(hash)
	stateDB, err := self.bc.StateAt(headerPos.Roots)

	electState, err := matrixstate.GetElectGraph(stateDB)
	if err != nil || nil == electState {
		log.Error(Module, "get electGraph from state err", err)
		return []mc.Alternative{}, err
	}

	electOnlineState, err := matrixstate.GetElectOnlineState(stateDB)
	if err != nil || nil == electOnlineState {
		log.ERROR(Module, "get electOnlineState from state err", err)
		return []mc.Alternative{}, err
	}

	TopoGrap, err := GetCurrentTopology(lastHash, common.RoleBackupValidator|common.RoleValidator)
	if err != nil {
		log.Error(Module, "获取CA当前拓扑图失败 err", err)
		return []mc.Alternative{}, err
	}
	antive := GetAllNativeDataForUpdate(*electState, *electOnlineState, TopoGrap)
	block := self.bc.GetBlockByHash(hash)
	if block == nil {
		log.ERROR(Module, "获取区块失败,hash", hash)
	}
	DiffValidatot, err := self.TopoUpdate(antive, TopoGrap, block.ParentHash())
	if err != nil {
		log.ERROR(Module, "拓扑更新失败 err", err, "高度", height)
	}

	olineStatus := GetOnlineAlter(offline, online, *electOnlineState)
	DiffValidatot = append(DiffValidatot, olineStatus...)
	log.DEBUG(Module, "获取拓扑改变 end ", DiffValidatot)
	return DiffValidatot, nil
}

func (self *ReElection) GetNetTopologyAll(hash common.Hash) (*ElectReturnInfo, error) {
	result := &ElectReturnInfo{}
	height, err := self.GetNumberByHash(hash)
	log.Trace(Module, "GetNetTopologyAll", "start", "height", height)
	defer log.Trace(Module, "GetNetTopologyAll", "end", "height", height)
	if err != nil {
		log.ERROR(Module, "根据hash获取高度失败 err", err)
		return nil, err
	}
	bcInterval, err := self.GetBroadcastIntervalByHash(hash)
	if err != nil {
		log.ERROR(Module, "根据hash获取广播周期信息 err", err)
		return nil, err
	}
	if bcInterval.IsReElectionNumber(height+2) == false {
		log.Trace(Module, "不是广播区间前一块 不处理 height", height+1)
		return result, nil
	}

	stateDB, err := self.bc.StateAtBlockHash(hash)
	if err != nil {
		log.ERROR(Module, "获取state失败", err)
		return nil, err
	}
	electGraph, err := matrixstate.GetElectGraph(stateDB)
	if err != nil || electGraph == nil {
		log.ERROR(Module, "获取elect graph 失败", err)
		return nil, errors.Errorf("获取elect graph 失败: %v", err)
	}

	masterV, backupV, _, err := self.GetNextElectNodeInfo(electGraph, common.RoleValidator)
	if err != nil {
		log.ERROR(Module, "获取验证者全拓扑图失败 err", err)
		return nil, err
	}
	masterM, backupM, _, err := self.GetNextElectNodeInfo(electGraph, common.RoleMiner)
	if err != nil {
		log.ERROR(Module, "获取矿工全拓扑图失败 err", err)
		return nil, err
	}

	result.MasterMiner = masterM
	result.BackUpMiner = backupM
	result.MasterValidator = masterV
	result.BackUpValidator = backupV

	return result, nil

}
