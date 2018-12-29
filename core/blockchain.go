// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

// Package core implements the Matrix consensus protocol.
package core

import (
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	mrand "math/rand"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/matrix/go-matrix/common/readstatedb"

	"github.com/hashicorp/golang-lru"
	"gopkg.in/karalabe/cookiejar.v2/collections/prque"

	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/common/mclock"
	"github.com/matrix/go-matrix/consensus"
	"github.com/matrix/go-matrix/consensus/mtxdpos"
	"github.com/matrix/go-matrix/core/matrixstate"
	"github.com/matrix/go-matrix/core/rawdb"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/core/vm"
	"github.com/matrix/go-matrix/crypto"
	"github.com/matrix/go-matrix/event"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mandb"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/metrics"
	"github.com/matrix/go-matrix/params"
	"github.com/matrix/go-matrix/params/manparams"

	"github.com/matrix/go-matrix/rlp"
	"github.com/matrix/go-matrix/trie"
	"github.com/pkg/errors"
	//"github.com/matrix/go-matrix/baseinterface"
)

var (
	blockInsertTimer = metrics.NewRegisteredTimer("chain/inserts", nil)

	ErrNoGenesis = errors.New("Genesis not found in chain")
)

const (
	bodyCacheLimit      = 256
	blockCacheLimit     = 256
	maxFutureBlocks     = 256
	maxTimeFutureBlocks = 30
	badBlockLimit       = 10
	triesInMemory       = 128

	// BlockChainVersion ensures that an incompatible database forces a resync from scratch.
	BlockChainVersion = 3
	ModuleName        = "blockchain"
)

// CacheConfig contains the configuration values for the trie caching/pruning
// that's resident in a blockchain.
type CacheConfig struct {
	Disabled      bool          // Whether to disable trie write caching (archive node)
	TrieNodeLimit int           // Memory limit (MB) at which to flush the current in-memory trie to disk
	TrieTimeLimit time.Duration // Time limit after which to flush the current in-memory trie to disk
}

// BlockChain represents the canonical chain given a database with a genesis
// block. The Blockchain manages chain imports, reverts, chain reorganisations.
//
// Importing blocks in to the block chain happens according to the set of rules
// defined by the two stage Validator. Processing of blocks is done using the
// Processor which processes the included transaction. The validation of the state
// is done in the second part of the Validator. Failing results in aborting of
// the import.
//
// The BlockChain also helps in returning blocks from **any** chain included
// in the database as well as blocks that represents the canonical chain. It's
// important to note that GetBlock can return any block and does not need to be
// included in the canonical one where as GetBlockByNumber always represents the
// canonical chain.
type BlockChain struct {
	chainConfig *params.ChainConfig // Chain & network configuration
	cacheConfig *CacheConfig        // Cache configuration for pruning

	db     mandb.Database // Low level persistent database to store final content in
	triegc *prque.Prque   // Priority queue mapping block numbers to tries to gc
	gcproc time.Duration  // Accumulates canonical block processing for trie dumping

	hc            *HeaderChain
	rmLogsFeed    event.Feed
	chainFeed     event.Feed
	chainSideFeed event.Feed
	chainHeadFeed event.Feed
	logsFeed      event.Feed
	scope         event.SubscriptionScope
	genesisBlock  *types.Block

	mu      sync.RWMutex // global mutex for locking chain operations
	chainmu sync.RWMutex // blockchain insertion lock
	procmu  sync.RWMutex // block processor lock

	checkpoint       int          // checkpoint counts towards the new checkpoint
	currentBlock     atomic.Value // Current head of the block chain
	currentFastBlock atomic.Value // Current head of the fast-sync chain (may be above the block chain!)

	stateCache   state.Database // State database to reuse between imports (contains state cache)
	bodyCache    *lru.Cache     // Cache for the most recent block bodies
	bodyRLPCache *lru.Cache     // Cache for the most recent block bodies in RLP encoded format
	blockCache   *lru.Cache     // Cache for the most recent entire blocks
	futureBlocks *lru.Cache     // future blocks are blocks added for later processing

	quit    chan struct{} // blockchain quit channel
	running int32         // running must be called atomically
	// procInterrupt must be atomically called
	procInterrupt int32          // interrupt signaler for block processing
	wg            sync.WaitGroup // chain processing wait group for shutting down

	engine     consensus.Engine
	dposEngine consensus.DPOSEngine
	processor  Processor // block processor interface
	validator  Validator // block and state validator interface
	vmConfig   vm.Config

	badBlocks *lru.Cache // Bad block cache
	msgceter  *mc.Center
	//lb ipfs
	bBlockSendIpfs bool
	qBlockQueue    *prque.Prque

	//matrix state
	matrixState *matrixstate.MatrixState
	graphStore  *matrixstate.GraphStore
}

// NewBlockChain returns a fully initialised block chain using information
// available in the database. It initialises the default Matrix Validator and
// Processor.
func NewBlockChain(db mandb.Database, cacheConfig *CacheConfig, chainConfig *params.ChainConfig, engine consensus.Engine, vmConfig vm.Config) (*BlockChain, error) {
	if cacheConfig == nil {
		cacheConfig = &CacheConfig{
			TrieNodeLimit: 256 * 1024 * 1024,
			TrieTimeLimit: 5 * time.Minute,
		}
	}
	bodyCache, _ := lru.New(bodyCacheLimit)
	bodyRLPCache, _ := lru.New(bodyCacheLimit)
	blockCache, _ := lru.New(blockCacheLimit)
	futureBlocks, _ := lru.New(maxFutureBlocks)
	badBlocks, _ := lru.New(badBlockLimit)

	bc := &BlockChain{
		chainConfig:  chainConfig,
		cacheConfig:  cacheConfig,
		db:           db,
		triegc:       prque.New(),
		stateCache:   state.NewDatabase(db),
		quit:         make(chan struct{}),
		bodyCache:    bodyCache,
		bodyRLPCache: bodyRLPCache,
		blockCache:   blockCache,
		futureBlocks: futureBlocks,
		engine:       engine,
		vmConfig:     vmConfig,
		badBlocks:    badBlocks,
		matrixState:  matrixstate.NewMatrixState(),
	}
	bc.graphStore = matrixstate.NewGraphStore(bc)

	bc.SetValidator(NewBlockValidator(chainConfig, bc, engine))
	bc.SetProcessor(NewStateProcessor(chainConfig, bc, engine))

	bc.dposEngine = mtxdpos.NewMtxDPOS()

	var err error
	err = bc.RegisterMatrixStateDataProducer(mc.MSKeyTopologyGraph, bc.graphStore.ProduceTopologyStateData)
	if err != nil {
		return nil, err
	}
	err = bc.RegisterMatrixStateDataProducer(mc.MSKeyBroadcastInterval, ProduceBroadcastIntervalData)
	if err != nil {
		return nil, err
	}

	bc.hc, err = NewHeaderChain(db, chainConfig, engine, bc.dposEngine, bc.getProcInterrupt)
	if err != nil {
		return nil, err
	}
	bc.genesisBlock = bc.GetBlockByNumber(0)
	if bc.genesisBlock == nil {
		return nil, ErrNoGenesis
	}
	err = bc.DPOSEngine().VerifyVersion(bc, bc.genesisBlock.Header())
	if err != nil {
		return nil, err
	}
	if err := bc.loadLastState(); err != nil {
		return nil, err
	}
	// Check the current state of the block hashes and make sure that we do not have any of the bad blocks in our chain
	for hash := range BadHashes {
		if header := bc.GetHeaderByHash(hash); header != nil {
			// get the canonical block corresponding to the offending header's number
			headerByNumber := bc.GetHeaderByNumber(header.Number.Uint64())
			// make sure the headerByNumber (if present) is in our current canonical chain
			if headerByNumber != nil && headerByNumber.Hash() == header.Hash() {
				log.Error("Found bad hash, rewinding chain", "number", header.Number, "hash", header.ParentHash)
				bc.SetHead(header.Number.Uint64() - 1)
				log.Error("Chain rewind was successful, resuming normal operation")
			}
		}
	}

	reqCh := make(chan struct{})
	sub, err := mc.SubscribeEvent(mc.CA_ReqCurrentBlock, reqCh)
	if err == nil {
		go func(chain *BlockChain, reqCh chan struct{}, sub event.Subscription) {
			time.Sleep(3 * time.Second)
			select {
			case <-reqCh:
				block := chain.CurrentBlock()
				num := block.Number().Uint64()
				log.INFO("MAIN", "本地区块插入消息已发送", num, "hash", block.Hash())
				mc.PublishEvent(mc.NewBlockMessage, block)
				sub.Unsubscribe()
				close(reqCh)
				return
			}
		}(bc, reqCh, sub)
	} else {
		log.ERROR(ModuleName, "订阅CA请求当前区块事件失败", err)
	}

	manparams.SetStateReader(bc)

	// Take ownership of this particular state
	go bc.update()
	return bc, nil
}

func (bc *BlockChain) getProcInterrupt() bool {
	return atomic.LoadInt32(&bc.procInterrupt) == 1
}

// loadLastState loads the last known chain state from the database. This method
// assumes that the chain manager mutex is held.
func (bc *BlockChain) loadLastState() error {
	// Restore the last known head block
	head := rawdb.ReadHeadBlockHash(bc.db)
	if head == (common.Hash{}) {
		// Corrupt or empty database, init from scratch
		log.Warn("Empty database, resetting chain")
		return bc.Reset()
	}
	// Make sure the entire head block is available
	currentBlock := bc.GetBlockByHash(head)
	if currentBlock == nil {
		// Corrupt or empty database, init from scratch
		log.Warn("Head block missing, resetting chain", "hash", head)
		return bc.Reset()
	}
	// Make sure the state associated with the block is available
	if _, err := state.New(currentBlock.Root(), bc.stateCache); err != nil {
		log.INFO("Get State Err", "root", currentBlock.Root().TerminalString(), "err", err)
		// Dangling block without a state associated, init from scratch
		log.Warn("Head state missing, repairing chain", "number", currentBlock.Number(), "hash", currentBlock.Hash())
		if err := bc.repair(&currentBlock); err != nil {
			return err
		}
	}
	// Everything seems to be fine, set as the head block
	bc.currentBlock.Store(currentBlock)

	// Restore the last known head header
	currentHeader := currentBlock.Header()
	if head := rawdb.ReadHeadHeaderHash(bc.db); head != (common.Hash{}) {
		if header := bc.GetHeaderByHash(head); header != nil {
			currentHeader = header
		}
	}
	bc.hc.SetCurrentHeader(currentHeader)

	// Restore the last known head fast block
	bc.currentFastBlock.Store(currentBlock)
	if head := rawdb.ReadHeadFastBlockHash(bc.db); head != (common.Hash{}) {
		if block := bc.GetBlockByHash(head); block != nil {
			bc.currentFastBlock.Store(block)
		}
	}

	// Issue a status log for the user
	currentFastBlock := bc.CurrentFastBlock()

	headerTd := bc.GetTd(currentHeader.Hash(), currentHeader.Number.Uint64())
	blockTd := bc.GetTd(currentBlock.Hash(), currentBlock.NumberU64())
	fastTd := bc.GetTd(currentFastBlock.Hash(), currentFastBlock.NumberU64())

	log.Info("Loaded most recent local header", "number", currentHeader.Number, "hash", currentHeader.Hash(), "td", headerTd)
	log.Info("Loaded most recent local full block", "number", currentBlock.Number(), "hash", currentBlock.Hash(), "td", blockTd)
	log.Info("Loaded most recent local fast block", "number", currentFastBlock.Number(), "hash", currentFastBlock.Hash(), "td", fastTd)

	return nil
}

