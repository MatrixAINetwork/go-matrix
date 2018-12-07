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

	active bool

	sub event.Subscription

	quit       chan struct{}
	activeQuit chan struct{}
	roleChan   chan mc.BlockToLinker

	mu    sync.Mutex
	topMu sync.RWMutex

	linkMap  map[discover.NodeID]uint32
	selfPeer map[common.RoleType][]*Peer

	topNode      map[common.RoleType]map[discover.NodeID][]uint8
	topNodeCache map[common.RoleType]map[discover.NodeID][]uint8
}

type NodeAliveInfo struct {
	Account    common.Address
	Position   uint16
	Type       common.RoleType
	Heartbeats []uint8
}

const MaxLinkers = 1000

var Link = &Linker{
	role:         common.RoleNil,
	selfPeer:     make(map[common.RoleType][]*Peer),
	quit:         make(chan struct{}),
	activeQuit:   make(chan struct{}),
	topNode:      make(map[common.RoleType]map[discover.NodeID][]uint8),
	topNodeCache: make(map[common.RoleType]map[discover.NodeID][]uint8),
}

func (l *Linker) Start() {
	defer func() {
		l.sub.Unsubscribe()

		close(l.roleChan)
		close(l.quit)
		close(l.activeQuit)
	}()

	l.roleChan = make(chan mc.BlockToLinker)
	l.sub, _ = mc.SubscribeEvent(mc.BlockToLinkers, l.roleChan)
	l.initTopNodeMap()
	l.initTopNodeMapCache()

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

				log.Info("self ide", "linker", l.role)

				l.maintainPeer()

				if common.IsReElectionNumber(height) {
					l.topNodeCache = l.topNode
					l.topNode = make(map[common.RoleType]map[discover.NodeID][]uint8)
					l.initTopNodeMap()
				}
				if common.IsReElectionNumber(height - 10) {
					l.topNodeCache = make(map[common.RoleType]map[discover.NodeID][]uint8)
					l.initTopNodeMapCache()
				}

				if !l.active {
					// recode top node active info
					go l.Active()
					l.active = true
				}

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
	if l.active {
		l.activeQuit <- struct{}{}
	}
}

func (l *Linker) initTopNodeMap() {
	l.topMu.Lock()
	for i := int(common.RoleBackupMiner); i <= int(common.RoleValidator); i = i << 1 {
		l.topNode[common.RoleType(i)] = make(map[discover.NodeID][]uint8)
	}
	l.topMu.Unlock()
}

func (l *Linker) initTopNodeMapCache() {
	l.topMu.Lock()
	for i := int(common.RoleBackupMiner); i <= int(common.RoleValidator); i = i << 1 {
		l.topNodeCache[common.RoleType(i)] = make(map[discover.NodeID][]uint8)
	}
	l.topMu.Unlock()
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
	log.Info("coming linker", "role", l.role)
	all := ca.GetTopologyInLinker()
	log.Info("len all", "length", len(all))
	for key, peers := range all {
		if key >= roleType {
			for _, peer := range peers {
				node := discover.NewNode(peer, nil, defaultPort, defaultPort)
				ServerP2p.AddPeer(node)
			}
		}
	}
	if roleType&(common.RoleValidator|common.RoleBackupValidator) != 0 {
		gap := ca.GetGapValidator()
		for _, val := range gap {
			node := discover.NewNode(val, nil, defaultPort, defaultPort)
			ServerP2p.AddPeer(node)
		}
	}
}

func (l *Linker) Active() {
	tk := time.NewTicker(time.Second * 15)
	defer tk.Stop()

	for {
		select {
		case <-tk.C:
			l.recordTopNodeActiveInfo()
		case <-l.activeQuit:
			l.active = false
			return
		}
	}
}

func (l *Linker) recordTopNodeActiveInfo() {
	l.topMu.Lock()
	defer l.topMu.Unlock()

	for i := int(common.RoleBackupMiner); i <= int(common.RoleValidator); i = i << 1 {
		topNodes := ca.GetRolesByGroup(common.RoleType(i))

		for _, tn := range topNodes {
			if tn == ServerP2p.Self().ID {
				continue
			}
			if _, ok := l.topNode[common.RoleType(i)][tn]; !ok {
				l.topNode[common.RoleType(i)][tn] = []uint8{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
			}
		}
	}

	for role := range l.topNode {
		for key := range l.topNode[role] {
			ok := false
			for _, peer := range ServerP2p.Peers() {
				if peer.ID() == key {
					ok = true
				}
			}
			if ok {
				l.topNode[role][key] = append(l.topNode[role][key], 1)
			} else {
				l.topNode[role][key] = append(l.topNode[role][key], 0)
			}
			if len(l.topNode[role][key]) > 20 {
				l.topNode[role][key] = l.topNode[role][key][len(l.topNode[role][key])-20:]
			}
		}
	}
}

// GetTopNodeAliveInfo
func GetTopNodeAliveInfo(roleType common.RoleType) (result []NodeAliveInfo) {
	Link.topMu.RLock()
	defer Link.topMu.RUnlock()

	if len(Link.topNode) <= 0 {
		for key, val := range Link.topNodeCache {
			if (key & roleType) != 0 {
				for nodeId, val := range val {
					addr, err := ca.ConvertNodeIdToAddress(nodeId)
					if err != nil {
						log.Error("convert info", "error", err)
						continue
					}
					result = append(result, NodeAliveInfo{Account: addr, Heartbeats: val})
				}
			}
		}
		return
	}

	for key, vals := range Link.topNode {
		if (key & roleType) != 0 {
			for nodeId, val := range vals {
				addr, err := ca.ConvertNodeIdToAddress(nodeId)
				if err != nil {
					log.Error("convert info", "error", err)
					continue
				}
				result = append(result, NodeAliveInfo{Account: addr, Heartbeats: val})
			}
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
