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
	"github.com/matrix/go-matrix/core"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/event"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/msgsend"
	"github.com/matrix/go-matrix/params"
	"gopkg.in/fatih/set.v0"
)

const (
	resultQueueSize   = 10
	chainHeadChanSize = 10
)

// Agent can register themself with the worker
type Agent interface {
	Work() chan<- *Work
	SetReturnCh(chan<- *Result)
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
	txs      []*types.Transaction
	receipts []*types.Receipt

	createdAt time.Time

	difficultyList  []*big.Int
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

	wg sync.WaitGroup

	agents map[Agent]struct{}
	recv   chan *Result

	coinbase common.Address
	extra    []byte

	currentMu sync.Mutex
	current   *Work

	unconfirmed *unconfirmedBlocks // set of locally mined blocks pending canonicalness confirmations

	// atomic status counters
	mining int32
	atWork int32

	dposEngine consensus.DPOSEngine
	bc         *core.BlockChain
	msgcenter  *mc.Center

	roleUpdateCh          chan *mc.RoleUpdatedMsg
	roleUpdateSub         event.Subscription
	miningRequestCh       chan *mc.HD_MiningReqMsg
	miningRequestSub      event.Subscription
	localMiningRequestCh  chan *mc.BlockGenor_BroadcastMiningReqMsg
	localMiningRequestSub event.Subscription

	currentHeight   uint64
	currentCAHeight uint64
	currentRole     common.RoleType
	currentHeader   *types.Header
	queue           []*mc.HD_MiningReqMsg

	minningReq mc.BlockData
	addr       common.Address
	hd         *msgsend.HD
	ca         *ca.Identity
}

func newWorker(config *params.ChainConfig, engine consensus.Engine, dposEngine consensus.DPOSEngine, coinbase common.Address, mux *event.TypeMux, hd *msgsend.HD, ca *ca.Identity) (*worker, error) {
	worker := &worker{
		config: config,
		engine: engine,
		mux:    mux,

		miningRequestCh:      make(chan *mc.HD_MiningReqMsg, 100),
		roleUpdateCh:         make(chan *mc.RoleUpdatedMsg, 100),
		recv:                 make(chan *Result, resultQueueSize),
		localMiningRequestCh: make(chan *mc.BlockGenor_BroadcastMiningReqMsg, 100),

		coinbase: coinbase,
		agents:   make(map[Agent]struct{}),

		dposEngine: dposEngine,

		currentHeight:   0,
		currentCAHeight: 0,
		queue:           make([]*mc.HD_MiningReqMsg, 0),
		currentRole:     common.RoleDefault, //从CA要当前身份！！！！！！！！！！！！！！！！！！！！！！！！！！！！！！！！！！！！！！！！！！！！！！！！！！！！！！！！！！！！！！！！！！
		hd:              hd,
	}

	atomic.StoreInt32(&worker.mining, 0)

	log.INFO(ModuleWork, "CurrentRole:", worker.currentRole)

	err := worker.init_SubscribeEvent()
	if err != nil {
		log.Error(ModuleWork, "worker创建失败", err)
		return nil, err
	}
	go worker.update()
	go worker.wait()
	log.INFO(ModuleWork, "worker创建成功", err)
	return worker, nil
}
func (self *worker) init_SubscribeEvent() error {
	var err error

	self.localMiningRequestSub, err = mc.SubscribeEvent(mc.HD_BroadcastMiningReq, self.localMiningRequestCh) //广播节点
	if err != nil {
		log.Error(ModuleWork, "广播节点挖矿请求订阅失败", err)
		return err
	} else {
		log.INFO(ModuleWork, "广播节点挖矿请求订阅成功", "")
	}

	self.roleUpdateSub, err = mc.SubscribeEvent(mc.CA_RoleUpdated, self.roleUpdateCh) //身份到达
	if err != nil {
		log.Error(ModuleWork, "身份更新订阅失败", err)
		return err
	} else {
		log.INFO(ModuleWork, "身份更新订阅成功", "")
	}

	self.miningRequestSub, err = mc.SubscribeEvent(mc.HD_MiningReq, self.miningRequestCh) //挖矿请求
	if err != nil {
		log.Error(ModuleWork, "普通矿工挖矿请求订阅失败", err)
		return err
	} else {
		log.INFO(ModuleWork, "普通矿工挖矿请求订阅成功", err)
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
			log.INFO(ModuleWork, "接收身份更新消息，高度", roleData.BlockNum, "身份", roleData.Role, "Leader", common.Bytes2Hex(roleData.Leader[:7]))
			self.roleUpdatedMsgHandler(roleData)

		case minerReqData := <-self.miningRequestCh:
			self.MiningRequestHandle(minerReqData)

		case data := <-self.localMiningRequestCh:
			log.INFO(ModuleWork, "received localMiningReq", data, "交易数量", data.BlockMainData.Txs.Len())
			self.minningReq = *data.BlockMainData
			self.BroadcastHashLocalMiningReqMsgHandle(data.BlockMainData.Header)

		case <-self.localMiningRequestSub.Err():
			return
		case <-self.miningRequestSub.Err():
			return
		case <-self.roleUpdateSub.Err():
			return

		}
	}
}
func (self *worker) roleUpdatedMsgHandler(data *mc.RoleUpdatedMsg) {

	if data.BlockNum > self.currentHeight {
		log.INFO(ModuleWork, "新区块已经生产，停止旧的. 新高度", data.BlockNum, "挖矿请求高度", self.currentHeight)
		self.Stop()
	}

	self.currentCAHeight = data.BlockNum

	self.HandleQueue()

	self.currentRole = data.Role
}

