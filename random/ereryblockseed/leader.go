// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package everyblockseed

import (
	"math/big"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/mc"

	"errors"

	"github.com/MatrixAINetwork/go-matrix/baseinterface"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/random/commonsupport"
)

func init() {
	EveryBlockSeedPlug1 := &EveryBlockSeedPlug1{privatekey: big.NewInt(0)}
	RegisterLotterySeedPlugs("Nonce&Address&Coinbase", EveryBlockSeedPlug1)
}

type EveryBlockSeedPlug1 struct {
	privatekey *big.Int
}

func (self *EveryBlockSeedPlug1) CalcSeed(hash common.Hash, support baseinterface.RandomChainSupport) (*big.Int, error) {
	currentHeader := support.BlockChain().GetHeaderByHash(hash)
	if currentHeader == nil {
		log.ERROR(ModulePreBlockSeed, "获取区块头失败:", hash.String())
		return nil, errors.New("根据hash获取区块头失败")
	}
	preBlockSeedSum := big.NewInt(0)
	preBlockSeedSum.SetUint64(currentHeader.Nonce.Uint64())
	preBlockSeedSum.Add(preBlockSeedSum, currentHeader.Leader.Big())
	return preBlockSeedSum, nil
}
func (self *EveryBlockSeedPlug1) Prepare(height uint64, hash common.Hash) error {
	privatekey, publickeySend, err := commonsupport.GetVoteData()
	if err != nil {
		log.ERROR(ModulePreBlockSeed, "获取投票数据失败:", err)
		return err
	}
	privatekeySend := common.BigToHash(self.privatekey).Bytes()
	self.privatekey = privatekey
	return mc.PublishEvent(mc.EveryBlockSeedRsp, mc.EveryBlockSeedRspMsg{PublicKey: publickeySend, Private: privatekeySend})
}

func GetCurrentPrivateKey(hash common.Hash) []byte {
	return []byte{}
}
func GetCurrentPublicKey(hash common.Hash) []byte {
	return []byte{}
}
