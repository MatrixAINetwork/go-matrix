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

func GetMinHash(hash common.Hash, support baseinterface.RandomChainSupport) common.Hash {

	height, err := GetNumberByHash(hash, support)
	if err != nil {
		log.Error("electionseed", "计算种子失败 err", err, "hash", hash.String())
		return common.Hash{}
	}
	minhash := hash
	BroadcastInterval := common.GetBroadcastInterval()
	for i := height - 1; i > height-BroadcastInterval; i-- {
		TempHash, err := GetHeaderHashByNumber(hash, i, support)
		if err != nil {
			break
		}
		if minhash.Big().Cmp(TempHash.Big()) == 1 { //前者大于后者
			minhash = TempHash
		}

	}
	return minhash
}

func GetNumberByHash(hash common.Hash, support baseinterface.RandomChainSupport) (uint64, error) {
	tHeader := support.BlockChain().GetHeaderByHash(hash)
	if tHeader == nil {
		log.Error("electionseed", "GetNumberByHash 根据hash算header失败 hash", hash.String())
		return 0, errors.New("根据hash算header失败")
	}
	if tHeader.Number == nil {
		log.Error("electionseed", "GetNumberByHash header 内的高度获取失败", hash.String())
		return 0, errors.New("header 内的高度获取失败")
	}
	return tHeader.Number.Uint64(), nil
}
func GetHeaderHashByNumber(hash common.Hash, height uint64, support baseinterface.RandomChainSupport) (common.Hash, error) {
	AimHash, err := support.BlockChain().GetAncestorHash(hash, height)
	if err != nil {
		log.Error("electionseed", "获取祖先hash失败 hash", hash.String(), "height", height, "err", err)
		return common.Hash{}, err
	}
	return AimHash, nil
}

//得到特殊交易
func getKeyTransInfo(hash common.Hash, Height uint64, types string, support baseinterface.RandomChainSupport) map[common.Address][]byte {
	aimHash, err := GetHeaderHashByNumber(hash, Height, support)
	if err != nil {
		log.Error("electionseed", "获取特殊交易阶段-获取祖先hash失败 hash", hash.String(), "height", Height)
		return make(map[common.Address][]byte)
	}

	ans, err := core.GetBroadcastTxs(aimHash, types)
	if err != nil {
		log.Error("electionseed", "获取特殊交易失败 Height", Height, "types", types)
	}
	return ans
}
func GetCurrentKeys(hash common.Hash, support baseinterface.RandomChainSupport) (*big.Int, error) {

	height, err := GetNumberByHash(hash, support)
	if err != nil {
		log.Error("electionseed", "计算种子失败 err", err, "hash", hash.String())
		return nil, errors.New("计算hash高度失败")
	}
	if height == 0 {
		return nil, errors.New("请求的高度不是广播高度")
	}

	broadcastInterval := common.GetBroadcastInterval()
	height_1 := height / broadcastInterval * broadcastInterval //上一个广播区块
	height_2 := height_1 - common.GetBroadcastInterval()

	PrivateMap := getKeyTransInfo(hash, height_1, mc.Privatekey, support)
	PublicMap := getKeyTransInfo(hash, height_2, mc.Publickey, support)

	return CompareMap(PrivateMap, PublicMap), nil
}

func GetMaxNonce(hash common.Hash, lastwHeight uint64, support baseinterface.RandomChainSupport) uint64 {
	height, err := GetNumberByHash(hash, support)
	if err != nil {
		log.Error("electionseed", "计算种子失败 err", err, "hash", hash.String())
		return 0
	}

	ans := support.BlockChain().GetBlockByNumber(lastwHeight).Header().Nonce.Uint64()
	lastwHeight = height - lastwHeight

	for k := lastwHeight; k <= height; k++ {
		fmt.Println("kkkk", k)
		tempHash, err := support.BlockChain().GetAncestorHash(hash, k)
		if err != nil {
			break
		}
		tempHeader := support.BlockChain().GetHeaderByHash(tempHash)
		if tempHeader == nil {
			break
		}
		temp := tempHeader.Nonce.Uint64()
		if ans < temp {
			ans = temp
		}
	}
	return ans
}
