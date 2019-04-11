// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package core

import (
	//"encoding/json"
	"encoding/json"
	"math/big"
	"testing"

	"github.com/MatrixAINetwork/go-matrix/core/matrixstate"
	"github.com/MatrixAINetwork/go-matrix/params/manparams"

	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/mandb"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/params"

	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/reward/cfg"

	"github.com/MatrixAINetwork/go-matrix/common"
)

type FakeTxsReward struct {
	reward map[string]map[common.Address]*big.Int
}

func (f *FakeTxsReward) CalcNodesRewards(blockReward *big.Int, Leader common.Address, num uint64, parentHash common.Hash, coinType string) map[common.Address]*big.Int {
	return f.reward[coinType]
}
func (f *FakeTxsReward) setNodesRewards(types string, coinReward map[common.Address]*big.Int) {
	f.reward[types] = coinReward
}
func (f *FakeTxsReward) CalcValidatorRewards(Leader common.Address, num uint64) map[common.Address]*big.Int {
	return nil
}

func (f *FakeTxsReward) CalcMinerRewards(num uint64, parentHash common.Hash) map[common.Address]*big.Int {
	return nil
}

func (f *FakeTxsReward) CalcMinerRateMount(blockReward *big.Int) (*big.Int, *big.Int, *big.Int) {
	return nil, nil, nil
}
func (f *FakeTxsReward) CalcValidatorRateMount(blockReward *big.Int) (*big.Int, *big.Int, *big.Int) {
	return nil, nil, nil

}
func (f *FakeTxsReward) GetRewardCfg() *cfg.RewardCfg {
	return nil
}

func TestStateProcessor_onlyman(t *testing.T) {
	//只有man奖励测试

	log.InitLog(3)
	chaindb := mandb.NewMemDatabase()
	roots := make([]common.CoinRoot, 0)
	roots = append(roots, common.CoinRoot{Cointyp: params.MAN_COIN, Root: common.Hash{}})
	preState, _ := state.NewStateDBManage(roots, chaindb, state.NewDatabase(chaindb))
	matrixstate.SetVersionInfo(preState, manparams.VersionAlpha)
	var coinlist []string
	data, _ := json.Marshal(coinlist)
	preState.SetMatrixData(types.RlpHash((params.COIN_NAME)), data)

	var coincfglist []common.CoinConfig
	data, _ = json.Marshal(coincfglist)

	preState.SetMatrixData(types.RlpHash(common.COINPREFIX+mc.MSCurrencyConfig), data)

	preState.SetBalance(params.MAN_COIN, common.MainAccount, common.TxGasRewardAddress, new(big.Int).SetUint64(16e18))

	//data2 := preState.GetBalance("A", common.TxGasRewardAddress)
	processor := &StateProcessor{}
	usedGas := make(map[string]*big.Int)
	usedGas[params.MAN_COIN] = new(big.Int).SetUint64(1e8)

	rewardlist := make([]common.RewarTx, 0)
	txs := &FakeTxsReward{make(map[string]map[common.Address]*big.Int)}
	ManMap := make(map[common.Address]*big.Int)
	ManMap[common.HexToAddress("1")] = big.NewInt(1)
	ManMap[common.HexToAddress("2")] = big.NewInt(2)
	ManMap[common.HexToAddress("3")] = big.NewInt(3)
	ManMap[common.HexToAddress("4")] = big.NewInt(4)
	ManMap[common.HexToAddress("5")] = big.NewInt(5)
	txs.setNodesRewards(params.MAN_COIN, ManMap)

	//txs.CalcNodesRewards(new(big.Int).SetUint64(1e10), common.HexToAddress("11"), 1, common.Hash{}, params.MAN_COIN)
	out := processor.processMultiCoinReward(usedGas, preState, txs, &types.Header{Number: new(big.Int).SetUint64(1), Leader: common.HexToAddress("11"), ParentHash: common.Hash{}}, rewardlist)
	if 1 != len(out) {
		t.Error("输出结果错误", len(out))
	}
	for _, v := range out {
		if v.CoinType != v.CoinRange {
			t.Error("发放币种错误")
		}
	}
}

