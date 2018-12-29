// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package blkgenor

import (
	"math/big"
	"time"

	"encoding/json"
	"github.com/matrix/go-matrix/baseinterface"
	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/matrixwork"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/params/manparams"
	"github.com/matrix/go-matrix/txpoolCache"
	"github.com/pkg/errors"
)

func (p *Process) processHeaderGen() error {
	log.INFO(p.logExtraInfo(), "processHeaderGen", "start")
	defer log.INFO(p.logExtraInfo(), "processHeaderGen", "end")

	if p.bcInterval == nil {
		log.ERROR(p.logExtraInfo(), "区块生成阶段", "广播周期信息为空")
		return errors.New("广播周期信息为空")
	}

	tstart := time.Now()
	log.Info(p.logExtraInfo(), "关键时间点", "区块头开始生成", "time", tstart, "块高", p.number)
	parent, err := p.getParentBlock()
	if err != nil {
		return err
	}
	parentHash := parent.Hash()

	NetTopology, onlineConsensusResults := p.getNetTopology(p.number, parentHash, p.bcInterval)
	if nil == NetTopology {
		log.Error(p.logExtraInfo(), "获取网络拓扑图错误 ", "")
		NetTopology = &common.NetTopology{common.NetTopoTypeChange, nil}
	}
	if nil == onlineConsensusResults {
		onlineConsensusResults = make([]*mc.HD_OnlineConsensusVoteResultMsg, 0)
	}

	log.Debug(p.logExtraInfo(), "获取拓扑结果 ", NetTopology, "高度", p.number)

	tstamp := tstart.Unix()
	if parent.Time().Cmp(new(big.Int).SetInt64(tstamp)) >= 0 {
		tstamp = parent.Time().Int64() + 1
	}
	// this will ensure we're not going off too far in the future
	if now := time.Now().Unix(); tstamp > now+1 {
		wait := time.Duration(tstamp-now) * time.Second
		log.Info(p.logExtraInfo(), "等待时间同步", common.PrettyDuration(wait))
		time.Sleep(wait)
	}
	account, vrfValue, vrfProof, err := p.getVrfValue(parent)
	if err != nil {
		log.Error(p.logExtraInfo(), "区块生成阶段 获取vrfValue失败 错误", err)
		return err
	}
	header := &types.Header{
		ParentHash:        parentHash,
		Leader:            ca.GetAddress(),
		Number:            new(big.Int).SetUint64(p.number),
		GasLimit:          core.CalcGasLimit(parent),
		Extra:             make([]byte, 0),
		Time:              big.NewInt(tstamp),
		NetTopology:       *NetTopology,
		Signatures:        make([]common.Signature, 0),
		Version:           parent.Header().Version, //param
		VersionSignatures: parent.Header().VersionSignatures,
		VrfValue:          baseinterface.NewVrf().GetHeaderVrf(account, vrfValue, vrfProof),
	}

	if err := p.engine().Prepare(p.blockChain(), header); err != nil {
		log.ERROR(p.logExtraInfo(), "Failed to prepare header for mining", err)
		return err
	}

	log.Info(p.logExtraInfo(), "关键时间点", "开始执行交易", "time", time.Now(), "块高", p.number)
	tsBlock, txsCode, stateDB, receipts, originalTxs, err := p.genHeaderTxs(header)
	if err != nil {
		log.Error(p.logExtraInfo(), "运行交易失败", err)
		return err
	}

	log.Info(p.logExtraInfo(), "关键时间点", "开始执行MatrixState", "time", time.Now(), "块高", p.number)
	err = p.blockChain().ProcessMatrixState(tsBlock, stateDB)
	if err != nil {
		log.Error(p.logExtraInfo(), "运行matrix状态树失败", err)
		return err
	}

	// 运行完状态树后，才能获取elect
	Elect := p.genElection(stateDB)
	if Elect == nil {
		return errors.New("生成elect信息错误")
	}
	log.Debug(p.logExtraInfo(), "获取选举结果 ", Elect, "高度", p.number)
	header = tsBlock.Header()
	header.Elect = Elect
	//运行完matrix状态树后，生成root
	finalTxs := tsBlock.Transactions()
	block, err := p.engine().Finalize(p.blockChain(), header, stateDB, finalTxs, nil, receipts)
	if err != nil {
		log.Error(p.logExtraInfo(), "最终finalize错误", err)
		return err
	}

	log.Info(p.logExtraInfo(), "关键时间点", "区块头生成完毕,发出共识请求", "time", time.Now(), "块高", p.number)
	if p.bcInterval.IsBroadcastNumber(block.NumberU64()) {
		header = block.Header()
		signHash := header.HashNoSignsAndNonce()
		sign, err := p.signHelper().SignHashWithValidate(signHash.Bytes(), true, p.preBlockHash)
		if err != nil {
			log.ERROR(p.logExtraInfo(), "广播区块生成，签名错误", err)
			return err
		}

		header.Signatures = make([]common.Signature, 0, 1)
		header.Signatures = append(header.Signatures, sign)
		sendMsg := &mc.BlockData{Header: header, Txs: finalTxs}
		log.INFO(p.logExtraInfo(), "广播挖矿请求(本地), number", sendMsg.Header.Number, "root", header.Root.TerminalString(), "tx数量", sendMsg.Txs.Len())
		mc.PublishEvent(mc.HD_BroadcastMiningReq, &mc.BlockGenor_BroadcastMiningReqMsg{sendMsg})
	} else {
		header = block.Header()
		p2pBlock := &mc.HD_BlkConsensusReqMsg{
			Header:                 header,
			TxsCode:                txsCode,
			ConsensusTurn:          p.consensusTurn,
			OnlineConsensusResults: onlineConsensusResults,
			From: ca.GetAddress()}
		//send to local block verify module
		localBlock := &mc.LocalBlockVerifyConsensusReq{BlkVerifyConsensusReq: p2pBlock, OriginalTxs: originalTxs, FinalTxs: finalTxs, Receipts: receipts, State: stateDB}
		if len(originalTxs) > 0 {
			txpoolCache.MakeStruck(originalTxs, header.HashNoSignsAndNonce(), p.number)
		}

		log.INFO(p.logExtraInfo(), "本地发送区块验证请求, root", p2pBlock.Header.Root.TerminalString(), "高度", p.number)
		mc.PublishEvent(mc.BlockGenor_HeaderVerifyReq, localBlock)
		p.startConsensusReqSender(p2pBlock)
	}

	return nil
}

