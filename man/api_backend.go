// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php


package man

import (
	"context"
	"math/big"
	"time"

	"github.com/matrix/go-matrix/accounts"
	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/common/math"
	"github.com/matrix/go-matrix/core"
	"github.com/matrix/go-matrix/core/bloombits"
	"github.com/matrix/go-matrix/core/rawdb"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/core/vm"
	"github.com/matrix/go-matrix/man/downloader"
	"github.com/matrix/go-matrix/man/gasprice"
	"github.com/matrix/go-matrix/mandb"
	"github.com/matrix/go-matrix/event"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/params"
	"github.com/matrix/go-matrix/rpc"
)

// ManAPIBackend implements manapi.Backend for full nodes
type ManAPIBackend struct {
	man *Matrix
	gpo *gasprice.Oracle
}

func (b *ManAPIBackend) ChainConfig() *params.ChainConfig {
	return b.man.chainConfig
}

func (b *ManAPIBackend) CurrentBlock() *types.Block {
	return b.man.blockchain.CurrentBlock()
}

func (b *ManAPIBackend) SetHead(number uint64) {
	b.man.protocolManager.downloader.Cancel()
	b.man.blockchain.SetHead(number)
}

func (b *ManAPIBackend) HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Header, error) {
	// Pending block is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block := b.man.miner.PendingBlock()
		return block.Header(), nil
	}
	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.man.blockchain.CurrentBlock().Header(), nil
	}
	return b.man.blockchain.GetHeaderByNumber(uint64(blockNr)), nil
}

func (b *ManAPIBackend) BlockByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Block, error) {
	// Pending block is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block := b.man.miner.PendingBlock()
		return block, nil
	}
	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.man.blockchain.CurrentBlock(), nil
	}
	return b.man.blockchain.GetBlockByNumber(uint64(blockNr)), nil
}

func (b *ManAPIBackend) StateAndHeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*state.StateDB, *types.Header, error) {
	// Pending state is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block, state := b.man.miner.Pending()
		return state, block.Header(), nil
	}
	// Otherwise resolve the block number and return its state
	header, err := b.HeaderByNumber(ctx, blockNr)
	if header == nil || err != nil {
		return nil, nil, err
	}
	stateDb, err := b.man.BlockChain().StateAt(header.Root)
	return stateDb, header, err
}

func (b *ManAPIBackend) GetBlock(ctx context.Context, hash common.Hash) (*types.Block, error) {
	return b.man.blockchain.GetBlockByHash(hash), nil
}

func (b *ManAPIBackend) GetReceipts(ctx context.Context, hash common.Hash) (types.Receipts, error) {
	if number := rawdb.ReadHeaderNumber(b.man.chainDb, hash); number != nil {
		return rawdb.ReadReceipts(b.man.chainDb, hash, *number), nil
	}
	return nil, nil
}

func (b *ManAPIBackend) GetLogs(ctx context.Context, hash common.Hash) ([][]*types.Log, error) {
	number := rawdb.ReadHeaderNumber(b.man.chainDb, hash)
	if number == nil {
		return nil, nil
	}
	receipts := rawdb.ReadReceipts(b.man.chainDb, hash, *number)
	if receipts == nil {
		return nil, nil
	}
	logs := make([][]*types.Log, len(receipts))
	for i, receipt := range receipts {
		logs[i] = receipt.Logs
	}
	return logs, nil
}

func (b *ManAPIBackend) GetTd(blockHash common.Hash) *big.Int {
	return b.man.blockchain.GetTdByHash(blockHash)
}

func (b *ManAPIBackend) GetEVM(ctx context.Context, msg core.Message, state *state.StateDB, header *types.Header, vmCfg vm.Config) (*vm.EVM, func() error, error) {
	state.SetBalance(msg.From(), math.MaxBig256)
	vmError := func() error { return nil }

	context := core.NewEVMContext(msg, header, b.man.BlockChain(), nil)
	return vm.NewEVM(context, state, b.man.chainConfig, vmCfg), vmError, nil
}

func (b *ManAPIBackend) SubscribeRemovedLogsEvent(ch chan<- core.RemovedLogsEvent) event.Subscription {
	return b.man.BlockChain().SubscribeRemovedLogsEvent(ch)
}

func (b *ManAPIBackend) SubscribeChainEvent(ch chan<- core.ChainEvent) event.Subscription {
	return b.man.BlockChain().SubscribeChainEvent(ch)
}

