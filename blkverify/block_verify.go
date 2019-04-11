// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package blkverify

import (
	"github.com/MatrixAINetwork/go-matrix/accounts/signhelper"
	"github.com/MatrixAINetwork/go-matrix/baseinterface"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/consensus/blkmanage"
	"github.com/MatrixAINetwork/go-matrix/core"
	"github.com/MatrixAINetwork/go-matrix/event"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mandb"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/msgsend"
	"github.com/MatrixAINetwork/go-matrix/reelection"
)

type Matrix interface {
	HD() *msgsend.HD
	BlockChain() *core.BlockChain
	TxPool() *core.TxPoolManager //Y
	SignHelper() *signhelper.SignHelper
	ReElection() *reelection.ReElection
	EventMux() *event.TypeMux
	Random() *baseinterface.Random
	ChainDb() mandb.Database
	ManBlkDeal() *blkmanage.ManBlkManage
}

type BlockVerify struct {
	quitCh               chan struct{}
	processManage        *ProcessManage
	roleUpdatedMsgCh     chan *mc.RoleUpdatedMsg
	leaderChangeNotifyCh chan *mc.LeaderChangeNotify
	requestCh            chan *mc.HD_BlkConsensusReqMsg
	localVerifyReqCh     chan *mc.LocalBlockVerifyConsensusReq
	voteMsgCh            chan *mc.HD_ConsensusVote
	recoveryCh           chan *mc.RecoveryStateMsg
	roleUpdatedMsgSub    event.Subscription
	leaderChangeSub      event.Subscription
	requestSub           event.Subscription
	localVerifyReqSub    event.Subscription
	voteMsgSub           event.Subscription
	recoverySub          event.Subscription
}

func NewBlockVerify(matrix Matrix) (*BlockVerify, error) {
	server := &BlockVerify{
		quitCh:               make(chan struct{}),
		roleUpdatedMsgCh:     make(chan *mc.RoleUpdatedMsg, 1),
		leaderChangeNotifyCh: make(chan *mc.LeaderChangeNotify, 1),
		requestCh:            make(chan *mc.HD_BlkConsensusReqMsg, 1),
		localVerifyReqCh:     make(chan *mc.LocalBlockVerifyConsensusReq, 1),
		voteMsgCh:            make(chan *mc.HD_ConsensusVote, 1),
		recoveryCh:           make(chan *mc.RecoveryStateMsg, 1),
		roleUpdatedMsgSub:    nil,
		leaderChangeSub:      nil,
		requestSub:           nil,
		localVerifyReqSub:    nil,
		voteMsgSub:           nil,
	}

	server.processManage = NewProcessManage(matrix)

	var err error
	if server.roleUpdatedMsgSub, err = mc.SubscribeEvent(mc.CA_RoleUpdated, server.roleUpdatedMsgCh); err != nil {
		return nil, err
	}
	if server.leaderChangeSub, err = mc.SubscribeEvent(mc.Leader_LeaderChangeNotify, server.leaderChangeNotifyCh); err != nil {
		return nil, err
	}
	if server.requestSub, err = mc.SubscribeEvent(mc.HD_BlkConsensusReq, server.requestCh); err != nil {
		return nil, err
	}
	if server.localVerifyReqSub, err = mc.SubscribeEvent(mc.BlockGenor_HeaderVerifyReq, server.localVerifyReqCh); err != nil {
		return nil, err
	}
	if server.voteMsgSub, err = mc.SubscribeEvent(mc.HD_BlkConsensusVote, server.voteMsgCh); err != nil {
		return nil, err
	}
	if server.recoverySub, err = mc.SubscribeEvent(mc.Leader_RecoveryState, server.recoveryCh); err != nil {
		return nil, err
	}

	go server.update()
	return server, nil
}

func (self *BlockVerify) Close() {
	close(self.quitCh)
}

