// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package state

import (
	"bytes"
	"fmt"
	"io"
	"math/big"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/crypto"
	"github.com/MatrixAINetwork/go-matrix/rlp"
	"sync"
)

var emptyCodeHash = crypto.Keccak256(nil)

type Code []byte

func (self Code) String() string {
	return string(self) //strings.Join(Disassemble(self), " ")
}

type Storage map[common.Hash]common.Hash

type StorageByteArray map[common.Hash][]byte

func (self Storage) String() (str string) {
	for key, value := range self {
		str += fmt.Sprintf("%X : %X\n", key, value)
	}

	return
}

func (self Storage) Copy() Storage {
	cpy := make(Storage)
	for key, value := range self {
		cpy[key] = value
	}

	return cpy
}

func (self StorageByteArray) Copy() StorageByteArray {
	cpy := make(StorageByteArray)
	for key, value := range self {
		cpy[key] = value
	}
	return cpy
}

// stateObject represents an Matrix account which is being modified.
//
// The usage pattern is as follows:
// First you need to obtain a state object.
// Account values can be accessed and modified through the object.
// Finally, call CommitTrie to write the modified storage trie into a database.
type stateObject struct {
	address  common.Address
	addrHash common.Hash // hash of matrix address of the account
	data     Account
	db       *StateDB

	readMu sync.Mutex
	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error

	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access
	code Code // contract bytecode, which gets set when code is loaded

	cachedStorage Storage // Storage entry cache to avoid duplicate reads
	dirtyStorage  Storage // Storage entries that need to be flushed to disk

	cachedStorageByteArray StorageByteArray
	dirtyStorageByteArray  StorageByteArray

	// Cache flags.
	// When an object is marked suicided it will be delete from the trie
	// during the "update" phase of the state transition.
	dirtyCode bool // true if the code was updated
	suicided  bool
	deleted   bool
}

// empty returns whether the account is considered empty.
func (s *stateObject) empty() bool {
	var amountIsZero bool
	for _, tAccount := range s.data.Balance {
		if tAccount.AccountType == common.MainAccount {
			amount := tAccount.Balance
			if amount.Cmp(big.NewInt(int64(0))) == 0 {
				amountIsZero = true
			}
			break
		}
	}
	return s.data.Nonce == 0 && amountIsZero && bytes.Equal(s.data.CodeHash, emptyCodeHash)
}

// Account is the Matrix consensus representation of accounts.
// These objects are stored in the main account trie.
type Account struct {
	Nonce    uint64
	Balance  common.BalanceType
	Root     common.Hash // merkle root of the storage trie
	CodeHash []byte
}

// newObject creates a state object.
func newObject(db *StateDB, address common.Address, data Account) *stateObject {
	if data.Balance == nil {
		//data.Balance = new(big.Int)
		//初始化账户
		data.Balance = make(common.BalanceType, 0)
		tmp := new(common.BalanceSlice)
		var i uint32
		for i = 0; i <= common.LastAccount; i++ {
			tmp.AccountType = i
			tmp.Balance = new(big.Int)
			data.Balance = append(data.Balance, *tmp)
		}
	}
	if data.CodeHash == nil {
		data.CodeHash = emptyCodeHash
	}
	return &stateObject{
		db:                     db,
		address:                address,
		addrHash:               crypto.Keccak256Hash(address[:]),
		data:                   data,
		cachedStorage:          make(Storage),
		dirtyStorage:           make(Storage),
		cachedStorageByteArray: make(StorageByteArray),
		dirtyStorageByteArray:  make(StorageByteArray),
	}
}

// EncodeRLP implements rlp.Encoder.
func (c *stateObject) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, c.data)
}

// setError remembers the first non-nil error it is called with.
func (self *stateObject) setError(err error) {
	if self.dbErr == nil {
		self.dbErr = err
	}
}

func (self *stateObject) markSuicided() {
	self.suicided = true
}

func (c *stateObject) touch() {
	c.db.journal.append(touchChange{
		account: &c.address,
	})
	if c.address == ripemd {
		// Explicitly put it in the dirty-cache, which is otherwise generated from
		// flattened journals.
		c.db.journal.dirty(c.address)
	}
}

