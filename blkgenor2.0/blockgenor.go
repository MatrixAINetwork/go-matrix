// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package blkgenorV2

import (
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/event"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/params/manversion"
)

type BlockGenor struct {
	pm                      *ProcessManage
	man                     Backend
	quitCh                  chan struct{}
	roleUpdatedMsgCh        chan *mc.RoleUpdatedMsg
	leaderChangeNotifyCh    chan *mc.LeaderChangeNotify
	powResultCh             chan *mc.HD_V2_PowMiningRspMsg
	aiResultCh              chan *mc.HD_V2_AIMiningRspMsg
	posBlockCh              chan *mc.BlockPOSFinishedV2
	blockInsertCh           chan *mc.HD_BlockInsertNotify
	broadcastBlockResultCh  chan *mc.HD_BroadcastMiningRspMsg
	basePowerCh             chan *mc.HD_BasePowerDifficulty
	recoveryCh              chan *mc.RecoveryStateMsg
	fullBlockReqCh          chan *mc.HD_V2_FullBlockReqMsg
	fullBlockRspCh          chan *mc.HD_V2_FullBlockRspMsg
	roleUpdatedMsgSub       event.Subscription
	leaderChangeSub         event.Subscription
	powResultSub            event.Subscription
	aiResultSub             event.Subscription
	posBlockSub             event.Subscription
	blockInsertSub          event.Subscription
	broadcastBlockResultSub event.Subscription
	basePowerSub            event.Subscription
	recoverySub             event.Subscription
	fullBlockReqSub         event.Subscription
	fullBlockRspSub         event.Subscription
}

func New(man Backend) (*BlockGenor, error) {
	if nil == &man {
		log.Error("区块生成模块，传入的参数为空")
		return nil, ParaNull
	}

	bg := &BlockGenor{
		man:                    man,
		quitCh:                 make(chan struct{}),
		roleUpdatedMsgCh:       make(chan *mc.RoleUpdatedMsg, 1),
		leaderChangeNotifyCh:   make(chan *mc.LeaderChangeNotify, 1),
		powResultCh:            make(chan *mc.HD_V2_PowMiningRspMsg, 1),
		aiResultCh:             make(chan *mc.HD_V2_AIMiningRspMsg, 1),
		posBlockCh:             make(chan *mc.BlockPOSFinishedV2, 1),
		blockInsertCh:          make(chan *mc.HD_BlockInsertNotify, 10),
		broadcastBlockResultCh: make(chan *mc.HD_BroadcastMiningRspMsg, 1),
		basePowerCh:            make(chan *mc.HD_BasePowerDifficulty, 1),
		recoveryCh:             make(chan *mc.RecoveryStateMsg, 1),
		fullBlockReqCh:         make(chan *mc.HD_V2_FullBlockReqMsg, 1),
		fullBlockRspCh:         make(chan *mc.HD_V2_FullBlockRspMsg, 1),
	}

	bg.pm = NewProcessManage(man)

	var err error
	if bg.roleUpdatedMsgSub, err = mc.SubscribeEvent(mc.CA_RoleUpdated, bg.roleUpdatedMsgCh); err != nil {
		log.Error(bg.logExtraInfo(), "订阅错误，消息号", mc.CA_RoleUpdated, "错误：", err)
		return nil, err
	}
	if bg.leaderChangeSub, err = mc.SubscribeEvent(mc.Leader_LeaderChangeNotify, bg.leaderChangeNotifyCh); err != nil {
		log.Error(bg.logExtraInfo(), "订阅错误，消息号", mc.Leader_LeaderChangeNotify, "错误：", err)
		return nil, err
	}
	if bg.powResultSub, err = mc.SubscribeEvent(mc.HD_V2_PowMiningRsp, bg.powResultCh); err != nil {
		log.Error(bg.logExtraInfo(), "订阅错误，消息号", mc.HD_V2_PowMiningRsp, "错误：", err)
		return nil, err
	}
	if bg.aiResultSub, err = mc.SubscribeEvent(mc.HD_V2_AIMiningRsp, bg.aiResultCh); err != nil {
		log.Error(bg.logExtraInfo(), "订阅错误，消息号", mc.HD_V2_AIMiningRsp, "错误：", err)
		return nil, err
	}
	if bg.posBlockSub, err = mc.SubscribeEvent(mc.BlkVerify_POSFinishedNotifyV2, bg.posBlockCh); err != nil {
		log.Error(bg.logExtraInfo(), "订阅错误，消息号", mc.BlkVerify_POSFinishedNotifyV2, "错误：", err)
		return nil, err
	}
	if bg.blockInsertSub, err = mc.SubscribeEvent(mc.HD_NewBlockInsert, bg.blockInsertCh); err != nil {
		log.Error(bg.logExtraInfo(), "订阅错误，消息号", mc.HD_NewBlockInsert, "错误：", err)
		return nil, err
	}
	if bg.broadcastBlockResultSub, err = mc.SubscribeEvent(mc.HD_BroadcastMiningRsp, bg.broadcastBlockResultCh); err != nil {
		log.Error("区块生成模块", "订阅错误，消息号", mc.HD_BroadcastMiningRsp, "错误：", err)
		return nil, err
	}
	if bg.basePowerSub, err = mc.SubscribeEvent(mc.HD_BasePowerResult, bg.basePowerCh); err != nil {
		log.Error(bg.logExtraInfo(), "订阅错误，消息号", mc.HD_BasePowerResult, "错误：", err)
		return nil, err
	}
	if bg.recoverySub, err = mc.SubscribeEvent(mc.Leader_RecoveryState, bg.recoveryCh); err != nil {
		log.Error(bg.logExtraInfo(), "订阅错误，消息号", mc.Leader_RecoveryState, "错误：", err)
		return nil, err
	}
	if bg.fullBlockReqSub, err = mc.SubscribeEvent(mc.HD_V2_FullBlockReq, bg.fullBlockReqCh); err != nil {
		log.Error(bg.logExtraInfo(), "订阅错误，消息号", mc.HD_V2_FullBlockReq, "错误：", err)
		return nil, err
	}
	if bg.fullBlockRspSub, err = mc.SubscribeEvent(mc.HD_V2_FullBlockRsp, bg.fullBlockRspCh); err != nil {
		log.Error(bg.logExtraInfo(), "订阅错误，消息号", mc.HD_V2_FullBlockRsp, "错误：", err)
		return nil, err
	}

	go bg.update()
	log.Info("区块生成模块对象创建成功")
	return bg, nil
}

