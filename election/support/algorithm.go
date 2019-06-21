// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package support

import (
	"errors"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/common/mt19937"
	"github.com/MatrixAINetwork/go-matrix/log"
)

func GetList_VIP(probnormalized []Pnormalized, needNum int, rand *mt19937.RandUniform) ([]Strallyint, []Pnormalized) {
	probnormalized = Normalize_VIP(probnormalized)

	if len(probnormalized) == 0 {
		return []Strallyint{}, probnormalized
	}
	if needNum > len(probnormalized) {
		needNum = len(probnormalized)
	}
	ChoseNode := []Strallyint{}
	RemainingProbNormalizedNodes := []Pnormalized{}
	dict := make(map[common.Address]int)
	orderAddress := []common.Address{}

	for i := 0; i < MaxSample; i++ {
		tempRand := float64(rand.Uniform(0.0, 1.0))

		node, status := Sample1NodesInValNodes_VIP(probnormalized, tempRand)
		if !status {
			continue
		}
		_, ok := dict[node]
		if ok == true {
			dict[node] = dict[node] + 1
		} else {
			dict[node] = 1

			orderAddress = append(orderAddress, node)
		}
		if len(dict) == (needNum) {
			break
		}
	}

	for _, v := range orderAddress {
		ChoseNode = append(ChoseNode, Strallyint{Addr: v, Value: dict[v]})
	}

	for _, item := range probnormalized {
		if _, ok := dict[item.Addr]; ok == true {
			continue
		}
		if len(ChoseNode) < needNum {
			ChoseNode = append(ChoseNode, Strallyint{Addr: item.Addr, Value: 1})
		} else {
			RemainingProbNormalizedNodes = append(RemainingProbNormalizedNodes, item)
		}
	}

	return ChoseNode, RemainingProbNormalizedNodes
}

func Normalize_VIP(probVal []Pnormalized) []Pnormalized {
	var pnormalizedlist []Pnormalized

	total := 0.0
	for _, item := range probVal {
		pnormalizedlist = append(pnormalizedlist, Pnormalized{Addr: item.Addr, Value: total})
		total += item.Value

	}
	for index := 0; index < len(probVal); index++ {
		pnormalizedlist[index].Value /= total
	}

	return pnormalizedlist
}

func Sample1NodesInValNodes_VIP(probnormalized []Pnormalized, rand01 float64) (common.Address, bool) {
	len := len(probnormalized)
	for index := len - 1; index >= 0; index-- {
		if rand01 >= probnormalized[index].Value {
			return probnormalized[index].Addr, true
		}
	}

	return common.Address{}, false
}

func GetList_Common(probnormalized []Pnormalized, needNum int, rand *mt19937.RandUniform) ([]Strallyint, []Pnormalized) {
	probnormalized = Normalize_Common(probnormalized)

	if len(probnormalized) == 0 {
		return []Strallyint{}, probnormalized
	}
	if needNum > len(probnormalized) {
		needNum = len(probnormalized)
	}
	ChoseNode := []Strallyint{}
	RemainingProbNormalizedNodes := []Pnormalized{}
	dict := make(map[common.Address]int)
	orderAddress := []common.Address{}

	for i := 0; i < MaxSample; i++ {
		tempRand := float64(rand.Uniform(0.0, 1.0))

		node, status := Sample1NodesInValNodes_Common(probnormalized, tempRand)
		if !status {
			continue
		}
		_, ok := dict[node]
		if ok == true {
			dict[node] = dict[node] + 1
		} else {
			dict[node] = 1

			orderAddress = append(orderAddress, node)
		}
		if len(dict) == (needNum) {
			break
		}
	}

	for _, v := range orderAddress {
		ChoseNode = append(ChoseNode, Strallyint{Addr: v, Value: dict[v]})
	}

	for _, item := range probnormalized {
		if _, ok := dict[item.Addr]; ok == true {
			continue
		}
		if len(ChoseNode) < needNum {
			ChoseNode = append(ChoseNode, Strallyint{Addr: item.Addr, Value: 1})
		} else {
			RemainingProbNormalizedNodes = append(RemainingProbNormalizedNodes, item)
		}
	}

	return ChoseNode, RemainingProbNormalizedNodes
}

func mapCounter(m map[common.Address]int, address common.Address) (error, bool) {
	if m == nil {
		return errors.New("input map is nil"), false
	}
	if _, ok := m[address]; ok {
		m[address] = m[address] + 1
		return nil, false
	} else {
		m[address] = 1
		return nil, true
	}
}

func isAllSuperNodeSampled(supeNodeMap map[common.Address]int) bool {
	if len(supeNodeMap) == 0 {
		return true
	}
	status := true
	for _, v := range supeNodeMap {
		if v == 0 {
			status = false
			break
		}
	}
	return status
}

