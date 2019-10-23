// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package types

import (
	"encoding/binary"
	"io"
	"math/big"
	"unsafe"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/common/hexutil"
	"github.com/MatrixAINetwork/go-matrix/crypto"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/params"
	"github.com/MatrixAINetwork/go-matrix/params/manversion"
	"github.com/MatrixAINetwork/go-matrix/rlp"
)

//go:generate gencodec -type Header -field-override headerMarshaling -out gen_header_json.go

// Header represents a block header in the Matrix blockchain.
type Header struct {
	ParentHash        common.Hash        `json:"parentHash"        gencodec:"required"`
	UncleHash         common.Hash        `json:"sha3Uncles"        gencodec:"required"`
	Leader            common.Address     `json:"leader"            gencodec:"required"`
	Coinbase          common.Address     `json:"miner"             gencodec:"required"`
	Roots             []common.CoinRoot  `json:"stateRoot"         gencodec:"required"`
	Sharding          []common.Coinbyte  `json:"sharding"          gencodec:"required"`
	Difficulty        *big.Int           `json:"difficulty"        gencodec:"required"`
	Number            *big.Int           `json:"number"            gencodec:"required"`
	GasLimit          uint64             `json:"gasLimit"          gencodec:"required"`
	GasUsed           uint64             `json:"gasUsed"           gencodec:"required"`
	Time              *big.Int           `json:"timestamp"         gencodec:"required"`
	Elect             []common.Elect     `json:"elect"             gencodec:"required"`
	NetTopology       common.NetTopology `json:"nettopology"       gencodec:"required"`
	Signatures        []common.Signature `json:"signatures"        gencodec:"required"`
	AIHash            common.Hash        `json:"aiHash" 		   gencodec:"required"`
	AICoinbase        common.Address     `json:"aiMiner"           gencodec:"required"`
	Extra             []byte             `json:"extraData"         gencodec:"required"`
	MixDigest         common.Hash        `json:"mixHash"           gencodec:"required"`
	Nonce             BlockNonce         `json:"nonce"             gencodec:"required"`
	Version           []byte             `json:"version"           gencodec:"required"`
	VersionSignatures []common.Signature `json:"versionSignatures" gencodec:"required"`
	VrfValue          []byte             `json:"vrfvalue"          gencodec:"required"`
	Sm3Nonce          BlockNonce         `json:"sm3Nonce"          gencodec:"required"`
	BasePowers        []BasePowers       `json:"basePowers"        gencodec:"required"`
}

type BasePowers struct {
	Miner     common.Address `json:"miner"             gencodec:"required"`
	Nonce     BlockNonce     `json:"nonce"             gencodec:"required"`
	MixDigest common.Hash    `json:"mixHash"           gencodec:"required"`
}

// field type overrides for gencodec
type headerMarshaling struct {
	Difficulty *hexutil.Big
	Number     *hexutil.Big
	GasLimit   hexutil.Uint64
	GasUsed    hexutil.Uint64
	Time       *hexutil.Big
	Extra      hexutil.Bytes
	Hash       common.Hash `json:"hash"` // adds call to Hash() in MarshalJSON
}

// Hash returns the block hash of the header, which is simply the keccak256 hash of its
// RLP encoding.
func (h *Header) Hash() common.Hash {
	return rlpHash(h)
}

// HashNoNonce returns the hash which is used as input for the proof-of-work search.
func (h *Header) HashNoNonce() common.Hash {
	if manversion.VersionCmp(string(h.Version), manversion.VersionAIMine) >= 0 {
		return rlpHash([]interface{}{
			h.ParentHash,
			h.UncleHash,
			h.Leader,
			h.Roots,
			h.Sharding,
			h.Difficulty,
			h.Number,
			h.GasLimit,
			h.GasUsed,
			h.Time,
			h.Elect,
			h.NetTopology,
			h.Signatures,
			h.Extra,
			h.Version,
			h.VersionSignatures,
			h.VrfValue,
		})
	} else {
		return rlpHash([]interface{}{
			h.ParentHash,
			h.UncleHash,
			h.Leader,
			h.Roots,
			h.Sharding,
			h.Difficulty,
			h.Number,
			h.GasLimit,
			h.GasUsed,
			h.Time,
			h.Elect,
			h.NetTopology,
			h.Signatures,
			h.Extra,
			h.Version,
			h.VersionSignatures,
		})
	}
}

