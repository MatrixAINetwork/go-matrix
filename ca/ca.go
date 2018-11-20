// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package ca

import (
	"errors"
	"math/big"
	"sync"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/core/vm"
	"github.com/matrix/go-matrix/depoistInfo"
	"github.com/matrix/go-matrix/event"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/p2p/discover"
	"github.com/matrix/go-matrix/params"
)

type TopologyGraphReader interface {
	GetTopologyGraphByNumber(number uint64) (*mc.TopologyGraph, error)
	GetOriginalElect(number uint64) ([]common.Elect, error)
	GetNextElect(number uint64) ([]common.Elect, error)
}

// Identity stand for node's identity.
type Identity struct {
	// self nodeId
	self discover.NodeID
	addr common.Address

	// if in elected duration
	duration      bool
	currentHeight *big.Int

	trChan         chan TopologyGraphReader
	topologyReader TopologyGraphReader
	topology       *mc.TopologyGraph
	prevElect      []common.Elect
	currentNodes   []discover.NodeID
	frontNodes     []discover.NodeID

	// self previous, current and next role type
	currentRole common.RoleType

	// chan to listen block coming and quit message
	blockChan chan *types.Block
	quit      chan struct{}

	// lock and once to sync
	lock sync.RWMutex
	once sync.Once

	// sub to unsubscribe block channel
	sub event.Subscription

	// logger
	log log.Logger

	// deposit in current height
	deposit []vm.DepositDetail

	// addrByGroup
	addrByGroup map[common.RoleType][]common.Address
}

var ide = newIde()

func newIde() *Identity {
	return &Identity{
		quit:        make(chan struct{}),
		currentRole: common.RoleNil,
		duration:    false,
		trChan:      make(chan TopologyGraphReader, 1),
		topology:    new(mc.TopologyGraph),
		prevElect:   make([]common.Elect, 0),
	}
}

// init to do something before run.
func (ide *Identity) init(id discover.NodeID, path string) {
	ide.once.Do(func() {
		// check bootNode and set identity
		ide.self = id
		ide.log = log.New()
	})
}

// Run this Identity.
func Start(id discover.NodeID, path string) {
	ide.init(id, path)

	defer func() {
		ide.sub.Unsubscribe()

		close(ide.quit)
		close(ide.blockChan)
	}()

	select {
	case tr := <-ide.trChan:
		ide.topologyReader = tr
	case <-ide.quit:
		return
	}

	ide.blockChan = make(chan *types.Block)
	ide.sub, _ = mc.SubscribeEvent(mc.NewBlockMessage, ide.blockChan)
	log.INFO("CA", "订阅区块事件", "完成")
	mc.PublishEvent(mc.CA_ReqCurrentBlock, struct{}{})

	for {
		select {
		case block := <-ide.blockChan:
			header := block.Header()
			hash := block.Hash()
			ide.currentHeight = header.Number

			log.INFO("CA", "leader", header.Leader, "height", header.Number.Uint64(), "block hash", hash)

			// init current height deposit
			ide.deposit, _ = GetElectedByHeightWithdraw(header.Number)
			// get self address from deposit
			ide.addr = GetAddress()

			// do topology
			tg, err := ide.topologyReader.GetTopologyGraphByNumber(header.Number.Uint64())
			if err != nil {
				ide.log.Error("get topology", "error", err)
				continue
			}
			ide.topology = tg

			// get elect
			elect, err := ide.topologyReader.GetNextElect(header.Number.Uint64())
			if err != nil {
				ide.log.Error("get next elect", "error", err)
				continue
			}
			ide.prevElect = elect

			// init topology
			initCurrentTopology()
			initNowTopologyResult()

			// get nodes in buckets
			nodesInBuckets := getNodesInBuckets(header.Number)

			// send role message to elect
			mc.PublishEvent(mc.CA_RoleUpdated, &mc.RoleUpdatedMsg{Role: ide.currentRole, BlockNum: header.Number.Uint64(), BlockHash: hash, Leader: header.Leader})
			log.Info("ca publish identity", "data", mc.RoleUpdatedMsg{Role: ide.currentRole, BlockNum: header.Number.Uint64(), Leader: header.Leader})
			// get nodes in buckets and send to buckets
			mc.PublishEvent(mc.BlockToBuckets, mc.BlockToBucket{Ms: nodesInBuckets, Height: block.Header().Number, Role: ide.currentRole})
			// send identity to linker
			mc.PublishEvent(mc.BlockToLinkers, mc.BlockToLinker{Height: header.Number, Role: ide.currentRole})
			mc.PublishEvent(mc.SendSyncRole, mc.SyncIdEvent{Role: ide.currentRole})//lb
			mc.PublishEvent(mc.TxPoolManager, ide.currentRole)
		case <-ide.quit:
			return
		}
	}
}

// Stop this Identity.
func Stop() {
	ide.log.Info("identity stop")

	ide.lock.Lock()
	ide.quit <- struct{}{}
	ide.lock.Unlock()
}

