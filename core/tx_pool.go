package core

import (
	"container/list"
	"encoding/json"
	"errors"
	//"github.com/MatrixAINetwork/go-matrix/p2p/discover"
	"math/big"
	"sync"
	"time"

	"github.com/MatrixAINetwork/go-matrix/ca"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/matrixstate"
	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/event"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/metrics"
	"github.com/MatrixAINetwork/go-matrix/p2p"
	"github.com/MatrixAINetwork/go-matrix/params"
	"github.com/MatrixAINetwork/go-matrix/rlp"
	"github.com/MatrixAINetwork/go-matrix/txpoolCache"
	"runtime"
)

//
const (
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

	// ErrKnownTransaction is returned if a transaction is known or existent.
	ErrKnownTransaction = errors.New("known transaction")

	// ErrReplaceUnderpriced is returned if a transaction is attempted to be replaced
	// with a different one without the required price bump.
	ErrReplaceUnderpriced = errors.New("replacement transaction underpriced")

	// ErrInsufficientFunds is returned if the total cost of executing a transaction
	// is higher than the balance of the user's account.
	ErrInsufficientFunds        = errors.New("insufficient funds for gas * price + value")
	ErrEntrustInsufficientFunds = errors.New("entrust insufficient funds for gas * price")

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

	//
	ErrTXCountOverflow = errors.New("transaction quantity spillover")
	ErrTXToNil         = errors.New("transaction`s to(common.address) is nil")
	ErrTXUnknownType   = errors.New("Unknown extra txtype")
	ErrTxToRepeat      = errors.New("contains duplicate transfer accounts")
	ErrTXWrongful      = errors.New("transaction is unlawful")
	ErrTXPoolFull      = errors.New("txpool is full")
	ErrTXNonceSame     = errors.New("the same Nonce transaction exists")
	ErrRepeatEntrust   = errors.New("Repeat Entrust")
	ErrWithoutAuth     = errors.New("gas entrust not set ")
	ErrinterestAmont   = errors.New("Incorrect total interest")
	ErrSpecialTxFailed = errors.New("Run special tx failed")
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

const (
	TxStatusUnknown TxStatus = iota
	TxStatusQueued
	TxStatusPending
	TxStatusIncluded
)

//======struct//
type mapst struct {
	slist []*big.Int
	mlock sync.RWMutex
}

//
type listst struct {
	list *list.List
}

//
type sendst struct {
	snlist mapst
	lst    listst
	lstMu  sync.Mutex
	notice chan *big.Int
}

//global  //
var gSendst sendst

// blockChain provides the state of blockchain and current gas limit to do
// some pre checks in tx pool and event subscribers.
type blockChain interface {
	CurrentBlock() *types.Block
	GetBlock(hash common.Hash, number uint64) *types.Block
	StateAt(root common.Hash) (*state.StateDB, error)
	State() (*state.StateDB, error)
	SubscribeChainHeadEvent(ch chan<- ChainHeadEvent) event.Subscription
	GetA0AccountFromAnyAccountAtSignHeight(account common.Address, blockHash common.Hash, signHeight uint64) (common.Address, common.Address, error)
}

type ConsensusNTx struct {
	Key   uint32
	Value types.SelfTransaction
}

// TxPoolConfig are the configuration parameters of the transaction pool.
type TxPoolConfig struct {
	PriceLimit   uint64 // Minimum gas price to enforce for acceptance into the pool
	AccountSlots uint64 // Minimum number of executable transaction slots guaranteed per account
	GlobalSlots  uint64 // Maximum number of executable transaction slots for all accounts
	AccountQueue uint64 // Maximum number of non-executable transaction slots permitted per account
	GlobalQueue  uint64 // Maximum number of non-executable transaction slots for all accounts
	txTimeout    time.Duration
}

// DefaultTxPoolConfig contains the default configurations for the transaction
// pool.
var DefaultTxPoolConfig = TxPoolConfig{
	PriceLimit:   params.TxGasPrice, // 2018-08-29 由1改为此值
	AccountSlots: 16,
	GlobalSlots:  4096 * 5 * 5 * 10, // 2018-08-30 改为乘以5
	AccountQueue: 64 * 1000,
	GlobalQueue:  1024 * 60,
	txTimeout:    180 * time.Second,
}

type NormalTxPool struct {
	config      TxPoolConfig
	chainconfig *params.ChainConfig
	chain       blockChain
	gasPrice    *big.Int
	//txFeed       event.Feed
	//scope        event.SubscriptionScope
	chainHeadCh  chan ChainHeadEvent
	sendTxCh     chan NewTxsEvent
	chainHeadSub event.Subscription
	signer       types.Signer
	mu           sync.RWMutex

	currentState  *state.StateDB      // Current state in the blockchain head
	pendingState  *state.ManagedState // Pending state tracking virtual nonces
	currentMaxGas uint64              // Current gas limit for transaction caps

	pending map[common.Address]*txList // All currently processable transactions
	all     *txLookup                  // All transactions to allow lookups
	//=================by  ==================//
	SContainer map[common.Hash]*types.Transaction
	NContainer map[uint32]*types.Transaction
	udptxsCh   chan []*types.Transaction_Mx //udp交易订阅
	quit       chan struct{}
	udptxsSub  event.Subscription //取消订阅
	//=================================================//
	//priced *txPricedList // All transactions sorted by price

	wg sync.WaitGroup // for shutdown sync

	//selfmlk sync.RWMutex //

	mapNs         sync.Map                         //
	mapCaclErrtxs map[common.Hash][]common.Address //  用来统计错误的交易
	mapDelErrtxs  map[common.Hash]*big.Int         //  用来删除mapErrorTxs
	mapErrorTxs   map[*big.Int]*types.Transaction  //  存放所有的错误交易（20个区块自动删除）
	mapTxsTiming  map[common.Hash]time.Time        //  需要做定时删除的交易
	mapHighttx    map[uint64][]uint32
}

// sanitize checks the provided user configurations and changes anything that's
// unreasonable or unworkable.
func (config *TxPoolConfig) sanitize() TxPoolConfig {
	conf := *config
	if conf.PriceLimit < 1 {
		log.Warn("Sanitizing invalid txpool price limit", "provided", conf.PriceLimit, "updated", DefaultTxPoolConfig.PriceLimit)
		conf.PriceLimit = DefaultTxPoolConfig.PriceLimit
	}
	return conf
}

func NewTxPool(config TxPoolConfig, chainconfig *params.ChainConfig, chain blockChain, sendch chan NewTxsEvent) *NormalTxPool {
	// Sanitize the input to ensure no vulnerable gas prices are set
	config = (&config).sanitize()
	// Create the transaction pool with its initial settings
	nPool := &NormalTxPool{
		config:        config,
		chainconfig:   chainconfig,
		chain:         chain,
		signer:        types.NewEIP155Signer(chainconfig.ChainId),
		pending:       make(map[common.Address]*txList),
		SContainer:    make(map[common.Hash]*types.Transaction), //by
		NContainer:    make(map[uint32]*types.Transaction),      //by
		udptxsCh:      make(chan []*types.Transaction_Mx, 0),    //
		sendTxCh:      make(chan NewTxsEvent),
		quit:          make(chan struct{}),
		all:           newTxLookup(),
		chainHeadCh:   make(chan ChainHeadEvent, chainHeadChanSize),
		gasPrice:      new(big.Int).SetUint64(config.PriceLimit),
		mapCaclErrtxs: make(map[common.Hash][]common.Address), //  用来统计错误的交易
		mapDelErrtxs:  make(map[common.Hash]*big.Int),         //  用来删除mapErrorTxs
		mapErrorTxs:   make(map[*big.Int]*types.Transaction),  //  存放所有的错误交易（20个区块自动删除）
		mapTxsTiming:  make(map[common.Hash]time.Time),        //  需要做定时删除的交易
		mapHighttx:    make(map[uint64][]uint32, 0),
	}
	//nPool.pool.priced = newTxPricedList(nPool.pool.all)
	nPool.reset(nil, chain.CurrentBlock().Header())

	// Subscribe events from blockchain
	nPool.chainHeadSub = nPool.chain.SubscribeChainHeadEvent(nPool.chainHeadCh)
	nPool.sendTxCh = sendch

	// Start the event loop and return
	nPool.wg.Add(3)

	gSendst.lst.list = list.New() //
	gSendst.snlist.slist = make([]*big.Int, 0)
	gSendst.notice = make(chan *big.Int, 1)

	//udp 交易订阅
	nPool.udptxsSub, _ = mc.SubscribeEvent(mc.SendUdpTx, nPool.udptxsCh)

	go nPool.loop()
	go nPool.checkList() //
	go nPool.ListenUdp()

	return nPool
}

// Type return txpool type.
func (nPool *NormalTxPool) Type() byte {
	return types.NormalTxIndex
}

// loop is the transaction pool's main event loop, waiting for and reacting to
// outside blockchain events as well as for various reporting and transaction
// eviction events.
func (nPool *NormalTxPool) loop() {
	defer nPool.wg.Done()

	delteTime := time.NewTicker(10 * time.Second)
	defer delteTime.Stop()
	// Track the previous head headers for transaction reorgs
	head := nPool.chain.CurrentBlock()

	// Keep waiting for and reacting to the various events
	for {
		select {
		// Handle ChainHeadEvent
		case ev := <-nPool.chainHeadCh:
			if ev.Block != nil {
				nPool.mu.Lock()
				nPool.reset(head.Header(), ev.Block.Header())
				head = ev.Block
				h := head.Number().Uint64() - 1
				if txlist, ok := nPool.mapHighttx[h]; ok {
					for _, n := range txlist {
						nPool.deletnTx(n)
					}
				}
				delete(nPool.mapHighttx, h)
				txpoolCache.DeleteTxCache(head.Header().HashNoSignsAndNonce(), head.Number().Uint64())
				nPool.mu.Unlock()
				nPool.getPendingTx() //
			}
			// Be unsubscribed due to system stopped
		case <-nPool.chainHeadSub.Err():
			return
		case <-delteTime.C:
			nPool.mu.Lock()
			nPool.blockTiming() //
			nPool.mu.Unlock()
			nPool.getPendingTx()

		}
	}
}

// sTxValIsNil verification transaction's N if nil
func (nPool *NormalTxPool) sTxValIsNil(s *big.Int, isLock bool) bool {
	if isLock {
		nPool.mu.Lock()
		defer nPool.mu.Unlock()
	}

	if tx, ok := nPool.SContainer[common.BigToHash(s)]; ok {
		return len(tx.N) == 0
	}
	return true
}

// setTxNum set N for transaction
func (nPool *NormalTxPool) setTxNum(tx *types.Transaction, num uint32, isLock bool) {
	if isLock {
		nPool.mu.Lock()
		defer nPool.mu.Unlock()
	}
	tx.N = append(tx.N, num)
}

func (nPool *NormalTxPool) setsTx(s *big.Int, tx *types.Transaction) {
	nPool.SContainer[common.BigToHash(s)] = tx
}

func (nPool *NormalTxPool) getTxbyS(s *big.Int, isLock bool) *types.Transaction {
	if isLock {
		nPool.mu.Lock()
		defer nPool.mu.Unlock()
	}
	return nPool.SContainer[common.BigToHash(s)]
}

func (nPool *NormalTxPool) setnTx(num uint32, tx *types.Transaction, isLock bool) {
	if isLock {
		nPool.mu.Lock()
		defer nPool.mu.Unlock()
	}
	nPool.NContainer[num] = tx
}

func (nPool *NormalTxPool) getTxbyN(num uint32, isLock bool) *types.Transaction {
	if isLock {
		nPool.mu.Lock()
		defer nPool.mu.Unlock()
	}
	return nPool.NContainer[num]
}

func (nPool *NormalTxPool) deletnTx(num uint32) {
	delete(nPool.NContainer, num)
}

func (nPool *NormalTxPool) deletsTx(s *big.Int) {
	delete(nPool.SContainer, common.BigToHash(s))
}

// packageSNList
func (nPool *NormalTxPool) packageSNList() {
	if len(gSendst.snlist.slist) == 0 {
		log.Trace("txpool:packageSNList_if", "len(gSendst.snlist.slist)", len(gSendst.snlist.slist))
		return
	} else {
		log.Trace("txpool:packageSNList_else", "len(gSendst.snlist.slist)", len(gSendst.snlist.slist))
	}
	lst := gSendst.snlist.slist
	gSendst.snlist.slist = make([]*big.Int, 0)
	go func(lst []*big.Int) {
		tmpsnlst := make(map[uint32]*big.Int)

		nPool.mu.Lock()
		for _, s := range lst {
			if nPool.sTxValIsNil(s, false) {
				tx := nPool.getTxbyS(s, false)
				if tx == nil {
					log.Error("packageSNList", "tx is nil", s)
					continue
				}
				tmpnum := byte4Number.catNumber()
				//log.Trace("txpool:packageSNList33", "tx N", tmpnum)
				nPool.setTxNum(tx, tmpnum, false)
				tmpsnlst[tmpnum] = s
				nPool.setnTx(tmpnum, tx, false)
			}
		}
		nPool.mu.Unlock()
		log.Trace("file tx_pool", "func packageSNList :send tmpsnlst", len(tmpsnlst))
		if len(tmpsnlst) > 0 {
			bt, _ := json.Marshal(tmpsnlst)
			nPool.SendMsg(MsgStruct{Msgtype: SendFloodSN, MsgData: bt})
		}
	}(lst)
}

// ProcessMsg
func (nPool *NormalTxPool) ProcessMsg(m NetworkMsgData) {
	if len(m.Data) <= 0 {
		log.Error("NormalTxPool::ProcessMsg  data is nil")
		return
	}

	var (
		msgData = m.Data[0]
		err     error
	)

	switch msgData.Msgtype {
	case SendFloodSN:
		snMap := make(map[uint32]*big.Int)
		if err = json.Unmarshal(msgData.MsgData, &snMap); err != nil {
			log.Error("func ProcessMsg", "case SendFloodSN:Unmarshal_err=", err)
			break
		}
		nPool.CheckTx(snMap, m.SendAddress)
	case GetTxbyN:
		listN := make([]uint32, 0)
		if err = json.Unmarshal(msgData.MsgData, &listN); err != nil {
			log.Error("func ProcessMsg", "case GetTxbyN:Unmarshal_err=", err)
			break
		}
		nPool.GetTxByN(listN, m.SendAddress)
	case GetConsensusTxbyN:
		listN := make([]uint32, 0)
		if err = json.Unmarshal(msgData.MsgData, &listN); err != nil {
			log.Error("func ProcessMsg", "case GetConsensusTxbyN:Unmarshal_err=", err)
			break
		}
		nPool.GetConsensusTxByN(listN, m.SendAddress)
	case RecvTxbyN:
		ntx := make(map[uint32]*types.Floodtxdata, 0)
		if err = json.Unmarshal(msgData.MsgData, &ntx); err != nil {
			log.Error("func ProcessMsg", "case RecvTxbyN:Unmarshal_err=", err)
			break
		}
		nPool.RecvFloodTx(ntx, m.SendAddress)
	case RecvConsensusTxbyN:
		mapNtx := make([]*ConsensusNTx, 0)
		err = rlp.DecodeBytes(msgData.MsgData, &mapNtx)
		if err != nil {
			log.Error("func ProcessMsg", "case GetConsensusTxbyN:DecodeBytes_err=", err)
			break
		}
		ntx := make(map[uint32]types.SelfTransaction, 0)
		for _, val := range mapNtx {
			ntx[val.Key] = val.Value
		}
		nPool.RecvConsensusFloodTx(ntx, m.SendAddress)
		//ntx := make(map[uint32]types.SelfTransaction, 0)
		//if err = json.Unmarshal(msgData.MsgData, &ntx); err != nil {
		//	log.Error("func ProcessMsg", "case RecvConsensusTxbyN:Unmarshal_err=", err)
		//	break
		//}
		//nPool.RecvConsensusFloodTx(ntx, m.NodeId)
	case RecvErrTx:
		listS := make([]*big.Int, 0)
		if err = json.Unmarshal(msgData.MsgData, &listS); err != nil {
			log.Error("func ProcessMsg", "case RecvErrTx:Unmarshal_err=", err)
			break
		}
		nPool.RecvErrTx(common.HexToAddress(m.SendAddress.String()), listS)
	}
}

// SendMsg
func (nPool *NormalTxPool) SendMsg(data MsgStruct) {
	selfRole := ca.GetRole()
	data.TxpoolType = types.NormalTxIndex
	switch data.Msgtype {
	case SendFloodSN:
		if selfRole == common.RoleValidator || selfRole == common.RoleMiner {
			log.Info("txpool transaction flood", "selfRole", selfRole)
			p2p.SendToGroupWithBackup(common.RoleValidator|common.RoleBackupValidator|common.RoleBroadcast, common.NetworkMsg, []interface{}{data})
		}
	case GetTxbyN, RecvTxbyN, GetConsensusTxbyN, RecvConsensusTxbyN: //
		//给固定的节点发送根据N获取Tx的请求
		p2p.SendToSingle(data.SendAddr, common.NetworkMsg, []interface{}{data})
	case RecvErrTx: // 给全部验证者发送错误交易做共识
		if selfRole == common.RoleValidator {
			p2p.SendToGroup(common.RoleValidator, common.NetworkMsg, []interface{}{data})
		}
	}
}

//by
func (nPool *NormalTxPool) checkList() {
	flood := time.NewTicker(params.FloodTime)
	defer func() {
		flood.Stop()
		nPool.wg.Done()
	}()

	for {
		select {
		case <-flood.C:
			nPool.packageSNList()

		case s := <-gSendst.notice:
			gSendst.snlist.slist = append(gSendst.snlist.slist, s)
			if len(gSendst.snlist.slist) >= params.FloodMaxTransactions {
				nPool.packageSNList()
			}
		case <-nPool.quit:
			return
		}
	}
}

func (nPool *NormalTxPool) ListenUdp() {
	defer func() {
		nPool.udptxsSub.Unsubscribe() //udp交易取消订阅
		nPool.wg.Done()
	}()

	for {
		select {
		//udp接收的交易，此处应该只发给验证节点
		case evtxs := <-nPool.udptxsCh:
			log.Info("txpool listenudp", "checklist: udptxs:", len(evtxs))
			selfRole := ca.GetRole()
			if selfRole == common.RoleValidator {
				tmptxs := make([]*types.Transaction, 0)
				for _, ftx := range evtxs {
					tx := types.ConvMxtotx(ftx)
					if nc := tx.Nonce(); nc < params.NonceAddOne {
						nc = nc | params.NonceAddOne
						tx.SetNonce(nc)
					}
					log.INFO("==udp tx hash","from",tx.From().String(),"tx.Nonce",tx.Nonce(),"hash",tx.Hash().String())
					tmptxs = append(tmptxs, tx)
				}
				nPool.getFromByTx(tmptxs)
				nPool.mu.Lock()
				for _, tx := range tmptxs {
					tmptx := nPool.getTxbyS(tx.GetTxS(), false)
					if tmptx == nil {
						_, err := nPool.add(tx, false)
						if err != nil {
							log.Info("txpool listenudp", "listenudp err:", err)
						}
					}
				}
				nPool.mu.Unlock()
			}
		case <-nPool.udptxsSub.Err():
			return
		}
	}
}
func (nPool *NormalTxPool) reset(oldHead, newHead *types.Header) {
	// Initialize the internal state to the current head
	if newHead == nil {
		newHead = nPool.chain.CurrentBlock().Header() // Special case during testing
	}
	statedb, err := nPool.chain.StateAt(newHead.Root)
	if err != nil {
		log.Error("Failed to reset txpool state", "err", err)
		return
	}
	nPool.currentState = statedb
	nPool.pendingState = state.ManageState(statedb)
	nPool.currentMaxGas = newHead.GasLimit

	// validate the pool of pending transactions, this will remove
	// any transactions that have been included in the block or
	// have been invalidated because of another transaction (e.g.
	// higher gas price)
	nPool.DemoteUnexecutables()

	// Update all accounts to the latest known pending nonce
	for addr, list := range nPool.pending {
		for _, txs := range list.txs {
			txs := txs.Flatten() // Heavy but will be cached and is needed by the miner anyway
			nPool.pendingState.SetNonce(addr, txs[len(txs)-1].Nonce()+1)
		}
	}
}

// Stop terminates the transaction pool.
func (nPool *NormalTxPool) Stop() {
	// Unsubscribe subscriptions registered from blockchain
	nPool.chainHeadSub.Unsubscribe()
	nPool.udptxsSub.Unsubscribe()
	nPool.quit <- struct{}{}
	nPool.wg.Wait()
	close(nPool.quit)

	log.Info("Transaction pool stopped")
}

// SubscribeNewTxsEvent registers a subscription of NewTxsEvent and
// starts sending event to the given channel.
//func (nPool *NormalTxPool) SubscribeNewTxsEvent(ch chan<- NewTxsEvent) event.Subscription {
//	return nPool.scope.Track(nPool.txFeed.Subscribe(ch))
//}

// State returns the virtual managed state of the transaction pool.
func (nPool *NormalTxPool) State() *state.ManagedState {
	nPool.mu.RLock()
	defer nPool.mu.RUnlock()

	return nPool.pendingState
}

// Stats retrieves the current pool stats, namely the number of pending and the
// number of queued (non-executable) transactions.
func (nPool *NormalTxPool) Stats() (int, int) {
	nPool.mu.RLock()
	defer nPool.mu.RUnlock()

	return nPool.stats()
}

// stats retrieves the current pool stats, namely the number of pending and the
// number of queued (non-executable) transactions.
func (nPool *NormalTxPool) stats() (int, int) {
	pending := 0
	for _, list := range nPool.pending {
		for _, txs := range list.txs {
			pending += txs.Len() //list.Len(typ)
		}
	}
	queued := 0
	return pending, queued
}

// Content retrieves the data content of the transaction pool, returning all the
// pending as well as queued transactions, grouped by account and sorted by nonce.
func (nPool *NormalTxPool) Content() map[common.Address][]*types.Transaction {
	nPool.mu.Lock()
	defer nPool.mu.Unlock()
	pending := make(map[common.Address][]*types.Transaction)
	for addr, list := range nPool.pending {
		txlist := make([]*types.Transaction, 0)
		for _, txs := range list.txs {
			txlist = append(txlist, txs.Flatten()...)
		}
		pending[addr] = txlist
	}
	//queued := make(map[common.Address][]*types.Transaction)
	return pending
}

// Pending retrieves all currently processable transactions, groupped by origin
// account and sorted by nonce. The returned transaction set is a copy and can be
// freely modified by calling code.
func (nPool *NormalTxPool) Pending() (map[common.Address][]types.SelfTransaction, error) {
	nPool.mu.Lock()
	defer nPool.mu.Unlock()
	pending := make(map[common.Address][]types.SelfTransaction)
	for addr, list := range nPool.pending {
		txlist := make([]*types.Transaction, 0)
		for _, txs := range list.txs {
			txlist = append(txlist, txs.Flatten()...)
		}
		var txser types.SelfTransactions
		for _, tx := range txlist {
			txser = append(txser, tx)
			if len(tx.N) <= 0 {
				continue
			}
			nPool.NContainer[tx.N[0]] = tx
		}
		pending[addr] = txser //.Flatten()
	}
	return pending, nil
}

// 获取pending中剩余的交易（广播区块头后触发）
//区块产生后将Pending中剩余的交易放入区块定时中，如果二十个区块还没有被打包则删除，如果已经被打包了则也删除
func (nPool *NormalTxPool) getPendingTx() {
	nPool.mu.Lock()
	pending := make(map[common.Address][]*types.Transaction)
	for addr, list := range nPool.pending {
		txlist := make([]*types.Transaction, 0)
		for _, txs := range list.txs {
			txlist = append(txlist, txs.Flatten()...)
		}
		pending[addr] = txlist
	}
	for _, txs := range pending {
		for _, tx := range txs {
			nPool.addBlockTiming(tx.Hash())
		}
	}
	nPool.mu.Unlock()
}

// 检查当前map中是否存在洪泛过来的交易
func (nPool *NormalTxPool) CheckTx(mapSN map[uint32]*big.Int, nid common.Address) {
	log.Info("txpool msg_CheckTx IN", "len(mapSN)", len(mapSN))
	defer log.Info("txpool msg_CheckTx OUT")
	listN := make([]uint32, 0)
	nPool.mu.Lock()
	for n, s := range mapSN {
		if s == nil || n == 0 { //如果S或者N 不合法则直接跳过
			continue
		}
		nPool.mapNs.Store(n, s)
		tx := nPool.getTxbyS(s, false)
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
				nPool.setTxNum(tx, n, false)
			}
			nPool.setnTx(n, tx, false)
		}
	}
	nPool.mu.Unlock()
	if len(listN) > 0 {
		msData, _ := json.Marshal(listN)
		nPool.SendMsg(MsgStruct{Msgtype: GetTxbyN, SendAddr: nid, MsgData: msData})
	}
}

