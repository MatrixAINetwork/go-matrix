// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package core

import (
	"fmt"
	"math/big"

	"github.com/MatrixAINetwork/go-matrix/consensus/amhash"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/matrixstate"
	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/depoistInfo"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/params"
	"github.com/MatrixAINetwork/go-matrix/params/manparams"
	"github.com/MatrixAINetwork/go-matrix/params/manversion"
	"github.com/MatrixAINetwork/go-matrix/reward/util"
	"github.com/pkg/errors"
)

// State returns a new mutable state based on the current HEAD block.
func (bc *BlockChain) State() (*state.StateDBManage, error) {
	return bc.StateAt(bc.CurrentBlock().Root())
}

// StateAt returns a new mutable state based on a particular point in time.
func (bc *BlockChain) StateAt(root []common.CoinRoot) (*state.StateDBManage, error) {
	return state.NewStateDBManage(root, bc.db, bc.stateCache)
	//return bc.getStateCache(root)
}
func (bc *BlockChain) getStateCache(root []common.CoinRoot) (*state.StateDBManage, error) {
	hash := types.RlpHash(root)
	if stCache, exist := bc.depCache.Get(hash); exist {
		return stCache.(*state.StateDBManage), nil
	}
	st, err := state.NewStateDBManage(root, bc.db, bc.stateCache)
	if err == nil {
		bc.depCache.Add(hash, st)
	}
	return st, err
}
func (bc *BlockChain) StateAtNumber(number uint64) (*state.StateDBManage, error) {
	block := bc.GetBlockByNumber(number)
	if block == nil {
		return nil, errors.Errorf("can't find block by number(%d)", number)
	}
	return bc.StateAt(block.Root())
}

func (bc *BlockChain) StateAtBlockHash(hash common.Hash) (*state.StateDBManage, error) {
	block := bc.GetBlockByHash(hash)
	if block == nil {
		return nil, errors.New("can't find block by hash")
	}
	return bc.StateAt(block.Root())
}

func (bc *BlockChain) RegisterMatrixStateDataProducer(key string, producer ProduceMatrixStateDataFn) {
	bc.matrixProcessor.RegisterProducer(key, producer)
}

func (bc *BlockChain) ProcessStateVersion(version []byte, st *state.StateDBManage) error {
	return bc.matrixProcessor.ProcessStateVersion(version, st)
}

