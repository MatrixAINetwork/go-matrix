// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package rewardexec

import (
	"fmt"
	"math/big"
	"math/rand"
	"testing"

	"bou.ke/monkey"
	"github.com/MatrixAINetwork/go-matrix/ca"
	"github.com/MatrixAINetwork/go-matrix/core/matrixstate"
	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/core/vm"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/params"
	"github.com/MatrixAINetwork/go-matrix/params/manparams"
	"github.com/MatrixAINetwork/go-matrix/reward/cfg"
	"github.com/MatrixAINetwork/go-matrix/reward/util"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/log"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	testAddress               = "0x8605cdbbdb6d264aa742e77020dcbc58fcdce182"
	MinersBlockRewardRate     = uint64(5000) //矿工网络奖励50%
	ValidatorsBlockRewardRate = uint64(5000) //验证者网络奖励50%

	MinerOutRewardRate        = uint64(4000) //出块矿工奖励40%
	ElectedMinerRewardRate    = uint64(5000) //当选矿工奖励50%
	FoundationMinerRewardRate = uint64(1000) //基金会网络奖励10%

	LeaderRewardRate               = uint64(4000) //出块验证者（leader）奖励40%
	ElectedValidatorsRewardRate    = uint64(5000) //当选验证者奖励60%
	FoundationValidatorsRewardRate = uint64(1000) //基金会网络奖励10%

	OriginElectOfflineRewardRate = uint64(5000) //初选下线验证者奖励50%
	BackupRate                   = uint64(5000) //当前替补验证者奖励50%
)

var myNodeId string = "4b2f638f46c7ae5b1564ca7015d716621848a0d9be66f1d1e91d566d2a70eedc2f11e92b743acb8d97dec3fb412c1b2f66afd7fbb9399d4fb2423619eaa51411"

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
func (chain *Chain) StateAt(root []common.CoinRoot) (*state.StateDB, error) {

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
	return nil, nil
}

func (chain *Chain) GetSuperBlockNum() (uint64, error) {

	return 0, nil
}

type InnerSeed struct {
}

func (s *InnerSeed) GetSeed(num uint64) *big.Int {
	random := rand.New(rand.NewSource(0))
	return new(big.Int).SetUint64(random.Uint64())
}

type State struct {
	balance int64
}

func (st *State) GetBalance(addr common.Address) common.BalanceType {
	return []common.BalanceSlice{{common.MainAccount, big.NewInt(st.balance)}}
}

func (st *State) GetMatrixData(hash common.Hash) (val []byte) {
	return nil

}
func (st *State) SetMatrixData(hash common.Hash, val []byte) {
	return
}

func TestBlockReward_setLeaderRewards(t *testing.T) {

	log.InitLog(3)

	rrc := mc.RewardRateCfg{
		MinerOutRate:        MinerOutRewardRate,
		ElectedMinerRate:    ElectedMinerRewardRate,
		FoundationMinerRate: FoundationMinerRewardRate,

		LeaderRate:              LeaderRewardRate,
		ElectedValidatorsRate:   ElectedValidatorsRewardRate,
		FoundationValidatorRate: FoundationValidatorsRewardRate,

		OriginElectOfflineRate: OriginElectOfflineRewardRate,
		BackupRewardRate:       BackupRate,
	}
	monkey.Patch(manparams.IsBroadcastNumber, func(number uint64, stateNumber uint64) bool {
		fmt.Println("use monkey  manparams.IsBroadcastNumber")

		return false
	})

	monkey.Patch(matrixstate.GetDataByState, func(key string, state matrixstate.StateDB) (interface{}, error) {
		fmt.Println("use monkey  manparams.IsBroadcastNumber")
		if key == mc.MSKeyMatrixAccount {
			return &mc.MatrixSpecialAccounts{FoundationAccount: mc.NodeInfo{Address: common.HexToAddress(testAddress)}}, nil
		}
		return nil, nil
	})
	monkey.Patch(manparams.NewBCIntervalWithInterval, func(interval interface{}) (*manparams.BCInterval, error) {
		fmt.Println("use monkey  manparams.IsBroadcastNumber")

		return nil, nil
	})
	RewardMount := &mc.BlkRewardCfg{MinerMount: 5, MinerAttenuationNum: 10000, ValidatorMount: 5, ValidatorAttenuationNum: 10000, RewardRate: rrc}
	rewardCfg := cfg.New(RewardMount, nil)
	rewardobject := New(&Chain{}, rewardCfg, &State{100})

	Convey("Leader测试0", t, func() {
		rewardobject.rewardCfg.SetReward.SetLeaderRewards(big.NewInt(0), common.HexToAddress(testAddress), 2)
	})

	Convey("Leader测试1", t, func() {

		rewardobject.rewardCfg.SetReward.SetLeaderRewards(big.NewInt(2e18), common.Address{}, 3)
	})

	Convey("Leader测试2", t, func() {

		rewardobject.rewardCfg.SetReward.SetLeaderRewards(big.NewInt(2e18), common.HexToAddress(testAddress), 10)
	})

	Convey("Leader测试3", t, func() {

		rewardobject.rewardCfg.SetReward.SetLeaderRewards(big.NewInt(-1), common.HexToAddress(testAddress), 100)
	})

	Convey("Leader测试4", t, func() {

		rewardobject.rewardCfg.SetReward.SetLeaderRewards(big.NewInt(100), common.HexToAddress(testAddress), 101)
	})

	Convey("Leader测试5", t, func() {

		rewardobject.rewardCfg.SetReward.SetLeaderRewards(big.NewInt(100), common.HexToAddress(testAddress), 99)
	})
}