func (c *stateObject) getTrie(db Database) Trie {
	if c.trie == nil {
		var err error
		c.trie, err = db.OpenStorageTrie(c.addrHash, c.data.Root)
		if err != nil {
			c.trie, _ = db.OpenStorageTrie(c.addrHash, common.Hash{})
			c.setError(fmt.Errorf("can't create storage trie: %v", err))
		}
	}
	return c.trie
}

// GetState returns a value in account storage.
func (self *stateObject) GetState(db Database, key common.Hash) common.Hash {
	self.readMu.Lock()
	defer self.readMu.Unlock()
	value, exists := self.cachedStorage[key]
	if exists {
		return value
	}
	// Load from DB in case it is missing.
	enc, err := self.getTrie(db).TryGet(key[:])
	if err != nil {
		self.setError(err)
		return common.Hash{}
	}
	if len(enc) > 0 {
		_, content, _, err := rlp.Split(enc)
		if err != nil {
			self.setError(err)
		}
		value.SetBytes(content)
	}
	self.cachedStorage[key] = value
	return value
}

func (self *stateObject) GetStateByteArray(db Database, key common.Hash) []byte {
	self.readMu.Lock()
	defer self.readMu.Unlock()
	value, exists := self.cachedStorageByteArray[key]
	if exists {
		return value
	}
	// Load from DB in case it is missing.
	value, err := self.getTrie(db).TryGet(key[:])
	if err == nil && len(value) != 0 {
		self.cachedStorageByteArray[key] = value
	}
	return value
}

// SetState updates a value in account storage.
func (self *stateObject) SetState(db Database, key, value common.Hash) {
	self.db.journal.append(storageChange{
		account:  &self.address,
		key:      key,
		prevalue: self.GetState(db, key),
	})
	self.setState(key, value)
}

func (self *stateObject) setState(key, value common.Hash) {
	self.cachedStorage[key] = value
	self.dirtyStorage[key] = value

}
func (self *stateObject) SetStateByteArray(db Database, key common.Hash, value []byte) {
	self.db.journal.append(storageByteArrayChange{
		account:  &self.address,
		key:      key,
		prevalue: self.GetStateByteArray(db, key),
	})
	self.setStateByteArray(key, value)
}
func (self *stateObject) setStateByteArray(key common.Hash, value []byte) {
	self.cachedStorageByteArray[key] = value
	self.dirtyStorageByteArray[key] = value
}

// updateTrie writes cached storage modifications into the object's storage trie.
func (self *stateObject) updateTrie(db Database) Trie {
	tr := self.getTrie(db)
	for key, value := range self.dirtyStorage {
		delete(self.dirtyStorage, key)
		if (value == common.Hash{}) {
			self.setError(tr.TryDelete(key[:]))
			continue
		}
		// Encoding []byte cannot fail, ok to ignore the error.
		v, _ := rlp.EncodeToBytes(bytes.TrimLeft(value[:], "\x00"))
		self.setError(tr.TryUpdate(key[:], v))
	}
	for key, value := range self.dirtyStorageByteArray {
		delete(self.dirtyStorageByteArray, key)
		if len(value) == 0 {
			self.setError(tr.TryDelete(key[:]))
			continue
		}
		self.setError(tr.TryUpdate(key[:], value))
	}
	return tr
}

// UpdateRoot sets the trie root to the current root hash of
func (self *stateObject) updateRoot(db Database) {
	self.updateTrie(db)
	self.data.Root = self.trie.Hash()
}

// CommitTrie the storage trie of the object to dwb.
// This updates the trie root.
func (self *stateObject) CommitTrie(db Database) error {
	self.updateTrie(db)
	if self.dbErr != nil {
		return self.dbErr
	}
	root, err := self.trie.Commit(nil)
	if err == nil {
		self.data.Root = root
	}
	return err
}

// AddBalance removes amount from c's balance.
// It is used to add funds to the destination account of a transfer.
func (c *stateObject) AddBalance(accountType uint32, amount *big.Int) {
	// EIP158: We must check emptiness for the objects such that the account
	// clearing (0,0,0 objects) can take effect.
	if amount.Sign() == 0 {
		if c.empty() {
			c.touch()
		}

		return
	}
	for _, tAccount := range c.Balance() {
		if tAccount.AccountType == accountType {
			if tAccount.Balance != nil {
				amt := new(big.Int).Add(tAccount.Balance, amount)
				c.SetBalance(accountType, amt)
			}
			break
		}
	}
}

