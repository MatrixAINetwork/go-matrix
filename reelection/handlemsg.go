//1542527827.018523
//1542527761.4027135
//1542526944.9498281
//1542526877.9396803
//1542526851.3072448
//1542526138.0412877
//1542525151.7506428
//1542524229.8355923
//1542523355.342893
//1542523332.5677123
//1542522607.4828787
// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package reelection

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
)

//role change message
func (self *ReElection) roleUpdateProcess(data *mc.RoleUpdatedMsg) error {
	self.lock.Lock()
	defer self.lock.Unlock()
	self.currentID = data.Role

	if common.RoleValidator != self.currentID { //Ignore Non-Validators
		log.ERROR(Module, "It is not a validator. No handling will be done", self.currentID)
		return nil
	}

	err := self.HandleTopGen(data.BlockNum) //Topology Generation
	if err != nil {
		log.ERROR(Module, "topology generation failure err", err)
		return err
	}

	/*
	err = self.HandleNative(data.BlockNum) //Initial Election List Update
	if err != nil {
		log.ERROR(Module, "Initial Election List Update Failure err", err)
		return err
	}
	*/
	log.INFO(Module, "roleUpdateProcess end height", data.BlockNum)
	return nil

}
/*
func (self *ReElection) HandleNative(height uint64) error {
	if true == IsinFristPeriod(height) { //Don't Update during the First Election Cycle
		log.INFO(Module, "BlockNum", height, "no need to update native list", "nil")
		return nil
	}
	if true == NeedReadTopoFromDB(height) { //300 600 900 re-load the cache
		return self.GetNativeFromDB(height)
	}

	allNative, err := self.readNativeData(height - 1) //
	if err != nil {
		log.Error(Module, "readNativeData failed height", height-1)
	}

	log.INFO(Module, "self,allNative", allNative)

	err = self.UpdateNative(height, allNative)
	log.INFO(Module,"finish updating initial election list  error message",err)
	return err
}
*/
func (self *ReElection) HandleTopGen(height uint64) error {
	var err error

	if IsMinerTopGenTiming(height) { //Miner Generation Time 240
		log.INFO(Module, "is miner topology generation time height", height)
		err = self.ToGenMinerTop(height)
		if err != nil {
			log.ERROR(Module, "miner topology generation time err", err)
		}
	}

	if IsValidatorTopGenTiming(height) { //验证者生成时间 260
		log.INFO(Module, "是驗證者拓撲生成時間點 height", height)
		err = self.ToGenValidatorTop(height)
		if err != nil {
			log.ERROR(Module, "驗證者拓撲生成時間點錯誤 err", err)
		}
	}

	return err
}
/*
func (self *ReElection) UpdateNative(height uint64, allNative AllNative) error {

	allNative, err := self.ToNativeMinerStateUpdate(height, allNative)
	if err != nil {
		log.ERROR(Module, "ToNativeMinerStateUpdate miner err ", err)
		return nil
	}

	allNative, err = self.ToNativeValidatorStateUpdate(height, allNative)
	if err != nil {
		log.INFO(Module, "ToNativeMinerStateUpdate validator err", err)
		return nil
	}

	err=self.writeNativeData(height, allNative)
	log.INFO(Module,"更新初选列表阶段","写初选列表到数据库","高度",height,"err",err)
	return err

}
*/
//是不是矿工拓扑生成时间段
func IsMinerTopGenTiming(height uint64) bool {

	now := height % common.GetReElectionInterval()
	if now == MinerTopGenTiming {
		return true
	}
	return false
}

//是不是验证者拓扑生成时间段
func IsValidatorTopGenTiming(height uint64) bool {

	now := height % common.GetReElectionInterval()
	if now == ValidatorTopGenTiming {
		return true
	}
	return false
}

func NeedReadTopoFromDB(height uint64) bool {
	if height%common.GetReElectionInterval() == 0 {
		return true
	}
	return false
}

func IsinFristPeriod(height uint64) bool {
	if height < common.GetReElectionInterval() {
		return true
	}
	return false
}
