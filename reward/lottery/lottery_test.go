// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package lottery

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strconv"
	"testing"

	"bou.ke/monkey"
	"github.com/MatrixAINetwork/go-matrix/core/matrixstate"
	"github.com/MatrixAINetwork/go-matrix/crypto"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/params/manparams"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/params"
)

type Chain struct {
	blockCache map[uint64]*types.Block
}

type randSeed struct {
}

func (r *randSeed) GetRandom(hash common.Hash, Type string) (*big.Int, error) {

	return big.NewInt(2000), nil
}

type State struct {
	balance int64
}

func (st *State) GetBalance(addr common.Address) common.BalanceType {
	return []common.BalanceSlice{{common.MainAccount, big.NewInt(st.balance)}}
}

func (st *State) GetMatrixData(hash common.Hash) (val []byte) {
	return nil

}
func (st *State) SetMatrixData(hash common.Hash, val []byte) {
	return
}
func (chain *Chain) New(num uint64) {
	header := &types.Header{}
	txs := make([]types.SelfTransaction, 0)
	key1, _ := crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	key2, _ := crypto.HexToECDSA("8a1f9a8f95be41cd7ccb6168179afb4504aefe388d1e14474d32c45c72ce7b7a")
	key3, _ := crypto.HexToECDSA("49a7b37aa6f6645917e7b807e9d1c00d4fa71f18343b0d4122a4d2df64dd6fee")
	key := []*ecdsa.PrivateKey{key1, key2, key3}
	chain.blockCache = make(map[uint64]*types.Block)
	if num == 298 {
		for i := 0; i < 100000; i++ {

			tx := types.NewTransactions(uint64(i), common.Address{}, big.NewInt(100), 100, big.NewInt(int64(100)), nil, nil, 0, common.ExtraNormalTxType, 0)
			addr := common.Address{}
			addr.SetString(strconv.Itoa(i))
			tx.SetFromLoad(addr)
			tx.SetTxV(big.NewInt(1))
			tx1, _ := types.SignTx(tx, types.NewEIP155Signer(big.NewInt(1)), key[i%3])
			txs = append(txs, tx1)

		}
		chain.blockCache[num] = types.NewBlockWithTxs(header, txs)
	}

}
func (chain *Chain) GetBlockByNumber(num uint64) *types.Block {
	header := &types.Header{}
	txs := make([]types.SelfTransaction, 0)
	//key1, _ := crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	//key2, _ := crypto.HexToECDSA("8a1f9a8f95be41cd7ccb6168179afb4504aefe388d1e14474d32c45c72ce7b7a")
	//key3, _ := crypto.HexToECDSA("49a7b37aa6f6645917e7b807e9d1c00d4fa71f18343b0d4122a4d2df64dd6fee")
	//key := []*ecdsa.PrivateKey{key1, key2, key3}
	if num == 298 {
		//for i := 0; i < 100000; i++ {
		//
		//	tx := types.NewTransactions(uint64(i), common.Address{}, big.NewInt(100), 100, big.NewInt(int64(100)), nil, nil, 0, common.ExtraNormalTxType, 0)
		//	addr := common.Address{}
		//	addr.SetString(strconv.Itoa(i))
		//	tx.SetFromLoad(addr)
		//	tx.SetTxV(big.NewInt(1))
		//	tx1, _ := types.SignTx(tx, types.NewEIP155Signer(big.NewInt(1)), key[i%3])
		//	txs = append(txs, tx1)
		//
		//}
		return chain.blockCache[298]
	}

	return types.NewBlockWithTxs(header, txs)
}
func (chain *Chain) Config() *params.ChainConfig {
	return &params.ChainConfig{ChainId: big.NewInt(1), EIP155Block: big.NewInt(2), HomesteadBlock: new(big.Int)}
}