func (self *BlockGenor) Close() {
	close(self.quitCh)
}

func (self *BlockGenor) update() {
	defer func() {
		self.fullBlockRspSub.Unsubscribe()
		self.fullBlockReqSub.Unsubscribe()
		self.recoverySub.Unsubscribe()
		self.blockInsertSub.Unsubscribe()
		self.posBlockSub.Unsubscribe()
		self.aiResultSub.Unsubscribe()
		self.powResultSub.Unsubscribe()
		self.leaderChangeSub.Unsubscribe()
		self.roleUpdatedMsgSub.Unsubscribe()
		self.basePowerSub.Unsubscribe()
		log.Info("区块生成模块退出成功")
	}()

	for {
		select {
		case roleMsg := <-self.roleUpdatedMsgCh:
			go self.roleUpdatedMsgHandle(roleMsg)

		case leaderMsg := <-self.leaderChangeNotifyCh:
			go self.leaderChangeNotifyHandle(leaderMsg)

		case powResult := <-self.powResultCh:
			go self.powResultHandle(powResult)

		case aiResult := <-self.aiResultCh:
			go self.aiResultHandle(aiResult)

		case posBlock := <-self.posBlockCh:
			go self.posBlockMsgHandle(posBlock)

		case basePowResult := <-self.basePowerCh:
			go self.basePowResultHandle(basePowResult)

		case blockInsertMsg := <-self.blockInsertCh:
			go self.blockInsertMsgHandle(blockInsertMsg)

		case broadcastBlockResult := <-self.broadcastBlockResultCh:
			go self.broadcastBlockHandle(broadcastBlockResult)

		case recoveryMsg := <-self.recoveryCh:
			go self.handleRecoveryMsg(recoveryMsg)

		case nbRepMsg := <-self.fullBlockReqCh:
			go self.handleFullBlockReqMsg(nbRepMsg)

		case nbRsqMsg := <-self.fullBlockRspCh:
			go self.handleFullBlockRspMsg(nbRsqMsg)

		case <-self.quitCh:
			return
		}
	}
}

