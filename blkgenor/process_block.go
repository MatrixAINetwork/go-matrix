// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package blkgenor

import (
	"math/big"

	"time"

	"github.com/MatrixAINetwork/go-matrix/ca"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/consensus/blkmanage"
	"github.com/MatrixAINetwork/go-matrix/core"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/params/manparams"
)

func (p *Process) ProcessRecoveryMsg(msg *mc.RecoveryStateMsg) {
	log.INFO(p.logExtraInfo(), "状态恢复消息处理", "开始", "类型", msg.Type, "高度", p.number, "leader", msg.Header.Leader.Hex())
	defer log.Debug(p.logExtraInfo(), "状态恢复消息处理", "结束", "类型", msg.Type, "高度", p.number, "leader", msg.Header.Leader.Hex())

	p.mu.Lock()
	defer p.mu.Unlock()

	header := msg.Header
	headerHash := header.HashNoSignsAndNonce()
	if msg.IsBroadcast {
		log.Debug(p.logExtraInfo(), "区块为广播区块", "直接发送全区块获取")
		p.sendFullBlockReq(headerHash, header.Number.Uint64(), msg.From)
		return
	}

	minerResult := &mc.HD_MiningRspMsg{
		From:       header.Coinbase,
		Number:     header.Number.Uint64(),
		BlockHash:  headerHash,
		Difficulty: header.Difficulty,
		Nonce:      header.Nonce,
		Coinbase:   header.Coinbase,
		MixDigest:  header.MixDigest,
		Signatures: header.Signatures,
	}
	log.Debug(p.logExtraInfo(), "状态恢复消息处理", "开始补全挖矿结果消息")
	if err := p.powPool.AddMinerResult(minerResult.BlockHash, minerResult.Difficulty, minerResult); err != nil {
		log.WARN(p.logExtraInfo(), "状态恢复消息处理", "挖矿结果入池失败", "err", err, "高度", p.number)
	}
	p.processMinerResultVerify(header.Leader, false)

	if p.state != StateEnd {
		//处理完成后，状态不是完成状态，说明缺少数据
		log.Debug(p.logExtraInfo(), "状态恢复消息处理", "处理完毕后，本地状态不是end", "本地状态", p.state, "hash", headerHash.TerminalString())
		p.sendFullBlockReq(headerHash, header.Number.Uint64(), msg.From)
	}
}

func (p *Process) sendFullBlockReq(hash common.Hash, number uint64, target common.Address) {
	if p.FullBlockReqCache.IsExistMsg(hash) {
		data, err := p.FullBlockReqCache.ReUseMsg(hash)
		if err != nil {
			return
		}
		reqMsg, _ := data.(*mc.HD_FullBlockReqMsg)
		log.Debug(p.logExtraInfo(), "状态恢复消息处理", "发送完整区块获取请求消息", "to", target.Hex(), "高度", reqMsg.Number, "hash", reqMsg.HeaderHash.TerminalString())
		p.pm.hd.SendNodeMsg(mc.HD_FullBlockReq, reqMsg, common.RoleNil, []common.Address{target})
	} else {
		reqMsg := &mc.HD_FullBlockReqMsg{
			HeaderHash: hash,
			Number:     number,
		}
		p.FullBlockReqCache.AddMsg(hash, reqMsg, time.Now().Unix())
		log.Debug(p.logExtraInfo(), "状态恢复消息处理", "发送完整区块获取请求消息", "to", target.Hex(), "高度", reqMsg.Number, "hash", reqMsg.HeaderHash.TerminalString())
		p.pm.hd.SendNodeMsg(mc.HD_FullBlockReq, reqMsg, common.RoleNil, []common.Address{target})
	}
}

func (p *Process) ProcessFullBlockReq(req *mc.HD_FullBlockReqMsg) {
	p.mu.Lock()
	defer p.mu.Unlock()

	blockData := p.blockCache.GetBlockDataByBlockHash(req.HeaderHash)
	if nil == blockData {
		log.ERROR(p.logExtraInfo(), "处理完整区块请求", "区块信息未找到", "高度", p.number, "hash", req.HeaderHash.TerminalString())
		return
	}

	if blockData.state != blockStateReady {
		log.ERROR(p.logExtraInfo(), "处理完整区块请求", "区块未生成", "高度", p.number, "hash", req.HeaderHash.TerminalString())
		return
	}

	rspMsg := &mc.HD_FullBlockRspMsg{
		Header: blockData.block.Header,
		Txs:    blockData.block.OriginalTxs,
	}
	log.Debug(p.logExtraInfo(), "处理完整区块请求", "发送响应消息", "to", req.From, "hash", rspMsg.Header.Hash(), "交易", rspMsg.Txs)
	p.pm.hd.SendNodeMsg(mc.HD_FullBlockRsp, rspMsg, common.RoleNil, []common.Address{req.From})
}

