// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php


package core

import (
	"container/list"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/event"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/metrics"
	"github.com/matrix/go-matrix/p2p/discover"
	"github.com/matrix/go-matrix/params"
	"github.com/syndtr/goleveldb/leveldb"
	"gopkg.in/karalabe/cookiejar.v2/collections/prque"
)

//消息类型
const (
	tmpEmpty = iota //YY
	SendFloodSN
	GetTxbyN
	RecvTxbyN //YY
	RecvErrTx //YY
	BroadCast //YY
	GetConsensusTxbyN
	RecvConsensusTxbyN
)

//YY
const (
	// chainHeadChanSize is the size of channel listening to ChainHeadEvent.
	chainHeadChanSize = 10
)

var (
	// ErrInvalidSender is returned if the transaction contains an invalid signature.
	ErrInvalidSender = errors.New("invalid sender")

	// ErrNonceTooLow is returned if the nonce of a transaction is lower than the
	// one present in the local chain.
	ErrNonceTooLow = errors.New("nonce too low")

	// ErrUnderpriced is returned if a transaction's gas price is below the minimum
	// configured for the transaction pool.
	ErrUnderpriced = errors.New("transaction underpriced")
	//YY 重复交易错误
	ErrKownTransaction = errors.New("known transaction")
	// ErrReplaceUnderpriced is returned if a transaction is attempted to be replaced
	// with a different one without the required price bump.
	ErrReplaceUnderpriced = errors.New("replacement transaction underpriced")

	// ErrInsufficientFunds is returned if the total cost of executing a transaction
	// is higher than the balance of the user's account.
	ErrInsufficientFunds = errors.New("insufficient funds for gas * price + value")

	// ErrIntrinsicGas is returned if the transaction is specified to use less gas
	// than required to start the invocation.
	ErrIntrinsicGas = errors.New("intrinsic gas too low")

	// ErrGasLimit is returned if a transaction's requested gas limit exceeds the
	// maximum allowance of the current block.
	ErrGasLimit = errors.New("exceeds block gas limit")

	// ErrNegativeValue is a sanity error to ensure noone is able to specify a
	// transaction with a negative value.
	ErrNegativeValue = errors.New("negative value")

	// ErrOversizedData is returned if the input data of a transaction is greater
	// than some meaningful limit a user might use. This is not a consensus error
	// making the transaction invalid, rather a DOS protection.
	ErrOversizedData = errors.New("oversized data")
	//YY
	ErrTXCountOverflow = errors.New("Transaction quantity spillover")
	ErrTxToRepeat      = errors.New("Contains duplicate transfer accounts")
	ErrTXWrongful      = errors.New("transaction is unlawful")
)

var (
	evictionInterval    = time.Minute     // Time interval to check for evictable transactions
	statsReportInterval = 8 * time.Second // Time interval to report transaction pool stats
)

var (
	// Metrics for the pending pool
	pendingDiscardCounter   = metrics.NewRegisteredCounter("txpool/pending/discard", nil)
	pendingReplaceCounter   = metrics.NewRegisteredCounter("txpool/pending/replace", nil)
	pendingRateLimitCounter = metrics.NewRegisteredCounter("txpool/pending/ratelimit", nil) // Dropped due to rate limiting
	pendingNofundsCounter   = metrics.NewRegisteredCounter("txpool/pending/nofunds", nil)   // Dropped due to out-of-funds

	// Metrics for the queued pool
	queuedDiscardCounter   = metrics.NewRegisteredCounter("txpool/queued/discard", nil)
	queuedReplaceCounter   = metrics.NewRegisteredCounter("txpool/queued/replace", nil)
	queuedRateLimitCounter = metrics.NewRegisteredCounter("txpool/queued/ratelimit", nil) // Dropped due to rate limiting
	queuedNofundsCounter   = metrics.NewRegisteredCounter("txpool/queued/nofunds", nil)   // Dropped due to out-of-funds

	// General tx metrics
	invalidTxCounter     = metrics.NewRegisteredCounter("txpool/invalid", nil)
	underpricedTxCounter = metrics.NewRegisteredCounter("txpool/underpriced", nil)
)

// TxStatus is the current status of a transaction as seen by the pool.
type TxStatus uint

var mapNs sync.Map  //YY
var mapCaclErrtxs = make(map[common.Hash][]common.Address) //YY  用来统计错误的交易
var mapDelErrtxs = make(map[common.Hash]*big.Int)          //YY  用来删除mapErrorTxs
var mapErrorTxs = make(map[*big.Int]*types.Transaction)    //YY  存放所有的错误交易（20个区块自动删除）
var mapTxsTiming = make(map[common.Hash]uint64)            //YY  需要做定时删除的交易
//YY
type RetChan struct {
	Rxs   types.Transactions
	Err   error
	Resqe int
}

// hezi
type NetworkMsgData struct {
	NodeId discover.NodeID
	Data   []*MsgStruct
}

//hezi
type SNStruct struct {
	Tx_S *big.Int
	Tx_N uint32
}

// hezi
type MsgStruct struct {
	Msgtype uint32
	NodeId  discover.NodeID
	MsgData []byte
}

var num uint32
var ldb *leveldb.DB
var whitemap = make(map[common.Address]bool)

//test====================================
var sendtxs = make([]*types.Transaction, 0) //for test
var sendtxsch = make(chan *types.Transaction, 1)

const (
	TxStatusUnknown TxStatus = iota
	TxStatusQueued
	TxStatusPending
	TxStatusIncluded
)

// blockChain provides the state of blockchain and current gas limit to do
// some pre checks in tx pool and event subscribers.
type blockChain interface {
	CurrentBlock() *types.Block
	GetBlock(hash common.Hash, number uint64) *types.Block
	StateAt(root common.Hash) (*state.StateDB, error)

	SubscribeChainHeadEvent(ch chan<- ChainHeadEvent) event.Subscription
}

// TxPoolConfig are the configuration parameters of the transaction pool.
type TxPoolConfig struct {
	NoLocals  bool          // Whether local transaction handling should be disabled
	Journal   string        // Journal of local transactions to survive node restarts
	Rejournal time.Duration // Time interval to regenerate the local transaction journal

	PriceLimit uint64 // Minimum gas price to enforce for acceptance into the pool
	PriceBump  uint64 // Minimum price bump percentage to replace an already existing transaction (nonce)

	AccountSlots uint64 // Minimum number of executable transaction slots guaranteed per account
	GlobalSlots  uint64 // Maximum number of executable transaction slots for all accounts
	AccountQueue uint64 // Maximum number of non-executable transaction slots permitted per account
	GlobalQueue  uint64 // Maximum number of non-executable transaction slots for all accounts

	Lifetime time.Duration // Maximum amount of time non-executable transaction are queued
}

// DefaultTxPoolConfig contains the default configurations for the transaction
// pool.
var DefaultTxPoolConfig = TxPoolConfig{
	Journal:   "transactions.rlp",
	Rejournal: time.Hour,

	PriceLimit: 18000000000, //YY 2018-08-29 由1改为此值
	PriceBump:  10,

	AccountSlots: 16,
	GlobalSlots:  4096 * 5 * 5 * 10, //YY 2018-08-30 改为乘以5
	AccountQueue: 64 * 1000,
	GlobalQueue:  1024 * 60,

	Lifetime: 3 * time.Hour,
}

// sanitize checks the provided user configurations and changes anything that's
// unreasonable or unworkable.
func (config *TxPoolConfig) sanitize() TxPoolConfig {
	conf := *config
	if conf.Rejournal < time.Second {
		log.Warn("Sanitizing invalid txpool journal time", "provided", conf.Rejournal, "updated", time.Second)
		conf.Rejournal = time.Second
	}
	if conf.PriceLimit < 1 {
		log.Warn("Sanitizing invalid txpool price limit", "provided", conf.PriceLimit, "updated", DefaultTxPoolConfig.PriceLimit)
		conf.PriceLimit = DefaultTxPoolConfig.PriceLimit
	}
	if conf.PriceBump < 1 {
		log.Warn("Sanitizing invalid txpool price bump", "provided", conf.PriceBump, "updated", DefaultTxPoolConfig.PriceBump)
		conf.PriceBump = DefaultTxPoolConfig.PriceBump
	}
	return conf
}

// TxPool contains all currently known transactions. Transactions
// enter the pool when they are received from the network or submitted
// locally. They exit the pool when they are included in the blockchain.
//
// The pool separates processable transactions (which can be applied to the
// current state) and future transactions. Transactions move between those
// two states over time as they are received and processed.
type TxPool struct {
	config       TxPoolConfig
	chainconfig  *params.ChainConfig
	chain        blockChain
	gasPrice     *big.Int
	txFeed       event.Feed
	scope        event.SubscriptionScope
	chainHeadCh  chan ChainHeadEvent
	chainHeadSub event.Subscription
	signer       types.Signer
	mu           sync.RWMutex

	currentState  *state.StateDB      // Current state in the blockchain head
	pendingState  *state.ManagedState // Pending state tracking virtual nonces
	currentMaxGas uint64              // Current gas limit for transaction caps

	locals  *accountSet // Set of local transaction to exempt from eviction rules
	journal *txJournal  // Journal of local transaction to back up to disk

	pending map[common.Address]*txList   // All currently processable transactions
	queue   map[common.Address]*txList   // Queued but non-processable transactions
	beats   map[common.Address]time.Time // Last heartbeat from each known account
	all     *txLookup                    // All transactions to allow lookups
	//=================by hezi==================//
	SContainer map[common.Hash]*types.Transaction
	NContainer map[uint32]*types.Transaction
	Special    map[common.Hash]*types.Transaction // All special transactions
	udptxsCh   chan []*types.Transaction_Mx       //udp交易订阅
	udptxsSub  event.Subscription                 //取消订阅
	//=================================================//
	priced *txPricedList // All transactions sorted by price

	wg sync.WaitGroup // for shutdown sync

	selfmlk sync.RWMutex //YY

	homestead bool
}

