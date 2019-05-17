// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package reelection

import (
	"errors"

	"github.com/MatrixAINetwork/go-matrix/ca"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	//"github.com/MatrixAINetwork/go-matrix/core/matrixstate"
	"github.com/MatrixAINetwork/go-matrix/baseinterface"
	"github.com/MatrixAINetwork/go-matrix/core/matrixstate"
)

func (self *ReElection) GetElectGenTimes(hash common.Hash) (*mc.ElectGenTimeStruct, error) {
	st, err := self.bc.StateAtBlockHash(hash)
	if err != nil {
		log.Error("GetElectGenTimes", "获取state失败", err, "hash", hash)
		return nil, err
	}
	electGenConfig, err := matrixstate.GetElectGenTime(st)
	if err != nil || nil == electGenConfig {
		log.Error("GetElectGenTimes", "获取选举时间点信息失败 err", err)
		return nil, err
	}
	return electGenConfig, nil
}
func (self *ReElection) GetElectConfig(hash common.Hash) (*mc.ElectConfigInfo_All, error) {
	st, err := self.bc.StateAtBlockHash(hash)
	if err != nil {
		log.Error("GetElectInfo", "获取state失败", err, "hash", hash)
		return nil, err
	}
	electInfo, err := matrixstate.GetElectConfigInfo(st)
	if err != nil || electInfo == nil {
		log.ERROR("GetElectInfo", "获取选举基础信息失败 err", err)
		return nil, err
	}

	electMinerNum, err := matrixstate.GetElectMinerNum(st)
	if err != nil {
		log.ERROR("MSKeyElectMinerNum", "获取MSKeyElectMinerNum err", err)
		return nil, err
	}

	blackList, err := matrixstate.GetElectBlackList(st)
	if err != nil {
		log.Error("MSKeyElectBlackList", "MSKeyElectBlackList", "反射失败", "hash", hash)
		return nil, err
	}

	innerMiners, err := matrixstate.GetInnerMinerAccounts(st)
	if err != nil {
		log.Error("MSKeyInnerList", "MSKeyAccountInnerMiners", "反射失败", "hash", hash)
		return nil, err
	}
	blackList = append(blackList, innerMiners...)

	whiteList, err := matrixstate.GetElectWhiteList(st)
	if err != nil {
		log.Error("MSKeyElectWhiteList", "MSKeyElectWhiteList", "反射失败", "hash", hash)
		return nil, err
	}

	whiteListSwitcher, err := matrixstate.GetElectWhiteListSwitcher(st)
	if err != nil {
		log.Error("MSKeyElectWhiteList", "MSKeyElectWhiteList", "反射失败", "hash", hash)
		return nil, err
	}
	elect := &mc.ElectConfigInfo_All{
		MinerNum:          electMinerNum.MinerNum,
		ValidatorNum:      electInfo.ValidatorNum,
		BackValidator:     electInfo.BackValidator,
		ElectPlug:         electInfo.ElectPlug,
		WhiteList:         whiteList,
		BlackList:         blackList,
		WhiteListSwitcher: whiteListSwitcher,
	}

	return elect, nil
}
func (self *ReElection) GetViPList(hash common.Hash) ([]mc.VIPConfig, error) {
	st, err := self.bc.StateAtBlockHash(hash)
	if err != nil {
		log.Error("GetViPList", "获取state失败", err, "number", hash)
		return nil, err
	}
	vipList, err := matrixstate.GetVIPConfig(st)
	if err != nil || vipList == nil {
		log.ERROR("GetViPList", "获取选举基础信息失败 err", err)
		return nil, err
	}
	return vipList, nil
}

func (self *ReElection) GetElectPlug(hash common.Hash) (baseinterface.ElectionInterface, error) {
	st, err := self.bc.StateAtBlockHash(hash)
	if err != nil {
		log.Error("GetElectPlug", "获取state失败", err, "number", hash)
		return nil, err
	}
	electInfo, err := matrixstate.GetElectConfigInfo(st)
	if err != nil || electInfo == nil {
		log.ERROR("GetElectPlug", "获取选举基础信息失败 err", err)
		return nil, err
	}
	return baseinterface.NewElect(electInfo.ElectPlug), nil
}

