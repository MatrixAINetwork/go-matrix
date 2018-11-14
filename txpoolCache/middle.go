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

package txpoolCache

import (
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/log"
	"sync"
)

type TxCaChe struct {
	Ntx map[uint32]*types.Transaction
	HeadHash common.Hash
	Height uint64
}
type TxCaCheListstruct struct {
	TxCaCheList []*TxCaChe
	mu sync.RWMutex
}
var TXCStruct = new(TxCaCheListstruct)
func MakeStruck(txs []*types.Transaction,hash common.Hash,h uint64){
	txc := &TxCaChe{
		Ntx : make(map[uint32]*types.Transaction),
	}
	for _,tx := range txs{
		if len(tx.N)>0{
			txc.Ntx[tx.N[0]] = tx
		}else {
			log.Info("package txpoolCache","MakeStruck()","tx`s N is nil")
		}
	}
	txc.HeadHash = hash
	txc.Height = h
	TXCStruct.mu.Lock()
	TXCStruct.TxCaCheList = append(TXCStruct.TxCaCheList,txc)
	TXCStruct.mu.Unlock()
}

func DeleteTxCache(hash common.Hash,h uint64)  {
	TXCStruct.mu.Lock()
	defer TXCStruct.mu.Unlock()
	for i,c := range TXCStruct.TxCaCheList{
		if c.Height < h{
			TXCStruct.TxCaCheList = TXCStruct.TxCaCheList[i:]
			return
		}else if c.HeadHash != hash && c.Height == h{
			TXCStruct.TxCaCheList = TXCStruct.TxCaCheList[i:]
			return
		}else {
			log.Info("package txpoolCache","DeleteTxCache()","unknown error",":c.HeadHash",c.HeadHash,"hash",hash,"c.Height",c.Height,"H",h)
		}
	}
}
//h 传过来时应该是当前区块高度，而在这存储的是下一区块的高度
func GetTxByN_Cache(listn []uint32,h uint64)map[uint32]*types.Transaction  {
	TXCStruct.mu.RLock()
	defer TXCStruct.mu.RUnlock()
	for _,txc:=range TXCStruct.TxCaCheList{
		if txc.Height == (h+1){
			return getMap(txc,listn)
		}
	}
	log.Info("package txpoolCache","GetTxByN_Cache()","Block height mismatch")
	return nil
}
func getMap(txc *TxCaChe,listn []uint32)(map[uint32]*types.Transaction)  {
	ntxmap := make(map[uint32]*types.Transaction,0)
	for _,n := range listn{
		if tx,ok := txc.Ntx[n];ok{
			ntxmap[n] = tx
		}
	}
	return ntxmap
}