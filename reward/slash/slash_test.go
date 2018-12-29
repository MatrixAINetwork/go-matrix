package slash

import (
	"fmt"
	"github.com/matrix/go-matrix/core/matrixstate"
	"github.com/matrix/go-matrix/core/vm"
	"github.com/matrix/go-matrix/depoistInfo"
	"github.com/matrix/go-matrix/params/manparams"
	"math/big"
	"testing"

	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/mandb"
	"github.com/matrix/go-matrix/params"

	"github.com/matrix/go-matrix/core/state"

	"bou.ke/monkey"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	. "github.com/smartystreets/goconvey/convey"
)

const account0 = "0x475baee143cf541ff3ee7b00c1c933129238d793"
const account1 = "0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"
const account2 = "0x519437b21e2a0b62788ab9235d0728dd7f1a7269"
const account3 = "0x29216818d3788c2505a593cbbb248907d47d9bce"

type Chain struct {
	curtop *mc.TopologyGraph
	elect  *mc.ElectGraph
}

func (chain *Chain) Config() *params.ChainConfig {
	return nil
}

// CurrentHeader retrieves the current header from the local chain.
func (chain *Chain) CurrentHeader() *types.Header {

	return nil
}

// GetHeader retrieves a block header from the database by hash and number.
func (chain *Chain) GetHeader(hash common.Hash, number uint64) *types.Header {
	return nil
}

// GetHeaderByHash retrieves a block header from the database by its hash.
func (chain *Chain) GetHeaderByHash(hash common.Hash) *types.Header {
	return nil
}

func (chain *Chain) GetBlockByNumber(number uint64) *types.Block {
	return nil
}

// GetBlock retrieves a block sfrom the database by hash and number.
func (chain *Chain) GetBlock(hash common.Hash, number uint64) *types.Block {
	return nil
}
func (chain *Chain) StateAt(root common.Hash) (*state.StateDB, error) {

	return nil, nil
}
func (chain *Chain) State() (*state.StateDB, error) {

	return nil, nil
}
func (chain *Chain) NewTopologyGraph(header *types.Header) (*mc.TopologyGraph, error) {
	return nil, nil
}

func (chain *Chain) GetHeaderByNumber(number uint64) *types.Header {
	header := &types.Header{
		Coinbase: common.Address{},
	}
	//txs := make([]types.SelfTransaction, 0)

	return header
}

func (chain *Chain) GetGraphByState(state matrixstate.StateDB) (*mc.TopologyGraph, *mc.ElectGraph, error) {
	return chain.curtop, chain.elect, nil
}
func (chain *Chain) SetGraphByState(node *mc.TopologyGraph, elect *mc.ElectGraph) {
	chain.curtop = node
	chain.elect = elect
	return
}
func (chain *Chain) GetMatrixStateData(key string) (interface{}, error) {
	return nil, nil
}

func (chain *Chain) GetMatrixStateDataByNumber(key string, number uint64) (interface{}, error) {
	if key == mc.MSKeyElectGraph {
		EleList := make([]mc.ElectNodeInfo, 0)
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress(account0), Position: 8192, Type: common.RoleValidator})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress(account1), Position: 8193, Type: common.RoleValidator})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress(account2), Position: 8194, Type: common.RoleValidator})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress(account3), Position: 8194, Type: common.RoleValidator})
		return chain.elect, nil
	}

	return nil, nil
}

func (chain *Chain) GetSuperBlockNum() (uint64, error) {

	return 0, nil
}

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

