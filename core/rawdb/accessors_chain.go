// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package rawdb

import (
	"bytes"
	"encoding/binary"
	"math/big"

	"github.com/MatrixAINetwork/go-matrix/params/manparams"

	"github.com/MatrixAINetwork/go-matrix/core/matrixstate"
	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/params/manversion"

	"encoding/json"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/rlp"
)

// ReadCanonicalHash retrieves the hash assigned to a canonical block number.
func ReadCanonicalHash(db DatabaseReader, number uint64) common.Hash {
	data, _ := db.Get(append(append(headerPrefix, encodeBlockNumber(number)...), headerHashSuffix...))
	if len(data) == 0 {
		return common.Hash{}
	}
	return common.BytesToHash(data)
}

// WriteCanonicalHash stores the hash assigned to a canonical block number.
func WriteCanonicalHash(db DatabaseWriter, hash common.Hash, number uint64) {
	key := append(append(headerPrefix, encodeBlockNumber(number)...), headerHashSuffix...)
	if err := db.Put(key, hash.Bytes()); err != nil {
		log.Crit("Failed to store number to hash mapping", "err", err)
	}
}

// DeleteCanonicalHash removes the number to hash canonical mapping.
func DeleteCanonicalHash(db DatabaseDeleter, number uint64) {
	if err := db.Delete(append(append(headerPrefix, encodeBlockNumber(number)...), headerHashSuffix...)); err != nil {
		log.Crit("Failed to delete number to hash mapping", "err", err)
	}
}

// ReadHeaderNumber returns the header number assigned to a hash.
func ReadHeaderNumber(db DatabaseReader, hash common.Hash) *uint64 {
	data, _ := db.Get(append(headerNumberPrefix, hash.Bytes()...))
	if len(data) != 8 {
		return nil
	}
	number := binary.BigEndian.Uint64(data)
	return &number
}

// ReadHeadHeaderHash retrieves the hash of the current canonical head header.
func ReadHeadHeaderHash(db DatabaseReader) common.Hash {
	data, _ := db.Get(headHeaderKey)
	if len(data) == 0 {
		return common.Hash{}
	}
	return common.BytesToHash(data)
}

// WriteHeadHeaderHash stores the hash of the current canonical head header.
func WriteHeadHeaderHash(db DatabaseWriter, hash common.Hash) {
	if err := db.Put(headHeaderKey, hash.Bytes()); err != nil {
		log.Crit("Failed to store last header's hash", "err", err)
	}
}

// ReadHeadBlockHash retrieves the hash of the current canonical head block.
func ReadHeadBlockHash(db DatabaseReader) common.Hash {
	data, _ := db.Get(headBlockKey)
	if len(data) == 0 {
		return common.Hash{}
	}
	return common.BytesToHash(data)
}

// WriteHeadBlockHash stores the head block's hash.
func WriteHeadBlockHash(db DatabaseWriter, hash common.Hash) {
	if err := db.Put(headBlockKey, hash.Bytes()); err != nil {
		log.Crit("Failed to store last block's hash", "err", err)
	}
}

// ReadHeadFastBlockHash retrieves the hash of the current fast-sync head block.
func ReadHeadFastBlockHash(db DatabaseReader) common.Hash {
	data, _ := db.Get(headFastBlockKey)
	if len(data) == 0 {
		return common.Hash{}
	}
	return common.BytesToHash(data)
}

// WriteHeadFastBlockHash stores the hash of the current fast-sync head block.
func WriteHeadFastBlockHash(db DatabaseWriter, hash common.Hash) {
	if err := db.Put(headFastBlockKey, hash.Bytes()); err != nil {
		log.Crit("Failed to store last fast block's hash", "err", err)
	}
}

// ReadFastTrieProgress retrieves the number of tries nodes fast synced to allow
// reporting correct numbers across restarts.
func ReadFastTrieProgress(db DatabaseReader) uint64 {
	data, _ := db.Get(fastTrieProgressKey)
	if len(data) == 0 {
		return 0
	}
	return new(big.Int).SetBytes(data).Uint64()
}

