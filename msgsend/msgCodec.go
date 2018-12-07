// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package msgsend

import (
	"encoding/json"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/pkg/errors"
)

func (self *HD) initCodec() {
	self.registerCodec(mc.HD_BlkConsensusReq, new(blkConsensusReqCodec))
	self.registerCodec(mc.HD_BlkConsensusVote, new(blkConsensusVoteCodec))
	self.registerCodec(mc.HD_MiningReq, new(miningReqCodec))
	self.registerCodec(mc.HD_MiningRsp, new(miningRspCodec))
	self.registerCodec(mc.HD_BroadcastMiningRsp, new(broadcastMiningRspCodec))
	self.registerCodec(mc.HD_NewBlockInsert, new(newBlockInsertCodec))
	self.registerCodec(mc.HD_TopNodeConsensusReq, new(onlineConsensusReqCodec))
	self.registerCodec(mc.HD_TopNodeConsensusVote, new(onlineConsensusVoteCodec))
	self.registerCodec(mc.HD_TopNodeConsensusVoteResult, new(onlineConsensusResultCodec))
	self.registerCodec(mc.HD_LeaderReelectInquiryReq, new(lrInquiryReqCodec))
	self.registerCodec(mc.HD_LeaderReelectInquiryRsp, new(lrInquiryRspCodec))
	self.registerCodec(mc.HD_LeaderReelectReq, new(lrReqCodec))
	self.registerCodec(mc.HD_LeaderReelectVote, new(lrVoteCodec))
	self.registerCodec(mc.HD_LeaderReelectResultBroadcast, new(lrResultBCCodec))
	self.registerCodec(mc.HD_LeaderReelectResultBroadcastRsp, new(lrResultBCRspCodec))
	self.registerCodec(mc.HD_FullBlockReq, new(fullBlockReqCodec))
	self.registerCodec(mc.HD_FullBlockRsp, new(fullBlockRspCodec))
}

//每个模块需要自己实现这两个接口
type MsgCodec interface {
	EncodeFn(msg interface{}) ([]byte, error)
	DecodeFn(data []byte, from common.Address) (interface{}, error)
}

////////////////////////////////////////////////////////////////////////
// 区块共识请求消息
// msg code = mc.HD_BlkConsensusReq
type blkConsensusReqCodec struct {
}

func (*blkConsensusReqCodec) EncodeFn(msg interface{}) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, errors.Errorf("json.Marshal failed: %s", err)
	}
	return data, nil
}

func (*blkConsensusReqCodec) DecodeFn(data []byte, from common.Address) (interface{}, error) {
	msg := new(mc.HD_BlkConsensusReqMsg)
	err := json.Unmarshal([]byte(data), msg)
	if err != nil {
		return nil, errors.Errorf("json.Unmarshal failed: %s", err)
	}
	if msg.Header == nil {
		return nil, errors.Errorf("'header' of the msg is nil")
	}
	msg.From.Set(from)
	return msg, nil
}

////////////////////////////////////////////////////////////////////////
// 区块共识投票消息
// msg code = mc.HD_BlkConsensusVote
type blkConsensusVoteCodec struct {
}

func (*blkConsensusVoteCodec) EncodeFn(msg interface{}) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, errors.Errorf("json.Marshal failed: %s", err)
	}
	return data, nil
}

func (*blkConsensusVoteCodec) DecodeFn(data []byte, from common.Address) (interface{}, error) {
	msg := new(mc.HD_ConsensusVote)
	err := json.Unmarshal([]byte(data), msg)
	if err != nil {
		return nil, errors.Errorf("json.Unmarshal failed: %s", err)
	}
	msg.From.Set(from)
	return msg, nil
}

////////////////////////////////////////////////////////////////////////
// 挖矿请求消息
// msg code = mc.HD_MiningReq
type miningReqCodec struct {
}

func (*miningReqCodec) EncodeFn(msg interface{}) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, errors.Errorf("json.Marshal failed: %s", err)
	}
	return data, nil
}

func (*miningReqCodec) DecodeFn(data []byte, from common.Address) (interface{}, error) {
	msg := new(mc.HD_MiningReqMsg)
	err := json.Unmarshal([]byte(data), msg)
	if err != nil {
		return nil, errors.Errorf("json.Unmarshal failed: %s", err)
	}
	if msg.Header == nil {
		return nil, errors.Errorf("'header' of the msg is nil")
	}
	msg.From.Set(from)
	return msg, nil
}