// NewTxPool creates a new transaction pool to gather, sort and filter inbound
// transactions from the network.
func NewTxPool(config TxPoolConfig, chainconfig *params.ChainConfig, chain blockChain, path string) *TxPool {
	// Sanitize the input to ensure no vulnerable gas prices are set
	config = (&config).sanitize()

	// Create the transaction pool with its initial settings
	pool := &TxPool{
		config:      config,
		chainconfig: chainconfig,
		chain:       chain,
		signer:      types.NewEIP155Signer(chainconfig.ChainId),
		pending:     make(map[common.Address]*txList),
		queue:       make(map[common.Address]*txList),
		beats:       make(map[common.Address]time.Time),
		SContainer:  make(map[common.Hash]*types.Transaction), //by hezi
		NContainer:  make(map[uint32]*types.Transaction),      //by hezi
		Special:     make(map[common.Hash]*types.Transaction), //by hezi
		udptxsCh:    make(chan []*types.Transaction_Mx, 0),    //hezi
		all:         newTxLookup(),
		chainHeadCh: make(chan ChainHeadEvent, chainHeadChanSize),
		gasPrice:    new(big.Int).SetUint64(config.PriceLimit),
	}
	pool.locals = newAccountSet(pool.signer)
	pool.priced = newTxPricedList(pool.all)
	pool.reset(nil, chain.CurrentBlock().Header())

	//go pool.testList() //for test

	// If local transactions and journaling is enabled, load from disk
	//if !config.NoLocals && config.Journal != "" {
	//	pool.journal = newTxJournal(config.Journal)
	//
	//	if err := pool.journal.load(pool.AddLocals); err != nil {
	//		log.Warn("Failed to load transaction journal", "err", err)
	//	}
	//	if err := pool.journal.rotate(pool.local()); err != nil {
	//		log.Warn("Failed to rotate transaction journal", "err", err)
	//	}
	//}
	// Subscribe events from blockchain
	pool.chainHeadSub = pool.chain.SubscribeChainHeadEvent(pool.chainHeadCh)

	// Start the event loop and return
	pool.wg.Add(1)

	ldb, _ = leveldb.OpenFile(path+"./broadcastdb", nil) //by hezi

	gSendst.lst.list = list.New() //hezi
	gSendst.snlist.slist = make([]*big.Int, 0)
	gSendst.notice = make(chan *big.Int, 1)

	//udp 交易订阅
	pool.udptxsSub, _ = mc.SubscribeEvent(mc.SendUdpTx, pool.udptxsCh)

	go pool.loop()
	go pool.checkList() //hezi
	go pool.listenudp()

	return pool
}

// loop is the transaction pool's main event loop, waiting for and reacting to
// outside blockchain events as well as for various reporting and transaction
// eviction events.
func (pool *TxPool) loop() {
	defer pool.wg.Done()

	// Start the stats reporting and transaction eviction tickers
	var prevPending, prevQueued, prevStales int

	report := time.NewTicker(statsReportInterval)
	defer report.Stop()

	evict := time.NewTicker(evictionInterval)
	defer evict.Stop()

	journal := time.NewTicker(pool.config.Rejournal)
	defer journal.Stop()

	// Track the previous head headers for transaction reorgs
	head := pool.chain.CurrentBlock()

	// Keep waiting for and reacting to the various events
	for {
		select {
		// Handle ChainHeadEvent
		case ev := <-pool.chainHeadCh:
			if ev.Block != nil {
				pool.mu.Lock()
				if pool.chainconfig.IsHomestead(ev.Block.Number()) {
					pool.homestead = true
				}
				pool.reset(head.Header(), ev.Block.Header())
				head = ev.Block
				pool.blockTiming() //YY
				//将广播区块写入db  hezi
				if head.Number().Uint64()%common.GetBroadcastInterval() == 0 {
					pubmap := make(map[common.Address][]byte)
					primap := make(map[common.Address][]byte)
					Heartmap := make(map[common.Address][]byte)
					CallNamemap := make(map[common.Address][]byte)

					var hash_key common.Hash
					log.Info("Block insert message", "height", head.Number().Uint64(), "head.Hash=", head.Hash())
					for _, tx := range head.Transactions() {
						tmpdt := make(map[string][]byte)

						if len(tx.GetMatrix_EX()) > 0 && tx.GetMatrix_EX()[0].TxType == 1 {
							from, _ := types.Sender(pool.signer, tx)
							json.Unmarshal(tx.Data(), &tmpdt)

							for keydata, valdata := range tmpdt {
								if strings.Contains(keydata, mc.Publickey) {
									pubmap[from] = valdata
								} else if strings.Contains(keydata, mc.Privatekey) {
									primap[from] = valdata
								} else if strings.Contains(keydata, mc.Heartbeat) {
									Heartmap[from] = valdata
								} else if strings.Contains(keydata, mc.CallTheRoll) {
									CallNamemap[from] = valdata
								}
							}
						}
					}

					if len(pubmap) > 0 {
						re := head.Number().Uint64() / common.GetBroadcastInterval()
						strVal := fmt.Sprintf("%v", re)
						hash_key = types.RlpHash(mc.Publickey + strVal)
						log.INFO("store publickey success", "height", head.Number().Uint64(), "str", mc.Publickey+strVal, "len", len(pubmap))
						insertdb(hash_key.Bytes(), pubmap)
					} else {
						log.ERROR("without publickey txs", "height", head.Number().Uint64())
					}

					if len(primap) > 0 {
						re := head.Number().Uint64() / common.GetBroadcastInterval()
						strVal := fmt.Sprintf("%v", re)
						hash_key = types.RlpHash(mc.Privatekey + strVal)
						log.INFO("store privatekey success", "height", head.Number().Uint64(), "str", mc.Privatekey+strVal, "len", len(primap))
						insertdb(hash_key.Bytes(), primap)
					} else {
						log.ERROR("without privatekey txs", "height", head.Number().Uint64())
					}

					if len(Heartmap) > 0 {
						re := head.Number().Uint64() / common.GetBroadcastInterval()
						strVal := fmt.Sprintf("%v", re)
						hash_key = types.RlpHash(mc.Heartbeat + strVal)
						log.INFO("store heartbeat success", "height", head.Number().Uint64(), "str", mc.Heartbeat+strVal, "len", len(Heartmap))
						insertdb(hash_key.Bytes(), Heartmap)
					} else {
						log.ERROR("without heartbeat txs", "height", head.Number().Uint64())
					}

					if len(CallNamemap) > 0 {
						re := head.Number().Uint64() / common.GetBroadcastInterval()
						strVal := fmt.Sprintf("%v", re)
						hash_key = types.RlpHash(mc.CallTheRoll + strVal)
						log.INFO("store callName success", "height", head.Number().Uint64(), "str", mc.CallTheRoll+strVal, "len", len(CallNamemap))
						insertdb(hash_key.Bytes(), CallNamemap)
					} else {
						log.ERROR("without callName txs", "height", head.Number().Uint64())
					}

				}
				pool.mu.Unlock()
				//log.INFO("==========", "getPendingTx:befer", 0)
				pool.getPendingTx() //YY
				//log.INFO("==========", "getPendingTx:after", 1)
			}
		// Be unsubscribed due to system stopped
		case <-pool.chainHeadSub.Err():
			return

		// Handle stats reporting ticks
		case <-report.C:
			pool.mu.RLock()
			pending, queued := pool.stats()
			stales := pool.priced.stales
			pool.mu.RUnlock()

			if pending != prevPending || queued != prevQueued || stales != prevStales {
				log.Debug("Transaction pool status report", "executable", pending, "queued", queued, "stales", stales)
				prevPending, prevQueued, prevStales = pending, queued, stales
			}

		// Handle inactive account transaction eviction
		case <-evict.C:
			pool.mu.Lock()
			for addr := range pool.queue {
				// Skip local transactions from the eviction mechanism
				if pool.locals.contains(addr) {
					continue
				}
				// Any non-locals old enough should be removed
				if time.Since(pool.beats[addr]) > pool.config.Lifetime {
					for _, tx := range pool.queue[addr].Flatten() {
						pool.removeTx(tx.Hash(), true)
					}
				}
			}
			pool.mu.Unlock()

		// Handle local transaction journal rotation
		case <-journal.C:
			if pool.journal != nil {
				pool.mu.Lock()
				if err := pool.journal.rotate(pool.local()); err != nil {
					log.Warn("Failed to rotate local tx journal", "err", err)
				}
				pool.mu.Unlock()
			}
		}
	}
}

//sTxmap->tx的编号N是否为nil ;hezi
func (pool *TxPool) sTxValIsNil(s *big.Int) bool {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	if tx, ok := pool.SContainer[common.BigToHash(s)]; ok {
		tmpN := tx.N
		if len(tmpN) == 0 {
			return true
		}
		return false
	}

	return true
}

//给tx设置num ;hezi
func (pool *TxPool) setTxNum(tx *types.Transaction, num uint32) {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	tx.N = append(tx.N, num)
}

//设置map[s]tx ;hezi
func (pool *TxPool) setsTx(s *big.Int, tx *types.Transaction) {
	pool.SContainer[common.BigToHash(s)] = tx
}

//根据s获取tx ;hezi
func (pool *TxPool) getTxbyS(s *big.Int) (tx *types.Transaction) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	tx = pool.SContainer[common.BigToHash(s)]
	return tx
}

//设置map[n]tx ;hezi
func (pool *TxPool) setnTx(num uint32, tx *types.Transaction) {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	pool.NContainer[num] = tx

}

//根据N获取tx ;hezi
func (pool *TxPool) getTxbyN(num uint32) (tx *types.Transaction) {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	tx = pool.NContainer[num]

	return tx
}

//hezi
func (pool *TxPool) deletnTx(num uint32) {
	delete(pool.NContainer, num)
}

//hezi
func (pool *TxPool) deletsTx(s *big.Int) {
	delete(pool.SContainer, common.BigToHash(s))
}

