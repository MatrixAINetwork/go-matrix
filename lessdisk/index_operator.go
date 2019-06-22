// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package lessdisk

import (
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/rlp"
	"github.com/pkg/errors"
)

type indexOperator struct {
	logInfo string
	db      DatabaseOperator
}

func newIndexOperator(logInfo string, db DatabaseOperator) *indexOperator {
	return &indexOperator{
		logInfo: logInfo,
		db:      db,
	}
}

func (im *indexOperator) UpdateMinNumberIndex(number uint64) error {
	minNumber := im.readMinNumberIndex()
	if minNumber != 0 && minNumber < number {
		log.Debug(im.logInfo, "更新最低高度索引", "无需更新", "number", number, "curMinNumber", minNumber)
		return nil
	}
	return im.writeMinNumberIndex(number)
}

func (im *indexOperator) readMinNumberIndex() uint64 {
	data, err := im.db.Get(minNumberIndex)
	if err != nil {
		log.Warn(im.logInfo, "获取本地最小区块高度索引失败", err)
		return 0
	}
	minNumber, err := decodeUint64(data)
	if err != nil {
		log.Error(im.logInfo, "本地最小区块高度索引解码失败", err)
		return 0
	}
	return minNumber
}

func (im *indexOperator) writeMinNumberIndex(minNumber uint64) error {
	if err := im.db.Put(minNumberIndex, encodeUint64(minNumber)); err != nil {
		return errors.Errorf("failed to write min number index: %v", err)
	}
	return nil
}

func (im *indexOperator) readBlkIndex(number uint64) []dbBlkIndex {
	data, _ := im.db.Get(append(blkIndexPrefix, encodeUint64(number)...))
	if len(data) == 0 {
		return []dbBlkIndex{}
	}

	index := make([]dbBlkIndex, 0)
	err := rlp.DecodeBytes(data, &index)
	if err != nil {
		log.Error(im.logInfo, "区块索引解码失败", err)
		return []dbBlkIndex{}
	}
	return index
}

func (im *indexOperator) writeBlkIndex(number uint64, index []dbBlkIndex) error {
	if len(index) == 0 {
		return errors.New("index is empty")
	}

	key := append(blkIndexPrefix, encodeUint64(number)...)
	data, err := rlp.EncodeToBytes(index)
	if err != nil {
		return errors.Errorf("failed to rlp encode block index: %v", err)
	}
	if err := im.db.Put(key, data); err != nil {
		return errors.Errorf("failed to write block index: %v", err)
	}
	return nil
}

func (im *indexOperator) deleteBlkIndex(number uint64) error {
	key := append(blkIndexPrefix, encodeUint64(number)...)
	if err := im.db.Delete(key); err != nil {
		return errors.Errorf("failed to delete block index: %v", err)
	}
	return nil
}
