// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

// Contains the block download scheduler to collect download tasks and schedule
// them in an ordered, and throttled way.

package downloader

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/metrics"
	"gopkg.in/karalabe/cookiejar.v2/collections/prque"
)

var (
	blockCacheItems      = 1200             //1024             //lb 8192             // Maximum number of blocks to cache before throttling the download
	blockCacheMemory     = 64 * 1024 * 1024 // Maximum amount of memory to use for block caching
	blockCacheSizeWeight = 0.1              // Multiplier to approximate the average block size based on past ones
	QSingleBlockStore    = false
	NumBlockInfoFromIpfs = 0
)

var (
	errNoFetchesPending = errors.New("no fetches pending")
	errStaleDelivery    = errors.New("stale delivery")
)

// fetchRequest is a currently running data retrieval operation.
type fetchRequest struct {
	Peer    *peerConnection // Peer to which the request was sent
	From    uint64          // [man/62] Requested chain element index (used for skeleton fills only)
	Headers []*types.Header // [man/62] Requested headers, sorted by request order
	Time    time.Time       // Time when the request was made
}

// fetchResult is a struct collecting partial results from data fetchers until
// all outstanding pieces complete and the result as a whole can be processed.
type fetchResult struct {
	Pending int // Number of data fetches still pending
	Flag    int
	Hash    common.Hash // Hash of the header to prevent recalculating

	Header       *types.Header
	Uncles       []*types.Header
	Transactions []types.CurrencyBlock //CoinSelfTransaction
	Receipts     []types.CoinReceipts
}
type ipfsRequest struct {
	Header  *types.Header
	Pending int // Number of data fetches still pending
	Flag    int
	Time    time.Time
}

// queue represents hashes that are either need fetching or are being fetched
type queue struct {
	mode SyncMode // Synchronisation mode to decide on the block parts to schedule for fetching

	// Headers are "special", they download in batches, supported by a skeleton chain
	headerHead      common.Hash                    // [man/62] Hash of the last queued header to verify order
	headerTaskPool  map[uint64]*types.Header       // [man/62] Pending header retrieval tasks, mapping starting indexes to skeleton headers
	headerTaskQueue *prque.Prque                   // [man/62] Priority queue of the skeleton indexes to fetch the filling headers for
	headerPeerMiss  map[string]map[uint64]struct{} // [man/62] Set of per-peer header batches known to be unavailable
	headerPendPool  map[string]*fetchRequest       // [man/62] Currently pending header retrieval operations
	headerResults   []*types.Header                // [man/62] Result cache accumulating the completed headers
	headerProced    int                            // [man/62] Number of headers already processed from the results
	headerOffset    uint64                         // [man/62] Number of the first header in the result cache
	headerContCh    chan bool                      // [man/62] Channel to notify when header download finishes
	headerReqExpireNum    int
	headerSkeletonErrNum  int 
	// All data retrievals below are based on an already assembles header chain
	blockTaskPool  map[common.Hash]*types.Header // [man/62] Pending block (body) retrieval tasks, mapping hashes to headers
	blockTaskQueue *prque.Prque                  // [man/62] Priority queue of the headers to fetch the blocks (bodies) for
	blockPendPool  map[string]*fetchRequest      // [man/62] Currently pending block (body) retrieval operations
	blockDonePool  map[common.Hash]struct{}      // [man/62] Set of the completed block (body) fetches
	blockIpfsPool  map[uint64]*ipfsRequest       //*types.Header

	receiptTaskPool  map[common.Hash]*types.Header // [man/63] Pending receipt retrieval tasks, mapping hashes to headers
	receiptTaskQueue *prque.Prque                  // [man/63] Priority queue of the headers to fetch the receipts for
	receiptPendPool  map[string]*fetchRequest      // [man/63] Currently pending receipt retrieval operations
	receiptDonePool  map[common.Hash]struct{}      // [man/63] Set of the completed receipt fetches

	resultCache  []*fetchResult // Downloaded but not yet delivered fetch results
	resultOffset uint64         // Offset of the first cached fetch result in the block chain
	resultBegin  uint64         //lb
	recvheadNum  int            //lb
	resultWait   bool           //lb
	getBlock     blockQRetrievalFn
	resultSize   common.StorageSize // Approximate size of a block (exponential moving average)

	lock   *sync.Mutex
	active *sync.Cond
	closed bool
}

// newQueue creates a new download queue for scheduling block retrieval.
func newQueue(getBlock blockQRetrievalFn) *queue {
	lock := new(sync.Mutex)
	return &queue{
		headerPendPool:   make(map[string]*fetchRequest),
		headerContCh:     make(chan bool),
		blockTaskPool:    make(map[common.Hash]*types.Header),
		blockTaskQueue:   prque.New(),
		blockPendPool:    make(map[string]*fetchRequest),
		blockDonePool:    make(map[common.Hash]struct{}),
		blockIpfsPool:    make(map[uint64]*ipfsRequest), //*types.Header),
		receiptTaskPool:  make(map[common.Hash]*types.Header),
		receiptTaskQueue: prque.New(),
		receiptPendPool:  make(map[string]*fetchRequest),
		receiptDonePool:  make(map[common.Hash]struct{}),
		resultCache:      make([]*fetchResult, blockCacheItems),
		active:           sync.NewCond(lock),
		getBlock:         getBlock,
		lock:             lock,
	}
}

// Reset clears out the queue contents.
func (q *queue) Reset() {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.closed = false
	q.mode = FullSync

	q.headerHead = common.Hash{}
	q.headerPendPool = make(map[string]*fetchRequest)
	q.headerReqExpireNum = 0
	q.blockTaskPool = make(map[common.Hash]*types.Header)
	q.blockTaskQueue.Reset()
	q.blockPendPool = make(map[string]*fetchRequest)
	q.blockDonePool = make(map[common.Hash]struct{})
	q.blockIpfsPool = make(map[uint64]*ipfsRequest) //*types.Header)

	q.receiptTaskPool = make(map[common.Hash]*types.Header)
	q.receiptTaskQueue.Reset()
	q.receiptPendPool = make(map[string]*fetchRequest)
	q.receiptDonePool = make(map[common.Hash]struct{})

	q.resultCache = make([]*fetchResult, blockCacheItems)
	q.resultOffset = 0
}

// Close marks the end of the sync, unblocking WaitResults.
// It may be called even if the queue is already closed.
func (q *queue) Close() {
	q.lock.Lock()
	q.closed = true
	q.lock.Unlock()
	q.active.Broadcast()
}

// PendingHeaders retrieves the number of header requests pending for retrieval.
func (q *queue) PendingHeaders() int {
	q.lock.Lock()
	defer q.lock.Unlock()

	return q.headerTaskQueue.Size()
}

// PendingBlocks retrieves the number of block (body) requests pending for retrieval.
func (q *queue) PendingBlocks() int {
	q.lock.Lock()
	defer q.lock.Unlock()

	return q.blockTaskQueue.Size()
}

// PendingReceipts retrieves the number of block receipts pending for retrieval.
func (q *queue) PendingReceipts() int {
	q.lock.Lock()
	defer q.lock.Unlock()

	return q.receiptTaskQueue.Size()
}

// InFlightHeaders retrieves whether there are header fetch requests currently
// in flight.
func (q *queue) InFlightHeaders() bool {
	q.lock.Lock()
	defer q.lock.Unlock()

	return len(q.headerPendPool) > 0
}

// InFlightBlocks retrieves whether there are block fetch requests currently in
// flight.
func (q *queue) InFlightBlocks() bool {
	q.lock.Lock()
	defer q.lock.Unlock()

	return len(q.blockPendPool) > 0
}

func (q *queue) BlockIpfsPoolBlocksNum() int {
	//q.lock.Lock()
	//defer q.lock.Unlock()
	return len(q.blockIpfsPool)
}
func (q *queue) BlockIpfsInsetPool(blockNum uint64, header *types.Header, pending, flag int) {
	//q.lock.Lock()
	//defer q.lock.Unlock()
	//q.blockIpfsPool[blockNum] = header
	tmp := &ipfsRequest{
		Header:  header,
		Pending: pending,
		Flag:    flag,
		Time:    time.Now(),
	}
	q.blockIpfsPool[blockNum] = tmp
}
func (q *queue) BlockIpfsdeletePool(blockNum uint64) {
	//q.lock.Lock()
	//defer q.lock.Unlock()
	delete(q.blockIpfsPool, blockNum)
}

