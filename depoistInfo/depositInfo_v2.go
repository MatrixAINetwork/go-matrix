// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package depoistInfo

import (
	"math/big"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/vm"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/params"
)

func (v2 *DepositMnger_v2) AddOnlineTime(stateDB vm.StateDBManager, address common.Address, ot *big.Int) error {
	return v2.MatrixDeposit.SetOnlineTime(v2.Contract, stateDB, address, ot)
}

func (v2 *DepositMnger_v2) GetOnlineTime(stateDB vm.StateDBManager, address common.Address) (*big.Int, error) {
	return v2.MatrixDeposit.GetOnlineTime(v2.Contract, stateDB, address), nil
}

func (v2 *DepositMnger_v2) SetOnlineTime(stateDB vm.StateDBManager, address common.Address, ot *big.Int) error {
	return v2.MatrixDeposit.SetOnlineTime(v2.Contract, stateDB, address, ot)
}

func (v2 *DepositMnger_v2) GetDepositList(tm *big.Int, getDeposit common.RoleType, statedb vm.StateDBManager, headtime uint64) ([]vm.DepositDetail, error) {
	//db, err := getDepositInfo(tm)
	//if err != nil {
	//	return nil, err
	//}

	contract := vm.NewContract(vm.AccountRef(common.HexToAddress("1337")), vm.AccountRef(common.BytesToAddress([]byte{10})), big.NewInt(0), 60000, params.MAN_COIN)
	//depositList := make([]vm.DepositDetail, 0)
	retDepositRoles := v2.MatrixDeposit.GetAllDepositList(contract, statedb, false, headtime)

	depositList := v2.convertionDetail(retDepositRoles, getDeposit, false)
	return depositList, nil
}
func (v2 *DepositMnger_v2) convertionDetail(retDepositRoles []common.DepositBase, getDeposit common.RoleType, checkDepositVal bool) []vm.DepositDetail {
	depositList := make([]vm.DepositDetail, 0)
	for _, droles := range retDepositRoles {
		deposit := vm.DepositDetail{
			Address:     common.Address{},
			SignAddress: common.Address{},
			Deposit:     new(big.Int),
			WithdrawH:   new(big.Int),
			OnlineTime:  new(big.Int),
			Role:        new(big.Int),
		}
		deposit.Address = droles.AddressA0
		deposit.SignAddress = droles.AddressA1
		deposit.OnlineTime = droles.OnlineTime
		for _, dam := range droles.Dpstmsg {
			deposit.Deposit = new(big.Int).Add(deposit.Deposit, dam.DepositAmount)
			if checkDepositVal {
				checkVal, _ := new(big.Int).SetString("10000000000000000000000000", 10)
				if deposit.Deposit.Cmp(checkVal) >= 0 {
					deposit.Deposit = checkVal
					break
				}
			}
		}
		deposit.Role = droles.Role
		if getDeposit > 0 {
			if droles.Role.Cmp(new(big.Int).SetUint64(uint64(common.RoleValidator&getDeposit))) == 0 {
				depositList = append(depositList, deposit)
			}
			if droles.Role.Cmp(new(big.Int).SetUint64(uint64(common.RoleMiner&getDeposit))) == 0 {
				depositList = append(depositList, deposit)
			}
		} else {
			depositList = append(depositList, deposit)
		}

	}
	return depositList
}
func (v2 *DepositMnger_v2) GetDepositAndWithDrawList(tm *big.Int, statedb vm.StateDBManager, headtime uint64) ([]vm.DepositDetail, error) {
	//db, err := getDepositInfo(tm)
	//if err != nil {
	//	return nil, err
	//}
	contract := vm.NewContract(vm.AccountRef(common.HexToAddress("1337")), vm.AccountRef(common.BytesToAddress([]byte{10})), big.NewInt(0), 60000, params.MAN_COIN)
	depositList := v2.MatrixDeposit.GetAllDepositList(contract, statedb, true, headtime)
	depositdetails := v2.convertionDetail(depositList, 0, false)
	return depositdetails, nil
}