func (self *BlockVerify) update() {
	defer func() {
		self.voteMsgSub.Unsubscribe()
		self.localVerifyReqSub.Unsubscribe()
		self.requestSub.Unsubscribe()
		self.leaderChangeSub.Unsubscribe()
		self.roleUpdatedMsgSub.Unsubscribe()
	}()

	// 启动时，先恢复DB中验证过的缓存
	self.reloadVerifiedBlocks()

	for {
		select {
		case roleMsg := <-self.roleUpdatedMsgCh:
			go self.handleRoleUpdatedMsg(roleMsg)

		case leaderMsg := <-self.leaderChangeNotifyCh:
			go self.handleLeaderChangeNotify(leaderMsg)

		case blkVerifyReq := <-self.requestCh:
			go self.handleRequestMsg(blkVerifyReq)

		case localVerifyReq := <-self.localVerifyReqCh:
			go self.handleLocalRequestMsg(localVerifyReq)

		case voteMsg := <-self.voteMsgCh:
			go self.handleVoteMsg(voteMsg)

		case recoveryMsg := <-self.recoveryCh:
			go self.handleRecoveryMsg(recoveryMsg)

		case <-self.quitCh:
			self.processManage.clearProcessMap()
			return
		}
	}
}

func (self *BlockVerify) handleRoleUpdatedMsg(roleMsg *mc.RoleUpdatedMsg) {
	if nil == roleMsg {
		log.Error(self.logExtraInfo(), "CA身份消息异常", "消息为nil")
		return
	}
	log.Debug(self.logExtraInfo(), "CA身份消息", "开始处理", "高度", roleMsg.BlockNum, "角色", roleMsg.Role.String(), "区块hash", roleMsg.BlockHash.TerminalString())

	curNumber := roleMsg.BlockNum + 1
	self.processManage.SetCurNumber(curNumber, roleMsg.SuperSeq)
	if roleMsg.Role == common.RoleValidator || roleMsg.Role == common.RoleBroadcast {
		curProcess := self.processManage.GetCurrentProcess()
		curProcess.StartRunning(roleMsg.Role)
	}

	return
}

func (self *BlockVerify) handleLeaderChangeNotify(leaderMsg *mc.LeaderChangeNotify) {
	if nil == leaderMsg {
		log.Error(self.logExtraInfo(), "leader变更消息异常", "消息为nil")
		return
	}
	log.Debug(self.logExtraInfo(), "Leader变更消息", "开始处理", "高度", leaderMsg.Number, "共识轮次",
		leaderMsg.ConsensusTurn.String(), "共识状态", leaderMsg.ConsensusState, "leader", leaderMsg.Leader.Hex(), "next leader", leaderMsg.NextLeader.Hex())

	msgNumber := leaderMsg.Number
	process, err := self.processManage.GetProcess(msgNumber)
	if err != nil {
		log.INFO(self.logExtraInfo(), "Leader变更消息 获取Process失败", err)
		return
	}

	process.SetLeaderInfo(leaderMsg)
}

func (self *BlockVerify) handleRequestMsg(reqMsg *mc.HD_BlkConsensusReqMsg) {
	if nil == reqMsg || nil == reqMsg.Header {
		log.Warn(self.logExtraInfo(), "区块共识请求消息", "msg is nil")
		return
	}
	log.Debug(self.logExtraInfo(), "区块共识请求消息", "开始处理", "高度", reqMsg.Header.Number, "共识轮次", reqMsg.ConsensusTurn.String(), "Leader", reqMsg.Header.Leader.Hex())
	if (reqMsg.Header.Leader == common.Address{}) {
		log.WARN(self.logExtraInfo(), "请求消息", "leader is nil")
		return
	}
	msgNumber := reqMsg.Header.Number.Uint64()
	process, err := self.processManage.GetProcess(msgNumber)
	if err != nil {
		log.INFO(self.logExtraInfo(), "请求消息 获取Process失败", err)
		return
	}

	process.AddReq(reqMsg)
}

