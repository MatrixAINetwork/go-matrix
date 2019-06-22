// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package entrust

import (
	"errors"
	"fmt"
	"sync"

	"github.com/MatrixAINetwork/go-matrix/accounts"
	"github.com/MatrixAINetwork/go-matrix/base58"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/params"
)

type EntrustValue struct {
	mu           sync.RWMutex
	entrustValue map[common.Address]string
}

func newEntrustValue() *EntrustValue {
	return &EntrustValue{
		entrustValue: make(map[common.Address]string, 0),
	}
}

var (
	EntrustAccountValue = newEntrustValue()
)

func (self *EntrustValue) SetEntrustValue(data map[common.Address]string) error {
	self.mu.RLock()
	defer self.mu.RUnlock()
	entrustData, noEntrustAccounts, flag := VerifyA2AccountAndPassword(data)
	if !flag {
		return errors.New(noEntrustAccounts)
	}
	for account, password := range entrustData {
		self.entrustValue[account] = password
	}
	return nil
}
func (self *EntrustValue) GetEntrustValue() map[common.Address]string {
	self.mu.RLock()
	defer self.mu.RUnlock()
	return self.entrustValue
}

type AccountChecker interface {
	CheckAccountAndPassword(a accounts.Account, passphrase string) error
}

func SetAccountChecker(checker AccountChecker) {
	accountChecker = checker
}

var accountChecker AccountChecker

func VerifyA2AccountAndPassword(data map[common.Address]string) (map[common.Address]string, string, bool) {
	if accountChecker == nil {
		log.Error("验证A2账户", "检查器未设置", "检查器 is nil")
		return nil, "", false
	}

	entrustData := make(map[common.Address]string, 0)
	noEntrustAccounts := fmt.Sprintf("Failed to import. Please check address，password，keyStore of the following accounts\n")
	flag := true
	for address, password := range data {
		err := accountChecker.CheckAccountAndPassword(accounts.Account{Address: address}, password)
		if err != nil {
			noEntrustAccounts += fmt.Sprintf("%s\n", base58.Base58EncodeToString(params.MAN_COIN, address))
			log.ERROR("验证A2账户", "错误配置账户", base58.Base58EncodeToString(params.MAN_COIN, address))
			flag = false
			continue
		}
		entrustData[address] = password
	}
	return entrustData, noEntrustAccounts, flag
}