// WriteFastTrieProgress stores the fast sync trie process counter to support
// retrieving it across restarts.
func WriteFastTrieProgress(db DatabaseWriter, count uint64) {
	if err := db.Put(fastTrieProgressKey, new(big.Int).SetUint64(count).Bytes()); err != nil {
		log.Crit("Failed to store fast sync trie progress", "err", err)
	}
}

// ReadHeaderRLP retrieves a block header in its raw RLP database encoding.
func ReadHeaderRLP(db DatabaseReader, hash common.Hash, number uint64) rlp.RawValue {
	data, _ := db.Get(append(append(headerPrefix, encodeBlockNumber(number)...), hash.Bytes()...))
	return data
}

// HasHeader verifies the existence of a block header corresponding to the hash.
func HasHeader(db DatabaseReader, hash common.Hash, number uint64) bool {
	key := append(append(append(headerPrefix, encodeBlockNumber(number)...), hash.Bytes()...))
	if has, err := db.Has(key); !has || err != nil {
		return false
	}
	return true
}

// ReadHeader retrieves the block header corresponding to the hash.
func ReadHeader(db DatabaseReader, hash common.Hash, number uint64) *types.Header {
	data := ReadHeaderRLP(db, hash, number)
	if len(data) == 0 {
		return nil
	}

	header := new(types.Header)
	//对区块数据解码
	if err := rlp.Decode(bytes.NewReader(data), header); err != nil {
		// 再次尝试使用旧header解析
		oldHeader := new(types.HeaderV1)
		if err := rlp.Decode(bytes.NewReader(data), oldHeader); err != nil {
			log.Error("Invalid block header RLP", "hash", hash, "err", err)
			return nil
		} else {
			header = oldHeader.TransferHeader()
		}
	}

	//创世区块，直接返回
	if number == 0 {
		return header
	}

	//备份区块头数据
	reader := new(types.Header)
	reader = types.CopyHeader(header)

	//根据number区块高度分别处理各段区块的写入，块高height1之前的块和height2（含）之后的块存储区块数据时需要删除多币种数据（重新构建主币种数据即可）
	if number < uint64(manparams.BlockHeaderModifyHeight) { //BlockheaderModifyHeight块高和VersionNumEpsilon块高作为height1和height2的临界点，需要替换成统一的宏配置管理
		//获取创世区块数据
		genesishash := ReadCanonicalHash(db, 0)
		genesisblock := ReadBlock(db, genesishash, 0)

		//根据多币种个数创建切片容量
		reader.Sharding = make([]common.Coinbyte, len(genesisblock.Sharding()))
		reader.Roots = make([]common.CoinRoot, len(genesisblock.Root()))

		//保存本区块number高度主币种数据
		reader.Sharding[0] = header.Sharding[0]
		reader.Roots[0] = header.Roots[0]

		//循环读取多币种数据
		for i := 1; i < len(genesisblock.Sharding()); i++ {
			reader.Sharding[i] = genesisblock.Sharding()[i]
			reader.Roots[i] = genesisblock.Root()[i]
		}
	} else if number >= uint64(manparams.BlockHeaderModifyHeight) && number < uint64(manversion.VersionNumAIMine) { //按原逻辑处理，直接返回读出来的区块数据信息
		return reader
	} else {
		//todo   状态树读取多币种数据的分支处理
		st, err := state.NewStateDBManage(reader.Roots, db, state.NewDatabase(db))
		if nil != err {
			log.Error("ReadHeader", "读取状态树施错误", err)
			return nil
		}
		readCurrencyHeader, err := matrixstate.GetCurrenyHeader(st)
		if nil != err {
			log.Error("ReadHeader", "读取多币种区块头错误", err)
			log.Error("ReadHeader", "readCurrencyHeader:", readCurrencyHeader)
			return nil
		}
		//根据多币种个数创建切片容量
		reader.Sharding = make([]common.Coinbyte, 1)
		reader.Roots = make([]common.CoinRoot, 1)

		//保存本区块number高度主币种数据
		reader.Sharding[0] = header.Sharding[0]
		reader.Roots[0] = header.Roots[0]

		reader.Sharding = append(reader.Sharding, readCurrencyHeader.Sharding...)
		reader.Roots = append(reader.Roots, readCurrencyHeader.Roots...)
	}

	return reader
}