func (v2 *DepositMnger_v2) GetAllDeposit(tm *big.Int, statedb vm.StateDBManager, headtime uint64) ([]vm.DepositDetail, error) {
	//db, err := getDepositInfo(tm)
	//if err != nil {
	//	return nil, err
	//}
	contract := vm.NewContract(vm.AccountRef(common.HexToAddress("1337")), vm.AccountRef(common.BytesToAddress([]byte{10})), big.NewInt(0), 60000, params.MAN_COIN)
	depositList := v2.MatrixDeposit.GetAllDepositList(contract, statedb, false, headtime)
	depositdetails := v2.convertionDetail(depositList, 0, false)
	return depositdetails, nil
}

func (v2 *DepositMnger_v2) GetDepositListByHash(hash common.Hash, getDeposit common.RoleType, statedb vm.StateDBManager, headtime uint64) ([]vm.DepositDetail, error) {
	//db, err := getDepositInfoByHash(hash)
	//if err != nil {
	//	return nil, err
	//}
	contract := vm.NewContract(vm.AccountRef(common.HexToAddress("1337")), vm.AccountRef(common.BytesToAddress([]byte{10})), big.NewInt(0), 60000, params.MAN_COIN)
	retDepositRoles := v2.MatrixDeposit.GetAllDepositList(contract, statedb, false, headtime)
	depositList := v2.convertionDetail(retDepositRoles, getDeposit, true)
	return depositList, nil
}

func (v2 *DepositMnger_v2) GetDepositAndWithDrawListByHash(hash common.Hash, statedb vm.StateDBManager, headtime uint64) ([]vm.DepositDetail, error) {
	//db, err := getDepositInfoByHash(hash)
	//if err != nil {
	//	return nil, err
	//}
	contract := vm.NewContract(vm.AccountRef(common.HexToAddress("1337")), vm.AccountRef(common.BytesToAddress([]byte{10})), big.NewInt(0), 60000, params.MAN_COIN)
	depositList := v2.MatrixDeposit.GetAllDepositList(contract, statedb, true, headtime)
	depositdetails := v2.convertionDetail(depositList, 0, false)
	return depositdetails, nil
}

func (v2 *DepositMnger_v2) GetAllDepositByHash(hash common.Hash, statedb vm.StateDBManager, headtime uint64) ([]vm.DepositDetail, error) {
	//db, err := getDepositInfoByHash(hash)
	//if err != nil {
	//	return nil, err
	//}
	contract := vm.NewContract(vm.AccountRef(common.HexToAddress("1337")), vm.AccountRef(common.BytesToAddress([]byte{10})), big.NewInt(0), 60000, params.MAN_COIN)
	depositList := v2.MatrixDeposit.GetAllDepositList(contract, statedb, false, headtime)
	depositdetails := v2.convertionDetail(depositList, 0, false)
	return depositdetails, nil
}

func (v2 *DepositMnger_v2) ResetSlash(stateDB vm.StateDBManager, address common.Address) error {
	return v2.MatrixDeposit.ResetSlash(v2.Contract, stateDB, address)
}

func (v2 *DepositMnger_v2) GetSlash(stateDB vm.StateDBManager, address common.Address) (*big.Int, error) {
	return nil, nil
}

func (v2 *DepositMnger_v2) AddSlash(stateDB vm.StateDBManager, address common.Address, slash *big.Int) error {
	return nil
}

func (v2 *DepositMnger_v2) GetInterest(stateDB vm.StateDBManager, address common.Address) (*big.Int, error) {
	return nil, nil
}

func (v2 *DepositMnger_v2) GetAllInterest(stateDB vm.StateDBManager) map[common.Address]*big.Int {
	return nil
}

func (v2 *DepositMnger_v2) AddInterest(stateDB vm.StateDBManager, address common.Address, reward *big.Int) error {
	return nil
}

func (v2 *DepositMnger_v2) SetDeposit(stateDB vm.StateDBManager, address common.Address) error {
	return nil //v2.MatrixDeposit.SetDeposit(v2.Contract, stateDB, address)
}

// 获取A0账户
func (v2 *DepositMnger_v2) GetDepositAccount(stateDB vm.StateDBManager, authAccount common.Address) common.Address {
	if v2.Contract == nil {
		v2.Contract = vm.NewContract(vm.AccountRef(common.HexToAddress("1337")), vm.AccountRef(common.BytesToAddress([]byte{10})), big.NewInt(0), 0, params.MAN_COIN)
	}
	return v2.MatrixDeposit.GetDepositAccount(v2.Contract, stateDB, authAccount)
}