//by hezi
func (pool *TxPool) testList() {

	//=============for test hezi=======================//
	timesendtxs := time.NewTicker(2 * time.Second) //for test hezi
	defer timesendtxs.Stop()
	//=============test=======================//
	for {
		select {
		//===================for test hezi==================//
		case <-timesendtxs.C:
			if len(sendtxs) > 0 {
				log.Info("=====hezi===:", "checklist: sendtxs num=", len(sendtxs))
				pool.txFeed.Send(NewTxsEvent{sendtxs})
				sendtxs = make([]*types.Transaction, 0)
			}

		case tx := <-sendtxsch:
			sendtxs = append(sendtxs, tx)
			//===================for test hezi==================//
			//default:
		}
	}
}

func (pool *TxPool) listenudp() {
	defer pool.udptxsSub.Unsubscribe() //udp交易取消订阅

	for {
		select {
		//udp接收的交易，此处应该只发给验证节点
		case evtxs := <-pool.udptxsCh:
			log.Info("======hezi=====", "checklist: udptxs:", evtxs)
			selfRole := ca.GetRole()
			if selfRole == common.RoleValidator {
				pool.selfmlk.Lock()
				for _, mxtx := range evtxs {
					tx := types.ConvMxtotx(mxtx)
					//YY ====begin======
					if nc := tx.Nonce(); nc < params.NonceAddOne {
						nc = nc | params.NonceAddOne
						tx.SetNonce(nc)
					}
					//=========end======
					//log.Info("====hezi===1","tx:",tx,"udptx add pool time:",time.Now().UnixNano())
					//udpCount++
					tmptx := pool.getTxbyS(tx.GetTxS())
					if tmptx == nil {
						err := pool.addTx(tx, false)
						if err != nil {
							log.Info("======error tx", "err:", err)
						}
					}
					//log.Info("====hezi===2","udptx add pool time:",time.Now().UnixNano())
				}
				pool.selfmlk.Unlock()
			}
			//log.Info("**********************","udp",0)
		case <-pool.udptxsSub.Err():
			return
		}
	}
}

// lockedReset is a wrapper around reset to allow calling it in a thread safe
// manner. This method is only ever used in the tester!
func (pool *TxPool) lockedReset(oldHead, newHead *types.Header) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	pool.reset(oldHead, newHead)
}

// reset retrieves the current state of the blockchain and ensures the content
// of the transaction pool is valid with regard to the chain state.
func (pool *TxPool) reset(oldHead, newHead *types.Header) {
	// If we're reorging an old state, reinject all dropped transactions
	var reinject types.Transactions

	if oldHead != nil && oldHead.Hash() != newHead.ParentHash {
		// If the reorg is too deep, avoid doing it (will happen during fast sync)
		oldNum := oldHead.Number.Uint64()
		newNum := newHead.Number.Uint64()

		if depth := uint64(math.Abs(float64(oldNum) - float64(newNum))); depth > 64 {
			log.Debug("Skipping deep transaction reorg", "depth", depth)
		} else {
			// Reorg seems shallow enough to pull in all transactions into memory
			var discarded, included types.Transactions

			var (
				rem = pool.chain.GetBlock(oldHead.Hash(), oldHead.Number.Uint64())
				add = pool.chain.GetBlock(newHead.Hash(), newHead.Number.Uint64())
			)
			for rem.NumberU64() > add.NumberU64() {
				discarded = append(discarded, rem.Transactions()...)
				if rem = pool.chain.GetBlock(rem.ParentHash(), rem.NumberU64()-1); rem == nil {
					log.Error("Unrooted old chain seen by tx pool", "block", oldHead.Number, "hash", oldHead.Hash())
					return
				}
			}
			for add.NumberU64() > rem.NumberU64() {
				included = append(included, add.Transactions()...)
				if add = pool.chain.GetBlock(add.ParentHash(), add.NumberU64()-1); add == nil {
					log.Error("Unrooted new chain seen by tx pool", "block", newHead.Number, "hash", newHead.Hash())
					return
				}
			}
			for rem.Hash() != add.Hash() {
				discarded = append(discarded, rem.Transactions()...)
				if rem = pool.chain.GetBlock(rem.ParentHash(), rem.NumberU64()-1); rem == nil {
					log.Error("Unrooted old chain seen by tx pool", "block", oldHead.Number, "hash", oldHead.Hash())
					return
				}
				included = append(included, add.Transactions()...)
				if add = pool.chain.GetBlock(add.ParentHash(), add.NumberU64()-1); add == nil {
					log.Error("Unrooted new chain seen by tx pool", "block", newHead.Number, "hash", newHead.Hash())
					return
				}
			}
			reinject = types.TxDifference(discarded, included)
		}
	}
	// Initialize the internal state to the current head
	if newHead == nil {
		newHead = pool.chain.CurrentBlock().Header() // Special case during testing
	}
	statedb, err := pool.chain.StateAt(newHead.Root)
	if err != nil {
		log.Error("Failed to reset txpool state", "err", err)
		return
	}
	pool.currentState = statedb
	pool.pendingState = state.ManageState(statedb)
	pool.currentMaxGas = newHead.GasLimit

	// Inject any transactions discarded due to reorgs
	log.Debug("Reinjecting stale transactions", "count", len(reinject))
	pool.addTxsLocked(reinject, false)

	// validate the pool of pending transactions, this will remove
	// any transactions that have been included in the block or
	// have been invalidated because of another transaction (e.g.
	// higher gas price)
	pool.demoteUnexecutables()

	// Update all accounts to the latest known pending nonce
	for addr, list := range pool.pending {
		txs := list.Flatten() // Heavy but will be cached and is needed by the miner anyway
		pool.pendingState.SetNonce(addr, txs[len(txs)-1].Nonce()+1)
	}
	// Check the queue and move transactions over to the pending if possible
	// or remove those that have become invalid
	pool.promoteExecutables(nil)
}

// Stop terminates the transaction pool.
func (pool *TxPool) Stop() {
	// Unsubscribe all subscriptions registered from txpool
	pool.scope.Close()

	// Unsubscribe subscriptions registered from blockchain
	pool.chainHeadSub.Unsubscribe()
	pool.wg.Wait()

	if pool.journal != nil {
		pool.journal.close()
	}
	log.Info("Transaction pool stopped")
}

// SubscribeNewTxsEvent registers a subscription of NewTxsEvent and
// starts sending event to the given channel.
func (pool *TxPool) SubscribeNewTxsEvent(ch chan<- NewTxsEvent) event.Subscription {
	return pool.scope.Track(pool.txFeed.Subscribe(ch))
}

// GasPrice returns the current gas price enforced by the transaction pool.
func (pool *TxPool) GasPrice() *big.Int {
	pool.mu.RLock()
	defer pool.mu.RUnlock()

	return new(big.Int).Set(pool.gasPrice)
}

// SetGasPrice updates the minimum price required by the transaction pool for a
// new transaction, and drops all transactions below this threshold.
func (pool *TxPool) SetGasPrice(price *big.Int) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	pool.gasPrice = price
	for _, tx := range pool.priced.Cap(price, pool.locals) {
		pool.removeTx(tx.Hash(), false)
	}
	log.Info("Transaction pool price threshold updated", "price", price)
}

// State returns the virtual managed state of the transaction pool.
func (pool *TxPool) State() *state.ManagedState {
	pool.mu.RLock()
	defer pool.mu.RUnlock()

	return pool.pendingState
}

// Stats retrieves the current pool stats, namely the number of pending and the
// number of queued (non-executable) transactions.
func (pool *TxPool) Stats() (int, int) {
	pool.mu.RLock()
	defer pool.mu.RUnlock()

	return pool.stats()
}

// stats retrieves the current pool stats, namely the number of pending and the
// number of queued (non-executable) transactions.
func (pool *TxPool) stats() (int, int) {
	pending := 0
	for _, list := range pool.pending {
		pending += list.Len()
	}
	queued := 0
	for _, list := range pool.queue {
		queued += list.Len()
	}
	return pending, queued
}

// Content retrieves the data content of the transaction pool, returning all the
// pending as well as queued transactions, grouped by account and sorted by nonce.
func (pool *TxPool) Content() (map[common.Address]types.Transactions, map[common.Address]types.Transactions) {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	pending := make(map[common.Address]types.Transactions)
	for addr, list := range pool.pending {
		pending[addr] = list.Flatten()
	}
	queued := make(map[common.Address]types.Transactions)
	for addr, list := range pool.queue {
		queued[addr] = list.Flatten()
	}
	//log.Info("============YY========","Content()::mapNS",len(mapNS))
	log.Info("============YY========", "Content()::SContainer", len(pool.SContainer))
	log.Info("============YY========", "Content()::NContainer", len(pool.NContainer))
	//log.Info("============YY========","Content()::NContainer",pool.NContainer)
	return pending, queued
}

// Pending retrieves all currently processable transactions, groupped by origin
// account and sorted by nonce. The returned transaction set is a copy and can be
// freely modified by calling code.
func (pool *TxPool) Pending() (map[common.Address]types.Transactions, error) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	pending := make(map[common.Address]types.Transactions)
	for addr, list := range pool.pending {
		pending[addr] = list.Flatten()
	}
	return pending, nil
}

//YY 获取pending中剩余的交易（广播区块头后触发）
//区块产生后将Pending中剩余的交易放入区块定时中，如果二十个区块还没有被打包则删除，如果已经被打包了则也删除
func (pool *TxPool) getPendingTx() {
	tmpPending, _ := pool.Pending()
	pool.mu.Lock()
	for _, txs := range tmpPending {
		for _, tx := range txs {
			pool.addBlockTiming(tx.Hash())
		}
	}
	pool.mu.Unlock()
}

