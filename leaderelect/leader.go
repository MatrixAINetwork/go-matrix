// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package leaderelect

import (
	"errors"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
)

var (
	modelName                     = "leader计算模块"
	ElectionInfoFromHeaderIsEmpty = errors.New("Election Info From Header Is Empty")
	ValidatorIsEmpty              = errors.New("Validator Is Empty")
	ValidatorNotFound             = errors.New("Validator Not Found")
)

/*func currBlockLeaer(blockHeight *big.Int, preLeader mc.TopologyNodeInfo, validator []*mc.TopologyNodeInfo, electionInfo []mc.TopologyNodeInfo) (ca.TopologyNodeInfo, error) {
	SubHeight := big.NewInt(0)
	SubHeight.Sub(blockHeight, big.NewInt(1))

	if SubHeight.Mod(SubHeight, big.NewInt(int64(common.GetBroadcastInterval()))).Cmp(big.NewInt(0)) != 0 { //不是广播周期的第一个块
		return nextLeader(validator, preLeader)
	}

		//if blockHeight.Cmp(big.NewInt(1)) == 0 { //是第一块:从配置中拿
		//	return getByParam()
		//}

	if len(electionInfo) == 0 { //如果选举信息为空
		return ca.TopologyNodeInfo{}, ElectionInfoFromHeaderIsEmpty
	}
	return electionInfo[0], nil

}
*/

/*func nextLeader(validator []*ca.TopologyNodeInfo, preLeader ca.TopologyNodeInfo) (ca.TopologyNodeInfo, error) {
	if len(validator) == 0 {
		return ca.TopologyNodeInfo{}, ValidatorIsEmpty
	}
	ValidatorNum := len(validator)
	for k, v := range validator {
		if v.Account == preLeader.Account {
			nextIndex := (k + 1) % ValidatorNum
			return *validator[nextIndex], nil
		}
	}
	return ca.TopologyNodeInfo{}, ValidatorNotFound
}*/

func nextLeaderByNum(validator []mc.TopologyNodeInfo, preLeader common.Address, skipNum uint8) (common.Address, error) {
	if validator == nil {
		log.ERROR(modelName, "计算Leader错误， 验证者列表为", "空")
		return common.Address{}, ValidatorIsEmpty
	}
	if skipNum == 0 {
		return preLeader, nil
	}
	ValidatorNum := len(validator)
	for k, v := range validator {
		if v.Account == preLeader {
			nextIndex := (k + int(skipNum)) % ValidatorNum
			return validator[nextIndex].Account, nil
		}
	}
	log.ERROR(modelName, "计算Leader错误， 没有找到Leader， 前一个", preLeader.Hex())
	return common.Address{}, ValidatorNotFound
}
