// Copyright (c) 2018-2019 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package blkgenor

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/matrixstate"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/params/manparams"
)

func (p *Process) genElection(state *state.StateDB) []common.Elect {
	info, err := p.reElection().GetElection(state, p.preBlockHash)
	if err != nil {
		log.Warn(p.logExtraInfo(), "获取选举信息错误", err)
		return nil
	}

	return p.reElection().TransferToElectionStu(info)
}

func (p *Process) getNetTopology(num uint64, parentHash common.Hash, bcInterval *mc.BCIntervalInfo) (*common.NetTopology, []*mc.HD_OnlineConsensusVoteResultMsg) {
	if bcInterval.IsReElectionNumber(num + 1) {
		return p.genAllNetTopology(parentHash)
	}

	return p.genChgNetTopology(parentHash)
}

func (p *Process) genAllNetTopology(parentHash common.Hash) (*common.NetTopology, []*mc.HD_OnlineConsensusVoteResultMsg) {
	info, err := p.reElection().GetNetTopologyAll(parentHash)
	if err != nil {
		log.Warn(p.logExtraInfo(), "获取拓扑图错误", err)
		return nil, nil
	}

	return p.reElection().TransferToNetTopologyAllStu(info), nil
}

func (p *Process) genChgNetTopology(parentHash common.Hash) (*common.NetTopology, []*mc.HD_OnlineConsensusVoteResultMsg) {
	state, err := p.blockChain().StateAtBlockHash(parentHash)
	if err != nil {
		log.Warn(p.logExtraInfo(), "生成拓扑变化", "获取父状态树失败", "err", err)
		return nil, nil
	}

	topology, err := matrixstate.GetTopologyGraph(state)
	if err != nil || topology == nil {
		log.Warn(p.logExtraInfo(), "生成拓扑变化", "状态树获取拓扑图失败", "err", err)
		return nil, nil
	}

	electState, err := matrixstate.GetElectOnlineState(state)
	if err != nil || electState == nil {
		log.Warn(p.logExtraInfo(), "生成拓扑变化", "状态树获取elect在线状态失败", "err", err)
		return nil, nil
	}

	onlineResults := p.topNode().GetConsensusOnlineResults()
	if len(onlineResults) == 0 {
		log.Info(p.logExtraInfo(), "生成拓扑变化信息", "无在线共识结果")
		return nil, nil
	}

	offlineNodes, onlineNods, consensusList := p.getOnlineStatus(onlineResults, topology, electState)

	// generate topology alter info
	alterInfo, err := p.reElection().GetTopoChange(parentHash, offlineNodes, onlineNods)
	if err != nil {
		log.Warn(p.logExtraInfo(), "获取拓扑变化信息错误", err)
		return nil, nil
	}
	for _, value := range alterInfo {
		log.Debug(p.logExtraInfo(), "获取拓扑变化地址", value.A, "位置", value.Position, "高度", p.number)
	}

	// generate self net topology
	ans := p.reElection().TransferToNetTopologyChgStu(alterInfo)
	return ans, consensusList
}

func (p *Process) getOnlineStatus(onlineResults []*mc.HD_OnlineConsensusVoteResultMsg, topology *mc.TopologyGraph, electState *mc.ElectOnlineStatus) ([]common.Address, []common.Address, []*mc.HD_OnlineConsensusVoteResultMsg) {
	offlineNodes := make([]common.Address, 0)
	onlineNods := make([]common.Address, 0)
	consensusList := make([]*mc.HD_OnlineConsensusVoteResultMsg, 0)
	// 筛选共识结果
	for i := 0; i < len(onlineResults); i++ {
		result := onlineResults[i]
		if result.Req == nil {
			log.Info(p.logExtraInfo(), "生成拓扑变化信息", "共识请求为空")
			continue
		}
		if result.IsValidity(p.number, manparams.OnlineConsensusValidityTime) == false {
			log.Info(p.logExtraInfo(), "生成拓扑变化信息", "高度 不合法", "请求高度", result.Req.Number, "当前高度", p.number)
			continue
		}

		node := result.Req.Node
		state := result.Req.OnlineState
		// 节点为当前拓扑图节点
		if topology.AccountIsInGraph(node) {
			if state == mc.OffLine {
				offlineNodes = append(offlineNodes, node)
				consensusList = append(consensusList, result)
			} else {
				log.Info(p.logExtraInfo(), "生成拓扑变化信息", "当前拓扑图中的节点，顶点共识状态错误", "状态", state)
			}
			continue
		}

		// 查看节点elect信息
		electInfo := electState.FindNodeElectOnlineState(node)
		if electInfo == nil {
			// 没有elect信息，表明节点非elect节点，不关心上下线信息
			continue
		}
		switch state {
		case mc.OnLine:
			if electInfo.Position == common.PosOffline {
				// 链上状态离线，当前共识结果在线，则需要上header
				onlineNods = append(onlineNods, node)
				consensusList = append(consensusList, result)
			}
		case mc.OffLine:
			if electInfo.Position == common.PosOnline {
				// 链上状态在线，当前共识结果离线，则需要上header
				offlineNodes = append(offlineNodes, node)
				consensusList = append(consensusList, result)
			}
		default:
			continue
		}
	}
	for i, value := range onlineNods {
		log.Debug(p.logExtraInfo(), "下线节点地址", value.String(), "序号", i)
	}
	for i, value := range offlineNodes {
		log.Debug(p.logExtraInfo(), "上线节点地址", value.String(), "序号", i)
	}
	return offlineNodes, onlineNods, consensusList
}
