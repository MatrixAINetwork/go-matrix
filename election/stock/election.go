// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package stock

import (
	"github.com/MatrixAINetwork/go-matrix/baseinterface"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/election/support"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
)

type StockElect struct {
}

func init() {
	baseinterface.RegElectPlug("stock", RegInit)
}

func RegInit() baseinterface.ElectionInterface {
	return &StockElect{}
}

func (self *StockElect) MinerTopGen(mmrerm *mc.MasterMinerReElectionReqMsg) *mc.MasterMinerReElectionRsp {
	log.INFO("选举种子", "矿工拓扑生成", len(mmrerm.MinerList))
	nodeElect := support.NewElelection(nil, mmrerm.MinerList, mmrerm.ElectConfig, mmrerm.RandSeed, mmrerm.SeqNum, common.RoleMiner)
	nodeElect.ProcessBlackNode()
	nodeElect.ProcessWhiteNode()
	//nodeElect.DisPlayNode()

	value := nodeElect.GetWeight(common.RoleMiner)
	//for _,v:=range value{
	//	fmt.Println(v.Addr.String(),v.Value)
	//}
	Master, value := support.GetList_Common(value, int(nodeElect.EleCfg.MinerNum), nodeElect.RandSeed)
	return support.MakeMinerAns(Master, nodeElect.SeqNum)
}

func (self *StockElect) ValidatorTopGen(mvrerm *mc.MasterValidatorReElectionReqMsg, stateDb *state.StateDBManage) *mc.MasterValidatorReElectionRsq {
	log.INFO("选举种子", "验证者拓扑生成", len(mvrerm.ValidatorList))
	nodeElect := support.NewElelection(nil, mvrerm.ValidatorList, mvrerm.ElectConfig, mvrerm.RandSeed, mvrerm.SeqNum, common.RoleValidator)

	nodeElect.ProcessBlackNode()
	nodeElect.ProcessWhiteNode()
	value := nodeElect.GetWeight(common.RoleValidator)
	//for _,v:=range value{
	//	fmt.Println(v.Value,v.Addr.String())
	//}
	Master, value := support.GetList_Common(value, int(nodeElect.EleCfg.ValidatorNum), nodeElect.RandSeed)

	BackUp, value := support.GetList_Common(value, int(nodeElect.EleCfg.BackValidator), nodeElect.RandSeed)

	Candid, value := support.GetList_Common(value, len(value), nodeElect.RandSeed)

	return support.MakeValidatoeTopGenAns(mvrerm.SeqNum, Master, BackUp, Candid)

}

func (self *StockElect) ToPoUpdate(allNative support.AllNative, topoG *mc.TopologyGraph) []mc.Alternative {

	return support.ToPoUpdate(allNative, topoG)
}

func (self *StockElect) PrimarylistUpdate(Q0, Q1, Q2 []mc.TopologyNodeInfo, online mc.TopologyNodeInfo, flag int) ([]mc.TopologyNodeInfo, []mc.TopologyNodeInfo, []mc.TopologyNodeInfo) {
	return support.PrimarylistUpdate(Q0, Q1, Q2, online, flag)
}
