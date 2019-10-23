// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package layereddp

import (
	"github.com/MatrixAINetwork/go-matrix/baseinterface"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/common/mt19937"
	"github.com/MatrixAINetwork/go-matrix/core/matrixstate"
	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/election/support"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/params/manparams"
)

type layeredDdp struct {
}

const (
	stockExp      = 1.45
	superFactor   = int64(19)
	minMinersBase = uint64(1024)
	MinersFactor  = uint64(64)
	addMinesNum   = uint64(2)
)

func init() {
	baseinterface.RegElectPlug(manparams.ElectPlug_layerdDP, RegInit)
}

func RegInit() baseinterface.ElectionInterface {
	return &layeredDdp{}
}

func (self *layeredDdp) MinerTopGen(mmrerm *mc.MasterMinerReElectionReqMsg, stateDb *state.StateDBManage) *mc.MasterMinerReElectionRsp {

	eleDpi := self.getEleDpi(stateDb, mmrerm)

	dpElect := self.CreateDpElect(stateDb, mmrerm)

	chosedNodes := self.minerdp(eleDpi, dpElect)

	//todo:正式上线删除日志
	log.Info("动态选举方案", "选出节点", len(chosedNodes))
	/*	for i, v := range chosedNodes {
		log.Trace("动态选举方案", "序号", i, "账户", v.String())
	}*/
	log.Info("动态选举方案", "候选节点个数", len(eleDpi.CandidateList))
	/*	for i, v := range eleDpi.CandidateList {
		log.Trace("动态选举方案", "序号", i, "账户", v.String())
	}*/
	log.Info("动态选举方案", "算力检测黑名单", len(dpElect.BpBlackList.BlackList))
	/*	for i, v := range dpElect.BpBlackList.BlackList {
		log.Trace("动态选举方案", "序号", i, "账户", v.Address.String(), "惩罚周期", v.ProhibitCycleCounter)
	}*/

	matrixstate.SetElectDynamicPollingInfo(stateDb, eleDpi)
	dpElect.DecrementBpSlashCount()
	matrixstate.SetBasePowerBlackList(stateDb, dpElect.BpBlackList)
	minerResult := self.transferMinerResult(mmrerm.SeqNum, chosedNodes)
	return minerResult
}

func (self *layeredDdp) CreateDpElect(stateDb *state.StateDBManage, mmrerm *mc.MasterMinerReElectionReqMsg) *support.ElectDP {
	bpBlackList, err := matrixstate.GetBasePowerBlackList(stateDb)
	if nil != err {
		log.Crit("动态选举方案", "获取算力黑名单错误", err)
	}
	if bpBlackList == nil {
		log.Crit("动态选举方案", "读取算力黑名单为空", "")
	}
	dpElect := support.NewDpElection(mmrerm, bpBlackList)
	return dpElect
}

func (self *layeredDdp) getEleDpi(stateDb *state.StateDBManage, mmrerm *mc.MasterMinerReElectionReqMsg) *mc.ElectDynamicPollingInfo {
	eleDpi, err := matrixstate.GetElectDynamicPollingInfo(stateDb)
	if nil != err {
		log.Crit("动态选举方案", "读取动态轮询信息失败，错误", err)
		//return nil
	}
	if eleDpi == nil {
		log.Crit("动态选举方案", "读取动态轮询信息为空", "")
	}

	log.Info("动态选举方案", "读取状态树,轮次", eleDpi.Seq, "高度", eleDpi.Number, "矿工选举个数", eleDpi.MinerNum, "候选节点个数", len(eleDpi.CandidateList))
	return eleDpi
}

func (self *layeredDdp) upDateEleDpi(dpElect *support.ElectDP, eleDpi *mc.ElectDynamicPollingInfo) {
	eleDpi.MinerNum = calcMinerNum(uint64(len(dpElect.DepositNode)), dpElect.EleCfg.MinerNum)
	addr := make([]common.Address, 0, 1024)
	for _, v := range dpElect.DepositNode {
		addr = append(addr, v.Address)
	}
	eleDpi.CandidateList = addr
	eleDpi.Seq++
	eleDpi.Number = dpElect.Num
	log.Trace("动态选举方案", "生成下一轮数据", eleDpi.Seq, "矿工选举数目", eleDpi.MinerNum, "高度", eleDpi.Number)

}

