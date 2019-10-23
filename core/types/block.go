// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

// Package types contains data types related to Matrix consensus.
package types

import (
	"encoding/binary"
	"io"
	"math/big"
	"sort"
	"sync/atomic"
	"time"

	"github.com/MatrixAINetwork/go-matrix/params"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/common/hexutil"
	"github.com/MatrixAINetwork/go-matrix/crypto/sha3"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/rlp"
)

var (
	EmptyRootHash  = common.HexToHash("0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421") //DeriveShaHash([]common.Hash{})
	EmptyUncleHash = CalcUncleHash(nil)
)

type SnapSaveInfo struct {
	Flg       int
	BlockNum  uint64
	BlockHash string
	SnapPath  string
}

// A BlockNonce is a 64-bit hash which proves (combined with the
// mix-hash) that a sufficient amount of computation has been carried
// out on a block.
type BlockNonce [8]byte

// EncodeNonce converts the given integer to a block nonce.
func EncodeNonce(i uint64) BlockNonce {
	var n BlockNonce
	binary.BigEndian.PutUint64(n[:], i)
	return n
}

// Uint64 returns the integer value of a block nonce.
func (n BlockNonce) Uint64() uint64 {
	return binary.BigEndian.Uint64(n[:])
}

// MarshalText encodes n as a hex string with 0x prefix.
func (n BlockNonce) MarshalText() ([]byte, error) {
	return hexutil.Bytes(n[:]).MarshalText()
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (n *BlockNonce) UnmarshalText(input []byte) error {
	return hexutil.UnmarshalFixedText("BlockNonce", input, n[:])
}

func rlpHash(x interface{}) (h common.Hash) {
	hw := sha3.NewKeccak256()
	rlp.Encode(hw, x)
	hw.Sum(h[:0])
	return h
}

func RlpHash(x interface{}) (h common.Hash) {
	hw := sha3.NewKeccak256()
	rlp.Encode(hw, x)
	hw.Sum(h[:0])
	return h
}
func RlpEncodeAndHash(x interface{}) (buff []byte, h common.Hash) {
	hw := sha3.NewKeccak256()
	buff, _ = rlp.EncodeToBytes(x)
	hw.Write(buff)
	hw.Sum(h[:0])
	return buff, h
}

//1. Complete Block transactions = []SelfTransactio
//2. Sharding Block transactions = []TxInfo
type BodyTransactions struct {
	Sharding         []uint            // if complete block sharding = []
	Transactions     []SelfTransaction //complete transactions
	TxHashs          []common.Hash     //所有交易的hash
	TransactionInfos []TransactionInfo //sharding transactions
}

func (bt *BodyTransactions) CheckHashs() bool {
	if len(bt.Sharding) == 0 {
		for i := 0; i < len(bt.Transactions); i++ {
			if bt.Transactions[i].Hash() != bt.TxHashs[i] {
				return false
			}
		}
		return true
	} else {
		for i := 0; i < len(bt.TransactionInfos); i++ {
			if bt.TransactionInfos[i].Tx.Hash() != bt.TxHashs[bt.TransactionInfos[i].Index] {
				return false
			}
		}
		return true
	}
}
func (bt *BodyTransactions) GetTransactionByIndex(index uint64) SelfTransaction {
	if len(bt.Sharding) == 0 {
		if index >= uint64(len(bt.Transactions)) {
			return nil
		}
		return bt.Transactions[index]
	} else {
		for _, txer := range bt.TransactionInfos {
			if txer.Index == index {
				return txer.Tx
			}
		}
		return nil
	}
}
func (bt *BodyTransactions) GetTransactions() []SelfTransaction {
	if len(bt.Sharding) == 0 {
		return bt.Transactions
	} else {
		txser := make([]SelfTransaction, 0)
		for _, txer := range bt.TransactionInfos {
			txser = append(txser, txer.Tx)
		}
		if len(txser) == 0 {
			return nil
		}
		return txser
	}
}

func (br *BodyReceipts) GetReceipts() Receipts {
	if len(br.Sharding) == 0 {
		return br.Rs
	} else {
		Receiptser := make([]*Receipt, 0)
		for _, Receipter := range br.ReceiptsInfos {
			Receiptser = append(Receiptser, &Receipter.Receipt)
		}
		if len(Receiptser) == 0 {
			return nil
		}
		return Receiptser
	}
}
func (bt *BodyReceipts) CheckRecptHashs() bool {
	if len(bt.Sharding) == 0 {
		for i := 0; i < len(bt.Rs); i++ {
			if bt.Rs[i].Hash() != bt.RsHashs[i] {
				return false
			}
		}
		return true
	} else {
		for i := 0; i < len(bt.ReceiptsInfos); i++ {
			if bt.ReceiptsInfos[i].Receipt.Hash() != bt.RsHashs[bt.ReceiptsInfos[i].Index] {
				return false
			}
		}
		return true
	}
}
func (bt *BodyReceipts) GetAReceiptsByIndex(index uint64) *Receipt {
	if len(bt.Sharding) == 0 {
		if index >= uint64(len(bt.Rs)) {
			return nil
		}
		return bt.Rs[index]
	} else {
		for _, re := range bt.ReceiptsInfos {
			if re.Index == index {
				return &re.Receipt
			}
		}
		return nil
	}
}
func SetTransactions(txser SelfTransactions, hashlist []common.Hash, shadings []uint) BodyTransactions {
	bt := BodyTransactions{}
	bt.TxHashs = make([]common.Hash, len(hashlist))
	copy(bt.TxHashs, hashlist)
	if len(shadings) == 0 {
		bt.Transactions = txser
	} else {
		bt.TransactionInfos = bt.setTransactionInfo(txser)
		bt.Sharding = shadings
	}
	return bt
}

func SetReceipts(Receiptser []*Receipt, hashlist []common.Hash, shadings []uint) BodyReceipts {
	br := BodyReceipts{}
	br.RsHashs = make([]common.Hash, len(hashlist))
	copy(br.RsHashs, hashlist)
	if len(shadings) == 0 {
		br.Rs = Receiptser
	} else {
		br.ReceiptsInfos = br.setSetReceiptInfo(Receiptser)
		br.Sharding = shadings
	}
	return br
}

func (br *BodyReceipts) setSetReceiptInfo(Receiptser Receipts) []ReceiptsInfo {
	for i, receipter := range Receiptser {
		if receipter == nil {
			continue
		}
		br.ReceiptsInfos = append(br.ReceiptsInfos, ReceiptsInfo{uint64(i), *receipter})
	}
	return br.ReceiptsInfos
}

func (bt *BodyTransactions) setTransactionInfo(txser SelfTransactions) []TransactionInfo {
	for i, txer := range txser {
		if txer == nil {
			continue
		}
		bt.TransactionInfos = append(bt.TransactionInfos, TransactionInfo{uint64(i), txer})
	}
	return bt.TransactionInfos
}

func (bt *BodyTransactions) EncodeRLP(w io.Writer) error {
	err := rlp.Encode(w, bt.Sharding)
	if err != nil {
		return err
	}
	if len(bt.Sharding) == 0 {
		return rlp.Encode(w, bt.Transactions)
	} else {
		return rlp.Encode(w, bt.TransactionInfos)
	}
}
func (bt *BodyTransactions) DecodeRLP(s *rlp.Stream) error {
	err := s.Decode(&bt.Sharding)
	if err != nil {
		return err
	}
	if len(bt.Sharding) == 0 {
		return s.Decode(&bt.Transactions)
	} else {
		return s.Decode(&bt.TransactionInfos)
	}
}

type TransactionInfo struct {
	Index uint64
	Tx    SelfTransaction
}

// Body is a simple (mutable, non-safe) data container for storing and moving
// a block's data contents (transactions and uncles) together.
type Body struct {
	CurrencyBody []CurrencyBlock
	Uncles       []*Header
}
type CurrencyHeader struct {
	Root        common.Hash
	TxHash      common.Hash
	ReceiptHash common.Hash
}

func MakeCurencyBlock(txser []CoinSelfTransaction, rece []CoinReceipts, shardings []uint) []CurrencyBlock {
	cb := make([]CurrencyBlock, 0, len(txser))
	for i, txer := range txser {
		var br BodyReceipts = BodyReceipts{}
		if len(rece) > 0 {
			br = SetReceipts(rece[i].Receiptlist, rece[i].Receiptlist.HashList(), shardings)
		}
		cb = append(cb, CurrencyBlock{CurrencyName: txer.CoinType, Transactions: SetTransactions(txer.Txser, TxHashList(txer.Txser), shardings), Receipts: br})
	}
	return cb
}

//币种block
//1 Validator :
//2 Miner : return len([]Sharding) == 0 discard tx
//Call Txs : input CurrencyBlock
type CurrencyBlock struct {
	CurrencyName string
	Header       CurrencyHeader
	Transactions BodyTransactions
	Receipts     BodyReceipts
	//Bloom 	 []byte
}

type BodyReceipts struct {
	Sharding      []uint   // if complete block sharding = []
	Rs            Receipts //complete transactions
	RsHashs       []common.Hash
	ReceiptsInfos []ReceiptsInfo //complete transactions
}

type ReceiptsInfo struct {
	Index   uint64
	Receipt Receipt
}

// if RPC
//Block: Block header uncles Currencies.Transactions
//Tx and TxReceipts : []Currencies.Transactions,[]Currencies.Receipts
//Save Raw Body []Currencies.Transactions
//Save Receipts []Currencies.Receipts
// Block represents an entire block in the Matrix blockchain.
type Block struct {
	header     *Header
	uncles     []*Header
	currencies []CurrencyBlock
	// caches
	hash atomic.Value
	size atomic.Value

	// Td is used by package core to store the total difficulty
	// of the chain up to and including the block.
	td *big.Int

	// These fields are used by package man to track
	// inter-peer block relay.
	ReceivedAt   time.Time
	ReceivedFrom interface{}
}

type BlockAllSt struct {
	Sblock *Block
	Pading uint64
}

// DeprecatedTd is an old relic for extracting the TD of a block. It is in the
// code solely to facilitate upgrading the database from the old format to the
// new, after which it should be deleted. Do not use!
func (b *Block) DeprecatedTd() *big.Int {
	return b.td
}

// [deprecated by man/63]
// StorageBlock defines the RLP encoding of a Block stored in the
// state database. The StorageBlock encoding contains fields that
// would otherwise need to be recomputed.
type StorageBlock Block

// "external" block encoding. used for man protocol, etc.
type extblock struct {
	Header     *Header
	Currencies []CurrencyBlock //BodyTransactions//[]SelfTransaction
	Uncles     []*Header
}

// [deprecated by man/63]
// "storage" block encoding. used for database.
type storageblock struct {
	Header     *Header
	Currencies []CurrencyBlock //[]SelfTransaction
	Uncles     []*Header
	TD         *big.Int
}

// NewBlock creates a new block. The input data is copied,
// changes to header and to the field values will not affect the
// block.
//
// The values of TxHash, UncleHash, ReceiptHash and Bloom in header
// are ignored and set to values derived from the given txs, uncles
// and receipts.
func NewBlock(header *Header, currencyBlocks []CurrencyBlock, uncles []*Header) *Block {
	b := &Block{header: CopyHeader(header), td: new(big.Int)}
	ischeck := len(b.header.Roots) > 0
	// TODO: panic if len(txs) != len(receipts)
	for i := 0; i < len(currencyBlocks); i++ {
		if len(currencyBlocks[i].Transactions.GetTransactions()) == 0 {
			if ischeck {
				for j, coinRoot := range b.header.Roots {
					if coinRoot.Cointyp == currencyBlocks[i].CurrencyName {
						b.header.Roots[j].TxHash = DeriveShaHash(currencyBlocks[i].Transactions.TxHashs)
						b.header.Roots[j].ReceiptHash = DeriveShaHash(currencyBlocks[i].Receipts.RsHashs)
						b.header.Roots[j].Bloom = CreateBloom(currencyBlocks[i].Receipts.GetReceipts())
						b.header.Roots[j].Cointyp = currencyBlocks[i].CurrencyName
						b.currencies = append(b.currencies, CurrencyBlock{CurrencyName: currencyBlocks[i].CurrencyName, Transactions: currencyBlocks[i].Transactions,
							Receipts: currencyBlocks[i].Receipts})
					}
				}
			} else {
				b.header.Roots = append(b.header.Roots, common.CoinRoot{Cointyp: currencyBlocks[i].CurrencyName, TxHash: EmptyRootHash, ReceiptHash: EmptyRootHash})
			}
		} else {
			if ischeck {
				for j, coinRoot := range b.header.Roots {
					if coinRoot.Cointyp == currencyBlocks[i].CurrencyName {
						b.header.Roots[j].TxHash = DeriveShaHash(currencyBlocks[i].Transactions.TxHashs)
						b.header.Roots[j].ReceiptHash = DeriveShaHash(currencyBlocks[i].Receipts.RsHashs)
						b.header.Roots[j].Bloom = CreateBloom(currencyBlocks[i].Receipts.GetReceipts())
						b.header.Roots[j].Cointyp = currencyBlocks[i].CurrencyName
						b.currencies = append(b.currencies, CurrencyBlock{CurrencyName: currencyBlocks[i].CurrencyName, Transactions: currencyBlocks[i].Transactions,
							Receipts: currencyBlocks[i].Receipts})
					} else {
						log.Info("coin name different", "header", coinRoot.Cointyp, "block", currencyBlocks[i].CurrencyName)
					}
				}
			} else {
				b.header.Roots = append(b.header.Roots, common.CoinRoot{Cointyp: currencyBlocks[i].CurrencyName, TxHash: DeriveShaHash(currencyBlocks[i].Transactions.TxHashs),
					ReceiptHash: DeriveShaHash(currencyBlocks[i].Receipts.RsHashs), Bloom: CreateBloom(currencyBlocks[i].Receipts.GetReceipts())})
				b.currencies = append(b.currencies, CurrencyBlock{CurrencyName: currencyBlocks[i].CurrencyName, Transactions: currencyBlocks[i].Transactions,
					Receipts: currencyBlocks[i].Receipts})
			}

		}
	}

	if len(uncles) == 0 {
		b.header.UncleHash = EmptyUncleHash
	} else {
		b.header.UncleHash = CalcUncleHash(uncles)
		b.uncles = make([]*Header, len(uncles))
		for i := range uncles {
			b.uncles[i] = CopyHeader(uncles[i])
		}
	}
	return b
}

func NewBlockMan(header *Header, currencyBlocks []CurrencyBlock, uncles []*Header) *Block {
	b := &Block{header: CopyHeader(header), td: new(big.Int)}
	ischeck := len(b.header.Roots) > 0
	// TODO: panic if len(txs) != len(receipts)
	for i := 0; i < len(currencyBlocks); i++ {
		if currencyBlocks[i].CurrencyName != params.MAN_COIN {
			continue
		}
		if len(currencyBlocks[i].Transactions.GetTransactions()) == 0 {
			if ischeck {
				for j, coinRoot := range b.header.Roots {
					if coinRoot.Cointyp == currencyBlocks[i].CurrencyName {
						b.header.Roots[j].TxHash = DeriveShaHash(currencyBlocks[i].Transactions.TxHashs)
						b.header.Roots[j].ReceiptHash = DeriveShaHash(currencyBlocks[i].Receipts.RsHashs)
						b.header.Roots[j].Bloom = CreateBloom(currencyBlocks[i].Receipts.GetReceipts())
						b.header.Roots[j].Cointyp = currencyBlocks[i].CurrencyName
						b.currencies = append(b.currencies, CurrencyBlock{CurrencyName: currencyBlocks[i].CurrencyName, Transactions: currencyBlocks[i].Transactions,
							Receipts: currencyBlocks[i].Receipts})
					}
				}
			} else {
				b.header.Roots = append(b.header.Roots, common.CoinRoot{Cointyp: currencyBlocks[i].CurrencyName, TxHash: EmptyRootHash, ReceiptHash: EmptyRootHash})
			}
		} else {
			if ischeck {
				for j, coinRoot := range b.header.Roots {
					if coinRoot.Cointyp == currencyBlocks[i].CurrencyName {
						b.header.Roots[j].TxHash = DeriveShaHash(currencyBlocks[i].Transactions.TxHashs)
						b.header.Roots[j].ReceiptHash = DeriveShaHash(currencyBlocks[i].Receipts.RsHashs)
						b.header.Roots[j].Bloom = CreateBloom(currencyBlocks[i].Receipts.GetReceipts())
						b.header.Roots[j].Cointyp = currencyBlocks[i].CurrencyName
						b.currencies = append(b.currencies, CurrencyBlock{CurrencyName: currencyBlocks[i].CurrencyName, Transactions: currencyBlocks[i].Transactions,
							Receipts: currencyBlocks[i].Receipts})
					} else {
						log.Info("coin name different", "header", coinRoot.Cointyp, "block", currencyBlocks[i].CurrencyName)
					}
				}
			} else {
				b.header.Roots = append(b.header.Roots, common.CoinRoot{Cointyp: currencyBlocks[i].CurrencyName, TxHash: DeriveShaHash(currencyBlocks[i].Transactions.TxHashs),
					ReceiptHash: DeriveShaHash(currencyBlocks[i].Receipts.RsHashs), Bloom: CreateBloom(currencyBlocks[i].Receipts.GetReceipts())})
				b.currencies = append(b.currencies, CurrencyBlock{CurrencyName: currencyBlocks[i].CurrencyName, Transactions: currencyBlocks[i].Transactions,
					Receipts: currencyBlocks[i].Receipts})
			}

		}
	}

	if len(uncles) == 0 {
		b.header.UncleHash = EmptyUncleHash
	} else {
		b.header.UncleHash = CalcUncleHash(uncles)
		b.uncles = make([]*Header, len(uncles))
		for i := range uncles {
			b.uncles[i] = CopyHeader(uncles[i])
		}
	}
	return b
}

// NewBlock creates a new block. The input data is copied,
// changes to header and to the field values will not affect the
// block.
//
// The values of TxHash, UncleHash, ReceiptHash and Bloom in header
// are ignored and set to values derived from the given txs, uncles
// and receipts.
func NewBlockCurrency(header *Header, currencyBlocks []CurrencyBlock, uncles []*Header) *Block {
	b := &Block{header: CopyHeader(header), td: new(big.Int)}
	ischeck := len(b.header.Roots) > 0
	// TODO: panic if len(txs) != len(receipts)
	for i := 0; i < len(currencyBlocks); i++ {
		if currencyBlocks[i].CurrencyName == params.MAN_COIN {
			continue
		}
		if len(currencyBlocks[i].Transactions.GetTransactions()) == 0 {
			if ischeck {
				for j, coinRoot := range b.header.Roots {
					if coinRoot.Cointyp == currencyBlocks[i].CurrencyName {
						b.header.Roots[j].TxHash = DeriveShaHash(currencyBlocks[i].Transactions.TxHashs)
						b.header.Roots[j].ReceiptHash = DeriveShaHash(currencyBlocks[i].Receipts.RsHashs)
						b.header.Roots[j].Bloom = CreateBloom(currencyBlocks[i].Receipts.GetReceipts())
						b.header.Roots[j].Cointyp = currencyBlocks[i].CurrencyName
						b.currencies = append(b.currencies, CurrencyBlock{CurrencyName: currencyBlocks[i].CurrencyName, Transactions: currencyBlocks[i].Transactions,
							Receipts: currencyBlocks[i].Receipts})
					}
				}
			} else {
				b.header.Roots = append(b.header.Roots, common.CoinRoot{Cointyp: currencyBlocks[i].CurrencyName, TxHash: EmptyRootHash, ReceiptHash: EmptyRootHash})
			}
		} else {
			if ischeck {
				for j, coinRoot := range b.header.Roots {
					if coinRoot.Cointyp == currencyBlocks[i].CurrencyName {
						b.header.Roots[j].TxHash = DeriveShaHash(currencyBlocks[i].Transactions.TxHashs)
						b.header.Roots[j].ReceiptHash = DeriveShaHash(currencyBlocks[i].Receipts.RsHashs)
						b.header.Roots[j].Bloom = CreateBloom(currencyBlocks[i].Receipts.GetReceipts())
						b.header.Roots[j].Cointyp = currencyBlocks[i].CurrencyName
						b.currencies = append(b.currencies, CurrencyBlock{CurrencyName: currencyBlocks[i].CurrencyName, Transactions: currencyBlocks[i].Transactions,
							Receipts: currencyBlocks[i].Receipts})
					} else {
						log.Info("coin name different", "header", coinRoot.Cointyp, "block", currencyBlocks[i].CurrencyName)
					}
				}
			} else {
				b.header.Roots = append(b.header.Roots, common.CoinRoot{Cointyp: currencyBlocks[i].CurrencyName, TxHash: DeriveShaHash(currencyBlocks[i].Transactions.TxHashs),
					ReceiptHash: DeriveShaHash(currencyBlocks[i].Receipts.RsHashs), Bloom: CreateBloom(currencyBlocks[i].Receipts.GetReceipts())})
				b.currencies = append(b.currencies, CurrencyBlock{CurrencyName: currencyBlocks[i].CurrencyName, Transactions: currencyBlocks[i].Transactions,
					Receipts: currencyBlocks[i].Receipts})
			}

		}
	}

	return b
}

// NewBlockWithHeader creates a block with the given header data. The
// header data is copied, changes to header and to the field values
// will not affect the block.
func NewBlockWithHeader(header *Header) *Block {
	return &Block{header: CopyHeader(header)}
}

// NewBlockWithHeader creates a block with the given header data. The
// header data is copied, changes to header and to the field values
// will not affect the block.
func NewBlockWithTxs(header *Header, currencyBlocks []CurrencyBlock) *Block {
	b := &Block{header: CopyHeader(header)}
	for _, Block := range currencyBlocks {
		if len(Block.Transactions.GetTransactions()) == 0 {
			for _, coinRoot := range b.header.Roots { //BB
				if coinRoot.Cointyp == Block.CurrencyName {
					coinRoot.TxHash = EmptyRootHash
				}
			}

		} else {
			for i, coinRoot := range b.header.Roots {
				if coinRoot.Cointyp == Block.CurrencyName {
					b.header.Roots[i].TxHash = DeriveShaHash(Block.Transactions.TxHashs)
					b.currencies = append(b.currencies, CurrencyBlock{CurrencyName: coinRoot.Cointyp, Transactions: Block.Transactions,
						Receipts: Block.Receipts})
				}
			}
		}
	}

	return b
}

// CopyHeader creates a deep copy of a block header to prevent side effects from
// modifying a header variable.
func CopyHeader(h *Header) *Header {
	cpy := *h
	if cpy.Time = new(big.Int); h.Time != nil {
		cpy.Time.Set(h.Time)
	}
	if cpy.Difficulty = new(big.Int); h.Difficulty != nil {
		cpy.Difficulty.Set(h.Difficulty)
	}
	if cpy.Number = new(big.Int); h.Number != nil {
		cpy.Number.Set(h.Number)
	}
	if len(h.Extra) > 0 {
		cpy.Extra = make([]byte, len(h.Extra))
		copy(cpy.Extra, h.Extra)
	}
	if len(h.Elect) > 0 {
		cpy.Elect = make([]common.Elect, len(h.Elect))
		copy(cpy.Elect, h.Elect)
	}
	if len(h.Roots) > 0 {
		cpy.Roots = make([]common.CoinRoot, len(h.Roots))
		copy(cpy.Roots, h.Roots)
	}
	if len(h.Sharding) > 0 {
		cpy.Sharding = make([]common.Coinbyte, len(h.Sharding))
		copy(cpy.Sharding, h.Sharding)
	}
	if len(h.NetTopology.NetTopologyData) > 0 {
		cpy.NetTopology.NetTopologyData = make([]common.NetTopologyData, len(h.NetTopology.NetTopologyData))
		copy(cpy.NetTopology.NetTopologyData, h.NetTopology.NetTopologyData)
		cpy.NetTopology.Type = h.NetTopology.Type
	}
	if len(h.Signatures) > 0 {
		cpy.Signatures = make([]common.Signature, len(h.Signatures))
		copy(cpy.Signatures, h.Signatures)
	}

	if len(h.Version) > 0 {
		cpy.Version = make([]byte, len(h.Version))
		copy(cpy.Version, h.Version)
	}
	if len(h.VersionSignatures) > 0 {
		cpy.VersionSignatures = make([]common.Signature, len(h.VersionSignatures))
		copy(cpy.VersionSignatures, h.VersionSignatures)
	}
	if len(h.VrfValue) > 0 {
		cpy.VrfValue = make([]byte, len(h.VrfValue))
		copy(cpy.VrfValue, h.VrfValue)
	}
	if len(h.BasePowers) > 0 {
		cpy.BasePowers = make([]BasePowers, len(h.BasePowers))
		copy(cpy.BasePowers, h.BasePowers)
	}
	return &cpy
}

// DecodeRLP decodes the Matrix
func (b *Block) DecodeRLP(s *rlp.Stream) error {
	var eb extblock
	_, size, _ := s.Kind()
	if err := s.Decode(&eb); err != nil {
		return err
	}
	b.header, b.uncles, b.currencies = eb.Header, eb.Uncles, eb.Currencies

	b.size.Store(common.StorageSize(rlp.ListSize(size)))
	return nil
}

// EncodeRLP serializes b into the Matrix RLP block format.
func (b *Block) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, extblock{
		Header:     b.header,
		Currencies: b.currencies,
		Uncles:     b.uncles,
	})
}