func TestBlockSlash_CalcSlash(t *testing.T) {
	log.InitLog(3)
	fakeFuntion()
	bc := &Chain{}
	slash := New(bc, &State{0})
	Convey("计算惩罚", t, func() {

		EleList := make([]mc.ElectNodeInfo, 0)
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress(account0), Position: 8192, Type: common.RoleValidator})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress(account1), Position: 8193, Type: common.RoleValidator})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress(account2), Position: 8194, Type: common.RoleValidator})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress(account3), Position: 8194, Type: common.RoleValidator})
		//EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Position: 8195})
		bc.SetGraphByState(nil, &mc.ElectGraph{Number: 4, ElectList: EleList})
		statedb, _ := state.New(common.Hash{}, state.NewDatabase(mandb.NewMemDatabase()))
		rewards := make(map[common.Address]*big.Int, 0)
		rewards[common.HexToAddress(account0)] = big.NewInt(4e+18)
		rewards[common.HexToAddress(account1)] = big.NewInt(4e+18)
		rewards[common.HexToAddress(account2)] = big.NewInt(4e+18)
		rewards[common.HexToAddress(account3)] = big.NewInt(4e+18)
		uptime := make(map[common.Address]uint64, 0)
		uptime[common.HexToAddress(account0)] = 97
		uptime[common.HexToAddress(account1)] = 75
		uptime[common.HexToAddress(account2)] = 48
		uptime[common.HexToAddress(account3)] = 20
		slash.CalcSlash(statedb, 101, uptime, rewards)
	})
}

func fakeFuntion() {
	monkey.Patch(matrixstate.GetDataByState, func(key string, state matrixstate.StateDB) (interface{}, error) {
		fmt.Println("use monkey  matrixstate.GetDataByState")
		if key == mc.MSKeyBroadcastInterval {
			return &mc.BCIntervalInfo{LastBCNumber: 100, LastReelectNumber: 0, BCInterval: 100}, nil
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

		if key == mc.MSKeySlashCfg {

			return &mc.SlashCfgStruct{SlashCalc: "1", SlashRate: 7500}, nil
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

		inteval1 := &mc.BCIntervalInfo{LastBCNumber: 100, LastReelectNumber: 0, BCInterval: 100}

		interval2, _ := manparams.NewBCIntervalWithInterval(inteval1)
		return interval2, nil
	})
	monkey.Patch(depoistInfo.AddSlash, func(stateDB vm.StateDB, address common.Address, slash *big.Int) error {
		fmt.Println("use monkey  depoistInfo.AddSlash", "address", address.String(), "slash", slash.Uint64())
		return nil
	})
}

func TestBlockSlash_CalcSlash22(t *testing.T) {
	log.InitLog(3)

	Convey("计算惩罚99", t, func() {
		monkey.Patch(matrixstate.GetDataByState, func(key string, state matrixstate.StateDB) (interface{}, error) {
			fmt.Println("use monkey  matrixstate.GetDataByState")
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

			if key == mc.MSKeySlashCfg {

				return &mc.SlashCfgStruct{SlashCalc: "1", SlashRate: 7500}, nil
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

			inteval1 := &mc.BCIntervalInfo{LastBCNumber: 100, LastReelectNumber: 0, BCInterval: 100}

			interval2, _ := manparams.NewBCIntervalWithInterval(inteval1)
			return interval2, nil
		})
		monkey.Patch(depoistInfo.AddSlash, func(stateDB vm.StateDB, address common.Address, slash *big.Int) error {
			fmt.Println("use monkey  depoistInfo.AddSlash", "address", address.String(), "slash", slash.Uint64())
			return nil
		})
		bc := &Chain{}
		slash := New(bc, &State{0})
		EleList := make([]mc.ElectNodeInfo, 0)
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress(account0), Position: 8192, Type: common.RoleValidator})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress(account1), Position: 8193, Type: common.RoleValidator})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress(account2), Position: 8194, Type: common.RoleValidator})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress(account3), Position: 8194, Type: common.RoleValidator})
		//EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Position: 8195})
		bc.SetGraphByState(nil, &mc.ElectGraph{Number: 4, ElectList: EleList})
		statedb, _ := state.New(common.Hash{}, state.NewDatabase(mandb.NewMemDatabase()))
		rewards := make(map[common.Address]*big.Int, 0)
		rewards[common.HexToAddress(account0)] = big.NewInt(4e+18)
		rewards[common.HexToAddress(account1)] = big.NewInt(4e+18)
		rewards[common.HexToAddress(account2)] = big.NewInt(4e+18)
		rewards[common.HexToAddress(account3)] = big.NewInt(4e+18)
		uptime := make(map[common.Address]uint64, 0)
		uptime[common.HexToAddress(account0)] = 97
		uptime[common.HexToAddress(account1)] = 75
		uptime[common.HexToAddress(account2)] = 48
		uptime[common.HexToAddress(account3)] = 20
		slash.CalcSlash(statedb, 99, uptime, rewards)
	})
	Convey("计算惩罚100", t, func() {
		monkey.Patch(matrixstate.GetDataByState, func(key string, state matrixstate.StateDB) (interface{}, error) {
			fmt.Println("use monkey  matrixstate.GetDataByState")
			if key == mc.MSKeyBroadcastInterval {
				return &mc.BCIntervalInfo{LastBCNumber: 100, LastReelectNumber: 0, BCInterval: 100}, nil
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

			if key == mc.MSKeySlashCfg {

				return &mc.SlashCfgStruct{SlashCalc: "1", SlashRate: 7500}, nil
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

			inteval1 := &mc.BCIntervalInfo{LastBCNumber: 100, LastReelectNumber: 0, BCInterval: 100}

			interval2, _ := manparams.NewBCIntervalWithInterval(inteval1)
			return interval2, nil
		})
		monkey.Patch(depoistInfo.AddSlash, func(stateDB vm.StateDB, address common.Address, slash *big.Int) error {
			fmt.Println("use monkey  depoistInfo.AddSlash", "address", address.String(), "slash", slash.Uint64())
			return nil
		})
		bc := &Chain{}
		slash := New(bc, &State{0})
		EleList := make([]mc.ElectNodeInfo, 0)
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress(account0), Position: 8192, Type: common.RoleValidator})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress(account1), Position: 8193, Type: common.RoleValidator})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress(account2), Position: 8194, Type: common.RoleValidator})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress(account3), Position: 8194, Type: common.RoleValidator})
		//EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Position: 8195})
		bc.SetGraphByState(nil, &mc.ElectGraph{Number: 4, ElectList: EleList})
		statedb, _ := state.New(common.Hash{}, state.NewDatabase(mandb.NewMemDatabase()))
		rewards := make(map[common.Address]*big.Int, 0)
		rewards[common.HexToAddress(account0)] = big.NewInt(4e+18)
		rewards[common.HexToAddress(account1)] = big.NewInt(4e+18)
		rewards[common.HexToAddress(account2)] = big.NewInt(4e+18)
		rewards[common.HexToAddress(account3)] = big.NewInt(4e+18)
		uptime := make(map[common.Address]uint64, 0)
		uptime[common.HexToAddress(account0)] = 97
		uptime[common.HexToAddress(account1)] = 75
		uptime[common.HexToAddress(account2)] = 48
		uptime[common.HexToAddress(account3)] = 20
		slash.CalcSlash(statedb, 100, uptime, rewards)
	})
}

