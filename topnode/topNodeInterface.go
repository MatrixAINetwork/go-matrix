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
package topnode

import (
	"github.com/matrix/go-matrix/accounts/signhelper"
	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/event"
	"github.com/matrix/go-matrix/hd"
	"github.com/matrix/go-matrix/mc"
)

type OnlineState uint8

const (
	Online OnlineState = iota + 1
	Offline
)

func (o OnlineState) String() string {
	switch o {
	case Online:
		return "在线"
	case Offline:
		return "下线"
	default:
		return "未知状态"
	}
}

type NodeOnLineInfo struct {
	Address     common.Address
	OnlineState [30]uint8
}

type TopNodeStateInterface interface {
	GetTopNodeOnlineState() []NodeOnLineInfo
}

type ValidatorAccountInterface interface {
	SignWithValidate(hash []byte, validate bool) (sig common.Signature, err error)
	IsSelfAddress(addr common.Address) bool
}

type MessageSendInterface interface {
	SendNodeMsg(subCode mc.EventCode, msg interface{}, Roles common.RoleType, address []common.Address)
}

type MessageCenterInterface interface {
	SubscribeEvent(aim mc.EventCode, ch interface{}) (event.Subscription, error)
	PublishEvent(aim mc.EventCode, data interface{}) error
}

////////////////////////////////////////////////////////////////////
type TopNodeInstance struct {
	signHelper *signhelper.SignHelper
	hd         *hd.HD
}

func NewTopNodeInstance(sh *signhelper.SignHelper, hd *hd.HD) *TopNodeInstance {
	return &TopNodeInstance{
		signHelper: sh,
		hd:         hd,
	}
}

func (self *TopNodeInstance) GetTopNodeOnlineState() []NodeOnLineInfo {
	return nil
}

func (self *TopNodeInstance) SignWithValidate(hash []byte, validate bool) (sig common.Signature, err error) {
	return self.signHelper.SignHashWithValidate(hash, validate)
}

func (self *TopNodeInstance) IsSelfAddress(addr common.Address) bool {
	return ca.GetAddress() == addr
}

func (self *TopNodeInstance) SendNodeMsg(subCode mc.EventCode, msg interface{}, Roles common.RoleType, address []common.Address) {
	self.hd.SendNodeMsg(subCode, msg, Roles, address)
}

func (self *TopNodeInstance) SubscribeEvent(aim mc.EventCode, ch interface{}) (event.Subscription, error) {
	return mc.SubscribeEvent(aim, ch)
}

func (self *TopNodeInstance) PublishEvent(aim mc.EventCode, data interface{}) error {
	return mc.PublishEvent(aim, data)
}