func (self *BlockGenor) roleUpdatedMsgHandle(roleMsg *mc.RoleUpdatedMsg) error {
	log.Info(self.logExtraInfo(), "CA身份消息处理", "开始", "高度", roleMsg.BlockNum, "角色", roleMsg.Role.String(), "block hash", roleMsg.BlockHash.TerminalString(), "version", roleMsg.Version)
	curNumber := roleMsg.BlockNum + 1
	self.pm.SetCurNumber(curNumber, roleMsg.SuperSeq)
	if manversion.VersionCmp(roleMsg.Version, manversion.VersionAIMine) < 0 && curNumber < manversion.VersionNumAIMine {
		log.Trace(self.logExtraInfo(), "CA身份消息处理", "版本号及高度不满足条件，不处理", "指定版本切换高度", manversion.VersionNumAIMine, "指定version", manversion.VersionAIMine)
		return nil
	}

	bcInterval, err := self.man.BlockChain().GetBroadcastIntervalByHash(roleMsg.BlockHash)
	if err != nil {
		log.Error(self.logExtraInfo(), "CA身份消息处理", "获取广播周期信息by hash 失败", "err", err)
		return err
	}
	role := roleMsg.Role
	if role == common.RoleValidator || role == common.RoleBroadcast {
		curProcess := self.pm.GetCurrentProcess()
		curProcess.StartRunning(role, bcInterval)
	}

	return nil
}

func (self *BlockGenor) leaderChangeNotifyHandle(leaderMsg *mc.LeaderChangeNotify) {
	log.Info(self.logExtraInfo(), "Leader变更消息处理", "开始", "高度", leaderMsg.Number, "轮次",
		leaderMsg.ReelectTurn, "有效", leaderMsg.ConsensusState, "leader", leaderMsg.Leader.Hex(), "next leader", leaderMsg.NextLeader.Hex())

	number := leaderMsg.Number
	var process, preProcess *Process
	var err error

	if number == 1 { // 第一个区块特殊处理
		process, err = self.pm.GetProcess(number)
	} else {
		process, preProcess, err = self.pm.GetProcessAndPreProcess(number)
	}

	if err != nil {
		log.Error(self.logExtraInfo(), "Leader变更消息 获取Process失败", err)
		return
	}

	if leaderMsg.ConsensusState {
		process.SetCurLeader(leaderMsg.Leader, leaderMsg.ConsensusTurn)
		process.SetNextLeader(leaderMsg.Leader, leaderMsg.NextLeader)
		if preProcess != nil {
			preProcess.SetNextLeader(leaderMsg.PreLeader, leaderMsg.Leader)
		}

		// 提前设置下个process的leader
		nextProcess, err := self.pm.GetProcess(number + 1)
		if err == nil {
			nextProcess.SetCurLeader(leaderMsg.NextLeader, mc.ConsensusTurnInfo{})
		} else {
			log.Warn(self.logExtraInfo(), "获取下个高度process失败", err)
		}
	} else {
		process.ReInit()
		if preProcess != nil {
			preProcess.ReInitNextLeader()
		}
	}
}

func (self *BlockGenor) powResultHandle(powResult *mc.HD_V2_PowMiningRspMsg) {
	process, err := self.pm.GetProcess(powResult.Number)
	if err != nil {
		log.Info(self.logExtraInfo(), "矿工Pow挖矿结果消息 获取Process失败", err)
		return
	}
	process.AddBasePowResult((*mc.HD_BasePowerDifficulty)(powResult))
	process.AddPowMinerResult(powResult)
}

func (self *BlockGenor) aiResultHandle(aiResult *mc.HD_V2_AIMiningRspMsg) {
	process, err := self.pm.GetProcess(aiResult.Number)
	if err != nil {
		log.Info(self.logExtraInfo(), "矿工AI挖矿结果消息 获取Process失败", err)
		return
	}
	process.AddAIMinerResult(aiResult)
}

func (self *BlockGenor) basePowResultHandle(basePowResult *mc.HD_BasePowerDifficulty) {
	process, err := self.pm.GetProcess(basePowResult.Number)
	if err != nil {
		log.Info(self.logExtraInfo(), "矿工挖矿结果消息 获取Process失败", err)
		return
	}
	process.AddBasePowResult(basePowResult)
	//log.Info(self.logExtraInfo(), "算力检测结果消息处理", "开始", "高度", basePowResult.Number, "难度", basePowResult.Difficulty.Uint64(), "parent hash", basePowResult.BlockHash.TerminalString(), "from", basePowResult.From.Hex())
}

func (self *BlockGenor) posBlockMsgHandle(data *mc.BlockPOSFinishedV2) {
	if manversion.VersionCmp(string(data.Header.Version), manversion.VersionAIMine) < 0 {
		log.Trace(self.logExtraInfo(), "共识结果消息处理", "版本小于指定版本，不处理", "msg version", data.Header.Version, "指定version", manversion.VersionAIMine)
		return
	}
	log.Info(self.logExtraInfo(), "共识结果消息处理", "开始", "高度", data.Header.Number, "block hash", data.BlockHash.TerminalString(), "header signs", len(data.Header.Signatures))
	process, err := self.pm.GetProcess(data.Header.Number.Uint64())
	if err != nil {
		log.Error(self.logExtraInfo(), "共识结果消息 获取Process失败", err)
		return
	}

	process.AddPOSBlock(data)
}