// InFlightReceipts retrieves whether there are receipt fetch requests currently
// in flight.
func (q *queue) InFlightReceipts() bool {
	q.lock.Lock()
	defer q.lock.Unlock()

	return len(q.receiptPendPool) > 0
}

// Idle returns if the queue is fully idle or has some data still inside.
func (q *queue) Idle() bool {
	q.lock.Lock()
	defer q.lock.Unlock()

	queued := q.blockTaskQueue.Size() + q.receiptTaskQueue.Size()
	pending := len(q.blockPendPool) + len(q.receiptPendPool)
	cached := len(q.blockDonePool) + len(q.receiptDonePool)

	return (queued + pending + cached) == 0
}

// ShouldThrottleBlocks checks if the download should be throttled (active block (body)
// fetches exceed block cache).
func (q *queue) ShouldThrottleBlocks() bool {
	q.lock.Lock()
	defer q.lock.Unlock()

	return q.resultSlots(q.blockPendPool, q.blockDonePool) <= 0
}

// ShouldThrottleReceipts checks if the download should be throttled (active receipt
// fetches exceed block cache).
func (q *queue) ShouldThrottleReceipts() bool {
	q.lock.Lock()
	defer q.lock.Unlock()

	return q.resultSlots(q.receiptPendPool, q.receiptDonePool) <= 0
}

// resultSlots calculates the number of results slots available for requests
// whilst adhering to both the item and the memory limit too of the results
// cache.
func (q *queue) resultSlots(pendPool map[string]*fetchRequest, donePool map[common.Hash]struct{}) int {
	// Calculate the maximum length capped by the memory limit
	limit := len(q.resultCache)
	if common.StorageSize(len(q.resultCache))*q.resultSize > common.StorageSize(blockCacheMemory) {
		limit = int((common.StorageSize(blockCacheMemory) + q.resultSize - 1) / q.resultSize)
	}
	// Calculate the number of slots already finished
	finished := 0
	for _, result := range q.resultCache[:limit] {
		if result == nil {
			break
		}
		if _, ok := donePool[result.Hash]; ok {
			finished++
		}
	}
	// Calculate the number of slots currently downloading
	pending := 0
	for _, request := range pendPool {
		for _, header := range request.Headers {
			if header.Number.Uint64() < q.resultOffset+uint64(limit) {
				pending++
			}
		}
	}
	// Return the free slots to distribute
	return limit - finished - pending
}

// ScheduleSkeleton adds a batch of header retrieval tasks to the queue to fill
// up an already retrieved header skeleton.
func (q *queue) ScheduleSkeleton(from uint64, skeleton []*types.Header) {
	q.lock.Lock()
	defer q.lock.Unlock()
	log.Trace("Filling up skeleton ScheduleSkeleton ", "from", from, "len", len(skeleton))
	// No skeleton retrieval can be in progress, fail hard if so (huge implementation bug)
	if q.headerResults != nil {
		panic("skeleton assembly already in progress")
	}
	// Schedule all the header retrieval tasks for the skeleton assembly
	q.headerTaskPool = make(map[uint64]*types.Header)
	q.headerTaskQueue = prque.New()
	q.headerPeerMiss = make(map[string]map[uint64]struct{}) // Reset availability to correct invalid chains
	q.headerResults = make([]*types.Header, len(skeleton)*MaxHeaderFetch)
	q.headerProced = 0
	q.headerOffset = from
	q.headerContCh = make(chan bool, 1)

	for i, header := range skeleton {
		index := from + uint64(i*MaxHeaderFetch)

		q.headerTaskPool[index] = header
		q.headerTaskQueue.Push(index, -float32(index))
		//log.Trace("Filling  ScheduleSkeleton ", "head", header.Number.Uint64)
	}
}

// RetrieveHeaders retrieves the header chain assemble based on the scheduled
// skeleton.
func (q *queue) RetrieveHeaders() ([]*types.Header, int) {
	q.lock.Lock()
	defer q.lock.Unlock()

	headers, proced := q.headerResults, q.headerProced
	q.headerResults, q.headerProced = nil, 0

	return headers, proced
}

// Schedule adds a set of headers for the download queue for scheduling, returning
// the new headers encountered.
func (q *queue) Schedule(headers []*types.Header, from uint64, bodymode int) []*types.Header {
	q.lock.Lock()
	defer q.lock.Unlock()
	log.Info("download queue  enter Schedule", "from", from, "len", len(headers))
	// Insert all the headers prioritised by the contained block number
	inserts := make([]*types.Header, 0, len(headers))
	for _, header := range headers {
		// Make sure chain order is honoured and preserved throughout
		hash := header.Hash()
		if header.Number == nil || header.Number.Uint64() != from {
			log.Warn("Header broke chain ordering", "number", header.Number, "hash", hash, "expected", from)
			break
		}
		if q.headerHead != (common.Hash{}) && q.headerHead != header.ParentHash {
			log.Warn("Header broke chain ancestry", "number", header.Number, "hash", hash)
			break
		}
		// Make sure no duplicate requests are executed
		if _, ok := q.blockTaskPool[hash]; ok {
			log.Warn("Header  already scheduled for block fetch", "number", header.Number, "hash", hash)
			continue
		}
		if _, ok := q.receiptTaskPool[hash]; ok {
			log.Warn("Header already scheduled for receipt fetch", "number", header.Number, "hash", hash)
			continue
		}
		// Queue the header for content retrieval

		if bodymode == 2 {
			//q.blockPendPool[]
		} else {
			q.blockTaskPool[hash] = header
			q.blockTaskQueue.Push(header, -float32(header.Number.Uint64()))
			if q.mode == FastSync {
				q.receiptTaskPool[hash] = header
				q.receiptTaskQueue.Push(header, -float32(header.Number.Uint64()))
			}
		}

		/*if q.mode == FastSync {
			q.receiptTaskPool[hash] = header
			q.receiptTaskQueue.Push(header, -float32(header.Number.Uint64()))
		}*/
		inserts = append(inserts, header)
		q.headerHead = hash
		from++
	}
	return inserts
}

// Results retrieves and permanently removes a batch of fetch results from
// the cache. the result slice will be empty if the queue has been closed.
func (q *queue) Results(block bool) []*fetchResult {
	q.lock.Lock()
	defer q.lock.Unlock()

	// Count the number of items available for processing
	nproc := q.countProcessableItems()
	if nproc == 0 {
		if q.resultCache[0] != nil && q.resultCache[0].Flag == 2 && q.resultCache[0].Pending > 0 { //qian20个还没收到body/receipt
			if q.resultCache[10] != nil { //保证后面有数据才认为0丢失
				hash := q.resultCache[0].Header.Hash()
				_, ok := q.blockTaskPool[hash]

				if !ok {
					log.Warn("download  Results check but begining is to receive, begin origin Req", "blockNum", q.resultCache[0].Header.Number.Uint64())
					q.blockTaskPool[hash] = q.resultCache[0].Header
					q.blockTaskQueue.Push(q.resultCache[0].Header, -float32(q.resultCache[0].Header.Number.Uint64()))
					if q.resultCache[0].Pending > 1 {
						q.receiptTaskPool[hash] = q.resultCache[0].Header
						q.receiptTaskQueue.Push(q.resultCache[0].Header, -float32(q.resultCache[0].Header.Number.Uint64()))
					}
				}
			}
		}
	}
	for nproc == 0 && !q.closed {
		if !block {
			return nil
		}
		log.Trace("download queue Wait1", "nproc", nproc)
		q.active.Wait()
		log.Trace("download queue Wait2", "nproc", nproc)
		nproc = q.countProcessableItems()
	}
	// Since we have a batch limit, don't pull more into "dangling" memory
	if nproc > maxResultsProcess {
		nproc = maxResultsProcess
	}
	results := make([]*fetchResult, nproc)
	copy(results, q.resultCache[:nproc])
	log.Warn("download queue  Results", "len ", len(results), "nproc", nproc, "closed", q.closed)
	if len(results) > 0 {
		// Mark results as done before dropping them from the cache.
		for _, result := range results {
			hash := result.Header.Hash()
			delete(q.blockDonePool, hash)
			delete(q.receiptDonePool, hash)
		}
		// Delete the results from the cache and clear the tail.
		copy(q.resultCache, q.resultCache[nproc:])
		for i := len(q.resultCache) - nproc; i < len(q.resultCache); i++ {
			q.resultCache[i] = nil
		}
		// Advance the expected block number of the first cache entry.
		q.resultOffset += uint64(nproc)

		// Recalculate the result item weights to prevent memory exhaustion
		for _, result := range results {
			size := result.Header.Size()
			for _, uncle := range result.Uncles {
				size += uncle.Size()
			}
			for _, receipt := range result.Receipts {
				for _, r := range receipt.Receiptlist {
					size += r.Size()
				}
			}
			for _, currblk := range result.Transactions {
				for _, tx := range currblk.Transactions.GetTransactions() {
					size += tx.Size()
				}
			}
			q.resultSize = common.StorageSize(blockCacheSizeWeight)*size + (1-common.StorageSize(blockCacheSizeWeight))*q.resultSize
		}
	}
	return results
}