func TestStateProcessor_ProcessMultiCoinReward(t *testing.T) {
	//多币种奖励测试

	log.InitLog(3)
	chaindb := mandb.NewMemDatabase()
	roots := make([]common.CoinRoot, 0)
	roots = append(roots, common.CoinRoot{Cointyp: params.MAN_COIN, Root: common.Hash{}})
	roots = append(roots, common.CoinRoot{Cointyp: "A", Root: common.Hash{}})
	roots = append(roots, common.CoinRoot{Cointyp: "B", Root: common.Hash{}})
	roots = append(roots, common.CoinRoot{Cointyp: "C", Root: common.Hash{}})
	preState, _ := state.NewStateDBManage(roots, chaindb, state.NewDatabase(chaindb))
	matrixstate.SetVersionInfo(preState, manparams.VersionAlpha)
	var coinlist []string
	coinlist = append(coinlist, "A")
	coinlist = append(coinlist, "B")
	coinlist = append(coinlist, "C")
	data, _ := json.Marshal(coinlist)
	preState.SetMatrixData(types.RlpHash((params.COIN_NAME)), data)

	var coincfglist []common.CoinConfig
	coincfglist = append(coincfglist, common.CoinConfig{CoinRange: "A", CoinType: "A", CoinAddress: common.TxGasRewardAddress})
	coincfglist = append(coincfglist, common.CoinConfig{CoinRange: "B", CoinType: "B", CoinAddress: common.TxGasRewardAddress})
	coincfglist = append(coincfglist, common.CoinConfig{CoinRange: "C", CoinType: "C", CoinAddress: common.TxGasRewardAddress})
	data, _ = json.Marshal(coincfglist)

	preState.SetMatrixData(types.RlpHash(common.COINPREFIX+mc.MSCurrencyConfig), data)

	preState.SetBalance(params.MAN_COIN, common.MainAccount, common.TxGasRewardAddress, new(big.Int).SetUint64(16e18))
	preState.SetBalance("A", common.MainAccount, common.TxGasRewardAddress, new(big.Int).SetUint64(16e18))
	preState.SetBalance("B", common.MainAccount, common.TxGasRewardAddress, new(big.Int).SetUint64(16e18))
	preState.SetBalance("C", common.MainAccount, common.TxGasRewardAddress, new(big.Int).SetUint64(16e18))

	//data2 := preState.GetBalance("A", common.TxGasRewardAddress)
	processor := &StateProcessor{}
	usedGas := make(map[string]*big.Int)
	usedGas[params.MAN_COIN] = new(big.Int).SetUint64(1e8)
	usedGas["A"] = new(big.Int).SetUint64(1e8)
	usedGas["B"] = new(big.Int).SetUint64(1e8)
	usedGas["C"] = new(big.Int).SetUint64(1e8)

	rewardlist := make([]common.RewarTx, 0)
	txs := &FakeTxsReward{make(map[string]map[common.Address]*big.Int)}
	ManMap := make(map[common.Address]*big.Int)
	ManMap[common.HexToAddress("1")] = big.NewInt(1)
	ManMap[common.HexToAddress("2")] = big.NewInt(2)
	ManMap[common.HexToAddress("3")] = big.NewInt(3)
	ManMap[common.HexToAddress("4")] = big.NewInt(4)
	ManMap[common.HexToAddress("5")] = big.NewInt(5)
	txs.setNodesRewards(params.MAN_COIN, ManMap)
	AMap := make(map[common.Address]*big.Int)
	AMap[common.HexToAddress("11")] = big.NewInt(11)
	AMap[common.HexToAddress("12")] = big.NewInt(12)
	AMap[common.HexToAddress("13")] = big.NewInt(13)
	AMap[common.HexToAddress("14")] = big.NewInt(14)
	AMap[common.HexToAddress("15")] = big.NewInt(15)

	txs.setNodesRewards("A", AMap)

	BMap := make(map[common.Address]*big.Int)
	BMap[common.HexToAddress("11")] = big.NewInt(11)
	BMap[common.HexToAddress("12")] = big.NewInt(12)
	BMap[common.HexToAddress("13")] = big.NewInt(13)
	BMap[common.HexToAddress("14")] = big.NewInt(14)
	BMap[common.HexToAddress("15")] = big.NewInt(15)

	txs.setNodesRewards("B", BMap)

	CMap := make(map[common.Address]*big.Int)
	CMap[common.HexToAddress("11")] = big.NewInt(11)
	CMap[common.HexToAddress("12")] = big.NewInt(12)
	CMap[common.HexToAddress("13")] = big.NewInt(13)
	CMap[common.HexToAddress("14")] = big.NewInt(14)
	CMap[common.HexToAddress("15")] = big.NewInt(15)

	txs.setNodesRewards("C", CMap)
	//txs.CalcNodesRewards(new(big.Int).SetUint64(1e10), common.HexToAddress("11"), 1, common.Hash{}, params.MAN_COIN)
	out := processor.processMultiCoinReward(usedGas, preState, txs, &types.Header{Number: new(big.Int).SetUint64(1), Leader: common.HexToAddress("11"), ParentHash: common.Hash{}}, rewardlist)
	if 4 != len(out) {
		t.Error("输出结果错误", len(out))
	}
	for _, v := range out {
		if v.CoinType != v.CoinRange {
			t.Error("发放币种错误")
		}
	}
}

