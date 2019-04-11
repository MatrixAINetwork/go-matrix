// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package commonsupport

import (
	"crypto/ecdsa"
	"errors"
	"math/big"

	"fmt"

	"github.com/MatrixAINetwork/go-matrix/accounts/keystore"
	"github.com/MatrixAINetwork/go-matrix/baseinterface"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/common/readstatedb"
	"github.com/MatrixAINetwork/go-matrix/core"
	"github.com/MatrixAINetwork/go-matrix/crypto"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/btcsuite/btcd/btcec"
)

type VoteData struct {
	PrivateData []byte
	PublicData  []byte
}

func GetCommonMap(private map[common.Address][]byte, public map[common.Address][]byte) map[common.Address]VoteData {
	commonMap := make(map[common.Address]VoteData)
	for address, privateData := range private {
		publicData, ok := public[address]
		if ok == false {
			continue
		}
		commonMap[address] = VoteData{
			PrivateData: privateData,
			PublicData:  publicData,
		}
	}
	return commonMap
}

func GetValidPrivateSum(commonMap map[common.Address]VoteData) *big.Int {
	PrivateSum := big.NewInt(0)
	all := uint64(0)
	for _, v := range commonMap {
		if CheckVoteDataIsCompare(v.PrivateData, v.PublicData) {
			PrivateSum.Add(PrivateSum, common.BytesToHash(v.PrivateData).Big())
			all++
		}
	}
	log.Trace(ModeleRandomCommon, "累加的有效私钥个数为", all)
	return PrivateSum
}

func CheckVoteDataIsCompare(private []byte, public []byte) bool {
	curve := btcec.S256()
	pk1, err := btcec.ParsePubKey(public, curve)
	if err != nil {
		log.WARN(ModeleRandomCommon, "比对公私钥数据阶段 转换公钥失败:", err)
		return false
	}
	if pk1 == nil {
		log.WARN(ModeleRandomCommon, "比对公私钥数据阶段 转换的公钥为空:", err)
		return false
	}

	pk1_1 := (*ecdsa.PublicKey)(pk1)
	if pk1_1 == nil {
		log.WARN(ModeleRandomCommon, "比对公私钥数据阶段 公钥转换失败", "空")
		return false
	}

	xx, yy := pk1_1.Curve.ScalarBaseMult(private)
	if xx.Cmp(pk1_1.X) != 0 {
		log.Warn(ModeleRandomCommon, "对比公私钥数据阶段,X值不匹配", "")
		return false
	}
	if yy.Cmp(pk1_1.Y) != 0 {
		log.Warn(ModeleRandomCommon, "对比公私要数据阶段,Y值不匹配", "")
		return false
	}
	return true
}

func GetVoteData() (*big.Int, []byte, error) {
	key, err := crypto.GenerateKey()
	if err != nil {
		log.Error(ModeleRandomCommon, "生成投票数据失败", err)
		return nil, nil, err
	}
	if key.D == nil {
		log.Error(ModeleRandomCommon, "生成投票数据失败", "D为空")
	}
	return key.D, keystore.ECDSAPKCompression(&key.PublicKey), err
}

func GetNumberByHash(hash common.Hash, support baseinterface.RandomChainSupport) (uint64, error) {
	tHeader := support.BlockChain().GetHeaderByHash(hash)
	if tHeader == nil {
		log.Error(ModeleRandomCommon, "获取Header失败", hash.String())
		return 0, errors.New("根据hash算header失败")
	}
	if tHeader.Number == nil {
		log.Error(ModeleRandomCommon, "Header高度错误", hash.String())
		return 0, errors.New("header 内的高度获取失败")
	}
	return tHeader.Number.Uint64(), nil
}

func GetAncestorHash(hash common.Hash, height uint64, support baseinterface.RandomChainSupport) (common.Hash, error) {
	aimHash, err := support.BlockChain().GetAncestorHash(hash, height)
	if err != nil {
		log.Error(ModeleRandomCommon, "获取祖先hash失败 hash", hash.String(), "height", height, "err", err)
		return common.Hash{}, err
	}
	return aimHash, nil
}

func getKeyTransInfo(root []common.CoinRoot, types string, support baseinterface.RandomChainSupport) map[common.Address][]byte {
	ans, err := core.GetBroadcastTxMap(support.BlockChain(), root, types)
	if err != nil {
		log.Error(ModeleRandomCommon, "获取特殊交易失败 root", root, "types", types)
	}
	return ans
}

func GetValidVoteSum(hash common.Hash, support baseinterface.RandomChainSupport) (*big.Int, error) {

	preBroadcastRoot, err := readstatedb.GetPreBroadcastRoot(support.BlockChain(), hash)
	if err != nil {
		log.Error(ModeleRandomCommon, "计算种子阶段,获取root值失败 err", err)
		return nil, fmt.Errorf("从状态树获取前2个广播区块root失败")
	}

	PrivateMap := getKeyTransInfo(preBroadcastRoot.LastStateRoot, mc.Privatekey, support)

	PublicMap := getKeyTransInfo(preBroadcastRoot.BeforeLastStateRoot, mc.Publickey, support)

	commonMap := GetCommonMap(PrivateMap, PublicMap)
	MapAns := GetValidPrivateSum(commonMap)
	return MapAns, nil
}
