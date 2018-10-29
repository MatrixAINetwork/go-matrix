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
package p2p

import (
	"container/ring"
	"math/big"
	"math/rand"
	"sync"
	"time"

	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/event"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/p2p/discover"
)

// hash bucket
type Bucket struct {
	role   common.RoleType
	bucket map[int64][]discover.NodeID

	rings *ring.Ring
	lock  *sync.RWMutex

	ids []discover.NodeID

	sub event.Subscription

	blockChain chan mc.BlockToBucket
	quit       chan struct{}

	log log.Logger
}

// Init bucket.
var Buckets = &Bucket{
	role:  common.RoleNil,
	lock:  new(sync.RWMutex),
	ids:   make([]discover.NodeID, 0),
	quit:  make(chan struct{}),
	rings: ring.New(4),
}

const MaxBucketContent = 2000

var (
	MaxLink = 3
)

// init bucket.
func (b *Bucket) init() {
	for i := 0; i < b.rings.Len(); i++ {
		b.rings.Value = int64(i)
		b.rings = b.rings.Next()
	}
	b.log = log.New()
}

// Start bucket.
func (b *Bucket) Start() {
	b.init()

	b.log.Info("buckets start!")

	defer func() {
		b.log.Info("buckets stop!")
		b.sub.Unsubscribe()

		close(b.quit)
		close(b.blockChain)
	}()

	b.blockChain = make(chan mc.BlockToBucket)
	b.sub, _ = mc.SubscribeEvent(mc.BlockToBuckets, b.blockChain)

	for {
		select {
		case h := <-b.blockChain:
			// only bottom nodes will into this buckets.
			if h.Role > common.RoleBucket {
				b.role = common.RoleNil
				break
			}

			if b.role != h.Role && b.role == common.RoleNil {
				fNodes := ca.GetFrontNodes()
				b.disconnectPeers(fNodes)
			}
			if b.role != h.Role {
				b.role = h.Role
			}

			b.ids = h.Ms
			// maintain nodes in buckets
			b.maintainNodes(h.Ms)

			// if not in bucket, do nothing
			if b.role != common.RoleBucket {
				b.linkBucketPeer()
				break
			}

			// modify max peers in buckets
			if b.nodesCount() >= MaxBucketContent {
				MaxLink = 2
				b.disconnectOnePeer()
			} else {
				MaxLink = 3
			}

			// adjust bucket order
			temp := &big.Int{}
			if temp.Mod(h.Height, big.NewInt(300)) == big.NewInt(50) {
				b.rings = b.rings.Prev()
			}

			if len(b.ids) <= 64 {
				b.maintainOuter()
				break
			}
			// maintain inner
			b.maintainInner()

			// maintain outer
			selfBucket, err := b.selfBucket()
			if err != nil {
				b.log.Error("bucket number wrong", "error", err)
				break
			}
			switch selfBucket {
			case b.rings.Value.(int64):
				b.maintainOuter()
			case b.rings.Next().Value.(int64):
				b.disconnectMiner()
			case b.rings.Prev().Value.(int64):
				miners := ca.GetRolesByGroupWithBackup(common.RoleMiner | common.RoleBackupValidator)
				b.outer(MaxLink, miners)
			}
		case <-b.quit:
			return
		}
	}
}

// Stop bucket running.
func (b *Bucket) Stop() {
	b.lock.Lock()
	b.quit <- struct{}{}
	b.lock.Unlock()
}

// maintainNodes maintain nodes in buckets.
func (b *Bucket) maintainNodes(elected []discover.NodeID) {
	// remake every time instead of delete
	b.bucket = make(map[int64][]discover.NodeID)
	for _, v := range elected {
		b.bucketAdd(v)
	}
}

// nodesCount return nodes number in buckets.
func (b *Bucket) nodesCount() (count int) {
	for _, val := range b.bucket {
		count += len(val)
	}
	return count
}

// DisconnectMiner older disconnect miner.
func (b *Bucket) disconnectMiner() {
	miners := ca.GetRolesByGroupWithBackup(common.RoleMiner | common.RoleBackupMiner)
	for _, miner := range miners {
		ServerP2p.RemovePeer(discover.NewNode(miner, nil, 0, 0))
	}
}

// disconnectPeers disconnect all peers
func (b *Bucket) disconnectPeers(drops []discover.NodeID) {
	for _, peer := range drops {
		ServerP2p.RemovePeer(discover.NewNode(peer, nil, 0, 0))
	}
	for _, peer := range ServerP2p.Peers() {
		ServerP2p.RemovePeer(discover.NewNode(peer.ID(), nil, 0, 0))
	}
}

// disconnectOnePeer if nodes in buckets more than 2 thousand, then disconnect one peer.
func (b *Bucket) disconnectOnePeer() {
	for _, peer := range ServerP2p.Peers() {
		ServerP2p.RemovePeer(discover.NewNode(peer.ID(), nil, 0, 0))
		break
	}
}