func (bd *BlockChain) processStateSwitchGamma(stateDB *state.StateDBManage) error {
	electCfg, err := matrixstate.GetElectConfigInfo(stateDB)
	if nil != err {
		log.Crit("blockChain", "选举配置错误", err)
		return err
	}
	err = matrixstate.SetElectConfigInfo(stateDB, &mc.ElectConfigInfo{ValidatorNum: electCfg.ValidatorNum, BackValidator: electCfg.BackValidator, ElectPlug: manparams.ElectPlug_layerdBSS})
	if nil != err {
		log.Crit("blockChain", "选举引擎切换,错误", err)
		return err
	}
	leaderCfg, err := matrixstate.GetLeaderConfig(stateDB)
	if nil != err {
		log.Crit("blockChain", "leader配置错误", err)
		return err
	}
	err = matrixstate.SetLeaderConfig(stateDB, &mc.LeaderConfig{ParentMiningTime: 20, PosOutTime: 40, ReelectOutTime: 40, ReelectHandleInterval: leaderCfg.ReelectHandleInterval})
	if nil != err {
		log.Crit("blockChain", "出块超时和投票超时改为60秒和40秒,错误", err)
		return err
	}
	blkSlash, err := matrixstate.GetBlockProduceSlashCfg(stateDB)
	if nil != err {
		log.Crit("blockChain", "读取惩罚配置错误", err)
		return err
	}
	err = matrixstate.SetBlockProduceSlashCfg(stateDB, &mc.BlockProduceSlashCfg{Switcher: blkSlash.Switcher, LowTHR: blkSlash.LowTHR, ProhibitCycleNum: 10})
	if nil != err {
		log.Crit("blockChain", "惩罚配置错误", err)
		return err
	}
	blkRewardCfg, err := matrixstate.GetBlkRewardCfg(stateDB)
	if nil != err {
		log.Crit("blockChain", "选举配置错误", err)
		return err
	}
	//每个块的奖励改为15MAN
	err = matrixstate.SetBlkRewardCfg(stateDB, &mc.BlkRewardCfg{
		MinerMount:               4800, //放大1000倍
		MinerAttenuationRate:     blkRewardCfg.MinerAttenuationRate,
		MinerAttenuationNum:      3000000,
		ValidatorMount:           8000, //放大1000倍
		ValidatorAttenuationRate: blkRewardCfg.ValidatorAttenuationRate,
		ValidatorAttenuationNum:  3000000,
		RewardRate: mc.RewardRateCfg{
			MinerOutRate:        4000, //出块矿工奖励
			ElectedMinerRate:    5000, //当选矿工奖励
			FoundationMinerRate: 1000, //基金会网络奖励

			LeaderRate:              2500, //出块验证者（leader）奖励
			ElectedValidatorsRate:   6500, //当选验证者奖励
			FoundationValidatorRate: 1000, //基金会网络奖励

			OriginElectOfflineRate: blkRewardCfg.RewardRate.OriginElectOfflineRate, //初选下线验证者奖励
			BackupRewardRate:       blkRewardCfg.RewardRate.BackupRewardRate,       //当前替补验证者奖励
		},
	})
	if nil != err {
		log.Crit("blockChain", "固定区块奖励修改为原来的1.5倍", err)
		return err
	}
	interestCfg, err := matrixstate.GetInterestCfg(stateDB)
	if nil != err {
		log.Crit("blockChain", "选举配置错误", err)
		return err
	}
	err = matrixstate.SetInterestCfg(stateDB, &mc.InterestCfg{
		RewardMount:       3200, //放大1000倍
		AttenuationRate:   interestCfg.AttenuationRate,
		AttenuationPeriod: 3000000,
		PayInterval:       interestCfg.PayInterval,
	})
	if nil != err {
		log.Crit("blockChain", "利息奖励修改为原来的1.5倍", err)
		return err
	}
	err = matrixstate.SetBlkCalc(stateDB, util.CalcGamma)
	if nil != err {
		log.Crit("blockChain", "固定区块奖励引擎设置错误", err)
		return err
	}
	err = matrixstate.SetInterestCalc(stateDB, util.CalcGamma)
	if nil != err {
		log.Crit("blockChain", "利息奖励引擎设置错误", err)
		return err
	}
	return nil
}

func (bd *BlockChain) processStateSwitchDelta(stateDB *state.StateDBManage, t uint64) error {
	err := matrixstate.SetInterestCalc(stateDB, util.CalcDelta)
	if nil != err {
		log.Crit("blockChain", "利息奖励引擎设置错误", err)
		return err
	}
	err = matrixstate.SetSlashCalc(stateDB, util.CalcDelta)
	if nil != err {
		log.Crit("blockChain", "惩罚奖励引擎设置错误", err)
		return err
	}
	err = depoistInfo.SetVersion(stateDB, t)
	if nil != err {
		log.Crit("blockChain", "合约设置新版本错误", err)
		return err
	}
	return nil
}

