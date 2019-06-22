// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package txsreward

import (
	"fmt"
	"math/big"
	"math/rand"
	"testing"

	"github.com/MatrixAINetwork/go-matrix/mandb"

	"github.com/MatrixAINetwork/go-matrix/params/manparams"

	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/params"

	"github.com/MatrixAINetwork/go-matrix/core/matrixstate"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/reward/util"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	testAddress             = "0x8605cdbbdb6d264aa742e77020dcbc58fcdce182"
	ValidatorsTxsRewardRate = uint64(util.RewardFullRate) //验证者交易奖励比例100%
	MinerTxsRewardRate      = uint64(0)                   //矿工交易奖励比例0%
	FoundationTxsRewardRate = uint64(0)                   //基金会交易奖励比例0%

	MinerOutRewardRate     = uint64(4000) //出块矿工奖励40%
	ElectedMinerRewardRate = uint64(6000) //当选矿工奖励60%

	LeaderRewardRate            = uint64(4000) //出块验证者（leader）奖励40%
	ElectedValidatorsRewardRate = uint64(6000) //当选验证者奖励60%

	OriginElectOfflineRewardRate = uint64(5000) //初选下线验证者奖励50%
	BackupRate                   = uint64(5000) //当前替补验证者奖励50%
)

var myNodeId string = "4b2f638f46c7ae5b1564ca7015d716621848a0d9be66f1d1e91d566d2a70eedc2f11e92b743acb8d97dec3fb412c1b2f66afd7fbb9399d4fb2423619eaa51411"

type Chain struct {
	curtop   *mc.TopologyGraph
	elect    *mc.ElectGraph
	coinBase common.Address
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
	header := &types.Header{
		Coinbase: chain.coinBase,
	}
	return header
}

func (chain *Chain) GetBlockByNumber(number uint64) *types.Block {
	return nil
}
func (chain *Chain) StateAtBlockHash(hash common.Hash) (*state.StateDBManage, error) {
	return nil, nil
}

// GetBlock retrieves a block sfrom the database by hash and number.
func (chain *Chain) GetBlock(hash common.Hash, number uint64) *types.Block {
	return nil
}
func (chain *Chain) StateAt(root []common.CoinRoot) (*state.StateDBManage, error) {

	return nil, nil
}
func (chain *Chain) State() (*state.StateDBManage, error) {

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
	return nil, nil
}

func (chain *Chain) GetSuperBlockNum() (uint64, error) {

	return 0, nil
}
func (chain *Chain) GetAncestorHash(sonHash common.Hash, ancestorNumber uint64) (common.Hash, error) {
	return common.Hash{}, nil
}

func (chain *Chain) GetBroadcastInterval() (*mc.BCIntervalInfo, error) {
	return &mc.BCIntervalInfo{LastBCNumber: 0, LastReelectNumber: 0, BCInterval: 100}, nil
}
func (chain *Chain) GetBroadcastIntervalByHash(hash common.Hash) (*mc.BCIntervalInfo, error) {
	return &mc.BCIntervalInfo{LastBCNumber: 0, LastReelectNumber: 0, BCInterval: 100}, nil
}
func (chain *Chain) GetBroadcastIntervalByNumber(number uint64) (*mc.BCIntervalInfo, error) {
	return &mc.BCIntervalInfo{LastBCNumber: 0, LastReelectNumber: 0, BCInterval: 100}, nil
}

type InnerSeed struct {
}

func (s *InnerSeed) GetSeed(num uint64) *big.Int {
	random := rand.New(rand.NewSource(0))
	return new(big.Int).SetUint64(random.Uint64())
}

type State struct {
	balance uint64
}

func (st *State) GetBalance(typ string, addr common.Address) common.BalanceType {
	return []common.BalanceSlice{{common.MainAccount, new(big.Int).SetUint64(st.balance)}}
}

func (st *State) GetMatrixData(hash common.Hash) (val []byte) {
	return nil

}
func (st *State) SetMatrixData(hash common.Hash, val []byte) {
	return
}