// SetHead rewinds the local chain to a new head. In the case of headers, everything
// above the new head will be deleted and the new one set. In the case of blocks
// though, the head may be further rewound if block bodies are missing (non-archive
// nodes after a fast sync).
func (bc *BlockChain) SetHead(head uint64) error {
	log.Warn("Rewinding blockchain", "target", head)

	bc.mu.Lock()
	defer bc.mu.Unlock()

	// Rewind the header chain, deleting all block bodies until then
	delFn := func(hash common.Hash, num uint64) {
		rawdb.DeleteBody(bc.db, hash, num)
	}
	bc.hc.SetHead(head, delFn)
	currentHeader := bc.hc.CurrentHeader()

	// Clear out any stale content from the caches
	bc.bodyCache.Purge()
	bc.bodyRLPCache.Purge()
	bc.blockCache.Purge()
	bc.futureBlocks.Purge()

	// Rewind the block chain, ensuring we don't end up with a stateless head block
	if currentBlock := bc.CurrentBlock(); currentBlock != nil && currentHeader.Number.Uint64() < currentBlock.NumberU64() {
		bc.currentBlock.Store(bc.GetBlock(currentHeader.Hash(), currentHeader.Number.Uint64()))
	}
	if currentBlock := bc.CurrentBlock(); currentBlock != nil {
		if _, err := state.New(currentBlock.Root(), bc.stateCache); err != nil {
			// Rewound state missing, rolled back to before pivot, reset to genesis
			bc.currentBlock.Store(bc.genesisBlock)
		}
	}
	// Rewind the fast block in a simpleton way to the target head
	if currentFastBlock := bc.CurrentFastBlock(); currentFastBlock != nil && currentHeader.Number.Uint64() < currentFastBlock.NumberU64() {
		bc.currentFastBlock.Store(bc.GetBlock(currentHeader.Hash(), currentHeader.Number.Uint64()))
	}
	// If either blocks reached nil, reset to the genesis state
	if currentBlock := bc.CurrentBlock(); currentBlock == nil {
		bc.currentBlock.Store(bc.genesisBlock)
	}
	if currentFastBlock := bc.CurrentFastBlock(); currentFastBlock == nil {
		bc.currentFastBlock.Store(bc.genesisBlock)
	}
	currentBlock := bc.CurrentBlock()
	currentFastBlock := bc.CurrentFastBlock()

	rawdb.WriteHeadBlockHash(bc.db, currentBlock.Hash())
	rawdb.WriteHeadFastBlockHash(bc.db, currentFastBlock.Hash())

	return bc.loadLastState()
}

// FastSyncCommitHead sets the current head block to the one defined by the hash
// irrelevant what the chain contents were prior.
func (bc *BlockChain) FastSyncCommitHead(hash common.Hash) error {
	// Make sure that both the block as well at its state trie exists
	block := bc.GetBlockByHash(hash)
	if block == nil {
		return fmt.Errorf("non existent block [%x…]", hash[:4])
	}
	if _, err := trie.NewSecure(block.Root(), bc.stateCache.TrieDB(), 0); err != nil {
		return err
	}
	// If all checks out, manually set the head block
	bc.mu.Lock()
	bc.currentBlock.Store(block)
	bc.mu.Unlock()

	log.Info("Committed new head block", "number", block.Number(), "hash", hash)
	return nil
}

// GasLimit returns the gas limit of the current HEAD block.
func (bc *BlockChain) GasLimit() uint64 {
	return bc.CurrentBlock().GasLimit()
}

// CurrentBlock retrieves the current head block of the canonical chain. The
// block is retrieved from the blockchain's internal cache.
func (bc *BlockChain) CurrentBlock() *types.Block {
	x := bc.currentBlock.Load()
	if x == nil {
		return nil
	}
	return x.(*types.Block)
}

// CurrentFastBlock retrieves the current fast-sync head block of the canonical
// chain. The block is retrieved from the blockchain's internal cache.
func (bc *BlockChain) CurrentFastBlock() *types.Block {
	x := bc.currentFastBlock.Load()
	if x == nil {
		return nil
	}
	return x.(*types.Block)
}

// SetProcessor sets the processor required for making state modifications.
func (bc *BlockChain) SetProcessor(processor Processor) {
	bc.procmu.Lock()
	defer bc.procmu.Unlock()
	bc.processor = processor
}

// SetValidator sets the validator which is used to validate incoming blocks.
func (bc *BlockChain) SetValidator(validator Validator) {
	bc.procmu.Lock()
	defer bc.procmu.Unlock()
	bc.validator = validator
}

// Validator returns the current validator.
func (bc *BlockChain) Validator() Validator {
	bc.procmu.RLock()
	defer bc.procmu.RUnlock()
	return bc.validator
}

// Processor returns the current processor.
func (bc *BlockChain) Processor() Processor {
	bc.procmu.RLock()
	defer bc.procmu.RUnlock()
	return bc.processor
}

// State returns a new mutable state based on the current HEAD block.
func (bc *BlockChain) State() (*state.StateDB, error) {
	return bc.StateAt(bc.CurrentBlock().Root())
}

// StateAt returns a new mutable state based on a particular point in time.
func (bc *BlockChain) StateAt(root common.Hash) (*state.StateDB, error) {
	return state.New(root, bc.stateCache)
}

func (bc *BlockChain) GetStateByHash(hash common.Hash) (*state.StateDB, error) {
	block := bc.GetBlockByHash(hash)
	if block == nil {
		return nil, errors.New("can't find block by hash")
	}
	return bc.StateAt(block.Root())
}

// Reset purges the entire blockchain, restoring it to its genesis state.
func (bc *BlockChain) Reset() error {
	return bc.ResetWithGenesisBlock(bc.genesisBlock)
}

// ResetWithGenesisBlock purges the entire blockchain, restoring it to the
// specified genesis state.
func (bc *BlockChain) ResetWithGenesisBlock(genesis *types.Block) error {
	// Dump the entire block chain and purge the caches
	if err := bc.SetHead(0); err != nil {
		return err
	}
	bc.mu.Lock()
	defer bc.mu.Unlock()

	// Prepare the genesis block and reinitialise the chain
	if err := bc.hc.WriteTd(genesis.Hash(), genesis.NumberU64(), genesis.Difficulty()); err != nil {
		log.Crit("Failed to write genesis block TD", "err", err)
	}
	rawdb.WriteBlock(bc.db, genesis)

	bc.genesisBlock = genesis
	bc.insert(bc.genesisBlock)
	bc.currentBlock.Store(bc.genesisBlock)
	bc.hc.SetGenesis(bc.genesisBlock.Header())
	bc.hc.SetCurrentHeader(bc.genesisBlock.Header())
	bc.currentFastBlock.Store(bc.genesisBlock)

	return nil
}

// repair tries to repair the current blockchain by rolling back the current block
// until one with associated state is found. This is needed to fix incomplete db
// writes caused either by crashes/power outages, or simply non-committed tries.
//
// This method only rolls back the current block. The current header and current
// fast block are left intact.
func (bc *BlockChain) repair(head **types.Block) error {
	for {
		// Abort if we've rewound to a head block that does have associated state
		if _, err := state.New((*head).Root(), bc.stateCache); err == nil {
			log.Info("Rewound blockchain to past state", "number", (*head).Number(), "hash", (*head).Hash())
			return nil
		}
		// Otherwise rewind one block and recheck state availability there
		(*head) = bc.GetBlock((*head).ParentHash(), (*head).NumberU64()-1)
	}
}

// Export writes the active chain to the given writer.
func (bc *BlockChain) Export(w io.Writer) error {
	return bc.ExportN(w, uint64(0), bc.CurrentBlock().NumberU64())
}

// ExportN writes a subset of the active chain to the given writer.
func (bc *BlockChain) ExportN(w io.Writer, first uint64, last uint64) error {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	if first > last {
		return fmt.Errorf("export failed: first (%d) is greater than last (%d)", first, last)
	}
	log.Info("Exporting batch of blocks", "count", last-first+1)

	for nr := first; nr <= last; nr++ {
		block := bc.GetBlockByNumber(nr)
		if block == nil {
			return fmt.Errorf("export failed on #%d: not found", nr)
		}

		if err := block.EncodeRLP(w); err != nil {
			return err
		}
	}

	return nil
}

// insert injects a new head block into the current block chain. This method
// assumes that the block is indeed a true head. It will also reset the head
// header and the head fast sync block to this very same block if they are older
// or if they are on a different side chain.
//
// Note, this function assumes that the `mu` mutex is held!
func (bc *BlockChain) insert(block *types.Block) {
	// If the block is on a side chain or an unknown one, force other heads onto it too
	var updateHeads bool
	if block.IsSuperBlock() {
		currentblock := bc.GetBlockByHash(bc.GetCurrentHash())

		if currentblock.NumberU64() > block.NumberU64() {
			log.INFO(ModuleName, "rewind to", block.NumberU64()-1)
			bc.bodyCache.Purge()
			bc.bodyRLPCache.Purge()
			bc.blockCache.Purge()
			bc.futureBlocks.Purge()
			delFn := func(hash common.Hash, num uint64) {
				rawdb.DeleteBody(bc.db, hash, num)
			}
			bc.hc.SetHead(block.NumberU64()-1, delFn)
		}

		updateHeads = true
	} else {
		updateHeads = rawdb.ReadCanonicalHash(bc.db, block.NumberU64()) != block.Hash()
	}

	// Add the block to the canonical chain number scheme and mark as the head
	rawdb.WriteCanonicalHash(bc.db, block.Hash(), block.NumberU64())
	rawdb.WriteHeadBlockHash(bc.db, block.Hash())

	bc.currentBlock.Store(block)

	// If the block is better than our head or is on a different chain, force update heads
	if updateHeads {
		bc.hc.SetCurrentHeader(block.Header())
		rawdb.WriteHeadFastBlockHash(bc.db, block.Hash())

		bc.currentFastBlock.Store(block)
	}
}

// Genesis retrieves the chain's genesis block.
func (bc *BlockChain) Genesis() *types.Block {
	return bc.genesisBlock
}

// GetBody retrieves a block body (transactions and uncles) from the database by
// hash, caching it if found.
func (bc *BlockChain) GetBody(hash common.Hash) *types.Body {
	// Short circuit if the body's already in the cache, retrieve otherwise
	if cached, ok := bc.bodyCache.Get(hash); ok {
		body := cached.(*types.Body)
		return body
	}
	number := bc.hc.GetBlockNumber(hash)
	if number == nil {
		return nil
	}
	body := rawdb.ReadBody(bc.db, hash, *number)
	if body == nil {
		return nil
	}
	// Cache the found body for next time and return
	bc.bodyCache.Add(hash, body)
	return body
}

// GetBodyRLP retrieves a block body in RLP encoding from the database by hash,
// caching it if found.
func (bc *BlockChain) GetBodyRLP(hash common.Hash) rlp.RawValue {
	// Short circuit if the body's already in the cache, retrieve otherwise
	if cached, ok := bc.bodyRLPCache.Get(hash); ok {
		return cached.(rlp.RawValue)
	}
	number := bc.hc.GetBlockNumber(hash)
	if number == nil {
		return nil
	}
	body := rawdb.ReadBodyRLP(bc.db, hash, *number)
	if len(body) == 0 {
		return nil
	}
	// Cache the found body for next time and return
	bc.bodyRLPCache.Add(hash, body)
	return body
}