func TestBlockReward_setMinerOut(t *testing.T) {
	log.InitLog(3)

	rrc := mc.RewardRateCfg{
		MinerOutRate:        MinerOutRewardRate,
		ElectedMinerRate:    ElectedMinerRewardRate,
		FoundationMinerRate: FoundationMinerRewardRate,

		LeaderRate:              LeaderRewardRate,
		ElectedValidatorsRate:   ElectedValidatorsRewardRate,
		FoundationValidatorRate: FoundationValidatorsRewardRate,

		OriginElectOfflineRate: OriginElectOfflineRewardRate,
		BackupRewardRate:       BackupRate,
	}
	monkey.Patch(manparams.IsBroadcastNumber, func(number uint64, stateNumber uint64) bool {
		fmt.Println("use monkey  manparams.IsBroadcastNumber")

		return false
	})

	monkey.Patch(matrixstate.GetDataByState, func(key string, state matrixstate.StateDB) (interface{}, error) {
		fmt.Println("use monkey  manparams.IsBroadcastNumber")
		if key == mc.MSKeyMatrixAccount {
			return &mc.MatrixSpecialAccounts{FoundationAccount: mc.NodeInfo{Address: common.HexToAddress(testAddress)}}, nil
		}
		if key == mc.MSKeyBroadcastInterval {
			return &mc.BCIntervalInfo{LastBCNumber: 0, LastReelectNumber: 0, BCInterval: 100}, nil
		}
		return nil, nil
	})
	//monkey.Patch(manparams.NewBCIntervalWithInterval, func(interval interface{}) (*manparams.BCInterval, error) {
	//	fmt.Println("use monkey  manparams.IsBroadcastNumber")
	//
	//	return nil, nil
	//})
	//
	monkey.Patch(manparams.NewBCIntervalByNumber, func(blockNumber uint64) (*manparams.BCInterval, error) {
		fmt.Println("use monkey NewBCIntervalByNumber")

		inteval1 := &mc.BCIntervalInfo{LastBCNumber: 0, LastReelectNumber: 0, BCInterval: 100}

		interval2, _ := manparams.NewBCIntervalWithInterval(inteval1)
		return interval2, nil
	})
	RewardMount := &mc.BlkRewardCfg{MinerMount: 5, MinerAttenuationNum: 10000, ValidatorMount: 5, ValidatorAttenuationNum: 10000, RewardRate: rrc}
	rewardCfg := cfg.New(RewardMount, nil)
	rewardobject := New(&Chain{}, rewardCfg, &State{100})

	Convey("挖矿测试0", t, func() {

		rewardobject.rewardCfg.SetReward.SetMinerOutRewards(big.NewInt(0), &State{1000}, nil, 2)
	})

	Convey("挖矿测试1", t, func() {

		rewardobject.rewardCfg.SetReward.SetMinerOutRewards(big.NewInt(2), &State{1000}, nil, 1)
	})

	Convey("挖矿测试高度错误", t, func() {
		rewardobject.rewardCfg.SetReward.SetMinerOutRewards(big.NewInt(2), &State{1000}, nil, 100)
	})
	Convey("挖矿账户nil", t, func() {

		rewardobject.rewardCfg.SetReward.SetMinerOutRewards(big.NewInt(2), &State{1000}, nil, 2)
	})
}

