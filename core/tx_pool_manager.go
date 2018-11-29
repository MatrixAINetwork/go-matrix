// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package core

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/matrix/go-matrix/p2p"

	"time"

	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/event"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/params"
)

var (
	ErrTxPoolAlreadyExist = errors.New("txpool already exist")
	ErrTxPoolIsNil        = errors.New("txpool is nil")
	ErrTxPoolNonexistent  = errors.New("txpool nonexistent")
)

//YY
type RetChan struct {
	//Rxs   []types.SelfTransaction
	AllTxs []*RetCallTx
	Err    error
	Resqe  int
}
type RetChan_txpool struct {
	Rxs  []types.SelfTransaction
	Err  error
	Tx_t common.TxTypeInt
}
type byteNumber struct {
	maxNum, num uint32
	mu          sync.Mutex
}

func (b3 *byteNumber) getNum() uint32 {
	if b3.num < b3.maxNum {
		b3.num++
	} else {
		b3.num = 0
	}
	return b3.num
}
func (b3 *byteNumber) catNumber() uint32 {
	b3.mu.Lock()
	defer b3.mu.Unlock()
	nodeNum, _ := ca.GetNodeNumber()
	num := b3.getNum()
	return (num << 7) + nodeNum
}

var byte3Number = &byteNumber{maxNum: 0x1ffff, num: 0}
var byte4Number = &byteNumber{maxNum: 0x1ffffff, num: 0}

// TxPoolManager
type TxPoolManager struct {
	txPoolsMutex sync.RWMutex
	once         sync.Once
	sub          event.Subscription
	txPools      map[common.TxTypeInt]TxPool
	roleChan     chan common.RoleType
	quit         chan struct{}
	addPool      chan TxPool
	delPool      chan TxPool
}

func NewTxPoolManager(config TxPoolConfig, chainconfig *params.ChainConfig, chain blockChain, path string) *TxPoolManager {
	txPoolManager := &TxPoolManager{
		txPoolsMutex: sync.RWMutex{},
		txPools:      make(map[common.TxTypeInt]TxPool),
		quit:         make(chan struct{}),
		roleChan:     make(chan common.RoleType),
		addPool:      make(chan TxPool),
		delPool:      make(chan TxPool),
	}
	go txPoolManager.loop(config, chainconfig, chain, path)
	return txPoolManager
}