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

////////////////////////////////////////////////////////////////////////
// 重选询问请求消息
type lrInquiryReqCodecV2 struct {
}

func (*lrInquiryReqCodecV2) EncodeFn(msg interface{}) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, errors.Errorf("json.Marshal failed: %s", err)
	}
	return data, nil
}

func (*lrInquiryReqCodecV2) DecodeFn(data []byte, from common.Address) (interface{}, error) {
	msg := new(mc.HD_V2_ReelectInquiryReqMsg)
	err := json.Unmarshal([]byte(data), msg)
	if err != nil {
		return nil, errors.Errorf("json.Unmarshal failed: %s", err)
	}
	msg.From.Set(from)
	return msg, nil
}

////////////////////////////////////////////////////////////////////////
// 重选询问响应消息
type lrInquiryRspCodecV2 struct {
}

func (*lrInquiryRspCodecV2) EncodeFn(msg interface{}) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, errors.Errorf("json.Marshal failed: %s", err)
	}
	return data, nil
}

func (*lrInquiryRspCodecV2) DecodeFn(data []byte, from common.Address) (interface{}, error) {
	msg := new(mc.HD_V2_ReelectInquiryRspMsg)
	err := json.Unmarshal([]byte(data), msg)
	if err != nil {
		return nil, errors.Errorf("json.Unmarshal failed: %s", err)
	}
	msg.From.Set(from)
	return msg, nil
}

////////////////////////////////////////////////////////////////////////
// leader重选请求消息
type lrReqCodecV2 struct {
}

func (*lrReqCodecV2) EncodeFn(msg interface{}) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, errors.Errorf("json.Marshal failed: %s", err)
	}
	return data, nil
}

func (*lrReqCodecV2) DecodeFn(data []byte, from common.Address) (interface{}, error) {
	msg := new(mc.HD_V2_ReelectLeaderReqMsg)
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
type lrVoteCodecV2 struct {
}

func (*lrVoteCodecV2) EncodeFn(msg interface{}) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, errors.Errorf("json.Marshal failed: %s", err)
	}
	return data, nil
}

func (*lrVoteCodecV2) DecodeFn(data []byte, from common.Address) (interface{}, error) {
	msg := new(mc.HD_V2_ConsensusVote)
	err := json.Unmarshal([]byte(data), msg)
	if err != nil {
		return nil, errors.Errorf("json.Unmarshal failed: %s", err)
	}
	msg.From.Set(from)
	return msg, nil
}

////////////////////////////////////////////////////////////////////////
// leader重选结果广播消息
type lrResultBCCodecV2 struct {
}

func (*lrResultBCCodecV2) EncodeFn(msg interface{}) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, errors.Errorf("json.Marshal failed: %s", err)
	}
	return data, nil
}

func (*lrResultBCCodecV2) DecodeFn(data []byte, from common.Address) (interface{}, error) {
	msg := new(mc.HD_V2_ReelectBroadcastMsg)
	err := json.Unmarshal([]byte(data), msg)
	if err != nil {
		return nil, errors.Errorf("json.Unmarshal failed: %s", err)
	}
	msg.From.Set(from)
	return msg, nil
}

////////////////////////////////////////////////////////////////////////
// leader重选结果广播响应消息
type lrResultBCRspCodecV2 struct {
}

func (*lrResultBCRspCodecV2) EncodeFn(msg interface{}) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, errors.Errorf("json.Marshal failed: %s", err)
	}
	return data, nil
}

func (*lrResultBCRspCodecV2) DecodeFn(data []byte, from common.Address) (interface{}, error) {
	msg := new(mc.HD_V2_ReelectBroadcastRspMsg)
	err := json.Unmarshal([]byte(data), msg)
	if err != nil {
		return nil, errors.Errorf("json.Unmarshal failed: %s", err)
	}
	msg.From.Set(from)
	return msg, nil
}

////////////////////////////////////////////////////////////////////////
// 完整区块获取请求
// msg code = mc.HD_V2_FullBlockReq
type fullBlockReqCodecV2 struct {
}

func (*fullBlockReqCodecV2) EncodeFn(msg interface{}) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, errors.Errorf("json.Marshal failed: %s", err)
	}
	return data, nil
}

func (*fullBlockReqCodecV2) DecodeFn(data []byte, from common.Address) (interface{}, error) {
	msg := new(mc.HD_V2_FullBlockReqMsg)
	err := json.Unmarshal([]byte(data), msg)
	if err != nil {
		return nil, errors.Errorf("json.Unmarshal failed: %s", err)
	}
	msg.From.Set(from)
	return msg, nil
}

