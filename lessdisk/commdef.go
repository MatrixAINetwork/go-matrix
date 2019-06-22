// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package lessdisk

import (
	"encoding/binary"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/pkg/errors"
)

var (
	ErrDataSize = errors.New("data size err")
)

type ChainOperator interface {
	CurrentHeader() *types.Header
	DelLocalBlocks(blocks []*mc.BlockInfo) (fails []*mc.BlockInfo, err error)
}

type DatabaseOperator interface {
	Has(key []byte) (bool, error)
	Get(key []byte) ([]byte, error)
	Put(key []byte, value []byte) error
	Delete(key []byte) error
}

var (
	minNumberIndex = []byte("LessDisk-MinNumber")
	blkIndexPrefix = []byte("LessDisk-Index-")
)

type dbBlkIndex struct {
	Hash       common.Hash
	InsertTime uint64
}

func encodeUint64(num uint64) []byte {
	data := make([]byte, 8)
	binary.BigEndian.PutUint64(data, num)
	return data
}

func decodeUint64(data []byte) (uint64, error) {
	if len(data) < 8 {
		return 0, ErrDataSize
	}
	return binary.BigEndian.Uint64(data[:8]), nil
}
