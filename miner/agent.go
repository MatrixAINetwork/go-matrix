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

package miner

import (
	"sync"

	"github.com/matrix/go-matrix/consensus"
	"github.com/matrix/go-matrix/log"
)

type CpuAgent struct {
	mu sync.Mutex

	workCh chan *Work
	stop   chan uint64

	mapQuit map[uint64]chan uint64

	quitCurrentOp chan uint64
	stopMineCh    chan struct{}
	returnCh      chan<- *consensus.Result
	foundMsgCh    chan *consensus.FoundMsg

	chain  consensus.ChainReader
	engine consensus.Engine

	isMining int32 // isMining indicates whether the agent is currently mining
}

func NewCpuAgent(chain consensus.ChainReader, engine consensus.Engine) *CpuAgent {
	miner := &CpuAgent{
		chain:      chain,
		engine:     engine,
		stop:       make(chan uint64, 1),
		workCh:     make(chan *Work, 1),
		foundMsgCh: make(chan *consensus.FoundMsg, 1),
		stopMineCh: make(chan struct{}, 1),
		mapQuit:    make(map[uint64]chan uint64, 0),
	}
	return miner
}

func (self *CpuAgent) Work() chan<- *Work                      { return self.workCh }
func (self *CpuAgent) SetReturnCh(ch chan<- *consensus.Result) { self.returnCh = ch }

func (self *CpuAgent) Stop(num uint64) {
	/*
		if !atomic.CompareAndSwapInt32(&self.isMining, 1, 0) {

			return // agent already stopped
		}
	*/

	self.stop <- num
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
	/*
		if !atomic.CompareAndSwapInt32(&self.isMining, 0, 1) {

			return // agent already started
		}
	*/
	go self.update()
}

func (self *CpuAgent) update() {
out:
	for {
		select {
		case work := <-self.workCh:
			self.mu.Lock()
			log.INFO("问题定位", "高度", work.header.Number.Uint64())

			tempQuit := make(chan uint64, 1)
			self.mapQuit[work.header.Number.Uint64()] = tempQuit
			go self.mine(work, tempQuit)
			self.mu.Unlock()
		case num := <-self.stop:
			self.mu.Lock()

			if tempquit, ok := self.mapQuit[num]; ok {
				tempquit <- num
				delete(self.mapQuit, num)
			}

			self.mu.Unlock()
			log.Info("miner", "CpuAgent Stop Minning", "", "高度", num)

			break out
		}
	}
}

func (self *CpuAgent) mine(work *Work, stop <-chan uint64) {

	//if result, err := self.engine.Seal(self.chain, work.header, stop, work.difficultyList, work.isBroadcastNode); result != nil {
	//	self.returnCh <- &Result{result.Difficulty, result.Header}
	//} else {
	//	if err != nil {
	//		log.Warn("Block sealing failed", "err", err)
	//	}
	//	self.returnCh <- nil
	//}

	self.engine.Seal(self.chain, work.header, stop, self.returnCh, work.difficultyList, work.isBroadcastNode)

	/*
		for {
			select {
			case result := <-self.foundMsgCh:
				self.returnCh <- &Result{result.Difficulty, result.Header}
			case <-self.stopMineCh:
				log.Info("miner", "quit agent mine")
				return
			}
		}
	*/
}

func (self *CpuAgent) GetHashRate() int64 {
	if pow, ok := self.engine.(consensus.PoW); ok {
		return int64(pow.Hashrate())
	}
	return 0
}
