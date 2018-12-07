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
)

func (self *ReElection) ToNativeValidatorStateUpdate(hash common.Hash, allNative support.AllNative) (support.AllNative, error) {
	height, err := self.GetNumberByHash(hash)
	if err != nil {
		log.Error(Module, "ToNativeValidatorStateUpdate err", err)
		return support.AllNative{}, err
	}

	block := self.bc.GetBlockByHash(hash)
	if block == nil {
		log.ERROR(Module, "获取指定高度的区块头失败 高度", height)
		return support.AllNative{}, errors.New("获取指定高度的区块头失败")
	}
	DiffFromBlock := block.Header().NetTopology

	lastHash, err := self.GetHeaderHashByNumber(hash, height-1)
	if err != nil {
		return support.AllNative{}, errors.New("根据hash获取高度失败")
	}
	TopoGrap, err := GetCurrentTopology(lastHash, common.RoleValidator|common.RoleBackupValidator)
	log.INFO(Module, "更新初选列表信息 拓扑的高度", height-1, "拓扑值", TopoGrap, "diff", DiffFromBlock)
	if err != nil {
		log.ERROR(Module, "从ca获取验证者拓扑图失败", err)
		return allNative, err
	}

	allNative = self.CalOnline(DiffFromBlock, TopoGrap, allNative)
	log.INFO(Module, "更新上下线状态", "结束", "高度", height)

	return allNative, nil
}

func deleteQueue(address common.Address, allNative support.AllNative) support.AllNative {
	log.INFO(Module, "在缓存中删除节点阶段-开始 地址", address, "缓存", allNative)
	for k, v := range allNative.MasterQ {
		if v == address {
			allNative.MasterQ = append(allNative.MasterQ[:k], allNative.MasterQ[k+1:]...)
			log.INFO(Module, "在缓存中删除节点阶段-master 地址 ", address, "缓存", allNative)
			return allNative
		}
	}
	for k, v := range allNative.BackUpQ {
		if v == address {
			allNative.BackUpQ = append(allNative.BackUpQ[:k], allNative.BackUpQ[k+1:]...)
			log.INFO(Module, "在缓存中删除节点阶段-backup 地址", address, "缓存", allNative)
			return allNative
		}
	}
	for k, v := range allNative.CandidateQ {
		if v == address {
			allNative.CandidateQ = append(allNative.CandidateQ[:k], allNative.CandidateQ[k+1:]...)
			log.INFO(Module, "在缓存中删除节点阶段-candidate 地址", address, "缓存", allNative)
			return allNative
		}
	}

	log.INFO(Module, "在缓存中删除节点阶段-结束-不再任何一个梯队 地址", address, "缓存", allNative)
	return allNative
}

func addQueue(address common.Address, allNative support.AllNative) support.AllNative {
	log.INFO(Module, "在缓存中增加节点阶段-开始 地址", address, "allNative", allNative)
	for _, v := range allNative.Master {
		if v.Account == address {

			allNative.MasterQ = append(allNative.MasterQ, address)
			log.INFO(Module, "在缓存中增加节点阶段-master 地址", address, "allNative", allNative)
			return allNative
		}
	}
	for _, v := range allNative.BackUp {
		if v.Account == address {
			allNative.BackUpQ = append(allNative.BackUpQ, address)
			log.INFO(Module, "在缓存中增加节点阶段-backup 地址", address, "allNative", allNative)
			return allNative
		}
	}
	for _, v := range allNative.Candidate {
		if v.Account == address {
			allNative.CandidateQ = append(allNative.CandidateQ, address)
			log.INFO(Module, "在缓存中增加节点阶段-candidate 地址", address, "allNative", allNative)
			return allNative
		}
	}
	log.INFO(Module, "在缓存中增加节点阶段-结束 地址-不在任何一个梯队", address, "allNative", allNative)
	return allNative
}
func (self *ReElection) CalOnline(diff common.NetTopology, top *mc.TopologyGraph, allNative support.AllNative) support.AllNative {

	log.INFO(Module, "更新上下线阶段 拓扑差值-开始", diff.NetTopologyData, "allNative", allNative)

	for _, v := range diff.NetTopologyData {
		if v.Position == common.PosOnline {
			allNative = addQueue(v.Account, allNative)
		} else {
			allNative = deleteQueue(v.Account, allNative)
		}
	}
	log.INFO(Module, "更新上下线阶段 拓扑差值-结束", diff.NetTopologyData, "allNative", allNative)
	return allNative

}

