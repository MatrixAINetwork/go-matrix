// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package blkgenor

import (
	"bou.ke/monkey"
	"fmt"
	"github.com/MatrixAINetwork/go-matrix/ca"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/crypto"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/params/manparams"
	. "github.com/smartystreets/goconvey/convey"
	"math/big"
	"testing"
	"time"
)

func TestBlockGenor_roleUpdatedMsgHandle(t *testing.T) {
	log.InitLog(3)
	eth := fakeEthNew(0)
	if nil == eth {
		fmt.Println("failed to create eth")
	}

	Convey("CA消息测试", t, func() {
		SkipConvey("高度0，当前身份为验证者", func() {
			//process := newProcess(1, pm)
			monkey.Patch(manparams.NewBCIntervalByHash, func(blockHash common.Hash) (*manparams.BCInterval, error) {
				fmt.Println("use monkey NewBCIntervalByNumber")

				inteval1 := &mc.BCIntervalInfo{LastBCNumber: 0, LastReelectNumber: 0, BCInterval: 100}

				interval2, _ := manparams.NewBCIntervalWithInterval(inteval1)
				return interval2, nil
			})
			blockgen, err := New(eth)
			if err != nil {
			}
			roleMsg := &mc.RoleUpdatedMsg{Role: common.RoleValidator, BlockNum: uint64(0), BlockHash: common.Hash{1}, Leader: common.HexToAddress(testAddress), IsSuperBlock: false}
			blockgen.roleUpdatedMsgHandle(roleMsg)
			p := blockgen.pm.GetCurrentProcess()
			So(blockgen.pm.curNumber, ShouldEqual, 1)
			So(p.state, ShouldEqual, StateBlockBroadcast)
			So(p.role, ShouldEqual, common.RoleValidator)
		})
		SkipConvey("高度0，当前身份为广播节点", func() {
			//process := newProcess(1, pm)
			monkey.Patch(manparams.NewBCIntervalByHash, func(blockHash common.Hash) (*manparams.BCInterval, error) {
				fmt.Println("use monkey NewBCIntervalByNumber")

				inteval1 := &mc.BCIntervalInfo{LastBCNumber: 0, LastReelectNumber: 0, BCInterval: 100}

				interval2, _ := manparams.NewBCIntervalWithInterval(inteval1)
				return interval2, nil
			})
			blockgen, err := New(eth)
			if err != nil {
			}
			roleMsg := &mc.RoleUpdatedMsg{Role: common.RoleBroadcast, BlockNum: uint64(0), BlockHash: common.Hash{1}, Leader: common.HexToAddress(testAddress), IsSuperBlock: false}
			blockgen.roleUpdatedMsgHandle(roleMsg)
			p := blockgen.pm.GetCurrentProcess()
			So(blockgen.pm.curNumber, ShouldEqual, 1)
			So(p.state, ShouldEqual, StateMinerResultVerify)
			So(p.role, ShouldEqual, common.RoleBroadcast)
		})
		//
		SkipConvey("高度0，当前身份为其它节点", func() {
			//process := newProcess(1, pm)
			monkey.Patch(manparams.NewBCIntervalByHash, func(blockHash common.Hash) (*manparams.BCInterval, error) {
				fmt.Println("use monkey NewBCIntervalByNumber")

				inteval1 := &mc.BCIntervalInfo{LastBCNumber: 0, LastReelectNumber: 0, BCInterval: 100}

				interval2, _ := manparams.NewBCIntervalWithInterval(inteval1)
				return interval2, nil
			})
			blockgen, err := New(eth)
			if err != nil {
			}
			roleMsg := &mc.RoleUpdatedMsg{Role: common.RoleMiner, BlockNum: uint64(0), BlockHash: common.Hash{1}, Leader: common.HexToAddress(testAddress), IsSuperBlock: false}
			blockgen.roleUpdatedMsgHandle(roleMsg)
			So(blockgen.pm.curNumber, ShouldEqual, 1)
			_, ok := blockgen.pm.processMap[1]
			So(ok, ShouldBeFalse)

		})

		SkipConvey("高度广播高度，当前身份为广播节点", func() {
			//process := newProcess(1, pm)
			monkey.Patch(manparams.NewBCIntervalByHash, func(blockHash common.Hash) (*manparams.BCInterval, error) {
				fmt.Println("use monkey NewBCIntervalByNumber")

				inteval1 := &mc.BCIntervalInfo{LastBCNumber: 100, LastReelectNumber: 0, BCInterval: 100}

				interval2, _ := manparams.NewBCIntervalWithInterval(inteval1)
				return interval2, nil
			})
			blockgen, err := New(eth)
			if err != nil {
			}
			roleMsg := &mc.RoleUpdatedMsg{Role: common.RoleBroadcast, BlockNum: uint64(99), BlockHash: common.Hash{1}, Leader: common.HexToAddress(testAddress), IsSuperBlock: false}
			blockgen.roleUpdatedMsgHandle(roleMsg)
			So(blockgen.pm.curNumber, ShouldEqual, 100)
			curProcess := blockgen.pm.GetCurrentProcess()
			So(curProcess.number, ShouldEqual, 100)
			So(curProcess.state, ShouldEqual, StateHeaderGen)

		})

		Convey("高度普通高度，超级区块测试", func() {
			//process := newProcess(1, pm)
			monkey.Patch(manparams.NewBCIntervalByHash, func(blockHash common.Hash) (*manparams.BCInterval, error) {
				fmt.Println("use monkey NewBCIntervalByNumber")

				inteval1 := &mc.BCIntervalInfo{LastBCNumber: 100, LastReelectNumber: 0, BCInterval: 100}

				interval2, _ := manparams.NewBCIntervalWithInterval(inteval1)
				return interval2, nil
			})
			blockgen, err := New(eth)
			if err != nil {
			}
			roleMsg := &mc.RoleUpdatedMsg{Role: common.RoleBroadcast, BlockNum: uint64(4), BlockHash: common.Hash{1}, Leader: common.HexToAddress(testAddress), IsSuperBlock: true}
			blockgen.roleUpdatedMsgHandle(roleMsg)
			roleMsg = &mc.RoleUpdatedMsg{Role: common.RoleBroadcast, BlockNum: uint64(3), BlockHash: common.Hash{1}, Leader: common.HexToAddress(testAddress), IsSuperBlock: true}
			blockgen.roleUpdatedMsgHandle(roleMsg)
			So(blockgen.pm.curNumber, ShouldEqual, 4)
			curProcess := blockgen.pm.GetCurrentProcess()
			So(curProcess.number, ShouldEqual, 4)
			So(curProcess.state, ShouldEqual, StateMinerResultVerify)

		})
	})

}

func TestLeaderChangeNotifyHeight1NoCA(t *testing.T) {
	log.InitLog(3)
	eth := fakeEthNew(0)
	if nil == eth {
		fmt.Println("failed to create eth")
	}

	//process := newProcess(1, pm)

	Convey("未收到ca消息 leader测试", t, func() {
		SkipConvey("未收到高度1的ca消息，当前节点是高度1的leader", func() {
			blockgen, err := New(eth)
			if err != nil {
			}
			//roleMsg := &mc.RoleUpdatedMsg{common.RoleValidator, uint64(0), common.HexToAddress(testAddress)}
			//blockgen.roleUpdatedMsgHandle(roleMsg)
			var guard *monkey.PatchGuard
			guard = monkey.Patch(ca.GetAddress, func() common.Address {
				guard.Unpatch()
				defer guard.Restore()
				return common.HexToAddress(testAddress1)
			})
			//blockgen.roleUpdatedMsgHandle(roleMsg)
			leaderMsg := &mc.LeaderChangeNotify{ConsensusState: true, PreLeader: common.HexToAddress(testAddress0), Leader: common.HexToAddress(testAddress1), NextLeader: common.HexToAddress(testAddress2), Number: 1, ConsensusTurn: mc.ConsensusTurnInfo{}}
			blockgen.leaderChangeNotifyHandle(leaderMsg)
			process, ok := blockgen.pm.processMap[1]
			So(ok, ShouldBeTrue)
			So(process.number, ShouldEqual, 1)
			So(process.state, ShouldEqual, StateIdle)
			So(process.curLeader, ShouldEqual, common.HexToAddress(testAddress1))
			So(process.nextLeader, ShouldEqual, common.HexToAddress(testAddress2))
			nextprocess, ok := blockgen.pm.processMap[2]
			So(ok, ShouldBeTrue)
			So(nextprocess.number, ShouldEqual, 2)
			So(nextprocess.state, ShouldEqual, StateIdle)
			So(nextprocess.curLeader, ShouldEqual, common.HexToAddress(testAddress2))
		})

		SkipConvey("未收到高度1的ca消息，当前节点是高度1的next leader", func() {
			blockgen, err := New(eth)
			if err != nil {
			}
			//roleMsg := &mc.RoleUpdatedMsg{common.RoleValidator, uint64(0), common.HexToAddress(testAddress)}
			//blockgen.roleUpdatedMsgHandle(roleMsg)
			var guard *monkey.PatchGuard
			guard = monkey.Patch(ca.GetAddress, func() common.Address {
				guard.Unpatch()
				defer guard.Restore()
				return common.HexToAddress(testAddress)
			})
			//blockgen.roleUpdatedMsgHandle(roleMsg)
			leaderMsg := &mc.LeaderChangeNotify{ConsensusState: true, PreLeader: common.HexToAddress(testAddress2), Leader: common.HexToAddress(testAddress1), NextLeader: common.HexToAddress(testAddress), Number: 1, ConsensusTurn: mc.ConsensusTurnInfo{}}
			blockgen.leaderChangeNotifyHandle(leaderMsg)
			process, ok := blockgen.pm.processMap[1]
			So(ok, ShouldBeTrue)
			So(process.number, ShouldEqual, 1)
			So(process.state, ShouldEqual, StateIdle)
			So(process.curLeader, ShouldEqual, common.HexToAddress(testAddress1))
			So(process.nextLeader, ShouldEqual, common.HexToAddress(testAddress))
			nextprocess, ok := blockgen.pm.processMap[2]
			So(ok, ShouldBeTrue)
			So(nextprocess.number, ShouldEqual, 2)
			So(nextprocess.state, ShouldEqual, StateIdle)
			So(nextprocess.curLeader, ShouldEqual, common.HexToAddress(testAddress))
		})

		Convey("未收到高度1的ca消息，当前是高度1无lead状态", func() {
			blockgen, err := New(eth)
			if err != nil {
			}
			//roleMsg := &mc.RoleUpdatedMsg{common.RoleValidator, uint64(0), common.HexToAddress(testAddress)}
			//blockgen.roleUpdatedMsgHandle(roleMsg)
			var guard *monkey.PatchGuard
			guard = monkey.Patch(ca.GetAddress, func() common.Address {
				guard.Unpatch()
				defer guard.Restore()
				return common.HexToAddress(testAddress)
			})
			//blockgen.roleUpdatedMsgHandle(roleMsg)
			leaderMsg := &mc.LeaderChangeNotify{ConsensusState: false, PreLeader: common.HexToAddress(testAddress2), Leader: common.HexToAddress(testAddress1), NextLeader: common.HexToAddress(testAddress), Number: 1, ConsensusTurn: mc.ConsensusTurnInfo{}}
			blockgen.leaderChangeNotifyHandle(leaderMsg)
			process, ok := blockgen.pm.processMap[1]
			So(ok, ShouldBeTrue)
			So(process.number, ShouldEqual, 1)
			So(process.state, ShouldEqual, StateIdle)
			So(process.curLeader, ShouldEqual, common.Address{})
			So(process.nextLeader, ShouldEqual, common.Address{})
			_, ok = blockgen.pm.processMap[2]
			So(ok, ShouldBeFalse)
		})

	})
}

