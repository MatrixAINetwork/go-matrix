// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package msgsend

import (
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
			log.Trace("HD", "SubCode", subCode, "from", data.Account.Hex())
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