func TestCalcTxsFees(t *testing.T) {
	log.InitLog(3)

	rrc := mc.RewardRateCfg{
		MinerOutRate:        MinerOutRewardRate,
		ElectedMinerRate:    ElectedMinerRewardRate,
		FoundationMinerRate: FoundationTxsRewardRate,

		LeaderRate:              LeaderRewardRate,
		ElectedValidatorsRate:   ElectedValidatorsRewardRate,
		FoundationValidatorRate: FoundationTxsRewardRate,

		OriginElectOfflineRate: OriginElectOfflineRewardRate,
		BackupRewardRate:       BackupRate,
	}

	chaindb := mandb.NewMemDatabase()
	currentState, _ := state.NewStateDBManage(nil, chaindb, state.NewDatabase(chaindb))
	currentState.SetBalance(params.MAN_COIN, common.MainAccount, common.TxGasRewardAddress, new(big.Int).SetUint64(16e18))
	//currentState := &State{balance: 10e18}
	preState := currentState
	ppreState := currentState
	matrixstate.SetVersionInfo(currentState, manparams.VersionAlpha)
	matrixstate.SetVersionInfo(ppreState, manparams.VersionAlpha)
	matrixstate.SetVersionInfo(preState, manparams.VersionAlpha)
	matrixstate.SetTxsCalc(preState, "01")
	matrixstate.SetTxsRewardCfg(preState, &mc.TxsRewardCfg{MinersRate: 10000, ValidatorsRate: 0, RewardRate: rrc})
	matrixstate.SetBroadcastInterval(preState, &mc.BCIntervalInfo{LastBCNumber: 0, LastReelectNumber: 0, BCInterval: 100})
	matrixstate.SetFoundationAccount(preState, common.HexToAddress(testAddress))
	matrixstate.SetInnerMinerAccounts(preState, []common.Address{common.HexToAddress("0x1")})
	bc := &Chain{}
	manparams.SetStateReader(bc)
	Convey("计算交易费", t, func() {
		NodeList := make([]mc.TopologyNodeInfo, 0)
		NodeList = append(NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("11"), Position: 0, Type: common.RoleMiner})
		NodeList = append(NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("12"), Position: 1, Type: common.RoleMiner})
		NodeList = append(NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("13"), Position: 2, Type: common.RoleMiner})
		NodeList = append(NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("14"), Position: 3, Type: common.RoleMiner})
		NodeList = append(NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("15"), Position: 4, Type: common.RoleMiner})
		EleList := make([]mc.ElectNodeInfo, 0)
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress("11"), Position: 0, Type: common.RoleMiner, Stock: 1})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress("12"), Position: 1, Type: common.RoleMiner, Stock: 1})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress("13"), Position: 2, Type: common.RoleMiner, Stock: 1})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress("14"), Position: 3, Type: common.RoleMiner, Stock: 1})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress("15"), Position: 4, Type: common.RoleMiner, Stock: 1})
		bc.SetGraphByState(&mc.TopologyGraph{NodeList: NodeList, CurNodeNumber: 5}, &mc.ElectGraph{Number: 4, ElectList: EleList})
		rewardobject := New(bc, currentState, preState, ppreState)
		out := rewardobject.CalcNodesRewards(new(big.Int).SetUint64(5e18), common.HexToAddress("02"), 1, common.HexToHash("03"), params.MAN_COIN)
		for _, nodelist := range NodeList {
			if mount, ok := out[nodelist.Account]; ok {
				if mount.Uint64() != 6e17 {
					t.Error("金额检查错误")
				}
			} else {
				t.Error("账户不存在错误")
			}

		}
	})
}

