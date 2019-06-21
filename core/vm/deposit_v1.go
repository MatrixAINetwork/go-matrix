// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package vm

import (
	"encoding/binary"
	"errors"
	"math/big"
	"strings"

	"github.com/MatrixAINetwork/go-matrix/reward/depositcfg"

	"github.com/MatrixAINetwork/go-matrix/accounts/abi"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/common/math"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/rlp"
)

var (
	man                = new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	minerThreshold     = new(big.Int).Mul(big.NewInt(10000), man)
	validatorThreshold = new(big.Int).Mul(big.NewInt(100000), man)
	withdrawState      = big.NewInt(1)

	errParameters        = errors.New("error parameters")
	errMethodId          = errors.New("error method id")
	errWithdraw          = errors.New("withdraw is not set")
	errExist             = errors.New("sign address exist")
	errClear             = errors.New("clear address invalid")
	errDeposit           = errors.New("deposit is not found")
	errOverflow          = errors.New("deposit is overflow")
	errDepositEmpty      = errors.New("depositList is Empty")
	errDepositRole       = errors.New("role is empty")
	errSlashOverflow     = errors.New("slash is overflow")
	errSlashEmpty        = errors.New("slash is empty")
	errInterestOverflow  = errors.New("interest id overflow")
	errInterestEmpty     = errors.New("interest is empty")
	errInterestAddrEmpty = errors.New("interest addr is empty")

	depositDef = ` [{"constant": true,"inputs": [],"name": "getDepositList","outputs": [{"name": "","type": "address[]"}],"payable": false,"stateMutability": "view","type": "function"},
			{"constant": true,"inputs": [{"name": "addr","type": "address"}],"name": "getDepositInfo","outputs": [{"name": "","type": "uint256"},{"name": "","type": "address"},{"name": "","type": "uint256"}, {"name": "","type": "uint256"}],"payable": false,"stateMutability": "view","type": "function"},
    		{"constant": false,"inputs": [{"name": "address","type": "address"}],"name": "valiDeposit","outputs": [],"payable": true,"stateMutability": "payable","type": "function"},
    		{"constant": false,"inputs": [{"name": "address","type": "address"}],"name": "minerDeposit","outputs": [],"payable": true,"stateMutability": "payable","type": "function"},
    		{"constant": false,"inputs": [],"name": "withdraw","outputs": [],"payable": false,"stateMutability": "nonpayable","type": "function"},
    		{"constant": false,"inputs": [],"name": "refund","outputs": [],"payable": false,"stateMutability": "nonpayable","type": "function"},
			{"constant": false,"inputs": [{"name": "addr","type": "address"}],"name": "interestAdd","outputs": [],"payable": true,"stateMutability": "payable","type": "function"},
			{"constant": false,"inputs": [{"name": "addr","type": "address"}],"name": "getinterest","outputs": [],"payable": false,"stateMutability": "payable","type": "function"}]`

	depositAbi, Abierr                                                                                                                  = abi.JSON(strings.NewReader(depositDef))
	valiDepositArr, minerDepositIdArr, withdrawIdArr, refundIdArr, getDepositListArr, getDepositInfoArr, interestAddArr, getinterestArr [4]byte
	emptyHash                                                                                                                           = common.Hash{}
)

func init() {
	if Abierr != nil {
		panic("err in deposit sc initialize")
	}

	copy(valiDepositArr[:], depositAbi.Methods["valiDeposit"].Id())
	copy(minerDepositIdArr[:], depositAbi.Methods["minerDeposit"].Id())
	copy(withdrawIdArr[:], depositAbi.Methods["withdraw"].Id())
	copy(refundIdArr[:], depositAbi.Methods["refund"].Id())
	copy(getDepositListArr[:], depositAbi.Methods["getDepositList"].Id())
	copy(getDepositInfoArr[:], depositAbi.Methods["getDepositInfo"].Id())
	copy(interestAddArr[:], depositAbi.Methods["interestAdd"].Id())
	copy(getinterestArr[:], depositAbi.Methods["getinterest"].Id())
}

type MatrixDeposit001 struct {
}

//func (md *MatrixDeposit001) RequiredGas(input []byte) uint64 {
//	if len(input) < 4 {
//		return 0
//	}
//	var methodIdArr [4]byte
//	copy(methodIdArr[:], input[:4])
//	if methodIdArr == interestAddArr {
//		return 0
//	}
//	return params.SstoreSetGas * 2
//}

