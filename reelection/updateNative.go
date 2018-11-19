//1542653173.1288269
//1542652291.3975606
//1542651564.622206
//1542650718.8423526
//1542650078.1190655
//1542649339.3541903
//1542648543.282788
//1542647696.3529923
//1542646901.5332875
// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package reelection

import (
	"encoding/json"
	"math/big"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/mc"
)

func locate(address common.Address, master []mc.TopologyNodeInfo, backUp []mc.TopologyNodeInfo, cand []mc.TopologyNodeInfo) (int, mc.TopologyNodeInfo) {
	for _, v := range master {
		if v.Account == address {
			return 0, v
		}
	}
	for _, v := range backUp {
		if v.Account == address {
			return 1, v
		}
	}
	for _, v := range cand {
		if v.Account == address {
			return 2, v
		}
	}
	return -1, mc.TopologyNodeInfo{}
}



func (self *ReElection) CalOnline(diff common.NetTopology, top *mc.TopologyGraph) ([]common.Address, []common.Address) {
	online := make([]common.Address, 0)
	offline := make([]common.Address, 0)

	for _, v := range diff.NetTopologyData {

		if v.Position == 0xF000 {
			offline = append(offline, v.Account)
			continue
		}
		if v.Position == 0xF001 {
			online = append(online, v.Account)
			continue
		}
		nativeAdd := checkInGraph(top, v.Position)
		if checkInDiff(diff, nativeAdd) == false {
			offline = append(offline, nativeAdd)
		}

	}

	return online, offline
}
func checkInGraph(top *mc.TopologyGraph, pos uint16) common.Address {
	for _, v := range top.NodeList {
		if v.Position == pos {
			return v.Account
		}
	}
	return common.Address{}
}
func checkInDiff(diff common.NetTopology, add common.Address) bool {
	for _, v := range diff.NetTopologyData {
		if v.Account == add {
			return true
		}
	}
	return false
}
func IsInArray(aimAddress common.Address, offline []common.Address) bool {
	for _, v := range offline {
		if v == aimAddress {
			return true
		}
	}
	return false
}
func (self *ReElection) writeNativeData(height uint64, data AllNative) error {
	key := MakeNativeDBKey(height)
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	err = self.ldb.Put([]byte(key), jsonData, nil)
	return err
}

func (self *ReElection) readNativeData(height uint64) (AllNative, error) {
	key := MakeNativeDBKey(height)
	ans, err := self.ldb.Get([]byte(key), nil)
	if err != nil {
		return AllNative{}, err
	}
	var realAns AllNative
	err = json.Unmarshal(ans, &realAns)
	if err != nil {
		return AllNative{}, err
	}

	return realAns, nil

}
func MakeNativeDBKey(height uint64) string {
	t := big.NewInt(int64(height))
	ss := t.String() + "---" + "Native"
	return ss
}