func TestCalcInnerMiner0(t *testing.T) {
	log.InitLog(3)

	rrc := mc.RewardRateCfg{
		MinerOutRate:        MinerOutRewardRate,
		ElectedMinerRate:    ElectedMinerRewardRate,
		FoundationMinerRate: FoundationTxsRewardRate,

		LeaderRate:              LeaderRewardRate,
		ElectedValidatorsRate:   ElectedValidatorsRewardRate,
		FoundationValidatorRate: FoundationTxsRewardRate,

		OriginElectOfflineRate: OriginElectOfflineRewardRate,
		BackupRewardRate:       BackupRate,
	}

	chaindb := mandb.NewMemDatabase()
	currentState, _ := state.NewStateDBManage(nil, chaindb, state.NewDatabase(chaindb))
	currentState.SetBalance(params.MAN_COIN, common.MainAccount, common.TxGasRewardAddress, new(big.Int).SetUint64(16e18))
	//currentState := &State{balance: 10e18}
	preState := currentState
	ppreState := currentState
	matrixstate.SetVersionInfo(currentState, manparams.VersionAlpha)
	matrixstate.SetVersionInfo(ppreState, manparams.VersionAlpha)
	matrixstate.SetVersionInfo(preState, manparams.VersionAlpha)
	matrixstate.SetTxsCalc(preState, "01")
	matrixstate.SetTxsRewardCfg(preState, &mc.TxsRewardCfg{MinersRate: 10000, ValidatorsRate: 0, RewardRate: rrc})
	matrixstate.SetBroadcastInterval(preState, &mc.BCIntervalInfo{LastBCNumber: 0, LastReelectNumber: 0, BCInterval: 100})
	matrixstate.SetFoundationAccount(preState, common.HexToAddress(testAddress))
	matrixstate.SetInnerMinerAccounts(ppreState, []common.Address{common.HexToAddress("0x1")})
	matrixstate.SetPreMinerTxsReward(preState, &mc.MinerOutReward{Reward: *new(big.Int).SetUint64(2e18)})
	bc := &Chain{}
	bc.coinBase = common.HexToAddress("0x1")
	manparams.SetStateReader(bc)
	Convey("计算交易费", t, func() {
		NodeList := make([]mc.TopologyNodeInfo, 0)
		NodeList = append(NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("11"), Position: 0, Type: common.RoleMiner})
		NodeList = append(NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("12"), Position: 1, Type: common.RoleMiner})
		NodeList = append(NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("13"), Position: 2, Type: common.RoleMiner})
		NodeList = append(NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("14"), Position: 3, Type: common.RoleMiner})
		NodeList = append(NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("15"), Position: 4, Type: common.RoleMiner})
		EleList := make([]mc.ElectNodeInfo, 0)
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress("11"), Position: 0, Type: common.RoleMiner, Stock: 1})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress("12"), Position: 1, Type: common.RoleMiner, Stock: 1})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress("13"), Position: 2, Type: common.RoleMiner, Stock: 1})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress("14"), Position: 3, Type: common.RoleMiner, Stock: 1})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress("15"), Position: 4, Type: common.RoleMiner, Stock: 1})
		bc.SetGraphByState(&mc.TopologyGraph{NodeList: NodeList, CurNodeNumber: 5}, &mc.ElectGraph{Number: 4, ElectList: EleList})
		rewardobject := New(bc, currentState, preState, ppreState)
		out := rewardobject.CalcNodesRewards(new(big.Int).SetUint64(5e18), common.HexToAddress("02"), 3, common.HexToHash("03"), params.MAN_COIN)
		for _, nodelist := range NodeList {
			if mount, ok := out[nodelist.Account]; ok {
				if mount.Uint64() != 6e17 {
					t.Error("金额检查错误")
				}
			} else {
				t.Error("账户不存在错误")
			}

		}
	})
}

