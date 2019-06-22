// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package miner

import (
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"math/big"
	"testing"
)

func TestUinit1(t *testing.T) {
	//初始化是否正常
	tempHeader := &types.Header{
		Number: big.NewInt(100),
	}
	reqCtrl := newMineReqData(common.Hash{}, tempHeader, types.SelfTransactions{}, false)

	if tempHeader != reqCtrl.header {
		panic("header不同")
	}
	if reqCtrl.headerHash.Equal(common.Hash{}) == false {
		panic("高度不同")
	}
	if false != reqCtrl.isBroadcastReq {
		panic("是否是广播节点错误")
	}
	if false != reqCtrl.mined {
		panic("miner 错误")
	}
	if nil != reqCtrl.mineDiff {
		panic("minerDiff 错误")
	}
	if 0 != reqCtrl.mineResultSendTime {
		panic("mineResultSendTime 错误")
	}
}

func TestUinit2(t *testing.T) {
	//检查ResendMineResult函数 正在挖矿 且入参合法
	tempHeader := &types.Header{
		Number: big.NewInt(100),
	}
	reqCtrl := newMineReqData(common.Hash{}, tempHeader, types.SelfTransactions{}, false)
	reqCtrl.mined = true
	reqCtrl.ResendMineResult(5)
	if reqCtrl.mineResultSendTime != 5 {
		panic("ResendMinerReSult失败")
	}
}
func TestUinit3(t *testing.T) {
	//检查ResendMineResult函数 正在挖矿 且入参小于最小间隔
	tempHeader := &types.Header{
		Number: big.NewInt(100),
	}
	reqCtrl := newMineReqData(common.Hash{}, tempHeader, types.SelfTransactions{}, false)
	reqCtrl.mined = true
	reqCtrl.ResendMineResult(1)
	if reqCtrl.mineResultSendTime != 0 {
		panic("ResendMinerReSult失败")
	}
}
func TestUinit4(t *testing.T) {
	//检查ResendMineResult函数 不在挖矿
	tempHeader := &types.Header{
		Number: big.NewInt(100),
	}
	reqCtrl := newMineReqData(common.Hash{}, tempHeader, types.SelfTransactions{}, false)
	//reqCtrl.mined=true
	reqCtrl.ResendMineResult(1)
	if reqCtrl.mineResultSendTime != 0 {
		panic("ResendMinerReSult失败")
	}
}
func TestUnit5(t *testing.T) {
	//检查交易
	tempHeader := &types.Header{
		Number: big.NewInt(100),
	}
	reqCtrl := newMineReqData(common.Hash{}, tempHeader, types.SelfTransactions{}, false)
	if len(reqCtrl.txs) != 0 {
		panic("交易赋值失败")
	}

}

type FakeValidatorReader struct {
}

func (self *FakeValidatorReader) GetCurrentHash() common.Hash {
	return common.Hash{}
}
func (self *FakeValidatorReader) GetValidatorByHash(hash common.Hash) (*mc.TopologyGraph, error) {
	return &mc.TopologyGraph{}, nil
}

func TestUnit6(t *testing.T) {
	//newMinerReqCrtl 初始化检查
	bc := &FakeValidatorReader{}
	tempCrtl := newMinReqCtrl(nil, bc)
	if tempCrtl.role != common.RoleNil {
		panic("身份不对")
	}
	if tempCrtl.curNumber != 0 {
		panic("当前高度不正确")
	}
	if tempCrtl.currentMineReq != nil {
		panic("当前挖矿请求未空")
	}
}

func TestUnit7(t *testing.T) {
	bc := &FakeValidatorReader{}
	tempCrtl := newMinReqCtrl(nil, bc)
	status := tempCrtl.CanMining()
	if status == true {
		panic("是否该挖矿状态不对")
	}
}
func TestUnit8(t *testing.T) {
	bc := &FakeValidatorReader{}
	tempCrtl := newMinReqCtrl(nil, bc)
	ans := tempCrtl.GetCurrentMineReq()
	if ans != nil {
		panic("获取当前挖矿请求错误")
	}
}
func TestUnit9(t *testing.T) {
	bc := &FakeValidatorReader{}
	tempCrtl := newMinReqCtrl(nil, bc)
	ans := tempCrtl.roleCanMine(common.RoleMiner, 1)
	if ans == false {
		panic("可以挖矿算法失败")
	}
}
func TestUnit10(t *testing.T) {
	bc := &FakeValidatorReader{}
	tempCrtl := newMinReqCtrl(nil, bc)
	ans := tempCrtl.roleCanMine(common.RoleBroadcast, 1)
	if ans == true {
		panic("可以挖矿算法失败")
	}
}
