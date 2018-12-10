// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package blkgenor

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/event"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
)

type BlockGenor struct {
	pm                      *ProcessManage
	man                     Backend
	roleUpdatedMsgCh        chan *mc.RoleUpdatedMsg
	leaderChangeNotifyCh    chan *mc.LeaderChangeNotify
	minerResultCh           chan *mc.HD_MiningRspMsg
	broadcastMinerResultCh  chan *mc.HD_BroadcastMiningRspMsg
	blockConsensusCh        chan *mc.BlockVerifyConsensusOK
	preBlockFinishedCh      chan *mc.PreBlockBroadcastFinished
	blockInsertCh           chan *mc.HD_BlockInsertNotify
	roleUpdatedMsgSub       event.Subscription
	leaderChangeSub         event.Subscription
	minerResultSub          event.Subscription
	broadcastMinerResultSub event.Subscription
	blockConsensusSub       event.Subscription
	preBlockFinishedSub     event.Subscription
	blockInsertSub          event.Subscription
}

func New(man Backend) (*BlockGenor, error) {
	if nil == &man {
		log.Error("nil == &man Error")
		return nil, ParaNull
	}
	if nil == man.BlockChain().Engine() {
		log.Error("man.BlockChain().Engine() Error")
		return nil, ParaNull
	}
	//if nil==man.ReElection(){
	//	return nil,ParaNull
	//}

	bg := &BlockGenor{
		man: man,

		roleUpdatedMsgCh:       make(chan *mc.RoleUpdatedMsg, 1),
		leaderChangeNotifyCh:   make(chan *mc.LeaderChangeNotify, 1),
		minerResultCh:          make(chan *mc.HD_MiningRspMsg, 1),
		broadcastMinerResultCh: make(chan *mc.HD_BroadcastMiningRspMsg, 1),
		blockConsensusCh:       make(chan *mc.BlockVerifyConsensusOK, 1),
		preBlockFinishedCh:     make(chan *mc.PreBlockBroadcastFinished, 1),
		blockInsertCh:          make(chan *mc.HD_BlockInsertNotify, 10),
	}

	bg.pm = NewProcessManage(man)

	var err error
	if bg.roleUpdatedMsgSub, err = mc.SubscribeEvent(mc.CA_RoleUpdated, bg.roleUpdatedMsgCh); err != nil {
		return nil, err
	}
	if bg.leaderChangeSub, err = mc.SubscribeEvent(mc.Leader_LeaderChangeNotify, bg.leaderChangeNotifyCh); err != nil {
		return nil, err
	}
	if bg.minerResultSub, err = mc.SubscribeEvent(mc.HD_MiningRsp, bg.minerResultCh); err != nil {
		return nil, err
	}
	if bg.broadcastMinerResultSub, err = mc.SubscribeEvent(mc.HD_BroadcastMiningRsp, bg.broadcastMinerResultCh); err != nil {
		return nil, err
	}
	if bg.blockConsensusSub, err = mc.SubscribeEvent(mc.BlkVerify_VerifyConsensusOK, bg.blockConsensusCh); err != nil {
		return nil, err
	}
	if bg.preBlockFinishedSub, err = mc.SubscribeEvent(mc.BlockGenor_PreBlockBroadcastFinished, bg.preBlockFinishedCh); err != nil {
		return nil, err
	}
	if bg.blockInsertSub, err = mc.SubscribeEvent(mc.HD_NewBlockInsert, bg.blockInsertCh); err != nil {
		return nil, err
	}

	go bg.update()

	return bg, nil
}

func (self *BlockGenor) update() {
	defer func() {
		self.preBlockFinishedSub.Unsubscribe()
		self.blockConsensusSub.Unsubscribe()
		self.broadcastMinerResultSub.Unsubscribe()
		self.minerResultSub.Unsubscribe()
		self.leaderChangeSub.Unsubscribe()
		self.roleUpdatedMsgSub.Unsubscribe()
	}()

	for {
		select {
		case roleMsg := <-self.roleUpdatedMsgCh:
			go self.roleUpdatedMsgHandle(roleMsg)

		case leaderMsg := <-self.leaderChangeNotifyCh:
			go self.leaderChangeNotifyHandle(leaderMsg)

		case minerResult := <-self.minerResultCh:
			go self.minerResultHandle(minerResult)

		case broadcastMinerResult := <-self.broadcastMinerResultCh:
			go self.broadcastMinerResultHandle(broadcastMinerResult)

		case consensusBlock := <-self.blockConsensusCh:
			go self.consensusBlockMsgHandle(consensusBlock)

		case preBlockMsg := <-self.preBlockFinishedCh:
			go self.preBlockFinishedMsgHandle(preBlockMsg)

		case blockInsertMsg := <-self.blockInsertCh:
			go self.blockInsertMsgHandle(blockInsertMsg)
		}
	}
}

func (self *BlockGenor) roleUpdatedMsgHandle(roleMsg *mc.RoleUpdatedMsg) error {
	log.INFO(self.logExtraInfo(), "CA身份消息处理", "开始", "高度", roleMsg.BlockNum, "角色", roleMsg.Role.String())
	defer log.INFO(self.logExtraInfo(), "CA身份消息处理", "结束", "高度", roleMsg.BlockNum)

	curNumber := roleMsg.BlockNum + 1
	self.pm.SetCurNumber(curNumber)
	if roleMsg.Role == common.RoleValidator || roleMsg.Role == common.RoleBroadcast {
		curProcess := self.pm.GetCurrentProcess()
		curProcess.StartRunning(roleMsg.Role)
	}

	return nil
}