func TestCalcInnerMiner1(t *testing.T) {
	log.InitLog(3)

	rrc := mc.RewardRateCfg{
		MinerOutRate:        MinerOutRewardRate,
		ElectedMinerRate:    ElectedMinerRewardRate,
		FoundationMinerRate: FoundationTxsRewardRate,

		LeaderRate:              LeaderRewardRate,
		ElectedValidatorsRate:   ElectedValidatorsRewardRate,
		FoundationValidatorRate: FoundationTxsRewardRate,

		OriginElectOfflineRate: OriginElectOfflineRewardRate,
		BackupRewardRate:       BackupRate,
	}
	chaindb := mandb.NewMemDatabase()
	currentState, _ := state.NewStateDBManage(nil, chaindb, state.NewDatabase(chaindb))
	currentState.SetBalance(params.MAN_COIN, common.MainAccount, common.TxGasRewardAddress, new(big.Int).SetUint64(16e18))
	//currentState := &State{balance: 10e18}
	preState := currentState
	chaindb2 := mandb.NewMemDatabase()
	ppreState, _ := state.NewStateDBManage(nil, chaindb, state.NewDatabase(chaindb2))

	matrixstate.SetVersionInfo(currentState, manparams.VersionAlpha)
	matrixstate.SetVersionInfo(ppreState, manparams.VersionAlpha)
	matrixstate.SetVersionInfo(preState, manparams.VersionAlpha)
	matrixstate.SetTxsCalc(preState, "01")
	matrixstate.SetTxsRewardCfg(preState, &mc.TxsRewardCfg{MinersRate: 10000, ValidatorsRate: 0, RewardRate: rrc})
	matrixstate.SetBroadcastInterval(preState, &mc.BCIntervalInfo{LastBCNumber: 0, LastReelectNumber: 0, BCInterval: 100})
	matrixstate.SetFoundationAccount(preState, common.HexToAddress(testAddress))
	matrixstate.SetInnerMinerAccounts(ppreState, []common.Address{common.HexToAddress("0x3")})
	matrixstate.SetInnerMinerAccounts(preState, []common.Address{common.HexToAddress("0x11")})
	matrixstate.SetPreMinerTxsReward(preState, &mc.MinerOutReward{Reward: *new(big.Int).SetUint64(2e18)})
	bc := &Chain{}
	bc.coinBase = common.HexToAddress("0x11")
	Convey("计算交易费", t, func() {
		NodeList := make([]mc.TopologyNodeInfo, 0)
		NodeList = append(NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("11"), Position: 0, Type: common.RoleMiner})
		NodeList = append(NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("12"), Position: 1, Type: common.RoleMiner})
		NodeList = append(NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("13"), Position: 2, Type: common.RoleMiner})
		NodeList = append(NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("14"), Position: 3, Type: common.RoleMiner})
		NodeList = append(NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("15"), Position: 4, Type: common.RoleMiner})
		EleList := make([]mc.ElectNodeInfo, 0)
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress("11"), Position: 0, Type: common.RoleMiner, Stock: 1})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress("12"), Position: 1, Type: common.RoleMiner, Stock: 1})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress("13"), Position: 2, Type: common.RoleMiner, Stock: 1})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress("14"), Position: 3, Type: common.RoleMiner, Stock: 1})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress("15"), Position: 4, Type: common.RoleMiner, Stock: 1})
		bc.SetGraphByState(&mc.TopologyGraph{NodeList: NodeList, CurNodeNumber: 5}, &mc.ElectGraph{Number: 4, ElectList: EleList})
		manparams.SetStateReader(bc)
		rewardobject := New(bc, currentState, preState, ppreState)
		out := rewardobject.CalcNodesRewards(new(big.Int).SetUint64(5e18), common.HexToAddress("02"), 3, common.HexToHash("03"), params.MAN_COIN)

		for _, nodelist := range NodeList {
			if nodelist.Account.Equal(common.HexToAddress("11")) {
				if mount, ok := out[nodelist.Account]; ok {
					if mount.Uint64() != 26e17 {
						t.Error("金额检查错误", mount)
					}
				} else {
					t.Error("账户不存在错误")
				}
				continue
			}
			if mount, ok := out[nodelist.Account]; ok {
				if mount.Uint64() != 6e17 {
					t.Error("金额检查错误")
				}
			} else {
				t.Error("账户不存在错误")
			}

		}
	})
}