func (self *BlockVerify) handleLocalRequestMsg(localReq *mc.LocalBlockVerifyConsensusReq) {
	if nil == localReq || nil == localReq.BlkVerifyConsensusReq {
		log.WARN(self.logExtraInfo(), "本地请求消息", "msg is nil")
		return
	}
	msgNumber := localReq.BlkVerifyConsensusReq.Header.Number.Uint64()
	log.Trace(self.logExtraInfo(), "本地请求消息处理", "开始", "高度", msgNumber)
	defer log.Trace(self.logExtraInfo(), "本地请求消息处理", "结束", "高度", msgNumber)
	if (localReq.BlkVerifyConsensusReq.Header.Leader == common.Address{}) {
		log.WARN(self.logExtraInfo(), "本地请求消息", "leader is nil")
		return
	}
	process, err := self.processManage.GetProcess(msgNumber)
	if err != nil {
		log.INFO(self.logExtraInfo(), "本地请求消息 获取Process失败", err)
		return
	}

	process.AddLocalReq(localReq)
}

func (self *BlockVerify) handleVoteMsg(voteMsg *mc.HD_ConsensusVote) {
	if nil == voteMsg {
		log.Error(self.logExtraInfo(), "投票消息处理", "消息为nil")
		return
	}

	log.Trace(self.logExtraInfo(), "投票消息处理", "开始", "from", voteMsg.From.Hex(), "signHash", voteMsg.SignHash.TerminalString())
	defer log.Trace(self.logExtraInfo(), "投票消息处理", "结束", "from", voteMsg.From.Hex(), "signHash", voteMsg.SignHash.TerminalString())

	process, err := self.processManage.GetProcess(voteMsg.Number)
	if err != nil {
		log.Debug(self.logExtraInfo(), "本地请求消息 获取Process失败", err)
		return
	}

	process.HandleVote(voteMsg.SignHash, voteMsg.Sign, voteMsg.From)
}

func (self *BlockVerify) handleRecoveryMsg(msg *mc.RecoveryStateMsg) {
	if nil == msg || nil == msg.Header {
		log.Error(self.logExtraInfo(), "状态恢复消息", "消息为nil")
		return
	}
	if msg.Type != mc.RecoveryTypePOS {
		log.Debug(self.logExtraInfo(), "状态恢复消息", "消息类型不是POS回复，忽略消息")
		return
	}

	number := msg.Header.Number.Uint64()
	log.Debug(self.logExtraInfo(), "状态恢复消息", "开始", "高度", number, "leader", msg.Header.Leader.Hex(), "header hash", msg.Header.HashNoSignsAndNonce().TerminalString())
	defer log.Debug(self.logExtraInfo(), "状态恢复消息", "结束", "高度", number, "leader", msg.Header.Leader.Hex())

	curProcess := self.processManage.GetCurrentProcess()
	if curProcess != nil {
		if curProcess.number != number {
			log.INFO(self.logExtraInfo(), "状态恢复消息", "高度不是当前处理高度，忽略消息", "高度", number, "当前高度", curProcess.number)
			return
		}
		curProcess.ProcessRecoveryMsg(msg)
	}
}

func (self *BlockVerify) reloadVerifiedBlocks() {
	blocks, err := readVerifiedBlocksFromDB(self.processManage.chainDB)
	if err != nil {
		log.Info(self.logExtraInfo(), "reloadVerifiedBlocks", "从DB中获取数据错误", "err", err)
		return
	}
	if len(blocks) == 0 {
		log.Info(self.logExtraInfo(), "reloadVerifiedBlocks", "DB中没有数据，不处理")
		return
	}

	curNumber := uint64(0)
	curBlock := self.processManage.bc.CurrentBlock()
	if curBlock != nil {
		curNumber = curBlock.Number().Uint64()
	}

	for _, block := range blocks {
		if block.req.Header.Number.Uint64() <= curNumber {
			log.Info(self.logExtraInfo(), "reloadVerifiedBlocks", "req 高度 <= 当前高度, 不处理", "block number", block.req.Header.Number.Uint64(), "cur number", curNumber)
			continue
		}
		log.Info(self.logExtraInfo(), "reloadVerifiedBlocks", "缓存 verified block", "block number", block.req.Header.Number.Uint64(), "hash", block.hash.TerminalString())
		self.processManage.AddVerifiedBlock(&block)
	}
}

func (self *BlockVerify) logExtraInfo() string {
	return "区块验证服务"
}
