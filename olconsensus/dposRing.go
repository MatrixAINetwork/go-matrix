// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package olconsensus

import (
	"sync"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/messageState"
)

type voteInfo struct {
	hash uint64
	data *mc.HD_ConsensusVote
}
type DPosVoteState struct {
	mu               sync.RWMutex
	Hash             common.Hash
	Proposal         interface{}
	Voted            bool //本地是否对请求投过票
	AffirmativeVotes []voteInfo
}

func (ds *DPosVoteState) setVoted() {
	ds.Voted = true
}

func (ds *DPosVoteState) hasHash(hash common.Hash) bool {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.Hash == hash
}
func (ds *DPosVoteState) addProposal(proposal interface{}, voted bool) bool {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	have := ds.Proposal != nil
	ds.Proposal = proposal
	ds.Voted = voted
	return have
}
func (ds *DPosVoteState) addVote(vote *mc.HD_ConsensusVote) (interface{}, []voteInfo) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	insVote := voteInfo{messageState.RlpFnvHash(vote), vote}
	log.Debug("共识节点状态", "fnvHash", insVote.hash, "From", vote.From.String())
	for _, item := range ds.AffirmativeVotes {
		if item.hash == insVote.hash {
			log.Error("共识节点状态", "添加投票,投票已经存在 vote", vote, "已经收到的票数", len(ds.AffirmativeVotes))
			return nil, nil
		}
	}
	ds.AffirmativeVotes = append(ds.AffirmativeVotes, insVote)
	log.Debug("共识节点状态", "添加投票length", len(ds.AffirmativeVotes), "proposal", ds.Proposal)
	return ds.Proposal, ds.AffirmativeVotes[:]
}
func (ds *DPosVoteState) clear(proposal common.Hash) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.Hash = proposal
	ds.Proposal = nil
	ds.AffirmativeVotes = make([]voteInfo, 0)
}
func (ds *DPosVoteState) getVotes() (interface{}, []voteInfo, bool) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.Proposal, ds.AffirmativeVotes[:], ds.Voted
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

func (ring *DPosVoteRing) getVotes(hash common.Hash) (interface{}, []voteInfo, bool) {
	ds, _ := ring.findProposal(hash)
	return ds.getVotes()
}
func (ring *DPosVoteRing) addProposal(hash common.Hash, proposal interface{}, voted bool) bool {
	ds, have := ring.findProposal(hash)
	add := ds.addProposal(proposal, voted)
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
