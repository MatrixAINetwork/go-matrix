// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package vm

import (
	"encoding/binary"
	"errors"
	"math/big"
	"strings"

	"github.com/matrix/go-matrix/accounts/abi"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/common/math"
	"github.com/matrix/go-matrix/p2p/discover"
	"github.com/matrix/go-matrix/params"
)

var (
	man                = new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	minerThreshold     = new(big.Int).Mul(big.NewInt(10000), man)
	validatorThreshold = new(big.Int).Mul(big.NewInt(100000), man)
	withdrawState      = big.NewInt(1)
	errParameters      = errors.New("error parameters")
	errMethodId        = errors.New("error method id")
	errDeposit         = errors.New("deposit is not found")
	errWithdraw        = errors.New("withdraw is not set")
	errOverflow        = errors.New("deposit is overflow")
	errDepositEmpty    = errors.New("depositList is Empty")

	depositDef = ` [{"constant": true,"inputs": [],"name": "getDepositList","outputs": [{"name": "","type": "address[]"}],"payable": false,"stateMutability": "view","type": "function"},
			{"constant": true,"inputs": [{"name": "addr","type": "address"}],"name": "getDepositInfo","outputs": [{"name": "","type": "uint256"},{"name": "","type": "bytes"},{"name": "","type": "uint256"}],"payable": false,"stateMutability": "view","type": "function"},
    		{"constant": false,"inputs": [{"name": "nodeID","type": "bytes"}],"name": "valiDeposit","outputs": [],"payable": true,"stateMutability": "payable","type": "function"},
    		{"constant": false,"inputs": [{"name": "nodeID","type": "bytes"}],"name": "minerDeposit","outputs": [],"payable": true,"stateMutability": "payable","type": "function"},
    		{"constant": false,"inputs": [],"name": "withdraw","outputs": [],"payable": false,"stateMutability": "nonpayable","type": "function"},
    		{"constant": false,"inputs": [],"name": "refund","outputs": [],"payable": false,"stateMutability": "nonpayable","type": "function"}]`

	depositAbi, Abierr                                                                                  = abi.JSON(strings.NewReader(depositDef))
	valiDepositArr, minerDepositIdArr, withdrawIdArr, refundIdArr, getDepositListArr, getDepositInfoArr [4]byte
	emptyHash                                                                                           = common.Hash{}
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
}

type MatrixDeposit struct {
}

func (md *MatrixDeposit) RequiredGas(input []byte) uint64 {
	if len(input) < 4 {
		return 0
	}
	return params.SstoreSetGas * 2
}

func (md *MatrixDeposit) Run(in []byte, contract *Contract, evm *EVM) ([]byte, error) {
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
	}
	return nil, errParameters
}

func (md *MatrixDeposit) valiDeposit(in []byte, contract *Contract, evm *EVM) ([]byte, error) {
	return md.deposit(in, contract, evm, validatorThreshold)
}

func (md *MatrixDeposit) minerDeposit(in []byte, contract *Contract, evm *EVM) ([]byte, error) {
	return md.deposit(in, contract, evm, minerThreshold)
}

