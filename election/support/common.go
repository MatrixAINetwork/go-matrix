// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package support

import (
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
)

func DelIndex(native AllNative, flag int) AllNative {
	ans := AllNative{
		Master:    native.Master,
		BackUp:    native.BackUp,
		Candidate: native.Candidate,
		ElectInfo: native.ElectInfo,
	}
	switch flag {
	case 1:
		for k, v := range native.MasterQ {
			if k == 0 {
				continue
			}
			ans.MasterQ = append(ans.MasterQ, v)
		}
		ans.BackUpQ = append(ans.BackUpQ, native.BackUpQ[0:]...)
		ans.CandidateQ = append(ans.CandidateQ, native.CandidateQ[0:]...)
		return ans
	case 2:
		for k, v := range native.BackUpQ {
			if k == 0 {
				continue
			}
			ans.BackUpQ = append(ans.BackUpQ, v)
		}
		ans.MasterQ = append(ans.MasterQ, native.MasterQ[0:]...)
		ans.CandidateQ = append(ans.CandidateQ, native.CandidateQ[0:]...)
		return ans
	case 3:
		for k, v := range native.CandidateQ {
			if k == 0 {
				continue
			}
			ans.CandidateQ = append(ans.CandidateQ, v)
		}
		ans.MasterQ = append(ans.MasterQ, native.MasterQ[0:]...)
		ans.BackUpQ = append(ans.BackUpQ, native.BackUpQ[0:]...)
		return ans
	default:
		return ans

	}
}
func GetNodeFromBuff(native AllNative) (AllNative, common.Address) {
	if len(native.MasterQ) > 0 {
		return DelIndex(native, 1), native.MasterQ[0]
	}
	if len(native.BackUpQ) > 0 {
		return DelIndex(native, 2), native.BackUpQ[0]
	}
	if len(native.CandidateQ) > 0 {
		return DelIndex(native, 3), native.CandidateQ[0]
	}
	return native, common.Address{}
}
func CalcGradeDependAddr(addr common.Address, native AllNative) int {
	for _, v := range native.Master {
		if v.Account.Equal(addr) {
			return 3
		}
	}

	for _, v := range native.BackUp {
		if v.Account.Equal(addr) {
			return 2
		}
	}
	for _, v := range native.Candidate {
		if v.Account.Equal(addr) {
			return 1
		}
	}
	return 0
}

func BackUpdata(top *mc.TopologyGraph, mapp map[uint16]common.Address, native AllNative) uint16 {
	ChoseStatus := -1
	AimPosition := uint16(0)
	for _, v := range top.NodeList {
		types := common.GetRoleTypeFromPosition(v.Position)
		if _, ok := mapp[v.Position]; ok == false {
			continue
		}
		if types != common.RoleBackupValidator {
			continue
		}
		AddrGrade := CalcGradeDependAddr(v.Account, native)
		//	fmt.Println("AddrGrade",AddrGrade,"v.Account",v.Account.String())
		if AddrGrade > ChoseStatus {
			AimPosition = v.Position
			ChoseStatus = AddrGrade
		}
	}
	//fmt.Println("AimPositon",AimPosition)
	return AimPosition
}
func KInTop(aim uint16, topoG *mc.TopologyGraph) bool {
	for _, v := range topoG.NodeList {
		if v.Position == aim {
			return true
		}
	}
	return false
}

