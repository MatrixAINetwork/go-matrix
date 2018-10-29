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
package broadcastTx

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"sync"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/event"
	"github.com/matrix/go-matrix/internal/manapi"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
)

const (
	// chainHeadChanSize is the size of channel listening to ChainHeadEvent.
	sendBroadCastCHSize = 10
)

type BroadCast struct {
	manBackend manapi.Backend

	sendBroadCastCH chan mc.BroadCastEvent
	broadCastSub    event.Subscription
	wg              sync.WaitGroup
}

func NewBroadCast(apiBackEnd manapi.Backend) *BroadCast {

	bc := &BroadCast{
		manBackend:      apiBackEnd,
		sendBroadCastCH: make(chan mc.BroadCastEvent, sendBroadCastCHSize),
	}
	bc.broadCastSub, _ = mc.SubscribeEvent(mc.SendBroadCastTx, bc.sendBroadCastCH)
	bc.wg.Add(1)
	go bc.loop()
	return bc
}
func (bc *BroadCast) Start() {
	//go bc.loop()
}
func (bc *BroadCast) Stop() {
	bc.broadCastSub.Unsubscribe()
	bc.wg.Wait()
	log.Info("BroadCast Server stopped.--YY")
}

func (bc *BroadCast) loop() {
	defer bc.wg.Done()
	for {
		select {
		case ev := <-bc.sendBroadCastCH:
			bc.sendBroadCastTransaction(ev.Txtyps, ev.Height, ev.Data)
		case <-bc.broadCastSub.Err():
			return
		}
	}
}

//YY 广播交易的接口
func (bc *BroadCast) sendBroadCastTransaction(t string, h *big.Int, data []byte) error {
	currBlockHeight := bc.manBackend.CurrentBlock().Number()
	//TODO sunchunfeng test
	if h.Cmp(currBlockHeight) < 0 {
		log.Info("===Send BroadCastTx===", "block height less than 100")
		return errors.New("===Send BroadCastTx===,block height less than 100")
	}
	log.Info("=========YY=========", "sendBroadCastTransaction", data)
	bType := false
	if t == mc.CallTheRoll {
		bType = true
	}
	log.Info("=========YY=========11111", "sendBroadCastTransaction:hhhhhhhh", h)
	h.Quo(h, big.NewInt(int64(common.GetBroadcastInterval())))
	log.Info("=========YY=========22222", "sendBroadCastTransaction:hhhhhhhh", h)
	t += h.String()
	log.Info("=========YY=========33333", "sendBroadCastTransaction:tttttttttt", t)
	tmpData := make(map[string][]byte)
	tmpData[t] = data
	msData, _ := json.Marshal(tmpData)
	var txtype byte
	txtype = byte(1)
	tx := types.NewHeartTransaction(txtype, msData)
	var chainID *big.Int
	if config := bc.manBackend.ChainConfig(); config.IsEIP155(currBlockHeight) {
		chainID = config.ChainId
	}
	signed, err := bc.manBackend.SignTx(tx, chainID)
	if err != nil {
		log.Info("=========YY=========", "sendBroadCastTransaction:SignTx=", err)
		return err
	}
	err1 := bc.manBackend.SendBroadTx(context.Background(), signed, bType)
	log.Info("=========YY=========", "sendBroadCastTransaction:Return=", err1)
	return nil
}
