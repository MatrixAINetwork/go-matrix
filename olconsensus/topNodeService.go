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
	"github.com/matrix/go-matrix/params/manparams"
)

var (
	voteFailed = errors.New("Vote error")
)

type TopNodeService struct {
	stateMap *topNodeState
	msgCheck *messageCheck
	dposRing *DPosVoteRing

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
		msgCheck:          newMessageCheck(3),
		dposRing:          NewDPosVoteRing(64),
		cd:                cd,
		roleUpdateCh:      make(chan *mc.RoleUpdatedMsg, 5),
		leaderChangeCh:    make(chan *mc.LeaderChangeNotify, 5),
		consensusReqCh:    make(chan *mc.HD_OnlineConsensusReqs, 5),
		consensusVoteCh:   make(chan *mc.HD_OnlineConsensusVotes, 5),
		consensusResultCh: make(chan *mc.HD_OnlineConsensusVoteResultMsg, 5),
		quitCh:            make(chan struct{}, 2),
		extraInfo:         "TopNodeOnline",
	}
	//	go t.update()

	t.stateMap = newTopNodeState(64, t.extraInfo)
	return t
}

func (serv *TopNodeService) SetValidatorReader(reader consensus.ValidatorReader) {
	serv.validatorReader = reader
}

func (serv *TopNodeService) SetTopNodeStateInterface(inter TopNodeStateInterface) {
	serv.topNodeState = inter
}

func (serv *TopNodeService) SetValidatorAccountInterface(inter ValidatorAccountInterface) {
	serv.validatorSign = inter
}

func (serv *TopNodeService) SetMessageSendInterface(inter MessageSendInterface) {
	serv.msgSender = inter
}

func (serv *TopNodeService) SetMessageCenterInterface(inter MessageCenterInterface) {
	serv.msgCenter = inter
}

func (serv *TopNodeService) Start() error {
	err := serv.subMsg()
	if err != nil {
		return err
	}

	go serv.update()
	return nil
}

func (serv *TopNodeService) subMsg() error {
	var err error

	serv.roleUpdateSub, err = mc.SubscribeEvent(mc.CA_RoleUpdated, serv.roleUpdateCh) //身份到达
	if err != nil {
		log.Error(serv.extraInfo, "身份更新订阅失败", err)
		return err
	}
	//订阅leader变化消息
	if serv.leaderChangeSub, err = serv.msgCenter.SubscribeEvent(mc.Leader_LeaderChangeNotify, serv.leaderChangeCh); err != nil {
		log.Error(serv.extraInfo, "SubscribeEvent LeaderChangeNotify failed.", err)
		return err
	}
	//订阅顶层节点状态共识请求消息
	if serv.consensusReqSub, err = serv.msgCenter.SubscribeEvent(mc.HD_TopNodeConsensusReq, serv.consensusReqCh); err != nil {
		log.Error(serv.extraInfo, "SubscribeEvent HD_TopNodeConsensusReq failed.", err)
		return err
	}
	//订共识投票消息
	if serv.consensusVoteSub, err = serv.msgCenter.SubscribeEvent(mc.HD_TopNodeConsensusVote, serv.consensusVoteCh); err != nil {
		log.Error(serv.extraInfo, "SubscribeEvent HD_TopNodeConsensusVote failed.", err)
		return err
	}
	//订阅共识结果消息
	if serv.consensusResultSub, err = serv.msgCenter.SubscribeEvent(mc.HD_TopNodeConsensusVoteResult, serv.consensusResultCh); err != nil {
		log.Error(serv.extraInfo, "SubscribeEvent HD_TopNodeConsensusVoteResult failed.", err)
		return err
	}

	log.Debug(serv.extraInfo, "服务订阅完成", "")
	return nil
}

func (serv *TopNodeService) unSubMsg() {
	serv.roleUpdateSub.Unsubscribe()
	serv.leaderChangeSub.Unsubscribe()
	serv.consensusReqSub.Unsubscribe()
	serv.consensusVoteSub.Unsubscribe()
	serv.consensusResultSub.Unsubscribe()
}

