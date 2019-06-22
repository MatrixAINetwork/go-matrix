// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package depoistInfo

import (
	"errors"
	"math/big"

	"strconv"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/vm"
	"github.com/MatrixAINetwork/go-matrix/params"
)

type DepositInterfaceer interface {
	AddOnlineTime(stateDB vm.StateDBManager, address common.Address, ot *big.Int) error
	GetOnlineTime(stateDB vm.StateDBManager, address common.Address) (*big.Int, error)
	SetOnlineTime(stateDB vm.StateDBManager, address common.Address, ot *big.Int) error
	GetDepositList(tm *big.Int, getDeposit common.RoleType, statedb vm.StateDBManager, headtime uint64) ([]vm.DepositDetail, error)
	GetDepositAndWithDrawList(tm *big.Int, statedb vm.StateDBManager, headtime uint64) ([]vm.DepositDetail, error)
	GetAllDeposit(tm *big.Int, statedb vm.StateDBManager, headtime uint64) ([]vm.DepositDetail, error)
	GetDepositListByHash(hash common.Hash, getDeposit common.RoleType, statedb vm.StateDBManager, headtime uint64) ([]vm.DepositDetail, error)
	GetDepositAndWithDrawListByHash(hash common.Hash, statedb vm.StateDBManager, headtime uint64) ([]vm.DepositDetail, error)
	GetAllDepositByHash(hash common.Hash, stateDB vm.StateDBManager, headtime uint64) ([]vm.DepositDetail, error)
	GetAllDepositByHash_v2(hash common.Hash, stateDB vm.StateDBManager, headtime uint64) ([]common.DepositBase, error)
	ResetSlash(stateDB vm.StateDBManager, address common.Address) error
	AddSlash(stateDB vm.StateDBManager, address common.Address, slash *big.Int) error
	GetSlash(stateDB vm.StateDBManager, address common.Address) (*big.Int, error)
	AddSlash_v2(stateDB vm.StateDBManager, address common.Address, slash common.CalculateDeposit) error
	GetSlash_v2(stateDB vm.StateDBManager, address common.Address) (common.CalculateDeposit, error)
	GetInterest_v2(stateDB vm.StateDBManager, address common.Address) (common.CalculateDeposit, error)
	GetAllInterest_v2(stateDB vm.StateDBManager, headTime uint64) map[common.Address]common.CalculateDeposit
	AddInterest_v2(stateDB vm.StateDBManager, address common.Address, reward common.CalculateDeposit) error
	GetInterest(stateDB vm.StateDBManager, address common.Address) (*big.Int, error)
	GetAllInterest(stateDB vm.StateDBManager) map[common.Address]*big.Int
	AddInterest(stateDB vm.StateDBManager, address common.Address, reward *big.Int) error
	SetDeposit(stateDB vm.StateDBManager, address common.Address) error
	GetDepositAccount(stateDB vm.StateDBManager, authAccount common.Address) common.Address
	GetAuthAccount(stateDB vm.StateDBManager, depositAccount common.Address) common.Address
	PayInterest(stateDB vm.StateDBManager, addrA0 common.Address, position uint64, amount *big.Int) error
}

func NewDepositInfo(manApi manBackend) {
	pre := vm.PrecompiledContractsByzantium[common.BytesToAddress([]byte{10})]
	depositInfo = &DepositInfo{manApi: manApi, p: pre}
	//getDepositListTest()
}

type DepositInfo struct {
	MatrixDeposit vm.MatrixDeposit001
	Contract      *vm.Contract
	manApi        manBackend
	p             vm.PrecompiledContract
}

var (
	depositInfo *DepositInfo
)

type DepositMnger_v1 struct {
}
type DepositMnger_v2 struct {
	MatrixDeposit vm.MatrixDeposit002
	Contract      *vm.Contract
}

func SetVersion(statedb vm.StateDBManager, t uint64) error {
	err := ConversionDeposit(statedb, t)
	if err != nil {
		return err
	}
	statedb.SetState(params.MAN_COIN, common.Address{}, common.BytesToHash([]byte(params.DepositVersionKey_1)), common.BytesToHash([]byte(params.DepositVersion_1)))
	return nil
}

