// Copyright (c) 2018-2019 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package broadcastTx

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"sync"

	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/event"
	"github.com/matrix/go-matrix/internal/manapi"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/params/manparams"
	"time"
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
	log.Info("BroadCast Server stopped.")
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

// 广播交易的接口
func (bc *BroadCast) sendBroadCastTransaction(t string, h *big.Int, data []byte) error {
	currBlock := bc.manBackend.CurrentBlock()
	currBlockHeight := currBlock.Number()
	//TODO sunchunfeng test
	if h.Cmp(currBlockHeight) < 0 {
		log.Info("===Send BroadCastTx===", "block height less than 100")
		return errors.New("===Send BroadCastTx===,block height less than 100")
	}
	bType := false
	if t == mc.CallTheRoll {
		bType = true
	}

	bcInterval, err := manparams.GetBCIntervalInfoByNumber(currBlockHeight.Uint64())
	if err != nil || bcInterval == nil {
		log.Info("===Send BroadCastTx===", "get broadcast interval err", err)
	}
	h.Quo(h, big.NewInt(int64(bcInterval.GetBroadcastInterval())))
	t += h.String()
	tmpData := make(map[string][]byte)
	tmpData[t] = data
	msData, _ := json.Marshal(tmpData)
	var txtype byte
	txtype = byte(1)
	tx := types.NewBroadCastTransaction(txtype, msData)
	var chainID *big.Int
	if config := bc.manBackend.ChainConfig(); config.IsEIP155(currBlockHeight) {
		chainID = config.ChainId
	}
	t1 := time.Now()
	usingEntrust := ca.GetRole() != common.RoleBroadcast
	signed, err := bc.manBackend.SignTx(tx, chainID, currBlock.ParentHash(), bcInterval.GetNextBroadcastNumber(currBlockHeight.Uint64()), usingEntrust)
	if err != nil {
		log.Info("file broadcast", "sendBroadCastTransaction:SignTx=", err)
		return err
	}
	t2 := time.Since(t1)
	err1 := bc.manBackend.SendBroadTx(context.Background(), signed, bType)
	t3 := time.Since(t1)
	log.Info("File BroadCast", "func sendBroadCastTransaction:t2", t2, "t3", t3)
	log.Info("==================", "sendBroadCastTransaction:Return=", err1)
	return nil
}