//YY 检查当前map中是否存在洪泛过来的交易
func (pool *TxPool) msg_CheckTx(mapSN map[uint32]*big.Int, nid discover.NodeID) {
	log.Info("**************msg_CheckTx IN")
	defer log.Info("**************msg_CheckTx OUT")
	log.Info("========YY===1", "msg_CheckTx", 0)
	//log.Info("========YY===2","msg_CheckTx:len(mapSN)=",len(mapSN))
	listN := make([]uint32, 0)
	//log.Info("==========YY=============","msg_CheckTx()mapSN",mapSN)
	pool.selfmlk.Lock()
	for n, s := range mapSN {
		if s == nil || n == 0 { //如果S或者N 不合法则直接跳过
			log.Info("========YY===:continue", "s", s, "n", n)
			continue
		}
		mapNs.Store(n,s)
		tx := pool.getTxbyS(s)
		//tx := pool.SContainer [s]
		if tx == nil {
			listN = append(listN, n)
		} else {
			isExist := true
			for _, txn := range tx.N { //如果有重复的N就不添加了
				if txn == n {
					isExist = false
					break
				}
			}
			if isExist {
				pool.setTxNum(tx, n)
			}

			pool.setnTx(n, tx)
		}
	}
	pool.selfmlk.Unlock()
	if len(listN) > 0 {
		log.Info("========YY=== ", "listN", len(listN))
		msData, _ := json.Marshal(listN)
		pool.sendMsg(MsgStruct{Msgtype: GetTxbyN, NodeId: nid, MsgData: msData})
	}
}

//YY 接收生成的广播交易
func (pool *TxPool) AddBroadTx(tx *types.Transaction, bType bool) (err error) {

	if bType { //true : 点名交易（广播节点自己产生）
		pool.addTx(tx, true)
	} else { // false : 其他广播交易（非广播节点产生）
		tx_mx := types.GetTransactionMx(tx)
		msData, tmperr := json.Marshal(tx_mx)
		if tmperr == nil {
			nid := ca.GetRolesByGroup(common.RoleBroadcast)
			if len(nid) > 0 {
				pool.sendMsg(MsgStruct{Msgtype: BroadCast, NodeId: nid[0], MsgData: msData})
			} else {
				tmperr = errors.New("ERROR : Send BroadCastTX Fail,BroadCast NodeID is Empty!")
			}
		}
		err = tmperr
	}
	return err
}

//YY 接收到Leader打包的交易共识消息时根据N获取tx (调用本方法需要启动协程)
func (pool *TxPool) ReturnAllTxsByN(listN []uint32, resqe int, addr common.Address, retch chan *RetChan) {
	log.Info("**************ReturnAllTxsByN IN")
	defer log.Info("**************ReturnAllTxsByN OUT")
	log.Info("========YY===1", "ReturnAllTxsByN:len(listN)", len(listN))
	if len(listN) <= 0 {
		retch <- &RetChan{nil, nil, resqe}
		return
	}
	log.Info("========YY===2", "ReturnAllTxsByN", 0)
	txs := make([]*types.Transaction, 0)
	ns := make([]uint32, 0)
	for _, n := range listN {
		tx := pool.getTxbyN(n)
		if tx != nil {
			txs = append(txs, tx)
		} else {
			//当根据N找不到对应的交易时需要跟对方索要一次
			ns = append(ns, n)
		}
	}
	log.Info("========YY===3", "ReturnAllTxsByN:len(ns)", len(ns), "len(txs):", len(txs))
	if len(ns) > 0 {
		log.Info("========YY===4", "ReturnAllTxsByN", 0)
		txs = make([]*types.Transaction, 0)
		nid, err1 := ca.ConvertAddressToNodeId(addr)
		log.Info("leader node", "addr::", addr, "id::", nid.String())
		if err1 != nil {
			log.Info("========YY===5", "ReturnAllTxsByN:discover=err", err1)
			retch <- &RetChan{nil, err1, resqe}
			return
		}
		msData, err2 := json.Marshal(ns)
		if err2 != nil {
			log.Info("========YY===6", "ReturnAllTxsByN:Marshal=err", err2)
			retch <- &RetChan{nil, err2, resqe}
			return
		}
		// 发送缺失交易N的列表
		//pool.sendMsg(MsgStruct{Msgtype: GetTxbyN, NodeId: nid, MsgData: msData})
		pool.sendMsg(MsgStruct{Msgtype: GetConsensusTxbyN, NodeId: nid, MsgData: msData}) //modi hezi(共识要的交易都带s)
		rettime := time.NewTimer(3 * time.Second) // 2秒后没有收到需要的交易则返回
	forBreak:
		for {
			select {
			case <-rettime.C:
				log.Info("========YY===", "ReturnAllTxsByN:Time Out=", 0)
				break forBreak
			case <-time.After(100 * time.Millisecond): //100毫秒轮训一次
				tmpns := make([]uint32, 0)
				for _, n := range ns {
					tx := pool.getTxbyN(n)
					if tx == nil {
						tmpns = append(tmpns, n)
					}
				}
				ns = tmpns
				if len(ns) == 0 {
					log.Info("========YY===", "ReturnAllTxsByN:recvTx Over=", 0)
					break forBreak
				}
			}
		}
		var txerr error
		if len(ns) > 0 {
			txerr = errors.New("loss tx")
		}else{
			for _, n := range listN {
				tx := pool.getTxbyN(n)
				if tx != nil {
					txs = append(txs, tx)
				} else {
			txerr = errors.New("loss tx")
					txs = make([]*types.Transaction, 0)
					break
				}
			}
		}
		retch <- &RetChan{txs, txerr, resqe}
		log.Info("========YY===end if", "ReturnAllTxsByN:len(ns)", len(ns))
	} else {
		retch <- &RetChan{txs, nil, resqe}
		log.Info("========YY===end else", "ReturnAllTxsByN", 0)
	}
}

// (共识要交易)根据N值获取对应的交易(modi hezi)
func (pool *TxPool) msg_GetConsensusTxByN(listN []uint32, nid discover.NodeID) {
	log.Info("==========YY", "msg_GetConsensusTxByN:len(listN)", len(listN))
	if len(listN) <= 0 {
		return
	}
	mapNtx := make(map[uint32]*types.Transaction)
	for _, n := range listN {
		tx := pool.getTxbyN(n) //TODO　循环调用这个方法，然后方法中有锁 这样的情况是否允许
		if tx != nil {
			//ftx := types.GetFloodData(tx)
			mapNtx[n] = tx
		} else {
			log.Info("===========YY==tx is nil")
		}
	}
	msData, _ := json.Marshal(mapNtx)
	log.Info("========YY===2", "GetConsensusTxByN:ntxMap", len(mapNtx),"nodeid",nid.String())
	pool.sendMsg(MsgStruct{Msgtype: RecvConsensusTxbyN, NodeId: nid, MsgData: msData})
	log.Info("========YY===3", "GetConsensusTxByN", 0)
}
//YY 根据N值获取对应的交易(洪泛)
func (pool *TxPool) msg_GetTxByN(listN []uint32, nid discover.NodeID) {
	log.Info("==========YY", "msg_GetTxByN:len(listN)", len(listN))
	if len(listN) <= 0 {
		return
	}
	mapNtx := make(map[uint32]*types.Floodtxdata)
	for _, n := range listN {
		tx := pool.getTxbyN(n) //TODO　循环调用这个方法，然后方法中有锁 这样的情况是否允许
		if tx != nil {
			ftx := types.GetFloodData(tx)
			mapNtx[n] = ftx
		} else {
			log.Info("===========YY==tx is nil")
		}
	}
	msData, _ := json.Marshal(mapNtx)

	log.Info("========YY===2", "msg_GetTxByN:ntxMap", len(mapNtx))
	pool.sendMsg(MsgStruct{Msgtype: RecvTxbyN, NodeId: nid, MsgData: msData})
	log.Info("========YY===3", "msg_GetTxByN", 0)
}

//此接口传的交易带s(modi hezi)
func (pool *TxPool) msg_RecvConsensusFloodTx(mapNtx map[uint32]*types.Transaction, nid discover.NodeID) {
	pool.selfmlk.Lock()
	log.INFO("===========","msg_RecvConsensusFloodTx",len(mapNtx))
	errorTxs := make([]*big.Int, 0)
	for n, tx := range mapNtx {
		mapNs.Store(n,tx.GetTxS())
		ts,ok:=mapNs.Load(n)
		if !ok{
			continue
		}
		s:=ts.(*big.Int)
		if s == nil || n == 0 { //如果S或者N 不合法则直接跳过
			continue
		}
		isExist := true
		for _, txn := range tx.N { //如果有重复的N我就不添加了
			if txn == n {
				isExist = false
				break
			}
		}
		if isExist {
			tmptx := pool.getTxbyS(s)
			if tmptx != nil {
				tx = tmptx
			}
			pool.setTxNum(tx, n)
			pool.setnTx(n, tx)
		}
		//tx.SetTxS(s)
		err := pool.addTx(tx, false)
		if err != nil && err != ErrKownTransaction {
			log.Info("========YY===3", "msg_RecvConsensusFloodTx::Error=", err)
			log.Info("========YY===4", "msg_RecvConsensusFloodTx::Tx=", tx)
			if _, ok := mapErrorTxs[s]; !ok {
				errorTxs = append(errorTxs, s)
				mapErrorTxs[s] = tx
				mapDelErrtxs[tx.Hash()] = s
			}
			//对于添加失败的交易要调用删除map方法
			pool.mu.Lock()
			pool.deleteMap(tx)
			pool.mu.Unlock()
		}
	}
	pool.selfmlk.Unlock()
	if len(errorTxs) > 0 {
		//TODO S在这如何进行签名？？如何获得本节点账户信息
		msData, _ := json.Marshal(errorTxs)
		pool.sendMsg(MsgStruct{Msgtype: RecvErrTx, NodeId: nid, MsgData: msData})
		//pool.msg_RecvErrTx(common.Address{}, errorTxs)
	}
}
//YY 接收洪泛的交易（根据N请求到的交易）
func (pool *TxPool) msg_RecvFloodTx(mapNtx map[uint32]*types.Floodtxdata, nid discover.NodeID) {
	errorTxs := make([]*big.Int, 0)
	pool.selfmlk.Lock()
	log.Info("=======YY===", "msg_RecvFloodTx: len(mapNtx)=", len(mapNtx))
	//aa := 0
	for n, ftx := range mapNtx {
		ts,ok:=mapNs.Load(n)
		if !ok{
			continue
		}
		s:=ts.(*big.Int)
		if s == nil || n == 0 { //如果S或者N 不合法则直接跳过
			continue
		}
		tx := types.SetFloodData(ftx)
		isExist := true
		for _, txn := range tx.N { //如果有重复的N我就不添加了
			if txn == n {
				isExist = false
				break
			}
		}
		if isExist {
			tmptx := pool.getTxbyS(s)
			if tmptx != nil {
				tx = tmptx
			}
			pool.setTxNum(tx, n)
			pool.setnTx(n, tx)
		}
		tx.SetTxS(s)
		err := pool.addTx(tx, false)
		//aa += 1
		//log.Info("========YY===3", "msg_RecvFloodTx::addTx:nid=",nid.String(),"err:", err,"count:",aa)
		if err != nil && err != ErrKownTransaction {
			log.Info("========YY===3", "msg_RecvFloodTx::Error=", err)
			log.Info("========YY===4", "msg_RecvFloodTx::Tx=", tx)
			if _, ok := mapErrorTxs[s]; !ok {
				errorTxs = append(errorTxs, s)
				mapErrorTxs[s] = tx
				mapDelErrtxs[tx.Hash()] = s
			}
			//对于添加失败的交易要调用删除map方法
			pool.mu.Lock()
			pool.deleteMap(tx)
			pool.mu.Unlock()
		}
	}
	pool.selfmlk.Unlock()
	if len(errorTxs) > 0 {
		//TODO S在这如何进行签名？？如何获得本节点账户信息
		msData, _ := json.Marshal(errorTxs)
		pool.sendMsg(MsgStruct{Msgtype: RecvErrTx, NodeId: nid, MsgData: msData})
		//pool.msg_RecvErrTx(common.Address{}, errorTxs)
	}
}