// SubBalance removes amount from c's balance.
// It is used to remove funds from the origin account of a transfer.
func (c *stateObject) SubBalance(accountType uint32, amount *big.Int) {
	if amount.Sign() == 0 {
		return
	}
	for _, tAccount := range c.Balance() {
		if tAccount.AccountType == accountType {
			if tAccount.Balance != nil {
				amt := new(big.Int).Sub(tAccount.Balance, amount)
				c.SetBalance(accountType, amt)
			}
			break
		}
	}
}

func (self *stateObject) SetBalance(accountType uint32, amount *big.Int) {
	tmpPrev := make(common.BalanceType, len(self.data.Balance))
	//copy(tmpPrev, self.data.Balance)
	for i := 0; i < len(self.data.Balance); i++ {
		tmpPrev[i].AccountType = self.data.Balance[i].AccountType
		tmpPrev[i].Balance = new(big.Int).Set(self.data.Balance[i].Balance)
	}
	self.db.journal.append(balanceChange{
		account: &self.address,
		//prev:    new(big.Int).Set(self.data.Balance),
		prev: tmpPrev,
	})
	self.setBalance(accountType, amount)
}

func (self *stateObject) setBalance(accountType uint32, amount *big.Int) {
	//self.data.Balance[accountType] = amount
	for index, tAccount := range self.data.Balance {
		if tAccount.AccountType == accountType {
			self.data.Balance[index].Balance = amount
			break
		}
	}
	//fmt.Println("ZH:balance:",self.data.Balance)
}

// Return the gas back to the origin. Used by the Virtual machine or Closures
func (c *stateObject) ReturnGas(gas *big.Int) {}

func (self *stateObject) deepCopy(db *StateDB) *stateObject {
	stateObject := newObject(db, self.address, self.data)
	if self.trie != nil {
		stateObject.trie = db.db.CopyTrie(self.trie)
	}
	stateObject.code = self.code
	stateObject.dirtyStorage = self.dirtyStorage.Copy()
	stateObject.cachedStorage = self.dirtyStorage.Copy()
	stateObject.dirtyStorageByteArray = self.dirtyStorageByteArray.Copy()
	stateObject.cachedStorageByteArray = self.cachedStorageByteArray.Copy()
	stateObject.suicided = self.suicided
	stateObject.dirtyCode = self.dirtyCode
	stateObject.deleted = self.deleted
	return stateObject
}

//
// Attribute accessors
//

// Returns the address of the contract/account
func (c *stateObject) Address() common.Address {
	return c.address
}

// Code returns the contract code associated with this object, if any.
func (self *stateObject) Code(db Database) []byte {
	if self.code != nil {
		return self.code
	}
	if bytes.Equal(self.CodeHash(), emptyCodeHash) {
		return nil
	}
	code, err := db.ContractCode(self.addrHash, common.BytesToHash(self.CodeHash()))
	if err != nil {
		self.setError(fmt.Errorf("can't load code hash %x: %v", self.CodeHash(), err))
	}
	self.code = code
	return code
}

func (self *stateObject) SetCode(codeHash common.Hash, code []byte) {
	prevcode := self.Code(self.db.db)
	self.db.journal.append(codeChange{
		account:  &self.address,
		prevhash: self.CodeHash(),
		prevcode: prevcode,
	})
	self.setCode(codeHash, code)
}

func (self *stateObject) setCode(codeHash common.Hash, code []byte) {
	self.code = code
	self.data.CodeHash = codeHash[:]
	self.dirtyCode = true
}

func (self *stateObject) SetNonce(nonce uint64) {
	self.db.journal.append(nonceChange{
		account: &self.address,
		prev:    self.data.Nonce,
	})
	self.setNonce(nonce)
}

func (self *stateObject) setNonce(nonce uint64) {
	self.data.Nonce = nonce
}

func (self *stateObject) CodeHash() []byte {
	return self.data.CodeHash
}

func (self *stateObject) Balance() common.BalanceType {
	return self.data.Balance
}

func (self *stateObject) Nonce() uint64 {
	return self.data.Nonce
}

// Never called, but must be present to allow stateObject to be used
// as a vm.Account interface that also satisfies the vm.ContractRef
// interface. Interfaces are awesome.
func (self *stateObject) Value() *big.Int {
	panic("Value on stateObject should never be called")
}