func (self *worker) MiningRequestHandle(data *mc.HD_MiningReqMsg) {

	if data.Header.Number.Uint64() > (self.currentCAHeight + 1) {
		self.queue = append(self.queue, data)
		return
	}
	if data.Header.Number.Uint64() != self.currentCAHeight+1 {
		return
	}
	//以下全是挖矿高度==当前CA高度+1
	log.INFO(ModuleWork, "接收挖矿请求,高度", data.Header.Number)
	err := self.CheckMiningReqError(data.Header)
	if err != nil {
		log.Error(ModuleWork, "挖矿请求错误", err)
		return
	}
	log.INFO(ModuleMiner, "处理挖矿请求 高度", data.Header.Number, "难度值", data.Header.Difficulty)
	self.currentHeader = types.CopyHeader(data.Header)
	self.currentHeight = data.Header.Number.Uint64()

	self.Stop()
	self.Start()

	difflist := self.CalDifflist(data.Header.Difficulty.Uint64())
	self.CommitNewWork(data.Header, difflist, false)

}
func (self *worker) BroadcastHashLocalMiningReqMsgHandle(header *types.Header) {
	log.INFO(ModuleWork, "BroadcastHashLocalMiningReqMsgHandle header--difficult", header.Difficulty.Uint64())
	err := self.CheckLocalMiningReqError(header)
	if err != nil {
		log.INFO(ModuleWork, "BroadcastHashLocalMiningReqMsgHandle CheckLocalMiningReqError", err)
		return
	}
	self.currentHeader = types.CopyHeader(header)
	self.currentHeight = header.Number.Uint64()
	self.Stop()
	self.Start()

	//log.INFO(ModuleWork, "CommitNewWork", "begin")

	difficultyList := make([]*big.Int, 1)
	difficultyList[0] = big.NewInt(int64(1))
	self.CommitNewWork(header, difficultyList, true)
}