// HasBlock checks if a block is fully present in the database or not.
func (bc *BlockChain) HasBlock(hash common.Hash, number uint64) bool {
	if bc.blockCache.Contains(hash) {
		return true
	}
	return rawdb.HasBody(bc.db, hash, number)
}

// HasState checks if state trie is fully present in the database or not.
func (bc *BlockChain) HasState(hash common.Hash) bool {
	_, err := bc.stateCache.OpenTrie(hash)
	return err == nil
}

// HasBlockAndState checks if a block and associated state trie is fully present
// in the database or not, caching it if present.
func (bc *BlockChain) HasBlockAndState(hash common.Hash, number uint64) bool {
	// Check first that the block itself is known
	block := bc.GetBlock(hash, number)
	if block == nil {
		return false
	}
	return bc.HasState(block.Root())
}

// GetBlock retrieves a block from the database by hash and number,
// caching it if found.
func (bc *BlockChain) GetBlock(hash common.Hash, number uint64) *types.Block {
	// Short circuit if the block's already in the cache, retrieve otherwise
	if block, ok := bc.blockCache.Get(hash); ok {
		return block.(*types.Block)
	}
	block := rawdb.ReadBlock(bc.db, hash, number)
	if block == nil {
		return nil
	}
	// Cache the found block for next time and return
	bc.blockCache.Add(block.Hash(), block)
	return block
}

// GetBlockByHash retrieves a block from the database by hash, caching it if found.
func (bc *BlockChain) GetBlockByHash(hash common.Hash) *types.Block {
	number := bc.hc.GetBlockNumber(hash)
	if number == nil {
		return nil
	}
	return bc.GetBlock(hash, *number)
}

// GetBlockByNumber retrieves a block from the database by number, caching it
// (associated with its hash) if found.
func (bc *BlockChain) GetBlockByNumber(number uint64) *types.Block {
	hash := rawdb.ReadCanonicalHash(bc.db, number)
	if hash == (common.Hash{}) {
		return nil
	}
	return bc.GetBlock(hash, number)
}

// GetReceiptsByHash retrieves the receipts for all transactions in a given block.
func (bc *BlockChain) GetReceiptsByHash(hash common.Hash) types.Receipts {
	number := rawdb.ReadHeaderNumber(bc.db, hash)
	if number == nil {
		return nil
	}
	return rawdb.ReadReceipts(bc.db, hash, *number)
}

// GetBlocksFromHash returns the block corresponding to hash and up to n-1 ancestors.
// [deprecated by man/62]
func (bc *BlockChain) GetBlocksFromHash(hash common.Hash, n int) (blocks []*types.Block) {
	number := bc.hc.GetBlockNumber(hash)
	if number == nil {
		return nil
	}
	for i := 0; i < n; i++ {
		block := bc.GetBlock(hash, *number)
		if block == nil {
			break
		}
		blocks = append(blocks, block)
		hash = block.ParentHash()
		*number--
	}
	return
}

// GetUnclesInChain retrieves all the uncles from a given block backwards until
// a specific distance is reached.
func (bc *BlockChain) GetUnclesInChain(block *types.Block, length int) []*types.Header {
	uncles := []*types.Header{}
	for i := 0; block != nil && i < length; i++ {
		uncles = append(uncles, block.Uncles()...)
		block = bc.GetBlock(block.ParentHash(), block.NumberU64()-1)
	}
	return uncles
}

// TrieNode retrieves a blob of data associated with a trie node (or code hash)
// either from ephemeral in-memory cache, or from persistent storage.
func (bc *BlockChain) TrieNode(hash common.Hash) ([]byte, error) {
	return bc.stateCache.TrieDB().Node(hash)
}

// Stop stops the blockchain service. If any imports are currently in progress
// it will abort them using the procInterrupt.
func (bc *BlockChain) Stop() {
	if !atomic.CompareAndSwapInt32(&bc.running, 0, 1) {
		return
	}
	// Unsubscribe all subscriptions registered from blockchain
	bc.scope.Close()
	close(bc.quit)
	atomic.StoreInt32(&bc.procInterrupt, 1)

	bc.wg.Wait()

	// Ensure the state of a recent block is also stored to disk before exiting.
	// We're writing three different states to catch different restart scenarios:
	//  - HEAD:     So we don't need to reprocess any blocks in the general case
	//  - HEAD-1:   So we don't do large reorgs if our HEAD becomes an uncle
	//  - HEAD-127: So we have a hard limit on the number of blocks reexecuted
	if !bc.cacheConfig.Disabled {
		triedb := bc.stateCache.TrieDB()

		for _, offset := range []uint64{0, 1, triesInMemory - 1} {
			if number := bc.CurrentBlock().NumberU64(); number > offset {
				recent := bc.GetBlockByNumber(number - offset)

				log.Info("Writing cached state to disk", "block", recent.Number(), "hash", recent.Hash(), "root", recent.Root())
				if err := triedb.Commit(recent.Root(), true); err != nil {
					log.Error("Failed to commit recent state trie", "err", err)
				}
			}
		}
		for !bc.triegc.Empty() {
			triedb.Dereference(bc.triegc.PopItem().(common.Hash), common.Hash{})
		}
		if size := triedb.Size(); size != 0 {
			log.Error("Dangling trie nodes after full cleanup")
		}
	}
	log.Info("Blockchain manager stopped")
}

func (bc *BlockChain) procFutureBlocks() {
	blocks := make([]*types.Block, 0, bc.futureBlocks.Len())
	for _, hash := range bc.futureBlocks.Keys() {
		if block, exist := bc.futureBlocks.Peek(hash); exist {
			blocks = append(blocks, block.(*types.Block))
		}
	}
	if len(blocks) > 0 {
		types.BlockBy(types.Number).Sort(blocks)

		// Insert one by one as chain insertion needs contiguous ancestry between blocks
		for i := range blocks {
			bc.InsertChain(blocks[i : i+1])
		}
	}
}

// WriteStatus status of write
type WriteStatus byte

const (
	NonStatTy WriteStatus = iota
	CanonStatTy
	SideStatTy
)

// Rollback is designed to remove a chain of links from the database that aren't
// certain enough to be valid.
func (bc *BlockChain) Rollback(chain []common.Hash) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	for i := len(chain) - 1; i >= 0; i-- {
		hash := chain[i]

		currentHeader := bc.hc.CurrentHeader()
		if currentHeader.Hash() == hash {
			bc.hc.SetCurrentHeader(bc.GetHeader(currentHeader.ParentHash, currentHeader.Number.Uint64()-1))
		}
		if currentFastBlock := bc.CurrentFastBlock(); currentFastBlock.Hash() == hash {
			newFastBlock := bc.GetBlock(currentFastBlock.ParentHash(), currentFastBlock.NumberU64()-1)
			bc.currentFastBlock.Store(newFastBlock)
			rawdb.WriteHeadFastBlockHash(bc.db, newFastBlock.Hash())
		}
		if currentBlock := bc.CurrentBlock(); currentBlock.Hash() == hash {
			newBlock := bc.GetBlock(currentBlock.ParentHash(), currentBlock.NumberU64()-1)
			bc.currentBlock.Store(newBlock)
			rawdb.WriteHeadBlockHash(bc.db, newBlock.Hash())
		}
	}
}

// SetReceiptsData computes all the non-consensus fields of the receipts
func SetReceiptsData(config *params.ChainConfig, block *types.Block, receipts types.Receipts) error {
	if block.IsSuperBlock() {
		return nil
	}
	signer := types.MakeSigner(config, block.Number())

	transactions, logIndex := block.Transactions(), uint(0)
	if len(transactions) != len(receipts) {
		return errors.New("transaction and receipt count mismatch")
	}

	for j := 0; j < len(receipts); j++ {
		// The transaction hash can be retrieved from the transaction itself
		receipts[j].TxHash = transactions[j].Hash()

		// The contract address can be derived from the transaction itself
		if transactions[j].To() == nil {
			// Deriving the signer is expensive, only do if it's actually needed
			from, _ := types.Sender(signer, transactions[j])
			receipts[j].ContractAddress = crypto.CreateAddress(from, transactions[j].Nonce())
		}
		// The used gas can be calculated based on previous receipts
		if j == 0 {
			receipts[j].GasUsed = receipts[j].CumulativeGasUsed
		} else {
			receipts[j].GasUsed = receipts[j].CumulativeGasUsed - receipts[j-1].CumulativeGasUsed
		}
		// The derived log fields can simply be set from the block and transaction
		for k := 0; k < len(receipts[j].Logs); k++ {
			receipts[j].Logs[k].BlockNumber = block.NumberU64()
			receipts[j].Logs[k].BlockHash = block.Hash()
			receipts[j].Logs[k].TxHash = receipts[j].TxHash
			receipts[j].Logs[k].TxIndex = uint(j)
			receipts[j].Logs[k].Index = logIndex
			logIndex++
		}
	}
	return nil
}

