// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package support

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
)

func ToPoUpdate(Q0, Q1, Q2 []mc.TopologyNodeInfo, nettopo mc.TopologyGraph, offline []common.Address) []mc.Alternative {

	log.Info("Elector ToPoUpdate")
	netmap := make(map[common.Address]mc.TopologyNodeInfo)
	Q0map := make(map[common.Address]mc.TopologyNodeInfo)
	Q1map := make(map[common.Address]mc.TopologyNodeInfo)
	Q2map := make(map[common.Address]mc.TopologyNodeInfo)

	for _, item := range Q0 {
		Q0map[item.Account] = item
	}

	for _, item := range Q1 {
		Q1map[item.Account] = item
	}

	for _, item := range Q2 {
		Q2map[item.Account] = item
	}

	for _, item := range nettopo.NodeList {
		netmap[item.Account] = item
	}

	for _, item := range nettopo.NodeList {
		var ok bool
		_, ok = Q0map[item.Account]
		if ok == true {
			delete(Q0map, item.Account)
		}

		_, ok = Q1map[item.Account]
		if ok == true {
			delete(Q1map, item.Account)
		}

		_, ok = Q2map[item.Account]
		if ok == true {
			delete(Q2map, item.Account)
		}
	}

	var substitute []mc.TopologyNodeInfo

	for _, v := range Q0map {
		substitute = append(substitute, v)
	}
	for _, v := range Q1map {
		substitute = append(substitute, v)
	}
	for _, v := range Q2map {
		substitute = append(substitute, v)
	}
	var sublen = len(substitute)

	var alternalist []mc.Alternative

	for index, item := range offline {
		if index < sublen {
			tmp := netmap[item]
			var talt mc.Alternative
			talt.B = item
			talt.A = substitute[index].Account
			talt.Position = tmp.Position

			alternalist = append(alternalist, talt)
		} else {
			break
		}
	}
	return alternalist
}

func PrimarylistUpdate(Q0, Q1, Q2 []mc.TopologyNodeInfo, online mc.TopologyNodeInfo, flag int) ([]mc.TopologyNodeInfo, []mc.TopologyNodeInfo, []mc.TopologyNodeInfo) {
	log.Info("Elector PrimarylistUpdate")
	if flag == 0 {
		var tQ0 []mc.TopologyNodeInfo
		tQ0 = append(tQ0, online)
		tQ0 = append(tQ0, Q0...)
		Q0 = tQ0
	}

	if flag == 1 {
		var tQ1 []mc.TopologyNodeInfo
		tQ1 = append(tQ1, Q1...)
		tQ1 = append(tQ1, online)
		Q1 = tQ1
	}

	if flag == 2 {
		var tQ2 []mc.TopologyNodeInfo
		tQ2 = append(tQ2, Q2...)
		tQ2 = append(tQ2, online)
		Q2 = tQ2
	}
	return Q0, Q1, Q2
}
