// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

// Package fetcher contains the block announcement based synchronisation.
package fetcher

import (
	"errors"
	"math/rand"
	"time"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/consensus"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/log"
	"gopkg.in/karalabe/cookiejar.v2/collections/prque"
)

const (
	arriveTimeout = 500 * time.Millisecond // Time allowance before an announced block is explicitly requested
	gatherSlack   = 100 * time.Millisecond // Interval used to collate almost-expired announces with fetches
	fetchTimeout  = 9 * time.Second        // Maximum allotted time to return an explicitly requested block
	maxUncleDist  = 7                      // Maximum allowed backward distance from the chain head
	maxQueueDist  = 32                     // Maximum allowed distance from the chain head to queue
	hashLimit     = 256                    // Maximum number of unique blocks a peer may have announced
	blockLimit    = 64                     // Maximum number of unique blocks a peer may have delivered
)

var (
	errTerminated = errors.New("terminated")
)

// blockRetrievalFn is a callback type for retrieving a block from the local chain.
type blockRetrievalFn func(common.Hash) *types.Block

// headerRequesterFn is a callback type for sending a header retrieval request.
type headerRequesterFn func(common.Hash , uint64) error

// bodyRequesterFn is a callback type for sending a body retrieval request.
type bodyRequesterFn func([]common.Hash) error

// headerVerifierFn is a callback type to verify a block's header for fast propagation.
type headerVerifierFn func(header *types.Header) error

// blockBroadcasterFn is a callback type for broadcasting a block to connected peers.
type blockBroadcasterFn func(block *types.Block, propagate bool)

// chainHeightFn is a callback type to retrieve the current chain height.
type chainHeightFn func() uint64

// chainInsertFn is a callback type to insert a batch of blocks into the local chain.
type chainInsertFn func(types.Blocks) (int, error)

// peerDropFn is a callback type for dropping a peer detected as malicious.
type peerDropFn func(id string,flg int)

// announce is the hash notification of the availability of a new block in the
// network.
type announce struct {
	hash   common.Hash   // Hash of the block being announced
	number uint64        // Number of the block being announced (0 = unknown | old protocol)
	header *types.Header // Header of the block partially reassembled (new protocol)
	time   time.Time     // Timestamp of the announcement
	
	origin string // Identifier of the peer originating the notification
	//hasSendReq  int
	fetchHeader headerRequesterFn // Fetcher function to retrieve the header of an announced block
	fetchBodies bodyRequesterFn   // Fetcher function to retrieve the body of an announced block
}

// headerFilterTask represents a batch of headers needing fetcher filtering.
type headerFilterTask struct {
	peer    string          // The source peer of block headers
	headers []*types.Header // Collection of headers to filter
	time    time.Time       // Arrival time of the headers
}

// headerFilterTask represents a batch of block bodies (transactions and uncles)
// needing fetcher filtering.
type bodyFilterTask struct {
	peer         string                  // The source peer of block bodies
	transactions [][]types.CurrencyBlock // Collection of transactions per block bodies
	uncles       [][]*types.Header       // Collection of uncles per block bodies
	time         time.Time               // Arrival time of the blocks' contents
}

// inject represents a schedules import operation.
type inject struct {
	origin string
	block  *types.Block
}

// Fetcher is responsible for accumulating block announcements from various peers
// and scheduling them for retrieval.
type Fetcher struct {
	// Various event channels
	notify chan *announce
	inject chan *inject

	blockFilter  chan chan []*types.Block
	headerFilter chan chan *headerFilterTask
	bodyFilter   chan chan *bodyFilterTask

	done chan common.Hash
	quit chan struct{}
	MaxChainHeight  uint64 //lb
	// Announce states
	announces  map[string]int              // Per peer announce counts to prevent memory exhaustion
	announced  map[common.Hash][]*announce // Announced blocks, scheduled for fetching
	fetching   map[common.Hash]*announce   // Announced blocks, currently fetching
	fetched    map[common.Hash][]*announce // Blocks with headers fetched, scheduled for body retrieval
	completing map[common.Hash]*announce   // Blocks with headers, currently body-completing
	fetchHeaderNum  map[common.Hash]int  //请求header数
	fetchBlockNum  map[common.Hash]int
	retransannounced  map[common.Hash][]*announce
	//retransfetched    map[common.Hash][]*announce
	// Block cache
	queue  *prque.Prque            // Queue containing the import operations (block number sorted)
	queues map[string]int          // Per peer block counts to prevent memory exhaustion
	queued map[common.Hash]*inject // Set of already queued blocks (to dedupe imports)

	// Callbacks
	getBlock       blockRetrievalFn   // Retrieves a block from the local chain
	verifyHeader   headerVerifierFn   // Checks if a block's headers have a valid proof of work
	broadcastBlock blockBroadcasterFn // Broadcasts a block to connected peers
	chainHeight    chainHeightFn      // Retrieves the current chain's height
	insertChain    chainInsertFn      // Injects a batch of blocks into the chain
	dropPeer       peerDropFn         // Drops a peer for misbehaving

	// Testing hooks
	announceChangeHook func(common.Hash, bool) // Method to call upon adding or deleting a hash from the announce list
	queueChangeHook    func(common.Hash, bool) // Method to call upon adding or deleting a block from the import queue
	fetchingHook       func([]common.Hash)     // Method to call upon starting a block (man/61) or header (man/62) fetch
	completingHook     func([]common.Hash)     // Method to call upon starting a block body fetch (man/62)
	importedHook       func(*types.Block)      // Method to call upon successful block import (both man/61 and man/62)
}

