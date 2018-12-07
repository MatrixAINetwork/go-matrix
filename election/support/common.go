// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package support

import (
	"fmt"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
)

func locate(addr common.Address, top *mc.TopologyGraph) int {
	for _, v := range top.NodeList {
		if v.Account != addr {
			continue
		}
		if v.Type == common.RoleValidator {
			return 0
		}
		if v.Type == common.RoleBackupValidator {
			return 1
		}
		return -1
	}
	return -1

}

func SloveZeroOffline(pos uint16, alter []mc.Alternative, native AllNative, top *mc.TopologyGraph, flag int, offline []common.Address) ([]mc.Alternative, AllNative) {

	for k, v := range native.MasterQ { //0级缓存
		temp := mc.Alternative{
			A:        v,
			Position: pos,
		}
		alter = append(alter, temp)
		native.MasterQ = append(native.MasterQ[:k], native.MasterQ[k+1:]...)
		log.INFO("选举计算阶段-0", "当前掉线的节点是", pos, "用0即缓存里的节点去顶替 缓存的节点", v.String())
		return alter, native
	}

	for kIndex, v := range top.NodeList { //1级在线
		if v.Account.Equal(common.Address{}) {
			//	fmt.Println("已经被顶替了，不能使用")
			continue
		}
		if findAddress(v.Account, offline) {
			continue
		}
		if v.Type == common.RoleBackupValidator {
			temp := mc.Alternative{
				A:        v.Account,
				Position: pos,
			}
			alter = append(alter, temp)
			top.NodeList[kIndex].Account = common.Address{}
			log.INFO("选举计算节点-0", "当前掉线的节点是", pos, "用一级在线的点去顶替", v.Account.String())
			for kB, vB := range native.BackUpQ { //1级缓存
				temp := mc.Alternative{
					A:        vB,
					Position: v.Position,
				}
				alter = append(alter, temp)
				native.BackUpQ = append(native.BackUpQ[:kB], native.BackUpQ[kB+1:]...)
				log.INFO("选举计算节点-0", "当前掉线的节点是", pos, "用一级在线的点去顶替", v.Account.String(), "一级缓存里有缓存 顶替一级在线的节点", vB.String())
				return alter, native
			}
			for kC, vC := range native.CandidateQ { //2级缓存
				temp := mc.Alternative{
					A:        vC,
					Position: v.Position,
				}
				alter = append(alter, temp)
				native.CandidateQ = append(native.CandidateQ[:kC], native.CandidateQ[kC+1:]...)
				log.INFO("选举计算节点-0", "当前掉线的节点是", pos, "用一级在线的点去顶替", v.Account.String(), "二级缓存里有缓存 顶替一级在线的节点", vC.String())
				return alter, native
			}
			temp = mc.Alternative{
				A: common.Address{},

				Position: v.Position,
			}

			alter = append(alter, temp)
			log.INFO("选举计算节点-0", "当前掉线的节点是", pos, "用一级在线的点去顶替", v.Account.String(), "无缓存可顶 原一级在线的顶点的位置直接删除", "")
			return alter, native
		}
	}
	for k, v := range native.BackUpQ { //1级缓存
		temp := mc.Alternative{
			A: v,

			Position: pos,
		}
		alter = append(alter, temp)
		native.BackUpQ = append(native.BackUpQ[:k], native.BackUpQ[k+1:]...)
		log.INFO("选举计算节点-0", "当前掉线节点是", pos, "用一级缓存的点去顶替", v.String())
		return alter, native

	}
	for k, v := range native.CandidateQ { //2级缓存
		temp := mc.Alternative{
			A:        v,
			Position: pos,
		}
		alter = append(alter, temp)
		native.CandidateQ = append(native.CandidateQ[:k], native.CandidateQ[k+1:]...)
		log.INFO("选举计算节点-0", "当前掉线节点是", pos, "用二级缓存的点去顶替", v.String())
		return alter, native
	}
	//该位置无人可补充
	temp := mc.Alternative{
		A: common.Address{},

		Position: pos,
	}
	if flag == IsOffline {
		alter = append(alter, temp)
	}

	log.INFO("选举计算节点-0", "当前掉线节点是", pos, "无候选节点可顶替 直接删除该位置", pos)
	return alter, native

}
func SloveFirstOffline(pos uint16, alter []mc.Alternative, native AllNative, flag int) ([]mc.Alternative, AllNative) {

	for k, v := range native.MasterQ { //0级缓存
		temp := mc.Alternative{
			A:        v,
			Position: pos,
		}
		alter = append(alter, temp)
		native.MasterQ = append(native.MasterQ[:k], native.MasterQ[k+1:]...)
		log.INFO("选举计算阶段-1", "当前掉线的节点是", pos, "用0即缓存里的节点去顶替 缓存的节点", v.String())
		return alter, native
	}

	for k, v := range native.BackUpQ { //1级缓存
		temp := mc.Alternative{
			A:        v,
			Position: pos,
		}
		alter = append(alter, temp)
		native.BackUpQ = append(native.BackUpQ[:k], native.BackUpQ[k+1:]...)
		log.INFO("选举计算节点-1", "当前掉线节点是", pos, "用一级缓存的点去顶替", v.String())
		return alter, native
	}
	for k, v := range native.CandidateQ { //2级缓存
		temp := mc.Alternative{
			A:        v,
			Position: pos,
		}
		alter = append(alter, temp)
		native.CandidateQ = append(native.CandidateQ[:k], native.CandidateQ[k+1:]...)
		log.INFO("选举计算节点-1", "当前掉线节点是", pos, "用二级缓存的点去顶替", v.String())
		return alter, native
	}
	temp := mc.Alternative{
		A:        common.Address{},
		Position: pos,
	}
	log.INFO("选举计算节点-1", "当前掉线节点是", pos, "无节点可顶替 直接删除该位置", "")
	if flag == IsOffline {
		alter = append(alter, temp)
	}
	return alter, native
}