func (md *MatrixDeposit) deposit(in []byte, contract *Contract, evm *EVM, threshold *big.Int) ([]byte, error) {
	if len(in) < 4 {
		return nil, errParameters
	}
	var nodeID []byte

	err := depositAbi.Methods["valiDeposit"].Inputs.Unpack(&nodeID, in)

	if err != nil || len(nodeID) != 64 {
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

	var discoverId discover.NodeID
	copy(discoverId[:], nodeID)
	md.modifyDepositState(contract, evm, discoverId)

	return []byte{1}, nil
}

func (md *MatrixDeposit) withdraw(in []byte, contract *Contract, evm *EVM) ([]byte, error) {
	err := md.modifyWithdrawState(contract, evm)
	if err != nil {
		return nil, err
	}

	return []byte{1}, nil
}

func (md *MatrixDeposit) refund(in []byte, contract *Contract, evm *EVM) ([]byte, error) {
	value, err := md.modifyRefundState(contract, evm)
	if err != nil {
		return nil, err
	}
	if !evm.CanTransfer(evm.StateDB, contract.Address(), value) {
		return nil, ErrInsufficientBalance
	}
	evm.Transfer(evm.StateDB, contract.Address(), contract.CallerAddress, value)
	return []byte{1}, nil
}

func (md *MatrixDeposit) GetOnlineTime(contract *Contract, stateDB StateDB, addr common.Address) *big.Int {
	onlineKey := append(addr[:], 'O', 'T')
	info := stateDB.GetState(contract.Address(), common.BytesToHash(onlineKey))
	if info != emptyHash {
		return info.Big()
	}
	return nil
}

func (md *MatrixDeposit) AddOnlineTime(contract *Contract, stateDB StateDB, address common.Address, ot *big.Int) error {
	onlineKey := append(address[:], 'O', 'T')
	info := stateDB.GetState(contract.Address(), common.BytesToHash(onlineKey))
	if info == emptyHash{
		info = common.BigToHash(big.NewInt(0))
	}
	dep := info.Big()
	dep.Add(dep, ot)
	if len(dep.Bytes()) > 32 {
		return errOverflow
	}
	stateDB.SetState(contract.Address(), common.BytesToHash(onlineKey), common.BigToHash(dep))
	return nil
}

func (md *MatrixDeposit) SetOnlineTime(contract *Contract, stateDB StateDB, address common.Address, tm *big.Int) error {
	onlineKey := append(address[:], 'O', 'T')
	stateDB.SetState(contract.Address(), common.BytesToHash(onlineKey), common.BigToHash(tm))
	return nil
}

func (md *MatrixDeposit) getDeposit(contract *Contract, stateDB StateDB, addr common.Address) *big.Int {
	depositKey := append(addr[:], 'D')
	info := stateDB.GetState(contract.Address(), common.BytesToHash(depositKey))
	if info != emptyHash {
		return info.Big()
	}
	return big.NewInt(0)
}

func (md *MatrixDeposit) addDeposit(contract *Contract, stateDB StateDB) error {
	depositKey := append(contract.CallerAddress[:], 'D')
	info := stateDB.GetState(contract.Address(), common.BytesToHash(depositKey))
	dep := info.Big()
	dep.Add(dep, contract.value)
	if len(dep.Bytes()) > 32 {
		return errOverflow
	}
	stateDB.SetState(contract.Address(), common.BytesToHash(depositKey), common.BigToHash(dep))
	return nil
}

func (md *MatrixDeposit) setDeposit(contract *Contract, stateDB StateDB, dep *big.Int) error {
	depositKey := append(contract.CallerAddress[:], 'D')
	stateDB.SetState(contract.Address(), common.BytesToHash(depositKey), common.BigToHash(dep))
	return nil
}

func (md *MatrixDeposit) getNodeID(contract *Contract, stateDB StateDB, addr common.Address) *discover.NodeID {
	nodeXKey := append(addr[:], 'N', 'X')
	nodeX := stateDB.GetState(contract.Address(), common.BytesToHash(nodeXKey))
	if nodeX == emptyHash {
		return &discover.NodeID{}
	}
	nodeYKey := append(addr[:], 'N', 'Y')
	nodeY := stateDB.GetState(contract.Address(), common.BytesToHash(nodeYKey))
	if nodeY == emptyHash {
		return &discover.NodeID{}
	}
	var nodeID discover.NodeID
	copy(nodeID[:32], nodeX[:])
	copy(nodeID[32:], nodeY[:])
	return &nodeID
}

func (md *MatrixDeposit) setNodeID(contract *Contract, stateDB StateDB, nodeID discover.NodeID) error {
	if (nodeID == discover.NodeID{}) {
		return nil
	}
	nodeXKey := append(contract.CallerAddress[:], 'N', 'X')
	stateDB.SetState(contract.Address(), common.BytesToHash(nodeXKey), common.BytesToHash(nodeID[:32]))
	nodeYKey := append(contract.CallerAddress[:], 'N', 'Y')
	stateDB.SetState(contract.Address(), common.BytesToHash(nodeYKey), common.BytesToHash(nodeID[32:]))
	return nil
}

func (md *MatrixDeposit) getWithdrawHeight(contract *Contract, stateDB StateDB, addr common.Address) *big.Int {
	WDKey := append(addr[:], 'W', 'H')
	withdraw := stateDB.GetState(contract.Address(), common.BytesToHash(WDKey))
	if withdraw == emptyHash {
		return big.NewInt(0)
	}
	return withdraw.Big()
}

func (md *MatrixDeposit) setWithdrawHeight(contract *Contract, stateDB StateDB, height *big.Int) error {
	withdrawKey := append(contract.CallerAddress[:], 'W', 'H')
	stateDB.SetState(contract.Address(), common.BytesToHash(withdrawKey), common.BigToHash(height))
	return nil
}

type DepositDetail struct {
	Address    common.Address
	NodeID     discover.NodeID
	Deposit    *big.Int
	WithdrawH  *big.Int
	OnlineTime *big.Int
}

func (md *MatrixDeposit) getValidatorDepositList(contract *Contract, stateDB StateDB) []DepositDetail {
	var detailList []DepositDetail
	contractAddr := contract.Address()
	numKey := append(contractAddr[:], 'D', 'N', 'U', 'M')
	//stateDB.get
	numHash := stateDB.GetState(contract.Address(), common.BytesToHash(numKey))
	num := numHash.Big()
	if num.Sign() != 0 {
		count := num.Uint64()
		for i := uint64(0); i < count; i++ {
			addr := md.getDepositListItem(contract, stateDB, i)
			detail, err := md.getDepositDetail(addr, contract, stateDB)
			if err == nil && detail.WithdrawH.Sign() == 0 && detail.Deposit.Cmp(validatorThreshold) >= 0 {
				detailList = append(detailList, *detail)
			}
		}
	}
	return detailList
}

func (md *MatrixDeposit) GetValidatorDepositList(contract *Contract, stateDB StateDB) []DepositDetail {
	return md.getValidatorDepositList(contract, stateDB)
}
func (md *MatrixDeposit) GetValidatorList(contract *Contract, stateDB StateDB) []DepositDetail {
	return md.getValidatorDepositList(contract, stateDB)
}

func (md *MatrixDeposit) GetMinerDepositList(contract *Contract, stateDB StateDB) []DepositDetail {
	return md.getMinerDepositList(contract, stateDB)
}
func (md *MatrixDeposit) GetMinerList(contract *Contract, stateDB StateDB) []DepositDetail {
	return md.getMinerDepositList(contract, stateDB)
}

func (md *MatrixDeposit) GetAllDepositList(contract *Contract, stateDB StateDB, withDraw bool) []DepositDetail {
	return md.getAllDepositList(contract, stateDB, withDraw)
}

func (md *MatrixDeposit) getDepositList(contract *Contract, evm *EVM) ([]byte, error) {
	var addrList []common.Address
	contractAddr := contract.Address()
	numKey := append(contractAddr[:], 'D', 'N', 'U', 'M')
	numHash := evm.StateDB.GetState(contract.Address(), common.BytesToHash(numKey))
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

func (md *MatrixDeposit) getMinerDepositList(contract *Contract, stateDB StateDB) []DepositDetail {
	var detailList []DepositDetail
	contractAddr := contract.Address()
	numKey := append(contractAddr[:], 'D', 'N', 'U', 'M')
	numHash := stateDB.GetState(contract.Address(), common.BytesToHash(numKey))
	num := numHash.Big()
	if num.Sign() != 0 {
		count := num.Uint64()
		for i := uint64(0); i < count; i++ {
			addr := md.getDepositListItem(contract, stateDB, i)
			detail, err := md.getDepositDetail(addr, contract, stateDB)
			if err == nil && detail.WithdrawH.Sign() == 0 && detail.Deposit.Cmp(validatorThreshold) < 0 {
				detailList = append(detailList, *detail)
			}
		}
	}
	return detailList
}

func (md *MatrixDeposit) getAllDepositList(contract *Contract, stateDB StateDB, withDraw bool) []DepositDetail {
	var detailList []DepositDetail
	contractAddr := contract.Address()
	numKey := append(contractAddr[:], 'D', 'N', 'U', 'M')
	numHash := stateDB.GetState(contract.Address(), common.BytesToHash(numKey))
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

func (md *MatrixDeposit) getDepositDetail(addr common.Address, contract *Contract, stateDB StateDB) (*DepositDetail, error) {
	detail := DepositDetail{Address: addr}
	detail.Deposit = md.getDeposit(contract, stateDB, addr)
	if detail.Deposit == nil || detail.Deposit.Sign() == 0 {
		return nil, errDepositEmpty
	}
	detail.NodeID = *(md.getNodeID(contract, stateDB, addr))
	detail.WithdrawH = md.getWithdrawHeight(contract, stateDB, addr)
	detail.OnlineTime = md.GetOnlineTime(contract, stateDB, addr)
	return &detail, nil
}

func (md *MatrixDeposit) getDepositListNum(contract *Contract, stateDB StateDB) *big.Int {
	contractAddr := contract.Address()
	numKey := append(contractAddr[:], 'D', 'N', 'U', 'M')
	num := stateDB.GetState(contract.Address(), common.BytesToHash(numKey))
	if num == emptyHash {
		return nil
	}
	return num.Big()
}

func (md *MatrixDeposit) setDepositListNum(contract *Contract, stateDB StateDB, num *big.Int) {
	contractAddr := contract.Address()
	numKey := append(contractAddr[:], 'D', 'N', 'U', 'M')
	stateDB.SetState(contract.Address(), common.BytesToHash(numKey), common.BigToHash(num))
}

func (md *MatrixDeposit) getDepositListItem(contract *Contract, stateDB StateDB, index uint64) common.Address {
	contractAddr := contract.Address()
	key := make([]byte, 8)
	binary.BigEndian.PutUint64(key, index)
	depKey := append(contractAddr[:], 'D', 'I')
	depKey = append(depKey, key...)
	addrHash := stateDB.GetState(contract.Address(), common.BytesToHash(depKey))
	return common.BytesToAddress(addrHash[:])
}

func (md *MatrixDeposit) SetDepositListItem(contract *Contract, stateDB StateDB, index uint64, addr common.Address) {
	key := make([]byte, 8)
	binary.BigEndian.PutUint64(key, index)
	contractAddr := contract.Address()
	depKey := append(contractAddr[:], 'D', 'I')
	depKey = append(depKey, key...)
	stateDB.SetState(contract.Address(), common.BytesToHash(depKey), common.BytesToHash(addr[:]))
}

func (md *MatrixDeposit) insertDepositList(contract *Contract, stateDB StateDB) {
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

func (md *MatrixDeposit) removeDepositList(contract *Contract, stateDB StateDB) error {
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

func (md *MatrixDeposit) getDepositInfo(in []byte, contract *Contract, evm *EVM) ([]byte, error) {
	var addr common.Address
	err := depositAbi.Methods["getDepositInfo"].Inputs.Unpack(&addr, in)
	if err != nil {
		return nil, err
	}
	deposit := md.getDeposit(contract, evm.StateDB, addr)
	if deposit == nil || deposit.Sign() == 0 {
		return nil, errDepositEmpty
	}
	nodeID := md.getNodeID(contract, evm.StateDB, addr)
	withdraw := md.getWithdrawHeight(contract, evm.StateDB, addr)
	return depositAbi.Methods["getDepositInfo"].Outputs.Pack(deposit, nodeID[:], withdraw)
}

func (md *MatrixDeposit) modifyDepositState(contract *Contract, evm *EVM, nodeID discover.NodeID) error {
	deposit := md.getDeposit(contract, evm.StateDB, contract.CallerAddress)
	bNew := deposit == nil || deposit.Sign() == 0
	err := md.addDeposit(contract, evm.StateDB)
	if err != nil {
		return err
	}
	md.setNodeID(contract, evm.StateDB, nodeID)
	if bNew {
		md.insertDepositList(contract, evm.StateDB)
	}
	return nil
}

func (md *MatrixDeposit) modifyWithdrawState(contract *Contract, evm *EVM) error {
	deposit := md.getDeposit(contract, evm.StateDB, contract.CallerAddress)
	if deposit == nil || deposit.Sign() == 0 {
		return errDeposit
	}
	md.setWithdrawHeight(contract, evm.StateDB, evm.BlockNumber)
	return nil
}

func (md *MatrixDeposit) modifyRefundState(contract *Contract, evm *EVM) (*big.Int, error) {
	deposit := md.getDeposit(contract, evm.StateDB, contract.CallerAddress)
	if deposit == nil || deposit.Sign() == 0 {
		return nil, errDeposit
	}
	withdrawHeight := md.getWithdrawHeight(contract, evm.StateDB, contract.CallerAddress)
	if withdrawHeight == nil || withdrawHeight.Sign() == 0 {
		return nil, errDeposit
	}
	withdrawHeight.Add(withdrawHeight, big.NewInt(1))
	if withdrawHeight.Cmp(evm.BlockNumber) > 0 {
		return nil, errDeposit
	}

	md.setDeposit(contract, evm.StateDB, big.NewInt(0))
	md.setNodeID(contract, evm.StateDB, discover.NodeID{})
	md.setWithdrawHeight(contract, evm.StateDB, big.NewInt(0))
	md.SetOnlineTime(contract, evm.StateDB, contract.CallerAddress, big.NewInt(0))
	md.removeDepositList(contract, evm.StateDB)
	return deposit, nil
}
func (md *MatrixDeposit) GetSlash(contract *Contract, stateDB StateDB, addr common.Address) *big.Int {
	slashKey := append(addr[:], 'S', 'L', 'A', 'S', 'H')
	info := stateDB.GetState(contract.Address(), common.BytesToHash(slashKey))
	if info != emptyHash {
		return info.Big()
	}
	return nil
}

func (md *MatrixDeposit) AddSlash(contract *Contract, stateDB StateDB, addr common.Address, slash *big.Int) error {
	slashKey := append(addr[:], 'S', 'L', 'A', 'S', 'H')
	info := stateDB.GetState(contract.Address(), common.BytesToHash(slashKey))
	dep := info.Big()
	dep.Add(dep, slash)
	if len(dep.Bytes()) > 32 {
		return errOverflow
	}
	stateDB.SetState(contract.Address(), common.BytesToHash(slashKey), common.BigToHash(dep))
	return nil
}

func (md *MatrixDeposit) SetSlash(contract *Contract, stateDB StateDB, addr common.Address, slash *big.Int) error {
	slashKey := append(addr[:], 'S', 'L', 'A', 'S', 'H')
	stateDB.SetState(contract.Address(), common.BytesToHash(slashKey), common.BigToHash(slash))
	return nil
}

func (md *MatrixDeposit) GetReward(contract *Contract, stateDB StateDB, addr common.Address) *big.Int {
	rewardKey := append(addr[:], 'R', 'E', 'W', 'A', 'R', 'D')
	info := stateDB.GetState(contract.Address(), common.BytesToHash(rewardKey))
	if info != emptyHash {
		return info.Big()
	}
	return nil
}

func (md *MatrixDeposit) AddReward(contract *Contract, stateDB StateDB, addr common.Address, slash *big.Int) error {
	rewardKey := append(addr[:], 'R', 'E', 'W', 'A', 'R', 'D')
	info := stateDB.GetState(contract.Address(), common.BytesToHash(rewardKey))
	dep := info.Big()
	dep.Add(dep, slash)
	if len(dep.Bytes()) > 32 {
		return errOverflow
	}
	stateDB.SetState(contract.Address(), common.BytesToHash(rewardKey), common.BigToHash(dep))
	return nil
}

func (md *MatrixDeposit) SetReward(contract *Contract, stateDB StateDB, addr common.Address, slash *big.Int) error {
	rewardKey := append(addr[:], 'R', 'E', 'W', 'A', 'R', 'D')
	stateDB.SetState(contract.Address(), common.BytesToHash(rewardKey), common.BigToHash(slash))
	return nil
}
