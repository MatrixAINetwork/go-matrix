package support

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/log"
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
		CapitalMap = append(CapitalMap, Pnormalized{Addr: self.Address, Value: float64(value)})
	}
	return CapitalMap
}
