// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php

// Package state provides a caching layer atop the Matrix state trie.
package state

import (
	"fmt"
	"math/big"
	"sort"
	"sync"

	"encoding/json"
	"github.com/matrix/go-matrix/base58"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/crypto"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/params"
	"github.com/matrix/go-matrix/rlp"
	"github.com/matrix/go-matrix/trie"
)

type revision struct {
	id           int
	journalIndex int
}

var (
	// emptyState is the known hash of an empty state trie entry.
	emptyState = crypto.Keccak256Hash(nil)

	// emptyCode is the known hash of the empty EVM bytecode.
	emptyCode = crypto.Keccak256Hash(nil)
)

// StateDBs within the matrix protocol are used to store anything
// within the merkle trie. StateDBs take care of caching and storing
// nested states. It's the general query interface to retrieve:
// * Contracts
// * Accounts
type StateDB struct {
	db   Database
	trie Trie

	// This map holds 'live' objects, which will get modified while processing a state transition.
	stateObjects      map[common.Address]*stateObject
	stateObjectsDirty map[common.Address]struct{}

	revocablebtrie trie.BTree //可撤销
	timebtrie      trie.BTree //定时

	btreeMap        []BtreeDietyStruct
	btreeMapDirty   []BtreeDietyStruct
	matrixData      map[common.Hash][]byte
	matrixDataDirty map[common.Hash][]byte

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error

	// The refund counter, also used by state transitioning.
	refund uint64

	thash, bhash common.Hash
	txIndex      int
	logs         map[common.Hash][]*types.Log
	logSize      uint

	preimages map[common.Hash][]byte

	// Journal of state modifications. This is the backbone of
	// Snapshot and RevertToSnapshot.
	journal        *journal
	validRevisions []revision
	nextRevisionId int

	lock sync.Mutex
}
type BtreeDietyStruct struct {
	Key  uint32
	Data map[common.Hash][]byte
	Typ  string
}

// Create a new state from a given trie.
func New(root common.Hash, db Database) (*StateDB, error) {
	tr, err := db.OpenTrie(root)
	if err != nil {
		return nil, err
	}
	st := &StateDB{
		db:                db,
		trie:              tr,
		stateObjects:      make(map[common.Address]*stateObject),
		stateObjectsDirty: make(map[common.Address]struct{}),
		matrixData:        make(map[common.Hash][]byte),
		matrixDataDirty:   make(map[common.Hash][]byte),
		btreeMap:          make([]BtreeDietyStruct, 0),
		btreeMapDirty:     make([]BtreeDietyStruct, 0),
		logs:              make(map[common.Hash][]*types.Log),
		preimages:         make(map[common.Hash][]byte),
		journal:           newJournal(),
	}
	b, err1 := tr.TryGet([]byte(common.StateDBRevocableBtree))
	if err1 == nil {
		hash1 := common.BytesToHash(b)
		st.NewBTrie(common.ExtraRevocable)
		trie.RestoreBtree(&st.revocablebtrie, nil, hash1, db.TrieDB(), common.ExtraRevocable)
	}
	b2, err2 := tr.TryGet([]byte(common.StateDBTimeBtree))
	if err2 == nil {
		hash2 := common.BytesToHash(b2)
		st.NewBTrie(common.ExtraTimeTxType)
		trie.RestoreBtree(&st.timebtrie, nil, hash2, db.TrieDB(), common.ExtraTimeTxType)
	}
	return st, nil
}

// setError remembers the first non-nil error it is called with.
func (self *StateDB) setError(err error) {
	if self.dbErr == nil {
		self.dbErr = err
	}
}

func (self *StateDB) Error() error {
	return self.dbErr
}

// Reset clears out all ephemeral state objects from the state db, but keeps
// the underlying state trie to avoid reloading data for the next operations.
func (self *StateDB) Reset(root common.Hash) error {
	tr, err := self.db.OpenTrie(root)
	if err != nil {
		return err
	}
	self.trie = tr
	self.stateObjects = make(map[common.Address]*stateObject)
	self.stateObjectsDirty = make(map[common.Address]struct{})
	self.btreeMap = make([]BtreeDietyStruct, 0)
	self.btreeMapDirty = make([]BtreeDietyStruct, 0)
	self.matrixData = make(map[common.Hash][]byte)
	self.matrixDataDirty = make(map[common.Hash][]byte)
	self.thash = common.Hash{}
	self.bhash = common.Hash{}
	self.txIndex = 0
	self.logs = make(map[common.Hash][]*types.Log)
	self.logSize = 0
	self.preimages = make(map[common.Hash][]byte)
	self.clearJournalAndRefund()
	return nil
}

