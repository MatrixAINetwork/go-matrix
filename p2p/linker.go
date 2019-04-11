// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package p2p

import (
	"encoding/json"
	"math/big"
	"sync"
	"time"

	"github.com/MatrixAINetwork/go-matrix/ca"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/event"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/p2p/discover"
)

type Linker struct {
	role common.RoleType

	active          bool
	broadcastActive bool

	sub event.Subscription

	quit       chan struct{}
	activeQuit chan struct{}
	roleChan   chan mc.BlockToLinker

	mu    sync.Mutex
	topMu sync.RWMutex

	linkMap  map[common.Address]uint32
	selfPeer map[common.RoleType][]*Peer

	topNode      map[common.RoleType]map[common.Address][]uint8
	topNodeCache map[common.RoleType]map[common.Address][]uint8
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
	topNode:      make(map[common.RoleType]map[common.Address][]uint8),
	topNodeCache: make(map[common.RoleType]map[common.Address][]uint8),
}

var (
	EmptyNodeId  = discover.NodeID{}
	EmptyAddress = common.Address{}
)

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
				if r.BroadCastInterval == nil {
					log.Error("p2p link", "broadcast interval err", "is nil")
					continue
				}

				if r.Role <= common.RoleBucket {
					l.role = common.RoleNil
					break
				}
				if l.role != r.Role {
					l.role = r.Role
				}
				dropNodes := ca.GetDropNode()
				l.dropNode(dropNodes)

				l.maintainPeer()

				if r.BroadCastInterval.IsReElectionNumber(height) {
					l.topNodeCache = l.topNode
					l.initTopNodeMap()
				}
				if r.BroadCastInterval.IsReElectionNumber(height - 10) {
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
				case r.BroadCastInterval.IsBroadcastNumber(height):
					l.ToLink()
					l.broadcastActive = true

				case r.BroadCastInterval.IsBroadcastNumber(height + 2):
					if len(l.linkMap) <= 0 {
						break
					}
					bytes, err := l.encodeData()
					if err != nil {
						log.Error("encode error", "error", err)
						break
					}
					mc.PublishEvent(mc.SendBroadCastTx, mc.BroadCastEvent{Txtyps: mc.CallTheRoll, Height: big.NewInt(r.Height.Int64() + 2), Data: bytes})
				case r.BroadCastInterval.IsBroadcastNumber(height + 1):
					break
				default:
					l.sendToAllPeersPing()
				}

				if !l.broadcastActive {
					l.ToLink()
					l.broadcastActive = true
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
	l.topNode = make(map[common.RoleType]map[common.Address][]uint8)
	for i := int(common.RoleBackupMiner); i <= int(common.RoleValidator); i = i << 1 {
		l.topNode[common.RoleType(i)] = make(map[common.Address][]uint8)
	}
	l.topMu.Unlock()
}

func (l *Linker) initTopNodeMapCache() {
	l.topMu.Lock()
	l.topNodeCache = make(map[common.RoleType]map[common.Address][]uint8)
	for i := int(common.RoleBackupMiner); i <= int(common.RoleValidator); i = i << 1 {
		l.topNodeCache[common.RoleType(i)] = make(map[common.Address][]uint8)
	}
	l.topMu.Unlock()
}

// MaintainPeer
func (l *Linker) maintainPeer() {
	l.link(l.role)
}

// disconnect all peers.
func (l *Linker) dropNode(drops []common.Address) {
	for _, drop := range drops {
		ServerP2p.RemovePeerByAddress(drop)
	}
}

// dropNodeDefer disconnect all peers.
func (l *Linker) dropNodeDefer(drops []common.Address) {
	select {
	case <-time.After(time.Second * 5):
		for _, drop := range drops {
			ServerP2p.RemovePeerByAddress(drop)
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
				ServerP2p.AddPeerTask(peer)
			}
		}
	}
	if roleType&(common.RoleValidator|common.RoleBackupValidator) != 0 {
		gap := ca.GetGapValidator()
		for _, val := range gap {
			ServerP2p.AddPeerTask(val)
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
			if tn == ServerP2p.ManAddress {
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
				id := ServerP2p.ConvertAddressToId(key)
				if id != EmptyNodeId && peer.ID() == id {
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
				for signAddr, hearts := range val {
					result = append(result, NodeAliveInfo{Account: signAddr, Heartbeats: hearts})
				}
			}
		}
		return
	}

	for key, vals := range Link.topNode {
		if (key & roleType) != 0 {
			for signAddr, val := range vals {
				result = append(result, NodeAliveInfo{Account: signAddr, Heartbeats: val})
			}
		}
	}
	return
}

func (l *Linker) ToLink() {
	l.linkMap = make(map[common.Address]uint32)
	h := ca.GetHash()
	elects, _ := ca.GetElectedByHeightByHash(h)

	if len(elects) <= MaxLinkers {
		for _, elect := range elects {
			ServerP2p.AddPeerTask(elect.SignAddress)
			l.linkMap[elect.SignAddress] = 0
		}
		return
	}

	randoms := Random(len(elects), MaxLinkers)
	for _, index := range randoms {
		ServerP2p.AddPeerTask(elects[index].SignAddress)
		l.linkMap[elects[index].SignAddress] = 0
	}
}

func Record(id discover.NodeID) error {
	Link.mu.Lock()
	defer Link.mu.Unlock()

	signAddr := ServerP2p.ConvertIdToAddress(id)
	if signAddr == EmptyAddress {
		return nil
	}

	if _, ok := Link.linkMap[signAddr]; ok {
		Link.linkMap[signAddr]++
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
		r[key.Hex()] = value
	}
	return json.Marshal(r)
}

// GetRollBook
func GetRollBook() (map[common.Address]struct{}, error) {
	Link.mu.Lock()
	defer Link.mu.Unlock()

	r := make(map[common.Address]struct{})
	for key := range Link.linkMap {
		r[key] = struct{}{}
	}
	return r, nil
}