// countProcessableItems counts the processable items.
func (q *queue) countProcessableItems() int {
	for i, result := range q.resultCache {
		if result == nil || result.Pending > 0 {
			return i
		}
	}
	return len(q.resultCache)
}
func (q *queue) leftResultCaceh() (int, uint64) {
	length := len(q.resultCache)
	var left int
	for i := length - 1; i > 0; i-- {
		if q.resultCache[i] != nil {
			//return left
			return left, q.resultCache[i].Header.Number.Uint64()
		}
		left++
	}
	return left, 0
}

// ReserveHeaders reserves a set of headers for the given peer, skipping any
// previously failed batches.
func (q *queue) ReserveHeaders(p *peerConnection, count int, ipfsmode int) (*fetchRequest, bool) {
	q.lock.Lock()
	defer q.lock.Unlock()

	/*if(q.headerReqExpireNum > 100){
		log.Warn("download queue  ReserveHeaders  exipre too many,then exit download")
		return nil,false
	}*/

	// Short circuit if the peer's already downloading something (sanity check to
	// not corrupt state)
	if _, ok := q.headerPendPool[p.id]; ok {
		return nil, false
	}
	//if d.bIpfsDownload == 2 && (d.mode == FullSync || d.mode == FastSync)
	/*if bHead {
		leftLen := q.leftResultCaceh()
		if leftLen < MaxHeaderFetch+3 {
			log.Debug("download queue ReserveHeaders left resulchche ", "leftlen", leftLen)
			return nil
		}
	}*/
	// Retrieve a batch of hashes, skipping previously failed ones
	send, skip := uint64(0), []uint64{}
	for send == 0 && !q.headerTaskQueue.Empty() {
		from, _ := q.headerTaskQueue.Pop()
		if q.headerPeerMiss[p.id] != nil {
			if _, ok := q.headerPeerMiss[p.id][from.(uint64)]; ok {
				skip = append(skip, from.(uint64))
				continue
			}
		}
		send = from.(uint64)
	}
	// Merge all the skipped batches back
	for _, from := range skip {
		q.headerTaskQueue.Push(from, -float32(from))
	}
	//lb
	/*leftLen, NumberB := q.leftResultCaceh()
	if NumberB != 0 {
		if NumberB < send && leftLen < MaxHeaderFetch+3 { //send 可能为重传
			q.headerTaskQueue.Push(send, -float32(send))
			log.Debug("download queue ReserveHeaders left resulchche ", "leftlen", leftLen, "NumberB", NumberB, "send", send)
			return nil
		}
	}*/
	// Assemble and return the block download request
	if send == 0 {
		return nil, false
	}
	//1 miao 10 ci, xiaoshi 36000
	index := int(int64(send)-int64(q.resultOffset)) + MaxHeaderFetch
	if index >= len(q.resultCache) || index < 0 {
		q.headerTaskQueue.Push(send, -float32(send))
		log.Debug("download queue ReserveHeaders left resultcache ", "send", send, "q.resultOffset", q.resultOffset, "index", index)
		gCurDownloadHeadReqWaitNum++
		if gCurDownloadHeadReqWaitNum > 72000 {
			log.Error("download queue ReserveHeaders left resultcache too many")
			return nil, false
		}
		return nil, true
	}

	if ipfsmode == 1 && gIpfsProcessBlockNumber != 0 { //广播节点
		if int(gCurDownloadHeadReqBeginNum-gIpfsProcessBlockNumber) > blockCacheItems {
			q.headerTaskQueue.Push(send, -float32(send))
			log.Trace("fetchParts Data Reserve header too big,wait for ippfs stroe", "HeadReqBeginNum", gCurDownloadHeadReqBeginNum, "IppfsProcessBlock", gIpfsProcessBlockNumber)
			gCurDownloadHeadReqWaitNum++
			if gCurDownloadHeadReqWaitNum > 72000 {
				log.Error("download queue wait for ippfs stroe too many")
				return nil, false
			}
			return nil, true
		}
		gCurDownloadHeadReqBeginNum = send
	}
	gCurDownloadHeadReqWaitNum = 0
	request := &fetchRequest{
		Peer: p,
		From: send,
		Time: time.Now(),
	}
	q.headerPendPool[p.id] = request
	return request, false
}
func (q *queue) ReserveLength(taskQueue *prque.Prque) int {
	q.lock.Lock()
	len := taskQueue.Size()
	q.lock.Unlock()
	return len
} 
// ReserveBodies reserves a set of body fetches for the given peer, skipping any
// previously failed downloads. Beside the next batch of needed fetches, it also
// returns a flag whether empty blocks were queued requiring processing.
func (q *queue) ReserveBodies(p *peerConnection, count int) (*fetchRequest, bool, bool, error) {
	isNoop := func(header *types.Header) bool {
		//flag:=true
		for _, cr := range header.Roots {
			if cr.TxHash != types.EmptyRootHash {
				return false
			}
		}
		return true
		//return header.TxHash == types.EmptyRootHash && header.UncleHash == types.EmptyUncleHash
	}
	q.lock.Lock()
	defer q.lock.Unlock()
	log.Debug("download queue ReserveBodies  = ", "count", count)
	return q.reserveHeaders(p, count, q.blockTaskPool, q.blockTaskQueue, q.blockPendPool, q.blockDonePool, isNoop)
}