func TestStateProcessor_ProcessMultiCoinReward2(t *testing.T) {
	//多币种奖励都用man币发放测试

	log.InitLog(3)
	chaindb := mandb.NewMemDatabase()
	roots := make([]common.CoinRoot, 0)
	roots = append(roots, common.CoinRoot{Cointyp: params.MAN_COIN, Root: common.Hash{}})
	roots = append(roots, common.CoinRoot{Cointyp: "A", Root: common.Hash{}})
	roots = append(roots, common.CoinRoot{Cointyp: "B", Root: common.Hash{}})
	roots = append(roots, common.CoinRoot{Cointyp: "C", Root: common.Hash{}})
	preState, _ := state.NewStateDBManage(roots, chaindb, state.NewDatabase(chaindb))
	matrixstate.SetVersionInfo(preState, manparams.VersionAlpha)
	var coinlist []string
	coinlist = append(coinlist, "A")
	coinlist = append(coinlist, "B")
	coinlist = append(coinlist, "C")
	data, _ := json.Marshal(coinlist)
	preState.SetMatrixData(types.RlpHash((params.COIN_NAME)), data)

	var coincfglist []common.CoinConfig
	coincfglist = append(coincfglist, common.CoinConfig{CoinRange: "A", CoinType: params.MAN_COIN, CoinAddress: common.TxGasRewardAddress})
	coincfglist = append(coincfglist, common.CoinConfig{CoinRange: "B", CoinType: params.MAN_COIN, CoinAddress: common.TxGasRewardAddress})
	coincfglist = append(coincfglist, common.CoinConfig{CoinRange: "C", CoinType: params.MAN_COIN, CoinAddress: common.TxGasRewardAddress})
	data, _ = json.Marshal(coincfglist)

	preState.SetMatrixData(types.RlpHash(common.COINPREFIX+mc.MSCurrencyConfig), data)

	preState.SetBalance(params.MAN_COIN, common.MainAccount, common.TxGasRewardAddress, new(big.Int).SetUint64(16e18))
	preState.SetBalance("A", common.MainAccount, common.TxGasRewardAddress, new(big.Int).SetUint64(16e18))
	preState.SetBalance("B", common.MainAccount, common.TxGasRewardAddress, new(big.Int).SetUint64(16e18))
	preState.SetBalance("C", common.MainAccount, common.TxGasRewardAddress, new(big.Int).SetUint64(16e18))

	//data2 := preState.GetBalance("A", common.TxGasRewardAddress)
	processor := &StateProcessor{}
	usedGas := make(map[string]*big.Int)
	usedGas[params.MAN_COIN] = new(big.Int).SetUint64(1e8)
	usedGas["A"] = new(big.Int).SetUint64(1e8)
	usedGas["B"] = new(big.Int).SetUint64(1e8)
	usedGas["C"] = new(big.Int).SetUint64(1e8)

	rewardlist := make([]common.RewarTx, 0)
	txs := &FakeTxsReward{make(map[string]map[common.Address]*big.Int)}
	ManMap := make(map[common.Address]*big.Int)
	ManMap[common.HexToAddress("1")] = big.NewInt(1)
	ManMap[common.HexToAddress("2")] = big.NewInt(2)
	ManMap[common.HexToAddress("3")] = big.NewInt(3)
	ManMap[common.HexToAddress("4")] = big.NewInt(4)
	ManMap[common.HexToAddress("5")] = big.NewInt(5)
	txs.setNodesRewards(params.MAN_COIN, ManMap)
	AMap := make(map[common.Address]*big.Int)
	AMap[common.HexToAddress("11")] = big.NewInt(11)
	AMap[common.HexToAddress("12")] = big.NewInt(12)
	AMap[common.HexToAddress("13")] = big.NewInt(13)
	AMap[common.HexToAddress("14")] = big.NewInt(14)
	AMap[common.HexToAddress("15")] = big.NewInt(15)

	txs.setNodesRewards("A", AMap)

	BMap := make(map[common.Address]*big.Int)
	BMap[common.HexToAddress("11")] = big.NewInt(11)
	BMap[common.HexToAddress("12")] = big.NewInt(12)
	BMap[common.HexToAddress("13")] = big.NewInt(13)
	BMap[common.HexToAddress("14")] = big.NewInt(14)
	BMap[common.HexToAddress("15")] = big.NewInt(15)

	txs.setNodesRewards("B", BMap)

	CMap := make(map[common.Address]*big.Int)
	CMap[common.HexToAddress("11")] = big.NewInt(11)
	CMap[common.HexToAddress("12")] = big.NewInt(12)
	CMap[common.HexToAddress("13")] = big.NewInt(13)
	CMap[common.HexToAddress("14")] = big.NewInt(14)
	CMap[common.HexToAddress("15")] = big.NewInt(15)

	txs.setNodesRewards("C", CMap)
	//txs.CalcNodesRewards(new(big.Int).SetUint64(1e10), common.HexToAddress("11"), 1, common.Hash{}, params.MAN_COIN)
	out := processor.processMultiCoinReward(usedGas, preState, txs, &types.Header{Number: new(big.Int).SetUint64(1), Leader: common.HexToAddress("11"), ParentHash: common.Hash{}}, rewardlist)
	if 4 != len(out) {
		t.Error("输出结果错误", len(out))
	}
	for _, v := range out {
		if v.CoinType != params.MAN_COIN {
			t.Error("发放币种错误", v.CoinType)
		}
	}
}