func (p *Process) ProcessFullBlockRsp(rsp *mc.HD_FullBlockRspMsg) {
	fullHash := rsp.Header.Hash()
	headerHash := rsp.Header.HashNoSignsAndNonce()
	log.INFO(p.logExtraInfo(), "处理完整区块响应", "开始", "区块 hash", fullHash.TerminalString(), "交易", rsp.Txs, "root", rsp.Header.Roots, "高度", p.number)
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.bcInterval == nil {
		log.Error(p.logExtraInfo(), "处理完整区块响应", "广播周期为nil", "header Hash", headerHash.TerminalString(), "高度", p.number)
		return
	}

	if blockData := p.blockCache.GetBlockDataByBlockHash(headerHash); blockData != nil {
		log.Error(p.logExtraInfo(), "处理完整区块响应", "已存在的区块信息", "header Hash", headerHash.TerminalString(), "高度", p.number)
		return
	}

	isBroadcast := p.bcInterval.IsBroadcastNumber(rsp.Header.Number.Uint64())
	if err := p.pm.bc.Engine(rsp.Header.Version).VerifyHeader(p.pm.bc, rsp.Header, !isBroadcast); err != nil {
		log.Error(p.logExtraInfo(), "处理完整区块响应", "POW验证未通过", "err", err, "高度", p.number)
		return
	}

	if err := p.pm.bc.DPOSEngine(rsp.Header.Version).VerifyBlock(p.pm.bc, rsp.Header); err != nil {
		log.ERROR(p.logExtraInfo(), "处理完整区块响应", "POS验证未通过", "err", err, "高度", p.number)
		return
	}

	blkType := blkmanage.CommonBlk
	if isBroadcast {
		blkType = blkmanage.BroadcastBlk
	}
	//运行交易
	stateDB, finalTxs, receipts, _, err := p.pm.manblk.VerifyTxsAndState(blkType, string(rsp.Header.Version), rsp.Header, rsp.Txs)
	if err != nil {
		log.Error(p.logExtraInfo(), "处理完整区块响应", "执行交易错误", "err", err, "高度", p.number)
		return
	}

	p.blockCache.SaveReadyBlock(&mc.BlockLocalVerifyOK{
		Header:      rsp.Header,
		BlockHash:   rsp.Header.HashNoSignsAndNonce(),
		OriginalTxs: rsp.Txs,
		FinalTxs:    finalTxs,
		Receipts:    receipts,
		State:       stateDB,
	})

	readyMsg := &mc.NewBlockReadyMsg{
		Header: rsp.Header,
		State:  stateDB.Copy(),
	}
	mc.PublishEvent(mc.BlockGenor_NewBlockReady, readyMsg)

	p.state = StateBlockInsert
	p.processBlockInsert(rsp.Header.Leader)
}

func (p *Process) AddMinerResult(minerResult *mc.HD_MiningRspMsg) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if err := p.powPool.AddMinerResult(minerResult.BlockHash, minerResult.Difficulty, minerResult); err != nil {
		return
	}
	log.Info(p.logExtraInfo(), "关键时间点", "收到矿工挖矿结果", "time", time.Now(), "块高", p.number)
	log.INFO(p.logExtraInfo(), "矿工挖矿结果消息处理", "开始", "高度", minerResult.Number, "难度", minerResult.Difficulty.Uint64(), "block hash", minerResult.BlockHash.TerminalString(), "from", minerResult.From.Hex())
	p.processMinerResultVerify(p.curLeader, true)
}

func (p *Process) minerPickTimeout() {
	p.mu.Lock()
	log.Debug(p.logExtraInfo(), "minerPickTimeout", "开始处理", "高度", p.number)
	defer func() {
		defer log.Debug(p.logExtraInfo(), "minerPickTimeout", "结束处理", "高度", p.number)
		p.mu.Unlock()
	}()

	p.stopMinerPikerTimer()
	p.processMinerResultVerify(p.curLeader, true)
}

func (p *Process) AddConsensusBlock(block *mc.BlockLocalVerifyOK) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.blockCache.SaveVerifiedBlock(block)
	p.processMinerResultVerify(p.curLeader, true)
}

func (p *Process) processMinerResultVerify(leader common.Address, checkState bool) {
	if checkState && p.checkState(StateMinerResultVerify) == false {
		log.WARN(p.logExtraInfo(), "准备进行挖矿结果验证，状态错误", p.state.String())
		return
	}

	if p.bcInterval == nil {
		log.WARN(p.logExtraInfo(), "准备进行挖矿结果验证", "广播周期信息为nil", "高度", p.number)
		return
	}

	if p.bcInterval.IsBroadcastNumber(p.number) {
		p.dealMinerResultVerifyBroadcast()
	} else {
		p.dealMinerResultVerifyCommon(leader)
	}
}

