// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package leaderelect

import (
	"time"

	"sync"

	"github.com/matrix/go-matrix/accounts/signhelper"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/consensus"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/crypto"
	"github.com/matrix/go-matrix/event"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/pkg/errors"
)

type vote struct {
	result   int //agree or disagree
	SignHash common.Hash
	Sign     common.Signature
}

var LDREReqTimeOut = 3 * time.Second

type ldreMaster struct {
	quitCh         chan bool
	resultCh       chan *mc.ReelectLeaderSuccMsg
	voteMsgPool    votePool
	extra          string
	matrix         Matrix
	ce             consensus.DPOSEngine
	signHelper     *signhelper.SignHelper
	voteReqReg     *mc.HD_LeaderReelectVoteReqMsg
	voteReqHash    common.Hash
	voteResultCh   chan *mc.HD_ConsensusVote
	voteResultSub  event.Subscription
	consensusState bool
}

type votePool struct {
	lock sync.Mutex
	pool map[common.Address]*common.VerifiedSign
}

var ErrVotePoolKeyRep = errors.New("Key已存在")
var ErrMasterConsensusFinished = errors.New("Master已经完成共识")

func (self *votePool) Reset() {
	self.lock.Lock()
	defer self.lock.Unlock()

	self.pool = make(map[common.Address]*common.VerifiedSign)
}
func (self *votePool) isKeyExist(address common.Address) bool {
	self.lock.Lock()
	defer self.lock.Unlock()

	_, ok := self.pool[address]

	return ok
}
func (self *votePool) Add(address common.Address, signVal *common.VerifiedSign) error {
	self.lock.Lock()
	defer self.lock.Unlock()

	if _, ok := self.pool[address]; !ok {
		self.pool[address] = signVal
		return nil
	} else {
		return ErrVotePoolKeyRep
	}
}
func (self *votePool) GetSignList() []*common.VerifiedSign {
	self.lock.Lock()
	defer self.lock.Unlock()

	var signList = make([]*common.VerifiedSign, 0)

	for _, v := range self.pool {
		signList = append(signList, v)
	}

	return signList
}
func newMaster(matrix Matrix, extra string, msg *mc.LeaderReelectMsg, resultCh chan *mc.ReelectLeaderSuccMsg) (*ldreMaster, error) {

	var err error
	master := &ldreMaster{
		quitCh:         make(chan bool, 1),
		resultCh:       resultCh,
		voteResultCh:   make(chan *mc.HD_ConsensusVote, 1),
		voteMsgPool:    votePool{},
		matrix:         matrix,
		ce:             matrix.DPOSEngine(),
		signHelper:     matrix.SignHelper(),
		consensusState: false,
		extra:          extra,
	}

	if master.voteResultSub, err = mc.SubscribeEvent(mc.HD_LeaderReelectVoteRsp, master.voteResultCh); err != nil {
		log.ERROR(master.extra, "Master订阅投票请求消息错误", err)
		return nil, err
	}

	master.voteMsgPool.Reset()
	master.voteReqReg = &mc.HD_LeaderReelectVoteReqMsg{Leader: msg.Leader, Height: msg.Number, ReelectTurn: msg.ReelectTurn}

	go master.run(msg)
	return master, nil
}
func (self *ldreMaster) broadcastVoteConsensusMsg(signs []common.Signature) {
	msg := &mc.HD_LeaderReelectConsensusBroadcastMsg{
		Signatures: signs,
		Req:        *self.voteReqReg,
	}
	self.matrix.HD().SendNodeMsg(mc.HD_LeaderReelectConsensusBroadcast, msg, common.RoleValidator, nil)
}
func (self *ldreMaster) reelectLeaderResultHandle(voteMsg *mc.HD_ConsensusVote, number uint64) error {
	if self.consensusState == true {
		return ErrMasterConsensusFinished
	}

	from, validate, err := self.voteMsgCheck(voteMsg)
	if err != nil {
		return err
	}

	if err := self.voteMsgPool.Add(from, &common.VerifiedSign{Sign: voteMsg.Sign, Validate: validate, Account: from}); err != nil {
		return err
	}

	signList := self.voteMsgPool.GetSignList()
	signs, err := self.ce.VerifyHashWithVerifiedSignsAndNumber(signList, number-1)
	if err != nil {
		return errors.Errorf("共识失败，总票数%d, err:%s", len(signList), err)
	}

	log.INFO(self.extra, "广播LDRE共识成功消息, sign count", len(signs))
	//广播发送重选leader成功消息给所有验证者
	self.broadcastVoteConsensusMsg(signs)

	log.INFO(self.extra, "发布LDRE成功", "至LD控制模块")
	//发出重选leader成功消息，给leader控制服务
	self.sendVoteConsensusToController(self.voteReqReg.Leader, self.voteReqReg.ReelectTurn, self.voteReqReg.Height)
	self.consensusState = true

	return nil

}
func (self *ldreMaster) sendVoteConsensusToController(masterAddress common.Address, turnNum uint8, height uint64) {
	msg := &mc.ReelectLeaderSuccMsg{
		Height:      height,
		Leader:      masterAddress,
		ReelectTurn: turnNum,
	}

	self.resultCh <- msg
}

