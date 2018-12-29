package interest

import (
	"errors"
	"fmt"
	"github.com/matrix/go-matrix/core/matrixstate"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/params/manparams"
	"github.com/matrix/go-matrix/reward/util"
	"math/big"
	"testing"

	"bou.ke/monkey"
	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/core/vm"
	"github.com/matrix/go-matrix/depoistInfo"
	"github.com/matrix/go-matrix/log"

	"github.com/matrix/go-matrix/common"
	. "github.com/smartystreets/goconvey/convey"
)

type State struct {
	balance int64
}

func (st *State) GetBalance(addr common.Address) common.BalanceType {
	return []common.BalanceSlice{{common.MainAccount, big.NewInt(st.balance)}}
}
func (st *State) GetBalanceByType(addr common.Address, accType uint32) *big.Int {
	return big.NewInt(st.balance)
}

func (st *State) CreateAccount(common.Address) {

}

func (st *State) SubBalance(uint32, common.Address, *big.Int) {}
func (st *State) AddBalance(uint32, common.Address, *big.Int) {}

func (st *State) GetNonce(common.Address) uint64  { return 0 }
func (st *State) SetNonce(common.Address, uint64) {}

func (st *State) GetCodeHash(common.Address) common.Hash { return common.Hash{} }
func (st *State) GetCode(common.Address) []byte          { return nil }
func (st *State) SetCode(common.Address, []byte)         {}
func (st *State) GetCodeSize(common.Address) int         { return 0 }

func (st *State) AddRefund(uint64)  {}
func (st *State) GetRefund() uint64 { return 0 }

func (st *State) GetState(common.Address, common.Hash) common.Hash  { return common.Hash{} }
func (st *State) SetState(common.Address, common.Hash, common.Hash) {}

func (st *State) Suicide(common.Address) bool     { return true }
func (st *State) HasSuicided(common.Address) bool { return true }

// Exist reports whether the given account exists in state.
// Notably this should also return true for suicided accounts.
func (st *State) Exist(common.Address) bool { return true }

// Empty returns whether the given account is empty. Empty
// is defined according to EIP161 (balance = nonce = code = 0).
func (st *State) Empty(common.Address) bool { return true }

func (st *State) RevertToSnapshot(int) {}
func (st *State) Snapshot() int        { return 0 }

func (st *State) AddLog(*types.Log)               {}
func (st *State) AddPreimage(common.Hash, []byte) {}

func (st *State) ForEachStorage(common.Address, func(common.Hash, common.Hash) bool) {}

func (st *State) GetMatrixData(hash common.Hash) (val []byte) {
	return nil

}
func (st *State) SetMatrixData(hash common.Hash, val []byte) {
	return
}

func (st *State) CommitSaveTx() {
	return
}
func (st *State) DeleteMxData(hash common.Hash, val []byte) {

}
func (st *State) Dump() []byte {
	return nil
}

func (st *State) Finalise(deleteEmptyObjects bool) {
	return
}

func (st *State) GetAllEntrustSignFrom(authFrom common.Address) []common.Address {
	return nil
}
func (st *State) GetAllEntrustGasFrom(authFrom common.Address) []common.Address {
	return nil
}

func (st *State) GetGasAuthFrom(entrustFrom common.Address, height uint64) common.Address {
	return common.Address{}
}
func (st *State) GetAuthFrom(entrustFrom common.Address, height uint64) common.Address {
	return common.Address{}
}
func (st *State) GetEntrustFrom(authFrom common.Address, height uint64) []common.Address {
	return nil
}

func (st *State) GetLogs(hash common.Hash) []*types.Log {
	return nil
}

func (st *State) GetSaveTx(typ byte, key uint32, hash []common.Hash, isdel bool) {

}
func (st *State) SaveTx(typ byte, key uint32, data map[common.Hash][]byte) {

}

func (st *State) GetStateByteArray(common.Address, common.Hash) []byte {
	return nil
}
func (st *State) SetStateByteArray(common.Address, common.Hash, []byte) {

}

