// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package nochoice

import (
	"github.com/matrix/go-matrix/baseinterface"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/election/support"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
)

const (
	DefauleStock = 1
)

type nochoice struct {
}

func init() {
	baseinterface.RegElectPlug("nochoice", RegInit)
}

func RegInit() baseinterface.ElectionInterface {
	return &nochoice{}
}

func (self *nochoice) MinerTopGen(mmrerm *mc.MasterMinerReElectionReqMsg) *mc.MasterMinerReElectionRsp {
	log.INFO("不选方案", "矿工拓扑生成", len(mmrerm.MinerList))
	MinerTopGenAns := mc.MasterMinerReElectionRsp{}

	for index, v := range mmrerm.MinerList {
		tempNode := mc.TopologyNodeInfo{
			Account:  v.Address,
			Position: uint16(index),
			Stock:    DefauleStock,
		}
		if index < support.M {
			tempNode.Type = common.RoleMiner
			MinerTopGenAns.MasterMiner = append(MinerTopGenAns.MasterMiner, tempNode)
			continue
		}
		tempNode.Type = common.RoleBackupMiner
		MinerTopGenAns.BackUpMiner = append(MinerTopGenAns.BackUpMiner, tempNode)
	}
	return &MinerTopGenAns

}

func (self *nochoice) ValidatorTopGen(mvrerm *mc.MasterValidatorReElectionReqMsg) *mc.MasterValidatorReElectionRsq {
	log.INFO("不选方案", "验证者拓扑生成", len(mvrerm.ValidatorList))
	ValidatorTop := mc.MasterValidatorReElectionRsq{}
	MasterNum := 0
	BackupNum := 0

	for index, v := range mvrerm.ValidatorList {
		tempNode := mc.TopologyNodeInfo{
			Account:  v.Address,
			Position: uint16(index),
			Stock:    DefauleStock,
		}
		if MasterNum < support.M {
			tempNode.Type = common.RoleValidator
			ValidatorTop.MasterValidator = append(ValidatorTop.MasterValidator, tempNode)
			MasterNum++
			continue
		}
		if BackupNum < support.P {
			tempNode.Type = common.RoleBackupValidator
			ValidatorTop.BackUpValidator = append(ValidatorTop.BackUpValidator, tempNode)
			BackupNum++
			continue
		}
		tempNode.Type = common.RoleCandidateValidator
		ValidatorTop.CandidateValidator = append(ValidatorTop.CandidateValidator, tempNode)

	}
	return &ValidatorTop
}

func (self *nochoice) ToPoUpdate(offline []common.Address, allNative support.AllNative, topoG *mc.TopologyGraph) []mc.Alternative {
	return support.ToPoUpdate(offline, allNative, topoG)
}

func (self *nochoice) PrimarylistUpdate(Q0, Q1, Q2 []mc.TopologyNodeInfo, online mc.TopologyNodeInfo, flag int) ([]mc.TopologyNodeInfo, []mc.TopologyNodeInfo, []mc.TopologyNodeInfo) {
	return support.PrimarylistUpdate(Q0, Q1, Q2, online, flag)
}
