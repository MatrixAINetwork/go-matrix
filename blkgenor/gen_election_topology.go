// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package blkgenor

import (
	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
)

func (p *Process) genElection(parentHash common.Hash) []common.Elect {
	info, err := p.reElection().GetElection(parentHash)
	if err != nil {
		log.Warn(p.logExtraInfo(), "verifyElection: get election err", err)
		return nil
	}

	return p.reElection().TransferToElectionStu(info)
}

func (p *Process) getNetTopology(currentNetTopology common.NetTopology, num uint64,parentHash common.Hash) *common.NetTopology {
	if common.IsReElectionNumber(num + 1) {
		return p.genAllNetTopology(parentHash)
	}

	return p.genChgNetTopology(currentNetTopology, parentHash)
}

func (p *Process) genAllNetTopology(parentHash common.Hash) *common.NetTopology {
	info, err := p.reElection().GetNetTopologyAll(parentHash)
	if err != nil {
		log.Warn(p.logExtraInfo(), "verifyNetTopology: get prev topology from ca err", err)
		return nil
	}

	return p.reElection().TransferToNetTopologyAllStu(info)
}

func (p *Process) getPrevTopology(parentHash common.Hash) (*mc.TopologyGraph, error) {
	reqRoles := common.RoleType(common.RoleValidator | common.RoleBackupValidator | common.RoleMiner | common.RoleBackupMiner)
	return ca.GetTopologyByHash(reqRoles, parentHash)
}

func (p *Process) parseOnlineState(currentNetTopology common.NetTopology, prevTopology *mc.TopologyGraph) ([]common.Address, []common.Address, []common.Address) {
	offlineTopNodes := p.reElection().ParseTopNodeOffline(currentNetTopology, prevTopology)
	onlinePrimaryNods, offlinePrimaryNodes := p.reElection().ParsePrimaryTopNodeState(currentNetTopology)
	return offlineTopNodes, onlinePrimaryNods, offlinePrimaryNodes
}

func (p *Process) genChgNetTopology(currentNetTopology common.NetTopology, parentHash common.Hash) *common.NetTopology {

	// get local consensus on-line state
	var eleNum uint64
	if p.number < common.GetReElectionInterval() {
		eleNum = 0
	} else {
		eleNum = common.GetLastReElectionNumber(p.number) - 1
	}
	originTopology, err := ca.GetTopologyByNumber(common.RoleValidator|common.RoleBackupValidator|common.RoleMiner|common.RoleBackupMiner, eleNum)
	if err != nil {
		log.Warn(p.logExtraInfo(), "get topology by number error", err)
		return nil
	}
	originTopNodes := make([]common.Address, 0)
	for _, node := range originTopology.NodeList {
		originTopNodes = append(originTopNodes, node.Account)
	}

	p.pm.olConsensus.SetElectNodes(originTopNodes, eleNum)
	//var currentNum uint64

	//if p.number < 1 {
	//	currentNum = 0
	//} else {
	//	currentNum = p.number - 1
	//}
	// get prev topology
	//currentTopology, err := ca.GetTopologyByNumber(common.RoleValidator|common.RoleBackupValidator|common.RoleMiner|common.RoleBackupMiner, currentNum)
	currentTopology, err := ca.GetTopologyByHash(common.RoleValidator|common.RoleBackupValidator|common.RoleMiner|common.RoleBackupMiner, parentHash)

	if err != nil {
		log.Warn(p.logExtraInfo(), "get topology by number error", err)
		return nil
	}

	// get online and offline info from header and prev topology
	onlineTopNodes := make([]common.Address, 0)
	for _, node := range currentTopology.NodeList {
		onlineTopNodes = append(onlineTopNodes, node.Account)
		log.Info(p.logExtraInfo(), "onlineTopNodes", node.Account)

	}

	onlineElectNodes := make([]common.Address, 0)
	for _, node := range currentTopology.ElectList {
		onlineElectNodes = append(onlineElectNodes, node.Account)
		log.Info(p.logExtraInfo(), "onlineElectNodes", node.Account)

	}
	log.Info(p.logExtraInfo(), "SetCurrentOnlineState:高度", p.number, "onlineElect", len(onlineElectNodes), "onlineTopnode", len(onlineTopNodes))

	p.pm.olConsensus.SetCurrentOnlineState(onlineTopNodes, onlineElectNodes)
	offlineTopNodes, onlineElectNods, offlineElectNodes := p.pm.olConsensus.GetConsensusOnlineState()

	for _, value := range offlineTopNodes {
		log.Info(p.logExtraInfo(), "offlineTopNodes", value.String())
	}
	for _, value := range onlineElectNods {
		log.Info(p.logExtraInfo(), "onlineElectNods", value.String())
	}

	for _, value := range offlineElectNodes {
		log.Info(p.logExtraInfo(), "offlineElectNodes", value.String())
	}

	saveData := []common.Address{}
	for k, v := range offlineElectNodes {
		flag := 0
		for _, vv := range offlineTopNodes {
			if v == vv {
				flag = 1
			}
		}
		if flag == 0 {
			saveData = append(saveData, offlineElectNodes[k])
		}
	}
	log.INFO("scfffff", "saveIndex", saveData)

	offlineElectNodes = saveData
	log.INFO("scffffff-Gen-GetTopoChange start ", "hash", parentHash.String(), "onlineElectNods", onlineElectNods, "offlineElectNodes", offlineElectNodes, "offlineTopNodes", offlineTopNodes)
	// generate topology alter info
	alterInfo, err := p.reElection().GetTopoChange(parentHash, offlineTopNodes)
	log.INFO("scffffff-Gen-GetTopoChange end", "alterInfo", alterInfo, "err", err)
	if err != nil {
		log.Warn(p.logExtraInfo(), "get topology change info by reelection server err", err)
		return nil
	}
	for _, value := range alterInfo {
		log.Info(p.logExtraInfo(), "alter-A", value.A, "alter-B","", "position", value.Position, "number", p.number)
	}

	// generate self net topology
	ans := p.reElection().TransferToNetTopologyChgStu(alterInfo, onlineElectNods, offlineElectNodes)
	log.INFO("scfffff-TransferToNetTopologyChgStu", "ans", ans)
	return ans
}
