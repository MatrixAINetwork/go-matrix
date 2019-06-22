// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package mc

import "github.com/MatrixAINetwork/go-matrix/log"

const ReelectionTimes = uint64(3) // 选举周期的倍数(广播周期*ReelectionTimes = 选举周期)

func (info *BCIntervalInfo) GetBroadcastInterval() uint64 {
	return info.BCInterval
}

func (info *BCIntervalInfo) GetReElectionInterval() uint64 {
	return info.BCInterval * ReelectionTimes
}

func (info *BCIntervalInfo) IsBroadcastNumber(number uint64) bool {
	if number < info.LastBCNumber {
		log.ERROR("广播周期", "IsBroadcastNumber", "false", "number", number, "period.lastBCNumber", info.LastBCNumber)
		return false
	}
	return (number-info.LastBCNumber)%info.BCInterval == 0
}

func (info *BCIntervalInfo) IsReElectionNumber(number uint64) bool {
	if number < info.LastReelectNumber {
		log.ERROR("广播周期", "IsReElectionNumber", "false", "number", number, "period.lastReelectNumber", info.LastReelectNumber)
		return false
	}

	if (number-info.LastReelectNumber)%info.BCInterval != 0 {
		// 高度不是广播区块，一定不是选举区块
		return false
	}

	bcCount := (number - info.LastReelectNumber) / info.BCInterval
	return bcCount%ReelectionTimes == 0
}

func (info *BCIntervalInfo) GetLastBroadcastNumber() uint64 {
	return info.LastBCNumber
}

func (info *BCIntervalInfo) GetLastReElectionNumber() uint64 {
	return info.LastReelectNumber
}

func (info *BCIntervalInfo) GetNextBroadcastNumber(number uint64) uint64 {
	if number <= info.LastBCNumber {
		return info.LastBCNumber
	}

	return ((number-info.LastBCNumber)/info.BCInterval+1)*info.BCInterval + info.LastBCNumber
}

func (info *BCIntervalInfo) GetNextReElectionNumber(number uint64) uint64 {
	if number <= info.LastReelectNumber {
		return info.LastReelectNumber
	}

	bcCountAfterReelect := (number - info.LastReelectNumber) / info.BCInterval
	NextReelectCount := bcCountAfterReelect/ReelectionTimes + 1
	return (NextReelectCount*ReelectionTimes)*info.BCInterval + info.LastReelectNumber
}

func (info *BCIntervalInfo) GetBackupEnableNumber() uint64 {
	return info.BackupEnableNumber
}

func (info *BCIntervalInfo) UsingBackupInterval() {
	info.BCInterval = info.BackupBCInterval
	info.BackupEnableNumber = 0
	info.BackupBCInterval = 0
}

func (info *BCIntervalInfo) SetLastBCNumber(number uint64) {
	info.LastBCNumber = number
}

func (info *BCIntervalInfo) SetLastReelectNumber(number uint64) {
	info.LastReelectNumber = number
}

func (info *BCIntervalInfo) SetBackupBCInterval(interval uint64, enableNumber uint64) {
	info.BackupEnableNumber = enableNumber
	info.BackupBCInterval = interval
}