func (bd *BlockChain) processStateSwitchAIMine(stateDB *state.StateDBManage) error {
	err := matrixstate.SetMinDifficulty(stateDB, params.AIManMinimumDifficulty)
	if nil != err {
		log.Crit("blockChain", "设置最小挖矿难度失败", err)
		return err
	}
	err = matrixstate.SetMaxDifficulty(stateDB, params.AIManMaxDifficulty)
	if nil != err {
		log.Crit("blockChain", "设置最大挖矿难度失败", err)
		return err
	}
	err = matrixstate.SetReelectionDifficulty(stateDB, params.AIManReelectionDifficulty)
	if nil != err {
		log.Crit("blockChain", "设置换届初始难度失败", err)
		return err
	}
	err = matrixstate.SetBasePowerSlashCfg(stateDB, &mc.BasePowerSlashCfg{Switcher: true, LowTHR: 2, ProhibitCycleNum: 1})
	if nil != err {
		log.Crit("blockChain", "算力黑名单惩罚配置错误", err)
		return err
	}
	electCfg, err := matrixstate.GetElectConfigInfo(stateDB)
	if nil != err {
		log.Crit("blockChain", "选举配置错误", err)
		return err
	}
	err = matrixstate.SetElectConfigInfo(stateDB, &mc.ElectConfigInfo{ValidatorNum: electCfg.ValidatorNum, BackValidator: electCfg.BackValidator, ElectPlug: manparams.ElectPlug_layerdDP})
	if nil != err {
		log.Crit("blockChain", "选举引擎切换,错误", err)
		return err
	}
	err = matrixstate.SetInterestCalcNum(stateDB, manversion.VersionNumAIMine)
	if nil != err {
		log.Crit("blockChain", "利息计算高度设置错误", err)
		return err
	}
	if err = matrixstate.SetSelMinerNum(stateDB, manversion.VersionNumAIMine); err != nil {
		log.Crit("blockChain", "设置参与矿工奖励状态错误", err)
		return err
	}
	if err = matrixstate.SetBLKSelValidatorNum(stateDB, manversion.VersionNumAIMine); err != nil {
		log.Crit("blockChain", "设置参与验证者奖励状态错误", err)
		return err
	}
	if err = matrixstate.SetTXSSelValidatorNum(stateDB, manversion.VersionNumAIMine); err != nil {
		log.Crit("blockChain", "设置交易费参与验证者奖励状态错误", err)
		return err
	}
	err = matrixstate.SetInterestCalc(stateDB, util.CalcEpsilon)
	if nil != err {
		log.Crit("blockChain", "利息奖励引擎设置错误", err)
		return err
	}

	err = matrixstate.SetBlkCalc(stateDB, util.CalcEpsilon)
	if nil != err {
		log.Crit("blockChain", "区块奖励引擎设置错误", err)
		return err
	}

	err = matrixstate.SetTxsCalc(stateDB, util.CalcEpsilon)
	if nil != err {
		log.Crit("blockChain", "交易奖励引擎设置错误", err)
		return err
	}

	leaderCfg, err := matrixstate.GetLeaderConfig(stateDB)
	if nil != err {
		log.Crit("blockChain", "读取leader配置错误", err)
		return err
	}
	err = matrixstate.SetLeaderConfig(stateDB, &mc.LeaderConfig{
		ParentMiningTime:      params.AIMineLeaderSrcParentMiningTime,
		PosOutTime:            params.AIMineLeaderSrcPosOutTime,
		ReelectOutTime:        params.AIMineLeaderSrcReelectOutTime,
		ReelectHandleInterval: leaderCfg.ReelectHandleInterval})
	if nil != err {
		log.Crit("blockChain", "出块超时和投票超时改为60秒和40秒,错误", err)
		return err
	}
	err = matrixstate.SetElectDynamicPollingInfo(stateDB, &mc.ElectDynamicPollingInfo{Number: 0, Seq: 0, MinerNum: 0, CandidateList: nil})
	if nil != err {
		log.Crit("blockChain", "初始化轮询选举信息错误", err)
		return err
	}
	blkRewardCfg, err := matrixstate.GetBlkRewardCfg(stateDB)
	if nil != err {
		log.Crit("blockChain", "读取区块奖励配置错误", err)
		return err
	}

	err = matrixstate.SetAIBlkRewardCfg(stateDB, &mc.AIBlkRewardCfg{
		MinerMount:               blkRewardCfg.MinerMount,
		MinerAttenuationRate:     blkRewardCfg.MinerAttenuationRate,
		MinerAttenuationNum:      blkRewardCfg.MinerAttenuationNum,
		ValidatorMount:           blkRewardCfg.ValidatorMount,
		ValidatorAttenuationRate: blkRewardCfg.ValidatorAttenuationRate,
		ValidatorAttenuationNum:  blkRewardCfg.ValidatorAttenuationNum,
		RewardRate: mc.AIRewardRateCfg{
			MinerOutRate:        4000, //出块矿工奖励
			AIMinerOutRate:      1000, //AI矿工奖励
			ElectedMinerRate:    4000, //当选矿工奖励
			FoundationMinerRate: 1000, //基金会网络奖励

			LeaderRate:              blkRewardCfg.RewardRate.LeaderRate,              //出块验证者（leader）奖励
			ElectedValidatorsRate:   blkRewardCfg.RewardRate.ElectedValidatorsRate,   //当选验证者奖励
			FoundationValidatorRate: blkRewardCfg.RewardRate.FoundationValidatorRate, //基金会网络奖励

			OriginElectOfflineRate: blkRewardCfg.RewardRate.OriginElectOfflineRate, //初选下线验证者奖励
			BackupRewardRate:       blkRewardCfg.RewardRate.BackupRewardRate,       //当前替补验证者奖励
		},
	})
	if nil != err {
		log.Crit("blockChain", "设置区块奖励配置错误", err)
		return err
	}
	if err = matrixstate.SetBlockProduceBlackList(stateDB, &mc.BlockProduceSlashBlackList{[]mc.UserBlockProduceSlash{}}); err != nil {
		log.Crit("blockChain", "清除区块生成黑名单错误", err)
		return err
	}
	if err = matrixstate.SetBlockDuration(stateDB, &mc.BlockDurationStatus{[]uint8{0}}); err != nil {
		log.Crit("blockChain", "设置出块状态错误", err)
		return err
	}
	return nil
}
func (bd *BlockChain) processStateSwitchZeta(stateDB *state.StateDBManage) error {
	err := matrixstate.SetMinDifficulty(stateDB, params.ZetaMinimumDifficulty)
	if nil != err {
		log.Crit("blockChain", "设置最小挖矿难度失败", err)
		return err
	}
	err = matrixstate.SetMaxDifficulty(stateDB, params.ZetaMaxDifficulty)
	if nil != err {
		log.Crit("blockChain", "设置最大挖矿难度失败", err)
		return err
	}
	err = matrixstate.SetReelectionDifficulty(stateDB, params.ZetaReelectionDifficulty)
	if nil != err {
		log.Crit("blockChain", "设置换届初始难度失败", err)
		return err
	}

	electCfg, err := matrixstate.GetElectConfigInfo(stateDB)
	if nil != err {
		log.Crit("blockChain", "选举配置错误", err)
		return err
	}
	err = matrixstate.SetElectConfigInfo(stateDB, &mc.ElectConfigInfo{ValidatorNum: electCfg.ValidatorNum, BackValidator: electCfg.BackValidator, ElectPlug: manparams.ElectPlug_layerdDPV2})
	if nil != err {
		log.Crit("blockChain", "选举引擎切换,错误", err)
		return err
	}
	if err := matrixstate.SetBlockProduceBlackList(stateDB, &mc.BlockProduceSlashBlackList{[]mc.UserBlockProduceSlash{}}); err != nil {
		log.Crit("blockChain", "清除区块生成黑名单错误", err)
		return err
	}
	if err := matrixstate.SetBasePowerBlackList(stateDB, &mc.BasePowerSlashBlackList{[]mc.BasePowerSlash{}}); err != nil {
		log.Crit("blockChain", "清除算力检测黑名单错误", err)
		return err
	}
	err = matrixstate.SetElectDynamicPollingInfo(stateDB, &mc.ElectDynamicPollingInfo{Number: 0, Seq: 0, MinerNum: 0, CandidateList: nil})
	if nil != err {
		log.Crit("blockChain", "初始化轮询选举信息错误", err)
		return err
	}
	return nil
}
func (bc *BlockChain) ProcessStateVersionSwitch(num uint64, t uint64, version []byte, stateDB *state.StateDBManage) error {
	//提前一个块设置各自算法引擎和配置，切换高度生效
	switch num {
	case manversion.VersionNumGamma - 1:
		if manversion.VersionCmp(string(version), manversion.VersionGamma) >= 0 {
			log.Info("blockchain", "切换版本Gamma高度", num, "当前版本大于等于Gamma版本, 不设置state", string(version))
			return nil
		}
		log.Info("blockchain", "切换版本Gamma高度", num)
		return bc.processStateSwitchGamma(stateDB)

	case manversion.VersionNumDelta - 1:
		if manversion.VersionCmp(string(version), manversion.VersionDelta) >= 0 {
			log.Info("blockchain", "切换版本Delta高度", num, "当前版本大于等于Delta版本, 不设置state", string(version))
			return nil
		}
		log.Info("blockchain", "切换版本Delta 高度", num)
		return bc.processStateSwitchDelta(stateDB, t)

	case manversion.VersionNumAIMine - 1:
		if manversion.VersionCmp(string(version), manversion.VersionAIMine) >= 0 {
			log.Info("blockchain", "切换版本AI Mine高度", num, "当前版本大于等于AI Mine版本, 不设置state", string(version))
			return nil
		}
		log.Info("blockchain", "切换版本AI Mine高度", num)
		return bc.processStateSwitchAIMine(stateDB)

	case manversion.VersionNumZeta - 1:
		if manversion.VersionCmp(string(version), manversion.VersionZeta) >= 0 {
			log.Info("blockchain", "切换版本Zeta高度", num, "当前版本大于等于Zeta版本, 不设置state", string(version))
			return nil
		}
		log.Info("blockchain", "切换版本Zeta高度", num)
		return bc.processStateSwitchZeta(stateDB)

	default:
		return nil
	}
	return nil
}
func (bc *BlockChain) ProcessMatrixState(block *types.Block, preVersion string, state *state.StateDBManage) error {
	return bc.matrixProcessor.ProcessMatrixState(block, preVersion, state)
}

