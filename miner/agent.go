// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php


package miner

import (
	"sync"

	"sync/atomic"

	"github.com/matrix/go-matrix/consensus"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/core/types"
)

type CpuAgent struct {
	mu sync.Mutex

	workCh chan *Work
	stop   chan struct{}

	quitCurrentOp chan struct{}
	returnCh      chan<- *types.Header

	chain  consensus.ChainReader
	engine consensus.Engine

	isMining int32 // isMining indicates whether the agent is currently mining
}

func NewCpuAgent(chain consensus.ChainReader, engine consensus.Engine) *CpuAgent {
	miner := &CpuAgent{
		chain:  chain,
		engine: engine,
		stop:   make(chan struct{}, 1),
		workCh: make(chan *Work, 1),
	}
	return miner
}

func (self *CpuAgent) Work() chan<- *Work                  { return self.workCh }
func (self *CpuAgent) SetReturnCh(ch chan<- *types.Header) { self.returnCh = ch }

func (self *CpuAgent) Stop() {
	if !atomic.CompareAndSwapInt32(&self.isMining, 1, 0) {
		return // agent already stopped
	}
	self.stop <- struct{}{}
done:
	// Empty work channel
	for {
		select {
		case <-self.workCh:
		default:
			break done
		}
	}
}

func (self *CpuAgent) Start() {
	if !atomic.CompareAndSwapInt32(&self.isMining, 0, 1) {
		return // agent already started
	}

	go self.update()
}

func (self *CpuAgent) update() {
out:
	for {
		select {
		case work := <-self.workCh:
			self.mu.Lock()
			if self.quitCurrentOp != nil {
				close(self.quitCurrentOp)
			}
			self.quitCurrentOp = make(chan struct{})
			go self.mine(work, self.quitCurrentOp)
			self.mu.Unlock()
		case <-self.stop:
			self.mu.Lock()
			if self.quitCurrentOp != nil {
				close(self.quitCurrentOp)
				self.quitCurrentOp = nil
			}
			self.mu.Unlock()
			log.Info("miner", "CpuAgent Stop Minning", "")
			break out
		}
	}
}

func (self *CpuAgent) mine(work *Work, stop <-chan struct{}) {
	if result, err := self.engine.Seal(self.chain, work.header, stop, work.isBroadcastNode); result != nil {
		log.Info("Successfully sealed new block", "number", result.Number, "hash", result.Hash())
		self.returnCh <- result
	} else {
		if err != nil {
			log.Warn("Block sealing failed", "err", err)
		}
		self.returnCh <- nil
	}
}

func (self *CpuAgent) GetHashRate() int64 {
	if pow, ok := self.engine.(consensus.PoW); ok {
		return int64(pow.Hashrate())
	}
	return 0
}