func (md *MatrixDeposit001) Run(in []byte, contract *Contract, evm *EVM) ([]byte, error) {
	if in == nil || len(in) == 0 {
		return nil, nil
	}
	if len(in) < 4 {
		return nil, errParameters
	}
	var methodIdArr [4]byte
	copy(methodIdArr[:], in[:4])

	if methodIdArr == valiDepositArr {
		return md.valiDeposit(in[4:], contract, evm)
	} else if methodIdArr == minerDepositIdArr {
		return md.minerDeposit(in[4:], contract, evm)
	} else if methodIdArr == withdrawIdArr {
		return md.withdraw(in[4:], contract, evm)
	} else if methodIdArr == refundIdArr {
		return md.refund(in[4:], contract, evm)
	} else if methodIdArr == getDepositListArr {
		return md.getDepositList(contract, evm)
	} else if methodIdArr == getDepositInfoArr {
		return md.getDepositInfo(in[4:], contract, evm)
	} else if methodIdArr == interestAddArr {
		return md.interestAdd(in[4:], contract, evm)
	} else if methodIdArr == getinterestArr {
		return md.getinterest(in[4:], contract, evm)
	}
	return nil, errParameters
}

func (md *MatrixDeposit001) getinterest(in []byte, contract *Contract, evm *EVM) ([]byte, error) {
	if len(in) < 20 {
		return nil, errParameters
	}
	var addr common.Address

	err := depositAbi.Methods["getinterest"].Inputs.Unpack(&addr, in)
	if err != nil || len(addr) != 20 {
		return nil, errInterestAddrEmpty
	}
	amont := md.GetInterest(contract, evm.StateDB, addr)

	return depositAbi.Methods["getinterest"].Outputs.Pack(amont)
}
func (md *MatrixDeposit001) interestAdd(in []byte, contract *Contract, evm *EVM) ([]byte, error) {
	if len(in) < 20 {
		return nil, errParameters
	}
	var addr common.Address

	err := depositAbi.Methods["interestAdd"].Inputs.Unpack(&addr, in)
	if err != nil || len(addr) != 20 {
		return nil, errInterestAddrEmpty
	}
	isok := false
	for _, from := range common.WhiteAddrlist {
		if from == contract.caller.Address() {
			isok = true
		}
	}
	if isok {
		err = md.AddDeposit(contract, evm.StateDB, addr)
	} else {
		err = errors.New("This from can not Send interest Transaction")
	}
	return []byte{1}, err
}
func (md *MatrixDeposit001) valiDeposit(in []byte, contract *Contract, evm *EVM) ([]byte, error) {
	return md.deposit(in, contract, evm, validatorThreshold, big.NewInt(common.RoleValidator))
}

func (md *MatrixDeposit001) minerDeposit(in []byte, contract *Contract, evm *EVM) ([]byte, error) {
	return md.deposit(in, contract, evm, minerThreshold, big.NewInt(common.RoleMiner))
}

func (md *MatrixDeposit001) deposit(in []byte, contract *Contract, evm *EVM, threshold *big.Int, depositRole *big.Int) ([]byte, error) {
	if len(in) < 4 {
		return nil, errParameters
	}

	var addr common.Address
	err := depositAbi.Methods["valiDeposit"].Inputs.Unpack(&addr, in[:])
	if err != nil || len(addr) != 20 {
		return nil, errDeposit
	}

	deposit := md.getDeposit(contract, evm.StateDB, contract.CallerAddress)
	if deposit == nil {
		deposit = big.NewInt(0)
	}

	withdraw := md.getWithdrawHeight(contract, evm.StateDB, contract.CallerAddress)
	if withdraw != nil && withdraw.Sign() > 0 {
		return nil, errDeposit
	}

	deposit.Add(deposit, contract.value)

	if deposit.Cmp(threshold) < 0 {
		return nil, errDeposit
	}

	var address common.Address
	copy(address[:], addr[:])
	err = md.modifyDepositState(contract, evm, address, depositRole)
	if err != nil {
		return nil, err
	}

	return []byte{1}, nil
}

func (md *MatrixDeposit001) withdraw(in []byte, contract *Contract, evm *EVM) ([]byte, error) {
	err := md.modifyWithdrawState(contract, evm)
	if err != nil {
		return nil, err
	}

	return []byte{1}, nil
}

func (md *MatrixDeposit001) refund(in []byte, contract *Contract, evm *EVM) ([]byte, error) {
	value, err := md.modifyRefundState(contract, evm)
	if err != nil {
		return nil, err
	}
	if !evm.CanTransfer(evm.StateDB, contract.Address(), value, evm.Cointyp) {
		return nil, ErrInsufficientBalance
	}
	evm.Transfer(evm.StateDB, contract.Address(), contract.CallerAddress, value, evm.Cointyp)
	return []byte{1}, nil
}

