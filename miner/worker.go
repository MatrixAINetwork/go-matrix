// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package miner

import (
	"math/big"
	"sync"
	"sync/atomic"
	"time"

	"errors"

	"github.com/MatrixAINetwork/go-matrix/ca"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/consensus"
	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/event"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/msgsend"
	"github.com/MatrixAINetwork/go-matrix/params"
	"github.com/MatrixAINetwork/go-matrix/params/manparams"
	"github.com/MatrixAINetwork/go-matrix/params/manversion"
)

const (
	resultQueueSize   = 10
	chainHeadChanSize = 10
)

// Agent can register themself with the worker
type Agent interface {
	Work() chan<- *Work
	SetReturnCh(chan<- *consensus.SealResult)
	Stop()
	Start()
	GetHashRate() int64
}

// Work is the workers current environment and holds
// all of the current state information
type Work struct {
	config *params.ChainConfig
	signer types.Signer

	state *state.StateDBManage // apply state changes here
	Block *types.Block         // the new block

	header   *types.Header
	txs      []types.SelfTransaction
	receipts []*types.Receipt

	createdAt time.Time

	isBroadcastNode bool

	mineType mineTaskType
}

type Result struct {
	Difficulty *big.Int
	Header     *types.Header
}

// worker is the main object which takes care of applying messages to the new state
type worker struct {
	config *params.ChainConfig
	bc     ChainReader

	mu sync.Mutex

	// update loop
	mux *event.TypeMux

	agents map[Agent]struct{}
	recv   chan *consensus.SealResult

	extra []byte

	currentMu sync.Mutex
	current   *Work
	// atomic status counters
	mining int32

	quitCh                chan struct{}
	roleUpdateCh          chan *mc.RoleUpdatedMsg
	roleUpdateSub         event.Subscription
	miningRequestCh       chan *mc.HD_MiningReqMsg
	miningRequestSub      event.Subscription
	localMiningRequestCh  chan *mc.BlockGenor_BroadcastMiningReqMsg
	localMiningRequestSub event.Subscription
	v2MiningRequestCh     chan *mc.HD_V2_MiningReqMsg
	v2MiningRequestSub    event.Subscription
	registerAgentCh       chan Agent
	unRegisterAgentCh     chan Agent
	mineReqCtrl           *mineReqCtrl
	hd                    *msgsend.HD
	mineResultSender      *common.ResendMsgCtrl

	curSuperSeq uint64
	curVersion  string
	handlerV2   *workerV2
}

type ChainReader interface {
	Config() *params.ChainConfig
	Engine(version []byte) consensus.Engine
	DPOSEngine(version []byte) consensus.DPOSEngine
	GetCurrentHash() common.Hash
	GetGraphByHash(hash common.Hash) (*mc.TopologyGraph, *mc.ElectGraph, error)
	GetBroadcastAccounts(blockHash common.Hash) ([]common.Address, error)
	GetInnerMinerAccounts(blockHash common.Hash) ([]common.Address, error)
	GetVersionSuperAccounts(blockHash common.Hash) ([]common.Address, error)
	GetBlockSuperAccounts(blockHash common.Hash) ([]common.Address, error)
	GetBroadcastIntervalByHash(blockHash common.Hash) (*mc.BCIntervalInfo, error)
	GetA0AccountFromAnyAccount(account common.Address, blockHash common.Hash) (common.Address, common.Address, error)
	CurrentHeader() *types.Header
	// GetBlock retrieves a block from the database by hash and number.
	GetBlock(hash common.Hash, number uint64) *types.Block
	GetHeader(hash common.Hash, number uint64) *types.Header

	// GetHeaderByNumber retrieves a block header from the database by number.
	GetHeaderByNumber(number uint64) *types.Header

	// GetHeaderByHash retrieves a block header from the database by its hash.
	GetHeaderByHash(hash common.Hash) *types.Header

	GetAncestorHash(sonHash common.Hash, ancestorNumber uint64) (common.Hash, error)

	GetMinDifficulty(blockHash common.Hash) (*big.Int, error)
	GetMaxDifficulty(blockHash common.Hash) (*big.Int, error)
	GetReelectionDifficulty(blockHash common.Hash) (*big.Int, error)
	GetBlockDurationStatus(blockHash common.Hash) (*mc.BlockDurationStatus, error)
}

