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
	info, err := p.reElection().GetElection(header.ParentHash)
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
	info, err := p.reElection().GetNetTopologyAll(header.ParentHash)
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
	prevTopology, err := ca.GetTopologyByHash(common.RoleValidator|common.RoleBackupValidator|common.RoleMiner|common.RoleBackupMiner, header.ParentHash)
	if err != nil {
		return err
	}

	// get online and offline info from header and prev topology
	offlineTopNodes, onlinePrimaryNods, offlinePrimaryNodes := p.parseOnlineState(header, prevTopology)

	log.INFO("scfffff-verify", "header.NetTop", header.NetTopology, "高度", header.Number.Uint64())
	log.INFO("scfffff--verify", "prevTopology", prevTopology)
	log.INFO("scfffff--verify", "offlineTopNodes", offlineTopNodes)
	log.INFO("scfffff--verify", "onlinePrimaryNods", onlinePrimaryNods)
	log.INFO("scfffff--verify", "offlinePrimaryNodes", offlinePrimaryNodes)

	originTopology, err := ca.GetTopologyByHash(common.RoleValidator|common.RoleBackupValidator|common.RoleMiner|common.RoleBackupMiner, header.ParentHash)
	if err != nil {
		return nil
	}
	originTopNodes := make([]common.Address, 0)
	for _, node := range originTopology.NodeList {
		originTopNodes = append(originTopNodes, node.Account)
		log.Info(p.logExtraInfo(), "originTopNode", node.Account)
	}
	onlineElectNodes := make([]common.Address, 0)
	for _, node := range originTopology.ElectList {
		onlineElectNodes = append(onlineElectNodes, node.Account)
		log.Info(p.logExtraInfo(), "onlineElectNodes", node.Account)

	}
	electNumber:=common.GetLastReElectionNumber(p.number)
	if electNumber > 0 {
		electNumber -= 1
	}
	p.pm.topNode.SetElectNodes(originTopNodes, electNumber)

	if false == p.topNode().CheckAddressConsensusOnlineState(offlineTopNodes, onlinePrimaryNods, offlinePrimaryNodes) {
		return errTopNodeState
	}

	// generate topology alter info
	log.INFO("scffffff---Verify---GetTopoChange start ", "p.number", p.number, "offlineTopNodes", offlineTopNodes, "onlinePrimaryNods", onlinePrimaryNods)
	alterInfo, err := p.reElection().GetTopoChange(header.ParentHash, offlineTopNodes)
	log.INFO("scffffff---Verify---GetTopoChange end", "alterInfo", alterInfo, "err", err)
	if err != nil {
		return err
	}

	for _, value := range alterInfo {
		log.Info(p.logExtraInfo(), "alter-A", value.A, "position", value.Position)
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
