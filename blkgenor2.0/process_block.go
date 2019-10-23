// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package blkgenorV2

import (
	"math/big"

	"github.com/MatrixAINetwork/go-matrix/params"

	"time"

	"errors"
	"fmt"
	"github.com/MatrixAINetwork/go-matrix/ca"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core"
	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/crypto"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
)

func (p *Process) AddAIMinerResult(aiResult *mc.HD_V2_AIMiningRspMsg) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if err := p.aiPool.AddAIResult(aiResult); err != nil {
		//log.Trace(p.logExtraInfo(), "AI挖矿结果消息处理", "加入AI池失败", "err", err)
		return
	}
	log.Info(p.logExtraInfo(), "AI挖矿结果消息处理", "开始", "高度", aiResult.Number, "AIHash", aiResult.AIHash.TerminalString(), "parent hash", aiResult.BlockHash.TerminalString(), "from", aiResult.From.Hex())
	p.processAIPick()
}

func (p *Process) AddPowMinerResult(minerResult *mc.HD_V2_PowMiningRspMsg) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.bcInterval == nil {
		log.Info(p.logExtraInfo(), "Pow挖矿结果消息处理", "广播周期信息为nil", "number", p.number)
		return
	}

	if params.IsPowBlock(p.number, p.bcInterval.GetBroadcastInterval()) == false {
		log.Debug(p.logExtraInfo(), "Pow挖矿结果消息处理", "非POW区块,抛弃消息", "number", p.number, "bc interval", p.bcInterval.GetBroadcastInterval())
		return
	}

	if err := p.powPool.AddMinerResult(minerResult.BlockHash, minerResult.Difficulty, minerResult); err != nil {
		//log.Trace(p.logExtraInfo(), "Pow挖矿结果消息处理", "加入POW池失败", "err", err)
		return
	}
	log.Info(p.logExtraInfo(), "Pow挖矿结果消息处理", "开始", "高度", minerResult.Number, "难度", minerResult.Difficulty.Uint64(), "mine hash", minerResult.BlockHash.TerminalString(), "from", minerResult.From.Hex())
	p.processPowCombine(true)
}

func (p *Process) AddBasePowResult(basePowerResult *mc.HD_BasePowerDifficulty) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if err := p.basePowPool.AddBasePowResult(basePowerResult.BlockHash, params.BasePowerDifficulty, basePowerResult); err != nil {
		return
	}
}

func (p *Process) AddPOSBlock(block *mc.BlockPOSFinishedV2) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.blockPool.SavePosBlock(block)
	p.processPosWaiting()
}

func (p *Process) startAIPick() {
	if p.checkState(StateBlockBroadcast) == false {
		log.Warn(p.logExtraInfo(), "准备进入AI结果选取，状态错误", p.state.String())
		return
	}
	if p.bcInterval == nil {
		log.Warn(p.logExtraInfo(), "准备进入AI结果选取", "广播周期为nil")
		return
	}
	log.Trace(p.logExtraInfo(), "AI结果选取", "开始", "number", p.number)
	p.state = StateAIPick

	if params.IsAIBlock(p.number, p.bcInterval.GetBroadcastInterval()) == false || p.bcInterval.IsReElectionNumber(p.number-1) {
		log.Trace(p.logExtraInfo(), "AI结果选取", "非AI区块", "number", p.number, "bc interval", p.bcInterval.GetBroadcastInterval())
		p.startHeaderGen(nil)
	} else {
		p.processAIPick()
	}
}

func (p *Process) processAIPick() {
	if p.checkState(StateAIPick) == false {
		log.Warn(p.logExtraInfo(), "AI结果选取阶段", "状态错误", "cur state", p.state.String())
		return
	}

	if (p.parentHash == common.Hash{}) {
		log.Info(p.logExtraInfo(), "AI结果选取阶段", "父区块hash为空")
		return
	}

	_, aiMineHash, err := p.getMineHeader(p.number-1, p.parentHash)
	if err != nil {
		log.Info(p.logExtraInfo(), "AI结果选取阶段", "获取 Mine header 失败", "err", err)
		return
	}
	aiResultList, err := p.aiPool.GetAIResults(aiMineHash)
	if err != nil {
		log.Warn(p.logExtraInfo(), "AI结果选取阶段", "从AI结果池获取数据失败", err, "高度", p.number, "mine hash", aiMineHash.TerminalString())
		return
	}
	if len(aiResultList) == 0 {
		log.Info(p.logExtraInfo(), "AI结果选取阶段", "当前没有AI结果", "高度", p.number, "mine hash", aiMineHash.TerminalString())
		return
	}

	version := p.pm.manblk.ProduceBlockVersion(p.number, string(p.parentHeader.Version))

	for _, aiResult := range aiResultList {
		if aiResult.verified {
			continue
		}
		aiResult.verified = true

		verifiedHeader := &types.Header{
			Number:     big.NewInt((int64)(p.number)),
			ParentHash: p.parentHash,
			AICoinbase: aiResult.aiMsg.AICoinbase,
			AIHash:     aiResult.aiMsg.AIHash,
		}

		if err := p.blockChain().Engine([]byte(version)).VerifyAISeal(p.blockChain(), verifiedHeader); err != nil {
			log.Warn(p.logExtraInfo(), "AI挖矿结果验证失败", err)
			aiResult.legal = false
			continue
		} else {
			aiResult.legal = true
		}

		p.startHeaderGen(aiResult.aiMsg)
		return
	}
}