func (st *State) NewBTrie(typ byte) {

}
func Test_interest_Calc(t *testing.T) {
	monkey.Patch(manparams.IsBroadcastNumber, func(number uint64, stateNumber uint64) bool {
		fmt.Println("use monkey  manparams.IsBroadcastNumber")

		return false
	})

	monkey.Patch(matrixstate.GetDataByState, func(key string, state matrixstate.StateDB) (interface{}, error) {
		fmt.Println("use monkey  manparams.IsBroadcastNumber")
		if key == mc.MSKeyBroadcastInterval {
			return &mc.BCIntervalInfo{LastBCNumber: 0, LastReelectNumber: 0, BCInterval: 100}, nil
		}
		if key == mc.MSKeyLotteryCfg {
			info := make([]mc.LotteryInfo, 0)
			info = append(info, mc.LotteryInfo{PrizeLevel: 0, PrizeNum: 1, PrizeMoney: 6})
			return &mc.LotteryCfgStruct{LotteryCalc: "1", LotteryInfo: info}, nil
		}
		if key == mc.MSKEYLotteryNum {
			info := make([]mc.LotteryInfo, 0)
			info = append(info, mc.LotteryInfo{PrizeLevel: 0, PrizeNum: 1, PrizeMoney: 6})
			return &mc.LotteryCfgStruct{LotteryCalc: "1", LotteryInfo: info}, nil
		}
		if key == mc.MSKeyInterestCfg {

			return &mc.InterestCfgStruct{InterestCalc: "1", CalcInterval: 100, PayInterval: 300}, nil
		}
		if key == mc.MSKeyVIPConfig {
			vip := make([]mc.VIPConfig, 0)
			vip = append(vip, mc.VIPConfig{MinMoney: 0, InterestRate: 5})
			vip = append(vip, mc.VIPConfig{MinMoney: 40000, InterestRate: 10})
			vip = append(vip, mc.VIPConfig{MinMoney: 100000, InterestRate: 15})
			return &vip, nil
		}
		return nil, nil
	})

	monkey.Patch(matrixstate.GetNumByState, func(key string, state matrixstate.StateDB) (uint64, error) {

		switch key {
		case mc.MSInterestCalcNum:
			return 1, nil
		case mc.MSInterestPayNum:
			return 1, nil
		}

		return uint64(3), nil
	})

	monkey.Patch(manparams.NewBCIntervalByNumber, func(blockNumber uint64) (*manparams.BCInterval, error) {
		fmt.Println("use monkey NewBCIntervalByNumber")

		inteval1 := &mc.BCIntervalInfo{LastBCNumber: 0, LastReelectNumber: 0, BCInterval: 100}

		interval2, _ := manparams.NewBCIntervalWithInterval(inteval1)
		return interval2, nil
	})

	monkey.Patch(ca.GetElectedByHeight, func(height *big.Int) ([]vm.DepositDetail, error) {
		fmt.Println("use monkey  ca.GetElectedByHeightAndRole")
		Deposit := make([]vm.DepositDetail, 0)
		Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Deposit: new(big.Int).Mul(big.NewInt(10000), util.ManPrice)})
		Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Deposit: new(big.Int).Mul(big.NewInt(40000), util.ManPrice)})
		Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Deposit: new(big.Int).Mul(big.NewInt(100000), util.ManPrice)})
		Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bcf"), Deposit: new(big.Int).Mul(big.NewInt(200000), util.ManPrice)})
		return Deposit, nil
	})
	insterestMap := make(map[common.Address]*big.Int, 0)
	monkey.Patch(depoistInfo.AddInterest, func(stateDB vm.StateDB, address common.Address, reward *big.Int) error {
		insterestMap[address] = reward
		return nil
	})
	monkey.Patch(depoistInfo.GetAllInterest, func(stateDB vm.StateDB) map[common.Address]*big.Int {
		return insterestMap
	})
	Convey("利息测试计算利息", t, func() {
		log.InitLog(3)

		interestTest := New(&State{5e+18})

		interestTest.InterestCalc(&State{0}, 101)

	})

	Convey("利息测试支付利息", t, func() {
		log.InitLog(3)

		interestTest := New(&State{5e+18})

		interestTest.InterestCalc(&State{5e+18}, 3601)

	})

}

