// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package olconsensus

import (
	"errors"
	"reflect"

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
	stateMap *topNodeState
	msgCheck messageCheck
	dposRing *DPosVoteRing

	validatorReader consensus.ValidatorReader
	topNodeState    TopNodeStateInterface
	validatorSign   ValidatorAccountInterface
	msgSender       MessageSendInterface
	msgCenter       MessageCenterInterface
	cd              consensus.DPOSEngine

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
}

func NewTopNodeService(cd consensus.DPOSEngine) *TopNodeService {
	t := &TopNodeService{
		stateMap:          newTopNodeState(64),
		msgCheck:          messageCheck{},
		dposRing:          NewDPosVoteRing(64),
		cd:                cd,
		leaderChangeCh:    make(chan *mc.LeaderChangeNotify, 5),
		consensusReqCh:    make(chan *mc.HD_OnlineConsensusReqs, 5),
		consensusVoteCh:   make(chan *mc.HD_OnlineConsensusVotes, 5),
		consensusResultCh: make(chan *mc.HD_OnlineConsensusVoteResultMsg, 5),
		quitCh:            make(chan struct{}, 2),
		extraInfo:         "TopnodeOnline",
	}
	//	go t.update()

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

	log.Info(self.extraInfo, "服务订阅完成", "")
	return nil
}

func (self *TopNodeService) unSubMsg() {
	log.Info(self.extraInfo, "开始取消服务订阅", "")
	//取消订阅leader变化消息

	self.leaderChangeSub.Unsubscribe()

	//取消订阅顶层节点状态共识请求消息

	self.consensusReqSub.Unsubscribe()
	//取消订共识投票消息

	self.consensusVoteSub.Unsubscribe()
	//取消订阅共识结果消息

	self.consensusResultSub.Unsubscribe()
	log.Info(self.extraInfo, "取消服务订阅完成", "")

}

func (serv *TopNodeService) update() {
	log.Info(serv.extraInfo, "启动顶层节点服务，等待接收消息", "")
	defer serv.unSubMsg()
	for {
		select {

		case data := <-serv.leaderChangeCh:
			if serv.msgCheck.checkLeaderChangeNotify(data) {
				log.Info(serv.extraInfo, "收到leader变更通知消息", "")
				go serv.LeaderChangeNotifyHandler(data)
			}

		case data := <-serv.consensusReqCh:
			log.Info(serv.extraInfo, "收到共识请求消息", "")
			go serv.consensusReqMsgHandler(data.ReqList)

		case data := <-serv.consensusVoteCh:
			log.Info(serv.extraInfo, "收到共识投票消息", "")
			go serv.consensusVoteMsgHandler(data.Votes)
			/*
				case data := <-serv.consensusResultCh:
					log.Info(serv.extraInfo, "收到共识结果消息", "")
					go serv.OnlineConsensusVoteResultMsgHandler(data)
			*/
		case <-serv.quitCh:
			log.Info(serv.extraInfo, "收到退出消息", "")
			return
		}
	}
}
func (self *TopNodeService) LeaderChangeNotifyHandler(msg *mc.LeaderChangeNotify) {
	if msg == nil {
		log.Error(self.extraInfo, "leader变更消息", "空消息")
		return
	}

	if self.validatorSign.IsSelfAddress(msg.Leader) {
		self.checkTopNodeState()
	} else {
		for _, item := range self.dposRing.DPosVoteS {
			go self.consensusVotes(item.getVotes())
		}
	}
}

func (serv *TopNodeService) getTopNodeState() (online, offline []common.Address) {
	return serv.stateMap.newTopNodeState(serv.topNodeState.GetTopNodeOnlineState())
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
		serv.msgSender.SendNodeMsg(mc.HD_TopNodeConsensusReq, &reqMsg, common.RoleValidator, nil)
	}
}

