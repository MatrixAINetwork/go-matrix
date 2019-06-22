// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package interest

import (
	"math/big"
	"testing"
	"time"

	"github.com/MatrixAINetwork/go-matrix/params/manparams"

	"github.com/MatrixAINetwork/go-matrix/params"

	"github.com/MatrixAINetwork/go-matrix/depoistInfo"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/matrixstate"
	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mandb"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/reward/depositcfg"
	"github.com/MatrixAINetwork/go-matrix/reward/util"
)

func Test_GetCurrent(t *testing.T) {
	log.InitLog(5)
	util.LogExtraDebug(PackageName, "test", "test")
	depositNodes := make([]common.DepositBase, 0)
	nodes0 := common.DepositBase{AddressA0: common.HexToAddress("01"), Dpstmsg: make([]common.DepositMsg, 0)}
	nodes0.Dpstmsg = append(nodes0.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes0.Dpstmsg = append(nodes0.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes0.Dpstmsg = append(nodes0.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes0.Dpstmsg = append(nodes0.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes0.Dpstmsg = append(nodes0.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	depositNodes = append(depositNodes, nodes0)
	nodes1 := common.DepositBase{AddressA0: common.HexToAddress("02"), Dpstmsg: make([]common.DepositMsg, 0)}
	nodes1.Dpstmsg = append(nodes1.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(100000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes1.Dpstmsg = append(nodes1.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(100000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes1.Dpstmsg = append(nodes1.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(100000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes1.Dpstmsg = append(nodes1.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(100000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes1.Dpstmsg = append(nodes1.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(100000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	depositNodes = append(depositNodes, nodes1)
	nodes2 := common.DepositBase{AddressA0: common.HexToAddress("03"), Dpstmsg: make([]common.DepositMsg, 0)}
	nodes2.Dpstmsg = append(nodes2.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(1000000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes2.Dpstmsg = append(nodes2.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(1000000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes2.Dpstmsg = append(nodes2.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(1000000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes2.Dpstmsg = append(nodes2.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(1000000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes2.Dpstmsg = append(nodes2.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(1000000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	depositNodes = append(depositNodes, nodes2)
	nodes3 := common.DepositBase{AddressA0: common.HexToAddress("04"), Dpstmsg: make([]common.DepositMsg, 0)}
	nodes3.Dpstmsg = append(nodes3.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(10000000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes3.Dpstmsg = append(nodes3.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(10000000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes3.Dpstmsg = append(nodes3.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(10000000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes3.Dpstmsg = append(nodes3.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(10000000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes3.Dpstmsg = append(nodes3.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(10000000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	depositNodes = append(depositNodes, nodes3)

	nodes4 := common.DepositBase{AddressA0: common.HexToAddress("05"), Dpstmsg: make([]common.DepositMsg, 0)}
	nodes4.Dpstmsg = append(nodes4.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(100), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes4.Dpstmsg = append(nodes4.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(100), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes4.Dpstmsg = append(nodes4.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(100), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes4.Dpstmsg = append(nodes4.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(100), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes4.Dpstmsg = append(nodes4.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(100), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	depositNodes = append(depositNodes, nodes4)
	ic := &interestDelta{
		interestConfig: &mc.InterestCfg{PayInterval: 3600, AttenuationRate: 8500, AttenuationPeriod: 3000000, RewardMount: 3200},
		depositCfg:     depositcfg.GetDepositCfg(depositcfg.VersionA),
	}

	ic.GetReward(new(big.Int).SetUint64(32e17), depositNodes)
	//if !reflect.DeepEqual(got, tt.want) {
	//	t.Errorf("interestDelta.GetWeightDeposit() got = %v, want %v", got, tt.want)
	//}

}

func Test_GetRegular(t *testing.T) {
	log.InitLog(5)

	depositNodes := make([]common.DepositBase, 0)
	nodes0 := common.DepositBase{AddressA0: common.HexToAddress("01"), Dpstmsg: make([]common.DepositMsg, 0)}
	nodes0.Dpstmsg = append(nodes0.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(100), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes0.Dpstmsg = append(nodes0.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_1, DepositAmount: new(big.Int).Mul(big.NewInt(100), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes0.Dpstmsg = append(nodes0.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_3, DepositAmount: new(big.Int).Mul(big.NewInt(100), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes0.Dpstmsg = append(nodes0.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_6, DepositAmount: new(big.Int).Mul(big.NewInt(100), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes0.Dpstmsg = append(nodes0.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_9, DepositAmount: new(big.Int).Mul(big.NewInt(100), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	depositNodes = append(depositNodes, nodes0)
	nodes1 := common.DepositBase{AddressA0: common.HexToAddress("02"), Dpstmsg: make([]common.DepositMsg, 0)}
	nodes1.Dpstmsg = append(nodes1.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(200), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes1.Dpstmsg = append(nodes1.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_1, DepositAmount: new(big.Int).Mul(big.NewInt(200), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes1.Dpstmsg = append(nodes1.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_3, DepositAmount: new(big.Int).Mul(big.NewInt(200), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes1.Dpstmsg = append(nodes1.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_6, DepositAmount: new(big.Int).Mul(big.NewInt(200), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes1.Dpstmsg = append(nodes1.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_9, DepositAmount: new(big.Int).Mul(big.NewInt(200), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	depositNodes = append(depositNodes, nodes1)
	nodes2 := common.DepositBase{AddressA0: common.HexToAddress("03"), Dpstmsg: make([]common.DepositMsg, 0)}
	nodes2.Dpstmsg = append(nodes2.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(300), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes2.Dpstmsg = append(nodes2.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_1, DepositAmount: new(big.Int).Mul(big.NewInt(300), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes2.Dpstmsg = append(nodes2.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_3, DepositAmount: new(big.Int).Mul(big.NewInt(300), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes2.Dpstmsg = append(nodes2.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_6, DepositAmount: new(big.Int).Mul(big.NewInt(300), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes2.Dpstmsg = append(nodes2.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_9, DepositAmount: new(big.Int).Mul(big.NewInt(300), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	depositNodes = append(depositNodes, nodes2)
	nodes3 := common.DepositBase{AddressA0: common.HexToAddress("04"), Dpstmsg: make([]common.DepositMsg, 0)}
	nodes3.Dpstmsg = append(nodes3.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(400), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes3.Dpstmsg = append(nodes3.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_1, DepositAmount: new(big.Int).Mul(big.NewInt(400), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes3.Dpstmsg = append(nodes3.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_3, DepositAmount: new(big.Int).Mul(big.NewInt(400), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes3.Dpstmsg = append(nodes3.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_6, DepositAmount: new(big.Int).Mul(big.NewInt(400), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes3.Dpstmsg = append(nodes3.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_9, DepositAmount: new(big.Int).Mul(big.NewInt(400), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	depositNodes = append(depositNodes, nodes3)

	nodes4 := common.DepositBase{AddressA0: common.HexToAddress("05"), Dpstmsg: make([]common.DepositMsg, 0)}
	nodes4.Dpstmsg = append(nodes4.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(500), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes4.Dpstmsg = append(nodes4.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_1, DepositAmount: new(big.Int).Mul(big.NewInt(500), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes4.Dpstmsg = append(nodes4.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_3, DepositAmount: new(big.Int).Mul(big.NewInt(500), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes4.Dpstmsg = append(nodes4.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_6, DepositAmount: new(big.Int).Mul(big.NewInt(500), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes4.Dpstmsg = append(nodes4.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_9, DepositAmount: new(big.Int).Mul(big.NewInt(500), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	depositNodes = append(depositNodes, nodes4)
	ic := &interestDelta{
		interestConfig: &mc.InterestCfg{PayInterval: 3600, AttenuationRate: 8500, AttenuationPeriod: 3000000, RewardMount: 3200},
		depositCfg:     depositcfg.GetDepositCfg(depositcfg.VersionA),
	}

	ic.GetReward(new(big.Int).SetUint64(32e17), depositNodes)
	//if !reflect.DeepEqual(got, tt.want) {
	//	t.Errorf("interestDelta.GetWeightDeposit() got = %v, want %v", got, tt.want)
	//}

}

func Test_interestDelta_GetWeightDeposit(t *testing.T) {
	log.InitLog(5)

	depositNodes := make([]common.DepositBase, 0)
	nodes0 := common.DepositBase{AddressA0: common.HexToAddress("01"), Dpstmsg: make([]common.DepositMsg, 0)}
	nodes0.Dpstmsg = append(nodes0.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes0.Dpstmsg = append(nodes0.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_1, DepositAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes0.Dpstmsg = append(nodes0.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_3, DepositAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes0.Dpstmsg = append(nodes0.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_6, DepositAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes0.Dpstmsg = append(nodes0.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_9, DepositAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	depositNodes = append(depositNodes, nodes0)
	nodes1 := common.DepositBase{AddressA0: common.HexToAddress("02"), Dpstmsg: make([]common.DepositMsg, 0)}
	nodes1.Dpstmsg = append(nodes1.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(100000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes1.Dpstmsg = append(nodes1.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_1, DepositAmount: new(big.Int).Mul(big.NewInt(100000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes1.Dpstmsg = append(nodes1.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_3, DepositAmount: new(big.Int).Mul(big.NewInt(100000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes1.Dpstmsg = append(nodes1.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_6, DepositAmount: new(big.Int).Mul(big.NewInt(100000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes1.Dpstmsg = append(nodes1.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_9, DepositAmount: new(big.Int).Mul(big.NewInt(100000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	depositNodes = append(depositNodes, nodes1)
	nodes2 := common.DepositBase{AddressA0: common.HexToAddress("03"), Dpstmsg: make([]common.DepositMsg, 0)}
	nodes2.Dpstmsg = append(nodes2.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(1000000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes2.Dpstmsg = append(nodes2.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_1, DepositAmount: new(big.Int).Mul(big.NewInt(1000000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes2.Dpstmsg = append(nodes2.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_3, DepositAmount: new(big.Int).Mul(big.NewInt(1000000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes2.Dpstmsg = append(nodes2.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_6, DepositAmount: new(big.Int).Mul(big.NewInt(1000000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes2.Dpstmsg = append(nodes2.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_9, DepositAmount: new(big.Int).Mul(big.NewInt(1000000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	depositNodes = append(depositNodes, nodes2)
	nodes3 := common.DepositBase{AddressA0: common.HexToAddress("04"), Dpstmsg: make([]common.DepositMsg, 0)}
	nodes3.Dpstmsg = append(nodes3.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(10000000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes3.Dpstmsg = append(nodes3.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_1, DepositAmount: new(big.Int).Mul(big.NewInt(10000000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes3.Dpstmsg = append(nodes3.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_3, DepositAmount: new(big.Int).Mul(big.NewInt(10000000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes3.Dpstmsg = append(nodes3.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_6, DepositAmount: new(big.Int).Mul(big.NewInt(10000000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes3.Dpstmsg = append(nodes3.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_9, DepositAmount: new(big.Int).Mul(big.NewInt(10000000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	depositNodes = append(depositNodes, nodes3)

	nodes4 := common.DepositBase{AddressA0: common.HexToAddress("05"), Dpstmsg: make([]common.DepositMsg, 0)}
	nodes4.Dpstmsg = append(nodes4.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(100), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes4.Dpstmsg = append(nodes4.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_1, DepositAmount: new(big.Int).Mul(big.NewInt(100), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes4.Dpstmsg = append(nodes4.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_3, DepositAmount: new(big.Int).Mul(big.NewInt(100), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes4.Dpstmsg = append(nodes4.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_6, DepositAmount: new(big.Int).Mul(big.NewInt(100), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes4.Dpstmsg = append(nodes4.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_9, DepositAmount: new(big.Int).Mul(big.NewInt(100), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	depositNodes = append(depositNodes, nodes4)
	ic := &interestDelta{
		interestConfig: &mc.InterestCfg{PayInterval: 3600, AttenuationRate: 8500, AttenuationPeriod: 3000000, RewardMount: 3200},
		depositCfg:     depositcfg.GetDepositCfg(depositcfg.VersionA),
	}

	ic.GetReward(new(big.Int).SetUint64(32e17), depositNodes)
	//if !reflect.DeepEqual(got, tt.want) {
	//	t.Errorf("interestDelta.GetWeightDeposit() got = %v, want %v", got, tt.want)
	//}

}

func Test_interestDelta_small(t *testing.T) {
	log.InitLog(5)

	depositNodes := make([]common.DepositBase, 0)
	nodes0 := common.DepositBase{AddressA0: common.HexToAddress("01"), Dpstmsg: make([]common.DepositMsg, 0)}
	nodes0.Dpstmsg = append(nodes0.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes0.Dpstmsg = append(nodes0.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_1, DepositAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes0.Dpstmsg = append(nodes0.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_3, DepositAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes0.Dpstmsg = append(nodes0.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_6, DepositAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes0.Dpstmsg = append(nodes0.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_9, DepositAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	depositNodes = append(depositNodes, nodes0)
	nodes1 := common.DepositBase{AddressA0: common.HexToAddress("02"), Dpstmsg: make([]common.DepositMsg, 0)}
	nodes1.Dpstmsg = append(nodes1.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(100000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes1.Dpstmsg = append(nodes1.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_1, DepositAmount: new(big.Int).Mul(big.NewInt(100000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes1.Dpstmsg = append(nodes1.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_3, DepositAmount: new(big.Int).Mul(big.NewInt(100000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes1.Dpstmsg = append(nodes1.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_6, DepositAmount: new(big.Int).Mul(big.NewInt(100000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes1.Dpstmsg = append(nodes1.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_9, DepositAmount: new(big.Int).Mul(big.NewInt(100000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	depositNodes = append(depositNodes, nodes1)
	nodes2 := common.DepositBase{AddressA0: common.HexToAddress("03"), Dpstmsg: make([]common.DepositMsg, 0)}
	nodes2.Dpstmsg = append(nodes2.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(1000000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes2.Dpstmsg = append(nodes2.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_1, DepositAmount: new(big.Int).Mul(big.NewInt(1000000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes2.Dpstmsg = append(nodes2.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_3, DepositAmount: new(big.Int).Mul(big.NewInt(1000000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes2.Dpstmsg = append(nodes2.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_6, DepositAmount: new(big.Int).Mul(big.NewInt(1000000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes2.Dpstmsg = append(nodes2.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_9, DepositAmount: new(big.Int).Mul(big.NewInt(1000000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	depositNodes = append(depositNodes, nodes2)
	nodes3 := common.DepositBase{AddressA0: common.HexToAddress("04"), Dpstmsg: make([]common.DepositMsg, 0)}
	nodes3.Dpstmsg = append(nodes3.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(10000000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes3.Dpstmsg = append(nodes3.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_1, DepositAmount: new(big.Int).Mul(big.NewInt(10000000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes3.Dpstmsg = append(nodes3.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_3, DepositAmount: new(big.Int).Mul(big.NewInt(10000000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes3.Dpstmsg = append(nodes3.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_6, DepositAmount: new(big.Int).Mul(big.NewInt(10000000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes3.Dpstmsg = append(nodes3.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_9, DepositAmount: new(big.Int).Mul(big.NewInt(10000000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	depositNodes = append(depositNodes, nodes3)

	nodes4 := common.DepositBase{AddressA0: common.HexToAddress("05"), Dpstmsg: make([]common.DepositMsg, 0)}
	nodes4.Dpstmsg = append(nodes4.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(100), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes4.Dpstmsg = append(nodes4.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_1, DepositAmount: new(big.Int).Mul(big.NewInt(100), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes4.Dpstmsg = append(nodes4.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_3, DepositAmount: new(big.Int).Mul(big.NewInt(100), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes4.Dpstmsg = append(nodes4.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_6, DepositAmount: new(big.Int).Mul(big.NewInt(100), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes4.Dpstmsg = append(nodes4.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_9, DepositAmount: new(big.Int).Mul(big.NewInt(100), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	depositNodes = append(depositNodes, nodes4)
	ic := &interestDelta{
		interestConfig: &mc.InterestCfg{PayInterval: 3600, AttenuationRate: 8500, AttenuationPeriod: 3000000, RewardMount: 3200},
		depositCfg:     depositcfg.GetDepositCfg(depositcfg.VersionA),
	}

	ic.GetReward(new(big.Int).SetUint64(32e17), depositNodes)
	//if !reflect.DeepEqual(got, tt.want) {
	//	t.Errorf("interestDelta.GetWeightDeposit() got = %v, want %v", got, tt.want)
	//}

}

func Test_interestDelta_big(t *testing.T) {
	log.InitLog(5)

	depositNodes := make([]common.DepositBase, 0)
	nodes0 := common.DepositBase{AddressA0: common.HexToAddress("01"), Dpstmsg: make([]common.DepositMsg, 0)}
	nodes0.Dpstmsg = append(nodes0.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(1e5), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes0.Dpstmsg = append(nodes0.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_1, DepositAmount: new(big.Int).Mul(big.NewInt(1e5), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes0.Dpstmsg = append(nodes0.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_3, DepositAmount: new(big.Int).Mul(big.NewInt(1e5), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes0.Dpstmsg = append(nodes0.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_6, DepositAmount: new(big.Int).Mul(big.NewInt(1e5), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes0.Dpstmsg = append(nodes0.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_9, DepositAmount: new(big.Int).Mul(big.NewInt(1e5), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	depositNodes = append(depositNodes, nodes0)
	nodes1 := common.DepositBase{AddressA0: common.HexToAddress("02"), Dpstmsg: make([]common.DepositMsg, 0)}
	nodes1.Dpstmsg = append(nodes1.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(1e6), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes1.Dpstmsg = append(nodes1.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_1, DepositAmount: new(big.Int).Mul(big.NewInt(1e6), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes1.Dpstmsg = append(nodes1.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_3, DepositAmount: new(big.Int).Mul(big.NewInt(1e6), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes1.Dpstmsg = append(nodes1.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_6, DepositAmount: new(big.Int).Mul(big.NewInt(1e6), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes1.Dpstmsg = append(nodes1.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_9, DepositAmount: new(big.Int).Mul(big.NewInt(1e6), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	depositNodes = append(depositNodes, nodes1)
	nodes2 := common.DepositBase{AddressA0: common.HexToAddress("03"), Dpstmsg: make([]common.DepositMsg, 0)}
	nodes2.Dpstmsg = append(nodes2.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(1e7), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes2.Dpstmsg = append(nodes2.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_1, DepositAmount: new(big.Int).Mul(big.NewInt(1e7), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes2.Dpstmsg = append(nodes2.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_3, DepositAmount: new(big.Int).Mul(big.NewInt(1e7), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes2.Dpstmsg = append(nodes2.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_6, DepositAmount: new(big.Int).Mul(big.NewInt(1e7), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes2.Dpstmsg = append(nodes2.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_9, DepositAmount: new(big.Int).Mul(big.NewInt(1e7), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	depositNodes = append(depositNodes, nodes2)
	nodes3 := common.DepositBase{AddressA0: common.HexToAddress("04"), Dpstmsg: make([]common.DepositMsg, 0)}
	nodes3.Dpstmsg = append(nodes3.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(1e8), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes3.Dpstmsg = append(nodes3.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_1, DepositAmount: new(big.Int).Mul(big.NewInt(1e8), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes3.Dpstmsg = append(nodes3.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_3, DepositAmount: new(big.Int).Mul(big.NewInt(1e8), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes3.Dpstmsg = append(nodes3.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_6, DepositAmount: new(big.Int).Mul(big.NewInt(1e8), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes3.Dpstmsg = append(nodes3.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_9, DepositAmount: new(big.Int).Mul(big.NewInt(1e8), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	depositNodes = append(depositNodes, nodes3)

	nodes4 := common.DepositBase{AddressA0: common.HexToAddress("05"), Dpstmsg: make([]common.DepositMsg, 0)}
	nodes4.Dpstmsg = append(nodes4.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(1e9), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes4.Dpstmsg = append(nodes4.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_1, DepositAmount: new(big.Int).Mul(big.NewInt(1e9), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes4.Dpstmsg = append(nodes4.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_3, DepositAmount: new(big.Int).Mul(big.NewInt(1e9), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes4.Dpstmsg = append(nodes4.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_6, DepositAmount: new(big.Int).Mul(big.NewInt(1e9), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	nodes4.Dpstmsg = append(nodes4.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_9, DepositAmount: new(big.Int).Mul(big.NewInt(1e9), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	depositNodes = append(depositNodes, nodes4)
	ic := &interestDelta{
		interestConfig: &mc.InterestCfg{PayInterval: 3600, AttenuationRate: 8500, AttenuationPeriod: 3000000, RewardMount: 3200},
		depositCfg:     depositcfg.GetDepositCfg(depositcfg.VersionA),
	}

	ic.GetReward(new(big.Int).SetUint64(32e17), depositNodes)
	//if !reflect.DeepEqual(got, tt.want) {
	//	t.Errorf("interestDelta.GetWeightDeposit() got = %v, want %v", got, tt.want)
	//}

}

func Test_interestDelta_repeatposition(t *testing.T) {
	log.InitLog(5)

	depositNodes := make([]common.DepositBase, 0)
	nodes0 := common.DepositBase{AddressA0: common.HexToAddress("01"), Dpstmsg: make([]common.DepositMsg, 0)}
	nodes0.Dpstmsg = append(nodes0.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(1e5), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 0})
	nodes0.Dpstmsg = append(nodes0.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_1, DepositAmount: new(big.Int).Mul(big.NewInt(1e5), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 0})
	nodes0.Dpstmsg = append(nodes0.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_3, DepositAmount: new(big.Int).Mul(big.NewInt(1e5), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 0})
	nodes0.Dpstmsg = append(nodes0.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_6, DepositAmount: new(big.Int).Mul(big.NewInt(1e5), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 0})
	nodes0.Dpstmsg = append(nodes0.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_9, DepositAmount: new(big.Int).Mul(big.NewInt(1e5), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 0})
	depositNodes = append(depositNodes, nodes0)
	nodes1 := common.DepositBase{AddressA0: common.HexToAddress("02"), Dpstmsg: make([]common.DepositMsg, 0)}
	nodes1.Dpstmsg = append(nodes1.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(1e6), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 0})
	nodes1.Dpstmsg = append(nodes1.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_1, DepositAmount: new(big.Int).Mul(big.NewInt(1e6), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 0})
	nodes1.Dpstmsg = append(nodes1.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_3, DepositAmount: new(big.Int).Mul(big.NewInt(1e6), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 0})
	nodes1.Dpstmsg = append(nodes1.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_6, DepositAmount: new(big.Int).Mul(big.NewInt(1e6), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 0})
	nodes1.Dpstmsg = append(nodes1.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_9, DepositAmount: new(big.Int).Mul(big.NewInt(1e6), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 0})
	depositNodes = append(depositNodes, nodes1)
	nodes2 := common.DepositBase{AddressA0: common.HexToAddress("03"), Dpstmsg: make([]common.DepositMsg, 0)}
	nodes2.Dpstmsg = append(nodes2.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(1e7), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 0})
	nodes2.Dpstmsg = append(nodes2.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_1, DepositAmount: new(big.Int).Mul(big.NewInt(1e7), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 0})
	nodes2.Dpstmsg = append(nodes2.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_3, DepositAmount: new(big.Int).Mul(big.NewInt(1e7), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 0})
	nodes2.Dpstmsg = append(nodes2.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_6, DepositAmount: new(big.Int).Mul(big.NewInt(1e7), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 0})
	nodes2.Dpstmsg = append(nodes2.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_9, DepositAmount: new(big.Int).Mul(big.NewInt(1e7), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 0})
	depositNodes = append(depositNodes, nodes2)
	nodes3 := common.DepositBase{AddressA0: common.HexToAddress("04"), Dpstmsg: make([]common.DepositMsg, 0)}
	nodes3.Dpstmsg = append(nodes3.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(1e8), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 0})
	nodes3.Dpstmsg = append(nodes3.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_1, DepositAmount: new(big.Int).Mul(big.NewInt(1e8), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 0})
	nodes3.Dpstmsg = append(nodes3.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_3, DepositAmount: new(big.Int).Mul(big.NewInt(1e8), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 0})
	nodes3.Dpstmsg = append(nodes3.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_6, DepositAmount: new(big.Int).Mul(big.NewInt(1e8), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 0})
	nodes3.Dpstmsg = append(nodes3.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_9, DepositAmount: new(big.Int).Mul(big.NewInt(1e8), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 0})
	depositNodes = append(depositNodes, nodes3)

	nodes4 := common.DepositBase{AddressA0: common.HexToAddress("05"), Dpstmsg: make([]common.DepositMsg, 0)}
	nodes4.Dpstmsg = append(nodes4.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(1e9), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 0})
	nodes4.Dpstmsg = append(nodes4.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_1, DepositAmount: new(big.Int).Mul(big.NewInt(1e9), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 0})
	nodes4.Dpstmsg = append(nodes4.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_3, DepositAmount: new(big.Int).Mul(big.NewInt(1e9), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 0})
	nodes4.Dpstmsg = append(nodes4.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_6, DepositAmount: new(big.Int).Mul(big.NewInt(1e9), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 0})
	nodes4.Dpstmsg = append(nodes4.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_9, DepositAmount: new(big.Int).Mul(big.NewInt(1e9), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 0})
	depositNodes = append(depositNodes, nodes4)
	ic := &interestDelta{
		interestConfig: &mc.InterestCfg{PayInterval: 3600, AttenuationRate: 8500, AttenuationPeriod: 3000000, RewardMount: 3200},
		depositCfg:     depositcfg.GetDepositCfg(depositcfg.VersionA),
	}

	ic.GetReward(new(big.Int).SetUint64(32e17), depositNodes)
	//if !reflect.DeepEqual(got, tt.want) {
	//	t.Errorf("interestDelta.GetWeightDeposit() got = %v, want %v", got, tt.want)
	//}

}

func Test_interestDelta_leakposition(t *testing.T) {
	log.InitLog(5)

	depositNodes := make([]common.DepositBase, 0)
	nodes0 := common.DepositBase{AddressA0: common.HexToAddress("01"), Dpstmsg: make([]common.DepositMsg, 0)}
	nodes0.Dpstmsg = append(nodes0.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(1e5), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 0})
	nodes0.Dpstmsg = append(nodes0.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_1, DepositAmount: new(big.Int).Mul(big.NewInt(1e5), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 1})
	nodes0.Dpstmsg = append(nodes0.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_3, DepositAmount: new(big.Int).Mul(big.NewInt(1e5), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 5})
	nodes0.Dpstmsg = append(nodes0.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_6, DepositAmount: new(big.Int).Mul(big.NewInt(1e5), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 7})
	nodes0.Dpstmsg = append(nodes0.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_9, DepositAmount: new(big.Int).Mul(big.NewInt(1e5), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 9})
	depositNodes = append(depositNodes, nodes0)
	nodes1 := common.DepositBase{AddressA0: common.HexToAddress("02"), Dpstmsg: make([]common.DepositMsg, 0)}
	nodes1.Dpstmsg = append(nodes1.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(1e6), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 0})
	nodes1.Dpstmsg = append(nodes1.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_1, DepositAmount: new(big.Int).Mul(big.NewInt(1e6), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 1})
	nodes1.Dpstmsg = append(nodes1.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_3, DepositAmount: new(big.Int).Mul(big.NewInt(1e6), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 4})
	nodes1.Dpstmsg = append(nodes1.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_6, DepositAmount: new(big.Int).Mul(big.NewInt(1e6), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 89})
	nodes1.Dpstmsg = append(nodes1.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_9, DepositAmount: new(big.Int).Mul(big.NewInt(1e6), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 5})
	depositNodes = append(depositNodes, nodes1)
	nodes2 := common.DepositBase{AddressA0: common.HexToAddress("03"), Dpstmsg: make([]common.DepositMsg, 0)}
	nodes2.Dpstmsg = append(nodes2.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(1e7), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 5})
	nodes2.Dpstmsg = append(nodes2.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_1, DepositAmount: new(big.Int).Mul(big.NewInt(1e7), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 9})
	nodes2.Dpstmsg = append(nodes2.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_3, DepositAmount: new(big.Int).Mul(big.NewInt(1e7), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 3})
	nodes2.Dpstmsg = append(nodes2.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_6, DepositAmount: new(big.Int).Mul(big.NewInt(1e7), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 2})
	nodes2.Dpstmsg = append(nodes2.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_9, DepositAmount: new(big.Int).Mul(big.NewInt(1e7), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 6})
	depositNodes = append(depositNodes, nodes2)
	nodes3 := common.DepositBase{AddressA0: common.HexToAddress("04"), Dpstmsg: make([]common.DepositMsg, 0)}
	nodes3.Dpstmsg = append(nodes3.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(1e8), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 3})
	nodes3.Dpstmsg = append(nodes3.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_1, DepositAmount: new(big.Int).Mul(big.NewInt(1e8), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 23})
	nodes3.Dpstmsg = append(nodes3.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_3, DepositAmount: new(big.Int).Mul(big.NewInt(1e8), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 45})
	nodes3.Dpstmsg = append(nodes3.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_6, DepositAmount: new(big.Int).Mul(big.NewInt(1e8), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 67})
	nodes3.Dpstmsg = append(nodes3.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_9, DepositAmount: new(big.Int).Mul(big.NewInt(1e8), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 45})
	depositNodes = append(depositNodes, nodes3)

	nodes4 := common.DepositBase{AddressA0: common.HexToAddress("05"), Dpstmsg: make([]common.DepositMsg, 0)}
	nodes4.Dpstmsg = append(nodes4.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(1e9), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 34})
	nodes4.Dpstmsg = append(nodes4.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_1, DepositAmount: new(big.Int).Mul(big.NewInt(1e9), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 56})
	nodes4.Dpstmsg = append(nodes4.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_3, DepositAmount: new(big.Int).Mul(big.NewInt(1e9), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 67})
	nodes4.Dpstmsg = append(nodes4.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_6, DepositAmount: new(big.Int).Mul(big.NewInt(1e9), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 45})
	nodes4.Dpstmsg = append(nodes4.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_9, DepositAmount: new(big.Int).Mul(big.NewInt(1e9), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 67})
	depositNodes = append(depositNodes, nodes4)
	ic := &interestDelta{
		interestConfig: &mc.InterestCfg{PayInterval: 3600, AttenuationRate: 8500, AttenuationPeriod: 3000000, RewardMount: 3200},
		depositCfg:     depositcfg.GetDepositCfg(depositcfg.VersionA),
	}

	ic.GetReward(new(big.Int).SetUint64(32e17), depositNodes)
	//if !reflect.DeepEqual(got, tt.want) {
	//	t.Errorf("interestDelta.GetWeightDeposit() got = %v, want %v", got, tt.want)
	//}

}

func Test_interestDelta_noposition(t *testing.T) {
	log.InitLog(5)

	depositNodes := make([]common.DepositBase, 0)
	nodes0 := common.DepositBase{AddressA0: common.HexToAddress("01"), Dpstmsg: make([]common.DepositMsg, 0)}
	nodes0.Dpstmsg = append(nodes0.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(1e5), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 0})
	nodes0.Dpstmsg = append(nodes0.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_1, DepositAmount: new(big.Int).Mul(big.NewInt(1e5), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 1})
	nodes0.Dpstmsg = append(nodes0.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_3, DepositAmount: new(big.Int).Mul(big.NewInt(1e5), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 5})
	nodes0.Dpstmsg = append(nodes0.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_6, DepositAmount: new(big.Int).Mul(big.NewInt(1e5), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 7})
	nodes0.Dpstmsg = append(nodes0.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_9, DepositAmount: new(big.Int).Mul(big.NewInt(1e5), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 9})
	depositNodes = append(depositNodes, nodes0)
	nodes1 := common.DepositBase{AddressA0: common.HexToAddress("02"), Dpstmsg: make([]common.DepositMsg, 0)}
	depositNodes = append(depositNodes, nodes1)
	nodes2 := common.DepositBase{AddressA0: common.HexToAddress("03"), Dpstmsg: make([]common.DepositMsg, 0)}
	nodes2.Dpstmsg = append(nodes2.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(1e7), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 5})
	nodes2.Dpstmsg = append(nodes2.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_1, DepositAmount: new(big.Int).Mul(big.NewInt(1e7), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 9})
	nodes2.Dpstmsg = append(nodes2.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_3, DepositAmount: new(big.Int).Mul(big.NewInt(1e7), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 3})
	nodes2.Dpstmsg = append(nodes2.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_6, DepositAmount: new(big.Int).Mul(big.NewInt(1e7), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 2})
	nodes2.Dpstmsg = append(nodes2.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_9, DepositAmount: new(big.Int).Mul(big.NewInt(1e7), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 6})
	depositNodes = append(depositNodes, nodes2)
	nodes3 := common.DepositBase{AddressA0: common.HexToAddress("04"), Dpstmsg: make([]common.DepositMsg, 0)}
	nodes3.Dpstmsg = append(nodes3.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(1e8), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 3})
	nodes3.Dpstmsg = append(nodes3.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_1, DepositAmount: new(big.Int).Mul(big.NewInt(1e8), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 23})
	nodes3.Dpstmsg = append(nodes3.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_3, DepositAmount: new(big.Int).Mul(big.NewInt(1e8), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 45})
	nodes3.Dpstmsg = append(nodes3.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_6, DepositAmount: new(big.Int).Mul(big.NewInt(1e8), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 67})
	nodes3.Dpstmsg = append(nodes3.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_9, DepositAmount: new(big.Int).Mul(big.NewInt(1e8), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 45})
	depositNodes = append(depositNodes, nodes3)

	nodes4 := common.DepositBase{AddressA0: common.HexToAddress("05"), Dpstmsg: make([]common.DepositMsg, 0)}
	nodes4.Dpstmsg = append(nodes4.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(1e9), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 34})
	nodes4.Dpstmsg = append(nodes4.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_1, DepositAmount: new(big.Int).Mul(big.NewInt(1e9), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 56})
	nodes4.Dpstmsg = append(nodes4.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_3, DepositAmount: new(big.Int).Mul(big.NewInt(1e9), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 67})
	nodes4.Dpstmsg = append(nodes4.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_6, DepositAmount: new(big.Int).Mul(big.NewInt(1e9), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 45})
	nodes4.Dpstmsg = append(nodes4.Dpstmsg, common.DepositMsg{DepositType: depositcfg.MONTH_9, DepositAmount: new(big.Int).Mul(big.NewInt(1e9), util.ManPrice), Interest: new(big.Int).SetUint64(0), Position: 67})
	depositNodes = append(depositNodes, nodes4)
	ic := &interestDelta{
		interestConfig: &mc.InterestCfg{PayInterval: 3600, AttenuationRate: 8500, AttenuationPeriod: 3000000, RewardMount: 3200},
		depositCfg:     depositcfg.GetDepositCfg(depositcfg.VersionA),
	}

	ic.GetReward(new(big.Int).SetUint64(32e17), depositNodes)
	//if !reflect.DeepEqual(got, tt.want) {
	//	t.Errorf("interestDelta.GetWeightDeposit() got = %v, want %v", got, tt.want)
	//}

}

func Test_interestDelta_nodeposit(t *testing.T) {
	log.InitLog(5)

	depositNodes := make([]common.DepositBase, 0)

	ic := &interestDelta{
		interestConfig: &mc.InterestCfg{PayInterval: 3600, AttenuationRate: 8500, AttenuationPeriod: 3000000, RewardMount: 3200},
		depositCfg:     depositcfg.GetDepositCfg(depositcfg.VersionA),
	}

	ic.GetReward(new(big.Int).SetUint64(32e17), depositNodes)
	//if !reflect.DeepEqual(got, tt.want) {
	//	t.Errorf("interestDelta.GetWeightDeposit() got = %v, want %v", got, tt.want)
	//}

}
func Test_interestDelta_benchMark(t *testing.T) {

	depositNodes := make([]common.DepositBase, 1000000)
	nodes0 := common.DepositBase{AddressA0: common.HexToAddress("01"), Dpstmsg: make([]common.DepositMsg, 0)}
	for i := 0; i < 1000000; i++ {
		nodes0.Dpstmsg = append(nodes0.Dpstmsg, common.DepositMsg{DepositType: depositcfg.CurrentDeposit, DepositAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Interest: new(big.Int).SetUint64(0)})
	}

	depositNodes = append(depositNodes, nodes0)

	ic := &interestDelta{
		interestConfig: &mc.InterestCfg{PayInterval: 3600, AttenuationRate: 8500, AttenuationPeriod: 3000000, RewardMount: 3200},
		depositCfg:     depositcfg.GetDepositCfg(depositcfg.VersionA),
	}

	t.Logf("start", time.Now().Unix())
	ic.GetReward(new(big.Int).SetUint64(32e17), depositNodes)
	t.Logf("end", time.Now().Unix())

}

func Test_interestDelta_PayInterest(t *testing.T) {
	log.InitLog(5)
	ic := &interestDelta{
		interestConfig: &mc.InterestCfg{PayInterval: 3600, AttenuationRate: 8500, AttenuationPeriod: 3000000, RewardMount: 3200},
		depositCfg:     depositcfg.GetDepositCfg(depositcfg.VersionA),
	}
	chaindb := mandb.NewMemDatabase()
	currentState, _ := state.NewStateDBManage(nil, chaindb, state.NewDatabase(chaindb))
	depoistInfo.NewDepositInfo(nil)
	currentState.SetState(params.MAN_COIN, common.Address{}, common.HexToHash(params.DepositVersionKey_1), common.HexToHash(params.DepositVersion_1))
	operate0 := make([]common.OperationalInterestSlash, 0)
	operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Position: 0})
	operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Position: 1})
	operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Position: 2})
	operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Position: 3})
	operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Position: 4})
	depoistInfo.AddInterest_v2(currentState, common.HexToAddress("0x01"), common.CalculateDeposit{AddressA0: common.HexToAddress("0x01"), CalcDeposit: operate0})

	operate1 := make([]common.OperationalInterestSlash, 0)
	operate1 = append(operate1, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(20000), util.ManPrice), Position: 0})
	operate1 = append(operate1, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(15000), util.ManPrice), Position: 1})
	operate1 = append(operate1, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(13000), util.ManPrice), Position: 2})
	operate1 = append(operate1, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(14000), util.ManPrice), Position: 3})
	operate1 = append(operate1, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(11000), util.ManPrice), Position: 4})
	depoistInfo.AddInterest_v2(currentState, common.HexToAddress("0x02"), common.CalculateDeposit{AddressA0: common.HexToAddress("0x02"), CalcDeposit: operate1})
	currentState.AddBalance(params.MAN_COIN, common.MainAccount, common.InterestRewardAddress, new(big.Int).Mul(big.NewInt(1000000), util.ManPrice))
	matrixstate.SetVersionInfo(currentState, manparams.VersionGamma)
	matrixstate.SetInterestPayNum(currentState, 3600)

	//depoistInfo.SetVersion(currentState, 1559219781)

	ic.PayInterest(currentState, 3601)
}

func Test_interestDelta_PayInterest_benchMark(t *testing.T) {

	ic := &interestDelta{
		interestConfig: &mc.InterestCfg{PayInterval: 3600, AttenuationRate: 8500, AttenuationPeriod: 3000000, RewardMount: 3200},
		depositCfg:     depositcfg.GetDepositCfg(depositcfg.VersionA),
	}
	chaindb := mandb.NewMemDatabase()
	currentState, _ := state.NewStateDBManage(nil, chaindb, state.NewDatabase(chaindb))
	operate0 := make([]common.OperationalInterestSlash, 0)
	for i := 0; i < 1000000; i++ {

		operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(100), util.ManPrice), Position: uint64(i)})
	}

	depoistInfo.AddInterest_v2(currentState, common.HexToAddress("01"), common.CalculateDeposit{AddressA0: common.HexToAddress("0x02"), CalcDeposit: operate0})
	currentState.AddBalance(params.MAN_COIN, common.MainAccount, common.InterestRewardAddress, new(big.Int).Mul(big.NewInt(3e12), util.ManPrice))
	t.Logf("start", time.Now().Unix())
	ic.PayInterest(currentState, 3601)
	t.Logf("end", time.Now().Unix())
}

func Test_interestDelta_PayInterest_not(t *testing.T) {

	ic := &interestDelta{
		interestConfig: &mc.InterestCfg{PayInterval: 3600, AttenuationRate: 8500, AttenuationPeriod: 3000000, RewardMount: 3200},
		depositCfg:     depositcfg.GetDepositCfg(depositcfg.VersionA),
	}
	chaindb := mandb.NewMemDatabase()
	currentState, _ := state.NewStateDBManage(nil, chaindb, state.NewDatabase(chaindb))
	operate0 := make([]common.OperationalInterestSlash, 0)
	operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Position: 0})
	operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Position: 1})
	operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Position: 2})
	operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Position: 3})
	operate0 = append(operate0, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(10000), util.ManPrice), Position: 4})
	depoistInfo.AddInterest_v2(currentState, common.HexToAddress("0x01"), common.CalculateDeposit{AddressA0: common.HexToAddress("0x01"), CalcDeposit: operate0})

	operate1 := make([]common.OperationalInterestSlash, 0)
	operate1 = append(operate1, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(20000), util.ManPrice), Position: 0})
	operate1 = append(operate1, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(15000), util.ManPrice), Position: 1})
	operate1 = append(operate1, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(13000), util.ManPrice), Position: 2})
	operate1 = append(operate1, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(14000), util.ManPrice), Position: 3})
	operate1 = append(operate1, common.OperationalInterestSlash{OperAmount: new(big.Int).Mul(big.NewInt(11000), util.ManPrice), Position: 4})
	depoistInfo.AddInterest_v2(currentState, common.HexToAddress("0x02"), common.CalculateDeposit{AddressA0: common.HexToAddress("0x02"), CalcDeposit: operate1})
	currentState.AddBalance(params.MAN_COIN, common.MainAccount, common.InterestRewardAddress, new(big.Int).Mul(big.NewInt(30000), util.ManPrice))
	ic.PayInterest(currentState, 3601)
}
