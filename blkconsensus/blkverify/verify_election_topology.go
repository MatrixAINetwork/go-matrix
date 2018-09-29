// Copyright 2018 The MATRIX Authors as well as Copyright 2014-2017 The go-ethereum Authors
// This file is consisted of the MATRIX library and part of the go-ethereum library.
//
// The MATRIX-ethereum library is free software: you can redistribute it and/or modify it under the terms of the MIT License.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, 
//and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject tothe following conditions:
//
//The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
//
//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, 
//WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISINGFROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE
//OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
package blkverify

import (
	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/topnode"
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
	stateMap map[common.Address]topnode.OnlineState) bool {

	for _, offlineNode := range offlineNodes {
		if consensusState, OK := stateMap[offlineNode]; OK == false {
			log.ERROR(p.logExtraInfo(), "拓扑变化检测(通过本地共识状态), 本地共识状态未找到, node", offlineNode, "区块请求中的状态", "下线")
			return false
		} else if consensusState != topnode.Offline {
			log.Warn(p.logExtraInfo(), "拓扑变化检测(通过本地共识状态), 本地共识状态不匹配, node", offlineNode, "区块请求中的状态", "下线")
			return false
		}
	}

	for _, onlineNode := range onlineNodes {
		if consensusState, OK := stateMap[onlineNode]; OK == false {
			log.Warn(p.logExtraInfo(), "拓扑变化检测(通过本地共识状态), 本地共识状态未找到, node", onlineNode, "区块请求中的状态", "上线")
			return false
		} else if consensusState != topnode.Online {
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
