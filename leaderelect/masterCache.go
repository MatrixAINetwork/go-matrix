// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package leaderelect

import (
	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/crypto"
	"github.com/matrix/go-matrix/mc"
	"github.com/pkg/errors"
	"time"
)

type masterCache struct {
	number                uint64
	inquiryResult         mc.ReelectRSPType
	inquiryHash           common.Hash
	inquiryMsg            *mc.HD_ReelectInquiryReqMsg
	inquiryAgreeSignCache map[common.Address]*common.VerifiedSign
	rlReqMsg              *mc.HD_ReelectLeaderReqMsg
	rlReqHash             common.Hash
	rlVoteCache           map[common.Address]*common.VerifiedSign
	resultBroadcastHash   common.Hash
	resultBroadcastMsg    *mc.HD_ReelectResultBroadcastMsg
	resultRspCache        map[common.Address]*common.VerifiedSign
}

func newMasterCache(number uint64) *masterCache {
	return &masterCache{
		number:                number,
		inquiryResult:         mc.ReelectRSPTypeNone,
		inquiryHash:           common.Hash{},
		inquiryMsg:            nil,
		inquiryAgreeSignCache: nil,
		rlReqHash:             common.Hash{},
		rlReqMsg:              nil,
		rlVoteCache:           nil,
		resultBroadcastHash:   common.Hash{},
		resultBroadcastMsg:    nil,
		resultRspCache:        nil,
	}
}

func (self *masterCache) SetInquiryReq(req *mc.HD_ReelectInquiryReqMsg) common.Hash {
	if nil == req {
		return common.Hash{}
	}
	self.inquiryResult = mc.ReelectRSPTypeNone
	self.inquiryMsg = req
	self.inquiryHash = types.RlpHash(req)
	self.inquiryAgreeSignCache = make(map[common.Address]*common.VerifiedSign)
	self.rlReqMsg = nil
	self.rlReqHash = common.Hash{}
	self.rlVoteCache = make(map[common.Address]*common.VerifiedSign)
	self.resultBroadcastHash = common.Hash{}
	self.resultBroadcastMsg = nil
	self.resultRspCache = make(map[common.Address]*common.VerifiedSign)

	return self.inquiryHash
}

func (self *masterCache) ClearSelfInquiryMsg() {
	self.inquiryResult = mc.ReelectRSPTypeNone
	self.inquiryMsg = nil
	self.inquiryHash = common.Hash{}
	self.inquiryAgreeSignCache = make(map[common.Address]*common.VerifiedSign)
	self.rlReqMsg = nil
	self.rlReqHash = common.Hash{}
	self.rlVoteCache = make(map[common.Address]*common.VerifiedSign)
	self.resultBroadcastHash = common.Hash{}
	self.resultBroadcastMsg = nil
	self.resultRspCache = make(map[common.Address]*common.VerifiedSign)
}

func (self *masterCache) CheckInquiryRspMsg(rsp *mc.HD_ReelectInquiryRspMsg) error {
	if nil == rsp {
		return ErrMsgIsNil
	}
	if (self.inquiryHash == common.Hash{}) {
		return ErrSelfReqIsNil
	}
	if rsp.ReqHash != self.inquiryHash {
		return errors.Errorf("reqHash不匹配, reqHash(%s)!=localHash(%s)", rsp.ReqHash.TerminalString(), self.inquiryHash.TerminalString())
	}
	return nil
}

func (self *masterCache) SaveInquiryAgreeSign(reqHash common.Hash, sign common.Signature, from common.Address) error {
	if _, exist := self.inquiryAgreeSignCache[from]; exist {
		return errors.Errorf("来自(%s)的签名已存在!", from.Hex())
	}
	signAccount, validate, err := crypto.VerifySignWithValidate(reqHash.Bytes(), sign.Bytes())
	if err != nil {
		return errors.Errorf("签名解析错误(%v)", err)
	}
	if signAccount != from {
		return errors.Errorf("签名账户(%s)与发送账户(%s)不匹配", signAccount.Hex(), from.Hex())
	}
	if !validate {
		return errors.New("签名为不同意签名")
	}
	self.inquiryAgreeSignCache[signAccount] = &common.VerifiedSign{Sign: sign, Account: signAccount, Validate: validate, Stock: 0}
	return nil
}