func (self *StateDB) AddLog(log *types.Log) {
	self.journal.append(addLogChange{txhash: self.thash})

	log.TxHash = self.thash
	log.BlockHash = self.bhash
	log.TxIndex = uint(self.txIndex)
	log.Index = self.logSize
	self.logs[self.thash] = append(self.logs[self.thash], log)
	self.logSize++
}

func (self *StateDB) GetLogs(hash common.Hash) []*types.Log {
	return self.logs[hash]
}

func (self *StateDB) Logs() []*types.Log {
	var logs []*types.Log
	for _, lgs := range self.logs {
		logs = append(logs, lgs...)
	}
	return logs
}

// AddPreimage records a SHA3 preimage seen by the VM.
func (self *StateDB) AddPreimage(hash common.Hash, preimage []byte) {
	if _, ok := self.preimages[hash]; !ok {
		self.journal.append(addPreimageChange{hash: hash})
		pi := make([]byte, len(preimage))
		copy(pi, preimage)
		self.preimages[hash] = pi
	}
}

// Preimages returns a list of SHA3 preimages that have been submitted.
func (self *StateDB) Preimages() map[common.Hash][]byte {
	return self.preimages
}

func (self *StateDB) AddRefund(gas uint64) {
	self.journal.append(refundChange{prev: self.refund})
	self.refund += gas
}

// Exist reports whether the given account address exists in the state.
// Notably this also returns true for suicided accounts.
func (self *StateDB) Exist(addr common.Address) bool {
	return self.getStateObject(addr) != nil
}

// Empty returns whether the state object is either non-existent
// or empty according to the EIP161 specification (balance = nonce = code = 0)
func (self *StateDB) Empty(addr common.Address) bool {
	so := self.getStateObject(addr)
	return so == nil || so.empty()
}

// Retrieve the balance from the given address or 0 if object not found
func (self *StateDB) GetBalance(addr common.Address) common.BalanceType {
	stateObject := self.getStateObject(addr)
	if stateObject != nil {
		return stateObject.Balance()
	}

	b := make(common.BalanceType, 0)
	tmp := new(common.BalanceSlice)
	var i uint32
	for i = 0; i <= common.LastAccount; i++ {
		tmp.AccountType = i
		tmp.Balance = new(big.Int)
		b = append(b, *tmp)
	}
	return b
}

// Retrieve the balance from the given address or 0 if object not found
func (self *StateDB) GetBalanceByType(addr common.Address, accType uint32) *big.Int {
	stateObject := self.getStateObject(addr)
	if stateObject != nil {
		for _, tAccount := range stateObject.data.Balance {
			if tAccount.AccountType == accType {
				return tAccount.Balance
			}
		}
	}

	return big.NewInt(0)
}
func (self *StateDB) GetNonce(addr common.Address) uint64 {
	stateObject := self.getStateObject(addr)
	if stateObject != nil {
		return stateObject.Nonce()
	}

	return 0 | params.NonceAddOne //YY
}

func (self *StateDB) GetCode(addr common.Address) []byte {
	stateObject := self.getStateObject(addr)
	if stateObject != nil {
		return stateObject.Code(self.db)
	}
	return nil
}

func (self *StateDB) GetCodeSize(addr common.Address) int {
	stateObject := self.getStateObject(addr)
	if stateObject == nil {
		return 0
	}
	if stateObject.code != nil {
		return len(stateObject.code)
	}
	size, err := self.db.ContractCodeSize(stateObject.addrHash, common.BytesToHash(stateObject.CodeHash()))
	if err != nil {
		self.setError(err)
	}
	return size
}

func (self *StateDB) GetCodeHash(addr common.Address) common.Hash {
	stateObject := self.getStateObject(addr)
	if stateObject == nil {
		return common.Hash{}
	}
	return common.BytesToHash(stateObject.CodeHash())
}

func (self *StateDB) GetState(addr common.Address, bhash common.Hash) common.Hash {
	stateObject := self.getStateObject(addr)
	if stateObject != nil {
		return stateObject.GetState(self.db, bhash)
	}
	return common.Hash{}
}

func (self *StateDB) GetStateByteArray(a common.Address, b common.Hash) []byte {
	stateObject := self.getStateObject(a)
	if stateObject != nil {
		return stateObject.GetStateByteArray(self.db, b)
	}
	return nil
}