// 接收到Leader打包的交易共识消息时根据N获取tx (调用本方法需要启动协程)
func (nPool *NormalTxPool) ReturnAllTxsByN(listN []uint32, resqe byte, addr common.Address, retch chan *RetChan_txpool) {
	log.Info("txpool returnAllTxsByN", "len(listN)", len(listN))
	if len(listN) <= 0 {
		retch <- &RetChan_txpool{nil, nil, resqe}
		return
	}
	txs := make([]types.SelfTransaction, 0)
	ns := make([]uint32, 0)
	nPool.mu.Lock()
	for _, n := range listN {
		tx := nPool.getTxbyN(n, false)
		if tx != nil {
			txs = append(txs, tx)
		} else {
			//当根据N找不到对应的交易时需要跟对方索要一次
			ns = append(ns, n)
		}
	}
	nPool.mu.Unlock()
	log.Trace("file txpool", "ReturnAllTxsByN:len(ns)", len(ns), "len(txs):", len(txs))
	if len(ns) > 0 {
		txs = make([]types.SelfTransaction, 0)
		//nid, err1 := ca.ConvertAddressToNodeId(addr)
		//log.Info("leader node", "addr::", addr, "id::", nid.String())
		//if err1 != nil {
		//	log.Error("file txpool", "ReturnAllTxsByN:discover=err", err1)
		//	retch <- &RetChan_txpool{nil, err1, resqe}
		//	return
		//}
		msData, err2 := json.Marshal(ns)
		if err2 != nil {
			log.Error("file txpool", "ReturnAllTxsByN:Marshal=err", err2)
			retch <- &RetChan_txpool{nil, err2, resqe}
			return
		}
		// 发送缺失交易N的列表
		nPool.SendMsg(MsgStruct{Msgtype: GetConsensusTxbyN, SendAddr: addr, MsgData: msData}) //modi  (共识要的交易都带s)

		rettime := time.NewTimer(4 * time.Second) // 2秒后没有收到需要的交易则返回
	forBreak:
		for {
			select {
			case <-rettime.C:
				log.Info("txpool returnAllTxsByN", "Time Out=", 0)
				break forBreak
			case <-time.After(500 * time.Millisecond): //500毫秒轮训一次
				tmpns := make([]uint32, 0)
				for _, n := range ns {
					tx := nPool.getTxbyN(n, true)
					if tx == nil {
						tmpns = append(tmpns, n)
					}
				}
				ns = tmpns
				if len(ns) == 0 {
					log.Trace("file txpool", "ReturnAllTxsByN:recvTx Over=", 0)
					break forBreak
				}
			}
		}
		var txerr error
		if len(ns) > 0 {
			txerr = errors.New("loss tx")
		} else {
			nPool.mu.Lock()
			for _, n := range listN {
				tx := nPool.getTxbyN(n, false)
				if tx != nil {
					txs = append(txs, tx)
				} else {
					txerr = errors.New("else loss tx")
					txs = make([]types.SelfTransaction, 0)
					break
				}
			}
			nPool.mu.Unlock()
		}
		retch <- &RetChan_txpool{txs, txerr, resqe}
		log.Trace("file txpool end if", "ReturnAllTxsByN:len(ns)", len(ns), "err", txerr)
	} else {
		retch <- &RetChan_txpool{txs, nil, resqe}
		log.Trace("file txpool end else", "ReturnAllTxsByN", "return success")
	}
}

