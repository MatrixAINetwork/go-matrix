// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package depositcfg

import (
	"errors"
	"math/big"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/log"
)

type depositCfgA struct {
	DepositMap map[uint64]DepositCfger
}

type DepositPositionConfig struct {
	DepositType          uint64
	Tmduration           uint64   //时长(活期是7天，定期是按月)
	CruWithDrawAmountMin *big.Int //参退选最小金额
	DepositRate          uint64   //利率
}
type Depositcurrent struct {
	DepositCur DepositPositionConfig
}

type Depositregular struct {
	Depositreg DepositPositionConfig
}

type DepositCfger interface {
	CheckwithdrawDeposit(index uint64, deposit *common.DepositBase, wdAm *big.Int) (bool, error)
	CheckAndcalcrefundDeposit(index uint64, deposit *common.DepositBase, t uint64) (bool, *big.Int, error)
	CheckAmountDeposit(addr common.Address, deposit *common.DepositBase, wdAm *big.Int) (bool, error)
	CalcDepositTime(index uint64, deposit *common.DepositBase, wdAm *big.Int, t uint64) error
	GetRate() uint64
}

const (
	CurrentDeposit  = uint64(0)
	MONTH_1         = uint64(1)
	MONTH_3         = uint64(3)
	MONTH_6         = uint64(6)
	MONTH_12        = uint64(12)
	SecondsPerMonth = 30 * 24 * 60 * 60 //每月秒数
	Days7Seconds    = 7 * 24 * 60 * 60
	Delay           = 60 * 60 * 2 //2小时
)

var (
	man                  = new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	FixDepositAmountMin  = new(big.Int).Mul(big.NewInt(2000), man) //定期抵押最小限额
	CurDepositAmountMin  = new(big.Int).Mul(big.NewInt(100), man)  //活期抵押最小限额
	CruWithDrawAmountMin = new(big.Int).Mul(big.NewInt(100), man)  //活期退选最小限额
	errDeposit           = errors.New("deposit is not found")
)

func newDepositCfgA() *depositCfgA {
	d := &depositCfgA{DepositMap: make(map[uint64]DepositCfger)}
	d.DepositMap[CurrentDeposit] = &Depositcurrent{DepositCur: DepositPositionConfig{DepositType: CurrentDeposit, Tmduration: Days7Seconds, CruWithDrawAmountMin: CurDepositAmountMin, DepositRate: 1}}
	d.DepositMap[MONTH_1] = &Depositregular{Depositreg: DepositPositionConfig{DepositType: MONTH_1, Tmduration: SecondsPerMonth, CruWithDrawAmountMin: FixDepositAmountMin, DepositRate: 2}}
	d.DepositMap[MONTH_3] = &Depositregular{Depositreg: DepositPositionConfig{DepositType: MONTH_3, Tmduration: SecondsPerMonth * MONTH_3, CruWithDrawAmountMin: FixDepositAmountMin, DepositRate: 3}}
	d.DepositMap[MONTH_6] = &Depositregular{Depositreg: DepositPositionConfig{DepositType: MONTH_6, Tmduration: SecondsPerMonth * MONTH_6, CruWithDrawAmountMin: FixDepositAmountMin, DepositRate: 4}}
	d.DepositMap[MONTH_12] = &Depositregular{Depositreg: DepositPositionConfig{DepositType: MONTH_12, Tmduration: SecondsPerMonth * MONTH_12, CruWithDrawAmountMin: FixDepositAmountMin, DepositRate: 6}}
	return d
}