// New creates a block fetcher to retrieve blocks based on hash announcements.
func New(getBlock blockRetrievalFn, verifyHeader headerVerifierFn, broadcastBlock blockBroadcasterFn, chainHeight chainHeightFn, insertChain chainInsertFn, dropPeer peerDropFn) *Fetcher {
	return &Fetcher{
		notify:         make(chan *announce),
		inject:         make(chan *inject),
		blockFilter:    make(chan chan []*types.Block),
		headerFilter:   make(chan chan *headerFilterTask),
		bodyFilter:     make(chan chan *bodyFilterTask),
		done:           make(chan common.Hash),
		quit:           make(chan struct{}),
		announces:      make(map[string]int),
		announced:      make(map[common.Hash][]*announce),
		fetching:       make(map[common.Hash]*announce),
		fetched:        make(map[common.Hash][]*announce),
		completing:     make(map[common.Hash]*announce),
		fetchHeaderNum: make(map[common.Hash]int),
		fetchBlockNum:  make(map[common.Hash]int),
		retransannounced: make(map[common.Hash][]*announce),
		//retransfetched : make(map[common.Hash][]*announce),
		queue:          prque.New(),
		queues:         make(map[string]int),
		queued:         make(map[common.Hash]*inject),
		getBlock:       getBlock,
		verifyHeader:   verifyHeader,
		broadcastBlock: broadcastBlock,
		chainHeight:    chainHeight,
		insertChain:    insertChain,
		dropPeer:       dropPeer,
	}
}

// Start boots up the announcement based synchroniser, accepting and processing
// hash notifications and block fetches until termination requested.
func (f *Fetcher) Start() {
	go f.loop()
}

// Stop terminates the announcement based synchroniser, canceling all pending
// operations.
func (f *Fetcher) Stop() {
	close(f.quit)
}

// Notify announces the fetcher of the potential availability of a new block in
// the network.
func (f *Fetcher) Notify(peer string, hash common.Hash, number uint64, time time.Time,
	headerFetcher headerRequesterFn, bodyFetcher bodyRequesterFn) error {
	block := &announce{
		hash:        hash,
		number:      number,
		time:        time,
		origin:      peer,
		fetchHeader: headerFetcher,
		fetchBodies: bodyFetcher,
	}
	select {
	case f.notify <- block:
		//log.Debug("fetcher notify ", "number", number, "hash", hash.Str(),"hash2",hash.Hex(),"peer",peer)
		return nil
	case <-f.quit:
		return errTerminated
	}
}

// Enqueue tries to fill gaps the the fetcher's future import queue.
func (f *Fetcher) Enqueue(peer string, block *types.Block) error {
	op := &inject{
		origin: peer,
		block:  block,
	}
	select {
	case f.inject <- op:
		return nil
	case <-f.quit:
		return errTerminated
	}
}

// FilterHeaders extracts all the headers that were explicitly requested by the fetcher,
// returning those that should be handled differently.
func (f *Fetcher) FilterHeaders(peer string, headers []*types.Header, time time.Time) []*types.Header {
	//log.Trace("download fetch Filtering headers", "peer", peer, "headers", len(headers))

	// Send the filter channel to the fetcher
	filter := make(chan *headerFilterTask)

	select {
	case f.headerFilter <- filter:
	case <-f.quit:
		return nil
	}
	// Request the filtering of the header list
	select {
	case filter <- &headerFilterTask{peer: peer, headers: headers, time: time}:
	case <-f.quit:
		return nil
	}
	// Retrieve the headers remaining after filtering
	select {
	case task := <-filter:
		return task.headers
	case <-f.quit:
		return nil
	}
}