// ReserveReceipts reserves a set of receipt fetches for the given peer, skipping
// any previously failed downloads. Beside the next batch of needed fetches, it
// also returns a flag whether empty receipts were queued requiring importing.
func (q *queue) ReserveReceipts(p *peerConnection, count int) (*fetchRequest, bool, bool, error) {
	isNoop := func(header *types.Header) bool {
		//return header.ReceiptHash == types.EmptyRootHash
		for _, cr := range header.Roots {
			if cr.TxHash != types.EmptyRootHash {
				return false
			}
		}
		return true
	}
	q.lock.Lock()
	defer q.lock.Unlock()
	log.Debug("download queue  ReserveReceipts  = ", "count", count)
	return q.reserveHeaders(p, count, q.receiptTaskPool, q.receiptTaskQueue, q.receiptPendPool, q.receiptDonePool, isNoop)
}
func (q *queue) Reserveipfs(recvheader []*types.Header, origin, remote uint64) ([]BlockIpfsReq, error) {
	//progress := false
	q.lock.Lock()
	defer q.lock.Unlock()
	//var curHeigh uint64
	if len(recvheader) == 0 {
		return nil, fmt.Errorf("Reserveipfs nil error")
	}
	//RequsetHeader := make([]*types.Header, 0)
	RequsetHeader := make([]BlockIpfsReq, 0)
	progress := false
	isNil := func(header *types.Header) bool {
		isOK := true
		for _, hr := range header.Roots {
			if hr.TxHash != types.EmptyRootHash || header.UncleHash != types.EmptyUncleHash {
				isOK = false
				break
			}
		}
		return isOK
	}
	components := 1
	if q.mode == FastSync { // fast
		components = 2
	}
	log.Trace("download queue  Reserveipfs begin ", "lenHeader", len(recvheader), "origin", origin, "remote", remote)
	for _, header := range recvheader {
		hash := header.Hash()
		index := int(header.Number.Int64() - int64(q.resultOffset))
		if index >= len(q.resultCache) || index < 0 {
			//common.Report("index allocation went beyond available resultCache space")
			log.Warn("index allocation went beyond available resultCache space")
			break
			//return nil, fmt.Errorf("index allocation went beyond available resultCache space ")
		}
		if q.resultCache[index] == nil {

			//log.Trace("download queue  enter Reserveipfs new", "index", index, " header number", header.Number.Uint64(), "components", components)
			q.resultCache[index] = &fetchResult{
				Pending: components,
				//lbSend:  1,
				Hash:   hash,
				Header: header,
			}
		}

		if isNil(header) {
			//log.Trace("download queue  enter Reserveipfs new", ".Pengding--", q.resultCache[index].Pending)
			//q.resultCache[index].Pending--
			q.resultCache[index].Pending = 0
			progress = true
			//log.Trace("download queue  enter Reserveipfs header new", "index", index, ".Pengding", q.resultCache[index].Pending)
			continue
		}
		log.Trace("download  queue  enter Reserveipfs new", "index", index, ".Pengding", q.resultCache[index].Pending, " header number", header.Number.Uint64())
		//RequsetHeader = append(RequsetHeader, header)
		//q.BlockIpfsInsetPool(header.Number.Uint64(), header)
		//lb curHeigh = header.Number.Uint64()
	}

	/*if progress {
		// WaitResults
		log.Warn("download queue  enter Reserveipfs q.active.Signal,")
		//
		q.active.Signal()
	}
	*/

	//rlen := len(q.resultCache)
	//CurJudgeBlockNum := recvheader[0].Number.Uint64()
	//先找当前最大的连续resultCache的值的头信息
	var pMaxRes *fetchResult
	var maxi int
	for maxi, pMaxRes = range q.resultCache {
		if pMaxRes == nil {
			if maxi > 0 {
				pMaxRes = q.resultCache[maxi-1]
			}
			break
		}
	}
	q.recvheadNum++

	if QSingleBlockStore == true {
		if pMaxRes != nil {
			hasbatch := false
			curHeigh := pMaxRes.Header.Number.Uint64()           //当前最大的连续resultCache头信息节点高度      //q.resultCache[rlen-1].Header.Number.Uint64()
			curOrigin := q.resultCache[0].Header.Number.Uint64() //当前待处理resultCache的头信息节点开始值      //q.resultBegin
			log.Warn("download queue  enter Reserveipfs len(q.resultCache),", "max resultCache index", maxi, "curOrigin resultCache", curOrigin, "maxnumber resultCache", curHeigh, "origin", origin, "remote", remote)
			if (remote-origin)/300 >= 1 {
				if (curHeigh-curOrigin+1) < 300 && (remote-curHeigh+1 > 300) {
					log.Warn("download queue header too little, wait for next time")
					return nil, fmt.Errorf("download queue header too little, wait for next time")
				}
				//if(curOrigin != origin+1)
			}
			//blockCacheItems 为 resultCache 的缓存数目
			if maxi >= blockCacheItems-300 {
				maxi = blockCacheItems - 300 - 1
			}

			for i := 0; i < maxi; i++ {

				if q.resultCache[i].Flag == 1 { //单个
					continue
				}
				if q.resultCache[i].Flag == 2 { //批量的
					//i += 299
					continue
				}
				//有可以300批量同步的区块
				if (q.resultCache[i].Header.Number.Uint64()%300 == 1) && ((maxi - i) >= 300) {
					if q.resultCache[i].Flag != 2 {
						//q.resultCache[i].Flag = 2
						//q.resultCache[i+299].Flag = 2
						for n := 0; n < 300; n++ {
							if q.resultCache[i+n] == nil {
								break
							}
							q.resultCache[i+n].Flag = 2
						}
						//
						tmpReq := BlockIpfsReq{
							ReqPendflg:  components, //q.resultCache[i].Pending,
							Flag:        2,
							coinstr:     "0",
							HeadReqipfs: q.resultCache[i].Header,
						}
						RequsetHeader = append(RequsetHeader, tmpReq)
						//q.BlockIpfsInsetPool(header.Number.Uint64(), header)
						q.BlockIpfsInsetPool(q.resultCache[i].Header.Number.Uint64(), q.resultCache[i].Header, components, 2)
						q.recvheadNum = 0
						log.Warn("download queue Reserveipfs continuous begin ", "i", i, "blockNum", tmpReq.HeadReqipfs.Number.Uint64())
						i += 299
						q.resultWait = false
					}
					hasbatch = true
				} else if hasbatch == false { //单个处理
					if q.recvheadNum < 6 {
						if (q.resultCache[i].Header.Number.Uint64()%300 == 1) && (remote-q.resultCache[i].Header.Number.Uint64() >= 300) { //((int(remote) - i) >= 300) {
							log.Warn("download queue Reserveipfs continuous not - break ", "i", i, "blockNum", q.resultCache[i].Header.Number.Uint64())
							q.resultWait = false
							break
						}
					}
					if q.resultCache[i].Pending > 0 { //空hash 不请求body等
						q.resultCache[i].Flag = 1
						tmpReq := BlockIpfsReq{
							ReqPendflg:  components, //q.resultCache[i].Pending,
							Flag:        1,
							coinstr:     "0",
							HeadReqipfs: q.resultCache[i].Header,
						}
						RequsetHeader = append(RequsetHeader, tmpReq)
						q.BlockIpfsInsetPool(q.resultCache[i].Header.Number.Uint64(), q.resultCache[i].Header, components, 1)
						log.Warn(" download queue Reserveipfs continuous not but is discontinuous ", "i", i, "blockNum", tmpReq.HeadReqipfs.Number.Uint64())
					}
					q.recvheadNum = 0
					// 延后处理
					if progress && q.resultWait {
						// WaitResults
						log.Warn("download queue  enter Reserveipfs q.active.Signal,")
						//
						q.active.Signal()
					}
				}
			}
		}
	} else {
		//300个区块批量下载,剩余的接近 remote 的不足300的走原来的下载方式
		//origin, remote uint64 // CurJudgeBlockNum
		remoteDiv := remote / 300
		remoteMod := remote % 300
		//	originDiv := origin / 300
		//	originMod := origin % 300
		MaxBatchreqNum := remoteDiv * 300
		//maxiDiv := maxi / 300
		if pMaxRes != nil {
			CurFethMaxNum := pMaxRes.Header.Number.Uint64()
			CurFethMaxDiv := CurFethMaxNum / 300
			if CurFethMaxNum > remote {
				remoteMod = CurFethMaxNum % 300
			}
			log.Warn("download queue Reserveipfs begin", "CurFethMaxNum", CurFethMaxNum, "remoteDiv", remoteDiv, "remoteMod", remoteMod, "MaxBatchreqNum", MaxBatchreqNum)
			for i := 0; i < maxi; i++ {
				//遍历有序的resultcache
				if q.resultCache[i].Flag != 0 { //已请求
					continue
				}
				curDiv := q.resultCache[i].Header.Number.Uint64() / 300
				//可以批量获取 remoteDiv*300前的数据
				if MaxBatchreqNum > 0 {
					//可以批量
					calcDiv := curDiv
					for calcDiv = curDiv; calcDiv < CurFethMaxDiv; calcDiv++ {
						curMod := q.resultCache[i].Header.Number.Uint64() % 300
						if curMod == 1 {

							for n := 0; n < 300; n++ {
								if q.resultCache[i+n] == nil {
									break
								}
								q.resultCache[i+n].Flag = 2
							}
							tmpReq := BlockIpfsReq{
								ReqPendflg:   components, //q.resultCache[i].Pending,
								Flag:         2,
								coinstr:      "0",
								realBeginNum: q.resultCache[i].Header.Number.Uint64(),
								HeadReqipfs:  q.resultCache[i].Header,
							}
							RequsetHeader = append(RequsetHeader, tmpReq)
							q.BlockIpfsInsetPool(q.resultCache[i].Header.Number.Uint64(), q.resultCache[i].Header, components, 2)
							log.Warn("download queue Reserveipfs continuous begin ", "i", i, "blockNum", tmpReq.HeadReqipfs.Number.Uint64())
						} else {
							if curMod < 270 {
								tmpBlock := q.getBlock(curDiv*300 + 1) // 取本地链
								if tmpBlock == nil {
									log.Warn("download queue Reserveipfs calc blockNum", "blcokNum", curDiv*300+1) //还未入链等待
									break
								}
								for n := 0; n < (300 - int(curMod) + 1); n++ {
									if q.resultCache[i+n] == nil {
										break
									}
									q.resultCache[i+n].Flag = 2
								}
								//tmpBlock := q.getBlock(curDiv*300 + 1) // 取本地链
								tmpReq := BlockIpfsReq{
									ReqPendflg:   components, //q.resultCache[i].Pending,
									Flag:         2,
									coinstr:      "0",
									realBeginNum: q.resultCache[i].Header.Number.Uint64(),
									HeadReqipfs:  tmpBlock.Header(),
								}
								log.Warn("download queue Reserveipfs continuous getlocalBlock ", "LocalBlockNum ", tmpBlock.NumberU64(), "hash", tmpBlock.Hash(), "realbeginNum", tmpReq.realBeginNum)
								RequsetHeader = append(RequsetHeader, tmpReq)
								q.BlockIpfsInsetPool(q.resultCache[i].Header.Number.Uint64(), q.resultCache[i].Header, components, 2)
								log.Warn("download queue Reserveipfs continuous begin mod ", "i", i, "blockNum", tmpReq.HeadReqipfs.Number.Uint64())
							} else { //剩余小于 30个 走原来
								for n := 0; n < (300 - int(curMod) + 1); n++ {
									if q.resultCache[i+n] == nil {
										break
									}
									q.resultCache[i+n].Flag = 1
									hash := q.resultCache[i+n].Header.Hash()
									q.blockTaskPool[hash] = q.resultCache[i+n].Header
									q.blockTaskQueue.Push(q.resultCache[i+n].Header, -float32(q.resultCache[i+n].Header.Number.Uint64()))
									if q.resultCache[i+n].Pending > 1 {
										q.receiptTaskPool[hash] = q.resultCache[i+n].Header
										q.receiptTaskQueue.Push(q.resultCache[i+n].Header, -float32(q.resultCache[i+n].Header.Number.Uint64()))
									}
									log.Warn("download queue Reserveipfs add mod to  quondam download ", "i", i+n, "blockNum", q.resultCache[i+n].Header.Number, "hash", hash.String())
								}
							}
						}
						i = i + (300 - int(curMod) + 1)
					}

					if calcDiv == remoteDiv && remoteMod > 0 && q.resultCache[i] != nil {
						q.resultCache[i].Flag = 1
						hash := q.resultCache[i].Header.Hash()
						q.blockTaskPool[hash] = q.resultCache[i].Header
						q.blockTaskQueue.Push(q.resultCache[i].Header, -float32(q.resultCache[i].Header.Number.Uint64()))
						if q.resultCache[i].Pending > 1 {
							q.receiptTaskPool[hash] = q.resultCache[i].Header
							q.receiptTaskQueue.Push(q.resultCache[i].Header, -float32(q.resultCache[i].Header.Number.Uint64()))
						}
						log.Warn("download queue Reserveipfs add to  quondam download", "i", i, "blockNum", q.resultCache[i].Header.Number, "hash", hash.String())
					}

				} else if q.resultCache[i] != nil {
					q.resultCache[i].Flag = 1
					hash := q.resultCache[i].Header.Hash()
					q.blockTaskPool[hash] = q.resultCache[i].Header
					q.blockTaskQueue.Push(q.resultCache[i].Header, -float32(q.resultCache[i].Header.Number.Uint64()))
					if q.resultCache[i].Pending > 1 {
						q.receiptTaskPool[hash] = q.resultCache[i].Header
						q.receiptTaskQueue.Push(q.resultCache[i].Header, -float32(q.resultCache[i].Header.Number.Uint64()))
					}
					log.Warn("download queue Reserveipfs add to  quondam download tail", "i", i, "blockNum", q.resultCache[i].Header.Number, "hash", hash.String())
				}

			}
		}

	}
	return RequsetHeader, nil
}