// (共识要交易)根据N值获取对应的交易(modi  )
func (nPool *NormalTxPool) GetConsensusTxByN(listN []uint32, nid common.Address) {
	log.Trace("txpool GetConsensusTxByN ", "msg_GetConsensusTxByN:len(listN)", len(listN))
	if len(listN) <= 0 {
		return
	}
	mapNtx := make([]*ConsensusNTx, 0)
	nPool.mu.Lock()
	for _, n := range listN {
		tx := nPool.getTxbyN(n, false)
		if tx != nil {
			ntx := &ConsensusNTx{n, tx}
			mapNtx = append(mapNtx, ntx)
		} else {
			log.Info("txpool getConsensusTxByN: tx is nil")
		}
	}
	if len(mapNtx) != len(listN) {
		tmpMap := txpoolCache.GetTxByN_Cache(listN, nPool.chain.CurrentBlock().Number().Uint64())
		log.Info("txpool getConsensusTxByN", "len(tmpMap)", len(tmpMap))
		if tmpMap != nil {
			if len(tmpMap) == len(listN) {
				for k, v := range tmpMap {
					ntx := &ConsensusNTx{k, v}
					mapNtx = append(mapNtx, ntx)
				}
			} else {
				log.Info("txpool getConsensusTxByN", "len(mapNtx)", len(mapNtx))
				for _, n := range listN {
					tx := nPool.getTxbyN(n, false)
					if tx != nil {
						ntx := &ConsensusNTx{n, tx}
						mapNtx = append(mapNtx, ntx)
					} else {
						if ttx, ok := tmpMap[n]; ok {
							ntx := &ConsensusNTx{n, ttx}
							mapNtx = append(mapNtx, ntx)
						}
					}
				}
			}
		}
	}
	nPool.mu.Unlock()
	//msData, _ := json.Marshal(mapNtx)
	msData, err := rlp.EncodeToBytes(mapNtx)
	if err == nil {
		nPool.SendMsg(MsgStruct{Msgtype: RecvConsensusTxbyN, SendAddr: nid, MsgData: msData})
		log.Trace("txpool getConsensusTxByN", "GetConsensusTxByN:ntxMap", len(mapNtx), "nodeid", nid.String())
	} else {
		log.Trace("tx_pool getConsensusTxByN", "GetConsensusTxByN:EncodeToBytes err", err)
	}
}