//根据授权人from和高度获取委托人的from列表,返回委托人地址列表(算法组调用,仅适用委托签名)
func (self *StateDB) GetEntrustFrom(authFrom common.Address, height uint64) []common.Address {
	EntrustMarsha1Data := self.GetStateByteArray(authFrom, common.BytesToHash(authFrom[:]))
	entrustDataList := make([]common.EntrustType, 0)
	err := json.Unmarshal(EntrustMarsha1Data, &entrustDataList)
	if err != nil {
		return nil
	}
	addressList := make([]common.Address, 0)
	for _, entrustData := range entrustDataList {
		if entrustData.IsEntrustSign == true && entrustData.StartHeight <= height && entrustData.EndHeight >= height {
			entrustFrom := base58.Base58DecodeToAddress(entrustData.EntrustAddres) //string地址转0x地址
			addressList = append(addressList, entrustFrom)
		}
	}
	return addressList
}

//根据委托人from和高度获取授权人的from,返回授权人地址(算法组调用,仅适用委托签名)
func (self *StateDB) GetAuthFrom(entrustFrom common.Address, height uint64) common.Address {
	AuthMarsha1Data := self.GetStateByteArray(entrustFrom, common.BytesToHash(entrustFrom[:]))
	AuthData := new(common.AuthType) //授权数据是单个结构
	err := json.Unmarshal(AuthMarsha1Data, AuthData)
	if err != nil {
		return common.Address{}
	}
	if AuthData.IsEntrustSign == true && AuthData.StartHeight <= height && AuthData.EndHeight >= height {
		return AuthData.AuthAddres
	}
	return common.Address{}
}

//根据授权人获取所有委托签名列表,(该方法用于取消委托时调用)
func (self *StateDB) GetAllEntrustSignFrom(authFrom common.Address) []common.Address {
	EntrustMarsha1Data := self.GetStateByteArray(authFrom, common.BytesToHash(authFrom[:]))
	entrustDataList := make([]common.EntrustType, 0)
	err := json.Unmarshal(EntrustMarsha1Data, &entrustDataList)
	if err != nil {
		return nil
	}
	addressList := make([]common.Address, 0)
	for _, entrustData := range entrustDataList {
		if entrustData.IsEntrustSign == true {
			entrustFrom := base58.Base58DecodeToAddress(entrustData.EntrustAddres) //string地址转0x地址
			addressList = append(addressList, entrustFrom)
		}
	}
	return addressList
}

//根据授权人获取所有委托gas列表,(该方法用于取消委托时调用)
func (self *StateDB) GetAllEntrustGasFrom(authFrom common.Address) []common.Address {
	EntrustMarsha1Data := self.GetStateByteArray(authFrom, common.BytesToHash(authFrom[:]))
	entrustDataList := make([]common.EntrustType, 0)
	err := json.Unmarshal(EntrustMarsha1Data, &entrustDataList)
	if err != nil {
		return nil
	}
	addressList := make([]common.Address, 0)
	for _, entrustData := range entrustDataList {
		if entrustData.IsEntrustGas == true {
			entrustFrom := base58.Base58DecodeToAddress(entrustData.EntrustAddres) //string地址转0x地址
			addressList = append(addressList, entrustFrom)
		}
	}
	return addressList
}

//根据委托人from和高度获取授权人的from,返回授权人地址(内部调用,仅适用委托gas)
func (self *StateDB) GetGasAuthFrom(entrustFrom common.Address, height uint64) common.Address {
	AuthMarsha1Data := self.GetStateByteArray(entrustFrom, common.BytesToHash(entrustFrom[:]))
	AuthData := new(common.AuthType) //授权数据是单个结构
	err := json.Unmarshal(AuthMarsha1Data, AuthData)
	if err != nil {
		return common.Address{}
	}
	if AuthData.IsEntrustGas == true && AuthData.StartHeight <= height && AuthData.EndHeight >= height {
		return AuthData.AuthAddres
	}
	return common.Address{}
}

// Database retrieves the low level database supporting the lower level trie ops.
func (self *StateDB) Database() Database {
	return self.db
}

