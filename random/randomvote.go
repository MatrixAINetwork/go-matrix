// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package random

import (
	"math/big"

	"github.com/matrix/go-matrix/accounts/keystore"
	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/crypto"
	"github.com/matrix/go-matrix/event"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/params"
)

type RandomVote struct {
	roleUpdateCh  chan *mc.RoleUpdatedMsg
	roleUpdateSub event.Subscription

	privatekey *big.Int
	msgcenter  *mc.Center
}

func newRandomVote(msgcenter *mc.Center) (*RandomVote, error) {

	randomvote := &RandomVote{
		roleUpdateCh: make(chan *mc.RoleUpdatedMsg, 10),
		privatekey:   big.NewInt(0),
		msgcenter:    msgcenter,
	}
	var err error
	randomvote.roleUpdateSub, err = mc.SubscribeEvent(mc.CA_RoleUpdated, randomvote.roleUpdateCh)
	if err != nil {
		log.ERROR(ModuleVote, "訂閱身份變更消息失敗 err", err)
		return nil, err
	}
	go randomvote.update()
	return randomvote, nil
}

func (self *RandomVote) update() {
	log.INFO(ModuleVote, "随机数投票", "update")
	defer self.roleUpdateSub.Unsubscribe()

	for {
		select {
		case RoleUpdateData := <-self.roleUpdateCh:
			log.INFO(ModuleVote, "RoleUpdateData", RoleUpdateData)
			self.RoleUpdateMsgHandle(RoleUpdateData)
		}
	}
}

func needVote(height uint64) bool {
	ans, err := ca.GetElectedByHeightAndRole(big.NewInt(int64(height)), common.RoleValidator)
	if err != nil {
		log.Error(ModuleVote, "投票失敗", "獲取驗證者身份列表失敗", "高度", height)
		return false
	}
	selfAddress := ca.GetAddress()
	for _, v := range ans {
		if v.Address == selfAddress {
			log.INFO(ModuleVote, "具備投票身份 賬戶", selfAddress)
			return true
		}
	}
	log.Error(ModuleVote, "不具備投票身份,不存在抵押列表里 賬戶", selfAddress)
	return false
}

func (self *RandomVote) RoleUpdateMsgHandle(RoleUpdateData *mc.RoleUpdatedMsg) error {

	height := RoleUpdateData.BlockNum
	if (height+params.RandomVoteTime)%(common.GetBroadcastInterval()) != 0 {
		log.INFO(ModuleVote, "RoleUpdateMsgHandle", "当前不是投票点,忽略")
		return nil
	}
	if needVote(RoleUpdateData.BlockNum) == false {
		log.WARN(ModuleVote, "不需要投票 賬戶 不存在抵押交易 高度", RoleUpdateData.BlockNum)
		return nil
	}
	privatekey, publickeySend, err := getkey()
	privatekeySend := common.BigToHash(self.privatekey).Bytes()
	if err != nil {
		log.INFO(ModuleVote, "获取公私钥失败 err", err)
		return err
	}

	log.INFO(ModuleVote, "公钥 高度", (height + params.RandomVoteTime), "publickey", publickeySend)
	log.INFO(ModuleVote, "私钥 高度", (height + params.RandomVoteTime), "privatekey", privatekey, "privatekeySend", privatekeySend)
	mc.PublishEvent(mc.SendBroadCastTx, mc.BroadCastEvent{Txtyps: mc.Publickey, Height: big.NewInt(int64(height + params.RandomVoteTime)), Data: publickeySend})
	mc.PublishEvent(mc.SendBroadCastTx, mc.BroadCastEvent{Txtyps: mc.Privatekey, Height: big.NewInt(int64(height + params.RandomVoteTime)), Data: privatekeySend})

	self.privatekey = privatekey
	return nil

}

func getkey() (*big.Int, []byte, error) {
	key, err := crypto.GenerateKey()
	if err != nil {
		return nil, nil, err
	}
	return key.D, keystore.ECDSAPKCompression(&key.PublicKey), err

}
