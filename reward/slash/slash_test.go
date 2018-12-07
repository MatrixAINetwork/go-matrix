package slash

import (
	"fmt"
	"math/big"
	"sync"
	"testing"

	"github.com/matrix/go-matrix/core/state"

	"github.com/matrix/go-matrix/ethdb"

	"github.com/matrix/go-matrix/depoistInfo"

	"bou.ke/monkey"
	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/consensus/ethash"
	"github.com/matrix/go-matrix/core"
	"github.com/matrix/go-matrix/core/vm"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	. "github.com/smartystreets/goconvey/convey"
)

type FakeEth struct {
	blockchain *core.BlockChain
	once       *sync.Once
}

const account0 = "0x475baee143cf541ff3ee7b00c1c933129238d793"
const account1 = "0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"
const account2 = "0x519437b21e2a0b62788ab9235d0728dd7f1a7269"
const account3 = "0x29216818d3788c2505a593cbbb248907d47d9bce"

func (s *FakeEth) BlockChain() *core.BlockChain { return s.blockchain }
func fakeEthNew(n int) *FakeEth {
	eth := &FakeEth{once: new(sync.Once)}
	eth.once.Do(func() {
		_, blockchain, err := core.NewCanonical(ethash.NewFaker(), n, true)
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
				newGraph.NodeList = append(newGraph.NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress(account0), Position: 8192})
				newGraph.NodeList = append(newGraph.NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress(account1), Position: 8193})
				newGraph.NodeList = append(newGraph.NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress(account2), Position: 8194})
				newGraph.NodeList = append(newGraph.NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress(account3), Position: 8195})
				newGraph.CurNodeNumber = 4
			}

			return newGraph, nil
		})

		//id, _ := discover.HexID(myNodeId)
		//ca.Start(id, "")
	})
	return eth
}
func TestBlockSlash_CalcSlash(t *testing.T) {
	log.InitLog(3)
	eth := fakeEthNew(0)
	slash := New(eth.blockchain)
	Convey("计算惩罚", t, func() {
		monkey.Patch(depoistInfo.AddSlash, func(stateDB vm.StateDB, address common.Address, slash *big.Int) error {
			fmt.Println("use monkey  depoistInfo.AddSlash", "address", address.String(), "slash", slash.Uint64())
			return nil
		})
		statedb, _ := state.New(common.Hash{}, state.NewDatabase(ethdb.NewMemDatabase()))
		monkey.Patch(depoistInfo.GetOnlineTime, func(stateDB vm.StateDB, address common.Address) (*big.Int, error) {
			fmt.Println("use monkey  ca.GetOnlineTime")
			onlineTime := big.NewInt(291)
			if stateDB == statedb {
				switch {
				case address.Equal(common.HexToAddress(account0)):
					onlineTime = big.NewInt(291 * 2) //100%
				case address.Equal(common.HexToAddress(account1)):
					onlineTime = big.NewInt(291) //0%
				case address.Equal(common.HexToAddress(account2)):
					onlineTime = big.NewInt(291 + 291/2) //50%
				case address.Equal(common.HexToAddress(account3)):
					onlineTime = big.NewInt(291 + 291/4) //25%

				}

			}

			return onlineTime, nil
		})

		rewards := make(map[common.Address]*big.Int, 0)
		rewards[common.HexToAddress(account0)] = big.NewInt(4e+18)
		rewards[common.HexToAddress(account1)] = big.NewInt(4e+18)
		rewards[common.HexToAddress(account2)] = big.NewInt(4e+18)
		rewards[common.HexToAddress(account3)] = big.NewInt(4e+18)
		slash.CalcSlash(statedb, common.GetReElectionInterval()-1)
	})
}
