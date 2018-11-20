// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package commonsupport

import (
	"crypto/ecdsa"
	"errors"
	"math/big"

	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/matrix/go-matrix/accounts/keystore"
	"github.com/matrix/go-matrix/baseinterface"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core"
	"github.com/matrix/go-matrix/crypto"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
)

const (
	ModuleRandomCommon = "随机数服务公共组件"
)

func CompareMap(private map[common.Address][]byte, public map[common.Address][]byte) *big.Int {
	if len(private) > len(public) {
		return rangePrivate(private, public)
	}
	ans := rangePublic(private, public)
	log.INFO(ModuleRandomCommon, "隨機數map匹配的公私鑰 data", ans)
	return ans
}

func rangePrivate(privateMap map[common.Address][]byte, publicMap map[common.Address][]byte) *big.Int {
	ans := big.NewInt(0)
	for address, privateV := range privateMap {
		publicV, ok := publicMap[address]
		if false == ok {
			continue
		}
		if Compare(privateV, publicV) {
			anst := common.BytesToHash(privateV).Big()
			ans.Add(ans, anst)
		}
	}
	return ans

}
func rangePublic(privateMap map[common.Address][]byte, publicMap map[common.Address][]byte) *big.Int {
	ans := big.NewInt(0)
	for adress, publicV := range publicMap {
		PrivateV, ok := privateMap[adress]

		if false == ok {
			continue
		}
		if Compare(PrivateV, publicV) {
			anst := common.BytesToHash(PrivateV).Big()
			ans.Add(ans, anst)
		}
	}
	return ans
}

func Compare(private []byte, public []byte) bool {
	curve := btcec.S256()
	pk1, err := btcec.ParsePubKey(public, curve)
	if err != nil {
		return false
	}

	pk1_1 := (*ecdsa.PublicKey)(pk1)
	xx, yy := pk1_1.Curve.ScalarBaseMult(private)
	if xx.Cmp(pk1_1.X) != 0 {
		return false
	}
	if yy.Cmp(pk1_1.Y) != 0 {
		return false
	}
	return true
}
func Getkey() (*big.Int, []byte, error) {
	key, err := crypto.GenerateKey()
	if err != nil {
		return nil, nil, err
	}
	return key.D, keystore.ECDSAPKCompression(&key.PublicKey), err

}

func GetKeyTransInfo(Height uint64, types string) map[common.Address][]byte {
	//asd()
	ans, err := core.GetBroadcastTxs(big.NewInt(int64(Height)), types)
	if err != nil {
		log.Error(ModuleRandomCommon, "获取特殊交易失败 Height", Height, "types", types)
	}
	return ans
}

func GetHashByNum(height uint64, bc *core.BlockChain) common.Hash {
	return bc.GetBlockByNumber(height).Hash()
}
func GetMinHash(height uint64, bc *core.BlockChain) common.Hash {

	minhash := GetHashByNum(height, bc)
	BroadcastInterval := common.GetBroadcastInterval()
	for i := height - 1; i > height-BroadcastInterval; i-- {
		fmt.Println("GetMinHash-i", i)
		blockhash := GetHashByNum(uint64(i), bc)
		if minhash.Big().Cmp(blockhash.Big()) == 1 { //前者大于后者
			minhash = blockhash
		}

	}
	return minhash
}

func GetCurrentKeys(data uint64) (*big.Int, error) {
	if common.IsBroadcastNumber(data) == false || data == 0 {
		return nil, errors.New("请求的高度不是广播高度")
	}

	LastBroadCastHeight := data - common.GetBroadcastInterval() //上一个广播区块
	PrivateMap := GetKeyTransInfo(LastBroadCastHeight, mc.Privatekey)
	PublicMap := GetKeyTransInfo(data, mc.Publickey)
	return CompareMap(PrivateMap, PublicMap), nil
}

func GetMaxNonce(nowHeight uint64, lastwHeight uint64, support baseinterface.RandomChainSupport) uint64 {
	fmt.Println("nowHeight", nowHeight, "lastHeight", lastwHeight)
	ans := support.BlockChain().GetBlockByNumber(lastwHeight).Header().Nonce.Uint64()
	fmt.Println("nowHeight-128", lastwHeight)
	fmt.Println("GetMaxNonce-129", ans)

	for k := lastwHeight; k <= nowHeight; k++ {
		fmt.Println("kkkk", k)
		temp := support.BlockChain().GetBlockByNumber(k).Header().Nonce.Uint64()
		if ans < temp {
			ans = temp
		}
	}
	return ans
}
