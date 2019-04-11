// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package messageState

import (
	"bytes"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/rlp"
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
