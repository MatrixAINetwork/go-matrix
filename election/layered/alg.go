// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package layered

import (
	"sort"

	"github.com/matrix/go-matrix/core/vm"
	"github.com/matrix/go-matrix/election/support"
)

func CalEchelonNum(vmData []vm.DepositDetail) ([]vm.DepositDetail, []vm.DepositDetail) {
	FirstQuota := []vm.DepositDetail{}
	SecondQuota := []vm.DepositDetail{}
	for _, v := range vmData {
		if v.Deposit.Uint64() >= FirstEchelon.MinMoney {
			FirstQuota = append(FirstQuota, v)
		} else if v.Deposit.Uint64() >= SecondEchelon.MinMoney {
			SecondQuota = append(SecondQuota, v)
		}
	}
	return FirstQuota, SecondQuota
}

func GetValueByDeposit(vm []vm.DepositDetail) []support.Stf {
	value := []support.Stf{}
	for _, v := range vm {
		temp := support.Stf{
			Str: v.NodeID.String(),
		}

		if v.Deposit.Uint64() >= 1000*1000 {
			temp.Flot = 1.5
		} else if v.Deposit.Uint64() >= 100*1000 {
			temp.Flot = 1.4
		} else if v.Deposit.Uint64() >= 10*1000 {
			temp.Flot = 1.3
		} else if v.Deposit.Uint64() >= 1*1000 {
			temp.Flot = 1.2
		} else if v.Deposit.Uint64() >= 0.1*1000 {
			temp.Flot = 1.1
		} else {
			temp.Flot = 1.0
		}
		value = append(value, temp)
	}
	return value
}

type VMS []vm.DepositDetail

func (self VMS) Len() int {
	return len(self)
}
func (self VMS) Less(i, j int) bool {
	if self[i].Deposit.Uint64() == self[j].Deposit.Uint64() {
		return self[i].OnlineTime.Uint64() > self[j].OnlineTime.Uint64()
	}
	return self[i].Deposit.Uint64() > self[j].Deposit.Uint64()
}
func (self VMS) Swap(i, j int) {

	self[i].Address, self[j].Address = self[j].Address, self[i].Address
	self[i].NodeID, self[j].NodeID = self[j].NodeID, self[i].NodeID
	self[i].Deposit, self[j].Deposit = self[j].Deposit, self[i].Deposit
	self[i].WithdrawH, self[j].WithdrawH = self[j].WithdrawH, self[i].WithdrawH
	self[i].OnlineTime, self[j].OnlineTime = self[j].OnlineTime, self[i].OnlineTime
}

func sortByDepositAndUptime(vm []vm.DepositDetail) []vm.DepositDetail {
	/*
		for _, v := range vm {
			fmt.Println("sort前", v.NodeID.String(), v.Address.String(), v.OnlineTime.String())
		}
	*/
	sort.Sort(VMS(vm))
	/*
		for _, v := range vm {
			fmt.Println("sort后", v.NodeID.String(), v.Address.String(), v.OnlineTime.String())
		}
	*/
	return vm
}

type Weight struct {
	NodeId string
	weight float64
}