func (self *worker) checkCurrentHeader() bool {
	if self.currentHeader != nil && self.currentHeader.Difficulty != nil {
		return true
	}
	return false
}
func (self *worker) foundHandle(data *Result) {

	if self.checkCurrentHeader() == false {
		log.Error(ModuleWork, "挖矿结果错误", data)
		return
	}

	if data.Difficulty.Cmp(self.currentHeader.Difficulty) == 0 {
		//log.INFO(ModuleWork, "FoundHandle seal successfully!!!!!!!!!!!!! height", data.Header.Number)
		log.INFO(ModuleWork, "挖矿成功，高度", data.Header.Number)
		self.Stop()
	}

	//self.hd.SendNodeMsg(mc.MD_MiningRsp, &mc.MD_MiningRspMsg{
	//	Blockhash:  data.Header.Hash(),
	//	Difficulty: data.Difficulty,
	//	Nonce:      data.Header.Nonce,
	//	Coinbase:   self.ca.SelfAddress(),
	//	MixDigest:  data.Header.MixDigest,
	//	From:       "",
	//	Signatures: data.Header.Signatures,
	//}, common.RoleValidator, nil)

	data.Header.Coinbase = ca.GetAddress()
	var rsp = mc.HD_MiningRspMsg{
		Blockhash:  data.Header.HashNoSignsAndNonce(),
		Difficulty: data.Difficulty,
		Number:     data.Header.Number.Uint64(),
		Nonce:      data.Header.Nonce,
		Coinbase:   ca.GetAddress(),
		MixDigest:  data.Header.MixDigest,
		Signatures: data.Header.Signatures}

	log.INFO(ModuleWork, "Rsp Header", data.Header)
	log.INFO(ModuleWork, "Rsp Hash", rsp.Blockhash)

	self.hd.SendNodeMsg(mc.HD_MiningRsp, &rsp, common.RoleValidator, nil)
	self.hd.SendNodeMsg(mc.HD_MiningRsp, &rsp, common.RoleBroadcast, nil)

	/*
		err := self.msgcenter.PostEvent(mc.MD_MiningRsp, &mc.MD_MiningRspMsg{
			Blockhash:  data.Header.Hash(),
			Difficulty: data.Difficulty,
			Nonce:      data.Header.Nonce,
			Coinbase:   data.Header.Coinbase,
			MixDigest:  data.Header.MixDigest,
			From:       self.addr,
			Signatures: data.Header.Signatures,
		})
		if err != nil {
			log.INFO(ModuleWork, "foundHandle PostEvent err ", err)
		}
	*/
	log.INFO(ModuleWork, "foundHandle", mc.HD_MiningRspMsg{
		Blockhash:  data.Header.Hash(),
		Difficulty: data.Difficulty,
		Nonce:      data.Header.Nonce,
		Coinbase:   data.Header.Coinbase,
		MixDigest:  data.Header.MixDigest,
		Signatures: data.Header.Signatures,
	})
	log.INFO(ModuleWork, "RSP, BlockHash", rsp.Blockhash[:15], "Diff", rsp.Difficulty, "Nonce", rsp.Nonce)
	log.INFO(ModuleWork, "RSP, CoinBase", common.Bytes2Hex(rsp.Coinbase[:7]), "Digest", rsp.MixDigest)

}

func (self *worker) broadcastHashFoundMsgHandle(data *Result) error {
	log.INFO(ModuleWork, "broadcastHashFoundMsgHandle--difficult", data.Header.Difficulty)
	if data == nil {
		log.Info("Parameter is invalid!")
		return invalidParameter
	}

	if data.Difficulty.Cmp(self.currentHeader.Difficulty) == 0 {
		log.Info("miner", "broadcast Node Worker Stop mine")

		self.Stop()
	}
	data.Header.Difficulty = self.minningReq.Header.Difficulty
	msg := &mc.HD_BroadcastMiningRspMsg{
		BlockMainData: &mc.BlockData{
			Header: data.Header,
			Txs:    self.minningReq.Txs,
		},
	}
	self.sendHeaderToValidator(msg)
	return nil
}

func (self *worker) sendHeaderToValidator(msg *mc.HD_BroadcastMiningRspMsg) {
	log.INFO(ModuleWork, "广播挖矿结果", "发送", "交易数量", msg.BlockMainData.Txs.Len(), "交易数据", msg.BlockMainData.Txs, "头leader", msg.BlockMainData.Header.Leader)
	//self.msgcenter.PostEvent(mc.BD_MiningRsp, msg)
	self.hd.SendNodeMsg(mc.HD_BroadcastMiningRsp, msg, common.RoleValidator, nil)
	/*
		err := self.msgcenter.PostEvent(mc.BD_MiningReq, &msg)
		if err != nil {
			log.Error("miner", "Send BD_MiningRspMsg failed", err)
		}
	*/
}

