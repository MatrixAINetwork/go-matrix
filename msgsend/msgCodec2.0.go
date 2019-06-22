// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package msgsend

import (
	"encoding/json"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/mc"
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
