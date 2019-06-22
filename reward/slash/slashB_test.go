// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package slash

import (
	"math/big"
	"testing"

	"github.com/MatrixAINetwork/go-matrix/params/manparams"

	"github.com/MatrixAINetwork/go-matrix/log"

	"bou.ke/monkey"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/matrixstate"
	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/core/vm"
	"github.com/MatrixAINetwork/go-matrix/depoistInfo"
	"github.com/MatrixAINetwork/go-matrix/mandb"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/reward/util"
)

func TestSlashDelta_GetCurrentInterest00(t *testing.T) {
	log.InitLog(5)
	bp := &SlashDelta{bcInterval: &mc.BCIntervalInfo{BCInterval: 100}}
	chaindb := mandb.NewMemDatabase()
	preState, _ := state.NewStateDBManage(nil, chaindb, state.NewDatabase(chaindb))
	matrixstate.SetVersionInfo(preState, manparams.VersionDelta)
	currentState, _ := state.NewStateDBManage(nil, chaindb, state.NewDatabase(chaindb))
	matrixstate.SetVersionInfo(currentState, manparams.VersionDelta)
	matrixstate.SetInterestPayNum(currentState, 3600)
	monkey.Patch(depoistInfo.GetAllInterest_v2, func(stateDB vm.StateDBManager) map[common.Address]common.CalculateDeposit {
		ret := make(map[common.Address]common.CalculateDeposit)
		if stateDB == preState {
			operate0 := make([]common.OperationalInterestSlash, 0)
			operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Position: 0})
			operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Position: 1})
			operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Position: 2})
			operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Position: 3})
			operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Position: 4})
			ret[common.HexToAddress("0x01")] = common.CalculateDeposit{AddressA0: common.HexToAddress("0x01"), CalcDeposit: operate0}
			return ret
		}

		if stateDB == currentState {
			operate1 := make([]common.OperationalInterestSlash, 0)
			operate1 = append(operate1, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(20000), util.ManPrice), Position: 0})
			operate1 = append(operate1, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(15000), util.ManPrice), Position: 1})
			operate1 = append(operate1, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(13000), util.ManPrice), Position: 2})
			operate1 = append(operate1, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(14000), util.ManPrice), Position: 3})
			operate1 = append(operate1, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(11000), util.ManPrice), Position: 4})
			ret[common.HexToAddress("0x01")] = common.CalculateDeposit{AddressA0: common.HexToAddress("0x01"), CalcDeposit: operate1}
			return ret
		}
		return nil
	})

	bp.getCurrentInterest(preState, currentState, 3700)

}

func TestSlashDelta_GetCurrentInterest01(t *testing.T) {
	log.InitLog(5)
	bp := &SlashDelta{bcInterval: &mc.BCIntervalInfo{BCInterval: 100}}
	chaindb := mandb.NewMemDatabase()
	preState, _ := state.NewStateDBManage(nil, chaindb, state.NewDatabase(chaindb))
	matrixstate.SetVersionInfo(preState, manparams.VersionDelta)
	currentState, _ := state.NewStateDBManage(nil, chaindb, state.NewDatabase(chaindb))
	matrixstate.SetVersionInfo(currentState, manparams.VersionDelta)
	matrixstate.SetInterestPayNum(currentState, 3600)
	monkey.Patch(depoistInfo.GetAllInterest_v2, func(stateDB vm.StateDBManager) map[common.Address]common.CalculateDeposit {
		ret := make(map[common.Address]common.CalculateDeposit)
		if stateDB == preState {
			operate0 := make([]common.OperationalInterestSlash, 0)
			operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Position: 0})
			operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Position: 1})
			operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Position: 2})
			operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Position: 3})
			operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Position: 4})
			ret[common.HexToAddress("0x01")] = common.CalculateDeposit{AddressA0: common.HexToAddress("0x01"), CalcDeposit: operate0}
			return ret
		}

		if stateDB == currentState {
			operate1 := make([]common.OperationalInterestSlash, 0)
			operate1 = append(operate1, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(20000), util.ManPrice), Position: 0})
			operate1 = append(operate1, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(15000), util.ManPrice), Position: 1})
			operate1 = append(operate1, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(13000), util.ManPrice), Position: 2})
			operate1 = append(operate1, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(14000), util.ManPrice), Position: 3})
			operate1 = append(operate1, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(11000), util.ManPrice), Position: 4})
			ret[common.HexToAddress("0x01")] = common.CalculateDeposit{AddressA0: common.HexToAddress("0x01"), CalcDeposit: operate1}
			return ret
		}
		return nil
	})

	bp.getCurrentInterest(preState, currentState, 3800)

}

