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
	"encoding/json"
	"errors"
	"math/big"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/params"
)

func locate(address common.Address, master []mc.TopologyNodeInfo, backUp []mc.TopologyNodeInfo, cand []mc.TopologyNodeInfo) (int, mc.TopologyNodeInfo) {
	for _, v := range master {
		if v.Account == address {
			return 0, v
		}
	}
	for _, v := range backUp {
		if v.Account == address {
			return 1, v
		}
	}
	for _, v := range cand {
		if v.Account == address {
			return 2, v
		}
	}
	return -1, mc.TopologyNodeInfo{}
}

func (self *ReElection) whereIsV(address common.Address, role common.RoleType, height uint64) (int, mc.TopologyNodeInfo, error) {
	switch {
	case role == common.RoleMiner:
		height = height / common.GetBroadcastInterval()
		height = height*common.GetBroadcastInterval() - params.MinerTopologyGenerateUptime
		ans, _, err := self.readElectData(common.RoleMiner, height)
		if err != nil {
			return -1, mc.TopologyNodeInfo{}, err
		}
		flag, aimOnline := locate(address, ans.MasterMiner, ans.BackUpMiner, []mc.TopologyNodeInfo{})
		return flag, aimOnline, nil

	case role == common.RoleValidator:
		height = height / common.GetBroadcastInterval()
		height = height*common.GetBroadcastInterval() - params.VerifyTopologyGenerateUpTime
		_, ans, err := self.readElectData(common.RoleValidator, height)
		if err != nil {
			return -1, mc.TopologyNodeInfo{}, err
		}
		flag, aimOnline := locate(address, ans.MasterValidator, ans.BackUpValidator, ans.CandidateValidator)
		return flag, aimOnline, nil
	default:
		log.ERROR(Module, "whereIsV ", "role must be role or validatoe")
		return -1, mc.TopologyNodeInfo{}, errors.New("whereIsV role must be role or validatoe")
	}
}
func (self *ReElection) ToNativeMinerStateUpdate(height uint64, allNative AllNative) (AllNative, error) {
	DiffFromBlock := self.bc.GetHeaderByNumber(height).NetTopology
	//测试
	//DiffFromBlock := common.NetTopology{}
	//aim := 0x04 + 0x08
	TopoGrap, _ := GetCurrentTopology(height-1, common.RoleMiner)
	online, offline := self.CalOnline(DiffFromBlock, TopoGrap)
	log.INFO(Module, "ToNativeMinerStateUpdate online", online, "offline", offline)
	allNative.MasterMiner, allNative.BackUpMiner = deleteOfflineNode(offline, allNative.MasterMiner, allNative.BackUpMiner)

	for _, v := range online {
		flag, aimonline, err := self.whereIsV(v, common.RoleMiner, height)
		if err != nil {
			return AllNative{}, err
		}
		if flag == -1 {
			continue
		}
		allNative.MasterMiner, allNative.BackUpMiner, _ = self.NativeUpdate(allNative.MasterMiner, allNative.BackUpMiner, []mc.TopologyNodeInfo{}, aimonline, flag)
	}

	return allNative, nil
}

func (self *ReElection) ToNativeValidatorStateUpdate(height uint64, allNative AllNative) (AllNative, error) {

	DiffFromBlock := self.bc.GetHeaderByNumber(height - 1).NetTopology

	//测试
	//DiffFromBlock := common.NetTopology{}
	//aim := 0x01 + 0x02
	TopoGrap, _ := GetCurrentTopology(height-1, common.RoleValidator)
	online, offline := self.CalOnline(DiffFromBlock, TopoGrap)
	log.INFO(Module, "ToNativeValidatorStateUpdate online", online, "offline", offline)
	allNative.MasterValidator, allNative.BackUpValidator = deleteOfflineNode(offline, allNative.MasterValidator, allNative.BackUpValidator)

	for _, v := range online {
		flag, aimonline, err := self.whereIsV(v, common.RoleMiner, height)
		if err != nil {
			return AllNative{}, err
		}

		if flag == -1 {
			continue
		}

		allNative.MasterValidator, allNative.BackUpValidator, allNative.CandidateValidator = self.NativeUpdate(allNative.MasterValidator, allNative.BackUpValidator, allNative.CandidateValidator, aimonline, flag)

	}

	return allNative, nil
}

