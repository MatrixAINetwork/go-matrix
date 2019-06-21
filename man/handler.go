// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package man

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"sync"
	"sync/atomic"
	"time"

	"github.com/MatrixAINetwork/go-matrix/params/manparams"

	"github.com/MatrixAINetwork/go-matrix/ca"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/consensus"
	"github.com/MatrixAINetwork/go-matrix/consensus/misc"
	"github.com/MatrixAINetwork/go-matrix/core"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/event"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/man/downloader"
	"github.com/MatrixAINetwork/go-matrix/man/fetcher"
	"github.com/MatrixAINetwork/go-matrix/mandb"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/msgsend"
	"github.com/MatrixAINetwork/go-matrix/p2p"
	"github.com/MatrixAINetwork/go-matrix/p2p/discover"
	"github.com/MatrixAINetwork/go-matrix/params"
	"github.com/MatrixAINetwork/go-matrix/rlp"
)

const (
	softResponseLimit = 18 * 1024 * 1024 // Target maximum size of returned blocks, headers or node data.
	estHeaderRlpSize  = 132000           //500              // Approximate size of an RLP encoded block header

	// txChanSize is the size of channel listening to NewTxsEvent.
	// The number is referenced from the size of tx pool.
	txChanSize = 4096
)

var (
	emptyNodeId = discover.NodeID{}

	daoChallengeTimeout = 15 * time.Second // Time allowance for a node to reply to the DAO handshake challenge
)

// errIncompatibleConfig is returned if the requested protocols and configs are
// not compatible (low protocol version restrictions and high requirements).
var errIncompatibleConfig = errors.New("incompatible configuration")

func errResp(code errCode, format string, v ...interface{}) error {
	return fmt.Errorf("%v - %v", code, fmt.Sprintf(format, v...))
}

//var MyPm *ProtocolManager

//var MsgCenter *mc.Center

type ProtocolManager struct {
	networkId uint64

	fastSync  uint32 // Flag whether fast sync is enabled (gets disabled if we already have blocks)
	acceptTxs uint32 // Flag whether we're considered synchronised (enables transaction processing)

	txpool      txPool
	blockchain  *core.BlockChain
	chainconfig *params.ChainConfig
	maxPeers    int

	downloader *downloader.Downloader
	fetcher    *fetcher.Fetcher
	//	peers      *peerSet
	Peers        *peerSet
	SubProtocols []p2p.Protocol

	eventMux      *event.TypeMux
	txsCh         chan core.NewTxsEvent
	txsSub        event.Subscription
	minedBlockSub *event.TypeMuxSubscription

	// channels for fetcher, syncer, txsyncLoop
	newPeerCh   chan *peer
	txsyncCh    chan *txsync
	quitSync    chan struct{}
	noMorePeers chan struct{}

	CheckDownloadNum      int
	LastCheckTime       int64
	LastCheckBlkNum     uint64
	Msgcenter *mc.Center
	// wait group is used for graceful shutdowns during downloading
	// and processing
	wg sync.WaitGroup
}

// NewProtocolManager returns a new Matrix sub protocol manager. The Matrix sub protocol manages peers capable
// with the Matrix network.
func NewProtocolManager(config *params.ChainConfig, mode downloader.SyncMode, networkId uint64, mux *event.TypeMux, txpool txPool, engine consensus.Engine, blockchain *core.BlockChain, chaindb mandb.Database, MsgCenter *mc.Center) (*ProtocolManager, error) {
	// Create the protocol manager with the base fields
	manager := &ProtocolManager{
		networkId:   networkId,
		eventMux:    mux,
		txpool:      txpool,
		blockchain:  blockchain,
		chainconfig: config,
		//		peers:       newPeerSet(),
		Peers:       newPeerSet(),
		newPeerCh:   make(chan *peer),
		noMorePeers: make(chan struct{}),
		txsyncCh:    make(chan *txsync),
		quitSync:    make(chan struct{}),
		Msgcenter:   MsgCenter,
	}
	// Figure out whether to allow fast sync or not
	if mode == downloader.FastSync && blockchain.CurrentBlock().NumberU64() > 0 {
		log.Warn("Blockchain not empty, fast sync disabled")
		mode = downloader.FullSync
	}
	if mode == downloader.FastSync {
		manager.fastSync = uint32(1)
	}
	// Initiate a sub-protocol for every implemented version we can handle
	manager.SubProtocols = make([]p2p.Protocol, 0, len(ProtocolVersions))
	for i, version := range ProtocolVersions {
		// Skip protocol version if incompatible with the mode of operation
		if mode == downloader.FastSync && version < man63 {
			continue
		}
		// Compatible; initialise the sub-protocol
		version := version // Closure for the run
		manager.SubProtocols = append(manager.SubProtocols, p2p.Protocol{
			Name:    ProtocolName,
			Version: version,
			Length:  ProtocolLengths[i],
			Run: func(p *p2p.Peer, rw p2p.MsgReadWriter) error {
				peer := manager.newPeer(int(version), p, rw)
				select {
				case manager.newPeerCh <- peer:
					manager.wg.Add(1)
					defer manager.wg.Done()
					return manager.handle(peer)
				case <-manager.quitSync:
					return p2p.DiscQuitting
				}
			},
			NodeInfo: func() interface{} {
				return manager.NodeInfo()
			},
			PeerInfo: func(id discover.NodeID) interface{} {
				//if p := manager.peers.Peer(fmt.Sprintf("%x", id[:8])); p != nil {
				if p := manager.Peers.Peer(fmt.Sprintf("%x", id[:8])); p != nil {
					return p.Info()
				}
				return nil
			},
		})
	}
	if len(manager.SubProtocols) == 0 {
		return nil, errIncompatibleConfig
	}
	// Construct the different synchronisation mechanisms
	manager.downloader = downloader.New(mode, chaindb, manager.eventMux, blockchain, nil, manager.removePeer, blockchain.GetBlockByNumber)

	validator := func(header *types.Header) error {
		//todo 无法连续验证，下载的区块全部不验证pow
		return engine.VerifyHeader(blockchain, header, false)
	}
	heighter := func() uint64 {
		return blockchain.CurrentBlock().NumberU64()
	}
	inserter := func(blocks types.Blocks) (int, error) {
		// If fast sync is running, deny importing weird blocks
		if atomic.LoadUint32(&manager.fastSync) == 1 {
			log.Warn("Discarded bad propagated block", "number", blocks[0].Number(), "hash", blocks[0].Hash())
			return 0, nil
		}
		atomic.StoreUint32(&manager.acceptTxs, 1) // Mark initial sync done on any fetcher import
		return manager.blockchain.InsertChain(blocks,1)
	}
	manager.fetcher = fetcher.New(blockchain.GetBlockByHash, validator, manager.BroadcastBlock, heighter, inserter, manager.removePeer)

	return manager, nil
}

