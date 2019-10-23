// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package reelection

import (
	"errors"
	"math/big"

	"github.com/MatrixAINetwork/go-matrix/ca"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/matrixstate"
	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/core/vm"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/params/manparams"
)

type TopGenStatus struct {
	MastV []mc.ElectNodeInfo
	BackV []mc.ElectNodeInfo
	CandV []mc.ElectNodeInfo

	MastM []mc.ElectNodeInfo
	BackM []mc.ElectNodeInfo
	CandM []mc.ElectNodeInfo
}


//得到随机种子
func (self *ReElection) GetSeed(hash common.Hash) (*big.Int, error) {
	seed, err := self.random.GetRandom(hash, manparams.ElectionSeed)
	//log.Info(Module, "common.Default seed", seed)
	return seed, err

}

func (self *ReElection) ToGenMinerTop(hash common.Hash, stateDb *state.StateDBManage) ([]mc.ElectNodeInfo, []mc.ElectNodeInfo, []mc.ElectNodeInfo, error) {
	//log.Info(Module, "准备生成矿工拓扑图", "start", "hash", hash.String())
	//defer log.Info(Module, "生成矿工拓扑图结束", "end", "hash", hash.String())
	height, err := self.GetNumberByHash(hash)
	if err != nil {
		log.Error(Module, "根据hash算高度失败 ToGenMinerTop hash", hash, "err", err)
		return []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, err
	}
	data, err := self.GetElectGenTimes(hash)
	if err != nil {
		log.Error(Module, "获取选举信息失败 高度", height, "err", err)
		return []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, err
	}
	minerGen := uint64(data.MinerGen)

	bcInterval, err := self.GetBroadcastIntervalByHash(hash)
	if err != nil {
		log.Error(Module, "get broadcast interval err", err)
		return []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, err
	}

	height = bcInterval.GetNextReElectionNumber(height) - minerGen
	AncestorHash, err := self.bc.GetAncestorHash(hash, height)
	if nil != err {
		log.Error(Module, "获取选举制定高度hash错误，高度", height, "err", err)
		return []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, err
	}
	minerDeposit, err := GetAllElectedByHash(AncestorHash, common.RoleMiner) //
	if err != nil {
		log.Error(Module, "获取矿工抵押列表失败 err", err)
		return []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, err
	}
	//log.Info(Module, "矿工抵押交易", minerDeposit)

	elect, err := self.GetElectPlug(hash)
	if err != nil {
		log.Error(Module, "获取选举插件失败 err", err, "高度", height)
		return []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, err
	}
	electConf, err := self.GetElectConfig(hash)
	if err != nil {
		log.Error(Module, "获取选举信息失败 err", err, "高度", height)
		return []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, err
	}

	seed, err := self.GetSeed(hash)
	if err != nil {
		log.Error(Module, "获取种子失败 err", err)
		return []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, err
	}
	//log.Info(Module, "矿工选举种子", seed)

	TopRsp := elect.MinerTopGen(&mc.MasterMinerReElectionReqMsg{SeqNum: height, RandSeed: seed, MinerList: minerDeposit, ElectConfig: *electConf}, stateDb)

	return TopRsp.MasterMiner, []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, nil
}