func (self *BlockGenor) blockInsertMsgHandle(blockInsert *mc.HD_BlockInsertNotify) {
	if manversion.VersionCmp(string(blockInsert.Header.Version), manversion.VersionAIMine) < 0 {
		return
	}
	number := blockInsert.Header.Number.Uint64()
	curNumber := self.pm.GetCurNumber()
	log.Info(self.logExtraInfo(), "收到的区块插入消息广播高度", number, "from", blockInsert.From.Hex(), "当前高度", curNumber, "sign count", len(blockInsert.Header.Signatures))

	if number > curNumber {
		log.Debug(self.logExtraInfo(), "fetch 区块高度", number, "from", blockInsert.From.Hex())
		self.pm.matrix.FetcherNotify(blockInsert.Header.Hash(), blockInsert.Header.Number.Uint64(), blockInsert.From)
		return
	}

	process, err := self.pm.GetProcess(number)
	if err != nil {
		log.Error(self.logExtraInfo(), "最终区块插入 获取Process失败", err)
		return
	}
	process.AddInsertBlockMsg(blockInsert)
}

func (self *BlockGenor) broadcastBlockHandle(result *mc.HD_BroadcastMiningRspMsg) {
	if manversion.VersionCmp(string(result.BlockMainData.Header.Version), manversion.VersionAIMine) < 0 {
		log.Trace(self.logExtraInfo(), "广播矿工挖矿结果消息处理", "版本小于指定版本，不处理", "msg version", string(result.BlockMainData.Header.Version), "指定version", manversion.VersionAIMine)
		return
	}

	number := result.BlockMainData.Header.Number.Uint64()
	log.Info(self.logExtraInfo(), "广播矿工挖矿结果消息处理", "开始", "高度", number, "交易数量", len(types.GetTX(result.BlockMainData.Txs)), "from", result.From.Hex())
	process, err := self.pm.GetProcess(number)
	if err != nil {
		log.Info(self.logExtraInfo(), "矿工挖矿结果消息 获取Process失败", err)
		return
	}
	process.AddBroadcastBlockResult(result)
}

func (self *BlockGenor) handleRecoveryMsg(msg *mc.RecoveryStateMsg) {
	if nil == msg || nil == msg.Header {
		log.Error(self.logExtraInfo(), "状态恢复消息", "消息为nil")
		return
	}
	if manversion.VersionCmp(string(msg.Header.Version), manversion.VersionAIMine) < 0 {
		return
	}
	if msg.Type != mc.RecoveryTypeFullHeader {
		log.Warn(self.logExtraInfo(), "状态恢复消息", "类型不是恢复区块，忽略消息")
		return
	}

	number := msg.Header.Number.Uint64()
	process, err := self.pm.GetProcess(number)
	if err != nil {
		log.Warn(self.logExtraInfo(), "状态恢复消息", "获取Process失败", "err", err)
		return
	}
	process.ProcessRecoveryMsg(msg)
}

func (self *BlockGenor) handleFullBlockReqMsg(req *mc.HD_V2_FullBlockReqMsg) {
	if nil == req {
		log.Error(self.logExtraInfo(), "完整区块请求消息", "消息为nil")
		return
	}

	log.Info(self.logExtraInfo(), "完整区块请求消息", "开始", "高度", req.Number)
	defer log.Debug(self.logExtraInfo(), "完整区块请求消息", "结束", "高度", req.Number)
	process, err := self.pm.GetProcess(req.Number)
	if err != nil {
		log.Warn(self.logExtraInfo(), "完整区块请求消息", "获取Process失败", "err", err)
		return
	}

	process.ProcessFullBlockReq(req)
}

func (self *BlockGenor) handleFullBlockRspMsg(rsp *mc.HD_V2_FullBlockRspMsg) {
	if nil == rsp || nil == rsp.Header {
		log.Error(self.logExtraInfo(), "完整区块响应消息", "消息为nil")
		return
	}

	number := rsp.Header.Number.Uint64()

	log.Info(self.logExtraInfo(), "完整区块响应消息", "开始", "高度", number)
	defer log.Debug(self.logExtraInfo(), "完整区块响应消息", "结束", "高度", number)
	process, err := self.pm.GetProcess(number)
	if err != nil {
		log.Warn(self.logExtraInfo(), "完整区块响应消息", "获取Process失败", "err", err)
		return
	}

	process.ProcessFullBlockRsp(rsp)
}

func (self *BlockGenor) logExtraInfo() string {
	return "区块生成 2.0"
}
