// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php


// Package miner implements Matrix block creation and mining.
package miner

import (
	"errors"
	"fmt"
	"sync/atomic"

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
)

const (
	ModuleWork  = "Miner_Work"
	ModuleMiner = "Miner Miner"
)

var (
	invalidParameter          = errors.New("Parameter is invalid!")
	currentNotMiner           = errors.New("Current Not Miner")
	smallThanCurrentHeightt   = errors.New("Small than current height")
	LimitBroadcastRole        = errors.New("change Broadcast to others")
	currentRoleIsNotBroadcast = errors.New("current Role Is Not Broadcast")
	difficultyIsZero          = errors.New("difficulty Is Zero")
)

// Miner creates blocks and searches for proof-of-work values.
type Miner struct {
	mux *event.TypeMux

	worker *worker

	coinbase common.Address
	mining   int32
	bc       *core.BlockChain
	engine   consensus.Engine

	canStart    int32 // can start indicates whether we can start the mining operation
	shouldStart int32 // should start indicates whether we should start after sync

	currentRole common.RoleType
	msgcenter   *mc.Center
	hd          *msgsend.HD
	ca          *ca.Identity
}

func (s *Miner) Getworker() *worker { return s.worker }

func New(bc *core.BlockChain, config *params.ChainConfig, mux *event.TypeMux, engine consensus.Engine, dposEngine consensus.DPOSEngine, hd *msgsend.HD, ca *ca.Identity) (*Miner, error) {
	miner := &Miner{
		mux:    mux,
		engine: engine,

		canStart:    1,
		currentRole: common.RoleBroadcast,
		hd:          hd,
		ca:          ca,
	}
	var err error
	miner.worker, err = newWorker(config, engine, dposEngine, common.Address{}, mux, hd, ca)
	if err != nil {
		log.DEBUG(ModuleMiner, "创建work失败")
		return miner, err
	}
	miner.Register(NewCpuAgent(bc, engine))
	//go miner.update()
	log.DEBUG(ModuleMiner, "创建miner成功")
	log.INFO(ModuleMiner, "�󹤷��񴴽��ɹ�", nil)
	return miner, nil
}

/*
func (self *Miner) downloadStartEventHandler() {
	atomic.StoreInt32(&self.canStart, 0)
	if self.Mining() {
		self.Stop()
		atomic.StoreInt32(&self.shouldStart, 1)
		log.Info("Mining aborted due to sync")
	}

}

func (self *Miner) downloadDoneEventHandler() {
	shouldStart := atomic.LoadInt32(&self.shouldStart) == 1

	atomic.StoreInt32(&self.canStart, 1)
	atomic.StoreInt32(&self.shouldStart, 0)
	if shouldStart {
		self.Start(self.coinbase)
	}

}

// update keeps track of the downloader events. Please be aware that this is a one shot type of update loop.
// It's entered once and as soon as `Done` or `Failed` has been broadcasted the events are unregistered and
// the loop is exited. This to prevent a major security vuln where external parties can DOS you with blocks
// and halt your mining operation for as long as the DOS continues.
func (self *Miner) update() {
	events := self.mux.Subscribe(downloader.StartEvent{}, downloader.DoneEvent{}, downloader.FailedEvent{})
out:
	for ev := range events.Chan() {
		switch ev.Data.(type) {
		case downloader.StartEvent:
			self.downloadStartEventHandler()
		case downloader.DoneEvent, downloader.FailedEvent:
			self.downloadDoneEventHandler()
			// unsubscribe. we're only interested in this event once
			events.Unsubscribe()
			// stop immediately and ignore all further pending events
			break out
		}
	}
}*/

//Start
func (self *Miner) Start(coinbase common.Address) {
	atomic.StoreInt32(&self.shouldStart, 1)
	if self.currentRole != common.RoleBroadcast {
		self.worker.setManerbase(coinbase)
		self.coinbase = coinbase
	}
	if atomic.LoadInt32(&self.canStart) == 0 {
		log.Info("Network syncing, will start miner afterwards")
		return
	}
	atomic.StoreInt32(&self.mining, 1)

	log.Info("Starting mining operation")
	if self.currentRole != common.RoleBroadcast {
		self.worker.Start()
	}
}

func (self *Miner) Stop() {
	// todo:
	if self.currentRole != common.RoleBroadcast {
		self.worker.Stop()
		atomic.StoreInt32(&self.mining, 0)
		atomic.StoreInt32(&self.shouldStart, 0)
	}
}

func (self *Miner) Register(agent Agent) {
	if self.Mining() {
		agent.Start()
	}
	self.worker.Register(agent)
}

/*

func (self *Miner) Unregister(agent Agent) {

	self.worker.Unregister(agent)
}
*/

func (self *Miner) Mining() bool {
	return atomic.LoadInt32(&self.Getworker().mining) > 0
}

func (self *Miner) HashRate() (tot int64) {
	if pow, ok := self.engine.(consensus.PoW); ok {
		tot += int64(pow.Hashrate())
	}
	// do we care this might race? is it worth we're rewriting some
	// aspects of the worker/locking up agents so we can get an accurate
	// hashrate?
	for agent := range self.worker.agents {
		if _, ok := agent.(*CpuAgent); !ok {
			tot += agent.GetHashRate()
		}
	}
	return
}

func (self *Miner) SetExtra(extra []byte) error {
	if uint64(len(extra)) > params.MaximumExtraDataSize {
		return fmt.Errorf("Extra exceeds max length. %d > %v", len(extra), params.MaximumExtraDataSize)
	}
	self.worker.setExtra(extra)
	return nil
}

// Pending returns the currently pending block and associated state.
func (self *Miner) Pending() (*types.Block, *state.StateDB) {
	return self.worker.pending()
}

/*
// PendingBlock returns the currently pending block.
//
// Note, to access both the pending block and the pending state
// simultaneously, please use Pending(), as the pending state can
// change between multiple method calls
*/
func (self *Miner) PendingBlock() *types.Block {
	return self.worker.pendingBlock()
}

func (self *Miner) SetManerbase(addr common.Address) {
	self.coinbase = addr
	self.worker.setManerbase(addr)
}