// InsertReceiptChain attempts to complete an already existing header chain with
// transaction and receipt data.
func (bc *BlockChain) InsertReceiptChain(blockChain types.Blocks, receiptChain []types.Receipts) (int, error) {
	bc.wg.Add(1)
	defer bc.wg.Done()

	// Do a sanity check that the provided chain is actually ordered and linked
	for i := 1; i < len(blockChain); i++ {
		if blockChain[i].NumberU64() != blockChain[i-1].NumberU64()+1 || blockChain[i].ParentHash() != blockChain[i-1].Hash() {
			log.Error("Non contiguous receipt insert", "number", blockChain[i].Number(), "hash", blockChain[i].Hash(), "parent", blockChain[i].ParentHash(),
				"prevnumber", blockChain[i-1].Number(), "prevhash", blockChain[i-1].Hash())
			return 0, fmt.Errorf("non contiguous insert: item %d is #%d [%x…], item %d is #%d [%x…] (parent [%x…])", i-1, blockChain[i-1].NumberU64(),
				blockChain[i-1].Hash().Bytes()[:4], i, blockChain[i].NumberU64(), blockChain[i].Hash().Bytes()[:4], blockChain[i].ParentHash().Bytes()[:4])
		}
	}

	var (
		stats = struct{ processed, ignored int32 }{}
		start = time.Now()
		bytes = 0
		batch = bc.db.NewBatch()
	)
	for i, block := range blockChain {
		receipts := receiptChain[i]
		// Short circuit insertion if shutting down or processing failed
		if atomic.LoadInt32(&bc.procInterrupt) == 1 {
			return 0, nil
		}
		// Short circuit if the owner header is unknown
		if !bc.HasHeader(block.Hash(), block.NumberU64()) {
			return i, fmt.Errorf("containing header #%d [%x…] unknown", block.Number(), block.Hash().Bytes()[:4])
		}
		// Skip if the entire data is already known
		if bc.HasBlock(block.Hash(), block.NumberU64()) {
			stats.ignored++
			continue
		}
		// Compute all the non-consensus fields of the receipts
		if err := SetReceiptsData(bc.chainConfig, block, receipts); err != nil {
			return i, fmt.Errorf("failed to set receipts data: %v", err)
		}
		// Write all the data out into the database
		rawdb.WriteBody(batch, block.Hash(), block.NumberU64(), block.Body())
		rawdb.WriteReceipts(batch, block.Hash(), block.NumberU64(), receipts)
		rawdb.WriteTxLookupEntries(batch, block)
		//lb
		if bc.bBlockSendIpfs && bc.qBlockQueue != nil {
			tmpBlock := &types.BlockAllSt{Sblock: block, SReceipt: receipts}
			//copy(tmpBlock.SReceipt, receipts)
			tmpBlock.SReceipt = receipts
			tmpBlock.Pading = uint64(len(block.Body().Transactions))
			bc.qBlockQueue.Push(tmpBlock, -float32(block.NumberU64()))
			log.Trace("BlockChain InsertReceiptChain ipfs save block data", "block", block.NumberU64())
			//bc.qBlockQueue.Push(block, -float32(block.NumberU64()))
		}
		stats.processed++

		if batch.ValueSize() >= mandb.IdealBatchSize {
			if err := batch.Write(); err != nil {
				return 0, err
			}
			bytes += batch.ValueSize()
			batch.Reset()
		}
	}
	if batch.ValueSize() > 0 {
		bytes += batch.ValueSize()
		if err := batch.Write(); err != nil {
			return 0, err
		}
	}

	// Update the head fast sync block if better
	bc.mu.Lock()
	head := blockChain[len(blockChain)-1]
	if td := bc.GetTd(head.Hash(), head.NumberU64()); td != nil { // Rewind may have occurred, skip in that case
		currentFastBlock := bc.CurrentFastBlock()
		if bc.GetTd(currentFastBlock.Hash(), currentFastBlock.NumberU64()).Cmp(td) < 0 {
			rawdb.WriteHeadFastBlockHash(bc.db, head.Hash())
			bc.currentFastBlock.Store(head)
		}
	}
	bc.mu.Unlock()

	log.Info("Imported new block receipts",
		"count", stats.processed,
		"elapsed", common.PrettyDuration(time.Since(start)),
		"number", head.Number(),
		"hash", head.Hash(),
		"size", common.StorageSize(bytes),
		"ignored", stats.ignored)
	return 0, nil
}

var lastWrite uint64

//lb ipfs
func (bc *BlockChain) SetbSendIpfsFlg(flg bool) {
	bc.bBlockSendIpfs = flg
	if flg {
		bc.qBlockQueue = prque.New()
		log.Info("BlockChain SetbSendIpfsFlg and new")
	}
}
func (bc *BlockChain) GetStoreBlockInfo() *prque.Prque {
	//(storeBlock types.Blocks) {
	return bc.qBlockQueue
}

// WriteBlockWithoutState writes only the block and its metadata to the database,
// but does not write any state. This is used to construct competing side forks
// up to the point where they exceed the canonical total difficulty.
func (bc *BlockChain) WriteBlockWithoutState(block *types.Block, td *big.Int) (err error) {
	bc.wg.Add(1)
	defer bc.wg.Done()

	if err := bc.hc.WriteTd(block.Hash(), block.NumberU64(), td); err != nil {
		return err
	}
	rawdb.WriteBlock(bc.db, block)

	return nil
}

// WriteBlockWithState writes the block and all associated state to the database.
func (bc *BlockChain) WriteBlockWithState(block *types.Block, receipts []*types.Receipt, state *state.StateDB) (status WriteStatus, err error) {
	bc.wg.Add(1)
	defer bc.wg.Done()

	// Calculate the total difficulty of the block
	ptd := bc.GetTd(block.ParentHash(), block.NumberU64()-1)
	if ptd == nil {
		return NonStatTy, consensus.ErrUnknownAncestor
	}
	// Make sure no inconsistent state is leaked during insertion
	bc.mu.Lock()
	defer bc.mu.Unlock()

	getBlock := bc.GetBlockByHash(block.Hash())
	if nil != getBlock {
		return NonStatTy, fmt.Errorf("插入区块失败，已存在区块")
	}
	currentBlock := bc.CurrentBlock()
	localTd := bc.GetTd(currentBlock.Hash(), currentBlock.NumberU64())
	externTd := new(big.Int).Add(block.Difficulty(), ptd)

	// Irrelevant of the canonical status, write the block itself to the database
	if err := bc.hc.WriteTd(block.Hash(), block.NumberU64(), externTd); err != nil {
		return NonStatTy, err
	}
	// Write other block data using a batch.
	batch := bc.db.NewBatch()
	rawdb.WriteBlock(batch, block)
	if bc.bBlockSendIpfs && bc.qBlockQueue != nil {
		//bc.qBlockQueue.Push(block, -float32(block.NumberU64()))
		tmpBlock := &types.BlockAllSt{Sblock: block, SReceipt: receipts}
		//copy(tmpBlock.SReceipt, receipts)
		//tmpBlock.SReceipt = receipts
		tmpBlock.Pading = uint64(len(block.Body().Transactions))
		bc.qBlockQueue.Push(tmpBlock, -float32(block.NumberU64()))
		log.Trace("BlockChain WriteBlockWithState ipfs save block data", "block", block.NumberU64())
	}

	//log.Info("miss tree node debug", "入链时", "commit前state状态")
	//state.MissTrieDebug()
	deleteEmptyObjects := bc.chainConfig.IsEIP158(block.Number())
	intermediateRoot := state.IntermediateRoot(deleteEmptyObjects)
	//fmt.Printf("===ZH1==:%s\n", state.Dump())
	root, err := state.Commit(deleteEmptyObjects)
	if err != nil {
		return NonStatTy, err
	}

	if root != block.Root() {
		//fmt.Printf("===ZH2==:%s\n", state.Dump())
		log.INFO("blockChain", "WriteBlockWithState", "root信息", "root", root.Hex(), "header root", block.Root().Hex(), "intermediateRoot", intermediateRoot.Hex(), "deleteEmptyObjects", deleteEmptyObjects)

		//log.Info("miss tree node debug", "入链时", "commit后state状态")
		//state.MissTrieDebug()

		return NonStatTy, errors.New("root not match")
	}

	triedb := bc.stateCache.TrieDB()

	// If we're running an archive node, always flush
	if bc.cacheConfig.Disabled {
		log.Info("file blockchain", "gcmode modify archive", "")
		if err := triedb.Commit(root, false); err != nil {
			return NonStatTy, err
		}
	} else {
		// Full but not archive node, do proper garbage collection
		triedb.Reference(root, common.Hash{}) // metadata reference to keep trie alive
		bc.triegc.Push(root, -float32(block.NumberU64()))

		if current := block.NumberU64(); current > triesInMemory {
			// Find the next state trie we need to commit
			header := bc.GetHeaderByNumber(current - triesInMemory)
			chosen := header.Number.Uint64()

			// Only write to disk if we exceeded our memory allowance *and* also have at
			// least a given number of tries gapped.
			var (
				size  = triedb.Size()
				limit = common.StorageSize(bc.cacheConfig.TrieNodeLimit) * 1024 * 1024
			)
			if size > limit || bc.gcproc > bc.cacheConfig.TrieTimeLimit {
				// If we're exceeding limits but haven't reached a large enough memory gap,
				// warn the user that the system is becoming unstable.
				if chosen < lastWrite+triesInMemory {
					switch {
					case size >= 2*limit:
						log.Warn("State memory usage too high, committing", "size", size, "limit", limit, "optimum", float64(chosen-lastWrite)/triesInMemory)
					case bc.gcproc >= 2*bc.cacheConfig.TrieTimeLimit:
						log.Info("State in memory for too long, committing", "time", bc.gcproc, "allowance", bc.cacheConfig.TrieTimeLimit, "optimum", float64(chosen-lastWrite)/triesInMemory)
					}
				}
				// If optimum or critical limits reached, write to disk
				if chosen >= lastWrite+triesInMemory || size >= 2*limit || bc.gcproc >= 2*bc.cacheConfig.TrieTimeLimit {
					triedb.Commit(header.Root, true)
					lastWrite = chosen
					bc.gcproc = 0
				}
			}
			// Garbage collect anything below our required write retention
			for !bc.triegc.Empty() {
				root, number := bc.triegc.Pop()
				if uint64(-number) > chosen {
					bc.triegc.Push(root, number)
					break
				}
				triedb.Dereference(root.(common.Hash), common.Hash{})
			}
		}
	}
	rawdb.WriteReceipts(batch, block.Hash(), block.NumberU64(), receipts)

	// If the total difficulty is higher than our known, add it to the canonical chain
	// Second clause in the if statement reduces the vulnerability to selfish mining.
	// Please refer to http://www.cs.cornell.edu/~ie53/publications/btcProcFC.pdf
	reorg := externTd.Cmp(localTd) > 0
	currentBlock = bc.CurrentBlock()
	if block.IsSuperBlock() {
		status = CanonStatTy
	} else {
		if !reorg && externTd.Cmp(localTd) == 0 {
			// Split same-difficulty blocks by number, then at random
			reorg = block.NumberU64() < currentBlock.NumberU64() || (block.NumberU64() == currentBlock.NumberU64() && mrand.Float64() < 0.5)
		}
		if reorg {
			// Reorganise the chain if the parent is not the head block
			if block.ParentHash() != currentBlock.Hash() {
				if err := bc.reorg(currentBlock, block); err != nil {
					return NonStatTy, err
				}
			}
			// Write the positional metadata for transaction/receipt lookups and preimages
			rawdb.WriteTxLookupEntries(batch, block)
			rawdb.WritePreimages(batch, block.NumberU64(), state.Preimages())

			status = CanonStatTy
		} else {
			status = SideStatTy
		}
	}

	if err := batch.Write(); err != nil {
		return NonStatTy, err
	}

	// Set new head.
	if status == CanonStatTy {
		bc.insert(block)
	}

	bc.futureBlocks.Remove(block.Hash())
	return status, nil
}

// InsertChain attempts to insert the given batch of blocks in to the canonical
// chain or, otherwise, create a fork. If an error is returned it will return
// the index number of the failing block as well an error describing what went
// wrong.
//
// After insertion is done, all accumulated events will be fired.
func (bc *BlockChain) InsertChain(chain types.Blocks) (int, error) {
	n, events, logs, err := bc.insertChain(chain)
	bc.PostChainEvents(events, logs)
	return n, err
}

func (bc *BlockChain) InsertChainNotify(chain types.Blocks, notify bool) (int, error) {
	n, events, logs, err := bc.insertChain(chain)
	if notify {
		bc.PostChainEvents(events, logs)
	}
	return n, err
}

type randSeed struct {
	bc *BlockChain
}

func (r *randSeed) GetRandom(hash common.Hash, Type string) (*big.Int, error) {
	parent := r.bc.GetBlockByHash(hash)
	if parent == nil {
		log.Error(ModuleName, "获取父区块错误,hash", hash)
		return big.NewInt(0), nil
	}
	//_, preVrfValue, _ := common.GetVrfInfoFromHeader(parent.Header().VrfValue)
	//seed := common.BytesToHash(preVrfValue).Big()
	return nil, nil
}

