// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package reelection

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

//func Post() {
//	blockNum := 20
//	for {
//
//		err := mc.PostEvent("CA_RoleUpdated", mc.RoleUpdatedMsg{Role: common.RoleValidator, BlockNum: uint64(blockNum)})
//		blockNum++
//		//fmt.Println("CA_RoleUpdated", mc.RoleUpdatedMsg{Role: common.RoleValidator, BlockNum: uint64(blockNum)})
//		log.Info("err", err)
//		time.Sleep(5 * time.Second)
//
//	}
//}
//
//func TestReElect(t *testing.T) {
//
//	electseed, err := random.NewElectionSeed()
//
//	log.Info("electseed", electseed)
//	log.Info("seed err", err)
//
//	var eth *eth.Ethereum
//	reElect, err := New(eth)
//	log.Info("err", err)
//
//	go Post()
//
//	time.Sleep(10000 * time.Second)
//	time.Sleep(3 * time.Second)
//	ans1, ans2, ans3 := reElect.readElectData(common.RoleMiner, 240)
//	fmt.Println("READ ELECT", ans1, ans2, ans3)
//	fmt.Println("READ ELECT", 240)
//
//	fmt.Println(reElect)
//}

func TestT(t *testing.T) {
	ans := big.NewInt(100)
	ans1 := common.BigToHash(ans)
	fmt.Println(ans1)

}
func TestCase(t *testing.T) {
	ans1, ans2 := GetAllElectedByHeight(big.NewInt(100), common.RoleMiner)
	fmt.Println(ans1, ans2)
}