func (md *MatrixDeposit001) GetOnlineTime(contract *Contract, stateDB StateDBManager, addr common.Address) *big.Int {
	onlineKey := append(addr[:], 'O', 'T')
	info := stateDB.GetState(contract.CoinTyp, contract.Address(), common.BytesToHash(onlineKey))
	if info != emptyHash {
		return info.Big()
	}
	return big.NewInt(0)
}

func (md *MatrixDeposit001) AddOnlineTime(contract *Contract, stateDB StateDBManager, address common.Address, ot *big.Int) error {
	onlineKey := append(address[:], 'O', 'T')
	info := stateDB.GetState(contract.CoinTyp, contract.Address(), common.BytesToHash(onlineKey))
	if info == emptyHash {
		info = common.BigToHash(big.NewInt(0))
	}
	dep := info.Big()
	dep.Add(dep, ot)
	if len(dep.Bytes()) > 32 {
		return errOverflow
	}
	stateDB.SetState(contract.CoinTyp, contract.Address(), common.BytesToHash(onlineKey), common.BigToHash(dep))
	return nil
}

func (md *MatrixDeposit001) SetOnlineTime(contract *Contract, stateDB StateDBManager, address common.Address, tm *big.Int) error {
	onlineKey := append(address[:], 'O', 'T')
	stateDB.SetState(contract.CoinTyp, contract.Address(), common.BytesToHash(onlineKey), common.BigToHash(tm))
	return nil
}

func (md *MatrixDeposit001) getDeposit(contract *Contract, stateDB StateDBManager, addr common.Address) *big.Int {
	depositKey := append(addr[:], 'D')
	info := stateDB.GetState(contract.CoinTyp, contract.Address(), common.BytesToHash(depositKey))
	if info != emptyHash {
		return info.Big()
	}
	return big.NewInt(0)
}

func (md *MatrixDeposit001) addDeposit(contract *Contract, stateDB StateDBManager) error {
	depositKey := append(contract.CallerAddress[:], 'D')
	info := stateDB.GetState(contract.CoinTyp, contract.Address(), common.BytesToHash(depositKey))
	dep := info.Big()
	dep.Add(dep, contract.value)
	if len(dep.Bytes()) > 32 {
		return errOverflow
	}
	stateDB.SetState(contract.CoinTyp, contract.Address(), common.BytesToHash(depositKey), common.BigToHash(dep))
	return nil
}

func (md *MatrixDeposit001) setDeposit(contract *Contract, stateDB StateDBManager, dep *big.Int) error {
	depositKey := append(contract.CallerAddress[:], 'D')
	stateDB.SetState(contract.CoinTyp, contract.Address(), common.BytesToHash(depositKey), common.BigToHash(dep))
	return nil
}

func (md *MatrixDeposit001) getAddress(contract *Contract, stateDB StateDBManager, addr common.Address) common.Address {
	// get signature address
	signAddrKey := append(addr[:], 'N', 'X')
	signAddr := stateDB.GetState(contract.CoinTyp, contract.Address(), common.BytesToHash(signAddrKey))
	if signAddr == emptyHash {
		return common.Address{}
	}
	return common.BytesToAddress(signAddr.Bytes())
}

func (md *MatrixDeposit001) setAddress(contract *Contract, stateDB StateDBManager, address common.Address) error {
	if (address == common.Address{}) {
		return nil
	}
	// set signature address : deposit address
	nodeXKey := append(contract.CallerAddress[:], 'N', 'X')
	addressA1 := stateDB.GetState(contract.CoinTyp, contract.Address(), common.BytesToHash(nodeXKey))
	if addressA1 == address.Hash() {
		return nil
	}

	nodeYKey := append(address[:], 'N', 'Y')
	addressA0 := stateDB.GetState(contract.CoinTyp, contract.Address(), common.BytesToHash(nodeYKey))
	if addressA0 != emptyHash {
		return errExist
	}
	//A0 A1都为空，
	//set A0
	stateDB.SetState(contract.CoinTyp, contract.Address(), common.BytesToHash(nodeYKey), contract.CallerAddress.Hash())

	// set old A0 address empty
	signAddr := md.getAddress(contract, stateDB, contract.CallerAddress)
	oldNodeYKey := append(signAddr[:], 'N', 'Y')
	stateDB.SetState(contract.CoinTyp, contract.Address(), common.BytesToHash(oldNodeYKey), common.Hash{})

	// set deposit address : signature address
	//set A1
	nodeXKey = append(contract.CallerAddress[:], 'N', 'X')
	stateDB.SetState(contract.CoinTyp, contract.Address(), common.BytesToHash(nodeXKey), address.Hash())
	return nil
}

