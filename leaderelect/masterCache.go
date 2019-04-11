// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package leaderelect

import (
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/types"

	"time"

	"github.com/MatrixAINetwork/go-matrix/accounts/signhelper"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/pkg/errors"
)

type posPool struct {
	reqHash   common.Hash
	reqMsg    interface{}
	voteCache map[common.Address]*common.VerifiedSign
}

func newPosPool() *posPool {
	return &posPool{
		reqHash:   common.Hash{},
		reqMsg:    nil,
		voteCache: make(map[common.Address]*common.VerifiedSign),
	}
}

func (pp *posPool) clear() {
	pp.reqHash = common.Hash{}
	pp.reqMsg = nil
	pp.voteCache = make(map[common.Address]*common.VerifiedSign)
}

func (pp *posPool) saveReqMsg(reqMsg interface{}) {
	pp.reqMsg = reqMsg
}

func (pp *posPool) saveReqMsgAndHash(reqMsg interface{}) common.Hash {
	pp.reqMsg = reqMsg
	pp.reqHash = types.RlpHash(reqMsg)
	pp.voteCache = make(map[common.Address]*common.VerifiedSign)
	return pp.reqHash
}

func (pp *posPool) getReqMsg() interface{} {
	return pp.reqMsg
}

func (pp *posPool) saveVoteMsg(reqHash common.Hash, sign common.Signature, from common.Address, cdc *cdc, signHelper *signhelper.SignHelper) error {
	if (pp.reqHash == common.Hash{} || pp.reqMsg == nil) {
		return ErrSelfReqIsNil
	}
	if cdc == nil || cdc.leaderCal == nil || signHelper == nil {
		return ErrCDCOrSignHelperisNil
	}

	if reqHash != pp.reqHash {
		return errors.Errorf("reqHash不匹配, reqHash(%s)!=localHash(%s)", reqHash.TerminalString(), pp.reqHash.TerminalString())
	}
	if _, exist := pp.voteCache[from]; exist {
		return errors.Errorf("响应已存在, from[%v]", from)
	}
	depositAccount, signAccount, validate, err := signHelper.VerifySignWithValidateByReader(cdc, reqHash.Bytes(), sign.Bytes(), cdc.leaderCal.preHash)
	if err != nil {
		return errors.Errorf("签名解析错误(%v)", err)
	}
	if signAccount != from {
		return errors.Errorf("签名账户(%s)与发送账户(%s)不匹配", signAccount.Hex(), from.Hex())
	}
	if !validate {
		return errors.New("签名为不同意签名")
	}
	pp.voteCache[signAccount] = &common.VerifiedSign{Sign: sign, Account: depositAccount, Validate: validate, Stock: 0}
	return nil
}

func (pp *posPool) getVotes() []*common.VerifiedSign {
	signs := make([]*common.VerifiedSign, 0)
	for _, v := range pp.voteCache {
		sign := v
		signs = append(signs, sign)
	}
	return signs
}

type masterCache struct {
	number                uint64
	selfAddr              common.Address
	selfNodeAddr          common.Address
	lastSignalInquiryTime int64
	inquiryResult         mc.ReelectRSPType
	inquiryPool           *posPool
	rlReqPool             *posPool
	broadcastPool         *posPool
}

func newMasterCache(number uint64) *masterCache {
	return &masterCache{
		number:                number,
		selfAddr:              common.Address{},
		selfNodeAddr:          common.Address{},
		lastSignalInquiryTime: 0,
		inquiryResult:         mc.ReelectRSPTypeNone,
		inquiryPool:           newPosPool(),
		rlReqPool:             newPosPool(),
		broadcastPool:         newPosPool(),
	}
}

func (self *masterCache) ClearSelfInquiryMsg() {
	self.lastSignalInquiryTime = 0
	self.inquiryResult = mc.ReelectRSPTypeNone
	self.inquiryPool.clear()
	self.rlReqPool.clear()
	self.broadcastPool.clear()
}

func (self *masterCache) CanSendSingleInquiryReq(time int64, interval int64) bool {
	if time-self.lastSignalInquiryTime <= interval {
		return false
	}
	return true
}

func (self *masterCache) SetLastSingleInquiryReqTime(time int64) {
	self.lastSignalInquiryTime = time
}

func (self *masterCache) GetInquiryResult() mc.ReelectRSPType {
	return self.inquiryResult
}

func (self *masterCache) SaveInquiryReq(req *mc.HD_ReelectInquiryReqMsg) common.Hash {
	if nil == req {
		log.Debug("leader masterCache", "SaveInquiryReq()", "param is nil")
		return common.Hash{}
	}

	self.ClearSelfInquiryMsg()
	self.inquiryResult = mc.ReelectRSPTypeNone
	return self.inquiryPool.saveReqMsgAndHash(req)
}

func (self *masterCache) IsMatchedInquiryRsp(rsp *mc.HD_ReelectInquiryRspMsg) error {
	if nil == rsp {
		return ErrParamsIsNil
	}
	if (self.inquiryPool.reqHash == common.Hash{}) {
		return ErrSelfReqIsNil
	}
	if rsp.ReqHash != self.inquiryPool.reqHash {
		return errors.Errorf("reqHash不匹配, reqHash(%s)!=localHash(%s)", rsp.ReqHash.TerminalString(), self.inquiryPool.reqHash.TerminalString())
	}
	return nil
}

func (self *masterCache) SaveInquiryVote(reqHash common.Hash, sign common.Signature, from common.Address, cdc *cdc, signHelper *signhelper.SignHelper) error {
	return self.inquiryPool.saveVoteMsg(reqHash, sign, from, cdc, signHelper)
}