//
func TestBlockReward_setSelectedBlockRewards(t *testing.T) {
	log.InitLog(3)

	rrc := mc.RewardRateCfg{
		MinerOutRate:        MinerOutRewardRate,
		ElectedMinerRate:    ElectedMinerRewardRate,
		FoundationMinerRate: FoundationMinerRewardRate,

		LeaderRate:              LeaderRewardRate,
		ElectedValidatorsRate:   ElectedValidatorsRewardRate,
		FoundationValidatorRate: FoundationValidatorsRewardRate,

		OriginElectOfflineRate: OriginElectOfflineRewardRate,
		BackupRewardRate:       BackupRate,
	}
	monkey.Patch(manparams.IsBroadcastNumber, func(number uint64, stateNumber uint64) bool {
		fmt.Println("use monkey  manparams.IsBroadcastNumber")

		return false
	})

	monkey.Patch(matrixstate.GetDataByState, func(key string, state matrixstate.StateDB) (interface{}, error) {
		fmt.Println("use monkey  manparams.IsBroadcastNumber")
		if key == mc.MSKeyPreMiner {
			return &mc.PreMinerStruct{PreMiner: common.HexToAddress(testAddress)}, nil
		}
		if key == mc.MSKeyMatrixAccount {
			return &mc.MatrixSpecialAccounts{FoundationAccount: mc.NodeInfo{Address: common.HexToAddress(testAddress)}}, nil
		}
		if key == mc.MSKeyBroadcastInterval {
			return &mc.BCIntervalInfo{LastBCNumber: 0, LastReelectNumber: 0, BCInterval: 100}, nil
		}
		return nil, nil
	})

	monkey.Patch(manparams.NewBCIntervalByNumber, func(blockNumber uint64) (*manparams.BCInterval, error) {
		fmt.Println("use monkey NewBCIntervalByNumber")

		inteval1 := &mc.BCIntervalInfo{LastBCNumber: 0, LastReelectNumber: 0, BCInterval: 100}

		interval2, _ := manparams.NewBCIntervalWithInterval(inteval1)
		return interval2, nil
	})
	RewardMount := &mc.BlkRewardCfg{MinerMount: 5, MinerAttenuationNum: 10000, ValidatorMount: 5, ValidatorAttenuationNum: 10000, RewardRate: rrc}
	rewardCfg := cfg.New(RewardMount, nil)
	bc := &Chain{}
	rewardobject := New(bc, rewardCfg, &State{100})
	a := new(big.Int).Exp(big.NewInt(10), big.NewInt(24), big.NewInt(0))
	fmt.Printf("%#x", a)
	SkipConvey("选中无节点变化测试", t, func() {

		NodeList := make([]mc.TopologyNodeInfo, 0)
		NodeList = append(NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"), Position: 8192})
		NodeList = append(NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Position: 8193})
		NodeList = append(NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Position: 8194})
		//NodeList = append(NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Position: 8195})

		EleList := make([]mc.ElectNodeInfo, 0)
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"), Position: 8192})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Position: 8193})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Position: 8194})
		//EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Position: 8195})
		bc.SetGraphByState(&mc.TopologyGraph{NodeList: NodeList, CurNodeNumber: 4}, &mc.ElectGraph{Number: 4, ElectList: EleList})
		monkey.Patch(ca.GetElectedByHeightAndRole, func(height *big.Int, roleType common.RoleType) ([]vm.DepositDetail, error) {
			fmt.Println("use monkey  ca.GetElectedByHeightAndRole")
			Deposit := make([]vm.DepositDetail, 0)
			if common.RoleValidator == roleType&common.RoleValidator {
				Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Deposit: big.NewInt(1e+18)})
				Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Deposit: big.NewInt(3e+18)})
				//Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Deposit: big.NewInt(4e+18)})
				Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"), Deposit: big.NewInt(2e+18)})

			}

			return Deposit, nil
		})
		rewardobject.rewardCfg.SetReward.GetSelectedRewards(big.NewInt(11e+17), &State{1000}, bc, common.RoleValidator|common.RoleBackupValidator, 3, rewardCfg.RewardMount.RewardRate.BackupRewardRate)

		//So(rewards[common.HexToAddress(testAddress)], ShouldEqual, reward.leaderBlkReward)
	})

	SkipConvey("选中节点全部替换测试", t, func() {

		NodeList := make([]mc.TopologyNodeInfo, 0)
		NodeList = append(NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7260"), Position: 8195})
		NodeList = append(NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7261"), Position: 8196})
		NodeList = append(NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7262"), Position: 8197})

		EleList := make([]mc.ElectNodeInfo, 0)
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"), Position: 8192})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Position: 8193})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Position: 8194})

		bc.SetGraphByState(&mc.TopologyGraph{NodeList: NodeList, CurNodeNumber: 4}, &mc.ElectGraph{Number: 4, ElectList: EleList})
		monkey.Patch(ca.GetElectedByHeightAndRole, func(height *big.Int, roleType common.RoleType) ([]vm.DepositDetail, error) {
			fmt.Println("use monkey  ca.GetElectedByHeightAndRole")
			Deposit := make([]vm.DepositDetail, 0)
			if common.RoleValidator == roleType&common.RoleValidator {
				Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Deposit: new(big.Int).Mul(big.NewInt(10000), util.ManPrice)})
				Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Deposit: new(big.Int).Mul(big.NewInt(10000), util.ManPrice)})
				Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"), Deposit: new(big.Int).Mul(big.NewInt(10000), util.ManPrice)})

				Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7260"), Deposit: new(big.Int).Mul(big.NewInt(10000), util.ManPrice)})
				Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7261"), Deposit: new(big.Int).Mul(big.NewInt(10000), util.ManPrice)})
				Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7262"), Deposit: new(big.Int).Mul(big.NewInt(10000), util.ManPrice)})
			}

			return Deposit, nil
		})
		rewardobject.rewardCfg.SetReward.GetSelectedRewards(big.NewInt(11e+17), &State{1000}, bc, common.RoleValidator|common.RoleBackupValidator, 3, rewardCfg.RewardMount.RewardRate.BackupRewardRate)

		//So(rewards[common.HexToAddress(testAddress)], ShouldEqual, reward.leaderBlkReward)
	})
	SkipConvey("选中有节点部分变化测试", t, func() {

		NodeList := make([]mc.TopologyNodeInfo, 0)
		NodeList = append(NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7260"), Position: 8192})
		NodeList = append(NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7261"), Position: 8193})
		NodeList = append(NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7262"), Position: 8194})

		EleList := make([]mc.ElectNodeInfo, 0)
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"), Position: 8192})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7261"), Position: 8193})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7262"), Position: 8194})

		bc.SetGraphByState(&mc.TopologyGraph{NodeList: NodeList, CurNodeNumber: 4}, &mc.ElectGraph{Number: 4, ElectList: EleList})
		monkey.Patch(ca.GetElectedByHeightAndRole, func(height *big.Int, roleType common.RoleType) ([]vm.DepositDetail, error) {
			fmt.Println("use monkey  ca.GetElectedByHeightAndRole")
			Deposit := make([]vm.DepositDetail, 0)
			if common.RoleValidator == roleType&common.RoleValidator {
				Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Deposit: new(big.Int).Mul(big.NewInt(20000), util.ManPrice)})
				Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Deposit: new(big.Int).Mul(big.NewInt(10000), util.ManPrice)})
				Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"), Deposit: new(big.Int).Mul(big.NewInt(10000), util.ManPrice)})

				Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7260"), Deposit: new(big.Int).Mul(big.NewInt(40000), util.ManPrice)})
				Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7261"), Deposit: new(big.Int).Mul(big.NewInt(10000), util.ManPrice)})
				Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7262"), Deposit: new(big.Int).Mul(big.NewInt(10000), util.ManPrice)})
			}

			return Deposit, nil
		})
		rewardobject.rewardCfg.SetReward.GetSelectedRewards(big.NewInt(11e+17), &State{1000}, bc, common.RoleValidator|common.RoleBackupValidator, 3, rewardCfg.RewardMount.RewardRate.BackupRewardRate)
	})
	//
	Convey("奖励金额0", t, func() {
		rewardobject.rewardCfg.SetReward.GetSelectedRewards(big.NewInt(0), &State{1000}, bc, common.RoleValidator|common.RoleBackupValidator, 3, rewardCfg.RewardMount.RewardRate.BackupRewardRate)
		//So(rewards[common.HexToAddress(testAddress)], ShouldEqual, reward.leaderBlkReward)
	})
	//
	Convey("奖励金额小于0", t, func() {

		rewardobject.rewardCfg.SetReward.GetSelectedRewards(big.NewInt(-1), &State{1000}, bc, common.RoleValidator|common.RoleBackupValidator, 3, rewardCfg.RewardMount.RewardRate.BackupRewardRate)
		//So(rewards[common.HexToAddress(testAddress)], ShouldEqual, reward.leaderBlkReward)
	})
	//
	Convey("抵押列表为空", t, func() {
		monkey.Patch(ca.GetElectedByHeightAndRole, func(height *big.Int, roleType common.RoleType) ([]vm.DepositDetail, error) {
			fmt.Println("use monkey  ca.GetElectedByHeightAndRole")

			return nil, nil
		})
		rewardobject.rewardCfg.SetReward.GetSelectedRewards(big.NewInt(5e+18), &State{1000}, bc, common.RoleValidator|common.RoleBackupValidator, 3, rewardCfg.RewardMount.RewardRate.BackupRewardRate)
	})

	Convey("原始拓扑图为空", t, func() {

		rewardobject.rewardCfg.SetReward.GetSelectedRewards(big.NewInt(5e+18), &State{1000}, bc, common.RoleValidator|common.RoleBackupValidator, 3, rewardCfg.RewardMount.RewardRate.BackupRewardRate)
		//So(rewards[common.HexToAddress(testAddress)], ShouldEqual, reward.leaderBlkReward)
	})

	Convey("抵押值非法", t, func() {

		NodeList := make([]mc.TopologyNodeInfo, 0)
		NodeList = append(NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7260"), Position: 8192})
		NodeList = append(NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7261"), Position: 8193})
		NodeList = append(NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7262"), Position: 8194})

		EleList := make([]mc.ElectNodeInfo, 0)
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"), Position: 8192})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7261"), Position: 8193})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7262"), Position: 8194})

		bc.SetGraphByState(&mc.TopologyGraph{NodeList: NodeList, CurNodeNumber: 4}, &mc.ElectGraph{Number: 4, ElectList: EleList})
		monkey.Patch(ca.GetElectedByHeightAndRole, func(height *big.Int, roleType common.RoleType) ([]vm.DepositDetail, error) {
			fmt.Println("use monkey  ca.GetElectedByHeightAndRole")
			Deposit := make([]vm.DepositDetail, 0)
			if common.RoleValidator == roleType&common.RoleValidator {
				Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Deposit: new(big.Int).Mul(big.NewInt(-20000), util.ManPrice)})
				Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Deposit: new(big.Int).Mul(big.NewInt(10000), util.ManPrice)})
				Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"), Deposit: new(big.Int).Mul(big.NewInt(10000), util.ManPrice)})

				Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7260"), Deposit: new(big.Int).Mul(big.NewInt(-40000), util.ManPrice)})
				Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7261"), Deposit: new(big.Int).Mul(big.NewInt(10000), util.ManPrice)})
				Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7262"), Deposit: new(big.Int).Mul(big.NewInt(10000), util.ManPrice)})
			}

			return Deposit, nil
		})
		rewardobject.rewardCfg.SetReward.GetSelectedRewards(big.NewInt(11e+17), &State{1000}, bc, common.RoleValidator|common.RoleBackupValidator, 3, rewardCfg.RewardMount.RewardRate.BackupRewardRate)
	})
}

