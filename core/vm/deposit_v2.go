// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package vm

import (
	"math/big"
	"strings"

	"github.com/MatrixAINetwork/go-matrix/accounts/abi"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/reward/depositcfg"
	"github.com/MatrixAINetwork/go-matrix/rlp"
	"github.com/pkg/errors"
)

//{"constant": true,"inputs": [{"name": "addr","type": "address"}],"name": "getDepositInfo","outputs": [{"name": "","type": "uint256"},{"name": "","type": "address"},{"name": "","type": "uint256"}, {"name": "","type": "uint256"}],"payable": false,"stateMutability": "view","type": "function"},
var (
	depositDef_v2 = ` [{"constant": true,"inputs": [],"name": "getDepositList","outputs": [{"name": "","type": "address[]"}],"payable": false,"stateMutability": "view","type": "function"},	
    		{"constant": false,"inputs": [{"name": "address","type": "address"},{"name": "depositType","type": "uint256"}],"name": "valiDeposit","outputs": [],"payable": true,"stateMutability": "payable","type": "function"},
    		{"constant": false,"inputs": [{"name": "address","type": "address"},{"name": "depositType","type": "uint256"}],"name": "minerDeposit","outputs": [],"payable": true,"stateMutability": "payable","type": "function"},
    		{"constant": false,"inputs": [{"name": "depositPosition","type": "uint256"},{"name": "withdrawAmount","type": "uint256"}],"name": "withdraw","outputs": [],"payable": false,"stateMutability": "nonpayable","type": "function"},
    		{"constant": false,"inputs": [{"name": "depositPosition","type": "uint256"}],"name": "refund","outputs": [],"payable": false,"stateMutability": "nonpayable","type": "function"},
			{"constant": false,"inputs": [{"name": "depositType","type": "uint256"},{"name": "amount","type": "uint256"}],"name": "modifyDepositType","outputs": [],"payable": true,"stateMutability": "payable","type": "function"},
			{"constant": true,"inputs": [{"name": "addr","type": "address"}],"name": "getInterest","outputs": [],"payable": false,"stateMutability": "payable","type": "function"}]`

	depositAbi_v2, Abierr_v2                                                                                                              = abi.JSON(strings.NewReader(depositDef_v2))
	valiDepositArr_v2, minerDepositIdArr_v2, withdrawIdArr_v2, refundIdArr_v2, getDepositListArr_v2, getinterestArr_v2, modifyDeposittype [4]byte
)

func init() {
	if Abierr_v2 != nil {
		panic("err in deposit sc initialize")
	}

	copy(valiDepositArr_v2[:], depositAbi_v2.Methods["valiDeposit"].Id())
	copy(minerDepositIdArr_v2[:], depositAbi_v2.Methods["minerDeposit"].Id())
	copy(withdrawIdArr_v2[:], depositAbi_v2.Methods["withdraw"].Id())
	copy(refundIdArr_v2[:], depositAbi_v2.Methods["refund"].Id())
	copy(getDepositListArr_v2[:], depositAbi_v2.Methods["getDepositList"].Id())
	//copy(getDepositInfoArr_v2[:], depositAbi_v2.Methods["getDepositInfo"].Id())
	copy(modifyDeposittype[:], depositAbi_v2.Methods["modifyDepositType"].Id())
	copy(getinterestArr_v2[:], depositAbi_v2.Methods["getinterest"].Id())
}

type MatrixDeposit002 struct {
}

const (
	KeyDepositNum    = "DepositNum"
	KeyDepositA0list = "DepositA0list"
	KeyDepositRole   = "DepositRole"
	KeyDepositInfo   = "DepositInfo"
)

func (md *MatrixDeposit002) Run(in []byte, contract *Contract, evm *EVM) ([]byte, error) {
	if in == nil || len(in) == 0 {
		return nil, nil
	}
	if len(in) < 4 {
		return nil, errParameters
	}
	var methodIdArr [4]byte
	copy(methodIdArr[:], in[:4])
	if methodIdArr == valiDepositArr_v2 {
		return md.valiDeposit_v2(in[4:], contract, evm)
	} else if methodIdArr == minerDepositIdArr_v2 {
		return md.minerDeposit_v2(in[4:], contract, evm)
	} else if methodIdArr == withdrawIdArr_v2 {
		return md.withdraw_v2(in[4:], contract, evm)
	} else if methodIdArr == refundIdArr_v2 {
		return md.refund_v2(in[4:], contract, evm)
	} else if methodIdArr == getDepositListArr_v2 {
		return md.getDepositList(contract, evm)
	} else if methodIdArr == getinterestArr_v2 {
		return md.getinterest(in[4:], contract, evm)
	} else if methodIdArr == modifyDeposittype { //修改抵押类型（活->定）
		return md.modifyDeposittype(in[4:], contract, evm)
	}
	return nil, errParameters
}

