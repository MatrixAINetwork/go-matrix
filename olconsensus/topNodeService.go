// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package olconsensus

import (
	"errors"

	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/consensus"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/event"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
)

var (
	voteFailed        = errors.New("Vote error")
	topologyValidator = 10
)

type TopNodeService struct {
	stateMap       *topNodeState
	msgCheck       messageCheck
	dposRing       *DPosVoteRing
	dposResultRing *DPosVoteRing

	validatorReader consensus.ValidatorReader
	topNodeState    TopNodeStateInterface
	validatorSign   ValidatorAccountInterface
	msgSender       MessageSendInterface
	msgCenter       MessageCenterInterface
	cd              consensus.DPOSEngine

	roleUpdateCh       chan *mc.RoleUpdatedMsg
	roleUpdateSub      event.Subscription
	leaderChangeCh     chan *mc.LeaderChangeNotify
	leaderChangeSub    event.Subscription
	consensusReqCh     chan *mc.HD_OnlineConsensusReqs //顶层节点共识请求消息
	consensusReqSub    event.Subscription
	consensusVoteCh    chan *mc.HD_OnlineConsensusVotes //顶层节点共识投票消息
	consensusVoteSub   event.Subscription
	consensusResultCh  chan *mc.HD_OnlineConsensusVoteResultMsg //顶层节点共识结果消息
	consensusResultSub event.Subscription
	quitCh             chan struct{}
	extraInfo          string
	recvCA             bool
	recvLeader         bool
}

func NewTopNodeService(cd consensus.DPOSEngine) *TopNodeService {
	t := &TopNodeService{
		//stateMap:          newTopNodeState(64),
		msgCheck:          messageCheck{},
		dposRing:          NewDPosVoteRing(64),
		dposResultRing:    NewDPosVoteRing(32),
		cd:                cd,
		roleUpdateCh:      make(chan *mc.RoleUpdatedMsg, 5),
		leaderChangeCh:    make(chan *mc.LeaderChangeNotify, 5),
		consensusReqCh:    make(chan *mc.HD_OnlineConsensusReqs, 5),
		consensusVoteCh:   make(chan *mc.HD_OnlineConsensusVotes, 5),
		consensusResultCh: make(chan *mc.HD_OnlineConsensusVoteResultMsg, 5),
		quitCh:            make(chan struct{}, 2),
		extraInfo:         "TopnodeOnline",
		recvCA:            false,
		recvLeader:        false,
	}
	//	go t.update()

	t.stateMap = newTopNodeState(64, t.extraInfo)
	return t
}

func (self *TopNodeService) SetValidatorReader(reader consensus.ValidatorReader) {
	self.validatorReader = reader
}

func (self *TopNodeService) SetTopNodeStateInterface(inter TopNodeStateInterface) {
	self.topNodeState = inter
}

func (self *TopNodeService) SetValidatorAccountInterface(inter ValidatorAccountInterface) {
	self.validatorSign = inter
}

func (self *TopNodeService) SetMessageSendInterface(inter MessageSendInterface) {
	self.msgSender = inter
}

func (self *TopNodeService) SetMessageCenterInterface(inter MessageCenterInterface) {
	self.msgCenter = inter
}

func (self *TopNodeService) Start() error {
	err := self.subMsg()
	if err != nil {
		return err
	}

	go self.update()
	return nil
}

