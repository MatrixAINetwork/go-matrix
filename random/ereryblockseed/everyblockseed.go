// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package everyblockseed

import (
	"math/big"

	"github.com/MatrixAINetwork/go-matrix/baseinterface"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/params/manparams"
)

var (
	ModulePreBlockSeed   = "区块种子"
	mapPreBlockSeedPlugs = make(map[string]PreBlockSeedPlug)
)

type PreBlockSeedPlug interface {
	CalcSeed(req common.Hash, support baseinterface.RandomChainSupport) (*big.Int, error)
	Prepare(uint64, common.Hash) error
}

func init() {
	baseinterface.RegRandom(manparams.EveryBlockSeed, NewSubService)
}
func NewSubService(plug string, support baseinterface.RandomChainSupport) (baseinterface.RandomSubService, error) {
	everyBlockSeed := &preBlockSeed{
		plug:    plug,
		support: support,
	}
	return everyBlockSeed, nil
}

type preBlockSeed struct {
	plug    string
	support baseinterface.RandomChainSupport
}

func RegisterLotterySeedPlugs(name string, plug PreBlockSeedPlug) {
	mapPreBlockSeedPlugs[name] = plug
}

func (self *preBlockSeed) Prepare(height uint64, hash common.Hash) error {
	err := mapPreBlockSeedPlugs[self.plug].Prepare(height, hash)
	return err
}

func (self *preBlockSeed) CalcData(calcData common.Hash) (*big.Int, error) {
	return mapPreBlockSeedPlugs[self.plug].CalcSeed(calcData, self.support)
}