// InitCurrentTopology init current topology.
func initCurrentTopology() {
	ide.currentRole = common.RoleDefault

	for _, t := range ide.topology.NodeList {
		if t.Account == ide.addr {
			ide.currentRole = t.Type
		}
	}
	for _, b := range params.BroadCastNodes {
		if b.NodeID == ide.self {
			ide.currentRole = common.RoleBroadcast
			break
		}
	}
	log.Info("current topology", "info:", ide.topology)
}

// initNowTopologyResult
func initNowTopologyResult() {
	ide.lock.Lock()
	ide.addrByGroup = make(map[common.RoleType][]common.Address)
	for _, node := range ide.topology.NodeList {
		ide.addrByGroup[node.Type] = append(ide.addrByGroup[node.Type], node.Account)
	}
	for _, b := range params.BroadCastNodes {
		ide.addrByGroup[common.RoleBroadcast] = append(ide.addrByGroup[common.RoleBroadcast], b.Address)
	}
	ide.lock.Unlock()
}

// SetTopologyReader
func SetTopologyReader(topologyReader TopologyGraphReader) {
	ide.trChan <- topologyReader
}

// GetRolesByGroup
func GetRolesByGroup(roleType common.RoleType) (result []discover.NodeID) {
	ide.lock.RLock()
	for k, v := range ide.addrByGroup {
		if (k & roleType) != 0 {
			for _, addr := range v {
				id, err := ConvertAddressToNodeId(addr)
				if err != nil {
					ide.log.Error("convert error", "ca", err)
					continue
				}
				result = append(result, id)
			}
		}
	}
	ide.lock.RUnlock()
	return
}

// GetRolesByGroupWithBackup
func GetRolesByGroupWithNextElect(roleType common.RoleType) (result []discover.NodeID) {
	result = GetRolesByGroup(roleType)
	for _, elect := range ide.prevElect {
		temp := true
		role := elect.Type.Transfer2CommonRole()
		if (role & roleType) != 0 {
			id, err := ConvertAddressToNodeId(elect.Account)
			if err != nil {
				ide.log.Error("convert error", "ca", err)
				continue
			}
			for _, r := range result {
				if r == id {
					temp = false
				}
			}
			if temp {
				result = append(result, id)
			}
		}
	}
	return
}

// GetRolesByGroupOnlyBackup
func GetRolesByGroupOnlyNextElect(roleType common.RoleType) (result []discover.NodeID) {
	for _, elect := range ide.prevElect {
		role := elect.Type.Transfer2CommonRole()
		if (role & roleType) != 0 {
			id, err := ConvertAddressToNodeId(elect.Account)
			if err != nil {
				ide.log.Error("convert error", "ca", err)
				continue
			}
			result = append(result, id)
		}
	}
	return
}

// Get self identity.
func GetRole() (role common.RoleType) {
	ide.lock.Lock()
	defer ide.lock.Unlock()

	return ide.currentRole
}

func GetHeight() *big.Int {
	ide.lock.Lock()
	defer ide.lock.Unlock()

	return ide.currentHeight
}

// InDuration
func InDuration() bool {
	ide.lock.Lock()
	defer ide.lock.Unlock()

	return ide.duration
}

// GetElectedByHeightAndRole get elected node, miner or validator by block height and type.
func GetElectedByHeightAndRole(height *big.Int, roleType common.RoleType) ([]vm.DepositDetail, error) {
	return depoistInfo.GetDepositList(height, roleType)
}

// GetElectedByHeight get all elected node by height.
func GetElectedByHeight(height *big.Int) ([]vm.DepositDetail, error) {
	return depoistInfo.GetAllDeposit(height)
}

// GetElectedByHeightWithdraw get all info in deposit.
func GetElectedByHeightWithdraw(height *big.Int) ([]vm.DepositDetail, error) {
	return depoistInfo.GetDepositAndWithDrawList(height)
}

// GetNodeNumber
func GetNodeNumber() (uint32, error) {
	ide.lock.RLock()
	defer ide.lock.RUnlock()

	for _, n := range ide.topology.NodeList {
		if n.Account == ide.addr {
			return uint32(n.NodeNumber), nil
		}
	}
	return 0, errors.New("No current node number. ")
}

// getNodesInBuckets get miner nodes that should be in buckets.
func getNodesInBuckets(height *big.Int) (result []discover.NodeID) {
	electedMiners, _ := GetElectedByHeightAndRole(height, common.RoleMiner)

	msMap := make(map[common.Address]discover.NodeID)
	for _, m := range electedMiners {
		msMap[m.Address] = m.NodeID
	}
	for _, node := range ide.topology.NodeList {
		for key := range msMap {
			if key == node.Account {
				delete(msMap, key)
				break
			}
		}
	}
	for key, val := range msMap {
		if ide.addr == key {
			ide.currentRole = common.RoleBucket
		}
		result = append(result, val)
	}
	return
}

