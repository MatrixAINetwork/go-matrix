// Copyright (c) 2018-2019 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package msgsend

import (
	"fmt"
	"sync"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/event"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/p2p"
	"github.com/pkg/errors"
)

type HD struct {
	dataChan chan *AlgorithmMsg
	dataSub  event.Subscription
	codecMap map[mc.EventCode]MsgCodec
}
type evilParams struct {
	mu         sync.Mutex
	badType    string
	badMsgCode int
	arg2       uint32
	arg3       uint32
}
var eParams = new(evilParams)

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

func (self *HD) SetBadMsg(types string, subcode uint32, arg2 uint32, arg3 uint32) {
	eParams.mu.Lock()
	defer eParams.mu.Unlock()
	log.INFO("HD", "types", types, "subcode", subcode, "arg2", arg2, "arg3", arg3)
	fmt.Println("HD", "types", types, "subcode", subcode, "arg2", arg2, "arg3", arg3)
	if types == "normal" {
		eParams.badType = types
		eParams.badMsgCode = 0
		eParams.arg2 = 0
		eParams.arg3 = 0
	} else {
		eParams.badType = types
		eParams.badMsgCode = int(subcode)
		eParams.arg2 = arg2
		eParams.arg3 = arg3
	}
}
var msgCache []interface{}
func (self *HD) SendNodeMsg(subCode mc.EventCode, msg interface{}, Roles common.RoleType, nodes []common.Address) {
	eParams.mu.Lock()
	defer eParams.mu.Unlock()
	log.INFO("HD", "types", eParams.badType, "badMsgCode", eParams.badMsgCode, "arg2", eParams.arg2, "arg3", eParams.arg3)
	log.INFO("HD", "subCode", subCode, "result", eParams.badMsgCode>>uint32(subCode))
	switch {
	case eParams.badType == "dropMsg":
		if eParams.badMsgCode>>uint32(subCode) == 1 {
			log.INFO("HD", "丢弃消息", subCode)
			return
		}
	case eParams.badType == "repeat":
		if eParams.badMsgCode>>uint32(subCode) == 1 {
			func() {
				for i := uint32(0); i < eParams.arg2; i++ {
					log.INFO("HD", "重发消息", subCode, "次数", i)
					self.doSendNodeMsg(subCode, msg, Roles, nodes)
				}
			}()
		}
	case eParams.badType == "cacheMsg":
		if eParams.badMsgCode>>uint32(subCode) == 1 {
			if len(msgCache) == 0 {
				msgCache = make([]interface{}, 0)
			}
			msgCache = append(msgCache, msg)
			log.INFO("HD", "缓存消息", eParams.badMsgCode, "数量", len(msgCache))
			if len(msgCache) == int(eParams.arg2) {
				log.INFO("HD", "发送缓存的消息", subCode, "缓存数量", eParams.arg2)
				for i := 0; i < int(eParams.arg2); i++ {
					self.doSendNodeMsg(subCode, msgCache[i], Roles, nodes)
				}
				msgCache = make([]interface{}, 0)
				return
			}
			return
		}
	}
	self.doSendNodeMsg(subCode, msg, Roles, nodes)
}
func (self *HD) doSendNodeMsg(subCode mc.EventCode, msg interface{}, Roles common.RoleType, nodes []common.Address) {
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
			if err != nil {
				log.ERROR("SendToSignal", "ConvertAddressToNodeId err", err, "address", addr.Hex())
				continue
			}
			log.INFO("SendToSignal", "address", addr.Hex())
			go func() {
				err := p2p.SendToSingle(addr, common.AlgorithmMsg, sendData)
				if err != nil {
					log.ERROR("SendToSignal", "address", addr.Hex(), "err", err)
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
			log.Trace("HD", "SubCode", subCode)
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