func newWorker(config *params.ChainConfig, bc ChainReader, mux *event.TypeMux, hd *msgsend.HD) (*worker, error) {
	worker := &worker{
		config: config,
		bc:     bc,
		mux:    mux,

		agents:               make(map[Agent]struct{}),
		quitCh:               make(chan struct{}),
		miningRequestCh:      make(chan *mc.HD_MiningReqMsg, 100),
		roleUpdateCh:         make(chan *mc.RoleUpdatedMsg, 100),
		v2MiningRequestCh:    make(chan *mc.HD_V2_MiningReqMsg, 10),
		recv:                 make(chan *consensus.SealResult, resultQueueSize),
		localMiningRequestCh: make(chan *mc.BlockGenor_BroadcastMiningReqMsg, 100),
		registerAgentCh:      make(chan Agent, 2),
		unRegisterAgentCh:    make(chan Agent, 2),
		mineReqCtrl:          newMinReqCtrl(bc),
		hd:                   hd,
		mineResultSender:     nil,
		curSuperSeq:          0,
		curVersion:           "",
	}

	worker.handlerV2 = newWorkerV2(worker)

	atomic.StoreInt32(&worker.mining, 0)

	err := worker.init_SubscribeEvent()
	if err != nil {
		log.Error(ModuleMiner, "worker创建失败", err)
		return nil, err
	}
	go worker.update()
	log.Info(ModuleMiner, "worker创建成功", err)
	return worker, nil
}
func (self *worker) init_SubscribeEvent() error {
	var err error

	self.localMiningRequestSub, err = mc.SubscribeEvent(mc.HD_BroadcastMiningReq, self.localMiningRequestCh) //广播节点
	if err != nil {
		log.Error(ModuleMiner, "广播节点挖矿请求订阅失败", err)
		return err
	}

	self.roleUpdateSub, err = mc.SubscribeEvent(mc.CA_RoleUpdated, self.roleUpdateCh) //身份到达
	if err != nil {
		log.Error(ModuleMiner, "身份更新订阅失败", err)
		return err
	}

	self.miningRequestSub, err = mc.SubscribeEvent(mc.HD_MiningReq, self.miningRequestCh) //挖矿请求
	if err != nil {
		log.Error(ModuleMiner, "普通矿工挖矿请求订阅失败", err)
		return err
	}

	self.v2MiningRequestSub, err = mc.SubscribeEvent(mc.HD_V2_MiningReq, self.v2MiningRequestCh) //V2挖矿请求
	if err != nil {
		log.Error(ModuleMiner, "V2挖矿请求订阅失败", err)
		return err
	}

	return nil

}

func (self *worker) Getmining() int32 { return atomic.LoadInt32(&self.mining) }

func (self *worker) update() {
	defer func() {
		if self.localMiningRequestSub != nil {
			self.localMiningRequestSub.Unsubscribe()
		}
		if self.miningRequestSub != nil {
			self.miningRequestSub.Unsubscribe()
		}
		if self.roleUpdateSub != nil {
			self.roleUpdateSub.Unsubscribe()
		}
		self.StopAgent()
		self.stopMineResultSender()
		self.mineReqCtrl.Clear()
		log.Info("矿工节点退出成功")
	}()

	for {
		select {
		case roleData := <-self.roleUpdateCh:
			self.RoleUpdatedMsgHandler(roleData)

		case minerReqData := <-self.miningRequestCh:
			self.MiningRequestHandle(minerReqData)

		case data := <-self.localMiningRequestCh:
			self.BroadcastHashLocalMiningReqMsgHandle(data.BlockMainData)

		case reqMsg := <-self.v2MiningRequestCh:
			self.handlerV2.MineReqMsgHandle(reqMsg)

		case minedResult := <-self.recv:
			if minedResult == nil || minedResult.Header == nil {
				log.Trace(ModuleMiner, "挖矿结果数据非法", "nil数据")
				continue
			}

			if manversion.VersionCmp(self.curVersion, manversion.VersionAIMine) >= 0 {
				self.handlerV2.foundHandle(minedResult)
			} else {
				self.foundHandle(minedResult)
			}

		case agent := <-self.registerAgentCh:
			self.addAgent(agent)

		case agent := <-self.unRegisterAgentCh:
			self.delAgent(agent)

		case err := <-self.localMiningRequestSub.Err():
			log.Error(ModuleMiner, "localMiningRequestSub err", err)
			return
		case err := <-self.miningRequestSub.Err():
			log.Error(ModuleMiner, "miningRequestSub err", err)
			return
		case err := <-self.roleUpdateSub.Err():
			log.Error(ModuleMiner, "roleUpdateSub err", err)
			return
		case err := <-self.v2MiningRequestSub.Err():
			log.Error(ModuleMiner, "v2MiningRequestSub err", err)
			return
		case <-self.quitCh:
			return
		}
	}
}
func (self *worker) RoleUpdatedMsgHandler(data *mc.RoleUpdatedMsg) {
	if data.SuperSeq > self.curSuperSeq {
		self.StopAgent()
		self.stopMineResultSender()
		self.mineReqCtrl.Clear()
		self.curSuperSeq = data.SuperSeq
		self.curVersion = data.Version
	} else if data.SuperSeq < self.curSuperSeq {
		return
	}

	switch manversion.VersionCmp(self.curVersion, data.Version) {
	case 1: // self.curVersion > data.Version
		log.Warn(ModuleMiner, "CA消息处理", "版本号回退,抛弃消息", "msg version", data.Version, "缓存version", self.curVersion)
		return
	case -1: // self.curVersion < data.Version return -1
		self.stopMineResultSender()
		self.curVersion = data.Version
	}

	if manversion.VersionCmp(self.curVersion, manversion.VersionDelta) == 0 && data.BlockNum+1 == manversion.VersionNumAIMine {
		// 版本切换零界点，使用新版本挖矿
		log.Trace(ModuleMiner, "CA身份消息处理", "版本切换零界点，使用新版本挖矿", "msg version", data.Version, "msg number", data.BlockNum)
		self.stopMineResultSender()
		self.curVersion = manversion.VersionAIMine
	}

	if manversion.VersionCmp(self.curVersion, manversion.VersionAIMine) >= 0 {
		// 新版本号，使用v2版处理流程
		self.handlerV2.RoleUpdatedMsgHandler(data)
		return
	}

	if data.BlockNum+1 > self.mineReqCtrl.curNumber {
		self.stopMineResultSender()
	}

	role := data.Role
	self.mineReqCtrl.SetNewNumber(data.BlockNum+1, role)
	canMining := self.mineReqCtrl.CanMining()
	log.Trace(ModuleMiner, "高度", data.BlockNum, "角色", role, "是否可以挖矿", canMining)
	if canMining {
		self.StartAgent()
		self.processMineReq()
	} else {
		self.StopAgent()
	}
}