//isdel:true 表示需要从map中删除hash，false 表示不需要删除
func (self *StateDB) GetSaveTx(typ byte, key uint32, hash common.Hash, isdel bool) {
	var item trie.Item
	var str string
	data := make(map[common.Hash][]byte)
	switch typ {
	case common.ExtraRevocable:
		item = self.revocablebtrie.Get(trie.SpcialTxData{key, nil})
		std, ok := item.(trie.SpcialTxData)
		if !ok {
			return
		}
		if isdel {
			delete(std.Value_Tx, hash)
			data = std.Value_Tx
		}
		str = common.StateDBRevocableBtree
	case common.ExtraTimeTxType:
		item = self.timebtrie.Get(trie.SpcialTxData{key, nil})
		std, ok := item.(trie.SpcialTxData)
		if !ok {
			return
		}
		if isdel {
			delete(std.Value_Tx, hash)
			data = std.Value_Tx
		}
		str = common.StateDBTimeBtree
	default:

	}
	tmpB := new(BtreeDietyStruct)
	tmpB.Typ = str
	tmpB.Key = key
	tmpB.Data = data
	self.btreeMap = append(self.btreeMap, *tmpB)
	tmpBD := new(BtreeDietyStruct)
	tmpBD.Typ = str
	tmpBD.Key = key
	tmpBD.Data = data
	self.btreeMapDirty = append(self.btreeMapDirty, *tmpBD)
	self.journal.append(addBtreeChange{typ: str, key: key})
	return
}
func (self *StateDB) SaveTx(typ byte, key uint32, data map[common.Hash][]byte) {
	var str string
	switch typ {
	case common.ExtraRevocable:
		str = common.StateDBRevocableBtree
	case common.ExtraTimeTxType:
		str = common.StateDBTimeBtree
	default:

	}
	key = key
	tmpB := new(BtreeDietyStruct)
	tmpB.Typ = str
	tmpB.Key = key
	tmpB.Data = data
	self.btreeMap = append(self.btreeMap, *tmpB)
	tmpBD := new(BtreeDietyStruct)
	tmpBD.Typ = str
	tmpBD.Key = key
	tmpBD.Data = data
	self.btreeMapDirty = append(self.btreeMapDirty, *tmpBD)
	self.journal.append(addBtreeChange{typ: str, key: key})
}
func (self *StateDB) CommitSaveTx() {
	var typ byte
	log.Info("file statedb", "func CommitSaveTx:len(self.btreeMap)", self.btreeMap)
	for _, btree := range self.btreeMap {
		var hash common.Hash
		var str string
		log.Info("file statedb", "func CommitSaveTx:Key", btree.Key, "mapData", btree.Data)
		self.revocablebtrie.ReplaceOrInsert(trie.SpcialTxData{btree.Key, btree.Data})
		tmproot := self.revocablebtrie.Root()
		switch btree.Typ {
		case common.StateDBRevocableBtree:
			typ = common.ExtraRevocable
		case common.StateDBTimeBtree:
			typ = common.ExtraTimeTxType
		default:

		}
		hash = trie.BtreeSaveHash(tmproot, self.db.TrieDB(), typ)
		str = common.StateDBRevocableBtree
		b := []byte(str)
		err := self.trie.TryUpdate(b, hash.Bytes())
		if err != nil {
			log.Error("file statedb", "func CommitSaveTx:err", err)
		}
		log.Info("file statedb", "func CommitSaveTx", "ooooooooooooooooooooooooooo")
	}
	self.btreeMap = make([]BtreeDietyStruct, 0)
	self.btreeMapDirty = make([]BtreeDietyStruct, 0)
}
func (self *StateDB) UpdateTxForBtree(key uint32) {
	out := make([]trie.Item, 0)
	self.revocablebtrie.DescendLessOrEqual(trie.SpcialTxData{Key_Time: key}, func(a trie.Item) bool {
		out = append(out, a)
		return true
	})
	log.Info("file statedb", "func UpdateTxForBtree:len(out)", len(out), "time", key)
	for _, it := range out {
		item, ok := it.(trie.SpcialTxData)
		if !ok {
			continue
		}
		log.Info("file statedb", "func UpdateTxForBtree:item.key", item.Key_Time, "item.Value", item.Value_Tx)
		for hash, tm := range item.Value_Tx {
			//self.GetMatrixData(hash)
			var rt common.RecorbleTx
			errRT := json.Unmarshal(tm, &rt)
			if errRT != nil {
				log.Error("file statedb", "func UpdateTxForBtree,Unmarshal err", errRT)
				continue
			}
			for _, vv := range rt.Adam { //一对多交易
				log.Info("file statedb", "func UpdateTxForBtree:vv.Addr", vv.Addr, "vv.Amont", vv.Amont)
				log.Info("file statedb", "func UpdateTxForBtree:from", rt.From, "vv.Amont", vv.Amont)
				if self.GetBalanceByType(rt.From, common.WithdrawAccount).Cmp(vv.Amont) >= 0 {
					self.SubBalance(common.WithdrawAccount, rt.From, vv.Amont)
					self.AddBalance(common.MainAccount, vv.Addr, vv.Amont)
				}
			}
			self.GetSaveTx(common.ExtraRevocable, key, hash, true)
		}
	}
}
func (self *StateDB) UpdateTxForBtreeBytime(key uint32) {
	out := make([]trie.Item, 0)
	self.timebtrie.DescendLessOrEqual(trie.SpcialTxData{Key_Time: key}, func(a trie.Item) bool {
		out = append(out, a)
		return true
	})
	log.Info("file statedb", "func UpdateTxForBtreeBytime:len(out)", len(out), "time", key)
	for _, it := range out {
		item, ok := it.(trie.SpcialTxData)
		if !ok {
			continue
		}
		log.Info("file statedb", "func UpdateTxForBtreeBytime:item.key", item.Key_Time, "item.Value", item.Value_Tx)
		for hash, tm := range item.Value_Tx {
			//self.GetMatrixData(hash)
			var rt common.RecorbleTx
			errRT := json.Unmarshal(tm, &rt)
			if errRT != nil {
				log.Error("file statedb", "func UpdateTxForBtreeBytime,Unmarshal err", errRT)
				continue
			}
			for _, vv := range rt.Adam { //一对多交易
				log.Info("file statedb", "func UpdateTxForBtreeBytime:vv.Addr", vv.Addr, "vv.Amont", vv.Amont)
				log.Info("file statedb", "func UpdateTxForBtreeBytime:from", rt.From, "vv.Amont", vv.Amont)
				if self.GetBalanceByType(rt.From, common.WithdrawAccount).Cmp(vv.Amont) >= 0 {
					self.SubBalance(common.WithdrawAccount, rt.From, vv.Amont)
					self.AddBalance(common.MainAccount, vv.Addr, vv.Amont)
				}
			}
			self.GetSaveTx(common.ExtraTimeTxType, key, hash, true)
		}
	}
}
func (self *StateDB) NewBTrie(typ byte) {
	switch typ {
	case common.ExtraRevocable:
		self.revocablebtrie = *trie.NewBtree(2, self.db.TrieDB())
	case common.ExtraTimeTxType:
		self.timebtrie = *trie.NewBtree(2, self.db.TrieDB())
	}
}

