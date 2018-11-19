// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package stock

import (
	"math/big"

	"github.com/matrix/go-matrix/baseinterface"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/vm"
	"github.com/matrix/go-matrix/election/support"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
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
	return support.MinerTopGen(mmrerm)
}

func (self *StockElect) ValidatorTopGen(mvrerm *mc.MasterValidatorReElectionReqMsg) *mc.MasterValidatorReElectionRsq {
	log.INFO("选举种子", "验证者拓扑生成", len(mvrerm.ValidatorList))
	ValidatorElectMap := make(map[string]vm.DepositDetail)
	for i, item := range mvrerm.ValidatorList {
		ValidatorElectMap[item.NodeID.String()] = item
		//todo: panic
		if item.Deposit == nil {
			mvrerm.ValidatorList[i].Deposit = big.NewInt(support.DefaultDeposit)
		}
		if item.WithdrawH == nil {
			mvrerm.ValidatorList[i].WithdrawH = big.NewInt(support.DefaultWithdrawH)
		}
		if item.OnlineTime == nil {
			mvrerm.ValidatorList[i].OnlineTime = big.NewInt(support.DefaultOnlineTime)
		}
	}

	ValidatorEleRs := new(mc.MasterValidatorReElectionRsq)
	ValidatorEleRs.SeqNum = mvrerm.SeqNum

	var a, b, c []support.Strallyint
	var value []support.Stf
	if len(mvrerm.FoundationValidatoeList) == 0 {
		value = support.CalcAllValueFunction(mvrerm.ValidatorList)
		a, b, c = support.ValNodesSelected(value, mvrerm.RandSeed.Int64(), 11, 5, 0) //mvrerm.RandSeed.Int64(), 11, 5, 0) //0x12217)
	} else {
		value = support.CalcAllValueFunction(mvrerm.ValidatorList)
		valuefound := support.CalcAllValueFunction(mvrerm.FoundationValidatoeList)
		a, b, c = support.ValNodesSelected(value, mvrerm.RandSeed.Int64(), 11, 5, len(mvrerm.FoundationValidatoeList)) //0x12217)
		a = support.CommbineFundNodesAndPricipal(value, valuefound, a, 0.25, 4.0)
	}

	for index, item := range a {
		tmp := ValidatorElectMap[item.Nodeid]
		var ToG mc.TopologyNodeInfo
		ToG.Account = tmp.Address
		ToG.Position = uint16(index)
		ToG.Type = common.RoleValidator
		ToG.Stock = uint16(item.Value)
		ValidatorEleRs.MasterValidator = append(ValidatorEleRs.MasterValidator, ToG)
	}

	for index, item := range b {
		tmp := ValidatorElectMap[item.Nodeid]
		var ToG mc.TopologyNodeInfo
		ToG.Account = tmp.Address
		ToG.Position = uint16(index)
		ToG.Type = common.RoleBackupValidator
		ToG.Stock = uint16(item.Value)
		ValidatorEleRs.BackUpValidator = append(ValidatorEleRs.BackUpValidator, ToG)
	}

	for index, item := range c {
		tmp := ValidatorElectMap[item.Nodeid]
		var ToG mc.TopologyNodeInfo
		ToG.Account = tmp.Address
		ToG.Position = uint16(index)
		ToG.Type = common.RoleCandidateValidator
		ToG.Stock = uint16(item.Value)
		ValidatorEleRs.CandidateValidator = append(ValidatorEleRs.CandidateValidator, ToG)
	}
	return ValidatorEleRs
}

func (self *StockElect) ToPoUpdate(offline []common.Address, allNative support.AllNative, topoG *mc.TopologyGraph) []mc.Alternative {

	return support.ToPoUpdate(offline, allNative, topoG)
}

func (self *StockElect) PrimarylistUpdate(Q0, Q1, Q2 []mc.TopologyNodeInfo, online mc.TopologyNodeInfo, flag int) ([]mc.TopologyNodeInfo, []mc.TopologyNodeInfo, []mc.TopologyNodeInfo) {
	return support.PrimarylistUpdate(Q0, Q1, Q2, online, flag)
}
