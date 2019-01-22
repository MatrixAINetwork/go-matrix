// Copyright (c) 2018-2019 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package manparams

import (
	"github.com/matrix/go-matrix/common"
)

type stateReader interface {
	GetMatrixStateData(key string) (interface{}, error)
	GetMatrixStateDataByHash(key string, hash common.Hash) (interface{}, error)
	GetMatrixStateDataByNumber(key string, number uint64) (interface{}, error)
}

type StateDB interface {
	GetMatrixData(hash common.Hash) (val []byte)
	SetMatrixData(hash common.Hash, val []byte)
}

type matrixConfig struct {
	stReader stateReader
}

var mtxCfg = newMatrixConfig()

func newMatrixConfig() *matrixConfig {
	return &matrixConfig{
		stReader: nil,
	}
}

func SetStateReader(stReader stateReader) {
	mtxCfg.stReader = stReader
}

func (mcfg *matrixConfig) getStateData(key string) (interface{}, error) {
	return mcfg.stReader.GetMatrixStateData(key)
}

func (mcfg *matrixConfig) getStateDataByNumber(key string, number uint64) (interface{}, error) {
	return mcfg.stReader.GetMatrixStateDataByNumber(key, number)
}

func (mcfg *matrixConfig) getStateDataByHash(key string, hash common.Hash) (interface{}, error) {
	return mcfg.stReader.GetMatrixStateDataByHash(key, hash)
}