// WriteHeader stores a block header into the database and also stores the hash-
// to-number mapping.
func WriteHeader(db DatabaseWriter, header *types.Header) {
	// Write the hash -> number mapping
	var (
		hash    = header.Hash().Bytes()
		number  = header.Number.Uint64()
		encoded = encodeBlockNumber(number) //对区块高度编码
	)

	//保存区块头数据
	storeHeader := types.CopyHeader(header)

	//根据number区块高度分别处理各段区块的写入，块高height1之前的块和height2（含）之后的块存储区块数据时需要删除多币种冗余数据（重新构建主币种数据即可）
	if (number < uint64(manparams.BlockHeaderModifyHeight) && number != 0) || number >= uint64(manversion.VersionNumAIMine) { //BlockheaderModifyHeight块高和VersionNumEpsilon块高作为height1和height2的临界点，需要替换成统一的宏配置管理
		//组装只有主币种Roots和Sharding的区块头数据（删除多余的区块头其他多币种）
		log.Debug("blockchain", "number:", number, "--删除多币种冗余数据，只写入主币种Roots和.Sharding")
		storeHeader.Roots = append([]common.CoinRoot{}, header.Roots[0])
		storeHeader.Sharding = append([]common.Coinbyte{}, header.Sharding[0])
	}

	//组建number 的 key值，带有 headerNumberPrefix前缀
	key := append(headerNumberPrefix, hash...)
	if err := db.Put(key, encoded); err != nil { //区块高度的key,value(经过编码）存入数据库
		log.Crit("Failed to store hash to number mapping", "err", err)
	}
	// Write the encoded header
	data, err := rlp.EncodeToBytes(storeHeader)
	if err != nil {
		log.Error("", "", err)
	}
	if err != nil {
		log.Crit("Failed to RLP encode header", "err", err)
	}
	key = append(append(headerPrefix, encoded...), hash...)
	if err := db.Put(key, data); err != nil { //区块数据的key,value(经过编码）存入数据库
		log.Crit("Failed to store header", "err", err)
	}
}

// DeleteHeader removes all block header data associated with a hash.
func DeleteHeader(db DatabaseDeleter, hash common.Hash, number uint64) {
	if err := db.Delete(append(append(headerPrefix, encodeBlockNumber(number)...), hash.Bytes()...)); err != nil {
		log.Crit("Failed to delete header", "err", err)
	}
	if err := db.Delete(append(headerNumberPrefix, hash.Bytes()...)); err != nil {
		log.Crit("Failed to delete hash to number mapping", "err", err)
	}
}

// ReadBodyRLP retrieves the block body (transactions and uncles) in RLP encoding.
func ReadBodyRLP(db DatabaseReader, hash common.Hash, number uint64) rlp.RawValue {
	data, _ := db.Get(append(append(blockBodyPrefix, encodeBlockNumber(number)...), hash.Bytes()...))
	return data
}

// WriteBodyRLP stores an RLP encoded block body into the database.
func WriteBodyRLP(db DatabaseWriter, hash common.Hash, number uint64, rlp rlp.RawValue) {
	key := append(append(blockBodyPrefix, encodeBlockNumber(number)...), hash.Bytes()...)
	if err := db.Put(key, rlp); err != nil {
		log.Crit("Failed to store block body", "err", err)
	}
}

// HasBody verifies the existence of a block body corresponding to the hash.
func HasBody(db DatabaseReader, hash common.Hash, number uint64) bool {
	key := append(append(blockBodyPrefix, encodeBlockNumber(number)...), hash.Bytes()...)
	if has, err := db.Has(key); !has || err != nil {
		return false
	}
	return true
}

// ReadBody retrieves the block body corresponding to the hash.
func ReadBody(db DatabaseReader, hash common.Hash, number uint64) *types.Body {
	data := ReadBodyRLP(db, hash, number)
	if len(data) == 0 {
		return nil
	}
	body := new(types.Body)
	if err := rlp.Decode(bytes.NewReader(data), body); err != nil {
		log.Error("Invalid block body RLP", "hash", hash, "err", err)
		return nil
	}
	return body
}