func selectDeposit(statedb vm.StateDBManager) DepositInterfaceer {
	ret := statedb.GetState(params.MAN_COIN, common.Address{}, common.BytesToHash([]byte(params.DepositVersionKey_1)))
	if ret.Equal(common.BytesToHash([]byte(params.DepositVersion_1))) {
		return depositmanagerversoin2
	} else {
		return depositmanagerversoin1
	}
}

var depositmanagerversoin1 *DepositMnger_v1
var depositmanagerversoin2 *DepositMnger_v2

func init() {
	depositmanagerversoin1 = new(DepositMnger_v1)
	depositmanagerversoin2 = &DepositMnger_v2{MatrixDeposit: vm.MatrixDeposit002{}}
}

func AddOnlineTime(stateDB vm.StateDBManager, address common.Address, ot *big.Int) error {
	dmv := selectDeposit(stateDB)
	if dmv != nil {
		return dmv.AddOnlineTime(stateDB, address, ot)
	} else {
		return errors.New("unknow version AddOnlineTime")
	}
}
func GetOnlineTime(stateDB vm.StateDBManager, address common.Address) (*big.Int, error) {
	dmv := selectDeposit(stateDB)
	if dmv != nil {
		return dmv.GetOnlineTime(stateDB, address)
	} else {
		return big.NewInt(0), errors.New("unknow version GetOnlineTime")
	}
}
func SetOnlineTime(stateDB vm.StateDBManager, address common.Address, ot *big.Int) error {
	dmv := selectDeposit(stateDB)
	if dmv != nil {
		return dmv.SetOnlineTime(stateDB, address, ot)
	} else {
		return errors.New("unknow version SetOnlineTime")
	}
}
func GetDepositList(tm *big.Int, getDeposit common.RoleType) ([]vm.DepositDetail, error) {
	statedb, headtime, err := getDepositInfo(tm)
	if err != nil {
		return nil, err
	}
	dmv := selectDeposit(statedb)
	if dmv != nil {
		return dmv.GetDepositList(tm, getDeposit, statedb, headtime)
	} else {
		return nil, errors.New("unknow version GetDepositList")
	}
}
func GetDepositAndWithDrawList(tm *big.Int) ([]vm.DepositDetail, error) {
	statedb, headtime, err := getDepositInfo(tm)
	if err != nil {
		return nil, err
	}
	dmv := selectDeposit(statedb)
	if dmv != nil {
		return dmv.GetDepositAndWithDrawList(tm, statedb, headtime)
	} else {
		return nil, errors.New("unknow version GetDepositAndWithDrawList")
	}
}
func GetAllDeposit(tm *big.Int) ([]vm.DepositDetail, error) {
	statedb, headtime, err := getDepositInfo(tm)
	if err != nil {
		return nil, err
	}
	dmv := selectDeposit(statedb)
	if dmv != nil {
		return dmv.GetAllDeposit(tm, statedb, headtime)
	} else {
		return nil, errors.New("unknow version GetAllDeposit")
	}
}
func GetDepositListByHash(hash common.Hash, getDeposit common.RoleType) ([]vm.DepositDetail, error) {
	statedb, headtime, err := getDepositInfoByHash(hash)
	if err != nil {
		return nil, err
	}

	dmv := selectDeposit(statedb)
	if dmv != nil {
		return dmv.GetDepositListByHash(hash, getDeposit, statedb, headtime)
	} else {
		return nil, errors.New("unknow version GetDepositListByHash")
	}
}
func GetDepositAndWithDrawListByHash(hash common.Hash) ([]vm.DepositDetail, error) {
	statedb, headtime, err := getDepositInfoByHash(hash)
	if err != nil {
		return nil, err
	}
	dmv := selectDeposit(statedb)
	if dmv != nil {
		return dmv.GetDepositAndWithDrawListByHash(hash, statedb, headtime)
	} else {
		return nil, errors.New("unknow version GetDepositAndWithDrawListByHash")
	}
}
func GetAllDepositByHash(hash common.Hash) ([]vm.DepositDetail, error) {
	statedb, headtime, err := getDepositInfoByHash(hash)
	if err != nil {
		return nil, err
	}
	dmv := selectDeposit(statedb)
	if dmv != nil {
		return dmv.GetAllDepositByHash(hash, statedb, headtime)
	} else {
		return nil, errors.New("unknow version GetAllDepositByHash")
	}
}
func GetAllDepositByHash_v2(hash common.Hash) ([]common.DepositBase, error) {
	statedb, headtime, err := getDepositInfoByHash(hash)
	if err != nil {
		return nil, err
	}
	dmv := selectDeposit(statedb)
	if dmv != nil {
		return dmv.GetAllDepositByHash_v2(hash, statedb, headtime)
	} else {
		return nil, errors.New("unknow version ResetSlash")
	}
}
func ResetSlash(stateDB vm.StateDBManager, address common.Address) error {
	dmv := selectDeposit(stateDB)
	if dmv != nil {
		return dmv.ResetSlash(stateDB, address)
	} else {
		return errors.New("unknow version ResetSlash")
	}
}
func GetSlash(stateDB vm.StateDBManager, address common.Address) (*big.Int, error) {
	dmv := selectDeposit(stateDB)
	if dmv != nil {
		return dmv.GetSlash(stateDB, address)
	} else {
		return big.NewInt(0), errors.New("unknow version GetSlash")
	}
}
func AddSlash(stateDB vm.StateDBManager, address common.Address, slash *big.Int) error {
	dmv := selectDeposit(stateDB)
	if dmv != nil {
		return dmv.AddSlash(stateDB, address, slash)
	} else {
		return errors.New("unknow version AddSlash")
	}
}
func GetInterest(stateDB vm.StateDBManager, address common.Address) (*big.Int, error) {
	dmv := selectDeposit(stateDB)
	if dmv != nil {
		return dmv.GetInterest(stateDB, address)
	} else {
		return big.NewInt(0), errors.New("unknow version GetInterest")
	}
}
func GetAllInterest(stateDB vm.StateDBManager) map[common.Address]*big.Int {
	dmv := selectDeposit(stateDB)
	if dmv != nil {
		return dmv.GetAllInterest(stateDB)
	} else {
		return nil
	}
}
func AddInterest(stateDB vm.StateDBManager, address common.Address, reward *big.Int) error {
	dmv := selectDeposit(stateDB)
	if dmv != nil {
		return dmv.AddInterest(stateDB, address, reward)
	} else {
		return errors.New("unknow version AddInterest")
	}
}