// FilterBodies extracts all the block bodies that were explicitly requested by
// the fetcher, returning those that should be handled differently.
func (f *Fetcher) FilterBodies(peer string, transactions [][]types.CurrencyBlock, uncles [][]*types.Header, time time.Time) ([][]types.CurrencyBlock, [][]*types.Header) {
	log.Trace("download fetch Filtering bodies", "peer", peer, "txs", len(transactions), "uncles", len(uncles))

	// Send the filter channel to the fetcher
	filter := make(chan *bodyFilterTask)

	select {
	case f.bodyFilter <- filter:
	case <-f.quit:
		return nil, nil
	}
	// Request the filtering of the body list
	select {
	case filter <- &bodyFilterTask{peer: peer, transactions: transactions, uncles: uncles, time: time}:
	case <-f.quit:
		return nil, nil
	}
	// Retrieve the bodies remaining after filtering
	select {
	case task := <-filter:
		return task.transactions, task.uncles
	case <-f.quit:
		return nil, nil
	}
}

// Loop is the main fetcher loop, checking and processing various notification
// events.
func (f *Fetcher) loop() {
	// Iterate the block fetching until a quit is requested
	fetchTimer := time.NewTimer(0)
	completeTimer := time.NewTimer(0)
	updateQueueTimer := time.NewTicker(1 * time.Minute)

	for {
		// Clean up any expired block fetches
		for hash, announce := range f.fetching {
			if time.Since(announce.time) > fetchTimeout {
				log.Trace("fetch f.fetching timeout", " blockNum", announce.number, "hash", hash.String(), "origin peer", announce.origin)
				if f.getBlock(hash) != nil || f.queued[hash] != nil {
					f.forgetHash(hash)
					f.forgetRetransHash(hash,1)
					log.Trace("fetch f.fetching local has exist")
					continue
				}
				if reqnum,ok := f.fetchHeaderNum[hash]; ok 	{
					if reqnum < 3 {
						//f.announces[rand.Intn(len(announces))]
						left,ok := f.retransannounced[hash]
						log.Trace("fetch f.fetching timeout retry","len",len(left))
						if ok{
							newannounce := left[rand.Intn(len(left))]
							newannounce.time = time.Now()
							f.fetching[hash] = newannounce
							fetchHeader:= newannounce.fetchHeader							
							log.Trace("fetch f.fetching timeout retry require", " newannounce",newannounce.origin,"reqnum",reqnum,"len",len(left),"id",announce.origin,"hash",hash)
							go func(hashls common.Hash) {
								fetchHeader(hashls,newannounce.number) 							
							}(hash)
						}else {
							f.forgetHash(hash)
						}
						f.fetchHeaderNum[hash] = reqnum +1 
						
					} else {
						f.forgetHash(hash)
						f.forgetRetransHash(hash,1)
					}
				} else {
					f.forgetHash(hash)
				}
			}
		}
		//lb
		for hash, announce := range f.completing {
			if time.Since(announce.time) > 3 *fetchTimeout {
				log.Trace("fetch f.completing timeout body", " blockNum", announce.number, "hash", hash.String(), "origin peer", announce.origin)
				//f.forgetHash(hash)
				//f.forgetBlock(hash)
				if f.getBlock(hash) != nil || f.queued[hash] != nil{
					log.Trace("fetch f.fetching body local has exist")
					f.forgetHash(hash)
					f.forgetRetransHash(hash,1)
					//f.forgetBlock(hash)					
					continue
				}
				if reqnum,ok := f.fetchBlockNum[hash]; ok 	{
					if reqnum < 3 {
						//f.announces[rand.Intn(len(announces))]
						left,ok := f.retransannounced[hash]//fetched[hash]
						if ok{
							newannounce := left[rand.Intn(len(left))]
							//fetchHeader:= newannounce.fetchHeader
							newannounce.time = time.Now()
							newannounce.header = announce.header
							f.completing[hash] =  newannounce
							log.Trace("fetch f.fetching timeout retry require body", " newannounce",newannounce.origin,"reqnum",reqnum,"len",len(left),"id",announce.origin)
							hashs := make([]common.Hash,0)
							hashs = append(hashs,hash)
							go newannounce.fetchBodies(hashs)
						}else {
							f.forgetHash(hash)
							//f.forgetBlock(hash)	
						}
						f.fetchBlockNum[hash] = reqnum +1 
						
					} else {
						f.forgetHash(hash)
						f.forgetRetransHash(hash,1)
						//f.forgetBlock(hash)	
					}
				} else {
					f.forgetHash(hash)
					//f.forgetBlock(hash)	
				}
			}
		}
		// Import any queued blocks that could potentially fit
		height := f.chainHeight()
		for !f.queue.Empty() {
			op := f.queue.PopItem().(*inject)
			if f.queueChangeHook != nil {
				f.queueChangeHook(op.block.Hash(), false)
			}
			// If too high up the chain or phase, continue later
			number := op.block.NumberU64()
			if number > height+1 {
				f.queue.Push(op, -float32(op.block.NumberU64()))
				if f.queueChangeHook != nil {
					f.queueChangeHook(op.block.Hash(), true)
				}
				break
			}
			// Otherwise if fresh and still unknown, try and import
			hash := op.block.Hash()
			log.Trace("fetch queue update", "blockNumber", number, "hash", hash.String(), "curHeight", height)
			if number+maxUncleDist < height || f.getBlock(hash) != nil {
				f.forgetBlock(hash)
				continue
			}
			if number > height {
				f.insert(op.origin, op.block)
			}
		}
		// Wait for an outside event to occur
		select {
		case <-f.quit:
			// Fetcher terminating, abort all operations
			return
		case <-updateQueueTimer.C:
			log.Trace("fetch update operate which is before select will rebegin")

		case notification := <-f.notify:
			// A block was announced, make sure the peer isn't DOSing us
			propAnnounceInMeter.Mark(1)

			count := f.announces[notification.origin] + 1
			if count > hashLimit {
				log.Debug("Peer exceeded outstanding announces", "peer", notification.origin, "limit", hashLimit)
				propAnnounceDOSMeter.Mark(1)
				break
			}
			// If we have a valid block number, check that it's potentially useful
			if notification.number > 0 {
				if dist := int64(notification.number) - int64(f.chainHeight()); dist < -maxUncleDist || dist > maxQueueDist {
					log.Debug("Peer discarded announcement", "peer", notification.origin, "number", notification.number, "hash", notification.hash.String(), "distance", dist)
					propAnnounceDropMeter.Mark(1)
					break
				}
			}
			if aucnd, ok := f.retransannounced[notification.hash]; ok {
				if  len(aucnd) < 6 {
					f.retransannounced[notification.hash] = append(f.retransannounced[notification.hash], notification)
				}
			}
			// All is well, schedule the announce if block's not yet downloading
			if _, ok := f.fetching[notification.hash]; ok {
				break
			}
			if _, ok := f.completing[notification.hash]; ok {
				break
			}
			if _,ok := f.queued[notification.hash]; ok {
				break
			}
			//lb
			if len(f.announced[notification.hash]) > 16 {
				log.Debug("fetch f.announced[] too big,discard ", "hash", notification.hash.String(), "peer", notification.origin, "Number", notification.number)
				break
			}
			f.announces[notification.origin] = count
			f.announced[notification.hash] = append(f.announced[notification.hash], notification)
			if f.announceChangeHook != nil && len(f.announced[notification.hash]) == 1 {
				f.announceChangeHook(notification.hash, true)
			}
			if len(f.announced) == 1 {
				f.rescheduleFetch(fetchTimer)
			}

		case op := <-f.inject:
			// A direct block insertion was requested, try and fill any pending gaps
			propBroadcastInMeter.Mark(1)
			//lb 
			if len(f.queued) < 64 {
				log.Trace("download fetcher enqueue  inject")
				f.enqueue(op.origin, op.block)
			}

		case hash := <-f.done:
			// A pending import finished, remove all traces of the notification
			f.forgetHash(hash)
			f.forgetBlock(hash)

		case <-fetchTimer.C:
			// At least one block's timer ran out, check for needing retrieval
			request := make(map[string][]common.Hash)

			for hash, announces := range f.announced {
				if time.Since(announces[0].time) > arriveTimeout-gatherSlack /*&& f.fetchHeaderNum[hash] == 0 */ {
					// Pick a random peer to retrieve from, reset all others
					announce := announces[rand.Intn(len(announces))]
					f.retransannounced[hash] = f.announced[hash]
					f.forgetHash(hash) //
					//f.fetchHeaderNum[hash]= 1  //请求header数
					// If the block still didn't arrive, queue for fetching
					if f.getBlock(hash) == nil {
						if _,ok := f.queued[hash]; ok{
							log.Trace("download fetcher Fetching scheduled headers not ,hash has exist ", "hash", hash)
							f.forgetHash(hash) 
							f.forgetRetransHash(hash,1)
						} else {
							request[announce.origin] = append(request[announce.origin], hash)
							f.fetching[hash] = announce
							f.fetchHeaderNum[hash]= 1  //请求header数
							//f.retransannounced[hash] = f.announced[hash] 
						}
					} else {
						//lb 加
						//f.forgetHash(hash)
						//f.forgetBlock(hash)
						f.forgetRetransHash(hash,1)
					}					
				}
			}
			// Send out all block header requests
			for peer, hashes := range request {
				for _, hashd := range hashes {
					log.Trace("download fetcher Fetching scheduled headers", "peer", peer, "list hash", hashd.String())
				}

				// Create a closure of the fetch and schedule in on a new thread
				fetchHeader, hashes := f.fetching[hashes[0]].fetchHeader, hashes
				go func() {
					if f.fetchingHook != nil {
						f.fetchingHook(hashes)
					}
					for _, hash := range hashes {
						headerFetchMeter.Mark(1)
						fetchHeader(hash,1) // Suboptimal, but protocol doesn't allow batch header retrievals
					}
				}()
			}
			// Schedule the next fetch if blocks are still pending
			f.rescheduleFetch(fetchTimer)

		case <-completeTimer.C:
			// At least one header's timer ran out, retrieve everything
			request := make(map[string][]common.Hash)
			var flg int
			for hash, announces := range f.fetched {
				if f.fetchBlockNum[hash] == 0 {
					// Pick a random peer to retrieve from, reset all others
					announce := announces[rand.Intn(len(announces))]
					flg = 0
					f.forgetHash(hash)//		
					//f.fetchBlockNum[hash] = 1
					// If the block still didn't arrive, queue for completion
					if f.getBlock(hash) == nil &&  f.queued[hash] == nil {
						request[announce.origin] = append(request[announce.origin], hash)
						f.completing[hash] = announce
						flg = 1
						f.fetchBlockNum[hash] = 1
						//f.retransfetched[hash] = f.fetched[hash] 
					} else {
						//lb
						//forget
						f.forgetRetransHash(hash,1)
					}
					log.Trace("download fetcher Fetching scheduled bodies block number", "flg", flg, "number", announce.number)
				}
			}
			// Send out all block body requests
			for peer, hashes := range request {
				for _, hashd := range hashes {
					log.Trace("download fetcher Fetching scheduled bodies", "peer", peer, "list hash", hashd.String())
				}
				// Create a closure of the fetch and schedule in on a new thread
				if f.completingHook != nil {
					f.completingHook(hashes)
				}
				bodyFetchMeter.Mark(int64(len(hashes)))

				go f.completing[hashes[0]].fetchBodies(hashes)
			}
			// Schedule the next fetch if blocks are still pending
			f.rescheduleComplete(completeTimer)

		case filter := <-f.headerFilter:
			// Headers arrived from a remote peer. Extract those that were explicitly
			// requested by the fetcher, and return everything else so it's delivered
			// to other parts of the system.
			var task *headerFilterTask
			select {
			case task = <-filter:
			case <-f.quit:
				return
			}
			headerFilterInMeter.Mark(int64(len(task.headers)))

			// Split the batch of headers into unknown ones (to return to the caller),
			// known incomplete ones (requiring body retrievals) and completed blocks.
			unknown, incomplete, complete := []*types.Header{}, []*announce{}, []*types.Block{}
			for _, header := range task.headers {
				hash := header.Hash()

				// Filter fetcher-requested headers from other synchronisation algorithms
				if announce := f.fetching[hash]; announce != nil && announce.origin == task.peer && f.fetched[hash] == nil && f.completing[hash] == nil && f.queued[hash] == nil {
					// If the delivered header does not match the promised number, drop the announcer
					log.Trace("fetch header recv sucess","hash",hash)
					if header.Number.Uint64() != announce.number {
						log.Trace("Invalid block number header fetched", "peer", announce.origin, "hash", header.Hash().String(), "announced", announce.number, "provided", header.Number)
						f.dropPeer(announce.origin,0)
						f.forgetHash(hash)
						continue
					}
					f.forgetRetransHash(hash,0)//lb add 
					// Only keep if not imported by other means
					if f.getBlock(hash) == nil {
						announce.header = header
						announce.time = task.time
						isok := false
						log.Trace("fetch header recv sucess local has block not")
						// If the block is empty (header only), short circuit into the final import queue
						for _, coinRoot := range header.Roots {
							if coinRoot.TxHash != types.DeriveShaHash([]common.Hash{}) {
								isok = true
							}
						}
						if !isok {
							log.Trace("fetch header Block empty, skipping body retrieval", "peer", announce.origin, "number", header.Number, "hash", header.Hash())
							block := types.NewBlockWithHeader(header)
							block.ReceivedAt = task.time

							complete = append(complete, block)
							f.completing[hash] = announce
							continue
						}
						// Otherwise add to the list of blocks needing completion
						incomplete = append(incomplete, announce)
						
					} else {
						log.Trace("fetch Block already imported, discarding header", "peer", announce.origin, "number", header.Number, "hash", header.Hash().String())
						f.forgetHash(hash)
						//lb ？？没有forget block
						//f.forgetBlock(hash)
					}
				} else {
					// Fetcher doesn't know about it, add to the return list
					unknown = append(unknown, header)
				}
			}
			log.Trace("download fetch headerFilter after match", "len tx", len(unknown))
			headerFilterOutMeter.Mark(int64(len(unknown)))
			select {
			case filter <- &headerFilterTask{headers: unknown, time: task.time}:
			case <-f.quit:
				return
			}
			// Schedule the retrieved headers for body completion
			for _, announce := range incomplete {
				hash := announce.header.Hash()
				if _, ok := f.completing[hash]; ok {
					continue
				}
				f.fetched[hash] = append(f.fetched[hash], announce)
				if len(f.fetched) == 1 {
					f.rescheduleComplete(completeTimer)
				}
			}
			// Schedule the header-only blocks for import
			for _, block := range complete {
				if announce := f.completing[block.Hash()]; announce != nil {
					f.enqueue(announce.origin, block)
				}
			}

		case filter := <-f.bodyFilter:
			// Block bodies arrived, extract any explicitly requested blocks, return the rest
			var task *bodyFilterTask
			select {
			case task = <-filter:
			case <-f.quit:
				return
			}
			height := f.chainHeight()
			log.Trace("download fetch bodyFilter", "len tx", len(task.transactions), "task.peer", task.peer, "len .completing", len(f.completing), "height", height)
			bodyFilterInMeter.Mark(int64(len(task.transactions)))

			blocks := []*types.Block{}
			for i := 0; i < len(task.transactions) && i < len(task.uncles); i++ {
				// Match up a body to any possible completion request
				matched := false
				tmpmap := make(map[string]common.Hash)
				for _, txer := range task.transactions[i] {
					tmpmap[txer.CurrencyName] = types.DeriveShaHash(types.TxHashList(txer.Transactions.GetTransactions()))
					//log.Trace("download fetch bodyFilter tmpmap", "CurrencyName", txer.CurrencyName, "hash1", tmpmap[txer.CurrencyName], "hash2", types.DeriveShaHash(txer.Transactions.TxHashs))
				}
				for hash, announce := range f.completing {
					if f.queued[hash] == nil {
						isok := true
						if announce.header != nil {
							for _, coinHeader := range announce.header.Roots {
								if announce.header != nil {
								txnHash := tmpmap[coinHeader.Cointyp]
								if txnHash == (common.Hash{}) {
									txnHash = types.DeriveShaHash([]common.Hash{})
								}
								uncleHash := types.CalcUncleHash(task.uncles[i])
								log.Trace("download fetch bodyFilter map", "hash", hash, "Cointyp", coinHeader.Cointyp, "announce", coinHeader.TxHash, "txnHash", txnHash, "origin id", announce.origin, "blockNum", announce.number)
								if txnHash != coinHeader.TxHash || uncleHash != announce.header.UncleHash || announce.origin != task.peer {
									log.Trace("fetchr err", "fetch body txhash != header txhash.  txnHash", txnHash.String(), "header txHash", coinHeader.TxHash.String(), "uncleHash", uncleHash.String(),
										"announce.header.UncleHash", announce.header.UncleHash.String(), "announce.origin", announce.origin, "task.peer", task.peer)
									isok = false
									break
								}
								} else {
									log.Trace("download fetch bodyFilter announce.header is nil", "number",announce.number)
								}
							}
						} else {
							log.Trace("download fetch bodyFilter  announce.header error","number",announce.number,"hash",announce.hash,"origin",announce.origin)
							f.forgetHash(hash)							
							continue
						}
						if isok {
							// Mark the body matched, reassemble if still unknown
							matched = true
							if f.getBlock(hash) == nil {
								log.Trace("download fetch bodyFilter getBlock")
								block := types.NewBlockWithHeader(announce.header).WithBody(task.transactions[i], task.uncles[i])
								block.ReceivedAt = task.time

								blocks = append(blocks, block)
							} else {
								log.Trace("download fetch bodyFilter forgetHash","hash", hash)
								f.forgetHash(hash)
							}						
							break
						}
					} else {
						log.Trace("fetch queued has this hash", "hash", hash.String())
					}
					//lb
					if dist := int64(announce.number) - int64(height); dist <= 0/*-maxUncleDist*/ || dist > maxQueueDist {
						log.Trace("fetch f.completing will delete hash", "hash", hash, "number", announce.number,"height",height)
						f.forgetHash(hash)
						//f.forgetBlock(hash)
					}
				}
				if matched {
					task.transactions = append(task.transactions[:i], task.transactions[i+1:]...)
					task.uncles = append(task.uncles[:i], task.uncles[i+1:]...)
					i--
					continue
				}
			}
			log.Trace("download fetch bodyFilter after match", "len tx", len(task.transactions))
			bodyFilterOutMeter.Mark(int64(len(task.transactions)))
			select {
			case filter <- task:
			case <-f.quit:
				return
			}
			// Schedule the retrieved blocks for ordered import
			for _, block := range blocks {
				if announce := f.completing[block.Hash()]; announce != nil {
					log.Trace("download fetch before  enqueue","block",block.NumberU64())
					f.enqueue(announce.origin, block)
						//lb
					f.forgetHash(block.Hash())
					f.forgetRetransHash(block.Hash(),1)
				}
			}
		}
	}
}