//节点无顶替
func TestCalcValidator0(t *testing.T) {
	log.InitLog(3)

	rrc := mc.RewardRateCfg{
		MinerOutRate:        MinerOutRewardRate,
		ElectedMinerRate:    ElectedMinerRewardRate,
		FoundationMinerRate: FoundationTxsRewardRate,

		LeaderRate:              LeaderRewardRate,
		ElectedValidatorsRate:   ElectedValidatorsRewardRate,
		FoundationValidatorRate: FoundationTxsRewardRate,

		OriginElectOfflineRate: OriginElectOfflineRewardRate,
		BackupRewardRate:       BackupRate,
	}
	chaindb := mandb.NewMemDatabase()
	currentState, _ := state.NewStateDBManage(nil, chaindb, state.NewDatabase(chaindb))
	currentState.SetBalance(params.MAN_COIN, common.MainAccount, common.TxGasRewardAddress, new(big.Int).SetUint64(16e18))
	//currentState := &State{balance: 10e18}
	preState := currentState
	chaindb2 := mandb.NewMemDatabase()
	ppreState, _ := state.NewStateDBManage(nil, chaindb, state.NewDatabase(chaindb2))

	matrixstate.SetVersionInfo(currentState, manparams.VersionAlpha)
	matrixstate.SetVersionInfo(ppreState, manparams.VersionAlpha)
	matrixstate.SetVersionInfo(preState, manparams.VersionAlpha)
	matrixstate.SetTxsCalc(preState, "01")
	matrixstate.SetTxsRewardCfg(preState, &mc.TxsRewardCfg{MinersRate: 0, ValidatorsRate: 10000, RewardRate: rrc})
	matrixstate.SetBroadcastInterval(preState, &mc.BCIntervalInfo{LastBCNumber: 0, LastReelectNumber: 0, BCInterval: 100})
	matrixstate.SetFoundationAccount(preState, common.HexToAddress(testAddress))
	matrixstate.SetInnerMinerAccounts(ppreState, []common.Address{common.HexToAddress("0x3")})
	matrixstate.SetInnerMinerAccounts(preState, []common.Address{common.HexToAddress("0x11")})
	bc := &Chain{}
	bc.coinBase = common.HexToAddress("0x11")
	Convey("计算交易费", t, func() {
		NodeList := make([]mc.TopologyNodeInfo, 0)
		NodeList = append(NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("11"), Position: 8192, Type: common.RoleValidator})
		NodeList = append(NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("12"), Position: 8193, Type: common.RoleValidator})
		NodeList = append(NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("13"), Position: 8194, Type: common.RoleValidator})
		NodeList = append(NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("14"), Position: 8195, Type: common.RoleValidator})
		NodeList = append(NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("15"), Position: 8196, Type: common.RoleValidator})
		EleList := make([]mc.ElectNodeInfo, 0)
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress("11"), Position: 8192, Type: common.RoleValidator, Stock: 1})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress("12"), Position: 8193, Type: common.RoleValidator, Stock: 1})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress("13"), Position: 8194, Type: common.RoleValidator, Stock: 1})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress("14"), Position: 8195, Type: common.RoleValidator, Stock: 1})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress("15"), Position: 8196, Type: common.RoleValidator, Stock: 1})
		bc.SetGraphByState(&mc.TopologyGraph{NodeList: NodeList, CurNodeNumber: 5}, &mc.ElectGraph{Number: 4, ElectList: EleList})
		manparams.SetStateReader(bc)
		rewardobject := New(bc, currentState, preState, ppreState)
		out := rewardobject.CalcNodesRewards(new(big.Int).SetUint64(5e18), common.HexToAddress("11"), 3, common.HexToHash("03"), params.MAN_COIN)

		for _, nodelist := range NodeList {
			if nodelist.Account.Equal(common.HexToAddress("11")) {
				if mount, ok := out[nodelist.Account]; ok {
					if mount.Uint64() != 26e17 {
						t.Error("金额检查错误", mount)
					}
				} else {
					t.Error("账户不存在错误")
				}
				continue
			}
			if mount, ok := out[nodelist.Account]; ok {
				if mount.Uint64() != 6e17 {
					t.Error("金额检查错误")
				}
			} else {
				t.Error("账户不存在错误")
			}

		}
	})
}

