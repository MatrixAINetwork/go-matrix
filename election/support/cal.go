// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package support

import (
	"math/big"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/vm"
	"github.com/matrix/go-matrix/mc"
)

func MinerTopGen(mmrerm *mc.MasterMinerReElectionReqMsg) *mc.MasterMinerReElectionRsp {
	MinerElectMap := make(map[string]vm.DepositDetail)
	for i, item := range mmrerm.MinerList {
		//				MinerElectMap[string(item.Account[:])] = item
		MinerElectMap[item.NodeID.String()] = item
		if item.Deposit == nil {
			mmrerm.MinerList[i].Deposit = big.NewInt(DefaultDeposit)
		}
		if item.WithdrawH == nil {
			mmrerm.MinerList[i].WithdrawH = big.NewInt(DefaultWithdrawH)
		}
		if item.OnlineTime == nil {
			mmrerm.MinerList[i].OnlineTime = big.NewInt(DefaultOnlineTime)
		}
	}

	value := CalcAllValueFunction(mmrerm.MinerList)

	a, b := MinerNodesSelected(value, mmrerm.RandSeed.Int64(), N) //Ele.Engine(value, mmrerm.RandSeed.Int64()) //0x12217)
	/*
		for index, item := range a {
			fmt.Println(index, "---", item.Nodeid, "===")
		}
		for index, item := range b {
			fmt.Println(index, "---", item.Nodeid, "===")
		}
	*/
	MinerEleRs := new(mc.MasterMinerReElectionRsp)
	MinerEleRs.SeqNum = mmrerm.SeqNum

	for index, item := range a {
		//	fmt.Println(item.Nodeid, []byte(item.Nodeid))
		tmp := MinerElectMap[item.Nodeid]
		var ToG mc.TopologyNodeInfo
		ToG.Account = tmp.Address
		ToG.Position = uint16(index)
		ToG.Type = common.RoleMiner
		ToG.Stock = uint16(item.Value)
		MinerEleRs.MasterMiner = append(MinerEleRs.MasterMiner, ToG)
	}

	for index, item := range b {
		tmp := MinerElectMap[item.Nodeid]
		var ToG mc.TopologyNodeInfo
		ToG.Account = tmp.Address
		//				ToG.OnlineState = true
		ToG.Position = uint16(index)
		ToG.Type = common.RoleBackupMiner
		ToG.Stock = uint16(item.Value)
		MinerEleRs.BackUpMiner = append(MinerEleRs.BackUpMiner, ToG)
	}
	return MinerEleRs
}