func Test_interest_pay(t *testing.T) {
	monkey.Patch(manparams.IsBroadcastNumber, func(number uint64, stateNumber uint64) bool {
		fmt.Println("use monkey  manparams.IsBroadcastNumber")

		return false
	})

	monkey.Patch(matrixstate.GetDataByState, func(key string, state matrixstate.StateDB) (interface{}, error) {
		fmt.Println("use monkey  manparams.IsBroadcastNumber")
		if key == mc.MSKeyBroadcastInterval {
			return &mc.BCIntervalInfo{LastBCNumber: 0, LastReelectNumber: 0, BCInterval: 100}, nil
		}
		if key == mc.MSKeyLotteryCfg {
			info := make([]mc.LotteryInfo, 0)
			info = append(info, mc.LotteryInfo{PrizeLevel: 0, PrizeNum: 1, PrizeMoney: 6})
			return &mc.LotteryCfgStruct{LotteryCalc: "1", LotteryInfo: info}, nil
		}
		if key == mc.MSKEYLotteryNum {
			info := make([]mc.LotteryInfo, 0)
			info = append(info, mc.LotteryInfo{PrizeLevel: 0, PrizeNum: 1, PrizeMoney: 6})
			return &mc.LotteryCfgStruct{LotteryCalc: "1", LotteryInfo: info}, nil
		}
		if key == mc.MSKeyInterestCfg {

			return &mc.InterestCfgStruct{InterestCalc: "1", CalcInterval: 100, PayInterval: 3600}, nil
		}
		if key == mc.MSKeyVIPConfig {
			vip := make([]mc.VIPConfig, 0)
			vip = append(vip, mc.VIPConfig{MinMoney: 0, InterestRate: 5})
			vip = append(vip, mc.VIPConfig{MinMoney: 40000, InterestRate: 10})
			vip = append(vip, mc.VIPConfig{MinMoney: 100000, InterestRate: 15})
			return &vip, nil
		}
		return nil, nil
	})

	monkey.Patch(matrixstate.GetNumByState, func(key string, state matrixstate.StateDB) (uint64, error) {

		switch key {
		case mc.MSInterestCalcNum:
			return 1, nil
		case mc.MSInterestPayNum:
			return 1, nil
		}

		return uint64(3), nil
	})

	monkey.Patch(manparams.NewBCIntervalByNumber, func(blockNumber uint64) (*manparams.BCInterval, error) {
		fmt.Println("use monkey NewBCIntervalByNumber")

		inteval1 := &mc.BCIntervalInfo{LastBCNumber: 0, LastReelectNumber: 0, BCInterval: 100}

		interval2, _ := manparams.NewBCIntervalWithInterval(inteval1)
		return interval2, nil
	})

	monkey.Patch(ca.GetElectedByHeight, func(height *big.Int) ([]vm.DepositDetail, error) {
		fmt.Println("use monkey  ca.GetElectedByHeightAndRole")
		Deposit := make([]vm.DepositDetail, 0)
		Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Deposit: new(big.Int).Mul(big.NewInt(10000), util.ManPrice)})
		Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Deposit: new(big.Int).Mul(big.NewInt(40000), util.ManPrice)})
		Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Deposit: new(big.Int).Mul(big.NewInt(100000), util.ManPrice)})
		Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bcf"), Deposit: new(big.Int).Mul(big.NewInt(200000), util.ManPrice)})
		return Deposit, nil
	})
	insterestMap := make(map[common.Address]*big.Int, 0)
	monkey.Patch(depoistInfo.AddInterest, func(stateDB vm.StateDB, address common.Address, reward *big.Int) error {
		insterestMap[address] = reward
		return nil
	})
	monkey.Patch(depoistInfo.GetAllInterest, func(stateDB vm.StateDB) map[common.Address]*big.Int {
		return insterestMap
	})
	Convey("计算利息0", t, func() {
		log.InitLog(3)

		interestTest := New(&State{5e+18})

		interestTest.InterestCalc(&State{5e+18}, 99)

	})
	Convey("计算利息1", t, func() {
		log.InitLog(3)

		interestTest := New(&State{5e+18})

		interestTest.InterestCalc(&State{5e+18}, 100)

	})

	Convey("支付利息0", t, func() {
		log.InitLog(3)

		interestTest := New(&State{5e+18})

		interestTest.InterestCalc(&State{5e+18}, 3599)

	})
	Convey("支付利息1", t, func() {
		log.InitLog(3)

		interestTest := New(&State{5e+18})

		interestTest.InterestCalc(&State{5e+18}, 3600)

	})
	Convey("支付利息2", t, func() {
		log.InitLog(3)

		interestTest := New(&State{5e+18})

		interestTest.InterestCalc(&State{5e+18}, 3601)

	})

}

