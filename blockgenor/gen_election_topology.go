//1542616417.2823353
//1542615530.1412451
//1542614826.4191656
// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package blockgenor

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

	// get prev topology
	prevTopology, err := p.getPrevTopology(parentHash)
	if err != nil {
		log.Warn(p.logExtraInfo(), "get prev topology err", err)
		return nil
	}

	// get online and offline info from header and prev topology
	offlineTopNodes, onlinePrimaryNods, offlinePrimaryNodes := p.parseOnlineState(currentNetTopology, prevTopology)

	// generate topology alter info
	alterInfo, err := p.reElection().GetTopoChange(parentHash, offlineTopNodes)
	if err != nil {
		log.Warn(p.logExtraInfo(), "get topology change info by reelection server err", err)
		return nil
	}

	// generate self net topology
	return p.reElection().TransferToNetTopologyChgStu(alterInfo, onlinePrimaryNods, offlinePrimaryNodes)
}