////////////////////////////////////////////////////////////////////////
// 挖矿结果消息
// msg code = mc.HD_MiningRsp
type miningRspCodec struct {
}

func (*miningRspCodec) EncodeFn(msg interface{}) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, errors.Errorf("json.Marshal failed: %s", err)
	}
	return data, nil
}

func (*miningRspCodec) DecodeFn(data []byte, from common.Address) (interface{}, error) {
	msg := new(mc.HD_MiningRspMsg)
	err := json.Unmarshal([]byte(data), msg)
	if err != nil {
		return nil, errors.Errorf("json.Unmarshal failed: %s", err)
	}
	if msg.Difficulty == nil {
		return nil, errors.Errorf("'Difficulty' of the msg is nil")
	}
	msg.From.Set(from)
	return msg, nil
}

////////////////////////////////////////////////////////////////////////
// 广播挖矿结果消息
// msg code = mc.HD_BroadcastMiningRsp
type broadcastMiningRspCodec struct {
}

func (*broadcastMiningRspCodec) EncodeFn(msg interface{}) ([]byte, error) {
	rsp, OK := msg.(*mc.HD_BroadcastMiningRspMsg)
	if !OK {
		return nil, errors.New("reflect err! broadcast_mining_rsp_msg")
	}

	size := rsp.BlockMainData.Txs.Len()
	marshalMsg := fullBlockMsgForMarshal{}
	marshalMsg.Txs = make([]*types.Transaction_Mx, 0, size)
	for i := 0; i < size; i++ {
		tx := rsp.BlockMainData.Txs[i]
		log.DEBUG("HD", "广播挖矿结果消息, Marshal前的tx", tx)
		marshalMsg.Txs = append(marshalMsg.Txs, types.GetTransactionMx(tx))
	}
	marshalMsg.Header = rsp.BlockMainData.Header
	data, err := json.Marshal(marshalMsg)
	if err != nil {
		return nil, errors.Errorf("json.Marshal failed: %s", err)
	}
	return data, nil
}

func (*broadcastMiningRspCodec) DecodeFn(data []byte, from common.Address) (interface{}, error) {
	msg := &fullBlockMsgForMarshal{}
	err := json.Unmarshal([]byte(data), msg)
	if err != nil {
		return nil, errors.Errorf("json.Unmarshal failed: %s", err)
	}
	if msg.Header == nil {
		return nil, errors.Errorf("'Header' of the msg is nil")
	}

	sendMsg := &mc.HD_BroadcastMiningRspMsg{
		From: from,
		BlockMainData: &mc.BlockData{
			Header: msg.Header,
			Txs:    make(types.SelfTransactions, 0),
		},
	}
	size := len(msg.Txs)
	for i := 0; i < size; i++ {
		tx := types.SetTransactionMx(msg.Txs[i])
		log.DEBUG("HD", "广播挖矿结果消息, Unmarshal后的tx", tx)
		sendMsg.BlockMainData.Txs = append(sendMsg.BlockMainData.Txs, tx)
	}

	return sendMsg, nil
}

////////////////////////////////////////////////////////////////////////
// 新区块插入消息
// msg code = mc.HD_NewBlockInsert
type newBlockInsertCodec struct {
}

func (*newBlockInsertCodec) EncodeFn(msg interface{}) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, errors.Errorf("json.Marshal failed: %s", err)
	}
	return data, nil
}

func (*newBlockInsertCodec) DecodeFn(data []byte, from common.Address) (interface{}, error) {
	msg := new(mc.HD_BlockInsertNotify)
	err := json.Unmarshal([]byte(data), msg)
	if err != nil {
		return nil, errors.Errorf("json.Unmarshal failed: %s", err)
	}
	if msg.Header == nil {
		return nil, errors.Errorf("'Header' of the msg is nil")
	}
	msg.From.Set(from)
	return msg, nil
}

////////////////////////////////////////////////////////////////////////
// 顶层节点在线共识请求消息
// msg code = mc.HD_TopNodeConsensusReq
type onlineConsensusReqCodec struct {
}

func (*onlineConsensusReqCodec) EncodeFn(msg interface{}) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, errors.Errorf("json.Marshal failed: %s", err)
	}
	return data, nil
}

