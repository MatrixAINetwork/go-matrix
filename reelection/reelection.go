// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package reelection

import (
	"sync"
	"time"

	"encoding/json"

	"github.com/matrix/go-matrix/baseinterface"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core"
	"github.com/matrix/go-matrix/core/matrixstate"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
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
	bc     *core.BlockChain
	random *baseinterface.Random
	lock   sync.Mutex
}

func New(bc *core.BlockChain, random *baseinterface.Random) (*ReElection, error) {
	reelection := &ReElection{
		bc:     bc,
		random: random,
	}
	return reelection, nil
}

func (self *ReElection) GetElection(state *state.StateDB, hash common.Hash) (*ElectReturnInfo, error) {
	log.INFO(Module, "GetElection", "start", "hash", hash)
	defer log.INFO(Module, "GetElection", "end", "hash", hash)
	preElectGraphBytes := state.GetMatrixData(matrixstate.GetKeyHash(mc.MSKeyElectGraph))
	var electState mc.ElectGraph
	if err := json.Unmarshal(preElectGraphBytes, &electState); err != nil {
		log.ERROR(Module, "GetElection Unmarshal err", err)
		return nil, err
	}
	log.INFO(Module, "开始获取选举信息 hash", hash.String())
	height, err := self.GetNumberByHash(hash)
	log.INFO(Module, "electStatte", electState, "高度", height, "err", err)
	if err != nil {
		log.Error(Module, "GetElection", "获取hash的高度失败")
		return nil, err
	}
	data := &ElectReturnInfo{}

	if self.IsMinerTopGenTiming(hash) {
		log.INFO(Module, "GetElection", "IsMinerTopGenTiming", "高度", height)
		for _, v := range electState.NextMinerElect {
			switch v.Type {
			case common.RoleMiner:
				data.MasterMiner = append(data.MasterMiner, v)

			}
		}
	}
	if self.IsValidatorTopGenTiming(hash) {
		log.INFO(Module, "GetElection", "IsValidatorTopGenTiming", "高度", height)
		for _, v := range electState.NextValidatorElect {
			switch v.Type {
			case common.RoleValidator:
				data.MasterValidator = append(data.MasterValidator, v)
			case common.RoleBackupValidator:
				data.BackUpValidator = append(data.BackUpValidator, v)
			}
		}
	}

	log.INFO(Module, "不是任何网络切换时间点 height", height)

	return data, nil
}
func (self *ReElection) GetTopoChange(hash common.Hash, offline []common.Address, online []common.Address) ([]mc.Alternative, error) {
	log.INFO(Module, "GetTopoChange", "start", "hash", hash, "online", online, "offline", offline)
	defer log.INFO(Module, "GetTopoChange", "end", "hash", hash, "online", online, "offline", offline)
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
	stateDB, err := self.bc.StateAt(headerPos.Root)

	ElectGraphBytes := stateDB.GetMatrixData(matrixstate.GetKeyHash(mc.MSKeyElectGraph))
	var electState mc.ElectGraph
	if err := json.Unmarshal(ElectGraphBytes, &electState); err != nil {
		log.ERROR(Module, "GetElection Unmarshal err", err)
		return []mc.Alternative{}, err
	}
	ElectOnlineBytes := stateDB.GetMatrixData(matrixstate.GetKeyHash(mc.MSKeyElectOnlineState))
	var electOnlineState mc.ElectOnlineStatus
	if err := json.Unmarshal(ElectOnlineBytes, &electOnlineState); err != nil {
		log.ERROR(Module, "GetElection Unmarshal err", err)
		return []mc.Alternative{}, err
	}

	TopoGrap, err := GetCurrentTopology(lastHash, common.RoleBackupValidator|common.RoleValidator)
	if err != nil {
		log.Error(Module, "获取CA当前拓扑图失败 err", err)
		return []mc.Alternative{}, err
	}
	antive := GetAllNativeDataForUpdate(electState, electOnlineState, TopoGrap)
	DiffValidatot, err := self.TopoUpdate(antive, TopoGrap, height-1)
	if err != nil {
		log.ERROR(Module, "拓扑更新失败 err", err, "高度", height)
	}

	olineStatus := GetOnlineAlter(offline, online, electOnlineState)
	DiffValidatot = append(DiffValidatot, olineStatus...)
	log.INFO(Module, "获取拓扑改变 end ", DiffValidatot)
	return DiffValidatot, nil
}

func (self *ReElection) GetNetTopologyAll(hash common.Hash) (*ElectReturnInfo, error) {

	result := &ElectReturnInfo{}
	//todo 从hash获取state， 得全拓扑
	height, err := self.GetNumberByHash(hash)
	log.INFO(Module, "GetNetTopologyAll", "start", "height", height)
	defer log.INFO(Module, "GetNetTopologyAll", "end", "height", height)
	if err != nil {
		log.ERROR(Module, "根据hash获取高度失败 err", err)
		return nil, err
	}
	bcInterval, err := self.GetBroadcastIntervalByHash(hash)
	if err != nil {
		log.ERROR(Module, "根据hash获取广播周期信息 err", err)
		return nil, err
	}
	if bcInterval.IsReElectionNumber(height + 2) {
		masterV, backupV, _, err := self.GetTopNodeInfo(hash, common.RoleValidator)
		if err != nil {
			log.ERROR(Module, "获取验证者全拓扑图失败 err", err)
			return nil, err
		}
		masterM, backupM, _, err := self.GetTopNodeInfo(hash, common.RoleMiner)
		if err != nil {
			log.ERROR(Module, "获取矿工全拓扑图失败 err", err)
			return nil, err
		}

		result = &ElectReturnInfo{
			MasterMiner:     masterM,
			BackUpMiner:     backupM,
			MasterValidator: masterV,
			BackUpValidator: backupV,
		}
		log.INFO(Module, "是299 height", height)
		return result, nil

	}
	log.Info(Module, "不是广播区间前一块 不处理 height", height)
	return result, nil
}
