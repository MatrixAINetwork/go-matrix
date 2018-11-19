package manclient

import (
	"context"
	"math/big"

	"github.com/matrix/go-matrix"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/p2p"
)

type ManClient interface {
	Close()

	// eth
	BlockNumber(ctx context.Context) (*big.Int, error)
	SendRawTransaction(ctx context.Context, tx *types.Transaction) error

	// admin
	AddPeer(ctx context.Context, nodeURL string) error
	AdminPeers(ctx context.Context) ([]*p2p.PeerInfo, error)
	NodeInfo(ctx context.Context) (*p2p.PeerInfo, error)

	// miner
	StartMining(ctx context.Context) error
	StopMining(ctx context.Context) error

	// eth client
	BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error)
	BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error)
	HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error)
	HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
	TransactionByHash(ctx context.Context, hash common.Hash) (*types.Transaction, bool, error)
	TransactionCount(ctx context.Context, blockHash common.Hash) (uint, error)
	TransactionInBlock(ctx context.Context, blockHash common.Hash, index uint) (*types.Transaction, error)
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
	SyncProgress(ctx context.Context) (*matrix.SyncProgress, error)
	SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) (matrix.Subscription, error)
	NetworkID(ctx context.Context) (*big.Int, error)
	BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error)
	StorageAt(ctx context.Context, account common.Address, key common.Hash, blockNumber *big.Int) ([]byte, error)
	CodeAt(ctx context.Context, account common.Address, blockNumber *big.Int) ([]byte, error)
	NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error)
	FilterLogs(ctx context.Context, q matrix.FilterQuery) ([]types.Log, error)
	SubscribeFilterLogs(ctx context.Context, q matrix.FilterQuery, ch chan<- types.Log) (matrix.Subscription, error)
	PendingBalanceAt(ctx context.Context, account common.Address) (*big.Int, error)
	PendingStorageAt(ctx context.Context, account common.Address, key common.Hash) ([]byte, error)
	PendingCodeAt(ctx context.Context, account common.Address) ([]byte, error)
	PendingNonceAt(ctx context.Context, account common.Address) (uint64, error)
	PendingTransactionCount(ctx context.Context) (uint, error)
	CallContract(ctx context.Context, msg matrix.CallMsg, blockNumber *big.Int) ([]byte, error)
	PendingCallContract(ctx context.Context, msg matrix.CallMsg) ([]byte, error)
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
	EstimateGas(ctx context.Context, msg matrix.CallMsg) (uint64, error)
	SendTransaction(ctx context.Context, tx *types.Transaction) error
}