// GetSlash get current slash with state db and address.
func (md *MatrixDeposit002) GetSlash(contract *Contract, stateDB StateDBManager, addr common.Address) common.CalculateDeposit {
	var rutlist common.CalculateDeposit
	dpb := md.GetDepositBase(contract, stateDB, addr)
	if dpb == nil {
		log.Error("GetSlash", "GetDepositBase err", "GetDepositBase is nil")
		return common.CalculateDeposit{}
	}
	rutlist.AddressA0 = dpb.AddressA0
	var opre []common.OperationalInterestSlash
	for _, slash := range dpb.Dpstmsg {
		tmp := common.OperationalInterestSlash{
			DepositAmount: new(big.Int),
			OperAmount:    new(big.Int), //用来增加的利息金额或者惩罚的金额
		}
		tmp.DepositAmount = slash.DepositAmount
		tmp.DepositType = slash.DepositType
		tmp.OperAmount = slash.Slash
		tmp.Position = slash.Position
		opre = append(opre, tmp)
	}
	rutlist.CalcDeposit = opre
	return rutlist
}

// AddSlash add current slash with state db and address.
func (md *MatrixDeposit002) AddSlash(contract *Contract, stateDB StateDBManager, addr common.Address, slash common.CalculateDeposit) error {
	return md.SetSlash(contract, stateDB, addr, slash)
}
func (md *MatrixDeposit002) GetAllDepositListByInterest(contract *Contract, stateDB StateDBManager, withDraw bool, headtime uint64) []common.DepositBase {
	var detailList []common.DepositBase
	listkey := common.BytesToHash([]byte(KeyDepositA0list))
	retbuf := stateDB.GetStateByteArray(contract.CoinTyp, contract.Address(), listkey)
	var addrA0list []common.Address
	err := rlp.DecodeBytes(retbuf, &addrA0list)
	if err != nil {
		log.ERROR("getAllDepositInfo KeyDepositA0list", "decode err", err)
		return nil
	}
	for _, addr := range addrA0list {
		dpb := md.GetDepositBase(contract, stateDB, addr)
		if dpb == nil {
			log.ERROR("GetDepositBase", "GetDepositBase err", "GetDepositBase is nil ")
			continue
		}

		if !withDraw { //withDraw=true表示退选的也要，false表示退选的不要了
			if !md.CheckWithdrawByInterest(dpb, headtime) {
				//表示总抵押值不满足角色要求的金额
				continue
			}
		}
		detailList = append(detailList, *dpb)
	}
	return detailList
}
func (md *MatrixDeposit002) CheckWithdrawByInterest(dpb *common.DepositBase, headtime uint64) bool {
	var depositlist []common.DepositMsg
	amount := big.NewInt(0)
	if dpb.Role.Cmp(new(big.Int).SetUint64(uint64(common.RoleValidator))) != 0 && dpb.Role.Cmp(new(big.Int).SetUint64(uint64(common.RoleMiner))) != 0 {
		return false
	}
	for i, d := range dpb.Dpstmsg {
		if i == 0 && d.DepositAmount.Sign() > 0 {
			amount.Add(amount, d.DepositAmount)
			depositlist = append(depositlist, dpb.Dpstmsg[i])
			continue //活期如果有钱就要计算利息不做检查
		}
		if d.EndTime > 0 && d.EndTime < headtime {
			continue
		}
		amount.Add(amount, d.DepositAmount)
		depositlist = append(depositlist, dpb.Dpstmsg[i])
	}
	dpb.Dpstmsg = depositlist
	return true
}
func (md *MatrixDeposit002) GetAllDepositList(contract *Contract, stateDB StateDBManager, withDraw bool, headtime uint64) []common.DepositBase {
	var detailList []common.DepositBase
	listkey := common.BytesToHash([]byte(KeyDepositA0list))
	retbuf := stateDB.GetStateByteArray(contract.CoinTyp, contract.Address(), listkey)
	var addrA0list []common.Address
	err := rlp.DecodeBytes(retbuf, &addrA0list)
	if err != nil {
		log.ERROR("getAllDepositInfo KeyDepositA0list", "decode err", err)
		return nil
	}
	for _, addr := range addrA0list {
		dpb := md.GetDepositBase(contract, stateDB, addr)
		if dpb == nil {
			log.ERROR("GetDepositBase", "GetDepositBase err", "GetDepositBase is nil ")
			continue
		}

		if !withDraw { //withDraw=true表示退选的也要，false表示退选的不要了
			if !md.CheckWithdraw(dpb, headtime) {
				//表示总抵押值不满足角色要求的金额
				continue
			}
		}
		detailList = append(detailList, *dpb)
	}
	return detailList
}
func (md *MatrixDeposit002) CheckWithdraw(dpb *common.DepositBase, headtime uint64) bool {
	var depositlist []common.DepositMsg
	amount := big.NewInt(0)
	if dpb.Role.Cmp(new(big.Int).SetUint64(uint64(common.RoleValidator))) != 0 && dpb.Role.Cmp(new(big.Int).SetUint64(uint64(common.RoleMiner))) != 0 {
		return false
	}
	for i, d := range dpb.Dpstmsg {
		if i == 0 && d.DepositAmount.Sign() > 0 {
			amount.Add(amount, d.DepositAmount)
			depositlist = append(depositlist, dpb.Dpstmsg[i])
			continue //活期如果有钱就要计算利息不做检查
		}
		if d.EndTime > 0 && d.EndTime < headtime {
			continue
		}
		amount.Add(amount, d.DepositAmount)
		depositlist = append(depositlist, dpb.Dpstmsg[i])
	}
	dpb.Dpstmsg = depositlist
	if dpb.Role.Cmp(new(big.Int).SetUint64(uint64(common.RoleValidator))) == 0 {
		if amount.Cmp(validatorThreshold) < 0 {
			return false
		}
	}
	if dpb.Role.Cmp(new(big.Int).SetUint64(uint64(common.RoleMiner))) == 0 {
		if amount.Cmp(minerThreshold) < 0 {
			return false
		}
	}
	return true
}