func TestBlockReward_calcTxsFees(t *testing.T) {
	log.InitLog(3)

	rrc := mc.RewardRateCfg{
		MinerOutRate:        MinerOutRewardRate,
		ElectedMinerRate:    ElectedMinerRewardRate,
		FoundationMinerRate: FoundationMinerRewardRate,

		LeaderRate:              LeaderRewardRate,
		ElectedValidatorsRate:   ElectedValidatorsRewardRate,
		FoundationValidatorRate: FoundationValidatorsRewardRate,

		OriginElectOfflineRate: OriginElectOfflineRewardRate,
		BackupRewardRate:       BackupRate,
	}
	monkey.Patch(manparams.IsBroadcastNumber, func(number uint64, stateNumber uint64) bool {
		fmt.Println("use monkey  manparams.IsBroadcastNumber")

		return false
	})

	monkey.Patch(matrixstate.GetDataByState, func(key string, state matrixstate.StateDB) (interface{}, error) {
		fmt.Println("use monkey  manparams.IsBroadcastNumber")
		if key == mc.MSKeyPreMiner {
			return &mc.PreMinerStruct{PreMiner: common.HexToAddress(testAddress)}, nil
		}
		if key == mc.MSKeyMatrixAccount {
			return &mc.MatrixSpecialAccounts{FoundationAccount: mc.NodeInfo{Address: common.HexToAddress(testAddress)}}, nil
		}
		if key == mc.MSKeyBroadcastInterval {
			return &mc.BCIntervalInfo{LastBCNumber: 0, LastReelectNumber: 0, BCInterval: 100}, nil
		}
		return nil, nil
	})

	monkey.Patch(manparams.NewBCIntervalByNumber, func(blockNumber uint64) (*manparams.BCInterval, error) {
		fmt.Println("use monkey NewBCIntervalByNumber")

		inteval1 := &mc.BCIntervalInfo{LastBCNumber: 0, LastReelectNumber: 0, BCInterval: 100}

		interval2, _ := manparams.NewBCIntervalWithInterval(inteval1)
		return interval2, nil
	})
	RewardMount := &mc.BlkRewardCfg{MinerMount: 5, MinerAttenuationNum: 10000, ValidatorMount: 5, ValidatorAttenuationNum: 10000, RewardRate: rrc}
	rewardCfg := cfg.New(RewardMount, nil)
	bc := &Chain{}
	rewardCfg.MinersRate = 0
	rewardCfg.ValidatorsRate = 10000
	rewardobject := New(bc, rewardCfg, &State{100})
	Convey("计算交易费", t, func() {
		NodeList := make([]mc.TopologyNodeInfo, 0)
		NodeList = append(NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"), Position: 8192, Type: common.RoleValidator})
		NodeList = append(NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Position: 8193, Type: common.RoleValidator})
		NodeList = append(NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Position: 8194, Type: common.RoleValidator})
		//NodeList = append(NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Position: 8195})

		EleList := make([]mc.ElectNodeInfo, 0)
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"), Position: 8192, Type: common.RoleValidator})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Position: 8193, Type: common.RoleValidator})
		EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Position: 8194, Type: common.RoleValidator})
		//EleList = append(EleList, mc.ElectNodeInfo{Account: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Position: 8195})
		bc.SetGraphByState(&mc.TopologyGraph{NodeList: NodeList, CurNodeNumber: 4}, &mc.ElectGraph{Number: 4, ElectList: EleList})
		monkey.Patch(ca.GetElectedByHeightAndRole, func(height *big.Int, roleType common.RoleType) ([]vm.DepositDetail, error) {
			fmt.Println("use monkey  ca.GetElectedByHeightAndRole")
			Deposit := make([]vm.DepositDetail, 0)
			if common.RoleValidator == roleType&common.RoleValidator {
				Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Deposit: new(big.Int).Mul(big.NewInt(10000), util.ManPrice)})
				Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Deposit: new(big.Int).Mul(big.NewInt(10000), util.ManPrice)})
				Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"), Deposit: new(big.Int).Mul(big.NewInt(10000), util.ManPrice)})

			}

			return Deposit, nil
		})
		rewardobject.CalcNodesRewards(util.ByzantiumTxsRewardDen, common.HexToAddress(testAddress), 3)
	})
}

