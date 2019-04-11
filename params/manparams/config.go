// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package manparams

import (
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/mc"
)

type BCIntervalReader interface {
	GetBroadcastInterval() (*mc.BCIntervalInfo, error)
	GetBroadcastIntervalByHash(hash common.Hash) (*mc.BCIntervalInfo, error)
	GetBroadcastIntervalByNumber(number uint64) (*mc.BCIntervalInfo, error)
}

type StateDB interface {
	GetMatrixData(hash common.Hash) (val []byte)
	SetMatrixData(hash common.Hash, val []byte)
}

type broadcastConfig struct {
	reader BCIntervalReader
}

var broadcastCfg = newBroadcastCfg()

func newBroadcastCfg() *broadcastConfig {
	return &broadcastConfig{
		reader: nil,
	}
}

func SetStateReader(stReader BCIntervalReader) {
	broadcastCfg.reader = stReader
}