func GetDepositAccount(stateDB vm.StateDBManager, authAccount common.Address) common.Address {
	dmv := selectDeposit(stateDB)
	if dmv != nil {
		return dmv.GetDepositAccount(stateDB, authAccount)
	} else {
		return common.Address{}
	}
}
func GetAuthAccount(stateDB vm.StateDBManager, depositAccount common.Address) common.Address {
	dmv := selectDeposit(stateDB)
	if dmv != nil {
		return dmv.GetAuthAccount(stateDB, depositAccount)
	} else {
		return common.Address{}
	}
}
func AddSlash_v2(stateDB vm.StateDBManager, address common.Address, slash common.CalculateDeposit) error {
	dmv := selectDeposit(stateDB)
	if dmv != nil {
		return dmv.AddSlash_v2(stateDB, address, slash)
	} else {
		return errors.New("AddSlash_v2 not find version,Maybe version is wrong")
	}
}
func GetSlash_v2(stateDB vm.StateDBManager, address common.Address) (common.CalculateDeposit, error) {
	dmv := selectDeposit(stateDB)
	if dmv == nil {
		return common.CalculateDeposit{}, errors.New("GetSlash_v2 not find version,Maybe version is wrong")
	}
	return dmv.GetSlash_v2(stateDB, address)
}
func GetInterest_v2(stateDB vm.StateDBManager, address common.Address) (common.CalculateDeposit, error) {
	dmv := selectDeposit(stateDB)
	if dmv == nil {
		return common.CalculateDeposit{}, errors.New("GetInterest_v2 not find version,Maybe version is wrong")
	}
	return dmv.GetInterest_v2(stateDB, address)

}

