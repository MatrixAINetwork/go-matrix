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
package blockgenor

import (
	"errors"
	"github.com/matrix/go-matrix/accounts"
	"github.com/matrix/go-matrix/accounts/signhelper"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core"
	"github.com/matrix/go-matrix/mandb"
	"github.com/matrix/go-matrix/event"
	"github.com/matrix/go-matrix/hd"
	"github.com/matrix/go-matrix/reelection"
	"math/big"
)

var (
	TimeStampError          = errors.New("Timestamp Error")
	NodeIDError             = errors.New("Node Error")
	PosHeaderError          = errors.New("PosHeader Error")
	MinerResultError        = errors.New("MinerResult Error")
	MinerPosfail            = errors.New("MinerResult POS Fail")
	AccountError            = errors.New("Acccount Error")
	TxsError                = errors.New("txs Error")
	NoWallets               = errors.New("No Wallets ")
	NoAccount               = errors.New("No Account   ")
	ParaNull                = errors.New("Para is null  ")
	Noleader                = errors.New("not leader  ")
	SignaturesError         = errors.New("Signatures Error")
	FakeHeaderError         = errors.New("FakeHeader Error")
	VoteResultError         = errors.New("VoteResultError Error")
	HeightError             = errors.New("Height Error")
	HaveNotOKResultError    = errors.New("have no satisfy miner result")
	HaveNoGenBlockError     = errors.New("have no gen block data")
	HashNoSignNotMatchError = errors.New("hash without sign not match")
	maxUint256              = new(big.Int).Exp(big.NewInt(2), big.NewInt(256), big.NewInt(0))
)

type Backend interface {
	AccountManager() *accounts.Manager
	BlockChain() *core.BlockChain
	TxPool() *core.TxPool
	ChainDb() mandb.Database
	EventMux() *event.TypeMux
	SignHelper() *signhelper.SignHelper
	HD() *hd.HD
	ReElection() *reelection.ReElection
	FetcherNotify(hash common.Hash, number uint64)
}
