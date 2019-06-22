// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package blkgenor

import (
	"fmt"
	"github.com/MatrixAINetwork/go-matrix"
	"github.com/MatrixAINetwork/go-matrix/common/readstatedb"
	"github.com/MatrixAINetwork/go-matrix/core/matrixstate"
	"github.com/MatrixAINetwork/go-matrix/params/manparams"
	"math/big"
	"testing"

	"bou.ke/monkey"
	"github.com/MatrixAINetwork/go-matrix/ca"

	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/core/vm"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	. "github.com/smartystreets/goconvey/convey"
)

func TestProcess_LeaderInsertAndBcBlock(t *testing.T) {
	log.InitLog(3)
	eth := fakeEthNew(0)
	if nil == eth {
		fmt.Println("failed to create eth")
	}
	Convey("可否插入区块测试", t, func() {
		Convey("当前高度0，leader可否生成高度1插入区块函数测试", func() {
			monkey.Patch(manparams.NewBCIntervalByHash, func(blockHash common.Hash) (*manparams.BCInterval, error) {
				fmt.Println("use monkey NewBCIntervalByNumber")

				inteval1 := &mc.BCIntervalInfo{LastBCNumber: 0, LastReelectNumber: 0, BCInterval: 100}

				interval2, _ := manparams.NewBCIntervalWithInterval(inteval1)
				return interval2, nil
			})

			monkey.Patch(ca.GetAddress, func() common.Address {

				return common.HexToAddress(testAddress1)
			})
			//blockgen.
			monkey.Patch(matrixstate.GetDataByState, func(key string, state matrixstate.StateDB) (interface{}, error) {
				fmt.Println("use monkey  manparams.IsBroadcastNumber")
				if key == mc.MSKeyPreMiner {
					return &mc.PreMinerStruct{PreMiner: common.HexToAddress(testAddress)}, nil
				}
				if key == mc.MSKeyMatrixAccount {
					return &mc.MatrixSpecialAccounts{FoundationAccount: mc.NodeInfo{Address: common.HexToAddress(testAddress)}}, nil
				}
				if key == mc.MSKeyBroadcastInterval {
					return &mc.BCIntervalInfo{LastBCNumber: 0, LastReelectNumber: 0, BCInterval: 100}, nil
				}
				return nil, nil
			})

			blockgen, err := New(eth)
			if err != nil {
			}
			header := types.CopyHeader(eth.BlockChain().CurrentHeader())
			monkey.Patch(readstatedb.GetPreBroadcastRoot, func(stateReader matrix.StateReader, height uint64) (*mc.PreBroadStateRoot, error) {
				fmt.Println("use monkey  GetPreBroadcastRoot")

				return &mc.PreBroadStateRoot{header.Root, header.Root}, nil
			})
			header.ParentHash = header.Hash()
			header.Number = big.NewInt(1)
			header.Leader = common.HexToAddress(testAddress)
			state, _ := eth.BlockChain().State()
			roleMsg := &mc.RoleUpdatedMsg{Role: common.RoleValidator, BlockNum: uint64(0), BlockHash: common.Hash{1}, Leader: common.HexToAddress(testAddress), IsSuperBlock: false}
			blockgen.roleUpdatedMsgHandle(roleMsg)
			p := blockgen.pm.GetCurrentProcess()
			p.blockCache.SaveReadyBlock(&mc.BlockLocalVerifyOK{Header: header, State: state})
			var guard *monkey.PatchGuard
			guard = monkey.Patch(ca.GetElectedByHeightWithdraw, func(height *big.Int) ([]vm.DepositDetail, error) {
				fmt.Println("use my GetElectedByHeightWithdraw")
				guard.Unpatch()
				defer guard.Restore()
				return nil, nil
			})

			hash, err := p.InsertAndBcBlock(true, common.HexToAddress(testAddress), header)
			So(err, ShouldEqual, nil)
			So(hash, ShouldNotEqual, common.Hash{})
			So(header.Hash(), ShouldEqual, eth.BlockChain().CurrentHeader().Hash())
			So(eth.BlockChain().CurrentHeader().Number.Int64(), ShouldEqual, int64(1))

		})
	})
}