// 根据N值获取对应的交易(洪泛)
func (nPool *NormalTxPool) GetTxByN(listN []uint32, nid common.Address) {
	log.Trace("txpool getTxByN", "msg_GetTxByN:len(listN)", len(listN))
	if len(listN) <= 0 {
		return
	}
	mapNtx := make(map[uint32]*types.Floodtxdata)
	nPool.mu.Lock()
	for _, n := range listN {
		tx := nPool.getTxbyN(n, false)
		if tx != nil {
			ftx := types.GetFloodData(tx)
			mapNtx[n] = ftx
		} else {
			log.Info("txpool getTxByN :tx is nil")
		}
	}
	nPool.mu.Unlock()
	msData, _ := json.Marshal(mapNtx)
	nPool.SendMsg(MsgStruct{Msgtype: RecvTxbyN, SendAddr: nid, MsgData: msData})
}

//此接口传的交易带s(modi  )
func (nPool *NormalTxPool) RecvConsensusFloodTx(mapNtx map[uint32]types.SelfTransaction, nid common.Address) {
	//nPool.selfmlk.Lock()
	//log.Info("txpool msg_RecvConsensusFloodTx", "len(mapNtx)=", len(mapNtx))
	//defer log.Info("txpool msg_RecvConsensusFloodTx defer ", "len(mapNtx)=", 0)
	errorTxs := make([]*big.Int, 0)
	txs := make([]*types.Transaction, 0)
	tmpNtx := make(map[uint32]*types.Transaction)
	nlist := make([]uint32, 0)
	for n, txer := range mapNtx {
		ss := txer.GetTxS()
		nPool.mapNs.Store(n, ss)
		ts, ok := nPool.mapNs.Load(n)
		if !ok {
			continue
		}
		s := ts.(*big.Int)
		if s == nil || n == 0 { //如果S或者N 不合法则直接跳过
			continue
		}
		tx, ok := txer.(*types.Transaction)
		if !ok {
			continue
		}
		txs = append(txs, tx)
		tmpNtx[n] = tx
	}
	nPool.getFromByTx(txs)
	log.Info("txpool msg_RecvConsensusFloodTx()", "len(mapNtx)=", len(tmpNtx))
	nPool.mu.Lock()
	for n, tx := range tmpNtx {
		ts, ok := nPool.mapNs.Load(n)
		if !ok {
			continue
		}
		s := ts.(*big.Int)
		if s == nil || n == 0 { //如果S或者N 不合法则直接跳过
			log.Info("txpool msg_RecvConsensusFloodTx()", "s or n is nil : s ", s, "n:", n)
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
			tmptx := nPool.getTxbyS(s, false)
			if tmptx != nil {
				tx = tmptx
			}
			nPool.setTxNum(tx, n, false)
			nPool.setnTx(n, tx, false)
			nlist = append(nlist, n)
		} else {
			log.Info("txpool msg_RecvConsensusFloodTx,tx N is same")
		}
	}
	if len(nlist) > 0 {
		nPool.mapHighttx[nPool.chain.CurrentBlock().Number().Uint64()] = nlist
	}
	nPool.mu.Unlock()
	//nPool.selfmlk.Unlock()

	if len(errorTxs) > 0 {
		//TODO S在这如何进行签名？？如何获得本节点账户信息
		msData, err := json.Marshal(errorTxs)
		if err != nil {
			log.Error("txpool msg_RecvConsensusFloodTx", "send error Tx,Marshal err:", err)
		} else {
			nPool.SendMsg(MsgStruct{Msgtype: RecvErrTx, SendAddr: nid, MsgData: msData})
			nPool.RecvErrTx(common.Address{}, errorTxs)
		}
	}
}