func (*onlineConsensusReqCodec) DecodeFn(data []byte, from common.Address) (interface{}, error) {
	msg := new(mc.HD_OnlineConsensusReqs)
	err := json.Unmarshal([]byte(data), msg)
	if err != nil {
		return nil, errors.Errorf("json.Unmarshal failed: %s", err)
	}
	msg.From.Set(from)
	return msg, nil
}

////////////////////////////////////////////////////////////////////////
// 顶层节点在线共识投票消息
// msg code = mc.HD_TopNodeConsensusVote
type onlineConsensusVoteCodec struct {
}

func (*onlineConsensusVoteCodec) EncodeFn(msg interface{}) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, errors.Errorf("json.Marshal failed: %s", err)
	}
	return data, nil
}

func (*onlineConsensusVoteCodec) DecodeFn(data []byte, from common.Address) (interface{}, error) {
	msg := new(mc.HD_OnlineConsensusVotes)
	err := json.Unmarshal([]byte(data), msg)
	if err != nil {
		return nil, errors.Errorf("json.Unmarshal failed: %s", err)
	}

	for _, vote := range msg.Votes {
		vote.From.Set(from)
	}

	return msg, nil
}

////////////////////////////////////////////////////////////////////////
// 顶层节点在线共识结果消息
// msg code = mc.HD_TopNodeConsensusVoteResult
type onlineConsensusResultCodec struct {
}

func (*onlineConsensusResultCodec) EncodeFn(msg interface{}) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, errors.Errorf("json.Marshal failed: %s", err)
	}
	return data, nil
}

func (*onlineConsensusResultCodec) DecodeFn(data []byte, from common.Address) (interface{}, error) {
	msg := new(mc.HD_OnlineConsensusVoteResultMsg)
	err := json.Unmarshal([]byte(data), msg)
	if err != nil {
		return nil, errors.Errorf("json.Unmarshal failed: %s", err)
	}
	return msg, nil
}

////////////////////////////////////////////////////////////////////////
// 重选询问请求消息
// msg code = mc.HD_LeaderReelectInquiryReq
type lrInquiryReqCodec struct {
}

func (*lrInquiryReqCodec) EncodeFn(msg interface{}) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, errors.Errorf("json.Marshal failed: %s", err)
	}
	return data, nil
}

func (*lrInquiryReqCodec) DecodeFn(data []byte, from common.Address) (interface{}, error) {
	msg := new(mc.HD_ReelectInquiryReqMsg)
	err := json.Unmarshal([]byte(data), msg)
	if err != nil {
		return nil, errors.Errorf("json.Unmarshal failed: %s", err)
	}
	msg.From.Set(from)
	return msg, nil
}

////////////////////////////////////////////////////////////////////////
// 重选询问响应消息
// msg code = mc.HD_LeaderReelectInquiryRsp
type lrInquiryRspCodec struct {
}

func (*lrInquiryRspCodec) EncodeFn(msg interface{}) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, errors.Errorf("json.Marshal failed: %s", err)
	}
	return data, nil
}

func (*lrInquiryRspCodec) DecodeFn(data []byte, from common.Address) (interface{}, error) {
	msg := new(mc.HD_ReelectInquiryRspMsg)
	err := json.Unmarshal([]byte(data), msg)
	if err != nil {
		return nil, errors.Errorf("json.Unmarshal failed: %s", err)
	}
	msg.From.Set(from)
	return msg, nil
}

////////////////////////////////////////////////////////////////////////
// leader重选请求消息
// msg code = mc.HD_LeaderReelectReq
type lrReqCodec struct {
}

func (*lrReqCodec) EncodeFn(msg interface{}) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, errors.Errorf("json.Marshal failed: %s", err)
	}
	return data, nil
}

func (*lrReqCodec) DecodeFn(data []byte, from common.Address) (interface{}, error) {
	msg := new(mc.HD_ReelectLeaderReqMsg)
	err := json.Unmarshal([]byte(data), msg)
	if err != nil {
		return nil, errors.Errorf("json.Unmarshal failed: %s", err)
	}
	if msg.InquiryReq == nil {
		return nil, errors.Errorf("'InquiryReq' of the msg is nil")
	}
	msg.InquiryReq.From.Set(from)
	return msg, nil
}

////////////////////////////////////////////////////////////////////////
// leader重选投票消息
// msg code = mc.HD_LeaderReelectVote
type lrVoteCodec struct {
}

func (*lrVoteCodec) EncodeFn(msg interface{}) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, errors.Errorf("json.Marshal failed: %s", err)
	}
	return data, nil
}