//func TestLeaderChangeNotifyHeight2NoCA(t *testing.T) {
//	log.InitLog(3)
//	eth := fakeEthNew(0)
//	if nil == eth {
//		fmt.Println("failed to create eth")
//	}
//
//	//process := newProcess(1, pm)
//	currentHeight := uint64(2)
//	Convey("leader测试", t, func() {
//Convey("生成验证区块头函数测试，当前节点是leader", func() {
//	blockgen, err := New(eth)
//	if err != nil {
//	}
//	//roleMsg := &mc.RoleUpdatedMsg{common.RoleValidator, uint64(0), common.HexToAddress(testAddress)}
//	//blockgen.roleUpdatedMsgHandle(roleMsg)
//	var guard *monkey.PatchGuard
//	guard = monkey.Patch(ca.GetAddress, func() common.Address {
//		guard.Unpatch()
//		defer guard.Restore()
//		return common.HexToAddress(testAddress)
//	})
//	//blockgen.roleUpdatedMsgHandle(roleMsg)
//	leaderMsg := &mc.LeaderChangeNotify{true, common.HexToAddress(testAddress), common.HexToAddress(testAddress1), currentHeight, 0}
//	blockgen.leaderChangeNotifyHandle(leaderMsg)
//
//	preprocess, ok := blockgen.pm.processMap[currentHeight-1]
//	So(ok, ShouldBeTrue)
//	So(preprocess.number, ShouldEqual, currentHeight-1)
//	So(preprocess.state, ShouldEqual, StateIdle)
//	So(preprocess.nextLeader, ShouldEqual, common.HexToAddress(testAddress))
//
//	process, ok := blockgen.pm.processMap[currentHeight]
//	So(ok, ShouldBeTrue)
//	So(process.number, ShouldEqual, currentHeight)
//	So(process.state, ShouldEqual, StateIdle)
//	So(process.curLeader, ShouldEqual, common.HexToAddress(testAddress))
//	So(process.nextLeader, ShouldEqual, common.HexToAddress(testAddress1))
//	_, ok = blockgen.pm.processMap[currentHeight+1]
//	So(ok, ShouldBeFalse)
//})

//Convey("生成区块测试，当前节点是next leader", func() {
//	blockgen, err := New(eth)
//	if err != nil {
//	}
//	//roleMsg := &mc.RoleUpdatedMsg{common.RoleValidator, uint64(0), common.HexToAddress(testAddress)}
//	//blockgen.roleUpdatedMsgHandle(roleMsg)
//	var guard *monkey.PatchGuard
//	guard = monkey.Patch(ca.GetAddress, func() common.Address {
//		guard.Unpatch()
//		defer guard.Restore()
//		return common.HexToAddress(testAddress)
//	})
//	//blockgen.roleUpdatedMsgHandle(roleMsg)
//	leaderMsg := &mc.LeaderChangeNotify{true, common.HexToAddress(testAddress1), common.HexToAddress(testAddress), currentHeight, 0}
//	blockgen.leaderChangeNotifyHandle(leaderMsg)
//
//	preprocess, ok := blockgen.pm.processMap[currentHeight-1]
//	So(ok, ShouldBeTrue)
//	So(preprocess.number, ShouldEqual, currentHeight-1)
//	So(preprocess.state, ShouldEqual, StateIdle)
//	So(preprocess.nextLeader, ShouldEqual, common.HexToAddress(testAddress1))
//
//	process, ok := blockgen.pm.processMap[currentHeight]
//	So(ok, ShouldBeTrue)
//	So(process.number, ShouldEqual, currentHeight)
//	So(process.state, ShouldEqual, StateIdle)
//	So(process.curLeader, ShouldEqual, common.HexToAddress(testAddress1))
//	So(process.nextLeader, ShouldEqual, common.HexToAddress(testAddress))
//	_, ok = blockgen.pm.processMap[currentHeight+1]
//	So(ok, ShouldBeFalse)
//})
//
//Convey("生成区块测试，当前无lead状态", func() {
//	blockgen, err := New(eth)
//	if err != nil {
//	}
//	//roleMsg := &mc.RoleUpdatedMsg{common.RoleValidator, uint64(0), common.HexToAddress(testAddress)}
//	//blockgen.roleUpdatedMsgHandle(roleMsg)
//	var guard *monkey.PatchGuard
//	guard = monkey.Patch(ca.GetAddress, func() common.Address {
//		guard.Unpatch()
//		defer guard.Restore()
//		return common.HexToAddress(testAddress)
//	})
//	//blockgen.roleUpdatedMsgHandle(roleMsg)
//	leaderMsg := &mc.LeaderChangeNotify{false, common.HexToAddress(testAddress1), common.HexToAddress(testAddress), currentHeight, 0}
//	blockgen.leaderChangeNotifyHandle(leaderMsg)
//	preprocess, ok := blockgen.pm.processMap[currentHeight-1]
//	So(ok, ShouldBeTrue)
//	So(preprocess.number, ShouldEqual, currentHeight-1)
//	So(preprocess.state, ShouldEqual, StateIdle)
//	So(preprocess.nextLeader, ShouldEqual, common.Address{})
//
//	blockgen.leaderChangeNotifyHandle(leaderMsg)
//	process, ok := blockgen.pm.processMap[currentHeight]
//	So(ok, ShouldBeTrue)
//	So(process.number, ShouldEqual, currentHeight)
//	So(process.state, ShouldEqual, StateIdle)
//	So(process.curLeader, ShouldEqual, common.Address{})
//	So(process.nextLeader, ShouldEqual, common.Address{})
//	_, ok = blockgen.pm.processMap[currentHeight+1]
//	So(ok, ShouldBeFalse)
//})

//	})
//}

