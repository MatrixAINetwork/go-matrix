// Copyright 2018 The MATRIX Authors as well as Copyright 2014-2017 The go-ethereum Authors
// This file is consisted of the MATRIX library and part of the go-ethereum library.
//
// The MATRIX-ethereum library is free software: you can redistribute it and/or modify it under the terms of the MIT License.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, 
//and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject tothe following conditions:
//
//The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
//
//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, 
//WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISINGFROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE
//OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// Contains all the wrappers from the accounts package to support client side key
// management on mobile platforms.

package gman

import (
	"errors"
	"time"

	"github.com/matrix/go-matrix/accounts"
	"github.com/matrix/go-matrix/accounts/keystore"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/crypto"
)

const (
	// StandardScryptN is the N parameter of Scrypt encryption algorithm, using 256MB
	// memory and taking approximately 1s CPU time on a modern processor.
	StandardScryptN = int(keystore.StandardScryptN)

	// StandardScryptP is the P parameter of Scrypt encryption algorithm, using 256MB
	// memory and taking approximately 1s CPU time on a modern processor.
	StandardScryptP = int(keystore.StandardScryptP)

	// LightScryptN is the N parameter of Scrypt encryption algorithm, using 4MB
	// memory and taking approximately 100ms CPU time on a modern processor.
	LightScryptN = int(keystore.LightScryptN)

	// LightScryptP is the P parameter of Scrypt encryption algorithm, using 4MB
	// memory and taking approximately 100ms CPU time on a modern processor.
	LightScryptP = int(keystore.LightScryptP)
)

// Account represents a stored key.
type Account struct{ account accounts.Account }

// Accounts represents a slice of accounts.
type Accounts struct{ accounts []accounts.Account }

// Size returns the number of accounts in the slice.
func (a *Accounts) Size() int {
	return len(a.accounts)
}

// Get returns the account at the given index from the slice.
func (a *Accounts) Get(index int) (account *Account, _ error) {
	if index < 0 || index >= len(a.accounts) {
		return nil, errors.New("index out of bounds")
	}
	return &Account{a.accounts[index]}, nil
}

// Set sets the account at the given index in the slice.
func (a *Accounts) Set(index int, account *Account) error {
	if index < 0 || index >= len(a.accounts) {
		return errors.New("index out of bounds")
	}
	a.accounts[index] = account.account
	return nil
}

// GetAddress retrieves the address associated with the account.
func (a *Account) GetAddress() *Address {
	return &Address{a.account.Address}
}

// GetURL retrieves the canonical URL of the account.
func (a *Account) GetURL() string {
	return a.account.URL.String()
}

// KeyStore manages a key storage directory on disk.
type KeyStore struct{ keystore *keystore.KeyStore }

// NewKeyStore creates a keystore for the given directory.
func NewKeyStore(keydir string, scryptN, scryptP int) *KeyStore {
	return &KeyStore{keystore: keystore.NewKeyStore(keydir, scryptN, scryptP)}
}

// HasAddress reports whether a key with the given address is present.
func (ks *KeyStore) HasAddress(address *Address) bool {
	return ks.keystore.HasAddress(address.address)
}

// GetAccounts returns all key files present in the directory.
func (ks *KeyStore) GetAccounts() *Accounts {
	return &Accounts{ks.keystore.Accounts()}
}

// DeleteAccount deletes the key matched by account if the passphrase is correct.
// If a contains no filename, the address must match a unique key.
func (ks *KeyStore) DeleteAccount(account *Account, passphrase string) error {
	return ks.keystore.Delete(account.account, passphrase)
}

// SignHash calculates a ECDSA signature for the given hash. The produced signature
// is in the [R || S || V] format where V is 0 or 1.
func (ks *KeyStore) SignHash(address *Address, hash []byte) (signature []byte, _ error) {
	return ks.keystore.SignHash(accounts.Account{Address: address.address}, common.CopyBytes(hash))
}

// SignTx signs the given transaction with the requested account.
func (ks *KeyStore) SignTx(account *Account, tx *Transaction, chainID *BigInt) (*Transaction, error) {
	if chainID == nil { // Null passed from mobile app
		chainID = new(BigInt)
	}
	signed, err := ks.keystore.SignTx(account.account, tx.tx, chainID.bigint)
	if err != nil {
		return nil, err
	}
	return &Transaction{signed}, nil
}

