// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php


package miner

import (
	"math/big"
	"sync"
	"sync/atomic"
	"time"

	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/consensus"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/event"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/params"
	"gopkg.in/fatih/set.v0"
	"github.com/matrix/go-matrix/msgsend"
	"github.com/matrix/go-matrix/params/manparams"
)

const (
	resultQueueSize   = 10
	chainHeadChanSize = 10
)

// Agent can register themself with the worker
type Agent interface {
	Work() chan<- *Work
	SetReturnCh(chan<- *types.Header)
	Stop()
	Start()
	GetHashRate() int64
}

// Work is the workers current environment and holds
// all of the current state information
type Work struct {
	config *params.ChainConfig
	signer types.Signer

	state     *state.StateDB // apply state changes here
	ancestors *set.Set       // ancestor set (used for checking uncle parent validity)
	family    *set.Set       // family set (used for checking uncle invalidity)
	uncles    *set.Set       // uncle set
	tcount    int            // tx count in cycle

	Block *types.Block // the new block

	header   *types.Header
	txs      []types.SelfTransaction
	receipts []*types.Receipt

	createdAt time.Time

	threadNum       int
	isBroadcastNode bool
}

type Result struct {
	Difficulty *big.Int
	Header     *types.Header
}

// worker is the main object which takes care of applying messages to the new state
type worker struct {
	config *params.ChainConfig
	engine consensus.Engine

	mu sync.Mutex

	// update loop
	mux *event.TypeMux

	agents map[Agent]struct{}
	recv   chan *types.Header

	extra []byte

	currentMu sync.Mutex
	current   *Work
	// atomic status counters
	mining int32
	atWork int32

	roleUpdateCh          chan *mc.RoleUpdatedMsg
	roleUpdateSub         event.Subscription
	miningRequestCh       chan *mc.HD_MiningReqMsg
	miningRequestSub      event.Subscription
	localMiningRequestCh  chan *mc.BlockGenor_BroadcastMiningReqMsg
	localMiningRequestSub event.Subscription
	mineReqCtrl           *mineReqCtrl
	hd                    *msgsend.HD
	mineResultSender      *common.ResendMsgCtrl
}

func newWorker(config *params.ChainConfig, engine consensus.Engine, validatorReader consensus.ValidatorReader, dposEngine consensus.DPOSEngine, mux *event.TypeMux, hd *msgsend.HD) (*worker, error) {
	worker := &worker{
		config: config,
		engine: engine,
		mux:    mux,

		agents: make(map[Agent]struct{}),

		miningRequestCh:      make(chan *mc.HD_MiningReqMsg, 100),
		roleUpdateCh:         make(chan *mc.RoleUpdatedMsg, 100),
		recv:                 make(chan *types.Header, resultQueueSize),
		localMiningRequestCh: make(chan *mc.BlockGenor_BroadcastMiningReqMsg, 100),
		mineReqCtrl:          newMinReqCtrl(dposEngine, validatorReader),
		hd:                   hd,
		mineResultSender:     nil,
	}

	atomic.StoreInt32(&worker.mining, 0)

	err := worker.init_SubscribeEvent()
	if err != nil {
		log.Error(ModuleMiner, "worker创建失败", err)
		return nil, err
	}
	go worker.update()
	go worker.wait()
	log.INFO(ModuleMiner, "worker创建成功", err)
	return worker, nil
}
func (self *worker) init_SubscribeEvent() error {
	var err error

	self.localMiningRequestSub, err = mc.SubscribeEvent(mc.HD_BroadcastMiningReq, self.localMiningRequestCh) //广播节点
	if err != nil {
		log.Error(ModuleMiner, "广播节点挖矿请求订阅失败", err)
		return err
	} else {
		log.INFO(ModuleMiner, "广播节点挖矿请求订阅成功", "")
	}

	self.roleUpdateSub, err = mc.SubscribeEvent(mc.CA_RoleUpdated, self.roleUpdateCh) //身份到达
	if err != nil {
		log.Error(ModuleMiner, "身份更新订阅失败", err)
		return err
	} else {
		log.INFO(ModuleMiner, "身份更新订阅成功", "")
	}

	self.miningRequestSub, err = mc.SubscribeEvent(mc.HD_MiningReq, self.miningRequestCh) //挖矿请求
	if err != nil {
		log.Error(ModuleMiner, "普通矿工挖矿请求订阅失败", err)
		return err
	} else {
		log.INFO(ModuleMiner, "普通矿工挖矿请求订阅成功", err)
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
	}()

	for {
		select {
		case roleData := <-self.roleUpdateCh:
			self.RoleUpdatedMsgHandler(roleData)

		case minerReqData := <-self.miningRequestCh:
			self.MiningRequestHandle(minerReqData)

		case data := <-self.localMiningRequestCh:
			self.BroadcastHashLocalMiningReqMsgHandle(data.BlockMainData)

		case <-self.localMiningRequestSub.Err():
			return
		case <-self.miningRequestSub.Err():
			return
		case <-self.roleUpdateSub.Err():
			return
		}
	}
}
func (self *worker) RoleUpdatedMsgHandler(data *mc.RoleUpdatedMsg) {
	if data.BlockNum+1 > self.mineReqCtrl.curNumber {
		self.stopMineResultSender()
	}
	self.mineReqCtrl.SetNewNumber(data.BlockNum+1, data.Role)
	canMining := self.mineReqCtrl.CanMining()
	log.INFO(ModuleMiner, "更新高度及身份", "完成", "高度", data.BlockNum, "角色", data.Role, "是否可以挖矿", canMining)
	if canMining {
		self.StartAgent()
		self.processMineReq()
	} else {
		self.StopAgent()
	}
}