func (self *TopNodeService) subMsg() error {
	var err error

	self.roleUpdateSub, err = mc.SubscribeEvent(mc.CA_RoleUpdated, self.roleUpdateCh) //身份到达
	if err != nil {
		log.Error(self.extraInfo, "身份更新订阅失败", err)
		return err
	}
	//订阅leader变化消息
	if self.leaderChangeSub, err = self.msgCenter.SubscribeEvent(mc.Leader_LeaderChangeNotify, self.leaderChangeCh); err != nil {
		log.Error(self.extraInfo, "SubscribeEvent LeaderChangeNotify failed.", err)
		return err
	}
	//订阅顶层节点状态共识请求消息
	if self.consensusReqSub, err = self.msgCenter.SubscribeEvent(mc.HD_TopNodeConsensusReq, self.consensusReqCh); err != nil {
		log.Error(self.extraInfo, "SubscribeEvent HD_TopNodeConsensusReq failed.", err)
		return err
	}
	//订共识投票消息
	if self.consensusVoteSub, err = self.msgCenter.SubscribeEvent(mc.HD_TopNodeConsensusVote, self.consensusVoteCh); err != nil {
		log.Error(self.extraInfo, "SubscribeEvent HD_TopNodeConsensusVote failed.", err)
		return err
	}
	//订阅共识结果消息
	if self.consensusResultSub, err = self.msgCenter.SubscribeEvent(mc.HD_TopNodeConsensusVoteResult, self.consensusResultCh); err != nil {
		log.Error(self.extraInfo, "SubscribeEvent HD_TopNodeConsensusVoteResult failed.", err)
		return err
	}

	log.Debug(self.extraInfo, "服务订阅完成", "")
	return nil
}

func (self *TopNodeService) unSubMsg() {
	log.Debug(self.extraInfo, "开始取消服务订阅", "")

	self.roleUpdateSub.Unsubscribe()
	//取消订阅leader变化消息

	self.leaderChangeSub.Unsubscribe()

	//取消订阅顶层节点状态共识请求消息

	self.consensusReqSub.Unsubscribe()
	//取消订共识投票消息

	self.consensusVoteSub.Unsubscribe()
	//取消订阅共识结果消息

	self.consensusResultSub.Unsubscribe()
	log.Debug(self.extraInfo, "取消服务订阅完成", "")

}

func (serv *TopNodeService) update() {
	defer serv.unSubMsg()
	for {
		select {
		case data := <-serv.roleUpdateCh:
			if !data.BlockHash.Equal(serv.msgCheck.getBlockHash()) {
				if serv.msgCheck.checkBlockHash(data.BlockHash) {
					log.Debug(serv.extraInfo, "收到CA通知消息", "", "块高", data.BlockNum)

					if serv.recvLeader {
						log.Info(serv.extraInfo, "已经收到Leader变更消息", "进行顶层节点在线状态共识", "块高", data.BlockNum)

						go serv.LeaderChangeNotifyHandler(data.Leader)
					} else {
						log.Debug(serv.extraInfo, "等待Leader变更消息", "", "块高", data.BlockNum)
					}
					serv.recvCA = true
				}
			}
		case data := <-serv.leaderChangeCh:
			if !data.Leader.Equal(serv.msgCheck.getLeader()) {
				if serv.msgCheck.checkLeaderChangeNotify(data) {
					log.Debug(serv.extraInfo, "收到leader变更通知消息", "", "块高", data.Number)
					if serv.recvCA {
						log.Info(serv.extraInfo, "已经收到CA通知消息", "进行顶层节点在线状态共识", "块高", data.Number)

						go serv.LeaderChangeNotifyHandler(data.Leader)
					} else {
						log.Debug(serv.extraInfo, "等待CA通知消息", "", "块高", data.Number)
					}
					serv.recvLeader = true
				}
			}
		case data := <-serv.consensusReqCh:
			log.Info(serv.extraInfo, "收到共识请求消息", "", "from", data.From.String())
			go serv.consensusReqMsgHandler(data.ReqList)

		case data := <-serv.consensusVoteCh:
			log.Info(serv.extraInfo, "收到共识投票消息", "")
			go serv.consensusVoteMsgHandler(data.Votes)
		case data := <-serv.consensusResultCh:
			log.Info(serv.extraInfo, "收到共识结果消息", "")
			go serv.OnlineConsensusVoteResultMsgHandler(data)
		case <-serv.quitCh:
			log.Info(serv.extraInfo, "收到退出消息", "")
			return
		}
	}
}
func (self *TopNodeService) LeaderChangeNotifyHandler(leader common.Address) {
	if leader.Equal(common.Address{}) {
		log.Error(self.extraInfo, "leader变更消息", "空消息")
		return
	}

	if self.validatorSign.IsSelfAddress(leader) {
		log.Info(self.extraInfo, "我是leader", "准备检查顶层节点在线状态")

		self.checkTopNodeState()
	} else {
		for _, item := range self.dposRing.DPosVoteS {
			go self.consensusVotes(item.getVotes())
		}
	}
}