func (self *layeredDdp) transferMinerResult(Number uint64, HasChosedNode []common.Address) *mc.MasterMinerReElectionRsp {
	minerResult := &mc.MasterMinerReElectionRsp{}
	minerResult.SeqNum = Number
	for k, v := range HasChosedNode {
		minerResult.MasterMiner = append(minerResult.MasterMiner, support.MakeElectNode(v, k, 1, common.VIP_Nil, common.RoleMiner))
	}
	return minerResult
}

func (self *layeredDdp) minerdp(eleDpi *mc.ElectDynamicPollingInfo, dpElect *support.ElectDP) (choosedNodes []common.Address) {
	//更新选举信息
	if self.canGenNewSeq(eleDpi) {
		log.Warn("动态选举方案", "进入下一轮序号", eleDpi.Seq, "矿工选举数目", eleDpi.MinerNum, "高度", eleDpi.Number)
		self.upDateEleDpi(dpElect, eleDpi)
		dpElect.UpdateSeq = true
	}
	usableNodeList := dpElect.GetUsableNodeList(eleDpi.CandidateList, nil)
	if 0 != len(usableNodeList) {
		//log.Error("动态选举方案", "无可用节点参与选举,序号", eleDpi.Seq, "矿工选举数目", eleDpi.MinerNum, "高度", eleDpi.Number)
		choosedNodes = self.enterElect(dpElect, usableNodeList, eleDpi, uint64(eleDpi.MinerNum))
	}
	if dpElect.ChosedNum == eleDpi.MinerNum {
		return choosedNodes
	}
	//更新候选列表
	usableNodeList = dpElect.GetUsableNodeList(nil, choosedNodes) //使用第二轮的节点，移除已选出的节点, 生成初选列表

	if self.canEnterNextSeq(dpElect, eleDpi, uint64(len(usableNodeList))) { //判断是否进入第二轮选举
		//使用当前轮次的选举信息选举
		//第二轮选举
		preNum := eleDpi.MinerNum
		self.upDateEleDpi(dpElect, eleDpi)
		dpElect.UpdateSeq = true
		choosedNodes = append(choosedNodes, self.enterElect(dpElect, usableNodeList, eleDpi, preNum-dpElect.ChosedNum)...)
	}
	return choosedNodes
}

func (self *layeredDdp) enterElect(dpElect *support.ElectDP, newCandidateList []common.Address, eleDpi *mc.ElectDynamicPollingInfo, chooseNum uint64) (choosedNodes []common.Address) {

	choosedNodes = self.electMiner(newCandidateList, chooseNum, dpElect.RandSeed)
	//生成下一轮的选举信息，完全进入下一轮
	eleDpi.CandidateList = self.removeChosedNode(eleDpi.CandidateList, choosedNodes) //更新选举信息，把选出的节点移除
	dpElect.AddChosedNum(uint64(len(choosedNodes)))

	return choosedNodes
}

func (self *layeredDdp) addChosedNum(dpElect *support.ElectDP, hasChosedNum uint64) {
	dpElect.ChosedNum = dpElect.ChosedNum + hasChosedNum
}

func (self *layeredDdp) canGenNewSeq(eleDpi *mc.ElectDynamicPollingInfo) bool {
	return eleDpi.MinerNum == 0 || len(eleDpi.CandidateList) == 0
}

func (self *layeredDdp) canEnterNextSeq(dpElect *support.ElectDP, eleDpi *mc.ElectDynamicPollingInfo, candidateLen uint64) bool {
	return dpElect.ChosedNum < eleDpi.MinerNum && 0 != candidateLen
}

func (self *layeredDdp) canEndElect(dpElect *support.ElectDP, eleDpi *mc.ElectDynamicPollingInfo) bool {
	if dpElect.ChosedNum == eleDpi.MinerNum {
		log.Info("动态选举方案", "轮次", eleDpi.Seq, "高度", eleDpi.Number, "矿工选举个数", eleDpi.MinerNum, "选出节点个数", dpElect.ChosedNum)
		return true
	}
	if len(eleDpi.CandidateList) == 0 {
		log.Info("动态选举方案", "候选节点为0，本轮选举结束，轮次", eleDpi.Seq)
		return true
	}
	return false
}

