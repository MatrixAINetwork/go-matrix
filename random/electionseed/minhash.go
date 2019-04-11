// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package electionseed

import (
	"math/big"

	"github.com/MatrixAINetwork/go-matrix/baseinterface"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/params/manparams"
	"github.com/MatrixAINetwork/go-matrix/random/commonsupport"
)

func init() {
	minHash := &MinHashPlug{privatekey: big.NewInt(0)}
	RegisterElectSeedPlugs(manparams.ElectionSeed_Plug_MinHash, minHash)
}

type MinHashPlug struct {
	privatekey *big.Int
}

func (self *MinHashPlug) Prepare(height uint64, hash common.Hash, support baseinterface.RandomChainSupport) error {
	data, err := commonsupport.GetElectGenTimes(support.BlockChain(), hash)
	if err != nil {
		log.ERROR(ModuleElectSeed, "获取通用配置失败:", err)
		return err
	}
	voteBeforeTime := uint64(data.VoteBeforeTime)
	bcInterval := manparams.GetBCIntervalInfo()
	if bcInterval.IsBroadcastNumber(height+voteBeforeTime) == false {
		return nil
	}
	if CanVote(hash) == false {
		return nil
	}
	privatekey, publickeySend, err := commonsupport.GetVoteData()
	if err != nil {
		log.ERROR(ModuleElectSeed, "获取投票数据失败:", err)
		return err
	}
	privatekeySend := common.BigToHash(self.privatekey).Bytes()

	log.TRACE(ModuleElectSeed, "投票:", (height + voteBeforeTime))
	mc.PublishEvent(mc.SendBroadCastTx, mc.BroadCastEvent{Txtyps: mc.Publickey, Height: big.NewInt(int64(height + voteBeforeTime)), Data: publickeySend})
	mc.PublishEvent(mc.SendBroadCastTx, mc.BroadCastEvent{Txtyps: mc.Privatekey, Height: big.NewInt(int64(height + voteBeforeTime)), Data: privatekeySend})

	self.privatekey = privatekey
	return nil
}

func (self *MinHashPlug) CalcSeed(hash common.Hash, support baseinterface.RandomChainSupport) (*big.Int, error) {
	SeedSum, err := commonsupport.GetValidVoteSum(hash, support)
	if err != nil {
		log.ERROR(ModuleElectSeed, "获取数据错误:", err)
		return nil, err
	}
	minHash, err := commonsupport.GetMinHash(hash, support)
	if err != nil {
		log.ERROR(ModuleElectSeed, "获取最小hash错误:", err)
		return nil, err
	}
	SeedSum.Add(SeedSum, minHash.Big())
	return SeedSum, nil
}

func CanVote(hash common.Hash) bool {
	depositList, err := commonsupport.GetDepositListByHeightAndRole(hash, common.RoleValidator)
	if err != nil {
		log.Error(ModuleElectSeed, "投票失敗", "获取验证者身份列表失败", "hash", hash)
		return false
	}
	selfAddress := commonsupport.GetSelfAddress()
	for _, v := range depositList {
		if v.Address == selfAddress {
			return true
		}
	}
	return false
}