func (p *Process) dealMinerResultVerifyCommon(leader common.Address) {
	var blockData *blockCacheData = nil
	log.INFO(p.logExtraInfo(), "当前高度为普通区块, 进行普通挖矿结果验证, 高度", p.number)
	if p.role == common.RoleBroadcast {
		blockData = p.blockCache.GetLastBlockData()
	} else {
		blockData = p.blockCache.GetBlockData(leader)
	}

	if nil == blockData {
		log.WARN(p.logExtraInfo(), "准备进行挖矿结果验证", "验证区块还未收到！等待验证区块", "高度", p.number, "身份", p.role, "leader", leader.Hex())
		return
	}

	if blockData.state == blockStateLocalVerified {
		diff := big.NewInt(blockData.block.Header.Difficulty.Int64())
		results, err := p.powPool.GetMinerResults(blockData.block.BlockHash, diff)
		if err != nil {
			log.WARN(p.logExtraInfo(), "挖矿结果验证，挖矿结果获取失败", err, "高度", p.number, "难度", diff, "block hash", blockData.block.BlockHash.TerminalString())
			return
		}
		if len(results) == 0 {
			log.WARN(p.logExtraInfo(), "进行挖矿结果验证", "当前没有挖矿结果", "高度", p.number, "block hash", blockData.block.BlockHash.TerminalString())
			return
		}

		passTime := time.Now().Unix() - blockData.block.Header.Time.Int64()
		innerMinerPick := passTime > manparams.MinerPickTimeout
		satisfyResult, err := p.pickSatisfyMinerResults(blockData.block.Header, results, innerMinerPick)
		if err != nil {
			log.WARN(p.logExtraInfo(), "挖矿结果验证，获取合适挖矿结果错误", err, "高度", p.number)
			//若未超时失败，则启动超时定时器
			if innerMinerPick == false {
				p.startMinerPikerTimer(manparams.MinerPickTimeout - passTime + 1)
			}
			return
		}
		blockData.block.Header = p.copyHeader(blockData.block.Header, satisfyResult)
		blockData.state = blockStateReady
	}
	p.stopMinerPikerTimer()
	p.closeConsensusReqSender()
	readyMsg := &mc.NewBlockReadyMsg{
		Header: blockData.block.Header,
		State:  blockData.block.State.Copy(),
	}
	log.INFO(p.logExtraInfo(), "普通区块验证完成", "发送新区块准备完毕消息", "高度", p.number, "leader", readyMsg.Header.Leader.Hex())
	mc.PublishEvent(mc.BlockGenor_NewBlockReady, readyMsg)

	p.state = StateBlockInsert
	p.processBlockInsert(p.curLeader)
}

func (p *Process) processBlockInsert(blockLeader common.Address) {
	if false == p.canGenBlock() {
		return
	}

	log.INFO(p.logExtraInfo(), "区块插入", "开始", "高度", p.number)
	hash, err := p.insertAndBcBlock(true, blockLeader, nil)
	if err != nil {
		log.ERROR(p.logExtraInfo(), "区块插入，错误", err)
		return
	}
	log.Info(p.logExtraInfo(), "关键时间点", "leader挂块成功", "time", time.Now(), "块高", p.number)
	log.Debug(p.logExtraInfo(), "区块插入", "完成", "高度", p.number, "插入区块hash", hash.TerminalString())
	p.state = StateEnd
}

func (p *Process) canGenBlock() bool {
	if p.state < StateBlockInsert {
		log.WARN(p.logExtraInfo(), "准备进行区块插入，状态错误", p.state.String(), "高度", p.number)
		return false
	}
	if p.bcInterval == nil {
		log.WARN(p.logExtraInfo(), "准备进行区块插入", "广播周期信息为nil")
		return false
	}
	if p.bcInterval.IsBroadcastNumber(p.number + 1) {
		if p.role != common.RoleBroadcast {
			log.WARN(p.logExtraInfo(), "准备进行区块插入，广播区块前一个区块，由广播节点插入", p.role.String(), "高度", p.number)
			return false
		}
	} else {
		if p.role != common.RoleValidator {
			log.WARN(p.logExtraInfo(), "准备进行区块插入，身份错误", "当前身份不是验证者", "高度", p.number, "身份", p.role.String())
			return false
		}

		if (p.nextLeader == common.Address{}) {
			log.WARN(p.logExtraInfo(), "准备进行区块插入", "下个区块leader为空", "需要等待leader的高度", p.number+1)
			return false
		}

		if p.nextLeader != ca.GetDepositAddress() {
			log.Debug(p.logExtraInfo(), "准备进行区块广播,自己不是下个区块leader,高度", p.number, "next leader", p.nextLeader.Hex(), "self", ca.GetDepositAddress().Hex())
			return false
		}
	}
	return true
}

