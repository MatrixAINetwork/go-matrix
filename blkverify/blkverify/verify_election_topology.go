// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package blkverify

import (
	"crypto/ecdsa"
	"encoding/json"
	"github.com/btcsuite/btcd/btcec"
	"github.com/matrix/go-matrix/baseinterface"
	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/crypto"
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

func (p *Process) verifyNetTopology(header *types.Header, onlineConsensusResults []*mc.HD_OnlineConsensusVoteResultMsg) error {
	if header.NetTopology.Type == common.NetTopoTypeAll {
		return p.verifyAllNetTopology(header)
	}

	return p.verifyChgNetTopology(header, onlineConsensusResults)
}
func (p *Process) verifyVrf(header *types.Header) error {
	log.Error("vrf", "len header.VrfValue", len(header.VrfValue), "data", header.VrfValue, "高度", header.Number.Uint64())
	account, _, _ := common.GetVrfInfoFromHeader(header.VrfValue)

	log.Error("vrf", "从区块头重算出账户户", account, "高度", header.Number.Uint64())

	public := account
	curve := btcec.S256()
	pk1, err := btcec.ParsePubKey(public, curve)
	if err != nil {
		log.Error("vrf转换失败", "err", err, "account", account, "len", len(account))
		return err
	}

	pk1_1 := (*ecdsa.PublicKey)(pk1)
	_, vrfValue, vrfProof := common.GetVrfInfoFromHeader(header.VrfValue)

	preBlock := p.blockChain().GetBlockByHash(header.ParentHash)
	if preBlock == nil {
		return errors.New("获取父区块失败")
	}
	_, preVrfValue, preVrfProof := common.GetVrfInfoFromHeader(preBlock.Header().VrfValue)

	preMsg := common.VrfMsg{
		VrfValue: preVrfValue,
		VrfProof: preVrfProof,
		Hash:     header.ParentHash,
	}

	preVrfMsg, err := json.Marshal(preMsg)
	if err != nil {
		log.Error(p.logExtraInfo(), "生成vefmsg出错", err, "parentMsg", preVrfMsg)
		return errors.New("生成vrfmsg出错")
	} else {
		log.Error("生成vrfmsg成功")
	}
	//log.Info("msgggggvrf_verify","preVrfMsg",preVrfMsg,"高度",header.Number.Uint64(),"VrfProof",preMsg.VrfProof,"VrfValue",preMsg.VrfValue,"Hash",preMsg.Hash)
	if err := baseinterface.NewVrf().VerifyVrf(pk1_1, preVrfMsg, vrfValue, vrfProof); err != nil {
		log.Error("vrf verify ", "err", err)
		return err
	}

	ans := crypto.PubkeyToAddress(*pk1_1)
	if ans.Equal(header.Leader) {
		log.Error("vrf leader comparre", "与leader不匹配", "nil")
		return nil
	}
	return errors.New("公钥与leader账户不匹配")
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

	// get prev topology
	prevTopology, err := ca.GetTopologyByHash(common.RoleValidator|common.RoleBackupValidator, header.ParentHash)
	if err != nil {
		return err
	}

	// get online and offline info from header and prev topology
	offlineTopNodes, onlineElectNods, offlineElectNodes := p.parseOnlineState(header, prevTopology)
	log.INFO("scfffff-verify", "header.NetTop", header.NetTopology, "高度", header.Number.Uint64())
	log.INFO("scfffff--verify", "prevTopology", prevTopology)
	log.INFO("scfffff--verify", "offlineTopNodes", offlineTopNodes)
	log.INFO("scfffff--verify", "onlinePrimaryNods", onlineElectNods)
	log.INFO("scfffff--verify", "offlinePrimaryNodes", offlineElectNodes)

	for _, node := range offlineTopNodes {
		if err := p.checkConsensusResult(node, header, onlineConsensusResults); err != nil {
			return err
		}
	}
	for _, node := range onlineElectNods {
		if err := p.checkConsensusResult(node, header, onlineConsensusResults); err != nil {
			return err
		}
		electInfo := prevTopology.GetAccountElectInfo(node)
		if electInfo == nil {
			return errors.Errorf("节点(%s)不是elect节点", node.Hex())
		}
		if electInfo.Position != common.PosOffline {
			return errors.Errorf("节点(%s)header中共识在线，但原链上状态不是离线")
		}
	}
	for _, node := range offlineElectNodes {
		if err := p.checkConsensusResult(node, header, onlineConsensusResults); err != nil {
			return err
		}
		electInfo := prevTopology.GetAccountElectInfo(node)
		if electInfo == nil {
			return errors.Errorf("节点(%s)不是elect节点", node.Hex())
		}
		if electInfo.Position != common.PosOnline {
			return errors.Errorf("节点(%s)header中共识离线，但原链上状态不是在线")
		}
	}

	// generate topology alter info
	log.INFO("scffffff---Verify---GetTopoChange start ", "p.number", p.number, "offlineTopNodes", offlineTopNodes, "onlineElectNods", onlineElectNods)
	alterInfo, err := p.reElection().GetTopoChange(header.ParentHash, offlineTopNodes)
	log.INFO("scffffff---Verify---GetTopoChange end", "alterInfo", alterInfo, "err", err)
	if err != nil {
		return err
	}

	for _, value := range alterInfo {
		log.Info(p.logExtraInfo(), "alter-A", value.A, "position", value.Position)
	}
	// generate self net topology
	netTopology := p.reElection().TransferToNetTopologyChgStu(alterInfo, onlineElectNods, offlineElectNodes)
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

func (p *Process) checkConsensusResult(node common.Address, header *types.Header, resultList []*mc.HD_OnlineConsensusVoteResultMsg) error {
	conResult := findResultInList(node, resultList)
	if conResult == nil {
		log.Error(p.logExtraInfo(), "检查拓扑变化消息", "online共识结果未找到", "node", node.Hex())
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

func (p *Process) parseOnlineState(header *types.Header, prevTopology *mc.TopologyGraph) ([]common.Address, []common.Address, []common.Address) {
	offlineTopNodes := p.reElection().ParseTopNodeOffline(header.NetTopology, prevTopology)
	onlineElectNods, offlineElectNodes := p.reElection().ParseElectTopNodeState(header.NetTopology)
	return offlineTopNodes, onlineElectNods, offlineElectNodes
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