func (p *Process) processPosWaiting() {
	if p.checkState(StatePOSWaiting) == false {
		log.Warn(p.logExtraInfo(), "POS结果等待阶段，状态错误", p.state.String())
		return
	}

	var blockData *blockInfoData = nil
	if p.role == common.RoleBroadcast {
		blockData = p.blockPool.GetLastBlockData()
	} else {
		blockData = p.blockPool.GetBlockData(p.curLeader)
	}

	if blockData == nil {
		log.Info(p.logExtraInfo(), "POS结果等待阶段", "尚未收到POS结果", "cur leader", p.curLeader.Hex(), "number", p.number)
		return
	}

	// 开始POW结果组合
	p.startPowCombine()
}

func (p *Process) startPowCombine() {
	if p.checkState(StatePOSWaiting) == false {
		log.Warn(p.logExtraInfo(), "准备进入POW结果组合阶段", "状态错误", "cur state", p.state.String())
		return
	}

	combinePow := params.IsPowBlock(p.number, p.bcInterval.GetBroadcastInterval())
	log.Trace(p.logExtraInfo(), "POW结果组合阶段", "开始", "number", p.number, "bc interval", p.bcInterval.GetBroadcastInterval(), "是否组合POW", combinePow)
	p.state = StatePowCombine
	p.processPowCombine(combinePow)

}

func (p *Process) processPowCombine(combinePow bool) {
	if p.checkState(StatePowCombine) == false {
		log.Warn(p.logExtraInfo(), "POW结果组合阶段", "状态错误", "cur state", p.state.String())
		return
	}

	var blockData *blockInfoData = nil
	if p.role == common.RoleBroadcast {
		blockData = p.blockPool.GetLastBlockData()
	} else {
		blockData = p.blockPool.GetBlockData(p.curLeader)
	}
	if nil == blockData {
		log.Warn(p.logExtraInfo(), "POW结果组合阶段", "区块数据获取失败", "高度", p.number, "身份", p.role, "leader", p.curLeader.Hex())
		return
	}

	if combinePow == false { // 当前区块不用组合POW消息
		blockData.state = blockStateComplete
	}
	if blockData.state == blockStatePosFinished {
		if err := p.combinePowInfo(blockData); err != nil {
			log.Warn(p.logExtraInfo(), "POW结果组合阶段", "组合POW结果失败", "err", err, "高度", p.number, "身份", p.role, "leader", p.curLeader.Hex())
			return
		}
	}

	// 关闭验证区块请求发送器
	p.closeMsgSender()
	// 发送新区块准备完毕消息
	readyMsg := &mc.NewBlockReadyMsg{
		Header: blockData.block.Header,
		State:  blockData.block.State.Copy(),
	}
	log.Info(p.logExtraInfo(), "POW结果组合阶段", "组合完毕", "发送新区块准备完毕消息", p.number, "leader", readyMsg.Header.Leader.Hex())
	mc.PublishEvent(mc.BlockGenor_NewBlockReady, readyMsg)

	p.state = StateBlockInsert
	p.startBlockInsert(p.curLeader)
}