// reserveHeaders reserves a set of data download operations for a given peer,
// skipping any previously failed ones. This method is a generic version used
// by the individual special reservation functions.
//
// Note, this method expects the queue lock to be already held for writing. The
// reason the lock is not obtained in here is because the parameters already need
// to access the queue, so they already need a lock anyway.
func (q *queue) reserveHeaders(p *peerConnection, count int, taskPool map[common.Hash]*types.Header, taskQueue *prque.Prque,
	pendPool map[string]*fetchRequest, donePool map[common.Hash]struct{}, isNoop func(*types.Header) bool) (*fetchRequest, bool, bool, error) {
	// Short circuit if the pool has been depleted, or if the peer's already
	// downloading something (sanity check not to corrupt state)
	if taskQueue.Empty() {
		log.Trace("download queue reserveHeaders taskQueue is nil")
		return nil, false, false, nil
	}
	if _, ok := pendPool[p.id]; ok {
		return nil, false, false, nil
	}
	// Calculate an upper limit on the items we might fetch (i.e. throttling)
	space := q.resultSlots(pendPool, donePool)

	// Retrieve a batch of tasks, skipping previously failed ones
	send := make([]*types.Header, 0, count)
	skip := make([]*types.Header, 0)

	progress := false
	for proc := 0; proc < space && len(send) < count && !taskQueue.Empty(); proc++ {
		header := taskQueue.PopItem().(*types.Header)
		hash := header.Hash()

		// If we're the first to request this task, initialise the result container
		index := int(header.Number.Int64() - int64(q.resultOffset))
		if index >= len(q.resultCache) || index < 0 {
			common.Report("index allocation went beyond available resultCache space")
			return nil, false, false, errInvalidChain
		}
		if q.resultCache[index] == nil {
			components := 1
			if q.mode == FastSync {
				components = 2
			}
			q.resultCache[index] = &fetchResult{
				Pending: components,
				Hash:    hash,
				Header:  header,
			}
		}
		// If this fetch task is a noop, skip this fetch operation
		if isNoop(header) {
			donePool[hash] = struct{}{}
			delete(taskPool, hash)

			space, proc = space-1, proc-1
			q.resultCache[index].Pending--
			progress = true
			continue
		}
		// Otherwise unless the peer is known not to have the data, add to the retrieve list
		if p.Lacks(hash) {
			skip = append(skip, header)
		} else {
			send = append(send, header)
		}
	}
	// Merge all the skipped headers back
	for _, header := range skip {
		taskQueue.Push(header, -float32(header.Number.Uint64()))
	}
	if progress {
		// Wake WaitResults, resultCache was modified
		q.active.Signal()
	}
	// Assemble and return the block download request
	if len(send) == 0 {
		return nil, false, progress, nil
	}

	log.Warn("download  queue reserveHeaders  reqest ", "p=", p.id, "sendcount", len(send))

	request := &fetchRequest{
		Peer:    p,
		Headers: send,
		Time:    time.Now(),
	}
	pendPool[p.id] = request

	return request, false, progress, nil
}

// CancelHeaders aborts a fetch request, returning all pending skeleton indexes to the queue.
func (q *queue) CancelHeaders(request *fetchRequest) {
	q.cancel(request, q.headerTaskQueue, q.headerPendPool)
}

// CancelBodies aborts a body fetch request, returning all pending headers to the
// task queue.
func (q *queue) CancelBodies(request *fetchRequest) {
	q.cancel(request, q.blockTaskQueue, q.blockPendPool)
}

// CancelReceipts aborts a body fetch request, returning all pending headers to
// the task queue.
func (q *queue) CancelReceipts(request *fetchRequest) {
	q.cancel(request, q.receiptTaskQueue, q.receiptPendPool)
}

// Cancel aborts a fetch request, returning all pending hashes to the task queue.
func (q *queue) cancel(request *fetchRequest, taskQueue *prque.Prque, pendPool map[string]*fetchRequest) {
	q.lock.Lock()
	defer q.lock.Unlock()

	if request.From > 0 {
		taskQueue.Push(request.From, -float32(request.From))
	}
	for _, header := range request.Headers {
		taskQueue.Push(header, -float32(header.Number.Uint64()))
	}
	delete(pendPool, request.Peer.id)
}