func Test_interest_number(t *testing.T) {
	monkey.Patch(manparams.IsBroadcastNumber, func(number uint64, stateNumber uint64) bool {
		fmt.Println("use monkey  manparams.IsBroadcastNumber")

		return false
	})

	monkey.Patch(matrixstate.GetDataByState, func(key string, state matrixstate.StateDB) (interface{}, error) {
		fmt.Println("use monkey  manparams.IsBroadcastNumber")
		if key == mc.MSKeyBroadcastInterval {
			return &mc.BCIntervalInfo{LastBCNumber: 0, LastReelectNumber: 0, BCInterval: 100}, nil
		}
		if key == mc.MSKeyLotteryCfg {
			info := make([]mc.LotteryInfo, 0)
			info = append(info, mc.LotteryInfo{PrizeLevel: 0, PrizeNum: 1, PrizeMoney: 6})
			return &mc.LotteryCfgStruct{LotteryCalc: "1", LotteryInfo: info}, nil
		}
		if key == mc.MSKEYLotteryNum {
			info := make([]mc.LotteryInfo, 0)
			info = append(info, mc.LotteryInfo{PrizeLevel: 0, PrizeNum: 1, PrizeMoney: 6})
			return &mc.LotteryCfgStruct{LotteryCalc: "1", LotteryInfo: info}, nil
		}
		if key == mc.MSKeyInterestCfg {

			return &mc.InterestCfgStruct{InterestCalc: "1", CalcInterval: 100, PayInterval: 3600}, nil
		}
		if key == mc.MSKeyVIPConfig {
			vip := make([]mc.VIPConfig, 0)
			vip = append(vip, mc.VIPConfig{MinMoney: 0, InterestRate: 5})
			vip = append(vip, mc.VIPConfig{MinMoney: 40000, InterestRate: 10})
			vip = append(vip, mc.VIPConfig{MinMoney: 100000, InterestRate: 15})
			return &vip, nil
		}
		return nil, nil
	})

	monkey.Patch(matrixstate.GetNumByState, func(key string, state matrixstate.StateDB) (uint64, error) {

		switch key {
		case mc.MSInterestCalcNum:
			return 1, nil
		case mc.MSInterestPayNum:
			return 1, nil
		}

		return uint64(3), nil
	})

	monkey.Patch(manparams.NewBCIntervalByNumber, func(blockNumber uint64) (*manparams.BCInterval, error) {
		fmt.Println("use monkey NewBCIntervalByNumber")

		inteval1 := &mc.BCIntervalInfo{LastBCNumber: 0, LastReelectNumber: 0, BCInterval: 100}

		interval2, _ := manparams.NewBCIntervalWithInterval(inteval1)
		return interval2, nil
	})

	monkey.Patch(ca.GetElectedByHeight, func(height *big.Int) ([]vm.DepositDetail, error) {
		fmt.Println("use monkey  ca.GetElectedByHeightAndRole")
		Deposit := make([]vm.DepositDetail, 0)
		Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Deposit: new(big.Int).Mul(big.NewInt(10000), util.ManPrice)})
		Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Deposit: new(big.Int).Mul(big.NewInt(40000), util.ManPrice)})
		Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Deposit: new(big.Int).Mul(big.NewInt(100000), util.ManPrice)})
		Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bcf"), Deposit: new(big.Int).Mul(big.NewInt(200000), util.ManPrice)})
		return Deposit, nil
	})
	insterestMap := make(map[common.Address]*big.Int, 0)
	monkey.Patch(depoistInfo.AddInterest, func(stateDB vm.StateDB, address common.Address, reward *big.Int) error {
		insterestMap[address] = reward
		return nil
	})
	monkey.Patch(depoistInfo.GetAllInterest, func(stateDB vm.StateDB) map[common.Address]*big.Int {
		return insterestMap
	})
	Convey("余额不足", t, func() {
		log.InitLog(3)

		interestTest := New(&State{5e+18})

		interestTest.InterestCalc(&State{0}, 3601)

	})

}

func Test_interest_number2(t *testing.T) {
	monkey.Patch(manparams.IsBroadcastNumber, func(number uint64, stateNumber uint64) bool {
		fmt.Println("use monkey  manparams.IsBroadcastNumber")

		return false
	})

	monkey.Patch(matrixstate.GetDataByState, func(key string, state matrixstate.StateDB) (interface{}, error) {
		fmt.Println("use monkey  manparams.IsBroadcastNumber")
		if key == mc.MSKeyBroadcastInterval {
			return &mc.BCIntervalInfo{LastBCNumber: 0, LastReelectNumber: 0, BCInterval: 100}, nil
		}
		if key == mc.MSKeyLotteryCfg {
			info := make([]mc.LotteryInfo, 0)
			info = append(info, mc.LotteryInfo{PrizeLevel: 0, PrizeNum: 1, PrizeMoney: 6})
			return &mc.LotteryCfgStruct{LotteryCalc: "1", LotteryInfo: info}, nil
		}
		if key == mc.MSKEYLotteryNum {
			info := make([]mc.LotteryInfo, 0)
			info = append(info, mc.LotteryInfo{PrizeLevel: 0, PrizeNum: 1, PrizeMoney: 6})
			return &mc.LotteryCfgStruct{LotteryCalc: "1", LotteryInfo: info}, nil
		}
		if key == mc.MSKeyInterestCfg {

			return &mc.InterestCfgStruct{InterestCalc: "1", CalcInterval: 100, PayInterval: 3600}, nil
		}
		if key == mc.MSKeyVIPConfig {
			vip := make([]mc.VIPConfig, 0)
			vip = append(vip, mc.VIPConfig{MinMoney: 0, InterestRate: 5})
			vip = append(vip, mc.VIPConfig{MinMoney: 40000, InterestRate: 10})
			vip = append(vip, mc.VIPConfig{MinMoney: 100000, InterestRate: 15})
			return &vip, nil
		}
		return nil, nil
	})

	monkey.Patch(matrixstate.GetNumByState, func(key string, state matrixstate.StateDB) (uint64, error) {

		switch key {
		case mc.MSInterestCalcNum:
			return 1, nil
		case mc.MSInterestPayNum:
			return 1, nil
		}

		return uint64(3), nil
	})

	monkey.Patch(manparams.NewBCIntervalByNumber, func(blockNumber uint64) (*manparams.BCInterval, error) {
		fmt.Println("use monkey NewBCIntervalByNumber")

		inteval1 := &mc.BCIntervalInfo{LastBCNumber: 0, LastReelectNumber: 0, BCInterval: 100}

		interval2, _ := manparams.NewBCIntervalWithInterval(inteval1)
		return interval2, nil
	})

	monkey.Patch(ca.GetElectedByHeight, func(height *big.Int) ([]vm.DepositDetail, error) {
		fmt.Println("use monkey  ca.GetElectedByHeightAndRole")
		Deposit := make([]vm.DepositDetail, 0)
		Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Deposit: new(big.Int).Mul(big.NewInt(10000), util.ManPrice)})
		Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Deposit: new(big.Int).Mul(big.NewInt(40000), util.ManPrice)})
		Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Deposit: new(big.Int).Mul(big.NewInt(100000), util.ManPrice)})
		Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bcf"), Deposit: new(big.Int).Mul(big.NewInt(200000), util.ManPrice)})
		return Deposit, nil
	})
	insterestMap := make(map[common.Address]*big.Int, 0)
	monkey.Patch(depoistInfo.AddInterest, func(stateDB vm.StateDB, address common.Address, reward *big.Int) error {
		insterestMap[address] = reward
		return nil
	})
	monkey.Patch(depoistInfo.GetAllInterest, func(stateDB vm.StateDB) map[common.Address]*big.Int {
		return insterestMap
	})
	Convey("利息测试计算利息", t, func() {
		log.InitLog(3)

		interestTest := New(&State{5e+18})

		interestTest.InterestCalc(nil, 101)

	})

}