func TestStateProcessor_ProcessMultiCoinReward3(t *testing.T) {
	//多币种奖励都用部分man币，部分自己的币种发放测试

	log.InitLog(3)
	chaindb := mandb.NewMemDatabase()
	roots := make([]common.CoinRoot, 0)
	roots = append(roots, common.CoinRoot{Cointyp: params.MAN_COIN, Root: common.Hash{}})
	roots = append(roots, common.CoinRoot{Cointyp: "A", Root: common.Hash{}})
	roots = append(roots, common.CoinRoot{Cointyp: "B", Root: common.Hash{}})
	roots = append(roots, common.CoinRoot{Cointyp: "C", Root: common.Hash{}})
	preState, _ := state.NewStateDBManage(roots, chaindb, state.NewDatabase(chaindb))
	matrixstate.SetVersionInfo(preState, manparams.VersionAlpha)
	var coinlist []string
	coinlist = append(coinlist, "A")
	coinlist = append(coinlist, "B")
	coinlist = append(coinlist, "C")
	data, _ := json.Marshal(coinlist)
	preState.SetMatrixData(types.RlpHash((params.COIN_NAME)), data)

	var coincfglist []common.CoinConfig
	coincfglist = append(coincfglist, common.CoinConfig{CoinRange: "A", CoinType: "A", CoinAddress: common.TxGasRewardAddress})
	coincfglist = append(coincfglist, common.CoinConfig{CoinRange: "B", CoinType: params.MAN_COIN, CoinAddress: common.TxGasRewardAddress})
	coincfglist = append(coincfglist, common.CoinConfig{CoinRange: "C", CoinType: params.MAN_COIN, CoinAddress: common.TxGasRewardAddress})
	data, _ = json.Marshal(coincfglist)

	preState.SetMatrixData(types.RlpHash(common.COINPREFIX+mc.MSCurrencyConfig), data)

	preState.SetBalance(params.MAN_COIN, common.MainAccount, common.TxGasRewardAddress, new(big.Int).SetUint64(16e18))
	preState.SetBalance("A", common.MainAccount, common.TxGasRewardAddress, new(big.Int).SetUint64(16e18))
	preState.SetBalance("B", common.MainAccount, common.TxGasRewardAddress, new(big.Int).SetUint64(16e18))
	preState.SetBalance("C", common.MainAccount, common.TxGasRewardAddress, new(big.Int).SetUint64(16e18))

	//data2 := preState.GetBalance("A", common.TxGasRewardAddress)
	processor := &StateProcessor{}
	usedGas := make(map[string]*big.Int)
	usedGas[params.MAN_COIN] = new(big.Int).SetUint64(1e8)
	usedGas["A"] = new(big.Int).SetUint64(1e8)
	usedGas["B"] = new(big.Int).SetUint64(1e8)
	usedGas["C"] = new(big.Int).SetUint64(1e8)

	rewardlist := make([]common.RewarTx, 0)
	txs := &FakeTxsReward{make(map[string]map[common.Address]*big.Int)}
	ManMap := make(map[common.Address]*big.Int)
	ManMap[common.HexToAddress("1")] = big.NewInt(1)
	ManMap[common.HexToAddress("2")] = big.NewInt(2)
	ManMap[common.HexToAddress("3")] = big.NewInt(3)
	ManMap[common.HexToAddress("4")] = big.NewInt(4)
	ManMap[common.HexToAddress("5")] = big.NewInt(5)
	txs.setNodesRewards(params.MAN_COIN, ManMap)
	AMap := make(map[common.Address]*big.Int)
	AMap[common.HexToAddress("11")] = big.NewInt(11)
	AMap[common.HexToAddress("12")] = big.NewInt(12)
	AMap[common.HexToAddress("13")] = big.NewInt(13)
	AMap[common.HexToAddress("14")] = big.NewInt(14)
	AMap[common.HexToAddress("15")] = big.NewInt(15)

	txs.setNodesRewards("A", AMap)

	BMap := make(map[common.Address]*big.Int)
	BMap[common.HexToAddress("11")] = big.NewInt(11)
	BMap[common.HexToAddress("12")] = big.NewInt(12)
	BMap[common.HexToAddress("13")] = big.NewInt(13)
	BMap[common.HexToAddress("14")] = big.NewInt(14)
	BMap[common.HexToAddress("15")] = big.NewInt(15)

	txs.setNodesRewards("B", BMap)

	CMap := make(map[common.Address]*big.Int)
	CMap[common.HexToAddress("11")] = big.NewInt(11)
	CMap[common.HexToAddress("12")] = big.NewInt(12)
	CMap[common.HexToAddress("13")] = big.NewInt(13)
	CMap[common.HexToAddress("14")] = big.NewInt(14)
	CMap[common.HexToAddress("15")] = big.NewInt(15)

	txs.setNodesRewards("C", CMap)
	//txs.CalcNodesRewards(new(big.Int).SetUint64(1e10), common.HexToAddress("11"), 1, common.Hash{}, params.MAN_COIN)
	out := processor.processMultiCoinReward(usedGas, preState, txs, &types.Header{Number: new(big.Int).SetUint64(1), Leader: common.HexToAddress("11"), ParentHash: common.Hash{}}, rewardlist)
	if 4 != len(out) {
		t.Error("输出结果错误", len(out))
	}
	for _, v := range out {
		if v.CoinRange == "A" {
			if v.CoinType != "A" {
				t.Error("发放币种错误", v.CoinType)
			}
			continue

		}
		if v.CoinType != params.MAN_COIN {
			t.Error("发放币种错误", v.CoinType)
		}
	}
}
