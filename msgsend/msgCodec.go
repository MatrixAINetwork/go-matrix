// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package msgsend

import (
	"encoding/json"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/rlp"
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
	self.registerCodec(mc.HD_LeaderReelectBroadcast, new(lrResultBCCodec))
	self.registerCodec(mc.HD_LeaderReelectBroadcastRsp, new(lrResultBCRspCodec))
	self.registerCodec(mc.HD_FullBlockReq, new(fullBlockReqCodec))
	self.registerCodec(mc.HD_FullBlockRsp, new(fullBlockRspCodec))

	self.registerCodec(mc.HD_V2_LeaderReelectInquiryReq, new(lrInquiryReqCodecV2))
	self.registerCodec(mc.HD_V2_LeaderReelectInquiryRsp, new(lrInquiryRspCodecV2))
	self.registerCodec(mc.HD_V2_LeaderReelectReq, new(lrReqCodecV2))
	self.registerCodec(mc.HD_V2_LeaderReelectVote, new(lrVoteCodecV2))
	self.registerCodec(mc.HD_V2_LeaderReelectBroadcast, new(lrResultBCCodecV2))
	self.registerCodec(mc.HD_V2_LeaderReelectBroadcastRsp, new(lrResultBCRspCodecV2))
	self.registerCodec(mc.HD_V2_FullBlockReq, new(fullBlockReqCodecV2))
	self.registerCodec(mc.HD_V2_FullBlockRsp, new(fullBlockRspCodecV2))
	self.registerCodec(mc.HD_V2_MiningReq, new(miningReqCodecV2))
	self.registerCodec(mc.HD_V2_PowMiningRsp, new(powMiningRspMsgcV2))
	self.registerCodec(mc.HD_V2_AIMiningRsp, new(aiMiningRspMsgcV2))
	self.registerCodec(mc.HD_BasePowerResult, new(basePowerDifficultyMsgcV2))
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
	data, err := rlp.EncodeToBytes(rsp)
	if err != nil {
		return nil, errors.Errorf("rlp encode err: %v", err)
	}
	return data, nil
}

func (*broadcastMiningRspCodec) DecodeFn(data []byte, from common.Address) (interface{}, error) {
	msg := new(mc.HD_BroadcastMiningRspMsg)
	err := rlp.DecodeBytes(data, &msg)
	if err == nil {
		if msg.BlockMainData == nil || msg.BlockMainData.Header == nil {
			return nil, errors.Errorf("'Header' of the msg is nil")
		}
		msg.From = from
		return msg, nil

	} else {
		oldmsg := new(mc.HD_BroadcastMiningRspMsgV1)
		err = rlp.DecodeBytes(data, &oldmsg)
		if err == nil {

			if oldmsg.BlockMainData == nil || oldmsg.BlockMainData.Header == nil {
				return nil, errors.Errorf("'Header' of the msg is nil")
			}
			msg.BlockMainData = new(mc.BlockData)
			msg.BlockMainData.Header = oldmsg.BlockMainData.Header.TransferHeader()
			msg.BlockMainData.Txs = oldmsg.BlockMainData.Txs
			msg.From = from
			return msg, nil
		}
	}
	return nil, errors.Errorf("rlp decode failed: %s", err)
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
	if msg.ReqList == nil {
		return nil, errors.New("`ReqList` of msg if nil")
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

	for i := 0; i < len(msg.Votes); i++ {
		msg.Votes[i].From.Set(from)
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
	if msg == nil {
		return nil, errors.New("msg is nil")
	}
	if msg.Req == nil {
		return nil, errors.New("`req` in msg struct is nil")
	}
	msg.From = from
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
	msg := new(mc.HD_ConsensusVote)
	err := json.Unmarshal([]byte(data), msg)
	if err != nil {
		return nil, errors.Errorf("json.Unmarshal failed: %s", err)
	}
	msg.From.Set(from)
	return msg, nil
}

////////////////////////////////////////////////////////////////////////
// leader重选结果广播消息
// msg code = mc.HD_LeaderReelectBroadcast
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
	msg := new(mc.HD_ReelectBroadcastMsg)
	err := json.Unmarshal([]byte(data), msg)
	if err != nil {
		return nil, errors.Errorf("json.Unmarshal failed: %s", err)
	}
	msg.From.Set(from)
	return msg, nil
}

////////////////////////////////////////////////////////////////////////
// leader重选结果广播响应消息
// msg code = mc.HD_LeaderReelectBroadcastRsp
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
	msg := new(mc.HD_ReelectBroadcastRspMsg)
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
	data, err := rlp.EncodeToBytes(rsp)
	if err != nil {
		return nil, errors.Errorf("rlp encode failed: %v", err)
	}
	return data, nil
}

func (*fullBlockRspCodec) DecodeFn(data []byte, from common.Address) (interface{}, error) {
	msg := new(mc.HD_FullBlockRspMsg)
	err := rlp.DecodeBytes(data, &msg)
	if err == nil {
		if msg.Header == nil {
			return nil, errors.Errorf("'header' of the msg is nil")
		}
		msg.From = from
		return msg, nil
	} else {
		oldmsg := new(mc.HD_FullBlockRspMsgV1)
		err = rlp.DecodeBytes(data, &oldmsg)
		if err == nil {
			if oldmsg.Header == nil {
				return nil, errors.Errorf("'header' of the msg is nil")
			}
			msg.Header = oldmsg.Header.TransferHeader()
			msg.Txs = oldmsg.Txs
			msg.From = from
			return msg, nil
		}

	}

	return nil, errors.Errorf("rlp decode failed: %v", err)
}
