// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package reelection

import (
	"encoding/json"
	"errors"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/election/support"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/params/manparams"
	"io/ioutil"
	"os"
	"strconv"
)

type Info struct {
	Position uint16
	Account  common.Address
}
type NodeSupport struct {
	First  []Info
	Second []Info
	Third  []Info
}

func CheckFileExist(filename string) bool {
	exist := true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist

}
func (self *ReElection) TestSupport(data *mc.RoleUpdatedMsg) error {
	if self.currentID != common.RoleBroadcast {
		return nil
	}
	filename := ""
	log.Info("测试支持", "开始存储", "start", "高度", data.BlockNum)
	nodeSupport := NodeSupport{}
	if data.BlockNum == 0 {
		blk := self.bc.GetBlockByHash(data.BlockHash)
		if blk == nil {
			log.Error("测试支持", "获取区块信息失败 高度", data.BlockNum)
			return errors.New("获取区块头失败")
		}
		for _, v := range blk.Header().NetTopology.NetTopologyData {
			switch common.GetRoleTypeFromPosition(v.Position) {
			case common.RoleValidator:
				nodeSupport.First = append(nodeSupport.First, Info{Position: v.Position, Account: v.Account})
			case common.RoleBackupValidator:
				nodeSupport.Second = append(nodeSupport.Second, Info{Position: v.Position, Account: v.Account})
			default:
				continue

			}
		}
		filename = "./0_top.json"

	} else if data.BlockNum%common.GetReElectionInterval() >= 292 && data.BlockNum%common.GetReElectionInterval() <= 299 {
		validatorHash, err := self.GetHeaderHashByNumber(data.BlockHash, common.GetNextReElectionNumber(data.BlockNum)-manparams.MinerTopologyGenerateUpTime)
		if err != nil {
			log.Error("测试支持 ", "获取292区块头hash失败 err", err, "高度", common.GetNextReElectionNumber(data.BlockNum)-manparams.MinerTopologyGenerateUpTime)
		}
		_, b, err := self.readElectData(common.RoleValidator, validatorHash)
		if err != nil {
			log.Error("测试支持", "获取选举信息失败", "err")
		}
		for _, v := range b.MasterValidator {
			nodeSupport.First = append(nodeSupport.First, Info{Position: v.Position, Account: v.Account})
		}
		for _, v := range b.BackUpValidator {
			nodeSupport.Second = append(nodeSupport.Second, Info{Position: v.Position, Account: v.Account})
		}
		for _, v := range b.CandidateValidator {
			nodeSupport.Third = append(nodeSupport.Third, Info{Position: v.Position, Account: v.Account})
		}
		aim := data.BlockNum / common.GetReElectionInterval()
		aim++
		aim = aim * common.GetReElectionInterval()
		aimF := strconv.Itoa(int(aim))
		filename = "./" + aimF + "_top.json"
	}

	if filename == "" {
		return nil
	}
	marshalData, err := json.Marshal(nodeSupport)
	if err != nil {
		log.Error("测试支持", "Marshal失败 data", nodeSupport)
		return err
	}
	err = ioutil.WriteFile(filename, marshalData, os.ModeAppend)
	if err != nil {
		log.Error("测试支持", "生成test文件成功")
	}
	return err
}

//身份变更消息带来
func (self *ReElection) roleUpdateProcess(data *mc.RoleUpdatedMsg) error {

	self.lock.Lock()
	defer self.lock.Unlock()
	self.currentID = data.Role
	self.TestSupport(data)
	//if common.RoleValidator != self.currentID { //不是验证者，不处理
	//	log.ERROR(Module, "當前不是驗證者，不處理", self.currentID)
	//	return nil
	//}

	err := self.HandleTopGen(data.BlockHash) //处理拓扑生成
	if err != nil {
		log.ERROR(Module, "處理拓撲生成失敗 err", err)
		return err
	}

	err = self.HandleNative(data.BlockHash) //处理初选列表更新
	if err != nil {
		log.ERROR(Module, "處理初選列表更新失敗 err", err)
		return err
	}

	log.INFO(Module, "roleUpdateProcess end height", data.BlockNum)
	return nil

}
func (self *ReElection) HandleNative(hash common.Hash) error {
	height, err := self.GetNumberByHash(hash)
	if err != nil {
		log.Error(Module, "HandleNative阶段 err", err)
		return err
	}

	if true == NeedReadTopoFromDB(height) { //300 600 900 重取缓存
		log.INFO(Module, "需要从db中读取native 高度", height)
		return self.GetNativeFromDB(hash)
	}

	lastHash, err := self.GetHeaderHashByNumber(hash, height-1)
	if err != nil {
		log.Error(Module, "HandleNative err", err)
		return err
	}
	self.checkUpdateStatus(lastHash)
	allNative, err := self.readNativeData(lastHash) //
	if err != nil {
		log.Error(Module, "readNativeData failed height", height-1)
	}

	log.INFO(Module, "self,allNative", allNative)

	err = self.UpdateNative(hash, allNative)
	log.INFO(Module, "更新初选列表结束 高度 ", height, "错误信息", err, "self,allNative", allNative)
	return err
}
func (self *ReElection) HandleTopGen(hash common.Hash) error {
	var err error

	if self.IsMinerTopGenTiming(hash) { //矿工生成时间 240
		log.INFO(Module, "是礦工拓撲生成時間點 height", hash.String())
		err = self.ToGenMinerTop(hash)
		if err != nil {
			log.ERROR(Module, "礦工拓撲生成時間點錯誤 err", err)
		}
	}

	if self.IsValidatorTopGenTiming(hash) { //验证者生成时间 260
		log.INFO(Module, "是驗證者拓撲生成時間點 height", hash)
		err = self.ToGenValidatorTop(hash)
		if err != nil {
			log.ERROR(Module, "驗證者拓撲生成時間點錯誤 err", err)
		}
	}

	return err
}
func (self *ReElection) UpdateNative(hash common.Hash, allNative support.AllNative) error {

	allNative, err := self.ToNativeValidatorStateUpdate(hash, allNative)
	if err != nil {
		log.INFO(Module, "ToNativeMinerStateUpdate validator err", err)
		return nil
	}

	err = self.writeNativeData(hash, allNative)

	log.ERROR(Module, "更新初选列表状态后-写入数据库状态 err", err, "高度对应的hash", hash.String())

	return err

}

//是不是矿工拓扑生成时间段
func (self *ReElection) IsMinerTopGenTiming(hash common.Hash) bool {

	height, err := self.GetNumberByHash(hash)
	if err != nil {
		return false
	}
	now := height % common.GetReElectionInterval()
	if now == MinerTopGenTiming {
		return true
	}
	return false
}

//是不是验证者拓扑生成时间段
func (self *ReElection) IsValidatorTopGenTiming(hash common.Hash) bool {

	height, err := self.GetNumberByHash(hash)
	if err != nil {
		return false
	}

	now := height % common.GetReElectionInterval()
	if now == ValidatorTopGenTiming {
		return true
	}
	return false
}
func NeedReadTopoFromDB(height uint64) bool {
	if (height)%common.GetReElectionInterval() == 0 || height == 0 {
		return true
	}
	return false
}