func TestBlockSlash_CalcSlash44(t *testing.T) {
	log.InitLog(3)

	Convey("拓扑图为nil", t, func() {
		monkey.Patch(matrixstate.GetDataByState, func(key string, state matrixstate.StateDB) (interface{}, error) {
			fmt.Println("use monkey  matrixstate.GetDataByState")
			if key == mc.MSKeyBroadcastInterval {
				return &mc.BCIntervalInfo{LastBCNumber: 100, LastReelectNumber: 0, BCInterval: 100}, nil
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

			if key == mc.MSKeySlashCfg {

				return &mc.SlashCfgStruct{SlashCalc: "1", SlashRate: 7500}, nil
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

			inteval1 := &mc.BCIntervalInfo{LastBCNumber: 100, LastReelectNumber: 0, BCInterval: 100}

			interval2, _ := manparams.NewBCIntervalWithInterval(inteval1)
			return interval2, nil
		})
		monkey.Patch(depoistInfo.AddSlash, func(stateDB vm.StateDB, address common.Address, slash *big.Int) error {
			fmt.Println("use monkey  depoistInfo.AddSlash", "address", address.String(), "slash", slash.Uint64())
			return nil
		})
		bc := &Chain{}
		slash := New(bc, &State{0})
		EleList := make([]mc.ElectNodeInfo, 0)
		//EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Position: 8195})
		bc.SetGraphByState(nil, &mc.ElectGraph{Number: 4, ElectList: EleList})
		statedb, _ := state.New(common.Hash{}, state.NewDatabase(mandb.NewMemDatabase()))
		rewards := make(map[common.Address]*big.Int, 0)
		rewards[common.HexToAddress(account0)] = big.NewInt(4e+18)
		rewards[common.HexToAddress(account1)] = big.NewInt(4e+18)
		rewards[common.HexToAddress(account2)] = big.NewInt(4e+18)
		rewards[common.HexToAddress(account3)] = big.NewInt(4e+18)
		uptime := make(map[common.Address]uint64, 0)
		uptime[common.HexToAddress(account0)] = 97
		uptime[common.HexToAddress(account1)] = 75
		uptime[common.HexToAddress(account2)] = 48
		uptime[common.HexToAddress(account3)] = 20
		slash.CalcSlash(statedb, 100, uptime, rewards)
	})
}

func TestBlockSlash_CalcSlash55(t *testing.T) {
	log.InitLog(3)
	fakeFuntion()
	bc := &Chain{}
	slash := New(bc, &State{0})
	Convey("uptime为nil", t, func() {

		EleList := make([]mc.ElectNodeInfo, 0)
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress(account0), Position: 8192, Type: common.RoleValidator})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress(account1), Position: 8193, Type: common.RoleValidator})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress(account2), Position: 8194, Type: common.RoleValidator})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress(account3), Position: 8194, Type: common.RoleValidator})
		//EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Position: 8195})
		bc.SetGraphByState(nil, &mc.ElectGraph{Number: 4, ElectList: EleList})
		statedb, _ := state.New(common.Hash{}, state.NewDatabase(mandb.NewMemDatabase()))
		rewards := make(map[common.Address]*big.Int, 0)
		rewards[common.HexToAddress(account0)] = big.NewInt(4e+18)
		rewards[common.HexToAddress(account1)] = big.NewInt(4e+18)
		rewards[common.HexToAddress(account2)] = big.NewInt(4e+18)
		rewards[common.HexToAddress(account3)] = big.NewInt(4e+18)
		uptime := make(map[common.Address]uint64, 0)
		uptime[common.HexToAddress(account0)] = 97
		uptime[common.HexToAddress(account1)] = 75
		uptime[common.HexToAddress(account2)] = 48
		uptime[common.HexToAddress(account3)] = 20
		slash.CalcSlash(statedb, 101, nil, rewards)
	})
}