// Revoke cancels all pending requests belonging to a given peer. This method is
// meant to be called during a peer drop to quickly reassign owned data fetches
// to remaining nodes.
func (q *queue) Revoke(peerId string) {
	q.lock.Lock()
	defer q.lock.Unlock()

	if request, ok := q.blockPendPool[peerId]; ok {
		for _, header := range request.Headers {
			q.blockTaskQueue.Push(header, -float32(header.Number.Uint64()))
		}
		delete(q.blockPendPool, peerId)
	}
	if request, ok := q.receiptPendPool[peerId]; ok {
		for _, header := range request.Headers {
			q.receiptTaskQueue.Push(header, -float32(header.Number.Uint64()))
		}
		delete(q.receiptPendPool, peerId)
	}
}

// ExpireHeaders checks for in flight requests that exceeded a timeout allowance,
// canceling them and returning the responsible peers for penalisation.
func (q *queue) ExpireHeaders(timeout time.Duration) map[string]int {
	q.lock.Lock()
	defer q.lock.Unlock()

	return q.expire(timeout, q.headerPendPool, q.headerTaskQueue, headerTimeoutMeter)
}

// ExpireBodies checks for in flight block body requests that exceeded a timeout
// allowance, canceling them and returning the responsible peers for penalisation.
func (q *queue) ExpireBodies(timeout time.Duration) map[string]int {
	q.lock.Lock()
	defer q.lock.Unlock()

	return q.expire(timeout, q.blockPendPool, q.blockTaskQueue, bodyTimeoutMeter)
}

// ExpireReceipts checks for in flight receipt requests that exceeded a timeout
// allowance, canceling them and returning the responsible peers for penalisation.
func (q *queue) ExpireReceipts(timeout time.Duration) map[string]int {
	q.lock.Lock()
	defer q.lock.Unlock()

	return q.expire(timeout, q.receiptPendPool, q.receiptTaskQueue, receiptTimeoutMeter)
}

// expire is the generic check that move expired tasks from a pending pool back
// into a task pool, returning all entities caught with expired tasks.
//
// Note, this method expects the queue lock to be already held. The
// reason the lock is not obtained in here is because the parameters already need
// to access the queue, so they already need a lock anyway.
func (q *queue) expire(timeout time.Duration, pendPool map[string]*fetchRequest, taskQueue *prque.Prque, timeoutMeter metrics.Meter) map[string]int {
	// Iterate over the expired requests and return each to the queue
	expiries := make(map[string]int)
	for id, request := range pendPool {
		if time.Since(request.Time) > timeout {
			// Update the metrics with the timeout
			timeoutMeter.Mark(1)

			// Return any non satisfied requests to the pool
			if request.From > 0 {
				taskQueue.Push(request.From, -float32(request.From))
			}
			for _, header := range request.Headers {
				taskQueue.Push(header, -float32(header.Number.Uint64()))
			}
			// Add the peer to the expiry report along the the number of failed requests
			expiries[id] = len(request.Headers)
		}
	}
	// Remove the expired requests from the pending pool
	for id := range expiries {
		delete(pendPool, id)
	}
	return expiries
}

// DeliverHeaders injects a header retrieval response into the header results
// cache. This method either accepts all headers it received, or none of them
// if they do not map correctly to the skeleton.
//
// If the headers are accepted, the method makes an attempt to deliver the set
// of ready headers to the processor to keep the pipeline full. However it will
// not block to prevent stalling other pending deliveries.
func (q *queue) DeliverHeaders(id string, headers []*types.Header, headerProcCh chan []*types.Header) (int, error) {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.headerReqExpireNum=0
	// Short circuit if the data was never requested
	request := q.headerPendPool[id]
	if request == nil {
		return 0, errNoFetchesPending
	}
	headerReqTimer.UpdateSince(request.Time)
	delete(q.headerPendPool, id)
	if len(headers) > 0 {
		log.Debug("download DeliverHeaders Header", "peer", id, "begin number", headers[0].Number.Uint64(), "len", len(headers))
	} else {
		log.Debug("download DeliverHeaders Header =0", "peer", id, "len", len(headers))
	}

	// Ensure headers can be mapped onto the skeleton chain
	target := q.headerTaskPool[request.From].Hash()

	accepted := len(headers) == MaxHeaderFetch
	if accepted {
		if headers[0].Number.Uint64() != request.From {
			log.Trace("First header broke chain ordering", "peer", id, "number", headers[0].Number, "hash", headers[0].Hash(), request.From)
			accepted = false
		} else if headers[len(headers)-1].Hash() != target {
			log.Trace("Last header broke skeleton structure ", "peer", id, "number", headers[len(headers)-1].Number, "hash", headers[len(headers)-1].Hash(), "expected", target)
			accepted = false
		}
	}
	if accepted {
		for i, header := range headers[1:] {
			hash := header.Hash()
			if want := request.From + 1 + uint64(i); header.Number.Uint64() != want {
				log.Warn("Header broke chain ordering", "peer", id, "number", header.Number, "hash", hash, "expected", want)
				accepted = false
				break
			}
			if headers[i].Hash() != header.ParentHash {
				log.Warn("Header broke chain ancestry", "peer", id, "number", header.Number, "hash", hash)
				accepted = false
				break
			}
		}
	}
	// If the batch of headers wasn't accepted, mark as unavailable
	if !accepted {
		q.headerSkeletonErrNum ++ 
		log.Trace("Skeleton filling not accepted", "peer", id, "from", request.From,"q.headerSkeletonErrNum",q.headerSkeletonErrNum)	
		miss := q.headerPeerMiss[id]
		if miss == nil {
			q.headerPeerMiss[id] = make(map[uint64]struct{})
			miss = q.headerPeerMiss[id]
		}
		miss[request.From] = struct{}{}

		q.headerTaskQueue.Push(request.From, -float32(request.From))
		return 0, errors.New("delivery not accepted")
	}
	// Clean up a successful fetch and try to deliver any sub-results
	copy(q.headerResults[request.From-q.headerOffset:], headers)
	delete(q.headerTaskPool, request.From)
	q.headerSkeletonErrNum = 0 
	log.Info("download  DeliverHeaders ", "num =", accepted, "from", request.From, "headers len", len(headers), "offset ", q.headerOffset)

	ready := 0
	for q.headerProced+ready < len(q.headerResults) && q.headerResults[q.headerProced+ready] != nil {
		ready += MaxHeaderFetch
	}
	if ready > 0 {
		// Headers are ready for delivery, gather them and push forward (non blocking)
		process := make([]*types.Header, ready)
		copy(process, q.headerResults[q.headerProced:q.headerProced+ready])

		select {
		case headerProcCh <- process:
			log.Trace("Pre-scheduled new headers", "peer", id, "count", len(process), "from", process[0].Number)
			q.headerProced += len(process)
		default:
		}
	}
	// Check for termination and return
	if len(q.headerTaskPool) == 0 {
		q.headerContCh <- false
	}
	return len(headers), nil
}

// DeliverBodies injects a block body retrieval response into the results queue.
// The method returns the number of blocks bodies accepted from the delivery and
// also wakes any threads waiting for data delivery.
func (q *queue) DeliverBodies(id string, txLists [][]types.CurrencyBlock, uncleLists [][]*types.Header) (int, error) {
	q.lock.Lock()
	defer q.lock.Unlock()

	reconstruct := func(header *types.Header, index int, result *fetchResult) error {
		for _, cointx := range txLists[index] {
			for _, hr := range header.Roots {
				if hr.Cointyp == cointx.CurrencyName {
					if types.DeriveShaHash(types.TxHashList(cointx.Transactions.GetTransactions())) != hr.TxHash || types.CalcUncleHash(uncleLists[index]) != header.UncleHash {
						log.Error("download queue DeliverBodies reconstruct error", "number", header.Number.Uint64(), "hash", hr.TxHash)
						return errInvalidBody
					}
					break
				}
			}
		}
		result.Transactions = txLists[index]
		result.Uncles = uncleLists[index]
		return nil
	}
	log.Info("download queue DeliverBodies  ", "id=%s", id, "len", len(txLists))
	return q.deliver(id, q.blockTaskPool, q.blockTaskQueue, q.blockPendPool, q.blockDonePool, bodyReqTimer, len(txLists), reconstruct)
}