// StorageTrie returns the storage trie of an account.
// The return value is a copy and is nil for non-existent accounts.
func (self *StateDB) StorageTrie(addr common.Address) Trie {
	stateObject := self.getStateObject(addr)
	if stateObject == nil {
		return nil
	}
	cpy := stateObject.deepCopy(self)
	return cpy.updateTrie(self.db)
}

func (self *StateDB) HasSuicided(addr common.Address) bool {
	stateObject := self.getStateObject(addr)
	if stateObject != nil {
		return stateObject.suicided
	}
	return false
}

/*
 * SETTERS
 */

// AddBalance adds amount to the account associated with addr.
func (self *StateDB) AddBalance(accountType uint32, addr common.Address, amount *big.Int) {
	stateObject := self.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.AddBalance(accountType, amount)
	}
}

// SubBalance subtracts amount from the account associated with addr.
func (self *StateDB) SubBalance(accountType uint32, addr common.Address, amount *big.Int) {
	stateObject := self.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SubBalance(accountType, amount)
	}
}

func (self *StateDB) SetBalance(accountType uint32, addr common.Address, amount *big.Int) {
	stateObject := self.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetBalance(accountType, amount)
	}
}

func (self *StateDB) SetNonce(addr common.Address, nonce uint64) {
	stateObject := self.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetNonce(nonce | params.NonceAddOne) //YY
	}
}

func (self *StateDB) SetCode(addr common.Address, code []byte) {
	stateObject := self.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetCode(crypto.Keccak256Hash(code), code)
	}
}

func (self *StateDB) SetState(addr common.Address, key, value common.Hash) {
	stateObject := self.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetState(self.db, key, value)
	}
}

func (self *StateDB) SetStateByteArray(addr common.Address, key common.Hash, value []byte) {
	stateObject := self.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetStateByteArray(self.db, key, value)
	}
}

// Suicide marks the given account as suicided.
// This clears the account balance.
//
// The account's state object is still available until the state is committed,
// getStateObject will return a non-nil account after Suicide.
func (self *StateDB) Suicide(addr common.Address) bool {
	stateObject := self.getStateObject(addr)
	if stateObject == nil {
		return false
	}
	self.journal.append(suicideChange{
		account: &addr,
		prev:    stateObject.suicided,
		//prevbalance: new(big.Int).Set(stateObject.Balance()),
		prevbalance: stateObject.Balance(),
	})
	stateObject.markSuicided()
	//stateObject.data.Balance = new(big.Int)
	stateObject.data.Balance = make(common.BalanceType, 0)

	return true
}