func (self *masterCache) GetInquiryVotes() []*common.VerifiedSign {
	return self.inquiryPool.getVotes()
}

func (self *masterCache) GenRLReqMsg(inquiryAgreeVotes []common.Signature) error {
	if self.inquiryResult != mc.ReelectRSPTypeNone {
		return errors.Errorf("已存在询问结果(%v)，无法生存重选请求", self.inquiryResult)
	}
	self.inquiryResult = mc.ReelectRSPTypeAgree

	inquiryMsg, OK := self.inquiryPool.getReqMsg().(*mc.HD_ReelectInquiryReqMsg)
	if OK == false || inquiryMsg == nil {
		return errors.New("获取询问消息失败")
	}

	reqMsg := &mc.HD_ReelectLeaderReqMsg{
		InquiryReq: inquiryMsg,
		AgreeSigns: inquiryAgreeVotes,
		TimeStamp:  0,
	}
	self.rlReqPool.saveReqMsg(reqMsg)
	self.inquiryPool.clear()
	return nil
}

func (self *masterCache) GenBroadcastMsgWithInquiryResult(result mc.ReelectRSPType, rsp *mc.HD_ReelectInquiryRspMsg) error {
	if result != mc.ReelectRSPTypePOS && result != mc.ReelectRSPTypeAlreadyRL {
		return errors.Errorf("设置目标结果类型错误! target result = %v", result)
	}
	if self.inquiryResult != mc.ReelectRSPTypeNone {
		return errors.Errorf("已存在询问结果(%v)", self.inquiryResult)
	}
	self.inquiryResult = result
	self.inquiryPool.clear()

	broadcastMsg := &mc.HD_ReelectBroadcastMsg{
		Number:    self.number,
		Type:      result,
		POSResult: rsp.POSResult,
		RLResult:  rsp.RLResult,
		From:      self.selfNodeAddr,
	}
	self.broadcastPool.saveReqMsg(broadcastMsg)
	return nil
}

func (self *masterCache) GenBroadcastMsgWithRLSuccess(rlAgreeVotes []common.Signature) error {
	if self.inquiryResult != mc.ReelectRSPTypeAgree {
		return errors.Errorf("当前询问结果(%v) != ReelectRSPTypeAgree", self.inquiryResult)
	}

	rlReq, OK := self.rlReqPool.getReqMsg().(*mc.HD_ReelectLeaderReqMsg)
	if OK == false || rlReq == nil {
		return errors.New("缓存中获取重选leader请求消息失败!")
	}

	self.inquiryResult = mc.ReelectRSPTypeAlreadyRL

	rlResult := &mc.HD_ReelectLeaderConsensus{
		Req:   rlReq,
		Votes: rlAgreeVotes,
	}
	broadcastMsg := &mc.HD_ReelectBroadcastMsg{
		Number:    self.number,
		Type:      mc.ReelectRSPTypeAgree,
		POSResult: nil,
		RLResult:  rlResult,
		From:      self.selfNodeAddr,
	}
	self.broadcastPool.saveReqMsg(broadcastMsg)
	self.rlReqPool.clear()
	return nil
}

func (self *masterCache) GetRLReqMsg() (*mc.HD_ReelectLeaderReqMsg, common.Hash, error) {
	if self.inquiryResult != mc.ReelectRSPTypeAgree {
		return nil, common.Hash{}, errors.Errorf("当前询问结果(%v) != ReelectRSPTypeAgree", self.inquiryResult)
	}

	reqMsg, OK := self.rlReqPool.getReqMsg().(*mc.HD_ReelectLeaderReqMsg)
	if OK == false || reqMsg == nil {
		return nil, common.Hash{}, errors.New("缓存中不存在请求消息")
	}
	reqMsg.TimeStamp = uint64(time.Now().Unix())
	return reqMsg, self.rlReqPool.saveReqMsgAndHash(reqMsg), nil
}

func (self *masterCache) SaveRLVote(signHash common.Hash, sign common.Signature, from common.Address, cdc *cdc, signHelper *signhelper.SignHelper) error {
	return self.rlReqPool.saveVoteMsg(signHash, sign, from, cdc, signHelper)
}

func (self *masterCache) GetRLVotes() []*common.VerifiedSign {
	return self.rlReqPool.getVotes()
}

func (self *masterCache) GetBroadcastMsg() (*mc.HD_ReelectBroadcastMsg, common.Hash, error) {
	broadcastMsg, OK := self.broadcastPool.getReqMsg().(*mc.HD_ReelectBroadcastMsg)
	if OK == false || broadcastMsg == nil {
		return nil, common.Hash{}, errors.Errorf("缓存中没有重选结果广播消息")
	}
	return broadcastMsg, self.broadcastPool.saveReqMsgAndHash(broadcastMsg), nil
}

func (self *masterCache) GetLocalBroadcastMsg() (*mc.HD_ReelectBroadcastMsg, error) {
	broadcastMsg, OK := self.broadcastPool.getReqMsg().(*mc.HD_ReelectBroadcastMsg)
	if OK == false || broadcastMsg == nil {
		return nil, errors.Errorf("缓存中没有重选结果广播消息")
	}
	return broadcastMsg, nil
}

func (self *masterCache) SaveBroadcastVote(broadcastHash common.Hash, sign common.Signature, from common.Address, cdc *cdc, signHelper *signhelper.SignHelper) error {
	return self.broadcastPool.saveVoteMsg(broadcastHash, sign, from, cdc, signHelper)
}

func (self *masterCache) GetBroadcastVotes() []*common.VerifiedSign {
	return self.broadcastPool.getVotes()
}