func (d *depositCfgA) GetDepositPositionCfg(depositType uint64) DepositCfger {
	if cfg, exit := d.DepositMap[depositType]; exit {
		return cfg
	}
	return nil
}
func (cur *Depositcurrent) CheckwithdrawDeposit(index uint64, deposit *common.DepositBase, wdAm *big.Int) (bool, error) {
	depositmsg := deposit.Dpstmsg[index]
	if cur.DepositCur.DepositType == depositmsg.DepositType {
		if depositmsg.DepositAmount.Cmp(wdAm) >= 0 && wdAm.Cmp(cur.DepositCur.CruWithDrawAmountMin) >= 0 {
			return true, nil
		} else {
			log.ERROR("CheckwithdrawDeposit", "deposit amount", depositmsg.DepositAmount, "wdAm", wdAm)
			return false, errors.New("CheckwithdrawDeposit deposit Amount insufficient or withdraw amount too less")
		}
	} else {
		return false, errors.New("CheckwithdrawDeposit deposit information Mismatch")
	}
}
func (cur *Depositcurrent) CalcDepositTime(index uint64, deposit *common.DepositBase, wdAm *big.Int, t uint64) error {
	endtime := t + cur.DepositCur.Tmduration
	deposit.Dpstmsg[index].WithDrawInfolist = append(deposit.Dpstmsg[index].WithDrawInfolist, common.WithDrawInfo{WithDrawAmount: wdAm, WithDrawTime: endtime})
	deposit.Dpstmsg[index].DepositAmount.Sub(deposit.Dpstmsg[index].DepositAmount, wdAm)
	if deposit.Dpstmsg[index].DepositAmount.Sign() == 0 {
		deposit.Dpstmsg[index].Interest = big.NewInt(0)
		deposit.Dpstmsg[index].Slash = big.NewInt(0)
	}
	return nil
}
func (cur *Depositcurrent) CheckAndcalcrefundDeposit(index uint64, deposit *common.DepositBase, t uint64) (bool, *big.Int, error) {
	var wdrawlist []common.WithDrawInfo
	retAmonut := big.NewInt(0)
	for i, zeroDeposit := range deposit.Dpstmsg[index].WithDrawInfolist {
		if zeroDeposit.WithDrawTime <= t {
			retAmonut = new(big.Int).Add(retAmonut, zeroDeposit.WithDrawAmount)
		} else {
			wdrawlist = append(wdrawlist, deposit.Dpstmsg[index].WithDrawInfolist[i])
		}
	}
	deposit.Dpstmsg[index].WithDrawInfolist = make([]common.WithDrawInfo, len(wdrawlist))
	copy(deposit.Dpstmsg[index].WithDrawInfolist, wdrawlist)
	return true, retAmonut, nil
}
func (cur *Depositcurrent) CheckAmountDeposit(addr common.Address, deposit *common.DepositBase, wdAm *big.Int) (bool, error) {
	if wdAm.Cmp(cur.DepositCur.CruWithDrawAmountMin) < 0 && wdAm.Sign() != 0 {
		return false, errDeposit
	}
	////活期抵押可填0，代表修改A1账户 (在设置账户时修改A1,不在这设置了)
	//if wdAm.Sign() == 0 {
	//	deposit.AddressA1 = addr
	//}
	return true, nil
}
func (cur *Depositcurrent) GetRate() uint64 {
	return cur.DepositCur.DepositRate
}
func (drg *Depositregular) CheckwithdrawDeposit(index uint64, deposit *common.DepositBase, wdAm *big.Int) (bool, error) {
	depositmsg := deposit.Dpstmsg[index]
	if drg.Depositreg.DepositType == depositmsg.DepositType {
		if /*depositmsg.DepositAmount.Cmp(wdAm) == 0 && */ len(depositmsg.WithDrawInfolist) <= 0 {
			return true, nil
		} else {
			return false, errors.New("CheckwithdrawDeposit deposit Amount err or this Position has been withdrawn")
		}
	} else {
		return false, errors.New("CheckwithdrawDeposit deposit information Mismatch")
	}
}
func (drg *Depositregular) CheckAndcalcrefundDeposit(index uint64, deposit *common.DepositBase, t uint64) (bool, *big.Int, error) {
	isok := false
	for _, zeroDeposit := range deposit.Dpstmsg[index].WithDrawInfolist {
		if zeroDeposit.WithDrawTime <= t {
			isok = true
		}
	}
	err := errors.New("regular deposit Time has not arrived")
	retAmonut := big.NewInt(0)
	if isok {
		retAmonut = deposit.Dpstmsg[index].DepositAmount
		newdepositlist := deposit.Dpstmsg[:index]
		if uint64(len(deposit.Dpstmsg)-1) >= index+1 {
			newdepositlist = append(newdepositlist, deposit.Dpstmsg[index+1:]...)
		}
		deposit.Dpstmsg = newdepositlist
		err = nil
	}
	return isok, retAmonut, err
}
func (drg *Depositregular) CheckAmountDeposit(addr common.Address, deposit *common.DepositBase, wdAm *big.Int) (bool, error) {
	if wdAm.Cmp(drg.Depositreg.CruWithDrawAmountMin) < 0 {
		return false, errDeposit
	}
	return true, nil
}
func (drg *Depositregular) CalcDepositTime(index uint64, deposit *common.DepositBase, wdAm *big.Int, t uint64) error {
	endtime := (t-deposit.Dpstmsg[index].BeginTime)/drg.Depositreg.Tmduration*drg.Depositreg.Tmduration + drg.Depositreg.Tmduration + deposit.Dpstmsg[index].BeginTime
	deposit.Dpstmsg[index].EndTime = endtime
	deposit.Dpstmsg[index].WithDrawInfolist = append(deposit.Dpstmsg[index].WithDrawInfolist, common.WithDrawInfo{WithDrawAmount: big.NewInt(0), WithDrawTime: endtime + uint64(Delay)})
	return nil
}
func (dc *Depositregular) GetRate() uint64 {
	return dc.Depositreg.DepositRate
}
