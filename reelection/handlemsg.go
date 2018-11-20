//1542678159.1199558
//1542677440.3277602
//1542676748.8989892
//1542675866.0266833
//1542675075.3106349
//1542674384.5508595
//1542673739.7782345
//1542673059.0441284
//1542672198.8333974
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

	if IsValidatorTopGenTiming(height) { //validator generation time 260
		log.INFO(Module, "is validator topology generation time height", height)
		err = self.ToGenValidatorTop(height)
		if err != nil {
			log.ERROR(Module, "validator topology generation time err", err)
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
	log.INFO(Module,"update initial election list","write initial election list to db","height",height,"err",err)
	return err

}
*/
//is miner topology generation time or not
func IsMinerTopGenTiming(height uint64) bool {

	now := height % common.GetReElectionInterval()
	if now == MinerTopGenTiming {
		return true
	}
	return false
}

//is validator topology generation time or not
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