//节点有顶替
func TestCalcValidator1(t *testing.T) {
	log.InitLog(3)

	rrc := mc.RewardRateCfg{
		MinerOutRate:        MinerOutRewardRate,
		ElectedMinerRate:    ElectedMinerRewardRate,
		FoundationMinerRate: FoundationTxsRewardRate,

		LeaderRate:              LeaderRewardRate,
		ElectedValidatorsRate:   ElectedValidatorsRewardRate,
		FoundationValidatorRate: FoundationTxsRewardRate,

		OriginElectOfflineRate: OriginElectOfflineRewardRate,
		BackupRewardRate:       BackupRate,
	}

	chaindb := mandb.NewMemDatabase()
	currentState, _ := state.NewStateDBManage(nil, chaindb, state.NewDatabase(chaindb))
	currentState.SetBalance(params.MAN_COIN, common.MainAccount, common.TxGasRewardAddress, new(big.Int).SetUint64(16e18))
	//currentState := &State{balance: 10e18}
	preState := currentState
	chaindb2 := mandb.NewMemDatabase()
	ppreState, _ := state.NewStateDBManage(nil, chaindb, state.NewDatabase(chaindb2))

	matrixstate.SetVersionInfo(currentState, manparams.VersionAlpha)
	matrixstate.SetVersionInfo(ppreState, manparams.VersionAlpha)
	matrixstate.SetVersionInfo(preState, manparams.VersionAlpha)
	matrixstate.SetTxsCalc(preState, "01")
	matrixstate.SetTxsRewardCfg(preState, &mc.TxsRewardCfg{MinersRate: 0, ValidatorsRate: 10000, RewardRate: rrc})
	matrixstate.SetBroadcastInterval(preState, &mc.BCIntervalInfo{LastBCNumber: 0, LastReelectNumber: 0, BCInterval: 100})
	matrixstate.SetFoundationAccount(preState, common.HexToAddress(testAddress))
	matrixstate.SetInnerMinerAccounts(ppreState, []common.Address{common.HexToAddress("0x3")})
	matrixstate.SetInnerMinerAccounts(preState, []common.Address{common.HexToAddress("0x11")})
	bc := &Chain{}
	bc.coinBase = common.HexToAddress("0x11")
	manparams.SetStateReader(bc)
	Convey("计算交易费", t, func() {
		NodeList := make([]mc.TopologyNodeInfo, 0)
		NodeList = append(NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("11"), Position: 8192, Type: common.RoleValidator})
		NodeList = append(NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("12"), Position: 8193, Type: common.RoleValidator})
		NodeList = append(NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("13"), Position: 8194, Type: common.RoleValidator})
		NodeList = append(NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("14"), Position: 8195, Type: common.RoleValidator})
		NodeList = append(NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("16"), Position: 8196, Type: common.RoleValidator})
		EleList := make([]mc.ElectNodeInfo, 0)
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress("11"), Position: 8192, Type: common.RoleValidator, Stock: 1})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress("12"), Position: 8193, Type: common.RoleValidator, Stock: 1})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress("13"), Position: 8194, Type: common.RoleValidator, Stock: 1})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress("14"), Position: 8195, Type: common.RoleValidator, Stock: 1})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress("15"), Position: 8196, Type: common.RoleValidator, Stock: 1})
		bc.SetGraphByState(&mc.TopologyGraph{NodeList: NodeList, CurNodeNumber: 5}, &mc.ElectGraph{Number: 4, ElectList: EleList})
		rewardobject := New(bc, currentState, preState, ppreState)
		out := rewardobject.CalcNodesRewards(new(big.Int).SetUint64(5e18), common.HexToAddress("11"), 3, common.HexToHash("03"), params.MAN_COIN)

		for _, nodelist := range NodeList {
			if nodelist.Account.Equal(common.HexToAddress("15")) || nodelist.Account.Equal(common.HexToAddress("16")) {
				if mount, ok := out[nodelist.Account]; ok {
					if mount.Uint64() != 3e17 {
						t.Error("金额检查错误", mount)
					}
				} else {
					t.Error("账户不存在错误")
				}
				continue
			}
			if nodelist.Account.Equal(common.HexToAddress("11")) {
				if mount, ok := out[nodelist.Account]; ok {
					if mount.Uint64() != 26e17 {
						t.Error("金额检查错误", mount)
					}
				} else {
					t.Error("账户不存在错误")
				}
				continue
			}
			if mount, ok := out[nodelist.Account]; ok {
				if mount.Uint64() != 6e17 {
					t.Error("金额检查错误")
				}
			} else {
				t.Error("账户不存在错误")
			}

		}
	})
}