func (bc *BlockChain) GetGraphByHash(hash common.Hash) (*mc.TopologyGraph, *mc.ElectGraph, error) {
	topologyGraph, err := bc.topologyStore.GetTopologyGraphByHash(hash)
	if err != nil {
		return nil, nil, err
	}
	electGraph, err := bc.topologyStore.GetElectGraphByHash(hash)
	if err != nil {
		return nil, nil, err
	}
	return topologyGraph, electGraph, nil
}

func (bc *BlockChain) GetGraphByState(state matrixstate.StateDB) (*mc.TopologyGraph, *mc.ElectGraph, error) {
	topologyGraph, err := matrixstate.GetTopologyGraph(state)
	if err != nil {
		return nil, nil, err
	}
	electGraph, err := matrixstate.GetElectGraph(state)
	if err != nil {
		return nil, nil, err
	}
	return topologyGraph, electGraph, nil
}

func (bc *BlockChain) GetTopologyStore() *TopologyStore {
	return bc.topologyStore
}

func (bc *BlockChain) GetBroadcastInterval() (*mc.BCIntervalInfo, error) {
	st, err := bc.State()
	if err != nil {
		return nil, errors.Errorf("get cur state err(%v)", err)
	}
	return matrixstate.GetBroadcastInterval(st)
}