func (self *ldreMaster) voteMsgCheck(msg *mc.HD_ConsensusVote) (common.Address, bool, error) {
	if msg == nil {
		return common.Address{}, false, errors.New("vote msg指针为空")
	}

	//投票结果是否是对本次的投票
	if !self.voteReqHash.Equal(msg.SignHash) {
		return common.Address{}, false, errors.New("投票Hash不匹配")
	}

	from, validate, err := crypto.VerifySignWithValidate(msg.SignHash.Bytes(), msg.Sign.Bytes())
	if err != nil {
		return common.Address{}, false, errors.Errorf("投票签名解析错误, %s", err)
	}

	if from != msg.From {
		return common.Address{}, false, errors.New("签名账户和消息来源账户不一致")
	}

	if self.voteMsgPool.isKeyExist(from) {
		return common.Address{}, false, errors.New("重复投票")
	}

	return from, validate, nil
}

func (self *ldreMaster) sendVoteReqMsg(msg *mc.HD_LeaderReelectVoteReqMsg) {
	self.matrix.HD().SendNodeMsg(mc.HD_LeaderReelectVoteReq, msg, common.RoleValidator, nil)
}
func (self *ldreMaster) run(reqMsg *mc.LeaderReelectMsg) {
	log.INFO(self.extra, "服务", "启动", "高度", reqMsg.Number, "轮次", reqMsg.ReelectTurn)
	defer func() {
		self.voteResultSub.Unsubscribe()
		log.INFO(self.extra, "服务", "退出", "高度", reqMsg.Number, "轮次", reqMsg.ReelectTurn)
	}()

	self.voteInit(reqMsg)
	var voteReqRetryTimer = time.NewTimer(LDREReqTimeOut)
	for {
		select {
		case vote := <-self.voteResultCh:
			log.INFO(self.extra, "收到LDRE投票RSP, From", vote.From.Hex(), "高度", reqMsg.Number)
			if err := self.reelectLeaderResultHandle(vote, reqMsg.Number); err != nil {
				log.INFO(self.extra, "投票消息处理错误", err, "高度", reqMsg.Number)
			}
			log.INFO(self.extra, "投票消息处理", "完成", "高度", reqMsg.Number)

		case <-voteReqRetryTimer.C:
			log.INFO(self.extra, "超时", "开启新一轮请求", "高度", reqMsg.Number)
			self.voteInit(reqMsg)
			voteReqRetryTimer.Reset(LDREReqTimeOut)

		case <-self.quitCh:
			log.INFO(self.extra, "主动退出", "Master处理")
			return
		}
	}
}

func (self *ldreMaster) voteInit(msg *mc.LeaderReelectMsg) {
	log.INFO(self.extra, "LDRE投票开始", "清除票池")
	self.voteMsgPool.Reset()

	//向所有验证者发送重选leader请求消息
	self.voteReqReg.TimeStamp = uint64(time.Now().Unix())
	self.voteReqHash = types.RlpHash(self.voteReqReg)

	log.INFO(self.extra, "发出LDRE投票请求, Turns", msg.ReelectTurn, "高度", msg.Number)
	self.sendVoteReqMsg(self.voteReqReg)

	sign, err := self.signHelper.SignHashWithValidate(self.voteReqHash.Bytes(), true)
	if err != nil {
		log.ERROR(self.extra, "LDRE投票签名失败, Turns", msg.ReelectTurn, "高度", msg.Number)
		return
	}

	self.voteMsgPool.Add(self.voteReqReg.Leader, &common.VerifiedSign{Sign: sign, Validate: true, Account: self.voteReqReg.Leader})
	return
}