func TestProcess_FowllerInsertAndBcBlock(t *testing.T) {
	log.InitLog(3)
	eth := fakeEthNew(0)
	if nil == eth {
		fmt.Println("failed to create eth")
	}
	Convey("可否插入区块测试", t, func() {
		Convey("当前高度0，leader可否生成高度1插入区块函数测试", func() {
			monkey.Patch(manparams.NewBCIntervalByHash, func(blockHash common.Hash) (*manparams.BCInterval, error) {
				fmt.Println("use monkey NewBCIntervalByNumber")

				inteval1 := &mc.BCIntervalInfo{LastBCNumber: 0, LastReelectNumber: 0, BCInterval: 100}

				interval2, _ := manparams.NewBCIntervalWithInterval(inteval1)
				return interval2, nil
			})

			monkey.Patch(ca.GetAddress, func() common.Address {

				return common.HexToAddress(testAddress1)
			})
			//blockgen.
			monkey.Patch(matrixstate.GetDataByState, func(key string, state matrixstate.StateDB) (interface{}, error) {
				fmt.Println("use monkey  manparams.IsBroadcastNumber")
				if key == mc.MSKeyPreMiner {
					return &mc.PreMinerStruct{PreMiner: common.HexToAddress(testAddress)}, nil
				}
				if key == mc.MSKeyMatrixAccount {
					return &mc.MatrixSpecialAccounts{FoundationAccount: mc.NodeInfo{Address: common.HexToAddress(testAddress)}}, nil
				}
				if key == mc.MSKeyBroadcastInterval {
					return &mc.BCIntervalInfo{LastBCNumber: 0, LastReelectNumber: 0, BCInterval: 100}, nil
				}
				return nil, nil
			})

			blockgen, err := New(eth)
			if err != nil {
			}
			header := types.CopyHeader(eth.BlockChain().CurrentHeader())
			monkey.Patch(readstatedb.GetPreBroadcastRoot, func(stateReader matrix.StateReader, height uint64) (*mc.PreBroadStateRoot, error) {
				fmt.Println("use monkey  GetPreBroadcastRoot")

				return &mc.PreBroadStateRoot{header.Root, header.Root}, nil
			})
			header.ParentHash = header.Hash()
			header.Number = big.NewInt(1)
			header.Leader = common.HexToAddress(testAddress)
			state, _ := eth.BlockChain().State()
			roleMsg := &mc.RoleUpdatedMsg{Role: common.RoleValidator, BlockNum: uint64(0), BlockHash: common.Hash{1}, Leader: common.HexToAddress(testAddress), IsSuperBlock: false}
			blockgen.roleUpdatedMsgHandle(roleMsg)
			p := blockgen.pm.GetCurrentProcess()
			p.blockCache.SaveReadyBlock(&mc.BlockLocalVerifyOK{Header: header, State: state})
			var guard *monkey.PatchGuard
			guard = monkey.Patch(ca.GetElectedByHeightWithdraw, func(height *big.Int) ([]vm.DepositDetail, error) {
				fmt.Println("use my GetElectedByHeightWithdraw")
				guard.Unpatch()
				defer guard.Restore()
				return nil, nil
			})

			hash, err := p.InsertAndBcBlock(false, common.HexToAddress(testAddress), header)
			So(err, ShouldEqual, nil)
			So(hash, ShouldNotEqual, common.Hash{})
			So(header.Hash(), ShouldEqual, eth.BlockChain().CurrentHeader().Hash())
			So(eth.BlockChain().CurrentHeader().Number.Int64(), ShouldEqual, int64(1))

		})
	})
}