func (pm *ProtocolManager) removePeer(id string,flg int) {
	// Short circuit if the peer was already removed
	//	peer := pm.peers.Peer(id)
	peer := pm.Peers.Peer(id)
	if peer == nil {
		return
	}
	log.Debug("Removing Matrix peer", "peer", id)

	// Unregister the peer from the downloader and Matrix peer set
	pm.downloader.UnregisterPeer(id,flg)
	//	if err := pm.peers.Unregister(id); err != nil {
	if err := pm.Peers.Unregister(id); err != nil {
		log.Error("Peer removal failed", "peer", id, "err", err)
	}
	// Hard disconnect at the networking layer
	if peer != nil {
		peer.Peer.Disconnect(p2p.DiscUselessPeer)
	}
}

/*func (pm *ProtocolManager) MySend() {

	fmt.Println("*************************Mysend Start")

	P2PSendDisPatcherMsgCH := make(chan *hd.P2PCommuHD, 10)
	_, err := mc.SubscribeEvent(mc.P2PSENDDISPATCHERMSG, P2PSendDisPatcherMsgCH) //pm.Msgcenter.SubscribeEvent(mc.P2PSENDDISPATCHERMSG, mcmsg.P2PSendDisPatcherMsgCH)
	if err != nil {
		fmt.Println("***************MySend err", err)
	}

	for {
		if pm.Peers.Len() != 0 {
			break
		}
		time.Sleep(time.Second * 10)
	}

	for {
		select {
		case mcmss := <-P2PSendDisPatcherMsgCH:
			log.INFO("SendToGroup", "data", mcmss.Role)
			p2p.SendToGroup(mcmss.Role, uint64(mcmss.SubCode), mcmss.NodeInfo)
		}
	}
}*/

func (pm *ProtocolManager) Start(maxPeers int) {
	pm.maxPeers = maxPeers

	// broadcast transactions
	pm.txsCh = make(chan core.NewTxsEvent, txChanSize)
	pm.txsSub = pm.txpool.SubscribeNewTxsEvent(pm.txsCh)
	go pm.txBroadcastLoop()

	// broadcast mined blocks
	pm.minedBlockSub = pm.eventMux.Subscribe(core.NewMinedBlockEvent{})
	go pm.minedBroadcastLoop()

	// start sync handlers
	//go pm.MySend()
	go pm.syncer()
	go pm.txsyncLoop()
	//	MyPm = pm

	//saveSnapshotPeriod ,allowSnapshotPoint 现在先定死 300 and 0  广播节点才能调用ipfs 上传接口
	//go pm.saveSnapshot(300, 0)

}

func (pm *ProtocolManager) Stop() {
	log.Info("Stopping Matrix protocol")

	pm.txsSub.Unsubscribe()        // quits txBroadcastLoop
	pm.minedBlockSub.Unsubscribe() // quits blockBroadcastLoop

	// Quit the sync loop.
	// After this send has completed, no new peers will be accepted.
	pm.noMorePeers <- struct{}{}

	// Quit fetcher, txsyncLoop.
	close(pm.quitSync)

	// Disconnect existing sessions.
	// This also closes the gate for any new registrations on the peer set.
	// sessions which are already established but not added to pm.peers yet
	// will exit when they try to register.
	//	pm.peers.Close()
	pm.Peers.Close()

	// Wait for all peer handler goroutines and the loops to come down.
	pm.wg.Wait()

	log.Info("Matrix protocol stopped")
}

func (pm *ProtocolManager) newPeer(pv int, p *p2p.Peer, rw p2p.MsgReadWriter) *peer {
	return newPeer(pv, p, newMeteredMsgWriter(rw))
}

