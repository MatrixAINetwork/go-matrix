// Copyright (c) 2018-2019 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package electionseed

import (
	"math/big"

	"github.com/matrix/go-matrix/baseinterface"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/params/manparams"
	"github.com/matrix/go-matrix/random/commonsupport"
)

func init() {
	minHash := &MinHashPlug{privatekey: big.NewInt(0)}
	RegisterElectSeedPlugs(manparams.ElectionSeed_Plug_MinHash, minHash)
}

type MinHashPlug struct {
	privatekey *big.Int
}

func (self *MinHashPlug) Prepare(height uint64, support baseinterface.RandomChainSupport) error {
	data, err := commonsupport.GetElectGenTimes(support.BlockChain(), height)
	if err != nil {
		log.ERROR(ModuleElectSeed, "获取通用配置失败 err", err)
		return err
	}
	voteBeforeTime := uint64(data.VoteBeforeTime)
	bcInterval := manparams.NewBCInterval()
	if bcInterval.IsBroadcastNumber(height+voteBeforeTime) == false {
		log.INFO(ModuleElectSeed, "RoleUpdateMsgHandle", "当前不是投票点,忽略")
		return nil
	}
	if CanVote(height) == false {
		log.WARN(ModuleElectSeed, "不需要投票  账户不存在抵押交易 高度", height)
		return nil
	}
	privatekey, publickeySend, err := commonsupport.GetVoteData()
	if err != nil {
		log.INFO(ModuleElectSeed, "获取投票数据失败 err", err)
		return err
	}
	privatekeySend := common.BigToHash(self.privatekey).Bytes()

	log.INFO(ModuleElectSeed, "投票高度", (height + voteBeforeTime))
	mc.PublishEvent(mc.SendBroadCastTx, mc.BroadCastEvent{Txtyps: mc.Publickey, Height: big.NewInt(int64(height + voteBeforeTime)), Data: publickeySend})
	mc.PublishEvent(mc.SendBroadCastTx, mc.BroadCastEvent{Txtyps: mc.Privatekey, Height: big.NewInt(int64(height + voteBeforeTime)), Data: privatekeySend})

	self.privatekey = privatekey
	return nil
}

func (self *MinHashPlug) CalcSeed(hash common.Hash, support baseinterface.RandomChainSupport) (*big.Int, error) {
	SeedSum, err := commonsupport.GetValidVoteSum(hash, support)
	if err != nil {
		log.ERROR(ModuleElectSeed, "计算阶段", "", "获取有效私钥出错 err", err)
		return nil, err
	}
	log.Info(ModuleElectSeed, "计算阶段,有效私钥之和", SeedSum)
	minHash, err := commonsupport.GetMinHash(hash, support)
	if err != nil {
		log.Error(ModuleElectSeed, "计算阶段,获取最小hash错误 err", err)
		return nil, err
	}
	SeedSum.Add(SeedSum, minHash.Big())
	log.INFO(ModuleElectSeed, "计算阶段", "", "计算结果为", SeedSum, "高度hash", hash.String())
	return SeedSum, nil
}

func CanVote(height uint64) bool {
	depositList, err := commonsupport.GetDepositListByHeightAndRole(big.NewInt(int64(height)), common.RoleValidator)
	if err != nil {
		log.Error(ModuleElectSeed, "投票失敗", "获取验证者身份列表失败", "高度", height)
		return false
	}
	selfAddress := commonsupport.GetSelfAddress()
	for _, v := range depositList {
		if v.Address == selfAddress {
			log.Info(ModuleElectSeed, "具备投票身份 账户", selfAddress)
			return true
		}
	}
	log.Info(ModuleElectSeed, "不具备投票身份,不存在抵押列表里 账户", selfAddress)
	return false
}