func (md *MatrixDeposit001) clearAddress(contract *Contract, stateDB StateDBManager, address common.Address) error {
	if (address != common.Address{}) {
		return errClear
	}

	signAddr := md.getAddress(contract, stateDB, contract.CallerAddress)
	// signature address : []
	nodeYKey := append(signAddr[:], 'N', 'Y')
	stateDB.SetState(contract.CoinTyp, contract.Address(), common.BytesToHash(nodeYKey), common.Hash{})

	// deposit address : []
	nodeXKey := append(contract.CallerAddress[:], 'N', 'X')
	stateDB.SetState(contract.CoinTyp, contract.Address(), common.BytesToHash(nodeXKey), common.Hash{})
	return nil
}

func (md *MatrixDeposit001) getWithdrawHeight(contract *Contract, stateDB StateDBManager, addr common.Address) *big.Int {
	WDKey := append(addr[:], 'W', 'H')
	withdraw := stateDB.GetState(contract.CoinTyp, contract.Address(), common.BytesToHash(WDKey))
	if withdraw == emptyHash {
		return big.NewInt(0)
	}
	return withdraw.Big()
}

func (md *MatrixDeposit001) setWithdrawHeight(contract *Contract, stateDB StateDBManager, height *big.Int) error {
	withdrawKey := append(contract.CallerAddress[:], 'W', 'H')
	stateDB.SetState(contract.CoinTyp, contract.Address(), common.BytesToHash(withdrawKey), common.BigToHash(height))
	return nil
}

type DepositDetail struct {
	Address     common.Address
	SignAddress common.Address
	Deposit     *big.Int
	WithdrawH   *big.Int
	OnlineTime  *big.Int
	Role        *big.Int
}

func (md *MatrixDeposit001) getValidatorDepositList(contract *Contract, stateDB StateDBManager) []DepositDetail {
	var detailList []DepositDetail
	contractAddr := contract.Address()
	numKey := append(contractAddr[:], 'D', 'N', 'U', 'M')
	numHash := stateDB.GetState(contract.CoinTyp, contract.Address(), common.BytesToHash(numKey))
	num := numHash.Big()
	if num.Sign() != 0 {
		count := num.Uint64()
		for i := uint64(0); i < count; i++ {
			addr := md.getDepositListItem(contract, stateDB, i)
			detail, err := md.getDepositDetail(addr, contract, stateDB)
			if err == nil && detail.WithdrawH.Sign() == 0 && detail.Role.Cmp(big.NewInt(common.RoleValidator)) == 0 {
				detailList = append(detailList, *detail)
			}
		}
	}
	return detailList
}

func (md *MatrixDeposit001) GetValidatorDepositList(contract *Contract, stateDB StateDBManager) []DepositDetail {
	return md.getValidatorDepositList(contract, stateDB)
}

func (md *MatrixDeposit001) GetMinerDepositList(contract *Contract, stateDB StateDBManager) []DepositDetail {
	return md.getMinerDepositList(contract, stateDB)
}

func (md *MatrixDeposit001) GetAllDepositList(contract *Contract, stateDB StateDBManager, withDraw bool, headtime uint64) []DepositDetail {
	return md.getAllDepositList(contract, stateDB, withDraw)
}

func (md *MatrixDeposit001) getDepositList(contract *Contract, evm *EVM) ([]byte, error) {
	var addrList []common.Address
	contractAddr := contract.Address()
	numKey := append(contractAddr[:], 'D', 'N', 'U', 'M')
	numHash := evm.StateDB.GetState(evm.Cointyp, contract.Address(), common.BytesToHash(numKey))
	num := numHash.Big()
	if num.Sign() != 0 {
		count := num.Uint64()
		addrList = make([]common.Address, count)
		for i := uint64(0); i < count; i++ {
			addrList[i] = md.getDepositListItem(contract, evm.StateDB, i)
		}
	}
	return depositAbi.Methods["getDepositList"].Outputs.Pack(addrList)
}

func (md *MatrixDeposit001) getMinerDepositList(contract *Contract, stateDB StateDBManager) []DepositDetail {
	var detailList []DepositDetail
	contractAddr := contract.Address()
	numKey := append(contractAddr[:], 'D', 'N', 'U', 'M')
	numHash := stateDB.GetState(contract.CoinTyp, contract.Address(), common.BytesToHash(numKey))
	num := numHash.Big()
	if num.Sign() != 0 {
		count := num.Uint64()
		for i := uint64(0); i < count; i++ {
			addr := md.getDepositListItem(contract, stateDB, i)
			detail, err := md.getDepositDetail(addr, contract, stateDB)
			if err == nil && detail.WithdrawH.Sign() == 0 && detail.Role.Cmp(big.NewInt(common.RoleMiner)) == 0 {
				detailList = append(detailList, *detail)
			}
		}
	}
	return detailList
}