// [deprecated by man/63]
func (b *StorageBlock) DecodeRLP(s *rlp.Stream) error {
	var sb storageblock
	if err := s.Decode(&sb); err != nil {
		return err
	}
	b.header, b.uncles, b.currencies, b.td = sb.Header, sb.Uncles, sb.Currencies, sb.TD
	return nil
}

func (b *Block) SignAccounts() []common.VerifiedSign {
	return b.header.SignAccounts()
}

func (b *Block) IsSuperBlock() bool {
	return b.header.IsSuperHeader()
}

func (b *Block) IsPowBlock(broadcastInterval uint64) bool {
	return b.header.IsPowHeader(broadcastInterval)
}
func (b *Block) IsAIBlock(broadcastInterval uint64) bool {
	return b.header.IsAIHeader(broadcastInterval)
}

// TODO: copies

func (b *Block) Uncles() []*Header           { return b.uncles }
func (b *Block) Currencies() []CurrencyBlock { return b.currencies }
func (b *Block) SetCurrencies(currbl []CurrencyBlock) {
	b.currencies = make([]CurrencyBlock, len(currbl))
	copy(b.currencies, currbl)
}
func (b *Block) Transaction(hash common.Hash) SelfTransaction {
	for _, currencyblock := range b.currencies {
		txser := currencyblock.Transactions.GetTransactions()
		for _, transaction := range txser {
			if transaction.Hash() == hash {
				return transaction
			}
		}
	}

	return nil
}