// insertChain will execute the actual chain insertion and event aggregation. The
// only reason this method exists as a separate one is to make locking cleaner
// with deferred statements.
func (bc *BlockChain) insertChain(chain types.Blocks) (int, []interface{}, []*types.Log, error) {
	// Do a sanity check that the provided chain is actually ordered and linked
	log.Trace("BlockChain insertChain in")
	for i := 1; i < len(chain); i++ {
		if chain[i].NumberU64() != chain[i-1].NumberU64()+1 || chain[i].ParentHash() != chain[i-1].Hash() {
			// Chain broke ancestry, log a messge (programming error) and skip insertion
			log.Error("Non contiguous block insert", "number", chain[i].Number(), "hash", chain[i].Hash(),
				"parent", chain[i].ParentHash(), "prevnumber", chain[i-1].Number(), "prevhash", chain[i-1].Hash())

			return 0, nil, nil, fmt.Errorf("non contiguous insert: item %d is #%d [%x…], item %d is #%d [%x…] (parent [%x…])", i-1, chain[i-1].NumberU64(),
				chain[i-1].Hash().Bytes()[:4], i, chain[i].NumberU64(), chain[i].Hash().Bytes()[:4], chain[i].ParentHash().Bytes()[:4])
		}
	}
	// Pre-checks passed, start the full block imports
	bc.wg.Add(1)
	defer bc.wg.Done()

	bc.chainmu.Lock()
	defer bc.chainmu.Unlock()
	log.Trace("BlockChain insertChain in2")
	// A queued approach to delivering events. This is generally
	// faster than direct delivery and requires much less mutex
	// acquiring.
	var (
		stats         = insertStats{startTime: mclock.Now()}
		events        = make([]interface{}, 0, len(chain))
		lastCanon     *types.Block
		coalescedLogs []*types.Log
	)

	// Iterate over the blocks and insert when the verifier permits
	for i, block := range chain {
		// If the chain is terminating, stop processing blocks
		log.Trace("BlockChain insertChain in3 range", "blockNum", block.NumberU64())
		if atomic.LoadInt32(&bc.procInterrupt) == 1 {
			log.Debug("Premature abort during blocks processing")
			break
		}
		// If the header is a banned one, straight out abort
		if BadHashes[block.Hash()] {
			bc.reportBlock(block, nil, ErrBlacklistedHash)
			return i, events, coalescedLogs, ErrBlacklistedHash
		}
		// Wait for the block's verification to complete
		bstart := time.Now()

		header := block.Header()
		seal := true
		if manparams.IsBroadcastNumberByHash(block.NumberU64(), block.ParentHash()) || block.IsSuperBlock() {
			seal = false
		}
		err := bc.engine.VerifyHeader(bc, header, seal)
		if err == nil {
			err = bc.Validator().ValidateBody(block)
		}
		switch {
		case err == ErrKnownBlock:
			// Block and state both already known. However if the current block is below
			// this number we did a rollback and we should reimport it nonetheless.
			log.Trace("BlockChain insertChain in3 ErrKnownBlock")
			if bc.CurrentBlock().NumberU64() >= block.NumberU64() {
				stats.ignored++
				continue
			}

		case err == consensus.ErrFutureBlock:
			// Allow up to MaxFuture second in the future blocks. If this limit is exceeded
			// the chain is discarded and processed at a later time if given.
			log.Trace("BlockChain insertChain in3 ErrFutureBlock")
			max := big.NewInt(time.Now().Unix() + maxTimeFutureBlocks)
			if block.Time().Cmp(max) > 0 {
				return i, events, coalescedLogs, fmt.Errorf("future block: %v > %v", block.Time(), max)
			}
			bc.futureBlocks.Add(block.Hash(), block)
			stats.queued++
			continue

		case err == consensus.ErrUnknownAncestor && bc.futureBlocks.Contains(block.ParentHash()):
			log.Trace("BlockChain insertChain in3 ErrUnknownAncestor")
			bc.futureBlocks.Add(block.Hash(), block)
			stats.queued++
			continue

		case err == consensus.ErrPrunedAncestor:
			// Block competing with the canonical chain, store in the db, but don't process
			// until the competitor TD goes above the canonical TD
			log.Trace("BlockChain insertChain in3 ErrPrunedAncestor")
			currentBlock := bc.CurrentBlock()
			localTd := bc.GetTd(currentBlock.Hash(), currentBlock.NumberU64())
			externTd := new(big.Int).Add(bc.GetTd(block.ParentHash(), block.NumberU64()-1), block.Difficulty())
			if localTd.Cmp(externTd) > 0 {
				if err = bc.WriteBlockWithoutState(block, externTd); err != nil {
					return i, events, coalescedLogs, err
				}
				continue
			}
			// Competitor chain beat canonical, gather all blocks from the common ancestor
			var winner []*types.Block

			parent := bc.GetBlock(block.ParentHash(), block.NumberU64()-1)
			for !bc.HasState(parent.Root()) {
				winner = append(winner, parent)
				parent = bc.GetBlock(parent.ParentHash(), parent.NumberU64()-1)
			}
			for j := 0; j < len(winner)/2; j++ {
				winner[j], winner[len(winner)-1-j] = winner[len(winner)-1-j], winner[j]
			}
			// Import all the pruned blocks to make the state available
			bc.chainmu.Unlock()
			_, evs, logs, err := bc.insertChain(winner)
			bc.chainmu.Lock()
			events, coalescedLogs = evs, logs

			if err != nil {
				return i, events, coalescedLogs, err
			}

		case err != nil:
			log.Trace("BlockChain insertChain in3 reportBlock")
			bc.reportBlock(block, nil, err)
			return i, events, coalescedLogs, err
		}

		// verify pos
		err = bc.dposEngine.VerifyBlock(bc, header)
		if err != nil {
			log.Error("block chain", "insertChain DPOS共识错误", err)
			return 0, nil, nil, fmt.Errorf("insert block dpos error")
		}

		// Create a new statedb using the parent block and report an
		// error if it fails.
		var parent *types.Block
		if i == 0 {
			parent = bc.GetBlock(block.ParentHash(), block.NumberU64()-1)
		} else {
			parent = chain[i-1]
		}
		log.Trace("BlockChain insertChain in3 parent")
		// Process block using the parent state as reference point.
		state, err := state.New(parent.Root(), bc.stateCache)
		if err != nil {
			return i, events, coalescedLogs, err
		}
		var (
			receipts types.Receipts = nil
			logs                    = make([]*types.Log, 0)
			usedGas  uint64         = 0
		)
		if block.IsSuperBlock() {
			log.Trace("BlockChain insertChain in3 IsSuperBlock")
			sbs, err := bc.GetSuperBlockSeq()
			if nil != err {
				return i, events, coalescedLogs, errors.Errorf("get super seq error")
			}
			if block.Header().SuperBlockSeq() <= sbs {
				return i, events, coalescedLogs, errors.Errorf("invalid super block seq (remote: %x local: %x)", block.Header().SuperBlockSeq(), sbs)
			}
			log.Trace("BlockChain insertChain in3 IsSuperBlock processSuperBlockState")
			err = bc.processSuperBlockState(block, state)
			if err != nil {
				bc.reportBlock(block, receipts, err)
				return i, events, coalescedLogs, err
			}

			root := state.IntermediateRoot(bc.chainConfig.IsEIP158(block.Number()))
			if root != block.Root() {
				return i, events, coalescedLogs, errors.Errorf("invalid super block root (remote: %x local: %x)", block.Root, root)
			}
		} else {
			log.Trace("BlockChain insertChain in3 Process Block")
			uptimeMap, err := bc.ProcessUpTime(state, block.Header())
			if err != nil {
				log.Trace("BlockChain insertChain in3 Process Block err1")
				bc.reportBlock(block, nil, err)
				return i, events, coalescedLogs, err
			}

			// Process block using the parent state as reference point.
			receipts, logs, usedGas, err = bc.processor.Process(block, state, bc.vmConfig, uptimeMap)
			if err != nil {
				log.Trace("BlockChain insertChain in3 Process Block err2")
				bc.reportBlock(block, receipts, err)
				return i, events, coalescedLogs, err
			}

			// Process matrix state
			err = bc.matrixState.ProcessMatrixState(block, state)
			if err != nil {
				log.Trace("BlockChain insertChain in3 Process Block err3")
				return i, events, coalescedLogs, err
			}

			// Validate the state using the default validator
			err = bc.Validator().ValidateState(block, parent, state, receipts, usedGas)
			if err != nil {
				log.Trace("BlockChain insertChain in3 Process Block err4")
				bc.reportBlock(block, receipts, err)
				return i, events, coalescedLogs, err
			}
		}

		proctime := time.Since(bstart)
		log.Trace("BlockChain insertChain in3 WriteBlockWithState")
		// Write the block to the chain and get the status.
		status, err := bc.WriteBlockWithState(block, receipts, state)
		if err != nil {
			return i, events, coalescedLogs, err
		}
		switch status {
		case CanonStatTy:
			log.Debug(" Inserted new block", "number", block.Number(), "hash", block.Hash(), "uncles", len(block.Uncles()),
				"txs", len(block.Transactions()), "gas", block.GasUsed(), "elapsed", common.PrettyDuration(time.Since(bstart)))

			coalescedLogs = append(coalescedLogs, logs...)
			blockInsertTimer.UpdateSince(bstart)
			events = append(events, ChainEvent{block, block.Hash(), logs})
			lastCanon = block

			// Only count canonical blocks for GC processing time
			bc.gcproc += proctime

		case SideStatTy:
			log.Debug("Inserted forked block", "number", block.Number(), "hash", block.Hash(), "diff", block.Difficulty(), "elapsed",
				common.PrettyDuration(time.Since(bstart)), "txs", len(block.Transactions()), "gas", block.GasUsed(), "uncles", len(block.Uncles()))

			blockInsertTimer.UpdateSince(bstart)
			events = append(events, ChainSideEvent{block})
		}
		stats.processed++
		stats.usedGas += usedGas
		stats.report(chain, i, bc.stateCache.TrieDB().Size())
		//lb
		tmp := block.Transactions()
		log.Trace("BlockChain insertChain mem", len(tmp))

		tmp = nil

		thd := block.Header()
		thd.Elect = nil
		thd.Difficulty = nil
		thd.Number = nil
		thd.Time = nil
		thd.Extra = nil
		thd.Signatures = nil
		thd.Version = nil
		thd = nil
		receipts = nil
		block = nil
		logs = nil
		time.Sleep(10 * time.Millisecond)
	}
	debug.FreeOSMemory() //lb

	log.Trace("BlockChain insertChain out")
	// Append a single chain head event if we've progressed the chain
	if lastCanon != nil && bc.CurrentBlock().Hash() == lastCanon.Hash() {
		events = append(events, ChainHeadEvent{lastCanon})
	}
	return 0, events, coalescedLogs, nil
}

// insertStats tracks and reports on block insertion.
type insertStats struct {
	queued, processed, ignored int
	usedGas                    uint64
	lastIndex                  int
	startTime                  mclock.AbsTime
}

// statsReportLimit is the time limit during import after which we always print
// out progress. This avoids the user wondering what's going on.
const statsReportLimit = 8 * time.Second

