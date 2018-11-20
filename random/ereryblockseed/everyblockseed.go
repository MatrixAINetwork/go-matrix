// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package everyblockseed

import (
	"math/big"

	"github.com/matrix/go-matrix/baseinterface"
	"github.com/matrix/go-matrix/log"
)

var (
	ModuleEveryBlockSeed   = "每个区块种子生成"
	mapEveryBlockSeedPlugs = make(map[string]EveryBlockSeedPlugs)
)

type EveryBlockSeedPlugs interface {
	CalcSeed(req uint64, support baseinterface.RandomChainSupport) (*big.Int, error)
	Prepare(uint64) error
}

func init() {
	baseinterface.RegRandom("everyblockseed", NewSubService)
}
func NewSubService(plug string, support baseinterface.RandomChainSupport) (baseinterface.RandomSubService, error) {
	everyBlockSeed := &EveryBlockSeed{
		plug:    plug,
		support: support,
	}
	return everyBlockSeed, nil

}

type EveryBlockSeed struct {
	plug    string
	support baseinterface.RandomChainSupport
}

func RegisterLotterySeedPlugs(name string, plug EveryBlockSeedPlugs) {
	log.INFO(ModuleEveryBlockSeed, "每个区块种子服务注册阶段", "", "插件名称", name)
	mapEveryBlockSeedPlugs[name] = plug
}

func (self *EveryBlockSeed) Prepare(height uint64) error {
	err := mapEveryBlockSeedPlugs[self.plug].Prepare(height)
	return err
}

func (self *EveryBlockSeed) CalcData(calcData uint64) (*big.Int, error) {
	return mapEveryBlockSeedPlugs[self.plug].CalcSeed(calcData, self.support)
}
