// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package everyblockseed

import (
	"math/big"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/mc"

	"fmt"

	"github.com/matrix/go-matrix/baseinterface"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/random/commonsupport"
)

func init() {
	fmt.Println("everyblockseed plug1")
	EveryBlockSeedPlug1 := &EveryBlockSeedPlug1{privatekey: big.NewInt(0)}
	RegisterLotterySeedPlugs("Nonce&Address&Coinbase", EveryBlockSeedPlug1)
}

type EveryBlockSeedPlug1 struct {
	privatekey *big.Int
}

func (self *EveryBlockSeedPlug1) CalcSeed(height uint64, support baseinterface.RandomChainSupport) (*big.Int, error) {
	ans := big.NewInt(0)
	ans.SetUint64(support.BlockChain().GetBlockByNumber(height).Nonce())
	selfAddress := support.BlockChain().GetBlockByNumber(height).Header().Leader

	prv := GetCurrentPrivateKey(height) //TODO:获取当前区块私钥
	KeyData := big.NewInt(0)
	for k := height - 1; k >= 0; k-- {
		if IsBlockOwner(support, k, selfAddress) {
			fmt.Println(ModuleEveryBlockSeed, "计算阶段", "", "找到上个区块高度", k, "当前区块高度", height)
			pub := GetCurrentPublicKey(k) //TODO:获取上一出块的区块的公钥
			if commonsupport.Compare(prv, pub) == true {
				log.INFO(ModuleEveryBlockSeed, "计算阶段", "", "找到上个区块高度", k, "当前区块高度", height, "公私钥配对", "")
				KeyData = common.BytesToHash(prv).Big()
			} else {
				fmt.Println(ModuleEveryBlockSeed, "计算阶段", "", "找到上个区块高度", k, "当前区块高度", height, "公私钥不配对 上一次出块的公钥", pub, "当前块私钥", prv)
			}
			break
		}
	}
	ans.Add(ans, selfAddress.Big())
	ans.Add(ans, KeyData)
	return ans, nil
}
func (self *EveryBlockSeedPlug1) Prepare(height uint64) error {
	log.Info(ModuleEveryBlockSeed, "准备阶段", "", "高度", height)
	privatekey, publickeySend, err := commonsupport.Getkey()
	log.Info(ModuleEveryBlockSeed, "准备阶段", "", "生成的公钥", publickeySend)
	log.Info(ModuleEveryBlockSeed, "准备阶段", "", "生成的私钥", common.BigToHash(privatekey).Bytes())
	privatekeySend := common.BigToHash(self.privatekey).Bytes()
	if err != nil {
		return err
	}
	self.privatekey = privatekey
	log.INFO(ModuleEveryBlockSeed, "准备阶段", "", "公钥", publickeySend)
	log.INFO(ModuleEveryBlockSeed, "准备阶段", "", "私钥", privatekeySend)
	return mc.PublishEvent(mc.EveryBlockSeedRsp, mc.EveryBlockSeedRspMsg{PublicKey: publickeySend, Private: privatekeySend})
}

func IsBlockOwner(support baseinterface.RandomChainSupport, height uint64, address common.Address) bool {
	if support.BlockChain().GetBlockByNumber(height).Header().Leader == address {
		return true
	} else {
		return false
	}
}
func GetCurrentPrivateKey(height uint64) []byte {
	return []byte{}
}
func GetCurrentPublicKey(height uint64) []byte {
	return []byte{}
}