//headerTime:=0表示获取所有定期的仓位包括退选
func GetAllInterest_v2(stateDB vm.StateDBManager, headTime uint64) map[common.Address]common.CalculateDeposit {
	dmv := selectDeposit(stateDB)
	if dmv == nil {
		return nil
	}
	return dmv.GetAllInterest_v2(stateDB, headTime)

}
func AddInterest_v2(stateDB vm.StateDBManager, address common.Address, reward common.CalculateDeposit) error {
	dmv := selectDeposit(stateDB)
	if dmv != nil {
		return dmv.AddInterest_v2(stateDB, address, reward)
	} else {
		return errors.New("AddInterest_v2 not find version,Maybe version is wrong")
	}
}

//支付利息，合约接口会清除对应账户的累加利息和惩罚
func PayInterest(stateDB vm.StateDBManager, time uint64, addrA0 common.Address, position uint64, amount *big.Int) error {
	dmv := selectDeposit(stateDB)
	if dmv != nil {
		err := dmv.PayInterest(stateDB, addrA0, position, amount)
		if err != nil {
			return err
		}
		return vm.NewTransInterestsInterface().TransferInterests(amount, position, time, addrA0, stateDB)
	} else {
		return errors.New("PayInterest not find version,Maybe version is wrong")
	}
}
func GetDepositBase(statedb vm.StateDBManager, addr0 common.Address) *common.DepositBase {
	return depositmanagerversoin2.GetDepositBase(statedb, addr0)
}

func ConversionDeposit(statedb vm.StateDBManager, t uint64) error {
	v1result := depositmanagerversoin1.ConversionDeposit(statedb, t)
	v2result := depositmanagerversoin2.CheckDeposit(statedb)
	if v1result == nil || len(v1result) <= 0 {
		return errors.New("ConversionDeposit rlp encode err")
	}
	if v2result == nil || len(v2result) <= 0 {
		return errors.New("ConversionDeposit err，not find new deposit")
	}
	if len(v1result) != len(v2result) {
		return errors.New("old deposit num ≠ new deposit num")
	}
	for _, v2 := range v2result {
		if d, ok := v1result[v2.AddressA0]; ok {
			if len(v2.Dpstmsg) > 0 {
				if d.Role.Cmp(v2.Role) != 0 || !d.AddressA1.Equal(v2.AddressA1) || (d.DepositAmount.Cmp(v2.Dpstmsg[0].DepositAmount) != 0 && v2.Dpstmsg[0].DepositAmount.Cmp(big.NewInt(0)) > 0) {
					err := "Conversion Deposit err old Role " + d.Role.String() + " new Role" + v2.Role.String() + " old a1 addr " + d.AddressA1.Hex() + " new a1 addr" + v2.AddressA1.Hex() +
						" old deposit amount" + d.DepositAmount.String() + " new deposit amount " + v2.Dpstmsg[0].DepositAmount.String()
					return errors.New(err)
				}
			} else {
				return errors.New("new deposit information not exist")
			}
			if d.Withdraw > 0 && len(v2.Dpstmsg[0].WithDrawInfolist) == 1 {
				continue
			} else if d.Withdraw <= 0 && len(v2.Dpstmsg[0].WithDrawInfolist) <= 0 {
				continue
			} else {
				err := "withdraw information err, old withdraw" + strconv.FormatUint(d.Withdraw, 10) + " new withdraw len" + strconv.FormatInt(int64(len(v2.Dpstmsg[0].WithDrawInfolist)), 10)
				return errors.New(err)
			}
		} else {
			err := "A0 address " + v2.AddressA0.Hex() + " not exist"
			return errors.New(err)
		}
	}
	return nil
}
