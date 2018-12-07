// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package layered

import (
	"sort"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/vm"
	"github.com/matrix/go-matrix/election/support"
	"math/big"
	"math/rand"
)

const (
	DefauleStock = 1
)

var (
	man  = new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	vip1 = new(big.Int).Mul(big.NewInt(100000), man)
	vip2 = new(big.Int).Mul(big.NewInt(40000), man)
)

func FindIndex(ans *big.Int) int {

	for k, v := range common.EchelonArrary {
		//fmt.Println("v.MinMoney",v.MinMoney,"ans",ans)
		if ans.Cmp(v.MinMoney) >= 0 {
			return k
		}
	}
	return -1
}
func CalEchelonNum(vmData []vm.DepositDetail) [][]vm.DepositDetail {
	ans := [][]vm.DepositDetail{}
	for index := 0; index < len(common.EchelonArrary); index++ {
		ans = append(ans, []vm.DepositDetail{})
	}

	for _, v := range vmData {
		//fmt.Println("56666666","big.Nit",v.Deposit.String(),"Uint64",v.Deposit.Uint64())
		index := FindIndex(v.Deposit)
		//	fmt.Println("index", index, "deposit", v.Deposit.Uint64())
		if index == -1 {
			continue
		}
		ans[index] = append(ans[index], v)
	}
	return ans
}

func GetRatio(aim uint64) float64 {
	switch {
	case aim >= 10000000:
		return 2.0
	case aim >= 1000000 && aim < 10000000:
		return 1.0
	default:
		return 0.5
	}
}
func GetValueByDeposit(vm []vm.DepositDetail) []support.Stf {
	value := []support.Stf{}
	for _, v := range vm {
		temp := support.Stf{
			Str: v.NodeID.String(),
		}

		temp.Flot = GetRatio(v.Deposit.Uint64())
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

func Knuth_Fisher_Yates_Algorithm(vm []vm.DepositDetail, randSeed *big.Int) []vm.DepositDetail {
	//高纳德置乱算法
	rand.Seed(randSeed.Int64())
	for index := len(vm) - 1; index > 0; index-- {
		aimIndex := rand.Intn(index + 1)
		t := vm[index]
		vm[index] = vm[aimIndex]
		vm[aimIndex] = t
	}
	return vm
}
func sortByDepositAndUptime(vm []vm.DepositDetail, random *big.Int) []vm.DepositDetail {

	vm = Knuth_Fisher_Yates_Algorithm(vm, random)

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
