package interest

import (
	"bou.ke/monkey"
	"fmt"
	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/consensus/manash"
	"github.com/matrix/go-matrix/core"
	"github.com/matrix/go-matrix/core/vm"
	"github.com/matrix/go-matrix/depoistInfo"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"math/big"
	"sync"
	"testing"

	"github.com/matrix/go-matrix/common"
	. "github.com/smartystreets/goconvey/convey"
)



type FakeEth struct {
	blockchain *core.BlockChain
	once       *sync.Once
}

func fakeEthNew(n int) *FakeEth {
	fakeman := &FakeEth{once: new(sync.Once)}
	fakeman.once.Do(func() {
		_, blockchain, err := core.NewCanonical(manash.NewFaker(), n, true)
		if err != nil {
			fmt.Println("failed to create pristine chain: ", err)
			return
		}
		defer blockchain.Stop()
		fakeman.blockchain = blockchain
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
		monkey.Patch(ca.GetElectedByHeight, func (height *big.Int) ([]vm.DepositDetail, error) {
			fmt.Println("use monkey  ca.GetElectedByHeight")
			Deposit := make([]vm.DepositDetail, 0)

				Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"), Deposit: new(big.Int).Exp(big.NewInt(10), big.NewInt(17), big.NewInt(0))})
				Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Deposit: new(big.Int).Exp(big.NewInt(10), big.NewInt(21), big.NewInt(0))})
				Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Deposit: new(big.Int).Exp(big.NewInt(10), big.NewInt(22), big.NewInt(0))})
				Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Deposit: new(big.Int).Exp(big.NewInt(10), big.NewInt(23), big.NewInt(0))})
				log.Info(PackageName,"deposit",Deposit[0].Deposit)



			return Deposit, nil
		})
		monkey.Patch(depoistInfo.AddReward, func (stateDB vm.StateDB, address common.Address, reward *big.Int) error  {

			fmt.Println("use monkey  depoistInfo.AddReward","acccount",address.String(),"interest",reward.String())



			return  nil
		})

	})
	return fakeman
}
func Test_interest_Calc(t *testing.T) {
	Convey("利息测试计算利息", t, func() {
		log.InitLog(3)
		eth := fakeEthNew(0)
		interestTest:=New(eth.blockchain)
		state,_:=eth.blockchain.State()
		interestTest.InterestCalc(state,99)

	})

}

func Test_interest_Send(t *testing.T) {
	Convey("利息测试计算利息", t, func() {
		log.InitLog(3)
		eth := fakeEthNew(0)
		interestTest:=New(eth.blockchain)
		state,_:=eth.blockchain.State()
		interestTest.InterestCalc(state,3599)

	})

}