//节点完全顶替
func TestCalcValidator2(t *testing.T) {
	log.InitLog(3)

	rrc := mc.RewardRateCfg{
		MinerOutRate:        MinerOutRewardRate,
		ElectedMinerRate:    ElectedMinerRewardRate,
		FoundationMinerRate: FoundationTxsRewardRate,

		LeaderRate:              LeaderRewardRate,
		ElectedValidatorsRate:   ElectedValidatorsRewardRate,
		FoundationValidatorRate: FoundationTxsRewardRate,

		OriginElectOfflineRate: OriginElectOfflineRewardRate,
		BackupRewardRate:       BackupRate,
	}
	//monkey.Patch(manparams.IsBroadcastNumber, func(number uint64, stateNumber uint64) bool {
	//	fmt.Println("use monkey  manparams.IsBroadcastNumber")
	//
	//	return false
	//})
	chaindb := mandb.NewMemDatabase()
	currentState, _ := state.NewStateDBManage(nil, chaindb, state.NewDatabase(chaindb))
	currentState.SetBalance(params.MAN_COIN, common.MainAccount, common.TxGasRewardAddress, new(big.Int).SetUint64(16e18))
	//currentState := &State{balance: 10e18}
	preState := currentState
	chaindb2 := mandb.NewMemDatabase()
	ppreState, _ := state.NewStateDBManage(nil, chaindb, state.NewDatabase(chaindb2))

	matrixstate.SetVersionInfo(currentState, manparams.VersionAlpha)
	matrixstate.SetVersionInfo(ppreState, manparams.VersionAlpha)
	matrixstate.SetVersionInfo(preState, manparams.VersionAlpha)
	matrixstate.SetTxsCalc(preState, "01")
	matrixstate.SetTxsRewardCfg(preState, &mc.TxsRewardCfg{MinersRate: 0, ValidatorsRate: 10000, RewardRate: rrc})
	matrixstate.SetBroadcastInterval(preState, &mc.BCIntervalInfo{LastBCNumber: 0, LastReelectNumber: 0, BCInterval: 100})
	matrixstate.SetFoundationAccount(preState, common.HexToAddress(testAddress))
	matrixstate.SetInnerMinerAccounts(ppreState, []common.Address{common.HexToAddress("0x3")})
	matrixstate.SetInnerMinerAccounts(preState, []common.Address{common.HexToAddress("0x11")})
	bc := &Chain{}
	bc.coinBase = common.HexToAddress("0x11")

	manparams.SetStateReader(bc)
	Convey("计算交易费", t, func() {
		NodeList := make([]mc.TopologyNodeInfo, 0)
		NodeList = append(NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("21"), Position: 8192, Type: common.RoleValidator})
		NodeList = append(NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("22"), Position: 8193, Type: common.RoleValidator})
		NodeList = append(NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("23"), Position: 8194, Type: common.RoleValidator})
		NodeList = append(NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("24"), Position: 8195, Type: common.RoleValidator})
		NodeList = append(NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("25"), Position: 8196, Type: common.RoleValidator})
		EleList := make([]mc.ElectNodeInfo, 0)
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress("11"), Position: 8192, Type: common.RoleValidator, Stock: 1})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress("12"), Position: 8193, Type: common.RoleValidator, Stock: 1})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress("13"), Position: 8194, Type: common.RoleValidator, Stock: 1})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress("14"), Position: 8195, Type: common.RoleValidator, Stock: 1})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress("15"), Position: 8196, Type: common.RoleValidator, Stock: 1})
		bc.SetGraphByState(&mc.TopologyGraph{NodeList: NodeList, CurNodeNumber: 5}, &mc.ElectGraph{Number: 4, ElectList: EleList})
		rewardobject := New(bc, currentState, preState, ppreState)
		out := rewardobject.CalcNodesRewards(new(big.Int).SetUint64(5e18), common.HexToAddress("11"), 3, common.HexToHash("03"), params.MAN_COIN)

		for _, nodelist := range NodeList {

			if mount, ok := out[nodelist.Account]; ok {
				if mount.Uint64() != 3e17 {
					t.Error("金额检查错误")
				}
			} else {
				t.Error("账户不存在错误")
			}

		}
	})
}

