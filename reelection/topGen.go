// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package reelection

import (
	"encoding/json"
	"errors"
	"math/big"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/core/vm"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/ca"
)

//得到随机种子
func (self *ReElection) GetSeed(hash common.Hash) (*big.Int, error) {
	return self.random.GetRandom(hash, "electionseed")
}

func (self *ReElection) ToGenMinerTop(hash common.Hash) error {
	height, err := self.GetNumberByHash(hash)
	if err != nil {
		return err
	}

	minerDeposit, err := GetAllElectedByHeight(big.NewInt(int64(height)), common.RoleMiner) //
	if err != nil {
		log.ERROR(Module, "获取矿工抵押列表失败 err", err)
		return err
	}
	log.INFO(Module, "矿工抵押交易", minerDeposit)

	seed, err := self.GetSeed(hash)
	if err != nil {
		log.ERROR(Module, "获取种子失败 err", err)
		return err
	}
	log.Info(Module, "矿工选举种子", seed)

	TopRsp := self.elect.MinerTopGen(&mc.MasterMinerReElectionReqMsg{SeqNum: height, RandSeed: seed, MinerList: minerDeposit})
	err = self.writeElectData(common.RoleMiner, hash, ElectMiner{MasterMiner: TopRsp.MasterMiner, BackUpMiner: TopRsp.BackUpMiner}, ElectValidator{})
	log.INFO(Module, "寫礦工的選舉信息到數據庫", err, "data", ElectMiner{MasterMiner: TopRsp.MasterMiner, BackUpMiner: TopRsp.BackUpMiner}, ElectValidator{})

	return err

}

func (self *ReElection) ToGenValidatorTop(hash common.Hash) error {
	height, err := self.GetNumberByHash(hash)
	if err != nil {
		return err
	}

	validatoeDeposit, err := GetAllElectedByHeight(big.NewInt(int64(height)), common.RoleValidator)
	if err != nil {
		log.ERROR(Module, "獲取驗證者抵押列表失敗 err", err)
		return err
	}
	log.INFO(Module, "验证者抵押账户", validatoeDeposit)
	foundDeposit := GetFound()

	seed, err := self.GetSeed(hash)
	if err != nil {
		log.ERROR(Module, "獲取驗證者種子生成失敗 err", err)
		return err
	}
	log.INFO(Module, "验证者随机种子", seed)
	TopRsp := self.elect.ValidatorTopGen(&mc.MasterValidatorReElectionReqMsg{SeqNum: height, RandSeed: seed, ValidatorList: validatoeDeposit, FoundationValidatoeList: foundDeposit})
	err = self.writeElectData(common.RoleValidator, hash, ElectMiner{}, ElectValidator{MasterValidator: TopRsp.MasterValidator,
		BackUpValidator:    TopRsp.BackUpValidator,
		CandidateValidator: TopRsp.CandidateValidator,
	})
	return err

}
func (self *ReElection) writeElectData(aim common.RoleType, hash common.Hash, minerData ElectMiner, validatorData ElectValidator) error {

	switch {
	case aim == common.RoleMiner:
		data, err := json.Marshal(minerData)
		if err != nil {
			log.INFO(Module, "Marshal 礦工數據失敗 err", err, "data", data)
			return err
		}
		key := MakeElectDBKey(hash, common.RoleMiner)

		err = self.ldb.Put([]byte(key), data, nil)
		if err != nil {
			log.ERROR(Module, "礦工 寫入數據庫失敗 err", err)
			return err
		}
		log.INFO(Module, "数据库矿工拓扑生成 err", err, "高度对应的hash", hash, "key", key)
		return nil

	case aim == common.RoleValidator:
		data, err := json.Marshal(validatorData)
		if err != nil {
			log.INFO(Module, "Marshal 驗證者數據失敗 err", err, "data", data)
			return err
		}
		key := MakeElectDBKey(hash, common.RoleValidator)
		err = self.ldb.Put([]byte(key), data, nil)
		if err != nil {
			log.ERROR(Module, "驗證者數據寫入數據庫失敗 err", err)
			return err
		}
		log.INFO(Module, "数据库 验证者拓扑生成 err", err, "高度对应的hash", hash, "key", key)
		return nil
	}
	return nil
}

func (self *ReElection) readElectData(aim common.RoleType, hash common.Hash) (ElectMiner, ElectValidator, error) {
	key := MakeElectDBKey(hash, aim)
	ans, err := self.ldb.Get([]byte(key), nil)
	if err != nil {
		log.ERROR(Module, "获取选举信息失败 err", err, "key", key)
		return ElectMiner{}, ElectValidator{}, err
	}

	switch {
	case aim == common.RoleMiner:
		var realAns ElectMiner
		err = json.Unmarshal(ans, &realAns)
		if err != nil {
			log.ERROR(Module, "db里的礦工選舉信息Unmarshal失敗", err, "data", ans)
			return ElectMiner{}, ElectValidator{}, err
		}
		return realAns, ElectValidator{}, nil

	case aim == common.RoleValidator:
		var realAns ElectValidator
		err = json.Unmarshal(ans, &realAns)
		if err != nil {
			log.INFO(Module, "db里的驗證者選舉信息Unmarshal失敗", err, "data", ans)
			return ElectMiner{}, ElectValidator{}, err
		}
		return ElectMiner{}, realAns, nil
	default:
		log.ERROR(Module, "讀選舉信息，請使用礦工或者驗證者，暫時不支持其他模式", "nil")
		return ElectMiner{}, ElectValidator{}, errors.New("選舉角色一定是礦工或者驗證者")
	}

}
func MakeElectDBKey(hash common.Hash, role common.RoleType) string {
	switch {
	case role == common.RoleMiner:
		key := hash.String() + "---" + "Miner---Elect"
		return key
	case role == common.RoleValidator:
		key := hash.String() + "---" + "Validator---Elect"
		return key
	default:
		log.ERROR("MakeElectDBKey failed role is not mathch role", role)
	}
	return ""
}
func GetFound() []vm.DepositDetail {
	return []vm.DepositDetail{}
}

func GetAllElectedByHeight(Heigh *big.Int, tp common.RoleType) ([]vm.DepositDetail, error) {

	switch tp {
	case common.RoleMiner:
		ans, err := ca.GetElectedByHeightAndRole(Heigh, common.RoleMiner)
		log.INFO("從CA獲取礦工抵押交易", "data", ans, "height", Heigh)
		if err != nil {
			return []vm.DepositDetail{}, errors.New("获取矿工交易身份不对")
		}
		return ans, nil
	case common.RoleValidator:
		ans, err := ca.GetElectedByHeightAndRole(Heigh, common.RoleValidator)
		log.Info("從CA獲取驗證者抵押交易", "data", ans, "height", Heigh)
		if err != nil {
			return []vm.DepositDetail{}, errors.New("获取验证者交易身份不对")
		}
		return ans, nil

	default:
		return []vm.DepositDetail{}, errors.New("获取抵押交易身份不对")
	}
}