//
// Setting, updating & deleting state object methods.
//

// updateStateObject writes the given object to the trie.
func (self *StateDB) updateStateObject(stateObject *stateObject) {
	addr := stateObject.Address()
	data, err := rlp.EncodeToBytes(stateObject)
	if err != nil {
		panic(fmt.Errorf("can't encode object at %x: %v", addr[:], err))
	}
	self.setError(self.trie.TryUpdate(addr[:], data))
}

// deleteStateObject removes the given object from the state trie.
func (self *StateDB) deleteStateObject(stateObject *stateObject) {
	stateObject.deleted = true
	addr := stateObject.Address()
	self.setError(self.trie.TryDelete(addr[:]))
}

// Retrieve a state object given by the address. Returns nil if not found.
func (self *StateDB) getStateObject(addr common.Address) (stateObject *stateObject) {
	// Prefer 'live' objects.
	if obj := self.stateObjects[addr]; obj != nil {
		if obj.deleted {
			return nil
		}
		return obj
	}

	// Load the object from the database.
	enc, err := self.trie.TryGet(addr[:])
	if len(enc) == 0 {
		self.setError(err)
		return nil
	}
	var data Account
	if err := rlp.DecodeBytes(enc, &data); err != nil {
		log.Error("Failed to decode state object", "addr", addr, "err", err)
		return nil
	}
	// Insert into the live set.
	obj := newObject(self, addr, data)
	self.setStateObject(obj)
	return obj
}

func (self *StateDB) setStateObject(object *stateObject) {
	self.stateObjects[object.Address()] = object
}
func (self *StateDB) DeleteMxData(hash common.Hash, val []byte) {
	self.deleteMatrixData(hash, val)
}

/************************11************************************************/
func (self *StateDB) updateMatrixData(hash common.Hash, val []byte) {
	self.setError(self.trie.TryUpdate(hash[:], val))
}

func (self *StateDB) deleteMatrixData(hash common.Hash, val []byte) {
	self.setError(self.trie.TryDelete(hash[:]))
}

func (self *StateDB) GetMatrixData(hash common.Hash) (val []byte) {
	self.lock.Lock()
	defer self.lock.Unlock()
	//if val = self.matrixData[hash]; val != nil{
	//	return val
	//}

	// Load the data from the database.
	val, err := self.trie.TryGet(hash[:])
	if len(val) == 0 {
		self.setError(err)
		return nil
	}
	return
}

func (self *StateDB) SetMatrixData(hash common.Hash, val []byte) {
	self.lock.Lock()
	defer self.lock.Unlock()
	self.journal.append(addMatrixDataChange{hash: hash})
	self.matrixData[hash] = val
	self.matrixDataDirty[hash] = val
}

/**************************22***********************************************/

// Retrieve a state object or create a new state object if nil.
func (self *StateDB) GetOrNewStateObject(addr common.Address) *stateObject {
	stateObject := self.getStateObject(addr)
	if stateObject == nil || stateObject.deleted {
		stateObject, _ = self.createObject(addr)
	}
	return stateObject
}

// createObject creates a new state object. If there is an existing account with
// the given address, it is overwritten and returned as the second return value.
func (self *StateDB) createObject(addr common.Address) (newobj, prev *stateObject) {
	prev = self.getStateObject(addr)
	newobj = newObject(self, addr, Account{})
	newobj.setNonce(0 | params.NonceAddOne) // sets the object to dirty    //YY
	if prev == nil {
		self.journal.append(createObjectChange{account: &addr})
	} else {
		self.journal.append(resetObjectChange{prev: prev})
	}
	self.setStateObject(newobj)
	return newobj, prev
}

// CreateAccount explicitly creates a state object. If a state object with the address
// already exists the balance is carried over to the new account.
//
// CreateAccount is called during the EVM CREATE operation. The situation might arise that
// a contract does the following:
//
//   1. sends funds to sha(account ++ (nonce + 1))
//   2. tx_create(sha(account ++ nonce)) (note that this gets the address of 1)
//
// Carrying over the balance ensures that Maner doesn't disappear.
func (self *StateDB) CreateAccount(addr common.Address) {
	new, prev := self.createObject(addr)
	if prev != nil {
		//new.setBalance(prev.data.Balance)
		for _, tAccount := range prev.data.Balance {
			new.setBalance(tAccount.AccountType, tAccount.Balance)
		}
	}
}