func TestTxsLottery_LotteryCalc(t *testing.T) {
	log.InitLog(5)
	monkey.Patch(manparams.IsBroadcastNumber, func(number uint64, stateNumber uint64) bool {
		fmt.Println("use monkey  manparams.IsBroadcastNumber")

		return false
	})

	monkey.Patch(matrixstate.GetDataByState, func(key string, state matrixstate.StateDB) (interface{}, error) {
		fmt.Println("use monkey  manparams.IsBroadcastNumber")
		if key == mc.MSKeyBroadcastInterval {
			return &mc.BCIntervalInfo{LastBCNumber: 0, LastReelectNumber: 0, BCInterval: 100}, nil
		}
		if key == mc.MSKeyLotteryCfg {
			info := make([]mc.LotteryInfo, 0)
			info = append(info, mc.LotteryInfo{PrizeLevel: 0, PrizeNum: 2, PrizeMoney: 6})
			return &mc.LotteryCfg{LotteryCalc: "1", LotteryInfo: info}, nil
		}
		return nil, nil
	})

	monkey.Patch(matrixstate.GetNumByState, func(key string, state matrixstate.StateDB) (uint64, error) {

		return uint64(3), nil
	})

	monkey.Patch(manparams.NewBCIntervalByNumber, func(blockNumber uint64) (*manparams.BCInterval, error) {
		fmt.Println("use monkey NewBCIntervalByNumber")

		inteval1 := &mc.BCIntervalInfo{LastBCNumber: 0, LastReelectNumber: 0, BCInterval: 100}

		interval2, _ := manparams.NewBCIntervalWithInterval(inteval1)
		return interval2, nil
	})
	lotterytest := New(&Chain{}, &State{0}, &randSeed{})
	lotterytest.LotteryCalc(common.Hash{}, 3)
}

func TestTxsLottery_LotteryCalc1(t *testing.T) {
	log.InitLog(5)
	monkey.Patch(manparams.IsBroadcastNumber, func(number uint64, stateNumber uint64) bool {
		fmt.Println("use monkey  manparams.IsBroadcastNumber")

		return false
	})

	monkey.Patch(matrixstate.GetDataByState, func(key string, state matrixstate.StateDB) (interface{}, error) {
		fmt.Println("use monkey  manparams.IsBroadcastNumber")
		if key == mc.MSKeyBroadcastInterval {
			return &mc.BCIntervalInfo{LastBCNumber: 0, LastReelectNumber: 0, BCInterval: 100}, nil
		}
		if key == mc.MSKeyLotteryCfg {
			info := make([]mc.LotteryInfo, 0)
			info = append(info, mc.LotteryInfo{PrizeLevel: 0, PrizeNum: 2, PrizeMoney: 6})
			return &mc.LotteryCfg{LotteryCalc: "1", LotteryInfo: info}, nil
		}
		if key == mc.MSKeyLotteryNum {
			info := make([]mc.LotteryInfo, 0)
			info = append(info, mc.LotteryInfo{PrizeLevel: 0, PrizeNum: 1, PrizeMoney: 6})
			return &mc.LotteryCfg{LotteryCalc: "1", LotteryInfo: info}, nil
		}
		return nil, nil
	})

	monkey.Patch(matrixstate.GetNumByState, func(key string, state matrixstate.StateDB) (uint64, error) {

		return uint64(3), nil
	})

	monkey.Patch(manparams.NewBCIntervalByNumber, func(blockNumber uint64) (*manparams.BCInterval, error) {
		fmt.Println("use monkey NewBCIntervalByNumber")

		inteval1 := &mc.BCIntervalInfo{LastBCNumber: 0, LastReelectNumber: 0, BCInterval: 100}

		interval2, _ := manparams.NewBCIntervalWithInterval(inteval1)
		return interval2, nil
	})
	lotterytest := New(&Chain{}, &State{0}, &randSeed{})
	lotterytest.LotteryCalc(common.Hash{}, 299)
}

func TestTxsLottery_LotteryCalc2(t *testing.T) {
	log.InitLog(5)
	monkey.Patch(manparams.IsBroadcastNumber, func(number uint64, stateNumber uint64) bool {
		fmt.Println("use monkey  manparams.IsBroadcastNumber")

		return false
	})

	monkey.Patch(matrixstate.GetDataByState, func(key string, state matrixstate.StateDB) (interface{}, error) {
		fmt.Println("use monkey  manparams.IsBroadcastNumber")
		if key == mc.MSKeyBroadcastInterval {
			return &mc.BCIntervalInfo{LastBCNumber: 0, LastReelectNumber: 0, BCInterval: 100}, nil
		}
		if key == mc.MSKeyLotteryCfg {
			info := make([]mc.LotteryInfo, 0)
			info = append(info, mc.LotteryInfo{PrizeLevel: 0, PrizeNum: 1, PrizeMoney: 6})
			return &mc.LotteryCfg{LotteryCalc: "1", LotteryInfo: info}, nil
		}
		if key == mc.MSKeyLotteryNum {
			info := make([]mc.LotteryInfo, 0)
			info = append(info, mc.LotteryInfo{PrizeLevel: 0, PrizeNum: 1, PrizeMoney: 6})
			return &mc.LotteryCfg{LotteryCalc: "1", LotteryInfo: info}, nil
		}
		return nil, nil
	})

	monkey.Patch(matrixstate.GetNumByState, func(key string, state matrixstate.StateDB) (uint64, error) {

		return uint64(3), nil
	})

	monkey.Patch(manparams.NewBCIntervalByNumber, func(blockNumber uint64) (*manparams.BCInterval, error) {
		fmt.Println("use monkey NewBCIntervalByNumber")

		inteval1 := &mc.BCIntervalInfo{LastBCNumber: 0, LastReelectNumber: 0, BCInterval: 100}

		interval2, _ := manparams.NewBCIntervalWithInterval(inteval1)
		return interval2, nil
	})
	lotterytest := New(&Chain{}, &State{0}, &randSeed{})
	lotterytest.LotteryCalc(common.Hash{}, 300)
}

