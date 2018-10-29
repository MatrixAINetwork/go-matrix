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
package mc

import (
	"fmt"
	"testing"
	"time"

	"github.com/matrix/go-matrix/event"
)

type Sub struct {
	MasterMinerReElectionReqMsgCH  chan MasterMinerReElectionReqMsg
	MasterMinerReElectionReqMsgSub event.Subscription
}

func NewSub() *Sub {
	sub := &Sub{
		MasterMinerReElectionReqMsgCH: make(chan MasterMinerReElectionReqMsg, 10),
	}
	sub.MasterMinerReElectionReqMsgSub, _ = SubscribeEvent("ReElect_MasterMinerReElectionReqMsg", sub.MasterMinerReElectionReqMsgCH)
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
		err := PublishEvent("ReElect_MasterMinerReElectionReqMsg", MasterMinerReElectionReqMsg{SeqNum: 666})
		fmt.Println("Post发送状态", err)
	}
}
func TestSub(t *testing.T) {
	sub := NewSub()
	go Post()
	time.Sleep(100 * time.Second)
	fmt.Println(sub)
}
