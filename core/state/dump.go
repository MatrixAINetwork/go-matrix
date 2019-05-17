// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package state

import (
	"encoding/json"
	"fmt"

	"math/big"

	"bytes"
	"strconv"

	"github.com/MatrixAINetwork/go-matrix/base58"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/rlp"
	"github.com/MatrixAINetwork/go-matrix/trie"
)

type DumpAccount struct {
	Balance  string            `json:"balance"`
	Nonce    uint64            `json:"nonce"`
	Root     string            `json:"root"`
	CodeHash string            `json:"codeHash"`
	Code     string            `json:"code"`
	Storage  map[string]string `json:"storage"`
}
type CoinDump struct {
	CoinTyp  string
	DumpList []Dump
}

type Dump struct {
	Root       string                 `json:"root"`
	Accounts   map[string]DumpAccount `json:"accounts"`
	MatrixData map[string]string      `json:"matrixData"`
}
type DumpValue struct {
	Key    []byte
	GetKey []byte
	Value  []byte
}

type CodeData struct {
	CodeHash []byte
	Code     []byte
}

//Root [Account ...] [Matrix...]
//Account -> Root -> []DumpValue
type DumpDB struct {
	Root    common.Hash
	Account []DumpValue
	Matrix  []DumpValue
	//	MapAccount map[common.Address][]DumpValue
	MapAccount []MapAccountArr
	CodeDatas  []CodeData
}

type MapAccountArr struct {
	Addr     common.Address
	DumpData []DumpValue
}

func (self *StateDB) RawDumpDB() DumpDB {
	dump := DumpDB{
		Root: self.trie.Hash(),
		//MapAccount: map[common.Address][]DumpValue{},
	}

	it := trie.NewIterator(self.trie.NodeIterator(nil))
	for it.Next() {
		addr := self.trie.GetKey(it.Key)
		matrixdt := it.Value[:4]
		if bytes.Compare(matrixdt, []byte("MAN-")) == 0 {
			dump.Matrix = append(dump.Matrix, DumpValue{it.Key, addr, it.Value})
			continue
		} else {
			dump.Account = append(dump.Account, DumpValue{it.Key, addr, it.Value})
		}
		var data Account
		if err := rlp.DecodeBytes(it.Value, &data); err != nil {
			panic(err)
		}

		obj := newObject(nil, common.BytesToAddress(addr), data)
		storageIt := trie.NewIterator(obj.getTrie(self.db).NodeIterator(nil))

		//code data
		code := obj.Code(self.db)
		if code != nil && common.Bytes2Hex(code) != "" {
			dump.CodeDatas = append(dump.CodeDatas, CodeData{
				CodeHash: data.CodeHash,
				Code:     code,
			})
		}

		var mapAccountArr MapAccountArr
		for storageIt.Next() {
			keyAddr := common.BytesToAddress(addr)
			//dump.MapAccount[keyAddr] = append(dump.MapAccount[keyAddr],DumpValue{storageIt.Key,self.trie.GetKey(storageIt.Key),storageIt.Value})
			mapAccountArr.Addr = keyAddr
			mapAccountArr.DumpData = append(mapAccountArr.DumpData, DumpValue{storageIt.Key, self.trie.GetKey(storageIt.Key), storageIt.Value})
		}
		dump.MapAccount = append(dump.MapAccount, mapAccountArr)
	}
	return dump
}

