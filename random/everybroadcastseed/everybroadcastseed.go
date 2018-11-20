// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package everybroadcastseed

import (
	"math/big"

	"github.com/matrix/go-matrix/baseinterface"
	"github.com/matrix/go-matrix/log"
)

var (
	ModuleEveryBroadcastSeed   = "每个广播区块种子生成"
	mapEveryBroadcastSeedPlugs = make(map[string]EveryBroadcastSeedPlugs)
)

func init() {
	baseinterface.RegRandom("everybroadcastseed", NewSubService)
}

type EveryBroadcastSeedPlugs interface {
	CalcSeed(data uint64, support baseinterface.RandomChainSupport) (*big.Int, error)
	Prepare(uint64) error
}

func NewSubService(plug string, support baseinterface.RandomChainSupport) (baseinterface.RandomSubService, error) {
	everyBroadcastSeed := &EveryBroadcastSeed{
		plug:    plug,
		support: support,
	}
	return everyBroadcastSeed, nil
}

type EveryBroadcastSeed struct {
	plug    string
	support baseinterface.RandomChainSupport
}

func (self *EveryBroadcastSeed) SetValue(plug string, support baseinterface.RandomChainSupport) error {
	self.plug = plug
	self.support = support
	log.INFO(ModuleEveryBroadcastSeed, "每个广播区块种子 赋值阶段", "", "使用的插件名", plug)
	return nil
}

func RegisterEveryBlockSeedPlugs(name string, plug EveryBroadcastSeedPlugs) {
	log.INFO(ModuleEveryBroadcastSeed, "每个广播区块种子 注册阶段", "", "注册的插件名", plug)
	mapEveryBroadcastSeedPlugs[name] = plug
}

func (self *EveryBroadcastSeed) Prepare(height uint64) error {
	log.INFO(ModuleEveryBroadcastSeed, "每个广播区块种子 准备阶段", "", "收到的高度", height)
	err := mapEveryBroadcastSeedPlugs[self.plug].Prepare(height)
	return err
}

func (self *EveryBroadcastSeed) CalcData(calcData uint64) (*big.Int, error) {
	return mapEveryBroadcastSeedPlugs[self.plug].CalcSeed(calcData, self.support)
}