func (p *Process) pickSatisfyMinerResults(header *types.Header, results []*mc.HD_MiningRspMsg, innerMinerPick bool) (*mc.HD_MiningRspMsg, error) {
	for _, result := range results {
		if innerMinerPick == false {
			role, _ := ca.GetAccountOriginalRole(result.Coinbase, header.ParentHash)
			if common.RoleInnerMiner == role {
				log.WARN(p.logExtraInfo(), "基金会矿工结果", "当前未超时，暂时不选用", "from", result.Coinbase.Hex(), "难度", result.Difficulty, "高度", p.number)
				continue
			}
		}
		if err := p.verifyOneResult(header, result); err != nil {
			log.WARN(p.logExtraInfo(), "验证挖矿结果失败，删除该挖矿结果, from", result.From, "难度", result.Difficulty,
				"高度", p.number, "block hash", result.BlockHash.TerminalString())
			p.powPool.DelOneResult(result.BlockHash, result.Difficulty, result.From)
			continue
		}
		log.INFO(p.logExtraInfo(), "选择挖矿结果", "完成", "矿工", result.Coinbase.Hex(), "难度", result.Difficulty, "高度", p.number)
		return result, nil
	}
	return nil, HaveNotOKResultError
}

func (p *Process) verifyOneResult(rawHeader *types.Header, result *mc.HD_MiningRspMsg) error {
	header := p.copyHeader(rawHeader, result)
	headerHash := header.HashNoSignsAndNonce()
	if headerHash != result.BlockHash {
		log.WARN(p.logExtraInfo(), "挖矿结果不匹配, header hash", headerHash.TerminalString(), "挖矿结果hash", result.BlockHash.TerminalString())
		return MinerResultError
	}

	if err := p.blockChain().DPOSEngine(header.Version).VerifyBlock(p.blockChain(), header); err != nil {
		log.WARN(p.logExtraInfo(), "挖矿结果DPOS共识失败", err)
		return err
	}

	if err := p.blockChain().Engine(header.Version).VerifySeal(p.blockChain(), header); err != nil {
		log.WARN(p.logExtraInfo(), "挖矿结果POW验证失败", err)
		return err
	}

	return nil
}

func (p *Process) copyHeader(header *types.Header, minerResult *mc.HD_MiningRspMsg) *types.Header {
	newHeader := types.CopyHeader(header)
	newHeader.Nonce = minerResult.Nonce
	newHeader.Coinbase = minerResult.Coinbase
	newHeader.MixDigest = minerResult.MixDigest
	newHeader.Signatures = make([]common.Signature, 0)
	newHeader.Signatures = append(newHeader.Signatures, minerResult.Signatures...)
	return newHeader
}

func (p *Process) insertAndBcBlock(isSelf bool, leader common.Address, header *types.Header) (common.Hash, error) {
	var blockData *blockCacheData = nil
	if p.role == common.RoleBroadcast {
		blockData = p.blockCache.GetLastBlockData()
	} else {
		blockData = p.blockCache.GetBlockData(leader)
	}
	if nil == blockData || blockData.state != blockStateReady {
		return common.Hash{}, HaveNoGenBlockError
	}

	insertHeader := blockData.block.Header
	if isSelf == false {
		if header == nil {
			return common.Hash{}, ParaNull
		}
		if header.HashNoSignsAndNonce() != insertHeader.HashNoSignsAndNonce() {
			return common.Hash{}, HashNoSignNotMatchError
		}
		insertHeader = header
	}

	txs := blockData.block.FinalTxs
	receipts := blockData.block.Receipts
	state := blockData.block.State
	block := types.NewBlockWithTxs(insertHeader, types.MakeCurencyBlock(txs, receipts, nil))

	stat, err := p.blockChain().WriteBlockWithState(block, state)
	if err != nil {
		log.ERROR(p.logExtraInfo(), "插入区块失败", err)
		return common.Hash{}, err
	}
	mc.PublishEvent(mc.BlockInserted, &mc.BlockInsertedMsg{Block: mc.BlockInfo{Hash: block.Hash(), Number: block.NumberU64()}, InsertTime: uint64(time.Now().Unix()), CanonState: stat == core.CanonStatTy})
	// Broadcast the block and announce chain insertion event
	hash := block.Hash()
	p.eventMux().Post(core.NewMinedBlockEvent{Block: block})
	var (
		events []interface{}
		logs   = state.Logs()
	)
	events = append(events, core.ChainEvent{Block: block, Hash: hash, Logs: logs})
	if stat == core.CanonStatTy {
		events = append(events, core.ChainHeadEvent{Block: block})
	}
	p.blockChain().PostChainEvents(events, logs)
	mc.PublishEvent(mc.BlockGenor_HeaderGenerateReq, p.number+1)
	return hash, nil
}