func Test_interest3(t *testing.T) {
	monkey.Patch(manparams.IsBroadcastNumber, func(number uint64, stateNumber uint64) bool {
		fmt.Println("use monkey  manparams.IsBroadcastNumber")

		return false
	})

	monkey.Patch(matrixstate.GetDataByState, func(key string, state matrixstate.StateDB) (interface{}, error) {
		fmt.Println("use monkey  manparams.IsBroadcastNumber")
		if key == mc.MSKeyBroadcastInterval {
			return &mc.BCIntervalInfo{LastBCNumber: 0, LastReelectNumber: 0, BCInterval: 100}, nil
		}
		if key == mc.MSKeyLotteryCfg {
			info := make([]mc.LotteryInfo, 0)
			info = append(info, mc.LotteryInfo{PrizeLevel: 0, PrizeNum: 1, PrizeMoney: 6})
			return &mc.LotteryCfgStruct{LotteryCalc: "1", LotteryInfo: info}, nil
		}
		if key == mc.MSKEYLotteryNum {
			info := make([]mc.LotteryInfo, 0)
			info = append(info, mc.LotteryInfo{PrizeLevel: 0, PrizeNum: 1, PrizeMoney: 6})
			return &mc.LotteryCfgStruct{LotteryCalc: "1", LotteryInfo: info}, nil
		}
		if key == mc.MSKeyInterestCfg {

			return &mc.InterestCfgStruct{InterestCalc: "1", CalcInterval: 100, PayInterval: 3600}, nil
		}
		if key == mc.MSKeyVIPConfig {
			vip := make([]mc.VIPConfig, 0)
			vip = append(vip, mc.VIPConfig{MinMoney: 0, InterestRate: 5})
			vip = append(vip, mc.VIPConfig{MinMoney: 40000, InterestRate: 10})
			vip = append(vip, mc.VIPConfig{MinMoney: 100000, InterestRate: 15})
			return &vip, nil
		}
		return nil, nil
	})

	monkey.Patch(matrixstate.GetNumByState, func(key string, state matrixstate.StateDB) (uint64, error) {

		switch key {
		case mc.MSInterestCalcNum:
			return 1, nil
		case mc.MSInterestPayNum:
			return 1, nil
		}

		return uint64(3), nil
	})

	monkey.Patch(manparams.NewBCIntervalByNumber, func(blockNumber uint64) (*manparams.BCInterval, error) {
		fmt.Println("use monkey NewBCIntervalByNumber")

		inteval1 := &mc.BCIntervalInfo{LastBCNumber: 0, LastReelectNumber: 0, BCInterval: 100}

		interval2, _ := manparams.NewBCIntervalWithInterval(inteval1)
		return interval2, nil
	})

	monkey.Patch(ca.GetElectedByHeight, func(height *big.Int) ([]vm.DepositDetail, error) {
		fmt.Println("use monkey  ca.GetElectedByHeightAndRole")
		Deposit := make([]vm.DepositDetail, 0)
		Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Deposit: new(big.Int).Mul(big.NewInt(10000), util.ManPrice)})
		Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Deposit: new(big.Int).Mul(big.NewInt(40000), util.ManPrice)})
		Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Deposit: new(big.Int).Mul(big.NewInt(100000), util.ManPrice)})
		Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bcf"), Deposit: new(big.Int).Mul(big.NewInt(200000), util.ManPrice)})
		return Deposit, nil
	})
	insterestMap := make(map[common.Address]*big.Int, 0)
	monkey.Patch(depoistInfo.AddInterest, func(stateDB vm.StateDB, address common.Address, reward *big.Int) error {
		insterestMap[address] = reward
		return nil
	})
	monkey.Patch(depoistInfo.GetAllInterest, func(stateDB vm.StateDB) map[common.Address]*big.Int {
		return insterestMap
	})
	Convey("利息测试计算利息", t, func() {
		log.InitLog(3)
		monkey.Patch(ca.GetElectedByHeight, func(height *big.Int) ([]vm.DepositDetail, error) {
			fmt.Println("use monkey  ca.GetElectedByHeightAndRole")
			Deposit := make([]vm.DepositDetail, 0)

			return Deposit, errors.New("获取抵押错误")
		})
		interestTest := New(&State{5e+18})

		interestTest.InterestCalc(&State{0}, 101)

	})

}
func Test_interest4(t *testing.T) {
	monkey.Patch(manparams.IsBroadcastNumber, func(number uint64, stateNumber uint64) bool {
		fmt.Println("use monkey  manparams.IsBroadcastNumber")

		return false
	})

	monkey.Patch(matrixstate.GetDataByState, func(key string, state matrixstate.StateDB) (interface{}, error) {
		fmt.Println("use monkey  manparams.IsBroadcastNumber")
		if key == mc.MSKeyBroadcastInterval {
			return &mc.BCIntervalInfo{LastBCNumber: 0, LastReelectNumber: 0, BCInterval: 100}, nil
		}
		if key == mc.MSKeyLotteryCfg {
			info := make([]mc.LotteryInfo, 0)
			info = append(info, mc.LotteryInfo{PrizeLevel: 0, PrizeNum: 1, PrizeMoney: 6})
			return &mc.LotteryCfgStruct{LotteryCalc: "1", LotteryInfo: info}, nil
		}
		if key == mc.MSKEYLotteryNum {
			info := make([]mc.LotteryInfo, 0)
			info = append(info, mc.LotteryInfo{PrizeLevel: 0, PrizeNum: 1, PrizeMoney: 6})
			return &mc.LotteryCfgStruct{LotteryCalc: "1", LotteryInfo: info}, nil
		}
		if key == mc.MSKeyInterestCfg {

			return &mc.InterestCfgStruct{InterestCalc: "1", CalcInterval: 100, PayInterval: 3600}, nil
		}
		if key == mc.MSKeyVIPConfig {
			vip := make([]mc.VIPConfig, 0)
			vip = append(vip, mc.VIPConfig{MinMoney: 0, InterestRate: 5})
			vip = append(vip, mc.VIPConfig{MinMoney: 40000, InterestRate: 10})
			vip = append(vip, mc.VIPConfig{MinMoney: 100000, InterestRate: 15})
			return &vip, nil
		}
		return nil, nil
	})

	monkey.Patch(matrixstate.GetNumByState, func(key string, state matrixstate.StateDB) (uint64, error) {

		switch key {
		case mc.MSInterestCalcNum:
			return 1, nil
		case mc.MSInterestPayNum:
			return 1, nil
		}

		return uint64(3), nil
	})

	monkey.Patch(manparams.NewBCIntervalByNumber, func(blockNumber uint64) (*manparams.BCInterval, error) {
		fmt.Println("use monkey NewBCIntervalByNumber")

		inteval1 := &mc.BCIntervalInfo{LastBCNumber: 0, LastReelectNumber: 0, BCInterval: 100}

		interval2, _ := manparams.NewBCIntervalWithInterval(inteval1)
		return interval2, nil
	})

	monkey.Patch(ca.GetElectedByHeight, func(height *big.Int) ([]vm.DepositDetail, error) {
		fmt.Println("use monkey  ca.GetElectedByHeightAndRole")
		Deposit := make([]vm.DepositDetail, 0)
		Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Deposit: new(big.Int).Mul(big.NewInt(10000), util.ManPrice)})
		Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Deposit: new(big.Int).Mul(big.NewInt(40000), util.ManPrice)})
		Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Deposit: new(big.Int).Mul(big.NewInt(100000), util.ManPrice)})
		Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bcf"), Deposit: new(big.Int).Mul(big.NewInt(200000), util.ManPrice)})
		return Deposit, nil
	})
	insterestMap := make(map[common.Address]*big.Int, 0)
	monkey.Patch(depoistInfo.AddInterest, func(stateDB vm.StateDB, address common.Address, reward *big.Int) error {
		insterestMap[address] = reward
		return nil
	})
	monkey.Patch(depoistInfo.GetAllInterest, func(stateDB vm.StateDB) map[common.Address]*big.Int {
		return insterestMap
	})
	Convey("抵押列表长度为0", t, func() {
		log.InitLog(3)
		monkey.Patch(ca.GetElectedByHeight, func(height *big.Int) ([]vm.DepositDetail, error) {
			fmt.Println("use monkey  ca.GetElectedByHeightAndRole")
			Deposit := make([]vm.DepositDetail, 0)

			return Deposit, nil
		})
		interestTest := New(&State{5e+18})

		interestTest.InterestCalc(&State{0}, 101)

	})

}