func (self *worker) MiningRequestHandle(data *mc.HD_MiningReqMsg) {
	if nil == data || nil == data.Header {
		log.Error(ModuleMiner, "挖矿请求Msg", "nil")
		return
	}
	log.Trace(ModuleMiner, "请求高度", data.Header.Number)

	reqData, err := self.mineReqCtrl.AddMineReq(data.Header, nil, false)
	if err != nil {
		log.Error(ModuleMiner, "缓存挖矿请求", err)
		return
	}
	if reqData != nil {
		self.processAppointedMineReq(reqData)
	}
}

func (self *worker) BroadcastHashLocalMiningReqMsgHandle(req *mc.BlockData) {
	if nil == req || nil == req.Header {
		log.Error(ModuleMiner, "广播挖矿请求Msg", "nil")
		return
	}
	log.Trace(ModuleMiner, "广播请求", req.Header.Number)
	reqData, err := self.mineReqCtrl.AddMineReq(req.Header, req.Txs, true)
	if err != nil {
		log.Error(ModuleMiner, "缓存请求Err", err)
		return
	}

	if reqData != nil {
		self.processAppointedMineReq(reqData)
	}
}

func (self *worker) Stop() {
	close(self.quitCh)
}

func (self *worker) foundHandle(result *consensus.SealResult) {
	if result.Type != consensus.SealTypePow {
		log.Info(ModuleMiner, "挖矿结果类型异常", result.Type)
		return
	}

	header := result.Header
	cache, err := self.mineReqCtrl.SetMiningResult(header)
	if err != nil {
		log.Error(ModuleMiner, "结果保存失败", err)
		return
	}
	self.startMineResultSender(cache)
}

func (self *worker) setExtra(extra []byte) {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.extra = extra
}

func (self *worker) pending() (*types.Block, *state.StateDBManage) {
	self.currentMu.Lock()
	defer self.currentMu.Unlock()

	tx, rx := types.GetCoinTXRS(self.current.txs, self.current.receipts)
	cb := types.MakeCurencyBlock(tx, rx, nil)
	if atomic.LoadInt32(&self.mining) == 0 {
		return types.NewBlock(
			self.current.header,
			cb,
			nil,
		), self.current.state.Copy()
	}
	return self.current.Block, self.current.state.Copy()
}

func (self *worker) pendingBlock() *types.Block {
	self.currentMu.Lock()
	defer self.currentMu.Unlock()

	if self == nil {
		return nil
	}
	if self.current == nil {
		return nil
	}
	if self.current.header == nil {
		return nil
	}
	return types.NewBlock(
		self.current.header,
		nil,
		nil,
	)

}