// rescheduleFetch resets the specified fetch timer to the next announce timeout.
func (f *Fetcher) rescheduleFetch(fetch *time.Timer) {
	// Short circuit if no blocks are announced
	if len(f.announced) == 0 {
		log.Trace("download fetch rescheduleFetch time shutdown")
		return
	}

	/*if len(f.announced) == 1 {
		for  hash,_ := range f.announced {
	       if f.fetchHeaderNum[hash] > 0  {
				log.Trace("download fetch rescheduleFetch time shutdown")
				return
		   }
		}
	}*/
	// Otherwise find the earliest expiring announcement
	earliest := time.Now()
	for _, announces := range f.announced {
		if earliest.After(announces[0].time) {
			earliest = announces[0].time
		}
	}
	
	gaptime := arriveTimeout - time.Since(earliest)
	if gaptime < 200*time.Millisecond{
		gaptime  =  200*time.Millisecond
	}
	log.Trace("download fetch rescheduleFetch time","time",arriveTimeout - time.Since(earliest),"gaptime",gaptime)
	fetch.Reset(gaptime)
	//fetch.Reset(arriveTimeout - time.Since(earliest))
}

// rescheduleComplete resets the specified completion timer to the next fetch timeout.
func (f *Fetcher) rescheduleComplete(complete *time.Timer) {
	// Short circuit if no headers are fetched
	if len(f.fetched) == 0 {
		return
	}
	// Otherwise find the earliest expiring announcement
	earliest := time.Now()
	for _, announces := range f.fetched {
		if earliest.After(announces[0].time) {
			earliest = announces[0].time
		}
	}
	//complete.Reset(gatherSlack - time.Since(earliest))
	gaptime := arriveTimeout - time.Since(earliest)
	if gaptime < 100*time.Millisecond{
		gaptime  =  100*time.Millisecond
	}
	log.Trace("download fetch rescheduleComplete time","time",arriveTimeout - time.Since(earliest),"gaptime",gaptime)
	complete.Reset(gaptime)
}