func TestSlashDelta_GetCurrentInterest02(t *testing.T) {

	log.InitLog(5)
	bp := &SlashDelta{bcInterval: &mc.BCIntervalInfo{BCInterval: 100}}
	chaindb := mandb.NewMemDatabase()
	preState, _ := state.NewStateDBManage(nil, chaindb, state.NewDatabase(chaindb))
	matrixstate.SetVersionInfo(preState, manparams.VersionDelta)
	currentState, _ := state.NewStateDBManage(nil, chaindb, state.NewDatabase(chaindb))
	matrixstate.SetVersionInfo(currentState, manparams.VersionDelta)
	matrixstate.SetInterestPayNum(currentState, 3600)
	monkey.Patch(depoistInfo.GetAllInterest_v2, func(stateDB vm.StateDBManager) map[common.Address]common.CalculateDeposit {
		ret := make(map[common.Address]common.CalculateDeposit)
		if stateDB == preState {
			operate0 := make([]common.OperationalInterestSlash, 0)
			operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Position: 0})
			operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Position: 1})
			operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Position: 2})
			operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Position: 3})
			operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Position: 4})
			ret[common.HexToAddress("0x01")] = common.CalculateDeposit{AddressA0: common.HexToAddress("0x01"), CalcDeposit: operate0}
			return ret
		}

		if stateDB == currentState {
			operate1 := make([]common.OperationalInterestSlash, 0)
			operate1 = append(operate1, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(20000), util.ManPrice), Position: 0})
			operate1 = append(operate1, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(15000), util.ManPrice), Position: 1})
			operate1 = append(operate1, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(13000), util.ManPrice), Position: 2})
			operate1 = append(operate1, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(14000), util.ManPrice), Position: 3})
			operate1 = append(operate1, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(11000), util.ManPrice), Position: 4})
			ret[common.HexToAddress("0x02")] = common.CalculateDeposit{AddressA0: common.HexToAddress("0x02"), CalcDeposit: operate1}
			return ret
		}
		return nil
	})

	bp.getCurrentInterest(preState, currentState, 3800)

}

func TestSlashDelta_GetCurrentInterest03(t *testing.T) {

	log.InitLog(5)
	bp := &SlashDelta{bcInterval: &mc.BCIntervalInfo{BCInterval: 100}}
	chaindb := mandb.NewMemDatabase()
	preState, _ := state.NewStateDBManage(nil, chaindb, state.NewDatabase(chaindb))
	matrixstate.SetVersionInfo(preState, manparams.VersionDelta)
	currentState, _ := state.NewStateDBManage(nil, chaindb, state.NewDatabase(chaindb))
	matrixstate.SetVersionInfo(currentState, manparams.VersionDelta)
	matrixstate.SetInterestPayNum(currentState, 3600)
	monkey.Patch(depoistInfo.GetAllInterest_v2, func(stateDB vm.StateDBManager) map[common.Address]common.CalculateDeposit {
		ret := make(map[common.Address]common.CalculateDeposit)
		if stateDB == preState {
			operate0 := make([]common.OperationalInterestSlash, 0)
			operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Position: 0})
			operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Position: 1})
			operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Position: 2})
			operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Position: 3})
			operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Position: 4})
			ret[common.HexToAddress("0x01")] = common.CalculateDeposit{AddressA0: common.HexToAddress("0x01"), CalcDeposit: operate0}
			return ret
		}

		if stateDB == currentState {
			operate1 := make([]common.OperationalInterestSlash, 0)
			operate1 = append(operate1, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(20000), util.ManPrice), Position: 0})
			operate1 = append(operate1, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(15000), util.ManPrice), Position: 1})
			operate1 = append(operate1, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(14000), util.ManPrice), Position: 3})
			operate1 = append(operate1, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(11000), util.ManPrice), Position: 4})
			operate1 = append(operate1, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(18000), util.ManPrice), Position: 5})
			ret[common.HexToAddress("0x01")] = common.CalculateDeposit{AddressA0: common.HexToAddress("0x01"), CalcDeposit: operate1}
			return ret
		}
		return nil
	})

	bp.getCurrentInterest(preState, currentState, 3800)
}