func (p *Process) combinePowInfo(blockData *blockInfoData) error {
	// 获取mine header 信息
	powMineHeader, powMineHash, err := p.getMineHeader(p.number, blockData.block.Header.ParentHash)
	if err != nil {
		log.Info(p.logExtraInfo(), "POW结果组合阶段", "获取 Mine header 失败", "err", err)
		return err
	}

	// 先处理普通矿工挖矿结果
	diff := powMineHeader.Difficulty
	powList := p.powPool.GetMinerResults(powMineHash, diff)
	log.Trace("POW结果组合阶段", "pow挖矿结果数量", len(powList), "number", p.number, "difficulty", diff, "parent hash", blockData.block.Header.ParentHash.TerminalString())
	if err := p.dealCombineWithPowList(blockData, powList, powMineHash); err == nil {
		return nil
	}

	// 后处理基金会矿工挖矿结果
	innerMinerPowList := p.powPool.GetMinerResults(powMineHash, params.InnerMinerDifficulty)
	log.Trace("POW结果组合阶段", "基金会Pow挖矿结果数量", len(innerMinerPowList), "number", p.number, "difficulty", params.InnerMinerDifficulty, "parent hash", blockData.block.Header.ParentHash.TerminalString())
	if err := p.dealCombineWithPowList(blockData, innerMinerPowList, powMineHash); err == nil {
		log.Info("POW结果组合阶段", "选取基金会挖矿结果", "成功")
		return nil
	}

	return errors.New("no legal pow result")
}

func (p *Process) dealCombineWithPowList(blockData *blockInfoData, powList []*powInfo, powMineHash common.Hash) error {
	if len(powList) == 0 {
		return errors.New("no pow result")
	}

	for _, powInfo := range powList {
		if powInfo.verified {
			continue
		}

		if powInfo.verified {
			if powInfo.legal {
				combinedHeader := p.copyHeader(blockData.block.Header, powInfo.powMsg)
				combinedHeader.BasePowers = p.getBasePowerList(blockData.block.Header, powMineHash)
				blockData.block.Header = combinedHeader
				blockData.state = blockStateComplete
				return nil
			}
		} else {
			verifyHeader := p.copyHeader(blockData.block.Header, powInfo.powMsg)
			if err := p.blockChain().Engine(verifyHeader.Version).VerifySeal(p.blockChain(), verifyHeader); err != nil {
				log.Warn(p.logExtraInfo(), "POW结果组合阶段", "POW结果验证失败", "miner", powInfo.powMsg.Coinbase.Hex(), "err", err)
				powInfo.legal = false
				continue
			} else {
				powInfo.legal = true
				verifyHeader.BasePowers = p.getBasePowerList(blockData.block.Header, powMineHash)
				blockData.block.Header = verifyHeader
				blockData.state = blockStateComplete
				return nil
			}
		}
	}
	return errors.New("no legal pow result")
}

func (p *Process) getPowDiff(parentHash common.Hash) *big.Int {
	realParentHash := parentHash
	parent := p.blockChain().GetHeaderByHash(realParentHash)
	if parent == nil {
		log.Info(p.logExtraInfo(), "获取POW难度值失败", "父区块获取失败", "parent hash", realParentHash.TerminalString())
		return nil
	}

	return parent.Difficulty
}

func (p *Process) getBasePowerList(header *types.Header, mineHash common.Hash) []types.BasePowers {
	headerBasePowerList := make([]types.BasePowers, 0)
	basePowInfoList, err := p.basePowPool.GetbasPowResults(mineHash, params.BasePowerDifficulty)
	if err != nil {
		log.Warn(p.logExtraInfo(), "进行算力检测验证，算力结果获取失败", err, "高度", p.number, "难度", params.BasePowerDifficulty, "mine hash", mineHash.TerminalString())
		return headerBasePowerList
	}

	headerBasePowMap := make(map[common.Address]bool, 0)
	for _, basePowInfo := range basePowInfoList {
		if _, ok := headerBasePowMap[basePowInfo.powMsg.Coinbase]; ok {
			continue
		}
		headerBasePower := types.BasePowers{Miner: basePowInfo.powMsg.Coinbase, Nonce: basePowInfo.powMsg.Nonce, MixDigest: basePowInfo.powMsg.MixDigest}
		if !basePowInfo.verified {
			basePowInfo.verified = true

			if err := p.blockChain().Engine(header.Version).VerifyBasePow(p.blockChain(), header, headerBasePower); err != nil {
				log.Warn(p.logExtraInfo(), "挖矿结果POW验证失败", err)
				basePowInfo.legal = false
				continue
			} else {
				basePowInfo.legal = true
			}
		}

		headerBasePowerList = append(headerBasePowerList, types.BasePowers{Miner: basePowInfo.powMsg.Coinbase, Nonce: basePowInfo.powMsg.Nonce, MixDigest: basePowInfo.powMsg.MixDigest})
		headerBasePowMap[basePowInfo.powMsg.Coinbase] = true
	}
	return headerBasePowerList
}

