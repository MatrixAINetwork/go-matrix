// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package support

import (
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/log"
	"math"
	"math/big"
)

type SelfNodeInfo struct {
	Address common.Address
	Stk     *big.Int
	Uptime  uint64
	Tps     uint64
}

func (self *SelfNodeInfo) TPSPowerStake() float64 {
	for _, v := range DefaultTpsRatio {
		if self.Tps >= v.MinNum {
			return v.Ratio
		}
	}
	log.Error(ModuleLogName, "tps参数配置有误,请调整,最后一个配置的MinNum需为0", DefaultTpsRatio)
	return 0
}

func (self *SelfNodeInfo) OnlineTimeStake() float64 {
	for _, v := range DefaultOnlineTimeRatio {
		if self.Uptime >= v.MinNum {
			return v.Ratio
		}
	}
	log.Error(ModuleLogName, "onlinetime参数配置有误,请调整,最后一个配置的MinNum需为0", DefaultOnlineTimeRatio)
	return 0
}

func (self *SelfNodeInfo) DepositStake(roles common.RoleType) float64 {
	temp := big.NewInt(0).Set(self.Stk)
	deposMan := temp.Div(temp, common.ManValue).Uint64()

	switch roles {
	case common.RoleMiner:
		return float64(deposMan / DefaultMinerDeposit)
	default:
		return float64(deposMan / DefaultValidatorDeposit)
	}

}

func getDepositList(types common.RoleType) []RatioList {
	ratio := []RatioList{}
	switch types {
	case common.RoleValidator:
		ratio = append(ratio, DefaultValidatorDepositRatio...)
	case common.RoleMiner:
		ratio = append(ratio, DefaultMinerDepositRatio...)
	}
	return ratio
}

func CalcValue(nodes []Node, role common.RoleType) []Pnormalized {
	var CapitalMap []Pnormalized
	for _, item := range nodes {
		self := SelfNodeInfo{Address: item.Address, Stk: item.Deposit, Uptime: item.OnlineTime.Uint64(), Tps: DefaultTps}
		//    a*A(b*B+c*C)+aa*A+cc*C
		// = a*b*A*B + a*c*A*C + aa*A + cc*C
		value := DefaultQuantificationRatio.Multi_Online * DefaultQuantificationRatio.Multi_Tps * self.OnlineTimeStake() * self.TPSPowerStake()
		value += DefaultQuantificationRatio.Multi_Online * DefaultQuantificationRatio.Multi_Deposit * self.OnlineTimeStake() * self.DepositStake(role)
		value += DefaultQuantificationRatio.Add_Online * self.OnlineTimeStake()
		value += DefaultQuantificationRatio.Add_Deposit * self.DepositStake(role)
		value *= (float64(item.Ratio) / float64(DefaultRatioDenominator))
		//保护价值函数值
		if 0 == value {
			value = 1
		}
		CapitalMap = append(CapitalMap, Pnormalized{Addr: self.Address, Value: float64(value)})
	}
	return CapitalMap
}

func CalcValueEW(nodes []Node, stockExp float64) []Pnormalized {
	var CapitalMap []Pnormalized
	for _, item := range nodes {
		self := SelfNodeInfo{Address: item.Address, Stk: item.Deposit, Uptime: item.OnlineTime.Uint64(), Tps: DefaultTps}
		val := big.NewInt(0).Div(item.Deposit, big.NewInt(1e18)).Int64()
		if val <= 0 {
			val = 1
		}
		value := math.Pow(float64(val), stockExp)
		CapitalMap = append(CapitalMap, Pnormalized{Addr: self.Address, Value: float64(value)})
	}
	return CapitalMap
}
