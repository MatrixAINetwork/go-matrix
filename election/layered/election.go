// Copyright (c) 2018-2019 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package layered

import (
	"github.com/matrix/go-matrix/baseinterface"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/election/support"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/params/manparams"
)

type layered struct {
}

func init() {
	baseinterface.RegElectPlug(manparams.ElectPlug_layerd, RegInit)
}

func RegInit() baseinterface.ElectionInterface {
	return &layered{}
}

func (self *layered) MinerTopGen(mmrerm *mc.MasterMinerReElectionReqMsg) *mc.MasterMinerReElectionRsp {
	log.INFO("分层方案", "矿工拓扑生成", mmrerm)
	vipEle := support.NewElelection(nil, mmrerm.MinerList, mmrerm.ElectConfig, mmrerm.RandSeed, mmrerm.SeqNum, common.RoleMiner)

	vipEle.ProcessBlackNode()
	vipEle.ProcessWhiteNode()
	nodeList := vipEle.GetNodeByLevel(common.VIP_Nil)
	value := support.CalcValue(nodeList, common.RoleMiner)
	Chosed, value := support.GetList_Common(value, vipEle.NeedNum, vipEle.RandSeed)
	return support.MakeMinerAns(Chosed, vipEle.SeqNum)

}

func (self *layered) ValidatorTopGen(mvrerm *mc.MasterValidatorReElectionReqMsg) *mc.MasterValidatorReElectionRsq {
	log.INFO("分层方案", "验证者拓扑生成", mvrerm)

	vipEle := support.NewElelection(mvrerm.VIPList, mvrerm.ValidatorList, mvrerm.ElectConfig, mvrerm.RandSeed, mvrerm.SeqNum, common.RoleValidator)
	vipEle.ProcessBlackNode()
	vipEle.ProcessWhiteNode()
	//vipEle.DisPlayNode()

	for vipEleLoop := len(vipEle.VipLevelCfg) - 1; vipEleLoop >= 0; vipEleLoop-- {
		if vipEle.VipLevelCfg[vipEleLoop].ElectUserNum <= 0 && vipEleLoop != 0 { //vip0继续处理
			continue
		}
		nodeList := vipEle.GetNodeByLevel(common.GetVIPLevel(vipEleLoop))

		value := support.CalcValue(nodeList, common.RoleValidator)
		curNeed := 0
		if vipEleLoop == 0 {
			curNeed = vipEle.NeedNum - vipEle.ChosedNum
		} else {
			curNeed = int(vipEle.VipLevelCfg[vipEleLoop].ElectUserNum)
		}
		if curNeed > vipEle.NeedNum-vipEle.ChosedNum {
			curNeed = vipEle.NeedNum - vipEle.ChosedNum
		}

		Chosed := []support.Strallyint{}

		if vipEleLoop == 0 {
			Chosed, value = support.GetList_Common(value, curNeed, vipEle.RandSeed)
		} else {
			Chosed, value = support.GetList_VIP(value, curNeed, vipEle.RandSeed)
		}

		vipEle.SetChosed(Chosed)

	}

	Master := []support.Strallyint{}
	Backup := []support.Strallyint{}
	Candidate := []support.Strallyint{}

	for k, v := range vipEle.HasChosedNode {
		for _, vv := range v {
			temp := support.Strallyint{}
			if k == len(vipEle.HasChosedNode)-1 {
				temp = support.Strallyint{Addr: vv.Addr, Value: vv.Value, VIPLevel: common.VIP_Nil}
			} else {
				temp = support.Strallyint{Addr: vv.Addr, Value: vipEle.GetVipStock(vv.Addr), VIPLevel: common.GetVIPLevel(len(vipEle.VipLevelCfg) - 1 - k)}
			}

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
	return support.MakeValidatoeTopGenAns(mvrerm.SeqNum, Master, Backup, Candidate)
}

func TransVIPNode(vipnode []support.Node) []support.Strallyint {
	ans := []support.Strallyint{}
	for _, v := range vipnode {
		ans = append(ans, support.Strallyint{Value: support.DefaultStock, Addr: v.Address})
	}
	return ans
}
func (self *layered) ToPoUpdate(allNative support.AllNative, topoG *mc.TopologyGraph) []mc.Alternative {
	return support.ToPoUpdate(allNative, topoG)
}

func (self *layered) PrimarylistUpdate(Q0, Q1, Q2 []mc.TopologyNodeInfo, online mc.TopologyNodeInfo, flag int) ([]mc.TopologyNodeInfo, []mc.TopologyNodeInfo, []mc.TopologyNodeInfo) {
	return support.PrimarylistUpdate(Q0, Q1, Q2, online, flag)
}
