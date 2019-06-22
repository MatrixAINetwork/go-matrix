// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package blkverify

import (
	"encoding/binary"
	"encoding/json"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/rawdb"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mandb"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/pkg/errors"
)

const verifiedBlockCacheSize = uint8(3)

type VerifiedBlockIndex struct {
	Last      uint8
	Capacity  uint8
	IndexRing []common.Hash
}

func newVerifiedBlockIndex(capacity uint8) *VerifiedBlockIndex {
	index := &VerifiedBlockIndex{
		Last:      capacity - 1,
		Capacity:  capacity,
		IndexRing: make([]common.Hash, capacity),
	}
	return index
}

func (vbi *VerifiedBlockIndex) saveBlock(hash common.Hash) uint8 {
	vbi.Last = (vbi.Last + 1) % vbi.Capacity
	vbi.IndexRing[vbi.Last] = hash
	return vbi.Last
}

func (vbi *VerifiedBlockIndex) findBlockPos(hash common.Hash) (uint8, error) {
	if hash == (common.Hash{}) {
		return 0, errors.New("hash is empty")
	}
	for pos, val := range vbi.IndexRing {
		if val == hash {
			return uint8(pos), nil
		}
	}
	return 0, errors.New("hash not find in index")
}

type verifiedBlock struct {
	hash common.Hash
	req  *mc.HD_BlkConsensusReqMsg
	txs  types.SelfTransactions
}

func saveVerifiedBlockToDB(db mandb.Database, hash common.Hash, req *mc.HD_BlkConsensusReqMsg, txs []types.CoinSelfTransaction) error {

	data, err := encodeVerifiedBlock(req, txs)
	if err != nil {
		return err
	}

	index, err := getVerifiedBlockIndex(db)
	if err != nil {
		return err
	}

	if index.Capacity == 0 {
		return errors.New("索引信息中capacity为0")
	}
	pos := index.saveBlock(hash)

	indexData, err := encodeVerifiedBlockIndex(index)
	if err != nil {
		return err
	}

	rawdb.WriteVerifiedBlockIndex(db, indexData)
	rawdb.WriteVerifiedBlock(db, pos, data)
	return nil
}

func readVerifiedBlocksFromDB(db mandb.Database) (blocks []verifiedBlock, err error) {
	// 读取DB中的索引
	index, err := getVerifiedBlockIndex(db)
	if err != nil {
		return nil, err
	}

	// 按索引信息读取req和txs
	for pos, hash := range index.IndexRing {
		if hash == (common.Hash{}) {
			continue
		}
		data, err := rawdb.ReadVerifiedBlock(db, uint8(pos))
		if err != nil {
			log.Info("verified block recovery", "read block data from db err", err, "pos", pos, "hash", hash.TerminalString())
			continue
		}
		req, txs, err := decodeVerifiedBlock(data)
		if err != nil {
			log.Info("verified block recovery", "block data decode err", err, "pos", pos, "hash", hash.TerminalString())
			continue
		}
		if req.Header.HashNoSignsAndNonce() != hash {
			log.Info("verified block recovery", "req data illegal", "hash not match", "pos", pos, "data hash", req.Header.HashNoSignsAndNonce().TerminalString(), "index hash", hash.TerminalString())
			continue
		}
		if req.TxsCodeCount() != len(txs) {
			log.Info("verified block recovery", "req data illegal", "txs size not match", "pos", pos, "txsCode size", req.TxsCodeCount(), "txs size", len(txs), "hash", hash.TerminalString())
			continue
		}

		blocks = append(blocks, verifiedBlock{hash: hash, req: req, txs: txs})
	}

	return blocks, nil
}

func getVerifiedBlockIndex(db rawdb.DatabaseReader) (*VerifiedBlockIndex, error) {
	if false == rawdb.HasVerifiedBlockIndex(db) {
		index := newVerifiedBlockIndex(verifiedBlockCacheSize)
		return index, nil
	}
	data, err := rawdb.ReadVerifiedBlockIndex(db)
	if err != nil {
		return nil, err
	}
	return decodeVerifiedBlockIndex(data)
}