// ResetInterest reset interest to zero with state db and address.
func (md *MatrixDeposit002) ResetInterest(contract *Contract, db StateDBManager, address common.Address) error {
	ret := md.GetInterest(contract, db, address)
	for i, _ := range ret.CalcDeposit {
		ret.CalcDeposit[i].OperAmount = big.NewInt(0)
	}
	return md.SetInterest(contract, db, address, ret)
}
func (md *MatrixDeposit002) SetInterest(contract *Contract, stateDB StateDBManager, addr common.Address, interest common.CalculateDeposit) error {
	dpb := md.GetDepositBase(contract, stateDB, addr)
	if dpb == nil {
		log.Error("SetInterest ", "GetDepositBase err", "GetDepositBase is nil")
		return errors.New("get DepositBase is nil")
	}
	for _, intrest := range interest.CalcDeposit {
		for i, msg := range dpb.Dpstmsg {
			if intrest.Position == msg.Position {
				dpb.Dpstmsg[i].Interest = intrest.OperAmount
				break
			}
		}
	}
	return md.SetDepositBase(contract, stateDB, addr, dpb)
}

// GetInterest get current interest with state db and address.
func (md *MatrixDeposit002) GetInterest(contract *Contract, stateDB StateDBManager, addr common.Address) common.CalculateDeposit {
	var rutlist common.CalculateDeposit
	dpb := md.GetDepositBase(contract, stateDB, addr)
	if dpb == nil {
		log.Error("GetInterest", "GetDepositBase err", "GetDepositBase is nil")
		return common.CalculateDeposit{}
	}
	rutlist.AddressA0 = addr
	var opre []common.OperationalInterestSlash
	for _, inter := range dpb.Dpstmsg {
		tmp := common.OperationalInterestSlash{
			DepositAmount: new(big.Int),
			OperAmount:    new(big.Int), //用来增加的利息金额或者惩罚的金额
		}
		tmp.OperAmount = inter.Interest
		tmp.DepositAmount = inter.DepositAmount
		tmp.Position = inter.Position
		tmp.DepositType = inter.DepositType
		opre = append(opre, tmp)
	}
	rutlist.CalcDeposit = opre
	return rutlist
}

// GetAllInterest get all account interest.
func (md *MatrixDeposit002) GetAllInterest(contract *Contract, stateDB StateDBManager, headTime uint64) map[common.Address]common.CalculateDeposit {
	interestList := make(map[common.Address]common.CalculateDeposit)

	listkey := common.BytesToHash([]byte(KeyDepositA0list))
	retbuf := stateDB.GetStateByteArray(contract.CoinTyp, contract.Address(), listkey)
	var addrA0list []common.Address
	err := rlp.DecodeBytes(retbuf, &addrA0list)
	if err != nil {
		log.ERROR("getAllDepositInfo KeyDepositA0list", "decode err", err)
		return nil
	}
	for _, addr := range addrA0list {
		dpb := md.GetDepositBase(contract, stateDB, addr)
		if dpb == nil {
			log.ERROR("GetAllInterest", "GetDepositBase err", "GetDepositBase is nil ")
			continue
		}
		var rutlist common.CalculateDeposit
		rutlist.AddressA0 = addr
		opre := make([]common.OperationalInterestSlash, 0)
		for _, inter := range dpb.Dpstmsg {
			if inter.DepositAmount.Sign() <= 0 {
				continue
			}
			if inter.EndTime > 0 && inter.EndTime < headTime {
				continue
			}
			if inter.Interest.Sign() <= 0 {
				continue
			}
			tmp := common.OperationalInterestSlash{
				DepositAmount: new(big.Int),
				OperAmount:    new(big.Int), //用来增加的利息金额或者惩罚的金额
			}
			tmp.OperAmount = inter.Interest
			tmp.DepositType = inter.DepositType
			tmp.Position = inter.Position
			tmp.DepositAmount = inter.DepositAmount
			opre = append(opre, tmp)
		}
		rutlist.CalcDeposit = make([]common.OperationalInterestSlash, len(opre))
		copy(rutlist.CalcDeposit, opre)
		interestList[addr] = rutlist
	}
	return interestList
}