func (self *ReElection) writeNativeData(hash common.Hash, data support.AllNative) error {
	key := MakeNativeDBKey(hash)
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	err = self.ldb.Put([]byte(key), jsonData, nil)
	log.INFO(Module, "数据库 初选列表 err", err, "高度对应的hash", hash.String(), "key", key)
	return err
}

func (self *ReElection) readNativeData(hash common.Hash) (support.AllNative, error) {

	key := MakeNativeDBKey(hash)
	ans, err := self.ldb.Get([]byte(key), nil)
	if err != nil {
		return support.AllNative{}, err
	}
	var realAns support.AllNative
	err = json.Unmarshal(ans, &realAns)
	if err != nil {
		return support.AllNative{}, err
	}

	return realAns, nil

}
func MakeNativeDBKey(hash common.Hash) string {
	ss := hash.String() + "---" + "Native"
	return ss
}
func needReadFromGenesis(height uint64) bool {
	if height == 0 {
		return true
	}
	return false
}
func (self *ReElection) wirteNativeFromGeneis() error {
	preBroadcast := support.AllNative{}
	block := self.bc.GetBlockByNumber(0)
	if block == nil {
		return errors.New("第0块区块拿不到")
	}
	header := block.Header()
	if header == nil {
		return errors.New("第0块区块头拿不到")
	}
	for _, v := range header.NetTopology.NetTopologyData {
		switch common.GetRoleTypeFromPosition(v.Position) {
		case common.RoleValidator:
			temp := mc.TopologyNodeInfo{
				Account:  v.Account,
				Position: v.Position,
				Type:     common.RoleValidator,
			}
			preBroadcast.Master = append(preBroadcast.Master, temp)
		case common.RoleBackupValidator:
			temp := mc.TopologyNodeInfo{
				Account:  v.Account,
				Position: v.Position,
				Type:     common.RoleBackupValidator,
			}
			preBroadcast.BackUp = append(preBroadcast.BackUp, temp)
		}
	}
	log.INFO(Module, "第0块到达处理阶段 更新初选列表", "从0的区块头中获取", "初选列表", preBroadcast)
	ZeroBlock := self.bc.GetBlockByNumber(0)
	if ZeroBlock == nil {
		return errors.New("不存在0块")
	}

	err := self.writeNativeData(ZeroBlock.Hash(), preBroadcast)
	log.INFO(Module, "第0块到达处理阶段 更新初选列表", "从0的区块头中获取 写数据到数据库", "err", err)
	return err
}
func (self *ReElection) GetNativeFromDB(hash common.Hash) error {
	height, err := self.GetNumberByHash(hash)
	if err != nil {
		log.Error(Module, "GetNativeFromDB 阶段err ", err)
		return err
	}
	if needReadFromGenesis(height) {
		return self.wirteNativeFromGeneis()
	}
	log.INFO(Module, "GetNativeFromDB", height)

	if err := self.checkTopGenStatus(hash); err != nil {
		log.ERROR(Module, "检查top生成出错 err", err)
	}
	_, validatorElect, err := self.readElectData(common.RoleValidator, hash)
	if err != nil {
		return err
	}
	preBroadcast := support.AllNative{
		Master:     validatorElect.MasterValidator,
		BackUp:     validatorElect.BackUpValidator,
		Candidate:  validatorElect.CandidateValidator,
		MasterQ:    []common.Address{},
		BackUpQ:    []common.Address{},
		CandidateQ: []common.Address{},
	}

	for _, v := range preBroadcast.Candidate {
		preBroadcast.CandidateQ = append(preBroadcast.CandidateQ, v.Account)
	}

	err = self.writeNativeData(hash, preBroadcast)
	log.INFO(Module, "writeNativeData", height, "err", err)
	return err
}
func (self *ReElection) boolNativeStatus(hash common.Hash) bool {
	if _, err := self.readNativeData(hash); err != nil {
		return false
	}
	return true
}
func (self *ReElection) TopoUpdate(offline []common.Address, allNative support.AllNative, top *mc.TopologyGraph) []mc.Alternative {
	return self.elect.ToPoUpdate(offline, allNative, top)
}