// DeliverReceipts injects a receipt retrieval response into the results queue.
// The method returns the number of transaction receipts accepted from the delivery
// and also wakes any threads waiting for data delivery.
func (q *queue) DeliverReceipts(id string, receiptList [][]types.CoinReceipts) (int, error) {
	q.lock.Lock()
	defer q.lock.Unlock()

	reconstruct := func(header *types.Header, index int, result *fetchResult) error {
		for _, cointx := range receiptList[index] {
			for _, hr := range header.Roots {
				if hr.Cointyp == cointx.CoinType {
					if types.DeriveShaHash(cointx.Receiptlist.HashList()) != hr.ReceiptHash {
						return errInvalidReceipt
					}
					break
				}
			}
		}
		result.Receipts = receiptList[index]
		return nil
	}
	log.Info("download queue DeliverReceipts  ", "id=%s", id, "len", len(receiptList))
	return q.deliver(id, q.receiptTaskPool, q.receiptTaskQueue, q.receiptPendPool, q.receiptDonePool, receiptReqTimer, len(receiptList), reconstruct)
}

// deliver injects a data retrieval response into the results queue.
//
// Note, this method expects the queue lock to be already held for writing. The
// reason the lock is not obtained in here is because the parameters already need
// to access the queue, so they already need a lock anyway.
func (q *queue) deliver(id string, taskPool map[common.Hash]*types.Header, taskQueue *prque.Prque,
	pendPool map[string]*fetchRequest, donePool map[common.Hash]struct{}, reqTimer metrics.Timer,
	results int, reconstruct func(header *types.Header, index int, result *fetchResult) error) (int, error) {

	// Short circuit if the data was never requested
	request := pendPool[id]
	if request == nil {
		return 0, errNoFetchesPending
	}
	reqTimer.UpdateSince(request.Time)
	delete(pendPool, id)

	// If no data items were retrieved, mark them as unavailable for the origin peer
	if results == 0 {
		for _, header := range request.Headers {
			request.Peer.MarkLacking(header.Hash())
		}
	}
	// Assemble each of the results with their headers and retrieved data parts
	var (
		accepted int
		failure  error
		useful   bool
	)
	for i, header := range request.Headers {
		// Short circuit assembly if no more fetch results are found
		if i == 0 {
			log.Trace("queue deliver body or receipt begin", "id=%s", id, "blockNum", header.Number.Uint64(), "len", results)
		}
		if i >= results {
			log.Trace("queue deliver body or receipt end ", "id=%s", id, "blockNum", header.Number.Uint64(), "len", results)
			break
		}
		// Reconstruct the next result if contents match up
		index := int(header.Number.Int64() - int64(q.resultOffset))
		if index >= len(q.resultCache) || index < 0 || q.resultCache[index] == nil {
			failure = errInvalidChain
			log.Trace("queue deliver body or receipt errInvalidChain index", "index", index)
			break
		}
		if err := reconstruct(header, i, q.resultCache[index]); err != nil {
			log.Trace("queue deliver body or receipt reconstruct", "index", index)
			failure = err
			break
		}
		hash := header.Hash()

		donePool[hash] = struct{}{}
		q.resultCache[index].Pending--
		useful = true
		accepted++

		// Clean up a successful fetch
		request.Headers[i] = nil
		delete(taskPool, hash)
	}
	// Return all failed or missing fetches to the queue
	for _, header := range request.Headers {
		if header != nil {
			taskQueue.Push(header, -float32(header.Number.Uint64()))
		}
	}
	// Wake up WaitResults
	if accepted > 0 {
		log.Trace("download queue q.active.Signal")
		q.active.Signal()
	}
	// If none of the data was good, it's a stale delivery
	switch {
	case failure == nil || failure == errInvalidChain:
		return accepted, failure
	case useful:
		return accepted, fmt.Errorf("partial failure: %v", failure)
	default:
		return accepted, errStaleDelivery
	}
}

// Prepare configures the result cache to allow accepting and caching inbound
// fetch results.
func (q *queue) Prepare(offset uint64, mode SyncMode) {
	q.lock.Lock()
	defer q.lock.Unlock()

	// Prepare the queue for sync results
	if q.resultOffset < offset {
		q.resultOffset = offset
	}
	q.mode = mode
}

//lb
func (q *queue) recvIpfsBody(bodyBlock *BlockIpfs) {

	q.lock.Lock()
	/*defer func() {
		q.lock.Unlock()
		q.BlockIpfsdeletePool(bodyBlock.Headeripfs.Number.Uint64())
	}()*/
	defer func() {
		q.lock.Unlock()
		//q.BlockIpfsdeletePool(trueBlockNumber)
	}()

	var index int
	//var trueBlockNumber uint64
	/*if bodyBlock.Flag == 0 {
		index = int(bodyBlock.Headeripfs.Number.Int64() - int64(q.resultOffset))
		trueBlockNumber = bodyBlock.Headeripfs.Number.Uint64()
	} else {
		index = int(int64(bodyBlock.BlockNum) - int64(q.resultOffset))
		trueBlockNumber = bodyBlock.BlockNum
	}*/

	index = int(int64(bodyBlock.BlockNum) - int64(q.resultOffset))

	log.Trace("####download queue recv a block insert to result# ", "number", bodyBlock.BlockNum, "index", index)
	// 需要增加对验证，头信息等， header

	if index >= len(q.resultCache) || index < 0 { //|| q.resultCache[index] == nil {
		if bodyBlock.Flag == 33 {
			log.Trace("download queue blockIpfsPool reqlist  delete resultCache is nil", "BlockNumber", bodyBlock.BlockNum)
			q.BlockIpfsdeletePool(bodyBlock.BlockNum)
			q.active.Signal()
		}
		return
	}
	if bodyBlock.Flag == 33 /*&& q.resultCache[index] != nil && q.resultCache[index].Pending == 0*/ {
		log.Trace("download queue blockIpfsPool reqlist  delete ", "trueBlockNumber", bodyBlock.BlockNum)
		q.BlockIpfsdeletePool(bodyBlock.BlockNum)
		q.active.Signal()
		return
	}

	if q.resultCache[index] == nil {
		log.Warn("download  syn recv a block insert reserveHeaders new discard ", "index", index, " header number", bodyBlock.BlockNum)
		return
		header := bodyBlock.Headeripfs
		hash := header.Hash()
		components := 1
		if q.mode == FastSync {
			components = 2
		}
		//log.Warn("download  syn recv a block insert reserveHeaders new", "index", index, " header number", header.Number.Uint64())
		//ResultCache
		q.resultCache[index] = &fetchResult{
			Pending: components,
			Hash:    hash,
			Header:  header,
		}
	}
	log.Trace("######download syn recv a block ", "index", index, ".Pending -- ", q.resultCache[index].Pending, "bodyBlock.Flag", bodyBlock.Flag, "res0", q.resultCache[0].Pending, "num", q.resultCache[0].Header.Number.Uint64())

	if index > 30 { //300 //ipfs 方式改为批量后,按理应该顺序，相差一定数目就就可以认为区块没有存储
		idx := 0 //index - 20
		if q.resultCache[idx] != nil {
			if q.resultCache[idx].Pending > 0 && q.resultCache[idx].Flag == 2 { //qian20个还没收到body/receipt
				hash := q.resultCache[idx].Header.Hash()
				_, ok := q.blockTaskPool[hash]
				if !ok {
					log.Warn("download  syn recv a block but begining is to receive, begin origin Req", "blockNum", q.resultCache[idx].Header.Number.Uint64())
					q.blockTaskPool[hash] = q.resultCache[idx].Header
					q.blockTaskQueue.Push(q.resultCache[idx].Header, -float32(q.resultCache[idx].Header.Number.Uint64()))
					if q.resultCache[idx].Pending > 1 {
						q.receiptTaskPool[hash] = q.resultCache[idx].Header
						q.receiptTaskQueue.Push(q.resultCache[idx].Header, -float32(q.resultCache[idx].Header.Number.Uint64()))
					}
				}
			}
		}
	}
	//head hash
	switch bodyBlock.Flag {
	case 0:
		/*
			q.resultCache[index].Transactions = bodyBlock.Transactionsipfs
			q.resultCache[index].Uncles = bodyBlock.Unclesipfs
			q.resultCache[index].Pending--*/
		if q.resultCache[index].Pending == 1 { //[]*types.Transaction
			for _, cointx := range bodyBlock.Transactionsipfs {
				for _, hr := range q.resultCache[index].Header.Roots {
					if hr.Cointyp == cointx.CurrencyName {
						if types.DeriveShaHash(types.TxHashList(cointx.Transactions.GetTransactions())) != hr.TxHash || types.CalcUncleHash(bodyBlock.Unclesipfs) != q.resultCache[index].Header.UncleHash {
							log.Warn("recvIpfsBody deal tx hash 0error")
							return
						}
					}
				}
			}
			q.resultCache[index].Transactions = bodyBlock.Transactionsipfs
			q.resultCache[index].Uncles = bodyBlock.Unclesipfs
			q.resultCache[index].Pending = 0
		} else if q.resultCache[index].Pending == 2 {
			for _, cointx := range bodyBlock.Transactionsipfs {
				for _, hr := range q.resultCache[index].Header.Roots {
					if hr.Cointyp == cointx.CurrencyName {
						if types.DeriveShaHash(types.TxHashList(cointx.Transactions.GetTransactions())) != hr.TxHash || types.CalcUncleHash(bodyBlock.Unclesipfs) != q.resultCache[index].Header.UncleHash {
							log.Warn("recvIpfsBody deal tx hash 02error")
							return
						}
					}
				}
			}
			for _, coinre := range bodyBlock.Receipt {
				for _, hr := range q.resultCache[index].Header.Roots {
					if hr.Cointyp == coinre.CoinType {
						if types.DeriveShaHash(coinre.Receiptlist.HashList()) != hr.ReceiptHash {
							log.Warn("recvIpfsBody deal receipt hash 02error")
							return
						}
					}
				}
			}

			q.resultCache[index].Transactions = bodyBlock.Transactionsipfs
			q.resultCache[index].Uncles = bodyBlock.Unclesipfs
			q.resultCache[index].Receipts = bodyBlock.Receipt
			q.resultCache[index].Pending = 0
		}
	case 1:

	case 2:
		if q.resultCache[index].Pending >= 1 {
			for _, cointx := range bodyBlock.Transactionsipfs {
				for _, hr := range q.resultCache[index].Header.Roots {
					if hr.Cointyp == cointx.CurrencyName {
						if types.DeriveShaHash(types.TxHashList(cointx.Transactions.GetTransactions())) != hr.TxHash || types.CalcUncleHash(bodyBlock.Unclesipfs) != q.resultCache[index].Header.UncleHash {
							log.Warn("recvIpfsBody deal tx hash 2error")
							return
						}
					}
				}
			}

			q.resultCache[index].Transactions = bodyBlock.Transactionsipfs
			q.resultCache[index].Uncles = bodyBlock.Unclesipfs
			q.resultCache[index].Pending--
		}
	case 3:
		if q.resultCache[index].Pending >= 1 {
			for _, coinre := range bodyBlock.Receipt {
				for _, hr := range q.resultCache[index].Header.Roots {
					if hr.Cointyp == coinre.CoinType {
						if types.DeriveShaHash(coinre.Receiptlist.HashList()) != hr.ReceiptHash {
							log.Warn("recvIpfsBody deal receipt hash 3error")
							return
						}
					}
				}
			}
			q.resultCache[index].Receipts = bodyBlock.Receipt
			q.resultCache[index].Pending--
		}

	}

	/*if q.resultCache[index].Pending == 0 {
		//log.Trace("recvIpfsBody resultCache delete ", "trueBlockNumber", trueBlockNumber)
		//q.BlockIpfsdeletePool(trueBlockNumber)
	}*/
	// Wake up
	NumBlockInfoFromIpfs++
	if NumBlockInfoFromIpfs > 30 {
		log.Trace("download queue q.active.Signal")
		q.active.Signal()
		NumBlockInfoFromIpfs = 0
	}

}

