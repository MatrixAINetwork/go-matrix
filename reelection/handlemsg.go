// Copyright 2018 The MATRIX Authors as well as Copyright 2014-2017 The go-ethereum Authors
// This file is consisted of the MATRIX library and part of the go-ethereum library.
//
// The MATRIX-ethereum library is free software: you can redistribute it and/or modify it under the terms of the MIT License.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, 
//and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject tothe following conditions:
//
//The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
//
//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, 
//WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISINGFROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE
//OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
package reelection

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
)

//身份变更消息带来
func (self *ReElection) roleUpdateProcess(data *mc.RoleUpdatedMsg) error {
	self.lock.Lock()
	defer self.lock.Unlock()
	self.currentID = data.Role

	if common.RoleValidator != self.currentID { //Ignore Non-Validators
		log.ERROR(Module, "當前不是驗證者，不處理", self.currentID)
		return nil
	}

	err := self.HandleTopGen(data.BlockNum) //处理拓扑生成
	if err != nil {
		log.ERROR(Module, "處理拓撲生成失敗 err", err)
		return err
	}

	/*
	err = self.HandleNative(data.BlockNum) //处理初选列表更新
	if err != nil {
		log.ERROR(Module, "處理初選列表更新失敗 err", err)
		return err
	}
	*/
	log.INFO(Module, "roleUpdateProcess end height", data.BlockNum)
	return nil

}
/*
func (self *ReElection) HandleNative(height uint64) error {
	if true == IsinFristPeriod(height) { //第一选举周期不更新
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

	if IsMinerTopGenTiming(height) { //矿工生成时间 240
		log.INFO(Module, "是礦工拓撲生成時間點 height", height)
		err = self.ToGenMinerTop(height)
		if err != nil {
			log.ERROR(Module, "礦工拓撲生成時間點錯誤 err", err)
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
