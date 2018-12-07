// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package blkgenor

import (
	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/params/manparams"
)

func (p *Process) genElection(parentHash common.Hash) []common.Elect {
	info, err := p.reElection().GetElection(parentHash)
	if err != nil {
		log.Warn(p.logExtraInfo(), "verifyElection: get election err", err)
		return nil
	}

	return p.reElection().TransferToElectionStu(info)
}

func (p *Process) getNetTopology(num uint64, parentHash common.Hash) (*common.NetTopology, []*mc.HD_OnlineConsensusVoteResultMsg) {
	if common.IsReElectionNumber(num + 1) {
		return p.genAllNetTopology(parentHash)
	}

	return p.genChgNetTopology(parentHash)
}

func (p *Process) genAllNetTopology(parentHash common.Hash) (*common.NetTopology, []*mc.HD_OnlineConsensusVoteResultMsg) {
	info, err := p.reElection().GetNetTopologyAll(parentHash)
	if err != nil {
		log.Warn(p.logExtraInfo(), "verifyNetTopology: get prev topology from ca err", err)
		return nil, nil
	}

	return p.reElection().TransferToNetTopologyAllStu(info), nil
}

func (p *Process) genChgNetTopology(parentHash common.Hash) (*common.NetTopology, []*mc.HD_OnlineConsensusVoteResultMsg) {
	currentTopology, err := ca.GetTopologyByHash(common.RoleValidator|common.RoleBackupValidator, parentHash)
	if err != nil {
		log.Warn(p.logExtraInfo(), "get topology by parentHash error", err)
		return nil, nil
	}

	onlineResults := p.topNode().GetConsensusOnlineResults()
	if len(onlineResults) == 0 {
		log.Info(p.logExtraInfo(), "生成拓扑变化信息", "无在线共识结果, return")
		return nil, nil
	}

	offlineTopNodes := make([]common.Address, 0)
	onlineElectNods := make([]common.Address, 0)
	offlineElectNodes := make([]common.Address, 0)
	consensusList := make([]*mc.HD_OnlineConsensusVoteResultMsg, 0)

	// 筛选共识结果
	for i := 0; i < len(onlineResults); i++ {
		result := onlineResults[i]
		if result.Req == nil {
			log.Info(p.logExtraInfo(), "生成拓扑变化信息", "result.Req = nil")
			continue
		}
		if result.IsValidity(p.number, manparams.OnlineConsensusValidityTime) == false {
			log.Info(p.logExtraInfo(), "生成拓扑变化信息", "result.Number 不合法", "result.Number", result.Req.Number, "curNumber", p.number)
			continue
		}

		node := result.Req.Node
		state := result.Req.OnlineState
		// 节点为当前拓扑图节点
		if currentTopology.AccountIsInGraph(node) {
			if state == mc.OffLine {
				offlineTopNodes = append(offlineTopNodes, node)
				consensusList = append(consensusList, result)
			} else {
				log.Info(p.logExtraInfo(), "生成拓扑变化信息", "当前拓扑图中的节点，顶点共识状态错误", "状态", state)
			}
			continue
		}

		// 查看节点elect信息
		electInfo := currentTopology.GetAccountElectInfo(node)
		if electInfo == nil {
			// 没有elect信息，表明节点非elect节点，不关心上下线信息
			continue
		}
		switch state {
		case mc.OnLine:
			if electInfo.Position == common.PosOffline {
				// 链上状态离线，当前共识结果在线，则需要上header
				onlineElectNods = append(onlineElectNods, node)
				consensusList = append(consensusList, result)
			}
		case mc.OffLine:
			if electInfo.Position == common.PosOnline {
				// 链上状态在线，当前共识结果离线，则需要上header
				offlineElectNodes = append(offlineElectNodes, node)
				consensusList = append(consensusList, result)
			}
		default:
			continue
		}
	}

	for i, value := range offlineTopNodes {
		log.Info(p.logExtraInfo(), "offlineTopNodes", value.String(), "index", i)
	}
	for i, value := range onlineElectNods {
		log.Info(p.logExtraInfo(), "onlineElectNods", value.String(), "index", i)
	}
	for i, value := range offlineElectNodes {
		log.Info(p.logExtraInfo(), "offlineElectNodes", value.String(), "index", i)
	}

	// generate topology alter info
	alterInfo, err := p.reElection().GetTopoChange(parentHash, offlineTopNodes)
	if err != nil {
		log.Warn(p.logExtraInfo(), "get topology change info by reelection server err", err)
		return nil, nil
	}
	for _, value := range alterInfo {
		log.Info(p.logExtraInfo(), "alter-A", value.A, "alter-B", "", "position", value.Position, "number", p.number)
	}

	// generate self net topology
	ans := p.reElection().TransferToNetTopologyChgStu(alterInfo, onlineElectNods, offlineElectNodes)
	log.INFO("scfffff-TransferToNetTopologyChgStu", "ans", ans)
	return ans, consensusList
}
