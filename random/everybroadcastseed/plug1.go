// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package everybroadcastseed

import (
	"math/big"

	"errors"

	"fmt"

	"github.com/matrix/go-matrix/baseinterface"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/random/commonsupport"
)

func init() {
	fmt.Println("everybroadcastseed plug1")
	EveryBroadcastSeedPlug1 := &EveryBroadcastSeedPlug1{}
	RegisterEveryBlockSeedPlugs("MaxNonce&Key", EveryBroadcastSeedPlug1)
}

type EveryBroadcastSeedPlug1 struct {
}

func (self *EveryBroadcastSeedPlug1) CalcSeed(hash common.Hash, support baseinterface.RandomChainSupport) (*big.Int, error) {
	ans, err := commonsupport.GetCurrentKeys(hash, support)
	if err != nil {
		return nil, errors.New("获取当前广播区块有效私钥之和")
	}

	maxNonce := big.NewInt(0)
	maxNonce.SetUint64(commonsupport.GetMaxNonce(hash, 100, support))

	ans.Add(ans, maxNonce)
	return ans, nil
}
func (self *EveryBroadcastSeedPlug1) Prepare(height uint64) error {
	log.Info(ModuleEveryBroadcastSeed, "每个广播区块产生一个随机数 准备阶段", "", "不需要处理 高度", height)
	return nil
}
