// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package lessdisk

import (
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/event"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/params"
	"sync"
	"time"
)

type Server struct {
	blkInsertedMsgCh  chan *mc.BlockInsertedMsg
	blkInsertedMsgSub event.Subscription
	logInfo           string
	funcSwitch        bool
	config            *params.LessDiskConfig
	timer             *time.Timer
	mu                sync.Mutex
	quit              chan struct{}
	indexOperator     *indexOperator
	chain             ChainOperator
}

func NewLessDiskSvr(config *params.LessDiskConfig, db DatabaseOperator, chain ChainOperator) *Server {
	logInfo := "LessDiskServer"
	svr := &Server{
		blkInsertedMsgCh:  make(chan *mc.BlockInsertedMsg, 10),
		blkInsertedMsgSub: nil,
		logInfo:           logInfo,
		funcSwitch:        false,
		config:            config,
		timer:             nil,
		quit:              make(chan struct{}),
		indexOperator:     newIndexOperator(logInfo, db),
		chain:             chain,
	}

	var err error
	if svr.blkInsertedMsgSub, err = mc.SubscribeEvent(mc.BlockInserted, svr.blkInsertedMsgCh); err != nil {
		log.Error(svr.logInfo, "订阅<BlockInserted>事件错误(%v)", err)
	}

	go svr.runIndexUpdate()
	go svr.runDelBlk()
	return svr
}

func (self *Server) Stop() {
	close(self.quit)
}

func (self *Server) FuncSwitch(enable bool) {
	self.funcSwitch = enable
}

func (self *Server) runIndexUpdate() {
	for {
		select {
		case msg := <-self.blkInsertedMsgCh:
			self.updateIndex(msg)
		case <-self.quit:
			return
		}
	}
}

func (self *Server) runDelBlk() {
	optInterval := time.Duration(self.config.OptInterval) * time.Second
	self.timer = time.NewTimer(optInterval)

	for {
		select {
		case <-self.timer.C:
			self.delBlk()
			self.timer = time.NewTimer(optInterval)
		case <-self.quit:
			return
		}
	}
}

func (self *Server) updateIndex(msg *mc.BlockInsertedMsg) {
	if msg == nil {
		return
	}

	if msg.Block.Number == 0 {
		log.Debug(self.logInfo, "更新索引", "创世区块不处理")
		return
	}

	log.Trace(self.logInfo, "更新索引", "BlockInsertedMsg", "高度", msg.Block.Number, "hash", msg.Block.Hash, "insertTime", msg.InsertTime)

	self.mu.Lock()
	defer self.mu.Unlock()

	index := self.indexOperator.readBlkIndex(msg.Block.Number)
	chg := false
	if index, chg = updateIndexSlice(msg.Block.Hash, msg.InsertTime, index); chg == false {
		log.Debug(self.logInfo, "更新索引", "区块索引已存在", "number", msg.Block.Number, "hash", msg.Block.Hash.Hex())
		return
	}
	if err := self.indexOperator.writeBlkIndex(msg.Block.Number, index); err != nil {
		log.Error(self.logInfo, "更新索引", "保存区块索引失败", "err", err, "number", msg.Block.Number, "hash", msg.Block.Hash.Hex())
		return
	}
	if err := self.indexOperator.UpdateMinNumberIndex(msg.Block.Number); err != nil {
		log.Error(self.logInfo, "更新索引", "更新最低高度索引失败", "err", err, "number", msg.Block.Number, "hash", msg.Block.Hash.Hex())
		return
	}
}

func (self *Server) delBlk() {
	if self.funcSwitch == false {
		log.Debug(self.logInfo, "删除区块", "功能未开启")
		return
	}

	header := self.chain.CurrentHeader()
	if header == nil {
		log.Warn(self.logInfo, "删除区块", "获取主链当前区块失败")
		return
	}

	self.mu.Lock()
	defer self.mu.Unlock()
	curTime := time.Now().Unix()
	curNumber := header.Number.Uint64()
	minNumber := self.indexOperator.readMinNumberIndex()
	log.Debug(self.logInfo, "删除区块", "开始", "当前高度", curNumber, "最低高度", minNumber, "高度阈值", self.config.HeightThreshold)
	if minNumber == 0 {
		log.Debug(self.logInfo, "删除区块", "获取最低高度失败")
		return
	}

	targetNumber := curNumber - self.config.HeightThreshold
	if targetNumber <= minNumber {
		log.Debug(self.logInfo, "删除区块", "当前区块高度未达到高度阈值")
		return
	}

	newMinNumber := minNumber
	delBlks := make([]*mc.BlockInfo, 0)
	for i := minNumber; i < targetNumber; i++ {
		blkIndex := self.indexOperator.readBlkIndex(i)
		if rlt, blk := hasBlockNotOutTime(curTime, self.config.TimeThreshold, blkIndex); rlt == true {
			log.Debug(self.logInfo, "删除区块", "有区块不满足时间阈值", "number", i, "check begin number", minNumber, "hash", blk.Hash.Hex(), "insertTime", blk.InsertTime, "timeThreshold", self.config.TimeThreshold)
			break
		}
		for j := 0; j < len(blkIndex); j++ {
			delBlks = append(delBlks, &mc.BlockInfo{Hash: blkIndex[j].Hash, Number: i})
		}
		newMinNumber = i + 1
	}

	fails, err := self.chain.DelLocalBlocks(delBlks)
	if err != nil {
		log.Debug(self.logInfo, "删除区块", "链删除本地区块数据失败", "err", err, "fails", fails)
		for _, item := range fails {
			if item.Number < newMinNumber {
				newMinNumber = item.Number
			}
		}
	}
	log.Debug(self.logInfo, "删除区块", "更新最低区块高度索引", "old", minNumber, "new", newMinNumber)
	if newMinNumber != minNumber {
		for i := minNumber; i < newMinNumber; i++ {
			if err := self.indexOperator.deleteBlkIndex(i); err != nil {
				log.Error(self.logInfo, "删除区块", "删除区块索引失败", "err", err, "number", i)
			}
		}
		self.indexOperator.writeMinNumberIndex(newMinNumber)
	}
}

func updateIndexSlice(hash common.Hash, insertTime uint64, index []dbBlkIndex) ([]dbBlkIndex, bool) {
	if len(index) == 0 {
		return append(index, dbBlkIndex{Hash: hash, InsertTime: insertTime}), true
	}
	for i := 0; i < len(index); i++ {
		if index[i].Hash == hash {
			if index[i].InsertTime == insertTime {
				return index, false
			} else {
				index[i].InsertTime = insertTime
				return index, true
			}
		}
	}
	return append(index, dbBlkIndex{Hash: hash, InsertTime: insertTime}), true
}

func hasBlockNotOutTime(curTime int64, timeThreshold int64, index []dbBlkIndex) (bool, *dbBlkIndex) {
	if len(index) == 0 {
		return false, nil
	}
	targetTime := curTime - timeThreshold
	for _, item := range index {
		if int64(item.InsertTime) > targetTime {
			return true, &item
		}
	}
	return false, nil
}