// handle is the callback invoked to manage the life cycle of an man peer. When
// this function terminates, the peer is disconnected.
func (pm *ProtocolManager) handle(p *peer) error {

	//	pm.Msgcenter = p.Msgcenter
	// Ignore maxPeers if this is a trusted peer
	//	if pm.peers.Len() >= pm.maxPeers && !p.Peer.Info().Network.Trusted {
	if pm.Peers.Len() >= pm.maxPeers && !p.Peer.Info().Network.Trusted {
		return p2p.DiscTooManyPeers
	}
	p.Log().Debug("Matrix peer connected", "name", p.Name())

	sbi, err := pm.blockchain.GetSuperBlockInfo()
	if nil != err {
		return errors.New("get super seq error")
	}
	// Execute the Matrix handshake
	var (
		genesis = pm.blockchain.Genesis()
		head    = pm.blockchain.CurrentHeader()
		hash    = head.Hash()
		number  = head.Number.Uint64()
		td      = pm.blockchain.GetTd(hash, number)
		bt      = pm.blockchain.CurrentHeader().Time.Uint64()
		sbs     = sbi.Seq
		sbHash  = sbi.Num
	)

	if manparams.CanSwitchGammaCanonicalChain(time.Now().Unix()) {
		if err := p.NewHandshake(pm.networkId, bt, hash, sbs, genesis.Hash(), sbHash, number); err != nil {
			p.Log().Debug("Matrix handshake failed", "err", err)
			return err
		}
	} else {
		if err := p.Handshake(pm.networkId, td, hash, sbs, genesis.Hash(), sbHash); err != nil {
			p.Log().Debug("Matrix handshake failed", "err", err)
			return err
		}
	}

	//	if rw, ok := p.rw.(*meteredMsgReadWriter); ok {
	if rw, ok := p.rw.(*meteredMsgReadWriter); ok {
		rw.Init(p.version)
	}
	p.Log().Debug("Matrix handshake with peer sucess ", "peerid=%d", pm.networkId)
	// Register the peer locally
	//	if err := pm.peers.Register(p); err != nil {
	if err := pm.Peers.Register(p); err != nil {
		p.Log().Error("Matrix peer registration failed", "err", err)
		return err
	}
	defer pm.removePeer(p.id,0)

	// Register the peer in the downloader. If the downloader considers it banned, we disconnect
	if err := pm.downloader.RegisterPeer(p.id, p.version, p); err != nil {
		return err
	}
	// Propagate existing transactions. new transactions appearing
	// after this will be sent via broadcasts.
	//pm.syncTransactions(p) // 2018-08-29 新节点连接时不去要其他的交易

	// If we're DAO hard-fork aware, validate any remote peer with regard to the hard-fork
	if daoBlock := pm.chainconfig.DAOForkBlock; daoBlock != nil {
		// Request the peer's DAO fork header for extra-data validation
		if err := p.RequestHeadersByNumber(daoBlock.Uint64(), 1, 0, false); err != nil {
			return err
		}
		// Start a timer to disconnect if the peer doesn't reply in time
		p.forkDrop = time.AfterFunc(daoChallengeTimeout, func() {
			p.Log().Debug("Timed out DAO fork-check, dropping")
			pm.removePeer(p.id,0)
		})
		// Make sure it's cleaned up if the peer dies off
		defer func() {
			if p.forkDrop != nil {
				p.forkDrop.Stop()
				p.forkDrop = nil
			}
		}()
	}
	// main loop. handle incoming messages.
	for {
		if err := pm.handleMsg(p); err != nil {
			p.Log().Debug("Matrix message handling failed", "err", err)
			return err
		}
	}
}