func (bc *BlockChain) GetBroadcastIntervalByHash(blockHash common.Hash) (*mc.BCIntervalInfo, error) {
	st, err := bc.StateAtBlockHash(blockHash)
	if err != nil {
		return nil, errors.Errorf("get state by hash(%s) err(%v)", blockHash.Hex(), err)
	}
	return matrixstate.GetBroadcastInterval(st)
}

func (bc *BlockChain) GetBroadcastIntervalByNumber(number uint64) (*mc.BCIntervalInfo, error) {
	st, err := bc.StateAtNumber(number)
	if err != nil {
		return nil, errors.Errorf("get state by number(%d) err(%v)", number, err)
	}
	return matrixstate.GetBroadcastInterval(st)
}

func (bc *BlockChain) GetBroadcastAccounts(blockHash common.Hash) ([]common.Address, error) {
	st, err := bc.StateAtBlockHash(blockHash)
	if err != nil {
		return nil, errors.Errorf("get state by hash(%s) err(%v)", blockHash.Hex(), err)
	}
	return matrixstate.GetBroadcastAccounts(st)
}

func (bc *BlockChain) GetInnerMinerAccounts(blockHash common.Hash) ([]common.Address, error) {
	st, err := bc.StateAtBlockHash(blockHash)
	if err != nil {
		return nil, errors.Errorf("get state by hash(%s) err(%v)", blockHash.Hex(), err)
	}
	return matrixstate.GetInnerMinerAccounts(st)
}

