// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package vm

import (
	"math/big"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/types"
)

//StateDB is an EVM database for full state querying.
type StateDB interface {
	SetMatrixData(hash common.Hash, val []byte)
	GetMatrixData(hash common.Hash) (val []byte)
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

type StateDBManager interface {
	CreateAccount(cointyp string, addr common.Address)
	MakeStatedb(cointyp string, isCheck bool)
	SetBalance(cointyp string, accountType uint32, addr common.Address, amount *big.Int)
	SubBalance(cointyp string, idx uint32, addr common.Address, am *big.Int)
	AddBalance(cointyp string, idx uint32, addr common.Address, am *big.Int)

	GetBalanceAll(common.Address) common.BalanceType
	GetBalance(cointyp string, addr common.Address) common.BalanceType
	GetBalanceByType(cointyp string, addr common.Address, accType uint32) *big.Int

	GetNonce(cointyp string, addr common.Address) uint64
	SetNonce(cointyp string, addr common.Address, noc uint64)

	GetCodeHash(cointyp string, addr common.Address) common.Hash
	GetCode(cointyp string, addr common.Address) []byte
	SetCode(cointyp string, addr common.Address, b []byte)
	GetCodeSize(cointyp string, addr common.Address) int

	AddRefund(cointyp string, address common.Address, gas uint64)
	GetRefund(cointyp string, address common.Address) uint64

	GetState(cointyp string, addr common.Address, hash common.Hash) common.Hash
	SetState(cointyp string, addr common.Address, hash, hash2 common.Hash)

	GetStateByteArray(cointyp string, addr common.Address, b common.Hash) []byte
	SetStateByteArray(cointyp string, addr common.Address, key common.Hash, value []byte)

	CommitSaveTx(cointyp string, addr common.Address)
	GetSaveTx(cointyp string, addr common.Address, typ byte, key uint32, hash []common.Hash, isdel bool)
	SaveTx(cointyp string, addr common.Address, typ byte, key uint32, data map[common.Hash][]byte)
	NewBTrie(cointyp string, addr common.Address, typ byte)

	Suicide(cointyp string, addr common.Address) bool
	HasSuicided(cointyp string, addr common.Address) bool

	GetEntrustStateByteArray(cointyp string, addr common.Address) []byte
	GetAuthStateByteArray(cointyp string, addr common.Address) []byte
	SetEntrustStateByteArray(cointyp string, addr common.Address, value []byte)
	SetAuthStateByteArray(cointyp string, addr common.Address, value []byte)

	//// Exist reports whether the given account exists in state.
	//// Notably this should also return true for suicided accounts.
	Exist(cointyp string, addr common.Address) bool
	// Empty returns whether the given account is empty. Empty
	// is defined according to EIP161 (balance = nonce = code = 0).
	Empty(cointyp string, addr common.Address) bool

	RevertToSnapshot(cointyp string, ss []int)
	Snapshot(cointyp string) []int

	Error() error

	AddLog(cointyp string, address common.Address, log *types.Log)
	GetLogs(cointyp string, address common.Address, hash common.Hash) []*types.Log
	Logs() []types.CoinLogs

	AddPreimage(cointype string, addr common.Address, hash common.Hash, preimage []byte)
	Preimages() map[string]map[common.Hash][]byte

	ForEachStorage(cointyp string, addr common.Address, cb func(key, value common.Hash) bool)
	IntermediateRoot(deleteEmptyObjects bool) ([]common.CoinRoot, []common.Coinbyte)
	IntermediateRootByCointype(cointype string, deleteEmptyObjects bool) common.Hash
	Prepare(thash, bhash common.Hash, ti int)
	Commit(deleteEmptyObjects bool) ([]common.CoinRoot, []common.Coinbyte, error)

	SetMatrixData(hash common.Hash, val []byte)
	GetMatrixData(hash common.Hash) (val []byte)
	DeleteMxData(hash common.Hash, val []byte)

	UpdateTxForBtree(key uint32)
	UpdateTxForBtreeBytime(key uint32)

	GetGasAuthFrom(cointyp string, entrustFrom common.Address, height uint64) common.Address
	GetAuthFrom(cointyp string, entrustFrom common.Address, height uint64) common.Address
	GetGasAuthFromByTime(cointyp string, entrustFrom common.Address, time uint64) common.Address
	GetEntrustFrom(cointyp string, authFrom common.Address, height uint64) []common.Address

	Finalise(cointyp string, deleteEmptyObjects bool)
	GetAllEntrustSignFrom(cointyp string, authFrom common.Address) []common.Address
	GetAllEntrustGasFrom(cointyp string, authFrom common.Address) []common.Address

	Dump(cointype string, address common.Address) []byte
}