// AddInterest add current interest with state db and address.
func (md *MatrixDeposit002) AddInterest(contract *Contract, stateDB StateDBManager, addr common.Address, interest common.CalculateDeposit) error {
	return md.SetInterest(contract, stateDB, addr, interest)
}
func (md *MatrixDeposit002) getinterest(in []byte, contract *Contract, evm *EVM) ([]byte, error) {
	if len(in) < 20 {
		return nil, errParameters
	}
	var addr common.Address

	err := depositAbi_v2.Methods["getinterest"].Inputs.Unpack(&addr, in)
	if err != nil || len(addr) != 20 {
		return nil, errInterestAddrEmpty
	}
	amont := md.GetInterest(contract, evm.StateDB, addr)

	return depositAbi_v2.Methods["getinterest"].Outputs.Pack(amont)
}
func (md *MatrixDeposit002) modifyDeposittype(in []byte, contract *Contract, evm *EVM) ([]byte, error) {
	if len(in) < 20 {
		return nil, errParameters
	}
	var addr common.Address
	depositType := big.NewInt(0)
	amount := big.NewInt(0)
	err := depositAbi_v2.Methods["modifyDepositType"].Inputs.Unpack(&[]interface{}{&depositType, &amount}, in[:])
	if err != nil || len(addr) != 20 {
		return nil, errInterestAddrEmpty
	}
	depocfg := depositcfg.GetDepositCfg(depositcfg.VersionA).GetDepositPositionCfg(depositType.Uint64())
	if depocfg == nil {
		return nil, errors.New("get depositype err")
	}
	if depositType.Sign() == 0 {
		return nil, errors.New("can not Convert 0 Position")
	}
	if amount.Cmp(depositcfg.FixDepositAmountMin) < 0 {
		return nil, errors.New("deposit amount must more 10000")
	}
	depositInfo := md.GetDepositBase(contract, evm.StateDB, contract.CallerAddress)
	if depositInfo != nil && len(depositInfo.Dpstmsg) > 0 {
		if depositInfo.Dpstmsg[0].DepositAmount.Cmp(amount) >= 0 {
			depositInfo.Dpstmsg[0].DepositAmount.Sub(depositInfo.Dpstmsg[0].DepositAmount, amount)
			depositInfo.Dpstmsg = append(depositInfo.Dpstmsg, common.DepositMsg{
				DepositType:      depositType.Uint64(),
				DepositAmount:    amount,
				Interest:         big.NewInt(0),
				Slash:            big.NewInt(0),
				BeginTime:        evm.Time.Uint64(),             //定期起始时间，为当前确认时间(evm.Time)
				Position:         depositInfo.PositionNonce + 1, //仓位
				WithDrawInfolist: make([]common.WithDrawInfo, 0),
			})
			depositInfo.PositionNonce += 1
		} else {
			log.Error("modifyDepositType", "0 Position amount not enough")
			return []byte{1}, errors.New("")
		}
	} else {
		str := "modifydeposittype err,deposit not exist.A0 addr = " + contract.CallerAddress.Hex()
		return []byte{1}, errors.New(str)
	}
	err = md.SetDepositBase(contract, evm.StateDB, contract.CallerAddress, depositInfo)
	if err != nil {
		return []byte{0}, err
	}
	return []byte{1}, nil
}

func (md *MatrixDeposit002) getDepositList(contract *Contract, evm *EVM) ([]byte, error) {
	a0key := common.BytesToHash([]byte(KeyDepositA0list))
	ret := evm.StateDB.GetStateByteArray(contract.CoinTyp, contract.Address(), a0key)
	var addrA0list []common.Address
	err := rlp.DecodeBytes(ret, &addrA0list)
	if err != nil {
		log.Error("getDepositList", "err", err)
		return []byte{0}, err
	}
	return depositAbi_v2.Methods["getDepositList"].Outputs.Pack(addrA0list)
}