func (serv *TopNodeService) getTopNodeState() (online, offline []common.Address) {
	return serv.stateMap.newTopNodeState(serv.topNodeState.GetTopNodeOnlineState(), serv.msgCheck.getLeader())
}
func (serv *TopNodeService) checkTopNodeState() {

	serv.sendRequest(serv.getTopNodeState())
}
func (serv *TopNodeService) sendRequest(online, offline []common.Address) {
	leader := ca.GetAddress()
	reqMsg := mc.HD_OnlineConsensusReqs{}
	turn := serv.msgCheck.getRound()
	for _, item := range online {
		val := mc.OnlineConsensusReq{
			OnlineState: onLine,
			Leader:      leader,
			Node:        item,
			Seq:         turn,
		}
		reqMsg.ReqList = append(reqMsg.ReqList, &val)
	}
	for _, item := range offline {
		val := mc.OnlineConsensusReq{
			OnlineState: offLine,
			Leader:      leader,
			Node:        item,
			Seq:         turn,
		}
		reqMsg.ReqList = append(reqMsg.ReqList, &val)
	}
	if len(reqMsg.ReqList) > 0 {
		log.Info(serv.extraInfo, "发送共识投票请求", "start", "轮次", turn, "共识数量", len(reqMsg.ReqList))
		serv.msgSender.SendNodeMsg(mc.HD_TopNodeConsensusReq, &reqMsg, common.RoleValidator, nil)
		go func() {
			serv.consensusReqCh <- &reqMsg
		}()
	}
}

func (serv *TopNodeService) consensusReqMsgHandler(requests []*mc.OnlineConsensusReq) {

	if requests == nil || len(requests) == 0 {
		log.Error(serv.extraInfo, "invalid parameter", "")
		return
	}
	var votes mc.HD_OnlineConsensusVotes
	log.Info(serv.extraInfo, "处理共识请求", "开始")
	for _, item := range requests {
		if serv.msgCheck.checkOnlineConsensusReq(item) {
			if serv.dposRing.addProposal(types.RlpHash(item), item) {
				sign, reqHash, err := serv.voteToReq(item)
				if err == nil {
					vote := mc.HD_ConsensusVote{}
					vote.SignHash.Set(reqHash)
					vote.Sign.Set(sign)
					vote.Round = item.Seq
					votes.Votes = append(votes.Votes, vote)
				} else {
					log.Error(serv.extraInfo, "处理共识请求", "", "error", err)
				}

			} else {
				log.Debug(serv.extraInfo, "处理共识请求", "", "addProposal", "false", "item", item.Node.String(), "seq", item.Seq, "leader", item.Leader.String())
			}
		} else {
			log.Debug(serv.extraInfo, "处理共识请求", "", "checkOnlineConsensusReq", "false", "node", item.Node.String(), "轮次", item.Seq)
		}
	}
	if len(votes.Votes) > 0 {
		log.Info(serv.extraInfo, "处理共识请求", "", "完成投票", "发送共识投票消息")
		serv.msgSender.SendNodeMsg(mc.HD_TopNodeConsensusVote, &votes, common.RoleValidator, nil)
		go func() {
			serv.consensusVoteCh <- &votes
		}()
	}
}
func (serv *TopNodeService) consensusVoteMsgHandler(msg []mc.HD_ConsensusVote) {
	if msg == nil || len(msg) == 0 {
		log.Error(serv.extraInfo, "处理共识投票", "", "invalid parameter", "", "len(msg)", len(msg))
		return
	}

	for _, item := range msg {
		serv.consensusVotes(serv.dposRing.addVote(item.SignHash, &item))
	}
}
func (serv *TopNodeService) OnlineConsensusVoteResultMsgHandler(msg *mc.HD_OnlineConsensusVoteResultMsg) {
	tempSigns, err := serv.cd.VerifyHash(serv.validatorReader, types.RlpHash(msg.Req), msg.SignList)

	if err != nil {
		log.Error(serv.extraInfo, "处理共识投票结果", err)
	} else {
		log.Info(serv.extraInfo, "处理共识投票结果", "sucess", "状态", msg.Req.OnlineState, "投票数", len(tempSigns))
		finishHash := getFinishedPropocalHash(msg.Req.Node, uint8(msg.Req.OnlineState))
		var vote = new(mc.HD_ConsensusVote)
		vote.SignHash = finishHash
		vote.Round = 0
		vote.Sign = common.Signature{}
		vote.From = msg.From
		proposal, voteInfo := serv.dposResultRing.addVote(finishHash, vote)
		for _, value := range voteInfo {
			log.Debug(serv.extraInfo, "处理共识投票结果", "", "投票from", value.data.From.String())

		}

		if serv.checkPosVoteResults(proposal, voteInfo) {
			log.Info(serv.extraInfo, "处理共识投票结果", "共识成功", "node", msg.Req.Node.String(), "在线状态", msg.Req.OnlineState)
			serv.stateMap.saveConsensusNodeState(msg.Req.Node, OnlineState(msg.Req.OnlineState))
		} else {
			log.Info(serv.extraInfo, "处理共识投票结果", "共识失败", "node", msg.Req.Node.String(), "在线状态",
				msg.Req.OnlineState, "投票数", len(voteInfo))

		}
	}
}
func (serv *TopNodeService) checkPosVoteResults(proposal interface{}, votes []voteInfo) bool {
	if votes == nil || len(votes) == 0 {
		log.Error(serv.extraInfo, "处理共识投票结果", "检查投票结果", "收到的投票结果", len(votes))
		return false
	}
	validators := make([]common.Address, 0)
	for _, value := range votes {
		validators = append(validators, value.data.From)
	}
	log.Info(serv.extraInfo, "处理共识投票结果", "检查投票结果", "收到的投票结果数量", len(validators))

	return serv.cd.VerifyStocksWithBlock(serv.validatorReader, validators, serv.msgCheck.getBlockHash())
}