func (self *masterCache) GetInquiryAgreeSigns() []*common.VerifiedSign {
	signs := make([]*common.VerifiedSign, 0)
	for _, v := range self.inquiryAgreeSignCache {
		sign := v
		signs = append(signs, sign)
	}
	return signs
}

func (self *masterCache) SetInquiryResultAgree(rightSign []common.Signature) error {
	if self.inquiryResult != mc.ReelectRSPTypeNone {
		return errors.Errorf("已存在询问结果(%v)", self.inquiryResult)
	}
	self.inquiryResult = mc.ReelectRSPTypeAgree
	self.rlReqMsg = &mc.HD_ReelectLeaderReqMsg{
		InquiryReq: self.inquiryMsg,
		AgreeSigns: rightSign,
		TimeStamp:  0,
	}
	self.inquiryMsg = nil
	self.inquiryHash = common.Hash{}
	self.inquiryAgreeSignCache = make(map[common.Address]*common.VerifiedSign)
	return nil
}

func (self *masterCache) SetInquiryResultNotAgree(result mc.ReelectRSPType, rsp *mc.HD_ReelectInquiryRspMsg) error {
	if result != mc.ReelectRSPTypePOS && result != mc.ReelectRSPTypeAlreadyRL {
		return errors.Errorf("设置目标结果类型错误! target result = %v", result)
	}
	if self.inquiryResult != mc.ReelectRSPTypeNone {
		return errors.Errorf("已存在询问结果(%v)", self.inquiryResult)
	}
	self.inquiryResult = result
	self.inquiryMsg = nil
	self.inquiryHash = common.Hash{}
	self.inquiryAgreeSignCache = make(map[common.Address]*common.VerifiedSign)
	self.rlReqMsg = nil
	self.resultBroadcastMsg = &mc.HD_ReelectResultBroadcastMsg{
		Number:    self.number,
		Type:      result,
		POSResult: rsp.POSResult,
		RLResult:  rsp.RLResult,
		From:      ca.GetAddress(),
	}
	return nil
}

func (self *masterCache) SetRLResultBroadcastSuccess(signs []common.Signature) error {
	if self.inquiryResult != mc.ReelectRSPTypeAgree {
		return errors.Errorf("当前询问结果(%v) != ReelectRSPTypeAgree", self.inquiryResult)
	}

	rlResult := &mc.HD_ReelectLeaderConsensus{
		Req:   self.rlReqMsg,
		Votes: signs,
	}

	self.inquiryResult = mc.ReelectRSPTypeAlreadyRL
	self.resultBroadcastMsg = &mc.HD_ReelectResultBroadcastMsg{
		Number:    self.number,
		Type:      mc.ReelectRSPTypeAgree,
		POSResult: nil,
		RLResult:  rlResult,
		From:      ca.GetAddress(),
	}
	return nil
}

func (self *masterCache) InquiryResult() mc.ReelectRSPType {
	return self.inquiryResult
}

func (self *masterCache) GetRLReqMsg() (*mc.HD_ReelectLeaderReqMsg, common.Hash, error) {
	if self.inquiryResult != mc.ReelectRSPTypeAgree {
		return nil, common.Hash{}, errors.Errorf("当前询问结果(%v) != ReelectRSPTypeAgree", self.inquiryResult)
	}
	self.rlReqMsg.TimeStamp = uint64(time.Now().Unix())
	self.rlReqHash = types.RlpHash(self.rlReqMsg)
	self.rlVoteCache = make(map[common.Address]*common.VerifiedSign)
	return self.rlReqMsg, self.rlReqHash, nil
}