//func TestLeaderChangeNotifyHeight1WithCA(t *testing.T) {
//	log.InitLog(3)
//	eth := fakeEthNew(0)
//	if nil == eth {
//		fmt.Println("failed to create eth")
//	}
//
//	//process := newProcess(1, pm)
//	LeaddercurrentHeight := uint64(1)
//	Convey("leader测试", t, func() {
//		Convey("生成验证区块头函数测试，先收ca再收leader", func() {
//			blockgen, err := New(eth)
//			if err != nil {
//			}
//			roleMsg := &mc.RoleUpdatedMsg{common.RoleValidator, uint64(LeaddercurrentHeight - 1), common.HexToAddress(testAddress)}
//			blockgen.roleUpdatedMsgHandle(roleMsg)
//			var guard *monkey.PatchGuard
//			guard = monkey.Patch(ca.GetAddress, func() common.Address {
//				guard.Unpatch()
//				defer guard.Restore()
//				return common.HexToAddress(testAddress)
//			})
//			blockgen.roleUpdatedMsgHandle(roleMsg)
//			leaderMsg := &mc.LeaderChangeNotify{true, common.HexToAddress(testAddress), common.HexToAddress(testAddress1), LeaddercurrentHeight, 0}
//			blockgen.leaderChangeNotifyHandle(leaderMsg)
//
//			_, ok := blockgen.pm.processMap[LeaddercurrentHeight-1]
//			So(ok, ShouldBeFalse)
//
//			process, ok := blockgen.pm.processMap[LeaddercurrentHeight]
//			So(ok, ShouldBeTrue)
//			So(process.number, ShouldEqual, LeaddercurrentHeight)
//			So(process.state, ShouldEqual, StateMinerResultVerify)
//			So(process.curLeader, ShouldEqual, common.HexToAddress(testAddress))
//			So(process.nextLeader, ShouldEqual, common.HexToAddress(testAddress1))
//			nextprocess, ok := blockgen.pm.processMap[LeaddercurrentHeight+1]
//			So(ok, ShouldBeTrue)
//			So(nextprocess.number, ShouldEqual, 2)
//			So(nextprocess.state, ShouldEqual, StateIdle)
//			So(nextprocess.curLeader, ShouldEqual, common.HexToAddress(testAddress1))
//		})
//
//		Convey("生成验证区块头函数测试，先收leader再收ca", func() {
//			blockgen, err := New(eth)
//			if err != nil {
//			}
//			roleMsg := &mc.RoleUpdatedMsg{common.RoleValidator, uint64(LeaddercurrentHeight - 1), common.HexToAddress(testAddress)}
//			blockgen.roleUpdatedMsgHandle(roleMsg)
//			var guard *monkey.PatchGuard
//			guard = monkey.Patch(ca.GetAddress, func() common.Address {
//				guard.Unpatch()
//				defer guard.Restore()
//				return common.HexToAddress(testAddress)
//			})
//
//			leaderMsg := &mc.LeaderChangeNotify{true, common.HexToAddress(testAddress), common.HexToAddress(testAddress1), LeaddercurrentHeight, 0}
//			blockgen.leaderChangeNotifyHandle(leaderMsg)
//			blockgen.roleUpdatedMsgHandle(roleMsg)
//			_, ok := blockgen.pm.processMap[LeaddercurrentHeight-1]
//			So(ok, ShouldBeFalse)
//
//			process, ok := blockgen.pm.processMap[LeaddercurrentHeight]
//			So(ok, ShouldBeTrue)
//			So(process.number, ShouldEqual, LeaddercurrentHeight)
//			So(process.state, ShouldEqual, StateMinerResultVerify)
//			So(process.curLeader, ShouldEqual, common.HexToAddress(testAddress))
//			So(process.nextLeader, ShouldEqual, common.HexToAddress(testAddress1))
//			nextprocess, ok := blockgen.pm.processMap[LeaddercurrentHeight+1]
//			So(ok, ShouldBeTrue)
//			So(nextprocess.number, ShouldEqual, LeaddercurrentHeight+1)
//			So(nextprocess.state, ShouldEqual, StateIdle)
//			So(nextprocess.curLeader, ShouldEqual, common.HexToAddress(testAddress1))
//		})
//
//		Convey("生成区块测试，先收leader再收ca，当前节点是next leader", func() {
//			blockgen, err := New(eth)
//			if err != nil {
//			}
//			roleMsg := &mc.RoleUpdatedMsg{common.RoleValidator, uint64(LeaddercurrentHeight - 1), common.HexToAddress(testAddress)}
//			blockgen.roleUpdatedMsgHandle(roleMsg)
//			var guard *monkey.PatchGuard
//			guard = monkey.Patch(ca.GetAddress, func() common.Address {
//				guard.Unpatch()
//				defer guard.Restore()
//				return common.HexToAddress(testAddress)
//			})
//			blockgen.roleUpdatedMsgHandle(roleMsg)
//			leaderMsg := &mc.LeaderChangeNotify{true, common.HexToAddress(testAddress1), common.HexToAddress(testAddress), LeaddercurrentHeight, 0}
//			blockgen.leaderChangeNotifyHandle(leaderMsg)
//
//			process, ok := blockgen.pm.processMap[LeaddercurrentHeight]
//			So(ok, ShouldBeTrue)
//			So(process.number, ShouldEqual, LeaddercurrentHeight)
//			So(process.state, ShouldEqual, StateMinerResultVerify)
//			So(process.curLeader, ShouldEqual, common.HexToAddress(testAddress1))
//			So(process.nextLeader, ShouldEqual, common.HexToAddress(testAddress))
//			nextprocess, ok := blockgen.pm.processMap[LeaddercurrentHeight+1]
//			So(ok, ShouldBeTrue)
//			So(nextprocess.number, ShouldEqual, LeaddercurrentHeight+1)
//			So(nextprocess.state, ShouldEqual, StateIdle)
//			So(nextprocess.curLeader, ShouldEqual, common.HexToAddress(testAddress))
//		})
//	})
//}

func TestBlockGenor_minerResultHandle(t *testing.T) {
	log.InitLog(3)

	eth := fakeEthNew(0)
	if nil == eth {
		fmt.Println("failed to create eth")
	}
	monkey.Patch(manparams.NewBCIntervalByHash, func(blockHash common.Hash) (*manparams.BCInterval, error) {
		fmt.Println("use monkey NewBCIntervalByNumber")

		inteval1 := &mc.BCIntervalInfo{LastBCNumber: 0, LastReelectNumber: 0, BCInterval: 100}

		interval2, _ := manparams.NewBCIntervalWithInterval(inteval1)
		return interval2, nil
	})
	Convey("矿工挖矿结果测试", t, func() {

		Convey("矿工挖矿结果测试，高度1", func() {
			blockgen, err := New(eth)
			if err != nil {
			}
			blockhash := common.Hash{0x01}
			diff := big.NewInt(100)
			from := common.HexToAddress(testAddress)
			Number := uint64(1)
			Signatures := []common.Signature{common.BytesToSignature(common.BigToHash(big.NewInt(100)).Bytes()), common.BytesToSignature(common.BigToHash(big.NewInt(101)).Bytes()), common.BytesToSignature(common.BigToHash(big.NewInt(102)).Bytes())}
			fromlist := []common.Address{common.HexToAddress(testAddress), common.HexToAddress(testAddress1), common.HexToAddress(testAddress2), common.HexToAddress(testAddress3)}
			minerResults := make([]*mc.HD_MiningRspMsg, 0)
			for i := 0; i < len(fromlist); i++ {
				tempResult := &mc.HD_MiningRspMsg{fromlist[i], Number, blockhash, diff, types.EncodeNonce(uint64(0)), from, common.Hash{0x01}, Signatures}
				minerResults = append(minerResults, tempResult)
				blockgen.minerResultHandle(tempResult)
			}
			p, ok := blockgen.pm.processMap[Number]
			So(ok, ShouldBeTrue)
			retMinerResults, err := p.powPool.GetMinerResults(blockhash, diff)
			So(err, ShouldBeNil)
			for i := 0; i < len(fromlist); i++ {
				So(retMinerResults, ShouldContain, minerResults[i])
			}

		})

		Convey("矿工挖矿结果测试，高度2", func() {
			blockgen, err := New(eth)
			if err != nil {
			}
			blockhash := common.Hash{0x01}
			diff := big.NewInt(100)
			from := common.HexToAddress(testAddress)
			Number := uint64(2)
			Signatures := []common.Signature{common.BytesToSignature(common.BigToHash(big.NewInt(100)).Bytes()), common.BytesToSignature(common.BigToHash(big.NewInt(101)).Bytes()), common.BytesToSignature(common.BigToHash(big.NewInt(102)).Bytes())}
			fromlist := []common.Address{common.HexToAddress(testAddress), common.HexToAddress(testAddress1), common.HexToAddress(testAddress2), common.HexToAddress(testAddress3)}
			minerResults := make([]*mc.HD_MiningRspMsg, 0)
			for i := 0; i < len(fromlist); i++ {
				tempResult := &mc.HD_MiningRspMsg{fromlist[i], Number, blockhash, diff, types.EncodeNonce(uint64(0)), from, common.Hash{0x01}, Signatures}
				minerResults = append(minerResults, tempResult)
				blockgen.minerResultHandle(tempResult)
			}
			p, ok := blockgen.pm.processMap[Number]
			So(ok, ShouldBeTrue)
			retMinerResults, err := p.powPool.GetMinerResults(blockhash, diff)
			So(err, ShouldBeNil)
			for i := 0; i < len(fromlist); i++ {
				So(retMinerResults, ShouldContain, minerResults[i])
			}

		})
		Convey("矿工挖矿结果测试，高度3", func() {
			blockgen, err := New(eth)
			if err != nil {
			}
			blockhash := common.Hash{0x01}
			diff := big.NewInt(100)
			from := common.HexToAddress(testAddress)
			Number := uint64(3)
			Signatures := []common.Signature{common.BytesToSignature(common.BigToHash(big.NewInt(100)).Bytes()), common.BytesToSignature(common.BigToHash(big.NewInt(101)).Bytes()), common.BytesToSignature(common.BigToHash(big.NewInt(102)).Bytes())}
			fromlist := []common.Address{common.HexToAddress(testAddress), common.HexToAddress(testAddress1), common.HexToAddress(testAddress2), common.HexToAddress(testAddress3)}
			minerResults := make([]*mc.HD_MiningRspMsg, 0)
			for i := 0; i < len(fromlist); i++ {
				tempResult := &mc.HD_MiningRspMsg{fromlist[i], Number, blockhash, diff, types.EncodeNonce(uint64(0)), from, common.Hash{0x01}, Signatures}
				minerResults = append(minerResults, tempResult)
				blockgen.minerResultHandle(tempResult)
			}
			_, ok := blockgen.pm.processMap[Number]
			So(ok, ShouldBeFalse)

		})
		Convey("矿工挖矿结果测试，收到CA高度为1， 挖矿结果高度3", func() {
			blockgen, err := New(eth)
			if err != nil {
			}
			blockhash := common.Hash{0x01}
			diff := big.NewInt(100)
			from := common.HexToAddress(testAddress)
			Number := uint64(3)
			Signatures := []common.Signature{common.BytesToSignature(common.BigToHash(big.NewInt(100)).Bytes()), common.BytesToSignature(common.BigToHash(big.NewInt(101)).Bytes()), common.BytesToSignature(common.BigToHash(big.NewInt(102)).Bytes())}
			fromlist := []common.Address{common.HexToAddress(testAddress), common.HexToAddress(testAddress1), common.HexToAddress(testAddress2), common.HexToAddress(testAddress3)}
			minerResults := make([]*mc.HD_MiningRspMsg, 0)
			roleMsg := &mc.RoleUpdatedMsg{Role: common.RoleValidator, BlockNum: uint64(1), Leader: common.HexToAddress(testAddress)}
			blockgen.roleUpdatedMsgHandle(roleMsg)
			for i := 0; i < len(fromlist); i++ {
				tempResult := &mc.HD_MiningRspMsg{fromlist[i], Number, blockhash, diff, types.EncodeNonce(uint64(0)), from, common.Hash{0x01}, Signatures}
				minerResults = append(minerResults, tempResult)
				blockgen.minerResultHandle(tempResult)
			}
			p, ok := blockgen.pm.processMap[Number]
			So(ok, ShouldBeTrue)
			retMinerResults, err := p.powPool.GetMinerResults(blockhash, diff)
			So(err, ShouldBeNil)
			for i := 0; i < len(fromlist); i++ {
				So(retMinerResults, ShouldContain, minerResults[i])
			}
		})

		Convey("矿工挖矿结果测试，收到CA高度为57， 挖矿结果高度57", func() {
			blockgen, err := New(eth)
			if err != nil {
			}
			blockhash := common.Hash{0x01}
			diff := big.NewInt(100)
			from := common.HexToAddress(testAddress)
			Number := uint64(57)
			Signatures := []common.Signature{common.BytesToSignature(common.BigToHash(big.NewInt(100)).Bytes()), common.BytesToSignature(common.BigToHash(big.NewInt(101)).Bytes()), common.BytesToSignature(common.BigToHash(big.NewInt(102)).Bytes())}
			fromlist := []common.Address{common.HexToAddress(testAddress), common.HexToAddress(testAddress1), common.HexToAddress(testAddress2), common.HexToAddress(testAddress3)}
			minerResults := make([]*mc.HD_MiningRspMsg, 0)
			roleMsg := &mc.RoleUpdatedMsg{Role: common.RoleValidator, BlockNum: uint64(57), Leader: common.HexToAddress(testAddress)}
			blockgen.roleUpdatedMsgHandle(roleMsg)
			for i := 0; i < len(fromlist); i++ {
				tempResult := &mc.HD_MiningRspMsg{fromlist[i], Number, blockhash, diff, types.EncodeNonce(uint64(0)), from, common.Hash{0x01}, Signatures}
				minerResults = append(minerResults, tempResult)
				blockgen.minerResultHandle(tempResult)
			}
			p, ok := blockgen.pm.processMap[Number]
			So(ok, ShouldBeTrue)
			retMinerResults, err := p.powPool.GetMinerResults(blockhash, diff)
			So(err, ShouldBeNil)
			for i := 0; i < len(fromlist); i++ {
				So(retMinerResults, ShouldContain, minerResults[i])
			}
		})

	})
}