func (md *MatrixDeposit002) modifyDepositState_v2(contract *Contract, depositType uint64, evm *EVM, addrA1 common.Address, depositRole *big.Int, dpb *common.DepositBase) error {
	if dpb == nil {
		dpb = new(common.DepositBase)
		num := md.getDepositListNum(contract, evm.StateDB)
		if num != nil {
			num.Add(num, big.NewInt(1))
		} else {
			num = big.NewInt(1)
		}
		md.setDepositListNum(contract, evm.StateDB, num)
		dpb.AddressA0 = contract.CallerAddress
		dpb.AddressA1 = addrA1
		dpb.OnlineTime = big.NewInt(0)
		var dm common.DepositMsg
		dm.BeginTime = evm.Time.Uint64()
		dm.DepositAmount = contract.Value()
		dm.DepositType = depositType
		dm.Interest = big.NewInt(0)
		dm.Slash = big.NewInt(0)
		if depositType == 0 {
			dm.Position = 0
		} else {
			dm.Position = 1
		}
		if depositType == depositcfg.CurrentDeposit {
			dpb.Dpstmsg = append(dpb.Dpstmsg, dm)
		} else {
			empt := common.DepositMsg{
				DepositAmount: big.NewInt(0),
				Interest:      big.NewInt(0),
				Slash:         big.NewInt(0),
			}
			dpb.Dpstmsg = append(dpb.Dpstmsg, empt)
			dpb.Dpstmsg = append(dpb.Dpstmsg, dm)
		}
		dpb.PositionNonce = dm.Position
		dpb.Role = depositRole
		err := md.SetA0list(contract, evm.StateDB)
		if err != nil {
			return err
		}
	} else {
		depositlen := len(dpb.Dpstmsg)
		if depositlen > 0 {
			if depositType == depositcfg.CurrentDeposit {
				dpb.Dpstmsg[0].DepositAmount.Add(dpb.Dpstmsg[0].DepositAmount, contract.Value())
			} else {
				dpb.Dpstmsg = append(dpb.Dpstmsg, common.DepositMsg{
					DepositType:      depositType,
					DepositAmount:    contract.Value(),
					Interest:         big.NewInt(0),
					Slash:            big.NewInt(0),
					BeginTime:        evm.Time.Uint64(),     //定期起始时间，为当前确认时间(evm.Time)
					Position:         dpb.PositionNonce + 1, //仓位
					WithDrawInfolist: make([]common.WithDrawInfo, 0),
				})
				dpb.PositionNonce += 1
			}
			dpb.Role = depositRole
		}
	}
	err := md.SetDepositRole(contract, evm.StateDB, depositRole)
	if err != nil {
		return err
	}
	err = md.setAddress(contract, evm.StateDB, addrA1, dpb)
	if err != nil {
		return err
	}
	return md.SetDepositBase(contract, evm.StateDB, contract.CallerAddress, dpb)
}
func (md *MatrixDeposit002) SetA0list(contract *Contract, stateDB StateDBManager) error {
	a0key := common.BytesToHash([]byte(KeyDepositA0list))
	addrA0list, err := md.GetA0list(contract, stateDB)
	if err != nil {
		return err
	}
	addrA0list = append(addrA0list, contract.CallerAddress)
	ba0, _ := rlp.EncodeToBytes(addrA0list)
	stateDB.SetStateByteArray(contract.CoinTyp, contract.Address(), a0key, ba0)
	return nil
}
func (md *MatrixDeposit002) GetA0list(contract *Contract, stateDB StateDBManager) ([]common.Address, error) {
	//a0key := common.HexToHash(KeyDepositA0list)
	a0key := common.BytesToHash([]byte(KeyDepositA0list))
	ret := stateDB.GetStateByteArray(contract.CoinTyp, contract.Address(), a0key)
	var addrA0list []common.Address
	if len(ret) == 0 {
		return addrA0list, nil
	}
	err := rlp.DecodeBytes(ret, &addrA0list)
	if err != nil {
		log.Error("SetA0list", "err", err)
		return nil, err
	}
	return addrA0list, nil
}
func (md *MatrixDeposit002) DelA0list(contract *Contract, stateDB StateDBManager, addr common.Address) error {
	a0key := common.BytesToHash([]byte(KeyDepositA0list))
	ret := stateDB.GetStateByteArray(contract.CoinTyp, contract.Address(), a0key)
	var addrA0list []common.Address
	var retaddrA0list []common.Address
	err := rlp.DecodeBytes(ret, &addrA0list)
	if err != nil {
		log.Error("SetA0list", "err", err)
		return err
	}
	for i, a := range addrA0list {
		if a.Equal(addr) {
			retaddrA0list = append(addrA0list[:i], addrA0list[i+1:]...)
		}
	}
	if len(retaddrA0list) <= 0 {
		str := "Delete A0 address err,not find " + addr.Hex()
		return errors.New(str)
	}
	ba0, _ := rlp.EncodeToBytes(retaddrA0list)
	stateDB.SetStateByteArray(contract.CoinTyp, contract.Address(), a0key, ba0)
	return nil
}
func (md *MatrixDeposit002) getDepositListNum(contract *Contract, stateDB StateDBManager) *big.Int {
	numkey := common.BytesToHash([]byte(KeyDepositNum))
	num_b := stateDB.GetState(contract.CoinTyp, contract.Address(), numkey)
	if len(num_b) <= 0 {
		return big.NewInt(0)
	}
	return num_b.Big()
}

func (md *MatrixDeposit002) setDepositListNum(contract *Contract, stateDB StateDBManager, num *big.Int) {
	numkey := common.BytesToHash([]byte(KeyDepositNum))
	//bnum, _ := rlp.EncodeToBytes(num)
	bytenum := common.BigToHash(num)
	stateDB.SetState(contract.CoinTyp, contract.Address(), numkey, bytenum)
}
func (md *MatrixDeposit002) deposit_v2(in []byte, contract *Contract, evm *EVM, threshold *big.Int, depositRole *big.Int) ([]byte, error) {
	if len(in) < 4 {
		return nil, errParameters
	}
	data, err := depositAbi_v2.Methods["valiDeposit"].Inputs.UnpackValues(in)
	if err != nil {
		return nil, errDeposit
	}
	addr := data[0].(common.Address)
	depositType := data[1].(*big.Int)
	depositInfo := md.GetDepositBase(contract, evm.StateDB, contract.CallerAddress)
	totalAmount := big.NewInt(0)
	if depositInfo != nil {
		for _, depositinfo := range depositInfo.Dpstmsg {
			totalAmount.Add(totalAmount, depositinfo.DepositAmount)
		}
	}
	totalAmount.Add(totalAmount, contract.value)
	if totalAmount.Cmp(threshold) < 0 {
		return nil, errDeposit
	}
	depocfg := depositcfg.GetDepositCfg(depositcfg.VersionA).GetDepositPositionCfg(depositType.Uint64())
	if depocfg == nil {
		return nil, errors.New("get depositype err")
	}
	ischeck, err := depocfg.CheckAmountDeposit(addr, depositInfo, contract.value)
	if !ischeck {
		return nil, err
	}

	var addressA1 common.Address
	copy(addressA1[:], addr[:])
	err = md.modifyDepositState_v2(contract, depositType.Uint64(), evm, addressA1, depositRole, depositInfo)
	if err != nil {
		return nil, err
	}

	return []byte{1}, nil
}
func (md *MatrixDeposit002) valiDeposit_v2(in []byte, contract *Contract, evm *EVM) ([]byte, error) {
	return md.deposit_v2(in, contract, evm, validatorThreshold, big.NewInt(common.RoleValidator))
}

