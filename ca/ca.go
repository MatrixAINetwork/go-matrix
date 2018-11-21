// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package ca

import (
	"encoding/json"
	"errors"
	"math/big"
	"sync"

	"github.com/syndtr/goleveldb/leveldb"

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

// Identity stand for node's identity.
type Identity struct {
	// self nodeId
	self discover.NodeID
	addr common.Address

	// if in elected duration
	duration      bool
	currentHeight *big.Int

	// levelDB
	ldb *leveldb.DB

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

	// save id list
	availableId uint32
	idList      map[discover.NodeID]uint32

	// deposit in current height
	deposit []vm.DepositDetail

	// elect
	tempElect    []common.Elect
	originalRole []common.Elect
	// elect result: [role type, [nodeId]]
	elect map[common.Address]common.RoleType
	// current topology: [role type, [nodeId]]
	topology map[uint16]common.Address

	// current nodes
	currentNodes []discover.NodeID
	// front nodes
	frontNodes []discover.NodeID

	// addrByGroup
	addrByGroup map[common.RoleType][]common.Address

	// temp position
	position []uint16
}

var ide = newIde()

func newIde() *Identity {
	return &Identity{
		quit:        make(chan struct{}),
		currentRole: common.RoleNil,
		duration:    false,
		idList:      make(map[discover.NodeID]uint32),
		elect:       make(map[common.Address]common.RoleType),
		topology:    make(map[uint16]common.Address),
	}
}

// init to do something before run.
func (ide *Identity) init(id discover.NodeID, path string) {
	ide.once.Do(func() {
		// check bootNode and set identity
		ide.self = id
		//ide.ldb, _ = leveldb.OpenFile("./db/ca", nil)
		ide.log = log.New()
		var err error
		ide.ldb, err = leveldb.OpenFile(path+"./db/ca", nil)
		if nil != err {
			ide.log.Error("identity init error", err)
		} else {
			ide.log.Info("identity init over")
		}
		for _, b := range params.BroadCastNodes {
			if b.NodeID == id {
				ide.addr = b.Address
				break
			}
		}
	})
}

// Run this Identity.
func Start(id discover.NodeID, path string) {
	ide.init(id, path)

	defer func() {
		ide.ldb.Close()
		ide.sub.Unsubscribe()

		close(ide.quit)
		close(ide.blockChan)
	}()

	ide.blockChan = make(chan *types.Block)
	ide.sub, _ = mc.SubscribeEvent(mc.NewBlockMessage, ide.blockChan)

	for {
		select {
		case block := <-ide.blockChan:
			header := block.Header()
			log.INFO("CA", "leader", header.Leader, "height", header.Number.Uint64(), "block hash", block.Hash())
			ide.currentHeight = header.Number
			height := header.Number.Uint64()

			// init current height deposit
			ide.deposit, _ = GetElectedByHeightWithdraw(header.Number)

			// get self address from deposit
			ide.addr = GetAddress()

			switch {
			// validator elected block
			case common.IsReElectionNumber(height + VerifyNetChangeUpTime):
				{
					ide.duration = true
					ide.tempElect = append(ide.tempElect, header.Elect...)

					// maintain topology and check self next role
					ide.lock.Lock()
					for _, e := range header.Elect {
						ide.elect[e.Account] = e.Type.Transfer2CommonRole()
					}
					ide.lock.Unlock()
				}
			// miner elected block
			case common.IsReElectionNumber(height + MinerNetChangeUpTime):
				{
					ide.duration = true
					ide.tempElect = append(ide.tempElect, header.Elect...)

					// maintain topology and check self next role
					ide.lock.Lock()
					for _, e := range header.Elect {
						ide.elect[e.Account] = e.Type.Transfer2CommonRole()
					}
					ide.lock.Unlock()
				}
			// formal elected block
			case common.IsReElectionNumber(height + 1):
				{
					ide.duration = false
					ide.lock.Lock()
					ide.elect = make(map[common.Address]common.RoleType)
					ide.lock.Unlock()
					ide.originalRole = make([]common.Elect, 0)
					ide.originalRole = ide.tempElect
					ide.currentRole = common.RoleNil
				}
			case common.IsReElectionNumber(height) && height > 0:
				{
					ide.tempElect = make([]common.Elect, 0)
				}
			case height == 0:
				{
					ide.originalRole = header.Elect
				}
			default:
			}
			log.INFO("身份不对问题定位", "ide.originalRole", ide.originalRole)
			// do topology
			initCurrentTopology(header.NetTopology)

			for _, b := range params.BroadCastNodes {
				if b.NodeID == ide.self {
					ide.currentRole = common.RoleBroadcast
					break
				}
			}
			// change default role
			if ide.currentRole == common.RoleNil {
				ide.currentRole = common.RoleDefault
			}

			// init now topology: self peer
			initNowTopologyResult()

			// set topology graph and original role to levelDB
			setToLevelDB(ide.originalRole)

			// send role message to elect
			mc.PublishEvent(mc.CA_RoleUpdated, &mc.RoleUpdatedMsg{Role: ide.currentRole, BlockNum: header.Number.Uint64(), Leader: header.Leader})
			log.INFO("公布身份变更消息", "data", mc.RoleUpdatedMsg{Role: ide.currentRole, BlockNum: header.Number.Uint64(), Leader: header.Leader})
			// get nodes in buckets and send to buckets
			nodesInBuckets := getNodesInBuckets(header.Number)
			mc.PublishEvent(mc.BlockToBuckets, mc.BlockToBucket{Ms: nodesInBuckets, Height: block.Header().Number, Role: ide.currentRole})
			// send identity to linker
			mc.PublishEvent(mc.BlockToLinkers, mc.BlockToLinker{Height: header.Number, Role: ide.currentRole})
			mc.PublishEvent(mc.SendSyncRole, mc.SyncIdEvent{Role: ide.currentRole}) //lb
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
func initCurrentTopology(tp common.NetTopology) {
	switch tp.Type {
	case common.NetTopoTypeChange:
		for _, v := range tp.NetTopologyData {
			ide.topology[v.Position] = v.Account
			if v.Account == ide.addr {
				ide.currentRole = common.GetRoleTypeFromPosition(v.Position)
			}
			id, err := ConvertAddressToNodeId(v.Account)
			if err != nil {
				ide.log.Error("convert error", "ca", err)
				continue
			}
			maintainIdList(id)
		}
	case common.NetTopoTypeAll:
		ide.position = make([]uint16, 0)
		ide.topology = make(map[uint16]common.Address)
		for _, v := range tp.NetTopologyData {
			ide.position = append(ide.position, v.Position)
			ide.topology[v.Position] = v.Account
			if v.Account == ide.addr {
				ide.currentRole = common.GetRoleTypeFromPosition(v.Position)
			}
			id, err := ConvertAddressToNodeId(v.Account)
			if err != nil {
				ide.log.Error("convert error", "ca", err)
				continue
			}
			maintainIdList(id)
		}
	}
	log.INFO("当前拓扑信息", "ide.topology", ide.topology)
}

// initNowTopologyResult
func initNowTopologyResult() {
	ide.lock.Lock()
	defer ide.lock.Unlock()
	ide.addrByGroup = make(map[common.RoleType][]common.Address)
	for po, addr := range ide.topology {
		role := common.GetRoleTypeFromPosition(po)
		if role == common.RoleBackupValidator {
			role = common.RoleValidator
		}
		if role == common.RoleBackupMiner {
			role = common.RoleMiner
		}
		ide.addrByGroup[role] = append(ide.addrByGroup[role], addr)
	}
	for _, b := range params.BroadCastNodes {
		ide.addrByGroup[common.RoleBroadcast] = append(ide.addrByGroup[common.RoleBroadcast], b.Address)
	}
}

// GetRolesByGroup
func GetRolesByGroup(roleType common.RoleType) (result []discover.NodeID) {
	ide.lock.RLock()
	defer ide.lock.RUnlock()
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
	return
}

// GetRolesByGroupWithBackup
func GetRolesByGroupWithBackup(roleType common.RoleType) (result []discover.NodeID) {
	ide.lock.RLock()
	defer ide.lock.RUnlock()

	result = GetRolesByGroup(roleType)
	for addr, role := range ide.elect {
		temp := true
		if (role & roleType) != 0 {
			id, err := ConvertAddressToNodeId(addr)
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
func GetRolesByGroupOnlyBackup(roleType common.RoleType) (result []discover.NodeID) {
	ide.lock.RLock()
	defer ide.lock.RUnlock()

	for addr, role := range ide.elect {
		if (role & roleType) != 0 {
			id, err := ConvertAddressToNodeId(addr)
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

// maintainIdList
func maintainIdList(d discover.NodeID) {
	ide.lock.Lock()
	if _, ok := ide.idList[d]; !ok {
		ide.idList[d] = ide.availableId
		ide.availableId++
	}
	ide.lock.Unlock()
}

// GetNodeNumber
func GetNodeNumber() (uint32, error) {
	ide.lock.RLock()
	defer ide.lock.RUnlock()
	for k, v := range ide.idList {
		if k == ide.self {
			return v, nil
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
	for _, v := range ide.topology {
		for key := range msMap {
			if key == v {
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
	defer ide.lock.RUnlock()
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

	for addr, role := range ide.elect {
		temp := true
		id, err := ConvertAddressToNodeId(addr)
		if err != nil {
			ide.log.Error("convert error", "ca", err)
			continue
		}
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
	return ide.frontNodes
}

func setToLevelDB(elect []common.Elect) error {
	var (
		tg = new(mc.TopologyGraph)
		tp = make([]mc.TopologyNodeInfo, 0)
	)

	tg.Number = GetHeight()

	// set topology graph
	for key, val := range ide.topology {
		st := mc.TopologyNodeInfo{}
		for _, e := range elect {
			if e.Account == val {
				st.Stock = e.Stock
				break
			}
		}
		st.Type = common.GetRoleTypeFromPosition(key)
		st.Account = val
		st.Position = key
		tp = append(tp, st)
	}
	tg.NodeList = tp

	bytes, err := json.Marshal(tg)
	if err != nil {
		return err
	}
	tgBytes := append(tg.Number.Bytes(), []byte(LevelDBTopologyGraph)...)
	ide.ldb.Put(tgBytes, bytes, nil)

	// set original role
	originalRoleBytes, err := json.Marshal(ide.originalRole)
	if err != nil {
		return err
	}
	orBytes := append(tg.Number.Bytes(), []byte(LevelDBOriginalRole)...)
	return ide.ldb.Put(orBytes, originalRoleBytes, nil)
}

// GetAddress
func GetAddress() common.Address {
	for _, node := range ide.deposit {
		if node.NodeID == ide.self {
			return node.Address
		}
	}
	for _, b := range params.BroadCastNodes {
		if b.NodeID == ide.self {
			return b.Address
		}
	}
	return common.Address{}
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
	tgBytes := append(big.NewInt(int64(number)).Bytes(), []byte(LevelDBTopologyGraph)...)
	val, err := ide.ldb.Get(tgBytes, nil)
	if err != nil {
		return nil, err
	}

	es := new(mc.TopologyGraph)
	err = json.Unmarshal(val, &es)
	if err != nil {
		return nil, err
	}

	newEs := new(mc.TopologyGraph)
	newEs.Number = es.Number
	tempList := make([]mc.TopologyNodeInfo, 0)
	for _, tg := range es.NodeList {
		if (tg.Type & reqTypes) != 0 {
			tempList = append(tempList, tg)
		}
	}

	for _, p := range ide.position {
		for _, n := range tempList {
			if p == n.Position {
				newEs.NodeList = append(newEs.NodeList, n)
				break
			}
		}
	}

	return newEs, nil
}

// GetAccountTopologyInfo
func GetAccountTopologyInfo(account common.Address, number uint64) (*mc.TopologyNodeInfo, error) {
	tgBytes := append(big.NewInt(int64(number)).Bytes(), []byte(LevelDBTopologyGraph)...)
	val, err := ide.ldb.Get(tgBytes, nil)
	if err != nil {
		return nil, err
	}

	es := new(mc.TopologyGraph)
	err = json.Unmarshal(val, &es)
	if err != nil {
		return nil, err
	}

	for _, tg := range es.NodeList {
		if tg.Account == account {
			return &tg, nil
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
	orBytes := append(big.NewInt(int64(number)).Bytes(), []byte(LevelDBOriginalRole)...)
	val, err := ide.ldb.Get(orBytes, nil)
	if err != nil {
		return common.RoleNil, err
	}

	es := make([]common.Elect, 0)
	err = json.Unmarshal(val, &es)
	if err != nil {
		return common.RoleNil, err
	}

	for _, tg := range es {
		if tg.Account == account {
			return tg.Type.Transfer2CommonRole(), nil
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