func (b *Block) Number() *big.Int     { return new(big.Int).Set(b.header.Number) }
func (b *Block) GasLimit() uint64     { return b.header.GasLimit }
func (b *Block) GasUsed() uint64      { return b.header.GasUsed }
func (b *Block) Difficulty() *big.Int { return new(big.Int).Set(b.header.Difficulty) }
func (b *Block) Time() *big.Int       { return new(big.Int).Set(b.header.Time) }

func (b *Block) NumberU64() uint64      { return b.header.Number.Uint64() }
func (b *Block) MixDigest() common.Hash { return b.header.MixDigest }
func (b *Block) Nonce() uint64          { return binary.BigEndian.Uint64(b.header.Nonce[:]) }
func (b *Block) Bloom(cointype string) Bloom {
	for _, h := range b.header.Roots {
		if h.Cointyp == cointype {
			return h.Bloom
		}
	}
	return Bloom{}
}
func (b *Block) Coinbase() common.Address    { return b.header.Coinbase }
func (b *Block) Root() []common.CoinRoot     { return b.header.Roots }
func (b *Block) Sharding() []common.Coinbyte { return b.header.Sharding }
func (b *Block) ParentHash() common.Hash     { return b.header.ParentHash }

func (b *Block) UncleHash() common.Hash               { return b.header.UncleHash }
func (b *Block) Extra() []byte                        { return common.CopyBytes(b.header.Extra) }
func (b *Block) Version() []byte                      { return b.header.Version }
func (b *Block) VersionSignature() []common.Signature { return b.header.VersionSignatures }
func (b *Block) Signatures() []common.Signature       { return b.header.Signatures }