func (self *worker) StartAgent() {
	self.mu.Lock()
	defer self.mu.Unlock()
	atomic.StoreInt32(&self.mining, 1)

	// spin up agents
	for agent := range self.agents {
		agent.Start()
	}
}

func (self *worker) StopAgent() {
	self.mu.Lock()
	defer self.mu.Unlock()
	if atomic.LoadInt32(&self.mining) != 0 {
		for agent := range self.agents {
			agent.Stop()
		}
	}
	atomic.StoreInt32(&self.mining, 0)
}

func (self *worker) Register(agent Agent) {
	self.registerAgentCh <- agent
}

func (self *worker) Unregister(agent Agent) {
	self.unRegisterAgentCh <- agent
}

func (self *worker) addAgent(agent Agent) {
	if atomic.LoadInt32(&self.mining) > 0 {
		agent.Start()
	}

	self.agents[agent] = struct{}{}
	agent.SetReturnCh(self.recv)

	if manversion.VersionCmp(self.curVersion, manversion.VersionAIMine) >= 0 {
		if isNilTask(self.handlerV2.curMineTask) {
			return
		}
		log.Debug(ModuleMiner, "add new agent", "当前有挖矿,向其推送挖矿任务V2", "number", self.handlerV2.curMineTask.MineNumber(),
			"type", self.handlerV2.curMineTask.TaskType(), "hash", self.handlerV2.curMineTask.MineHash().Hex())
		agent.Work() <- self.current
	} else {
		if curMineReq := self.mineReqCtrl.GetCurrentMineReq(); curMineReq != nil && curMineReq.mined == false {
			log.Debug(ModuleMiner, "add new agent", "当前有挖矿,向其推送挖矿任务", "number", curMineReq.header.Number, "hash", curMineReq.headerHash.Hex())
			agent.Work() <- self.current
		}
	}
}

func (self *worker) delAgent(agent Agent) {
	delete(self.agents, agent)
	agent.Stop()
}

// push sends a new work task to currently live miner agents.
func (self *worker) push(work *Work) {
	if atomic.LoadInt32(&self.mining) != 1 {
		return
	}
	for agent := range self.agents {
		if ch := agent.Work(); ch != nil {
			ch <- work
		}
	}
}

// makeCurrent creates a new environment for the current cycle.
func (self *worker) makeCurrent(header *types.Header, isBroadcastNode bool) error {
	work := &Work{
		header:          types.CopyHeader(header),
		isBroadcastNode: isBroadcastNode,
		mineType:        mineTaskTypePow,
		createdAt:       time.Now(),
	}

	work.header.Coinbase = ca.GetDepositAddress()

	self.current = work
	return nil
}

func (self *worker) CommitNewWork(header *types.Header, isBroadcastNode bool) {
	if manversion.VersionCmp(self.curVersion, manversion.VersionAIMine) >= 0 {
		log.Error(ModuleMiner, "调用CommitNewWork版本错误", self.curVersion)
		return
	}

	err := self.makeCurrent(header, isBroadcastNode)
	if err != nil {
		log.Error(ModuleMiner, "创建挖矿work失败", err)
		return
	}

	log.Info(ModuleMiner, "挖矿", "开始")

	//// Create the current work task and check any fork transitions needed
	work := self.current
	self.push(work)
}

func (self *worker) makeCurrentV2(task mineTask, version []byte) error {
	if task == nil {
		return errors.New("task is nil")
	}

	work := &Work{
		header: &types.Header{
			Number:     task.MineNumber(),
			ParentHash: task.MineHash(),
			Difficulty: task.MineDifficulty(),
			VrfValue:   task.VrfValue(),
			Version:    version,
			Coinbase:   ca.GetDepositAddress(),
			AICoinbase: ca.GetDepositAddress(),
		},
		isBroadcastNode: false,
		mineType:        task.TaskType(),
		createdAt:       time.Now(),
	}

	self.current = work
	return nil
}

func (self *worker) CommitNewWorkV2(task mineTask) {
	if manversion.VersionCmp(self.curVersion, manversion.VersionAIMine) < 0 {
		log.Error(ModuleMiner, "调用CommitNewWorkV2版本错误", self.curVersion)
		return
	}

	err := self.makeCurrentV2(task, []byte(self.curVersion))
	if err != nil {
		log.Error(ModuleMiner, "创建挖矿work v2失败", err)
		return
	}
	log.Info(ModuleMiner, "挖矿", "开始", "coinbase", self.current.header.Coinbase.Hex())

	//// Create the current work task and check any fork transitions needed
	work := self.current
	self.push(work)
}

