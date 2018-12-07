// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package layered

import (
	//"fmt"

	"math/big"

	"github.com/matrix/go-matrix/baseinterface"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/vm"
	"github.com/matrix/go-matrix/election/support"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
)

type layered struct {
}

func init() {
	baseinterface.RegElectPlug("layered", RegInit)
}

func RegInit() baseinterface.ElectionInterface {
	return &layered{}
}

func (self *layered) MinerTopGen(mmrerm *mc.MasterMinerReElectionReqMsg) *mc.MasterMinerReElectionRsp {
	log.INFO("分层方案", "矿工拓扑生成", len(mmrerm.MinerList))
	for k, v := range mmrerm.MinerList {
		if v.Deposit == nil {
			mmrerm.MinerList[k].Deposit = big.NewInt(100)
		}
		if v.WithdrawH == nil {
			mmrerm.MinerList[k].WithdrawH = big.NewInt(100)
		}
		if v.OnlineTime == nil {
			mmrerm.MinerList[k].OnlineTime = big.NewInt(100)
		}
	}
	return support.MinerTopGen(mmrerm)

}

func (self *layered) ValidatorTopGen(mvrerm *mc.MasterValidatorReElectionReqMsg) *mc.MasterValidatorReElectionRsq {
	for k, v := range mvrerm.ValidatorList {
		if v.Deposit == nil {
			mvrerm.ValidatorList[k].Deposit = big.NewInt(100)
		}
		if v.WithdrawH == nil {
			mvrerm.ValidatorList[k].WithdrawH = big.NewInt(100)
		}
		if v.OnlineTime == nil {
			mvrerm.ValidatorList[k].OnlineTime = big.NewInt(100)
		}
	}

	log.INFO("分层方案", "验证者拓扑生成", len(mvrerm.ValidatorList))
	ValidatorTopGen := mc.MasterValidatorReElectionRsq{}

	ChoiceToMaster := make(map[common.Address]int, 0)

	InitMapList := make(map[string]vm.DepositDetail, 0)

	for _, v := range mvrerm.ValidatorList {
		InitMapList[v.NodeID.String()] = v
	}

	QuotaArray := CalEchelonNum(mvrerm.ValidatorList)
	for k, v := range QuotaArray {
		log.ERROR("quotaarray", "等级", k, "长度", len(v), "data", v)
	}
	//fmt.Println(len(FirstQuota), len(SecondQuota))
	//fmt.Println("QuotaArray", len(QuotaArray[0]))

	sumCount := 0
	//fmt.Println("EchelonArrary", EchelonArrary)
	for k, v := range QuotaArray {
		sumCount += common.EchelonArrary[k].Quota
		if len(v) > common.EchelonArrary[k].Quota {
			v = sortByDepositAndUptime(v, mvrerm.RandSeed)
		}
		for _, vv := range v {
			tempNodeInfo := mc.TopologyNodeInfo{
				Account:  vv.Address,
				Position: uint16(len(ValidatorTopGen.MasterValidator)),
				Stock:    DefauleStock,
				Type:     common.RoleValidator,
			}
			ValidatorTopGen.MasterValidator = append(ValidatorTopGen.MasterValidator, tempNodeInfo)
			ChoiceToMaster[vv.Address] = 1
			if len(ValidatorTopGen.MasterValidator) >= sumCount || len(ValidatorTopGen.MasterValidator) >= common.MasterValidatorNum {
				//fmt.Println(len(ValidatorTopGen.MasterValidator), sumCount, len(ValidatorTopGen.MasterValidator), common.MasterValidatorNum)
				break
			}
		}
	}

	NowList := []vm.DepositDetail{}
	for _, v := range mvrerm.ValidatorList {
		_, ok := ChoiceToMaster[v.Address]
		if ok {
			continue
		}
		NowList = append(NowList, v)
	}
	weight := GetValueByDeposit(NowList)
	//fmt.Println("weight", len(weight))
	//fmt.Println("zzz", support.M-len(ValidatorTopGen.MasterValidator))

	a, b, c := support.ValNodesSelected(weight, mvrerm.RandSeed.Int64(), support.M-len(ValidatorTopGen.MasterValidator), 5, 0) //mvrerm.RandSeed.Int64(), 11, 5, 0) //0x12217)

	//fmt.Println(len(a), len(b), len(c))
	for _, v := range a {
		tempNodeInfo := mc.TopologyNodeInfo{
			Account:  InitMapList[v.Nodeid].Address,
			Position: uint16(len(ValidatorTopGen.MasterValidator)),
			Stock:    DefauleStock,
			Type:     common.RoleValidator,
		}
		ValidatorTopGen.MasterValidator = append(ValidatorTopGen.MasterValidator, tempNodeInfo)
	}
	for _, v := range b {
		tempNodeInfo := mc.TopologyNodeInfo{
			Account:  InitMapList[v.Nodeid].Address,
			Position: uint16(len(ValidatorTopGen.BackUpValidator)),
			Stock:    DefauleStock,
			Type:     common.RoleBackupValidator,
		}
		ValidatorTopGen.BackUpValidator = append(ValidatorTopGen.BackUpValidator, tempNodeInfo)
	}
	for _, v := range c {
		tempNodeInfo := mc.TopologyNodeInfo{
			Account:  InitMapList[v.Nodeid].Address,
			Position: uint16(len(ValidatorTopGen.CandidateValidator)),
			Stock:    DefauleStock,
			Type:     common.RoleCandidateValidator,
		}
		ValidatorTopGen.CandidateValidator = append(ValidatorTopGen.CandidateValidator, tempNodeInfo)
	}

	return &ValidatorTopGen

}

func (self *layered) ToPoUpdate(offline []common.Address, allNative support.AllNative, topoG *mc.TopologyGraph) []mc.Alternative {
	return support.ToPoUpdate(offline, allNative, topoG)
}

func (self *layered) PrimarylistUpdate(Q0, Q1, Q2 []mc.TopologyNodeInfo, online mc.TopologyNodeInfo, flag int) ([]mc.TopologyNodeInfo, []mc.TopologyNodeInfo, []mc.TopologyNodeInfo) {
	return support.PrimarylistUpdate(Q0, Q1, Q2, online, flag)
}
