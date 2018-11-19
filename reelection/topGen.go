// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package reelection

import (
	"encoding/json"
	"errors"
	"math/big"
	"time"

	"github.com/matrix/go-matrix/params"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
)

//得到随机种子
func (self *ReElection) GetSeed(height uint64) (*big.Int, error) {

	sendData := self.CalcbeforeSeedGen(height)

	var err error
	self.electionSeedSub, err = mc.SubscribeEvent(mc.Random_TopoSeedRsp, self.electionSeedCh)
	if err != nil {
		return nil, err
	}
	mc.PublishEvent(mc.ReElec_TopoSeedReq, &sendData)

	select {
	case seed := <-self.electionSeedCh:
		log.INFO(Module, "received seed", seed)
		self.electionSeedSub.Unsubscribe()
		return seed.Seed, nil

	case <-time.After(Time_Out_Limit):
		self.electionSeedSub.Unsubscribe()
		log.INFO(Module, "received seed", "Time_Out_Falied")
		return nil, errors.New("Seed Gen failed")
	}
}

//随机种子生成前的消息准备
func (self *ReElection) CalcbeforeSeedGen(height uint64) mc.RandomRequest {
	broadcastInterval := common.GetBroadcastInterval()
	height_1 := height / broadcastInterval * broadcastInterval //上一个广播区块
	height_2 := height_1 - common.GetBroadcastInterval()

	minHash := self.getMinHash(height)
	PrivateMap := getKeyTransInfo(height_1, mc.Privatekey)
	PublicMap := getKeyTransInfo(height_2, mc.Publickey)
	log.INFO(Module, "获取到的公私钥匙, 公钥长度", len(PublicMap), "私钥长度", len(PrivateMap), "当前高度", height)
	for k, v := range PublicMap {
		log.INFO(Module, "公钥key", k, "value", v, "当前高度", height)
	}
	for k, v := range PrivateMap {
		log.INFO(Module, "私钥key", k, "value", v, "当前高度", height)
	}

	return mc.RandomRequest{MinHash: minHash, PrivateMap: PrivateMap, PublicMap: PublicMap}
}
func (self *ReElection) GetHashByNum(height uint64) common.Hash {
	return self.bc.GetBlockByNumber(height).Hash()
}
func (self *ReElection) getMinHash(height uint64) common.Hash {

	minhash := self.GetHashByNum(height)
	BroadcastInterval := common.GetBroadcastInterval()
	for i := height - 1; i > height-BroadcastInterval; i-- {
		blockhash := self.GetHashByNum(uint64(i))
		if minhash.Big().Cmp(blockhash.Big()) == 1 { //前者大于后者
			minhash = blockhash
		}

	}
	return minhash
}

func (self *ReElection) ToGenMinerTop(height uint64) error {

	minerDeposit, err := GetAllElectedByHeight(big.NewInt(int64(height)), common.RoleMiner) //
	if err != nil {
		log.ERROR(Module, "獲取礦工抵押交易失敗 err", err)
		return err
	}
	log.INFO(Module, "矿工抵押交易", minerDeposit)

	seed, err := self.GetSeed(height)
	if err != nil {
		log.ERROR(Module, "獲取種子失敗 err", err)
		return err
	}
	log.Info(Module, "矿工选举种子", seed)

	self.minerGenSub, err = mc.SubscribeEvent(mc.Topo_MasterMinerElectionRsp, self.minerGenCh)
	if err != nil {
		log.ERROR(Module, "訂閱Topo_MasterMinerElectionRsp err", err)
		return err
	}

	err = mc.PublishEvent(mc.ReElec_MasterMinerReElectionReq, &mc.MasterMinerReElectionReqMsg{SeqNum: height, RandSeed: seed, MinerList: minerDeposit})
	log.INFO(Module, "發送-礦工拓撲生成請求", mc.MasterMinerReElectionReqMsg{SeqNum: height, RandSeed: seed, MinerList: minerDeposit})

	select {
	case TopRsp := <-self.minerGenCh:
		log.INFO(Module, "收到礦工拓撲生成相應,data", TopRsp)

		self.minerGenSub.Unsubscribe()
		err := self.writeElectData(common.RoleMiner, height+params.MinerTopologyGenerateUptime-params.MinerNetChangeUpTime, ElectMiner{MasterMiner: TopRsp.MasterMiner, BackUpMiner: TopRsp.BackUpMiner}, ElectValidator{})
		log.INFO(Module, "寫礦工的選舉信息到數據庫", err, "data", ElectMiner{MasterMiner: TopRsp.MasterMiner, BackUpMiner: TopRsp.BackUpMiner}, ElectValidator{})

		return err

	case <-time.After(Time_Out_Limit):
		self.minerGenSub.Unsubscribe()
		log.INFO(Module, "礦工拓撲生成相應失敗 err TimeOut", Time_Out_Limit)
		return nil
	}

}