func (db *StateDB) ForEachStorage(addr common.Address, cb func(key, value common.Hash) bool) {
	so := db.getStateObject(addr)
	if so == nil {
		return
	}

	// When iterating over the storage check the cache first
	for h, value := range so.cachedStorage {
		cb(h, value)
	}

	it := trie.NewIterator(so.getTrie(db.db).NodeIterator(nil))
	for it.Next() {
		// ignore cached values
		key := common.BytesToHash(db.trie.GetKey(it.Key))
		if _, ok := so.cachedStorage[key]; !ok {
			cb(key, common.BytesToHash(it.Value))
		}
	}
}

// Copy creates a deep, independent copy of the state.
// Snapshots of the copied state cannot be applied to the copy.
func (self *StateDB) Copy() *StateDB {
	self.lock.Lock()
	defer self.lock.Unlock()

	// Copy all the basic fields, initialize the memory ones
	state := &StateDB{
		db:                self.db,
		trie:              self.db.CopyTrie(self.trie),
		stateObjects:      make(map[common.Address]*stateObject, len(self.journal.dirties)),
		stateObjectsDirty: make(map[common.Address]struct{}, len(self.journal.dirties)),
		btreeMap:          make([]BtreeDietyStruct, 0),
		btreeMapDirty:     make([]BtreeDietyStruct, 0),
		matrixData:        make(map[common.Hash][]byte),
		matrixDataDirty:   make(map[common.Hash][]byte),
		refund:            self.refund,
		logs:              make(map[common.Hash][]*types.Log, len(self.logs)),
		logSize:           self.logSize,
		preimages:         make(map[common.Hash][]byte),
		journal:           newJournal(),
	}
	// Copy the dirty states, logs, and preimages
	for addr := range self.journal.dirties {
		// As documented [here](https://github.com/matrix/go-matrix/pull/16485#issuecomment-380438527),
		// and in the Finalise-method, there is a case where an object is in the journal but not
		// in the stateObjects: OOG after touch on ripeMD prior to Byzantium. Thus, we need to check for
		// nil
		if object, exist := self.stateObjects[addr]; exist {
			state.stateObjects[addr] = object.deepCopy(state)
			state.stateObjectsDirty[addr] = struct{}{}
		}
	}
	// Above, we don't copy the actual journal. This means that if the copy is copied, the
	// loop above will be a no-op, since the copy's journal is empty.
	// Thus, here we iterate over stateObjects, to enable copies of copies
	for addr := range self.stateObjectsDirty {
		if _, exist := state.stateObjects[addr]; !exist {
			state.stateObjects[addr] = self.stateObjects[addr].deepCopy(state)
			state.stateObjectsDirty[addr] = struct{}{}
		}
	}

	for hash, logs := range self.logs {
		state.logs[hash] = make([]*types.Log, len(logs))
		copy(state.logs[hash], logs)
	}
	for hash, preimage := range self.preimages {
		state.preimages[hash] = preimage
	}

	//for hash := range self.matrixDataDirty {
	//	if _, exist := state.matrixData[hash]; !exist {
	//		state.stateObjects[addr] = self.matrixData[hash].deepCopy(state)
	//		state.stateObjectsDirty[addr] = struct{}{}
	//	}
	//}
	for hash, mandata := range self.matrixData {
		state.matrixData[hash] = mandata
		state.matrixDataDirty[hash] = mandata
	}

	state.btreeMap = self.btreeMap
	state.btreeMapDirty = self.btreeMapDirty

	return state
}

// Snapshot returns an identifier for the current revision of the state.
func (self *StateDB) Snapshot() int {
	id := self.nextRevisionId
	self.nextRevisionId++
	self.validRevisions = append(self.validRevisions, revision{id, self.journal.length()})
	return id
}

// RevertToSnapshot reverts all state changes made since the given revision.
func (self *StateDB) RevertToSnapshot(revid int) {
	// Find the snapshot in the stack of valid snapshots.
	idx := sort.Search(len(self.validRevisions), func(i int) bool {
		return self.validRevisions[i].id >= revid
	})
	if idx == len(self.validRevisions) || self.validRevisions[idx].id != revid {
		panic(fmt.Errorf("revision id %v cannot be reverted", revid))
	}
	snapshot := self.validRevisions[idx].journalIndex

	// Replay the journal to undo changes and remove invalidated snapshots
	self.journal.revert(self, snapshot)
	self.validRevisions = self.validRevisions[:idx]
}

