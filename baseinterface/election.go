// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package baseinterface

import (
	"fmt"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/election/support"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/params/manparams"
)

const (
	ModuleElection   = "选举模块"
	DefaultElectPlug = "stock"
)

var (
	electionPlugs = make(map[string]func() ElectionInterface)
)

func RegElectPlug(name string, value func() ElectionInterface) {
	fmt.Println("选举服务 注册函数", "name", name)
	electionPlugs[name] = value
}

func NewElect() ElectionInterface {
	//从配置中获取参数
	if _, ok := electionPlugs[manparams.ElectPlugs]; ok {
		return electionPlugs[manparams.ElectPlugs]()
	}
	return electionPlugs[DefaultElectPlug]()
}

type ElectionInterface interface {
	MinerTopGen(*mc.MasterMinerReElectionReqMsg) *mc.MasterMinerReElectionRsp
	ValidatorTopGen(*mc.MasterValidatorReElectionReqMsg) *mc.MasterValidatorReElectionRsq
	ToPoUpdate([]common.Address, support.AllNative, *mc.TopologyGraph) []mc.Alternative
	PrimarylistUpdate([]mc.TopologyNodeInfo, []mc.TopologyNodeInfo, []mc.TopologyNodeInfo, mc.TopologyNodeInfo, int) ([]mc.TopologyNodeInfo, []mc.TopologyNodeInfo, []mc.TopologyNodeInfo)
}