func findCommonValue(a []int32, b []int32) []int32 {
	recordMap := make(map[int32]bool)
	result := make([]int32, 0)
	lengthA := len(a)
	lengthB := len(b)
	if lengthA == 0 || lengthB == 0 {
		return result
	}

	for _, v := range a {
		recordMap[v] = true
	}
	for _, v := range b {
		if _, ok := recordMap[v]; ok {
			result = append(result, v)
		}
	}

	return result
}
func Test_commonValue(t *testing.T) {
	a := []int32{1, 10, 20, 40, 60}
	b := []int32{2, 30, 50}
	c := findCommonValue(a, b)
	fmt.Println(c)
}

var dy = []int{-2, -1, 1, 2}
var dx = []int{1, 2, 2, 1}
var n = 5
var m = 5
var out int

type roader struct {
	i int
	j int
}

var bestRoader = []roader{}
var used [][]bool

func search(x int, y int, data [][]int) {
	if x == 0 {
		out = data[x][y]
	}
	if x < 0 || x > n {
		return
	}
	if y < 0 || y > m {
		return
	}
	out = out + data[x][y]
	fmt.Println(x, y, " ")
	if x == n-1 {
		fmt.Println("best")
		for _, v := range bestRoader {
			fmt.Println(v.i, v.j, " ")
		}
		return
	}
	var flag bool
	for i := 0; i < 4; i++ {
		xx := x + dx[i]
		yy := y + dy[i]
		if xx >= 0 && xx < n && yy >= 0 && yy < m {
			//used[xx][yy] = true
			if data[xx][yy]+out < out {
				out = data[x][y] + out
				bestRoader = append(bestRoader, roader{x, y})
			}
			flag = true
			search(xx, yy, data)
			//used[xx][yy] = false
		}
	}
	if flag == false {
		bestRoader = bestRoader[:len(bestRoader)-1]
		return
	}
}

func Test_Search(t *testing.T) {
	data := [][]int{{3, 0, -2, 4, 0}, {-1, 2, -2, 1, 4}, {3, 1, -2, 3, 3}, {2, -4, -3, -3, 2}, {-1, 2, -2, 1, 4}}
	search(0, 3, data)
	//fmt.Println("best")
	//for _, v := range bestRoader {
	//	fmt.Println(v.i, v.j, " ")
	//}

}
