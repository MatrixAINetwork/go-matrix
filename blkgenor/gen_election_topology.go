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

func (p *Process) genElection(num uint64) []common.Elect {
	info, err := p.reElection().GetElection(num)
	if err != nil {
		log.Warn(p.logExtraInfo(), "verifyElection: get election err", err)
		return nil
	}

	return p.reElection().TransferToElectionStu(info)
}

func (p *Process) getNetTopology(currentNetTopology common.NetTopology, num uint64) *common.NetTopology {
	if common.IsReElectionNumber(num + 1) {
		return p.genAllNetTopology(num)
	}

	return p.genChgNetTopology(currentNetTopology, num)
}

func (p *Process) genAllNetTopology(num uint64) *common.NetTopology {
	info, err := p.reElection().GetNetTopologyAll(num)
	if err != nil {
		log.Warn(p.logExtraInfo(), "verifyNetTopology: get prev topology from ca err", err)
		return nil
	}

	return p.reElection().TransferToNetTopologyAllStu(info)
}

func (p *Process) getPrevTopology(num uint64) (*mc.TopologyGraph, error) {
	reqRoles := common.RoleType(common.RoleValidator | common.RoleBackupValidator | common.RoleMiner | common.RoleBackupMiner)

	return ca.GetTopologyByNumber(reqRoles, num-1)
}

func (p *Process) parseOnlineState(currentNetTopology common.NetTopology, prevTopology *mc.TopologyGraph) ([]common.Address, []common.Address, []common.Address) {
	offlineTopNodes := p.reElection().ParseTopNodeOffline(currentNetTopology, prevTopology)
	onlinePrimaryNods, offlinePrimaryNodes := p.reElection().ParsePrimaryTopNodeState(currentNetTopology)
	return offlineTopNodes, onlinePrimaryNods, offlinePrimaryNodes
}

func (p *Process) genChgNetTopology(currentNetTopology common.NetTopology, num uint64) *common.NetTopology {

	// get prev topology
	prevTopology, err := p.getPrevTopology(num)
	if err != nil {
		log.Warn(p.logExtraInfo(), "get prev topology err", err)
		return nil
	}

	// get online and offline info from header and prev topology
	offlineTopNodes, onlinePrimaryNods, offlinePrimaryNodes := p.parseOnlineState(currentNetTopology, prevTopology)

	// generate topology alter info
	alterInfo, err := p.reElection().GetTopoChange(num, offlineTopNodes)
	if err != nil {
		log.Warn(p.logExtraInfo(), "get topology change info by reelection server err", err)
		return nil
	}

	// generate self net topology
	return p.reElection().TransferToNetTopologyChgStu(alterInfo, onlinePrimaryNods, offlinePrimaryNodes)
}