//
//func TestBlockGenor_minerVerifyHandle(t *testing.T) {
//	log.InitLog(3)
//
//	eth := fakeEthNew(0)
//	if nil == eth {
//		fmt.Println("failed to create eth")
//	}
//
//	Convey("本地共识结果测试", t, func() {
//
//		Convey("本地共识结果测试，高度1", func() {
//			blockgen, err := New(eth)
//			if err != nil {
//			}
//			header := eth.BlockChain().CurrentHeader()
//			newheader := types.CopyHeader(header)
//			newheader.Number = big.NewInt(1)
//			state, err := eth.blockchain.StateAt(eth.blockchain.GetBlockByHash(header.Hash()).Root())
//			So(err, ShouldBeNil)
//			blockConsensus := &mc.BlockVerifyConsensusOK{newheader, common.Hash{0x10}, nil, nil, state}
//			blockgen.consensusBlockMsgHandle(blockConsensus)
//			p, ok := blockgen.pm.processMap[1]
//			So(ok, ShouldBeTrue)
//			So(p.consensusBlock, ShouldEqual, blockConsensus)
//		})
//	})
//
//}
//
func TestBlockGenor_broadcastMinerResultHandle(t *testing.T) {

	log.InitLog(3)
	eth := fakeEthNew(0)
	if nil == eth {
		fmt.Println("failed to create eth")
	}
	monkey.Patch(manparams.NewBCIntervalByHash, func(blockHash common.Hash) (*manparams.BCInterval, error) {
		fmt.Println("use monkey NewBCIntervalByNumber")

		inteval1 := &mc.BCIntervalInfo{LastBCNumber: 0, LastReelectNumber: 0, BCInterval: 100}

		interval2, _ := manparams.NewBCIntervalWithInterval(inteval1)
		return interval2, nil
	})
	Convey("广播矿工挖矿结果测试", t, func() {

		Convey("广播矿工挖矿结果测试，高度100", func() {
			blockgen, err := New(eth)
			if err != nil {
			}
			header := eth.BlockChain().CurrentHeader()
			Number := uint64(100)
			newheader := types.CopyHeader(header)
			newheader.Number = big.NewInt(int64(Number))
			newheader.Leader = common.HexToAddress(testAddress)
			newheader.Signatures = append(newheader.Signatures, common.BytesToSignature(common.BigToHash(big.NewInt(100)).Bytes()))
			fromlist := []common.Address{common.HexToAddress(testAddress), common.HexToAddress(testAddress1), common.HexToAddress(testAddress2), common.HexToAddress(testAddress3)}
			roleMsg := &mc.RoleUpdatedMsg{Role: common.RoleValidator, BlockNum: uint64(Number - 1), Leader: common.HexToAddress(testAddress)}
			blockgen.roleUpdatedMsgHandle(roleMsg)

			broadcastResult := &mc.HD_BroadcastMiningRspMsg{fromlist[0], &mc.BlockData{newheader, nil}}
			var guard *monkey.PatchGuard
			guard = monkey.Patch(ca.GetAccountOriginalRole, func(account common.Address, number common.Hash) (common.RoleType, error) {
				guard.Unpatch()
				defer guard.Restore()
				return common.RoleDefault, nil
			})

			var guard1 *monkey.PatchGuard
			guard1 = monkey.Patch(crypto.VerifySignWithValidate, func(sighash []byte, sig []byte) (common.Address, bool, error) {
				guard1.Unpatch()
				defer guard1.Restore()
				return common.HexToAddress(testAddress), true, nil
			})

			blockgen.broadcastMinerResultHandle(broadcastResult)
			p, ok := blockgen.pm.processMap[Number]
			So(ok, ShouldBeTrue)
			So(p.broadcastRstCache, ShouldContain, broadcastResult.BlockMainData)

		})

		Convey("广播矿工挖矿结果测试，高度99", func() {
			blockgen, err := New(eth)
			if err != nil {
			}
			header := eth.BlockChain().CurrentHeader()
			Number := uint64(100 - 2)
			newheader := types.CopyHeader(header)
			newheader.ParentHash = header.Hash()
			newheader.Number = big.NewInt(int64(Number + 1))
			newheader.Leader = common.HexToAddress(testAddress)
			newheader.Signatures = append(newheader.Signatures, common.BytesToSignature(common.BigToHash(big.NewInt(100)).Bytes()))
			fromlist := []common.Address{common.HexToAddress(testAddress), common.HexToAddress(testAddress1), common.HexToAddress(testAddress2), common.HexToAddress(testAddress3)}
			roleMsg := &mc.RoleUpdatedMsg{Role: common.RoleValidator, BlockNum: uint64(Number), Leader: common.HexToAddress(testAddress)}
			blockgen.roleUpdatedMsgHandle(roleMsg)

			broadcastResult := &mc.HD_BroadcastMiningRspMsg{fromlist[0], &mc.BlockData{newheader, nil}}
			var guard *monkey.PatchGuard
			guard = monkey.Patch(ca.GetAccountOriginalRole, func(account common.Address, number common.Hash) (common.RoleType, error) {
				guard.Unpatch()
				defer guard.Restore()
				return common.RoleDefault, nil
			})
			var guard1 *monkey.PatchGuard
			guard1 = monkey.Patch(crypto.VerifySignWithValidate, func(sighash []byte, sig []byte) (common.Address, bool, error) {
				guard1.Unpatch()
				defer guard1.Restore()
				return common.HexToAddress(testAddress), true, nil
			})

			blockgen.broadcastMinerResultHandle(broadcastResult)
			p, ok := blockgen.pm.processMap[Number+1]
			So(ok, ShouldBeTrue)
			So(p.state, ShouldEqual, StateBlockBroadcast)
			So(p.broadcastRstCache, ShouldNotContain, broadcastResult.BlockMainData)

		})

		Convey("广播矿工挖矿结果测试，高度99，身份错误", func() {
			blockgen, err := New(eth)
			if err != nil {
			}
			header := eth.BlockChain().CurrentHeader()
			Number := uint64(100)
			newheader := types.CopyHeader(header)
			newheader.Number = big.NewInt(int64(Number))
			newheader.Leader = common.HexToAddress(testAddress)
			newheader.Signatures = append(newheader.Signatures, common.BytesToSignature(common.BigToHash(big.NewInt(100)).Bytes()))
			fromlist := []common.Address{common.HexToAddress(testAddress), common.HexToAddress(testAddress1), common.HexToAddress(testAddress2), common.HexToAddress(testAddress3)}
			roleMsg := &mc.RoleUpdatedMsg{Role: common.RoleValidator, BlockNum: uint64(Number - 1), Leader: common.HexToAddress(testAddress)}
			blockgen.roleUpdatedMsgHandle(roleMsg)

			broadcastResult := &mc.HD_BroadcastMiningRspMsg{fromlist[0], &mc.BlockData{newheader, nil}}
			var guard *monkey.PatchGuard
			guard = monkey.Patch(ca.GetAccountOriginalRole, func(account common.Address, number common.Hash) (common.RoleType, error) {
				guard.Unpatch()
				defer guard.Restore()
				return common.RoleDefault, nil
			})

			var guard1 *monkey.PatchGuard
			guard1 = monkey.Patch(crypto.VerifySignWithValidate, func(sighash []byte, sig []byte) (common.Address, bool, error) {
				guard1.Unpatch()
				defer guard1.Restore()
				return common.HexToAddress(testAddress), true, nil
			})

			blockgen.broadcastMinerResultHandle(broadcastResult)
			p, ok := blockgen.pm.processMap[Number]
			So(ok, ShouldBeTrue)
			So(p.broadcastRstCache, ShouldContain, broadcastResult.BlockMainData)

		})

		Convey("广播矿工挖矿结果测试，高度99,签名错误", func() {
			blockgen, err := New(eth)
			if err != nil {
			}
			header := eth.BlockChain().CurrentHeader()
			Number := uint64(100)
			newheader := types.CopyHeader(header)
			newheader.Number = big.NewInt(int64(Number))
			newheader.Leader = common.HexToAddress(testAddress)
			newheader.Signatures = append(newheader.Signatures, common.BytesToSignature(common.BigToHash(big.NewInt(100)).Bytes()))
			fromlist := []common.Address{common.HexToAddress(testAddress), common.HexToAddress(testAddress1), common.HexToAddress(testAddress2), common.HexToAddress(testAddress3)}
			roleMsg := &mc.RoleUpdatedMsg{Role: common.RoleValidator, BlockNum: uint64(Number - 1), Leader: common.HexToAddress(testAddress)}
			blockgen.roleUpdatedMsgHandle(roleMsg)

			broadcastResult := &mc.HD_BroadcastMiningRspMsg{fromlist[0], &mc.BlockData{newheader, nil}}
			var guard *monkey.PatchGuard
			guard = monkey.Patch(ca.GetAccountOriginalRole, func(account common.Address, number common.Hash) (common.RoleType, error) {
				guard.Unpatch()
				defer guard.Restore()
				return common.RoleDefault, nil
			})
			var guard1 *monkey.PatchGuard
			guard1 = monkey.Patch(crypto.VerifySignWithValidate, func(sighash []byte, sig []byte) (common.Address, bool, error) {
				guard1.Unpatch()
				defer guard1.Restore()
				return common.HexToAddress(testAddress1), true, nil
			})

			blockgen.broadcastMinerResultHandle(broadcastResult)
			p, ok := blockgen.pm.processMap[Number]
			So(ok, ShouldBeTrue)
			So(p.broadcastRstCache, ShouldContain, broadcastResult.BlockMainData)

		})

		Convey("广播矿工挖矿结果测试，当前是nextleader", func() {
			blockgen, err := New(eth)
			if err != nil {
			}
			header := eth.BlockChain().CurrentHeader()
			Number := uint64(99)
			newheader := types.CopyHeader(header)
			newheader.ParentHash = header.Hash()
			newheader.Number = big.NewInt(int64(Number) + 1)
			newheader.Leader = common.HexToAddress(testAddress)
			newheader.Signatures = append(newheader.Signatures, common.BytesToSignature(common.BigToHash(big.NewInt(100)).Bytes()))
			fromlist := []common.Address{common.HexToAddress(testAddress), common.HexToAddress(testAddress1), common.HexToAddress(testAddress2), common.HexToAddress(testAddress3)}
			roleMsg := &mc.RoleUpdatedMsg{Role: common.RoleValidator, BlockNum: uint64(Number), Leader: common.HexToAddress(testAddress)}
			blockgen.roleUpdatedMsgHandle(roleMsg)

			broadcastResult := &mc.HD_BroadcastMiningRspMsg{fromlist[0], &mc.BlockData{newheader, nil}}
			var guard *monkey.PatchGuard
			guard = monkey.Patch(ca.GetAccountOriginalRole, func(account common.Address, number common.Hash) (common.RoleType, error) {
				guard.Unpatch()
				defer guard.Restore()
				return common.RoleDefault, nil
			})
			var guard1 *monkey.PatchGuard
			guard1 = monkey.Patch(crypto.VerifySignWithValidate, func(sighash []byte, sig []byte) (common.Address, bool, error) {
				guard1.Unpatch()
				defer guard1.Restore()
				return common.HexToAddress(testAddress), true, nil
			})
			var guard3 *monkey.PatchGuard
			guard3 = monkey.Patch(ca.GetAddress, func() common.Address {
				guard3.Unpatch()
				defer guard3.Restore()
				return common.HexToAddress(testAddress)
			})
			blockgen.broadcastMinerResultHandle(broadcastResult)
			leaderMsg := &mc.LeaderChangeNotify{ConsensusState: true, Leader: common.HexToAddress(testAddress1), NextLeader: common.HexToAddress(testAddress), Number: Number + 1, ConsensusTurn: mc.ConsensusTurnInfo{}}
			blockgen.leaderChangeNotifyHandle(leaderMsg)

			p, ok := blockgen.pm.processMap[Number+1]
			So(ok, ShouldBeTrue)
			So(p.state, ShouldEqual, StateEnd)

		})
	})
}

