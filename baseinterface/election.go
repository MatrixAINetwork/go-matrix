// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package baseinterface

import (
	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/election/support"
	"github.com/MatrixAINetwork/go-matrix/mc"
)

const (
	ModuleElection   = "选举模块"
	DefaultElectPlug = "layered"
)

var (
	electionPlugs = make(map[string]func() ElectionInterface)
)

func RegElectPlug(name string, value func() ElectionInterface) {
	//	fmt.Println("选举服务 注册函数", "name", name)
	electionPlugs[name] = value
}

func NewElect(ElectPlugs string) ElectionInterface {
	//从配置中获取参数
	if _, ok := electionPlugs[ElectPlugs]; ok {
		return electionPlugs[ElectPlugs]()
	}
	return electionPlugs[DefaultElectPlug]()
}

type ElectionInterface interface {
	MinerTopGen(*mc.MasterMinerReElectionReqMsg) *mc.MasterMinerReElectionRsp
	ValidatorTopGen(*mc.MasterValidatorReElectionReqMsg, *state.StateDBManage) *mc.MasterValidatorReElectionRsq
	ToPoUpdate(support.AllNative, *mc.TopologyGraph) []mc.Alternative
	//	PrimarylistUpdate([]mc.TopologyNodeInfo, []mc.TopologyNodeInfo, []mc.TopologyNodeInfo, mc.TopologyNodeInfo, int) ([]mc.TopologyNodeInfo, []mc.TopologyNodeInfo, []mc.TopologyNodeInfo)
}