// WriteBody storea a block body into the database.
func WriteBody(db DatabaseWriter, hash common.Hash, number uint64, body *types.Body) {
	var tempBody types.Body
	tempBody.CurrencyBody = make([]types.CurrencyBlock, len(body.CurrencyBody))
	tempBody.Uncles = make([]*types.Header, len(body.Uncles))
	copy(tempBody.CurrencyBody, body.CurrencyBody)
	copy(tempBody.Uncles, body.Uncles)
	for i, _ := range tempBody.CurrencyBody {
		tempBody.CurrencyBody[i].Receipts = types.BodyReceipts{}
	}
	data, err := rlp.EncodeToBytes(tempBody)
	if err != nil {
		log.Crit("Failed to RLP encode body", "err", err)
	}
	WriteBodyRLP(db, hash, number, data)
}

// DeleteBody removes all block body data associated with a hash.
func DeleteBody(db DatabaseDeleter, hash common.Hash, number uint64) {
	if err := db.Delete(append(append(blockBodyPrefix, encodeBlockNumber(number)...), hash.Bytes()...)); err != nil {
		log.Crit("Failed to delete block body", "err", err)
	}
}

// ReadTd retrieves a block's total difficulty corresponding to the hash.
func ReadTd(db DatabaseReader, hash common.Hash, number uint64) *big.Int {
	data, _ := db.Get(append(append(append(headerPrefix, encodeBlockNumber(number)...), hash[:]...), headerTDSuffix...))
	if len(data) == 0 {
		return nil
	}
	td := new(big.Int)
	if err := rlp.Decode(bytes.NewReader(data), td); err != nil {
		log.Error("Invalid block total difficulty RLP", "hash", hash, "err", err)
		return nil
	}
	return td
}

// WriteTd stores the total difficulty of a block into the database.
func WriteTd(db DatabaseWriter, hash common.Hash, number uint64, td *big.Int) {
	data, err := rlp.EncodeToBytes(td)
	if err != nil {
		log.Crit("Failed to RLP encode block total difficulty", "err", err)
	}
	key := append(append(append(headerPrefix, encodeBlockNumber(number)...), hash.Bytes()...), headerTDSuffix...)
	if err := db.Put(key, data); err != nil {
		log.Crit("Failed to store block total difficulty", "err", err)
	}
}

// DeleteTd removes all block total difficulty data associated with a hash.
func DeleteTd(db DatabaseDeleter, hash common.Hash, number uint64) {
	if err := db.Delete(append(append(append(headerPrefix, encodeBlockNumber(number)...), hash.Bytes()...), headerTDSuffix...)); err != nil {
		log.Crit("Failed to delete block total difficulty", "err", err)
	}
}

// ReadReceipts retrieves all the transaction receipts belonging to a block.
func ReadReceipts(db DatabaseReader, hash common.Hash, number uint64) []types.CoinReceipts {
	// Retrieve the flattened receipt slice
	data, _ := db.Get(append(append(blockReceiptsPrefix, encodeBlockNumber(number)...), hash[:]...))
	if len(data) == 0 {
		return nil
	}
	// Convert the revceipts from their storage form to their internal representation

	var cr []types.CurrencyReceipts
	//storageReceipts := []*types.ReceiptForStorage{}
	if err := rlp.DecodeBytes(data, &cr); err != nil {
		log.Error("Invalid receipt array RLP", "hash", hash, "err", err)
		return nil
	}
	creceipts := make([]types.CoinReceipts, 0)
	for _, receipt := range cr {
		receipts := make(types.Receipts, 0)
		for _, r := range receipt.StorageReceipts {
			receipts = append(receipts, (*types.Receipt)(r))
		}
		creceipts = append(creceipts, types.CoinReceipts{CoinType: receipt.Currency, Receiptlist: receipts})
	}
	return creceipts
}

