// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package main

import (
	"encoding/binary"
	"errors"
	"math"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/common/hexutil"
	"github.com/matrix/go-matrix/consensus/manash"
	"github.com/matrix/go-matrix/core"
	"github.com/matrix/go-matrix/params"
)

// cppMatrixGenesisSpec represents the genesis specification format used by the
// C++ Matrix implementation.
type cppMatrixGenesisSpec struct {
	SealEngine string `json:"sealEngine"`
	Params     struct {
		AccountStartNonce       hexutil.Uint64 `json:"accountStartNonce"`
		HomesteadForkBlock      hexutil.Uint64 `json:"homesteadForkBlock"`
		EIP150ForkBlock         hexutil.Uint64 `json:"EIP150ForkBlock"`
		EIP158ForkBlock         hexutil.Uint64 `json:"EIP158ForkBlock"`
		ByzantiumForkBlock      hexutil.Uint64 `json:"byzantiumForkBlock"`
		ConstantinopleForkBlock hexutil.Uint64 `json:"constantinopleForkBlock"`
		NetworkID               hexutil.Uint64 `json:"networkID"`
		ChainID                 hexutil.Uint64 `json:"chainID"`
		MaximumExtraDataSize    hexutil.Uint64 `json:"maximumExtraDataSize"`
		MinGasLimit             hexutil.Uint64 `json:"minGasLimit"`
		MaxGasLimit             hexutil.Uint64 `json:"maxGasLimit"`
		GasLimitBoundDivisor    hexutil.Uint64 `json:"gasLimitBoundDivisor"`
		MinimumDifficulty       *hexutil.Big   `json:"minimumDifficulty"`
		DifficultyBoundDivisor  *hexutil.Big   `json:"difficultyBoundDivisor"`
		DurationLimit           *hexutil.Big   `json:"durationLimit"`
		BlockReward             *hexutil.Big   `json:"blockReward"`
	} `json:"params"`

	Genesis struct {
		Nonce      hexutil.Bytes  `json:"nonce"`
		Difficulty *hexutil.Big   `json:"difficulty"`
		MixHash    common.Hash    `json:"mixHash"`
		Author     common.Address `json:"author"`
		Timestamp  hexutil.Uint64 `json:"timestamp"`
		ParentHash common.Hash    `json:"parentHash"`
		ExtraData  hexutil.Bytes  `json:"extraData"`
		GasLimit   hexutil.Uint64 `json:"gasLimit"`
	} `json:"genesis"`

	Accounts map[common.Address]*cppMatrixGenesisSpecAccount `json:"accounts"`
}

// cppMatrixGenesisSpecAccount is the prefunded genesis account and/or precompiled
// contract definition.
type cppMatrixGenesisSpecAccount struct {
	Balance     *hexutil.Big                   `json:"balance"`
	Nonce       uint64                         `json:"nonce,omitempty"`
	Precompiled *cppMatrixGenesisSpecBuiltin `json:"precompiled,omitempty"`
}

// cppMatrixGenesisSpecBuiltin is the precompiled contract definition.
type cppMatrixGenesisSpecBuiltin struct {
	Name          string                               `json:"name,omitempty"`
	StartingBlock hexutil.Uint64                       `json:"startingBlock,omitempty"`
	Linear        *cppMatrixGenesisSpecLinearPricing `json:"linear,omitempty"`
}

type cppMatrixGenesisSpecLinearPricing struct {
	Base uint64 `json:"base"`
	Word uint64 `json:"word"`
}

