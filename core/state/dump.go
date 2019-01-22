// Copyright (c) 2018-2019 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package state

import (
	"encoding/json"
	"fmt"

	"math/big"

	"bytes"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/rlp"
	"github.com/matrix/go-matrix/trie"
	"strconv"
)

type DumpAccount struct {
	Balance  string            `json:"balance"`
	Nonce    uint64            `json:"nonce"`
	Root     string            `json:"root"`
	CodeHash string            `json:"codeHash"`
	Code     string            `json:"code"`
	Storage  map[string]string `json:"storage"`
}
type Dump struct {
	Root       string                 `json:"root"`
	Accounts   map[string]DumpAccount `json:"accounts"`
	MatrixData map[string]string      `json:"matrixData"`
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
		dump.Accounts[common.Bytes2Hex(addr)] = account
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
