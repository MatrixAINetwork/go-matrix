// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

// Package manapi implements the general Matrix API functions.
package manapi

import (
	"context"
	"math/big"

	"github.com/MatrixAINetwork/go-matrix/accounts"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core"
	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/core/txinterface"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/core/vm"
	"github.com/MatrixAINetwork/go-matrix/event"
	"github.com/MatrixAINetwork/go-matrix/man/downloader"
	"github.com/MatrixAINetwork/go-matrix/mandb"
	"github.com/MatrixAINetwork/go-matrix/params"
	"github.com/MatrixAINetwork/go-matrix/rpc"
)

// Backend interface provides the common API services (that are provided by
// both full and light clients) with access to necessary functions.
type Backend interface {
	// General Matrix API
	Downloader() *downloader.Downloader
	ProtocolVersion() int
	SuggestPrice(ctx context.Context) (*big.Int, error)
	ChainDb() mandb.Database
	EventMux() *event.TypeMux
	AccountManager() *accounts.Manager

	// BlockChain API
	SetHead(number uint64)
	HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Header, error)
	BlockByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Block, error)
	StateAndHeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*state.StateDBManage, *types.Header, error)
	GetBlock(ctx context.Context, blockHash common.Hash) (*types.Block, error)
	GetReceipts(ctx context.Context, blockHash common.Hash) ([]types.CoinReceipts, error)
	GetTd(blockHash common.Hash) *big.Int
	GetEVM(ctx context.Context, msg txinterface.Message, state *state.StateDBManage, header *types.Header, vmCfg vm.Config) (*vm.EVM, func() error, error)
	SubscribeChainEvent(ch chan<- core.ChainEvent) event.Subscription
	SubscribeChainHeadEvent(ch chan<- core.ChainHeadEvent) event.Subscription
	SubscribeChainSideEvent(ch chan<- core.ChainSideEvent) event.Subscription
	ImportSuperBlock(ctx context.Context, filePath string) (common.Hash, error)
	GetState() (*state.StateDBManage, error)

	// TxPool API
	SendTx(ctx context.Context, signedTx types.SelfTransaction) error
	GetPoolTransactions() (types.SelfTransactions, error)
	GetPoolTransaction(txHash common.Hash) types.SelfTransaction
	GetPoolNonce(cointyp string, ctx context.Context, addr common.Address) (uint64, error)
	Stats() (pending int, queued int)
	GetTxNmap() map[uint32]*types.Transaction
	TxPoolContent() (map[common.Address]types.SelfTransactions, map[common.Address]types.SelfTransactions)
	SubscribeNewTxsEvent(chan core.NewTxsEvent) event.Subscription //Y

	SignTx(signedTx types.SelfTransaction, chainID *big.Int, blkHash common.Hash, signHeight uint64, usingEntrust bool) (types.SelfTransaction, error) //
	SendBroadTx(ctx context.Context, signedTx types.SelfTransaction, bType bool) error                                                                 //
	FetcherNotify(hash common.Hash, number uint64)                                                                                                     //

	ChainConfig() *params.ChainConfig
	//Config() *man.Config
	NetWorkID() uint64
	SyncMode() downloader.SyncMode
	NetRPCService() *PublicNetAPI
	CurrentBlock() *types.Block
	GetDepositAccount(signAccount common.Address, blockHash common.Hash) (common.Address, error)
	GetFutureRewards(*state.StateDBManage, rpc.BlockNumber) (interface{}, error)
	Genesis() *types.Block
}

func GetAPIs(apiBackend Backend) []rpc.API {
	nonceLock := new(AddrLocker)
	return []rpc.API{
		{
			Namespace: "eth",
			Version:   "1.0",
			Service:   NewPublicMatrixAPI(apiBackend),
			Public:    true,
		},
		{
			Namespace: "man",
			Version:   "1.0",
			Service:   NewPublicMatrixAPI(apiBackend),
			Public:    true,
		}, {
			Namespace: "eth",
			Version:   "1.0",
			Service:   NewPublicBlockChainAPI(apiBackend),
			Public:    true,
		}, {
			Namespace: "man",
			Version:   "1.0",
			Service:   NewPublicBlockChainAPI(apiBackend),
			Public:    true,
		}, {
			Namespace: "eth",
			Version:   "1.0",
			Service:   NewPublicTransactionPoolAPI(apiBackend, nonceLock),
			Public:    true,
		}, {
			Namespace: "man",
			Version:   "1.0",
			Service:   NewPublicTransactionPoolAPI(apiBackend, nonceLock),
			Public:    true,
		}, {
			Namespace: "txpool",
			Version:   "1.0",
			Service:   NewPublicTxPoolAPI(apiBackend),
			Public:    true,
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPublicDebugAPI(apiBackend),
			Public:    true,
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPrivateDebugAPI(apiBackend),
		}, {
			Namespace: "eth",
			Version:   "1.0",
			Service:   NewPublicAccountAPI(apiBackend),
			Public:    true,
		}, {
			Namespace: "man",
			Version:   "1.0",
			Service:   NewPublicAccountAPI(apiBackend),
			Public:    true,
		}, {
			Namespace: "personal",
			Version:   "1.0",
			Service:   NewPrivateAccountAPI(apiBackend, nonceLock),
			Public:    false,
		},
	}
}