func (self *masterCache) SaveRLVote(signHash common.Hash, sign common.Signature, from common.Address) error {
	if (self.rlReqHash == common.Hash{}) {
		return ErrSelfReqIsNil
	}
	if signHash != self.rlReqHash {
		return errors.Errorf("signHash不匹配, signHash(%s)!=localHash(%s)", signHash.TerminalString(), self.rlReqHash.TerminalString())
	}
	if _, exist := self.rlVoteCache[from]; exist {
		return errors.Errorf("来自(%s)的签名已存在", from.Hex())
	}
	signAccount, validate, err := crypto.VerifySignWithValidate(signHash.Bytes(), sign.Bytes())
	if err != nil {
		return errors.Errorf("签名解析错误(%v)", err)
	}
	if signAccount != from {
		return errors.Errorf("签名账户(%s)与发送账户(%s)不匹配", signAccount.Hex(), from.Hex())
	}
	if !validate {
		return errors.New("签名为不同意签名")
	}
	self.rlVoteCache[signAccount] = &common.VerifiedSign{Sign: sign, Account: signAccount, Validate: validate, Stock: 0}
	return nil
}

func (self *masterCache) GetRLSigns() []*common.VerifiedSign {
	signs := make([]*common.VerifiedSign, 0)
	for _, v := range self.rlVoteCache {
		sign := v
		signs = append(signs, sign)
	}
	return signs
}

func (self *masterCache) GetResultBroadcastMsg() (*mc.HD_ReelectResultBroadcastMsg, common.Hash, error) {
	if self.resultBroadcastMsg == nil {
		return nil, common.Hash{}, errors.Errorf("缓存中没有重选结果广播消息")
	}
	self.resultBroadcastMsg.TimeStamp = uint64(time.Now().Unix())
	self.resultBroadcastHash = types.RlpHash(self.resultBroadcastMsg)
	self.resultRspCache = make(map[common.Address]*common.VerifiedSign)
	return self.resultBroadcastMsg, self.resultBroadcastHash, nil
}

func (self *masterCache) GetLocalResultMsg() (*mc.HD_ReelectResultBroadcastMsg, error) {
	if self.resultBroadcastMsg == nil {
		return nil, errors.Errorf("缓存中没有重选结果广播消息")
	}
	return self.resultBroadcastMsg, nil
}

func (self *masterCache) SaveResultRsp(resultHash common.Hash, sign common.Signature, from common.Address) error {
	if (self.resultBroadcastHash == common.Hash{}) {
		return ErrBroadcastIsNil
	}
	if resultHash != self.resultBroadcastHash {
		return errors.Errorf("ResultHash不匹配, ResultHash(%s)!=localHash(%s)", resultHash.TerminalString(), self.resultBroadcastHash.TerminalString())
	}
	if _, exist := self.resultRspCache[from]; exist {
		return errors.Errorf("响应已存在, from[%v]", from)
	}
	signAccount, validate, err := crypto.VerifySignWithValidate(resultHash.Bytes(), sign.Bytes())
	if err != nil {
		return errors.Errorf("签名解析错误(%v)", err)
	}
	if signAccount != from {
		return errors.Errorf("签名账户(%s)与发送账户(%s)不匹配", signAccount.Hex(), from.Hex())
	}
	if !validate {
		return errors.New("签名为不同意签名")
	}
	self.resultRspCache[signAccount] = &common.VerifiedSign{Sign: sign, Account: signAccount, Validate: validate, Stock: 0}
	return nil
}

func (self *masterCache) GetResultRspSigns() []*common.VerifiedSign {
	signs := make([]*common.VerifiedSign, 0)
	for _, v := range self.resultRspCache {
		sign := v
		signs = append(signs, sign)
	}
	return signs
}
