// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package reelection

import (
	"sync"
	"time"

	"errors"

	"github.com/matrix/go-matrix/accounts"
	"github.com/matrix/go-matrix/baseinterface"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core"
	"github.com/matrix/go-matrix/event"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mandb"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/params/manparams"
	"github.com/syndtr/goleveldb/leveldb"
)

var (
	/*
		MinerTopologyAlreadyGenerate     = errors.New("Miner Topology Already Generate")
		ValidatorTopologyAlreadyGenerate = errors.New("Validator Topology Already Generate")
		MinerNotRecviveTopology          = errors.New("Miner Not Recvive Topology")
		ValidatorNotReceiveTopology      = errors.New("Validator Not Receive Topology")
		TopNotBeLocal                    = errors.New("Top Not Be Local")
	*/

	BroadCastInterval        = common.GetBroadcastInterval()
	MinerTopGenTiming        = common.GetReElectionInterval() - manparams.MinerTopologyGenerateUpTime
	MinerNetchangeTiming     = common.GetReElectionInterval() - manparams.MinerNetChangeUpTime
	ValidatorTopGenTiming    = common.GetReElectionInterval() - manparams.VerifyTopologyGenerateUpTime
	ValidatorNetChangeTiming = common.GetReElectionInterval() - manparams.VerifyNetChangeUpTime
	Time_Out_Limit           = 2 * time.Second
	ChanSize                 = 10
)

const (
	Module = "换届服务"
)

// Backend wraps all methods required for mining.
type Backend interface {
	AccountManager() *accounts.Manager
	BlockChain() *core.BlockChain
	TxPool() *core.TxPool
	ChainDb() mandb.Database
}
type AllNative struct {
	MasterMiner        []mc.TopologyNodeInfo //矿工主节点
	BackUpMiner        []mc.TopologyNodeInfo //矿工备份
	MasterValidator    []mc.TopologyNodeInfo //验证者主节点
	BackUpValidator    []mc.TopologyNodeInfo //验证者备份
	CandidateValidator []mc.TopologyNodeInfo //验证者候选

}

type ElectMiner struct {
	MasterMiner []mc.TopologyNodeInfo
	BackUpMiner []mc.TopologyNodeInfo
}

type ElectValidator struct {
	MasterValidator    []mc.TopologyNodeInfo
	BackUpValidator    []mc.TopologyNodeInfo
	CandidateValidator []mc.TopologyNodeInfo
}

type ElectReturnInfo struct {
	MasterMiner     []mc.TopologyNodeInfo
	BackUpMiner     []mc.TopologyNodeInfo
	MasterValidator []mc.TopologyNodeInfo
	BackUpValidator []mc.TopologyNodeInfo
}
type ReElection struct {
	bc  *core.BlockChain //eth实例：生成种子时获取一周期区块的最小hash
	ldb *leveldb.DB      //本都db数据库

	roleUpdateCh    chan *mc.RoleUpdatedMsg //身份变更信息通道
	roleUpdateSub   event.Subscription
	minerGenCh      chan *mc.MasterMinerReElectionRsp //矿工主节点生成消息通道
	minerGenSub     event.Subscription
	validatorGenCh  chan *mc.MasterValidatorReElectionRsq //验证者主节点生成消息通道
	validatorGenSub event.Subscription
	electionSeedCh  chan *mc.ElectionEvent //选举种子请求消息通道
	electionSeedSub event.Subscription

	//allNative AllNative

	currentID common.RoleType //当前身份

	elect  baseinterface.ElectionInterface
	random *baseinterface.Random
	lock   sync.Mutex
}

func New(bc *core.BlockChain, dbDir string, random *baseinterface.Random) (*ReElection, error) {
	reelection := &ReElection{
		bc:             bc,
		roleUpdateCh:   make(chan *mc.RoleUpdatedMsg, ChanSize),
		minerGenCh:     make(chan *mc.MasterMinerReElectionRsp, ChanSize),
		validatorGenCh: make(chan *mc.MasterValidatorReElectionRsq, ChanSize),
		electionSeedCh: make(chan *mc.ElectionEvent, ChanSize),
		random:         random,

		currentID: common.RoleDefault,
	}
	reelection.elect = baseinterface.NewElect()
	var err error
	dbDir = dbDir + "/reElection"
	reelection.ldb, err = leveldb.OpenFile(dbDir, nil)
	if err != nil {
		return nil, err
	}
	err = reelection.initSubscribeEvent()
	if err != nil {
		return nil, err
	}
	go reelection.update()
	return reelection, nil
}

func (self *ReElection) initSubscribeEvent() error {
	var err error

	self.roleUpdateSub, err = mc.SubscribeEvent(mc.CA_RoleUpdated, self.roleUpdateCh)

	if err != nil {
		return err
	}
	log.INFO(Module, "CA_RoleUpdated", "订阅成功")
	return nil
}
func (self *ReElection) update() {
	defer func() {
		if self.roleUpdateSub != nil {
			self.roleUpdateSub.Unsubscribe()
		}

	}()
	for {
		select {
		case roleData := <-self.roleUpdateCh:
			log.INFO(Module, "roleData", roleData)
			go self.roleUpdateProcess(roleData)
		}
	}
}