func (serv *TopNodeService) consensusReqMsgHandler(requests []*mc.OnlineConsensusReq) {

	if requests == nil || len(requests) == 0 {
		log.Error(serv.extraInfo, "invalid parameter", "")
		return
	}
	var votes mc.HD_OnlineConsensusVotes
	log.Info(serv.extraInfo, "开始投票", "")
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
					log.Error(serv.extraInfo, "error", err)
				}

			}
		}
	}
	serv.msgSender.SendNodeMsg(mc.HD_TopNodeConsensusVote, &votes, common.RoleValidator, nil)
	log.Info("test info", "type", reflect.TypeOf(votes))
}
func (serv *TopNodeService) consensusVoteMsgHandler(msg []mc.HD_ConsensusVote) {
	if msg == nil || len(msg) == 0 {
		log.Error(serv.extraInfo, "invalid parameter", "")
		return
	}

	for _, item := range msg {
		serv.consensusVotes(serv.dposRing.addVote(item.SignHash, &item))
	}
}
func (serv *TopNodeService) consensusVotes(proposal interface{}, votes []voteInfo) {
	if proposal == nil || votes == nil || len(votes) == 0 {
		return
	}
	prop := proposal.(*mc.OnlineConsensusReq)
	if serv.msgCheck.getLeader() != prop.Leader {
		log.Info(serv.extraInfo, "invalid Leader", prop.Leader)
		return
	}
	if !serv.msgCheck.checkRound(prop.Seq) {
		log.Info(serv.extraInfo, "invalid Round", prop.Seq)
	}
	signList := make([]common.Signature, 0)
	for _, value := range votes {
		signList = append(signList, value.data.Sign)
	}
	tempSigns, err := serv.cd.VerifyHashWithNumber(serv.validatorReader, votes[0].data.SignHash, signList, 10)
	if err != nil {
		log.Info(serv.extraInfo, "DPOS共识失败", err)
		return
	}
	log.Error(serv.extraInfo, "处理共识投票消息", "DPOS共识成功", "投票数", len(tempSigns))
	serv.stateMap.finishedProposal.addProposal(getFinishedPropocalHash(prop.Node, uint8(prop.OnlineState)), proposal)
}

func (serv *TopNodeService) voteToReq(tempReq *mc.OnlineConsensusReq) (common.Signature, common.Hash, error) {
	var sign common.Signature
	var err error

	reqHash := types.RlpHash(tempReq)

	var ok bool
	if tempReq.OnlineState == onLine {
		ok = serv.stateMap.checkNodeOnline(tempReq.Node, serv.topNodeState.GetTopNodeOnlineState())
	} else {
		ok = serv.stateMap.checkNodeOffline(tempReq.Node, serv.topNodeState.GetTopNodeOnlineState())
	}
	if ok {
		//投赞成票
		sign, err = serv.validatorSign.SignWithValidate(reqHash.Bytes(), true)
		if err != nil {
			log.Info(serv.extraInfo, "Vote failed:", err)
			return common.Signature{}, common.Hash{}, voteFailed
		}
		log.Info(serv.extraInfo, "投赞成票", "", "reqNode", tempReq.Node)
	} else {
		//投反对票
		sign, err = serv.validatorSign.SignWithValidate(reqHash.Bytes(), false)
		if err != nil {
			log.Info(serv.extraInfo, "Vote failed:", err)
			return common.Signature{}, common.Hash{}, voteFailed
		}
		log.Info(serv.extraInfo, "投反对票", "", "reqNode", tempReq.Node)

	}
	log.Info(serv.extraInfo, "保存共识请求", "", "reqhash", reqHash)
	//	p.consensusReqCache[reqHash] = tempReq
	return sign, reqHash, nil
}

//提供换届服务获取当前经过共识的在线状态
func (serv *TopNodeService) GetConsensusOnlineState() (map[common.Address]OnlineState, map[common.Address]OnlineState) {
	/*if req == common.RoleMiner {
	   return getMinerOnlineState()
	} else if req == common.RoleValidator {
	   return getValidatorOnlineState()
	} else {
	   return nil
	}*/
	// 返回两个map，
	// 第一个map = currentRole是验证者和矿工的在线共识状态
	// 第二个map = originalRole是验证者和矿工，但currentRole不是验证者或矿工的在线共识状态
	return nil, nil
}