func (h *Header) HashNoSigns() common.Hash {
	if manversion.VersionCmp(string(h.Version), manversion.VersionAIMine) >= 0 {
		return rlpHash([]interface{}{
			h.ParentHash,
			h.UncleHash,
			h.Leader,
			h.Coinbase,
			h.Roots,
			h.Sharding,
			h.Difficulty,
			h.Number,
			h.GasLimit,
			h.GasUsed,
			h.Time,
			h.Elect,
			h.NetTopology,
			h.Extra,
			h.MixDigest,
			h.Nonce,
			h.Version,
			h.VersionSignatures,
			h.VrfValue,
			h.AIHash,
			h.AICoinbase,
			h.Sm3Nonce,
			h.BasePowers,
		})
	} else {
		return rlpHash([]interface{}{
			h.ParentHash,
			h.UncleHash,
			h.Leader,
			h.Coinbase,
			h.Roots,
			h.Sharding,
			h.Difficulty,
			h.Number,
			h.GasLimit,
			h.GasUsed,
			h.Time,
			h.Elect,
			h.NetTopology,
			h.Extra,
			h.MixDigest,
			h.Nonce,
			h.Version,
			h.VersionSignatures,
		})
	}
}

func (h *Header) HashNoSignsAndNonce() common.Hash {
	if manversion.VersionCmp(string(h.Version), manversion.VersionAIMine) >= 0 {
		return rlpHash([]interface{}{
			h.ParentHash,
			h.UncleHash,
			h.Leader,
			h.Roots,
			h.Sharding,
			h.Difficulty,
			h.Number,
			h.GasLimit,
			h.GasUsed,
			h.Time,
			h.Elect,
			h.NetTopology,
			h.Extra,
			h.Version,
			h.VersionSignatures,
			h.VrfValue,
			h.AIHash,
			h.AICoinbase,
		})
	} else {
		return rlpHash([]interface{}{
			h.ParentHash,
			h.UncleHash,
			h.Leader,
			h.Roots,
			h.Sharding,
			h.Difficulty,
			h.Number,
			h.GasLimit,
			h.GasUsed,
			h.Time,
			h.Elect,
			h.NetTopology,
			h.Extra,
			h.Version,
			h.VersionSignatures,
		})
	}
}

// Size returns the approximate memory used by all internal contents. It is used
// to approximate and limit the memory consumption of various caches.
func (h *Header) Size() common.StorageSize {
	return common.StorageSize(unsafe.Sizeof(*h)) + common.StorageSize(len(h.Extra)+(h.Difficulty.BitLen()+h.Number.BitLen()+h.Time.BitLen())/8)
}

func (h *Header) SignAccounts() []common.VerifiedSign {
	accounts := make([]common.VerifiedSign, 0)
	hash := h.HashNoSignsAndNonce().Bytes()
	for i := 0; i < len(h.Signatures); i++ {
		sign := h.Signatures[i]
		signAccount, validate, err := crypto.VerifySignWithValidate(hash, sign.Bytes())
		if err != nil {
			log.Warn("header SignAccounts", "VerifySignWithValidate err", err, "sign", sign)
			continue
		}

		if !validate {
			log.Warn("header SignAccounts", "VerifySignWithValidate illegal", validate, "sign", sign)
			continue
		}
		accounts = append(accounts, common.VerifiedSign{
			Sign:     sign,
			Account:  signAccount,
			Validate: validate,
			Stock:    0,
		})
	}
	return accounts
}

func (h *Header) IsSuperHeader() bool {
	return h.Leader == common.HexToAddress("0x8111111111111111111111111111111111111111")
}

func (h *Header) IsPowHeader(broadcastInterval uint64) bool {
	return params.IsPowBlock(h.Number.Uint64(), broadcastInterval)
}

func (h *Header) IsAIHeader(broadcastInterval uint64) bool {
	return params.IsAIBlock(h.Number.Uint64(), broadcastInterval)
}

func (h *Header) SuperBlockSeq() uint64 {
	if h.Number.Uint64() == 0 {
		return 0
	}
	if len(h.Extra) < 8 {
		return 0
	}

	return uint64(binary.BigEndian.Uint64(h.Extra[:8]))
}

type StorageHeader Header