func (bc *BlockChain) GetVersionSuperAccounts(blockHash common.Hash) ([]common.Address, error) {
	st, err := bc.StateAtBlockHash(blockHash)
	if err != nil {
		return nil, errors.Errorf("get state by hash(%s) err(%v)", blockHash.Hex(), err)
	}
	return matrixstate.GetVersionSuperAccounts(st)
}

func (bc *BlockChain) GetMultiCoinSuperAccounts(blockHash common.Hash) ([]common.Address, error) {
	st, err := bc.StateAtBlockHash(blockHash)
	if err != nil {
		return nil, errors.Errorf("get state by hash(%s) err(%v)", blockHash.Hex(), err)
	}
	return matrixstate.GetMultiCoinSuperAccounts(st)
}

func (bc *BlockChain) GetSubChainSuperAccounts(blockHash common.Hash) ([]common.Address, error) {
	st, err := bc.StateAtBlockHash(blockHash)
	if err != nil {
		return nil, errors.Errorf("get state by hash(%s) err(%v)", blockHash.Hex(), err)
	}
	return matrixstate.GetSubChainSuperAccounts(st)
}

func (bc *BlockChain) GetBlockSuperAccounts(blockHash common.Hash) ([]common.Address, error) {
	st, err := bc.StateAtBlockHash(blockHash)
	if err != nil {
		return nil, errors.Errorf("get state err by hash(%s) err(%v)", blockHash.Hex(), err)
	}
	return matrixstate.GetBlockSuperAccounts(st)
}

func (bc *BlockChain) GetSuperBlockSeq() (uint64, error) {
	st, err := bc.State()
	if err != nil {
		return 0, errors.Errorf("get cur state err (%v)", err)
	}
	superBlkCfg, err := matrixstate.GetSuperBlockCfg(st)
	if err != nil {
		return 0, err
	}
	log.Info("blockChain", "超级区块序号", superBlkCfg.Seq)
	return superBlkCfg.Seq, nil
}

func (bc *BlockChain) GetSuperBlockNum() (uint64, error) {
	st, err := bc.State()
	if err != nil {
		return 0, errors.Errorf("get cur state err (%v)", err)
	}
	superBlkCfg, err := matrixstate.GetSuperBlockCfg(st)
	if err != nil {
		return 0, err
	}
	log.Info("blockChain", "超级区块高度", superBlkCfg.Num)
	return superBlkCfg.Num, nil
}