func (b *ManAPIBackend) SubscribeChainHeadEvent(ch chan<- core.ChainHeadEvent) event.Subscription {
	return b.man.BlockChain().SubscribeChainHeadEvent(ch)
}

func (b *ManAPIBackend) SubscribeChainSideEvent(ch chan<- core.ChainSideEvent) event.Subscription {
	return b.man.BlockChain().SubscribeChainSideEvent(ch)
}

func (b *ManAPIBackend) SubscribeLogsEvent(ch chan<- []*types.Log) event.Subscription {
	return b.man.BlockChain().SubscribeLogsEvent(ch)
}

func (b *ManAPIBackend) SendTx(ctx context.Context, signedTx *types.Transaction) error {
	return b.man.txPool.AddLocal(signedTx)
}

func (b *ManAPIBackend) GetPoolTransactions() (types.Transactions, error) {
	pending, err := b.man.txPool.Pending()
	if err != nil {
		return nil, err
	}
	var txs types.Transactions
	for _, batch := range pending {
		txs = append(txs, batch...)
	}
	return txs, nil
}

func (b *ManAPIBackend) GetPoolTransaction(hash common.Hash) *types.Transaction {
	return b.man.txPool.Get(hash)
}

func (b *ManAPIBackend) GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error) {
	return b.man.txPool.State().GetNonce(addr), nil
}

func (b *ManAPIBackend) Stats() (pending int, queued int) {
	return b.man.txPool.Stats()
}

func (b *ManAPIBackend) TxPoolContent() (map[common.Address]types.Transactions, map[common.Address]types.Transactions) {
	return b.man.TxPool().Content()
}

func (b *ManAPIBackend) SubscribeNewTxsEvent(ch chan<- core.NewTxsEvent) event.Subscription {
	return b.man.TxPool().SubscribeNewTxsEvent(ch)
}

func (b *ManAPIBackend) Downloader() *downloader.Downloader {
	return b.man.Downloader()
}

func (b *ManAPIBackend) ProtocolVersion() int {
	return b.man.ManVersion()
}

func (b *ManAPIBackend) SuggestPrice(ctx context.Context) (*big.Int, error) {
	return b.gpo.SuggestPrice(ctx)
}

func (b *ManAPIBackend) ChainDb() mandb.Database {
	return b.man.ChainDb()
}

func (b *ManAPIBackend) EventMux() *event.TypeMux {
	return b.man.EventMux()
}

func (b *ManAPIBackend) AccountManager() *accounts.Manager {
	return b.man.AccountManager()
}

func (b *ManAPIBackend) BloomStatus() (uint64, uint64) {
	sections, _, _ := b.man.bloomIndexer.Sections()
	return params.BloomBitsBlocks, sections
}

func (b *ManAPIBackend) ServiceFilter(ctx context.Context, session *bloombits.MatcherSession) {
	for i := 0; i < bloomFilterThreads; i++ {
		go session.Multiplex(bloomRetrievalBatch, bloomRetrievalWait, b.man.bloomRequests)
	}
}

//YY
func (b *ManAPIBackend) SignTx(signedTx *types.Transaction, chainID *big.Int) (*types.Transaction, error) {
	return b.man.signHelper.SignTx(signedTx, chainID)
}

//YY
func (b *ManAPIBackend) SendBroadTx(ctx context.Context, signedTx *types.Transaction, bType bool) error {
	return b.man.txPool.AddBroadTx(signedTx, bType)
}

//YY
func (b *ManAPIBackend) FetcherNotify(hash common.Hash, number uint64) {
	ids := ca.GetRolesByGroup(common.RoleValidator)
	log.Info("==========YY===========", "FetcherNotify()��Validator`s count", len(ids))
	for _, id := range ids {
		peer := b.man.protocolManager.Peers.Peer(id.String())
		log.Info("==========YY===========", "FetcherNotify()��Validator`s NodeID", id)
		log.Info("==========YY===========", "FetcherNotify()��get PeerID by Validator ID", peer.id)
		b.man.protocolManager.fetcher.Notify(id.String(), hash, number, time.Now(), peer.RequestOneHeader, peer.RequestBodies)
		log.Info("==========YY===========", "FetcherNotify()��send Notify completed", 111111111111111)
	}
}
