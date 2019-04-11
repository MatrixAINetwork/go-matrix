// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package everybroadcastseed

import (
	"math/big"

	"fmt"

	"github.com/MatrixAINetwork/go-matrix/baseinterface"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/params/manparams"
	"github.com/MatrixAINetwork/go-matrix/random/commonsupport"
)

func init() {
	EveryBroadcastSeedPlug1 := &EveryBroadcastSeedPlug1{}
	RegisterEveryBlockSeedPlugs(manparams.EveryBroadcastSeed_Plug_MaxNonce, EveryBroadcastSeedPlug1)
}

type EveryBroadcastSeedPlug1 struct {
}

func (self *EveryBroadcastSeedPlug1) CalcSeed(hash common.Hash, support baseinterface.RandomChainSupport) (*big.Int, error) {
	ans, err := commonsupport.GetValidVoteSum(hash, support)
	if err != nil {
		log.Error(ModuleEveryBroadcastSeed, "获取广播区块有效私钥之和失败,err", err)
		return nil, fmt.Errorf("获取广播区块有效私钥之和失败,err:%v", err)
	}
	maxNonce, err := commonsupport.GetMaxNonce(hash, support)
	if err != nil {
		log.Error(ModuleEveryBroadcastSeed, "获取最大nonce失败 err:", err, "hash", hash)
		return nil, fmt.Errorf("获取最大nonce失败,hash值:%v err:%v", hash, err)
	}
	ans.Add(ans, big.NewInt(int64(maxNonce)))
	return ans, nil
}
func (self *EveryBroadcastSeedPlug1) Prepare(height uint64, hash common.Hash) error {
	log.Info(ModuleEveryBroadcastSeed, "每个广播区块产生一个随机数 准备阶段", "", "不需要处理 高度", height)
	return nil
}