func (bc *BlockChain) GetBlockSuperBlockInfo(blockHash common.Hash) (uint64, uint64, error) {
	st, err := bc.StateAtBlockHash(blockHash)
	if err != nil {
		return 0, 0, errors.Errorf("get state err by hash(%s) err(%v)", blockHash.Hex(), err)
	}
	superBlkCfg, err := matrixstate.GetSuperBlockCfg(st)
	if err != nil {
		return 0, 0, err
	}
	return superBlkCfg.Num, superBlkCfg.Seq, nil
}

func (bc *BlockChain) GetSuperBlockInfo() (*mc.SuperBlkCfg, error) {
	st, err := bc.State()
	if err != nil {
		return nil, errors.Errorf("get cur state err (%v)", err)
	}
	superBlkCfg, err := matrixstate.GetSuperBlockCfg(st)
	if err != nil {
		return nil, err
	}
	log.Trace("blockChain", "超级区块高度", superBlkCfg.Num, "超级区块序号", superBlkCfg.Seq)
	return superBlkCfg, nil
}

func (bc *BlockChain) GetMinDifficulty(blockHash common.Hash) (*big.Int, error) {
	st, err := bc.StateAtBlockHash(blockHash)
	if err != nil {
		return nil, errors.Errorf("get state by hash(%s) err(%v)", blockHash.Hex(), err)
	}
	return matrixstate.GetMinDifficulty(st)
}

func (bc *BlockChain) GetMaxDifficulty(blockHash common.Hash) (*big.Int, error) {
	st, err := bc.StateAtBlockHash(blockHash)
	if err != nil {
		return nil, errors.Errorf("get state by hash(%s) err(%v)", blockHash.Hex(), err)
	}
	return matrixstate.GetMaxDifficulty(st)
}
func (bc *BlockChain) GetReelectionDifficulty(blockHash common.Hash) (*big.Int, error) {
	st, err := bc.StateAtBlockHash(blockHash)
	if err != nil {
		return nil, errors.Errorf("get state by hash(%s) err(%v)", blockHash.Hex(), err)
	}
	return matrixstate.GetReelectionDifficulty(st)
}

func (bc *BlockChain) GetBlockDurationStatus(blockHash common.Hash) (*mc.BlockDurationStatus, error) {
	st, err := bc.StateAtBlockHash(blockHash)
	if err != nil {
		return nil, errors.Errorf("get state by hash(%s) err(%v)", blockHash.Hex(), err)
	}
	return matrixstate.GetBlockDuration(st)
}

func (bc *BlockChain) getMineHeader(number uint64, sonHash common.Hash, bcInterval *mc.BCIntervalInfo) (*types.Header, common.Hash, error) {
	mineHashNumber := params.GetCurAIBlockNumber(number, bcInterval.GetBroadcastInterval())
	mineHeaderHash, err := bc.GetAncestorHash(sonHash, mineHashNumber)
	if err != nil {
		return nil, common.Hash{}, fmt.Errorf("get mine header hash err: %v", err)
	}
	mineHeader := bc.GetHeaderByHash(mineHeaderHash)
	if mineHeader == nil {
		return nil, common.Hash{}, fmt.Errorf("get mine header err")
	}

	return mineHeader, mineHeader.HashNoSignsAndNonce(), nil
}

func (bc *BlockChain) SetBlockDurationStatus(header *types.Header, state *state.StateDBManage) error {
	if string(header.Version) < manversion.VersionAIMine || header.Number.Uint64() == 0 {
		return nil
	}
	bcInterval, err := bc.GetBroadcastIntervalByHash(header.ParentHash)
	if err != nil {
		return errors.Errorf("get state by hash(%s) err(%v)", header.ParentHash, err)
	}

	if !header.IsAIHeader(bcInterval.BCInterval) {
		return nil
	}

	mineHeader, _, err := bc.getMineHeader(header.Number.Uint64()-1, header.ParentHash, bcInterval)
	if err != nil {
		return err
	}
	blockDurationStatus := &mc.BlockDurationStatus{[]uint8{0}}
	if bcInterval.IsReElectionNumber(header.Number.Uint64() - 1) {
		blockDurationStatus.Status = []uint8{0}
		return matrixstate.SetBlockDuration(state, blockDurationStatus)
	}
	innerMiners, err := bc.GetInnerMinerAccounts(header.ParentHash)
	isTimeout := amhash.IsPowTimeout(mineHeader.Coinbase, innerMiners)
	if isTimeout {
		blockDurationStatus.Status = []uint8{2}
	}
	if uint64(header.Time.Uint64())-mineHeader.Time.Uint64() <= new(big.Int).Mul(params.MinBlockInterval, new(big.Int).SetUint64(params.PowBlockPeriod)).Uint64() {
		blockDurationStatus.Status = []uint8{1}
	}
	return matrixstate.SetBlockDuration(state, blockDurationStatus)
}