// 接收洪泛的交易（根据N请求到的交易）
func (nPool *NormalTxPool) RecvFloodTx(mapNtx map[uint32]*types.Floodtxdata, nid common.Address) {
	//nPool.selfmlk.Lock()
	log.Info("txpool msg_RecvFloodTx", "len(mapNtx)=", len(mapNtx))
	//defer log.Info("func msg_RecvFloodTx defer ", "msg_RecvFloodTx: len(mapNtx)=", 0)
	errorTxs := make([]*big.Int, 0)
	txs := make([]*types.Transaction, 0)
	tmpNtx := make(map[uint32]*types.Transaction)
	for n, ftx := range mapNtx {
		ts, ok := nPool.mapNs.Load(n)
		if !ok {
			continue
		}
		s := ts.(*big.Int)
		if s == nil || n == 0 { //如果S或者N 不合法则直接跳过
			continue
		}
		tx := types.SetFloodData(ftx)
		if tx.GetTxS() == nil {
			tx.SetTxS(s)
		}
		txs = append(txs, tx)
		tmpNtx[n] = tx
	}
	nPool.getFromByTx(txs)
	nPool.mu.Lock()
	for n, tx := range tmpNtx {
		ts, ok := nPool.mapNs.Load(n)
		if !ok {
			continue
		}
		s := ts.(*big.Int)
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
			tmptx := nPool.getTxbyS(s, false)
			if tmptx != nil {
				tx = tmptx
			}
			nPool.setTxNum(tx, n, false)
			nPool.setnTx(n, tx, false)
		}
		if tx.GetTxS() == nil {
			tx.SetTxS(s)
		}
		_, err := nPool.add(tx, false)
		if err != nil && err != ErrKnownTransaction {
			log.Error("file txpool", "msg_RecvFloodTx::Error=", err)
			if _, ok := nPool.mapErrorTxs[s]; !ok {
				errorTxs = append(errorTxs, s)
				nPool.mapErrorTxs[s] = tx
				nPool.mapDelErrtxs[tx.Hash()] = s
			}
			//对于添加失败的交易要调用删除map方法
			nPool.deleteMap(tx)
		}
	}
	nPool.mu.Unlock()
	//nPool.selfmlk.Unlock()
	if len(errorTxs) > 0 {
		//TODO S在这如何进行签名？？如何获得本节点账户信息
		msData, err := json.Marshal(errorTxs)
		if err != nil {
			log.Error("function msg_RecvFloodTx", "send error Tx,json.Marshal is err:", err)
		} else {
			nPool.SendMsg(MsgStruct{Msgtype: RecvErrTx, SendAddr: nid, MsgData: msData})
			nPool.RecvErrTx(common.Address{}, errorTxs)
		}
	}
}

