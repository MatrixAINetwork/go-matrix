// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

// Package state provides a caching layer atop the Matrix state trie.
package state

import (
	"fmt"
	"math/big"
	"sort"
	"sync"

	"bytes"
	"encoding/json"
	"github.com/MatrixAINetwork/go-matrix/base58"
	"github.com/MatrixAINetwork/go-matrix/btrie"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/crypto"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/params"
	"github.com/MatrixAINetwork/go-matrix/rlp"
	"github.com/MatrixAINetwork/go-matrix/trie"
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
	readMu            sync.Mutex
	stateObjects      map[common.Address]*stateObject
	stateObjectsDirty map[common.Address]struct{}

	revocablebtrie btrie.BTree //可撤销
	timebtrie      btrie.BTree //定时

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

	//	thash, bhash common.Hash
	//	txIndex      int
	logs    map[common.Hash][]*types.Log
	logSize uint

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
func newStatedb(root common.Hash, db Database) (*StateDB, error) {
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

	b, err1 := st.tryGetMatrixData(types.RlpHash(common.StateDBRevocableBtree))
	if err1 == nil {
		hash1 := common.BytesToHash(b)
		st.NewBTrie(common.ExtraRevocable)
		btrie.RestoreBtree(&st.revocablebtrie, nil, hash1, db.TrieDB(), common.ExtraRevocable, st)
	}
	b2, err2 := st.tryGetMatrixData(types.RlpHash(common.StateDBTimeBtree))
	if err2 == nil {
		hash2 := common.BytesToHash(b2)
		st.NewBTrie(common.ExtraTimeTxType)
		btrie.RestoreBtree(&st.timebtrie, nil, hash2, db.TrieDB(), common.ExtraTimeTxType, st)
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
	//self.thash = common.Hash{}
	//self.bhash = common.Hash{}
	//self.txIndex = 0
	self.logs = make(map[common.Hash][]*types.Log)
	self.logSize = 0
	self.preimages = make(map[common.Hash][]byte)
	self.clearJournalAndRefund()
	return nil
}

//func (self *StateDB) AddLog(log *types.Log) {
//	self.journal.append(addLogChange{txhash: self.thash})
//
//	log.TxHash = self.thash
//	log.BlockHash = self.bhash
//	log.TxIndex = uint(self.txIndex)
//	log.Index = self.logSize
//	self.logs[self.thash] = append(self.logs[self.thash], log)
//	self.logSize++
//}

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

	return 0 | params.NonceAddOne //
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

func (self *StateDB) GetEntrustStateByteArray(addr common.Address) []byte {
	hashkey := append([]byte("ET"), addr[:]...)
	return self.GetStateByteArray(addr, common.BytesToHash(hashkey[:]))
}
func (self *StateDB) GetAuthStateByteArray(addr common.Address) []byte {
	hashkey := append([]byte("AU"), addr[:]...)
	return self.GetStateByteArray(addr, common.BytesToHash(hashkey[:]))
}

//根据授权人from和高度获取委托人的from列表,返回委托人地址列表(算法组调用,仅适用委托签名) A2 s
func (self *StateDB) GetEntrustFrom(authFrom common.Address, height uint64) []common.Address {
	EntrustMarsha1Data := self.GetEntrustStateByteArray(authFrom)
	if len(EntrustMarsha1Data) == 0 {
		return nil
	}
	entrustDataList := make([]common.EntrustType, 0)
	err := json.Unmarshal(EntrustMarsha1Data, &entrustDataList)
	if err != nil {
		return nil
	}
	addressList := make([]common.Address, 0)
	for _, entrustData := range entrustDataList {
		if entrustData.EnstrustSetType == params.EntrustByHeight && entrustData.IsEntrustSign == true && entrustData.StartHeight <= height && entrustData.EndHeight >= height {
			entrustFrom, err := base58.Base58DecodeToAddress(entrustData.EntrustAddres) //string地址转0x地址
			if err != nil {
				return nil
			}
			addressList = append(addressList, entrustFrom)
		}
	}
	return addressList
}

//根据委托人from和高度获取授权人的from,返回授权人地址(算法组调用,仅适用委托签名)	 A1
func (self *StateDB) GetAuthFrom(entrustFrom common.Address, height uint64) common.Address {
	AuthMarsha1Data := self.GetAuthStateByteArray(entrustFrom)
	if len(AuthMarsha1Data) == 0 {
		return common.Address{}
	}
	AuthDataList := make([]common.AuthType, 0) //授权数据是结构体切片
	err := json.Unmarshal(AuthMarsha1Data, &AuthDataList)
	if err != nil {
		return common.Address{}
	}
	for _, AuthData := range AuthDataList {
		if AuthData.EnstrustSetType == params.EntrustByHeight && AuthData.IsEntrustSign == true && AuthData.StartHeight <= height && AuthData.EndHeight >= height {
			return AuthData.AuthAddres
		}
	}
	return common.Address{}
}

//根据授权人获取所有委托签名列表,(该方法用于取消委托时调用)
func (self *StateDB) GetAllEntrustSignFrom(authFrom common.Address) []common.Address {
	EntrustMarsha1Data := self.GetEntrustStateByteArray(authFrom)
	entrustDataList := make([]common.EntrustType, 0)
	err := json.Unmarshal(EntrustMarsha1Data, &entrustDataList)
	if err != nil {
		return nil
	}
	addressList := make([]common.Address, 0)
	for _, entrustData := range entrustDataList {
		if entrustData.IsEntrustSign == true {
			entrustFrom, err := base58.Base58DecodeToAddress(entrustData.EntrustAddres) //string地址转0x地址
			if err != nil {
				return nil
			}
			addressList = append(addressList, entrustFrom)
		}
	}
	return addressList
}

//根据委托人from和时间获取授权人的from,返回授权人地址(内部调用,仅适用委托gas)
func (self *StateDB) GetGasAuthFromByTime(entrustFrom common.Address, time uint64) common.Address {
	AuthMarsha1Data := self.GetAuthStateByteArray(entrustFrom)
	if len(AuthMarsha1Data) == 0 {
		return common.Address{}
	}
	AuthDataList := make([]common.AuthType, 0) //授权数据是结构体切片
	err := json.Unmarshal(AuthMarsha1Data, &AuthDataList)
	if err != nil {
		return common.Address{}
	}
	for _, AuthData := range AuthDataList {
		if AuthData.EnstrustSetType == params.EntrustByTime && AuthData.IsEntrustGas == true && AuthData.StartTime <= time && AuthData.EndTime >= time {
			return AuthData.AuthAddres
		}
	}
	return common.Address{}
}

//获取按次数返回的授权人
func (self *StateDB) GetGasAuthFromByCount(entrustFrom common.Address) common.Address {
	AuthMarsha1Data := self.GetAuthStateByteArray(entrustFrom)
	if len(AuthMarsha1Data) == 0 {
		return common.Address{}
	}
	AuthDataList := make([]common.AuthType, 0) //授权数据是结构体切片
	err := json.Unmarshal(AuthMarsha1Data, &AuthDataList)
	if err != nil {
		return common.Address{}
	}
	for _, AuthData := range AuthDataList {
		if AuthData.EnstrustSetType == params.EntrustByCount && AuthData.IsEntrustGas == true && AuthData.EntrustCount > 0 {
			return AuthData.AuthAddres
		}
	}
	return common.Address{}
}

//授权次数减1
func (self *StateDB) GasAuthCountSubOne(entrustFrom common.Address) bool {
	AuthMarsha1Data := self.GetAuthStateByteArray(entrustFrom)
	if len(AuthMarsha1Data) == 0 {
		return false
	}
	AuthDataList := make([]common.AuthType, 0) //授权数据是结构体切片
	err := json.Unmarshal(AuthMarsha1Data, &AuthDataList)
	if err != nil {
		return false
	}
	newAuthDataList := make([]common.AuthType, 0)
	for _, AuthData := range AuthDataList {
		if AuthData.EnstrustSetType == params.EntrustByCount && AuthData.IsEntrustGas == true && AuthData.EntrustCount > 0 {
			AuthData.EntrustCount--
		}
		newAuthDataList = append(newAuthDataList, AuthData)
	}
	if len(newAuthDataList) > 0 {
		marshalData, _ := json.Marshal(newAuthDataList)
		self.SetAuthStateByteArray(entrustFrom, marshalData)
	}
	return true
}

//委托人次数减1（用于钱包展示时反向查找）
func (self *StateDB) GasEntrustCountSubOne(authFrom common.Address) {
	EntrustMarsha1Data := self.GetEntrustStateByteArray(authFrom)
	if len(EntrustMarsha1Data) == 0 {
		return
	}
	EntrustDataList := make([]common.EntrustType, 0) //委托数据是结构体切片
	err := json.Unmarshal(EntrustMarsha1Data, &EntrustDataList)
	if err != nil {
		return
	}
	newEntrustDataList := make([]common.EntrustType, 0)
	for _, EntrustData := range EntrustDataList {
		if EntrustData.EnstrustSetType == params.EntrustByCount && EntrustData.IsEntrustGas == true && EntrustData.EntrustCount > 0 {
			EntrustData.EntrustCount--
		}
		newEntrustDataList = append(newEntrustDataList, EntrustData)
	}
	if len(newEntrustDataList) > 0 {
		marshalData, _ := json.Marshal(newEntrustDataList)
		self.SetEntrustStateByteArray(authFrom, marshalData)
	}
	return
}

//根据委托人from和高度获取授权人的from,返回授权人地址(内部调用,仅适用委托gas)
func (self *StateDB) GetGasAuthFrom(entrustFrom common.Address, height uint64) common.Address {
	AuthMarsha1Data := self.GetAuthStateByteArray(entrustFrom)
	if len(AuthMarsha1Data) > 0 {
		AuthDataList := make([]common.AuthType, 0) //授权数据是结构体切片
		err := json.Unmarshal(AuthMarsha1Data, &AuthDataList)
		if err != nil {
			return common.Address{}
		}
		for _, AuthData := range AuthDataList {
			if AuthData.EnstrustSetType == params.EntrustByHeight && AuthData.IsEntrustGas == true && AuthData.StartHeight <= height && AuthData.EndHeight >= height {
				return AuthData.AuthAddres
			}
		}
	}
	return common.Address{}
}

//rpc调用，获取当时状态的委托gas信息
func (self *StateDB) GetGasAuthFromByHeightAddTime(entrustFrom common.Address) common.Address {
	AuthMarsha1Data := self.GetAuthStateByteArray(entrustFrom)
	if len(AuthMarsha1Data) > 0 {
		AuthDataList := make([]common.AuthType, 0) //授权数据是结构体切片
		err := json.Unmarshal(AuthMarsha1Data, &AuthDataList)
		if err != nil {
			return common.Address{}
		}
		for _, AuthData := range AuthDataList {
			if AuthData.IsEntrustGas == true {
				return AuthData.AuthAddres
			}
		}
	}
	return common.Address{}
}

//根据授权人获取所有委托gas列表,(该方法用于取消委托时调用)
func (self *StateDB) GetAllEntrustGasFrom(authFrom common.Address) []common.Address {
	EntrustMarsha1Data := self.GetEntrustStateByteArray(authFrom)
	entrustDataList := make([]common.EntrustType, 0)
	err := json.Unmarshal(EntrustMarsha1Data, &entrustDataList)
	if err != nil {
		return nil
	}
	addressList := make([]common.Address, 0)
	for _, entrustData := range entrustDataList {
		if entrustData.IsEntrustGas == true {
			entrustFrom, err := base58.Base58DecodeToAddress(entrustData.EntrustAddres) //string地址转0x地址
			if err != nil {
				return nil
			}
			addressList = append(addressList, entrustFrom)
		}
	}
	return addressList
}

func (self *StateDB) GetEntrustFromByTime(authFrom common.Address, time uint64) []common.Address {
	EntrustMarsha1Data := self.GetEntrustStateByteArray(authFrom)
	if len(EntrustMarsha1Data) == 0 {
		return nil
	}
	entrustDataList := make([]common.EntrustType, 0)
	err := json.Unmarshal(EntrustMarsha1Data, &entrustDataList)
	if err != nil {
		return nil
	}
	addressList := make([]common.Address, 0)
	for _, entrustData := range entrustDataList {
		if entrustData.EnstrustSetType == params.EntrustByTime && entrustData.IsEntrustGas == true && entrustData.StartHeight <= time && entrustData.EndHeight >= time {
			entrustFrom, err := base58.Base58DecodeToAddress(entrustData.EntrustAddres) //string地址转0x地址
			if err != nil {
				return nil
			}
			addressList = append(addressList, entrustFrom)
		}
	}
	return addressList
}

//判断根据时间委托是否满足条件，用于执行按时间委托的交易(跑交易),此处time应该为header里的时间戳
func (self *StateDB) GetIsEntrustByTime(entrustFrom common.Address, time uint64) bool {
	AuthMarsha1Data := self.GetAuthStateByteArray(entrustFrom)
	if len(AuthMarsha1Data) == 0 {
		return false
	}
	AuthDataList := make([]common.AuthType, 0) //授权数据是结构体切片
	err := json.Unmarshal(AuthMarsha1Data, &AuthDataList)
	if err != nil {
		return false
	}
	for _, AuthData := range AuthDataList {
		if AuthData.EnstrustSetType == params.EntrustByTime && AuthData.IsEntrustGas == true && AuthData.StartTime <= time && AuthData.EndTime >= time {
			return true
		}
	}
	return false
}

//钱包调用显示
func (self *StateDB) GetAllEntrustList(authFrom common.Address) []common.EntrustType {
	//EntrustMarsha1Data := self.GetStateByteArray(authFrom, common.BytesToHash(authFrom[:]))
	EntrustMarsha1Data := self.GetEntrustStateByteArray(authFrom)
	if len(EntrustMarsha1Data) == 0 {
		return nil
	}
	entrustDataList := make([]common.EntrustType, 0)
	err := json.Unmarshal(EntrustMarsha1Data, &entrustDataList)
	if err != nil {
		return nil
	}

	return entrustDataList
}

// Database retrieves the low level database supporting the lower level trie ops.
func (self *StateDB) Database() Database {
	return self.db
}

//isdel:true 表示需要从map中删除hash，false 表示不需要删除
func (self *StateDB) GetSaveTx(typ byte, key uint32, hashlist []common.Hash, isdel bool) {
	//var item trie.Item
	var str string
	data := make(map[common.Hash][]byte)

	switch typ {
	case common.ExtraRevocable:
		log.Info("StateDBManage", "GetSaveTx:ExtraRevocable", key)
		item := self.revocablebtrie.Get(btrie.SpcialTxData{key, nil})
		std, ok := item.(btrie.SpcialTxData)
		if !ok {
			log.Info("StateDBManage", "GetSaveTx:ExtraRevocable", "item is nil")
			return
		}
		//self.revocablebtrie.Root().Printree(2)
		delitem := self.revocablebtrie.Delete(item)
		//self.revocablebtrie.Root().Printree(2)

		log.Info("StateDBManage", "revocablebtrie GetSaveTx:del item key", delitem.(btrie.SpcialTxData).Key_Time, "len(delitem.(trie.SpcialTxData).Value_Tx)", len(delitem.(btrie.SpcialTxData).Value_Tx))
		log.Info("StateDBManage", "revocablebtrie  GetSaveTx:del item key", std.Key_Time)
		if isdel {
			log.Info("StateDBManage", "revocablebtrie GetSaveTx:del item val:begin", len(std.Value_Tx))
			for _, hash := range hashlist {
				delete(std.Value_Tx, hash)
			}
			data = std.Value_Tx
			log.Info("StateDBManage", "revocablebtrie GetSaveTx:del item val:end", len(std.Value_Tx))
		}
		str = common.StateDBRevocableBtree
	case common.ExtraTimeTxType:
		log.Info("StateDBManage", "GetSaveTx:ExtraTimeTxType:Key", key)
		item := self.timebtrie.Get(btrie.SpcialTxData{key, nil})
		std, ok := item.(btrie.SpcialTxData)
		if !ok {
			log.Info("StateDBManage", "GetSaveTx:ExtraTimeTxType", "item is nil")
			return
		}
		//self.timebtrie.Root().Printree(2)
		delitem := self.timebtrie.Delete(item)
		//self.timebtrie.Root().Printree(2)

		log.Info("StateDBManage", "timebtrie GetSaveTx:del item key", delitem.(btrie.SpcialTxData).Key_Time, "len(delitem.(trie.SpcialTxData).Value_Tx)", len(delitem.(btrie.SpcialTxData).Value_Tx))
		log.Info("StateDBManage", "timebtrie GetSaveTx:del item key", std.Key_Time)
		if isdel {
			log.Info("StateDBManage", "timebtrie GetSaveTx:del item val:begin", len(std.Value_Tx))
			for _, hash := range hashlist {
				delete(std.Value_Tx, hash)
			}
			data = std.Value_Tx
			log.Info("StateDBManage", "timebtrie GetSaveTx:del item val:end", len(std.Value_Tx))
		}
		str = common.StateDBTimeBtree
	default:

	}
	var tmpB BtreeDietyStruct
	tmpB.Typ = str
	tmpB.Key = key
	tmpB.Data = data
	self.btreeMap = append(self.btreeMap, tmpB)
	var tmpBD BtreeDietyStruct
	tmpBD.Typ = str
	tmpBD.Key = key
	tmpBD.Data = data
	self.btreeMapDirty = append(self.btreeMapDirty, tmpBD)
	self.journal.append(addBtreeChange{typ: str, key: key})

	self.CommitSaveTx()
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
	var tmpB BtreeDietyStruct
	tmpB.Typ = str
	tmpB.Key = key
	tmpB.Data = data
	self.btreeMap = append(self.btreeMap, tmpB)
	var tmpBD BtreeDietyStruct
	tmpBD.Typ = str
	tmpBD.Key = key
	tmpBD.Data = data
	self.btreeMapDirty = append(self.btreeMapDirty, tmpBD)
	self.journal.append(addBtreeChange{typ: str, key: key})
}
func (self *StateDB) CommitSaveTx() {
	//var typ byte
	for _, btree := range self.btreeMap {
		var hash common.Hash
		//var btrie *trie.BTree
		log.Info("StateDBManage", "CommitSaveTx:Key", btree.Key, "mapData", btree.Data)
		switch btree.Typ {
		case common.StateDBRevocableBtree:
			if len(btree.Data) > 0 {
				self.revocablebtrie.ReplaceOrInsert(btrie.SpcialTxData{btree.Key, btree.Data})
			}
			tmproot := self.revocablebtrie.Root()
			hash = btrie.BtreeSaveHash(tmproot, self.db.TrieDB(), common.ExtraRevocable, self)
			self.updateMatrixData(types.RlpHash(common.StateDBRevocableBtree), hash[:])
		case common.StateDBTimeBtree:
			if len(btree.Data) > 0 {
				self.timebtrie.ReplaceOrInsert(btrie.SpcialTxData{btree.Key, btree.Data})
			}
			tmproot := self.timebtrie.Root()
			hash = btrie.BtreeSaveHash(tmproot, self.db.TrieDB(), common.ExtraTimeTxType, self)
			self.updateMatrixData(types.RlpHash(common.StateDBTimeBtree), hash[:])
		default:

		}
	}
	self.btreeMap = make([]BtreeDietyStruct, 0)
	self.btreeMapDirty = make([]BtreeDietyStruct, 0)
}
func (self *StateDB) GetBtreeItem(key uint32, typ byte) []btrie.Item {
	out := make([]btrie.Item, 0)
	switch typ {
	case common.ExtraRevocable:
		self.revocablebtrie.DescendLessOrEqual(btrie.SpcialTxData{Key_Time: key}, func(a btrie.Item) bool {
			out = append(out, a)
			return true
		})
	case common.ExtraTimeTxType:
		self.timebtrie.DescendLessOrEqual(btrie.SpcialTxData{Key_Time: key}, func(a btrie.Item) bool {
			out = append(out, a)
			return true
		})
	}
	return out
}

func (self *StateDB) NewBTrie(typ byte) {
	switch typ {
	case common.ExtraRevocable:
		self.revocablebtrie = *btrie.NewBtree(2, self.db.TrieDB())
	case common.ExtraTimeTxType:
		self.timebtrie = *btrie.NewBtree(2, self.db.TrieDB())
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
		stateObject.SetNonce(nonce | params.NonceAddOne) //
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
func (self *StateDB) SetEntrustStateByteArray(addr common.Address, value []byte) {
	hashkey := append([]byte("ET"), addr[:]...)
	self.SetStateByteArray(addr, common.BytesToHash(hashkey[:]), value)
}
func (self *StateDB) SetAuthStateByteArray(addr common.Address, value []byte) {
	hashkey := append([]byte("AU"), addr[:]...)
	self.SetStateByteArray(addr, common.BytesToHash(hashkey[:]), value)
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
	self.readMu.Lock()
	defer self.readMu.Unlock()
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
	vl := append([]byte("MAN-"), val...)
	self.setError(self.trie.TryUpdate(hash[:], vl))
}

func (self *StateDB) tryGetMatrixData(hash common.Hash) (val []byte, err error) {
	tmpval, err := self.trie.TryGet(hash[:])
	if err != nil || len(tmpval) == 0 {
		return nil, err
	}
	if bytes.Compare(tmpval[:4], []byte("MAN-")) == 0 {
		val = tmpval[4:] //去掉"MAN-"前綴
	} else {
		val = tmpval
	}
	return
}
func (self *StateDB) deleteMatrixData(hash common.Hash, val []byte) {
	self.setError(self.trie.TryDelete(hash[:]))
}

func (self *StateDB) GetMatrixData(hash common.Hash) (val []byte) {
	self.readMu.Lock()
	defer self.readMu.Unlock()
	val, exist := self.matrixData[hash]
	if exist {
		return val
	}

	// Load the data from the database.
	tmpval, err := self.trie.TryGet(hash[:])
	if len(tmpval) == 0 {
		self.setError(err)
		return nil
	}
	if bytes.Compare(tmpval[:4], []byte("MAN-")) == 0 {
		val = tmpval[4:] //去掉"MAN-"前綴
	} else {
		val = tmpval
	}
	self.matrixData[hash] = val
	return
}

func (self *StateDB) SetMatrixData(hash common.Hash, val []byte) {
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
	newobj.setNonce(0 | params.NonceAddOne) // sets the object to dirty    //
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
		// As documented [here](https://github.com/MatrixAINetwork/go-matrix/pull/16485#issuecomment-380438527),
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
	len1 := len(self.validRevisions)
	if len1 == 0 || self.validRevisions[len1-1].journalIndex < self.journal.length() {
		self.validRevisions = append(self.validRevisions, revision{id, self.journal.length()})
	}
	return id
}

// RevertToSnapshot reverts all state changes made since the given revision.
func (self *StateDB) RevertToSnapshot(revid int) {
	// Find the snapshot in the stack of valid snapshots.
	/*
		idx := sort.Search(len(self.validRevisions), func(i int) bool {
			if i == len(self.validRevisions)-1 {
				return self.validRevisions[i].id <= revid
			} else {
				return self.validRevisions[i].id <= revid && self.validRevisions[i+1].id > revid
			}
		})
		if idx == len(self.validRevisions) || self.validRevisions[idx].id > revid {
			//		panic(fmt.Errorf("revision id %v cannot be reverted", revid))
			idx--
		}
	*/
	idx := 0
	for i := len(self.validRevisions) - 1; i >= 0; i-- {
		if self.validRevisions[i].id <= revid {
			idx = i
			break
		}
	}
	if self.validRevisions[idx].id > revid {
		panic(fmt.Errorf("revision id %v cannot be reverted", revid))
		//		idx--
	}
	snapshot := self.validRevisions[idx].journalIndex

	// Replay the journal to undo changes and remove invalidated snapshots
	self.journal.revert(self, snapshot)
	if self.validRevisions[idx].id == revid {
		self.validRevisions = self.validRevisions[:idx]
	} else {
		self.validRevisions = self.validRevisions[:idx+1]
	}
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
			//stateObject.updateTrie(s.db)
			s.updateStateObject(stateObject)
		}
		s.stateObjectsDirty[addr] = struct{}{}
	}

	for hash, val := range s.matrixData {
		_, isDirty := s.matrixDataDirty[hash]
		if !isDirty {
			continue
		}
		s.updateMatrixData(hash, val)
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
//func (self *StateDB) Prepare(thash, bhash common.Hash, ti int) {
//	self.thash = thash
//	self.bhash = bhash
//	self.txIndex = ti
//}

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
		if !isDirty {
			continue
		}
		s.updateMatrixData(hash, val)
		delete(s.matrixDataDirty, hash)
	}
	s.CommitSaveTx()
	// Write trie changes.
	root, err = s.trie.Commit(func(leaf []byte, parent common.Hash) error {
		var account Account
		if err := rlp.DecodeBytes(leaf, &account); err != nil {
			return err
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
	//log.Debug("Trie cache stats after commit", "misses", trie.CacheMisses(), "unloads", trie.CacheUnloads())
	return root, err
}

func (self *StateDB) MissTrieDebug() {
	self.lock.Lock()
	defer self.lock.Unlock()

	log.Info("miss tree node debug", "data amount", len(self.matrixData), "dirty amount", len(self.matrixDataDirty))
	matrixKeys := make([]common.Hash, 0)
	for key := range self.matrixData {
		matrixKeys = append(matrixKeys, key)
	}

	sort.Slice(matrixKeys, func(i, j int) bool {
		return matrixKeys[i].Big().Cmp(matrixKeys[j].Big()) <= 0
	})

	for i, k := range matrixKeys {
		log.Info("miss tree node debug", "data index", i, "hash", k.TerminalString(), "data", self.matrixData[k])
	}

	dirtyKeys := make([]common.Hash, 0)
	for key := range self.matrixDataDirty {
		dirtyKeys = append(dirtyKeys, key)
	}

	sort.Slice(dirtyKeys, func(i, j int) bool {
		return dirtyKeys[i].Big().Cmp(dirtyKeys[j].Big()) <= 0
	})

	for i, k := range dirtyKeys {
		log.Info("miss tree node debug", "dirty index", i, "hash", k.TerminalString(), "data", self.matrixDataDirty[k])
	}
}