func (self *ReElection) addBlockProduceBlackList(hash common.Hash) (*mc.BlockProduceSlashBlackList, error) {
	st, err := self.bc.StateAtBlockHash(hash)
	if err != nil {
		log.Error(Module, "获取state 错误", err, "number", hash)
		return &mc.BlockProduceSlashBlackList{}, err
	}

	slashCfg, err := matrixstate.GetBlockProduceSlashCfg(st)
	if err != nil {
		log.Error(Module, "slashCfg 错误", err)
		return &mc.BlockProduceSlashBlackList{}, err
	}

	if !slashCfg.Switcher {
		log.Debug(Module, "slashCfg 状态关闭", nil)
		return &mc.BlockProduceSlashBlackList{}, nil
	}

	produceBlackList, err := matrixstate.GetBlockProduceBlackList(st)
	if err != nil {
		log.Error(Module, "获取produce blackList 错误", err)
		return &mc.BlockProduceSlashBlackList{}, err
	}

	return produceBlackList, nil
}
func (self *ReElection) ToGenValidatorTop(hash common.Hash, stateDb *state.StateDBManage) ([]mc.ElectNodeInfo, []mc.ElectNodeInfo, []mc.ElectNodeInfo, error) {
	//log.Info(Module, "准备生成验证者拓扑图", "start", "hash", hash.String())
	//defer log.Info(Module, "生成验证者拓扑图结束", "end", "hash", hash.String())
	height, err := self.GetNumberByHash(hash)
	if err != nil {
		log.Error(Module, "根据hash算高度失败 ToGenValidatorTop hash", hash.String())
		return []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, err
	}
	data, err := self.GetElectGenTimes(hash)
	if err != nil {
		log.Error(Module, "获取选举信息失败 err", err)
		return []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, err
	}
	verifyGenTime := uint64(data.ValidatorGen)
	bcInterval, err := self.GetBroadcastIntervalByHash(hash)
	if err != nil {
		log.Error(Module, "根据hash获取广播周期信息 err", err)
		return []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, err
	}
	height = bcInterval.GetNextReElectionNumber(height) - verifyGenTime
	AncestorHash, err := self.bc.GetAncestorHash(hash, height)
	if nil != err {
		log.Error(Module, "获取选举制定高度hash错误，高度", height, "err", err)
		return []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, err
	}

	validatoeDeposit, err := GetAllElectedByHash(AncestorHash, common.RoleValidator)
	if err != nil {
		log.Error(Module, "获取验证者列表失败 err", err)
		return []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, err
	}
	//log.Info(Module, "验证者抵押账户", validatoeDeposit)
	foundDeposit := GetFound()

	elect, err := self.GetElectPlug(hash)
	if err != nil {
		log.Error(Module, "获取选举插件失败 err", err, "高度", height)
		return []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, err
	}
	electConf, err := self.GetElectConfig(hash)
	if err != nil {
		log.Error(Module, "获取选举信息失败 err", err, "高度", height)
		return []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, err
	}
	seed, err := self.GetSeed(hash)
	if err != nil {
		log.Error(Module, "获取验证者种子失败 err", err)
		return []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, err
	}
	//log.Info(Module, "验证者随机种子", seed)

	vipList, err := self.GetViPList(hash)
	if err != nil {
		log.Error(Module, "获取viplist为空 err", err, "高度", height)
		return []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, err
	}
	produceBlackList, err := self.addBlockProduceBlackList(hash)
	if err != nil {
		log.Error(Module, "获取区块生产惩罚错误", err, "高度", height)
		return []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, err
	}

	TopRsp := elect.ValidatorTopGen(&mc.MasterValidatorReElectionReqMsg{SeqNum: height, RandSeed: seed, ValidatorList: validatoeDeposit, FoundationValidatorList: foundDeposit, ElectConfig: *electConf, VIPList: vipList, BlockProduceBlackList: *produceBlackList}, stateDb)

	return TopRsp.MasterValidator, TopRsp.BackUpValidator, TopRsp.CandidateValidator, nil

}
func GetFound() []vm.DepositDetail {
	return []vm.DepositDetail{}
}
func GetAllElectedByHash(hash common.Hash, tp common.RoleType) ([]vm.DepositDetail, error) {

	switch tp {
	case common.RoleMiner:
		ans, err := ca.GetElectedByHeightAndRoleByHash(hash, common.RoleMiner)
		//log.Info("從CA獲取礦工抵押交易", "data", ans, "height", Heigh)
		if err != nil {
			return []vm.DepositDetail{}, errors.New("获取矿工交易身份不对")
		}
		return ans, nil
	case common.RoleValidator:
		ans, err := ca.GetElectedByHeightAndRoleByHash(hash, common.RoleValidator)
		//log.Info("從CA獲取驗證者抵押交易", "data", ans, "height", Heigh)
		if err != nil {
			return []vm.DepositDetail{}, errors.New("获取验证者交易身份不对")
		}
		return ans, nil

	default:
		return []vm.DepositDetail{}, errors.New("获取抵押交易身份不对")
	}
}