func (h *Header) EncodeRLP(w io.Writer) error {
	if manversion.VersionCmp(string(h.Version), manversion.VersionAIMine) >= 0 {
		sh := StorageHeader{
			ParentHash:        h.ParentHash,
			UncleHash:         h.UncleHash,
			Leader:            h.Leader,
			Coinbase:          h.Coinbase,
			Roots:             h.Roots,
			Sharding:          h.Sharding,
			Difficulty:        h.Difficulty,
			Number:            h.Number,
			GasLimit:          h.GasLimit,
			GasUsed:           h.GasUsed,
			Time:              h.Time,
			Elect:             h.Elect,
			NetTopology:       h.NetTopology,
			Signatures:        h.Signatures,
			AIHash:            h.AIHash,
			AICoinbase:        h.AICoinbase,
			Extra:             h.Extra,
			MixDigest:         h.MixDigest,
			Nonce:             h.Nonce,
			Version:           h.Version,
			VersionSignatures: h.VersionSignatures,
			VrfValue:          h.VrfValue,
			Sm3Nonce:          h.Sm3Nonce,
			BasePowers:        h.BasePowers,
		}
		return rlp.Encode(w, &sh)
	} else {
		oh := HeaderV1{
			ParentHash:        h.ParentHash,
			UncleHash:         h.UncleHash,
			Leader:            h.Leader,
			Coinbase:          h.Coinbase,
			Roots:             h.Roots,
			Sharding:          h.Sharding,
			Difficulty:        h.Difficulty,
			Number:            h.Number,
			GasLimit:          h.GasLimit,
			GasUsed:           h.GasUsed,
			Time:              h.Time,
			Elect:             h.Elect,
			NetTopology:       h.NetTopology,
			Signatures:        h.Signatures,
			Extra:             h.Extra,
			MixDigest:         h.MixDigest,
			Nonce:             h.Nonce,
			Version:           h.Version,
			VersionSignatures: h.VersionSignatures,
			VrfValue:          h.VrfValue,
		}
		return rlp.Encode(w, &oh)
	}
}

// DecodeRLP implements rlp.Decoder
/*func (h *Header) DecodeRLP(s *rlp.Stream) error {
	var sh StorageHeader
	nanoTime := time.Now().Nanosecond()
	log.Info("header decodeRlp", "开始解析数据", nanoTime)

	_, size, _ := s.Kind()
	var data []byte
	var err error
	if size != 0 {
		data, err = s.Raw()
		if err != nil {
			log.Error("header decodeRlp", "获取steam数据失败", err, "time", nanoTime)
			return err
		}
	}

	err = s.Decode(&sh)
	if err == nil {
		if manversion.VersionCmp(string(h.Version), manversion.VersionAIMine) < 0 {
			return errors.New("header version err")
		}

		h.ParentHash = sh.ParentHash
		h.UncleHash = sh.UncleHash
		h.Leader = sh.Leader
		h.Coinbase = sh.Coinbase
		h.Roots = sh.Roots
		h.Sharding = sh.Sharding
		h.Difficulty = sh.Difficulty
		h.Number = sh.Number
		h.GasLimit = sh.GasLimit
		h.GasUsed = sh.GasUsed
		h.Time = sh.Time
		h.Elect = sh.Elect
		h.NetTopology = sh.NetTopology
		h.Signatures = sh.Signatures
		h.AIHash = sh.AIHash
		h.Extra = sh.Extra
		h.MixDigest = sh.MixDigest
		h.Nonce = sh.Nonce
		h.Version = sh.Version
		h.VersionSignatures = sh.VersionSignatures
		h.VrfValue = sh.VrfValue
	}

	r := bytes.NewReader(data)
	s.Reset(r, 0)

	var oh HeaderV1
	err = s.Decode(&oh)
	if err != nil {
		return err
	}

	h.ParentHash = oh.ParentHash
	h.UncleHash = oh.UncleHash
	h.Leader = oh.Leader
	h.Coinbase = oh.Coinbase
	h.Roots = oh.Roots
	h.Sharding = oh.Sharding
	h.Difficulty = oh.Difficulty
	h.Number = oh.Number
	h.GasLimit = oh.GasLimit
	h.GasUsed = oh.GasUsed
	h.Time = oh.Time
	h.Elect = oh.Elect
	h.NetTopology = oh.NetTopology
	h.Signatures = oh.Signatures
	h.AIHash = common.Hash{}
	h.Extra = oh.Extra
	h.MixDigest = oh.MixDigest
	h.Nonce = oh.Nonce
	h.Version = oh.Version
	h.VersionSignatures = oh.VersionSignatures
	h.VrfValue = oh.VrfValue

	return nil
}
*/