// 接收错误交易
func (nPool *NormalTxPool) RecvErrTx(addr common.Address, listS []*big.Int) {
	//nPool.selfmlk.Lock()
	//defer nPool.selfmlk.Unlock()
	nPool.mu.Lock()
	for _, s := range listS {
		tmptx := nPool.getTxbyS(s, false)
		if tmptx != nil {
			//如果本地有该笔交易则说明本节点认为这笔交易是对的，不做其他操作。
			continue
		}
		if _, ok := nPool.mapErrorTxs[s]; !ok {
			//如果本地没有该笔错误交易则等待。因为如果本地没有这笔交易其他节点肯定会给其洪泛的，洪泛后就会收到这笔交易
			continue
		}
		tx := nPool.mapErrorTxs[s]
		hash := tx.Hash()
		isRepeat := false
		for _, acc := range nPool.mapCaclErrtxs[hash] {
			if addr == acc { //判断当前错误交易，同一个节点是否发送过
				isRepeat = true
				break
			}
		}
		if isRepeat {
			continue
		}
		nPool.mapCaclErrtxs[hash] = append(nPool.mapCaclErrtxs[hash], addr)
		if uint64(len(nPool.mapCaclErrtxs[hash])) >= params.ErrTxConsensus {
			nPool.removeTx(hash, true)
		} else {
			nPool.addBlockTiming(hash)
		}
	}
	nPool.mu.Unlock()
}

// 刪除新增加的map中的数据
func (nPool *NormalTxPool) deleteMap(tx *types.Transaction) {
	//在调用的地方已经加锁了所以在此不用加锁
	s := tx.GetTxS()
	listn := tx.N
	for _, n := range listn {
		nPool.deletnTx(n)
		nPool.mapNs.Delete(n)
	}
	delete(nPool.mapTxsTiming, tx.Hash())
	nPool.deletsTx(s)
}

// 添加区块定时
func (nPool *NormalTxPool) addBlockTiming(hash common.Hash) {
	if _, ok := nPool.mapTxsTiming[hash]; ok {
		return
	}
	nPool.mapTxsTiming[hash] = time.Now()
}

// 20个区块定时删除(每次收到新区快头广播时触发)
func (nPool *NormalTxPool) blockTiming() {
	//外侧已经有锁在此不用再加锁
	//blockNum := nPool.chain.CurrentBlock().Number()
	listHash := make([]common.Hash, 0)
	for hash, t := range nPool.mapTxsTiming {
		if time.Since(t) > nPool.config.txTimeout {
			listHash = append(listHash, hash)
		}
	}
	if len(listHash) > 0 {
		for _, hash := range listHash {
			nPool.removeTx(hash, true)
			delete(nPool.mapCaclErrtxs, hash)
			delete(nPool.mapTxsTiming, hash)
			if s, ok := nPool.mapDelErrtxs[hash]; ok {
				delete(nPool.mapErrorTxs, s)
				delete(nPool.mapDelErrtxs, hash)
			}
		}
	}
}