//YY 接收错误交易
//问题：如果leader作弊其交易池中有一笔错误交易，其他验证者没有该笔交易，这时候洪泛给其他节点肯定会跟leader要该笔交易，这时候如果广播错误交易的S给其他节点，其他节点可能都没有这个记录
func (pool *TxPool) msg_RecvErrTx(addr common.Address, listS []*big.Int) {
	/*
		1、传输时需要将交易传输过去，否则其他节点可能会找不到这笔交易
		2、接收到之后先检查本地是否存在，如果存在就不错其他的操作了如果不存在就调用入池的方法，然后如果入池成功就不做操作如果
			入池失败就签名后发送出去。
			洪泛时还需要检查这笔交易是否存在于错误交易列表中，如果存在则不需要再去请求这笔交易
		3、签名时需要私钥 该怎么获取git 		1、针对一个交易，如果5个验证者认为是错误的6个认为是对的，那么在挖矿成功后打包区块时，那5个验证者无法打包生成区块。
		2、如果6个V认为是错误的5个认为是对的，那么对打包生成区块没影响。
	*/
	pool.selfmlk.Lock()
	for _, s := range listS {
		tmptx := pool.getTxbyS(s)
		if tmptx != nil {
			//如果本地有该笔交易则说明本节点认为这笔交易是对的，不做其他操作。
			continue
		}
		if _, ok := mapErrorTxs[s]; !ok {
			//如果本地没有该笔错误交易则等待。因为如果本地没有这笔交易其他节点肯定会给其洪泛的，洪泛后就会收到这笔交易
			continue
		}
		tx := mapErrorTxs[s]
		hash := tx.Hash()
		isRepeat := false
		for _, acc := range mapCaclErrtxs[hash] {
			if addr == acc { //判断当前错误交易，同一个节点是否发送过
				isRepeat = true
				break
			}
		}
		if isRepeat {
			continue
		}
		pool.mu.Lock()
		mapCaclErrtxs[hash] = append(mapCaclErrtxs[hash], addr)
		if uint64(len(mapCaclErrtxs[hash])) >= params.ErrTxConsensus {
			pool.removeTx(hash, true)
		} else {
			pool.addBlockTiming(hash)
		}
		pool.mu.Unlock()
	}
	pool.selfmlk.Unlock()
}

//YY 刪除新增加的map中的数据
func (pool *TxPool) deleteMap(tx *types.Transaction) {
	//log.Info("========YY===1","begin deleteMap",0)
	//在调用的地方已经加锁了所以在此不用加锁
	s := tx.GetTxS()
	listn := tx.N
	for _, n := range listn {
		pool.deletnTx(n)
		mapNs.Delete(n)
	}
	//delete(mapSNForTx,s)
	delete(mapTxsTiming, tx.Hash())
	pool.deletsTx(s)
	//log.Info("========YY===2","end deleteMap")
}

//YY 添加区块定时
func (pool *TxPool) addBlockTiming(hash common.Hash) {
	if _, ok := mapTxsTiming[hash]; ok {
		return
	}
	mapTxsTiming[hash] = pool.chain.CurrentBlock().Number().Uint64()
}

//YY 20个区块定时删除(每次收到新区快头广播时触发)
func (pool *TxPool) blockTiming() {
	//外侧已经有锁在此不用再加锁 TODO mapTxsTiming 等全局变量操作是否需要单独加锁
	blockNum := pool.chain.CurrentBlock().Number()
	listHash := make([]common.Hash, 0)
	for hash, num := range mapTxsTiming {
		if n := new(big.Int).Sub(blockNum, new(big.Int).SetUint64(num)); n.Uint64() >= params.SubBlockNum {
			listHash = append(listHash, hash)
		}
	}
	if len(listHash) > 0 {
		for _, hash := range listHash {
			//delete(mapErrtxsTiming, hash)
			pool.removeTx(hash, true)
			delete(mapCaclErrtxs, hash)
			delete(mapTxsTiming, hash)
			if s, ok := mapDelErrtxs[hash]; ok {
				delete(mapErrorTxs, s)
				//delete(mapSNForTx,s)
				delete(mapDelErrtxs, hash)
			}
		}
	}
}

// local retrieves all currently known local transactions, groupped by origin
// account and sorted by nonce. The returned transaction set is a copy and can be
// freely modified by calling code.
func (pool *TxPool) local() map[common.Address]types.Transactions {
	txs := make(map[common.Address]types.Transactions)
	for addr := range pool.locals.accounts {
		if pending := pool.pending[addr]; pending != nil {
			txs[addr] = append(txs[addr], pending.Flatten()...)
		}
		if queued := pool.queue[addr]; queued != nil {
			txs[addr] = append(txs[addr], queued.Flatten()...)
		}
	}
	return txs
}

// validateTx checks whether a transaction is valid according to the consensus
// rules and adheres to some heuristic limits of the local node (price and size).
func (pool *TxPool) validateTx(tx *types.Transaction, local bool) error {

	//YY add if
	txEx := tx.GetMatrix_EX()
	var txcount uint64
	if len(txEx) > 0 {
		txcount = 1
		if len(txEx[0].ExtraTo) > 0 {
			maptx := map[common.Address]bool{}
			maptx[*tx.To()] = true
			for _, tx_list := range txEx {
				for _, txs := range tx_list.ExtraTo {
					txcount++
					if txs.Amount.Sign() < 0 { //验证每个被转账的金额不能小于0
						return ErrNegativeValue
					}
					if txs.Recipient != nil {
						_, ok := maptx[*txs.Recipient] //判断是否有重复的被转账地址
						if ok {
							return ErrTxToRepeat
						} else {
							maptx[*txs.Recipient] = true
						}
					}

				}
			}
		}
		// Heuristic limit, reject transactions over 32KB to prevent DOS attacks
		var txsize uint64 = 32 * 1024
		if uint64(tx.Size()) > txsize*txcount {
			return ErrOversizedData
		}
		if txcount > params.TxCount { //验证一对多交易最多支持1000条
			return ErrTXCountOverflow
		}
	} else {
		// Heuristic limit, reject transactions over 32KB to prevent DOS attacks
		if tx.Size() > 32*1024 {
			return ErrOversizedData
		}
	}
	// Transactions can't be negative. This may never happen using RLP decoded
	// transactions but may occur if you create a transaction using the RPC.
	if tx.Value().Sign() < 0 {
		return ErrNegativeValue
	}

	// Ensure the transaction doesn't exceed the current block limit gas.
	if pool.currentMaxGas < tx.Gas() {
		return ErrGasLimit
	}

	// Make sure the transaction is signed properly
	from, err := types.Sender(pool.signer, tx)
	if err != nil {
		return ErrInvalidSender
	}
	//YY 验证当V值大于128时，如果扩展交易为空则直接丢弃该交易并返回交易不合法
	if tx.GetTxV().Cmp(big.NewInt(128)) > 0 && len(txEx) <= 0 {
		return ErrTXWrongful
	}
	// Drop non-local transactions under our own minimal accepted gas price
	//local = local || pool.locals.contains(from) // account may be local even if the transaction arrived from the network
	if /*!local && */ pool.gasPrice.Cmp(tx.GasPrice()) > 0 {
		return ErrUnderpriced
	}
	// Ensure the transaction adheres to nonce ordering
	if pool.currentState.GetNonce(from) > tx.Nonce() {
		return ErrNonceTooLow
	}
	//YY add if
	if len(txEx) > 0 && len(txEx[0].ExtraTo) > 0 {
		// Transactor should have enough funds to cover the costs
		// cost == V + GP * GL
		if pool.currentState.GetBalance(from).Cmp(tx.CostALL()) < 0 {
			return ErrInsufficientFunds
		}
	} else {
		if pool.currentState.GetBalance(from).Cmp(tx.Cost()) < 0 {
			return ErrInsufficientFunds
		}
	}
	intrGas, err := IntrinsicGas(tx.Data(), tx.To() == nil, pool.homestead)
	if err != nil {
		return err
	}
	//YY add if
	if len(txEx) > 0 && len(txEx[0].ExtraTo) > 0 {
		for _, tx_list := range txEx {
			for _, txs := range tx_list.ExtraTo {
				tmpintrGas, tmperr := IntrinsicGas(txs.Payload, txs.Recipient == nil, pool.homestead)
				if tmperr != nil {
					return tmperr
				}
				intrGas += tmpintrGas
			}
		}
	}
	if tx.Gas() < intrGas {
		return ErrIntrinsicGas
	}
	return nil
}

