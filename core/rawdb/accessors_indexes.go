// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package rawdb

import (
	"encoding/binary"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/rlp"
)

// ReadTxLookupEntry retrieves the positional metadata associated with a transaction
// hash to allow retrieving the transaction or receipt by hash.
func ReadTxLookupEntry(db DatabaseReader, hash common.Hash) (common.Hash, uint64, uint64, string) {
	data, _ := db.Get(append(txLookupPrefix, hash.Bytes()...))
	if len(data) == 0 {
		return common.Hash{}, 0, 0, ""
	}
	var entry TxLookupEntry
	if err := rlp.DecodeBytes(data, &entry); err != nil {
		log.Error("Invalid transaction lookup entry RLP", "hash", hash, "err", err)
		return common.Hash{}, 0, 0, ""
	}
	return entry.BlockHash, entry.BlockIndex, entry.Index, entry.Cointype
}

// WriteTxLookupEntries stores a positional metadata for every transaction from
// a block, enabling hash based transaction and receipt lookups.
func WriteTxLookupEntries(db DatabaseWriter, block *types.Block) {
	for _, currencies := range block.Currencies() {
		if len(currencies.Transactions.Sharding) > 0 {
			for _, tx := range currencies.Transactions.TransactionInfos {
				entry := TxLookupEntry{
					BlockHash:  block.Hash(),
					BlockIndex: block.NumberU64(),
					Index:      tx.Index,
					Cointype:   tx.Tx.GetTxCurrency(),
				}
				data, err := rlp.EncodeToBytes(entry)
				if err != nil {
					log.Crit("Failed to encode transaction lookup entry", "err", err)
				}
				if err := db.Put(append(txLookupPrefix, tx.Tx.Hash().Bytes()...), data); err != nil {
					log.Crit("Failed to store transaction lookup entry", "err", err)
				}
			}
		} else {
			var i uint64
			for _, tx := range currencies.Transactions.GetTransactions() {
				entry := TxLookupEntry{
					BlockHash:  block.Hash(),
					BlockIndex: block.NumberU64(),
					Index:      i,
					Cointype:   tx.GetTxCurrency(),
				}
				data, err := rlp.EncodeToBytes(entry)
				if err != nil {
					log.Crit("Failed to encode transaction lookup entry", "err", err)
				}
				if err := db.Put(append(txLookupPrefix, tx.Hash().Bytes()...), data); err != nil {
					log.Crit("Failed to store transaction lookup entry", "err", err)
				}
				i++
			}
		}
	}
}

// DeleteTxLookupEntry removes all transaction data associated with a hash.
func DeleteTxLookupEntry(db DatabaseDeleter, hash common.Hash) {
	db.Delete(append(txLookupPrefix, hash.Bytes()...))
}

// ReadTransaction retrieves a specific transaction from the database, along with
// its added positional metadata.
func ReadTransaction(db DatabaseReader, hash common.Hash) (types.SelfTransaction, common.Hash, uint64, uint64) {
	blockHash, blockNumber, txIndex, cointy := ReadTxLookupEntry(db, hash)
	if blockHash == (common.Hash{}) {
		return nil, common.Hash{}, 0, 0
	}

	body := ReadBody(db, blockHash, blockNumber)
	var tx types.SelfTransaction
	for _, currencyBlock := range body.CurrencyBody {
		if currencyBlock.CurrencyName == cointy {
			tx = currencyBlock.Transactions.GetTransactionByIndex(txIndex)
		}
	}
	if tx == nil {
		log.Error("Transaction referenced missing", "number", blockNumber, "hash", blockHash, "index", txIndex)
		return nil, common.Hash{}, 0, 0
	}
	//currencyBlock.Transactions.GetTransactions()[txIndex]
	return tx, blockHash, blockNumber, txIndex
}

// ReadReceipt retrieves a specific transaction receipt from the database, along with
// its added positional metadata.
func ReadReceipt(db DatabaseReader, hash common.Hash) (*types.Receipt, common.Hash, uint64, uint64) {
	blockHash, blockNumber, receiptIndex, cointy := ReadTxLookupEntry(db, hash)
	if blockHash == (common.Hash{}) {
		return nil, common.Hash{}, 0, 0
	}
	creceipts := ReadReceipts(db, blockHash, blockNumber)
	receipts := make(types.Receipts, 0)
	for _, rcps := range creceipts {
		if rcps.CoinType == cointy {
			receipts = append(receipts, rcps.Receiptlist...)
		}
	}
	if len(receipts) <= int(receiptIndex) {
		log.Error("Receipt refereced missing", "number", blockNumber, "hash", blockHash, "index", receiptIndex)
		return nil, common.Hash{}, 0, 0
	}
	return receipts[receiptIndex], blockHash, blockNumber, receiptIndex
}

// ReadBloomBits retrieves the compressed bloom bit vector belonging to the given
// section and bit index from the.
func ReadBloomBits(db DatabaseReader, bit uint, section uint64, head common.Hash) ([]byte, error) {
	key := append(append(bloomBitsPrefix, make([]byte, 10)...), head.Bytes()...)

	binary.BigEndian.PutUint16(key[1:], uint16(bit))
	binary.BigEndian.PutUint64(key[3:], section)

	return db.Get(key)
}

// WriteBloomBits stores the compressed bloom bits vector belonging to the given
// section and bit index.
func WriteBloomBits(db DatabaseWriter, bit uint, section uint64, head common.Hash, bits []byte) {
	key := append(append(bloomBitsPrefix, make([]byte, 10)...), head.Bytes()...)

	binary.BigEndian.PutUint16(key[1:], uint16(bit))
	binary.BigEndian.PutUint64(key[3:], section)

	if err := db.Put(key, bits); err != nil {
		log.Crit("Failed to store bloom bits", "err", err)
	}
}

///////////////////////////////////////////////////////////////////////
// 验证服务，验证过的区块缓存
func HasVerifiedBlockIndex(db DatabaseReader) bool {
	if has, err := db.Has(verifiedBlockIndex); !has || err != nil {
		return false
	}
	return true
}

func ReadVerifiedBlockIndex(db DatabaseReader) ([]byte, error) {
	return db.Get(verifiedBlockIndex)
}

func WriteVerifiedBlockIndex(db DatabaseWriter, data []byte) {
	if err := db.Put(verifiedBlockIndex, data); err != nil {
		log.Crit("Failed to store verified block index", "err", err)
	}
}

func ReadVerifiedBlock(db DatabaseReader, index uint8) ([]byte, error) {
	key := append(verifiedBlockPrefix, index)
	return db.Get(key)
}

func WriteVerifiedBlock(db DatabaseWriter, index uint8, blockData []byte) {
	key := append(verifiedBlockPrefix, index)

	if err := db.Put(key, blockData); err != nil {
		log.Crit("Failed to store verified block", "err", err)
	}
}
