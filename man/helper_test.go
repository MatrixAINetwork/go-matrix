//// Copyright (c) 2018 The MATRIX Authors
//// Distributed under the MIT software license, see the accompanying
//// file COPYING or http://www.opensource.org/licenses/mit-license.php
//
//// This file contains some shares testing functionality, common to  multiple
//// different files and modules being tested.
//
package man

//
//import (
//	"crypto/ecdsa"
//	"crypto/rand"
//	"math/big"
//	"sort"
//	"sync"
//	"testing"
//
//	"github.com/MatrixAINetwork/go-matrix/common"
//	"github.com/MatrixAINetwork/go-matrix/consensus/manash"
//	"github.com/MatrixAINetwork/go-matrix/core"
//	"github.com/MatrixAINetwork/go-matrix/core/types"
//	"github.com/MatrixAINetwork/go-matrix/core/vm"
//	"github.com/MatrixAINetwork/go-matrix/crypto"
//	"github.com/MatrixAINetwork/go-matrix/event"
//	"github.com/MatrixAINetwork/go-matrix/man/downloader"
//	"github.com/MatrixAINetwork/go-matrix/mandb"
//	"github.com/MatrixAINetwork/go-matrix/p2p"
//	"github.com/MatrixAINetwork/go-matrix/p2p/discover"
//	"github.com/MatrixAINetwork/go-matrix/params"
//)
//
//var (
//	testBankKey, _ = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
//	testBank       = crypto.PubkeyToAddress(testBankKey.PublicKey)
//)
//
//// newTestProtocolManager creates a new protocol manager for testing purposes,
//// with the given number of blocks already known, and potential notification
//// channels for different events.
//func newTestProtocolManager(mode downloader.SyncMode, blocks int, generator func(int, *core.BlockGen), newtx chan<- []*types.Transaction) (*ProtocolManager, *mandb.MemDatabase, error) {
//	var (
//		evmux  = new(event.TypeMux)
//		engine = manash.NewFaker()
//		db     = mandb.NewMemDatabase()
//		gspec  = &core.Genesis{
//			Config: params.TestChainConfig,
//			Alloc:  core.GenesisAlloc{testBank: {Balance: big.NewInt(1000000)}},
//		}
//		genesis       = gspec.MustCommit(db)
//		blockchain, _ = core.NewBlockChain(db, nil, gspec.Config, engine, vm.Config{})
//	)
//	chain, _ := core.GenerateChain(gspec.Config, genesis, manash.NewFaker(), db, blocks, generator)
//	if _, err := blockchain.InsertChain(chain); err != nil {
//		panic(err)
//	}
//
//	pm, err := NewProtocolManager(gspec.Config, mode, DefaultConfig.NetworkId, evmux, &testTxPool{added: newtx}, engine, blockchain, db)
//	if err != nil {
//		return nil, nil, err
//	}
//	pm.Start(1000)
//	return pm, db, nil
//}
//
//// newTestProtocolManagerMust creates a new protocol manager for testing purposes,
//// with the given number of blocks already known, and potential notification
//// channels for different events. In case of an error, the constructor force-
//// fails the test.
//func newTestProtocolManagerMust(t *testing.T, mode downloader.SyncMode, blocks int, generator func(int, *core.BlockGen), newtx chan<- []*types.Transaction) (*ProtocolManager, *mandb.MemDatabase) {
//	pm, db, err := newTestProtocolManager(mode, blocks, generator, newtx)
//	if err != nil {
//		t.Fatalf("Failed to create protocol manager: %v", err)
//	}
//	return pm, db
//}
//
//// testTxPool is a fake, helper transaction pool for testing purposes
//type testTxPool struct {
//	txFeed event.Feed
//	pool   []*types.Transaction        // Collection of all transactions
//	added  chan<- []*types.Transaction // Notification channel for new transactions
//
//	lock sync.RWMutex // Protects the transaction pool
//}
//
//// AddRemotes appends a batch of transactions to the pool, and notifies any
//// listeners if the addition channel is non nil
//func (p *testTxPool) AddRemotes(txs []*types.Transaction) []error {
//	p.lock.Lock()
//	defer p.lock.Unlock()
//
//	p.pool = append(p.pool, txs...)
//	if p.added != nil {
//		p.added <- txs
//	}
//	return make([]error, len(txs))
//}
//
//// Pending returns all the transactions known to the pool
//func (p *testTxPool) Pending() (map[common.Address]types.Transactions, error) {
//	p.lock.RLock()
//	defer p.lock.RUnlock()
//
//	batches := make(map[common.Address]types.Transactions)
//	for _, tx := range p.pool {
//		from, _ := types.Sender(types.HomesteadSigner{}, tx)
//		batches[from] = append(batches[from], tx)
//	}
//	for _, batch := range batches {
//		sort.Sort(types.TxByNonce(batch))
//	}
//	return batches, nil
//}
//
//func (p *testTxPool) SubscribeNewTxsEvent(ch chan<- core.NewTxsEvent) event.Subscription {
//	return p.txFeed.Subscribe(ch)
//}
//
//// newTestTransaction create a new dummy transaction.
//func newTestTransaction(from *ecdsa.PrivateKey, nonce uint64, datasize int) *types.Transaction {
//	tx := types.NewTransaction(nonce, common.Address{}, big.NewInt(0), 100000, big.NewInt(0), make([]byte, datasize))
//	tx, _ = types.SignTx(tx, types.HomesteadSigner{}, from)
//	return tx
//}
//
//// testPeer is a simulated peer to allow testing direct network calls.
//type testPeer struct {
//	net p2p.MsgReadWriter // Network layer reader/writer to simulate remote messaging
//	app *p2p.MsgPipeRW    // Application layer reader/writer to simulate the local side
//	*peer
//}
//
//// newTestPeer creates a new peer registered at the given protocol manager.
//func newTestPeer(name string, version int, pm *ProtocolManager, shake bool) (*testPeer, <-chan error) {
//	// Create a message pipe to communicate through
//	app, net := p2p.MsgPipe()
//
//	// Generate a random id and create the peer
//	var id discover.NodeID
//	rand.Read(id[:])
//
//	peer := pm.newPeer(version, p2p.NewPeer(id, name, nil), net)
//
//	// Start the peer on a new thread
//	errc := make(chan error, 1)
//	go func() {
//		select {
//		case pm.newPeerCh <- peer:
//			errc <- pm.handle(peer)
//		case <-pm.quitSync:
//			errc <- p2p.DiscQuitting
//		}
//	}()
//	tp := &testPeer{app: app, net: net, peer: peer}
//	// Execute any implicitly requested handshakes and return
//	if shake {
//		var (
//			genesis = pm.blockchain.Genesis()
//			head    = pm.blockchain.CurrentHeader()
//			td      = pm.blockchain.GetTd(head.Hash(), head.Number.Uint64())
//		)
//		tp.handshake(nil, td, head.Hash(), genesis.Hash())
//	}
//	return tp, errc
//}
//
//// handshake simulates a trivial handshake that expects the same state from the
//// remote side as we are simulating locally.
//func (p *testPeer) handshake(t *testing.T, td *big.Int, head common.Hash, genesis common.Hash) {
//	msg := &statusData{
//		ProtocolVersion: uint32(p.version),
//		NetworkId:       DefaultConfig.NetworkId,
//		TD:              td,
//		CurrentBlock:    head,
//		GenesisBlock:    genesis,
//	}
//	if err := p2p.ExpectMsg(p.app, StatusMsg, msg); err != nil {
//		t.Fatalf("status recv: %v", err)
//	}
//	if err := p2p.Send(p.app, StatusMsg, msg); err != nil {
//		t.Fatalf("status send: %v", err)
//	}
//}
//
//// close terminates the local side of the peer, notifying the remote protocol
//// manager of termination.
//func (p *testPeer) close() {
//	p.app.Close()
//}