//
//func TestBlockGenor_blockInsertMsgHandle(t *testing.T) {
//	log.InitLog(3)
//	eth := fakeEthNew(0)
//	if nil == eth {
//		fmt.Println("failed to create eth")
//	}
//	Convey("区块插入测试", t, func() {
//		Convey("高度相同，本地共识不过", func() {
//			blockgen, err := New(eth)
//			if err != nil {
//			}
//			header := types.CopyHeader(eth.BlockChain().CurrentHeader())
//			header.ParentHash = header.Hash()
//			header.Number = big.NewInt(2)
//			roleMsg := &mc.RoleUpdatedMsg{common.RoleValidator, uint64(1), common.HexToAddress(testAddress)}
//			blockgen.roleUpdatedMsgHandle(roleMsg)
//
//			blockInsert := &mc.HD_BlockInsertNotify{common.HexToAddress(testAddress), header}
//			blockgen.blockInsertMsgHandle(blockInsert)
//			_, ok := blockgen.pm.processMap[2]
//			So(ok, ShouldBeTrue)
//			So(eth.fetchhash, ShouldEqual, header.Hash())
//			So(eth.fetchnum, ShouldEqual, 2)
//		})
//	})
//
//	Convey("区块插入测试", t, func() {
//		Convey("高度相同，本地共识通过", func() {
//			blockgen, err := New(eth)
//			if err != nil {
//			}
//			header := types.CopyHeader(eth.BlockChain().CurrentHeader())
//			header.ParentHash = header.Hash()
//			header.Number = big.NewInt(2)
//			state, _ := eth.BlockChain().State()
//			roleMsg := &mc.RoleUpdatedMsg{common.RoleValidator, uint64(1), common.HexToAddress(testAddress)}
//			blockgen.roleUpdatedMsgHandle(roleMsg)
//			p := blockgen.pm.GetCurrentProcess()
//			p.genBlockData = &mc.BlockVerifyConsensusOK{header, common.Hash{}, nil, nil, state}
//			blockInsert := &mc.HD_BlockInsertNotify{common.HexToAddress(testAddress), header}
//			blockgen.blockInsertMsgHandle(blockInsert)
//			_, ok := blockgen.pm.processMap[2]
//			So(ok, ShouldBeTrue)
//			So(header.Hash(), ShouldEqual, eth.BlockChain().CurrentHeader().Hash())
//			So(eth.BlockChain().CurrentHeader().Number.Int64(), ShouldEqual, int64(2))
//		})
//	})
//
//	Convey("区块插入测试", t, func() {
//		Convey("高度大于当前处理高度，启动fetch流程", func() {
//			blockgen, err := New(eth)
//			if err != nil {
//			}
//			header := types.CopyHeader(eth.BlockChain().CurrentHeader())
//			header.ParentHash = header.Hash()
//			header.Number = big.NewInt(10)
//			roleMsg := &mc.RoleUpdatedMsg{common.RoleValidator, uint64(1), common.HexToAddress(testAddress)}
//			blockgen.roleUpdatedMsgHandle(roleMsg)
//
//			blockInsert := &mc.HD_BlockInsertNotify{common.HexToAddress(testAddress), header}
//			blockgen.blockInsertMsgHandle(blockInsert)
//			_, ok := blockgen.pm.processMap[2]
//			So(ok, ShouldBeTrue)
//			So(eth.fetchhash, ShouldEqual, header.Hash())
//			So(eth.fetchnum, ShouldEqual, 10)
//		})
//	})
//}
//
func TestBlockGenor_consensusBlockMsgHandle(t *testing.T) {
	log.InitLog(3)

	eth := fakeEthNew(0)
	if nil == eth {
		fmt.Println("failed to create eth")
	}
	monkey.Patch(manparams.NewBCIntervalByHash, func(blockHash common.Hash) (*manparams.BCInterval, error) {
		fmt.Println("use monkey NewBCIntervalByNumber")

		inteval1 := &mc.BCIntervalInfo{LastBCNumber: 0, LastReelectNumber: 0, BCInterval: 100}

		interval2, _ := manparams.NewBCIntervalWithInterval(inteval1)
		return interval2, nil
	})
	var guard *monkey.PatchGuard
	guard = monkey.Patch(ca.GetAddress, func() common.Address {
		guard.Unpatch()
		defer guard.Restore()
		return common.HexToAddress(testAddress)
	})

	Convey("本地共识结果和矿工挖坑结果测试", t, func() {

		SkipConvey("本地共识结果测试高度1,矿工挖坑结果高度1，本地高度0", func() {
			blockgen, err := New(eth)
			if err != nil {
			}
			currentNum := uint64(0)
			roleMsg := &mc.RoleUpdatedMsg{Role: common.RoleValidator, BlockNum: currentNum, Leader: common.HexToAddress(testAddress)}
			blockgen.roleUpdatedMsgHandle(roleMsg)

			blockhash := common.Hash{0x01}
			diff := big.NewInt(100)
			from := common.HexToAddress(testAddress)
			Number := currentNum + 1
			Signatures := []common.Signature{common.BytesToSignature(common.BigToHash(big.NewInt(100)).Bytes()), common.BytesToSignature(common.BigToHash(big.NewInt(101)).Bytes()), common.BytesToSignature(common.BigToHash(big.NewInt(102)).Bytes())}
			fromlist := []common.Address{common.HexToAddress(testAddress), common.HexToAddress(testAddress1), common.HexToAddress(testAddress2), common.HexToAddress(testAddress3)}
			minerResults := make([]*mc.HD_MiningRspMsg, 0)
			for i := 0; i < len(fromlist); i++ {
				tempResult := &mc.HD_MiningRspMsg{fromlist[i], Number, blockhash, diff, types.EncodeNonce(uint64(0)), from, common.Hash{0x01}, Signatures}
				minerResults = append(minerResults, tempResult)
				blockgen.minerResultHandle(tempResult)
			}

			header := eth.BlockChain().CurrentHeader()
			newheader := types.CopyHeader(header)
			newheader.Number = big.NewInt(int64(currentNum + 1))
			now := time.Now().Unix()
			newheader.Time = big.NewInt(now)
			state, err := eth.blockchain.StateAt(eth.blockchain.GetBlockByHash(header.Hash()).Root())
			So(err, ShouldBeNil)
			blockConsensus := &mc.BlockLocalVerifyOK{Header: newheader, BlockHash: common.Hash{0x10}, OriginalTxs: nil, FinalTxs: nil, Receipts: nil, State: state}
			blockgen.consensusBlockMsgHandle(blockConsensus)
			p, ok := blockgen.pm.processMap[1]
			So(ok, ShouldBeTrue)
			So(p.blockCache.GetBlockData(blockConsensus.Header.Leader), ShouldNotBeNil)
			So(p.state, ShouldEqual, StateBlockBroadcast)

		})

		Convey("本地共识结果测试，高度1,矿工挖坑结果，高度1，本地高度0,当前是leader", func() {
			blockgen, err := New(eth)
			if err != nil {
			}
			currentNum := uint64(0)
			roleMsg := &mc.RoleUpdatedMsg{Role: common.RoleValidator, BlockNum: currentNum, Leader: common.HexToAddress(testAddress)}
			blockgen.roleUpdatedMsgHandle(roleMsg)
			leaderMsg := &mc.LeaderChangeNotify{ConsensusState: true, Leader: common.HexToAddress(testAddress), NextLeader: common.HexToAddress(testAddress1), Number: currentNum + 1, ConsensusTurn: mc.ConsensusTurnInfo{}}
			blockgen.leaderChangeNotifyHandle(leaderMsg)
			blockhash := common.Hash{0x01}
			diff := big.NewInt(100)
			from := common.HexToAddress(testAddress)
			Number := currentNum + 1
			Signatures := []common.Signature{common.BytesToSignature(common.BigToHash(big.NewInt(100)).Bytes()), common.BytesToSignature(common.BigToHash(big.NewInt(101)).Bytes()), common.BytesToSignature(common.BigToHash(big.NewInt(102)).Bytes())}
			fromlist := []common.Address{common.HexToAddress(testAddress), common.HexToAddress(testAddress1), common.HexToAddress(testAddress2), common.HexToAddress(testAddress3)}
			minerResults := make([]*mc.HD_MiningRspMsg, 0)
			for i := 0; i < len(fromlist); i++ {
				tempResult := &mc.HD_MiningRspMsg{fromlist[i], Number, blockhash, diff, types.EncodeNonce(uint64(0)), from, common.Hash{0x01}, Signatures}
				minerResults = append(minerResults, tempResult)
				blockgen.minerResultHandle(tempResult)
			}

			header := eth.BlockChain().CurrentHeader()
			p, ok := blockgen.pm.processMap[1]
			p.preBlockHash = header.Hash()
			newheader := types.CopyHeader(header)
			newheader.Number = big.NewInt(int64(currentNum + 1))
			now := time.Now().Unix()
			newheader.Time = big.NewInt(now)
			state, err := eth.blockchain.StateAt(eth.blockchain.GetBlockByHash(header.Hash()).Root())
			So(err, ShouldBeNil)
			blockConsensus := &mc.BlockLocalVerifyOK{newheader, common.Hash{0x10}, nil, nil, nil, state}
			blockgen.consensusBlockMsgHandle(blockConsensus)
			p, ok = blockgen.pm.processMap[1]
			So(ok, ShouldBeTrue)
			So(p.blockCache.GetBlockData(blockConsensus.Header.Leader), ShouldNotBeNil)
			So(p.state, ShouldEqual, StateMinerResultVerify)

		})
		//
		//Convey("本地共识结果测试，高度1,矿工挖坑结果，高度1，本地高度0,当前是nextleader", func() {
		//	blockgen, err := New(eth)
		//	if err != nil {
		//	}
		//	currentNum := uint64(0)
		//	roleMsg := &mc.RoleUpdatedMsg{Role: common.RoleValidator, BlockNum: currentNum, Leader: common.HexToAddress(testAddress)}
		//	blockgen.roleUpdatedMsgHandle(roleMsg)
		//
		//	var guard *monkey.PatchGuard
		//	guard = monkey.Patch(ca.GetAddress, func() common.Address {
		//		guard.Unpatch()
		//		defer guard.Restore()
		//		return common.HexToAddress(testAddress)
		//	})
		//	leaderMsg := &mc.LeaderChangeNotify{true, common.HexToAddress(testAddress1), common.HexToAddress(testAddress), currentNum + 1, 0}
		//	blockgen.leaderChangeNotifyHandle(leaderMsg)
		//
		//	diff := big.NewInt(100)
		//	Number := currentNum + 1
		//	Signatures := []common.Signature{common.BytesToSignature(common.BigToHash(big.NewInt(100)).Bytes()), common.BytesToSignature(common.BigToHash(big.NewInt(101)).Bytes()), common.BytesToSignature(common.BigToHash(big.NewInt(102)).Bytes())}
		//	fromlist := []common.Address{common.HexToAddress(testAddress), common.HexToAddress(testAddress1), common.HexToAddress(testAddress2), common.HexToAddress(testAddress3)}
		//	minerResults := make([]*mc.HD_MiningRspMsg, 0)
		//	header := eth.BlockChain().CurrentHeader()
		//	newheader := types.CopyHeader(header)
		//	newheader.ParentHash = header.Hash()
		//	newheader.Number = big.NewInt(int64(currentNum + 1))
		//	newheader.Difficulty = diff
		//	now := time.Now().Unix()
		//	newheader.Time = big.NewInt(now)
		//	blockhash := newheader.HashNoSignsAndNonce()
		//	for i := 0; i < len(fromlist); i++ {
		//		tempResult := &mc.HD_MiningRspMsg{fromlist[i], Number, blockhash, diff, types.EncodeNonce(uint64(0)), fromlist[i], common.Hash{0x01}, Signatures}
		//		minerResults = append(minerResults, tempResult)
		//		blockgen.minerResultHandle(tempResult)
		//	}
		//
		//	state, err := eth.blockchain.StateAt(eth.blockchain.GetBlockByHash(header.Hash()).Root())
		//	So(err, ShouldBeNil)
		//	blockConsensus := &mc.BlockVerifyConsensusOK{newheader, blockhash, nil, nil, state}
		//	blockgen.consensusBlockMsgHandle(blockConsensus)
		//	p, ok := blockgen.pm.processMap[1]
		//	So(ok, ShouldBeTrue)
		//	So(p.state, ShouldEqual, StateEnd)
		//
		//})
		//
		//Convey("本地共识结果测试，高度1,矿工挖坑结果，高度1，本地高度0,当前是nextleader,leader消息再共识后面处理", func() {
		//	blockgen, err := New(eth)
		//	if err != nil {
		//	}
		//	currentNum := uint64(0)
		//	roleMsg := &mc.RoleUpdatedMsg{Role: common.RoleValidator, BlockNum: currentNum, Leader: common.HexToAddress(testAddress)}
		//	blockgen.roleUpdatedMsgHandle(roleMsg)
		//
		//	var guard *monkey.PatchGuard
		//	guard = monkey.Patch(ca.GetAddress, func() common.Address {
		//		guard.Unpatch()
		//		defer guard.Restore()
		//		return common.HexToAddress(testAddress)
		//	})
		//
		//	diff := big.NewInt(100)
		//	Number := currentNum + 1
		//	Signatures := []common.Signature{common.BytesToSignature(common.BigToHash(big.NewInt(100)).Bytes()), common.BytesToSignature(common.BigToHash(big.NewInt(101)).Bytes()), common.BytesToSignature(common.BigToHash(big.NewInt(102)).Bytes())}
		//	fromlist := []common.Address{common.HexToAddress(testAddress), common.HexToAddress(testAddress1), common.HexToAddress(testAddress2), common.HexToAddress(testAddress3)}
		//	minerResults := make([]*mc.HD_MiningRspMsg, 0)
		//	header := eth.BlockChain().CurrentHeader()
		//	newheader := types.CopyHeader(header)
		//	newheader.ParentHash = header.Hash()
		//	newheader.Number = big.NewInt(int64(currentNum + 1))
		//	newheader.Difficulty = diff
		//	now := time.Now().Unix()
		//	newheader.Time = big.NewInt(now)
		//	newheader.Leader = common.HexToAddress(testAddress1)
		//	blockhash := newheader.HashNoSignsAndNonce()
		//	for i := 0; i < len(fromlist); i++ {
		//		tempResult := &mc.HD_MiningRspMsg{fromlist[i], Number, blockhash, diff, types.EncodeNonce(uint64(0)), fromlist[i], common.Hash{0x01}, Signatures}
		//		minerResults = append(minerResults, tempResult)
		//		blockgen.minerResultHandle(tempResult)
		//	}
		//
		//	state, err := eth.blockchain.StateAt(eth.blockchain.GetBlockByHash(header.Hash()).Root())
		//	So(err, ShouldBeNil)
		//	blockConsensus := &mc.BlockVerifyConsensusOK{newheader, blockhash, nil, nil, state}
		//	blockgen.consensusBlockMsgHandle(blockConsensus)
		//
		//	leaderMsg := &mc.LeaderChangeNotify{true, common.HexToAddress(testAddress1), common.HexToAddress(testAddress), currentNum + 1, 0}
		//	blockgen.leaderChangeNotifyHandle(leaderMsg)
		//
		//	p, ok := blockgen.pm.processMap[1]
		//	So(ok, ShouldBeTrue)
		//	So(p.state, ShouldEqual, StateEnd)
		//
		//})
	})

}

