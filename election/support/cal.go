// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package support

import (
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/mc"
)

func MakeElectNode(address common.Address, Pos int, Stock int, VIPLevel common.VIPRoleType, Type common.RoleType) mc.ElectNodeInfo {
	return mc.ElectNodeInfo{
		Account:  address,
		Position: uint16(Pos),
		Stock:    uint16(Stock),
		VIPLevel: VIPLevel,
		Type:     Type,
	}
}

func MakeMinerAns(chosed []Strallyint, seqnum uint64) *mc.MasterMinerReElectionRsp {
	minerResult := &mc.MasterMinerReElectionRsp{}
	minerResult.SeqNum = seqnum
	for k, v := range chosed {
		minerResult.MasterMiner = append(minerResult.MasterMiner, MakeElectNode(v.Addr, k, v.Value, common.VIP_Nil, common.RoleMiner))
		//log.Info(ModuleLogName, "Master", v.Addr.String())
	}
	return minerResult
}

func MakeValidatoeTopGenAns(seqnum uint64, master []Strallyint, backup []Strallyint, candiate []Strallyint) *mc.MasterValidatorReElectionRsq {
	ans := &mc.MasterValidatorReElectionRsq{
		SeqNum: seqnum,
	}

	for _, v := range master {
		ans.MasterValidator = append(ans.MasterValidator, MakeElectNode(v.Addr, len(ans.MasterValidator), v.Value, v.VIPLevel, common.RoleValidator))
	}
	for _, v := range backup {
		ans.BackUpValidator = append(ans.BackUpValidator, MakeElectNode(v.Addr, len(ans.BackUpValidator), v.Value, v.VIPLevel, common.RoleBackupValidator))
	}
	for _, v := range candiate {
		ans.CandidateValidator = append(ans.CandidateValidator, MakeElectNode(v.Addr, len(ans.CandidateValidator), v.Value, v.VIPLevel, common.RoleCandidateValidator))
	}
	return ans
}
