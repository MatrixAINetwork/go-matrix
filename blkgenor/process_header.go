// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package blkgenor

import (
	"github.com/MatrixAINetwork/go-matrix/consensus/blkmanage"

	"github.com/MatrixAINetwork/go-matrix/ca"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/params/manparams"
	"github.com/MatrixAINetwork/go-matrix/txpoolCache"
	"github.com/pkg/errors"
)

func (p *Process) processBcHeaderGen() error {
	log.INFO(p.logExtraInfo(), "processBCHeaderGen", "start")
	defer log.INFO(p.logExtraInfo(), "processBCHeaderGen", "end")
	if p.bcInterval == nil {
		log.ERROR(p.logExtraInfo(), "区块生成阶段", "广播周期信息为空")
		return errors.New("广播周期信息为空")
	}
	parent, err := p.getParentBlock(p.number)
	if err != nil {
		return err
	}
	version := p.pm.manblk.ProduceBlockVersion(p.number, string(parent.Version()))

	originHeader, _, err := p.pm.manblk.Prepare(blkmanage.BroadcastBlk, version, p.number, p.bcInterval, p.preBlockHash)
	if err != nil {
		log.Error(p.logExtraInfo(), "准备去看失败", err)
		return err
	}

	_, stateDB, receipts, _, finalTxs, _, err := p.pm.manblk.ProcessState(blkmanage.BroadcastBlk, version, originHeader, nil)
	if err != nil {
		log.Error(p.logExtraInfo(), "运行交易和状态树失败", err)
		return err
	}

	//运行完matrix状态树后，生成root
	block, _, err := p.pm.manblk.Finalize(blkmanage.BroadcastBlk, version, originHeader, stateDB, finalTxs, nil, receipts, nil)
	if err != nil {
		log.Error(p.logExtraInfo(), "Finalize失败", err)
		return err
	}
	finalHeader := block.Header()
	err = p.setSignatures(finalHeader)
	if err != nil {
		return err
	}
	p.sendBroadcastMiningReq(finalHeader, finalTxs)
	return nil
}

func (p *Process) processHeaderGen() error {
	log.INFO(p.logExtraInfo(), "processHeaderGen", "start")
	defer log.INFO(p.logExtraInfo(), "processHeaderGen", "end")
	if p.bcInterval == nil {
		log.ERROR(p.logExtraInfo(), "区块生成阶段", "广播周期信息为空")
		return errors.New("广播周期信息为空")
	}
	parent, err := p.getParentBlock(p.number)
	if err != nil {
		return err
	}
	version := p.pm.manblk.ProduceBlockVersion(p.number, string(parent.Version()))

	originHeader, extraData, err := p.pm.manblk.Prepare(blkmanage.CommonBlk, version, p.number, p.bcInterval, p.preBlockHash)
	if err != nil {
		log.Error(p.logExtraInfo(), "准备阶段失败", err)
		return err
	}
	onlineConsensusResults, ok := extraData.([]*mc.HD_OnlineConsensusVoteResultMsg)
	if !ok {
		log.Error(p.logExtraInfo(), "反射在线状态失败", "")
		return errors.New("反射在线状态失败")
	}

	txsCode, stateDB, receipts, originalTxs, finalTxs, _, err := p.pm.manblk.ProcessState(blkmanage.CommonBlk, version, originHeader, nil)
	if err != nil {
		log.Error(p.logExtraInfo(), "运行交易和状态树失败", err)
		return err
	}
	//运行完matrix状态树后，生成root (p.blockChain(), header, stateDB, nil, tsBlock.Currencies())
	block, _, err := p.pm.manblk.Finalize(blkmanage.CommonBlk, version, originHeader, stateDB, finalTxs, nil, receipts, nil)
	if err != nil {
		log.Error(p.logExtraInfo(), "Finalize失败", err)
		return err
	}
	p.sendHeaderVerifyReq(block.Header(), txsCode, onlineConsensusResults, originalTxs, finalTxs, receipts, stateDB)
	return nil
}