func Test_interest5(t *testing.T) {
	monkey.Patch(manparams.IsBroadcastNumber, func(number uint64, stateNumber uint64) bool {
		fmt.Println("use monkey  manparams.IsBroadcastNumber")

		return false
	})

	monkey.Patch(matrixstate.GetDataByState, func(key string, state matrixstate.StateDB) (interface{}, error) {
		fmt.Println("use monkey  manparams.IsBroadcastNumber")
		if key == mc.MSKeyBroadcastInterval {
			return &mc.BCIntervalInfo{LastBCNumber: 0, LastReelectNumber: 0, BCInterval: 100}, nil
		}
		if key == mc.MSKeyLotteryCfg {
			info := make([]mc.LotteryInfo, 0)
			info = append(info, mc.LotteryInfo{PrizeLevel: 0, PrizeNum: 1, PrizeMoney: 6})
			return &mc.LotteryCfgStruct{LotteryCalc: "1", LotteryInfo: info}, nil
		}
		if key == mc.MSKEYLotteryNum {
			info := make([]mc.LotteryInfo, 0)
			info = append(info, mc.LotteryInfo{PrizeLevel: 0, PrizeNum: 1, PrizeMoney: 6})
			return &mc.LotteryCfgStruct{LotteryCalc: "1", LotteryInfo: info}, nil
		}
		if key == mc.MSKeyInterestCfg {

			return &mc.InterestCfgStruct{InterestCalc: "1", CalcInterval: 100, PayInterval: 3600}, nil
		}
		if key == mc.MSKeyVIPConfig {
			vip := make([]mc.VIPConfig, 0)
			vip = append(vip, mc.VIPConfig{MinMoney: 0, InterestRate: 5})
			vip = append(vip, mc.VIPConfig{MinMoney: 40000, InterestRate: 10})
			vip = append(vip, mc.VIPConfig{MinMoney: 100000, InterestRate: 15})
			return &vip, nil
		}
		return nil, nil
	})

	monkey.Patch(matrixstate.GetNumByState, func(key string, state matrixstate.StateDB) (uint64, error) {

		switch key {
		case mc.MSInterestCalcNum:
			return 1, nil
		case mc.MSInterestPayNum:
			return 1, nil
		}

		return uint64(3), nil
	})

	monkey.Patch(manparams.NewBCIntervalByNumber, func(blockNumber uint64) (*manparams.BCInterval, error) {
		fmt.Println("use monkey NewBCIntervalByNumber")

		inteval1 := &mc.BCIntervalInfo{LastBCNumber: 0, LastReelectNumber: 0, BCInterval: 100}

		interval2, _ := manparams.NewBCIntervalWithInterval(inteval1)
		return interval2, nil
	})

	monkey.Patch(ca.GetElectedByHeight, func(height *big.Int) ([]vm.DepositDetail, error) {
		fmt.Println("use monkey  ca.GetElectedByHeightAndRole")
		Deposit := make([]vm.DepositDetail, 0)
		Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Deposit: new(big.Int).Mul(big.NewInt(10000), util.ManPrice)})
		Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Deposit: new(big.Int).Mul(big.NewInt(40000), util.ManPrice)})
		Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Deposit: new(big.Int).Mul(big.NewInt(100000), util.ManPrice)})
		Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bcf"), Deposit: new(big.Int).Mul(big.NewInt(200000), util.ManPrice)})
		return Deposit, nil
	})
	insterestMap := make(map[common.Address]*big.Int, 0)
	monkey.Patch(depoistInfo.AddInterest, func(stateDB vm.StateDB, address common.Address, reward *big.Int) error {
		insterestMap[address] = reward
		return nil
	})
	monkey.Patch(depoistInfo.GetAllInterest, func(stateDB vm.StateDB) map[common.Address]*big.Int {
		return insterestMap
	})
	Convey("利息错误", t, func() {
		log.InitLog(3)
		monkey.Patch(ca.GetElectedByHeight, func(height *big.Int) ([]vm.DepositDetail, error) {
			fmt.Println("use monkey  ca.GetElectedByHeightAndRole")
			Deposit := make([]vm.DepositDetail, 0)
			Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Deposit: big.NewInt(-1)})
			Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Deposit: big.NewInt(3e+18)})
			Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Deposit: big.NewInt(4e+18)})
			Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bcf"), Deposit: big.NewInt(2e+18)})
			return Deposit, nil
		})

		monkey.Patch(depoistInfo.AddInterest, func(stateDB vm.StateDB, address common.Address, reward *big.Int) error {
			return nil
		})

		interestTest := New(&State{5e+18})

		interestTest.InterestCalc(&State{0}, 101)

	})

}