// 获取A1账户
func (v2 *DepositMnger_v2) GetAuthAccount(stateDB vm.StateDBManager, depositAccount common.Address) common.Address {
	if v2.Contract == nil {
		v2.Contract = vm.NewContract(vm.AccountRef(common.HexToAddress("1337")), vm.AccountRef(common.BytesToAddress([]byte{10})), big.NewInt(0), 0, params.MAN_COIN)
	}
	return v2.MatrixDeposit.GetAuthAccount(v2.Contract, stateDB, depositAccount)
}
func (v2 *DepositMnger_v2) AddSlash_v2(stateDB vm.StateDBManager, address common.Address, slash common.CalculateDeposit) error {
	return v2.MatrixDeposit.AddSlash(v2.Contract, stateDB, address, slash)
}
func (v2 *DepositMnger_v2) GetSlash_v2(stateDB vm.StateDBManager, address common.Address) (common.CalculateDeposit, error) {
	return v2.MatrixDeposit.GetSlash(v2.Contract, stateDB, address), nil
}
func (v2 *DepositMnger_v2) GetInterest_v2(stateDB vm.StateDBManager, address common.Address) (common.CalculateDeposit, error) {
	return v2.MatrixDeposit.GetInterest(v2.Contract, stateDB, address), nil
}
func (v2 *DepositMnger_v2) GetAllInterest_v2(stateDB vm.StateDBManager, headTime uint64) map[common.Address]common.CalculateDeposit {
	return v2.MatrixDeposit.GetAllInterest(v2.Contract, stateDB, headTime)
}
func (v2 *DepositMnger_v2) AddInterest_v2(stateDB vm.StateDBManager, address common.Address, reward common.CalculateDeposit) error {
	return v2.MatrixDeposit.AddInterest(v2.Contract, stateDB, address, reward)
}
func (v2 *DepositMnger_v2) GetAllDepositByHash_v2(hash common.Hash, stateDB vm.StateDBManager, headtime uint64) ([]common.DepositBase, error) {
	//db, err := getDepositInfoByHash(hash)
	//if err != nil {
	//	return nil, err
	//}
	//该方法仅用于利息和惩罚
	contract := vm.NewContract(vm.AccountRef(common.HexToAddress("1337")), vm.AccountRef(common.BytesToAddress([]byte{10})), big.NewInt(0), 60000, params.MAN_COIN)
	depositList := v2.MatrixDeposit.GetAllDepositListByInterest(contract, stateDB, false, headtime)
	return depositList, nil
}
func (v2 *DepositMnger_v2) PayInterest(stateDB vm.StateDBManager, addrA0 common.Address, position uint64, amount *big.Int) error {
	if v2.Contract == nil {
		v2.Contract = vm.NewContract(vm.AccountRef(common.HexToAddress("1337")), vm.AccountRef(common.BytesToAddress([]byte{10})), big.NewInt(0), 0, params.MAN_COIN)
	}
	return v2.MatrixDeposit.PayInterest(v2.Contract, stateDB, addrA0, position, amount)

}
func (v2 *DepositMnger_v2) GetDepositBase(stateDB vm.StateDBManager, addrA0 common.Address) *common.DepositBase {
	if v2.Contract == nil {
		v2.Contract = vm.NewContract(vm.AccountRef(common.HexToAddress("1337")), vm.AccountRef(common.BytesToAddress([]byte{10})), big.NewInt(0), 0, params.MAN_COIN)
	}
	return v2.MatrixDeposit.GetDepositBase(v2.Contract, stateDB, addrA0)
}
func (v2 *DepositMnger_v2) CheckDeposit(statedb vm.StateDBManager) []common.DepositBase {
	if v2.Contract == nil {
		v2.Contract = vm.NewContract(vm.AccountRef(common.HexToAddress("1337")), vm.AccountRef(common.BytesToAddress([]byte{10})), big.NewInt(0), 0, params.MAN_COIN)
	}
	addrlist, err := v2.MatrixDeposit.GetA0list(v2.Contract, statedb)
	if err != nil {
		log.ERROR("CheckDeposit err", "can not get A0 address list")
		return nil
	}
	var depositlist []common.DepositBase
	for _, addr := range addrlist {
		deposit := v2.MatrixDeposit.GetDepositBase(v2.Contract, statedb, addr)
		depositlist = append(depositlist, *deposit)
	}
	return depositlist
}