func (p *Process) copyHeader(header *types.Header, minerResult *mc.HD_V2_PowMiningRspMsg) *types.Header {
	newHeader := types.CopyHeader(header)
	newHeader.Nonce = minerResult.Nonce
	newHeader.Coinbase = minerResult.Coinbase
	newHeader.MixDigest = minerResult.MixDigest
	newHeader.Sm3Nonce = minerResult.Sm3Nonce
	return newHeader
}

func (p *Process) startBlockInsert(blockLeader common.Address) {
	if p.canInsertBlock() == false {
		return
	}

	log.Info(p.logExtraInfo(), "区块插入", "开始", "高度", p.number)
	defer log.Debug(p.logExtraInfo(), "区块插入", "完成", "高度", p.number)
	if err := p.processBlockInsert(true, blockLeader, nil); err != nil {
		log.Error(p.logExtraInfo(), "区块插入，错误", err)
		return
	}
	log.Info(p.logExtraInfo(), "关键时间点", "leader挂块成功", "time", time.Now(), "块高", p.number)
	p.state = StateEnd
}

func (p *Process) canInsertBlock() bool {
	if p.state < StateBlockInsert {
		log.Warn(p.logExtraInfo(), "准备进行区块插入，状态错误", p.state.String(), "高度", p.number)
		return false
	}
	if p.bcInterval == nil {
		log.Warn(p.logExtraInfo(), "准备进行区块插入", "广播周期信息为nil")
		return false
	}
	if p.bcInterval.IsBroadcastNumber(p.number + 1) {
		if p.role != common.RoleBroadcast {
			log.Warn(p.logExtraInfo(), "准备进行区块插入，广播区块前一个区块，由广播节点插入", p.role.String(), "高度", p.number)
			return false
		}
	} else {
		if p.role != common.RoleValidator {
			log.Warn(p.logExtraInfo(), "准备进行区块插入，身份错误", "当前身份不是验证者", "高度", p.number, "身份", p.role.String())
			return false
		}

		if (p.nextLeader == common.Address{}) {
			log.Warn(p.logExtraInfo(), "准备进行区块插入", "下个区块leader为空", "需要等待leader的高度", p.number+1)
			return false
		}

		if p.nextLeader != ca.GetDepositAddress() {
			log.Debug(p.logExtraInfo(), "准备进行区块广播,自己不是下个区块leader,高度", p.number, "next leader", p.nextLeader.Hex(), "self", ca.GetDepositAddress().Hex())
			return false
		}
	}
	return true
}

func (p *Process) processBlockInsert(isSelf bool, leader common.Address, header *types.Header) error {
	var blockData *blockInfoData = nil
	if p.role == common.RoleBroadcast {
		blockData = p.blockPool.GetLastBlockData()
	} else {
		blockData = p.blockPool.GetBlockData(leader)
	}
	if nil == blockData {
		return HaveNoBlockDataError
	}

	insertHeader := blockData.block.Header
	if isSelf == false {
		if header == nil {
			return ParaNull
		}
		if header.HashNoSignsAndNonce() != insertHeader.HashNoSignsAndNonce() {
			return HashNoSignNotMatchError
		}
		insertHeader = header
	} else {
		if blockData.state != blockStateComplete {
			return BlockDataNotCompleteError
		}
	}

	return p.insertBlock(insertHeader, blockData.block.FinalTxs, blockData.block.Receipts, blockData.block.State)
}

func (p *Process) insertBlock(header *types.Header, finalTxs []types.CoinSelfTransaction, receipts []types.CoinReceipts, stateDB *state.StateDBManage) error {
	block := types.NewBlockWithTxs(header, types.MakeCurencyBlock(finalTxs, receipts, nil))
	stat, err := p.blockChain().WriteBlockWithState(block, stateDB)
	if err != nil {
		log.Error(p.logExtraInfo(), "processInsertBlock 失败", err)
		return err
	}
	mc.PublishEvent(mc.BlockInserted, &mc.BlockInsertedMsg{Block: mc.BlockInfo{Hash: block.Hash(), Number: block.NumberU64()}, InsertTime: uint64(time.Now().Unix()), CanonState: stat == core.CanonStatTy})
	// Broadcast the block and announce chain insertion event
	hash := block.Hash()
	var (
		events []interface{}
		logs   = stateDB.Logs()
	)
	events = append(events, core.ChainEvent{Block: block, Hash: hash, Logs: logs})
	if stat == core.CanonStatTy {
		events = append(events, core.ChainHeadEvent{Block: block})
	}
	p.blockChain().PostChainEvents(events, logs)
	return nil
}