func (p *Process) sendHeaderVerifyReq(header *types.Header, txsCode []*common.RetCallTxN, onlineConsensusResults []*mc.HD_OnlineConsensusVoteResultMsg, originalTxs []types.CoinSelfTransaction,
	finalTxs []types.CoinSelfTransaction, receipts []types.CoinReceipts, stateDB *state.StateDBManage) {
	p2pBlock := &mc.HD_BlkConsensusReqMsg{
		Header:                 header,
		TxsCode:                txsCode,
		ConsensusTurn:          p.consensusTurn,
		OnlineConsensusResults: onlineConsensusResults,
		From:                   ca.GetSignAddress(),
	}
	//send to local block verify module
	localBlock := &mc.LocalBlockVerifyConsensusReq{BlkVerifyConsensusReq: p2pBlock, OriginalTxs: originalTxs, FinalTxs: finalTxs, Receipts: receipts, State: stateDB}
	if len(originalTxs) > 0 {
		txpoolCache.MakeStruck(types.GetTX(originalTxs), header.HashNoSignsAndNonce(), p.number)
	}
	log.INFO(p.logExtraInfo(), "本地发送区块验证请求, root", p2pBlock.Header.Roots, "高度", p.number)
	mc.PublishEvent(mc.BlockGenor_HeaderVerifyReq, localBlock)
	p.startConsensusReqSender(p2pBlock)
}

func (p *Process) sendBroadcastMiningReq(header *types.Header, finalTxs []types.CoinSelfTransaction) {
	sendMsg := &mc.BlockData{Header: header, Txs: finalTxs}
	log.INFO(p.logExtraInfo(), "广播挖矿请求(本地), number", sendMsg.Header.Number, "root", header.Roots, "tx数量", len(types.GetTX(finalTxs)))
	mc.PublishEvent(mc.HD_BroadcastMiningReq, &mc.BlockGenor_BroadcastMiningReqMsg{sendMsg})
}

func (p *Process) setSignatures(header *types.Header) error {

	signHash := header.HashNoSignsAndNonce()
	sign, err := p.signHelper().SignHashWithValidateByAccount(signHash.Bytes(), true, ca.GetDepositAddress())
	if err != nil {
		log.ERROR(p.logExtraInfo(), "广播区块生成，签名错误", err)
		return err
	}

	header.Signatures = make([]common.Signature, 0, 1)
	header.Signatures = append(header.Signatures, sign)

	return nil
}

func (p *Process) startConsensusReqSender(req *mc.HD_BlkConsensusReqMsg) {
	p.closeConsensusReqSender()
	sender, err := common.NewResendMsgCtrl(req, p.sendConsensusReqFunc, manparams.BlkPosReqSendInterval, manparams.BlkPosReqSendTimes)
	if err != nil {
		log.ERROR(p.logExtraInfo(), "创建POS完成的req发送器", "失败", "err", err)
		return
	}
	p.consensusReqSender = sender
}

func (p *Process) closeConsensusReqSender() {
	if p.consensusReqSender == nil {
		return
	}
	p.consensusReqSender.Close()
	p.consensusReqSender = nil
}

func (p *Process) sendConsensusReqFunc(data interface{}, times uint32) {
	req, OK := data.(*mc.HD_BlkConsensusReqMsg)
	if !OK {
		log.ERROR(p.logExtraInfo(), "发出区块共识req", "反射消息失败", "次数", times)
		return
	}
	log.INFO(p.logExtraInfo(), "!!!!网络发送区块验证请求, hash", req.Header.HashNoSignsAndNonce(), "tx数量", req.TxsCodeCount(), "次数", times)
	p.pm.hd.SendNodeMsg(mc.HD_BlkConsensusReq, req, common.RoleValidator, nil)
}

func (p *Process) getParentBlock(num uint64) (*types.Block, error) {
	if num == 1 { // 第一个块直接返回创世区块作为父区块
		return p.blockChain().Genesis(), nil
	}

	if (p.preBlockHash == common.Hash{}) {
		return nil, errors.Errorf("未知父区块hash[%s]", p.preBlockHash.TerminalString())
	}

	parent := p.blockChain().GetBlockByHash(p.preBlockHash)
	if nil == parent {
		return nil, errors.Errorf("未知的父区块[%s]", p.preBlockHash.TerminalString())
	}

	return parent, nil
}