// report prints statistics if some number of blocks have been processed
// or more than a few seconds have passed since the last message.
func (st *insertStats) report(chain []*types.Block, index int, cache common.StorageSize) {
	// Fetch the timings for the batch
	var (
		now     = mclock.Now()
		elapsed = time.Duration(now) - time.Duration(st.startTime)
	)
	// If we're at the last block of the batch or report period reached, log
	if index == len(chain)-1 || elapsed >= statsReportLimit {
		var (
			end = chain[index]
			txs = countTransactions(chain[st.lastIndex : index+1])
		)
		context := []interface{}{
			"blocks", st.processed, "txs", txs, "mgas", float64(st.usedGas) / 1000000,
			"elapsed", common.PrettyDuration(elapsed), "mgasps", float64(st.usedGas) * 1000 / float64(elapsed),
			"number", end.Number(), "hash", end.Hash(), "cache", cache,
		}
		if st.queued > 0 {
			context = append(context, []interface{}{"queued", st.queued}...)
		}
		if st.ignored > 0 {
			context = append(context, []interface{}{"ignored", st.ignored}...)
		}
		log.Info("Imported new chain segment", context...)

		*st = insertStats{startTime: now, lastIndex: index + 1}
	}
}

func countTransactions(chain []*types.Block) (c int) {
	for _, b := range chain {
		c += len(b.Transactions())
	}
	return c
}

// reorgs takes two blocks, an old chain and a new chain and will reconstruct the blocks and inserts them
// to be part of the new canonical chain and accumulates potential missing transactions and post an
// event about them
func (bc *BlockChain) reorg(oldBlock, newBlock *types.Block) error {
	var (
		newChain    types.Blocks
		oldChain    types.Blocks
		commonBlock *types.Block
		deletedTxs  types.SelfTransactions
		deletedLogs []*types.Log
		// collectLogs collects the logs that were generated during the
		// processing of the block that corresponds with the given hash.
		// These logs are later announced as deleted.
		collectLogs = func(hash common.Hash) {
			// Coalesce logs and set 'Removed'.
			number := bc.hc.GetBlockNumber(hash)
			if number == nil {
				return
			}
			receipts := rawdb.ReadReceipts(bc.db, hash, *number)
			for _, receipt := range receipts {
				for _, log := range receipt.Logs {
					del := *log
					del.Removed = true
					deletedLogs = append(deletedLogs, &del)
				}
			}
		}
	)

	// first reduce whoever is higher bound
	if oldBlock.NumberU64() > newBlock.NumberU64() {
		// reduce old chain
		for ; oldBlock != nil && oldBlock.NumberU64() != newBlock.NumberU64(); oldBlock = bc.GetBlock(oldBlock.ParentHash(), oldBlock.NumberU64()-1) {
			oldChain = append(oldChain, oldBlock)
			deletedTxs = append(deletedTxs, oldBlock.Transactions()...)

			collectLogs(oldBlock.Hash())
		}
	} else {
		// reduce new chain and append new chain blocks for inserting later on
		for ; newBlock != nil && newBlock.NumberU64() != oldBlock.NumberU64(); newBlock = bc.GetBlock(newBlock.ParentHash(), newBlock.NumberU64()-1) {
			newChain = append(newChain, newBlock)
		}
	}
	if oldBlock == nil {
		return fmt.Errorf("Invalid old chain")
	}
	if newBlock == nil {
		return fmt.Errorf("Invalid new chain")
	}

	for {
		if oldBlock.Hash() == newBlock.Hash() {
			commonBlock = oldBlock
			break
		}

		oldChain = append(oldChain, oldBlock)
		newChain = append(newChain, newBlock)
		deletedTxs = append(deletedTxs, oldBlock.Transactions()...)
		collectLogs(oldBlock.Hash())

		oldBlock, newBlock = bc.GetBlock(oldBlock.ParentHash(), oldBlock.NumberU64()-1), bc.GetBlock(newBlock.ParentHash(), newBlock.NumberU64()-1)
		if oldBlock == nil {
			return fmt.Errorf("Invalid old chain")
		}
		if newBlock == nil {
			return fmt.Errorf("Invalid new chain")
		}
	}
	// Ensure the user sees large reorgs
	if len(oldChain) > 0 && len(newChain) > 0 {
		logFn := log.Debug
		if len(oldChain) > 63 {
			logFn = log.Warn
		}
		logFn("Chain split detected", "number", commonBlock.Number(), "hash", commonBlock.Hash(),
			"drop", len(oldChain), "dropfrom", oldChain[0].Hash(), "add", len(newChain), "addfrom", newChain[0].Hash())
	} else {
		log.Error("Impossible reorg, please file an issue", "oldnum", oldBlock.Number(), "oldhash", oldBlock.Hash(), "newnum", newBlock.Number(), "newhash", newBlock.Hash())
	}
	// Insert the new chain, taking care of the proper incremental order
	var addedTxs types.SelfTransactions
	for i := len(newChain) - 1; i >= 0; i-- {
		// insert the block in the canonical way, re-writing history
		bc.insert(newChain[i])
		// write lookup entries for hash based transaction/receipt searches
		rawdb.WriteTxLookupEntries(bc.db, newChain[i])
		addedTxs = append(addedTxs, newChain[i].Transactions()...)
	}
	// calculate the difference between deleted and added transactions
	diff := types.TxDifference(deletedTxs, addedTxs)
	// When transactions get deleted from the database that means the
	// receipts that were created in the fork must also be deleted
	for _, tx := range diff {
		rawdb.DeleteTxLookupEntry(bc.db, tx.Hash())
	}
	if len(deletedLogs) > 0 {
		go bc.rmLogsFeed.Send(RemovedLogsEvent{deletedLogs})
	}
	if len(oldChain) > 0 {
		go func() {
			for _, block := range oldChain {
				bc.chainSideFeed.Send(ChainSideEvent{Block: block})
			}
		}()
	}

	return nil
}

// PostChainEvents iterates over the events generated by a chain insertion and
// posts them into the event feed.
// TODO: Should not expose PostChainEvents. The chain events should be posted in WriteBlock.
func (bc *BlockChain) PostChainEvents(events []interface{}, logs []*types.Log) {
	// post event logs for further processing
	if logs != nil {
		bc.logsFeed.Send(logs)
	}
	for _, event := range events {
		switch ev := event.(type) {
		case ChainEvent:
			bc.chainFeed.Send(ev)

		case ChainHeadEvent:
			bc.chainHeadFeed.Send(ev)
			//YY=========Begin===============
			bc.sendBroadTx()
			//=============end===============
			mc.PublishEvent(mc.NewBlockMessage, ev.Block)

		case ChainSideEvent:
			bc.chainSideFeed.Send(ev)
		}
	}
}

//YY 发送心跳交易
var viSendHeartTx bool = false         //是否验证过发送心跳交易，每100块内只验证一次 //YY
var saveBroacCastblockHash common.Hash //YY 广播区块的hash  默认值应该为创世区块的hash
func (bc *BlockChain) sendBroadTx() {
	block := bc.CurrentBlock()
	bcInterval, err := bc.GetBroadcastInterval(block.Hash())
	if err != nil {
		log.ERROR("sendBroadTx", "获取广播周期失败", err)
		return
	}

	blockNum := block.Number()
	subVal := bcInterval.LastBCNumber
	log.Info("sendBroadTx", "获取广播高度", subVal)
	//没验证过心跳交易
	if !viSendHeartTx {
		viSendHeartTx = true
		//广播区块的roothash与99取余如果与广播账户与99取余的结果一样那么发送广播交易

		log.INFO(ModuleName, "sendBroadTx获取所有心跳交易", "")
		preBroadcastRoot, err := readstatedb.GetPreBroadcastRoot(bc, blockNum.Uint64())
		if err != nil {
			log.Error(ModuleName, "获取之前广播区块的root值失败 err", err)
			return
		}
		log.INFO(ModuleName, "sendBroadTx获取最新的root", preBroadcastRoot.LastStateRoot.Hex())
		currentAcc := ca.GetAddress().Big() //YY TODO 这里应该是广播账户。后期需要修改. 后期可能需要使用委托账户
		ret := new(big.Int).Rem(currentAcc, big.NewInt(int64(bcInterval.BCInterval)-1))
		broadcastBlock := preBroadcastRoot.LastStateRoot.Big()
		val := new(big.Int).Rem(broadcastBlock, big.NewInt(int64(bcInterval.BCInterval)-1))
		if ret.Cmp(val) == 0 {
			height := new(big.Int).Add(new(big.Int).SetUint64(subVal), big.NewInt(int64(bcInterval.BCInterval))) //下一广播区块的高度
			data := new([]byte)
			mc.PublishEvent(mc.SendBroadCastTx, mc.BroadCastEvent{mc.Heartbeat, height, *data})
			log.Trace("file blockchain", "blockChian:sendBroadTx()", ret, "val", val)
		}
		log.Trace("file blockchain", "blockChian:sendBroadTx()", ret, "val", val)
	}

	if blockNum.Uint64()%bcInterval.BCInterval == 0 { //到整百的区块后需要重置数据以便下一区块验证是否发送心跳交易
		saveBroacCastblockHash = block.Hash()
		viSendHeartTx = false
	}
}

func (bc *BlockChain) update() {
	futureTimer := time.NewTicker(5 * time.Second)
	defer futureTimer.Stop()
	for {
		select {
		case <-futureTimer.C:
			bc.procFutureBlocks()
		case <-bc.quit:
			return
		}
	}
}

// BadBlockArgs represents the entries in the list returned when bad blocks are queried.
type BadBlockArgs struct {
	Hash   common.Hash   `json:"hash"`
	Header *types.Header `json:"header"`
}

// BadBlocks returns a list of the last 'bad blocks' that the client has seen on the network
func (bc *BlockChain) BadBlocks() ([]BadBlockArgs, error) {
	headers := make([]BadBlockArgs, 0, bc.badBlocks.Len())
	for _, hash := range bc.badBlocks.Keys() {
		if hdr, exist := bc.badBlocks.Peek(hash); exist {
			header := hdr.(*types.Header)
			headers = append(headers, BadBlockArgs{header.Hash(), header})
		}
	}
	return headers, nil
}

// addBadBlock adds a bad block to the bad-block LRU cache
func (bc *BlockChain) addBadBlock(block *types.Block) {
	bc.badBlocks.Add(block.Header().Hash(), block.Header())
}

// reportBlock logs a bad block error.
func (bc *BlockChain) reportBlock(block *types.Block, receipts types.Receipts, err error) {
	bc.addBadBlock(block)

	var receiptString string
	for _, receipt := range receipts {
		receiptString += fmt.Sprintf("\t%v\n", receipt)
	}
	log.Error(fmt.Sprintf(`
########## BAD BLOCK #########
Chain config: %v

Number: %v
Hash: 0x%x
%v

Error: %v
##############################
`, bc.chainConfig, block.Number(), block.Hash(), receiptString, err))
}

// InsertHeaderChain attempts to insert the given header chain in to the local
// chain, possibly creating a reorg. If an error is returned, it will return the
// index number of the failing header as well an error describing what went wrong.
//
// The verify parameter can be used to fine tune whether nonce verification
// should be done or not. The reason behind the optional check is because some
// of the header retrieval mechanisms already need to verify nonces, as well as
// because nonces can be verified sparsely, not needing to check each.
func (bc *BlockChain) InsertHeaderChain(chain []*types.Header, checkFreq int) (int, error) {
	start := time.Now()
	if i, err := bc.hc.ValidateHeaderChain(chain, checkFreq); err != nil {
		return i, err
	}

	// Make sure only one thread manipulates the chain at once
	bc.chainmu.Lock()
	defer bc.chainmu.Unlock()

	bc.wg.Add(1)
	defer bc.wg.Done()

	whFunc := func(header *types.Header) error {
		bc.mu.Lock()
		defer bc.mu.Unlock()

		_, err := bc.hc.WriteHeader(header)
		return err
	}

	return bc.hc.InsertHeaderChain(chain, whFunc, start)
}