// add validates a transaction and inserts it into the non-executable queue for
// later pending promotion and execution. If the transaction is a replacement for
// an already pending or queued one, it overwrites the previous and returns this
// so outer code doesn't uselessly call promote.
//
// If a newly added transaction is marked as local, its sending account will be
// whitelisted, preventing any associated transaction from being dropped out of
// the pool due to pricing constraints.
func (pool *TxPool) add(tx *types.Transaction, local bool) (bool, error) {
	//======================by hezi============================//
	if len(tx.GetMatrix_EX()) != 0 {
		log.Info("GetMatrix_EX has data")
		if tx.GetMatrix_EX()[0].TxType == 1 {
			log.Info("========broadcast tx add pool")
			from, _ := types.Sender(pool.signer, tx)
			tmpdt := make(map[string][]byte)
			err := json.Unmarshal(tx.Data(), &tmpdt)
			if err != nil {
				log.Error("add broadcast tx pool", "json.Unmarshal failed", err)
				log.Info("======hezi===", "json.Unmarshal failed", err)
				return false, err
			}
			log.Info("======hezi===", "broadcast count", len(tmpdt))
			log.Info("======hezi===", "broadcast======", tmpdt)
			for keydata, _ := range tmpdt {
				hash := types.RlpHash(keydata + from.String())
				if pool.Special[hash] != nil {
					log.Info("======hezi===", "known broadcast transaction", hash)
					log.Trace("Discarding already known broadcast transaction", "hash", hash)
					return false, fmt.Errorf("known broadcast transaction: %x", hash)
				}

				//var val big.Int
				//val.Quo(pool.chain.CurrentBlock().Number(), big.NewInt(100)) //z = x/y
				log.Info("======hezi===", "pool.chain.CurrentBlock().Number()========", pool.chain.CurrentBlock().Number())
				//val := new(big.Int).Quo(pool.chain.CurrentBlock().Number(), big.NewInt(10)) //z = x/y
				tval := pool.chain.CurrentBlock().Number().Uint64() / common.GetBroadcastInterval()
				strVal := fmt.Sprintf("%v", (tval + 1))
				log.Info("======hezi===", "strVal()========", strVal)
				//strVal := fmt.Sprintf("%v", val)
				if strings.Contains(keydata, strVal) {
					log.Info("======hezi===", "add specialtxs map", tx)
					pool.Special[hash] = tx
				} else {
					aa := keydata + from.String()
					log.Info("======hezi===", "keydata + from.String()", aa)
					log.Info("======hezi===", "add finish", strVal)
				}
			}
			log.Info("======hezi===", "broadcast end", 0)
			return true, nil
		}

	}

	//log.Info("=====hezi====CommonTx add tx")
	//普通交易
	hash := tx.Hash()
	// If the transaction is already known, discard it
	if pool.all.Get(hash) != nil {
		log.Trace("Discarding already known transaction", "hash", hash)
		return false, ErrKownTransaction //fmt.Errorf("known transaction: %x", hash) // YY
	}

	// If the transaction fails basic validation, discard it
	if err := pool.validateTx(tx, local); err != nil {
		log.Trace("Discarding invalid transaction", "hash", hash, "err", err)
		invalidTxCounter.Inc(1)
		return false, err
	}

	// If the transaction pool is full, discard underpriced transactions
	if uint64(pool.all.Count()) >= pool.config.GlobalSlots+pool.config.GlobalQueue {
		// If the new transaction is underpriced, don't accept it
		if !local && pool.priced.Underpriced(tx, pool.locals) {
			log.Trace("Discarding underpriced transaction", "hash", hash, "price", tx.GasPrice())
			underpricedTxCounter.Inc(1)
			return false, ErrUnderpriced
		}
		// New transaction is better than our worse ones, make room for it
		drop := pool.priced.Discard(pool.all.Count()-int(pool.config.GlobalSlots+pool.config.GlobalQueue-1), pool.locals)
		for _, tx := range drop {
			log.Trace("Discarding freshly underpriced transaction", "hash", tx.Hash(), "price", tx.GasPrice())
			underpricedTxCounter.Inc(1)
			pool.removeTx(tx.Hash(), false)
		}
	}
	// If the transaction is replacing an already pending one, do directly
	from, _ := types.Sender(pool.signer, tx) // already validated
	if list := pool.pending[from]; list != nil && list.Overlaps(tx) {
		// Nonce already pending, check if required price bump is met
		inserted, old := list.Add(tx, pool.config.PriceBump)
		if !inserted {
			pendingDiscardCounter.Inc(1)
			return false, ErrReplaceUnderpriced
		}
		// New transaction is better, replace old one
		if old != nil {
			pool.all.Remove(old.Hash())
			pool.priced.Removed()
			pendingReplaceCounter.Inc(1)
		}
		pool.all.Add(tx)
		pool.priced.Put(tx)
		pool.journalTx(from, tx)

		log.Trace("Pooled new executable transaction", "hash", hash, "from", from, "to", tx.To())

		//by hezi
		selfRole := ca.GetRole()
		if selfRole == common.RoleMiner || selfRole == common.RoleValidator {
			tx_s := tx.GetTxS()
			pool.setsTx(tx_s, tx)
			//log.Info("=====hezi====", "add pending finish", 0)
			if len(tx.N) == 0 {
				gSendst.notice <- tx.GetTxS()
			}
		} else if selfRole == common.RoleDefault {
			// We've directly injected a replacement transaction, notify subsystems
			//log.Info("======hezi====send NewTxsEvent")

			//========================for test hezi=================================//
			//if tx != nil {
			//	sendtxsch <- tx
			//}
			//=====================================================================//
			go pool.txFeed.Send(NewTxsEvent{types.Transactions{tx}})
		}

		return old != nil, nil
	}
	// New transaction isn't replacing a pending one, push into queue
	replace, err := pool.enqueueTx(hash, tx)
	if err != nil {
		return false, err
	}
	// Mark local addresses and journal local transactions
	if local {
		pool.locals.add(from)
	}
	pool.journalTx(from, tx)

	//log.Trace("Pooled new future transaction", "hash", hash, "from", from, "to", tx.To())
	return replace, nil
}

// enqueueTx inserts a new transaction into the non-executable transaction queue.
//
// Note, this method assumes the pool lock is held!
func (pool *TxPool) enqueueTx(hash common.Hash, tx *types.Transaction) (bool, error) {
	// Try to insert the transaction into the future queue
	from, _ := types.Sender(pool.signer, tx) // already validated
	if pool.queue[from] == nil {
		pool.queue[from] = newTxList(false)
	}
	inserted, old := pool.queue[from].Add(tx, pool.config.PriceBump)
	if !inserted {
		// An older transaction was better, discard this
		queuedDiscardCounter.Inc(1)
		return false, ErrReplaceUnderpriced
	}
	// Discard any previous transaction and mark this
	if old != nil {
		pool.all.Remove(old.Hash())
		pool.priced.Removed()
		queuedReplaceCounter.Inc(1)
	}
	if pool.all.Get(hash) == nil {
		pool.all.Add(tx)
		pool.priced.Put(tx)
	}
	return old != nil, nil
}

// journalTx adds the specified transaction to the local disk journal if it is
// deemed to have been sent from a local account.
func (pool *TxPool) journalTx(from common.Address, tx *types.Transaction) {
	// Only journal if it's enabled and the transaction is local
	if pool.journal == nil || !pool.locals.contains(from) {
		return
	}
	if err := pool.journal.insert(tx); err != nil {
		log.Warn("Failed to journal local transaction", "err", err)
	}
}

// promoteTx adds a transaction to the pending (processable) list of transactions
// and returns whether it was inserted or an older was better.
//
// Note, this method assumes the pool lock is held!
func (pool *TxPool) promoteTx(addr common.Address, hash common.Hash, tx *types.Transaction) bool {
	// Try to insert the transaction into the pending queue
	if pool.pending[addr] == nil {
		pool.pending[addr] = newTxList(true)
	}
	list := pool.pending[addr]

	inserted, old := list.Add(tx, pool.config.PriceBump)
	if !inserted {
		// An older transaction was better, discard this
		pool.all.Remove(hash)
		pool.priced.Removed()

		pendingDiscardCounter.Inc(1)
		return false
	}
	// Otherwise discard any previous transaction and mark this
	if old != nil {
		pool.all.Remove(old.Hash())
		pool.priced.Removed()

		pendingReplaceCounter.Inc(1)
	}
	// Failsafe to work around direct pending inserts (tests)
	if pool.all.Get(hash) == nil {
		pool.all.Add(tx)
		pool.priced.Put(tx)
	}
	// Set the potentially new pending nonce and notify any subsystems of the new tx
	pool.beats[addr] = time.Now()
	pool.pendingState.SetNonce(addr, tx.Nonce()+1)

	return true
}

// AddLocal enqueues a single transaction into the pool if it is valid, marking
// the sender as a local one in the mean time, ensuring it goes around the local
// pricing constraints.
func (pool *TxPool) AddLocal(tx *types.Transaction) error {
	return pool.addTx(tx, !pool.config.NoLocals)
}

// AddRemote enqueues a single transaction into the pool if it is valid. If the
// sender is not among the locally tracked ones, full pricing constraints will
// apply.
func (pool *TxPool) AddRemote(tx *types.Transaction) error {
	return pool.addTx(tx, false)
}

// AddLocals enqueues a batch of transactions into the pool if they are valid,
// marking the senders as a local ones in the mean time, ensuring they go around
// the local pricing constraints.
func (pool *TxPool) AddLocals(txs []*types.Transaction) []error {
	return pool.addTxs(txs, !pool.config.NoLocals)
}

// AddRemotes enqueues a batch of transactions into the pool if they are valid.
// If the senders are not among the locally tracked ones, full pricing constraints
// will apply.
func (pool *TxPool) AddRemotes(txs []*types.Transaction) []error {
	return pool.addTxs(txs, false)
}

