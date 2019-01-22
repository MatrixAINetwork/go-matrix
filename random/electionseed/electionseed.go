// Copyright (c) 2018-2019 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package electionseed

import (
	"math/big"

	"github.com/matrix/go-matrix/baseinterface"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/params/manparams"
)

var (
	ModuleElectSeed   = "随机种子生成"
	mapElectSeedPlugs = make(map[string]ElectSeedPlugs)
)

type ElectSeedPlugs interface {
	CalcSeed(common.Hash, baseinterface.RandomChainSupport) (*big.Int, error)
	Prepare(uint64, baseinterface.RandomChainSupport) error
}

func init() {
	baseinterface.RegRandom(manparams.ElectionSeed, NewSubService)
}
func NewSubService(plug string, support baseinterface.RandomChainSupport) (baseinterface.RandomSubService, error) {
	electSeed := &ElectionSeed{
		plug:    plug,
		support: support,
	}
	return electSeed, nil
}

type ElectionSeed struct {
	plug    string
	support baseinterface.RandomChainSupport
}

func RegisterElectSeedPlugs(name string, plug ElectSeedPlugs) {
	mapElectSeedPlugs[name] = plug
}

func (self *ElectionSeed) CalcData(calcData common.Hash) (*big.Int, error) {
	ans, err := mapElectSeedPlugs[self.plug].CalcSeed(calcData, self.support)
	log.INFO(ModuleElectSeed, "计算阶段,收到的数据", calcData, "结果", ans, "err", err)
	return ans, err

}

func (self *ElectionSeed) Prepare(height uint64) error {
	err := mapElectSeedPlugs[self.plug].Prepare(height, self.support)
	log.INFO(ModuleElectSeed, "准备阶段,高度", height, "使用的插件", self.plug)
	return err
}