// enqueue schedules a new future import operation, if the block to be imported
// has not yet been seen.
func (f *Fetcher) enqueue(peer string, block *types.Block) {
	hash := block.Hash()

	// Ensure the peer isn't DOSing us
	count := f.queues[peer] + 1
	if count > blockLimit {
		log.Debug("Discarded propagated block, exceeded allowance", "peer", peer, "number", block.Number(), "hash", hash.String(), "limit", blockLimit)
		propBroadcastDOSMeter.Mark(1)
		f.forgetHash(hash)
		return
	}
	// Discard any past or too distant blocks
	/*if dist := int64(block.NumberU64()) - int64(f.chainHeight()); dist < -maxUncleDist || dist > maxQueueDist {
		log.Debug("Discarded propagated block, too far away", "peer", peer, "number", block.Number(), "hash", hash.String(), "distance", dist)
		propBroadcastDropMeter.Mark(1)
		f.forgetHash(hash)
		return
	}*/
	// Schedule the block for future importing
	if _, ok := f.queued[hash]; !ok {
		op := &inject{
			origin: peer,
			block:  block,
		}
		f.queues[peer] = count
		f.queued[hash] = op
		f.queue.Push(op, -float32(block.NumberU64()))
		if f.queueChangeHook != nil {
			f.queueChangeHook(op.block.Hash(), true)
		}
		log.Debug("Queued propagated block", "peer", peer, "number", block.Number(), "hash", hash.String(), "queued", f.queue.Size())
	}
}