func TestBlockSlash_CalcSlash6(t *testing.T) {
	log.InitLog(3)
	fakeFuntion()
	bc := &Chain{}
	slash := New(bc, &State{0})
	Convey("利息为nil", t, func() {

		EleList := make([]mc.ElectNodeInfo, 0)
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress(account0), Position: 8192, Type: common.RoleValidator})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress(account1), Position: 8193, Type: common.RoleValidator})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress(account2), Position: 8194, Type: common.RoleValidator})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress(account3), Position: 8194, Type: common.RoleValidator})
		//EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Position: 8195})
		bc.SetGraphByState(nil, &mc.ElectGraph{Number: 4, ElectList: EleList})
		statedb, _ := state.New(common.Hash{}, state.NewDatabase(mandb.NewMemDatabase()))
		rewards := make(map[common.Address]*big.Int, 0)
		rewards[common.HexToAddress(account0)] = big.NewInt(4e+18)
		rewards[common.HexToAddress(account1)] = big.NewInt(4e+18)
		rewards[common.HexToAddress(account2)] = big.NewInt(4e+18)
		rewards[common.HexToAddress(account3)] = big.NewInt(4e+18)
		uptime := make(map[common.Address]uint64, 0)
		uptime[common.HexToAddress(account0)] = 97
		uptime[common.HexToAddress(account1)] = 75
		uptime[common.HexToAddress(account2)] = 48
		uptime[common.HexToAddress(account3)] = 20
		slash.CalcSlash(statedb, 101, uptime, nil)
	})

	Convey("利息值非法", t, func() {

		EleList := make([]mc.ElectNodeInfo, 0)
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress(account0), Position: 8192, Type: common.RoleValidator})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress(account1), Position: 8193, Type: common.RoleValidator})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress(account2), Position: 8194, Type: common.RoleValidator})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress(account3), Position: 8194, Type: common.RoleValidator})
		//EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Position: 8195})
		bc.SetGraphByState(nil, &mc.ElectGraph{Number: 4, ElectList: EleList})
		statedb, _ := state.New(common.Hash{}, state.NewDatabase(mandb.NewMemDatabase()))
		rewards := make(map[common.Address]*big.Int, 0)
		rewards[common.HexToAddress(account0)] = big.NewInt(-4e+18)
		rewards[common.HexToAddress(account1)] = big.NewInt(4e+18)
		rewards[common.HexToAddress(account2)] = big.NewInt(4e+18)
		rewards[common.HexToAddress(account3)] = big.NewInt(4e+18)
		uptime := make(map[common.Address]uint64, 0)
		uptime[common.HexToAddress(account0)] = 97
		uptime[common.HexToAddress(account1)] = 75
		uptime[common.HexToAddress(account2)] = 48
		uptime[common.HexToAddress(account3)] = 20
		slash.CalcSlash(statedb, 101, uptime, rewards)
	})

	Convey("插入超级区块", t, func() {

		EleList := make([]mc.ElectNodeInfo, 0)
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress(account0), Position: 8192, Type: common.RoleInnerMiner})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress(account1), Position: 8193, Type: common.RoleValidator})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress(account2), Position: 8194, Type: common.RoleValidator})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress(account3), Position: 8194, Type: common.RoleValidator})
		//EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Position: 8195})
		bc.SetGraphByState(nil, &mc.ElectGraph{Number: 4, ElectList: EleList})
		statedb, _ := state.New(common.Hash{}, state.NewDatabase(mandb.NewMemDatabase()))
		rewards := make(map[common.Address]*big.Int, 0)
		rewards[common.HexToAddress(account0)] = big.NewInt(4e+18)
		rewards[common.HexToAddress(account1)] = big.NewInt(4e+18)
		rewards[common.HexToAddress(account2)] = big.NewInt(4e+18)
		rewards[common.HexToAddress(account3)] = big.NewInt(4e+18)
		uptime := make(map[common.Address]uint64, 0)
		uptime[common.HexToAddress(account0)] = 97
		uptime[common.HexToAddress(account1)] = 75
		uptime[common.HexToAddress(account2)] = 48
		uptime[common.HexToAddress(account3)] = 20
		slash.CalcSlash(statedb, 105, uptime, rewards)
	})
}