func encodeVerifiedBlockIndex(index *VerifiedBlockIndex) ([]byte, error) {
	if index == nil {
		return nil, errors.New("param `index` is inl")
	}

	data, err := json.Marshal(index)
	if err != nil {
		return nil, errors.Errorf("index json.Marshal failed: %s", err)
	}
	return data, nil
}

func decodeVerifiedBlockIndex(data []byte) (*VerifiedBlockIndex, error) {
	index := new(VerifiedBlockIndex)
	err := json.Unmarshal(data, index)
	if err != nil {
		return nil, errors.Errorf("index json.Unmarshal failed: %s", err)
	}
	return index, nil
}

func encodeVerifiedBlock(req *mc.HD_BlkConsensusReqMsg, txs []types.CoinSelfTransaction) ([]byte, error) {
	if req == nil {
		return nil, errors.New("req msg is nil")
	}
	txss := types.GetTX(txs)
	txSize := req.TxsCodeCount()
	if txSize != len(txss) {
		return nil, errors.New("txs count is not match txCodes count")
	}

	reqData, err := json.Marshal(req)
	if err != nil {
		return nil, errors.Errorf("req msg json.Marshal failed: %s", err)
	}

	marshalTxs := make([]*types.Transaction_Mx, txSize)
	for i := 0; i < txSize; i++ {
		tx := txss[i]
		if mtx := types.SetTransactionToMx(tx); mtx == nil {
			return nil, errors.Errorf("tx(%d/%d) transfer to mtx err,tx=%v", i+1, txSize, tx)
		} else {
			marshalTxs[i] = mtx
		}
	}
	txsData, err := json.Marshal(marshalTxs)
	if err != nil {
		return nil, errors.Errorf("txs json.Marshal failed: %s", err)
	}

	reqDataSize := uint64(len(reqData))
	txsDataSize := uint64(len(txsData))

	data := append(encodeUint64(reqDataSize), encodeUint64(txsDataSize)...)
	data = append(data, reqData...)
	data = append(data, txsData...)
	return data, nil
}

func decodeVerifiedBlock(data []byte) (*mc.HD_BlkConsensusReqMsg, types.SelfTransactions, error) {
	totalSize := uint64(len(data))
	if totalSize < 16 {
		return nil, nil, errors.New("data size err, < 16")
	}

	reqDataSize := binary.BigEndian.Uint64(data[:8])
	txsDataSize := binary.BigEndian.Uint64(data[8:16])

	if totalSize != 16+reqDataSize+txsDataSize {
		return nil, nil, errors.Errorf("data size err, total size(%d) != 16 + reqDataSize(%d) + txsDataSize(%d)", totalSize, reqDataSize, txsDataSize)
	}
	reqData := data[16 : 16+reqDataSize]
	txsData := data[16+reqDataSize:]

	req := new(mc.HD_BlkConsensusReqMsg)
	err := json.Unmarshal(reqData, req)
	if err != nil {
		return nil, nil, errors.Errorf("req msg json.Unmarshal failed: %s", err)
	}

	marshalTxs := make([]*types.Transaction_Mx, 0)
	err = json.Unmarshal(txsData, &marshalTxs)
	if err != nil {
		return nil, nil, errors.Errorf("txs json.Unmarshal failed: %s", err)
	}

	txs := make([]types.SelfTransaction, 0)
	txSize := len(marshalTxs)
	for i := 0; i < txSize; i++ {
		tx := types.SetMxToTransaction(marshalTxs[i])
		if nil == tx {
			return nil, nil, errors.Errorf("decode tx err: the (%d/%d) tx is nil", i, txSize)
		}
		txs = append(txs, tx)
	}
	return req, txs, nil
}

func encodeUint64(number uint64) []byte {
	enc := make([]byte, 8)
	binary.BigEndian.PutUint64(enc, number)
	return enc
}
