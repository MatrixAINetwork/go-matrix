// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package layeredBss

import (
	"github.com/MatrixAINetwork/go-matrix/baseinterface"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/matrixstate"
	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/election/support"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/params/manparams"
)

type layeredBss struct {
}

const (
	stockExp    = 1.45
	superFactor = int64(19)
)

func init() {
	baseinterface.RegElectPlug(manparams.ElectPlug_layerdBSS, RegInit)
}

func RegInit() baseinterface.ElectionInterface {
	return &layeredBss{}
}

func (self *layeredBss) MinerTopGen(mmrerm *mc.MasterMinerReElectionReqMsg) *mc.MasterMinerReElectionRsp {
	log.Trace("MEP分层方案", "矿工拓扑生成", mmrerm)

	vipEle := support.NewMEPElection(nil, mmrerm.MinerList, mmrerm.ElectConfig, mmrerm.RandSeed, mmrerm.SeqNum, common.RoleMiner)

	if mmrerm.ElectConfig.WhiteListSwitcher {
		vipEle.ProcessWhiteNode()
	}
	vipEle.ProcessBlackNode()
	nodeList := vipEle.GetNodeByLevel(common.VIP_Nil)
	value := support.CalcValue(nodeList, common.RoleMiner)
	Chosed, value := support.GetList_MEP(value, vipEle.NeedNum, vipEle.RandSeed)

	return support.MakeMinerAns(Chosed, vipEle.SeqNum)
}

//cal the sum of deposit, return the sum value
func (self *layeredBss) ValidatorTopGen(mvrerm *mc.MasterValidatorReElectionReqMsg, stateDb *state.StateDBManage) *mc.MasterValidatorReElectionRsq {
	log.Trace("分层方案", "验证者拓扑生成", mvrerm)
	vipEle := support.NewElelection(mvrerm.VIPList, mvrerm.ValidatorList, mvrerm.ElectConfig, mvrerm.RandSeed, mvrerm.SeqNum, common.RoleValidator)
	vipEle.SetBlockBlackList(mvrerm.BlockProduceBlackList)
	if mvrerm.ElectConfig.WhiteListSwitcher {
		vipEle.ProcessWhiteNode()
	}
	vipEle.ProcessBlackNode()

	//supernode election
	superNodeS, superNodeN := vipEle.GenSuperNode(superFactor)
	vipEle.SetChosed(superNodeS)

	//get randElect List
	randPickNodeList := vipEle.GetUsableNode()
	//cal randElect node values
	randPickValue := support.CalcValueEW(randPickNodeList, stockExp)
	superNodeValue := support.CalcValueEW(superNodeN, stockExp)

	//rand Elect
	Chosed, superNodeStock := support.RandSampleFilterBlackList(randPickValue, superNodeValue, vipEle.NeedNum-vipEle.ChosedNum, vipEle.RandSeed, vipEle.BlockBlackProc)
	vipEle.SetChosed(Chosed)

	//re-calculate superNode stocks
	vipEle.SuperNodeStockProc(superNodeStock, stockExp)

	//set all blackslah node Unavailable
	vipEle.FilterBlockSlashList()

	Master := []support.Strallyint{}
	Backup := []support.Strallyint{}
	Candidate := []support.Strallyint{}

	for _, v := range vipEle.HasChosedNode {
		for _, vv := range v {
			temp := support.Strallyint{Addr: vv.Addr, Value: vv.Value, VIPLevel: common.VIP_Nil}
			if len(Master) < int(vipEle.EleCfg.ValidatorNum) {
				Master = append(Master, temp)
				continue
			}
			if len(Backup) < int(vipEle.EleCfg.BackValidator) {
				Backup = append(Backup, temp)
				continue
			}
		}
	}

	lastNode := vipEle.GetLastNode()
	for _, v := range lastNode {
		if len(Candidate) < int(4*vipEle.EleCfg.ValidatorNum-vipEle.EleCfg.BackValidator) {
			Candidate = append(Candidate, support.Strallyint{Addr: v.Address, Value: 1})
		}
	}
	ans := support.MakeValidatoeTopGenAns(mvrerm.SeqNum, Master, Backup, Candidate)
	//ans.UpDateList = &mc.BlockProduceSlashBlackList{vipEle.BlockBlackProc.List}
	matrixstate.SetBlockProduceBlackList(stateDb, &mc.BlockProduceSlashBlackList{vipEle.BlockBlackProc.List})
	return ans
}

func TransVIPNode(vipnode []support.Node) []support.Strallyint {
	ans := []support.Strallyint{}
	for _, v := range vipnode {
		ans = append(ans, support.Strallyint{Value: support.DefaultStock, Addr: v.Address})
	}
	return ans
}
func (self *layeredBss) ToPoUpdate(allNative support.AllNative, topoG *mc.TopologyGraph) []mc.Alternative {
	return support.ToPoUpdate(allNative, topoG)
}

func (self *layeredBss) PrimarylistUpdate(Q0, Q1, Q2 []mc.TopologyNodeInfo, online mc.TopologyNodeInfo, flag int) ([]mc.TopologyNodeInfo, []mc.TopologyNodeInfo, []mc.TopologyNodeInfo) {
	return support.PrimarylistUpdate(Q0, Q1, Q2, online, flag)
}
