// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package manparams

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/pkg/errors"
)

const ReelectionTimes = uint64(3) // 选举周期的倍数(广播周期*ReelectionTimes = 选举周期)

type BCInterval struct {
	lastBCNumber       uint64 // 最后的广播区块高度
	lastReelectNumber  uint64 // 最后的选举区块高度
	bcInterval         uint64 // 广播周期
	backupEnableNumber uint64 // 预备广播周期启用高度
	backupBCInterval   uint64 // 预备广播周期
}

func NewBCInterval() *BCInterval {
	interval := GetBCIntervalInfo()
	log.INFO("params", "NewBCInterval广播周期", interval.BCInterval, "上个广播高度", interval.LastBCNumber)
	return &BCInterval{
		lastBCNumber:       interval.LastBCNumber,
		lastReelectNumber:  interval.LastReelectNumber,
		bcInterval:         interval.BCInterval,
		backupEnableNumber: interval.BackupEnableNumber,
		backupBCInterval:   interval.BackupBCInterval,
	}
}

func NewBCIntervalWithInterval(interval interface{}) (*BCInterval, error) {
	if interval == nil {
		return nil, errors.New("param is nil")
	}
	infoStu, err := bcIntervalInfo2Stu(interval)
	if err != nil {
		return nil, errors.Errorf("transfer broadcast interval data to struct err(%v)", err)
	}

	//log.INFO("params", "NewBCIntervalWithInterval广播周期", infoStu.BCInterval, "上个广播高度", infoStu.LastBCNumber)
	return &BCInterval{
		lastBCNumber:       infoStu.LastBCNumber,
		lastReelectNumber:  infoStu.LastReelectNumber,
		bcInterval:         infoStu.BCInterval,
		backupEnableNumber: infoStu.BackupEnableNumber,
		backupBCInterval:   infoStu.BackupBCInterval,
	}, nil
}

func NewBCIntervalByNumber(blockNumber uint64) (*BCInterval, error) {
	interval, err := GetBCIntervalInfoByNumber(blockNumber)
	if err != nil {
		return nil, err
	}

	log.INFO("params", "NewBCIntervalByNumber广播周期", interval.BCInterval, "上个广播高度", interval.LastBCNumber)
	return &BCInterval{
		lastBCNumber:       interval.LastBCNumber,
		lastReelectNumber:  interval.LastReelectNumber,
		bcInterval:         interval.BCInterval,
		backupEnableNumber: interval.BackupEnableNumber,
		backupBCInterval:   interval.BackupBCInterval,
	}, nil
}

func NewBCIntervalByHash(blockHash common.Hash) (*BCInterval, error) {
	interval, err := GetBCIntervalInfoByHash(blockHash)
	if err != nil {
		return nil, err
	}

	log.INFO("params", "NewBCIntervalByHash广播周期", interval.BCInterval, "上个广播高度", interval.LastBCNumber)
	return &BCInterval{
		lastBCNumber:       interval.LastBCNumber,
		lastReelectNumber:  interval.LastReelectNumber,
		bcInterval:         interval.BCInterval,
		backupEnableNumber: interval.BackupEnableNumber,
		backupBCInterval:   interval.BackupBCInterval,
	}, nil
}

func (period *BCInterval) GetBroadcastInterval() uint64 {
	return period.bcInterval
}

func (period *BCInterval) GetReElectionInterval() uint64 {
	return period.bcInterval * ReelectionTimes
}

func (period *BCInterval) IsBroadcastNumber(number uint64) bool {
	if number < period.lastBCNumber {
		log.ERROR("广播周期", "IsBroadcastNumber", "false", "number", number, "period.lastBCNumber", period.lastBCNumber)
		return false
	}
	return (number-period.lastBCNumber)%period.bcInterval == 0
}

func (period *BCInterval) IsReElectionNumber(number uint64) bool {
	if number < period.lastReelectNumber {
		log.ERROR("广播周期", "IsReElectionNumber", "false", "number", number, "period.lastReelectNumber", period.lastReelectNumber)
		return false
	}

	if (number-period.lastReelectNumber)%period.bcInterval != 0 {
		// 高度不是广播区块，一定不是选举区块
		return false
	}

	bcCount := (number - period.lastReelectNumber) / period.bcInterval
	return bcCount%ReelectionTimes == 0
}