// WriteReceipts stores all the transaction receipts belonging to a block.
func WriteReceipts(db DatabaseWriter, hash common.Hash, number uint64, receipts []types.CoinReceipts) {
	// Convert the receipts into their storage form and serialize them
	currRcps := make([]types.CurrencyReceipts, 0)
	for _, cr := range receipts {
		storageReceipts := make([]*types.ReceiptForStorage, len(cr.Receiptlist))
		for i, receipt := range cr.Receiptlist {
			storageReceipts[i] = (*types.ReceiptForStorage)(receipt)
		}
		currRcps = append(currRcps, types.CurrencyReceipts{Currency: cr.CoinType, StorageReceipts: storageReceipts})
	}
	bytes, err := rlp.EncodeToBytes(currRcps)
	if err != nil {
		log.Crit("Failed to encode block receipts", "err", err)
	}
	// Store the flattened receipt slice
	key := append(append(blockReceiptsPrefix, encodeBlockNumber(number)...), hash.Bytes()...)
	if err := db.Put(key, bytes); err != nil {
		log.Crit("Failed to store block receipts", "err", err)
	}
}

// DeleteReceipts removes all receipt data associated with a block hash.
func DeleteReceipts(db DatabaseDeleter, hash common.Hash, number uint64) {
	if err := db.Delete(append(append(blockReceiptsPrefix, encodeBlockNumber(number)...), hash.Bytes()...)); err != nil {
		log.Crit("Failed to delete block receipts", "err", err)
	}
}

// ReadBlock retrieves an entire block corresponding to the hash, assembling it
// back from the stored header and body. If either the header or body could not
// be retrieved nil is returned.
//
// Note, due to concurrent download of header and block body the header and thus
// canonical hash can be stored in the database but the body data not (yet).
func ReadBlock(db DatabaseReader, hash common.Hash, number uint64) *types.Block {
	header := ReadHeader(db, hash, number)
	if header == nil {
		return nil
	}
	body := ReadBody(db, hash, number)
	if body == nil {
		return nil
	}
	return types.NewBlockWithHeader(header).WithBody(body.CurrencyBody, body.Uncles)
}

// WriteBlock serializes a block into the database, header and body separately.
func WriteBlock(db DatabaseWriter, block *types.Block) {
	WriteBody(db, block.Hash(), block.NumberU64(), block.Body())
	WriteHeader(db, block.Header())
}

// DeleteBlock removes all block data associated with a hash.
func DeleteBlock(db DatabaseDeleter, hash common.Hash, number uint64) {
	DeleteReceipts(db, hash, number)
	DeleteHeader(db, hash, number)
	DeleteBody(db, hash, number)
	DeleteTd(db, hash, number)
}

// FindCommonAncestor returns the last common ancestor of two block headers
func FindCommonAncestor(db DatabaseReader, a, b *types.Header) *types.Header {
	for bn := b.Number.Uint64(); a.Number.Uint64() > bn; {
		a = ReadHeader(db, a.ParentHash, a.Number.Uint64()-1)
		if a == nil {
			return nil
		}
	}
	for an := a.Number.Uint64(); an < b.Number.Uint64(); {
		b = ReadHeader(db, b.ParentHash, b.Number.Uint64()-1)
		if b == nil {
			return nil
		}
	}
	for a.Hash() != b.Hash() {
		a = ReadHeader(db, a.ParentHash, a.Number.Uint64()-1)
		if a == nil {
			return nil
		}
		b = ReadHeader(db, b.ParentHash, b.Number.Uint64()-1)
		if b == nil {
			return nil
		}
	}
	return a
}

func ReadSuperBlockIndex(db DatabaseReader) *SuperBlockIndexData {

	data, err := db.Get([]byte("SBLK"))
	if err != nil {
		//log.Error("ReadSuperBlockIndex ", "err", err)
		return nil
	}
	if len(data) == 0 {
		log.Error("ReadSuperBlockIndex ", "data len ", 0)
		return nil
	}

	sbi := new(SuperBlockIndexData)
	if err := json.Unmarshal(data, &sbi); err != nil {
		log.Error("Invalid SuperBlockIndexData data", "err", err)
		return nil
	}
	return sbi
}

func WriteSuperBlockIndex(db DatabaseWriter, sbi *SuperBlockIndexData) {

	bytes, err := json.Marshal(sbi)
	if err != nil {
		log.Crit("Failed to encode elect index", "err", err)
	}
	if err := db.Put([]byte("SBLK"), bytes); err != nil {
		log.Crit("Failed to store elect index", "err", err)
	}
}