func (q *queue) checkIpfsPool() (bool, int, []uint64) {
	q.lock.Lock()
	defer q.lock.Unlock()

	bFind := false

	ipfsNum := q.BlockIpfsPoolBlocksNum()
	log.Trace("download queue checkIpfsPool len", "len", ipfsNum)
	if ipfsNum == 0 {
		return false, ipfsNum, nil
	}
	//log.Trace("download queue checkIpfsPool len", "len", q.blockIpfsPool)
	var blockNumlist []uint64
	var duration time.Duration = 20 * time.Minute
	for num, requset := range q.blockIpfsPool {
		//log.Trace("download queue checkIpfsPool len", "requset number", requset.Header.Number.Uint64())
		if time.Since(requset.Time) > duration {
			if requset.Header == nil {
				continue
			}
			log.Warn("download queue checkIpfsPool fail", "num", num, "blockNum", requset.Header.Number.Uint64())
			hash := requset.Header.Hash()
			//q.blockTaskPool[hash] = requset.Header
			//q.blockTaskQueue.Push(requset.Header, -float32(requset.Header.Number.Uint64()))
			if requset.Flag == 2 {
				if q.checkBatchBlockReq(requset.Header.Number.Uint64(), requset.Pending) == false {
					break
				}
			} else {
				q.blockTaskPool[hash] = requset.Header
				q.blockTaskQueue.Push(requset.Header, -float32(requset.Header.Number.Uint64()))
				if requset.Pending > 1 {
					q.receiptTaskPool[hash] = requset.Header
					q.receiptTaskQueue.Push(requset.Header, -float32(requset.Header.Number.Uint64()))
				}
			}
			bFind = true
			blockNumlist = append(blockNumlist, num)
		}
	} //blockIpfsPool

	/*if bFind {
		for _, num := range blockNumlist {
			q.BlockIpfsdeletePool(num)
		}
	}*/
	return bFind, ipfsNum, blockNumlist
}
func (q *queue) BlockIpfsdeleteBatch(blockNumlist []uint64) {
	q.lock.Lock()
	defer q.lock.Unlock()
	for _, num := range blockNumlist {
		q.BlockIpfsdeletePool(num)
	}
}
func (q *queue) BlockRegetByOldMode(bFlg int, pending int, header *types.Header, realReqNum uint64) {
	q.lock.Lock()
	defer q.lock.Unlock()
	hash := header.Hash()
	log.Warn("download queue BlockRegetByOld", "header number", header.Number.Uint64())
	if 2 == bFlg { //批量请求的
		//q.checkBatchBlockReq(header.Number.Uint64(), pending)
		q.checkBatchBlockReq(realReqNum, pending)
	} else {
		q.blockTaskPool[hash] = header
		q.blockTaskQueue.Push(header, -float32(header.Number.Uint64()))
		if pending > 1 {
			q.receiptTaskPool[hash] = header
			q.receiptTaskQueue.Push(header, -float32(header.Number.Uint64()))
		}
	}
	q.BlockIpfsdeletePool(realReqNum) //header.Number.Uint64())
}
func (q *queue) checkBatchBlockReq(blockNum uint64, pending int) bool {
	if q.resultCache[0] == nil || q.resultCache[0].Header == nil {
		log.Warn("download queue checkBatchBlockReq nil")
		return false
	}
	firstBlockNum := q.resultCache[0].Header.Number.Uint64()
	index := int(blockNum) - int(firstBlockNum)
	log.Warn("download queue checkBatchBlockReq reto old sysc", "index", index, "blockNum", blockNum)
	if index >= len(q.resultCache) || index < 0 || q.resultCache[index] == nil {
		log.Warn("download queue checkBatchBlockReq fail", "index", index)
		return false
	}
	if q.resultCache[index].Header.Number.Uint64() != blockNum {
		log.Warn("download queue checkBatchBlockReq fail", "index", index, "q.resultCache[index].Header.Number.Uint64()", q.resultCache[index].Header.Number.Uint64())
		return false
	}
	//for i := 0; i < 300; i++ {
	tailNum := (blockNum/300 + 1) * 300
	crNum := int(tailNum - blockNum + 1)
	for i := 0; i < crNum; i++ {
		if q.resultCache[index+i] == nil {
			log.Warn("download queue checkBatchBlockReq q.resultCache[index+i] fail", "index", index+i)
			return false //break
		}
		hash := q.resultCache[index+i].Header.Hash()
		q.blockTaskPool[hash] = q.resultCache[index+i].Header
		q.blockTaskQueue.Push(q.resultCache[index+i].Header, -float32(q.resultCache[index+i].Header.Number.Uint64()))
		if pending > 1 {
			q.receiptTaskPool[hash] = q.resultCache[index+i].Header
			q.receiptTaskQueue.Push(q.resultCache[index+i].Header, -float32(q.resultCache[index+i].Header.Number.Uint64()))
		}
	}
	return true
}