func (self *worker) processAppointedMineReq(reqData *mineReqData) {
	if nil == reqData {
		return
	}

	if self.mineReqCtrl.CanMining() == false {
		return
	}

	if reqData.mined {
		log.Trace(ModuleMiner, "请求已完成，直接发送结果", reqData.headerHash.TerminalString())
		self.sendMineResultFunc(reqData, 0)
	} else {
		log.Trace(ModuleMiner, "接收请求，开始处理", reqData.headerHash.TerminalString())
		self.beginMine(reqData)
	}
}

func (self *worker) processMineReq() {
	reqData := self.mineReqCtrl.GetUnMinedReq()
	if reqData == nil {
		return
	}
	log.Trace(ModuleMiner, "开始挖矿", reqData.headerHash.TerminalString())
	self.beginMine(reqData)
}

func (self *worker) beginMine(reqData *mineReqData) {
	if nil == reqData {
		return
	}
	if atomic.LoadInt32(&self.mining) == 0 {
		return
	}

	if curMineReq := self.mineReqCtrl.GetCurrentMineReq(); curMineReq != nil {
		if curMineReq.mined == false && curMineReq.header.Time.Cmp(reqData.header.Time) >= 0 {
			log.Debug(ModuleMiner, "beginMine", "当前挖矿时间较大，不处理本次挖矿",
				"当前挖矿header时间", curMineReq.header.Time, "请求挖矿header时间", reqData.header.Time)
			return
		}
	}

	if err := self.mineReqCtrl.SetCurrentMineReq(reqData.headerHash); err != nil {
		log.Error(ModuleMiner, "保存挖矿请求:", err)
		return
	}
	self.CommitNewWork(reqData.header, reqData.isBroadcastReq)
}

func (self *worker) startMineResultSender(data *mineReqData) {
	self.stopMineResultSender()
	sender, err := common.NewResendMsgCtrl(data, self.sendMineResultFunc, manparams.MinerResultSendInterval, 0)
	if err != nil {
		log.Error(ModuleMiner, "创建挖矿结果发送器", "失败", "err", err)
		return
	}
	log.Trace(ModuleMiner, "创建挖矿结果发送器", "成功", "高度", data.header.Number.Uint64(), "hash", data.headerHash.TerminalString())
	self.mineResultSender = sender
}

func (self *worker) stopMineResultSender() {
	if self.mineResultSender == nil {
		return
	}
	self.mineResultSender.Close()
	self.mineResultSender = nil
	log.Trace(ModuleMiner, "挖矿结果发送器", "停止")
}

func (self *worker) sendMineResultFunc(data interface{}, times uint32) {
	resultData, OK := data.(*mineReqData)
	if !OK {
		log.Error(ModuleMiner, "发出挖矿结果", "反射消息失败", "次数", times)
		return
	}

	if nil == resultData || nil == resultData.header || resultData.mined == false {
		log.Error(ModuleMiner, "发出挖矿结果", "入参错误", "次数", times)
		return
	}

	if err := resultData.ResendMineResult(time.Now().Unix()); err != nil {
		log.Error(ModuleMiner, "发出挖矿结果", "发送挖矿结果失败", "原因", err, "次数", times)
		return
	}

	if resultData.isBroadcastReq {
		rsp := &mc.HD_BroadcastMiningRspMsg{
			BlockMainData: &mc.BlockData{
				Header: resultData.header,
				Txs:    resultData.txs,
			},
		}
		log.Trace(ModuleMiner, "广播挖矿结果", "发送", "交易数量", len(types.GetTX(rsp.BlockMainData.Txs)), "次数", times, "高度", rsp.BlockMainData.Header.Number)
		self.hd.SendNodeMsg(mc.HD_BroadcastMiningRsp, rsp, common.RoleValidator, nil)
	} else {
		rsp := &mc.HD_MiningRspMsg{
			BlockHash:  resultData.headerHash,
			Difficulty: resultData.mineDiff,
			Number:     resultData.header.Number.Uint64(),
			Nonce:      resultData.header.Nonce,
			Coinbase:   resultData.header.Coinbase,
			MixDigest:  resultData.header.MixDigest,
			Signatures: resultData.header.Signatures}

		self.hd.SendNodeMsg(mc.HD_MiningRsp, rsp, common.RoleValidator|common.RoleBroadcast, nil)
		log.Trace(ModuleMiner, "挖矿结果", "发送", "hash", rsp.BlockHash.TerminalString(), "次数", times, "高度", rsp.Number, "Nonce", rsp.Nonce)
	}
}