func (b *Block) Header() *Header { return CopyHeader(b.header) }

// Body returns the non-header content of the block.
func (b *Block) Body() *Body { return &Body{b.currencies, b.uncles} }

func (b *Block) HashNoNonce() common.Hash {
	return b.header.HashNoNonce()
}

func (b *Block) HashNoSigns() common.Hash {
	return b.header.HashNoSigns()
}

func (b *Block) HashNoSignsAndNonce() common.Hash {
	return b.header.HashNoSignsAndNonce()
}

// Size returns the true RLP encoded storage size of the block, either by encoding
// and returning it, or returning a previsouly cached value.
func (b *Block) Size() common.StorageSize {
	if size := b.size.Load(); size != nil {
		return size.(common.StorageSize)
	}
	c := writeCounter(0)
	rlp.Encode(&c, b)
	b.size.Store(common.StorageSize(c))
	return common.StorageSize(c)
}

type writeCounter common.StorageSize

func (c *writeCounter) Write(b []byte) (int, error) {
	*c += writeCounter(len(b))
	return len(b), nil
}

func CalcUncleHash(uncles []*Header) common.Hash {
	return rlpHash(uncles)
}

// WithSeal returns a new block with the data from b but the header replaced with
// the sealed one.
func (b *Block) WithSeal(header *Header) *Block {
	cpy := *header

	return &Block{
		header:     &cpy,
		currencies: b.currencies,
		uncles:     b.uncles,
	}
}