// handleMsg is invoked whenever an inbound message is received from a remote
// peer. The remote connection is torn down upon returning any error.
func (pm *ProtocolManager) handleMsg(p *peer) error {
	// Read the next message from the remote peer, and ensure it's fully consumed
	msg, err := p.rw.ReadMsg()
	if err != nil {
		return err
	}
	if msg.Size > ProtocolMaxMsgSize {
		return errResp(ErrMsgTooLarge, "%v > %v", msg.Size, ProtocolMaxMsgSize)
	}
	defer msg.Discard()

	// Handle the message depending on its contents
	switch {
	case msg.Code == StatusMsg:
		// Status messages should never arrive after the handshake
		return errResp(ErrExtraStatusMsg, "uncontrolled status message")

	// Block header query, collect the requested headers and reply
	case msg.Code == GetBlockHeadersMsg:
		// Decode the complex header query
		var query getBlockHeadersData
		if err := msg.Decode(&query); err != nil {
			return errResp(ErrDecode, "%v: %v", msg, err)
		}
		hashMode := query.Origin.Hash != (common.Hash{})

		// Gather headers until the fetch or network limits is reached
		var (
			bytes   common.StorageSize
			headers []*types.Header
			unknown bool
		)
		for !unknown && len(headers) < int(query.Amount) && bytes < softResponseLimit && len(headers) < downloader.MaxHeaderFetch {
			// Retrieve the next header satisfying the query
			var origin *types.Header
			if hashMode {
				origin = pm.blockchain.GetHeaderByHash(query.Origin.Hash)
			} else {
				origin = pm.blockchain.GetHeaderByNumber(query.Origin.Number)
			}
			if origin == nil {
				break
			}
			number := origin.Number.Uint64()
			headers = append(headers, origin)
			bytes += estHeaderRlpSize

			// Advance to the next header of the query
			switch {
			case query.Origin.Hash != (common.Hash{}) && query.Reverse:
				// Hash based traversal towards the genesis block
				for i := 0; i < int(query.Skip)+1; i++ {
					if header := pm.blockchain.GetHeader(query.Origin.Hash, number); header != nil {
						query.Origin.Hash = header.ParentHash
						number--
					} else {
						unknown = true
						break
					}
				}
			case query.Origin.Hash != (common.Hash{}) && !query.Reverse:
				// Hash based traversal towards the leaf block
				var (
					current = origin.Number.Uint64()
					next    = current + query.Skip + 1
				)
				if next <= current {
					infos, _ := json.MarshalIndent(p.Peer.Info(), "", "  ")
					p.Log().Warn("GetBlockHeaders skip overflow attack", "current", current, "skip", query.Skip, "next", next, "attacker", infos)
					unknown = true
				} else {
					if header := pm.blockchain.GetHeaderByNumber(next); header != nil {
						if pm.blockchain.GetBlockHashesFromHash(header.Hash(), query.Skip+1)[query.Skip] == query.Origin.Hash {
							query.Origin.Hash = header.Hash()
						} else {
							unknown = true
						}
					} else {
						unknown = true
					}
				}
			case query.Reverse:
				// Number based traversal towards the genesis block
				if query.Origin.Number >= query.Skip+1 {
					query.Origin.Number -= query.Skip + 1
				} else {
					unknown = true
				}

			case !query.Reverse:
				// Number based traversal towards the leaf block
				query.Origin.Number += query.Skip + 1
			}
		}
		if len(headers) > 0 {
			p.Log().Trace("download handleMsg recv GetBlockHeadersMsg", "headers len", len(headers), "number", headers[0].Number.Uint64())
		} else {
			p.Log().Trace("download handleMsg recv GetBlockHeadersMsg", "headers len", len(headers))
		}
		return p.SendBlockHeaders(headers)

	case msg.Code == BlockHeadersMsg:
		// A batch of headers arrived to one of our previous requests
		var headers []*types.Header
		if err := msg.Decode(&headers); err != nil {
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}

		p.Log().Trace("download handleMsg BlockHeadersMsg", "len", len(headers))

		// Filter out any explicitly requested headers, deliver the rest to the downloader
		filter := len(headers) == 1
		if filter {
			// If it's a potential DAO fork check, validate against the rules
			if p.forkDrop != nil && pm.chainconfig.DAOForkBlock.Cmp(headers[0].Number) == 0 {
				// Disable the fork drop timer
				p.forkDrop.Stop()
				p.forkDrop = nil

				// Validate the header and either drop the peer or continue
				if err := misc.VerifyDAOHeaderExtraData(pm.chainconfig, headers[0]); err != nil {
					p.Log().Debug("Verified to be on the other side of the DAO fork, dropping")
					return err
				}
				p.Log().Debug("Verified to be on the same side of the DAO fork")
				return nil
			}
			// Irrelevant of the fork checks, send the header to the fetcher just in case
			p.Log().Trace("download handleMsg BlockHeadersMsg fetch","headers",headers[0].Hash(),"number",headers[0].Number.Uint64())
			headers = pm.fetcher.FilterHeaders(p.id, headers, time.Now())
		}
		p.Log().Trace("download handleMsg BlockHeadersMsg after", "len", len(headers), "!filter", !filter)
		if len(headers) > 0 || !filter {
			err := pm.downloader.DeliverHeaders(p.id, headers)
			if err != nil {
				log.Debug("Failed to deliver headers", "err", err)
			}
		}

	case msg.Code == GetBlockBodiesMsg:
		// Decode the retrieval message
		msgStream := rlp.NewStream(msg.Payload, uint64(msg.Size))
		if _, err := msgStream.List(); err != nil {
			return err
		}
		// Gather blocks until the fetch or network limits is reached
		var (
			hash   common.Hash
			bytes  int
			bodies []rlp.RawValue
		)
		for bytes < softResponseLimit && len(bodies) < downloader.MaxBlockFetch {
			// Retrieve the hash of the next block
			if err := msgStream.Decode(&hash); err == rlp.EOL {
				break
			} else if err != nil {
				return errResp(ErrDecode, "msg %v: %v", msg, err)
			}
			// Retrieve the requested block body, stopping if enough was found
			if data := pm.blockchain.GetBodyRLP(hash); len(data) != 0 {
				bodies = append(bodies, data)
				bytes += len(data)
			}
		}
		p.Log().Trace("download handleMsg recv GetBlockBodiesMsg", "bodies len", len(bodies), "hash", hash)
		return p.SendBlockBodiesRLP(bodies)

	case msg.Code == BlockBodiesMsg:
		// A batch of block bodies arrived to one of our previous requests
		var request blockBodiesData
		if err := msg.Decode(&request); err != nil {
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		// Deliver them all to the downloader for queuing
		//transactions := make([][]types.SelfTransaction, len(request))
		//transactions := make([][]types.CoinSelfTransaction, len(request))
		transCrBlock := make([][]types.CurrencyBlock, len(request))
		uncles := make([][]*types.Header, len(request))

		for i, body := range request {
			/*
				cointx := make([]types.CoinSelfTransaction,0)
				for _,curr := range body.Transactions{
					cointx = append(cointx,types.CoinSelfTransaction{CoinType:curr.CurrencyName,Txser:curr.Transactions.GetTransactions()})
				}
				transactions[i] = cointx//.GetTransactions()*/
			transCrBlock[i] = body.Transactions
			uncles[i] = body.Uncles
		}
		// Filter out any explicitly requested bodies, deliver the rest to the downloader
		filter := len(transCrBlock) > 0 || len(uncles) > 0
		if filter {
			transCrBlock, uncles = pm.fetcher.FilterBodies(p.id, transCrBlock, uncles, time.Now())
		}

		p.Log().Trace("download handleMsg BlockBodiesMsg after filter", "len transaction", len(transCrBlock), "!filter", !filter)
		if len(transCrBlock) > 0 || len(uncles) > 0 || !filter {
			err := pm.downloader.DeliverBodies(p.id, transCrBlock, uncles)
			if err != nil {
				log.Debug("Failed to deliver bodies", "err", err)
			}
		}

	case p.version >= man63 && msg.Code == GetNodeDataMsg:
		// Decode the retrieval message
		msgStream := rlp.NewStream(msg.Payload, uint64(msg.Size))
		if _, err := msgStream.List(); err != nil {
			return err
		}
		// Gather state data until the fetch or network limits is reached
		var (
			hash  common.Hash
			bytes int
			data  [][]byte
		)
		for bytes < softResponseLimit && len(data) < downloader.MaxStateFetch {
			// Retrieve the hash of the next state entry
			if err := msgStream.Decode(&hash); err == rlp.EOL {
				break
			} else if err != nil {
				return errResp(ErrDecode, "msg %v: %v", msg, err)
			}
			// Retrieve the requested state entry, stopping if enough was found
			if entry, err := pm.blockchain.TrieNode(hash); err == nil {
				data = append(data, entry)
				bytes += len(entry)
			}
		}
		return p.SendNodeData(data)

	case p.version >= man63 && msg.Code == NodeDataMsg:
		// A batch of node state data arrived to one of our previous requests
		var data [][]byte
		if err := msg.Decode(&data); err != nil {
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		// Deliver all to the downloader
		if err := pm.downloader.DeliverNodeData(p.id, data); err != nil {
			p.Log().Debug("Failed to deliver node state data", "err", err)
		}

	case p.version >= man63 && msg.Code == GetReceiptsMsg:
		// Decode the retrieval message
		msgStream := rlp.NewStream(msg.Payload, uint64(msg.Size))
		if _, err := msgStream.List(); err != nil {
			return err
		}
		// Gather state data until the fetch or network limits is reached
		var (
			hash     common.Hash
			bytes    int
			receipts []rlp.RawValue
		)
		for bytes < softResponseLimit && len(receipts) < downloader.MaxReceiptFetch {
			// Retrieve the hash of the next block
			if err := msgStream.Decode(&hash); err == rlp.EOL {
				break
			} else if err != nil {
				return errResp(ErrDecode, "msg %v: %v", msg, err)
			}
			// Retrieve the requested block's receipts, skipping if unknown to us
			results := pm.blockchain.GetReceiptsByHash(hash)
			if results == nil {
				p.Log().Info("Get receipt err", "Get receipt err", "Get receipt err")
				continue
			}
			// If known, encode and queue for response packet
			if encoded, err := rlp.EncodeToBytes(results); err != nil {
				p.Log().Error("Failed to encode receipt", "err", err)
			} else {
				receipts = append(receipts, encoded)
				bytes += len(encoded)
			}
		}
		return p.SendReceiptsRLP(receipts)

	case p.version >= man63 && msg.Code == ReceiptsMsg:
		// A batch of receipts arrived to one of our previous requests
		var receipts [][]types.CoinReceipts
		if err := msg.Decode(&receipts); err != nil {
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		// Deliver all to the downloader
		if err := pm.downloader.DeliverReceipts(p.id, receipts); err != nil {
			p.Log().Debug("Failed to deliver receipts", "err", err)
		}

	case msg.Code == NewBlockHashesMsg:
		var announces newBlockHashesData
		if err := msg.Decode(&announces); err != nil {
			return errResp(ErrDecode, "%v: %v", msg, err)
		}
		// Mark the hashes as present at the remote node
		for _, block := range announces {
			p.MarkBlock(block.Hash)
		}
		if len(announces) > 0 {
			p.Log().Trace("download fetch handleMsg receive NewBlockHashesMsg0", "BlockNum", announces[0].Number, "hash", announces[0].Hash.String())
			if (announces[0].Number > pm.fetcher.MaxChainHeight){
				pm.fetcher.MaxChainHeight = announces[0].Number
			}
			if pm.blockchain.CurrentBlock().NumberU64() + 20 < announces[0].Number {
				p.SetHeadPart(announces[0].Hash,announces[0].Number)
				go pm.synchronise(p,10)//强制检查
			}
		}
		// Schedule all the unknown hashes for retrieval
		unknown := make(newBlockHashesData, 0, len(announces))
		for _, block := range announces {
			if !pm.blockchain.HasBlock(block.Hash, block.Number) {
				unknown = append(unknown, block)
			}
		}

		for _, block := range unknown {
			if ( block.Number > pm.blockchain.CurrentBlock().NumberU64()) {
				pm.fetcher.Notify(p.id, block.Hash, block.Number, time.Now(), p.RequestOneHeader, p.RequestBodies)
			}
		}

	case msg.Code == NewBlockMsg:
		// Retrieve and decode the propagated block
		var request newBlockData
		if err := msg.Decode(&request); err != nil {
			return errResp(ErrDecode, "%v: %v", msg, err)
		}
		request.Block.ReceivedAt = msg.ReceivedAt
		request.Block.ReceivedFrom = p

		// Mark the peer as owning the block and schedule it for import
		p.MarkBlock(request.Block.Hash())
		if (request.Block.NumberU64()> pm.fetcher.MaxChainHeight){
			pm.fetcher.MaxChainHeight = request.Block.NumberU64()
		}
		currentBlock := pm.blockchain.CurrentBlock()
		flg := 0
		if(pm.fetcher.MaxChainHeight - currentBlock.NumberU64() < 32) && ( request.Block.NumberU64() > currentBlock.NumberU64() ){
			pm.fetcher.Enqueue(p.id, request.Block)
			flg = 1
		}
		p.Log().Trace("download fetch handleMsg receive NewBlockMsg", "number", request.Block.NumberU64(), "request.TD", request.TD,"MaxChainHeight",pm.fetcher.MaxChainHeight,"flg",flg)
	

		// Assuming the block is importable by the peer, but possibly not yet done so,
		// calculate the head hash and TD that the peer truly must have.

		// Update the peers total difficulty if better than the previous
		var (
			trueHead = request.Block.ParentHash()
			trueTD   = new(big.Int).Sub(request.TD, request.Block.Difficulty())
			trueSBS  = request.SBS
		)
		_, td, sbs, _, bt, bn := p.Head()
		if manparams.CanSwitchGammaCanonicalChain(time.Now().Unix()) {

			p.Log().Trace("handleMsg receive NewBlockMsg", "超级区块序号", trueSBS, "缓存序号", sbs, "高度", request.Block.NumberU64(), "时间", request.Block.Time())
			if common.IsGreaterLink(common.LinkInfo{Sbs: trueSBS, Bn: request.Block.NumberU64(), Bt: request.Block.Time().Uint64()}, common.LinkInfo{Sbs: sbs, Bn: bn, Bt: bt}) {
				p.SetHead(trueHead, trueTD, trueSBS, request.SBH, request.Block.Time().Uint64(),request.Block.NumberU64())// request.Block.NumberU64()-1)
				p.Log().Trace("handleMsg receive NewBlockMsg SetHead") 
				// Schedule a sync if above ours. Note, this will not fire a sync for a gap of
				// a singe block (as the true TD is below the propagated block), however this
				// scenario should easily be covered by the fetcher.
				//currentBlock := pm.blockchain.CurrentBlock()
				/*
				td := pm.blockchain.GetTd(currentBlock.Hash(), currentBlock.NumberU64())
				if td == nil {
					log.Error("td is nil", "peer", p.id)
					break
				}
				sbs, err := pm.blockchain.GetSuperBlockSeq()
				if nil != err {
					log.Error("get super seq error")
					break
				}

				if common.IsGreaterLink(common.LinkInfo{Sbs: trueSBS, Bn: request.Block.NumberU64(), Bt: request.Block.Time().Uint64()}, common.LinkInfo{Sbs: sbs, Bn: currentBlock.NumberU64(), Bt: currentBlock.Time().Uint64()}) {

					log.Trace("handleMsg receive NewBlockMsg", "超级区块序号", trueSBS, "本地序号", sbs,
						"远程高度", request.Block.NumberU64(), "本地高度", currentBlock.NumberU64(), "远程时间", request.Block.Time(), "本地时间", currentBlock.Time())
					if currentBlock.NumberU64() + 1 < request.Block.NumberU64(){
						go pm.synchronise(p,3)
					}
				}
				*/
				if currentBlock.NumberU64() + 1 < request.Block.NumberU64(){
					go pm.synchronise(p,3)
				}
			}
		} else {

			p.Log().Trace("handleMsg receive NewBlockMsg", "超级区块序号", trueSBS, "缓存序号", sbs, "trueTD", trueTD)
			if trueSBS < sbs {
				//todo:日志
				break
			}

			if trueSBS > sbs || trueTD.Cmp(td) > 0 {
				p.SetHead(trueHead, trueTD, trueSBS, request.SBH, request.Block.Time().Uint64(), request.Block.NumberU64())

				// Schedule a sync if above ours. Note, this will not fire a sync for a gap of
				// a singe block (as the true TD is below the propagated block), however this
				// scenario should easily be covered by the fetcher.
				/*currentBlock := pm.blockchain.CurrentBlock()
				td := pm.blockchain.GetTd(currentBlock.Hash(), currentBlock.NumberU64())
				if td == nil {
					log.Error("td is nil", "peer", p.id)
					break
				}
				sbs, err := pm.blockchain.GetSuperBlockSeq()
				if nil != err {
					log.Error("get super seq error")
					break
				}

				if trueSBS > sbs || trueTD.Cmp(td) > 0 {
					log.Trace("handleMsg receive NewBlockMsg", "超级区块序号", trueSBS, "本地序号", sbs, "远程td", trueTD, "本地td", td)
					if currentBlock.NumberU64() + 1 < request.Block.NumberU64(){
						go pm.synchronise(p,4)
					}
				}*/
				if currentBlock.NumberU64() + 1 < request.Block.NumberU64(){
					go pm.synchronise(p,4)
				}
			}

		}

	case msg.Code == TxMsg:
		// Transactions arrived, make sure we have a valid and fresh chain to handle them
		selfRole := ca.GetRole()
		if selfRole == common.RoleBroadcast {
			break
		}
		// Transactions can be processed, parse all of them and deliver to the pool

		var txs []types.SelfTransaction
		if err := msg.Decode(&txs); err != nil {
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		for i, tx := range txs {
			// Validate and mark the remote transaction
			if tx == nil {
				return errResp(ErrDecode, "transaction %d is nil", i)
			}
			if nc := tx.Nonce(); nc < params.NonceAddOne {
				nc = nc | params.NonceAddOne
				tx.SetNonce(nc)
			}
			hash := tx.Hash()
			p.MarkTransaction(hash)
			log.INFO("==tcp tx hash", "from", tx.From().String(), "tx.Nonce", tx.Nonce(), "hash", hash.String(), "sender addr", p2p.ServerP2p.ConvertIdToAddress(p.ID()).String(),
				"node id", p.ID().String())
		}
		pm.txpool.AddRemotes(txs)
	case msg.Code == common.NetworkMsg:
		var m []*core.MsgStruct
		if err := msg.Decode(&m); err != nil {
			log.Info("handler", "mag NetworkMsg err", err)
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		log.Info("handler", "msg NetworkMsg ", "ProcessMsg")

		addr := p2p.ServerP2p.ConvertIdToAddress(p.ID())
		go pm.txpool.ProcessMsg(core.NetworkMsgData{SendAddress: addr, Data: m})

	case msg.Code == common.AlgorithmMsg:
		var m msgsend.NetData
		if err := msg.Decode(&m); err != nil {
			log.Error("algorithm message", "error", err)
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		addr := p2p.ServerP2p.ConvertIdToAddress(p.ID())
		if addr == p2p.EmptyAddress {
			log.Error("algorithm message", "addr", "is empty address", "node id", p.ID().TerminalString())
		}
		return mc.PublishEvent(mc.P2P_HDMSG, &msgsend.AlgorithmMsg{Account: addr, Data: m})

	case msg.Code == common.BroadcastReqMsg:
		return p.SendPongToBroad([]uint8{0})

	case msg.Code == common.BroadcastRespMsg:
		return p2p.Record(p.ID())

	default:
		return errResp(ErrInvalidMsgCode, "%v", msg.Code)
	}
	return nil
}

// BroadcastBlock will either propagate a block to a subset of it's peers, or
// will only announce it's availability (depending what's requested).
func (pm *ProtocolManager) BroadcastBlock(block *types.Block, propagate bool) {
	role := ca.GetRole()
	pairOfPeer := make(map[bool][]*peer)

	hash := block.Hash()
	sbi, err := pm.blockchain.GetSuperBlockInfo()
	if nil != err {
		log.Error("get super seq error")
		return
	}
	peers := pm.Peers.PeersWithoutBlock(hash)
	//超级区块都广播
	if block.IsSuperBlock() {
		pairOfPeer[true] = peers
	} else {
		switch role {
		case common.RoleMiner, common.RoleBucket:
			if len(peers) == 0 {
				return
			}
			if len(peers) == 1 {
				pairOfPeer[true] = append(pairOfPeer[true], peers[0])
			}
			if len(peers) > 1 {
				in := p2p.Random(len(peers)-1, 1)
				if len(in) <= 0 {
					return
				}

				for index, peer := range peers {
					if index == in[0] {
						pairOfPeer[true] = append(pairOfPeer[true], peer)
						continue
					}
					pairOfPeer[false] = append(pairOfPeer[false], peer)
				}
			}

		case common.RoleValidator:

			miners := ca.GetRolesByGroup(common.RoleMiner | common.RoleBackupMiner | common.RoleInnerMiner)
			broads := ca.GetRolesByGroup(common.RoleBroadcast | common.RoleBackupBroadcast)
			sender := make(map[string]struct{})
			for _, m := range miners {
				if id := p2p.ServerP2p.ConvertAddressToId(m); id != emptyNodeId {
					sender[id.String()] = struct{}{}
				}
			}
			for _, b := range broads {
				if id := p2p.ServerP2p.ConvertAddressToId(b); id != emptyNodeId {
					sender[id.String()] = struct{}{}
				}
			}

			for _, peer := range peers {
				if _, ok := sender[peer.ID().String()]; ok {
					pairOfPeer[true] = append(pairOfPeer[true], peer)
				} else {
					pairOfPeer[false] = append(pairOfPeer[false], peer)
				}
			}

		default:
			//roles ,_:=ca.GetElectedByHeightAndRole(new(big.Int).Sub(block.Header().Number,big.NewInt(1)),common.RoleValidator)
			//isgoon := false
			//for _,role := range roles{
			//	if role.SignAddress == ca.GetAddress(){
			for _, peer := range peers {
				pairOfPeer[false] = append(pairOfPeer[false], peer)
			}
			//		isgoon = true
			//		break
			//	}
			//}
			//if !isgoon{
			//	return
			//}

		}
	}

	if peerSender, ok := pairOfPeer[true]; ok {
		// Calculate the TD of the block (it's not imported yet, so block.Td is not valid)
		var td *big.Int
		if parent := pm.blockchain.GetBlock(block.ParentHash(), block.NumberU64()-1); parent != nil {
			td = new(big.Int).Add(block.Difficulty(), pm.blockchain.GetTd(block.ParentHash(), block.NumberU64()-1))
		} else {
			log.Error("Propagating dangling block", "number", block.Number(), "hash", hash)
			return
		}

		for _, peer := range peerSender {
			peer.AsyncSendNewBlock(block, td, sbi.Num, sbi.Seq)
		}
		log.Trace("Propagated block", "hash", hash, "recipients", len(peerSender), "duration", common.PrettyDuration(time.Since(block.ReceivedAt)))
	}

	if peerOther, ok := pairOfPeer[false]; ok {
		// Otherwise if the block is indeed in out own chain, announce it
		if pm.blockchain.HasBlock(hash, block.NumberU64()) {
			for _, peer := range peerOther {
				peer.AsyncSendNewBlockHash(block)
			}
			log.Trace("Announced block", "hash", hash, "recipients", len(peerOther), "duration", common.PrettyDuration(time.Since(block.ReceivedAt)))
		}
	}
}

func (pm *ProtocolManager) AllBroadcastBlock(block *types.Block, propagate bool) {
	hash := block.Hash()
	sbi, err := pm.blockchain.GetSuperBlockInfo()
	if nil != err {
		log.ERROR("get super seq error")
		return
	}
	//	peers := pm.peers.PeersWithoutBlock(hash)
	peers := pm.Peers.PeersWithoutBlock(hash)

	// If propagation is requested, send to a subset of the peer
	if propagate {
		// Calculate the TD of the block (it's not imported yet, so block.Td is not valid)
		var td *big.Int
		if parent := pm.blockchain.GetBlock(block.ParentHash(), block.NumberU64()-1); parent != nil {
			td = new(big.Int).Add(block.Difficulty(), pm.blockchain.GetTd(block.ParentHash(), block.NumberU64()-1))
		} else {
			log.Error("Propagating dangling block", "number", block.Number(), "hash", hash)
			return
		}
		// Send the block to a subset of our peers
		for _, peer := range peers {
			peer.AsyncSendNewBlock(block, td, sbi.Num, sbi.Seq)
		}
		log.Trace("Propagated block", "hash", hash, "recipients", len(peers), "duration", common.PrettyDuration(time.Since(block.ReceivedAt)))
		return
	}
	// Otherwise if the block is indeed in out own chain, announce it
	if pm.blockchain.HasBlock(hash, block.NumberU64()) {
		for _, peer := range peers {
			peer.AsyncSendNewBlockHash(block)
		}
		log.Trace("Announced block", "hash", hash, "recipients", len(peers), "duration", common.PrettyDuration(time.Since(block.ReceivedAt)))
	}
}

// BroadcastBlock will either propagate a block to a subset of it's peers, or
// will only announce it's availability (depending what's requested).
//todo: debug
func (pm *ProtocolManager) BroadcastBlockHeader(block *types.Block, propagate bool) {
	hash := block.Hash()
	peers := pm.Peers.PeersWithoutBlock(hash)

	// If propagation is requested, send to a subset of the peer
	if propagate {
		// Calculate the TD of the block (it's not imported yet, so block.Td is not valid)
		// Send the block to a subset of our peers
		transfer := peers[:int(math.Sqrt(float64(len(peers))))]
		headers := []*types.Header{block.Header()}
		for _, peer := range transfer {
			peer.SendBlockHeaders(headers)
		}
		log.Trace("Propagated block", "hash", hash, "recipients", len(transfer), "duration", common.PrettyDuration(time.Since(block.ReceivedAt)))
		return
	}
	// Otherwise if the block is indeed in out own chain, announce it
	if pm.blockchain.HasBlock(hash, block.NumberU64()) {
		for _, peer := range peers {
			peer.AsyncSendNewBlockHash(block)
		}
		log.Trace("Announced block", "hash", hash, "recipients", len(peers), "duration", common.PrettyDuration(time.Since(block.ReceivedAt)))
	}
}

// BroadcastTxs will propagate a batch of transactions to all peers which are not known to
// already have the given transaction.
func (pm *ProtocolManager) BroadcastTxs(txs types.SelfTransactions) {
	var txset = make(map[*peer]types.SelfTransactions)

	// Broadcast transactions to a batch of peers not knowing about it
	for _, tx := range txs {
		//		peers := pm.peers.PeersWithoutTx(tx.Hash())
		peers := pm.Peers.PeersWithoutTx(tx.Hash())
		for _, peer := range peers {
			txset[peer] = append(txset[peer], tx)
		}
		log.Trace("Broadcast transaction", "hash", tx.Hash(), "recipients", len(peers))
	}
	// udp send
	if ca.GetRole() == common.RoleDefault {
		SendUdpTransactions(txs)
	}
	// FIXME include this again: peers = peers[:int(math.Sqrt(float64(len(peers))))]
	for peer, txs := range txset {
		peer.AsyncSendTransactions(txs)
	}
}

// Mined broadcast loop
func (pm *ProtocolManager) minedBroadcastLoop() {
	// automatically stops if unsubscribe
	for obj := range pm.minedBlockSub.Chan() {
		switch ev := obj.Data.(type) {
		case core.NewMinedBlockEvent:
			pm.BroadcastBlock(ev.Block, true)  // First propagate block to peers
			pm.BroadcastBlock(ev.Block, false) // Only then announce to the rest
			//TOdo: broadcast block header
			//pm.BroadcastBlockHeader(ev.Block, true)
			//pm.BroadcastBlockHeader(ev.Block, false)
		}
	}
}

func (pm *ProtocolManager) txBroadcastLoop() {
	for {
		select {
		case event := <-pm.txsCh:
			pm.BroadcastTxs(event.Txs)

		// Err() channel will be closed when unsubscribing.
		case <-pm.txsSub.Err():
			return
		}
	}
}

// NodeInfo represents a short summary of the Matrix sub-protocol metadata
// known about the host peer.
type NodeInfo struct {
	Network    uint64              `json:"network"`    // Matrix network ID (1=Frontier, 2=Morden, Ropsten=3, Rinkeby=4)
	Difficulty *big.Int            `json:"difficulty"` // Total difficulty of the host's blockchain
	Genesis    common.Hash         `json:"genesis"`    // SHA3 hash of the host's genesis block
	Config     *params.ChainConfig `json:"config"`     // Chain configuration for the fork rules
	Head       common.Hash         `json:"head"`       // SHA3 hash of the host's best owned block
}

// NodeInfo retrieves some protocol metadata about the running host node.
func (pm *ProtocolManager) NodeInfo() *NodeInfo {
	currentBlock := pm.blockchain.CurrentBlock()
	return &NodeInfo{
		Network:    pm.networkId,
		Difficulty: pm.blockchain.GetTd(currentBlock.Hash(), currentBlock.NumberU64()),
		Genesis:    pm.blockchain.Genesis().Hash(),
		Config:     pm.blockchain.Config(),
		Head:       currentBlock.Hash(),
	}
}