// writeHeader writes a header into the local chain, given that its parent is
// already known. If the total difficulty of the newly inserted header becomes
// greater than the current known TD, the canonical chain is re-routed.
//
// Note: This method is not concurrent-safe with inserting blocks simultaneously
// into the chain, as side effects caused by reorganisations cannot be emulated
// without the real blocks. Hence, writing headers directly should only be done
// in two scenarios: pure-header mode of operation (light clients), or properly
// separated header/block phases (non-archive clients).
func (bc *BlockChain) writeHeader(header *types.Header) error {
	bc.wg.Add(1)
	defer bc.wg.Done()

	bc.mu.Lock()
	defer bc.mu.Unlock()

	_, err := bc.hc.WriteHeader(header)
	return err
}

// CurrentHeader retrieves the current head header of the canonical chain. The
// header is retrieved from the HeaderChain's internal cache.
func (bc *BlockChain) CurrentHeader() *types.Header {
	return bc.hc.CurrentHeader()
}

// GetTd retrieves a block's total difficulty in the canonical chain from the
// database by hash and number, caching it if found.
func (bc *BlockChain) GetTd(hash common.Hash, number uint64) *big.Int {
	return bc.hc.GetTd(hash, number)
}

// GetTdByHash retrieves a block's total difficulty in the canonical chain from the
// database by hash, caching it if found.
func (bc *BlockChain) GetTdByHash(hash common.Hash) *big.Int {
	return bc.hc.GetTdByHash(hash)
}

// GetHeader retrieves a block header from the database by hash and number,
// caching it if found.
func (bc *BlockChain) GetHeader(hash common.Hash, number uint64) *types.Header {
	return bc.hc.GetHeader(hash, number)
}

// GetHeaderByHash retrieves a block header from the database by hash, caching it if
// found.
func (bc *BlockChain) GetHeaderByHash(hash common.Hash) *types.Header {
	return bc.hc.GetHeaderByHash(hash)
}

// HasHeader checks if a block header is present in the database or not, caching
// it if present.
func (bc *BlockChain) HasHeader(hash common.Hash, number uint64) bool {
	return bc.hc.HasHeader(hash, number)
}

// GetBlockHashesFromHash retrieves a number of block hashes starting at a given
// hash, fetching towards the genesis block.
func (bc *BlockChain) GetBlockHashesFromHash(hash common.Hash, max uint64) []common.Hash {
	return bc.hc.GetBlockHashesFromHash(hash, max)
}

// GetHeaderByNumber retrieves a block header from the database by number,
// caching it (associated with its hash) if found.
func (bc *BlockChain) GetHeaderByNumber(number uint64) *types.Header {
	return bc.hc.GetHeaderByNumber(number)
}

// Config retrieves the blockchain's chain configuration.
func (bc *BlockChain) Config() *params.ChainConfig { return bc.chainConfig }

// Engine retrieves the blockchain's consensus engine.
func (bc *BlockChain) Engine() consensus.Engine { return bc.engine }

func (bc *BlockChain) DPOSEngine() consensus.DPOSEngine { return bc.dposEngine }

// SubscribeRemovedLogsEvent registers a subscription of RemovedLogsEvent.
func (bc *BlockChain) SubscribeRemovedLogsEvent(ch chan<- RemovedLogsEvent) event.Subscription {
	return bc.scope.Track(bc.rmLogsFeed.Subscribe(ch))
}

// SubscribeChainEvent registers a subscription of ChainEvent.
func (bc *BlockChain) SubscribeChainEvent(ch chan<- ChainEvent) event.Subscription {
	return bc.scope.Track(bc.chainFeed.Subscribe(ch))
}

// SubscribeChainHeadEvent registers a subscription of ChainHeadEvent.
func (bc *BlockChain) SubscribeChainHeadEvent(ch chan<- ChainHeadEvent) event.Subscription {
	return bc.scope.Track(bc.chainHeadFeed.Subscribe(ch))
}

// SubscribeChainSideEvent registers a subscription of ChainSideEvent.
func (bc *BlockChain) SubscribeChainSideEvent(ch chan<- ChainSideEvent) event.Subscription {
	return bc.scope.Track(bc.chainSideFeed.Subscribe(ch))
}

// SubscribeLogsEvent registers a subscription of []*types.Log.
func (bc *BlockChain) SubscribeLogsEvent(ch chan<- []*types.Log) event.Subscription {
	return bc.scope.Track(bc.logsFeed.Subscribe(ch))
}

func (bc *BlockChain) VerifyHeader(header *types.Header) error {
	return bc.engine.VerifyHeader(bc, header, false)
}

func (bc *BlockChain) SetDposEngine(dposEngine consensus.DPOSEngine) {
	bc.dposEngine = dposEngine
}

func (bc *BlockChain) GetHashByNumber(number uint64) common.Hash {
	block := bc.GetBlockByNumber(number)
	if block == nil {
		return common.Hash{}
	}
	return block.Hash()
}

func (bc *BlockChain) GetCurrentHash() common.Hash {
	block := bc.CurrentBlock()
	if block == nil {
		return common.Hash{}
	}
	return block.Hash()
}

func (bc *BlockChain) GetGraphByHash(hash common.Hash) (*mc.TopologyGraph, *mc.ElectGraph, error) {
	topologyGraph, err := bc.graphStore.GetTopologyGraphByHash(hash)
	if err != nil {
		return nil, nil, err
	}
	electGraph, err := bc.graphStore.GetElectGraphByHash(hash)
	if err != nil {
		return nil, nil, err
	}
	return topologyGraph, electGraph, nil
}

func (bc *BlockChain) GetGraphByState(state matrixstate.StateDB) (*mc.TopologyGraph, *mc.ElectGraph, error) {
	topologyGraph, err := matrixstate.GetDataByState(mc.MSKeyTopologyGraph, state)
	if err != nil {
		return nil, nil, err
	}
	electGraph, err := matrixstate.GetDataByState(mc.MSKeyElectGraph, state)
	if err != nil {
		return nil, nil, err
	}
	return topologyGraph.(*mc.TopologyGraph), electGraph.(*mc.ElectGraph), nil
}

func (bc *BlockChain) ProcessMatrixState(block *types.Block, state *state.StateDB) error {
	return bc.matrixState.ProcessMatrixState(block, state)
}

func (bc *BlockChain) GetGraphStore() *matrixstate.GraphStore {
	return bc.graphStore
}

func (bc *BlockChain) RegisterMatrixStateDataProducer(key string, producer matrixstate.ProduceMatrixStateDataFn) error {
	return bc.matrixState.RegisterProducer(key, producer)
}

func (bc *BlockChain) GetAncestorHash(sonHash common.Hash, ancestorNumber uint64) (common.Hash, error) {
	return bc.hc.GetAncestorHash(sonHash, ancestorNumber)
}

func (bc *BlockChain) GetMatrixStateData(key string) (interface{}, error) {
	state, err := bc.State()
	if err != nil {
		return nil, errors.Errorf("get cur state err(%v)", err)
	}
	if state == nil {
		return nil, errors.New("cur state is nil")
	}
	return matrixstate.GetDataByState(key, state)
}

func (bc *BlockChain) GetMatrixStateDataByHash(key string, hash common.Hash) (interface{}, error) {
	header := bc.GetHeaderByHash(hash)
	if header == nil {
		return nil, errors.Errorf("can't find block by hash(%s)", hash.Hex())
	}
	state, err := bc.StateAt(header.Root)
	if err != nil {
		return nil, errors.Errorf("can't find state by root(%s): %v", header.Root.TerminalString(), err)
	}
	if state == nil {
		return nil, errors.Errorf("state of root(%s) is nil", header.Root.TerminalString())
	}
	return matrixstate.GetDataByState(key, state)
}

func (bc *BlockChain) GetMatrixStateDataByNumber(key string, number uint64) (interface{}, error) {
	header := bc.GetHeaderByNumber(number)
	if header == nil {
		return nil, errors.Errorf("can't find block by number(%d)", number)
	}
	state, err := bc.StateAt(header.Root)
	if err != nil {
		return nil, errors.Errorf("can't find state by root(%s): %v", header.Root.TerminalString(), err)
	}
	if state == nil {
		return nil, errors.Errorf("state of root(%s) is nil", header.Root.TerminalString())
	}
	return matrixstate.GetDataByState(key, state)
}

func (bc *BlockChain) GetBroadcastAccount(blockHash common.Hash) (common.Address, error) {
	data, err := bc.GetMatrixStateDataByHash(mc.MSKeyAccountBroadcast, blockHash)
	if err != nil {
		return common.Address{}, err
	}
	broadcast, OK := data.(common.Address)
	if OK == false {
		return common.Address{}, errors.New("反射结构体失败")
	}
	return broadcast, nil
}

func (bc *BlockChain) GetInnerMinerAccounts(blockHash common.Hash) ([]common.Address, error) {
	data, err := bc.GetMatrixStateDataByHash(mc.MSKeyAccountInnerMiners, blockHash)
	if err != nil {
		return nil, err
	}
	accounts, OK := data.([]common.Address)
	if OK == false {
		return nil, errors.New("反射结构体失败")
	}
	return accounts, nil
}

func (bc *BlockChain) GetVersionSuperAccounts(blockHash common.Hash) ([]common.Address, error) {
	data, err := bc.GetMatrixStateDataByHash(mc.MSKeyAccountVersionSupers, blockHash)
	if err != nil {
		return nil, err
	}
	accounts, OK := data.([]common.Address)
	if OK == false {
		return nil, errors.New("反射结构体失败")
	}
	return accounts, nil
}

func (bc *BlockChain) GetBlockSuperAccounts(blockHash common.Hash) ([]common.Address, error) {
	data, err := bc.GetMatrixStateDataByHash(mc.MSKeyAccountBlockSupers, blockHash)
	if err != nil {
		return nil, err
	}
	accounts, OK := data.([]common.Address)
	if OK == false {
		return nil, errors.New("反射结构体失败")
	}
	return accounts, nil
}

func (bc *BlockChain) GetBroadcastInterval(blockHash common.Hash) (*mc.BCIntervalInfo, error) {
	data, err := bc.GetMatrixStateDataByHash(mc.MSKeyBroadcastInterval, blockHash)
	if err != nil {
		return nil, err
	}

	interval, OK := data.(*mc.BCIntervalInfo)
	if OK == false {
		return nil, errors.New("反射广播周期失败")
	}
	log.INFO("blockChain", "广播周期", interval.BCInterval, "上个广播高度", interval.LastBCNumber)
	return interval, nil
}

func (bc *BlockChain) GetSuperBlockSeq() (uint64, error) {

	data, err := bc.GetMatrixStateData(mc.MSKeySuperBlockCfg)
	if err != nil {
		return 0, err
	}

	superBlkCfg, OK := data.(*mc.SuperBlkCfg)
	if OK == false {
		return 0, errors.New("反射广播周期失败")
	}
	log.INFO("blockChain", "超级区块序号", superBlkCfg.Seq)

	return superBlkCfg.Seq, nil
}