func (self *worker) MiningRequestHandle(data *mc.HD_MiningReqMsg) {
	if nil == data || nil == data.Header {
		log.ERROR(ModuleMiner, "接受挖矿请求", "消息为nil")
		return
	}
	log.INFO(ModuleMiner, "接收挖矿请求,高度", data.Header.Number, "难度值", data.Header.Difficulty)

	reqData, err := self.mineReqCtrl.AddMineReq(data.Header, nil, false)
	if err != nil {
		log.ERROR(ModuleMiner, "接受挖矿请求", "缓存挖矿请求错误", "err", err)
		return
	}

	if reqData != nil {
		self.processAppointedMineReq(reqData)
	}
}

func (self *worker) BroadcastHashLocalMiningReqMsgHandle(req *mc.BlockData) {
	if nil == req || nil == req.Header {
		log.ERROR(ModuleMiner, "接受广播挖矿请求", "消息为nil")
		return
	}
	log.INFO(ModuleMiner, "接受广播挖矿请求 header--difficult", req.Header.Difficulty.Uint64())
	reqData, err := self.mineReqCtrl.AddMineReq(req.Header, req.Txs, true)
	if err != nil {
		log.ERROR(ModuleMiner, "接受广播挖矿请求", "缓存挖矿请求错误", "err", err)
		return
	}

	if reqData != nil {
		self.processAppointedMineReq(reqData)
	}
}

func (self *worker) wait() {
	for {
		for header := range self.recv {
			atomic.AddInt32(&self.atWork, -1)

			if header == nil {
				continue
			}
			log.Info(ModuleMiner, "挖矿成功, Nonce", header.Nonce, "难度", header.Difficulty, "高度", header.Number.Uint64())
			self.foundHandle(header)
		}
	}
}

func (self *worker) foundHandle(header *types.Header) {
	cache, err := self.mineReqCtrl.SetMiningResult(header)
	if err != nil {
		log.ERROR(ModuleMiner, "挖矿结果处理", "保存失败", "err", err)
		return
	}
	self.startMineResultSender(cache)
}

func (self *worker) setExtra(extra []byte) {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.extra = extra
}

func (self *worker) pending() (*types.Block, *state.StateDB) {
	self.currentMu.Lock()
	defer self.currentMu.Unlock()

	if atomic.LoadInt32(&self.mining) == 0 {
		return types.NewBlock(
			self.current.header,
			self.current.txs,
			nil,
			self.current.receipts,
		), self.current.state.Copy()
	}
	return self.current.Block, self.current.state.Copy()
}

func (self *worker) pendingBlock() *types.Block {
	self.currentMu.Lock()
	defer self.currentMu.Unlock()

	if atomic.LoadInt32(&self.mining) == 0 {
		return types.NewBlock(
			self.current.header,
			self.current.txs,
			nil,
			self.current.receipts,
		)
	}

	self.currentMu.Lock()
	defer self.currentMu.Unlock()
	return self.current.Block
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
	atomic.StoreInt32(&self.atWork, 0)
}

func (self *worker) Register(agent Agent) {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.agents[agent] = struct{}{}
	agent.SetReturnCh(self.recv)
}

func (self *worker) Unregister(agent Agent) {
	self.mu.Lock()
	defer self.mu.Unlock()
	delete(self.agents, agent)
	agent.Stop()
}

// push sends a new work task to currently live miner agents.
func (self *worker) push(work *Work) {
	if atomic.LoadInt32(&self.mining) != 1 {
		return
	}
	for agent := range self.agents {
		atomic.AddInt32(&self.atWork, 1)
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
	}

	work.header.Coinbase = ca.GetAddress()

	self.current = work
	return nil
}