func (md *MatrixDeposit001) getAllDepositList(contract *Contract, stateDB StateDBManager, withDraw bool) []DepositDetail {
	var detailList []DepositDetail
	contractAddr := contract.Address()
	numKey := append(contractAddr[:], 'D', 'N', 'U', 'M')
	numHash := stateDB.GetState(contract.CoinTyp, contract.Address(), common.BytesToHash(numKey))
	num := numHash.Big()
	if num.Sign() != 0 {
		count := num.Uint64()
		for i := uint64(0); i < count; i++ {
			addr := md.getDepositListItem(contract, stateDB, i)
			detail, err := md.getDepositDetail(addr, contract, stateDB)
			if err != nil {
				//todo:返回空错误
				return detailList
			}
			if withDraw {
				detailList = append(detailList, *detail)
			} else if detail.WithdrawH.Sign() == 0 {
				detailList = append(detailList, *detail)
			}
		}
	}
	return detailList
}

func (md *MatrixDeposit001) getDepositDetail(addr common.Address, contract *Contract, stateDB StateDBManager) (*DepositDetail, error) {
	detail := DepositDetail{Address: addr}
	detail.Deposit = md.getDeposit(contract, stateDB, addr)
	if detail.Deposit == nil || detail.Deposit.Sign() == 0 {
		return nil, errDepositEmpty
	}
	detail.SignAddress = md.getAddress(contract, stateDB, addr)
	detail.WithdrawH = md.getWithdrawHeight(contract, stateDB, addr)
	detail.OnlineTime = md.GetOnlineTime(contract, stateDB, addr)
	detail.Role = md.getDepositRole(contract, stateDB, addr)
	return &detail, nil
}

func (md *MatrixDeposit001) getDepositListNum(contract *Contract, stateDB StateDBManager) *big.Int {
	contractAddr := contract.Address()
	numKey := append(contractAddr[:], 'D', 'N', 'U', 'M')
	num := stateDB.GetState(contract.CoinTyp, contract.Address(), common.BytesToHash(numKey))
	if num == emptyHash {
		return nil
	}
	return num.Big()
}

func (md *MatrixDeposit001) setDepositListNum(contract *Contract, stateDB StateDBManager, num *big.Int) {
	contractAddr := contract.Address()
	numKey := append(contractAddr[:], 'D', 'N', 'U', 'M')
	stateDB.SetState(contract.CoinTyp, contract.Address(), common.BytesToHash(numKey), common.BigToHash(num))
}

func (md *MatrixDeposit001) getDepositListItem(contract *Contract, stateDB StateDBManager, index uint64) common.Address {
	contractAddr := contract.Address()
	key := make([]byte, 8)
	binary.BigEndian.PutUint64(key, index)
	depKey := append(contractAddr[:], 'D', 'I')
	depKey = append(depKey, key...)
	addrHash := stateDB.GetState(contract.CoinTyp, contract.Address(), common.BytesToHash(depKey))
	return common.BytesToAddress(addrHash[:])
}

func (md *MatrixDeposit001) SetDepositListItem(contract *Contract, stateDB StateDBManager, index uint64, addr common.Address) {
	key := make([]byte, 8)
	binary.BigEndian.PutUint64(key, index)
	contractAddr := contract.Address()
	depKey := append(contractAddr[:], 'D', 'I')
	depKey = append(depKey, key...)
	stateDB.SetState(contract.CoinTyp, contract.Address(), common.BytesToHash(depKey), common.BytesToHash(addr[:]))
}

func (md *MatrixDeposit001) insertDepositList(contract *Contract, stateDB StateDBManager) {
	num := md.getDepositListNum(contract, stateDB)
	var count uint64
	if num != nil {
		count = num.Uint64()
		num.Add(num, big.NewInt(1))
	} else {
		num = big.NewInt(1)
	}
	md.setDepositListNum(contract, stateDB, num)
	md.SetDepositListItem(contract, stateDB, count, contract.CallerAddress)
}

func (md *MatrixDeposit001) removeDepositList(contract *Contract, stateDB StateDBManager) error {
	num := md.getDepositListNum(contract, stateDB)
	if num == nil {
		return errDepositEmpty
	}
	count := num.Uint64()
	insert := uint64(math.MaxUint64)
	for i := uint64(0); i < count; i++ {
		addr := md.getDepositListItem(contract, stateDB, i)
		if addr == contract.CallerAddress {
			insert = i
			break
		}
	}
	if insert != math.MaxUint64 {
		addr := md.getDepositListItem(contract, stateDB, count-1)
		md.SetDepositListItem(contract, stateDB, insert, addr)
		num.Sub(num, big.NewInt(1))
		md.setDepositListNum(contract, stateDB, num)
		return nil
	} else {
		return errDeposit
	}
}