////////////////////////////////////////////////////////////////////////
// 完整区块获取响应
// msg code = mc.HD_V2_FullBlockRsp
type fullBlockRspCodecV2 struct {
}

func (*fullBlockRspCodecV2) EncodeFn(msg interface{}) ([]byte, error) {
	rsp, OK := msg.(*mc.HD_V2_FullBlockRspMsg)
	if !OK {
		return nil, errors.New("reflect err! HD_V2_FullBlockRspMsg")
	}
	data, err := rlp.EncodeToBytes(rsp)
	if err != nil {
		return nil, errors.Errorf("rlp encode failed: %v", err)
	}
	return data, nil
}

func (*fullBlockRspCodecV2) DecodeFn(data []byte, from common.Address) (interface{}, error) {
	msg := new(mc.HD_V2_FullBlockRspMsg)
	err := rlp.DecodeBytes(data, &msg)
	if err != nil {
		return nil, errors.Errorf("rlp decode failed: %v", err)
	}
	if msg.Header == nil {
		return nil, errors.Errorf("'header' of the msg is nil")
	}
	msg.From = from
	return msg, nil
}

////////////////////////////////////////////////////////////////////////
// V2挖矿请求消息
// msg code = mc.HD_V2_MiningReq
type miningReqCodecV2 struct {
}

func (*miningReqCodecV2) EncodeFn(msg interface{}) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, errors.Errorf("json.Marshal failed: %s", err)
	}
	return data, nil
}

func (*miningReqCodecV2) DecodeFn(data []byte, from common.Address) (interface{}, error) {
	msg := new(mc.HD_V2_MiningReqMsg)
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
// V2 pow挖矿响应
// msg code = mc.HD_V2_PowMiningRsp
type powMiningRspMsgcV2 struct {
}

func (*powMiningRspMsgcV2) EncodeFn(msg interface{}) ([]byte, error) {
	rsp, OK := msg.(*mc.HD_V2_PowMiningRspMsg)
	if !OK {
		return nil, errors.New("reflect err! HD_V2_PowMiningRspMsg")
	}
	data, err := rlp.EncodeToBytes(rsp)
	if err != nil {
		return nil, errors.Errorf("rlp encode failed: %v", err)
	}
	return data, nil
}

func (*powMiningRspMsgcV2) DecodeFn(data []byte, from common.Address) (interface{}, error) {
	msg := new(mc.HD_V2_PowMiningRspMsg)
	err := rlp.DecodeBytes(data, &msg)
	if err != nil {
		return nil, errors.Errorf("rlp decode failed: %v", err)
	}
	msg.From = from
	return msg, nil
}

////////////////////////////////////////////////////////////////////////
// V2 ai挖矿响应
// msg code = mc.HD_V2_AIMiningRsp
type aiMiningRspMsgcV2 struct {
}

func (*aiMiningRspMsgcV2) EncodeFn(msg interface{}) ([]byte, error) {
	rsp, OK := msg.(*mc.HD_V2_AIMiningRspMsg)
	if !OK {
		return nil, errors.New("reflect err! HD_V2_AIMiningRspMsg")
	}
	data, err := rlp.EncodeToBytes(rsp)
	if err != nil {
		return nil, errors.Errorf("rlp encode failed: %v", err)
	}
	return data, nil
}

func (*aiMiningRspMsgcV2) DecodeFn(data []byte, from common.Address) (interface{}, error) {
	msg := new(mc.HD_V2_AIMiningRspMsg)
	err := rlp.DecodeBytes(data, &msg)
	if err != nil {
		return nil, errors.Errorf("rlp decode failed: %v", err)
	}
	msg.From = from
	return msg, nil
}

////////////////////////////////////////////////////////////////////////
// 算力检测
// msg code = mc.HD_BasePowerResult
type basePowerDifficultyMsgcV2 struct {
}

func (*basePowerDifficultyMsgcV2) EncodeFn(msg interface{}) ([]byte, error) {
	rsp, OK := msg.(*mc.HD_BasePowerDifficulty)
	if !OK {
		return nil, errors.New("reflect err! HD_BasePowerDifficulty")
	}
	data, err := rlp.EncodeToBytes(rsp)
	if err != nil {
		return nil, errors.Errorf("rlp encode failed: %v", err)
	}
	return data, nil
}

func (*basePowerDifficultyMsgcV2) DecodeFn(data []byte, from common.Address) (interface{}, error) {
	msg := new(mc.HD_BasePowerDifficulty)
	err := rlp.DecodeBytes(data, &msg)
	if err != nil {
		return nil, errors.Errorf("rlp decode failed: %v", err)
	}
	msg.From = from
	return msg, nil
}