func Test_interest6(t *testing.T) {
	monkey.Patch(manparams.IsBroadcastNumber, func(number uint64, stateNumber uint64) bool {
		fmt.Println("use monkey  manparams.IsBroadcastNumber")

		return false
	})

	monkey.Patch(matrixstate.GetDataByState, func(key string, state matrixstate.StateDB) (interface{}, error) {
		fmt.Println("use monkey  manparams.IsBroadcastNumber")
		if key == mc.MSKeyBroadcastInterval {
			return &mc.BCIntervalInfo{LastBCNumber: 0, LastReelectNumber: 0, BCInterval: 100}, nil
		}
		if key == mc.MSKeyLotteryCfg {
			info := make([]mc.LotteryInfo, 0)
			info = append(info, mc.LotteryInfo{PrizeLevel: 0, PrizeNum: 1, PrizeMoney: 6})
			return &mc.LotteryCfgStruct{LotteryCalc: "1", LotteryInfo: info}, nil
		}
		if key == mc.MSKEYLotteryNum {
			info := make([]mc.LotteryInfo, 0)
			info = append(info, mc.LotteryInfo{PrizeLevel: 0, PrizeNum: 1, PrizeMoney: 6})
			return &mc.LotteryCfgStruct{LotteryCalc: "1", LotteryInfo: info}, nil
		}
		if key == mc.MSKeyInterestCfg {

			return &mc.InterestCfgStruct{InterestCalc: "1", CalcInterval: 100, PayInterval: 3600}, nil
		}
		if key == mc.MSKeyVIPConfig {
			vip := make([]mc.VIPConfig, 0)
			vip = append(vip, mc.VIPConfig{MinMoney: 0, InterestRate: 5})
			vip = append(vip, mc.VIPConfig{MinMoney: 40000, InterestRate: 10})
			vip = append(vip, mc.VIPConfig{MinMoney: 100000, InterestRate: 15})
			return &vip, nil
		}
		return nil, nil
	})

	monkey.Patch(matrixstate.GetNumByState, func(key string, state matrixstate.StateDB) (uint64, error) {

		switch key {
		case mc.MSInterestCalcNum:
			return 1, nil
		case mc.MSInterestPayNum:
			return 1, nil
		}

		return uint64(3), nil
	})

	monkey.Patch(manparams.NewBCIntervalByNumber, func(blockNumber uint64) (*manparams.BCInterval, error) {
		fmt.Println("use monkey NewBCIntervalByNumber")

		inteval1 := &mc.BCIntervalInfo{LastBCNumber: 0, LastReelectNumber: 0, BCInterval: 100}

		interval2, _ := manparams.NewBCIntervalWithInterval(inteval1)
		return interval2, nil
	})

	insterestMap := make(map[common.Address]*big.Int, 0)
	monkey.Patch(depoistInfo.AddInterest, func(stateDB vm.StateDB, address common.Address, reward *big.Int) error {
		insterestMap[address] = reward
		return nil
	})
	monkey.Patch(depoistInfo.GetAllInterest, func(stateDB vm.StateDB) map[common.Address]*big.Int {
		return insterestMap
	})
	Convey("利息错误", t, func() {
		log.InitLog(3)
		monkey.Patch(ca.GetElectedByHeight, func(height *big.Int) ([]vm.DepositDetail, error) {
			fmt.Println("use monkey  ca.GetElectedByHeightAndRole")
			Deposit := make([]vm.DepositDetail, 0)
			Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Deposit: big.NewInt(-1)})
			Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Deposit: big.NewInt(3e+18)})
			Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Deposit: big.NewInt(4e+18)})
			Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bcf"), Deposit: big.NewInt(2e+18)})
			return Deposit, nil
		})

		monkey.Patch(depoistInfo.AddInterest, func(stateDB vm.StateDB, address common.Address, reward *big.Int) error {
			return nil
		})

		interestTest := New(&State{5e+18})

		interestTest.InterestCalc(&State{0}, 3601)
	})
}