// addTx enqueues a single transaction into the pool if it is valid.
func (pool *TxPool) addTx(tx *types.Transaction, local bool) error {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	// Try to inject the transaction and update any state
	replace, err := pool.add(tx, local)
	if err != nil {
		return err
	}
	// If we added a new transaction, run promotion checks and return
	if !replace {
		from, _ := types.Sender(pool.signer, tx) // already validated
		pool.promoteExecutables([]common.Address{from})
	}
	return nil
}

// addTxs attempts to queue a batch of transactions if they are valid.
func (pool *TxPool) addTxs(txs []*types.Transaction, local bool) []error {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	return pool.addTxsLocked(txs, local)
}

// addTxsLocked attempts to queue a batch of transactions if they are valid,
// whilst assuming the transaction pool lock is already held.
func (pool *TxPool) addTxsLocked(txs []*types.Transaction, local bool) []error {
	// Add the batch of transaction, tracking the accepted ones
	dirty := make(map[common.Address]struct{})
	errs := make([]error, len(txs))

	for i, tx := range txs {
		var replace bool
		if replace, errs[i] = pool.add(tx, local); errs[i] == nil {
			if !replace {
				from, _ := types.Sender(pool.signer, tx) // already validated
				dirty[from] = struct{}{}
			}
		}
	}
	// Only reprocess the internal state if something was actually added
	if len(dirty) > 0 {
		addrs := make([]common.Address, 0, len(dirty))
		for addr := range dirty {
			addrs = append(addrs, addr)
		}
		pool.promoteExecutables(addrs)
	}
	return errs
}

// Status returns the status (unknown/pending/queued) of a batch of transactions
// identified by their hashes.
func (pool *TxPool) Status(hashes []common.Hash) []TxStatus {
	pool.mu.RLock()
	defer pool.mu.RUnlock()

	status := make([]TxStatus, len(hashes))
	for i, hash := range hashes {
		if tx := pool.all.Get(hash); tx != nil {
			from, _ := types.Sender(pool.signer, tx) // already validated
			if pool.pending[from] != nil && pool.pending[from].txs.items[tx.Nonce()] != nil {
				status[i] = TxStatusPending
			} else {
				status[i] = TxStatusQueued
			}
		}
	}
	return status
}

// Get returns a transaction if it is contained in the pool
// and nil otherwise.
func (pool *TxPool) Get(hash common.Hash) *types.Transaction {
	return pool.all.Get(hash)
}

// removeTx removes a single transaction from the queue, moving all subsequent
// transactions back to the future queue.
func (pool *TxPool) removeTx(hash common.Hash, outofbound bool) {
	// Fetch the transaction we wish to delete
	log.Info("========YY=======1", "removeTx", 0)
	tx := pool.all.Get(hash)
	if tx == nil {
		return
	}
	log.Info("========YY=======2", "removeTx", 0)
	addr, _ := types.Sender(pool.signer, tx) // already validated during insertion

	// Remove it from the list of known transactions
	pool.all.Remove(hash)
	if outofbound {
		pool.priced.Removed()
	}
	//YY ========begin=========
	pool.deleteMap(tx)
	//===========end===========
	// Remove the transaction from the pending lists and reset the account nonce
	if pending := pool.pending[addr]; pending != nil {
		if removed, invalids := pending.Remove(tx); removed {
			// If no more pending transactions are left, remove the list
			if pending.Empty() {
				delete(pool.pending, addr)
				delete(pool.beats, addr)
			}
			// Postpone any invalidated transactions
			for _, tx := range invalids {
				pool.enqueueTx(tx.Hash(), tx)
			}
			// Update the account nonce if needed
			if nonce := tx.Nonce(); pool.pendingState.GetNonce(addr) > nonce {
				pool.pendingState.SetNonce(addr, nonce)
			}
			return
		}
	}
	// Transaction is in the future queue
	if future := pool.queue[addr]; future != nil {
		future.Remove(tx)
		if future.Empty() {
			delete(pool.queue, addr)
		}
	}
}

// promoteExecutables moves transactions that have become processable from the
// future queue to the set of pending transactions. During this process, all
// invalidated transactions (low nonce, low balance) are deleted.
func (pool *TxPool) promoteExecutables(accounts []common.Address) {
	// Track the promoted transactions to broadcast them at once
	var promoted []*types.Transaction

	// Gather all the accounts potentially needing updates
	if accounts == nil {
		accounts = make([]common.Address, 0, len(pool.queue))
		for addr := range pool.queue {
			accounts = append(accounts, addr)
		}
	}
	// Iterate over all accounts and promote any executable transactions
	for _, addr := range accounts {
		list := pool.queue[addr]
		if list == nil {
			continue // Just in case someone calls with a non existing account
		}
		// Drop all transactions that are deemed too old (low nonce)
		for _, tx := range list.Forward(pool.currentState.GetNonce(addr)) {
			hash := tx.Hash()
			log.Trace("Removed old queued transaction", "hash", hash)
			pool.all.Remove(hash)
			pool.priced.Removed()
		}
		// Drop all transactions that are too costly (low balance or out of gas)
		drops, _ := list.Filter(pool.currentState.GetBalance(addr), pool.currentMaxGas)
		for _, tx := range drops {
			hash := tx.Hash()
			log.Trace("Removed unpayable queued transaction", "hash", hash)
			pool.all.Remove(hash)
			pool.priced.Removed()
			queuedNofundsCounter.Inc(1)
		}
		// Gather all executable transactions and promote them
		for _, tx := range list.Ready(pool.pendingState.GetNonce(addr)) {
			hash := tx.Hash()
			if pool.promoteTx(addr, hash, tx) {
				//log.Trace("Promoting queued transaction", "hash", hash)
				promoted = append(promoted, tx)
			}
		}
		// Drop all transactions over the allowed limit
		if !pool.locals.contains(addr) {
			for _, tx := range list.Cap(int(pool.config.AccountQueue)) {
				hash := tx.Hash()
				pool.all.Remove(hash)
				pool.priced.Removed()
				queuedRateLimitCounter.Inc(1)
				log.Trace("Removed cap-exceeding queued transaction", "hash", hash)
			}
		}
		// Delete the entire queue entry if it became empty.
		if list.Empty() {
			delete(pool.queue, addr)
		}
	}
	// Notify subsystem for new promoted transactions.
	if len(promoted) > 0 {
		//hezi
		selfRole := ca.GetRole()
		if selfRole == common.RoleMiner || selfRole == common.RoleValidator {
			for _, tx := range promoted {
				tx_s := tx.GetTxS()
				pool.setsTx(tx_s, tx)
				//pool.setsTx(tx.GetTxS(), tx)
				//log.Info("=====hezi====", "promoted:Nlist", tx.N)
				if len(tx.N) == 0 {
					gSendst.notice <- tx.GetTxS()
				}
			}
		} else if selfRole == common.RoleDefault {
			//log.Info("======hezi====send NewTxsEvent")
			//========================for test hezi=================================//
			//for _, tx := range promoted {
			//	sendtxsch <- tx
			//}
			//=====================================================================//
			pool.txFeed.Send(NewTxsEvent{promoted})
		}
	}
	// If the pending limit is overflown, start equalizing allowances
	pending := uint64(0)
	for _, list := range pool.pending {
		pending += uint64(list.Len())
	}
	if pending > pool.config.GlobalSlots {
		pendingBeforeCap := pending
		// Assemble a spam order to penalize large transactors first
		spammers := prque.New()
		for addr, list := range pool.pending {
			// Only evict transactions from high rollers
			if !pool.locals.contains(addr) && uint64(list.Len()) > pool.config.AccountSlots {
				spammers.Push(addr, float32(list.Len()))
			}
		}
		// Gradually drop transactions from offenders
		offenders := []common.Address{}
		for pending > pool.config.GlobalSlots && !spammers.Empty() {
			// Retrieve the next offender if not local address
			offender, _ := spammers.Pop()
			offenders = append(offenders, offender.(common.Address))

			// Equalize balances until all the same or below threshold
			if len(offenders) > 1 {
				// Calculate the equalization threshold for all current offenders
				threshold := pool.pending[offender.(common.Address)].Len()

				// Iteratively reduce all offenders until below limit or threshold reached
				for pending > pool.config.GlobalSlots && pool.pending[offenders[len(offenders)-2]].Len() > threshold {
					for i := 0; i < len(offenders)-1; i++ {
						list := pool.pending[offenders[i]]
						for _, tx := range list.Cap(list.Len() - 1) {
							// Drop the transaction from the global pools too
							hash := tx.Hash()
							pool.all.Remove(hash)
							pool.priced.Removed()

							// Update the account nonce to the dropped transaction
							if nonce := tx.Nonce(); pool.pendingState.GetNonce(offenders[i]) > nonce {
								pool.pendingState.SetNonce(offenders[i], nonce)
							}
							log.Trace("Removed fairness-exceeding pending transaction", "hash", hash)
						}
						pending--
					}
				}
			}
		}
		// If still above threshold, reduce to limit or min allowance
		if pending > pool.config.GlobalSlots && len(offenders) > 0 {
			for pending > pool.config.GlobalSlots && uint64(pool.pending[offenders[len(offenders)-1]].Len()) > pool.config.AccountSlots {
				for _, addr := range offenders {
					list := pool.pending[addr]
					for _, tx := range list.Cap(list.Len() - 1) {
						// Drop the transaction from the global pools too
						hash := tx.Hash()
						pool.all.Remove(hash)
						pool.priced.Removed()

						// Update the account nonce to the dropped transaction
						if nonce := tx.Nonce(); pool.pendingState.GetNonce(addr) > nonce {
							pool.pendingState.SetNonce(addr, nonce)
						}
						log.Trace("Removed fairness-exceeding pending transaction", "hash", hash)
					}
					pending--
				}
			}
		}
		pendingRateLimitCounter.Inc(int64(pendingBeforeCap - pending))
	}
	// If we've queued more transactions than the hard limit, drop oldest ones
	queued := uint64(0)
	for _, list := range pool.queue {
		queued += uint64(list.Len())
	}
	if queued > pool.config.GlobalQueue {
		// Sort all accounts with queued transactions by heartbeat
		addresses := make(addresssByHeartbeat, 0, len(pool.queue))
		for addr := range pool.queue {
			if !pool.locals.contains(addr) { // don't drop locals
				addresses = append(addresses, addressByHeartbeat{addr, pool.beats[addr]})
			}
		}
		sort.Sort(addresses)

		// Drop transactions until the total is below the limit or only locals remain
		for drop := queued - pool.config.GlobalQueue; drop > 0 && len(addresses) > 0; {
			addr := addresses[len(addresses)-1]
			list := pool.queue[addr.address]

			addresses = addresses[:len(addresses)-1]

			// Drop all transactions if they are less than the overflow
			if size := uint64(list.Len()); size <= drop {
				for _, tx := range list.Flatten() {
					pool.removeTx(tx.Hash(), true)
				}
				drop -= size
				queuedRateLimitCounter.Inc(int64(size))
				continue
			}
			// Otherwise drop only last few transactions
			txs := list.Flatten()
			for i := len(txs) - 1; i >= 0 && drop > 0; i-- {
				pool.removeTx(txs[i].Hash(), true)
				drop--
				queuedRateLimitCounter.Inc(1)
			}
		}
	}
}

