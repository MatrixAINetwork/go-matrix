// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package reward

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"math/rand"
	"testing"
)

//const (
//	testAddress = "0x8605cdbbdb6d264aa742e77020dcbc58fcdce182"
//)
//
//var myNodeId string = "4b2f638f46c7ae5b1564ca7015d716621848a0d9be66f1d1e91d566d2a70eedc2f11e92b743acb8d97dec3fb412c1b2f66afd7fbb9399d4fb2423619eaa51411"
//
//type FakeEth struct {
//	blockchain *core.BlockChain
//	once       *sync.Once
//}
//
//func (s *FakeEth) BlockChain() *core.BlockChain { return s.blockchain }
//
//func fakeEthNew(n int) *FakeEth {
//	eth := &FakeEth{once: new(sync.Once)}
//	eth.once.Do(func() {
//		_, blockchain, err := core.NewCanonical(manash.NewFaker(), n, true)
//		if err != nil {
//			fmt.Println("failed to create pristine chain: ", err)
//			return
//		}
//		defer blockchain.Stop()
//		eth.blockchain = blockchain
//		monkey.Patch(ca.GetTopologyByNumber, func(reqTypes common.RoleType, number uint64) (*mc.TopologyGraph, error) {
//			fmt.Println("use monkey  ca.GetTopologyByNumber")
//			newGraph := &mc.TopologyGraph{
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
//		monkey.Patch(ca.GetElectedByHeightAndRole, func(height *big.Int, roleType common.RoleType) ([]vm.DepositDetail, error) {
//			fmt.Println("use monkey  ca.GetTopologyByNumber")
//			Deposit := make([]vm.DepositDetail, 0)
//			if common.RoleValidator == roleType&common.RoleValidator {
//				Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"), Deposit: big.NewInt(1e+18)})
//				Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Deposit: big.NewInt(2e+18)})
//				Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Deposit: big.NewInt(3e+18)})
//				Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Deposit: big.NewInt(4e+18)})
//				Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bcf"), Deposit: big.NewInt(2e+18)})
//
//			}
//
//			return Deposit, nil
//		})
//		//id, _ := discover.HexID(myNodeId)
//		//ca.Start(id, "")
//	})
//	return eth
//}

//type InnerSeed struct {
//}
//
//func (s *InnerSeed) GetSeed(num uint64) *big.Int {
//	random := rand.New(rand.NewSource(0))
//	return new(big.Int).SetUint64(random.Uint64())
//}
//func TestNew(t *testing.T) {
//	type args struct {
//		chain reward.ChainReader
//	}
//	log.InitLog(3)
//	eth := fakeEthNew(0)
//	blkreward.New(eth.blockchain)
//
//}
//
//func TestBlockReward_setLeaderRewards(t *testing.T) {
//
//	log.InitLog(3)
//	eth := fakeEthNew(0)
//	rewardCfg := cfg.New(nil, nil)
//	rewardobject := reward.New(eth.blockchain, rewardCfg)
//	Convey("Leader测试", t, func() {
//		rewards := make(map[common.Address]*big.Int, 0)
//		validatorsBlkReward := util.CalcRateReward(ByzantiumBlockReward, rewardobject.rewardCfg.RewardMount.ValidatorsRate)
//		leaderBlkReward := util.CalcRateReward(validatorsBlkReward, rewardobject.rewardCfg.RewardMount.LeaderRate)
//		rewardobject.rewardCfg.SetReward.SetLeaderRewards(leaderBlkReward, rewards, common.HexToAddress(testAddress), new(big.Int).SetUint64(uint64(1)))
//
//	})
//}
//
//func TestBlockReward_setSelectedBlockRewards(t *testing.T) {
//	type args struct {
//		chain ChainReader
//	}
//	log.InitLog(3)
//	eth := fakeEthNew(0)
//	slash := slash.New(eth.blockchain)
//	seed := &InnerSeed{}
//	lottery := lottery.New(eth.blockchain, seed)
//	reward := New(eth.blockchain, lottery, slash)
//	SkipConvey("选中无节点变化测试", t, func() {
//
//		rewards := make(map[common.Address]*big.Int, 0)
//		header := eth.BlockChain().CurrentHeader()
//		newheader := types.CopyHeader(header)
//		newheader.Number = big.NewInt(1)
//		newheader.NetTopology.Type = common.NetTopoTypeAll
//		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"), Position: 8192})
//		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Position: 8193})
//		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Position: 8194})
//		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Position: 8195})
//		reward.setSelectedBlockRewards(reward.electedValidatorsReward, rewards, common.RoleValidator|common.RoleBackupValidator, newheader, BackupRewardRate)
//		//So(rewards[common.HexToAddress(testAddress)], ShouldEqual, reward.leaderBlkReward)
//	})
//
//	Convey("选中有节点变化测试", t, func() {
//
//		rewards := make(map[common.Address]*big.Int, 0)
//		header := eth.BlockChain().CurrentHeader()
//		newheader := types.CopyHeader(header)
//		newheader.Number = big.NewInt(1)
//		newheader.NetTopology.Type = common.NetTopoTypeAll
//		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"), Position: 8192})
//		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Position: 8193})
//		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Position: 8194})
//		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bcf"), Position: 8195})
//		reward.setSelectedBlockRewards(reward.electedValidatorsReward, rewards, common.RoleValidator|common.RoleBackupValidator, newheader, BackupRewardRate)
//		//So(rewards[common.HexToAddress(testAddress)], ShouldEqual, reward.leaderBlkReward)
//	})
//}
//
//func TestBlockReward_calcTxsFees(t *testing.T) {
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
//		txsReward := txsreward.New(eth.BlockChain())
//		txsReward.CalcNodesRewards(util.ByzantiumTxsRewardDen, common.HexToAddress(testAddress), 1)
//
//	})
//}

func Test(t *testing.T) {
	sbs := uint64(100)
	s1 := make([]byte, 8)

	buf := bytes.NewBuffer(s1)
	binary.Write(buf, binary.BigEndian, sbs)
	fmt.Println("buf ", buf.Bytes())

	buf2 := bytes.NewBuffer(buf.Bytes())
	test := make(map[int]int)
	test = nil
	binary.Read(buf2, binary.BigEndian, sbs)
	fmt.Println("超级区块序号", len(test))

	rand.Seed(-4522357180955875538)
	for i := 0; i < 100; i++ {
		fmt.Println("随机数1", rand.Uint64())
	}

	rand.Seed(-7711789739634500565)
	for i := 0; i < 100; i++ {
		fmt.Println("随机数2", rand.Uint64())
	}
}
func TestPower(t *testing.T) {

	data := make([]uint64, 10000)
	for i := 1; i < 10000; i++ {
		data[i] = uint64(i)
	}

	for i := 1; i < 10000; i++ {
		math.Pow(float64(i), 1.3)
	}
}