// insert spawns a new goroutine to run a block insertion into the chain. If the
// block's number is at the same height as the current import phase, it updates
// the phase states accordingly.
func (f *Fetcher) insert(peer string, block *types.Block) {
	hash := block.Hash()

	// Run the import on a new thread
	log.Debug("feth Importing propagated block", "peer", peer, "number", block.Number(), "hash", hash.String())
	go func() {
		defer func() { f.done <- hash }()

		// If the parent's unknown, abort insertion
		parent := f.getBlock(block.ParentHash())
		if parent == nil {
			log.Debug("Unknown parent of propagated block", "peer", peer, "number", block.Number(), "hash", hash.String(), "parent", block.ParentHash().String())
			return
		}
		// Quickly validate the header and propagate the block if it passes
		switch err := f.verifyHeader(block.Header()); err {
		case nil:
			// All ok, quickly propagate to our peers
			propBroadcastOutTimer.UpdateSince(block.ReceivedAt)
			go f.broadcastBlock(block, true)

		case consensus.ErrFutureBlock:
			// Weird future block, don't fail, but neither propagate

		default:
			// Something went very wrong, drop the peer
			log.Debug("Propagated block verification failed", "peer", peer, "number", block.Number(), "hash", hash.String(), "err", err)
			f.dropPeer(peer,0)
			return
		}
		// Run the actual import and log any issues
		if _, err := f.insertChain(types.Blocks{block}); err != nil {
			log.Debug("Propagated block import failed", "peer", peer, "number", block.Number(), "hash", hash.String(), "err", err)
			return
		}
		// If import succeeded, broadcast the block
		propAnnounceOutTimer.UpdateSince(block.ReceivedAt)
		go f.broadcastBlock(block, false)

		// Invoke the testing hook if needed
		if f.importedHook != nil {
			f.importedHook(block)
		}
	}()
}