func deleteOfflineNode(offline []common.Address, Master []mc.TopologyNodeInfo, BackUp []mc.TopologyNodeInfo) ([]mc.TopologyNodeInfo, []mc.TopologyNodeInfo) {
	for k, v := range Master {
		if IsInArray(v.Account, offline) {
			Master = append(Master[:k], Master[k+1:]...)
		}
	}
	for k, v := range BackUp {
		if IsInArray(v.Account, offline) {
			BackUp = append(BackUp[:k], BackUp[k+1:]...)
		}
	}
	return Master, BackUp
}

func (self *ReElection) CalOnline(diff common.NetTopology, top *mc.TopologyGraph) ([]common.Address, []common.Address) {
	online := make([]common.Address, 0)
	offline := make([]common.Address, 0)

	for _, v := range diff.NetTopologyData {

		if v.Position == 0xF000 {
			offline = append(offline, v.Account)
			continue
		}
		if v.Position == 0xF001 {
			online = append(online, v.Account)
			continue
		}
		nativeAdd := checkInGraph(top, v.Position)
		if checkInDiff(diff, nativeAdd) == false {
			offline = append(offline, nativeAdd)
		}

	}

	return online, offline
}
func checkInGraph(top *mc.TopologyGraph, pos uint16) common.Address {
	for _, v := range top.NodeList {
		if v.Position == pos {
			return v.Account
		}
	}
	return common.Address{}
}
func checkInDiff(diff common.NetTopology, add common.Address) bool {
	for _, v := range diff.NetTopologyData {
		if v.Account == add {
			return true
		}
	}
	return false
}
func IsInArray(aimAddress common.Address, offline []common.Address) bool {
	for _, v := range offline {
		if v == aimAddress {
			return true
		}
	}
	return false
}
func (self *ReElection) writeNativeData(height uint64, data AllNative) error {
	key := MakeNativeDBKey(height)
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	err = self.ldb.Put([]byte(key), jsonData, nil)
	return err
}

func (self *ReElection) readNativeData(height uint64) (AllNative, error) {
	key := MakeNativeDBKey(height)
	ans, err := self.ldb.Get([]byte(key), nil)
	if err != nil {
		return AllNative{}, err
	}
	var realAns AllNative
	err = json.Unmarshal(ans, &realAns)
	if err != nil {
		return AllNative{}, err
	}

	return realAns, nil

}
func MakeNativeDBKey(height uint64) string {
	t := big.NewInt(int64(height))
	ss := t.String() + "---" + "Native"
	return ss
}
func (self *ReElection) GetNativeFromDB(height uint64) error {
	log.INFO(Module, "GetNativeFromDB", height)
	minerH := height - params.MinerNetChangeUpTime
	minerELect, _, err := self.readElectData(common.RoleMiner, minerH)
	if err != nil {
		return err
	}

	validatorH := height - params.VerifyNetChangeUpTime
	_, validatorElect, err := self.readElectData(common.RoleValidator, validatorH)
	if err != nil {
		return err
	}
	log.INFO(Module, "GetNativeFromDB", height, "ready to writeNativeData data", AllNative{
		MasterMiner:        minerELect.MasterMiner,
		BackUpMiner:        minerELect.BackUpMiner,
		MasterValidator:    validatorElect.MasterValidator,
		BackUpValidator:    validatorElect.BackUpValidator,
		CandidateValidator: validatorElect.CandidateValidator,
	})
	return self.writeNativeData(height, AllNative{
		MasterMiner:        minerELect.MasterMiner,
		BackUpMiner:        minerELect.BackUpMiner,
		MasterValidator:    validatorElect.MasterValidator,
		BackUpValidator:    validatorElect.BackUpValidator,
		CandidateValidator: validatorElect.CandidateValidator,
	})

}