func (*lrVoteCodec) DecodeFn(data []byte, from common.Address) (interface{}, error) {
	msg := new(mc.HD_ReelectLeaderVoteMsg)
	err := json.Unmarshal([]byte(data), msg)
	if err != nil {
		return nil, errors.Errorf("json.Unmarshal failed: %s", err)
	}
	msg.Vote.From.Set(from)
	return msg, nil
}

////////////////////////////////////////////////////////////////////////
// leader重选结果广播消息
// msg code = mc.HD_LeaderReelectResultBroadcast
type lrResultBCCodec struct {
}

func (*lrResultBCCodec) EncodeFn(msg interface{}) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, errors.Errorf("json.Marshal failed: %s", err)
	}
	return data, nil
}

func (*lrResultBCCodec) DecodeFn(data []byte, from common.Address) (interface{}, error) {
	msg := new(mc.HD_ReelectResultBroadcastMsg)
	err := json.Unmarshal([]byte(data), msg)
	if err != nil {
		return nil, errors.Errorf("json.Unmarshal failed: %s", err)
	}
	msg.From.Set(from)
	return msg, nil
}

////////////////////////////////////////////////////////////////////////
// leader重选结果广播响应消息
// msg code = mc.HD_LeaderReelectResultBroadcastRsp
type lrResultBCRspCodec struct {
}

func (*lrResultBCRspCodec) EncodeFn(msg interface{}) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, errors.Errorf("json.Marshal failed: %s", err)
	}
	return data, nil
}

func (*lrResultBCRspCodec) DecodeFn(data []byte, from common.Address) (interface{}, error) {
	msg := new(mc.HD_ReelectResultRspMsg)
	err := json.Unmarshal([]byte(data), msg)
	if err != nil {
		return nil, errors.Errorf("json.Unmarshal failed: %s", err)
	}
	msg.From.Set(from)
	return msg, nil
}

////////////////////////////////////////////////////////////////////////
// 完整区块获取请求
// msg code = mc.HD_FullBlockReq
type fullBlockReqCodec struct {
}

func (*fullBlockReqCodec) EncodeFn(msg interface{}) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, errors.Errorf("json.Marshal failed: %s", err)
	}
	return data, nil
}

func (*fullBlockReqCodec) DecodeFn(data []byte, from common.Address) (interface{}, error) {
	msg := new(mc.HD_FullBlockReqMsg)
	err := json.Unmarshal([]byte(data), msg)
	if err != nil {
		return nil, errors.Errorf("json.Unmarshal failed: %s", err)
	}
	msg.From.Set(from)
	return msg, nil
}

////////////////////////////////////////////////////////////////////////
// 完整区块获取响应
// msg code = mc.HD_FullBlockRsp
type fullBlockRspCodec struct {
}

func (*fullBlockRspCodec) EncodeFn(msg interface{}) ([]byte, error) {
	rsp, OK := msg.(*mc.HD_FullBlockRspMsg)
	if !OK {
		return nil, errors.New("reflect err! HD_FullBlockRspMsg")
	}

	size := rsp.Txs.Len()
	marshalMsg := fullBlockMsgForMarshal{}
	marshalMsg.Txs = make([]*types.Transaction_Mx, 0, size)
	for i := 0; i < size; i++ {
		tx := rsp.Txs[i]
		marshalMsg.Txs = append(marshalMsg.Txs, types.SetTransactionToMx(tx))
	}
	marshalMsg.Header = rsp.Header
	data, err := json.Marshal(marshalMsg)
	if err != nil {
		return nil, errors.Errorf("json.Marshal failed: %s", err)
	}
	return data, nil
}

func (*fullBlockRspCodec) DecodeFn(data []byte, from common.Address) (interface{}, error) {
	msg := &fullBlockMsgForMarshal{}
	err := json.Unmarshal([]byte(data), msg)
	if err != nil {
		return nil, errors.Errorf("json.Unmarshal failed: %s", err)
	}

	sendMsg := &mc.HD_FullBlockRspMsg{
		From:   from,
		Header: msg.Header,
		Txs:    make(types.SelfTransactions, 0),
	}
	size := len(msg.Txs)
	for i := 0; i < size; i++ {
		tx := types.SetMxToTransaction(msg.Txs[i])
		sendMsg.Txs = append(sendMsg.Txs, tx)
	}

	return sendMsg, nil
}
