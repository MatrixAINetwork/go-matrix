// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package mc

import (
	"fmt"
	"testing"
	"time"

	"github.com/MatrixAINetwork/go-matrix/event"
)

type Sub struct {
	MasterMinerReElectionReqMsgCH  chan MasterMinerReElectionReqMsg
	MasterMinerReElectionReqMsgSub event.Subscription
}

func NewSub() *Sub {
	sub := &Sub{
		MasterMinerReElectionReqMsgCH: make(chan MasterMinerReElectionReqMsg, 10),
	}
	//sub.MasterMinerReElectionReqMsgSub, _ = SubscribeEvent("ReElect_MasterMinerReElectionReqMsg", sub.MasterMinerReElectionReqMsgCH)
	go sub.update()
	return sub
}

func (self *Sub) update() {
	for {
		select {
		case data := <-self.MasterMinerReElectionReqMsgCH:
			fmt.Println("收到数据", data)
		}
	}
}

func Post() {
	for {
		time.Sleep(5 * time.Second)
		//err := PublishEvent("ReElect_MasterMinerReElectionReqMsg", MasterMinerReElectionReqMsg{SeqNum: 666})
		//fmt.Println("Post发送状态", err)
	}
}
func TestSub(t *testing.T) {
	sub := NewSub()
	go Post()
	time.Sleep(100 * time.Second)
	fmt.Println(sub)
}