// newCppMatrixGenesisSpec converts a go-matrix genesis block into a Parity specific
// chain specification format.
func newCppMatrixGenesisSpec(network string, genesis *core.Genesis) (*cppMatrixGenesisSpec, error) {
	// Only manash is currently supported between go-matrix and cpp-matrix
	if genesis.Config.Ethash == nil {
		return nil, errors.New("unsupported consensus engine")
	}
	// Reconstruct the chain spec in Parity's format
	spec := &cppMatrixGenesisSpec{
		SealEngine: "Ethash",
	}
	spec.Params.AccountStartNonce = 0
	spec.Params.HomesteadForkBlock = (hexutil.Uint64)(genesis.Config.HomesteadBlock.Uint64())
	spec.Params.EIP150ForkBlock = (hexutil.Uint64)(genesis.Config.EIP150Block.Uint64())
	spec.Params.EIP158ForkBlock = (hexutil.Uint64)(genesis.Config.EIP158Block.Uint64())
	spec.Params.ByzantiumForkBlock = (hexutil.Uint64)(genesis.Config.ByzantiumBlock.Uint64())
	spec.Params.ConstantinopleForkBlock = (hexutil.Uint64)(math.MaxUint64)

	spec.Params.NetworkID = (hexutil.Uint64)(genesis.Config.ChainId.Uint64())
	spec.Params.ChainID = (hexutil.Uint64)(genesis.Config.ChainId.Uint64())

	spec.Params.MaximumExtraDataSize = (hexutil.Uint64)(params.MaximumExtraDataSize)
	spec.Params.MinGasLimit = (hexutil.Uint64)(params.MinGasLimit)
	spec.Params.MaxGasLimit = (hexutil.Uint64)(math.MaxUint64)
	spec.Params.MinimumDifficulty = (*hexutil.Big)(params.MinimumDifficulty)
	spec.Params.DifficultyBoundDivisor = (*hexutil.Big)(params.DifficultyBoundDivisor)
	spec.Params.GasLimitBoundDivisor = (hexutil.Uint64)(params.GasLimitBoundDivisor)
	spec.Params.DurationLimit = (*hexutil.Big)(params.DurationLimit)
	spec.Params.BlockReward = (*hexutil.Big)(manash.FrontierBlockReward)

	spec.Genesis.Nonce = (hexutil.Bytes)(make([]byte, 8))
	binary.LittleEndian.PutUint64(spec.Genesis.Nonce[:], genesis.Nonce)

	spec.Genesis.MixHash = genesis.Mixhash
	spec.Genesis.Difficulty = (*hexutil.Big)(genesis.Difficulty)
	spec.Genesis.Author = genesis.Coinbase
	spec.Genesis.Timestamp = (hexutil.Uint64)(genesis.Timestamp)
	spec.Genesis.ParentHash = genesis.ParentHash
	spec.Genesis.ExtraData = (hexutil.Bytes)(genesis.ExtraData)
	spec.Genesis.GasLimit = (hexutil.Uint64)(genesis.GasLimit)

	spec.Accounts = make(map[common.Address]*cppMatrixGenesisSpecAccount)
	for address, account := range genesis.Alloc {
		spec.Accounts[address] = &cppMatrixGenesisSpecAccount{
			Balance: (*hexutil.Big)(account.Balance),
			Nonce:   account.Nonce,
		}
	}
	spec.Accounts[common.BytesToAddress([]byte{1})].Precompiled = &cppMatrixGenesisSpecBuiltin{
		Name: "ecrecover", Linear: &cppMatrixGenesisSpecLinearPricing{Base: 3000},
	}
	spec.Accounts[common.BytesToAddress([]byte{2})].Precompiled = &cppMatrixGenesisSpecBuiltin{
		Name: "sha256", Linear: &cppMatrixGenesisSpecLinearPricing{Base: 60, Word: 12},
	}
	spec.Accounts[common.BytesToAddress([]byte{3})].Precompiled = &cppMatrixGenesisSpecBuiltin{
		Name: "ripemd160", Linear: &cppMatrixGenesisSpecLinearPricing{Base: 600, Word: 120},
	}
	spec.Accounts[common.BytesToAddress([]byte{4})].Precompiled = &cppMatrixGenesisSpecBuiltin{
		Name: "identity", Linear: &cppMatrixGenesisSpecLinearPricing{Base: 15, Word: 3},
	}
	if genesis.Config.ByzantiumBlock != nil {
		spec.Accounts[common.BytesToAddress([]byte{5})].Precompiled = &cppMatrixGenesisSpecBuiltin{
			Name: "modexp", StartingBlock: (hexutil.Uint64)(genesis.Config.ByzantiumBlock.Uint64()),
		}
		spec.Accounts[common.BytesToAddress([]byte{6})].Precompiled = &cppMatrixGenesisSpecBuiltin{
			Name: "alt_bn128_G1_add", StartingBlock: (hexutil.Uint64)(genesis.Config.ByzantiumBlock.Uint64()), Linear: &cppMatrixGenesisSpecLinearPricing{Base: 500},
		}
		spec.Accounts[common.BytesToAddress([]byte{7})].Precompiled = &cppMatrixGenesisSpecBuiltin{
			Name: "alt_bn128_G1_mul", StartingBlock: (hexutil.Uint64)(genesis.Config.ByzantiumBlock.Uint64()), Linear: &cppMatrixGenesisSpecLinearPricing{Base: 40000},
		}
		spec.Accounts[common.BytesToAddress([]byte{8})].Precompiled = &cppMatrixGenesisSpecBuiltin{
			Name: "alt_bn128_pairing_product", StartingBlock: (hexutil.Uint64)(genesis.Config.ByzantiumBlock.Uint64()),
		}
	}
	return spec, nil
}