func (self *ReElection) GetBroadcastIntervalByHash(hash common.Hash) (*mc.BCIntervalInfo, error) {
	st, err := self.bc.StateAtBlockHash(hash)
	if err != nil {
		log.Error("GetBroadcastIntervalByHash", "获取state失败", err, "hash", hash.Hex())
		return nil, err
	}

	bcInterval, err := matrixstate.GetBroadcastInterval(st)
	if err != nil {
		return nil, err
	}
	return bcInterval, nil
}

func (self *ReElection) GetNumberByHash(hash common.Hash) (uint64, error) {
	tHeader := self.bc.GetHeaderByHash(hash)
	if tHeader == nil {
		log.Error(Module, "GetNumberByHash 根据hash算header失败 hash", hash.String())
		return 0, errors.New("根据hash算header失败")
	}
	if tHeader.Number == nil {
		log.Error(Module, "GetNumberByHash header 内的高度获取失败", hash.String())
		return 0, errors.New("header 内的高度获取失败")
	}
	return tHeader.Number.Uint64(), nil
}

func (self *ReElection) GetHeaderHashByNumber(hash common.Hash, height uint64) (common.Hash, error) {
	AimHash, err := self.bc.GetAncestorHash(hash, height)
	if err != nil {
		log.Error(Module, "获取祖先hash失败 hash", hash.String(), "height", height, "err", err)
		return common.Hash{}, err
	}
	return AimHash, nil
}

func GetCurrentTopology(hash common.Hash, reqtypes common.RoleType) (*mc.TopologyGraph, error) {
	return ca.GetTopologyByHash(reqtypes, hash)
	//return ca.GetTopologyByNumber(reqtypes, height)
}

func CheckBlock(block *types.Block) error {
	if block == nil {
		return errors.New("block为空")
	}
	if block.Header() == nil {
		return errors.New("block.Header()为空")
	}
	if block.Header().Number == nil {
		return errors.New("block.Header.Number为空 ")
	}
	return nil
}

func (self *ReElection) TransferToElectionStu(info *ElectReturnInfo) []common.Elect {
	result := make([]common.Elect, 0)

	srcMap := make(map[common.ElectRoleType][]mc.ElectNodeInfo)
	srcMap[common.ElectRoleMiner] = info.MasterMiner
	//srcMap[common.ElectRoleMinerBackUp] = info.BackUpMiner
	srcMap[common.ElectRoleValidator] = info.MasterValidator
	srcMap[common.ElectRoleValidatorBackUp] = info.BackUpValidator
	orderIndex := []common.ElectRoleType{common.ElectRoleValidator, common.ElectRoleValidatorBackUp, common.ElectRoleMiner}

	for _, role := range orderIndex {
		src := srcMap[role]
		for _, node := range src {
			e := common.Elect{
				Account: node.Account,
				Stock:   node.Stock,
				Type:    role,
				VIP:     node.VIPLevel,
			}

			result = append(result, e)
		}
	}

	return result
}

func (self *ReElection) TransferToNetTopologyAllStu(info *ElectReturnInfo) *common.NetTopology {
	result := &common.NetTopology{
		Type:            common.NetTopoTypeAll,
		NetTopologyData: make([]common.NetTopologyData, 0),
	}

	srcMap := make(map[common.ElectRoleType][]mc.ElectNodeInfo)
	srcMap[common.ElectRoleMiner] = info.MasterMiner
	//srcMap[common.ElectRoleMinerBackUp] = info.BackUpMiner
	srcMap[common.ElectRoleValidator] = info.MasterValidator
	srcMap[common.ElectRoleValidatorBackUp] = info.BackUpValidator
	orderIndex := []common.ElectRoleType{common.ElectRoleMiner, common.ElectRoleValidator, common.ElectRoleValidatorBackUp}

	for _, role := range orderIndex {
		src := srcMap[role]
		for i, node := range src {
			data := common.NetTopologyData{
				Account:  node.Account,
				Position: common.GeneratePosition(uint16(i), role),
			}
			result.NetTopologyData = append(result.NetTopologyData, data)
		}
	}

	return result
}

func (self *ReElection) TransferToNetTopologyChgStu(alterInfo []mc.Alternative) *common.NetTopology {
	result := &common.NetTopology{
		Type:            common.NetTopoTypeChange,
		NetTopologyData: make([]common.NetTopologyData, 0),
	}

	for _, alter := range alterInfo {
		data := common.NetTopologyData{
			Account:  alter.A,
			Position: alter.Position,
		}
		result.NetTopologyData = append(result.NetTopologyData, data)
	}

	return result
}