func (md *MatrixDeposit002) minerDeposit_v2(in []byte, contract *Contract, evm *EVM) ([]byte, error) {
	return md.deposit_v2(in, contract, evm, minerThreshold, big.NewInt(common.RoleMiner))
}

func (md *MatrixDeposit002) modifyWithdrawState_v2(contract *Contract, depositNum uint64, withdrawAmount *big.Int, evm *EVM) error {
	deposit := md.GetDepositBase(contract, evm.StateDB, contract.CallerAddress)
	if deposit == nil {
		str := "modifyWithdrawState_v2 err,deposit not exist.A0 addr = " + contract.CallerAddress.Hex()
		return errors.New(str)
	}
	totalAmount := big.NewInt(0)
	for _, info := range deposit.Dpstmsg {
		totalAmount.Add(totalAmount, info.DepositAmount)
	}
	if totalAmount.Sign() == 0 {
		return errDeposit
	}

	iscontinue := false
	for i, dpb := range deposit.Dpstmsg {
		if dpb.Position == depositNum {
			depositNum = uint64(i)
			iscontinue = true
			break
		}
	}
	if !iscontinue {
		return errOverflow
	}
	deptcfg := depositcfg.GetDepositCfg(depositcfg.VersionA).GetDepositPositionCfg(deposit.Dpstmsg[depositNum].DepositType)
	if deptcfg == nil {
		return errors.New("get depositype err")
	}
	ischeck, err := deptcfg.CheckwithdrawDeposit(depositNum, deposit, withdrawAmount)
	if ischeck {
		deptcfg.CalcDepositTime(depositNum, deposit, withdrawAmount, evm.Time.Uint64())
	} else {
		return err
	}
	return md.SetDepositBase(contract, evm.StateDB, contract.CallerAddress, deposit)
}

func (md *MatrixDeposit002) withdraw_v2(in []byte, contract *Contract, evm *EVM) ([]byte, error) {
	depositNum := big.NewInt(0)
	withdrawAmount := big.NewInt(0)
	err := depositAbi_v2.Methods["withdraw"].Inputs.Unpack(&[]interface{}{&depositNum, &withdrawAmount}, in[:])
	if err != nil {
		return nil, errDeposit
	}

	err = md.modifyWithdrawState_v2(contract, depositNum.Uint64(), withdrawAmount, evm)
	if err != nil {
		return nil, err
	}

	return []byte{1}, nil
}

func (md *MatrixDeposit002) modifyRefundState_v2(contract *Contract, depositNum uint64, evm *EVM) (*big.Int, error) {
	deposit := md.GetDepositBase(contract, evm.StateDB, contract.CallerAddress)
	if deposit == nil {
		str := "modifyRefundState_v2 err,deposit not exist.A0 addr = " + contract.CallerAddress.Hex()
		return big.NewInt(0), errors.New(str)
	}
	iscontinue := false
	for i, dpb := range deposit.Dpstmsg {
		if dpb.Position == depositNum {
			depositNum = uint64(i)
			iscontinue = true
			break
		}
	}
	if !iscontinue {
		return nil, errOverflow
	}
	depocfg := depositcfg.GetDepositCfg(depositcfg.VersionA).GetDepositPositionCfg(deposit.Dpstmsg[depositNum].DepositType)
	if depocfg == nil {
		return nil, errors.New("get depositype err")
	}
	ischeck, retval, err := depocfg.CheckAndcalcrefundDeposit(depositNum, deposit, evm.Time.Uint64())
	if !ischeck {
		return big.NewInt(0), err
	}
	//获取当前抵押身份
	depositRole := deposit.Role
	remainAmount := big.NewInt(0)
	for _, dept := range deposit.Dpstmsg {
		remainAmount.Add(remainAmount, dept.DepositAmount)
		for _, with := range dept.WithDrawInfolist {
			remainAmount.Add(remainAmount, with.WithDrawAmount)
		}
	}

	if depositRole.Cmp(big.NewInt(common.RoleMiner)) == 0 {
		if remainAmount.Cmp(minerThreshold) < 0 {
			//md.SetDepositRole(contract, evm.StateDB, big.NewInt(0)) //清除身份
			//md.clearAddress(contract, evm.StateDB, common.Address{})
			md.ResetSlash(contract, evm.StateDB, contract.CallerAddress)
			md.ResetInterest(contract, evm.StateDB, contract.CallerAddress)
			md.SetOnlineTime(contract, evm.StateDB, contract.CallerAddress, big.NewInt(0))
		}
	}

	if depositRole.Cmp(big.NewInt(common.RoleValidator)) == 0 {
		if remainAmount.Cmp(validatorThreshold) < 0 {
			//md.SetDepositRole(contract, evm.StateDB, big.NewInt(0)) //清除身份
			//md.clearAddress(contract, evm.StateDB, common.Address{})
			md.ResetSlash(contract, evm.StateDB, contract.CallerAddress)
			md.ResetInterest(contract, evm.StateDB, contract.CallerAddress)
			md.SetOnlineTime(contract, evm.StateDB, contract.CallerAddress, big.NewInt(0))
		}
	}
	//如果抵押金额为0则删除抵押信息
	if remainAmount.Sign() <= 0 {
		//md.removeDepositList(contract, evm.StateDB)
		md.SetDepositRole(contract, evm.StateDB, big.NewInt(0)) //清除身份
		num := md.getDepositListNum(contract, evm.StateDB)
		if num != nil {
			num.Sub(num, big.NewInt(1))
		} else {
			num = big.NewInt(0)
		}
		md.clearAddress(contract, evm.StateDB, deposit.AddressA1)
		md.setDepositListNum(contract, evm.StateDB, num)
		md.DelA0list(contract, evm.StateDB, deposit.AddressA0)
		deposit = nil
	}
	md.SetDepositBase(contract, evm.StateDB, contract.CallerAddress, deposit)
	return retval, nil
}