// 根据交易获取交易中的from
func (nPool *NormalTxPool) getFromByTx(txs []*types.Transaction) {
	var waitG = &sync.WaitGroup{}
	maxProcs := runtime.NumCPU() //获取cpu个数
	if maxProcs >= 2 {
		runtime.GOMAXPROCS(maxProcs - 1) //限制同时运行的goroutines数量
	}
	for _, tx := range txs {
		waitG.Add(1)
		ttx := tx
		go types.Sender_self(nPool.signer, ttx, waitG)
	}
	waitG.Wait()
}

// 检查交易中是否存在from
func (nPool *NormalTxPool) checkTxFrom(tx *types.Transaction) (common.Address, error) {
	from, err := tx.GetTxFrom()
	if err == nil {
		return from, nil
	} else {
		f, err := types.Sender(nPool.signer, tx)
		if err != nil {
			return common.Address{}, ErrInvalidSender
		} else {
			return f, nil
		}
	}
}

// validateTx checks whether a transaction is valid according to the consensus
// rules and adheres to some heuristic limits of the local node (price and size).
func (nPool *NormalTxPool) validateTx(tx *types.Transaction, local bool) error {
	// add if
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
		var txsize uint64 = params.TxSize
		if uint64(tx.Size()) > txsize*txcount {
			return ErrOversizedData
		}
		if txcount > params.TxCount { //验证一对多交易最多支持1000条
			return ErrTXCountOverflow
		}
	} else {
		// Heuristic limit, reject transactions over 32KB to prevent DOS attacks
		if uint64(tx.Size()) > params.TxSize {
			return ErrOversizedData
		}
	}
	// Transactions can't be negative. This may never happen using RLP decoded
	// transactions but may occur if you create a transaction using the RPC.
	if tx.Value().Sign() < 0 {
		return ErrNegativeValue
	}

	// Ensure the transaction doesn't exceed the current block limit gas.
	if nPool.currentMaxGas < tx.Gas() {
		return ErrGasLimit
	}
	// 如果交易中已经有了from就不需要在做解签
	from, addrerr := nPool.checkTxFrom(tx)
	if addrerr != nil {
		return addrerr
	}
	// 验证当V值大于128时，如果扩展交易为空则直接丢弃该交易并返回交易不合法
	//if tx.GetTxV().Cmp(big.NewInt(128)) > 0 && len(txEx) <= 0 {
	//	return ErrTXWrongful
	//}
	// Drop non-local transactions under our own minimal accepted gas price
	gasprice, err := matrixstate.GetTxpoolGasLimit(nPool.currentState)
	if err != nil {
		return errors.New("get txpool gasPrice err")
	}
	nPool.gasPrice.Set(gasprice)
	if nPool.gasPrice.Cmp(tx.GasPrice()) > 0 {
		return ErrUnderpriced
	}
	// Ensure the transaction adheres to nonce ordering
	if nPool.currentState.GetNonce(from) > tx.Nonce() {
		return ErrNonceTooLow
	}
	// add if
	var balance *big.Int
	var entrustbalance *big.Int
	//当前账户余额
	for _, tAccount := range nPool.currentState.GetBalance(from) {
		if tAccount.AccountType == common.MainAccount {
			balance = tAccount.Balance
			break
		}
	}
	//委托账户的余额
	if tx.IsEntrustGas {
		for _, tAccount := range nPool.currentState.GetBalance(tx.AmontFrom()) {
			if tAccount.AccountType == common.MainAccount {
				entrustbalance = tAccount.Balance
				break
			}
		}
	}
	if len(txEx) > 0 && len(txEx[0].ExtraTo) > 0 {
		//如果是委托gas，检查授权人的账户余额是否大于gas;委托人的余额是否大于转账金额
		if tx.IsEntrustGas {
			totalGas := new(big.Int).Mul(tx.GasPrice(), new(big.Int).SetUint64(tx.Gas()))
			if entrustbalance.Cmp(totalGas) < 0 {
				return ErrEntrustInsufficientFunds
			}
			if balance.Cmp(tx.TotalAmount()) < 0 {
				return ErrInsufficientFunds
			}
		} else {
			if balance.Cmp(tx.CostALL()) < 0 {
				return ErrInsufficientFunds
			}
		}
	} else {
		if tx.IsEntrustGas {
			totalGas := new(big.Int).Mul(tx.GasPrice(), new(big.Int).SetUint64(tx.Gas()))
			if entrustbalance.Cmp(totalGas) < 0 {
				return ErrEntrustInsufficientFunds
			}
			if balance.Cmp(tx.Value()) < 0 {
				return ErrInsufficientFunds
			}
		} else {
			if balance.Cmp(tx.Cost()) < 0 {
				return ErrInsufficientFunds

			}
		}
	}
	intrGas, err := IntrinsicGas(tx.Data())
	if err != nil {
		return err
	}
	// add if
	if len(txEx) > 0 && len(txEx[0].ExtraTo) > 0 {
		for _, tx_list := range txEx {
			for _, txs := range tx_list.ExtraTo {
				tmpintrGas, tmperr := IntrinsicGas(txs.Payload)
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

func (nPool *NormalTxPool) add(tx *types.Transaction, local bool) (bool, error) {
	if tx.IsEntrustTx() {
		//通过from获得的数据为授权人marsha1过的数据
		from := tx.From()
		//from = base58.Base58DecodeToAddress("MAN.3oW6eUV7MmQcHiD4WGQcRnsN8ho1aFTWPaYADwnqu2wW3WcJzbEfZNw2") //******测试用，要删除
		entrustFrom := nPool.currentState.GetGasAuthFrom(from, nPool.chain.CurrentBlock().NumberU64()+1) //当前块高加1，因为该笔交易要上到下一个区块
		if !entrustFrom.Equal(common.Address{}) {
			tx.Setentrustfrom(entrustFrom)
			tx.IsEntrustGas = true
		} else {
			entrustFrom := nPool.currentState.GetGasAuthFromByTime(from, uint64(time.Now().Unix()))
			if !entrustFrom.Equal(common.Address{}) {
				tx.Setentrustfrom(entrustFrom)
				tx.IsEntrustGas = true
				tx.IsEntrustByTime = true
			} else {
				log.Error("该用户没有被授权过委托Gas")
				return false, ErrWithoutAuth
			}
		}
	}

	//普通交易
	hash := tx.Hash()
	// If the transaction is already known, discard it
	if nPool.all.Get(hash) != nil {
		log.Trace("Discarding already known transaction", "hash", hash)
		return false, ErrKnownTransaction
	}

	// If the transaction fails basic validation, discard it
	if err := nPool.validateTx(tx, local); err != nil {
		log.Trace("Discarding invalid transaction", "hash", hash, "err", err)
		invalidTxCounter.Inc(1)
		return false, err
	}
	// 池子满了之后就不再加入
	// If the transaction pool is full, discard transactions
	if uint64(nPool.all.Count()) >= nPool.config.GlobalSlots+nPool.config.GlobalQueue {
		return false, ErrTXPoolFull
	}

	// 如果交易中已经有了from就不需要在做解签
	from, addrerr := nPool.checkTxFrom(tx)
	if addrerr != nil {
		return false, addrerr
	}

	if list := nPool.pending[from]; list != nil && list.Overlaps(tx) {
		return false, ErrTXNonceSame
	}
	//将交易加入pending
	if nPool.pending[from] == nil {
		nPool.pending[from] = newTxList(false, tx.GetTxCurrency())
	}
	nPool.pending[from].Add(tx, 0)
	nPool.all.Add(tx)
	nPool.pendingState.SetNonce(from, tx.Nonce()+1)
	selfRole := ca.GetRole()
	if selfRole == common.RoleMiner || selfRole == common.RoleValidator {
		tx_s := tx.GetTxS()
		nPool.setsTx(tx_s, tx)
		if len(tx.N) == 0 {
			log.Trace("txpool:add()", "gSendst.notice", "")
			gSendst.notice <- tx.GetTxS()
		} else {
			log.Trace("txpool:add()", "gSendst.notice::tx N ", tx.N)
		}
	} else if selfRole == common.RoleDefault || selfRole == common.RoleBucket {
		promoted := make([]types.SelfTransaction, 0)
		promoted = append(promoted, tx)
		nPool.sendTxCh <- NewTxsEvent{promoted, types.NormalTxIndex}
		//nPool.txFeed.Send(NewTxsEvent{promoted, types.NormalTxIndex})
		log.Trace("txpool:add()", "selfRole == common.RoleDefault", selfRole)
	} else {
		log.Trace("txpool:add()", "unknown selfRole ", selfRole)
	}
	return true, nil
}

// AddLocal enqueues a single transaction into the pool if it is valid, marking
// the sender as a local one in the mean time, ensuring it goes around the local
// pricing constraints.
func (nPool *NormalTxPool) AddTxPool(txer types.SelfTransaction) error {
	txs := make([]*types.Transaction, 0)
	tx := txer.(*types.Transaction)
	txs = append(txs, tx)
	err := nPool.addTxs(txs, false)
	return err
}

// addTxs attempts to queue a batch of transactions if they are valid.
func (nPool *NormalTxPool) addTxs(txs []*types.Transaction, local bool) error {
	//nPool.selfmlk.Lock()
	nPool.getFromByTx(txs) //
	//nPool.selfmlk.Unlock()
	nPool.mu.Lock()
	err := nPool.addTxsLocked(txs, local)
	nPool.mu.Unlock()
	return err
}

// addTxsLocked attempts to queue a batch of transactions if they are valid,
// whilst assuming the transaction pool lock is already held.
func (nPool *NormalTxPool) addTxsLocked(txs []*types.Transaction, local bool) (err error) {
	//errs := make([]error, len(txs))
	for _, tx := range txs {
		_, err = nPool.add(tx, local)
	}
	return err
}

// Status returns the status (unknown/pending/queued) of a batch of transactions
// identified by their hashes.
func (nPool *NormalTxPool) Status(hashes []common.Hash) []TxStatus {
	nPool.mu.RLock()
	defer nPool.mu.RUnlock()
	status := make([]TxStatus, len(hashes))
	for i, hash := range hashes {
		if tx := nPool.all.Get(hash); tx != nil {
			// 如果交易中已经有了from就不需要在做解签
			from, _ := nPool.checkTxFrom(tx)
			if nPool.pending[from] != nil && nPool.pending[from].txs[tx.GetTxCurrency()].items[tx.Nonce()] != nil {
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
func (nPool *NormalTxPool) Get(hash common.Hash) *types.Transaction {
	tx := nPool.all.Get(hash)
	return tx
}

// removeTx removes a single transaction from the queue, moving all subsequent
// transactions back to the future queue.
func (nPool *NormalTxPool) removeTx(hash common.Hash, outofbound bool) {
	// Fetch the transaction we wish to delete
	tx := nPool.all.Get(hash)
	if tx == nil {
		return
	}
	// 如果交易中已经有了from就不需要在做解签
	addr, _ := nPool.checkTxFrom(tx)

	// Remove it from the list of known transactions
	nPool.all.Remove(hash)

	// ========begin=========
	nPool.deleteMap(tx)
	//===========end===========
	// Remove the transaction from the pending lists and reset the account nonce
	if pending := nPool.pending[addr]; pending != nil {
		if removed, _ := pending.Remove(tx); removed {
			// If no more pending transactions are left, remove the list
			if pending.Empty(tx.GetTxCurrency()) {
				delete(nPool.pending, addr)
			}
			// Update the account nonce if needed
			if nonce := tx.Nonce(); nPool.pendingState.GetNonce(addr) > nonce {
				nPool.pendingState.SetNonce(addr, nonce)
			}
			return
		}
	}
}

// demoteUnexecutables removes invalid and processed transactions from the pools
// executable/pending queue and any subsequent transactions that become unexecutable
// are moved back into the future queue.
func (nPool *NormalTxPool) DemoteUnexecutables() {
	// Iterate over all accounts and demote any non-executable transactions
	for addr, list := range nPool.pending {
		for typ, txs := range list.txs {
			nonce := nPool.currentState.GetNonce(addr)
			// Drop all transactions that are deemed too old (low nonce)
			for _, tx := range txs.Forward(nonce) {
				// ========begin=========
				nPool.deleteMap(tx)
				//===========end===========
				hash := tx.Hash()
				//log.Trace("Removed old pending transaction", "hash", hash)
				nPool.all.Remove(hash)
				//nPool.priced.Removed()
			}
			// Drop all transactions that are too costly (low balance or out of gas), and queue any invalids back for later
			tBalance := new(big.Int)
			for _, tAccount := range nPool.currentState.GetBalance(addr) {
				if tAccount.AccountType == common.MainAccount {
					tBalance = tAccount.Balance
					break
				}
			}
			drops, _ := list.Filter(tBalance, nPool.currentMaxGas, typ)
			for _, tx := range drops {
				// ========begin=========
				nPool.deleteMap(tx)
				//===========end===========
				hash := tx.Hash()
				log.Trace("Removed unpayable pending transaction", "hash", hash)
				nPool.all.Remove(hash)
				//nPool.priced.Removed()
				pendingNofundsCounter.Inc(1)
			}
			// Delete the entire queue entry if it became empty.
			if list.Empty(typ) {
				delete(nPool.pending, addr)
			}
		}
	}
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
	tx := t.all[hash]
	return tx
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
	hash := tx.Hash()
	//log.Info("file tx_pool", "all.Add()", hash.String())
	t.all[hash] = tx
}

// Remove removes a transaction from the lookup.
func (t *txLookup) Remove(hash common.Hash) {
	t.lock.Lock()
	defer t.lock.Unlock()

	delete(t.all, hash)
}
