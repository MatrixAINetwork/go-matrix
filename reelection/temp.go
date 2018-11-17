// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package reelection

import (
	"errors"
	"math/big"
	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/core/vm"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/core"
)

func (self *ReElection) TopoUpdate(Q0, Q1, Q2 []mc.TopologyNodeInfo, netTop mc.TopologyGraph, offline []common.Address) []mc.Alternative {
	return self.elect.ToPoUpdate(Q0, Q1, Q2, netTop, offline)
}
func (self *ReElection) NativeUpdate(Q0, Q1, Q2 []mc.TopologyNodeInfo, online mc.TopologyNodeInfo, flag int) ([]mc.TopologyNodeInfo, []mc.TopologyNodeInfo, []mc.TopologyNodeInfo) {
	return self.elect.PrimarylistUpdate(Q0, Q1, Q1, online, flag)
}

func (self *ReElection)GetNumberByHash(hash common.Hash)(uint64,error){
	tHeader:=self.bc.GetHeaderByHash(hash)
	if tHeader==nil {
		log.Error(Module,"GetNumberByHash 根据hash算header失败 hash" ,hash.String())
		return 0,errors.New("根据hash算header失败")
	}
	if tHeader.Number==nil {
		log.Error(Module,"GetNumberByHash header 内的高度获取失败" ,hash.String())
		return 0,errors.New("header 内的高度获取失败")
	}
	return tHeader.Number.Uint64(),nil
}

func (self *ReElection)GetHeaderHashByNumber(hash common.Hash,height uint64)(common.Hash,error){
	AimHash,err:=self.bc.GetAncestorHash(hash,height)
	if err!=nil{
		log.Error(Module,"获取祖先hash失败 hash",hash.String(),"height",height)
		return common.Hash{},err
	}
	return AimHash,nil
}
//得到特殊交易
func (self *ReElection)getKeyTransInfo(hash common.Hash,Height uint64, types string) map[common.Address][]byte {
	aimHash,err:=self.GetHeaderHashByNumber(hash,Height)
	if err!=nil{
		log.Error(Module,"获取特殊交易阶段-获取祖先hash失败 hash",hash.String(),"height",Height)
		return make(map[common.Address][]byte)
	}

	ans, err := core.GetBroadcastTxs(aimHash, types)
	if err != nil {
			log.Error(Module, "获取特殊交易失败 Height", Height, "types", types)
	}
	return ans
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

func GetFound() []vm.DepositDetail {
	return []vm.DepositDetail{}
}
