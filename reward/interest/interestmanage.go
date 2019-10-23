// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package interest

import (
	"math/big"
	"os"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/matrixstate"
	"github.com/MatrixAINetwork/go-matrix/core/vm"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/reward/depositcfg"
	"github.com/MatrixAINetwork/go-matrix/reward/util"
)

type InterestOperator interface {
	PayInterest(state vm.StateDBManager, num uint64, time uint64) map[common.Address]*big.Int
	CalcReward(state vm.StateDBManager, num uint64, parentHash common.Hash)
}

const INTERESTDIR = "./interestdir"

func init() {
	_, e := os.Stat(INTERESTDIR)
	if e != nil {
		os.Mkdir(INTERESTDIR, os.ModePerm)
	}
}

func ManageNew(st util.StateDB, preSt util.StateDB) InterestOperator {
	calc, err := matrixstate.GetInterestCalc(preSt)
	if nil != err {
		log.Error(PackageName, "获取状态树配置错误")
		return nil
	}

	if calc == util.Stop {
		log.Error(PackageName, "停止发放区块奖励", "")
		return nil
	}

	switch calc {
	case util.CalcAlpha, util.CalcGamma:
		return New(st, preSt)
	case util.CalcDelta:
		return DeltaNew(st, preSt, depositcfg.VersionA)
	case util.CalcEpsilon:
		return EpsilonNew(st, preSt, depositcfg.VersionA)
	default:
		log.Error(PackageName, "获取利息计算引擎不存在", "")
		return nil

	}
	return nil
}
