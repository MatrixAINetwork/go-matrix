// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package man

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/MatrixAINetwork/go-matrix/params/manparams"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/man/downloader"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/p2p/discover"
	"github.com/MatrixAINetwork/go-matrix/snapshot"
)

const (
	forceSyncCycle      = 10 * time.Second // Time interval to force syncs, even if few peers are available
	minDesiredPeerCount = 5                // Amount of peers desired to start syncing

	// This is the target size for the packs of transactions sent by txsyncLoop.
	// A pack can get larger than this if a single transactions exceeds this size.
	txsyncPackSize = 100 * 1024
)

var (
	SnapshootNumber   uint64
	SnapshootHash     string
	SaveSnapStart     uint64
	SaveSnapPeriod    uint64 = 300
	SnaploadFromLocal int    = 0
	ManualSaveSnapNum uint64 = 0
	AutoLoadSnapFlg   bool   = false
	SnapLoadFile      string
)

type txsync struct {
	p   *peer
	txs []types.SelfTransaction
}

// syncTransactions starts sending all currently pending transactions to the given peer.
func (pm *ProtocolManager) syncTransactions(p *peer) {
	var txs types.SelfTransactions
	pending, _ := pm.txpool.Pending()
	for _, txsmap := range pending {
		for _, batch := range txsmap {
			txs = append(txs, batch...)
		}
	}
	if len(txs) == 0 {
		return
	}
	select {
	case pm.txsyncCh <- &txsync{p, txs}:
	case <-pm.quitSync:
	}
}

// txsyncLoop takes care of the initial transaction sync for each new
// connection. When a new peer appears, we relay all currently pending
// transactions. In order to minimise egress bandwidth usage, we send
// the transactions in small packs to one peer at a time.
func (pm *ProtocolManager) txsyncLoop() {
	var (
		pending = make(map[discover.NodeID]*txsync)
		sending = false               // whether a send is active
		pack    = new(txsync)         // the pack that is being sent
		done    = make(chan error, 1) // result of the send
	)

	// send starts a sending a pack of transactions from the sync.
	send := func(s *txsync) {
		// Fill pack with transactions up to the target size.
		size := common.StorageSize(0)
		pack.p = s.p
		pack.txs = pack.txs[:0]
		for i := 0; i < len(s.txs) && size < txsyncPackSize; i++ {
			pack.txs = append(pack.txs, s.txs[i])
			size += s.txs[i].Size()
		}
		// Remove the transactions that will be sent.
		s.txs = s.txs[:copy(s.txs, s.txs[len(pack.txs):])]
		if len(s.txs) == 0 {
			delete(pending, s.p.ID())
		}
		// Send the pack in the background.
		s.p.Log().Trace("Sending batch of transactions", "count", len(pack.txs), "bytes", size)
		sending = true
		go func() { done <- pack.p.SendTransactions(pack.txs) }()
	}

	// pick chooses the next pending sync.
	pick := func() *txsync {
		if len(pending) == 0 {
			return nil
		}
		n := rand.Intn(len(pending)) + 1
		for _, s := range pending {
			if n--; n == 0 {
				return s
			}
		}
		return nil
	}

	for {
		select {
		case s := <-pm.txsyncCh:
			pending[s.p.ID()] = s
			if !sending {
				send(s)
			}
		case err := <-done:
			sending = false
			// Stop tracking peers that cause send failures.
			if err != nil {
				pack.p.Log().Debug("Transaction send failed", "err", err)
				delete(pending, pack.p.ID())
			}
			// Schedule the next send.
			if s := pick(); s != nil {
				send(s)
			}
		case <-pm.quitSync:
			return
		}
	}
}

