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
package messageState

import (
	"bytes"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/rlp"
)

type testMessage struct {
	nRound uint64
}

func (tmsg *testMessage) runloop() {
	tm := time.NewTicker(time.Second / 10)
	for {
		select {
		case <-tm.C:
			atomic.AddUint64(&tmsg.nRound, 1)
		}
	}
}
func (tmsg *testMessage) checkMessage(aim mc.EventCode, value interface{}) (uint64, bool) {
	return tmsg.nRound, true
}
func (tmsg *testMessage) checkState(state []byte) bool {
	temp := []byte{1, 1, 1, 1, 1}
	return bytes.Equal(state, temp)
}
func (tmsg *testMessage) getKeyBytes(value interface{}) []byte {
	val, _ := rlp.EncodeToBytes(value)
	return val
}

type testInfo struct {
	Value uint
	Round uint
}

func TestMessageStateProcess(t *testing.T) {
	log.InitLog(3)
	testCh := &testMessage{}
	msgState := NewMessageStatePool(10, 3, testCh)
	msgState.SubscribeEvent(1, make(chan *testInfo, 5))
	msgState.SubscribeEvent(2, make(chan *testInfo, 5))
	msgState.SubscribeEvent(3, make(chan *testInfo, 5))
	msgState.SubscribeEvent(4, make(chan *testInfo, 5))
	msgState.SubscribeEvent(5, make(chan *testInfo, 5))
	msgChan := make(chan MessageSend, 5)
	msgState.SetStateChan(msgChan)
	go testCh.runloop()
	go msgState.RunLoop()
	go func(msgCh chan MessageSend) {
		for {
			select {
			case data := <-msgCh:
				log.Info("Message Data", "data", data.Message, "round", data.Round)
			}
		}
	}(msgChan)
	wg := sync.WaitGroup{}
	wg.Add(5)
	for i := 0; i < 5; i++ {
		go func(index int) {
			for i := 0; i < 100; i++ {
				mc.PublishEvent(mc.EventCode(index+1), &testInfo{uint(index), uint((index + i) * 10)})
				time.Sleep(time.Second / 50)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	req := RequireInfo{2, testCh.getKeyBytes(&testInfo{uint(2), uint((96 + 2) * 10)}), make(chan []interface{}, 10)}
	go msgState.Require(req)
	select {
	case data := <-req.requireChan:
		log.Info("get Data:", "data", data)
	}
}
func TestMessageStateProcess_Quit(t *testing.T) {
	log.InitLog(3)
	testCh := &testMessage{}
	msgState := NewMessageStatePool(10, 3, testCh)
	msgState.SubscribeEvent(1, make(chan *testInfo, 5))
	msgState.SubscribeEvent(2, make(chan *testInfo, 5))
	msgState.SubscribeEvent(3, make(chan *testInfo, 5))
	msgState.SubscribeEvent(4, make(chan *testInfo, 5))
	msgState.SubscribeEvent(5, make(chan *testInfo, 5))
	msgChan := make(chan MessageSend, 5)
	msgState.SetStateChan(msgChan)
	go testCh.runloop()
	go msgState.RunLoop()
	go func(msgCh chan MessageSend) {
		for {
			select {
			case data := <-msgCh:
				log.Info("Message Data", "data", data.Message, "round", data.Round)
			}
		}
	}(msgChan)
	msgState.Quit()
	wg := sync.WaitGroup{}
	wg.Add(5)
	for i := 0; i < 5; i++ {
		go func(index int) {
			for i := 0; i < 100; i++ {
				mc.PublishEvent(mc.EventCode(index+1), &testInfo{uint(index), uint(index * index * 10)})
				time.Sleep(time.Second / 10)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
}
