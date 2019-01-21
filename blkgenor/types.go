// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package blkgenor

import (
	"errors"
	"math/big"

	"github.com/matrix/go-matrix/accounts/signhelper"
	"github.com/matrix/go-matrix/baseinterface"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core"
	"github.com/matrix/go-matrix/event"
	"github.com/matrix/go-matrix/msgsend"
	"github.com/matrix/go-matrix/olconsensus"
	"github.com/matrix/go-matrix/reelection"
)

var (
	MinerResultError        = errors.New("MinerResult Error")
	ParaNull                = errors.New("Para is null  ")
	HaveNotOKResultError    = errors.New("have no satisfy miner result")
	HaveNoGenBlockError     = errors.New("have no gen block data")
	HashNoSignNotMatchError = errors.New("hash without sign not match")
)

type Backend interface {
	BlockChain() *core.BlockChain
	TxPool() *core.TxPoolManager //Y
	EventMux() *event.TypeMux
	SignHelper() *signhelper.SignHelper
	HD() *msgsend.HD
	ReElection() *reelection.ReElection
	FetcherNotify(hash common.Hash, number uint64, addr common.Address)
	OLConsensus() *olconsensus.TopNodeService
	Random() *baseinterface.Random
	ManBlkDeal() *blkmanage.ManBlkManage
}

type VrfMsg struct {
	VrfValue []byte
	VrfProof []byte
	Hash     common.Hash
}
