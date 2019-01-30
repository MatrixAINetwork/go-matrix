// Copyright (c) 2018-2019 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package electionseed

import (
	"github.com/matrix/go-matrix/log"
	"math/big"

	"github.com/matrix/go-matrix/baseinterface"
	"github.com/matrix/go-matrix/common"
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
	if err != nil{
		log.ERROR(ModuleElectSeed, "随机数计算错误:", "err", err)
	}
	return ans, err

}

func (self *ElectionSeed) Prepare(height uint64) error {
	err := mapElectSeedPlugs[self.plug].Prepare(height, self.support)
	if err != nil{
		log.ERROR(ModuleElectSeed, "随机数计算错误:", "err", err)
	}
	return err
}