func ProduceBroadcastIntervalData(block *types.Block, state *state.StateDBManage, readFn PreStateReadFn) (interface{}, error) {
	bciData, err := readFn(mc.MSKeyBroadcastInterval)
	if err != nil {
		log.Error("ProduceBroadcastIntervalData", "read pre broadcast interval err", err)
		return nil, err
	}

	bcInterval, OK := bciData.(*mc.BCIntervalInfo)
	if OK == false {
		return nil, errors.New("pre broadcast interval reflect failed")
	}

	modify := false
	number := block.NumberU64()
	backupEnableNumber := bcInterval.GetBackupEnableNumber()
	if number == backupEnableNumber {
		// 备选生效时间点
		if bcInterval.IsReElectionNumber(number) == false || bcInterval.IsBroadcastNumber(number) == false {
			// 生效时间点不是原周期的选举点，数据错误
			log.Crit("ProduceBroadcastIntervalData", "backup enable number illegal", backupEnableNumber,
				"old interval", bcInterval.GetBroadcastInterval(), "last broadcast number", bcInterval.GetLastBroadcastNumber(), "last reelect number", bcInterval.GetLastReElectionNumber())
		}

		oldInterval := bcInterval.GetBroadcastInterval()

		// 设置最后的广播区块和选举区块
		bcInterval.SetLastBCNumber(backupEnableNumber)
		bcInterval.SetLastReelectNumber(backupEnableNumber)
		// 启动备选周期
		bcInterval.UsingBackupInterval()
		log.Info("ProduceBroadcastIntervalData", "old interval", oldInterval, "new interval", bcInterval.GetBroadcastInterval())
		modify = true
	} else {
		if bcInterval.IsBroadcastNumber(number) {
			bcInterval.SetLastBCNumber(number)
			modify = true
		}

		if bcInterval.IsReElectionNumber(number) {
			bcInterval.SetLastReelectNumber(number)
			modify = true
		}
	}

	if modify {
		log.Info("ProduceBroadcastIntervalData", "生成广播区块内容", "成功", "block number", number, "data", bcInterval)
		return bcInterval, nil
	} else {
		return nil, nil
	}
}
func (bc *BlockChain) UpdateCurrencyHeaderState(st *state.StateDBManage, version string, Roots []common.CoinRoot, Sharding []common.Coinbyte) error {
	if manversion.VersionCmp(version, manversion.VersionAIMine) >= 0 {
		readCurrencyHeader, err := matrixstate.GetCurrenyHeader(st)
		if nil != err {
			log.Error("blockchain", "读取多币种区块头错误", err)
		}
		readCurrencyHeaderHash := types.RlpHash(readCurrencyHeader)
		CurrentCurrencyHeader := mc.CurrencyHeader{Roots: Roots, Sharding: Sharding}
		if readCurrencyHeaderHash != types.RlpHash(CurrentCurrencyHeader) {
			if err := matrixstate.SetCurrenyHeader(st, &CurrentCurrencyHeader); nil != err {
				log.Crit("blockchain", "写入多币种区块头错误", err)
			}
		}
	}
	return nil

}
