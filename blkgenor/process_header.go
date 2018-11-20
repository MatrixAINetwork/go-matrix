// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package blkgenor

import (
	"math/big"
	"time"

	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/matrixwork"
	"github.com/matrix/go-matrix/mc"
	"github.com/pkg/errors"
	"github.com/matrix/go-matrix/txpoolCache"
	"github.com/matrix/go-matrix/params"
)

func (p *Process) processHeaderGen() error {
	log.INFO(p.logExtraInfo(), "processHeaderGen", "start")
	defer log.INFO(p.logExtraInfo(), "processHeaderGen", "end")

	tstart := time.Now()
	parent, err := p.getParentBlock()
	if err != nil {
		return err
	}

	tstamp := tstart.Unix()
	NetTopology := p.getNetTopology(parent.Header().NetTopology, p.number)
	if nil == NetTopology {
		log.Error(p.logExtraInfo(), "获取网络拓扑图错误 ", "")
		NetTopology = &common.NetTopology{common.NetTopoTypeChange, nil}
	}

	Elect := p.genElection(p.number)

	log.Info(p.logExtraInfo(), "++++++++获取选举结果 ", Elect, "高度", p.number)
	log.Info(p.logExtraInfo(), "++++++++获取拓扑结果 ", NetTopology, "高度", p.number)
	if parent.Time().Cmp(new(big.Int).SetInt64(tstamp)) >= 0 {
		tstamp = parent.Time().Int64() + 1
	}
	// this will ensure we're not going off too far in the future
	if now := time.Now().Unix(); tstamp > now+1 {
		wait := time.Duration(tstamp-now) * time.Second
		log.Info("Mining too far in the future", "wait", common.PrettyDuration(wait))
		time.Sleep(wait)
	}
	header := &types.Header{
		ParentHash:  parent.Hash(),
		Leader:      ca.GetAddress(),
		Number:      new(big.Int).SetUint64(p.number),
		GasLimit:    core.CalcGasLimit(parent),
		Extra:       make([]byte, 0),
		Time:        big.NewInt(tstamp),
		Elect:       Elect,
		NetTopology: *NetTopology,
		Signatures:  make([]common.Signature, 0),
		Version:     parent.Header().Version, //param
	}
	if err := p.engine().Prepare(p.blockChain(), header); err != nil {
		log.ERROR(p.logExtraInfo(), "Failed to prepare header for mining", err)
		return err
	}
	//broadcast txs deal,remove no validators txs
	if common.IsBroadcastNumber(header.Number.Uint64()) {
		work, err := matrixwork.NewWork(p.blockChain().Config(), p.blockChain(), nil, header)
		if err != nil {
			log.ERROR(p.logExtraInfo(), "NewWork!", err, "高度", p.number)
			return err
		}
		mapTxs := p.pm.matrix.TxPool().GetAllSpecialTxs()

		Txs := make([]types.SelfTransaction, 0)
		for _, txs := range mapTxs {
			for _, tx := range txs {
				log.INFO(p.logExtraInfo(), "交易数据 t", tx)
			}
			Txs = append(Txs, txs...)
		}
		work.ProcessBroadcastTransactions(p.pm.matrix.EventMux(), Txs, p.pm.bc)

		for _, tx := range Txs {
			log.INFO("==========", "Finalize:GasPrice", tx.GasPrice(), "amount", tx.Value())
		}

		//validators, _ := self.ca.GetPreValidatorsAddress()
		//for validator := range pending {
		//	for i, v := range validators {
		//		if validator.String() == v.String() {
		//			continue
		//		}
		//		if i == len(validators)-1 {
		//			delete(pending, validator)
		//		}
		//	}
		//}
		//send to local block mining module
		block, err := p.engine().Finalize(p.blockChain(), header, work.State, Txs, nil, work.Receipts)
		if err != nil {
			log.ERROR(p.logExtraInfo(), "Failed to finalize block for sealing", err)
			return err
		}
		header = block.Header()
		signHash := header.HashNoSignsAndNonce()
		sign, err := p.signHelper().SignHashWithValidate(signHash.Bytes(), true)
		if err != nil {
			log.ERROR(p.logExtraInfo(), "广播区块生成，签名错误", err)
			return err
		}

		header.Signatures = make([]common.Signature, 0, 1)
		header.Signatures = append(header.Signatures, sign)
		sendMsg := &mc.BlockData{Header: header, Txs: Txs}
		log.INFO(p.logExtraInfo(), "广播挖矿请求(本地), number", sendMsg.Header.Number, "root", header.Root.TerminalString(), "tx数量", sendMsg.Txs.Len())
		mc.PublishEvent(mc.HD_BroadcastMiningReq, &mc.BlockGenor_BroadcastMiningReqMsg{sendMsg})

	} else {
		log.INFO(p.logExtraInfo(), "区块验证请求生成，交易部分", "开始创建work")
		work, err := matrixwork.NewWork(p.blockChain().Config(), p.blockChain(), nil, header)
		if err != nil {
			log.ERROR(p.logExtraInfo(), "NewWork!", err, "高度", p.number)
			return err
		}

		//work.commitTransactions(self.mux, Txs, self.chain)
		// todo： update uptime
		/*if common.IsBroadcastNumber(p.number-1) && p.number > common.GetBroadcastInterval() {
			upTimeAccounts, err := work.GetUpTimeAccounts(p.number)
			if err != nil {
				log.ERROR(p.logExtraInfo(), "获取所有抵押账户错误!", err, "高度", p.number)
				return err
			}
			calltherollMap, heatBeatUnmarshallMMap, err := work.GetUpTimeData(p.number)
			if err != nil {
				log.ERROR(p.logExtraInfo(), "获取心跳交易错误!", err, "高度", p.number)
			}

			err = work.HandleUpTime(work.State, upTimeAccounts, calltherollMap, heatBeatUnmarshallMMap, p.number, p.blockChain())
			if nil != err {
				log.ERROR(p.logExtraInfo(), "处理uptime错误", err)
				return err
			}
		}*/
		log.INFO(p.logExtraInfo(), "区块验证请求生成，交易部分", "完成创建work, 开始执行交易")
		txsCode, Txs := work.ProcessTransactions(p.pm.matrix.EventMux(), p.pm.txPool, p.pm.bc)
		log.INFO("=========", "ProcessTransactions finish", len(txsCode))
		log.INFO(p.logExtraInfo(), "区块验证请求生成，交易部分", "完成执行交易, 开始finalize")
		block, err := p.engine().Finalize(p.blockChain(), header, work.State, Txs, nil, work.Receipts)
		if err != nil {
			log.ERROR(p.logExtraInfo(), "Failed to finalize block for sealing", err)
			return err
		}
		log.INFO(p.logExtraInfo(), "区块验证请求生成，交易部分", "完成finalize")
		header = block.Header()
		p2pBlock := &mc.HD_BlkConsensusReqMsg{Header: header, TxsCode: txsCode, ConsensusTurn: p.consensusTurn, From: ca.GetAddress()}
		//send to local block verify module
		localBlock := &mc.LocalBlockVerifyConsensusReq{BlkVerifyConsensusReq: p2pBlock, Txs: Txs, Receipts: work.Receipts, State: work.State}
		if len(Txs) > 0{
			txpoolCache.MakeStruck(Txs,header.HashNoSignsAndNonce(),p.number)
		}
		log.INFO(p.logExtraInfo(), "!!!!本地发送区块验证请求, root", p2pBlock.Header.Root.TerminalString(), "高度", p.number)
		mc.PublishEvent(mc.BlockGenor_HeaderVerifyReq, localBlock)
		p.startConsensusReqSender(p2pBlock)
	}

	return nil
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
	sender, err := common.NewResendMsgCtrl(req, p.sendConsensusReqFunc, params.BlkPosReqSendInterval, params.BlkPosReqSendTimes)
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
