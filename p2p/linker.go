// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package p2p

import (
	"encoding/json"
	"math/big"
	"sync"
	"time"

	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/event"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/p2p/discover"
)

type Linker struct {
	role common.RoleType

	quit     chan struct{}
	roleChan chan mc.BlockToLinker
	sub      event.Subscription
	mu       *sync.Mutex

	linkMap  map[discover.NodeID]uint32
	selfPeer map[common.RoleType][]*Peer

	topNode      map[discover.NodeID][]uint8
	topNodeCache map[discover.NodeID][]uint8
}

type NodeAliveInfo struct {
	Account    common.Address
	Position   uint16
	Type       common.RoleType
	Heartbeats []uint8
}

const MaxLinkers = 1000

var Link = &Linker{
	mu:           &sync.Mutex{},
	role:         common.RoleNil,
	selfPeer:     make(map[common.RoleType][]*Peer),
	quit:         make(chan struct{}),
	topNode:      make(map[discover.NodeID][]uint8),
	topNodeCache: make(map[discover.NodeID][]uint8),
}

func (l *Linker) Start() {
	defer func() {
		l.sub.Unsubscribe()

		close(l.roleChan)
		close(l.quit)
	}()

	l.roleChan = make(chan mc.BlockToLinker)
	l.sub, _ = mc.SubscribeEvent(mc.BlockToLinkers, l.roleChan)

	for {
		select {
		case r := <-l.roleChan:
			{
				height := r.Height.Uint64()

				if r.Role <= common.RoleBucket {
					l.role = common.RoleNil
					break
				}
				if l.role != r.Role {
					l.role = r.Role
				}
				dropNodes := ca.GetDropNode()
				l.dropNode(dropNodes)
				go l.dropNodeDefer(dropNodes)

				l.maintainPeer()

				if common.IsReElectionNumber(height) {
					l.topNodeCache = l.topNode
					l.topNode = make(map[discover.NodeID][]uint8)
				}
				if common.IsReElectionNumber(height - 10) {
					l.topNodeCache = make(map[discover.NodeID][]uint8)
				}

				// recode top node active info
				l.recordTopNodeActiveInfo()

				// broadcast link and message
				if l.role != common.RoleBroadcast {
					break
				}

				switch {
				case common.IsBroadcastNumber(height):
					l.ToLink()
				case common.IsBroadcastNumber(height + 2):
					if len(l.linkMap) <= 0 {
						break
					}
					bytes, err := l.encodeData()
					if err != nil {
						log.Error("encode error", "error", err)
						break
					}
					mc.PublishEvent(mc.SendBroadCastTx, mc.BroadCastEvent{Txtyps: mc.CallTheRoll, Height: big.NewInt(r.Height.Int64() + 2), Data: bytes})
				case common.IsBroadcastNumber(height + 1):
					break
				default:
					l.sendToAllPeersPing()
				}
			}
		case <-l.quit:
			return
		}
	}
}

func (l *Linker) Stop() {
	l.quit <- struct{}{}
}

// MaintainPeer
func (l *Linker) maintainPeer() {
	l.link(l.role)
}

// disconnect all peers.
func (l *Linker) dropNode(drops []discover.NodeID) {
	for _, drop := range drops {
		ServerP2p.RemovePeer(discover.NewNode(drop, nil, 0, 0))
	}
}

// dropNodeDefer disconnect all peers.
func (l *Linker) dropNodeDefer(drops []discover.NodeID) {
	select {
	case <-time.After(time.Second * 5):
		for _, drop := range drops {
			ServerP2p.RemovePeer(discover.NewNode(drop, nil, 0, 0))
		}
	}
}

// Link peers that should to link.
// link peers by group
func (l *Linker) link(roleType common.RoleType) {
	all := ca.GetTopologyInLinker()
	for key, peers := range all {
		if key >= roleType {
			for _, peer := range peers {
				node := discover.NewNode(peer, nil, defaultPort, defaultPort)
				ServerP2p.AddPeer(node)
			}
		}
	}
}

func (l *Linker) recordTopNodeActiveInfo() {
	topNode := ca.GetRolesByGroup(common.RoleMiner | common.RoleValidator)
	for _, tn := range topNode {
		if _, ok := l.topNode[tn]; !ok {
			l.topNode[tn] = []uint8{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
		}
	}

	for key := range l.topNode {
		ok := false
		for _, peer := range ServerP2p.Peers() {
			if peer.ID() == key {
				ok = true
			}
		}
		if ok {
			l.topNode[key] = append(l.topNode[key], 1)
		} else {
			l.topNode[key] = append(l.topNode[key], 0)
		}
		if len(l.topNode[key]) > 20 {
			l.topNode[key] = l.topNode[key][len(l.topNode[key])-20:]
		}
	}
}

// GetTopNodeAliveInfo
func GetTopNodeAliveInfo(address []common.Address) (result []NodeAliveInfo) {
	for _, addr := range address {
		id, err := ca.ConvertAddressToNodeId(addr)
		if err != nil {
			continue
		}
		node, err := ca.GetAccountTopologyInfo(addr, ca.GetHeight().Uint64())
		if err != nil {
			continue
		}
		if val, ok := Link.topNode[id]; ok {
			result = append(result, NodeAliveInfo{Account: addr, Position: node.Position, Type: node.Type, Heartbeats: val})
			continue
		}
		if val, ok := Link.topNodeCache[id]; ok {
			result = append(result, NodeAliveInfo{Account: addr, Position: node.Position, Type: node.Type, Heartbeats: val})
		}
	}
	return
}

func (l *Linker) ToLink() {
	l.linkMap = make(map[discover.NodeID]uint32)
	h := ca.GetHeight()
	elects, _ := ca.GetElectedByHeight(h)

	if len(elects) <= MaxLinkers {
		for _, elect := range elects {
			node := discover.NewNode(elect.NodeID, nil, defaultPort, defaultPort)
			ServerP2p.AddPeer(node)
			l.linkMap[elect.NodeID] = 0
		}
		return
	}

	randoms := Random(len(elects), MaxLinkers)
	for _, index := range randoms {
		node := discover.NewNode(elects[index].NodeID, nil, defaultPort, defaultPort)
		ServerP2p.AddPeer(node)
		l.linkMap[elects[index].NodeID] = 0
	}
}

func Record(id discover.NodeID) error {
	Link.mu.Lock()
	defer Link.mu.Unlock()
	if _, ok := Link.linkMap[id]; ok {
		Link.linkMap[id]++
	}
	return nil
}

func (l *Linker) sendToAllPeersPing() {
	peers := ServerP2p.Peers()
	for _, peer := range peers {
		Send(peer.msgReadWriter, common.BroadcastReqMsg, []uint8{0})
	}
}

func (l *Linker) encodeData() ([]byte, error) {
	Link.mu.Lock()
	defer Link.mu.Unlock()
	r := make(map[string]uint32)
	for key, value := range l.linkMap {
		addr, err := ca.ConvertNodeIdToAddress(key)
		if err != nil {
			return nil, err
		}
		r[addr.Hex()] = value
	}
	return json.Marshal(r)
}

// GetRollBook
func GetRollBook() (map[common.Address]struct{}, error) {
	Link.mu.Lock()
	defer Link.mu.Unlock()

	r := make(map[common.Address]struct{})
	for key := range Link.linkMap {
		addr, err := ca.ConvertNodeIdToAddress(key)
		if err != nil {
			return nil, err
		}
		r[addr] = struct{}{}
	}
	return r, nil
}
