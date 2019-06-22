// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package slash

import (
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/matrixstate"
	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/reward/util"
)

type SlashOperator interface {
	CalcSlash(currentState *state.StateDBManage, num uint64, upTimeMap map[common.Address]uint64, parentHash common.Hash, time uint64)
}

func ManageNew(chain util.ChainReader, st util.StateDB, preSt *state.StateDBManage) SlashOperator {
	data, err := matrixstate.GetSlashCalc(preSt)
	if nil != err {
		log.ERROR(PackageName, "获取状态树配置错误")
		return nil
	}

	if data == util.Stop {
		log.ERROR(PackageName, "停止发放区块奖励", "")
		return nil
	}
	switch data {
	case util.CalcAlpha, util.CalcGamma:
		return New(chain, st, preSt)
	case util.CalcDelta:
		return DeltaNew(chain, st, preSt)
	default:
		log.ERROR(PackageName, "获取惩罚计算引擎不存在")
		return nil

	}
	return nil
}
