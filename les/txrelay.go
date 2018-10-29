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

package les

import (
	"sync"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/types"
)

type ltrInfo struct {
	tx     *types.Transaction
	sentTo map[*peer]struct{}
}

type LesTxRelay struct {
	txSent       map[common.Hash]*ltrInfo
	txPending    map[common.Hash]struct{}
	ps           *peerSet
	peerList     []*peer
	peerStartPos int
	lock         sync.RWMutex

	reqDist *requestDistributor
}

func NewLesTxRelay(ps *peerSet, reqDist *requestDistributor) *LesTxRelay {
	r := &LesTxRelay{
		txSent:    make(map[common.Hash]*ltrInfo),
		txPending: make(map[common.Hash]struct{}),
		ps:        ps,
		reqDist:   reqDist,
	}
	ps.notify(r)
	return r
}

func (self *LesTxRelay) registerPeer(p *peer) {
	self.lock.Lock()
	defer self.lock.Unlock()

	self.peerList = self.ps.AllPeers()
}

func (self *LesTxRelay) unregisterPeer(p *peer) {
	self.lock.Lock()
	defer self.lock.Unlock()

	self.peerList = self.ps.AllPeers()
}

// send sends a list of transactions to at most a given number of peers at
// once, never resending any particular transaction to the same peer twice
func (self *LesTxRelay) send(txs types.Transactions, count int) {
	sendTo := make(map[*peer]types.Transactions)

	self.peerStartPos++ // rotate the starting position of the peer list
	if self.peerStartPos >= len(self.peerList) {
		self.peerStartPos = 0
	}

	for _, tx := range txs {
		hash := tx.Hash()
		ltr, ok := self.txSent[hash]
		if !ok {
			ltr = &ltrInfo{
				tx:     tx,
				sentTo: make(map[*peer]struct{}),
			}
			self.txSent[hash] = ltr
			self.txPending[hash] = struct{}{}
		}

		if len(self.peerList) > 0 {
			cnt := count
			pos := self.peerStartPos
			for {
				peer := self.peerList[pos]
				if _, ok := ltr.sentTo[peer]; !ok {
					sendTo[peer] = append(sendTo[peer], tx)
					ltr.sentTo[peer] = struct{}{}
					cnt--
				}
				if cnt == 0 {
					break // sent it to the desired number of peers
				}
				pos++
				if pos == len(self.peerList) {
					pos = 0
				}
				if pos == self.peerStartPos {
					break // tried all available peers
				}
			}
		}
	}

	for p, list := range sendTo {
		pp := p
		ll := list

		reqID := genReqID()
		rq := &distReq{
			getCost: func(dp distPeer) uint64 {
				peer := dp.(*peer)
				return peer.GetRequestCost(SendTxMsg, len(ll))
			},
			canSend: func(dp distPeer) bool {
				return dp.(*peer) == pp
			},
			request: func(dp distPeer) func() {
				peer := dp.(*peer)
				cost := peer.GetRequestCost(SendTxMsg, len(ll))
				peer.fcServer.QueueRequest(reqID, cost)
				return func() { peer.SendTxs(reqID, cost, ll) }
			},
		}
		self.reqDist.queue(rq)
	}
}

func (self *LesTxRelay) Send(txs types.Transactions) {
	self.lock.Lock()
	defer self.lock.Unlock()

	self.send(txs, 3)
}

func (self *LesTxRelay) NewHead(head common.Hash, mined []common.Hash, rollback []common.Hash) {
	self.lock.Lock()
	defer self.lock.Unlock()

	for _, hash := range mined {
		delete(self.txPending, hash)
	}

	for _, hash := range rollback {
		self.txPending[hash] = struct{}{}
	}

	if len(self.txPending) > 0 {
		txs := make(types.Transactions, len(self.txPending))
		i := 0
		for hash := range self.txPending {
			txs[i] = self.txSent[hash].tx
			i++
		}
		self.send(txs, 1)
	}
}

func (self *LesTxRelay) Discard(hashes []common.Hash) {
	self.lock.Lock()
	defer self.lock.Unlock()

	for _, hash := range hashes {
		delete(self.txSent, hash)
		delete(self.txPending, hash)
	}
}
