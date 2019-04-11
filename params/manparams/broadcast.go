// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package manparams

import (
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
)

func IsBroadcastNumber(number uint64, stateNumber uint64) bool {
	interval, err := GetBCIntervalInfoByNumber(stateNumber)
	if err != nil {
		log.Error("config", "获取广播区块周期失败", err, "stateNumber", stateNumber)
		return false
	}
	return interval.IsBroadcastNumber(number)
}

func IsBroadcastNumberByHash(number uint64, blockHash common.Hash) bool {
	interval, err := GetBCIntervalInfoByHash(blockHash)
	if err != nil {
		log.Error("config", "获取广播区块周期失败", err, "block hash", blockHash.Hex())
		return false
	}
	return interval.IsBroadcastNumber(number)
}

func IsReElectionNumber(number uint64, stateNumber uint64) bool {
	interval, err := GetBCIntervalInfoByNumber(stateNumber)
	if err != nil {
		log.Error("config", "获取广播区块周期失败", err, "stateNumber", stateNumber)
		return false
	}
	return interval.IsReElectionNumber(number)
}

func GetBCIntervalInfo() *mc.BCIntervalInfo {
	info, err := broadcastCfg.reader.GetBroadcastInterval()
	if err != nil || info == nil {
		log.Crit("config", "get broadcast interval from state err", err)
	}
	return info
}

func GetBCIntervalInfoByNumber(number uint64) (*mc.BCIntervalInfo, error) {
	return broadcastCfg.reader.GetBroadcastIntervalByNumber(number)
}

func GetBCIntervalInfoByHash(hash common.Hash) (*mc.BCIntervalInfo, error) {
	return broadcastCfg.reader.GetBroadcastIntervalByHash(hash)
}