// ResetSlash reset slash to zero with state db and address.
func (md *MatrixDeposit002) ResetSlash(contract *Contract, db StateDBManager, address common.Address) error {
	ret := md.GetSlash(contract, db, address)
	for i, _ := range ret.CalcDeposit {
		ret.CalcDeposit[i].OperAmount = big.NewInt(0)
	}
	return md.SetSlash(contract, db, address, ret)
}
func (md *MatrixDeposit002) SetSlash(contract *Contract, stateDB StateDBManager, addr common.Address, slash common.CalculateDeposit) error {
	dpb := md.GetDepositBase(contract, stateDB, addr)
	if dpb == nil {
		log.Error("SetSlash ", "GetDepositBase err", "GetDepositBase is nil")
		return errors.New("get DepositBase is nil")
	}
	for _, val := range slash.CalcDeposit {
		for i, depo := range dpb.Dpstmsg {
			if val.Position == depo.Position {
				dpb.Dpstmsg[i].Slash = val.OperAmount
				break
			}

		}
	}
	return md.SetDepositBase(contract, stateDB, addr, dpb)
}
func (md *MatrixDeposit002) GetDepositBase(contract *Contract, stateDB StateDBManager, addr common.Address) *common.DepositBase {
	a0keystr := append(addr[:], 'A', '0')
	a0key := common.BytesToHash(a0keystr)
	ret := stateDB.GetStateByteArray(contract.CoinTyp, contract.Address(), a0key)
	if len(ret) == 0 {
		log.Info("GetDepositBase", "not find DepositBase", addr.String())
		return nil
	}
	var dpb common.DepositBase
	err := rlp.DecodeBytes(ret, &dpb)
	if err != nil {
		log.Error("GetDepositBase ", "decode err", err)
		return nil
	}
	return &dpb
}
func (md *MatrixDeposit002) SetDepositBase(contract *Contract, stateDB StateDBManager, addr common.Address, dpb *common.DepositBase) error {
	a1keystr := append(addr[:], 'A', '0')
	a1key := common.BytesToHash(a1keystr)
	var b []byte = []byte{}
	var err error
	if dpb != nil {
		b, err = rlp.EncodeToBytes(dpb)
		if err != nil {
			log.Error("SetDepositBase", "Encode err", err)
			return err
		}
	}
	stateDB.SetStateByteArray(contract.CoinTyp, contract.Address(), a1key, b)
	return nil
}
func (md *MatrixDeposit002) SetDepositRole(contract *Contract, stateDB StateDBManager, role *big.Int) error {
	retrolelist := md.GetDepositRole(contract, stateDB)
	for i, r := range retrolelist {
		var addrlist []common.Address
		for j, addr := range retrolelist[i].Address {
			if addr.Equal(contract.CallerAddress) {
				addrlist = append(addrlist, retrolelist[i].Address[:j]...)
				if len(retrolelist[i].Address) > j {
					addrlist = append(addrlist, retrolelist[i].Address[j+1:]...)
				}
			}
		}
		if len(addrlist) > 0 {
			retrolelist[i].Address = addrlist
		}
		if r.Role.Cmp(role) == 0 {
			retrolelist[i].Address = append(retrolelist[i].Address, contract.CallerAddress)
		}
	}
	key := common.BytesToHash([]byte(KeyDepositRole))
	b, err := rlp.EncodeToBytes(retrolelist)
	if err != nil {
		log.Error("SetDepositBase", "Encode err", err)
		return err
	}
	stateDB.SetStateByteArray(contract.CoinTyp, contract.Address(), key, b)
	return nil
}
func (md *MatrixDeposit002) GetDepositRole(contract *Contract, stateDB StateDBManager) (dproles []common.DepositRoles) {
	key := common.BytesToHash([]byte(KeyDepositRole))
	ret := stateDB.GetStateByteArray(contract.CoinTyp, contract.Address(), key)
	err := rlp.DecodeBytes(ret, &dproles)
	if err != nil {
		log.Error("GetDepositRole", "err", err)
		return nil
	}
	return
}
func (md *MatrixDeposit002) setAddress(contract *Contract, stateDB StateDBManager, addressA1 common.Address, dpb *common.DepositBase) error {
	if (addressA1 == common.Address{}) {
		return nil
	}
	a0addr := md.GetDepositAccount(contract, stateDB, addressA1)
	if a0addr != (common.Address{}) {
		if !a0addr.Equal(dpb.AddressA0) {
			return errors.New("different A0 address deposit A1 address")
		}
	}
	if !addressA1.Equal(dpb.AddressA1) {
		//如果当前的A1和之前的A1 不一致但是A0是一样的，那么就将原来的A1替换掉 	 (暂时不考虑替换)
		md.clearAddress(contract, stateDB, dpb.AddressA1)
		dpb.AddressA1 = addressA1
		//return errors.New("A0 address or A1 address Already mortgaged")
	}
	a1keystr := append(addressA1[:], 'A', '1')
	a1key := common.BytesToHash(a1keystr)
	ba0 := dpb.AddressA0
	stateDB.SetStateByteArray(contract.CoinTyp, contract.Address(), a1key, ba0[:])
	return nil
}
func (md *MatrixDeposit002) clearAddress(contract *Contract, stateDB StateDBManager, address common.Address) error {
	if (address == common.Address{}) {
		return errClear
	}
	a1keystr := append(address[:], 'A', '1')
	a1key := common.BytesToHash(a1keystr)

	stateDB.SetStateByteArray(contract.CoinTyp, contract.Address(), a1key, []byte{})
	return nil
}