func (bc *BlockChain) GetSuperBlockNum() (uint64, error) {
	data, err := bc.GetMatrixStateData(mc.MSKeySuperBlockCfg)
	if err != nil {
		return 0, err
	}

	superBlkCfg, OK := data.(*mc.SuperBlkCfg)
	if OK == false {
		return 0, errors.New("反射广播周期失败")
	}
	log.INFO("blockChain", "超级区块高度", superBlkCfg.Num)

	return superBlkCfg.Num, nil
}

func (bc *BlockChain) GetSuperBlockInfo() (*mc.SuperBlkCfg, error) {
	data, err := bc.GetMatrixStateData(mc.MSKeySuperBlockCfg)
	if err != nil {
		return nil, err
	}

	superBlkCfg, OK := data.(*mc.SuperBlkCfg)
	if OK == false {
		return nil, errors.New("反射广播周期失败")
	}
	log.INFO("blockChain", "超级区块高度", superBlkCfg.Num)

	return superBlkCfg, nil
}
func (bc *BlockChain) getBCIntervalByState(st *state.StateDB) (*manparams.BCInterval, error) {
	data, err := matrixstate.GetDataByState(mc.MSKeyBroadcastInterval, st)
	if err != nil {
		return nil, err
	}
	return manparams.NewBCIntervalWithInterval(data)
}

func (bc *BlockChain) InsertSuperBlock(superBlockGen *Genesis, notify bool) (*types.Block, error) {
	if nil == superBlockGen {
		return nil, errors.New("super block is nil")
	}
	if superBlockGen.Number <= 0 {
		return nil, errors.Errorf("super block`s number(%d) is too low", superBlockGen.Number)
	}
	parent := bc.GetBlockByHash(superBlockGen.ParentHash)
	if nil == parent {
		return nil, errors.Errorf("get parent block by hash(%s) err", superBlockGen.ParentHash.Hex())
	}
	if parent.NumberU64()+1 != superBlockGen.Number {
		return nil, errors.Errorf("parent block number(%d) + 1 != super block number(%d)", parent.NumberU64(), superBlockGen.Number)
	}

	block := superBlockGen.GenSuperBlock(parent.Header(), bc.stateCache, bc.chainConfig)
	if nil == block {
		return nil, errors.New("genesis super block failed")
	}

	if !block.IsSuperBlock() {
		return nil, errors.New("err, genesis block is not super block!")
	}
	if block.Root() != superBlockGen.Root {
		return nil, errors.Errorf("root not match, calc root(%s) != genesis root(%s)", block.Root().TerminalString(), superBlockGen.Root.TerminalString())
	}
	if block.TxHash() != superBlockGen.TxHash {
		return nil, errors.Errorf("txHash not match, calc txHash(%s) != genesis txHash(%s)", block.TxHash().TerminalString(), superBlockGen.TxHash.TerminalString())
	}
	sbh, err := bc.GetSuperBlockNum()
	if nil != err {
		return nil, errors.Errorf("get super seq error")
	}
	superBlock := bc.GetBlockByNumber(sbh)
	if nil != superBlock {
		if block.Hash() == superBlock.Hash() {
			log.WARN(ModuleName, "has the same super block", "")
			return block, nil
		}
	}
	sbs, err := bc.GetSuperBlockSeq()
	if nil != err {
		return nil, errors.Errorf("get super seq error")
	}
	if block.Header().SuperBlockSeq() <= sbs {
		return nil, errors.Errorf("SuperBlockSeq not match, current seq(%v) < genesis block(%v)", sbs, block.Header().SuperBlockSeq())
	}

	if err := bc.DPOSEngine().VerifyBlock(bc, block.Header()); err != nil {
		return nil, errors.Errorf("verify super block err(%v)", err)
	}
	//todo 应该在InsertChain时确定权威链，从而进行回滚
	//if err := bc.SetHead(superBlockGen.Number - 1); err != nil {
	//	return nil, errors.Errorf("rollback chain err(%v)", err)
	//}

	if _, err := bc.InsertChainNotify(types.Blocks{block}, notify); err != nil {
		return nil, errors.Errorf("insert super block err(%v)", err)
	}

	if _, err := bc.StateAt(block.Root()); err != nil {
		log.Error("hyk", "get state err", err, "root", block.Root())
	}

	return block, nil
}

func (bc *BlockChain) processSuperBlockState(block *types.Block, stateDB *state.StateDB) error {
	if nil == block || nil == stateDB {
		return errors.New("param is nil")
	}

	txs := block.Transactions()
	if len(txs) == 0 || len(txs) > 2 {
		return errors.Errorf("super block's txs count(%d) err", len(txs))
	}

	tx := txs[0]
	if tx.GetMatrixType() != common.ExtraSuperBlockTx {
		return errors.Errorf("super block's tx type(%d) err", tx.TxType())
	}
	if tx.Nonce() != block.NumberU64() {
		return errors.Errorf("super block's tx nonce(%d) err, != block number(%d)", tx.Nonce(), block.NumberU64())
	}

	var alloc GenesisAlloc
	if err := alloc.UnmarshalJSON(tx.Data()); err != nil {
		return errors.Errorf("super block: unmarshal alloc info err(%v)", err)
	}

	for addr, account := range alloc {
		stateDB.SetBalance(common.MainAccount, addr, account.Balance)
		stateDB.SetCode(addr, account.Code)
		stateDB.SetNonce(addr, account.Nonce)
		for key, value := range account.Storage {
			stateDB.SetState(addr, key, value)
		}
	}
	mState := new(GenesisMState)
	if 2 == len(txs) {
		tx1 := txs[1]

		if err := json.Unmarshal(tx1.Data(), mState); err != nil {
			return errors.Errorf("super block: unmarshal matrix state info err(%v)", err)
		}
		mState.setMatrixState(stateDB, block.Header().NetTopology, block.Header().Elect, block.Header().Number.Uint64())

	}
	if err := mState.SetSuperBlkToState(stateDB, block.Header().Extra, block.Header().Number.Uint64()); err != nil {
		log.Error("genesis", "设置matrix状态树错误", err)
		return errors.Errorf("设置超级区块状态树错误", err)
	}
	return nil
}

func ProduceBroadcastIntervalData(block *types.Block, readFn matrixstate.PreStateReadFn) (interface{}, error) {
	bciData, err := readFn(mc.MSKeyBroadcastInterval)
	if err != nil {
		log.Error("ProduceBroadcastIntervalData", "read pre broadcast interval err", err)
		return nil, err
	}

	bcInterval, err := manparams.NewBCIntervalWithInterval(bciData)
	if err != nil {
		return nil, err
	}

	modify := false
	number := block.NumberU64()
	backupEnableNumber := bcInterval.GetBackupEnableNumber()
	if number == backupEnableNumber {
		// 备选生效时间点
		if bcInterval.IsReElectionNumber(number) == false || bcInterval.IsBroadcastNumber(number) == false {
			// 生效时间点不是原周期的选举点，数据错误
			log.Crit("ProduceBroadcastIntervalData", "backup enable number illegal", backupEnableNumber,
				"old interval", bcInterval.GetBroadcastInterval(), "last broadcast number", bcInterval.GetLastBroadcastNumber(), "last reelect number", bcInterval.GetLastReElectionNumber())
		}

		oldInterval := bcInterval.GetBroadcastInterval()

		// 设置最后的广播区块和选举区块
		bcInterval.SetLastBCNumber(backupEnableNumber)
		bcInterval.SetLastReelectNumber(backupEnableNumber)
		// 启动备选周期
		bcInterval.UsingBackupInterval()
		log.INFO("ProduceBroadcastIntervalData", "old interval", oldInterval, "new interval", bcInterval.GetBroadcastInterval())
		modify = true
	} else {
		if bcInterval.IsBroadcastNumber(number) {
			bcInterval.SetLastBCNumber(number)
			modify = true
		}

		if bcInterval.IsReElectionNumber(number) {
			bcInterval.SetLastReelectNumber(number)
			modify = true
		}
	}

	if modify {
		data := bcInterval.ToInfoStu()
		log.INFO("ProduceBroadcastIntervalData", "生成广播区块内容", "成功", "block number", number, "data", data)
		return data, nil
	} else {
		return nil, nil
	}
}

func (bc *BlockChain) GetSignAccountPassword(signAccounts []common.Address) (common.Address, string, error) {
	entrustValue := manparams.EntrustAccountValue.GetEntrustValue()
	for _, signAccount := range signAccounts {
		for account, password := range entrustValue {
			if signAccount != account {
				continue
			}
			return signAccount, password, nil
		}
	}
	log.Info(common.SignLog, "获取签名账户密码", "失败, 未找到")
	return common.Address{}, "", errors.New("未找到密码")
}

func (bc *BlockChain) GetSignAccounts(authFrom common.Address, blockHash common.Hash) ([]common.Address, error) {
	if common.TopAccountType == common.TopAccountA0 {
		//TODO 暂定根据ca提供的接口获取委托账户，
	}
	block := bc.GetBlockByHash(blockHash)
	if block == nil {
		log.ERROR(common.SignLog, "获取签名账户阶段", "BlockChain 最终结果", "根据区块hash获取区块失败 hash", blockHash)
		return nil, errors.Errorf("获取区块(%s)失败", blockHash.TerminalString())
	}
	st, err := bc.StateAt(block.Root())
	if err != nil {
		log.ERROR(common.SignLog, "获取签名账户阶段", "BlockChain 最终结果", "根据区块root获取statedb失败 err", err)
		return nil, errors.New("获取stateDB失败")
	}

	height := block.NumberU64()

	ans := []common.Address{}
	ans = st.GetEntrustFrom(authFrom, height)
	if len(ans) == 0 {
		ans = append(ans, authFrom)
		log.INFO(common.SignLog, "获取签名账户阶段", ModuleName, "无委托交易,使用本地账户", authFrom.String())
	}
	return ans, nil
}

//TransSignAccontToDeposit(signAccount common.Address, height uint64) (common.Address, error) {
func (bc *BlockChain) GetAuthAccount(signAccount common.Address, blockHash common.Hash) (common.Address, error) {
	block := bc.GetBlockByHash(blockHash)
	if block == nil {
		log.ERROR(common.SignLog, "获取委托账户阶段", "BlockChain 最终结果", "根据区块hash算区块失败", "err")
		return common.Address{}, errors.Errorf("获取区块(%s)失败", blockHash.TerminalString())
	}
	st, err := bc.StateAt(block.Root())
	if err != nil {
		log.ERROR(common.SignLog, "获取委托账户阶段", "BlockChain 最终结果", "根据区块root获取状态树失败 err", err)
		return common.Address{}, errors.New("获取stateDB失败")
	}

	height := block.NumberU64()
	addr := st.GetAuthFrom(signAccount, height)
	if addr.Equal(common.Address{}) {
		addr = signAccount
		log.WARN(common.SignLog, "获取委托账户阶段", ModuleName, "不存在委托账户 signAccount", signAccount, "高度", height, "委托账户", addr)
	} else {
		log.WARN(common.SignLog, "获取委托账户阶段", ModuleName, "存在委托 signAccount", signAccount, "height", height, "addr", addr)
	}
	log.Info(common.SignLog, "获取委托账户阶段", "BlockChain 最终结果", "高度", height, "签名账户", signAccount, "真实账户", addr)
	if common.TopAccountType == common.TopAccountA0 {
		//TODO 利用CA接口将A1转换为A0
	}
	return addr, nil
}
