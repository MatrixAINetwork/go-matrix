// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package reelection

import (
	"github.com/MatrixAINetwork/go-matrix/baseinterface"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/params/manparams"
	"github.com/MatrixAINetwork/go-matrix/params/manversion"
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

func (p *ReElection) VerifyElection(header *types.Header, state *state.StateDBManage) error {
	info, err := p.GetElection(state, header.ParentHash)
	if err != nil {
		return errGetElection
	}

	electInfo := p.TransferToElectionStu(info)
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

func (p *ReElection) VerifyNetTopology(version string, header *types.Header, onlineConsensusResults []*mc.HD_OnlineConsensusVoteResultMsg) error {
	if header.NetTopology.Type == common.NetTopoTypeAll {
		return p.verifyAllNetTopology(header)
	}

	return p.verifyChgNetTopology(version, header, onlineConsensusResults)
}
func (p *ReElection) VerifyVrf(header *types.Header) error {

	preBlock := p.bc.GetBlockByHash(header.ParentHash)
	if preBlock == nil {
		return errors.New("获取父区块失败")
	}
	if preBlock.Header() == nil {
		return errors.New("区块头为空")
	}

	vrfSignAccount, err := baseinterface.NewVrf().DecodeVrf(header, preBlock.Header())
	if err != nil {
		log.Error(Module, "DecodeVrf err", err)
		return err
	}

	accountA0, _, err := p.bc.GetA0AccountFromAnyAccount(vrfSignAccount, header.ParentHash)
	if err != nil {
		log.Error(Module, "vrf 获取A0账户失败", err, "signAccount", vrfSignAccount.Hex())
		return err
	}

	if accountA0 != header.Leader {
		log.Error(Module, "vrf验证", "与leader不匹配", "accountA0", accountA0.Hex(), "leader", header.Leader.Hex())
		return errors.New("公钥与leader账户不匹配")
	}

	return nil

}

func (p *ReElection) verifyAllNetTopology(header *types.Header) error {
	info, err := p.GetNetTopologyAll(header.ParentHash)
	if err != nil {
		return err
	}

	netTopology := p.TransferToNetTopologyAllStu(info)
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

func (p *ReElection) verifyChgNetTopology(version string, header *types.Header, onlineConsensusResults []*mc.HD_OnlineConsensusVoteResultMsg) error {
	if len(header.NetTopology.NetTopologyData) == 0 {
		return nil
	}
	if manversion.VersionCmp(version, manversion.VersionGamma) >= 0 {
		for _, item := range onlineConsensusResults {
			if item.Req.Node == header.Leader {
				log.Warn(Module, "verifyChgNetTopology", "leader出块共识中存在自己的下线共识", "leader", header.Leader.Hex())
				return errors.New("节点下线共识中出现出块leader自己")
			}
		}
	}
	// get online and offline info from header and prev topology
	offlineNodes, onlineNods := p.parseOnlineState(header)
	//log.Info(Module, "header.NetTop", header.NetTopology, "高度", header.Number.Uint64())
	//log.Info(Module, "offlineNodes", offlineNodes)
	//log.Info(Module, "onlineNods", onlineNods)

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
	//log.Info(Module, "p.number", header.Number.Uint64(), "offlineNodes", offlineNodes, "onlineNods", onlineNods)
	alterInfo, err := p.GetTopoChange(header.ParentHash, offlineNodes, onlineNods)
	//log.Info(Module, "alterInfo", alterInfo, "err", err)
	if err != nil {
		return err
	}

	for _, value := range alterInfo {
		log.Trace(Module, "alter-A", value.A, "position", value.Position)
	}
	// generate self net topology
	netTopology := p.TransferToNetTopologyChgStu(alterInfo)
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

func (p *ReElection) checkConsensusResult(node common.Address, state mc.OnlineState, header *types.Header, resultList []*mc.HD_OnlineConsensusVoteResultMsg) error {
	conResult := findResultInList(node, resultList)
	if conResult == nil {
		log.Error(Module, "检查拓扑变化消息", "online共识结果未找到", "node", node.Hex())
		return errors.New("online共识结果未找到")
	}
	if conResult.Req.OnlineState != state {
		log.Error(Module, "检查拓扑变化消息", "共识结果的状态不匹配", "node", node.Hex(), "结果状态", conResult.Req.OnlineState.String(), "头中状态", state.String())
		return errors.New("online共识结果未找到")
	}
	if conResult.IsValidity(header.Number.Uint64(), manparams.OnlineConsensusValidityTime) == false {
		log.Error(Module, "检查拓扑变化消息", "online共识结果过期", "result.Number", conResult.Req.Number, "curNumber", header.Number.Uint64(), "node", node.Hex())
		return errors.New("online共识结果过期")
	}
	reqHash := types.RlpHash(conResult.Req)
	blockHash, err := p.bc.GetAncestorHash(header.ParentHash, conResult.Req.Number-1)
	if err != nil {
		log.Error(Module, "检查拓扑变化消息", "online共识结果验证区块hash获取失败", "err", err, "node", node.Hex())
		return errors.New("online共识结果验证区块hash获取失败")
	}
	_, err = p.bc.DPOSEngine(header.Version).VerifyHashWithBlock(p.bc, reqHash, conResult.SignList, blockHash)
	if err != nil {
		log.Error(Module, "检查拓扑变化消息", "online共识结果POS失败", "node", node.Hex(), "err", err)
		return err
	}
	return nil
}

func (p *ReElection) parseOnlineState(header *types.Header) ([]common.Address, []common.Address) {
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