// parityChainSpec is the chain specification format used by Parity.
type parityChainSpec struct {
	Name   string `json:"name"`
	Engine struct {
		Ethash struct {
			Params struct {
				MinimumDifficulty      *hexutil.Big `json:"minimumDifficulty"`
				DifficultyBoundDivisor *hexutil.Big `json:"difficultyBoundDivisor"`
				DurationLimit          *hexutil.Big `json:"durationLimit"`
				BlockReward            *hexutil.Big `json:"blockReward"`
				HomesteadTransition    uint64       `json:"homesteadTransition"`
				EIP150Transition       uint64       `json:"eip150Transition"`
				EIP160Transition       uint64       `json:"eip160Transition"`
				EIP161abcTransition    uint64       `json:"eip161abcTransition"`
				EIP161dTransition      uint64       `json:"eip161dTransition"`
				EIP649Reward           *hexutil.Big `json:"eip649Reward"`
				EIP100bTransition      uint64       `json:"eip100bTransition"`
				EIP649Transition       uint64       `json:"eip649Transition"`
			} `json:"params"`
		} `json:"Ethash"`
	} `json:"engine"`

	Params struct {
		MaximumExtraDataSize hexutil.Uint64 `json:"maximumExtraDataSize"`
		MinGasLimit          hexutil.Uint64 `json:"minGasLimit"`
		GasLimitBoundDivisor hexutil.Uint64 `json:"gasLimitBoundDivisor"`
		NetworkID            hexutil.Uint64 `json:"networkID"`
		MaxCodeSize          uint64         `json:"maxCodeSize"`
		EIP155Transition     uint64         `json:"eip155Transition"`
		EIP98Transition      uint64         `json:"eip98Transition"`
		EIP86Transition      uint64         `json:"eip86Transition"`
		EIP140Transition     uint64         `json:"eip140Transition"`
		EIP211Transition     uint64         `json:"eip211Transition"`
		EIP214Transition     uint64         `json:"eip214Transition"`
		EIP658Transition     uint64         `json:"eip658Transition"`
	} `json:"params"`

	Genesis struct {
		Seal struct {
			Matrix struct {
				Nonce   hexutil.Bytes `json:"nonce"`
				MixHash hexutil.Bytes `json:"mixHash"`
			} `json:"matrix"`
		} `json:"seal"`

		Difficulty *hexutil.Big   `json:"difficulty"`
		Author     common.Address `json:"author"`
		Timestamp  hexutil.Uint64 `json:"timestamp"`
		ParentHash common.Hash    `json:"parentHash"`
		ExtraData  hexutil.Bytes  `json:"extraData"`
		GasLimit   hexutil.Uint64 `json:"gasLimit"`
	} `json:"genesis"`

	Nodes    []string                                   `json:"nodes"`
	Accounts map[common.Address]*parityChainSpecAccount `json:"accounts"`
}

// parityChainSpecAccount is the prefunded genesis account and/or precompiled
// contract definition.
type parityChainSpecAccount struct {
	Balance *hexutil.Big            `json:"balance"`
	Nonce   uint64                  `json:"nonce,omitempty"`
	Builtin *parityChainSpecBuiltin `json:"builtin,omitempty"`
}

// parityChainSpecBuiltin is the precompiled contract definition.
type parityChainSpecBuiltin struct {
	Name       string                  `json:"name,omitempty"`
	ActivateAt uint64                  `json:"activate_at,omitempty"`
	Pricing    *parityChainSpecPricing `json:"pricing,omitempty"`
}

// parityChainSpecPricing represents the different pricing models that builtin
// contracts might advertise using.
type parityChainSpecPricing struct {
	Linear       *parityChainSpecLinearPricing       `json:"linear,omitempty"`
	ModExp       *parityChainSpecModExpPricing       `json:"modexp,omitempty"`
	AltBnPairing *parityChainSpecAltBnPairingPricing `json:"alt_bn128_pairing,omitempty"`
}

type parityChainSpecLinearPricing struct {
	Base uint64 `json:"base"`
	Word uint64 `json:"word"`
}