func (md *MatrixDeposit001) getDepositInfo(in []byte, contract *Contract, evm *EVM) ([]byte, error) {
	var addr common.Address
	err := depositAbi.Methods["getDepositInfo"].Inputs.Unpack(&addr, in)
	if err != nil {
		return nil, err
	}
	deposit := md.getDeposit(contract, evm.StateDB, addr)
	if deposit == nil || deposit.Sign() == 0 {
		return nil, errDepositEmpty
	}

	/*get deposit role*/
	depositRole := md.getDepositRole(contract, evm.StateDB, addr)
	if depositRole.Cmp(big.NewInt(common.RoleMiner)) != 0 && depositRole.Cmp(big.NewInt(common.RoleValidator)) != 0 {
		return nil, errDepositRole
	}

	signAddr := md.getAddress(contract, evm.StateDB, addr)
	withdraw := md.getWithdrawHeight(contract, evm.StateDB, addr)

	return depositAbi.Methods["getDepositInfo"].Outputs.Pack(deposit, signAddr, withdraw, depositRole)
}

func (md *MatrixDeposit001) modifyDepositState(contract *Contract, evm *EVM, addr common.Address, depositRole *big.Int) error {
	deposit := md.getDeposit(contract, evm.StateDB, contract.CallerAddress)
	bNew := deposit == nil || deposit.Sign() == 0
	err := md.addDeposit(contract, evm.StateDB)
	if err != nil {
		return err
	}
	err = md.setAddress(contract, evm.StateDB, addr)
	if err != nil {
		return err
	}
	if bNew {
		md.insertDepositList(contract, evm.StateDB)
	}
	/*set deposit Role*/
	err = md.setDepositRole(contract, evm.StateDB, depositRole)
	if err != nil {
		return err
	}
	return nil
}

func (md *MatrixDeposit001) setDepositRole(contract *Contract, stateDB StateDBManager, depositRole *big.Int) error {
	depositRoleKey := append(contract.CallerAddress[:], 'R')
	stateDB.SetState(contract.CoinTyp, contract.Address(), common.BytesToHash(depositRoleKey), common.BigToHash(depositRole))
	return nil
}

func (md *MatrixDeposit001) getDepositRole(contract *Contract, stateDB StateDBManager, addr common.Address) *big.Int {
	depositRoleKey := append(addr[:], 'R')
	role := stateDB.GetState(contract.CoinTyp, contract.Address(), common.BytesToHash(depositRoleKey))

	if role != emptyHash {
		return role.Big()
	}
	return big.NewInt(common.RoleDefault)
}

func (md *MatrixDeposit001) modifyWithdrawState(contract *Contract, evm *EVM) error {
	deposit := md.getDeposit(contract, evm.StateDB, contract.CallerAddress)
	if deposit == nil || deposit.Sign() == 0 {
		return errDeposit
	}
	md.setWithdrawHeight(contract, evm.StateDB, evm.BlockNumber)
	return nil
}

func (md *MatrixDeposit001) modifyRefundState(contract *Contract, evm *EVM) (*big.Int, error) {
	deposit := md.getDeposit(contract, evm.StateDB, contract.CallerAddress)
	if deposit == nil || deposit.Sign() == 0 {
		return nil, errDeposit
	}
	withdrawHeight := md.getWithdrawHeight(contract, evm.StateDB, contract.CallerAddress)
	if withdrawHeight == nil || withdrawHeight.Sign() == 0 {
		return nil, errDeposit
	}
	withdrawHeight.Add(withdrawHeight, big.NewInt(600))
	if withdrawHeight.Cmp(evm.BlockNumber) > 0 {
		return nil, errDeposit
	}

	md.ResetSlash(contract, evm.StateDB, contract.CallerAddress)
	md.ResetInterest(contract, evm.StateDB, contract.CallerAddress)
	md.setDeposit(contract, evm.StateDB, big.NewInt(0))
	md.setDepositRole(contract, evm.StateDB, big.NewInt(0))
	md.clearAddress(contract, evm.StateDB, common.Address{})
	md.setWithdrawHeight(contract, evm.StateDB, big.NewInt(0))
	md.SetOnlineTime(contract, evm.StateDB, contract.CallerAddress, big.NewInt(0))
	md.removeDepositList(contract, evm.StateDB)
	return deposit, nil
}

// GetAllSlash get all account slash.
func (md *MatrixDeposit001) GetAllSlash(contract *Contract, stateDB StateDBManager) map[common.Address]*big.Int {
	slashList := make(map[common.Address]*big.Int)

	depositList := md.getAllDepositList(contract, stateDB, true)
	for _, deposit := range depositList {
		slash := md.GetSlash(contract, stateDB, deposit.Address)
		slashList[deposit.Address] = slash
	}

	return slashList
}