func TestSlashDelta_addSlash00(t *testing.T) {
	log.InitLog(5)
	bp := &SlashDelta{}
	chaindb := mandb.NewMemDatabase()
	currentState, _ := state.NewStateDBManage(nil, chaindb, state.NewDatabase(chaindb))
	monkey.Patch(depoistInfo.GetSlash_v2, func(stateDB vm.StateDBManager, address common.Address) (common.CalculateDeposit, error) {
		if stateDB == currentState {
			operate0 := make([]common.OperationalInterestSlash, 0)
			operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Position: 0})
			operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Position: 1})
			operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Position: 2})
			operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Position: 3})
			operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Position: 4})

			return common.CalculateDeposit{AddressA0: common.HexToAddress("0x01"), CalcDeposit: operate0}, nil
		}
		return common.CalculateDeposit{}, nil
	})
	operate0 := make([]common.OperationalInterestSlash, 0)
	operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Position: 0})
	operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(20000), util.ManPrice), Position: 1})
	operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(30000), util.ManPrice), Position: 2})
	operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(40000), util.ManPrice), Position: 3})
	operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(50000), util.ManPrice), Position: 4})
	bp.addSlash(common.HexToAddress("0x01"), operate0, currentState, uint64(7500))

}

func TestSlashDelta_addSlash01(t *testing.T) {
	log.InitLog(5)
	bp := &SlashDelta{}
	chaindb := mandb.NewMemDatabase()
	currentState, _ := state.NewStateDBManage(nil, chaindb, state.NewDatabase(chaindb))
	monkey.Patch(depoistInfo.GetSlash_v2, func(stateDB vm.StateDBManager, address common.Address) (common.CalculateDeposit, error) {
		if stateDB == currentState {
			operate0 := make([]common.OperationalInterestSlash, 0)
			operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Position: 0})
			operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Position: 1})
			operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Position: 3})
			operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Position: 4})

			return common.CalculateDeposit{AddressA0: common.HexToAddress("0x01"), CalcDeposit: operate0}, nil
		}
		return common.CalculateDeposit{}, nil
	})
	operate0 := make([]common.OperationalInterestSlash, 0)
	operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Position: 0})
	operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(20000), util.ManPrice), Position: 1})
	operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(30000), util.ManPrice), Position: 2})
	operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(40000), util.ManPrice), Position: 3})
	operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(50000), util.ManPrice), Position: 4})
	bp.addSlash(common.HexToAddress("0x01"), operate0, currentState, uint64(7500))

}

func TestSlashDelta_addSlash02(t *testing.T) {
	log.InitLog(5)
	bp := &SlashDelta{}
	chaindb := mandb.NewMemDatabase()
	currentState, _ := state.NewStateDBManage(nil, chaindb, state.NewDatabase(chaindb))
	monkey.Patch(depoistInfo.GetSlash_v2, func(stateDB vm.StateDBManager, address common.Address) (common.CalculateDeposit, error) {
		if stateDB == currentState {
			operate0 := make([]common.OperationalInterestSlash, 0)
			operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Position: 0})
			operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Position: 1})
			operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Position: 2})
			operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Position: 3})
			operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Position: 4})

			return common.CalculateDeposit{AddressA0: common.HexToAddress("0x01"), CalcDeposit: operate0}, nil
		}
		return common.CalculateDeposit{}, nil
	})
	operate0 := make([]common.OperationalInterestSlash, 0)
	operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Position: 0})
	operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(20000), util.ManPrice), Position: 1})
	operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(40000), util.ManPrice), Position: 3})
	operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(50000), util.ManPrice), Position: 4})
	bp.addSlash(common.HexToAddress("0x01"), operate0, currentState, uint64(7500))

}
