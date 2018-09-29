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
package hd

import (
	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/event"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/p2p"
	"github.com/pkg/errors"
)

type HD struct {
	dataChan chan *AlgorithmMsg
	dataSub  event.Subscription
	codecMap map[mc.EventCode]MsgCodec
}

func NewHD() (*HD, error) {
	hd := &HD{
		dataChan: make(chan *AlgorithmMsg, 10),
		codecMap: make(map[mc.EventCode]MsgCodec),
	}
	//订阅网络消息
	var err error
	hd.dataSub, err = mc.SubscribeEvent(mc.P2P_HDMSG, hd.dataChan)
	if err != nil {
		return nil, err
	}
	//初始化编解码器
	hd.initCodec()
	//run
	go hd.receive()

	return hd, nil
}

func (self *HD) SendNodeMsg(subCode mc.EventCode, msg interface{}, Roles common.RoleType, nodes []common.Address) {
	codec, err := self.findCodec(subCode)
	if err != nil {
		log.ERROR("HD", "send findCodec err", err)
		return
	}

	data, err := codec.EncodeFn(msg)
	if err != nil {
		log.ERROR("HD", "EncodeFn err", err, "subCode", subCode)
		return
	}

	sendData := NetData{
		SubCode: uint32(subCode),
		Msg:     data,
	}

	if nodes == nil {
		log.INFO("SendToGroup", "roles", Roles.String(), "SubCode", subCode)
		go p2p.SendToGroup(Roles, common.AlgorithmMsg, sendData)
	} else {
		log.INFO("SendToSignal", "total address count", len(nodes), "SubCode", subCode)
		for _, addr := range nodes {
			sendNode, err := ca.ConvertAddressToNodeId(addr)
			if err != nil {
				log.ERROR("SendToSignal", "ConvertAddressToNodeId err", err, "address", addr.Hex())
				continue
			}
			log.INFO("SendToSignal", "address", addr.Hex(), "node id", sendNode)
			go func() {
				err := p2p.SendToSingle(sendNode, common.AlgorithmMsg, sendData)
				if err != nil {
					log.ERROR("SendToSignal", "address", addr.Hex(), "node id", sendNode, "err", err)
				}
			}()
		}
	}
}

func (self *HD) receive() {
	for {
		select {
		case data := <-self.dataChan:
			subCode := mc.EventCode(data.Data.SubCode)
			log.INFO("HD", "SubCode", subCode)
			codec, err := self.findCodec(subCode)
			if err != nil {
				log.ERROR("HD", "receive findCodec err", err)
				break
			}
			msg, err := codec.DecodeFn(data.Data.Msg, data.Account)
			if err != nil {
				log.ERROR("HD", "DecodeFn err", err, "subCode", subCode, "from", data.Account.Hex())
				break
			}
			mc.PublishEvent(subCode, msg)
		}
	}
}

func (self *HD) registerCodec(subCode mc.EventCode, codec MsgCodec) {
	_, exist := self.codecMap[subCode]
	if exist {
		log.ERROR("HD", "注册编解码器失败, 已存在的消息码", subCode)
		return
	}
	self.codecMap[subCode] = codec
}

func (self *HD) findCodec(subCode mc.EventCode) (MsgCodec, error) {
	codec, OK := self.codecMap[subCode]
	if !OK {
		return nil, errors.Errorf("消息码[%v]的编解码器不存在", subCode)
	}
	return codec, nil
}