// GetSlash get current slash with state db and address.
func (md *MatrixDeposit001) GetSlash(contract *Contract, stateDB StateDBManager, addr common.Address) *big.Int {
	slashKey := append(addr[:], 'S', 'L', 'A', 'S', 'H')
	info := stateDB.GetState(contract.CoinTyp, contract.Address(), common.BytesToHash(slashKey))
	if info != emptyHash {
		return info.Big()
	}
	return big.NewInt(0)
}

// AddSlash add current slash with state db and address.
func (md *MatrixDeposit001) AddSlash(contract *Contract, stateDB StateDBManager, addr common.Address, slash *big.Int) error {
	info := md.GetSlash(contract, stateDB, addr)
	if info == nil {
		return errSlashEmpty
	}
	info.Add(info, slash)
	if len(info.Bytes()) > 32 {
		return errSlashOverflow
	}
	return md.SetSlash(contract, stateDB, addr, info)
}

// ResetSlash reset slash to zero with state db and address.
func (md *MatrixDeposit001) ResetSlash(contract *Contract, db StateDBManager, address common.Address) error {
	return md.SetSlash(contract, db, address, big.NewInt(0))
}

func (md *MatrixDeposit001) SetSlash(contract *Contract, stateDB StateDBManager, addr common.Address, slash *big.Int) error {
	slashKey := append(addr[:], 'S', 'L', 'A', 'S', 'H')
	stateDB.SetState(contract.CoinTyp, contract.Address(), common.BytesToHash(slashKey), common.BigToHash(slash))
	return nil
}

// GetAllInterest get all account interest.
func (md *MatrixDeposit001) GetAllInterest(contract *Contract, stateDB StateDBManager) map[common.Address]*big.Int {
	interestList := make(map[common.Address]*big.Int)

	depositList := md.getAllDepositList(contract, stateDB, true)
	for _, deposit := range depositList {
		interest := md.GetInterest(contract, stateDB, deposit.Address)
		interestList[deposit.Address] = interest
	}

	return interestList
}

// GetInterest get current interest with state db and address.
func (md *MatrixDeposit001) GetInterest(contract *Contract, stateDB StateDBManager, addr common.Address) *big.Int {
	interestKey := append(addr[:], 'R', 'E', 'W', 'A', 'R', 'D')
	info := stateDB.GetState(contract.CoinTyp, contract.Address(), common.BytesToHash(interestKey))
	if info != emptyHash {
		return info.Big()
	}
	return big.NewInt(0)
}

// AddInterest add current interest with state db and address.
func (md *MatrixDeposit001) AddInterest(contract *Contract, stateDB StateDBManager, addr common.Address, interest *big.Int) error {
	info := md.GetInterest(contract, stateDB, addr)
	if info == nil {
		return errInterestEmpty
	}
	info.Add(info, interest)
	if len(info.Bytes()) > 32 {
		return errInterestOverflow
	}
	return md.SetInterest(contract, stateDB, addr, info)
}

// ResetInterest reset interest to zero with state db and address.
func (md *MatrixDeposit001) ResetInterest(contract *Contract, db StateDBManager, address common.Address) error {
	return md.SetInterest(contract, db, address, big.NewInt(0))
}

func (md *MatrixDeposit001) SetInterest(contract *Contract, stateDB StateDBManager, addr common.Address, interest *big.Int) error {
	interestKey := append(addr[:], 'R', 'E', 'W', 'A', 'R', 'D')
	stateDB.SetState(contract.CoinTyp, contract.Address(), common.BytesToHash(interestKey), common.BigToHash(interest))
	return nil
}

// SetDeposit set deposit.
func (md *MatrixDeposit001) SetDeposit(contract *Contract, stateDB StateDBManager, address common.Address) error {
	depositKey := append(address[:], 'D')
	stateDB.SetState(contract.CoinTyp, contract.Address(), common.BytesToHash(depositKey), common.BigToHash(contract.value))
	return nil
}

// GetDeposit get deposit.
func (md *MatrixDeposit001) GetDeposit(contract *Contract, stateDB StateDBManager, address common.Address) *big.Int {
	return md.getDeposit(contract, stateDB, address)
}

// AddDeposit add deposit.
func (md *MatrixDeposit001) AddDeposit(contract *Contract, stateDB StateDBManager, address common.Address) error {
	dep := md.getDeposit(contract, stateDB, address)
	dep.Add(dep, contract.value)
	if len(dep.Bytes()) > 32 {
		return errOverflow
	}
	depositKey := append(address[:], 'D')
	stateDB.SetState(contract.CoinTyp, contract.Address(), common.BytesToHash(depositKey), common.BigToHash(dep))
	return md.ResetInterest(contract, stateDB, address)
}