func (serv *TopNodeService) consensusVotes(proposal interface{}, votes []voteInfo) {
	if proposal == nil || votes == nil || len(votes) == 0 {
		return
	}
	log.Info(serv.extraInfo, "处理共识投票", "开始")

	prop := proposal.(*mc.OnlineConsensusReq)
	if serv.msgCheck.getLeader() != prop.Leader {
		log.Error(serv.extraInfo, "处理共识投票", "leader无效", "invalid Leader", prop.Leader, "leader", serv.msgCheck.getLeader())
		return
	}
	if !serv.msgCheck.checkRound(prop.Seq) {
		log.Error(serv.extraInfo, "处理共识投票", "轮次无效", "invalid Round", prop.Seq, "Round", serv.msgCheck.getRound())
		return
	}
	signList := make([]common.Signature, 0)
	for _, value := range votes {
		signList = append(signList, value.data.Sign)
	}
	tempSigns, err := serv.cd.VerifyHash(serv.validatorReader, votes[0].data.SignHash, signList)
	if err != nil {
		log.Error(serv.extraInfo, "处理共识投票", "共识失败", "投票数", len(tempSigns), "err", err)
		return
	}
	log.Info(serv.extraInfo, "处理共识投票", "共识成功", "节点", prop.Node.String(), "状态", prop.OnlineState,
		"投票数", len(tempSigns), "finishHash", getFinishedPropocalHash(prop.Node, uint8(prop.OnlineState)))
	serv.stateMap.finishedProposal.addProposal(getFinishedPropocalHash(prop.Node, uint8(prop.OnlineState)), proposal)
	//send DPos Success message
	result := mc.HD_OnlineConsensusVoteResultMsg{
		Req:      prop,
		SignList: signList,
		From:     ca.GetAddress(),
	}

	serv.msgSender.SendNodeMsg(mc.HD_TopNodeConsensusVoteResult, &result, common.RoleValidator, nil)
	go func() {
		serv.consensusResultCh <- &result
	}()

}