func findAddress(addr common.Address, aim []common.Address) bool {
	for _, v := range aim {
		if v == addr {
			return true
		}
	}
	return false
}
func findAddressInTop(addr common.Address, aim []mc.TopologyNodeInfo) bool {
	for _, v := range aim {
		if v.Account == addr {
			return true
		}
	}
	return false
}

const (
	IsOffline  = 1
	NotOffline = 0
)

func Slove(topG *mc.TopologyGraph, native AllNative) *mc.TopologyGraph {

	backInMaster := []mc.TopologyNodeInfo{}
	backInitList := []mc.TopologyNodeInfo{}
	backNoInitList := []mc.TopologyNodeInfo{}
	for _, v := range topG.NodeList {
		if v.Type != common.RoleBackupValidator {
			continue
		}
		if findAddressInTop(v.Account, native.Master) {
			backInMaster = append(backInMaster, v)
		} else if findAddressInTop(v.Account, native.BackUp) {
			backInitList = append(backInitList, v)
		} else {
			backNoInitList = append(backNoInitList, v)
		}
	}
	for _, v := range backInitList {
		backInMaster = append(backInMaster, v)
	}
	for _, v := range backNoInitList {
		backInMaster = append(backInMaster, v)
	}

	aimList := []mc.TopologyNodeInfo{}
	flag := 0
	for _, v := range topG.NodeList {
		if v.Type != common.RoleBackupValidator {
			aimList = append(aimList, v)
			continue
		}
		if flag == 0 {
			aimList = append(aimList, backInMaster...)
			flag = 1
		}
	}
	ans := topG
	ans.NodeList = aimList

	return ans
}
func findAddr(addr common.Address, list []common.Address) bool {
	for _, v := range list {
		if v.Equal(addr) {
			return true
		}
	}
	return false
}
func DelIndex(native AllNative, flag int) AllNative {
	ans := AllNative{}
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
func BackUpdata(top *mc.TopologyGraph, mapp map[uint16]common.Address) uint16 {
	for _, v := range top.NodeList {
		types := common.GetRoleTypeFromPosition(v.Position)
		if _, ok := mapp[v.Position]; ok == false {
			continue
		}
		if types == common.RoleBackupValidator {
			return v.Position
		}
	}
	return 0
}
func KInTop(aim uint16, topoG *mc.TopologyGraph) bool {
	for _, v := range topoG.NodeList {
		if v.Position == aim {
			return true
		}
	}
	return false
}
func ToPoUpdate(offline []common.Address, allNative AllNative, topoG *mc.TopologyGraph) []mc.Alternative {
	ans := []mc.Alternative{}
	mapMaster := make(map[uint16]common.Address)
	mapBackup := make(map[uint16]common.Address)

	//fmt.Println("len topG.NodeList",len(topoG.NodeList))
	//for _,v:=range topoG.NodeList{
	//	fmt.Println(v.Position,v.Account.String())
	//}
	for _, v := range topoG.NodeList {
		if findAddr(v.Account, offline) {
			continue //删除节点
		}
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
	for index := 0; index < common.MasterValidatorNum; index++ { //用一级在线去补
		k := common.GeneratePosition(uint16(index), common.ElectRoleValidator)
		_, ok := mapMaster[k]
		if ok == true {
			//	fmt.Println("该位置已存在",k)
			continue
		}

		trans := BackUpdata(topoG, mapBackup)
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
	for index := 0; index < common.MasterValidatorNum; index++ { //用buff去补
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
	for index := 0; index < common.BackupValidatorNum; index++ {
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

	for index := 0; index < common.MasterValidatorNum; index++ { //算一级下线
		k := common.GeneratePosition(uint16(index), common.ElectRoleValidator)
		if KInTop(k, topoG) == false {
			fmt.Println("一级 该点不在顶层内", "不处理")
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
	for index := 0; index < common.BackupValidatorNum; index++ { //算二级下线
		k := common.GeneratePosition(uint16(index), common.ElectRoleValidatorBackUp)
		//fmt.Println("222222222---k",k)
		if KInTop(k, topoG) == false {
			fmt.Println("二级 该点不在顶层内", "不处理")
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
func ToPoUpdate_back(offline []common.Address, allNative AllNative, topoG *mc.TopologyGraph) []mc.Alternative {
	ans := []mc.Alternative{}
	mapMaster := make(map[uint16]common.Address)
	mapBackup := make(map[uint16]common.Address)

	top := Slove(topoG, allNative)

	for _, v := range top.NodeList {
		types := common.GetRoleTypeFromPosition(v.Position)
		if types == common.RoleValidator {
			mapMaster[v.Position] = v.Account
		}
		if types == common.RoleBackupValidator {
			mapBackup[v.Position] = v.Account
		}
	}

	for index := 0; index < common.MasterValidatorNum; index++ {
		k := common.GeneratePosition(uint16(index), common.ElectRoleValidator)
		_, ok := mapMaster[k]
		if ok == false {
			ans, allNative = SloveZeroOffline(k, ans, allNative, top, NotOffline, offline)
			continue
		}
		if findAddress(mapMaster[k], offline) == true {
			ans, allNative = SloveZeroOffline(k, ans, allNative, top, IsOffline, offline)
			continue
		}

	}
	for index := 0; index < common.BackupValidatorNum; index++ {
		k := common.GeneratePosition(uint16(index), common.ElectRoleValidatorBackUp)
		_, ok := mapBackup[k]
		if ok == false {
			ans, allNative = SloveFirstOffline(k, ans, allNative, NotOffline)
			continue
		}
		if findAddress(mapBackup[k], offline) == true {
			ans, allNative = SloveFirstOffline(k, ans, allNative, IsOffline)
			continue
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
