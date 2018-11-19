// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package electionseed

import (
	"math/big"

	"fmt"

	"github.com/matrix/go-matrix/baseinterface"
	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/random/commonsupport"
	"github.com/matrix/go-matrix/params/manparams"
)

func init() {
	fmt.Println("electionseed Minhash&Key")
	electSeedPlug1 := &ElectSeedPlug1{privatekey: big.NewInt(0)}
	RegisterElectSeedPlugs("Minhash&Key", electSeedPlug1)
}

type ElectSeedPlug1 struct {
	privatekey *big.Int
}

func (self *ElectSeedPlug1) Prepare(height uint64) error {
	if (height+manparams.RandomVoteTime)%(common.GetBroadcastInterval()) != 0 {
		log.INFO(ModuleElectSeed, "RoleUpdateMsgHandle", "当前不是投票点,忽略")
		return nil
	}
	if NeedVote(height) == false {
		log.WARN(ModuleElectSeed, "不需要投票 賬戶 不存在抵押交易 高度", height)
		return nil
	}
	privatekey, publickeySend, err := commonsupport.Getkey()
	privatekeySend := common.BigToHash(self.privatekey).Bytes()
	if err != nil {
		log.INFO(ModuleElectSeed, "获取公私钥失败 err", err)
		return err
	}

	log.INFO(ModuleElectSeed, "公钥 高度", (height + manparams.RandomVoteTime), "publickey", publickeySend)
	log.INFO(ModuleElectSeed, "私钥 高度", (height + manparams.RandomVoteTime), "privatekey", common.BigToHash(privatekey).Bytes(), "privatekeySend", privatekeySend)
	mc.PublishEvent(mc.SendBroadCastTx, mc.BroadCastEvent{Txtyps: mc.Publickey, Height: big.NewInt(int64(height + manparams.RandomVoteTime)), Data: publickeySend})
	mc.PublishEvent(mc.SendBroadCastTx, mc.BroadCastEvent{Txtyps: mc.Privatekey, Height: big.NewInt(int64(height + manparams.RandomVoteTime)), Data: privatekeySend})

	self.privatekey = privatekey
	return nil
}

func (self *ElectSeedPlug1) CalcSeed(hash common.Hash, support baseinterface.RandomChainSupport) (*big.Int, error) {
	ans, err := commonsupport.GetCurrentKeys(hash, support)
	if err != nil {
		log.ERROR(ModuleElectSeed, "计算阶段", "", "获取有效私钥出错 err", err)
		return nil, err
	}
	minHash := commonsupport.GetMinHash(hash, support)
	ans.Add(ans, minHash.Big())
	log.INFO(ModuleElectSeed, "计算阶段", "", "计算结果未", ans, "高度hash", hash.String())
	return ans, nil
}

func NeedVote(height uint64) bool {
	ans, err := ca.GetElectedByHeightAndRole(big.NewInt(int64(height)), common.RoleValidator)
	if err != nil {
		log.Error(ModuleElectSeed, "投票失敗", "獲取驗證者身份列表失敗", "高度", height)
		return false
	}
	selfAddress := ca.GetAddress()
	for _, v := range ans {
		if v.Address == selfAddress {
			log.INFO(ModuleElectSeed, "具備投票身份 賬戶", selfAddress)
			return true
		}
	}
	log.Error(ModuleElectSeed, "不具備投票身份,不存在抵押列表里 賬戶", selfAddress)
	return false
}