func (md *MatrixDeposit001) GetDepositAccount(contract *Contract, stateDB StateDBManager, authAccount common.Address) common.Address {
	signAddrKey := append(authAccount[:], 'N', 'Y')
	signAddr := stateDB.GetState(contract.CoinTyp, contract.Address(), common.BytesToHash(signAddrKey))
	if signAddr == emptyHash {
		return common.Address{}
	}
	return common.BytesToAddress(signAddr.Bytes())
}

func (md *MatrixDeposit001) GetAuthAccount(contract *Contract, stateDB StateDBManager, depositAccount common.Address) common.Address {
	return md.getAddress(contract, stateDB, depositAccount)
}
func (md *MatrixDeposit001) ConversionDeposit(contract *Contract, statedb StateDBManager, t uint64) map[common.Address]common.CheckDepositInfo {
	depositdetails := md.getAllDepositList(contract, statedb, true)
	var (
		addrA0list    []common.Address
		addrMinerlist []common.Address
		addrValiList  []common.Address
	)
	if len(depositdetails) <= 0 {
		return nil
	}
	retVal := make(map[common.Address]common.CheckDepositInfo)
	for _, deposit := range depositdetails {
		var depositlist common.DepositBase
		var depositmsg common.DepositMsg
		depositmsg.DepositType = depositcfg.CurrentDeposit
		if deposit.WithdrawH.Cmp(big.NewInt(0)) > 0 {
			wi := common.WithDrawInfo{
				WithDrawAmount: new(big.Int),
			}
			wi.WithDrawTime = t + uint64(depositcfg.Days7Seconds)
			wi.WithDrawAmount = deposit.Deposit
			depositmsg.WithDrawInfolist = append(depositmsg.WithDrawInfolist, wi)
			depositmsg.DepositAmount = big.NewInt(0)
		} else {
			depositmsg.DepositAmount = deposit.Deposit
		}
		depositmsg.Interest = big.NewInt(0)
		depositmsg.Slash = big.NewInt(0)
		depositmsg.Position = 0
		addrA0list = append(addrA0list, deposit.Address)
		depositlist.AddressA0 = deposit.Address
		depositlist.AddressA1 = deposit.SignAddress
		depositlist.OnlineTime = deposit.OnlineTime
		depositlist.Dpstmsg = append(depositlist.Dpstmsg, depositmsg) //0仓-活期
		depositlist.Role = deposit.Role
		depositlist.PositionNonce = 0
		if deposit.Role.Cmp(big.NewInt(common.RoleValidator)) == 0 {
			addrValiList = append(addrValiList, deposit.Address)
		}
		if deposit.Role.Cmp(big.NewInt(common.RoleMiner)) == 0 {
			addrMinerlist = append(addrMinerlist, deposit.Address)
		}
		encodeDeposit, err := rlp.EncodeToBytes(depositlist)
		if err != nil {
			log.ERROR("translateDeposit", "rlp encode err", err)
			return nil
		}
		depositKeystr := append(deposit.Address[:], 'A', '0')
		depositKey := common.BytesToHash(depositKeystr)
		statedb.SetStateByteArray(contract.CoinTyp, contract.Address(), depositKey, encodeDeposit)
		a1keystr := append(deposit.SignAddress[:], 'A', '1')
		a1key := common.BytesToHash(a1keystr)
		value := deposit.Address
		statedb.SetStateByteArray(contract.CoinTyp, contract.Address(), a1key, value[:])
		retVal[deposit.Address] = common.CheckDepositInfo{
			DepositAmount: deposit.Deposit,
			Withdraw:      deposit.WithdrawH.Uint64(),
			Role:          deposit.Role,
			AddressA1:     deposit.SignAddress,
		}
	}
	numkey := common.BytesToHash([]byte(KeyDepositNum))
	bytenum := common.BigToHash(new(big.Int).SetUint64(uint64(len(depositdetails))))
	statedb.SetState(contract.CoinTyp, contract.Address(), numkey, bytenum)

	a0key := common.BytesToHash([]byte(KeyDepositA0list))
	ba0, _ := rlp.EncodeToBytes(addrA0list)
	statedb.SetStateByteArray(contract.CoinTyp, contract.Address(), a0key, ba0)

	var dr []common.DepositRoles
	dr = append(dr, common.DepositRoles{big.NewInt(common.RoleMiner), addrMinerlist})
	dr = append(dr, common.DepositRoles{big.NewInt(common.RoleValidator), addrValiList})
	rolekey := common.BytesToHash([]byte(KeyDepositRole))
	brole, _ := rlp.EncodeToBytes(dr)
	statedb.SetStateByteArray(contract.CoinTyp, contract.Address(), rolekey, brole)
	return retVal
}