//lb WaitForDownLoadMode ...
func (pm *ProtocolManager) WaitForDownLoadMode() {

	syncRoleCH := make(chan mc.SyncIdEvent, 1)
	sub, _ := mc.SubscribeEvent(mc.SendSyncRole, syncRoleCH)
	fmt.Println("download sync.go  WaitForDownLoadMode enter")
	log.WARN("download sync.go  WaitForDownLoadMode enter")
	select {
	case id := <-syncRoleCH:
		if id.Role == common.RoleBroadcast {
			log.Warn("download sync.go syncer wait role is Broadcast")
			fmt.Println("download sync.go syncer wait role is Broadcast")
			pm.downloader.SetbStoreSendIpfsFlg(true)
			go pm.downloader.SynBlockFormBlockchain()
			//go pm.downloader.StatusSnapshootDeal()
			//return
		} else {
			log.Warn("download sync.go syncer wait role is generaler")
			fmt.Println("download sync.go syncer wait role is generaler")
			pm.downloader.SetbStoreSendIpfsFlg(false)
			go pm.downloader.IpfsProcessRcvHead()
			go pm.downloader.WaitBlockInfoFromIpfs()
			go pm.downloader.SynIPFSCheck()
			//return
		}
	}
	sub.Unsubscribe()
	log.WARN("download sync.go  WaitForDownLoadMode out")

}

// syncer is responsible for periodically synchronising with the network, both
// downloading hashes and blocks as well as handling the announcement handler.
func (pm *ProtocolManager) syncer() {
	// Start and ensure cleanup of sync mechanisms
	pm.fetcher.Start()
	defer pm.fetcher.Stop()
	defer pm.downloader.Terminate()
	log.WARN("syncer", "syncer IpfsDownloadflg", pm.downloader.IpfsMode)
	if pm.downloader.IpfsMode {
		pm.WaitForDownLoadMode()
	}
	pm.blockchain.SetSnapshotParam(SaveSnapPeriod, SaveSnapStart)
	//手动保存快照功能
	if ManualSaveSnapNum != 0 {
		fmt.Println("matrix begin save snapshot! snapshot block Number", ManualSaveSnapNum)
		go pm.blockchain.SaveSnapshot(0, 0, ManualSaveSnapNum)
	}
	curBlkNum := pm.blockchain.CurrentBlock().NumberU64()
	//自动加载手动保存的快照
	if SnapLoadFile != "" {
		filePath := path.Join(snapshot.SNAPDIR, SnapLoadFile)
		if common.FileExist(filePath) == false {
			fmt.Println("matrix  load local snapshoot not find file", SnapLoadFile)
			log.Error("matrix  load local snapshoot not find file", "file", SnapLoadFile)
			os.Exit(1)
		} else {
			fmt.Println("matrix  load local snapshoot start", SnapLoadFile)
			blockNum, flg := pm.blockchain.SynSnapshot(0, "", filePath)
			if flg == false {
				fmt.Println("Info!  matrix  can't load and use snapshoot because local chaindata is heigher than snap or other, SnapNum=", blockNum, curBlkNum)
				log.Warn("Info!  matrix  can't load and use snapshoot because local chaindata is heigher than snap or other", "SnapNum", blockNum, "curBlkNum", curBlkNum)
			} else {
				pm.downloader.SetSnapshootNum(blockNum)
				fmt.Println("matrix  load local snapshoot and use sucess, blockNum=", blockNum)
				log.Debug("matrix  load local snapshoot and use sucess", "blockNum", blockNum)
			}
		}
	} else if AutoLoadSnapFlg == true {
		var (
			maxTime  int64
			fileName string
		)
		files, _ := ioutil.ReadDir(snapshot.SNAPDIR)
		for _, onefile := range files {
			if onefile.ModTime().Unix() > maxTime {
				maxTime = onefile.ModTime().Unix()
				fileName = onefile.Name()
			}
		}
		if fileName == "" {
			fmt.Println("matrix auto load local snapshoot can't find snap file")
			log.Error("matrix auto load local snapshoot can't find snap file")
		} else {
			fmt.Println("matrix auto load local snapshoot start", fileName)
			filePath := path.Join(snapshot.SNAPDIR, fileName)
			blockNum, flg := pm.blockchain.SynSnapshot(0, "", filePath)
			if flg == false {
				fmt.Println("Info!  matrix  can't auto load and use snapshoot because local chaindata is heigher than snap or other,SnapNum=", blockNum, curBlkNum)
				log.Warn("Info!  matrix  can't auto load and use snapshoot because local chaindata is heigher than snap or other,please check", "SnapNum", blockNum, "curBlkNum", curBlkNum)
			} else {
				pm.downloader.SetSnapshootNum(blockNum)
				fmt.Println("matrix auto load local snapshoot and use sucess, blockNum=", blockNum)
				log.Debug("matrix auto load local snapshoot and use sucess", "AutoLoadSnapNum", blockNum)
			}
		}

	} else if SnapshootNumber != 0 { //快照下载 SnaploadFromLoacl
		if SnaploadFromLocal == 0 {
			fmt.Println("snapshoot  DownLoad start blockNum=", SnapshootNumber)
			pm.downloader.SetSnapshootNum(SnapshootNumber)
			log.Warn("download  Snapshoot status will begin", "number", SnapshootNumber, "shash", SnapshootHash)
			time.Sleep(10 * time.Second)
			err := pm.downloader.ProcessSnapshoot(uint64(SnapshootNumber), SnapshootHash)
			if err != nil {
				log.Error(" ipfs download snapshoot  error ", "err", err)
				os.Exit(1)
			}
			res := <-pm.downloader.WaitSnapshoot
			log.Debug(" ipfs download DownloadBatchBlock get status MPT over")
			if res == 0 {
				log.Error(" ipfs download snapshoot or deal error and exit,please check")
				os.Exit(1)
			}
		} else {
			pm.downloader.SetSnapshootNum(SnapshootNumber)
			filePath := path.Join(snapshot.SNAPDIR, "/TrieData"+strconv.Itoa(int(SnapshootNumber)))
			if blkNum, flg := pm.blockchain.SynSnapshot(SnapshootNumber, "", filePath); flg == false {
				log.Error(" ipfs local snapshoot deal error and exit,please check", "blkNum", blkNum)
				os.Exit(1)
			}
		}
		if SnapshootNumber != curBlkNum {
			log.Error(" snapshoot deal over,but block Num is illegal", "SnapshootNumber", SnapshootNumber, "current block", curBlkNum)
			os.Exit(1)
		}
		fmt.Println("snapshoot  DownLoad and use sucess, blockNum=", SnapshootNumber)
	}
	// Wait for different events to fire synchronisation operations
	forceSync := time.NewTicker(forceSyncCycle)
	defer forceSync.Stop()
	log.Warn("download  syncer will begin running")
	for {
		select {
		case <-pm.newPeerCh:
			// Make sure we have peers to select from, then sync
			if pm.Peers.Len() < minDesiredPeerCount {
				break
			}
			go pm.synchronise(pm.Peers.BestPeer(), 0)

		case <-forceSync.C:
			// Force a sync even if not enough peers are present
			go pm.synchronise(pm.Peers.BestPeer(), 1)

		case <-pm.noMorePeers:
			return
		}
	}
}

