// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package blkgenor

import (
	"bou.ke/monkey"
	"fmt"
	"github.com/MatrixAINetwork/go-matrix/params/manparams"
	"testing"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	. "github.com/smartystreets/goconvey/convey"
)

func TestProcess_processHeaderGen(t *testing.T) {
	log.InitLog(3)
	eth := fakeEthNew(0)
	if nil == eth {
		fmt.Println("failed to create eth")
	}

	blockgen, err := New(eth)
	if err != nil {
	}

	//process := newProcess(1, pm)
	Convey("可否生成普通验证区块头函数测试", t, func() {
		monkey.Patch(manparams.NewBCIntervalByHash, func(blockHash common.Hash) (*manparams.BCInterval, error) {
			fmt.Println("use monkey NewBCIntervalByNumber")

			inteval1 := &mc.BCIntervalInfo{LastBCNumber: 0, LastReelectNumber: 0, BCInterval: 100}

			interval2, _ := manparams.NewBCIntervalWithInterval(inteval1)
			return interval2, nil
		})
		Convey("当前高度0，可否生成高度1验证区块头函数测试", func() {
			roleMsg := &mc.RoleUpdatedMsg{Role: common.RoleValidator, BlockNum: uint64(0), BlockHash: common.Hash{1}, Leader: common.HexToAddress(testAddress), IsSuperBlock: false}
			blockgen.roleUpdatedMsgHandle(roleMsg)
			//p := blockgen.pm.GetCurrentProcess()
			//err := p.processHeaderGen()
			So(err, ShouldEqual, nil)

		})

		//Convey("当前高度1，可否生成高度2验证区块头函数测试", func() {
		//	roleMsg := &mc.RoleUpdatedMsg{common.RoleValidator, uint64(1), common.HexToAddress(testAddress)}
		//	blockgen.roleUpdatedMsgHandle(roleMsg)
		//	p := blockgen.pm.GetCurrentProcess()
		//	p.preBlockHash = eth.blockchain.GetBlockByNumber(0).Hash()
		//	err := p.processHeaderGen()
		//	So(err, ShouldEqual, nil)
		//
		//})
		//
		//Convey("当前高度99，可否生成高度100验证区块头函数测试", func() {
		//	roleMsg := &mc.RoleUpdatedMsg{common.RoleValidator, uint64(common.GetBroadcastInterval() - 1), common.HexToAddress(testAddress)}
		//	blockgen.roleUpdatedMsgHandle(roleMsg)
		//	p := blockgen.pm.GetCurrentProcess()
		//	p.preBlockHash = eth.blockchain.GetBlockByNumber(0).Hash()
		//	err := p.processHeaderGen()
		//	So(err, ShouldEqual, nil)
		//
		//})Convey("当前高度1，可否生成高度2验证区块头函数测试", func() {
		//	roleMsg := &mc.RoleUpdatedMsg{common.RoleValidator, uint64(1), common.HexToAddress(testAddress)}
		//	blockgen.roleUpdatedMsgHandle(roleMsg)
		//	p := blockgen.pm.GetCurrentProcess()
		//	p.preBlockHash = eth.blockchain.GetBlockByNumber(0).Hash()
		//	err := p.processHeaderGen()
		//	So(err, ShouldEqual, nil)
		//
		//})
		//
		//Convey("当前高度99，可否生成高度100验证区块头函数测试", func() {
		//	roleMsg := &mc.RoleUpdatedMsg{common.RoleValidator, uint64(common.GetBroadcastInterval() - 1), common.HexToAddress(testAddress)}
		//	blockgen.roleUpdatedMsgHandle(roleMsg)
		//	p := blockgen.pm.GetCurrentProcess()
		//	p.preBlockHash = eth.blockchain.GetBlockByNumber(0).Hash()
		//	err := p.processHeaderGen()
		//	So(err, ShouldEqual, nil)
		//
		//})
	})
}

//
//func TestProcess_processBroadcastHeaderGen(t *testing.T) {
//	log.InitLog(3)
//	eth := fakeEthNew(0)
//	if nil == eth {
//		fmt.Println("failed to create eth")
//	}
//
//	blockgen, err := New(eth)
//	if err != nil {
//	}
//
//	//process := newProcess(1, pm)
//	Convey("可否生成验证区块头函数测试", t, func() {
//		Convey("当前高度99，可否生成高度1验证区块头函数测试", func() {
//			roleMsg := &mc.RoleUpdatedMsg{common.RoleValidator, uint64(common.GetBroadcastInterval() - 1), common.HexToAddress(testAddress)}
//			blockgen.roleUpdatedMsgHandle(roleMsg)
//			p := blockgen.pm.GetCurrentProcess()
//			p.preBlockHash = eth.blockchain.GetBlockByNumber(0).Hash()
//			err := p.processHeaderGen()
//			So(err, ShouldEqual, nil)
//		})
//	})
//}
//
func TestProcess_getParentBlock(t *testing.T) {
	log.InitLog(3)
	eth := fakeEthNew(0)
	if nil == eth {
		fmt.Println("failed to create eth")
	}

	blockgen, err := New(eth)
	if err != nil {
	}

	//process := newProcess(1, pm)
	Convey("获取父区块测试", t, func() {
		Convey("获取创世区块", func() {
			roleMsg := &mc.RoleUpdatedMsg{Role: common.RoleValidator, BlockNum: uint64(0), BlockHash: common.Hash{1}, Leader: common.HexToAddress(testAddress), IsSuperBlock: false}
			blockgen.roleUpdatedMsgHandle(roleMsg)
			p := blockgen.pm.GetCurrentProcess()
			block, err := p.getParentBlock()
			So(err, ShouldBeNil)
			So(block.Number().Uint64(), ShouldEqual, 0)
			So(block, ShouldEqual, p.blockChain().Genesis())
		})

		Convey("前一块blockhash为空", func() {
			roleMsg := &mc.RoleUpdatedMsg{Role: common.RoleValidator, BlockNum: uint64(10), BlockHash: common.Hash{1}, Leader: common.HexToAddress(testAddress), IsSuperBlock: false}
			blockgen.roleUpdatedMsgHandle(roleMsg)
			p := blockgen.pm.GetCurrentProcess()
			block, err := p.getParentBlock()
			So(err, ShouldBeError)
			So(block, ShouldBeNil)
		})

		Convey("前一块blockhash为非法值", func() {
			roleMsg := &mc.RoleUpdatedMsg{Role: common.RoleValidator, BlockNum: uint64(10), BlockHash: common.Hash{1}, Leader: common.HexToAddress(testAddress), IsSuperBlock: false}
			blockgen.roleUpdatedMsgHandle(roleMsg)
			p := blockgen.pm.GetCurrentProcess()
			p.preBlockHash = common.HexToHash("1232")
			block, err := p.getParentBlock()
			So(err, ShouldBeError)
			So(block, ShouldBeNil)
		})

		Convey("获取合法的blockhash", func() {
			roleMsg := &mc.RoleUpdatedMsg{Role: common.RoleValidator, BlockNum: uint64(10), BlockHash: common.Hash{1}, Leader: common.HexToAddress(testAddress), IsSuperBlock: false}
			blockgen.roleUpdatedMsgHandle(roleMsg)
			p := blockgen.pm.GetCurrentProcess()
			p.preBlockHash = eth.BlockChain().GetBlockByNumber(0).Hash()
			_, err := p.getParentBlock()
			So(err, ShouldBeNil)
		})
	})
}