func (p *Process) genHeaderTxs(header *types.Header) (*types.Block, []*common.RetCallTxN, *state.StateDB, []*types.Receipt, []types.SelfTransaction, error) {
	//broadcast txs deal,remove no validators txs
	if p.bcInterval.IsBroadcastNumber(header.Number.Uint64()) {
		work, err := matrixwork.NewWork(p.blockChain().Config(), p.blockChain(), nil, header, p.pm.random)
		if err != nil {
			log.ERROR(p.logExtraInfo(), "NewWork!", err, "高度", p.number)
			return nil, nil, nil, nil, nil, err
		}
		mapTxs := p.pm.matrix.TxPool().GetAllSpecialTxs()

		Txs := make([]types.SelfTransaction, 0)
		for _, txs := range mapTxs {
			for _, tx := range txs {
				log.Trace(p.logExtraInfo(), "交易数据", tx)
			}
			Txs = append(Txs, txs...)
		}
		// todo: add rewward and run
		//rewardList:=p.calcRewardAndSlash(work.State, header)

		work.ProcessBroadcastTransactions(p.pm.matrix.EventMux(), Txs, p.pm.bc)
		//work.ProcessBroadcastTransactions(p.pm.matrix.EventMux(), Txs, p.pm.bc)
		retTxs := work.GetTxs()
		block := types.NewBlock(header, retTxs, nil, work.Receipts)
		return block, nil, work.State, work.Receipts, retTxs, nil

	} else {
		work, err := matrixwork.NewWork(p.blockChain().Config(), p.blockChain(), nil, header, p.pm.random)
		if err != nil {
			log.ERROR(p.logExtraInfo(), "NewWork!", err, "高度", p.number)
			return nil, nil, nil, nil, nil, err
		}

		upTimeMap, err := p.blockChain().ProcessUpTime(work.State, header)
		if err != nil {
			log.ERROR(p.logExtraInfo(), "执行uptime错误", err, "高度", p.number)
			return nil, nil, nil, nil, nil, err
		}
		txsCode, originalTxs, finalTxs := work.ProcessTransactions(p.pm.matrix.EventMux(), p.pm.txPool, p.blockChain(), upTimeMap)
		//txsCode, Txs := work.ProcessTransactions(p.pm.matrix.EventMux(), p.pm.txPool, p.blockChain(),nil,nil)
		block := types.NewBlock(header, finalTxs, nil, work.Receipts)
		log.Debug(p.logExtraInfo(), "区块验证请求生成，交易部分,完成 tx hash", block.TxHash())
		return block, txsCode, work.State, work.Receipts, originalTxs, nil
	}
}

func (p *Process) getParentBlock() (*types.Block, error) {
	if p.number == 1 { // 第一个块直接返回创世区块作为父区块
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
	log.INFO(p.logExtraInfo(), "!!!!网络发送区块验证请求, hash", req.Header.HashNoSignsAndNonce(), "tx数量", len(req.TxsCode), "次数", times)
	p.pm.hd.SendNodeMsg(mc.HD_BlkConsensusReq, req, common.RoleValidator, nil)
}

func (p *Process) getVrfValue(parent *types.Block) ([]byte, []byte, []byte, error) {
	_, preVrfValue, preVrfProof := baseinterface.NewVrf().GetVrfInfoFromHeader(parent.Header().VrfValue)
	parentMsg := VrfMsg{
		VrfProof: preVrfProof,
		VrfValue: preVrfValue,
		Hash:     parent.Hash(),
	}
	vrfmsg, err := json.Marshal(parentMsg)
	if err != nil {
		log.Error(p.logExtraInfo(), "生成vrfmsg出错", err, "parentMsg", parentMsg)
		return []byte{}, []byte{}, []byte{}, errors.New("生成vrfmsg出错")
	}
	return p.signHelper().SignVrf(vrfmsg, p.preBlockHash)
}