// MaintainInner maintain bucket inner.
func (b *Bucket) maintainInner() {
	count := 0
	for _, peer := range ServerP2p.Peers() {
		pid, err := b.peerBucket(peer.ID())
		if err != nil {
			b.log.Error("bucket number wrong", "error", err)
			continue
		}
		if pid == b.rings.Next().Value {
			count++
		}
	}
	if count < MaxLink {
		if MaxLink < len(b.bucket[b.rings.Next().Value.(int64)]) {
			b.inner(MaxLink-count, b.rings.Next().Value.(int64))
			return
		}
		b.inner(len(b.bucket[b.rings.Next().Value.(int64)])-count, b.rings.Next().Value.(int64))
	}
}

// MaintainOuter maintain bucket outer.
func (b *Bucket) maintainOuter() {
	count := 0
	miners := ca.GetRolesByGroupWithBackup(common.RoleMiner | common.RoleBackupMiner)
	b.log.Info("maintainOuter", "peer info", miners)
	for _, peer := range ServerP2p.Peers() {
		for _, miner := range miners {
			if peer.ID() == miner {
				count++
				break
			}
		}
	}
	b.log.Info("maintainOuter", "peer count", count)
	if count < MaxLink {
		if MaxLink < len(miners) {
			b.outer(MaxLink-count, miners)
			return
		}
		b.outer(len(miners)-count, miners)
	}
}

// SelfBucket return self bucket number.
func (b *Bucket) selfBucket() (int64, error) {
	return b.peerBucket(ServerP2p.Self().ID)
}

func (b *Bucket) peerBucket(node discover.NodeID) (int64, error) {
	addr, err := ca.ConvertNodeIdToAddress(node)
	if err != nil {
		b.log.Error("bucket add", "error:", err)
		return 0, err
	}
	m := big.Int{}
	return m.Mod(addr.Hash().Big(), big.NewInt(4)).Int64(), nil
}

func (b *Bucket) linkBucketPeer() {
	if len(b.ids) <= 64 {
		b.maintainOuter()
		return
	}
	self, err := b.selfBucket()
	if err != nil {
		b.log.Error("bucket number wrong", "error", err)
		return
	}
	count := 0
	for _, peer := range ServerP2p.Peers() {
		pid, err := b.peerBucket(peer.ID())
		if err != nil {
			b.log.Error("bucket number wrong", "error", err)
			continue
		}
		if pid == self {
			count++
		}
	}

	if count < MaxLink {
		if MaxLink < len(b.bucket[self]) {
			b.inner(MaxLink-count, b.rings.Value.(int64))
			return
		}
		b.inner(len(b.bucket[self])-count, b.rings.Value.(int64))
	}
}

// BucketAdd add to bucket.
func (b *Bucket) bucketAdd(nodeId discover.NodeID) {
	b.lock.Lock()
	defer b.lock.Unlock()

	addr, err := ca.ConvertNodeIdToAddress(nodeId)
	if err != nil {
		b.log.Error("bucket add", "error:", err)
		return
	}
	m := big.Int{}
	mod := m.Mod(addr.Hash().Big(), big.NewInt(4)).Int64()

	for _, n := range b.bucket[mod] {
		if n == nodeId {
			return
		}
	}
	b.bucket[mod] = append(b.bucket[mod], nodeId)
}

// inner adjust inner network.
func (b *Bucket) inner(num int, bucket int64) {
	if num <= 0 {
		return
	}
	peers := b.randomInnerPeersByBucketNumber(num, bucket)

	for _, value := range peers {
		b.log.Info("peer", "p2p", value)
		node := discover.NewNode(value, nil, defaultPort, defaultPort)
		ServerP2p.AddPeer(node)
	}
}

// outer adjust outer network.
func (b *Bucket) outer(num int, ids []discover.NodeID) {
	if num <= 0 {
		return
	}
	peers := b.randomOuterPeers(num, ids)

	for _, value := range peers {
		b.log.Info("peer", "p2p", value)
		node := discover.NewNode(value, nil, defaultPort, defaultPort)
		ServerP2p.AddPeer(node)
	}
}

// RandomPeers random peers from next buckets.
func (b *Bucket) randomInnerPeersByBucketNumber(num int, bucket int64) (nodes []discover.NodeID) {
	length := len(b.bucket[b.rings.Next().Value.(int64)])

	if length <= MaxLink {
		return b.bucket[bucket]
	}

	randoms := Random(length, num)
	for _, ran := range randoms {
		for index := range b.bucket[bucket] {
			if index == ran {
				nodes = append(nodes, b.bucket[bucket][index])
				break
			}
		}
	}
	return nodes
}

// RandomOuterPeers random peers from overstory.
func (b *Bucket) randomOuterPeers(num int, ids []discover.NodeID) (nodes []discover.NodeID) {
	if len(ids) <= MaxLink {
		return ids
	}

	randoms := Random(len(ids), num)
	for _, ran := range randoms {
		for index := range ids {
			if ran == index {
				nodes = append(nodes, ids[index])
			}
		}
	}
	return nodes
}

// Random a int number.
func Random(max, num int) (randoms []int) {
	rand.Seed(time.Now().UnixNano())
	for m := 0; m < num; m++ {
		randoms = append(randoms, rand.Intn(max))
	}
	return randoms
}