func (self *worker) CommitNewWork(header *types.Header, isBroadcastNode bool) {
	err := self.makeCurrent(header, isBroadcastNode)
	if err != nil {
		log.Error(ModuleMiner, "创建挖矿work失败", err)
		return
	}
	log.Error(ModuleMiner, "挖矿work", "开始")

	//// Create the current work task and check any fork transitions needed
	work := self.current

	self.push(work)
}

func (self *worker) processAppointedMineReq(reqData *mineReqData) {
	if nil == reqData {
		return
	}

	if self.mineReqCtrl.CanMining() == false {
		log.INFO(ModuleMiner, "processAppointedMineReq", "当前身份不可挖矿", "身份", self.mineReqCtrl.role.String(), "高度", self.mineReqCtrl.curNumber)
		return
	}

	if reqData.mined {
		log.INFO(ModuleMiner, "processAppointedMineReq", "请求已挖完，直接发送挖矿结果", "hash", reqData.headerHash.TerminalString())
		self.sendMineResultFunc(reqData, 0)
	} else {
		log.INFO(ModuleMiner, "processAppointedMineReq", "开始挖矿", "hash", reqData.headerHash.TerminalString())
		self.beginMine(reqData)
	}
}

func (self *worker) processMineReq() {
	reqData := self.mineReqCtrl.GetUnMinedReq()
	if reqData == nil {
		log.INFO(ModuleMiner, "processMineReq", "获取未处理过的请求失败")
		return
	}
	log.INFO(ModuleMiner, "processMineReq", "开始挖矿", "hash", reqData.headerHash.TerminalString())
	self.beginMine(reqData)
}

func (self *worker) beginMine(reqData *mineReqData) {
	if nil == reqData {
		return
	}
	if atomic.LoadInt32(&self.mining) == 0 {
		log.INFO(ModuleMiner, "beginMine", "当前不处于mining状态，不处理")
		return
	}

	if curMineReq := self.mineReqCtrl.GetCurrentMineReq(); curMineReq != nil {
		if curMineReq.mined == false && curMineReq.header.Time.Cmp(reqData.header.Time) >= 0 {
			log.INFO(ModuleMiner, "beginMine", "当前挖矿时间较大，不处理本次挖矿",
				"当前挖矿header时间", curMineReq.header.Time, "请求挖矿header时间", reqData.header.Time)
			return
		}
	}

	if err := self.mineReqCtrl.SetCurrentMineReq(reqData.headerHash); err != nil {
		log.ERROR(ModuleMiner, "beginMine", "保存当前挖矿请求错误", "err", err)
		return
	}
	self.CommitNewWork(reqData.header, reqData.isBroadcastReq)
}

func (self *worker) startMineResultSender(data *mineReqData) {
	self.stopMineResultSender()
	sender, err := common.NewResendMsgCtrl(data, self.sendMineResultFunc, manparams.MinerResultSendInterval, 0)
	if err != nil {
		log.ERROR(ModuleMiner, "创建挖矿结果发送器", "失败", "err", err)
		return
	}
	log.INFO(ModuleMiner, "创建挖矿结果发送器", "成功", "高度", data.header.Number.Uint64(), "hash", data.headerHash.TerminalString())
	self.mineResultSender = sender
}

func (self *worker) stopMineResultSender() {
	if self.mineResultSender == nil {
		return
	}
	self.mineResultSender.Close()
	self.mineResultSender = nil
	log.INFO(ModuleMiner, "挖矿结果发送器", "停止")
}

func (self *worker) sendMineResultFunc(data interface{}, times uint32) {
	resultData, OK := data.(*mineReqData)
	if !OK {
		log.ERROR(ModuleMiner, "发出挖矿结果", "反射消息失败", "次数", times)
		return
	}

	if nil == resultData || nil == resultData.header || resultData.mined == false {
		log.ERROR(ModuleMiner, "发出挖矿结果", "入参错误", "次数", times)
		return
	}

	if err := resultData.ResendMineResult(time.Now().Unix()); err != nil {
		log.WARN(ModuleMiner, "发出挖矿结果", "发送挖矿结果失败", "原因", err, "次数", times)
		return
	}

	if resultData.isBroadcastReq {
		rsp := &mc.HD_BroadcastMiningRspMsg{
			BlockMainData: &mc.BlockData{
				Header: resultData.header,
				Txs:    resultData.txs,
			},
		}
		log.INFO(ModuleMiner, "广播挖矿结果", "发送", "交易数量", rsp.BlockMainData.Txs.Len(), "次数", times, "高度", rsp.BlockMainData.Header.Number)
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
		log.INFO(ModuleMiner, "挖矿结果", "发送", "hash", rsp.BlockHash.TerminalString(), "次数", times, "高度", rsp.Number, "Nonce", rsp.Nonce)
	}
}
