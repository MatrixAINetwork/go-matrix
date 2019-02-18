package txsreward

import (
	"fmt"
	"math/big"
	"math/rand"
	"testing"

	"github.com/MatrixAINetwork/go-matrix/ca"
	"github.com/MatrixAINetwork/go-matrix/core/vm"
	"github.com/MatrixAINetwork/go-matrix/params/manparams"

	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/params"

	"github.com/MatrixAINetwork/go-matrix/core/matrixstate"

	"bou.ke/monkey"
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
func TestBlockReward_calcTxsFees(t *testing.T) {
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
		if key == mc.MSKeyTxsRewardCfg {

			return &mc.TxsRewardCfg{TxsRewardCalc: "1", MinersRate: 0, ValidatorsRate: 10000, RewardRate: rrc}, nil
		}
		return nil, nil
	})

	monkey.Patch(manparams.NewBCIntervalByNumber, func(blockNumber uint64) (*manparams.BCInterval, error) {
		fmt.Println("use monkey NewBCIntervalByNumber")

		inteval1 := &mc.BCIntervalInfo{LastBCNumber: 0, LastReelectNumber: 0, BCInterval: 100}

		interval2, _ := manparams.NewBCIntervalWithInterval(inteval1)
		return interval2, nil
	})

	bc := &Chain{}
	rewardobject := New(bc, &State{100})
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

//func TestNew2(t *testing.T) {
//	Convey("计算交易费", t, func() {
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
//		reward := New(eth.blockchain)
//		reward.CalcNodesRewards(big.NewInt(0), common.HexToAddress(testAddress), 1)
//	})
//}
//func TestNew3(t *testing.T) {
//	Convey("计算交易费", t, func() {
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
//		reward := New(eth.blockchain)
//		reward.CalcNodesRewards(big.NewInt(-1), common.HexToAddress(testAddress), 1)
//	})
//}