// WithBody returns a new block with the given transaction and uncle contents.
func (b *Block) WithBody(cb []CurrencyBlock, uncles []*Header) *Block {
	block := &Block{
		header:     CopyHeader(b.header),
		currencies: cb,
		uncles:     make([]*Header, len(uncles)),
	}
	//copy(block.transactions, transactions)
	for i := range uncles {
		block.uncles[i] = CopyHeader(uncles[i])
	}
	return block
}

// Hash returns the keccak256 hash of b's header.
// The hash is computed on the first call and cached thereafter.
func (b *Block) Hash() common.Hash {
	if hash := b.hash.Load(); hash != nil {
		return hash.(common.Hash)
	}
	v := b.header.Hash()
	b.hash.Store(v)
	return v
}

//will change header num!!
func (b *Block) SetHeadNum(num int64) {
	b.header.Number.SetInt64(num)
}

type Blocks []*Block

type BlockBy func(b1, b2 *Block) bool

func (self BlockBy) Sort(blocks Blocks) {
	bs := blockSorter{
		blocks: blocks,
		by:     self,
	}
	sort.Sort(bs)
}

type blockSorter struct {
	blocks Blocks
	by     func(b1, b2 *Block) bool
}

func (self blockSorter) Len() int { return len(self.blocks) }
func (self blockSorter) Swap(i, j int) {
	self.blocks[i], self.blocks[j] = self.blocks[j], self.blocks[i]
}
func (self blockSorter) Less(i, j int) bool { return self.by(self.blocks[i], self.blocks[j]) }

func Number(b1, b2 *Block) bool { return b1.header.Number.Cmp(b2.header.Number) < 0 }
