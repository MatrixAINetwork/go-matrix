// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package miner

import (
	"math/big"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/params"
	"github.com/pkg/errors"
)

type mineTaskManager struct {
	curNumber        uint64
	curRole          common.RoleType
	powTaskCache     map[common.Hash]*powMineTask
	aiTaskCache      map[common.Hash]*aiMineTask
	unusedMineHeader map[common.Hash]*types.Header
	bc               ChainReader
	logInfo          string
}

func newMineTaskManager(bc ChainReader, logInfo string) *mineTaskManager {
	return &mineTaskManager{
		curNumber:        0,
		curRole:          common.RoleNil,
		powTaskCache:     make(map[common.Hash]*powMineTask),
		aiTaskCache:      make(map[common.Hash]*aiMineTask),
		unusedMineHeader: make(map[common.Hash]*types.Header),
		bc:               bc,
		logInfo:          logInfo,
	}
}

func (mgr *mineTaskManager) Clear() {
	mgr.curNumber = 0
	mgr.powTaskCache = make(map[common.Hash]*powMineTask)
	mgr.aiTaskCache = make(map[common.Hash]*aiMineTask)
	mgr.unusedMineHeader = make(map[common.Hash]*types.Header)
	return
}

func (mgr *mineTaskManager) SetNewNumberAndRole(number uint64, role common.RoleType) {
	if mgr.curNumber > number {
		return
	}

	if mgr.curNumber < number {
		mgr.curNumber = number
		mgr.fixMap() // todo 删除和更新分开做
	}

	mgr.curRole = role
	return
}

func (mgr *mineTaskManager) AddTaskByInsertedHeader(headerHash common.Hash) {
	insertedHeader := mgr.bc.GetHeaderByHash(headerHash)
	if nil == insertedHeader {
		log.Info(mgr.logInfo, "AddTaskByInsertedHeader", "get cur header failed")
		return
	}

	bcInterval, err := mgr.bc.GetBroadcastIntervalByHash(headerHash)
	if err != nil || bcInterval == nil {
		log.Info(mgr.logInfo, "AddTaskByInsertedHeader", "get broadcast interval failed")
		return
	}

	insertedNumber := insertedHeader.Number.Uint64()
	if bcInterval.IsReElectionNumber(insertedNumber) {
		log.Trace(mgr.logInfo, "AddTaskByInsertedHeader", "忽略区块", "插入区块是选举换届区块", insertedNumber)
		return
	}

	var aiHeader *types.Header = nil
	if insertedHeader.IsAIHeader(bcInterval.GetBroadcastInterval()) {
		// 插入区块即是AI区块
		aiHeader = insertedHeader
	} else {
		// 获取上一个AI区块
		aiHeaderNumber := params.GetCurAIBlockNumber(insertedHeader.Number.Uint64(), bcInterval.GetBroadcastInterval())
		aiHeaderHash, err := mgr.bc.GetAncestorHash(insertedHeader.ParentHash, aiHeaderNumber)
		if err != nil {
			log.Info(mgr.logInfo, "AddTaskByInsertedHeader", "获取pre ai header hash 失败", "err", err, "aiHeaderNumber", aiHeaderNumber, "cur header number", insertedHeader.Number)
			return
		}
		aiHeader = mgr.bc.GetHeaderByHash(aiHeaderHash)
		if aiHeader == nil {
			log.Info(mgr.logInfo, "AddTaskByInsertedHeader", "get pre ai header failed")
			return
		}
	}

	powTask, aiTask, err := mgr.createMineTask(aiHeader, false)
	if err != nil {
		log.Trace(mgr.logInfo, "AddTaskByInsertedHeader", "创建任务失败", "err", err)
		return
	}
	mgr.addPowTask(powTask)
	mgr.addAITask(aiTask)
}

