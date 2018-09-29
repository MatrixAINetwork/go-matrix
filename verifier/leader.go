// Copyright 2018 The MATRIX Authors as well as Copyright 2014-2017 The go-ethereum Authors
// This file is consisted of the MATRIX library and part of the go-ethereum library.
//
// The MATRIX-ethereum library is free software: you can redistribute it and/or modify it under the terms of the MIT License.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, 
//and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject tothe following conditions:
//
//The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
//
//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, 
//WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISINGFROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE
//OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
package verifier

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