func (p *Process) processBlockInsertMsg(blkInsertMsg *mc.HD_BlockInsertNotify) {
	if blkInsertMsg == nil || blkInsertMsg.Header == nil {
		log.Warn(p.logExtraInfo(), "区块插入", "消息为空")
		return
	}

	blockHash := blkInsertMsg.Header.Hash()
	log.Info(p.logExtraInfo(), "区块插入", "启动", "区块 hash", blockHash.TerminalString(), "from", blkInsertMsg.From.Hex(), "高度", p.number)

	if p.checkRepeatInsert(blockHash) {
		log.Trace(p.logExtraInfo(), "插入区块已处理", p.number, "区块 hash", blockHash.TerminalString())
		return
	}

	parentBlock := p.blockChain().GetBlockByHash(blkInsertMsg.Header.ParentHash)
	if parentBlock == nil {
		log.Warn(p.logExtraInfo(), "区块插入", "缺少父区块, 进行fetch", "父区块 hash", blkInsertMsg.Header.ParentHash.TerminalString())
		p.backend().FetcherNotify(blkInsertMsg.Header.ParentHash, p.number-1, blkInsertMsg.From)
		return
	}

	bcInterval, err := p.blockChain().GetBroadcastIntervalByHash(blkInsertMsg.Header.ParentHash)
	if err != nil {
		log.Error(p.logExtraInfo(), "区块插入", "获取广播周期错误", "err", err)
		return
	}

	header := blkInsertMsg.Header
	if false == p.checkInsertedHeader(bcInterval, header) {
		return
	}

	if err := p.processBlockInsert(false, header.Leader, header); err != nil {
		log.Warn(p.logExtraInfo(), "区块插入失败, err", err, "fetch 高度", p.number, "fetch hash", blockHash.TerminalString(), "source", blkInsertMsg.From.Hex())
		p.backend().FetcherNotify(blockHash, p.number, blkInsertMsg.From)
	}

	p.saveInsertedBlockHash(blockHash)
}

func (p *Process) checkInsertedHeader(bcInterval *mc.BCIntervalInfo, header *types.Header) bool {
	if bcInterval.IsBroadcastNumber(p.number) {
		signAccount, _, err := crypto.VerifySignWithValidate(header.HashNoSigns().Bytes(), header.Signatures[0].Bytes())
		if err != nil {
			log.Error(p.logExtraInfo(), "广播区块插入消息非法, 签名解析错误", err)
			return false
		}

		if signAccount != header.Leader {
			log.Warn(p.logExtraInfo(), "广播区块插入消息非法, 签名不匹配，签名人", signAccount.Hex(), "Leader", header.Leader.Hex())
			return false
		}

		if role, _ := ca.GetAccountOriginalRole(signAccount, header.ParentHash); common.RoleBroadcast != role {
			log.Warn(p.logExtraInfo(), "广播区块插入消息非法，签名人不是广播身份, 角色", role.String())
			return false
		}
	} else {
		if err := p.blockChain().DPOSEngine(header.Version).VerifyBlock(p.blockChain(), header); err != nil {
			log.Error(p.logExtraInfo(), "区块插入消息DPOS共识失败", err, "signs", len(header.Signatures))
			return false
		}

		if err := p.blockChain().Engine(header.Version).VerifySeal(p.blockChain(), header); err != nil {
			log.Error(p.logExtraInfo(), "区块插入消息POW验证失败", err)
			return false
		}
	}
	return true
}

func (p *Process) getMineHeader(number uint64, sonHash common.Hash) (*types.Header, common.Hash, error) {
	if p.bcInterval == nil {
		return nil, common.Hash{}, errors.New("broadcast interval is nil")
	}

	mineHashNumber := params.GetCurAIBlockNumber(number, p.bcInterval.GetBroadcastInterval())
	mineHeaderHash, err := p.pm.bc.GetAncestorHash(sonHash, mineHashNumber)
	if err != nil {
		log.Info(p.logExtraInfo(), "getMineHeader", "get mine header hash failed", "err", err, "number", number, "mine number", mineHashNumber)
		return nil, common.Hash{}, fmt.Errorf("get mine header hash err: %v", err)
	}
	mineHeader := p.pm.bc.GetHeaderByHash(mineHeaderHash)
	if mineHeader == nil {
		log.Info(p.logExtraInfo(), "getMineHeader", "get mine header failed", "cur number", p.number, "mine number", mineHashNumber, "hash", mineHeaderHash.TerminalString())
		return nil, common.Hash{}, fmt.Errorf("get mine header err")
	}

	return mineHeader, mineHeader.HashNoSignsAndNonce(), nil
}
