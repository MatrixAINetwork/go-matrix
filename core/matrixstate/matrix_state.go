// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package matrixstate

import (
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/params/manversion"
)

const logInfo = "matrix state"

var mangerAlpha *Manager
var mangerBeta *Manager
var mangerGamma *Manager
var mangerDelta *Manager
var mangerAIMine *Manager
var versionOpt MatrixOperator

func init() {
	mangerAlpha = newManger(manversion.VersionAlpha)
	mangerBeta = newManger(manversion.VersionBeta)
	mangerGamma = newManger(manversion.VersionGamma)
	mangerDelta = newManger(manversion.VersionDelta)
	mangerAIMine = newManger(manversion.VersionAIMine)
	versionOpt = newVersionInfoOpt()
}

type MatrixOperator interface {
	KeyHash() common.Hash
	GetValue(st StateDB) (interface{}, error)
	SetValue(st StateDB, value interface{}) error
}

type Manager struct {
	version   string
	operators map[string]MatrixOperator
}

func GetManager(version string) *Manager {
	switch version {
	case manversion.VersionAlpha:
		return mangerAlpha
	case manversion.VersionBeta:
		return mangerBeta
	case manversion.VersionGamma:
		return mangerGamma
	case manversion.VersionDelta:
		return mangerDelta
	case manversion.VersionAIMine:
		return mangerAIMine
	default:
		log.Error(logInfo, "get Manger err", "version not exist", "version", version)
		return nil
	}
}

func (self *Manager) Version() string {
	return self.version
}

func (self *Manager) FindOperator(key string) (MatrixOperator, error) {
	opt, exist := self.operators[key]
	if !exist {
		log.Warn(logInfo, "find operator err", "not exist", "key", key, "version", self.version)
		return nil, ErrOptNotExist
	}
	return opt, nil
}

