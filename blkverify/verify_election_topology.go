// Copyright (c) 2018-2019 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package blkverify

import (
	"github.com/matrix/go-matrix/baseinterface"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/params/manparams"
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

func (p *Process) verifyElection(header *types.Header, state *state.StateDB) error {
	info, err := p.reElection().GetElection(state, header.ParentHash)
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

func (p *Process) verifyNetTopology(header *types.Header, onlineConsensusResults []*mc.HD_OnlineConsensusVoteResultMsg) error {
	if header.NetTopology.Type == common.NetTopoTypeAll {
		return p.verifyAllNetTopology(header)
	}

	return p.verifyChgNetTopology(header, onlineConsensusResults)
}
func (p *Process) verifyVrf(header *types.Header) error {

	preBlock := p.blockChain().GetBlockByHash(header.ParentHash)
	if preBlock == nil {
		return errors.New("获取父区块失败")
	}
	if preBlock.Header() == nil {
		return errors.New("区块头为空")
	}

	SignAddr, err := p.blockChain().GetSignAccounts(header.Leader, header.ParentHash)
	if err != nil {
		log.Error(p.logExtraInfo(), "验证vrf失败", "获取真实签名账户失败")
		return errors.New("获取真实签名账户失败")
	}

	if len(SignAddr) <= 0 {
		log.Error(p.logExtraInfo(), "验证vrf失败", "真实签名账户数量为空")
		return errors.New("获取真实签名账户失败")
	}

	return baseinterface.NewVrf().VerifyVrf(header, preBlock.Header(), SignAddr[0])
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

func (p *Process) verifyChgNetTopology(header *types.Header, onlineConsensusResults []*mc.HD_OnlineConsensusVoteResultMsg) error {
	if len(header.NetTopology.NetTopologyData) == 0 {
		return nil
	}

	// get online and offline info from header and prev topology
	offlineNodes, onlineNods := p.parseOnlineState(header)
	log.INFO("scfffff-verify", "header.NetTop", header.NetTopology, "高度", header.Number.Uint64())
	log.INFO("scfffff--verify", "offlineNodes", offlineNodes)
	log.INFO("scfffff--verify", "onlineNods", onlineNods)

	for _, node := range offlineNodes {
		if err := p.checkConsensusResult(node, mc.OffLine, header, onlineConsensusResults); err != nil {
			return err
		}
	}
	for _, node := range onlineNods {
		if err := p.checkConsensusResult(node, mc.OnLine, header, onlineConsensusResults); err != nil {
			return err
		}
	}

	// generate topology alter info
	log.INFO("scffffff---Verify---GetTopoChange start ", "p.number", p.number, "offlineNodes", offlineNodes, "onlineNods", onlineNods)
	alterInfo, err := p.reElection().GetTopoChange(header.ParentHash, offlineNodes, onlineNods)
	log.INFO("scffffff---Verify---GetTopoChange end", "alterInfo", alterInfo, "err", err)
	if err != nil {
		return err
	}

	for _, value := range alterInfo {
		log.Info(p.logExtraInfo(), "alter-A", value.A, "position", value.Position)
	}
	// generate self net topology
	netTopology := p.reElection().TransferToNetTopologyChgStu(alterInfo)
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

func (p *Process) checkConsensusResult(node common.Address, state mc.OnlineState, header *types.Header, resultList []*mc.HD_OnlineConsensusVoteResultMsg) error {
	conResult := findResultInList(node, resultList)
	if conResult == nil {
		log.Error(p.logExtraInfo(), "检查拓扑变化消息", "online共识结果未找到", "node", node.Hex())
		return errors.New("online共识结果未找到")
	}
	if conResult.Req.OnlineState != state {
		log.Error(p.logExtraInfo(), "检查拓扑变化消息", "共识结果的状态不匹配", "node", node.Hex(), "结果状态", conResult.Req.OnlineState.String(), "头中状态", state.String())
		return errors.New("online共识结果未找到")
	}
	if conResult.IsValidity(p.number, manparams.OnlineConsensusValidityTime) == false {
		log.Error(p.logExtraInfo(), "检查拓扑变化消息", "online共识结果过期", "result.Number", conResult.Req.Number, "curNumber", p.number, "node", node.Hex())
		return errors.New("online共识结果过期")
	}
	reqHash := types.RlpHash(conResult.Req)
	blockHash, err := p.blockChain().GetAncestorHash(header.ParentHash, conResult.Req.Number-1)
	if err != nil {
		log.Error(p.logExtraInfo(), "检查拓扑变化消息", "online共识结果验证区块hash获取失败", "err", err, "node", node.Hex())
		return errors.New("online共识结果验证区块hash获取失败")
	}
	_, err = p.blockChain().DPOSEngine().VerifyHashWithBlock(p.blockChain(), reqHash, conResult.SignList, blockHash)
	if err != nil {
		log.Error(p.logExtraInfo(), "检查拓扑变化消息", "online共识结果POS失败", "node", node.Hex(), "err", err)
		return err
	}
	return nil
}

func (p *Process) parseOnlineState(header *types.Header) ([]common.Address, []common.Address) {
	if header.NetTopology.Type != common.NetTopoTypeChange {
		return nil, nil
	}

	online := make([]common.Address, 0)
	offline := make([]common.Address, 0)
	for _, v := range header.NetTopology.NetTopologyData {

		if v.Position == common.PosOffline {
			offline = append(offline, v.Account)
			continue
		}
		if v.Position == common.PosOnline {
			online = append(online, v.Account)
			continue
		}
	}
	return offline, online
}

func findResultInList(node common.Address, resultList []*mc.HD_OnlineConsensusVoteResultMsg) *mc.HD_OnlineConsensusVoteResultMsg {
	if len(resultList) == 0 {
		return nil
	}
	for i := 0; i < len(resultList); i++ {
		result := resultList[i]
		if nil == result || nil == result.Req {
			continue
		}
		if result.Req.Node == node {
			return result
		}
	}
	return nil
}