func RandSampleFilterBlackList(randNodeValue []Pnormalized, superNodeValue []Pnormalized, needNum int, rand *mt19937.RandUniform, blackList *BlockProduceProc) ([]Strallyint, map[common.Address]int) {
	probnormalized := append(randNodeValue, superNodeValue...)
	probnormalized = Normalize_Common(probnormalized)
	ChoseNode := make([]Strallyint, 0)

	//init vip stock 0
	superNodeStcok := make(map[common.Address]int)
	for _, v := range superNodeValue {
		superNodeStcok[v.Addr] = 0
	}

	if len(randNodeValue) == 0 {
		return ChoseNode, superNodeStcok
	}

	if needNum > len(randNodeValue) {
		needNum = len(randNodeValue)
	}

	nonBlackListDict := make(map[common.Address]int)
	orderAddress := []common.Address{}
	blackListDict := make(map[common.Address]int)

	for i := 0; i < PowerWeightMaxSmple; i++ {
		tempRand := float64(rand.Uniform(0.0, 1.0))
		node, status := Sample1NodesInValNodes_Common(probnormalized, tempRand)
		if !status {
			continue
		}
		//if select superNode, superNode stock add 1. It's not a rand node
		if val, ok := superNodeStcok[node]; ok {
			superNodeStcok[node] = val + 1
			continue
		}

		if _, ok := blackList.IsBlackList(node); ok {
			if err, _ := mapCounter(blackListDict, node); err != nil {
				log.ERROR("Election Module", "blackListDict", "uninitialized")
			}
		} else {
			if err, status := mapCounter(nonBlackListDict, node); err != nil {
				log.ERROR("Election Module", "blackListDict", "uninitialized")
			} else if status {
				orderAddress = append(orderAddress, node)
			}
		}

		if len(nonBlackListDict) >= (needNum) && isAllSuperNodeSampled(superNodeStcok) {
			break
		}
	}

	//super node stock protect
	for key, val := range superNodeStcok {
		if val == 0 {
			superNodeStcok[key] = 1
		}
	}

	for k, v := range blackListDict {
		log.Trace("Layered_BSS", "RandNode", k.String(), "Elect Slash", true, "Rand PickNum", v)
	}

	for _, v := range orderAddress {
		ChoseNode = append(ChoseNode, Strallyint{Addr: v, Value: nonBlackListDict[v]})
	}

	for _, item := range probnormalized {

		if _, ok := superNodeStcok[item.Addr]; ok == true {
			continue
		}

		if _, ok := nonBlackListDict[item.Addr]; ok == true {
			continue
		}

		if _, ok := blackList.IsBlackList(item.Addr); ok {
			continue
		}

		if len(ChoseNode) < needNum {
			ChoseNode = append(ChoseNode, Strallyint{Addr: item.Addr, Value: 1})
		}
	}

	//modify block produce slash counter
	for address, _ := range blackListDict {
		blackList.DecrementCount(address)
	}

	return ChoseNode, superNodeStcok
}

func GetList_MEP(probnormalized []Pnormalized, needNum int, rand *mt19937.RandUniform) ([]Strallyint, []Pnormalized) {
	probnormalized = Normalize_Common(probnormalized)
	if len(probnormalized) == 0 {
		return []Strallyint{}, probnormalized
	}
	if needNum > len(probnormalized) {
		needNum = len(probnormalized)
	}
	ChoseNode := []Strallyint{}
	RemainingProbNormalizedNodes := []Pnormalized{}
	dict := make(map[common.Address]int)
	orderAddress := []common.Address{}
	for i := 0; i < MaxSample; i++ {
		tempRand := float64(rand.Uniform(0.0, 1.0))
		node, status := Sample1NodesInValNodes_Common(probnormalized, tempRand)
		if !status {
			continue
		}
		_, ok := dict[node]
		if ok == true {
			dict[node] = dict[node] + 1
		} else {
			dict[node] = 1
			orderAddress = append(orderAddress, node)
		}
		if len(dict) == (needNum) {
			break
		}
	}
	for _, v := range orderAddress {
		ChoseNode = append(ChoseNode, Strallyint{Addr: v, Value: DefaultMinerStock})
	}
	for _, item := range probnormalized {
		if _, ok := dict[item.Addr]; ok == true {
			continue
		}
		if len(ChoseNode) < needNum {
			ChoseNode = append(ChoseNode, Strallyint{Addr: item.Addr, Value: DefaultMinerStock})
		} else {
			RemainingProbNormalizedNodes = append(RemainingProbNormalizedNodes, item)
		}
	}
	return ChoseNode, RemainingProbNormalizedNodes
}
func Normalize_Common(probVal []Pnormalized) []Pnormalized {

	var total float64
	for _, item := range probVal {
		total += item.Value
	}

	var pnormalizedlist []Pnormalized
	for _, item := range probVal {

		var tmp Pnormalized
		tmp.Value = item.Value / total
		tmp.Addr = item.Addr
		pnormalizedlist = append(pnormalizedlist, tmp)
	}
	return pnormalizedlist
}

func Sample1NodesInValNodes_Common(probnormalized []Pnormalized, rand01 float64) (common.Address, bool) {

	for _, iterm := range probnormalized {
		rand01 -= iterm.Value
		if rand01 < 0 {
			return iterm.Addr, true
		}
	}
	return probnormalized[0].Addr, false
}