func TestTxsLottery_LotteryCalc3(t *testing.T) {
	log.InitLog(5)
	monkey.Patch(manparams.IsBroadcastNumber, func(number uint64, stateNumber uint64) bool {
		fmt.Println("use monkey  manparams.IsBroadcastNumber")

		return false
	})

	monkey.Patch(matrixstate.GetDataByState, func(key string, state matrixstate.StateDB) (interface{}, error) {
		fmt.Println("use monkey  manparams.IsBroadcastNumber")
		if key == mc.MSKeyBroadcastInterval {
			return &mc.BCIntervalInfo{LastBCNumber: 300, LastReelectNumber: 300, BCInterval: 100}, nil
		}
		if key == mc.MSKeyLotteryCfg {
			info := make([]mc.LotteryInfo, 0)
			info = append(info, mc.LotteryInfo{PrizeLevel: 0, PrizeNum: 1, PrizeMoney: 1})
			info = append(info, mc.LotteryInfo{PrizeLevel: 1, PrizeNum: 2, PrizeMoney: 1})
			info = append(info, mc.LotteryInfo{PrizeLevel: 2, PrizeNum: 3, PrizeMoney: 1})
			return &mc.LotteryCfg{LotteryCalc: "1", LotteryInfo: info}, nil
		}
		if key == mc.MSKeyLotteryNum {
			info := make([]mc.LotteryInfo, 0)
			info = append(info, mc.LotteryInfo{PrizeLevel: 0, PrizeNum: 50, PrizeMoney: 1})
			return &mc.LotteryCfg{LotteryCalc: "1", LotteryInfo: info}, nil
		}
		return nil, nil
	})

	monkey.Patch(matrixstate.GetNumByState, func(key string, state matrixstate.StateDB) (uint64, error) {

		return uint64(3), nil
	})

	monkey.Patch(manparams.NewBCIntervalByNumber, func(blockNumber uint64) (*manparams.BCInterval, error) {
		fmt.Println("use monkey NewBCIntervalByNumber")

		inteval1 := &mc.BCIntervalInfo{LastBCNumber: 300, LastReelectNumber: 300, BCInterval: 100}

		interval2, _ := manparams.NewBCIntervalWithInterval(inteval1)
		return interval2, nil
	})
	chain := &Chain{}
	chain.New(298)
	lotterytest := New(chain, &State{6e18}, &randSeed{})
	test := lotterytest.LotteryCalc(common.Hash{}, 301)
	log.Info(PackageName, "奖励", test)
}

//
//func TestTxsLottery_LotteryChoose(t *testing.T) {
//	log.InitLog(3)
//	lotterytest := New(&Chain{}, &State{0}, &randSeed{})
//	TxCmpResultList
//}

//func TestTxsLottery_LotteryCalc2(t *testing.T) {
//	log.InitLog(3)
//	lotterytest := New(&Chain{}, &randSeed{})
//	lotterytest.LotteryCalc(&State{-1}, 299)
//}
//
//func TestTxsLottery_LotteryCalc3(t *testing.T) {
//	log.InitLog(3)
//	lotterytest := New(&Chain{}, &randSeed{})
//	lotterytest.LotteryCalc(&State{3e18}, 299)
//}
//
//func TestTxsLottery_LotteryCalc4(t *testing.T) {
//	log.InitLog(3)
//	lotterytest := New(&Chain{}, &randSeed{})
//	lotterytest.LotteryCalc(&State{6e18}, 299)
//}
