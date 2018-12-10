// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package blkverify

import (
	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/olconsensus"
	"github.com/pkg/errors"
)

var (
	errGetElection      = errors.New("get election info err")
	errElectionSize     = errors.New("election count not match")
	errElectionInfo     = errors.New("election info not match")
	errTopoSize         = errors.New("topology count not match")
	errTopoInfo         = errors.New("topology info not match")
	errTopNodeState     = errors.New("cur top node consensus state not match")
	errPrimaryNodeState = errors.New("primary node consensus state not match")
)

func (p *Process) verifyElection(header *types.Header) error {
	info, err := p.reElection().GetElection(p.number)
	if err != nil {
		return errGetElection
	}

	electInfo := p.reElection().TransferToElectionStu(info)
	if len(electInfo) != len(header.Elect) {
		return errElectionSize
	}

	if len(electInfo) == 0 {
		return nil
	}

	targetRlp := types.RlpHash(header.Elect)
	selfRlp := types.RlpHash(electInfo)
	if targetRlp != selfRlp {
		return errElectionInfo
	}
	return nil
}

func (p *Process) verifyNetTopology(header *types.Header) error {
	if header.NetTopology.Type == common.NetTopoTypeAll {
		return p.verifyAllNetTopology(header)
	}

	return p.verifyChgNetTopology(header)
}

func (p *Process) verifyAllNetTopology(header *types.Header) error {
	info, err := p.reElection().GetNetTopologyAll(header.Number.Uint64())
	if err != nil {
		return err
	}

	netTopology := p.reElection().TransferToNetTopologyAllStu(info)
	if len(netTopology.NetTopologyData) != len(header.NetTopology.NetTopologyData) {
		return errTopoSize
	}

	targetRlp := types.RlpHash(&header.NetTopology)
	selfRlp := types.RlpHash(netTopology)

	if targetRlp != selfRlp {
		return errTopoInfo
	}
	return nil
}

func (p *Process) verifyChgNetTopology(header *types.Header) error {
	if len(header.NetTopology.NetTopologyData) == 0 {
		return nil
	}

	// get prev topology
	prevTopology, err := ca.GetTopologyByNumber(common.RoleValidator|common.RoleBackupValidator|common.RoleMiner|common.RoleBackupMiner, p.number-1)
	if err != nil {
		return err
	}

	// get online and offline info from header and prev topology
	offlineTopNodes, onlinePrimaryNods, offlinePrimaryNodes := p.parseOnlineState(header, prevTopology)

	// get local consensus on-line state
	curTopNodeState, primaryNodeState := p.topNode().GetConsensusOnlineState()
	//check state from header with state in local consensus
	if false == p.checkStateByConsensus(offlineTopNodes, nil, curTopNodeState) {
		return errTopNodeState
	}
	if false == p.checkStateByConsensus(offlinePrimaryNodes, onlinePrimaryNods, primaryNodeState) {
		return errPrimaryNodeState
	}

	// generate topology alter info
	alterInfo, err := p.reElection().GetTopoChange(p.number, offlineTopNodes)
	if err != nil {
		return err
	}

	// generate self net topology
	netTopology := p.reElection().TransferToNetTopologyChgStu(alterInfo, onlinePrimaryNods, offlinePrimaryNodes)
	if len(netTopology.NetTopologyData) != len(header.NetTopology.NetTopologyData) {
		return errTopoSize
	}

	targetRlp := types.RlpHash(&header.NetTopology)
	selfRlp := types.RlpHash(netTopology)
	if targetRlp != selfRlp {
		return errTopoInfo
	}
	return nil
}

func (p *Process) checkStateByConsensus(offlineNodes []common.Address,
	onlineNodes []common.Address,
	stateMap map[common.Address]olconsensus.OnlineState) bool {

	for _, offlineNode := range offlineNodes {
		if consensusState, OK := stateMap[offlineNode]; OK == false {
			log.ERROR(p.logExtraInfo(), "拓扑变化检测(通过本地共识状态), 本地共识状态未找到, node", offlineNode, "区块请求中的状态", "下线")
			return false
		} else if consensusState != olconsensus.Offline {
			log.Warn(p.logExtraInfo(), "拓扑变化检测(通过本地共识状态), 本地共识状态不匹配, node", offlineNode, "区块请求中的状态", "下线")
			return false
		}
	}

	for _, onlineNode := range onlineNodes {
		if consensusState, OK := stateMap[onlineNode]; OK == false {
			log.Warn(p.logExtraInfo(), "拓扑变化检测(通过本地共识状态), 本地共识状态未找到, node", onlineNode, "区块请求中的状态", "上线")
			return false
		} else if consensusState != olconsensus.Online {
			log.Warn(p.logExtraInfo(), "拓扑变化检测(通过本地共识状态), 本地共识状态不匹配, node", onlineNode, "区块请求中的状态", "上线")
			return false
		}
	}

	return true
}

func (p *Process) parseOnlineState(header *types.Header, prevTopology *mc.TopologyGraph) ([]common.Address, []common.Address, []common.Address) {
	offlineTopNodes := p.reElection().ParseTopNodeOffline(header.NetTopology, prevTopology)
	onlinePrimaryNods, offlinePrimaryNodes := p.reElection().ParsePrimaryTopNodeState(header.NetTopology)
	return offlineTopNodes, onlinePrimaryNods, offlinePrimaryNodes
}
