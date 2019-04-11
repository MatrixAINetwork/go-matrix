// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package p2p

import (
	"container/ring"
	"errors"
	"math/big"
	"math/rand"
	"sync"
	"time"

	"github.com/MatrixAINetwork/go-matrix/ca"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/event"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/p2p/discover"
)

// hash bucket
type Bucket struct {
	role   common.RoleType
	bucket map[int64][]common.Address

	rings *ring.Ring
	lock  sync.RWMutex

	ids  []common.Address
	self int64

	sub event.Subscription

	blockChain chan mc.BlockToBucket
	quit       chan struct{}

	log log.Logger
}

// Init bucket.
var Buckets = &Bucket{
	role:  common.RoleNil,
	ids:   make([]common.Address, 0),
	quit:  make(chan struct{}),
	rings: ring.New(4),
}

const (
	MaxBucketContent = 2000
	BucketLimit      = 10
)

var (
	errAddressEmpty = errors.New("address is empty")

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

	timeoutTimer := time.NewTimer(time.Second * 60)

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
		case <-timeoutTimer.C:
			b.maintainOuter()
			if !timeoutTimer.Stop() && len(timeoutTimer.C) > 0 {
				<-timeoutTimer.C
			}
			timeoutTimer.Reset(time.Second * 60)
		case h := <-b.blockChain:
			if !timeoutTimer.Stop() && len(timeoutTimer.C) > 0 {
				<-timeoutTimer.C
			}
			timeoutTimer.Reset(time.Second * 60)

			// only bottom nodes will into this buckets.
			if h.Role > common.RoleBucket {
				b.role = common.RoleNil
				break
			}

			// down to default, disconnect all peers first
			if b.role != h.Role && h.Role == common.RoleDefault {
				b.disconnectPeers()
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

			if len(b.ids) <= BucketLimit {
				b.maintainOuter()
				break
			}

			// maintain outer
			selfBucket, err := b.selfBucket()
			if err != nil {
				b.log.Error("bucket number wrong", "error", err)
				break
			}
			b.self = selfBucket

			// maintain inner
			b.maintainInner()
			switch selfBucket {
			case b.rings.Value.(int64):
				b.maintainOuter()
			case b.rings.Next().Value.(int64):
				b.disconnectMiner()
			case b.rings.Prev().Value.(int64):
				miners := ca.GetRolesByGroupWithNextElect(common.RoleMiner | common.RoleBackupValidator)
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
func (b *Bucket) maintainNodes(elected []common.Address) {
	// remake every time instead of delete
	b.bucket = make(map[int64][]common.Address)
	for _, v := range elected {
		b.bucketAdd(v)
	}
	for index, bkt := range b.bucket {
		b.log.Info("bucket info", "index", index, "length", len(bkt))
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
	miners := ca.GetRolesByGroupWithNextElect(common.RoleMiner | common.RoleBackupMiner)
	for _, miner := range miners {
		ServerP2p.RemovePeerByAddress(miner)
	}
}

// disconnectPeers disconnect all peers
func (b *Bucket) disconnectPeers() {
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
	next := (b.self + 1) % 4
	for _, peer := range ServerP2p.Peers() {
		signAddr := ServerP2p.ConvertIdToAddress(peer.ID())
		if signAddr == EmptyAddress {
			continue
		}
		pid, err := b.peerBucket(signAddr)
		if err != nil {
			b.log.Error("bucket number wrong", "error", err)
			continue
		}
		if pid == next {
			count++
		}
	}
	if count < MaxLink {
		if MaxLink < len(b.bucket[next]) {
			b.inner(MaxLink-count, next)
			return
		}
		b.inner(len(b.bucket[next])-count, next)
	}
}

// MaintainOuter maintain bucket outer.
func (b *Bucket) maintainOuter() {
	count := 0
	miners := ca.GetRolesByGroupWithNextElect(common.RoleMiner | common.RoleBackupMiner)
	b.log.Info("maintainOuter", "peer info", miners)
	for _, peer := range ServerP2p.Peers() {
		for _, miner := range miners {
			id := ServerP2p.ConvertAddressToId(miner)
			if id != EmptyNodeId && peer.ID() == id {
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
	return b.peerBucket(ServerP2p.ManAddress)
}

func (b *Bucket) peerBucket(addr common.Address) (int64, error) {
	m := big.Int{}
	if b.self < common.RoleBucket {
		return m.Mod(MockHash(ServerP2p.Self().ID).Big(), big.NewInt(4)).Int64(), nil
	}

	if addr != EmptyAddress {
		return m.Mod(common.BytesToHash(addr.Bytes()).Big(), big.NewInt(4)).Int64(), nil
	}

	return 0, errAddressEmpty
}

func (b *Bucket) linkBucketPeer() {
	if len(b.ids) <= BucketLimit {
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
		signAddr := ServerP2p.ConvertIdToAddress(peer.ID())
		if signAddr == EmptyAddress {
			b.log.Error("not found sign address", "id", peer.ID())
			continue
		}
		pid, err := b.peerBucket(signAddr)
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
			b.inner(MaxLink-count, self)
			return
		}
		b.inner(len(b.bucket[self])-count, self)
	}
}

// BucketAdd add to bucket.
func (b *Bucket) bucketAdd(addr common.Address) {
	b.lock.Lock()
	defer b.lock.Unlock()

	m := big.Int{}
	mod := m.Mod(addr.Hash().Big(), big.NewInt(4)).Int64()

	for _, n := range b.bucket[mod] {
		if n == addr {
			return
		}
	}
	b.bucket[mod] = append(b.bucket[mod], addr)
}

// inner adjust inner network.
func (b *Bucket) inner(num int, bucket int64) {
	if num <= 0 {
		return
	}
	peers := b.randomInnerPeersByBucketNumber(num, bucket)

	for _, value := range peers {
		b.log.Info("peer", "p2p", value)
		ServerP2p.AddPeerTask(value)
	}
}

// outer adjust outer network.
func (b *Bucket) outer(num int, ids []common.Address) {
	if num <= 0 {
		return
	}
	peers := b.randomOuterPeers(num, ids)

	for _, value := range peers {
		b.log.Info("peer", "p2p", value)
		ServerP2p.AddPeerTask(value)
	}
}

// RandomPeers random peers from next buckets.
func (b *Bucket) randomInnerPeersByBucketNumber(num int, bucket int64) (nodes []common.Address) {
	length := len(b.bucket[bucket])

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
	if len(nodes) <= 0 {
		return b.randomInnerPeersByBucketNumber(num, (bucket+1)%4)
	}
	return nodes
}

// RandomOuterPeers random peers from overstory.
func (b *Bucket) randomOuterPeers(num int, ids []common.Address) (nodes []common.Address) {
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

func MockHash(id discover.NodeID) common.Hash {
	return common.BytesToHash([]byte(id.String()))
}
