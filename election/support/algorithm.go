// Copyright (c) 2018-2019 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package support

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/common/mt19937"
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
		if !status{
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

	return common.Address{},false
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
		if !status{
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