func (serv *TopNodeService) update() {
	defer serv.unSubMsg()
	for {
		select {
		case data := <-serv.roleUpdateCh:
			topology, err := ca.GetTopologyByHash(common.RoleValidator|common.RoleBackupValidator, data.BlockHash)
			if err != nil {
				log.Error(serv.extraInfo, "处理CA通知消息", "获取拓扑图错误", "err", err)
				continue
			}
			if serv.msgCheck.CheckRoleUpdateMsg(data, topology) {
				log.Debug(serv.extraInfo, "处理CA通知消息", "", "块高", data.BlockNum)
				serv.stateMap.SetCurStates(data.BlockNum+1, topology.NodeList, topology.ElectList)
				go serv.LeaderChangeNotifyHandler(serv.msgCheck.GetCurLeader())
			}
		case data := <-serv.leaderChangeCh:
			log.Debug(serv.extraInfo, "处理leader变更通知消息", "", "块高", data.Number)
			if serv.msgCheck.CheckAndSaveLeaderChangeNotify(data) {
				go serv.LeaderChangeNotifyHandler(data.Leader)
			}

		case data := <-serv.consensusReqCh:
			go serv.consensusReqMsgHandler(data)
		case data := <-serv.consensusVoteCh:
			go serv.consensusVoteMsgHandler(data.Votes)
		case data := <-serv.consensusResultCh:
			go serv.OnlineConsensusVoteResultMsgHandler(data)
		case <-serv.quitCh:
			log.Info(serv.extraInfo, "收到退出消息", "")
			return
		}
	}
}

func (serv *TopNodeService) LeaderChangeNotifyHandler(leader common.Address) {
	if leader.Equal(common.Address{}) {
		log.Error(serv.extraInfo, "leader变更消息", "空消息")
		return
	}

	if serv.validatorSign.IsSelfAddress(leader) {
		log.Info(serv.extraInfo, "我是leader", "准备检查顶层节点在线状态")
		serv.sendRequest(serv.getTopNodeState(leader))
	} else {
		for _, item := range serv.dposRing.DPosVoteS {
			go serv.consensusVotes(item.getVotes())
		}
	}
}

func (serv *TopNodeService) getTopNodeState(leader common.Address) (online, offline []common.Address) {
	return serv.stateMap.newTopNodeState(serv.topNodeState.GetTopNodeOnlineState(), leader)
}

