// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package signhelper

import (
	"math/big"

	"github.com/matrix/go-matrix/accounts"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/crypto"
	"github.com/pkg/errors"

	"sync"

	"github.com/matrix/go-matrix/accounts/keystore"
	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/log"
)

type MatrixEth interface {
	BlockChain() *core.BlockChain
}

type AuthReader interface {
	GetSignAccountPassword(signAccounts []common.Address) (common.Address, string, error)
	GetSignAccounts(authFrom common.Address, blockHash common.Hash) ([]common.Address, error)
	GetAuthAccount(signAccount common.Address, blockHash common.Hash) (common.Address, error)
}

var (
	ModeLog                  = "签名助手"
	ErrNilAccountManager     = errors.New("account manager is nil")
	ErrNilKeyStore           = errors.New("key store is nil")
	ErrKeyStoreCount         = errors.New("key stores is empty")
	ErrKeyStoreReflect       = errors.New("reflect key stores failed")
	ErrIllegalSignAccount    = errors.New("sign account is illegal")
	ErrReader                = errors.New("auth reader is nil")
	ErrGetAccountAndPassword = errors.New("get account and password  error")
)

type SignHelper struct {
	mu         sync.RWMutex
	keyStore   *keystore.KeyStore
	authReader AuthReader
}

func NewSignHelper() *SignHelper {
	return &SignHelper{
		keyStore:   nil,
		authReader: nil,
	}
}

func (sh *SignHelper) SetAuthReader(reader AuthReader) error {
	if reader == nil {
		return ErrReader
	}
	sh.authReader = reader
	return nil
}

func (sh *SignHelper) SetAccountManager(am *accounts.Manager) error {
	if am == nil {
		return ErrNilAccountManager
	}

	sh.mu.Lock()
	defer sh.mu.Unlock()

	keyStores := am.Backends(keystore.KeyStoreType)
	if len(keyStores) <= 0 {
		return ErrKeyStoreCount
	}
	ks, OK := keyStores[0].(*keystore.KeyStore)
	if OK == false || ks == nil {
		return ErrKeyStoreReflect
	}
	sh.keyStore = ks

	return nil
}

func (sh *SignHelper) SignHashWithValidateByReader(reader AuthReader, hash []byte, validate bool, blkHash common.Hash) (common.Signature, error) {
	signAccount, signPassword, err := sh.getSignAccountAndPassword(reader, blkHash)
	if err != nil {
		return common.Signature{}, ErrGetAccountAndPassword
	}
	if (signAccount.Address == common.Address{}) {
		return common.Signature{}, ErrIllegalSignAccount
	}

	sh.mu.RLock()
	defer sh.mu.RUnlock()
	if nil == sh.keyStore {
		return common.Signature{}, ErrNilKeyStore
	}
	sign, err := sh.keyStore.SignHashValidateWithPass(signAccount, signPassword, hash, validate)
	if err != nil {
		return common.Signature{}, err
	}
	return common.BytesToSignature(sign), nil
}

func (sh *SignHelper) SignHashWithValidate(hash []byte, validate bool, blkHash common.Hash) (common.Signature, error) {
	return sh.SignHashWithValidateByReader(sh.authReader, hash, validate, blkHash)
}

func (sh *SignHelper) SignTx(tx types.SelfTransaction, chainID *big.Int, blkHash common.Hash) (types.SelfTransaction, error) {
	// Sign the requested hash with the wallet
	signAccount, signPassword, err := sh.getSignAccountAndPassword(sh.authReader, blkHash)
	if err != nil {
		return nil, ErrGetAccountAndPassword
	}
	if (signAccount.Address == common.Address{}) {
		return nil, ErrIllegalSignAccount
	}
	sh.mu.RLock()
	defer sh.mu.RUnlock()
	if nil == sh.keyStore {
		return nil, ErrNilKeyStore
	}
	return sh.keyStore.SignTxWithPassAndTemp(signAccount, signPassword, tx, chainID)
}

func (sh *SignHelper) SignVrf(msg []byte, blkHash common.Hash) ([]byte, []byte, []byte, error) {
	signAccount, signPassword, err := sh.getSignAccountAndPassword(sh.authReader, blkHash)
	//log.ERROR(ModeLog, "signAccount", signAccount, "signPassword", signPassword, "err", err, "blkhash", blkHash)
	if err != nil {
		return []byte{}, []byte{}, []byte{}, ErrGetAccountAndPassword
	}
	if (signAccount.Address == common.Address{}) {
		return []byte{}, []byte{}, []byte{}, ErrIllegalSignAccount
	}

	sh.mu.RLock()
	defer sh.mu.RUnlock()
	if nil == sh.keyStore {
		return []byte{}, []byte{}, []byte{}, ErrNilKeyStore
	}
	return sh.keyStore.SignVrfWithPass(signAccount, signPassword, msg)
}

func (sh *SignHelper) getSignAccountAndPassword(reader AuthReader, blkHash common.Hash) (accounts.Account, string, error) {
	account := accounts.Account{}
	addrs, err := reader.GetSignAccounts(ca.GetAddress(), blkHash)
	if err != nil {
		return account, "", err
	}

	addr, password, err := reader.GetSignAccountPassword(addrs)
	account.Address = addr
	return account, password, err
}

func (sh *SignHelper) VerifySignWithValidateDependHash(signHash []byte, sig []byte, blkHash common.Hash) (common.Address, bool, error) {
	addr, flag, err := crypto.VerifySignWithValidate(signHash, sig)

	authAddr, err := sh.authReader.GetAuthAccount(addr, blkHash)
	log.ERROR(ModeLog, "addr", addr, "height", blkHash.TerminalString(), "err", err, "authAddr", authAddr)
	return authAddr, flag, err
}

func (sh *SignHelper) VerifySignWithValidateByReader(reader AuthReader, signHash []byte, sig []byte, blkHash common.Hash) (common.Address, bool, error) {
	if reader == nil {
		return common.Address{}, false, ErrReader
	}
	addr, flag, err := crypto.VerifySignWithValidate(signHash, sig)

	authAddr, err := reader.GetAuthAccount(addr, blkHash)
	log.ERROR(ModeLog, "addr", addr, "height", blkHash.TerminalString(), "err", err, "authAddr", authAddr)
	return authAddr, flag, err
}