type parityChainSpecModExpPricing struct {
	Divisor uint64 `json:"divisor"`
}

type parityChainSpecAltBnPairingPricing struct {
	Base uint64 `json:"base"`
	Pair uint64 `json:"pair"`
}

// newParityChainSpec converts a go-matrix genesis block into a Parity specific
// chain specification format.
func newParityChainSpec(network string, genesis *core.Genesis, bootnodes []string) (*parityChainSpec, error) {
	// Only manash is currently supported between go-matrix and Parity
	if genesis.Config.Ethash == nil {
		return nil, errors.New("unsupported consensus engine")
	}
	// Reconstruct the chain spec in Parity's format
	spec := &parityChainSpec{
		Name:  network,
		Nodes: bootnodes,
	}
	spec.Engine.Ethash.Params.MinimumDifficulty = (*hexutil.Big)(params.MinimumDifficulty)
	spec.Engine.Ethash.Params.DifficultyBoundDivisor = (*hexutil.Big)(params.DifficultyBoundDivisor)
	spec.Engine.Ethash.Params.DurationLimit = (*hexutil.Big)(params.DurationLimit)
	spec.Engine.Ethash.Params.BlockReward = (*hexutil.Big)(manash.FrontierBlockReward)
	spec.Engine.Ethash.Params.HomesteadTransition = genesis.Config.HomesteadBlock.Uint64()
	spec.Engine.Ethash.Params.EIP150Transition = genesis.Config.EIP150Block.Uint64()
	spec.Engine.Ethash.Params.EIP160Transition = genesis.Config.EIP155Block.Uint64()
	spec.Engine.Ethash.Params.EIP161abcTransition = genesis.Config.EIP158Block.Uint64()
	spec.Engine.Ethash.Params.EIP161dTransition = genesis.Config.EIP158Block.Uint64()
	spec.Engine.Ethash.Params.EIP649Reward = (*hexutil.Big)(manash.ByzantiumBlockReward)
	spec.Engine.Ethash.Params.EIP100bTransition = genesis.Config.ByzantiumBlock.Uint64()
	spec.Engine.Ethash.Params.EIP649Transition = genesis.Config.ByzantiumBlock.Uint64()

	spec.Params.MaximumExtraDataSize = (hexutil.Uint64)(params.MaximumExtraDataSize)
	spec.Params.MinGasLimit = (hexutil.Uint64)(params.MinGasLimit)
	spec.Params.GasLimitBoundDivisor = (hexutil.Uint64)(params.GasLimitBoundDivisor)
	spec.Params.NetworkID = (hexutil.Uint64)(genesis.Config.ChainId.Uint64())
	spec.Params.MaxCodeSize = params.MaxCodeSize
	spec.Params.EIP155Transition = genesis.Config.EIP155Block.Uint64()
	spec.Params.EIP98Transition = math.MaxUint64
	spec.Params.EIP86Transition = math.MaxUint64
	spec.Params.EIP140Transition = genesis.Config.ByzantiumBlock.Uint64()
	spec.Params.EIP211Transition = genesis.Config.ByzantiumBlock.Uint64()
	spec.Params.EIP214Transition = genesis.Config.ByzantiumBlock.Uint64()
	spec.Params.EIP658Transition = genesis.Config.ByzantiumBlock.Uint64()

	spec.Genesis.Seal.Matrix.Nonce = (hexutil.Bytes)(make([]byte, 8))
	binary.LittleEndian.PutUint64(spec.Genesis.Seal.Matrix.Nonce[:], genesis.Nonce)

	spec.Genesis.Seal.Matrix.MixHash = (hexutil.Bytes)(genesis.Mixhash[:])
	spec.Genesis.Difficulty = (*hexutil.Big)(genesis.Difficulty)
	spec.Genesis.Author = genesis.Coinbase
	spec.Genesis.Timestamp = (hexutil.Uint64)(genesis.Timestamp)
	spec.Genesis.ParentHash = genesis.ParentHash
	spec.Genesis.ExtraData = (hexutil.Bytes)(genesis.ExtraData)
	spec.Genesis.GasLimit = (hexutil.Uint64)(genesis.GasLimit)

	spec.Accounts = make(map[common.Address]*parityChainSpecAccount)
	for address, account := range genesis.Alloc {
		spec.Accounts[address] = &parityChainSpecAccount{
			Balance: (*hexutil.Big)(account.Balance),
			Nonce:   account.Nonce,
		}
	}
	spec.Accounts[common.BytesToAddress([]byte{1})].Builtin = &parityChainSpecBuiltin{
		Name: "ecrecover", Pricing: &parityChainSpecPricing{Linear: &parityChainSpecLinearPricing{Base: 3000}},
	}
	spec.Accounts[common.BytesToAddress([]byte{2})].Builtin = &parityChainSpecBuiltin{
		Name: "sha256", Pricing: &parityChainSpecPricing{Linear: &parityChainSpecLinearPricing{Base: 60, Word: 12}},
	}
	spec.Accounts[common.BytesToAddress([]byte{3})].Builtin = &parityChainSpecBuiltin{
		Name: "ripemd160", Pricing: &parityChainSpecPricing{Linear: &parityChainSpecLinearPricing{Base: 600, Word: 120}},
	}
	spec.Accounts[common.BytesToAddress([]byte{4})].Builtin = &parityChainSpecBuiltin{
		Name: "identity", Pricing: &parityChainSpecPricing{Linear: &parityChainSpecLinearPricing{Base: 15, Word: 3}},
	}
	if genesis.Config.ByzantiumBlock != nil {
		spec.Accounts[common.BytesToAddress([]byte{5})].Builtin = &parityChainSpecBuiltin{
			Name: "modexp", ActivateAt: genesis.Config.ByzantiumBlock.Uint64(), Pricing: &parityChainSpecPricing{ModExp: &parityChainSpecModExpPricing{Divisor: 20}},
		}
		spec.Accounts[common.BytesToAddress([]byte{6})].Builtin = &parityChainSpecBuiltin{
			Name: "alt_bn128_add", ActivateAt: genesis.Config.ByzantiumBlock.Uint64(), Pricing: &parityChainSpecPricing{Linear: &parityChainSpecLinearPricing{Base: 500}},
		}
		spec.Accounts[common.BytesToAddress([]byte{7})].Builtin = &parityChainSpecBuiltin{
			Name: "alt_bn128_mul", ActivateAt: genesis.Config.ByzantiumBlock.Uint64(), Pricing: &parityChainSpecPricing{Linear: &parityChainSpecLinearPricing{Base: 40000}},
		}
		spec.Accounts[common.BytesToAddress([]byte{8})].Builtin = &parityChainSpecBuiltin{
			Name: "alt_bn128_pairing", ActivateAt: genesis.Config.ByzantiumBlock.Uint64(), Pricing: &parityChainSpecPricing{AltBnPairing: &parityChainSpecAltBnPairingPricing{Base: 100000, Pair: 80000}},
		}
	}
	return spec, nil
}

