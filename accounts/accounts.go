// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

// Package accounts implements high level Matrix account management.
package accounts

import (
	"math/big"

	matrix "github.com/MatrixAINetwork/go-matrix"
	"github.com/MatrixAINetwork/go-matrix/base58"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/event"
	"github.com/MatrixAINetwork/go-matrix/params"
)

// Account represents an Matrix account located at a specific location defined
// by the optional URL field.
type Account struct {
	Address common.Address `json:"address"` // Matrix account address derived from the key
	//	ManAddress string         `json:"manAddress"` //
	URL URL `json:"url"` // Optional resource locator within a backend
}

func (Ac *Account) ManAddress() string {
	return base58.Base58EncodeToString(params.MAN_COIN, Ac.Address)
}

// Wallet represents a software or hardware wallet that might contain one or more
// accounts (derived from the same seed).
type Wallet interface {
	// URL retrieves the canonical path under which this wallet is reachable. It is
	// user by upper layers to define a sorting order over all wallets from multiple
	// backends.
	URL() URL

	// Status returns a textual status to aid the user in the current state of the
	// wallet. It also returns an error indicating any failure the wallet might have
	// encountered.
	Status() (string, error)

	// Open initializes access to a wallet instance. It is not meant to unlock or
	// decrypt account keys, rather simply to establish a connection to hardware
	// wallets and/or to access derivation seeds.
	//
	// The passphrase parameter may or may not be used by the implementation of a
	// particular wallet instance. The reason there is no passwordless open method
	// is to strive towards a uniform wallet handling, oblivious to the different
	// backend providers.
	//
	// Please note, if you open a wallet, you must close it to release any allocated
	// resources (especially important when working with hardware wallets).
	Open(passphrase string) error

	// Close releases any resources held by an open wallet instance.
	Close() error

	// Accounts retrieves the list of signing accounts the wallet is currently aware
	// of. For hierarchical deterministic wallets, the list will not be exhaustive,
	// rather only contain the accounts explicitly pinned during account derivation.
	Accounts() []Account

	// Contains returns whether an account is part of this particular wallet or not.
	Contains(account Account) bool

	// Derive attempts to explicitly derive a hierarchical deterministic account at
	// the specified derivation path. If requested, the derived account will be added
	// to the wallet's tracked account list.
	Derive(path DerivationPath, pin bool) (Account, error)

	// SelfDerive sets a base account derivation path from which the wallet attempts
	// to discover non zero accounts and automatically add them to list of tracked
	// accounts.
	//
	// Note, self derivaton will increment the last component of the specified path
	// opposed to decending into a child path to allow discovering accounts starting
	// from non zero components.
	//
	// You can disable automatic account discovery by calling SelfDerive with a nil
	// chain state reader.
	SelfDerive(base DerivationPath, chain matrix.ChainStateReader)

	// SignHash requests the wallet to sign the given hash.
	//
	// It looks up the account specified either solely via its address contained within,
	// or optionally with the aid of any location metantication to sign the request (e.g.
	// a password to decrypt the account, or a PIN code o verify the transaction),
	// an AuthNeededError instance will be returned, containing infos for the user
	// about which fields or actions are needed. The user may retry by providing
	// the needed details via SignHashWithPassphrase, or by other means (e.g. unlock
	// the account in a keystore).
	SignHash(account Account, hash []byte) ([]byte, error)

	// SignTx requests the wallet to sign the given transaction.
	//
	// It looks up the account specified either solely via its address contained within,
	// or optionally with the aid of any location metadata from the embedded URL field.
	//
	// If the wallet requires additional authentication to sign the request (e.g.
	// a password to decrypt the account, or a PIN code o verify the transaction),
	// an AuthNeededError instance will be returned, containing infos for the user
	// about which fields or actions are needed. The user may retry by providing
	// the needed details via SignTxWithPassphrase, or by other means (e.g. unlock
	// the account in a keystore).
	SignTx(account Account, tx types.SelfTransaction, chainID *big.Int) (types.SelfTransaction, error)

	// SignHashWithPassphrase requests the wallet to sign the given hash with the
	// given passphrase as extra authentication information.
	//
	// It looks up the account specified either solely via its address contained within,
	// or optionally with the aid of any location metadata from the embedded URL field.
	SignHashWithPassphrase(account Account, passphrase string, hash []byte) ([]byte, error)

	// SignTxWithPassphrase requests the wallet to sign the given transaction, with the
	// given passphrase as extra authentication information.
	//
	// It looks up the account specified either solely via its address contained within,
	// or optionally with the aid of any location metadata from the embedded URL field.
	SignTxWithPassphrase(account Account, passphrase string, tx types.SelfTransaction, chainID *big.Int) (types.SelfTransaction, error)
}

// Backend is a "wallet provider" that may contain a batch of accounts they can
// sign transactions with and upon request, do so.
type Backend interface {
	// Wallets retrieves the list of wallets the backend is currently aware of.
	//
	// The returned wallets are not opened by default. For software HD wallets this
	// means that no base seeds are decrypted, and for hardware wallets that no actual
	// connection is established.
	//
	// The resulting wallet list will be sorted alphabetically based on its internal
	// URL assigned by the backend. Since wallets (especially hardware) may come and
	// go, the same wallet might appear at a different positions in the list during
	// subsequent retrievals.
	Wallets() []Wallet

	// Subscribe creates an async subscription to receive notifications when the
	// backend detects the arrival or departure of a wallet.
	Subscribe(sink chan<- WalletEvent) event.Subscription
}

// WalletEventType represents the different event types that can be fired by
// the wallet subscription subsystem.
type WalletEventType int

const (
	// WalletArrived is fired when a new wallet is detected either via USB or via
	// a filesystem event in the keystore.
	WalletArrived WalletEventType = iota

	// WalletOpened is fired when a wallet is successfully opened with the purpose
	// of starting any background processes such as automatic key derivation.
	WalletOpened

	// WalletDropped
	WalletDropped
)

// WalletEvent is an event fired by an account backend when a wallet arrival or
// departure is detected.
type WalletEvent struct {
	Wallet Wallet          // Wallet instance arrived or departed
	Kind   WalletEventType // Event type that happened in the system
}
