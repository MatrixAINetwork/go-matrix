// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package blkverify

import (
	"github.com/matrix/go-matrix/accounts/signhelper"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core"
	"github.com/matrix/go-matrix/event"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/msgsend"
	"github.com/matrix/go-matrix/reelection"
)

type Matrix interface {
	HD() *msgsend.HD
	BlockChain() *core.BlockChain
	TxPool() *core.TxPoolManager //YYY
	SignHelper() *signhelper.SignHelper
	ReElection() *reelection.ReElection
	EventMux() *event.TypeMux
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
			return
		}
	}
}

func (self *BlockVerify) handleRoleUpdatedMsg(roleMsg *mc.RoleUpdatedMsg) error {
	log.INFO(self.logExtraInfo(), "CA身份消息处理", "开始", "高度", roleMsg.BlockNum, "角色", roleMsg.Role.String())
	defer log.INFO(self.logExtraInfo(), "CA身份消息", "结束", "高度", roleMsg.BlockNum, "角色", roleMsg.Role.String())

	curNumber := roleMsg.BlockNum + 1
	self.processManage.SetCurNumber(curNumber)
	if roleMsg.Role == common.RoleValidator || roleMsg.Role == common.RoleBroadcast {
		curProcess := self.processManage.GetCurrentProcess()
		curProcess.StartRunning(roleMsg.Role)
	}

	return nil
}

func (self *BlockVerify) handleLeaderChangeNotify(leaderMsg *mc.LeaderChangeNotify) {
	if nil == leaderMsg {
		log.ERROR(self.logExtraInfo(), "leader变更消息异常", "消息为nil")
		return
	}
	log.INFO(self.logExtraInfo(), "Leader变更消息处理", "开始", "高度", leaderMsg.Number, "共识轮次",
		leaderMsg.ConsensusTurn, "有效", leaderMsg.ConsensusState, "leader", leaderMsg.Leader.Hex(), "next leader", leaderMsg.NextLeader.Hex())
	defer log.INFO(self.logExtraInfo(), "Leader变更消息处理", "结束", "高度", leaderMsg.Number, "共识轮次", leaderMsg.ConsensusTurn, "有效", leaderMsg.ConsensusState)

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
		log.WARN(self.logExtraInfo(), "请求消息", "msg is nil")
		return
	}
	log.INFO(self.logExtraInfo(), "请求消息处理", "开始", "高度", reqMsg.Header.Number, "共识轮次", reqMsg.ConsensusTurn, "Leader", reqMsg.Header.Leader.Hex())
	defer log.INFO(self.logExtraInfo(), "请求消息处理", "结束", "高度", reqMsg.Header.Number, "共识轮次", reqMsg.ConsensusTurn, "Leader", reqMsg.Header.Leader.Hex())
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
	log.INFO(self.logExtraInfo(), "本地请求消息处理", "开始", "高度", msgNumber)
	defer log.INFO(self.logExtraInfo(), "本地请求消息处理", "结束", "高度", msgNumber)
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
		log.ERROR(self.logExtraInfo(), "投票消息处理", "消息为nil")
		return
	}
	log.INFO(self.logExtraInfo(), "投票消息处理", "开始", "from", voteMsg.From.Hex(), "signHash", voteMsg.SignHash.TerminalString())
	defer log.INFO(self.logExtraInfo(), "投票消息处理", "结束", "from", voteMsg.From.Hex(), "signHash", voteMsg.SignHash.TerminalString())
	if err := self.processManage.AddVoteToPool(voteMsg.SignHash, voteMsg.Sign, voteMsg.From, voteMsg.Number); err != nil {
		log.ERROR(self.logExtraInfo(), "投票消息，加入票池失败", err)
		return
	}

	curProcess := self.processManage.GetCurrentProcess()
	if curProcess != nil {
		curProcess.ProcessDPOSOnce()
	}
}

func (self *BlockVerify) handleRecoveryMsg(msg *mc.RecoveryStateMsg) {
	if nil == msg || nil == msg.Header {
		log.ERROR(self.logExtraInfo(), "状态恢复消息", "消息为nil")
		return
	}
	number := msg.Header.Number.Uint64()
	log.INFO(self.logExtraInfo(), "状态恢复消息", "开始", "类型", msg.Type, "高度", number, "leader", msg.Header.Leader.Hex())
	defer log.INFO(self.logExtraInfo(), "状态恢复消息", "结束", "类型", msg.Type, "高度", number, "leader", msg.Header.Leader.Hex())
	curProcess := self.processManage.GetCurrentProcess()
	if curProcess != nil {
		if curProcess.number != number {
			log.INFO(self.logExtraInfo(), "状态恢复消息", "高度不是当前处理高度，忽略消息", "高度", number, "当前高度", curProcess.number)
			return
		}
		curProcess.ProcessRecoveryMsg(msg)
	}
}

func (self *BlockVerify) logExtraInfo() string {
	return "区块验证服务"
}
