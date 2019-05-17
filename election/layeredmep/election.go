// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package layeredmep

import (
	"github.com/MatrixAINetwork/go-matrix/baseinterface"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/election/support"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/params/manparams"
)

type layeredMep struct {
}

func init() {
	baseinterface.RegElectPlug(manparams.ElectPlug_layerdMEP, RegInit)
}

func RegInit() baseinterface.ElectionInterface {
	return &layeredMep{}
}

func (self *layeredMep) MinerTopGen(mmrerm *mc.MasterMinerReElectionReqMsg) *mc.MasterMinerReElectionRsp {
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

func TryFilterBlockProduceBlackList(vipElec *support.Electoion, blackList []mc.UserBlockProduceSlash, minRemainNum int) int {
	//计算目前可用Node
	var availableNodeNum = vipElec.GetAvailableNodeNum()
	for i := 0; i < len(blackList); i++ {
		//如果剩余可用数小于等于最小保留数，不再过滤
		if availableNodeNum <= minRemainNum {
			return availableNodeNum
		}
		if k, status := vipElec.GetNodeByAccount(blackList[i].Address); status {
			if vipElec.NodeList[k].Usable {
				vipElec.NodeList[k].Usable = false
				log.Trace("VIP选举黑名单处理", "过滤账户", vipElec.NodeList[k].Address, "禁止周期", blackList[i].ProhibitCycleCounter)
				availableNodeNum--
			}
		}
	}
	return availableNodeNum
}
func printVipBlackList(blackList []mc.UserBlockProduceSlash) {
	if len(blackList) == 0 {
		log.Trace("VIP选举黑名单处理", "无黑名单", nil)
	} else {
		for _, v := range blackList {
			log.Trace("VIP选举黑名单处理", "账户", v.Address.String(), "禁止周期", v.ProhibitCycleCounter)
		}
	}
}
func (self *layeredMep) ValidatorTopGen(mvrerm *mc.MasterValidatorReElectionReqMsg, stateDb *state.StateDBManage) *mc.MasterValidatorReElectionRsq {
	log.Trace("分层方案", "验证者拓扑生成", mvrerm)

	vipEle := support.NewElelection(mvrerm.VIPList, mvrerm.ValidatorList, mvrerm.ElectConfig, mvrerm.RandSeed, mvrerm.SeqNum, common.RoleValidator)
	if mvrerm.ElectConfig.WhiteListSwitcher {
		vipEle.ProcessWhiteNode()
	}
	vipEle.ProcessBlackNode()

	for vipEleLoop := len(vipEle.VipLevelCfg) - 1; vipEleLoop >= 0; vipEleLoop-- {
		if vipEle.VipLevelCfg[vipEleLoop].ElectUserNum <= 0 && vipEleLoop != 0 { //vip0继续处理
			continue
		}

		//普通选举之前过滤区块生成黑名单
		if vipEleLoop == 0 {
			printVipBlackList(mvrerm.BlockProduceBlackList.BlackList)
			if vipEle.NeedNum >= vipEle.ChosedNum {
				//TryFilterBlockProduceBlackList(vipEle, mvrerm.BlockProduceBlackList.BlackList, vipEle.NeedNum-vipEle.ChosedNum)
				TryFilterBlockProduceBlackList(vipEle, mvrerm.BlockProduceBlackList.BlackList, 0)
			}
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
func (self *layeredMep) ToPoUpdate(allNative support.AllNative, topoG *mc.TopologyGraph) []mc.Alternative {
	return support.ToPoUpdate(allNative, topoG)
}

func (self *layeredMep) PrimarylistUpdate(Q0, Q1, Q2 []mc.TopologyNodeInfo, online mc.TopologyNodeInfo, flag int) ([]mc.TopologyNodeInfo, []mc.TopologyNodeInfo, []mc.TopologyNodeInfo) {
	return support.PrimarylistUpdate(Q0, Q1, Q2, online, flag)
}
