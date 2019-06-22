// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package txpoolCache

import (
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/log"
	"sync"
)

type TxCaChe struct {
	Ntx      map[uint32]types.SelfTransaction
	HeadHash common.Hash
	Height   uint64
}
type TxCaCheListstruct struct {
	TxCaCheList []*TxCaChe
	mu          sync.RWMutex
}

var TXCStruct = new(TxCaCheListstruct)

func MakeStruck(txser []types.SelfTransaction, hash common.Hash, h uint64) {
	txc := &TxCaChe{
		Ntx: make(map[uint32]types.SelfTransaction),
	}
	for _, tx := range txser {
		if tx.GetTxNLen() > 0 {
			txc.Ntx[tx.GetTxN(0)] = tx
		} else {
			log.Info("package txpoolCache", "MakeStruck()", "tx`s N is nil")
		}
	}
	txc.HeadHash = hash
	txc.Height = h
	TXCStruct.mu.Lock()
	TXCStruct.TxCaCheList = append(TXCStruct.TxCaCheList, txc)
	TXCStruct.mu.Unlock()
}

func DeleteTxCache(hash common.Hash, h uint64) {
	TXCStruct.mu.Lock()
	defer TXCStruct.mu.Unlock()
	for i, c := range TXCStruct.TxCaCheList {
		if c.Height < h {
			TXCStruct.TxCaCheList = TXCStruct.TxCaCheList[i:]
			return
		} else if c.HeadHash != hash && c.Height == h {
			TXCStruct.TxCaCheList = TXCStruct.TxCaCheList[i:]
			return
		} else {
			log.Info("package txpoolCache", "DeleteTxCache()", "unknown error", ":c.HeadHash", c.HeadHash, "hash", hash, "c.Height", c.Height, "H", h)
		}
	}
}

//h 传过来时应该是当前区块高度，而在这存储的是下一区块的高度
func GetTxByN_Cache(listn []uint32, h uint64) map[uint32]types.SelfTransaction {
	TXCStruct.mu.RLock()
	defer TXCStruct.mu.RUnlock()
	for _, txc := range TXCStruct.TxCaCheList {
		if txc.Height == (h + 1) {
			return getMap(txc, listn)
		}
	}
	log.Info("package txpoolCache", "GetTxByN_Cache()", "Block height mismatch")
	return nil
}
func getMap(txc *TxCaChe, listn []uint32) map[uint32]types.SelfTransaction {
	ntxmap := make(map[uint32]types.SelfTransaction, 0)
	for _, n := range listn {
		if tx, ok := txc.Ntx[n]; ok {
			ntxmap[n] = tx
		}
	}
	return ntxmap
}