// GetTopologyInLinker
func GetTopologyInLinker() (result map[common.RoleType][]discover.NodeID) {
	ide.frontNodes = make([]discover.NodeID, 0)
	ide.frontNodes = ide.currentNodes
	ide.currentNodes = make([]discover.NodeID, 0)

	result = make(map[common.RoleType][]discover.NodeID)
	ide.lock.RLock()
	for k, v := range ide.addrByGroup {
		for _, addr := range v {
			id, err := ConvertAddressToNodeId(addr)
			if err != nil {
				ide.log.Error("convert error", "ca", err)
				continue
			}
			ide.currentNodes = append(ide.currentNodes, id)
			result[k] = append(result[k], id)
		}
	}
	ide.lock.RUnlock()
	for _, elect := range ide.prevElect {
		id, err := ConvertAddressToNodeId(elect.Account)
		if err != nil {
			ide.log.Error("convert error", "ca", err)
			continue
		}

		temp := true
		role := elect.Type.Transfer2CommonRole()
		for _, i := range result[role] {
			if i == id {
				temp = false
			}
		}
		if temp {
			ide.currentNodes = append(ide.currentNodes, id)
			result[role] = append(result[role], id)
		}
	}
	return
}

// GetDropNode
func GetDropNode() (result []discover.NodeID) {
	for _, fn := range ide.frontNodes {
		temp := false
		for _, cn := range ide.currentNodes {
			if cn == fn {
				temp = true
				break
			}
		}
		if !temp {
			result = append(result, fn)
		}
	}
	return
}

// GetFrontNodes
func GetFrontNodes() []discover.NodeID {
	ide.lock.RLock()
	defer ide.lock.RUnlock()
	return ide.frontNodes
}

// GetAddress
func GetAddress() common.Address {
	addr, err := ConvertNodeIdToAddress(ide.self)
	if err != nil {
		log.Error("ca get self address", "error", err)
	}
	return addr
}

// GetSelfLevel
func GetSelfLevel() int {
	switch {
	case ide.currentRole > common.RoleBucket:
		return TopNode
	case ide.currentRole == common.RoleBucket:
		m := big.Int{}
		return int(m.Mod(ide.addr.Hash().Big(), big.NewInt(4)).Int64()) + 1
	case ide.currentRole <= common.RoleDefault:
		return DefaultNode
	default:
		return ErrNode
	}
}

// GetTopologyByNumber
func GetTopologyByNumber(reqTypes common.RoleType, number uint64) (*mc.TopologyGraph, error) {
	tg, err := ide.topologyReader.GetTopologyGraphByNumber(number)
	if err != nil {
		log.Error("GetAccountTopologyInfo", "error", err, "number", number)
		return nil, err
	}

	rlt := &mc.TopologyGraph{
		Number:        tg.Number,
		CurNodeNumber: tg.CurNodeNumber,
	}
	for _, node := range tg.NodeList {
		if node.Type&reqTypes != 0 {
			rlt.NodeList = append(rlt.NodeList, node)
		}
	}

	return rlt, nil
}

// GetAccountTopologyInfo
func GetAccountTopologyInfo(account common.Address, number uint64) (*mc.TopologyNodeInfo, error) {
	tg, err := ide.topologyReader.GetTopologyGraphByNumber(number)
	if err != nil {
		ide.log.Error("GetAccountTopologyInfo", "error", err)
		return nil, err
	}
	for _, node := range tg.NodeList {
		if node.Account == account {
			return &node, nil
		}
	}
	return nil, errors.New("not found")
}

// GetAccountOriginalRole
func GetAccountOriginalRole(account common.Address, number uint64) (common.RoleType, error) {
	for _, b := range params.BroadCastNodes {
		if b.Address == account {
			return common.RoleBroadcast, nil
		}
	}
	ori, err := ide.topologyReader.GetOriginalElect(number)
	if err != nil {
		ide.log.Error("get original elect", "error", err)
		return common.RoleNil, err
	}

	for _, elect := range ori {
		if elect.Account == account {
			return elect.Type.Transfer2CommonRole(), nil
		}
	}
	return common.RoleNil, errors.New("not found")
}

// ConvertNodeIdToAddress
func ConvertNodeIdToAddress(id discover.NodeID) (addr common.Address, err error) {
	for _, node := range ide.deposit {
		if node.NodeID == id {
			return node.Address, nil
		}
	}
	for _, b := range params.BroadCastNodes {
		if b.NodeID == id {
			return b.Address, nil
		}
	}
	return addr, errors.New("not found")
}

// ConvertAddressToNodeId
func ConvertAddressToNodeId(address common.Address) (id discover.NodeID, err error) {
	for _, node := range ide.deposit {
		if node.Address == address {
			return node.NodeID, nil
		}
	}
	for _, b := range params.BroadCastNodes {
		if b.Address == address {
			return b.NodeID, nil
		}
	}
	return id, errors.New("not found")
}
