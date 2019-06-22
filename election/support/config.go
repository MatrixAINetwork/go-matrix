// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package support

const (
	ModuleLogName       = "选举基础模块"
	MaxSample           = 1000 //配置参数,采样最多发生1000次,是一个离P+M较远的值
	PowerWeightMaxSmple = 1000
	J                   = 0 //基金会验证节点个数tps_weight
	DefaultStock        = 1
)

const (
	DefaultNodeConfig       = 0
	MaxVipEleLevelNum       = 2
	DefaultRatio            = 1000
	DefaultRatioDenominator = 1000
	DefaultMinerDeposit     = 10000
	DefaultValidatorDeposit = 100000
)
const (
	ValidatorElectPlug_Direct = "Direct"
	ValidatorElectPlug_Order  = "Order"
)

type QuantRatio struct {
	Multi_Online  float64
	Multi_Tps     float64
	Multi_Deposit float64
	Add_Online    float64
	Add_Deposit   float64
}

var (
	DefaultTps                 = uint64(1000)
	DefalutValidatorElectPlug  = ValidatorElectPlug_Order //选举所要用到的插件     1.直接选11+5     2.依次选11+5
	DefaultVIPStock            = []int{3, 2, 1}           //默认股权能否配载vip列表里(创世文件)
	DefaultQuantificationRatio = QuantRatio{
		Multi_Online:  0,
		Multi_Tps:     0,
		Multi_Deposit: 0.0,
		Add_Online:    0,
		Add_Deposit:   1.0,
	}
	DefaultTpsRatio = []RatioList{
		RatioList{
			MinNum: 16000,
			Ratio:  5.0,
		},
		RatioList{
			MinNum: 8000,
			Ratio:  4.0,
		},
		RatioList{
			MinNum: 4000,
			Ratio:  3.0,
		},
		RatioList{
			MinNum: 2000,
			Ratio:  2.0,
		},
		RatioList{
			MinNum: 1000,
			Ratio:  1.0,
		},
		RatioList{
			MinNum: 0,
			Ratio:  0.0,
		},
	}
	DefaultOnlineTimeRatio = []RatioList{
		RatioList{
			MinNum: 512,
			Ratio:  4.0,
		},
		RatioList{
			MinNum: 256,
			Ratio:  2.0,
		},
		RatioList{
			MinNum: 128,
			Ratio:  1.0,
		},
		RatioList{
			MinNum: 64,
			Ratio:  0.5,
		},
		RatioList{
			MinNum: 0,
			Ratio:  0.25,
		},
	}
	DefaultMinerDepositRatio = []RatioList{
		RatioList{
			MinNum: 50000,
			Ratio:  5.0,
		},
		RatioList{
			MinNum: 40000,
			Ratio:  4.0,
		},
		RatioList{
			MinNum: 30000,
			Ratio:  3.0,
		},
		RatioList{
			MinNum: 20000,
			Ratio:  2.0,
		},
		RatioList{
			MinNum: 0,
			Ratio:  0.0,
		},
	}
	DefaultValidatorDepositRatio = []RatioList{
		RatioList{
			MinNum: 800000,
			Ratio:  4.0,
		},
		RatioList{
			MinNum: 600000,
			Ratio:  3.0,
		},
		RatioList{
			MinNum: 400000,
			Ratio:  2.0,
		},
		RatioList{
			MinNum: 200000,
			Ratio:  1.0,
		},
		RatioList{
			MinNum: 0,
			Ratio:  0.0,
		},
	}
)