func (mgr *mineTaskManager) AddMineHeader(mineHeader *types.Header) error {
	if nil == mineHeader {
		return errors.New("mine header为nil")
	}

	if mineHeader.Number.Uint64()+params.PowBlockPeriod+1 < mgr.curNumber {
		// 高度过低
		return errors.Errorf("mine header number(%d) is too less than cur number(%d)", mineHeader.Number.Uint64(), mgr.curNumber)
	}

	mineHash := mineHeader.HashNoSignsAndNonce()
	if mgr.isExistMineHash(mineHash) {
		return errors.Errorf("mine hash(%s) already exist", mineHash.TerminalString())
	}

	if mgr.bc.GetHeaderByHash(mineHeader.ParentHash) == nil {
		// 没有父区块的挖矿header 先缓存
		for len(mgr.unusedMineHeader) > OVERFLOWLEN {
			var earliestHeaderTime *big.Int = nil
			var earliestHash common.Hash
			for hash, header := range mgr.unusedMineHeader {
				if earliestHeaderTime != nil && earliestHeaderTime.Cmp(header.Time) <= 0 {
					continue
				}
				earliestHeaderTime = header.Time
				earliestHash = hash
			}
			delete(mgr.unusedMineHeader, earliestHash)
		}
		mgr.unusedMineHeader[mineHash] = mineHeader
		return nil
	} else {
		// 已有父区块，将mineHeader转换成mineTask
		powTask, aiTask, err := mgr.createMineTask(mineHeader, true)
		if err != nil {
			log.Info(mgr.logInfo, "create mine task err", err)
			return err
		}

		mgr.addPowTask(powTask)
		mgr.addAITask(aiTask)
		return nil
	}
}

func (mgr *mineTaskManager) CanMining() bool {
	return mgr.curRole == common.RoleMiner || mgr.curRole == common.RoleInnerMiner
}

func (mgr *mineTaskManager) GetBestPowTask() (bestTask *powMineTask) {
	bestTask = nil
	for hash, task := range mgr.powTaskCache {
		if task.minedPow {
			continue
		}
		if task.powMiningNumber < mgr.curNumber {
			log.Info(mgr.logInfo, "GetBestPowTask", "task mining number < cur number", "hash", hash.TerminalString(), "task mining number", task.powMiningNumber, "cur number", mgr.curNumber)
			delete(mgr.aiTaskCache, hash)
			continue
		}
		if bestTask != nil && bestTask.mineHeader.Time.Cmp(task.mineHeader.Time) >= 0 {
			// 时间搓最大的task 为最好的task
			continue
		}
		bestTask = task
	}

	return bestTask
}

func (mgr *mineTaskManager) GetBestAITask() (bestTask *aiMineTask) {
	bestTask = nil
	for hash, task := range mgr.aiTaskCache {
		if task.minedAI {
			continue
		}
		if task.aiMiningNumber < mgr.curNumber {
			log.Info(mgr.logInfo, "GetBestAITask", "task mining number < cur number", "hash", hash.TerminalString(), "task mining number", task.aiMiningNumber, "cur number", mgr.curNumber)
			delete(mgr.aiTaskCache, hash)
			continue
		}
		if bestTask != nil && bestTask.mineHeader.Time.Cmp(task.mineHeader.Time) >= 0 {
			// 时间搓最大的task 为最好的task
			continue
		}
		bestTask = task
	}

	return bestTask
}

func (mgr *mineTaskManager) createMineTask(mineHeader *types.Header, verify bool) (powTask *powMineTask, aiTask *aiMineTask, returnErr error) {
	bcInterval, err := mgr.bc.GetBroadcastIntervalByHash(mineHeader.ParentHash)
	if err != nil || bcInterval == nil {
		return nil, nil, errors.Errorf("get broadcast interval err: %v", err)
	}

	if verify {
		if mineHeader.Difficulty.Uint64() == 0 {
			return nil, nil, difficultyIsZero
		}

		if mineHeader.IsAIHeader(bcInterval.GetBroadcastInterval()) == false {
			return nil, nil, errors.Errorf("mine header is not ai header")
		}

		err = mgr.bc.DPOSEngine(mineHeader.Version).VerifyBlock(mgr.bc, mineHeader)
		if err != nil {
			return nil, nil, errors.Errorf("verify mine header err: %v", err)
		}
	}

	powMiningNumber := mineHeader.Number.Uint64() + params.PowBlockPeriod - 1
	aiMiningNumber := params.GetNextAIBlockNumber(mineHeader.Number.Uint64(), bcInterval.GetBroadcastInterval())

	mineHash := mineHeader.HashNoSignsAndNonce()
	difficulty := mineHeader.Difficulty
	if mgr.curRole == common.RoleInnerMiner {
		difficulty = params.InnerMinerDifficulty
	}

	powTask = newPowMineTask(mineHash, mineHeader, powMiningNumber, bcInterval, difficulty)
	if bcInterval.IsReElectionNumber(aiMiningNumber - 1) {
		aiTask = nil
	} else {
		aiTask = newAIMineTask(mineHash, mineHeader, aiMiningNumber, bcInterval)
	}

	return powTask, aiTask, nil
}