func (md *MatrixDeposit002) refund_v2(in []byte, contract *Contract, evm *EVM) ([]byte, error) {
	depositNum := big.NewInt(0)
	err := depositAbi_v2.Methods["refund"].Inputs.Unpack(&depositNum, in[:])
	if err != nil {
		return nil, errDeposit
	}

	value, err := md.modifyRefundState_v2(contract, depositNum.Uint64(), evm)
	if err != nil {
		return nil, err
	}
	if !evm.CanTransfer(evm.StateDB, contract.Address(), value, evm.Cointyp) {
		return nil, ErrInsufficientBalance
	}
	evm.Transfer(evm.StateDB, contract.Address(), contract.CallerAddress, value, evm.Cointyp)
	return []byte{1}, nil
}

func (md *MatrixDeposit002) SetOnlineTime(contract *Contract, stateDB StateDBManager, address common.Address, tm *big.Int) error {
	dposit := md.GetDepositBase(contract, stateDB, address)
	if dposit == nil {
		err := "SetOnlineTime" + "get deposit err,deposit is nil"
		return errors.New(err)
	}
	//ot := new(big.Int).Add(dposit.OnlineTime, tm)
	dposit.OnlineTime = tm
	md.SetDepositBase(contract, stateDB, address, dposit)
	return nil
}
func (md *MatrixDeposit002) GetOnlineTime(contract *Contract, stateDB StateDBManager, addr common.Address) *big.Int {
	dposit := md.GetDepositBase(contract, stateDB, addr)
	if dposit == nil {
		log.Error("GetOnlineTime", "get deposit err", "deposit is nil")
		return big.NewInt(0)
	}
	return dposit.OnlineTime
}

func (md *MatrixDeposit002) GetDepositAccount(contract *Contract, stateDB StateDBManager, authAccount common.Address) common.Address {
	a1keystr := append(authAccount[:], 'A', '1')
	a1key := common.BytesToHash(a1keystr)
	addrA0 := stateDB.GetStateByteArray(contract.CoinTyp, contract.Address(), a1key)
	if len(addrA0) == 0 {
		return common.Address{}
	}
	return common.BytesToAddress(addrA0)
}
func (md *MatrixDeposit002) GetAuthAccount(contract *Contract, stateDB StateDBManager, depositAccount common.Address) common.Address {
	ret := md.GetDepositBase(contract, stateDB, depositAccount)
	if ret == nil {
		return common.Address{}
	}
	return ret.AddressA1
}
func (md *MatrixDeposit002) PayInterest(contract *Contract, stateDB StateDBManager, addrA0 common.Address, position uint64, amount *big.Int) error {
	ret := md.GetDepositBase(contract, stateDB, addrA0)
	if ret == nil {
		str := "PayInterest err,deposit not exist.A0 addr = " + addrA0.Hex()
		return errors.New(str)
	}
	for i, retinfo := range ret.Dpstmsg {
		if retinfo.Position == position {
			ret.Dpstmsg[i].DepositAmount.Add(ret.Dpstmsg[i].DepositAmount, amount)
			ret.Dpstmsg[i].Interest = big.NewInt(0)
			ret.Dpstmsg[i].Slash = big.NewInt(0)
			break
		}
	}

	return md.SetDepositBase(contract, stateDB, addrA0, ret)
}