func (self *BlockGenor) leaderChangeNotifyHandle(leaderMsg *mc.LeaderChangeNotify) {
	log.INFO(self.logExtraInfo(), "Leader变更消息处理", "开始", "高度", leaderMsg.Number, "轮次",
		leaderMsg.ReelectTurn, "有效", leaderMsg.ConsensusState, "leader", leaderMsg.Leader.Hex(), "next leader", leaderMsg.NextLeader.Hex())
	defer log.INFO(self.logExtraInfo(), "Leader变更消息处理", "结束", "高度", leaderMsg.Number, "轮次", leaderMsg.ReelectTurn, "有效", leaderMsg.ConsensusState)

	number := leaderMsg.Number
	var process, preProcess *Process
	var err error

	if number == 1 { // 第一个区块特殊处理
		process, err = self.pm.GetProcess(number)
	} else {
		process, preProcess, err = self.pm.GetProcessAndPreProcess(number)
	}

	if err != nil {
		log.INFO(self.logExtraInfo(), "Leader变更消息 获取Process失败", err)
		return
	}

	if leaderMsg.ConsensusState {
		process.SetCurLeader(leaderMsg.Leader)
		process.SetNextLeader(leaderMsg.NextLeader)
		if preProcess != nil {
			preProcess.SetNextLeader(leaderMsg.Leader)
		}

		// 提前设置下个process的leader
		nextProcess, err := self.pm.GetProcess(number + 1)
		if err == nil {
			nextProcess.SetCurLeader(leaderMsg.NextLeader)
		}
	} else {
		process.ReInit()
		if preProcess != nil {
			preProcess.ReInitNextLeader()
		}
	}
}

func (self *BlockGenor) minerResultHandle(minerResult *mc.HD_MiningRspMsg) {
	log.INFO(self.logExtraInfo(), "矿工挖矿结果消息处理", "开始", "高度", minerResult.Number, "block hash", minerResult.Blockhash.TerminalString())
	defer log.INFO(self.logExtraInfo(), "矿工挖矿结果消息处理", "结束", "高度", minerResult.Number, "block hash", minerResult.Blockhash.TerminalString())
	process, err := self.pm.GetProcess(minerResult.Number)
	if err != nil {
		log.INFO(self.logExtraInfo(), "矿工挖矿结果消息 获取Process失败", err)
		return
	}
	process.AddMinerResult(minerResult)
}

func (self *BlockGenor) broadcastMinerResultHandle(result *mc.HD_BroadcastMiningRspMsg) {
	number := result.BlockMainData.Header.Number.Uint64()
	log.INFO(self.logExtraInfo(), "广播矿工挖矿结果消息处理", "开始", "高度", number, "交易数量", result.BlockMainData.Txs.Len())
	defer log.INFO(self.logExtraInfo(), "广播矿工挖矿结果消息处理", "结束", "高度", number)
	for _, tx := range result.BlockMainData.Txs {
		log.INFO(self.logExtraInfo(), "广播矿工挖矿结果消息 高度", number, "交易", tx)
	}
	process, err := self.pm.GetProcess(number)
	if err != nil {
		log.INFO(self.logExtraInfo(), "矿工挖矿结果消息 获取Process失败", err)
		return
	}
	process.AddBroadcastMinerResult(result)
}

func (self *BlockGenor) consensusBlockMsgHandle(data *mc.BlockVerifyConsensusOK) {
	log.INFO(self.logExtraInfo(), "共识结果消息处理", "开始", "高度", data.Header.Number, "block hash", data.BlockHash.TerminalString(), "计算hash", data.Header.HashNoSignsAndNonce().TerminalString())
	defer log.INFO(self.logExtraInfo(), "共识结果消息处理", "结束", "高度", data.Header.Number)
	process, err := self.pm.GetProcess(data.Header.Number.Uint64())
	if err != nil {
		log.INFO(self.logExtraInfo(), "共识结果消息 获取Process失败", err)
		return
	}

	process.AddConsensusBlock(data)
}

func (self *BlockGenor) preBlockFinishedMsgHandle(preBlock *mc.PreBlockBroadcastFinished) {
	log.INFO(self.logExtraInfo(), "前一区块广播完成消息处理", "开始", "高度", preBlock.Number, "block hash", preBlock.BlockHash.TerminalString())
	defer log.INFO(self.logExtraInfo(), "前一区块广播完成消息处理", "结束", "高度", preBlock.Number, "block hash", preBlock.BlockHash.TerminalString())
	process, err := self.pm.GetProcess(preBlock.Number + 1)
	if err != nil {
		log.INFO(self.logExtraInfo(), "共识结果消息 获取Process失败", err)
		return
	}

	process.AddPreBlockInfo(preBlock)
}

func (self *BlockGenor) blockInsertMsgHandle(blockInsert *mc.HD_BlockInsertNotify) {
	number := blockInsert.Header.Number.Uint64()
	curNumber := self.pm.GetCurNumber()
	log.INFO(self.logExtraInfo(), "收到的区块插入消息广播高度", number, "from", blockInsert.From.Hex(), "当前高度", curNumber)

	if number > curNumber {
		log.INFO(self.logExtraInfo(), "+++++fetch 区块高度", number)
		self.pm.matrix.FetcherNotify(blockInsert.Header.Hash(), blockInsert.Header.Number.Uint64())
		return
	}

	process, err := self.pm.GetProcess(number)
	if err != nil {
		log.INFO(self.logExtraInfo(), "最终区块插入 获取Process失败", err)
		return
	}
	//log.INFO(self.logExtraInfo(), "最终区块插入 获取Process成功", err)
	process.AddInsertBlockInfo(blockInsert)
}

func (self *BlockGenor) logExtraInfo() string {
	return "区块生成"
}
