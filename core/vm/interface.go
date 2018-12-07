// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php

package vm

import (
	"math/big"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/types"
)

// StateDB is an EVM database for full state querying.
type StateDB interface {
	CreateAccount(common.Address)

	SubBalance(uint32, common.Address, *big.Int)
	AddBalance(uint32, common.Address, *big.Int)
	GetBalance(common.Address) common.BalanceType

	GetNonce(common.Address) uint64
	SetNonce(common.Address, uint64)

	GetCodeHash(common.Address) common.Hash
	GetCode(common.Address) []byte
	SetCode(common.Address, []byte)
	GetCodeSize(common.Address) int

	AddRefund(uint64)
	GetRefund() uint64

	GetState(common.Address, common.Hash) common.Hash
	SetState(common.Address, common.Hash, common.Hash)

	CommitSaveTx()
	GetSaveTx(typ byte, key uint32, hash common.Hash, isdel bool)
	SaveTx(typ byte, key uint32, data map[common.Hash][]byte)
	NewBTrie(typ byte)

	GetStateByteArray(common.Address, common.Hash) []byte
	SetStateByteArray(common.Address, common.Hash, []byte)

	Suicide(common.Address) bool
	HasSuicided(common.Address) bool

	// Exist reports whether the given account exists in state.
	// Notably this should also return true for suicided accounts.
	Exist(common.Address) bool
	// Empty returns whether the given account is empty. Empty
	// is defined according to EIP161 (balance = nonce = code = 0).
	Empty(common.Address) bool

	RevertToSnapshot(int)
	Snapshot() int

	AddLog(*types.Log)
	GetLogs(hash common.Hash) []*types.Log
	AddPreimage(common.Hash, []byte)

	ForEachStorage(common.Address, func(common.Hash, common.Hash) bool)
	SetMatrixData(hash common.Hash, val []byte)
	GetMatrixData(hash common.Hash) (val []byte)
	DeleteMxData(hash common.Hash, val []byte)

	GetGasAuthFrom(entrustFrom common.Address, height uint64) common.Address
	GetAuthFrom(entrustFrom common.Address, height uint64) common.Address
	GetEntrustFrom(authFrom common.Address, height uint64) []common.Address
	GetAllEntrustSignFrom(authFrom common.Address) []common.Address
	GetAllEntrustGasFrom(authFrom common.Address) []common.Address
}

// CallContext provides a basic interface for the EVM calling conventions. The EVM EVM
// depends on this context being implemented for doing subcalls and initialising new EVM contracts.
type CallContext interface {
	// Call another contract
	Call(env *EVM, me ContractRef, addr common.Address, data []byte, gas, value *big.Int) ([]byte, error)
	// Take another's contract code and execute within our own context
	CallCode(env *EVM, me ContractRef, addr common.Address, data []byte, gas, value *big.Int) ([]byte, error)
	// Same as CallCode except sender and value is propagated from parent to child scope
	DelegateCall(env *EVM, me ContractRef, addr common.Address, data []byte, gas *big.Int) ([]byte, error)
	// Create a new contract
	Create(env *EVM, me ContractRef, data []byte, gas, value *big.Int) ([]byte, common.Address, error)
}