func (self *ReElection) ToGenValidatorTop(height uint64) error {
	validatoeDeposit, err := GetAllElectedByHeight(big.NewInt(int64(height)), common.RoleValidator)
	if err != nil {
		log.ERROR(Module, "獲取驗證者抵押列表失敗 err", err)
		return err
	}
	log.INFO(Module, "验证者抵押账户", validatoeDeposit)
	foundDeposit := GetFound()

	seed, err := self.GetSeed(height)
	if err != nil {
		log.ERROR(Module, "獲取驗證者種子生成失敗 err", err)
		return err
	}
	log.INFO(Module, "验证者随机种子", seed)

	self.validatorGenSub, err = mc.SubscribeEvent(mc.Topo_MasterValidatorElectionRsp, self.validatorGenCh)
	if err != nil {
		log.ERROR(Module, "訂閱Topo_MasterValidatorElectionRsp err", err)
		return err
	}

	err = mc.PublishEvent(mc.ReElec_MasterValidatorElectionReq, &mc.MasterValidatorReElectionReqMsg{SeqNum: height, RandSeed: seed, ValidatorList: validatoeDeposit, FoundationValidatoeList: foundDeposit})
	if err != nil {
		log.ERROR(Module, "發送 礦工拓撲生成請求 err", err)
	}

	select {
	case TopRsp := <-self.validatorGenCh:
		log.INFO(Module, "收到驗證者拓撲生成相應 data", TopRsp)
		self.validatorGenSub.Unsubscribe()
		err := self.writeElectData(common.RoleValidator, height+params.VerifyTopologyGenerateUpTime-params.VerifyNetChangeUpTime, ElectMiner{}, ElectValidator{MasterValidator: TopRsp.MasterValidator,
			BackUpValidator:    TopRsp.BackUpValidator,
			CandidateValidator: TopRsp.CandidateValidator,
		})
		return err
	case <-time.After(Time_Out_Limit):
		self.validatorGenSub.Unsubscribe()
		log.ERROR(Module, "驗證者拓撲生成相應獲取超時 ", Time_Out_Limit)

		return nil
	}

}
func (self *ReElection) writeElectData(aim common.RoleType, height uint64, minerData ElectMiner, validatorData ElectValidator) error {

	switch {
	case aim == common.RoleMiner:
		data, err := json.Marshal(minerData)
		if err != nil {
			log.INFO(Module, "Marshal 礦工數據失敗 err", err, "data", data)
			return err
		}
		key := MakeElectDBKey(height, common.RoleMiner)

		err = self.ldb.Put([]byte(key), data, nil)
		if err != nil {
			log.ERROR(Module, "礦工 寫入數據庫失敗 err", err)
			return err
		}
		log.INFO(Module, "key", key, "value", minerData)
		return nil

	case aim == common.RoleValidator:
		data, err := json.Marshal(validatorData)
		if err != nil {
			log.INFO(Module, "Marshal 驗證者數據失敗 err", err, "data", data)
			return err
		}
		key := MakeElectDBKey(height, common.RoleValidator)
		err = self.ldb.Put([]byte(key), data, nil)
		if err != nil {
			log.ERROR(Module, "驗證者數據寫入數據庫失敗 err", err)
			return err
		}
		log.INFO(Module, "key", key, "value", validatorData)
		return nil
	}
	return nil
}

func (self *ReElection) readElectData(aim common.RoleType, height uint64) (ElectMiner, ElectValidator, error) {

	switch {
	case aim == common.RoleMiner:
		key := MakeElectDBKey(height, common.RoleMiner)
		ans, err := self.ldb.Get([]byte(key), nil)
		if err != nil {
			log.ERROR(Module, "獲取db礦工選舉信息失敗 err", err, "key", key)
			return ElectMiner{}, ElectValidator{}, err
		}
		var realAns ElectMiner
		err = json.Unmarshal(ans, &realAns)
		if err != nil {
			log.ERROR(Module, "db里的礦工選舉信息Unmarshal失敗", err, "data", ans)
			return ElectMiner{}, ElectValidator{}, err
		}
		return realAns, ElectValidator{}, nil

	case aim == common.RoleValidator:
		key := MakeElectDBKey(height, common.RoleValidator)
		ans, err := self.ldb.Get([]byte(key), nil)
		if err != nil {
			log.INFO(Module, "獲取db驗證者選舉信息失敗 err", err, "key", key)
			return ElectMiner{}, ElectValidator{}, err
		}
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
func MakeElectDBKey(height uint64, role common.RoleType) string {
	t := big.NewInt(int64(height))
	switch {
	case role == common.RoleMiner:
		key := t.String() + "---" + "Miner---Elect"
		return key
	case role == common.RoleValidator:
		key := t.String() + "---" + "Validator---Elect"
		return key
	default:
		log.ERROR("MakeElectDBKey failed role is not mathch role", role)
	}
	return ""
}