func (period *BCInterval) GetLastBroadcastNumber() uint64 {
	return period.lastBCNumber
}

func (period *BCInterval) GetLastReElectionNumber() uint64 {
	return period.lastReelectNumber
}

func (period *BCInterval) GetNextBroadcastNumber(number uint64) uint64 {
	if number <= period.lastBCNumber {
		return period.lastBCNumber
	}

	return ((number-period.lastBCNumber)/period.bcInterval+1)*period.bcInterval + period.lastBCNumber
}

func (period *BCInterval) GetNextReElectionNumber(number uint64) uint64 {
	if number <= period.lastReelectNumber {
		return period.lastReelectNumber
	}

	bcCountAfterReelect := (number - period.lastReelectNumber) / period.bcInterval
	NextReelectCount := bcCountAfterReelect/ReelectionTimes + 1
	return (NextReelectCount*ReelectionTimes)*period.bcInterval + period.lastReelectNumber
}

func (period *BCInterval) GetBackupEnableNumber() uint64 {
	return period.backupEnableNumber
}

func (period *BCInterval) UsingBackupInterval() {
	period.bcInterval = period.backupBCInterval
	period.backupEnableNumber = 0
	period.backupBCInterval = 0
}

func (period *BCInterval) SetLastBCNumber(number uint64) {
	period.lastBCNumber = number
}

func (period *BCInterval) SetLastReelectNumber(number uint64) {
	period.lastReelectNumber = number
}

func (period *BCInterval) SetBackupBCInterval(interval uint64, enableNumber uint64) {
	period.backupEnableNumber = enableNumber
	period.backupBCInterval = interval
}

func (period *BCInterval) ToInfoStu() *mc.BCIntervalInfo {
	return &mc.BCIntervalInfo{
		LastBCNumber:       period.lastBCNumber,
		LastReelectNumber:  period.lastReelectNumber,
		BCInterval:         period.bcInterval,
		BackupEnableNumber: period.backupEnableNumber,
		BackupBCInterval:   period.backupBCInterval,
	}
}

func IsBroadcastNumber(number uint64, stateNumber uint64) bool {
	interval, err := NewBCIntervalByNumber(stateNumber)
	if err != nil {
		log.Error("config", "获取广播区块周期失败", err, "stateNumber", stateNumber)
		return false
	}
	return interval.IsBroadcastNumber(number)
}

func IsBroadcastNumberByHash(number uint64, blockHash common.Hash) bool {
	interval, err := NewBCIntervalByHash(blockHash)
	if err != nil {
		log.Error("config", "获取广播区块周期失败", err, "block hash", blockHash.Hex())
		return false
	}
	return interval.IsBroadcastNumber(number)
}

func IsReElectionNumber(number uint64, stateNumber uint64) bool {
	interval, err := NewBCIntervalByNumber(stateNumber)
	if err != nil {
		log.Error("config", "获取广播区块周期失败", err, "stateNumber", stateNumber)
		return false
	}
	return interval.IsReElectionNumber(number)
}

func GetBCIntervalInfo() *mc.BCIntervalInfo {
	data, err := mtxCfg.getStateData(mc.MSKeyBroadcastInterval)
	if err != nil {
		log.Crit("config", "get broadcast interval from state err", err)
	}
	info, err := bcIntervalInfo2Stu(data)
	if err != nil {
		log.Crit("config", "transfer broadcast interval struct err", err)
	}
	return info
}

func GetBCIntervalInfoByNumber(number uint64) (*mc.BCIntervalInfo, error) {
	data, err := mtxCfg.getStateDataByNumber(mc.MSKeyBroadcastInterval, number)
	if err != nil {
		return nil, err
	}
	return bcIntervalInfo2Stu(data)
}

func GetBCIntervalInfoByHash(hash common.Hash) (*mc.BCIntervalInfo, error) {
	data, err := mtxCfg.getStateDataByHash(mc.MSKeyBroadcastInterval, hash)
	if err != nil {
		return nil, err
	}
	return bcIntervalInfo2Stu(data)
}

func bcIntervalInfo2Stu(data interface{}) (*mc.BCIntervalInfo, error) {
	info, OK := data.(*mc.BCIntervalInfo)
	if OK == false {
		return nil, errors.New("BCIntervalInfo反射失败")
	}
	if nil == info {
		return nil, errors.New("BCIntervalInfo is nil")
	}
	return info, nil
}