//
//func TestBlockGenor_MulBlockMsgHandle(t *testing.T) {
//	log.InitLog(3)
//
//	eth := fakeEthNew(0)
//	if nil == eth {
//		fmt.Println("failed to create eth")
//	}
//	Convey("多难度测试", t, func() {
//		Convey("难度减半测试", func() {
//
//			blockgen, err := New(eth)
//			if err != nil {
//			}
//			currentNum := uint64(0)
//			roleMsg := &mc.RoleUpdatedMsg{common.RoleValidator, currentNum, common.HexToAddress(testAddress)}
//			blockgen.roleUpdatedMsgHandle(roleMsg)
//
//			var guard *monkey.PatchGuard
//			guard = monkey.Patch(ca.GetAddress, func() common.Address {
//				guard.Unpatch()
//				defer guard.Restore()
//				return common.HexToAddress(testAddress)
//			})
//			//leader消息
//			leaderMsg := &mc.LeaderChangeNotify{true, common.HexToAddress(testAddress1), common.HexToAddress(testAddress), currentNum + 1, 0}
//			blockgen.leaderChangeNotifyHandle(leaderMsg)
//			//挖矿消息
//			diff := big.NewInt(100)
//			Number := currentNum + 1
//			Signatures := []common.Signature{common.BytesToSignature(common.BigToHash(big.NewInt(100)).Bytes()), common.BytesToSignature(common.BigToHash(big.NewInt(101)).Bytes()), common.BytesToSignature(common.BigToHash(big.NewInt(102)).Bytes())}
//			fromlist := []common.Address{common.HexToAddress(testAddress), common.HexToAddress(testAddress1), common.HexToAddress(testAddress2), common.HexToAddress(testAddress3)}
//			minerResults := make([]*mc.HD_MiningRspMsg, 0)
//			header := eth.BlockChain().CurrentHeader()
//			newheader := types.CopyHeader(header)
//			newheader.ParentHash = header.Hash()
//			newheader.Number = big.NewInt(int64(currentNum + 1))
//			newheader.Difficulty = diff
//			newheader.Leader = common.HexToAddress(testAddress1)
//			now := time.Now().Unix()
//			newheader.Time = big.NewInt(now)
//			blockhash := newheader.HashNoSignsAndNonce()
//			for i := 0; i < len(fromlist); i++ {
//				tempResult := &mc.HD_MiningRspMsg{fromlist[i], Number, blockhash, big.NewInt(newheader.Difficulty.Int64() / 2), types.EncodeNonce(uint64(0)), fromlist[i], common.Hash{0x01}, Signatures}
//				minerResults = append(minerResults, tempResult)
//				blockgen.minerResultHandle(tempResult)
//			}
//
//			state, err := eth.blockchain.StateAt(eth.blockchain.GetBlockByHash(header.Hash()).Root())
//			So(err, ShouldBeNil)
//			//共识消息
//			blockConsensus := &mc.BlockVerifyConsensusOK{newheader, blockhash, nil, nil, state}
//			blockgen.consensusBlockMsgHandle(blockConsensus)
//			p, ok := blockgen.pm.processMap[1]
//			So(ok, ShouldBeTrue)
//			time.Sleep(time.Second * time.Duration(params.MinerPickTimeout+5))
//			So(p.state, ShouldEqual, StateEnd)
//		})
//		Convey("基金会矿工20s超时测试，原始难度", func() {
//			blockgen, err := New(eth)
//			if err != nil {
//			}
//			currentNum := uint64(0)
//			roleMsg := &mc.RoleUpdatedMsg{common.RoleValidator, currentNum, common.HexToAddress(testAddress)}
//			blockgen.roleUpdatedMsgHandle(roleMsg)
//
//			var guard *monkey.PatchGuard
//			guard = monkey.Patch(ca.GetAddress, func() common.Address {
//				guard.Unpatch()
//				defer guard.Restore()
//				return common.HexToAddress(testAddress)
//			})
//
//			var guard1 *monkey.PatchGuard
//			//基金会矿工打桩
//			guard1 = monkey.Patch(ca.GetAccountOriginalRole, func(account common.Address, number uint64) (common.RoleType, error) {
//				guard1.Unpatch()
//				defer guard1.Restore()
//				if account.Equal(common.HexToAddress(testAddress)) {
//					return common.RoleInnerMiner, nil
//				}
//				return common.RoleMiner, nil
//			})
//			//leader消息
//			leaderMsg := &mc.LeaderChangeNotify{true, common.HexToAddress(testAddress1), common.HexToAddress(testAddress), currentNum + 1, 0}
//			blockgen.leaderChangeNotifyHandle(leaderMsg)
//			//挖矿消息
//			diff := big.NewInt(100)
//			Number := currentNum + 1
//			Signatures := []common.Signature{common.BytesToSignature(common.BigToHash(big.NewInt(100)).Bytes()), common.BytesToSignature(common.BigToHash(big.NewInt(101)).Bytes()), common.BytesToSignature(common.BigToHash(big.NewInt(102)).Bytes())}
//			fromlist := []common.Address{common.HexToAddress(testAddress), common.HexToAddress(testAddress1), common.HexToAddress(testAddress2), common.HexToAddress(testAddress3)}
//			minerResults := make([]*mc.HD_MiningRspMsg, 0)
//			header := eth.BlockChain().CurrentHeader()
//			newheader := types.CopyHeader(header)
//			newheader.ParentHash = header.Hash()
//			newheader.Number = big.NewInt(int64(currentNum + 1))
//			newheader.Difficulty = diff
//			now := time.Now().Unix()
//			newheader.Time = big.NewInt(now)
//			newheader.Leader = common.HexToAddress(testAddress1)
//			blockhash := newheader.HashNoSignsAndNonce()
//			freeMinerResult := &mc.HD_MiningRspMsg{fromlist[0], Number, blockhash, newheader.Difficulty, types.EncodeNonce(uint64(0)), fromlist[0], common.Hash{0x01}, Signatures}
//			blockgen.minerResultHandle(freeMinerResult)
//			for i := 1; i < len(fromlist); i++ {
//				tempResult := &mc.HD_MiningRspMsg{fromlist[i], Number, blockhash, big.NewInt(newheader.Difficulty.Int64() / 2), types.EncodeNonce(uint64(0)), fromlist[i], common.Hash{0x01}, Signatures}
//				minerResults = append(minerResults, tempResult)
//				blockgen.minerResultHandle(tempResult)
//			}
//
//			state, err := eth.blockchain.StateAt(eth.blockchain.GetBlockByHash(header.Hash()).Root())
//			So(err, ShouldBeNil)
//			//共识消息
//			blockConsensus := &mc.BlockVerifyConsensusOK{newheader, blockhash, nil, nil, state}
//			blockgen.consensusBlockMsgHandle(blockConsensus)
//			p, ok := blockgen.pm.processMap[1]
//			So(ok, ShouldBeTrue)
//			time.Sleep(time.Second * time.Duration(params.MinerPickTimeout+5))
//			So(p.minerPickTimer, ShouldEqual, nil)
//			So(p.state, ShouldEqual, StateEnd)
//			So(common.HexToAddress(testAddress), ShouldEqual, eth.BlockChain().CurrentHeader().Coinbase)
//			So(eth.BlockChain().CurrentHeader().Number.Int64(), ShouldEqual, int64(1))
//		})
//
//		Convey("基金会矿工20s超时测试，1/2原始难度", func() {
//			blockgen, err := New(eth)
//			if err != nil {
//			}
//			currentNum := uint64(0)
//			roleMsg := &mc.RoleUpdatedMsg{common.RoleValidator, currentNum, common.HexToAddress(testAddress)}
//			blockgen.roleUpdatedMsgHandle(roleMsg)
//
//			var guard *monkey.PatchGuard
//			guard = monkey.Patch(ca.GetAddress, func() common.Address {
//				guard.Unpatch()
//				defer guard.Restore()
//				return common.HexToAddress(testAddress)
//			})
//
//			var guard1 *monkey.PatchGuard
//			//基金会矿工打桩
//			guard1 = monkey.Patch(ca.GetAccountOriginalRole, func(account common.Address, number uint64) (common.RoleType, error) {
//				guard1.Unpatch()
//				defer guard1.Restore()
//				if account.Equal(common.HexToAddress(testAddress)) {
//					return common.RoleInnerMiner, nil
//				}
//				return common.RoleMiner, nil
//			})
//			//leader消息
//			leaderMsg := &mc.LeaderChangeNotify{true, common.HexToAddress(testAddress1), common.HexToAddress(testAddress), currentNum + 1, 0}
//			blockgen.leaderChangeNotifyHandle(leaderMsg)
//			//挖矿消息
//			diff := big.NewInt(100)
//			Number := currentNum + 1
//			Signatures := []common.Signature{common.BytesToSignature(common.BigToHash(big.NewInt(100)).Bytes()), common.BytesToSignature(common.BigToHash(big.NewInt(101)).Bytes()), common.BytesToSignature(common.BigToHash(big.NewInt(102)).Bytes())}
//			fromlist := []common.Address{common.HexToAddress(testAddress), common.HexToAddress(testAddress1), common.HexToAddress(testAddress2), common.HexToAddress(testAddress3)}
//			minerResults := make([]*mc.HD_MiningRspMsg, 0)
//			header := eth.BlockChain().CurrentHeader()
//			newheader := types.CopyHeader(header)
//			newheader.ParentHash = header.Hash()
//			newheader.Number = big.NewInt(int64(currentNum + 1))
//			newheader.Difficulty = diff
//			newheader.Leader = common.HexToAddress(testAddress1)
//			now := time.Now().Unix()
//			newheader.Time = big.NewInt(now)
//			blockhash := newheader.HashNoSignsAndNonce()
//			state, err := eth.blockchain.StateAt(eth.blockchain.GetBlockByHash(header.Hash()).Root())
//			So(err, ShouldBeNil)
//			//共识消息
//			blockConsensus := &mc.BlockVerifyConsensusOK{newheader, blockhash, nil, nil, state}
//			blockgen.consensusBlockMsgHandle(blockConsensus)
//
//			freeMinerResult := &mc.HD_MiningRspMsg{fromlist[0], Number, blockhash, big.NewInt(newheader.Difficulty.Int64() / 2), types.EncodeNonce(uint64(0)), fromlist[0], common.Hash{0x01}, Signatures}
//			blockgen.minerResultHandle(freeMinerResult)
//			for i := 1; i < len(fromlist); i++ {
//				tempResult := &mc.HD_MiningRspMsg{fromlist[i], Number, blockhash, big.NewInt(newheader.Difficulty.Int64() / 2), types.EncodeNonce(uint64(0)), fromlist[i], common.Hash{0x01}, Signatures}
//				minerResults = append(minerResults, tempResult)
//				blockgen.minerResultHandle(tempResult)
//			}
//
//			p, ok := blockgen.pm.processMap[1]
//			So(ok, ShouldBeTrue)
//			time.Sleep(time.Second * time.Duration(params.MinerPickTimeout+5))
//			So(p.minerPickTimer, ShouldEqual, nil)
//			So(p.state, ShouldEqual, StateEnd)
//			So(common.HexToAddress(testAddress), ShouldEqual, eth.BlockChain().CurrentHeader().Coinbase)
//			So(eth.BlockChain().CurrentHeader().Number.Int64(), ShouldEqual, int64(1))
//		})
//
//		Convey("基金会矿工最低难度测试", func() {
//			blockgen, err := New(eth)
//			if err != nil {
//			}
//			currentNum := uint64(0)
//			roleMsg := &mc.RoleUpdatedMsg{common.RoleValidator, currentNum, common.HexToAddress(testAddress)}
//			blockgen.roleUpdatedMsgHandle(roleMsg)
//
//			var guard *monkey.PatchGuard
//			guard = monkey.Patch(ca.GetAddress, func() common.Address {
//				guard.Unpatch()
//				defer guard.Restore()
//				return common.HexToAddress(testAddress)
//			})
//
//			var guard1 *monkey.PatchGuard
//			//基金会矿工打桩
//			guard1 = monkey.Patch(ca.GetAccountOriginalRole, func(account common.Address, number uint64) (common.RoleType, error) {
//				guard1.Unpatch()
//				defer guard1.Restore()
//				if account.Equal(common.HexToAddress(testAddress)) {
//					return common.RoleInnerMiner, nil
//				}
//				return common.RoleMiner, nil
//			})
//			//leader消息
//			leaderMsg := &mc.LeaderChangeNotify{true, common.HexToAddress(testAddress1), common.HexToAddress(testAddress), currentNum + 1, 0}
//			blockgen.leaderChangeNotifyHandle(leaderMsg)
//			//挖矿消息
//			diff := big.NewInt(100)
//			Number := currentNum + 1
//			Signatures := []common.Signature{common.BytesToSignature(common.BigToHash(big.NewInt(100)).Bytes()), common.BytesToSignature(common.BigToHash(big.NewInt(101)).Bytes()), common.BytesToSignature(common.BigToHash(big.NewInt(102)).Bytes())}
//			fromlist := []common.Address{common.HexToAddress(testAddress), common.HexToAddress(testAddress1), common.HexToAddress(testAddress2), common.HexToAddress(testAddress3)}
//			minerResults := make([]*mc.HD_MiningRspMsg, 0)
//			header := eth.BlockChain().CurrentHeader()
//			newheader := types.CopyHeader(header)
//			newheader.ParentHash = header.Hash()
//			newheader.Number = big.NewInt(int64(currentNum + 1))
//			newheader.Difficulty = diff
//			newheader.Leader = common.HexToAddress(testAddress1)
//			now := time.Now().Unix()
//			newheader.Time = big.NewInt(now)
//			blockhash := newheader.HashNoSignsAndNonce()
//			state, err := eth.blockchain.StateAt(eth.blockchain.GetBlockByHash(header.Hash()).Root())
//			So(err, ShouldBeNil)
//			//共识消息
//			blockConsensus := &mc.BlockVerifyConsensusOK{newheader, blockhash, nil, nil, state}
//			blockgen.consensusBlockMsgHandle(blockConsensus)
//
//			freeMinerResult := &mc.HD_MiningRspMsg{fromlist[0], Number, blockhash, big.NewInt(newheader.Difficulty.Int64() / 50), types.EncodeNonce(uint64(0)), fromlist[0], common.Hash{0x01}, Signatures}
//			blockgen.minerResultHandle(freeMinerResult)
//			for i := 1; i < len(fromlist); i++ {
//				tempResult := &mc.HD_MiningRspMsg{fromlist[i], Number, blockhash, big.NewInt(newheader.Difficulty.Int64() / 50), types.EncodeNonce(uint64(0)), fromlist[i], common.Hash{0x01}, Signatures}
//				minerResults = append(minerResults, tempResult)
//				blockgen.minerResultHandle(tempResult)
//			}
//
//			p, ok := blockgen.pm.processMap[1]
//			So(ok, ShouldBeTrue)
//			time.Sleep(time.Second * time.Duration(100+5))
//			So(p.minerPickTimer, ShouldEqual, nil)
//			So(p.state, ShouldEqual, StateEnd)
//			So(common.HexToAddress(testAddress), ShouldEqual, eth.BlockChain().CurrentHeader().Coinbase)
//			So(eth.BlockChain().CurrentHeader().Number.Int64(), ShouldEqual, int64(1))
//		})
//	})
//}
//
//func TestBlockGenor_RandMsgHandle(t *testing.T) {
//	log.InitLog(3)
//
//	eth := fakeEthNew(0)
//	if nil == eth {
//		fmt.Println("failed to create eth")
//	}
//	Convey("同一高度消息乱序测试", t, func() {
//
//		currentNum := uint64(0)
//		var guard1 *monkey.PatchGuard
//		var guard *monkey.PatchGuard
//		guard = monkey.Patch(ca.GetAddress, func() common.Address {
//			guard.Unpatch()
//			defer guard.Restore()
//			return common.HexToAddress(testAddress)
//		})
//
//		guard1 = monkey.Patch(ca.GetAccountOriginalRole, func(account common.Address, number uint64) (common.RoleType, error) {
//			guard1.Unpatch()
//			defer guard1.Restore()
//			return common.RoleMiner, nil
//		})
//		for {
//			blockgen, err := New(eth)
//			if err != nil {
//			}
//			roleMsg := &mc.RoleUpdatedMsg{common.RoleBroadcast, currentNum, common.HexToAddress(testAddress)}
//			go func() {
//				time.Sleep(time.Duration(rand.Uint32()%8) * time.Second)
//				blockgen.roleUpdatedMsgHandle(roleMsg)
//			}()
//			go func() {
//				roleMsg2 := &mc.RoleUpdatedMsg{common.RoleBroadcast, currentNum + 1, common.HexToAddress(testAddress)}
//				time.Sleep(time.Duration(rand.Uint32()%5) * time.Second)
//				blockgen.roleUpdatedMsgHandle(roleMsg2)
//			}()
//			diff := big.NewInt(100)
//			Number := currentNum + 1
//			Signatures := []common.Signature{common.BytesToSignature(common.BigToHash(big.NewInt(100)).Bytes()), common.BytesToSignature(common.BigToHash(big.NewInt(101)).Bytes()), common.BytesToSignature(common.BigToHash(big.NewInt(102)).Bytes())}
//			fromlist := []common.Address{common.HexToAddress(testAddress), common.HexToAddress(testAddress1), common.HexToAddress(testAddress2), common.HexToAddress(testAddress3)}
//			minerResults := make([]*mc.HD_MiningRspMsg, 0)
//			header := eth.BlockChain().CurrentHeader()
//			newheader := types.CopyHeader(header)
//			newheader.ParentHash = header.Hash()
//			newheader.Number = big.NewInt(int64(currentNum + 1))
//			newheader.Difficulty = diff
//			now := time.Now().Unix()
//			newheader.Time = big.NewInt(now)
//			newheader.Leader = common.HexToAddress(testAddress1)
//			blockhash := newheader.HashNoSignsAndNonce()
//
//			newheader2 := types.CopyHeader(newheader)
//			newheader2.Leader = common.HexToAddress(testAddress2)
//			difflist := []*big.Int{big.NewInt(100 / 2), big.NewInt(100 / 10), big.NewInt(100 / 50)}
//			for i := 0; i < len(fromlist); i++ {
//				tempResult := &mc.HD_MiningRspMsg{fromlist[i], Number, blockhash, difflist[2], types.EncodeNonce(uint64(0)), fromlist[i], common.Hash{0x01}, Signatures}
//				minerResults = append(minerResults, tempResult)
//				go func() {
//					time.Sleep(time.Duration(rand.Uint32()%5) * time.Second)
//					blockgen.minerResultHandle(tempResult)
//				}()
//			}
//
//			state, err := eth.blockchain.StateAt(eth.blockchain.GetBlockByHash(header.Hash()).Root())
//			So(err, ShouldBeNil)
//			blockConsensus := &mc.BlockVerifyConsensusOK{newheader, blockhash, nil, nil, state}
//			go func() {
//				time.Sleep(time.Duration(rand.Uint32()%5) * time.Second)
//				blockgen.consensusBlockMsgHandle(blockConsensus)
//			}()
//
//			leaderMsg := &mc.LeaderChangeNotify{true, common.HexToAddress(testAddress1), common.HexToAddress(testAddress), currentNum + 1, 0}
//			go func() {
//				time.Sleep(time.Duration(rand.Uint32()%5) * time.Second)
//				blockgen.leaderChangeNotifyHandle(leaderMsg)
//			}()
//			time.Sleep(10 * time.Second)
//			_, ok := blockgen.pm.processMap[1]
//			So(ok, ShouldBeTrue)
//			//So(p.state, ShouldEqual, StateEnd)
//		}
//
//	})
//
//}