func (f *Fetcher) forgetRetransHash(hash common.Hash,flg int) {
	log.Debug("fether head forgetHash", "hash", hash.String())	
	delete(f.fetchHeaderNum,hash)
	if flg > 0 {
		delete(f.retransannounced, hash)
		delete(f.fetchBlockNum,hash)
	}
}
// forgetHash removes all traces of a block announcement from the fetcher's
// internal state.
func (f *Fetcher) forgetHash(hash common.Hash) {
	// Remove all pending announces and decrement DOS counters
	log.Debug("fether forgetHash", "hash", hash.String())
	for _, announce := range f.announced[hash] {
		f.announces[announce.origin]--
		if f.announces[announce.origin] == 0 {
			delete(f.announces, announce.origin)
		}
	}
	delete(f.announced, hash)
	if f.announceChangeHook != nil {
		f.announceChangeHook(hash, false)
	}
	// Remove any pending fetches and decrement the DOS counters
	if announce := f.fetching[hash]; announce != nil {
		f.announces[announce.origin]--
		if f.announces[announce.origin] == 0 {
			delete(f.announces, announce.origin)
		}
		delete(f.fetching, hash)
	}

	// Remove any pending completion requests and decrement the DOS counters
	for _, announce := range f.fetched[hash] {
		f.announces[announce.origin]--
		if f.announces[announce.origin] == 0 {
			delete(f.announces, announce.origin)
		}
	}
	delete(f.fetched, hash)

	// Remove any pending completions and decrement the DOS counters
	if announce := f.completing[hash]; announce != nil {
		f.announces[announce.origin]--
		if f.announces[announce.origin] == 0 {
			delete(f.announces, announce.origin)
		}
		delete(f.completing, hash)
	}
}

// forgetBlock removes all traces of a queued block from the fetcher's internal
// state.
func (f *Fetcher) forgetBlock(hash common.Hash) {
	if insert := f.queued[hash]; insert != nil {
		f.queues[insert.origin]--
		if f.queues[insert.origin] == 0 {
			delete(f.queues, insert.origin)
		}
		delete(f.queued, hash)
	}
}