func (self *layeredDdp) updateElectInfo(eleDpi *mc.ElectDynamicPollingInfo, dpElect *support.ElectDP, choosedNodes []common.Address, nextSeqHasChosedNode []common.Address) {
	eleDpi.CandidateList = append(eleDpi.CandidateList, choosedNodes...)
	choosedNodes = append(choosedNodes, nextSeqHasChosedNode...)
	eleDpi.MinerNum = calcMinerNum(uint64(len(dpElect.DepositNode)), dpElect.EleCfg.MinerNum)
	eleDpi.Seq = eleDpi.Seq + 1
	dpElect.DecrementBpSlashCount()
	log.Info("动态选举方案", "本轮选举结束，计算下一轮轮询数据,轮次", eleDpi.Seq, "高度", eleDpi.Number, "矿工选举个数", eleDpi.MinerNum, "候选节点个数", len(eleDpi.CandidateList), "选出节点个数", dpElect.ChosedNum)

}

func (self *layeredDdp) getOutboundNode(dpElect *support.ElectDP, choosedNodes []common.Address) []common.Address {
	addr := make([]common.Address, 0)
	//更新候选列表,上一轮当前选出来的节点暂时从抵押列表移除
	for _, v := range dpElect.DepositNode {
		if !support.FindAddress(v.Address, choosedNodes) {
			addr = append(addr, v.Address)
		} else {
			//log.Info("动态选举方案", "移除上一轮选出来的节点，地址", v.Address)
		}
	}
	return addr
}
func (self *layeredDdp) electMiner(usableNodeList []common.Address, chooseNum uint64, RandSeed *mt19937.RandUniform) []common.Address {
	if uint64(len(usableNodeList)) <= chooseNum {
		return self.allElect(usableNodeList)
	} else {
		return self.randomElect(usableNodeList, chooseNum, RandSeed)
	}
}

func (self *layeredDdp) allElect(usableNodeList []common.Address) []common.Address {
	choosedNodes := make([]common.Address, 0)
	for i := 0; i < len(usableNodeList); i++ {
		choosedNodes = append(choosedNodes, usableNodeList[i])
	}

	return choosedNodes
}

func (self *layeredDdp) randomElect(usableNodeList []common.Address, chooseNum uint64, RandSeed *mt19937.RandUniform) []common.Address {
	usableLength := len(usableNodeList)
	choosedNodes := make([]common.Address, 0)
	for i := 0; i < int(chooseNum) && i < usableLength; i++ {
		randomData := uint64(RandSeed.Uniform(0, float64(^uint64(0))))
		index := randomData % (uint64(len(usableNodeList)))
		choosedNodes = append(choosedNodes, usableNodeList[index])
		usableNodeList = append(usableNodeList[:index], usableNodeList[index+1:]...)
	}

	return choosedNodes
}

func (self *layeredDdp) removeChosedNode(CandidateList []common.Address, choosedNodes []common.Address) []common.Address {
	newCandidateList := make([]common.Address, 0)
	for i := 0; i < len(CandidateList); i++ {
		if !support.FindAddress(CandidateList[i], choosedNodes) {
			newCandidateList = append(newCandidateList, CandidateList[i])
		}

	}
	return newCandidateList

}

func calcMinerNum(depsitnodeNum uint64, minMinerNum uint16) uint64 {
	if depsitnodeNum > minMinersBase {
		return uint64(minMinerNum) + (depsitnodeNum-minMinersBase)/MinersFactor*addMinesNum
	} else {
		return uint64(minMinerNum)
	}

}

//cal the sum of deposit, return the sum value
func (self *layeredDdp) ValidatorTopGen(mvrerm *mc.MasterValidatorReElectionReqMsg, stateDb *state.StateDBManage) *mc.MasterValidatorReElectionRsq {
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

func (self *layeredDdp) ToPoUpdate(allNative support.AllNative, topoG *mc.TopologyGraph) []mc.Alternative {
	return support.ToPoUpdate(allNative, topoG)
}

func (self *layeredDdp) PrimarylistUpdate(Q0, Q1, Q2 []mc.TopologyNodeInfo, online mc.TopologyNodeInfo, flag int) ([]mc.TopologyNodeInfo, []mc.TopologyNodeInfo, []mc.TopologyNodeInfo) {
	return support.PrimarylistUpdate(Q0, Q1, Q2, online, flag)
}