func ToPoUpdate(allNative AllNative, topoG *mc.TopologyGraph) []mc.Alternative {
	ans := []mc.Alternative{}
	mapMaster := make(map[uint16]common.Address)
	mapBackup := make(map[uint16]common.Address)

	//fmt.Println("len topG.NodeList",len(topoG.NodeList))
	//for _,v:=range topoG.NodeList{
	//	fmt.Println(v.Position,v.Account.String())
	//}
	for _, v := range topoG.NodeList {
		//fmt.Println("v.Pos",v.Position,"v.addr",v.Account.String())
		types := common.GetRoleTypeFromPosition(v.Position)
		if types == common.RoleValidator {
			mapMaster[v.Position] = v.Account
		}
		if types == common.RoleBackupValidator {
			mapBackup[v.Position] = v.Account
		}
	}
	//fmt.Println("mapMaster",mapMaster,"len",len(mapMaster))
	//fmt.Println("mapBackup",mapBackup,"len",len(mapBackup))
	for index := 0; index < int(allNative.ElectInfo.ValidatorNum); index++ { //用一级在线去补
		k := common.GeneratePosition(uint16(index), common.ElectRoleValidator)
		_, ok := mapMaster[k]
		if ok == true {
			//	fmt.Println("该位置已存在",k)
			continue
		}

		trans := BackUpdata(topoG, mapBackup, allNative)
		if trans == 0 {
			//fmt.Println("已空")
			continue
		}
		if _, ok := mapBackup[trans]; ok == false {
			//	fmt.Println("mapBack没有",trans)
			continue
		}

		ans = append(ans, mc.Alternative{
			A:        mapBackup[trans],
			Position: k,
		})
		//	fmt.Println("一级在线去补",mapBackup[trans].String(),k)
		mapMaster[k] = mapBackup[trans]
		delete(mapBackup, trans)

	}
	for index := 0; index < int(allNative.ElectInfo.ValidatorNum); index++ { //用buff去补
		k := common.GeneratePosition(uint16(index), common.ElectRoleValidator)
		_, ok := mapMaster[k]
		if ok == true {
			continue
		}
		var addr common.Address
		allNative, addr = GetNodeFromBuff(allNative)
		if addr.Equal(common.Address{}) {
			continue
		}
		ans = append(ans, mc.Alternative{
			A:        addr,
			Position: k,
		})
		mapMaster[k] = addr
		//	fmt.Println("用buff去补",addr,k)
	}
	for index := 0; index < int(allNative.ElectInfo.BackValidator); index++ {
		k := common.GeneratePosition(uint16(index), common.ElectRoleValidatorBackUp)
		_, ok := mapBackup[k]
		if ok == true {
			continue
		}
		var addr common.Address
		allNative, addr = GetNodeFromBuff(allNative)
		if addr.Equal(common.Address{}) {
			continue
		}
		ans = append(ans, mc.Alternative{
			A:        addr,
			Position: k,
		})
		mapBackup[k] = addr
	}

	for index := 0; index < int(allNative.ElectInfo.ValidatorNum); index++ { //算一级下线
		k := common.GeneratePosition(uint16(index), common.ElectRoleValidator)
		if KInTop(k, topoG) == false {
			log.Trace(ModuleLogName, "一级 该点不在顶层内", "不处理")
			continue
		}
		if _, ok := mapMaster[k]; ok == false {
			//fmt.Println("该店直接下线-一级")
			ans = append(ans, mc.Alternative{
				A:        common.Address{},
				Position: k,
			})
		}
	}
	for index := 0; index < int(allNative.ElectInfo.BackValidator); index++ { //算二级下线
		k := common.GeneratePosition(uint16(index), common.ElectRoleValidatorBackUp)
		//fmt.Println("222222222---k",k)
		if KInTop(k, topoG) == false {
			log.Trace(ModuleLogName, "二级 该点不在顶层内", "不处理")
			continue
		}
		if _, ok := mapBackup[k]; ok == false {
			//	fmt.Println("该店直接下线-二级")
			ans = append(ans, mc.Alternative{
				A:        common.Address{},
				Position: k,
			})
		}
	}

	return ans

}

func PrimarylistUpdate(Q0, Q1, Q2 []mc.TopologyNodeInfo, online mc.TopologyNodeInfo, flag int) ([]mc.TopologyNodeInfo, []mc.TopologyNodeInfo, []mc.TopologyNodeInfo) {
	log.Info("Elector PrimarylistUpdate")
	if flag == 0 {
		var tQ0 []mc.TopologyNodeInfo
		tQ0 = append(tQ0, online)
		tQ0 = append(tQ0, Q0...)
		Q0 = tQ0
	}

	if flag == 1 {
		var tQ1 []mc.TopologyNodeInfo
		tQ1 = append(tQ1, Q1...)
		tQ1 = append(tQ1, online)
		Q1 = tQ1
	}

	if flag == 2 {
		var tQ2 []mc.TopologyNodeInfo
		tQ2 = append(tQ2, Q2...)
		tQ2 = append(tQ2, online)
		Q2 = tQ2
	}
	return Q0, Q1, Q2
}