// pyMatrixGenesisSpec represents the genesis specification format used by the
// Python Matrix implementation.
type pyMatrixGenesisSpec struct {
	Nonce      hexutil.Bytes     `json:"nonce"`
	Timestamp  hexutil.Uint64    `json:"timestamp"`
	ExtraData  hexutil.Bytes     `json:"extraData"`
	GasLimit   hexutil.Uint64    `json:"gasLimit"`
	Difficulty *hexutil.Big      `json:"difficulty"`
	Mixhash    common.Hash       `json:"mixhash"`
	Coinbase   common.Address    `json:"coinbase"`
	Alloc      core.GenesisAlloc `json:"alloc"`
	ParentHash common.Hash       `json:"parentHash"`
}

// newPyMatrixGenesisSpec converts a go-matrix genesis block into a Parity specific
// chain specification format.
func newPyMatrixGenesisSpec(network string, genesis *core.Genesis) (*pyMatrixGenesisSpec, error) {
	// Only manash is currently supported between go-matrix and pymatrix
	if genesis.Config.Ethash == nil {
		return nil, errors.New("unsupported consensus engine")
	}
	spec := &pyMatrixGenesisSpec{
		Timestamp:  (hexutil.Uint64)(genesis.Timestamp),
		ExtraData:  genesis.ExtraData,
		GasLimit:   (hexutil.Uint64)(genesis.GasLimit),
		Difficulty: (*hexutil.Big)(genesis.Difficulty),
		Mixhash:    genesis.Mixhash,
		Coinbase:   genesis.Coinbase,
		Alloc:      genesis.Alloc,
		ParentHash: genesis.ParentHash,
	}
	spec.Nonce = (hexutil.Bytes)(make([]byte, 8))
	binary.LittleEndian.PutUint64(spec.Nonce[:], genesis.Nonce)

	return spec, nil
}