func (serv *TopNodeService) sendRequest(online, offline []common.Address) {
	leader := ca.GetAddress()
	reqMsg := mc.HD_OnlineConsensusReqs{
		From: leader,
	}
	number, turn := serv.msgCheck.GetRound()
	for _, item := range online {
		val := mc.OnlineConsensusReq{
			OnlineState: mc.OnLine,
			Leader:      leader,
			Node:        item,
			Number:      number,
			LeaderTurn:  turn,
		}
		reqMsg.ReqList = append(reqMsg.ReqList, &val)
	}
	for _, item := range offline {
		val := mc.OnlineConsensusReq{
			OnlineState: mc.OffLine,
			Leader:      leader,
			Node:        item,
			Number:      number,
			LeaderTurn:  turn,
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

func (serv *TopNodeService) consensusReqMsgHandler(msg *mc.HD_OnlineConsensusReqs) {
	if msg == nil || msg.ReqList == nil || len(msg.ReqList) == 0 {
		log.Error(serv.extraInfo, "invalid parameter", "")
		return
	}
	var votes mc.HD_OnlineConsensusVotes
	requests := msg.ReqList
	log.Info(serv.extraInfo, "处理共识请求", "开始", "处理总数", len(requests), "from", msg.From.Hex(), "leader", msg.ReqList[0].Leader.Hex())
	for i := 0; i < len(requests); i++ {
		item := requests[i]
		switch serv.msgCheck.CheckRound(item.Number, item.LeaderTurn) {
		case 1: // localRound > reqRound
			log.DEBUG(serv.extraInfo, "处理共识请求", "轮次过低，抛弃请求", "req Number", item.Number, "req turn", item.LeaderTurn)
			continue
		case -1: // localRound < reqRound
			serv.dposRing.addProposal(types.RlpHash(item), item)
			continue
		case 0: // localRound == reqRound
			if serv.dposRing.addProposal(types.RlpHash(item), item) {
				// todo 共识的节点判断，是否是顶层节点 或 elect节点
				sign, reqHash, err := serv.voteToReq(item)
				if err == nil {
					vote := mc.HD_ConsensusVote{}
					vote.SignHash.Set(reqHash)
					vote.Sign.Set(sign)
					vote.From.Set(ca.GetAddress())
					votes.Votes = append(votes.Votes, vote)
				} else {
					log.Error(serv.extraInfo, "处理共识请求", "签名失败", "error", err)
				}
			}
		}
	}

	if len(votes.Votes) > 0 {
		log.Info(serv.extraInfo, "处理共识请求", "发送投票消息")
		serv.msgSender.SendNodeMsg(mc.HD_TopNodeConsensusVote, &votes, common.RoleValidator, nil)
		go func() {
			serv.consensusVoteCh <- &votes
		}()
	}
}
func (serv *TopNodeService) consensusVoteMsgHandler(msg []mc.HD_ConsensusVote) {
	//log.Info(serv.extraInfo, "收到共识投票消息", "")
	if msg == nil || len(msg) == 0 {
		log.Error(serv.extraInfo, "处理共识投票", "", "invalid parameter", "", "len(msg)", len(msg))
		return
	}
	for i := 0; i < len(msg); i++ {
		item := msg[i]
		serv.consensusVotes(serv.dposRing.addVote(item.SignHash, &item))
	}
}

func (serv *TopNodeService) OnlineConsensusVoteResultMsgHandler(msg *mc.HD_OnlineConsensusVoteResultMsg) {
	//log.Info(serv.extraInfo, "收到共识结果消息", "")
	if msg == nil || msg.Req == nil {
		return
	}
	curNumber, _ := serv.msgCheck.GetRound()
	if msg.IsValidity(curNumber, manparams.OnlineConsensusValidityTime) == false {
		log.Error(serv.extraInfo, "处理共识结果消息", "共识消息已过期")
		return
	}

	tempSigns, err := serv.cd.VerifyHash(serv.validatorReader, types.RlpHash(msg.Req), msg.SignList)
	if err != nil {
		log.Error(serv.extraInfo, "处理共识结果消息", "POS验证失败", "err", err)
	} else {
		log.Info(serv.extraInfo, "处理共识结果消息", "验证通过，缓存状态", "状态", msg.Req.OnlineState.String(), "投票数", len(tempSigns))
		serv.stateMap.SaveConsensusResult(msg)
	}
}

func (serv *TopNodeService) consensusVotes(proposal interface{}, votes []voteInfo) {
	if proposal == nil || votes == nil || len(votes) == 0 {
		return
	}
	prop := proposal.(*mc.OnlineConsensusReq)
	curLeader := serv.msgCheck.GetCurLeader()
	if curLeader != prop.Leader {
		return
	}
	if serv.msgCheck.CheckRound(prop.Number, prop.LeaderTurn) != 0 {
		return
	}

	log.Info(serv.extraInfo, "处理共识投票", "开始")
	signList := make([]common.Signature, 0)
	for _, value := range votes {
		signList = append(signList, value.data.Sign)
	}
	rightSigns, err := serv.cd.VerifyHash(serv.validatorReader, votes[0].data.SignHash, signList)
	if err != nil {
		log.Debug(serv.extraInfo, "处理共识投票", "POS失败", "投票数", len(signList), "err", err)
		return
	}
	log.Info(serv.extraInfo, "处理共识投票", "POS通过，发送共识结果消息", "节点", prop.Node.String(), "状态", prop.OnlineState.String())
	//send DPos Success message
	result := mc.HD_OnlineConsensusVoteResultMsg{
		Req:      prop,
		SignList: rightSigns,
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

	if tempReq.Node.Equal(common.Address{}) || tempReq.Leader.Equal(common.Address{}) {
		log.Error(serv.extraInfo, "处理共识请求", "对共识请求进行投票", "无效的参数", "", "leader", tempReq.Leader.String(),
			"请求共识的节点", tempReq.Node.String())
		return common.Signature{}, common.Hash{}, voteFailed
	}
	reqHash := types.RlpHash(tempReq)
	if (reqHash == common.Hash{}) {
		log.Error(serv.extraInfo, "处理共识请求", "对请求的hash错误")
		return common.Signature{}, common.Hash{}, voteFailed
	}
	// TODO 优化，一次获取一个节点的在线状态 GetTopNodeOnlineState
	ok = serv.stateMap.checkNodeState(tempReq.Node, serv.topNodeState.GetTopNodeOnlineState(), tempReq.OnlineState)
	log.Info(serv.extraInfo, "处理共识请求", "对共识请求进行投票", "高度", tempReq.Number, "轮次", tempReq.LeaderTurn,
		"检查状态", tempReq.OnlineState.String(), "ok", ok, "node", tempReq.Node.Hex(), "hash", reqHash.TerminalString(), "leader", tempReq.Leader.Hex())

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
	return sign, reqHash, nil
}

//提供需要上区块头的顶点共识结果，只有leader调用
func (serv *TopNodeService) GetConsensusOnlineResults() []*mc.HD_OnlineConsensusVoteResultMsg {
	return serv.stateMap.GetConsensusResults()
}