func (mgr *mineTaskManager) addPowTask(powTask *powMineTask) {
	if powTask == nil {
		return
	}
	if powTask.powMiningNumber < mgr.curNumber {
		log.Trace(mgr.logInfo, "add pow task failed", "task number < cur number", "task number", powTask.powMiningNumber, "cur number", mgr.curNumber)
		return
	}

	_, exist := mgr.powTaskCache[powTask.mineHash]
	if exist {
		log.Trace(mgr.logInfo, "add pow task failed", "already exist", "task number", powTask.powMiningNumber, "mine hash", powTask.mineHash.TerminalString())
		return
	}
	mgr.powTaskCache[powTask.mineHash] = powTask
	log.Info(mgr.logInfo, "add pow task success", powTask.mineHash.TerminalString(), "mining number", powTask.powMiningNumber, "cur number", mgr.curNumber)
}

func (mgr *mineTaskManager) addAITask(aiTask *aiMineTask) {
	if aiTask == nil {
		return
	}
	if aiTask.aiMiningNumber < mgr.curNumber {
		log.Trace(mgr.logInfo, "add ai task failed", "task number < cur number", "task number", aiTask.aiMiningNumber, "cur number", mgr.curNumber)
		return
	}

	_, exist := mgr.aiTaskCache[aiTask.mineHash]
	if exist {
		log.Trace(mgr.logInfo, "add ai task failed", "already exist", "task number", aiTask.aiMiningNumber, "mine hash", aiTask.mineHash.TerminalString())
		return
	}
	mgr.aiTaskCache[aiTask.mineHash] = aiTask
	log.Info(mgr.logInfo, "add ai task success", aiTask.mineHash.TerminalString(), "mining number", aiTask.aiMiningNumber, "cur number", mgr.curNumber)
}

func (mgr *mineTaskManager) isExistMineHash(mineHash common.Hash) bool {
	if _, exist := mgr.powTaskCache[mineHash]; exist {
		log.Trace(mgr.logInfo, "mine hash exist in pow task cache", mineHash.TerminalString())
		return true
	}

	if _, exist := mgr.aiTaskCache[mineHash]; exist {
		log.Trace(mgr.logInfo, "mine hash exist in ai task cache", mineHash.TerminalString())
		return true
	}

	if _, exist := mgr.unusedMineHeader[mineHash]; exist {
		log.Trace(mgr.logInfo, "mine hash exist in unusedMineHeader cache", mineHash.TerminalString())
		return true
	}

	return false
}

func (mgr *mineTaskManager) fixMap() {
	// 删除高度过低的task
	for hash, task := range mgr.powTaskCache {
		if task.powMiningNumber < mgr.curNumber {
			log.Trace(mgr.logInfo, "fix map", "delete pow task", "task number", task.powMiningNumber, "cur number", mgr.curNumber, "key hash", task.mineHash.TerminalString())
			delete(mgr.powTaskCache, hash)
		}
	}
	for hash, task := range mgr.aiTaskCache {
		if task.aiMiningNumber < mgr.curNumber {
			log.Trace(mgr.logInfo, "fix map", "delete ai task", "task number", task.aiMiningNumber, "cur number", mgr.curNumber, "key hash", task.mineHash.TerminalString())
			delete(mgr.aiTaskCache, hash)
		}
	}

	// 检查是否有可用header
	for hash, header := range mgr.unusedMineHeader {
		if mgr.bc.GetHeaderByHash(header.ParentHash) == nil {
			if header.Number.Uint64()+params.PowBlockPeriod+1 < mgr.curNumber {
				// 高度过低，删除
				delete(mgr.unusedMineHeader, hash)
			}
		} else {
			delete(mgr.unusedMineHeader, hash)
			powTask, aiTask, err := mgr.createMineTask(header, true)
			if err != nil {
				log.Trace(mgr.logInfo, "create mine task err", err)
				continue
			}

			mgr.addPowTask(powTask)
			mgr.addAITask(aiTask)
		}
	}
}