func (self *StateDB) RawDump1(dbDump *DumpDB) Dump {
	dump := Dump{
		Root:       fmt.Sprintf("%x", dbDump.Root),
		Accounts:   make(map[string]DumpAccount),
		MatrixData: make(map[string]string),
	}

	for _, item := range dbDump.Account {
		var data Account
		if err := rlp.DecodeBytes(item.Value, &data); err != nil {
			panic(err)
		}
		tBalance := new(big.Int)
		var total_balance string
		for _, tAccount := range data.Balance {
			tBalance = tAccount.Balance
			str_account := strconv.Itoa(int(tAccount.AccountType))
			str_balance := str_account + ":" + tBalance.String()
			total_balance += str_balance + ","
		}
		obj := newObject(nil, common.BytesToAddress(item.GetKey), data)
		account := DumpAccount{
			Balance:  total_balance[:len(total_balance)-1],
			Nonce:    data.Nonce,
			Root:     common.Bytes2Hex(data.Root[:]),
			CodeHash: common.Bytes2Hex(data.CodeHash),
			Code:     common.Bytes2Hex(obj.Code(self.db)),
			Storage:  make(map[string]string),
		}

		//storage := dbDump.MapAccount[common.BytesToAddress(item.getKey)]

		for _, storage := range dbDump.MapAccount {
			if storage.Addr.Equal(common.BytesToAddress(item.GetKey)) {
				for _, storageIt := range storage.DumpData {
					account.Storage[common.Bytes2Hex(storageIt.GetKey)] = common.Bytes2Hex(storageIt.Value)

				}
			}
		}

		//for _, storageIt := range storage {
		//	account.Storage[common.Bytes2Hex(storageIt.getKey)] = common.Bytes2Hex(storageIt.value)
		//
		//}
		dump.Accounts[common.Bytes2Hex(item.GetKey)] = account
	}

	for _, it := range dbDump.Matrix {
		dump.MatrixData[common.Bytes2Hex(it.GetKey)] = common.Bytes2Hex(it.Value)
	}
	return dump

}
func (self *StateDB) RawDump() Dump {
	dump := Dump{
		Root:       fmt.Sprintf("%x", self.trie.Hash()),
		Accounts:   make(map[string]DumpAccount),
		MatrixData: make(map[string]string),
	}

	it := trie.NewIterator(self.trie.NodeIterator(nil))
	for it.Next() {
		addr := self.trie.GetKey(it.Key)
		matrixdt := it.Value[:4]
		if bytes.Compare(matrixdt, []byte("MAN-")) == 0 {
			dump.MatrixData[common.Bytes2Hex(addr)] = common.Bytes2Hex(it.Value[4:])
			continue
		}
		var data Account
		if err := rlp.DecodeBytes(it.Value, &data); err != nil {
			panic(err)
		}

		tBalance := new(big.Int)
		var total_balance string
		for _, tAccount := range data.Balance {
			tBalance = tAccount.Balance
			str_account := strconv.Itoa(int(tAccount.AccountType))
			str_balance := str_account + ":" + tBalance.String()
			total_balance += str_balance + ","
		}
		obj := newObject(nil, common.BytesToAddress(addr), data)
		account := DumpAccount{
			Balance:  total_balance[:len(total_balance)-1],
			Nonce:    data.Nonce,
			Root:     common.Bytes2Hex(data.Root[:]),
			CodeHash: common.Bytes2Hex(data.CodeHash),
			Code:     common.Bytes2Hex(obj.Code(self.db)),
			Storage:  make(map[string]string),
		}
		storageIt := trie.NewIterator(obj.getTrie(self.db).NodeIterator(nil))
		for storageIt.Next() {
			account.Storage[common.Bytes2Hex(self.trie.GetKey(storageIt.Key))] = common.Bytes2Hex(storageIt.Value)
		}
		strAddrs := common.Bytes2Hex(addr)
		manAddrs := base58.Base58EncodeToString("MAN", common.HexToAddress(strAddrs))
		dump.Accounts[manAddrs] = account
		//dump.Accounts[common.Bytes2Hex(addr)] = account
	}
	return dump
}

func (self *StateDB) Dump() []byte {
	json, err := json.MarshalIndent(self.RawDump(), "", "    ")
	if err != nil {
		fmt.Println("dump err", err)
	}

	return json
}

func (self *StateDB) RawDumpAcccount(address common.Address) Dump {
	dump := Dump{
		Root:     fmt.Sprintf("%x", self.trie.Hash()),
		Accounts: make(map[string]DumpAccount),
	}

	value, err := self.trie.TryGet(address[:])
	if value != nil && err == nil {
		var data Account
		if err := rlp.DecodeBytes(value, &data); err != nil {
			panic(err)
		}

		tBalance := new(big.Int)
		var total_balance string
		for _, tAccount := range data.Balance {
			tBalance = tAccount.Balance
			str_account := strconv.Itoa(int(tAccount.AccountType))
			str_balance := str_account + ":" + tBalance.String()
			total_balance += str_balance + ","
		}
		obj := newObject(nil, address, data)
		account := DumpAccount{
			Balance:  total_balance[:len(total_balance)-1],
			Nonce:    data.Nonce,
			Root:     common.Bytes2Hex(data.Root[:]),
			CodeHash: common.Bytes2Hex(data.CodeHash),
			Code:     common.Bytes2Hex(obj.Code(self.db)),
			Storage:  make(map[string]string),
		}
		storageIt := trie.NewIterator(obj.getTrie(self.db).NodeIterator(nil))
		for storageIt.Next() {
			account.Storage[common.Bytes2Hex(self.trie.GetKey(storageIt.Key))] = common.Bytes2Hex(storageIt.Value)
		}
		dump.Accounts[address.String()] = account

	}
	return dump
}

func (dbDump *DumpDB) PrintAccountMsg() {
	fmt.Println("PrintAccountMsg info")
	type EasyAccount struct {
		Balance string
		Addr    string
	}
	for _, item := range dbDump.Account {
		var data Account
		if err := rlp.DecodeBytes(item.Value, &data); err != nil {
			panic(err)
		}
		tBalance := new(big.Int)
		var total_balance string
		for _, tAccount := range data.Balance {
			tBalance = tAccount.Balance
			str_account := strconv.Itoa(int(tAccount.AccountType))
			str_balance := str_account + ":" + tBalance.String()
			total_balance += str_balance + ","
		}

		account := EasyAccount{
			Balance: total_balance[:len(total_balance)-1],
			Addr:    common.Bytes2Hex(item.GetKey),
		}

		log.Debug("PrintAccountMsg info", "Addr", account.Addr, "Balance", account.Balance)
		fmt.Println("PrintAccountMsg info", "Addr", account.Addr, "Balance", account.Balance)
	}
}