//
//func TestBlockReward_CalcRewardMountByNumber(t *testing.T) {
//	Convey("初始奖励金额小于0", t, func() {
//
//		log.InitLog(3)
//		eth := fakeEthNew(0)
//		header := eth.BlockChain().CurrentHeader()
//		newheader := types.CopyHeader(header)
//		newheader.Number = big.NewInt(1)
//		newheader.NetTopology.Type = common.NetTopoTypeAll
//		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"), Position: 8192})
//		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Position: 8193})
//		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Position: 8194})
//		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Position: 8195})
//		rewardCfg := cfg.New(nil, nil)
//		reward := New(eth.blockchain, rewardCfg)
//		//reward.CalcNodesRewards(util.ByzantiumTxsRewardDen, common.HexToAddress(testAddress), header)
//		reward.CalcRewardMountByNumber(&State{100}, 2, big.NewInt(-1), 100, common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"))
//	})
//
//	Convey("发放余额等于0", t, func() {
//
//		log.InitLog(3)
//		eth := fakeEthNew(0)
//		header := eth.BlockChain().CurrentHeader()
//		newheader := types.CopyHeader(header)
//		newheader.Number = big.NewInt(1)
//		newheader.NetTopology.Type = common.NetTopoTypeAll
//		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"), Position: 8192})
//		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Position: 8193})
//		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Position: 8194})
//		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Position: 8195})
//		rewardCfg := cfg.New(nil, nil)
//		reward := New(eth.blockchain, rewardCfg)
//		//reward.CalcNodesRewards(util.ByzantiumTxsRewardDen, common.HexToAddress(testAddress), header)
//		state00, _ := eth.blockchain.State()
//		reward.CalcRewardMountByNumber(state00, 2, big.NewInt(0), 100, common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"))
//	})
//
//	Convey("状态树为nil", t, func() {
//
//		log.InitLog(3)
//		eth := fakeEthNew(0)
//		header := eth.BlockChain().CurrentHeader()
//		newheader := types.CopyHeader(header)
//		newheader.Number = big.NewInt(1)
//		newheader.NetTopology.Type = common.NetTopoTypeAll
//		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"), Position: 8192})
//		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Position: 8193})
//		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Position: 8194})
//		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Position: 8195})
//		rewardCfg := cfg.New(nil, nil)
//		reward := New(eth.blockchain, rewardCfg)
//		//reward.CalcNodesRewards(util.ByzantiumTxsRewardDen, common.HexToAddress(testAddress), header)
//		reward.CalcRewardMountByNumber(nil, 2, big.NewInt(100), 100, common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"))
//	})
//
//	Convey("账户余额为0", t, func() {
//
//		log.InitLog(3)
//		eth := fakeEthNew(0)
//		header := eth.BlockChain().CurrentHeader()
//		newheader := types.CopyHeader(header)
//		newheader.Number = big.NewInt(1)
//		newheader.NetTopology.Type = common.NetTopoTypeAll
//		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"), Position: 8192})
//		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Position: 8193})
//		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Position: 8194})
//		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Position: 8195})
//		rewardCfg := cfg.New(nil, nil)
//		reward := New(eth.blockchain, rewardCfg)
//		//reward.CalcNodesRewards(util.ByzantiumTxsRewardDen, common.HexToAddress(testAddress), header)
//		reward.CalcRewardMountByNumber(&State{0}, 2, big.NewInt(100), 100, common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"))
//	})
//
//	Convey("账户余额为负值", t, func() {
//
//		log.InitLog(3)
//		eth := fakeEthNew(0)
//		header := eth.BlockChain().CurrentHeader()
//		newheader := types.CopyHeader(header)
//		newheader.Number = big.NewInt(1)
//		newheader.NetTopology.Type = common.NetTopoTypeAll
//		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"), Position: 8192})
//		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Position: 8193})
//		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Position: 8194})
//		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Position: 8195})
//		rewardCfg := cfg.New(nil, nil)
//		reward := New(eth.blockchain, rewardCfg)
//		//reward.CalcNodesRewards(util.ByzantiumTxsRewardDen, common.HexToAddress(testAddress), header)
//		reward.CalcRewardMountByNumber(&State{-1000}, 2, big.NewInt(100), 100, common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"))
//	})
//}