// demoteUnexecutables removes invalid and processed transactions from the pools
// executable/pending queue and any subsequent transactions that become unexecutable
// are moved back into the future queue.
func (pool *TxPool) demoteUnexecutables() {
	log.Info("========YY===1", "demoteUnexecutables():len(pool.pending)=", len(pool.pending))
	// Iterate over all accounts and demote any non-executable transactions
	for addr, list := range pool.pending {
		nonce := pool.currentState.GetNonce(addr)

		// Drop all transactions that are deemed too old (low nonce)
		for _, tx := range list.Forward(nonce) {
			//YY ========begin=========
			//log.Info("========YY===2","demoteUnexecutables()",0)
			pool.deleteMap(tx)
			//===========end===========
			hash := tx.Hash()
			//log.Trace("Removed old pending transaction", "hash", hash)
			pool.all.Remove(hash)
			pool.priced.Removed()
		}
		// Drop all transactions that are too costly (low balance or out of gas), and queue any invalids back for later
		drops, invalids := list.Filter(pool.currentState.GetBalance(addr), pool.currentMaxGas)
		for _, tx := range drops {
			//YY ========begin=========
			//log.Info("========YY===3","demoteUnexecutables()",0)
			pool.deleteMap(tx)
			//===========end===========
			hash := tx.Hash()
			log.Trace("Removed unpayable pending transaction", "hash", hash)
			pool.all.Remove(hash)
			pool.priced.Removed()
			pendingNofundsCounter.Inc(1)
		}
		for _, tx := range invalids {
			hash := tx.Hash()
			log.Trace("Demoting pending transaction", "hash", hash)
			pool.enqueueTx(hash, tx)
		}
		/* //YY 2018-08-30 因为我们的链Nonce不是必须连续的所以不需要这个操作
		// If there's a gap in front, warn (should never happen) and postpone all transactions
		if list.Len() > 0 && list.txs.Get(nonce) == nil {
			for _, tx := range list.Cap(0) {
				hash := tx.Hash()
				log.Error("Demoting invalidated transaction", "hash", hash)
				pool.enqueueTx(hash, tx)
			}
		}
		*/
		// Delete the entire queue entry if it became empty.
		if list.Empty() {
			delete(pool.pending, addr)
			delete(pool.beats, addr)
		}
	}
}

// addressByHeartbeat is an account address tagged with its last activity timestamp.
type addressByHeartbeat struct {
	address   common.Address
	heartbeat time.Time
}

type addresssByHeartbeat []addressByHeartbeat

func (a addresssByHeartbeat) Len() int           { return len(a) }
func (a addresssByHeartbeat) Less(i, j int) bool { return a[i].heartbeat.Before(a[j].heartbeat) }
func (a addresssByHeartbeat) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

// accountSet is simply a set of addresses to check for existence, and a signer
// capable of deriving addresses from transactions.
type accountSet struct {
	accounts map[common.Address]struct{}
	signer   types.Signer
}

// newAccountSet creates a new address set with an associated signer for sender
// derivations.
func newAccountSet(signer types.Signer) *accountSet {
	return &accountSet{
		accounts: make(map[common.Address]struct{}),
		signer:   signer,
	}
}

// contains checks if a given address is contained within the set.
func (as *accountSet) contains(addr common.Address) bool {
	_, exist := as.accounts[addr]
	return exist
}

// containsTx checks if the sender of a given tx is within the set. If the sender
// cannot be derived, this method returns false.
func (as *accountSet) containsTx(tx *types.Transaction) bool {
	if addr, err := types.Sender(as.signer, tx); err == nil {
		return as.contains(addr)
	}
	return false
}

// add inserts a new address into the set to track.
func (as *accountSet) add(addr common.Address) {
	as.accounts[addr] = struct{}{}
}

// txLookup is used internally by TxPool to track transactions while allowing lookup without
// mutex contention.
//
// Note, although this type is properly protected against concurrent access, it
// is **not** a type that should ever be mutated or even exposed outside of the
// transaction pool, since its internal state is tightly coupled with the pools
// internal mechanisms. The sole purpose of the type is to permit out-of-bound
// peeking into the pool in TxPool.Get without having to acquire the widely scoped
// TxPool.mu mutex.
type txLookup struct {
	all  map[common.Hash]*types.Transaction
	lock sync.RWMutex
}

// newTxLookup returns a new txLookup structure.
func newTxLookup() *txLookup {
	return &txLookup{
		all: make(map[common.Hash]*types.Transaction),
	}
}

// Range calls f on each key and value present in the map.
func (t *txLookup) Range(f func(hash common.Hash, tx *types.Transaction) bool) {
	t.lock.RLock()
	defer t.lock.RUnlock()

	for key, value := range t.all {
		if !f(key, value) {
			break
		}
	}
}

// Get returns a transaction if it exists in the lookup, or nil if not found.
func (t *txLookup) Get(hash common.Hash) *types.Transaction {
	t.lock.RLock()
	defer t.lock.RUnlock()

	return t.all[hash]
}

// Count returns the current number of items in the lookup.
func (t *txLookup) Count() int {
	t.lock.RLock()
	defer t.lock.RUnlock()

	return len(t.all)
}

// Add adds a transaction to the lookup.
func (t *txLookup) Add(tx *types.Transaction) {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.all[tx.Hash()] = tx
}

// Remove removes a transaction from the lookup.
func (t *txLookup) Remove(hash common.Hash) {
	t.lock.Lock()
	defer t.lock.Unlock()

	delete(t.all, hash)
}

//
////by hezi
//func GetTxtype(tx *types.Transaction) (txType string) {
//	tmpdt := make(map[string][]byte)
//	json.Unmarshal(tx.Data(), &tmpdt)
//	var data string
//	for data, _ = range tmpdt {
//	}
//	if strings.Contains(data, Heartbeat) {
//		txType = Heartbeat
//	} else if strings.Contains(data, Publickey) {
//		txType = Publickey
//	} else if strings.Contains(data, Privatekey) {
//		txType = Privatekey
//	} else {
//		txType = CommTx
//	}
//	return txType
//}

//by hezi
//keydata: hash值作为key
func insertdb(keydata []byte, val map[common.Address][]byte) error {
	dataval, err := json.Marshal(val)
	if err != nil {
		log.INFO("insertdb", "json.Marshal(val) err", err)
		return err
	}
	log.INFO("=====insertdb", "keydata", keydata, "dataval", dataval)
	return ldb.Put(keydata, dataval, nil)
}

//by hezi
func GetBroadcastTxs(height *big.Int, txtype string) (reqVal map[common.Address][]byte, err error) {
	//var val big.Int
	if height.Uint64() < common.GetBroadcastInterval() {
		return nil, errors.New("Invalid height")
	}
	//val.Quo(height, big.NewInt(100)) // val = a/100
	val := height.Uint64() / common.GetBroadcastInterval()

	strVal := fmt.Sprintf("%v", val)
	hv := types.RlpHash(txtype + strVal)
	log.INFO("GetBroadcastTxs", "keydata:", txtype+strVal, "write hash key:", hv)
	dataval, err := ldb.Get(hv.Bytes(), nil)
	if err != nil {
		log.ERROR("GetBroadcastTxs", "Get broadcast failed", err)
		return nil, err
	}
	err = json.Unmarshal(dataval, &reqVal)
	if err != nil {
		log.ERROR("GetBroadcastTxs", "Unmarshal failed", 1)
	}
	return reqVal, err
}

//by hezi
//该接口已废弃
//func (pool *TxPool) getBroadcastTxs(height *big.Int, txtype string) map[common.Address]*types.Transaction {
//	pool.mu.Lock()
//	defer pool.mu.Unlock()
//
//	var val big.Int
//	SpecialMem := make(map[common.Hash]map[common.Address]*types.Transaction)
//	val.Quo(height, big.NewInt(100))
//	strVal := fmt.Sprintf("%v", val)
//	hv := types.RlpHash(strVal + txtype)
//	return SpecialMem[hv]
//}

//by hezi
//func GetSeedTxKey(tx *types.Transaction) (seedKey interface{}) {
//	tmpdt := make(map[string][]byte)
//	json.Unmarshal(tx.Data(), &tmpdt)
//
//	for _, val := range tmpdt {
//		rlp.DecodeBytes(val, &seedKey)
//	}
//	return seedKey
//}

//by hezi
func (pool *TxPool) GetAllSpecialTxs() (reqVal map[common.Address]types.Transactions) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	var from common.Address
	reqVal = make(map[common.Address]types.Transactions, 0)
	for _, tx := range pool.Special {
		from, _ = types.Sender(pool.signer, tx)
		reqVal[from] = append(reqVal[from], tx)
	}

	pool.Special = make(map[common.Hash]*types.Transaction, 0)
	return reqVal
}

//hezi
func SetBroadcastTxsPoolFilter(FilterType string, Filter interface{}) {
	whitemap = Filter.(map[common.Address]bool)
}

//hezi
func GetBroadcastTxsPoolFilter(FilterType string) map[common.Address]bool {
	return whitemap
}
