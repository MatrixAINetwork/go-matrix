// Copyright 2018 The MATRIX Authors as well as Copyright 2014-2017 The go-ethereum Authors
// This file is consisted of the MATRIX library and part of the go-ethereum library.
//
// The MATRIX-ethereum library is free software: you can redistribute it and/or modify it under the terms of the MIT License.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, 
//and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject tothe following conditions:
//
//The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
//
//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, 
//WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISINGFROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE
//OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
package verifier

import (
	"errors"
	"github.com/matrix/go-matrix/accounts/signhelper"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/consensus"
	"github.com/matrix/go-matrix/core"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/hd"
	"github.com/matrix/go-matrix/mc"
)

var (
	ErrMsgAccountIsNull  = errors.New("不合法的账户：空账户")
	ErrValidatorsIsNil   = errors.New("验证者列表为空")
	ErrValidatorNotFound = errors.New("验证者未找到")
	ErrNoMsgInCache      = errors.New("缓存中没有目标消息")
	ErrMsgIsNil          = errors.New("消息为nil")
	ErrSelfReqIsNil      = errors.New("self请求不在缓存中")
	ErrBroadcastIsNil    = errors.New("缓存没有广播消息")
	ErrPOSResultIsNil    = errors.New("POS结果为nil/header为nil")
	ErrLeaderResultIsNil = errors.New("leader共识结果为nil")
)

type Matrix interface {
	BlockChain() *core.BlockChain
	SignHelper() *signhelper.SignHelper
	DPOSEngine() consensus.DPOSEngine
	Engine() consensus.Engine
	HD() *hd.HD
	FetcherNotify(hash common.Hash, number uint64)
}

type state uint8

const (
	stIdle state = iota
	stPos
	stReelect
	stMining
)

func (s state) String() string {
	switch s {
	case stIdle:
		return "未运行阶段"
	case stPos:
		return "POS阶段"
	case stReelect:
		return "重选阶段"
	case stMining:
		return "挖矿结果等待阶段"
	default:
		return "未知状态"
	}
}

type leaderData struct {
	leader     common.Address
	nextLeader common.Address
}

func (self *leaderData) copyData() *leaderData {
	newData := &leaderData{
		leader:     common.Address{},
		nextLeader: common.Address{},
	}

	newData.leader.Set(self.leader)
	newData.nextLeader.Set(self.nextLeader)
	return newData
}

type startControllerMsg struct {
	role         common.RoleType
	validators   []mc.TopologyNodeInfo
	parentHeader *types.Header
}

type sendNewBlockReadyRsp struct {
	repHash   common.Hash
	target    common.Address
	rspNumber uint64
}