func newManger(version string) *Manager {
	switch version {
	case manversion.VersionAlpha:
		return &Manager{
			version: version,
			operators: map[string]MatrixOperator{
				mc.MSKeyBroadcastTx:            newBroadcastTxOpt(),
				mc.MSKeyTopologyGraph:          newTopologyGraphOpt(),
				mc.MSKeyElectGraph:             newELectGraphOpt(),
				mc.MSKeyElectOnlineState:       newELectOnlineStateOpt(),
				mc.MSKeyBroadcastInterval:      newBroadcastIntervalOpt(),
				mc.MSKeyElectGenTime:           newElectGenTimeOpt(),
				mc.MSKeyElectMinerNum:          newElectMinerNumOpt(),
				mc.MSKeyElectConfigInfo:        newElectConfigInfoOpt(),
				mc.MSKeyElectBlackList:         newElectBlackListOpt(),
				mc.MSKeyElectWhiteList:         newElectWhiteListOpt(),
				mc.MSKeyElectWhiteListSwitcher: newElectWhiteListSwitcherOpt(),
				mc.MSKeyAccountBroadcasts:      newBroadcastAccountsOpt(),
				mc.MSKeyAccountInnerMiners:     newInnerMinerAccountsOpt(),
				mc.MSKeyAccountFoundation:      newFoundationAccountOpt(),
				mc.MSKeyAccountVersionSupers:   newVersionSuperAccountsOpt(),
				mc.MSKeyAccountBlockSupers:     newBlockSuperAccountsOpt(),
				mc.MSKeyAccountMultiCoinSupers: newMultiCoinSuperAccountsOpt(),
				mc.MSKeyAccountSubChainSupers:  newSubChainSuperAccountsOpt(),
				mc.MSKeyVIPConfig:              newVIPConfigOpt(),
				mc.MSKeyPreBroadcastRoot:       newPreBroadcastRootOpt(),
				mc.MSKeyLeaderConfig:           newLeaderConfigOpt(),
				mc.MSKeyMinHash:                newMinHashOpt(),
				mc.MSKeySuperBlockCfg:          newSuperBlockCfgOpt(),

				mc.MSKeyBlkRewardCfg:      newBlkRewardCfgOpt(),
				mc.MSKeyTxsRewardCfg:      newTxsRewardCfgOpt(),
				mc.MSKeyInterestCfg:       newInterestCfgOpt(),
				mc.MSKeyLotteryCfg:        newLotteryCfgOpt(),
				mc.MSKeySlashCfg:          newSlashCfgOpt(),
				mc.MSKeyPreMinerBlkReward: newPreMinerBlkRewardOpt(),
				mc.MSKeyPreMinerTxsReward: newPreMinerTxsRewardOpt(),
				mc.MSKeyUpTimeNum:         newUpTimeNumOpt(),
				mc.MSKeyLotteryNum:        newLotteryNumOpt(),
				mc.MSKeyLotteryAccount:    newLotteryAccountOpt(),
				mc.MSKeyInterestCalcNum:   newInterestCalcNumOpt(),
				mc.MSKeyInterestPayNum:    newInterestPayNumOpt(),
				mc.MSKeySlashNum:          newSlashNumOpt(),

				mc.MSKeyBlkCalc:      newBlkCalcOpt(),
				mc.MSKeyTxsCalc:      newTxsCalcOpt(),
				mc.MSKeyInterestCalc: newInterestCalcOpt(),
				mc.MSKeyLotteryCalc:  newLotteryCalcOpt(),
				mc.MSKeySlashCalc:    newSlashCalcOpt(),

				mc.MSTxpoolGasLimitCfg: newTxpoolGasLimitOpt(),
				mc.MSCurrencyConfig:    newCurrencyPackOpt(),
				mc.MSAccountBlackList:  newAccountBlackListOpt(),

				mc.MSKeyBlockProduceStatsStatus: newBlockProduceStatsStatusOpt(),
				mc.MSKeyBlockProduceSlashCfg:    newBlockProduceSlashCfgOpt(),
				mc.MSKeyBlockProduceStats:       newBlockProduceStatsOpt(),
				mc.MSKeyBlockProduceBlackList:   newBlockProduceBlackListOpt(),
			},
		}
	case manversion.VersionBeta:
		return &Manager{
			version: version,
			operators: map[string]MatrixOperator{
				mc.MSKeyBroadcastTx:            newBroadcastTxOpt(),
				mc.MSKeyTopologyGraph:          newTopologyGraphOpt(),
				mc.MSKeyElectGraph:             newELectGraphOpt(),
				mc.MSKeyElectOnlineState:       newELectOnlineStateOpt(),
				mc.MSKeyBroadcastInterval:      newBroadcastIntervalOpt(),
				mc.MSKeyElectGenTime:           newElectGenTimeOpt(),
				mc.MSKeyElectMinerNum:          newElectMinerNumOpt(),
				mc.MSKeyElectConfigInfo:        newElectConfigInfoOpt(),
				mc.MSKeyElectBlackList:         newElectBlackListOpt(),
				mc.MSKeyElectWhiteList:         newElectWhiteListOpt(),
				mc.MSKeyElectWhiteListSwitcher: newElectWhiteListSwitcherOpt(),
				mc.MSKeyAccountBroadcasts:      newBroadcastAccountsOpt(),
				mc.MSKeyAccountInnerMiners:     newInnerMinerAccountsOpt(),
				mc.MSKeyAccountFoundation:      newFoundationAccountOpt(),
				mc.MSKeyAccountVersionSupers:   newVersionSuperAccountsOpt(),
				mc.MSKeyAccountBlockSupers:     newBlockSuperAccountsOpt(),
				mc.MSKeyAccountMultiCoinSupers: newMultiCoinSuperAccountsOpt(),
				mc.MSKeyAccountSubChainSupers:  newSubChainSuperAccountsOpt(),
				mc.MSKeyVIPConfig:              newVIPConfigOpt(),
				mc.MSKeyPreBroadcastRoot:       newPreBroadcastRootOpt(),
				mc.MSKeyLeaderConfig:           newLeaderConfigOpt(),
				mc.MSKeyMinHash:                newMinHashOpt(),
				mc.MSKeySuperBlockCfg:          newSuperBlockCfgOpt(),

				mc.MSKeyBlkRewardCfg:      newBlkRewardCfgOpt(),
				mc.MSKeyTxsRewardCfg:      newTxsRewardCfgOpt(),
				mc.MSKeyInterestCfg:       newInterestCfgOpt(),
				mc.MSKeyLotteryCfg:        newLotteryCfgOpt(),
				mc.MSKeySlashCfg:          newSlashCfgOpt(),
				mc.MSKeyPreMinerBlkReward: newPreMinerBlkRewardOpt(),
				mc.MSKeyPreMinerTxsReward: newPreMinerMultiCoinTxsRewardOpt(),
				mc.MSKeyUpTimeNum:         newUpTimeNumOpt(),
				mc.MSKeyLotteryNum:        newLotteryNumOpt(),
				mc.MSKeyLotteryAccount:    newLotteryAccountOpt(),
				mc.MSKeyInterestCalcNum:   newInterestCalcNumOpt(),
				mc.MSKeyInterestPayNum:    newInterestPayNumOpt(),
				mc.MSKeySlashNum:          newSlashNumOpt(),

				mc.MSKeyBlkCalc:      newBlkCalcOpt(),
				mc.MSKeyTxsCalc:      newTxsCalcOpt(),
				mc.MSKeyInterestCalc: newInterestCalcOpt(),
				mc.MSKeyLotteryCalc:  newLotteryCalcOpt(),
				mc.MSKeySlashCalc:    newSlashCalcOpt(),

				mc.MSTxpoolGasLimitCfg: newTxpoolGasLimitOpt(),
				mc.MSCurrencyConfig:    newCurrencyPackOpt(),
				mc.MSAccountBlackList:  newAccountBlackListOpt(),

				mc.MSKeyBlockProduceStatsStatus: newBlockProduceStatsStatusOpt(),
				mc.MSKeyBlockProduceSlashCfg:    newBlockProduceSlashCfgOpt(),
				mc.MSKeyBlockProduceStats:       newBlockProduceStatsOpt(),
				mc.MSKeyBlockProduceBlackList:   newBlockProduceBlackListOpt(),
			},
		}
	case manversion.VersionGamma:
		return &Manager{
			version: version,
			operators: map[string]MatrixOperator{
				mc.MSKeyBroadcastTx:            newBroadcastTxOpt(),
				mc.MSKeyTopologyGraph:          newTopologyGraphOpt(),
				mc.MSKeyElectGraph:             newELectGraphOpt(),
				mc.MSKeyElectOnlineState:       newELectOnlineStateOpt(),
				mc.MSKeyBroadcastInterval:      newBroadcastIntervalOpt(),
				mc.MSKeyElectGenTime:           newElectGenTimeOpt(),
				mc.MSKeyElectMinerNum:          newElectMinerNumOpt(),
				mc.MSKeyElectConfigInfo:        newElectConfigInfoOpt(),
				mc.MSKeyElectBlackList:         newElectBlackListOpt(),
				mc.MSKeyElectWhiteList:         newElectWhiteListOpt(),
				mc.MSKeyElectWhiteListSwitcher: newElectWhiteListSwitcherOpt(),
				mc.MSKeyAccountBroadcasts:      newBroadcastAccountsOpt(),
				mc.MSKeyAccountInnerMiners:     newInnerMinerAccountsOpt(),
				mc.MSKeyAccountFoundation:      newFoundationAccountOpt(),
				mc.MSKeyAccountVersionSupers:   newVersionSuperAccountsOpt(),
				mc.MSKeyAccountBlockSupers:     newBlockSuperAccountsOpt(),
				mc.MSKeyAccountMultiCoinSupers: newMultiCoinSuperAccountsOpt(),
				mc.MSKeyAccountSubChainSupers:  newSubChainSuperAccountsOpt(),
				mc.MSKeyVIPConfig:              newVIPConfigOpt(),
				mc.MSKeyPreBroadcastRoot:       newPreBroadcastRootOpt(),
				mc.MSKeyLeaderConfig:           newLeaderConfigOpt(),
				mc.MSKeyMinHash:                newMinHashOpt(),
				mc.MSKeySuperBlockCfg:          newSuperBlockCfgOpt(),

				mc.MSKeyBlkRewardCfg:      newBlkRewardCfgOpt(),
				mc.MSKeyTxsRewardCfg:      newTxsRewardCfgOpt(),
				mc.MSKeyInterestCfg:       newInterestCfgOpt(),
				mc.MSKeyLotteryCfg:        newLotteryCfgOpt(),
				mc.MSKeySlashCfg:          newSlashCfgOpt(),
				mc.MSKeyPreMinerBlkReward: newPreMinerBlkRewardOpt(),
				mc.MSKeyPreMinerTxsReward: newPreMinerMultiCoinTxsRewardOpt(),
				mc.MSKeyUpTimeNum:         newUpTimeNumOpt(),
				mc.MSKeyLotteryNum:        newLotteryNumOpt(),
				mc.MSKeyLotteryAccount:    newLotteryAccountOpt(),
				mc.MSKeyInterestCalcNum:   newInterestCalcNumOpt(),
				mc.MSKeyInterestPayNum:    newInterestPayNumOpt(),
				mc.MSKeySlashNum:          newSlashNumOpt(),

				mc.MSKeyBlkCalc:      newBlkCalcOpt(),
				mc.MSKeyTxsCalc:      newTxsCalcOpt(),
				mc.MSKeyInterestCalc: newInterestCalcOpt(),
				mc.MSKeyLotteryCalc:  newLotteryCalcOpt(),
				mc.MSKeySlashCalc:    newSlashCalcOpt(),

				mc.MSTxpoolGasLimitCfg: newTxpoolGasLimitOpt(),
				mc.MSCurrencyConfig:    newCurrencyPackOpt(),
				mc.MSAccountBlackList:  newAccountBlackListOpt(),

				mc.MSKeyBlockProduceStatsStatus: newBlockProduceStatsStatusOpt(),
				mc.MSKeyBlockProduceSlashCfg:    newBlockProduceSlashCfgOpt(),
				mc.MSKeyBlockProduceStats:       newBlockProduceStatsOpt(),
				mc.MSKeyBlockProduceBlackList:   newBlockProduceBlackListOpt(),
				mc.MSKeySelMinerNum:             newSelMinerNumOpt(),
			},
		}
	case manversion.VersionDelta, manversion.VersionAIMine:
		return &Manager{
			version: version,
			operators: map[string]MatrixOperator{
				mc.MSKeyBroadcastTx:            newBroadcastTxOpt(),
				mc.MSKeyTopologyGraph:          newTopologyGraphOpt(),
				mc.MSKeyElectGraph:             newELectGraphOpt(),
				mc.MSKeyElectOnlineState:       newELectOnlineStateOpt(),
				mc.MSKeyBroadcastInterval:      newBroadcastIntervalOpt(),
				mc.MSKeyElectGenTime:           newElectGenTimeOpt(),
				mc.MSKeyElectMinerNum:          newElectMinerNumOpt(),
				mc.MSKeyElectConfigInfo:        newElectConfigInfoOpt(),
				mc.MSKeyElectBlackList:         newElectBlackListOpt(),
				mc.MSKeyElectWhiteList:         newElectWhiteListOpt(),
				mc.MSKeyElectWhiteListSwitcher: newElectWhiteListSwitcherOpt(),
				mc.MSKeyAccountBroadcasts:      newBroadcastAccountsOpt(),
				mc.MSKeyAccountInnerMiners:     newInnerMinerAccountsOpt(),
				mc.MSKeyAccountFoundation:      newFoundationAccountOpt(),
				mc.MSKeyAccountVersionSupers:   newVersionSuperAccountsOpt(),
				mc.MSKeyAccountBlockSupers:     newBlockSuperAccountsOpt(),
				mc.MSKeyAccountMultiCoinSupers: newMultiCoinSuperAccountsOpt(),
				mc.MSKeyAccountSubChainSupers:  newSubChainSuperAccountsOpt(),
				mc.MSKeyVIPConfig:              newVIPConfigOpt(),
				mc.MSKeyPreBroadcastRoot:       newPreBroadcastRootOpt(),
				mc.MSKeyLeaderConfig:           newLeaderConfigOpt(),
				mc.MSKeyMinHash:                newMinHashOpt(),
				mc.MSKeySuperBlockCfg:          newSuperBlockCfgOpt(),
				mc.MSKeyMinimumDifficulty:      newMinDiffcultyOpt(),
				mc.MSKeyMaximumDifficulty:      newMaxDiffcultyOpt(),
				mc.MSKeyReelectionDifficulty:   newReelectionDiffcultyOpt(),
				mc.MSKeyBlockDurationStatus:    newBlockDurationOpt(),

				mc.MSKeyBlkRewardCfg:       newBlkRewardCfgOpt(),
				mc.MSKeyAIBlkRewardCfg:     newAIBlkRewardCfgOpt(),
				mc.MSKeyTxsRewardCfg:       newTxsRewardCfgOpt(),
				mc.MSKeyInterestCfg:        newInterestCfgOpt(),
				mc.MSKeyLotteryCfg:         newLotteryCfgOpt(),
				mc.MSKeySlashCfg:           newSlashCfgOpt(),
				mc.MSKeyPreMinerBlkReward:  newPreMinerBlkRewardOpt(),
				mc.MSKeyPreMinerTxsReward:  newPreMinerMultiCoinTxsRewardOpt(),
				mc.MSKeyUpTimeNum:          newUpTimeNumOpt(),
				mc.MSKeyLotteryNum:         newLotteryNumOpt(),
				mc.MSKeyLotteryAccount:     newLotteryAccountOpt(),
				mc.MSKeyInterestCalcNum:    newInterestCalcNumOpt(),
				mc.MSKeyInterestPayNum:     newInterestPayNumOpt(),
				mc.MSKeySlashNum:           newSlashNumOpt(),
				mc.MSKeySelMinerNum:        newSelMinerNumOpt(),
				mc.MSKeyBLKSelValidatorNum: newSelValidatorBLKNumOpt(),
				mc.MSKeyBLKSelValidator:    newValidatorBLKSelRewardOpt(),
				mc.MSKeyTXSSelValidatorNum: newSelValidatorTXSNumOpt(),
				mc.MSKeyTXSSelValidator:    newValidatorTXSSelRewardOpt(),

				mc.MSKeyBlkCalc:      newBlkCalcOpt(),
				mc.MSKeyTxsCalc:      newTxsCalcOpt(),
				mc.MSKeyInterestCalc: newInterestCalcOpt(),
				mc.MSKeyLotteryCalc:  newLotteryCalcOpt(),
				mc.MSKeySlashCalc:    newSlashCalcOpt(),

				mc.MSTxpoolGasLimitCfg: newTxpoolGasLimitOpt(),
				mc.MSCurrencyConfig:    newCurrencyPackOpt(),
				mc.MSAccountBlackList:  newAccountBlackListOpt(),

				mc.MSKeyBlockProduceStatsStatus: newBlockProduceStatsStatusOpt(),
				mc.MSKeyBlockProduceSlashCfg:    newBlockProduceSlashCfgOpt(),
				mc.MSKeyBlockProduceStats:       newBlockProduceStatsOpt(),
				mc.MSKeyBlockProduceBlackList:   newBlockProduceBlackListOpt(),
				mc.MSKeyBasePowerStatsStatus:    newBasePowerStatsStatusOpt(),
				mc.MSKeyBasePowerSlashCfg:       newBasePowerSlashCfgOpt(),
				mc.MSKeyBasePowerStats:          newBasePowerStatsOpt(),
				mc.MSKeyBasePowerBlackList:      newBasePowerBlackListOpt(),
				mc.MSKeyElectDynamicPollingInfo: newDynamicPollingOpt(),
				mc.MSCurrencyHeader:             newCurrencyHeaderCfgOpt(),
			},
		}
	default:
		log.Error(logInfo, "创建管理类", "失败", "版本", version)
		return nil
	}
}
