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
package topnode

import (
	"sync"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/messageState"
)

type voteInfo struct {
	hash uint64
	data *mc.HD_ConsensusVote
}
type DPosVoteState struct {
	mu               sync.RWMutex
	Hash             common.Hash
	Proposal         interface{}
	AffirmativeVotes []voteInfo
}

func (ds *DPosVoteState) hasHash(hash common.Hash) bool {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.Hash == hash
}
func (ds *DPosVoteState) addProposal(proposal interface{}) bool {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	have := ds.Proposal != nil
	ds.Proposal = proposal
	return have
}
func (ds *DPosVoteState) addVote(vote *mc.HD_ConsensusVote) (interface{}, []voteInfo) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	insVote := voteInfo{messageState.RlpFnvHash(vote), vote}
	for _, item := range ds.AffirmativeVotes {
		if item.hash == insVote.hash {
			return nil, nil
		}
	}
	ds.AffirmativeVotes = append(ds.AffirmativeVotes, insVote)
	log.Info("DPosVoteState", "length", len(ds.AffirmativeVotes))
	return ds.Proposal, ds.AffirmativeVotes[:]
}
func (ds *DPosVoteState) clear(proposal common.Hash) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.Hash = proposal
	ds.Proposal = nil
	ds.AffirmativeVotes = make([]voteInfo, 0)
}
func (ds *DPosVoteState) getVotes() (interface{}, []voteInfo) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.Proposal, ds.AffirmativeVotes[:]
}

type DPosVoteRing struct {
	DPosVoteS []*DPosVoteState
	capacity  int
	mu        sync.RWMutex
	last      int
}

func NewDPosVoteRing(capacity int) *DPosVoteRing {
	ring := &DPosVoteRing{
		DPosVoteS: make([]*DPosVoteState, capacity),
		capacity:  capacity,
		mu:        sync.RWMutex{},
		last:      capacity - 1,
	}
	for i := 0; i < capacity; i++ {
		ring.DPosVoteS[i] = &DPosVoteState{}
	}
	return ring
}
func (ring *DPosVoteRing) insertLast() int {
	ring.mu.Lock()
	defer ring.mu.Unlock()
	ring.last = (ring.last + 1) % ring.capacity
	return ring.last
}
func (ring *DPosVoteRing) getLast() int {
	ring.mu.RLock()
	defer ring.mu.RUnlock()
	return ring.last
}
func (ring *DPosVoteRing) insertNewProposal(hash common.Hash) *DPosVoteState {
	last := ring.insertLast()
	ring.DPosVoteS[last].clear(hash)
	return ring.DPosVoteS[last]
}

func (ring *DPosVoteRing) getVotes(hash common.Hash) (interface{}, []voteInfo) {
	ds, _ := ring.findProposal(hash)
	return ds.getVotes()
}
func (ring *DPosVoteRing) addProposal(hash common.Hash, proposal interface{}) bool {
	ds, have := ring.findProposal(hash)
	add := ds.addProposal(proposal)
	return !(have && add)
}
func (ring *DPosVoteRing) addVote(hash common.Hash, vote *mc.HD_ConsensusVote) (interface{}, []voteInfo) {
	ds, _ := ring.findProposal(hash)
	return ds.addVote(vote)
}
func (ring *DPosVoteRing) findProposal(hash common.Hash) (*DPosVoteState, bool) {
	for _, item := range ring.DPosVoteS {
		if item.hasHash(hash) {
			return item, true
		}
	}
	return ring.insertNewProposal(hash), false
}