// synchronise tries to sync up our local block chain with a remote peer.
func (pm *ProtocolManager) synchronise(peer *peer, flg int) {
	// Short circuit if no peers are available
	if peer == nil {
		return
	}
	// Make sure the peer's TD is higher than our own
	if flg == 3 || flg == 4 || flg == 10 || flg == 0 {
		nowCheckTime := time.Now().Unix()
		//if pm.LastCheckBlkNum == bn {
		if nowCheckTime-pm.LastCheckTime < 2 {
			return
		}
		pm.LastCheckTime = nowCheckTime
		//} else {
		//	pm.LastCheckBlkNum = bn
		//}
	}

	currentBlock := pm.blockchain.CurrentBlock()
	td := pm.blockchain.GetTd(currentBlock.Hash(), currentBlock.NumberU64())
	sbi, err := pm.blockchain.GetSuperBlockInfo()
	if nil != err {
		log.Error("get super seq error")
		return
	}
	needForceDownload := false
	if flg == 1 {
		pm.downloader.CheckNum++
		if pm.downloader.CheckNum > 180 {
			//needForceDownload = true   //30 分钟强制同步
		}
	}

	log.Trace("download sync.go enter Synchronise peer", "peer", peer, "flg", flg, "currentBlock", currentBlock.NumberU64())
	sbs := sbi.Seq
	sbh := sbi.Num
	pHead, pTd, pSbs, pSbh, bt, bn := peer.Head()

	if flg == 10 {
		needForceDownload = true
	}

	log.Trace("download sync.go enter Synchronise td", "td", td, "pTd", pTd, "Sbs", sbs, "pSbs", pSbs, "bheight", bn, "pSbh", pSbh, "currentBlock", currentBlock.NumberU64())
	if pSbs < sbs {
		go peer.SendBlockHeaders([]*types.Header{currentBlock.Header()})
		go peer.AsyncSendNewBlock(currentBlock, td, sbh, sbs)
		log.Trace("对端peer超级序号小于本地的序号", "本地序号", sbs, "peer序号", pSbs, "peer hex", peer.id)
		return
	}
	if needForceDownload {
		log.Trace("download sync.go force into")
	} else if pSbs == sbs {
		if manparams.CanSwitchGammaCanonicalChain(time.Now().Unix()) {
			if bn < currentBlock.NumberU64() {
				log.Trace("对端peer高度小于本地的高度", "本地高度", currentBlock.NumberU64(), "对端高度", bn, "peer hex", peer.id)
				return
			} else if bn == currentBlock.NumberU64() {
				if bt < currentBlock.Time().Uint64() {
					log.Trace("对端peer高度小于本地的高度", "本地时间", currentBlock.Time(), "对端时间", bt, "peer hex", peer.id)
					return
				}
			} else if (flg == 3) && (bn == currentBlock.NumberU64()+1) { //若是对端比本地大于1,多等一次机会
				if pm.downloader.CheckWaitNum++; pm.downloader.CheckWaitNum < 2 {
					log.Trace("download 需等待下次调度")
					return
				}

			}
		} else {
			if nil == td || pTd.Cmp(td) <= 0 {
				log.Trace("对端peer超级td小于本地的td", "本地td", td, "peertd", pTd, "peer hex", peer.id)
				return
			}
		}

	}
	pm.downloader.CheckWaitNum = 0
	log.Warn("download sync.go enter Synchronise", "currentBlock", currentBlock.NumberU64())
	// Otherwise try to sync with the downloader
	mode := downloader.FullSync
	/* //fast 模式不启用
	if atomic.LoadUint32(&pm.fastSync) == 1 {
		// Fast sync was explicitly requested, and explicitly granted
		mode = downloader.FastSync
		log.Trace("download sync.go enter Synchronise fastSync", "currentBlock", currentBlock.NumberU64())
	} else if currentBlock.NumberU64() == 0 && pm.blockchain.CurrentFastBlock().NumberU64() > 0 {
		// The database seems empty as the current block is the genesis. Yet the fast
		// block is ahead, so fast sync was enabled for this node at a certain point.
		// The only scenario where this can happen is if the user manually (or via a
		// bad block) rolled back a fast sync node below the sync point. In this case
		// however it's safe to reenable fast sync.
		atomic.StoreUint32(&pm.fastSync, 1)
		mode = downloader.FastSync
		log.Trace("download sync.go enter Synchronise set fastSync", "currentBlock", currentBlock.NumberU64())
	}*/

	//log.Trace("download sync.go enter Synchronise downloader", "currentBlock", currentBlock.NumberU64())
	// Run the sync cycle, and disable fast sync if we've went past the pivot block
	if err := pm.downloader.Synchronise(peer.id, pHead, pTd, pSbs, pSbh, mode); err != nil {
		return
	}
	if atomic.LoadUint32(&pm.fastSync) == 1 {
		log.Info("Fast sync complete, auto disabling")
		atomic.StoreUint32(&pm.fastSync, 0)
	}
	atomic.StoreUint32(&pm.acceptTxs, 1) // Mark initial sync done

	if pSbs < sbs {
		go peer.SendBlockHeaders([]*types.Header{currentBlock.Header()})
		go peer.AsyncSendNewBlock(currentBlock, td, sbh, sbs)
	} else if head := pm.blockchain.CurrentBlock(); head.NumberU64() > 0 {
		// We've completed a sync cycle, notify all peers of new state. This path is
		// essential in star-topology networks where a gateway node needs to notify
		// all its out-of-date peers of the availability of a new block. This failure
		// scenario will most often crop up in private and hackathon networks with
		// degenerate connectivity, but it should be healthy for the mainnet too to
		// more reliably update peers or the local TD state.
		go pm.BroadcastBlock(head, false)
	}
}