func (serv *TopNodeService) voteToReq(tempReq *mc.OnlineConsensusReq) (common.Signature, common.Hash, error) {
	var sign common.Signature
	var err error
	var ok bool

	if tempReq.Seq == 0 || tempReq.Node.Equal(common.Address{}) || tempReq.Leader.Equal(common.Address{}) {
		log.Error(serv.extraInfo, "处理共识请求", "对共识请求进行投票", "无效的参数", "", "轮次", tempReq.Seq, "leader", tempReq.Leader.String(),
			"请求共识的节点", tempReq.Node.String())
		return common.Signature{}, common.Hash{}, voteFailed
	}
	reqHash := types.RlpHash(tempReq)
	log.Info(serv.extraInfo, "处理共识请求", "对共识请求进行投票", "轮次", tempReq.Seq, "node", tempReq.Node,
		"状态", tempReq.OnlineState, "leader", tempReq.Leader)
	if tempReq.OnlineState == onLine {
		ok = serv.stateMap.checkNodeOnline(tempReq.Node, serv.topNodeState.GetTopNodeOnlineState())
		log.Info(serv.extraInfo, "处理共识请求", "对共识请求进行投票", "检查状态", "在线", "node", tempReq.Node.String(), "ok", ok)
	} else {
		ok = serv.stateMap.checkNodeOffline(tempReq.Node, serv.topNodeState.GetTopNodeOnlineState())
		log.Info(serv.extraInfo, "处理共识请求", "对共识请求进行投票", "检查状态", "离线", "node", tempReq.Node.String(), "ok", ok)

	}
	if ok {
		//投赞成票
		sign, err = serv.validatorSign.SignWithValidate(reqHash.Bytes(), true)
		if err != nil {
			log.Error(serv.extraInfo, "处理共识请求", "对共识请求进行投票", "投票失败", err)
			return common.Signature{}, common.Hash{}, voteFailed
		}
		log.Info(serv.extraInfo, "处理共识请求", "对共识请求进行投票", "投赞成票", "", "reqNode", tempReq.Node.String())
	} else {
		//投反对票
		sign, err = serv.validatorSign.SignWithValidate(reqHash.Bytes(), false)
		if err != nil {
			log.Error(serv.extraInfo, "处理共识请求", "对共识请求进行投票", "投票失败", err)
			return common.Signature{}, common.Hash{}, voteFailed
		}
		log.Info(serv.extraInfo, "处理共识请求", "对共识请求进行投票", "投反对票", "", "reqNode", tempReq.Node.String())

	}
	log.Info(serv.extraInfo, "处理共识请求", "对共识请求进行投票", "保存共识请求", "", "reqhash", reqHash.String())
	return sign, reqHash, nil
}

//set ElectNodes
func (serv *TopNodeService) SetElectNodes(nodes []common.Address, height uint64) {
	serv.stateMap.setElectNodes(nodes, height)
}

//设置当前区块的顶点拓扑图，每个验证者都调用
func (serv *TopNodeService) SetCurrentOnlineState(onLineNode, onElectNode []common.Address) {
	//	offLineNode1 := append(offLineNode, offElectNode...)
	serv.stateMap.setCurrentTopNodeState(onLineNode, onElectNode)
}

//提供换届服务获取当前经过共识的在线状态，只有leader调用
func (serv *TopNodeService) GetConsensusOnlineState() (ret_offLineNode, ret_onElectNode, ret_offElectNode []common.Address) {
	ret_offLineNode, ret_onElectNode, ret_offElectNode = serv.stateMap.getCurrentTopNodeChange()
	log.Info(serv.extraInfo, "区块生成", "获取经过共识的在线状态", "offline长度", len(ret_offLineNode), "online长度", len(ret_onElectNode), "offlineElect长度", len(ret_offElectNode))

	return
}

//verify Online State
func (serv *TopNodeService) CheckAddressConsensusOnlineState(offLineNode, onElectNode, offElectNode []common.Address) bool {
	offLineNode1 := append(offLineNode, offElectNode...)
	check := true
	for _, item := range onElectNode {
		if !serv.stateMap.checkAddressConsesusOnlineState(item, onLine) {
			check = false
			log.Info(serv.extraInfo, "区块验证", "", "检查在线状态", check)
			break
		}
	}

	if check {
		for _, item := range offLineNode1 {
			if !serv.stateMap.checkAddressConsesusOnlineState(item, offLine) {
				check = false
				log.Info(serv.extraInfo, "区块验证", "", "检查离线状态", check)
				break
			}
		}

	}
	return check
}
