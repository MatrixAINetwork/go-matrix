package txsreward

import (
	"bou.ke/monkey"
	"fmt"
	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/consensus/manash"
	"github.com/matrix/go-matrix/core"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/core/vm"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/reward/util"
	. "github.com/smartystreets/goconvey/convey"
	"math/big"
	"sync"
	"testing"
)

type FakeEth struct {
	blockchain *core.BlockChain
	once       *sync.Once
}

const (
	testAddress = "0x8605cdbbdb6d264aa742e77020dcbc58fcdce182"
)

func (s *FakeEth) BlockChain() *core.BlockChain { return s.blockchain }

func fakeEthNew(n int) *FakeEth {
	eth := &FakeEth{once: new(sync.Once)}
	eth.once.Do(func() {
		_, blockchain, err := core.NewCanonical(manash.NewFaker(), n, true)
		if err != nil {
			fmt.Println("failed to create pristine chain: ", err)
			return
		}
		defer blockchain.Stop()
		eth.blockchain = blockchain
		monkey.Patch(ca.GetTopologyByNumber, func(reqTypes common.RoleType, number uint64) (*mc.TopologyGraph, error) {
			fmt.Println("use monkey  ca.GetTopologyByNumber")
			newGraph := &mc.TopologyGraph{
				Number:        number,
				NodeList:      make([]mc.TopologyNodeInfo, 0),
				CurNodeNumber: 0,
			}
			if common.RoleValidator == reqTypes&common.RoleValidator {
				newGraph.NodeList = append(newGraph.NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"), Position: 8192})
				newGraph.NodeList = append(newGraph.NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Position: 8193})
				newGraph.NodeList = append(newGraph.NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Position: 8194})
				newGraph.NodeList = append(newGraph.NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Position: 8195})
				newGraph.CurNodeNumber = 4
			}

			return newGraph, nil
		})
		monkey.Patch(ca.GetElectedByHeightAndRole, func(height *big.Int, roleType common.RoleType) ([]vm.DepositDetail, error) {
			fmt.Println("use monkey  ca.GetTopologyByNumber")
			Deposit := make([]vm.DepositDetail, 0)
			if common.RoleValidator == roleType&common.RoleValidator {
				Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"), Deposit: big.NewInt(1e+18)})
				Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Deposit: big.NewInt(2e+18)})
				Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Deposit: big.NewInt(3e+18)})
				Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Deposit: big.NewInt(4e+18)})
				Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bcf"), Deposit: big.NewInt(2e+18)})

			}

			return Deposit, nil
		})

	})
	return eth
}

func TestNew(t *testing.T) {
	Convey("计算交易费", t, func() {

		log.InitLog(3)
		eth := fakeEthNew(0)
		header := eth.BlockChain().CurrentHeader()
		newheader := types.CopyHeader(header)
		newheader.Number = big.NewInt(1)
		newheader.NetTopology.Type = common.NetTopoTypeAll
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"), Position: 8192})
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Position: 8193})
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Position: 8194})
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Position: 8195})
		reward := New(eth.blockchain)
		reward.CalcNodesRewards(util.ByzantiumTxsRewardDen, common.HexToAddress(testAddress), header)
	})
}