//
//func TestBlockSlash_CalcSlash44(t *testing.T) {
//	log.InitLog(3)
//
//	slash := New(&Chain{})
//	Convey("计算惩罚", t, func() {
//		monkey.Patch(depoistInfo.AddSlash, func(stateDB vm.StateDB, address common.Address, slash *big.Int) error {
//			fmt.Println("use monkey  depoistInfo.AddSlash", "address", address.String(), "slash", slash.Uint64())
//			return nil
//		})
//		statedb, _ := state.New(common.Hash{}, state.NewDatabase(mandb.NewMemDatabase()))
//		monkey.Patch(depoistInfo.GetOnlineTime, func(stateDB vm.StateDB, address common.Address) (*big.Int, error) {
//			fmt.Println("use monkey  ca.GetOnlineTime")
//			onlineTime := big.NewInt(291)
//			if stateDB == statedb {
//				switch {
//				case address.Equal(common.HexToAddress(account0)):
//					onlineTime = big.NewInt(291 * 2) //100%
//				case address.Equal(common.HexToAddress(account1)):
//					onlineTime = big.NewInt(291) //0%
//				case address.Equal(common.HexToAddress(account2)):
//					onlineTime = big.NewInt(291 + 291/2) //50%
//				case address.Equal(common.HexToAddress(account3)):
//					onlineTime = big.NewInt(291 + 291/4) //25%
//
//				}
//
//			}
//
//			return onlineTime, nil
//		})
//		monkey.Patch(ca.GetTopologyByNumber, func(reqTypes common.RoleType, number uint64) (*mc.TopologyGraph, error) {
//			fmt.Println("use monkey  ca.GetTopologyByNumber")
//			newGraph := &mc.TopologyGraph{
//				Number:        number,
//				NodeList:      make([]mc.TopologyNodeInfo, 0),
//				CurNodeNumber: 0,
//			}
//			if common.RoleValidator == reqTypes&common.RoleValidator {
//				newGraph.NodeList = append(newGraph.NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"), Position: 8192})
//				newGraph.NodeList = append(newGraph.NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Position: 8193})
//				newGraph.NodeList = append(newGraph.NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Position: 8194})
//				newGraph.NodeList = append(newGraph.NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Position: 8195})
//				newGraph.CurNodeNumber = 4
//			}
//
//			return newGraph, nil
//		})
//		monkey.Patch(depoistInfo.GetInterest, func(stateDB vm.StateDB, address common.Address) (*big.Int, error) {
//			fmt.Println("use monkey  ca.GetInterest")
//
//			return nil, errors.New("利息非法")
//		})
//		rewards := make(map[common.Address]*big.Int, 0)
//		rewards[common.HexToAddress(account0)] = big.NewInt(4e+18)
//		rewards[common.HexToAddress(account1)] = big.NewInt(4e+18)
//		rewards[common.HexToAddress(account2)] = big.NewInt(4e+18)
//		rewards[common.HexToAddress(account3)] = big.NewInt(4e+18)
//		slash.CalcSlash(statedb, common.GetReElectionInterval()+1)
//	})
//}
//
//func TestBlockSlash_CalcSlash55(t *testing.T) {
//	log.InitLog(3)
//
//	slash := New(&Chain{})
//	Convey("计算惩罚", t, func() {
//		monkey.Patch(depoistInfo.AddSlash, func(stateDB vm.StateDB, address common.Address, slash *big.Int) error {
//			fmt.Println("use monkey  depoistInfo.AddSlash", "address", address.String(), "slash", slash.Uint64())
//			return nil
//		})
//		statedb, _ := state.New(common.Hash{}, state.NewDatabase(mandb.NewMemDatabase()))
//		monkey.Patch(depoistInfo.GetOnlineTime, func(stateDB vm.StateDB, address common.Address) (*big.Int, error) {
//			fmt.Println("use monkey  ca.GetOnlineTime")
//			onlineTime := big.NewInt(291)
//			if stateDB == statedb {
//				switch {
//				case address.Equal(common.HexToAddress(account0)):
//					onlineTime = big.NewInt(291 * 2) //100%
//				case address.Equal(common.HexToAddress(account1)):
//					onlineTime = big.NewInt(291) //0%
//				case address.Equal(common.HexToAddress(account2)):
//					onlineTime = big.NewInt(291 + 291/2) //50%
//				case address.Equal(common.HexToAddress(account3)):
//					onlineTime = big.NewInt(291 + 291/4) //25%
//
//				}
//
//			}
//
//			return onlineTime, nil
//		})
//		monkey.Patch(ca.GetTopologyByNumber, func(reqTypes common.RoleType, number uint64) (*mc.TopologyGraph, error) {
//			fmt.Println("use monkey  ca.GetTopologyByNumber")
//			newGraph := &mc.TopologyGraph{
//				Number:        number,
//				NodeList:      make([]mc.TopologyNodeInfo, 0),
//				CurNodeNumber: 0,
//			}
//			if common.RoleValidator == reqTypes&common.RoleValidator {
//				newGraph.NodeList = append(newGraph.NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"), Position: 8192})
//				newGraph.NodeList = append(newGraph.NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Position: 8193})
//				newGraph.NodeList = append(newGraph.NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Position: 8194})
//				newGraph.NodeList = append(newGraph.NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Position: 8195})
//				newGraph.CurNodeNumber = 4
//			}
//
//			return newGraph, nil
//		})
//		monkey.Patch(depoistInfo.GetInterest, func(stateDB vm.StateDB, address common.Address) (*big.Int, error) {
//			fmt.Println("use monkey  ca.GetInterest")
//
//			return big.NewInt(100), nil
//		})
//
//		monkey.Patch(depoistInfo.GetOnlineTime, func(stateDB vm.StateDB, address common.Address) (*big.Int, error) {
//			fmt.Println("use monkey  ca.GetInterest")
//
//			return nil, errors.New("利息非法")
//		})
//		rewards := make(map[common.Address]*big.Int, 0)
//		rewards[common.HexToAddress(account0)] = big.NewInt(4e+18)
//		rewards[common.HexToAddress(account1)] = big.NewInt(4e+18)
//		rewards[common.HexToAddress(account2)] = big.NewInt(4e+18)
//		rewards[common.HexToAddress(account3)] = big.NewInt(4e+18)
//		slash.CalcSlash(statedb, common.GetReElectionInterval()+1)
//	})
//}