func (self *worker) wait() {
	for {
		for result := range self.recv {
			atomic.AddInt32(&self.atWork, -1)

			if result == nil {
				continue
			}
			log.Info(ModuleWork, "挖矿成功, Nonce", result.Header.Nonce, "难度", result.Difficulty)
			if self.currentRole == common.RoleBroadcast {
				self.broadcastHashFoundMsgHandle(result)
			} else {
				self.foundHandle(result)
			}
		}
	}
}

func (self *worker) CalDifflist(difficulty uint64) []*big.Int {

	difflist := make([]*big.Int, 0)
	//if difficulty == 1 {
	//	difflist = []*big.Int{big.NewInt(int64(1))}
	//
	//}
	for i := 0; i < len(params.Difficultlist); i++ {
		temp := difficulty / params.Difficultlist[i]
		difflist = append(difflist, big.NewInt(int64(temp)))
	}
	log.INFO(ModuleWork, "难度列表", difflist)
	return difflist
}

func (self *worker) CheckMiningReqError(header *types.Header) error {

	if common.RoleMiner != self.currentRole {
		log.Error(ModuleWork, "当前不是矿工,", "不处理")
		return currentNotMiner
	}
	if header.Number.Uint64() <= self.currentHeight {
		log.Error(ModuleWork, "挖矿请求块高过小, 请求高度", header.Number.Uint64(), "当前处理高度", self.currentHeader.Number.Uint64())
		return smallThanCurrentHeightt
	}
	if header.Difficulty.Uint64() == 0 {
		log.Error(ModuleWork, "挖矿请求难度值为0", "")
		return difficultyIsZero
	}

	err := self.dposEngine.VerifyBlock(header)
	if err != nil {
		log.Error(ModuleWork, "挖矿请求DPOS验证错误", err)
		return err
	}
	return nil
}
func (self *worker) CheckLocalMiningReqError(header *types.Header) error {
	if common.RoleBroadcast != self.currentRole {
		return currentRoleIsNotBroadcast
	}
	if header.Number.Uint64() <= self.currentHeight {
		return smallThanCurrentHeightt
	}
	return nil
}

/////original
func (self *worker) setManerbase(addr common.Address) {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.coinbase = addr
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

func (self *worker) Start() {
	self.mu.Lock()
	defer self.mu.Unlock()

	atomic.StoreInt32(&self.mining, 1)

	// spin up agents
	for agent := range self.agents {
		agent.Start()
	}
}

func (self *worker) Stop() {
	self.wg.Wait()

	self.mu.Lock()
	defer self.mu.Unlock()
	if atomic.LoadInt32(&self.mining) == 1 {
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

//lcq todo
// makeCurrent creates a new environment for the current cycle.
func (self *worker) makeCurrent(header *types.Header, diffList []*big.Int, isBroadcastNode bool) error {
	work := &Work{
		header:          header,
		difficultyList:  diffList,
		isBroadcastNode: isBroadcastNode,
	}

	self.current = work
	return nil
}

func (self *worker) CommitNewWork(header *types.Header, difficultyList []*big.Int, isBroadcastNode bool) {
	err := self.makeCurrent(header, difficultyList, isBroadcastNode)
	if err != nil {
		log.Error(ModuleWork, "创建挖矿worker失败", err)
		return
	}
	log.Error(ModuleWork, "开始挖矿worker", "")

	//// Create the current work task and check any fork transitions needed
	work := self.current

	self.push(work)
}
func (self *worker) HandleQueue() {
	Need_Delete := make([]int, 0)
	if len(self.queue) == 0 {
		return
	}
	log.INFO(ModuleWork, "挖矿请求队列不为空,len", self.queue)
	for k, v := range self.queue {
		if v.Header.Number.Uint64() == self.currentCAHeight+1 {
			self.MiningRequestHandle(v)
			Need_Delete = append(Need_Delete, k)
		}
	}
	for i := len(Need_Delete) - 1; i >= 0; i-- {
		index := Need_Delete[i]
		self.queue = append(self.queue[:index], self.queue[index+1:]...)
	}

}