// SignHashPassphrase signs hash if the private key matching the given address can
// be decrypted with the given passphrase. The produced signature is in the
// [R || S || V] format where V is 0 or 1.
func (ks *KeyStore) SignHashPassphrase(account *Account, passphrase string, hash []byte) (signature []byte, _ error) {
	return ks.keystore.SignHashWithPassphrase(account.account, passphrase, common.CopyBytes(hash))
}

// SignTxPassphrase signs the transaction if the private key matching the
// given address can be decrypted with the given passphrase.
func (ks *KeyStore) SignTxPassphrase(account *Account, passphrase string, tx *Transaction, chainID *BigInt) (*Transaction, error) {
	if chainID == nil { // Null passed from mobile app
		chainID = new(BigInt)
	}
	signed, err := ks.keystore.SignTxWithPassphrase(account.account, passphrase, tx.tx, chainID.bigint)
	if err != nil {
		return nil, err
	}
	return &Transaction{signed}, nil
}

// Unlock unlocks the given account indefinitely.
func (ks *KeyStore) Unlock(account *Account, passphrase string) error {
	return ks.keystore.TimedUnlock(account.account, passphrase, 0)
}

// Lock removes the private key with the given address from memory.
func (ks *KeyStore) Lock(address *Address) error {
	return ks.keystore.Lock(address.address)
}

// TimedUnlock unlocks the given account with the passphrase. The account stays
// unlocked for the duration of timeout (nanoseconds). A timeout of 0 unlocks the
// account until the program exits. The account must match a unique key file.
//
// If the account address is already unlocked for a duration, TimedUnlock extends or
// shortens the active unlock timeout. If the address was previously unlocked
// indefinitely the timeout is not altered.
func (ks *KeyStore) TimedUnlock(account *Account, passphrase string, timeout int64) error {
	return ks.keystore.TimedUnlock(account.account, passphrase, time.Duration(timeout))
}

// NewAccount generates a new key and stores it into the key directory,
// encrypting it with the passphrase.
func (ks *KeyStore) NewAccount(passphrase string) (*Account, error) {
	account, err := ks.keystore.NewAccount(passphrase)
	if err != nil {
		return nil, err
	}
	return &Account{account}, nil
}

// UpdateAccount changes the passphrase of an existing account.
func (ks *KeyStore) UpdateAccount(account *Account, passphrase, newPassphrase string) error {
	return ks.keystore.Update(account.account, passphrase, newPassphrase)
}

// ExportKey exports as a JSON key, encrypted with newPassphrase.
func (ks *KeyStore) ExportKey(account *Account, passphrase, newPassphrase string) (key []byte, _ error) {
	return ks.keystore.Export(account.account, passphrase, newPassphrase)
}

// ImportKey stores the given encrypted JSON key into the key directory.
func (ks *KeyStore) ImportKey(keyJSON []byte, passphrase, newPassphrase string) (account *Account, _ error) {
	acc, err := ks.keystore.Import(common.CopyBytes(keyJSON), passphrase, newPassphrase)
	if err != nil {
		return nil, err
	}
	return &Account{acc}, nil
}

// ImportECDSAKey stores the given encrypted JSON key into the key directory.
func (ks *KeyStore) ImportECDSAKey(key []byte, passphrase string) (account *Account, _ error) {
	privkey, err := crypto.ToECDSA(common.CopyBytes(key))
	if err != nil {
		return nil, err
	}
	acc, err := ks.keystore.ImportECDSA(privkey, passphrase)
	if err != nil {
		return nil, err
	}
	return &Account{acc}, nil
}

// ImportPreSaleKey decrypts the given Matrix presale wallet and stores
// a key file in the key directory. The key file is encrypted with the same passphrase.
func (ks *KeyStore) ImportPreSaleKey(keyJSON []byte, passphrase string) (ccount *Account, _ error) {
	account, err := ks.keystore.ImportPreSaleKey(common.CopyBytes(keyJSON), passphrase)
	if err != nil {
		return nil, err
	}
	return &Account{account}, nil
}