func (self *ReElection) GetTopoChange(hash common.Hash, offline []common.Address) ([]mc.Alternative, error) {

	height, err := self.GetNumberByHash(hash)
	if err != nil {
		return []mc.Alternative{}, errors.New("根据hash获取高度失败")
	}
	height = height + 1
	self.lock.Lock()
	defer self.lock.Unlock()
	if common.IsReElectionNumber(height) {
		log.INFO(Module, "是换届区块", "无差值")
		return []mc.Alternative{}, nil
	}

	log.INFO(Module, "获取拓扑改变 start height", height, "offline", offline)
	lastHash, err := self.GetHeaderHashByNumber(hash, height-1)
	if err != nil {
		log.Error(Module, "根据hash获取高度失败 err", err)
		return []mc.Alternative{}, err
	}
	self.checkUpdateStatus(lastHash)
	antive, err := self.readNativeData(lastHash)
	if err != nil {
		log.Error(Module, "获取上一个高度的初选列表失败 height-1", height-1)
		return []mc.Alternative{}, err
	}

	//aim := 0x04 + 0x08
	TopoGrap, err := GetCurrentTopology(lastHash, common.RoleBackupValidator|common.RoleValidator)
	if err != nil {
		log.Error(Module, "获取CA当前拓扑图失败 err", err)
		return []mc.Alternative{}, err
	}

	log.Info(Module, "获取拓扑变化 start 上一个高度缓存allNative-M", antive.MasterQ, "B", antive.BackUpQ, "Can", antive.CandidateQ)
	DiffValidatot := self.TopoUpdate(offline, antive, TopoGrap)
	log.INFO(Module, "获取拓扑改变 end ", DiffValidatot)
	return DiffValidatot, nil

}

func (self *ReElection) GetElection(hash common.Hash) (*ElectReturnInfo, error) {
	log.INFO(Module, "开始获取选举信息 hash", hash.String())
	height, err := self.GetNumberByHash(hash)
	if err != nil {
		log.Error(Module, "GetElection", "获取hash的高度失败")
		return nil, err
	}
	if common.IsReElectionNumber(height + 1 + manparams.MinerNetChangeUpTime) {
		log.Error(Module, "是矿工网络生成切换时间点 height", height)
		MinerHash, err := self.GetHeaderHashByNumber(hash, height+1+manparams.MinerNetChangeUpTime-manparams.MinerTopologyGenerateUpTime)
		if err != nil {
			return nil, err
		}
		if err := self.checkTopGenStatus(MinerHash); err != nil {
			log.ERROR(Module, "检查top生成出错 err", err)
		}
		ans, _, err := self.readElectData(common.RoleMiner, MinerHash)
		if err != nil {
			log.ERROR(Module, "获取本地矿工选举信息失败", "miner", "heightminer", height+manparams.MinerNetChangeUpTime)
			return nil, err
		}
		resultM := &ElectReturnInfo{
			MasterMiner: ans.MasterMiner,
			BackUpMiner: ans.BackUpMiner,
		}
		return resultM, nil
	} else if common.IsReElectionNumber(height + 1 + manparams.VerifyNetChangeUpTime) {
		log.Error(Module, "是验证者网络切换时间点 height", height)
		ValidatorHash, err := self.GetHeaderHashByNumber(hash, height+1+manparams.VerifyNetChangeUpTime-manparams.VerifyTopologyGenerateUpTime)
		if err != nil {
			return nil, err
		}
		if err := self.checkTopGenStatus(ValidatorHash); err != nil {
			log.ERROR(Module, "检查top生成出错 err", err)
		}
		_, ans, err := self.readElectData(common.RoleValidator, ValidatorHash)
		if err != nil {
			log.ERROR(Module, "获取本地验证者选举信息失败", "miner", "heightValidator", height+manparams.VerifyNetChangeUpTime)
			return nil, err
		}
		resultV := &ElectReturnInfo{
			MasterValidator: ans.MasterValidator,
			BackUpValidator: ans.BackUpValidator,
		}
		return resultV, nil
	}
	log.INFO(Module, "GetElection end height", height)
	log.INFO(Module, "不是任何网络切换时间点 height", height)
	temp := &ElectReturnInfo{}
	return temp, nil

}

func (self *ReElection) GetNetTopologyAll(hash common.Hash) (*ElectReturnInfo, error) {

	height, err := self.GetNumberByHash(hash)
	if err != nil {
		return nil, err
	}
	if common.IsReElectionNumber(height + 2) {
		heightMiner := height + 2 - manparams.MinerTopologyGenerateUpTime
		MinerHash, err := self.GetHeaderHashByNumber(hash, heightMiner)
		if err != nil {
			return nil, err
		}
		if err := self.checkTopGenStatus(MinerHash); err != nil {
			log.ERROR(Module, "检查top生成出错 err", err)
		}
		ans, _, err := self.readElectData(common.RoleMiner, MinerHash)
		if err != nil {
			return nil, err
		}

		heightValidator := height + 2 - manparams.VerifyTopologyGenerateUpTime
		ValidatorHash, err := self.GetHeaderHashByNumber(hash, heightValidator)
		if err != nil {
			return nil, err
		}
		if err := self.checkTopGenStatus(ValidatorHash); err != nil {
			log.ERROR(Module, "检查top生成出错 err", err)
		}
		_, ans1, err := self.readElectData(common.RoleValidator, ValidatorHash)
		if err != nil {
			return nil, err
		}
		result := &ElectReturnInfo{
			MasterMiner:     ans.MasterMiner,
			BackUpMiner:     ans.BackUpMiner,
			MasterValidator: ans1.MasterValidator,
			BackUpValidator: ans1.BackUpValidator,
		}
		return result, nil

	}
	result := &ElectReturnInfo{
		MasterMiner:     make([]mc.TopologyNodeInfo, 0),
		BackUpMiner:     make([]mc.TopologyNodeInfo, 0),
		MasterValidator: make([]mc.TopologyNodeInfo, 0),
		BackUpValidator: make([]mc.TopologyNodeInfo, 0),
	}
	return result, nil
}