// GetRefund returns the current value of the refund counter.
func (self *StateDB) GetRefund() uint64 {
	return self.refund
}

// Finalise finalises the state by removing the self destructed objects
// and clears the journal as well as the refunds.
func (s *StateDB) Finalise(deleteEmptyObjects bool) {
	for addr := range s.journal.dirties {
		stateObject, exist := s.stateObjects[addr]
		if !exist {
			// ripeMD is 'touched' at block 1714175, in tx 0x1237f737031e40bcde4a8b7e717b2d15e3ecadfe49bb1bbc71ee9deb09c6fcf2
			// That tx goes out of gas, and although the notion of 'touched' does not exist there, the
			// touch-event will still be recorded in the journal. Since ripeMD is a special snowflake,
			// it will persist in the journal even though the journal is reverted. In this special circumstance,
			// it may exist in `s.journal.dirties` but not in `s.stateObjects`.
			// Thus, we can safely ignore it here
			continue
		}

		if stateObject.suicided || (deleteEmptyObjects && stateObject.empty()) {
			s.deleteStateObject(stateObject)
		} else {
			stateObject.updateRoot(s.db)
			s.updateStateObject(stateObject)
		}
		s.stateObjectsDirty[addr] = struct{}{}
	}

	for hash, val := range s.matrixData {
		_, isDirty := s.matrixDataDirty[hash]
		if isDirty {
			s.updateMatrixData(hash, val)
		}
		delete(s.matrixDataDirty, hash)
	}
	s.CommitSaveTx()
	// Invalidate journal because reverting across transactions is not allowed.
	s.clearJournalAndRefund()
}

// IntermediateRoot computes the current root hash of the state trie.
// It is called in between transactions to get the root hash that
// goes into transaction receipts.
func (s *StateDB) IntermediateRoot(deleteEmptyObjects bool) common.Hash {
	s.Finalise(deleteEmptyObjects)
	return s.trie.Hash()
}

// Prepare sets the current transaction hash and index and block hash which is
// used when the EVM emits new state logs.
func (self *StateDB) Prepare(thash, bhash common.Hash, ti int) {
	self.thash = thash
	self.bhash = bhash
	self.txIndex = ti
}

func (s *StateDB) clearJournalAndRefund() {
	s.journal = newJournal()
	s.validRevisions = s.validRevisions[:0]
	s.refund = 0
}

// Commit writes the state to the underlying in-memory trie database.
func (s *StateDB) Commit(deleteEmptyObjects bool) (root common.Hash, err error) {
	defer s.clearJournalAndRefund()

	for addr := range s.journal.dirties {
		s.stateObjectsDirty[addr] = struct{}{}
	}
	// Commit objects to the trie.
	for addr, stateObject := range s.stateObjects {
		_, isDirty := s.stateObjectsDirty[addr]
		switch {
		case stateObject.suicided || (isDirty && deleteEmptyObjects && stateObject.empty()):
			// If the object has been removed, don't bother syncing it
			// and just mark it for deletion in the trie.
			s.deleteStateObject(stateObject)
		case isDirty:
			// Write any contract code associated with the state object
			if stateObject.code != nil && stateObject.dirtyCode {
				s.db.TrieDB().Insert(common.BytesToHash(stateObject.CodeHash()), stateObject.code)
				stateObject.dirtyCode = false
			}
			// Write any storage changes in the state object to its storage trie.
			if err := stateObject.CommitTrie(s.db); err != nil {
				return common.Hash{}, err
			}
			// Update the object in the main account trie.
			s.updateStateObject(stateObject)
		}
		delete(s.stateObjectsDirty, addr)
	}

	for hash, val := range s.matrixData {
		_, isDirty := s.matrixDataDirty[hash]
		if isDirty {
			s.updateMatrixData(hash, val)
		}
		delete(s.matrixDataDirty, hash)
	}
	s.CommitSaveTx()
	// Write trie changes.
	root, err = s.trie.Commit(func(leaf []byte, parent common.Hash) error {
		var account Account
		if err := rlp.DecodeBytes(leaf, &account); err != nil {
			return nil
		}
		if account.Root != emptyState {
			s.db.TrieDB().Reference(account.Root, parent)
		}
		code := common.BytesToHash(account.CodeHash)
		if code != emptyCode {
			s.db.TrieDB().Reference(code, parent)
		}
		return nil
	})
	log.Debug("Trie cache stats after commit", "misses", trie.CacheMisses(), "unloads", trie.CacheUnloads())
	return root, err
}
