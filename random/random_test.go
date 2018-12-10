// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package random

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/crypto"
	"github.com/matrix/go-matrix/event"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
)

func TestIsChanInFeed(t *testing.T) {
	seedSrv, err := NewElectionSeed()
	if err != nil {
		t.Error("新建随机种子对象失败")
	}
	log.Info("asd", seedSrv, err)
	ans := mc.IsChInFeed("ReElec_TopoSeedReq", seedSrv.randomSeedReqCh)
	fmt.Println("订阅成功后的状态", ans)
	seedSrv.randomSeedReqSub.Unsubscribe()
	ans = mc.IsChInFeed("ReElec_TopoSeedReq", seedSrv.randomSeedReqCh)
	fmt.Println("取消后的状态", ans)
	ans1, _ := mc.SubscribeEvent("ReElec_TopoSeedReq", seedSrv.randomSeedReqCh)
	ans = mc.IsChInFeed("ReElec_TopoSeedReq", seedSrv.randomSeedReqCh)
	fmt.Println("重新订阅后的状态", ans)

	var ch chan int
	ch = make(chan int, 10)
	ans = mc.IsChInFeed("ReElec_TopoSeedReq", ch)
	fmt.Println("查看其他通道的状态", ans)

	time.Sleep(100 * time.Second)
	fmt.Println(ans1)
}
func testRandomVote_1(t *testing.T) {
	randomvote, _ := newRandomVote()
	randomvote.currentRole = common.RoleValidator
	mc.PostEvent("CA_RoleUpdated", mc.RoleUpdatedMsg{Role: common.RoleMiner, BlockNum: 90})
	randomvote.roleUpdateSub.Unsubscribe()

}
func testRandomVote_2(t *testing.T) {
	randomvote, _ := newRandomVote()
	randomvote.currentRole = common.RoleValidator
	mc.PostEvent("CA_RoleUpdated", mc.RoleUpdatedMsg{Role: common.RoleMiner, BlockNum: 70})
	randomvote.roleUpdateSub.Unsubscribe()

}
func testRandomVote_3(t *testing.T) {
	randomvote, _ := newRandomVote()
	randomvote.currentRole = common.RoleValidator
	mc.PostEvent("CA_RoleUpdated", mc.RoleUpdatedMsg{Role: common.RoleValidator, BlockNum: 90})
	randomvote.roleUpdateSub.Unsubscribe()

}
func testRandomVote_4(t *testing.T) {
	randomvote, _ := newRandomVote()
	randomvote.currentRole = common.RoleValidator
	mc.PostEvent("CA_RoleUpdated", mc.RoleUpdatedMsg{Role: common.RoleValidator, BlockNum: 70})
	randomvote.roleUpdateSub.Unsubscribe()

}
func testRandomVote_5(t *testing.T) {
	randomvote, _ := newRandomVote()
	randomvote.currentRole = common.RoleMiner
	mc.PostEvent("CA_RoleUpdated", mc.RoleUpdatedMsg{Role: common.RoleValidator, BlockNum: 90})
	randomvote.roleUpdateSub.Unsubscribe()

}
func testRandomVote_6(t *testing.T) {
	randomvote, _ := newRandomVote()
	randomvote.currentRole = common.RoleMiner
	mc.PostEvent("CA_RoleUpdated", mc.RoleUpdatedMsg{Role: common.RoleValidator, BlockNum: 70})
	randomvote.roleUpdateSub.Unsubscribe()

}
func testRandomVote_7(t *testing.T) {
	randomvote, _ := newRandomVote()
	randomvote.currentRole = common.RoleMiner
	mc.PostEvent("CA_RoleUpdated", mc.RoleUpdatedMsg{Role: common.RoleMiner, BlockNum: 90})
	randomvote.roleUpdateSub.Unsubscribe()

}
func testRandomVote_8(t *testing.T) {
	randomvote, _ := newRandomVote()
	randomvote.currentRole = common.RoleMiner
	mc.PostEvent("CA_RoleUpdated", mc.RoleUpdatedMsg{Role: common.RoleMiner, BlockNum: 70})
	randomvote.roleUpdateSub.Unsubscribe()

}
func TestRandomVote(t *testing.T) {
	t.Run("testRandomVote_1", testRandomVote_1)
	time.Sleep(1 * time.Second)
	t.Run("testRandomVote_2", testRandomVote_2)
	time.Sleep(1 * time.Second)
	t.Run("testRandomVote_3", testRandomVote_3)
	time.Sleep(1 * time.Second)
	t.Run("testRandomVote_4", testRandomVote_4)
	time.Sleep(1 * time.Second)
	t.Run("testRandomVote_5", testRandomVote_5)
	time.Sleep(1 * time.Second)
	t.Run("testRandomVote_6", testRandomVote_6)
	time.Sleep(1 * time.Second)
	t.Run("testRandomVote_7", testRandomVote_7)
	time.Sleep(1 * time.Second)
	t.Run("testRandomVote_8", testRandomVote_8)
	time.Sleep(5 * time.Second)

}

//公用方法：
func bigToHash(data int) common.Hash {
	data_1 := big.NewInt(int64(data))
	return common.BigToHash(data_1)
}
func bigToBytes(data int) []byte {
	data_1 := bigToHash(data)
	return data_1.Bytes()
}

func testSeed_1(t *testing.T) {
	seedSrv, err := NewElectionSeed()
	if err != nil {
		t.Error("新建随机种子对象失败")
	}
	seedSrv.randomSeedReqSub.Unsubscribe()
	seedSrv.randomSeedReqSub, err = mc.SubscribeEvent("ReElec_TopoSeedReq_1", seedSrv.randomSeedReqCh)
	if err != mc.SubErrorNoThisEvent {
		t.Error("Need ", mc.SubErrorNoThisEvent, "is", err)
	}

}

func testSeed_2(t *testing.T) {
	ans := big.NewInt(100)
	err := mc.PostEvent("Random_TopoSeedRsp_1", mc.ElectionEvent{Seed: ans})
	if err != mc.PostErrorNoThisEvent {
		t.Error("Need", mc.PostErrorNoThisEvent, "is", err)
	}

}

func GetminHash(data int) common.Hash {
	return bigToHash(data)
}

func GetMap(data int, count int) (map[common.Address][]byte, map[common.Address][]byte, *big.Int) {
	prv := make(map[common.Address][]byte)
	pub := make(map[common.Address][]byte)
	temp := 400
	need := big.NewInt(int64(0))
	for i := 0; i < count; i++ {
		p1, p2, err := getkey()
		if err != nil {
			fmt.Println("得到公私钥失败,err", err)
		}
		//fmt.Println("私钥：", p1, p1.Bytes())
		address := common.BigToAddress(big.NewInt(int64(temp)))
		prv[address] = p1.Bytes()
		pub[address] = p2
		need.Add(need, common.BytesToHash(prv[address]).Big())
		temp += 100
	}

	return prv, pub, need

}

func testSeed_3(t *testing.T) {
	seedSrv, err := NewElectionSeed()
	if err != nil {
		t.Error("新建随机种子对象失败")
	}
	log.Info("asd", seedSrv, err)

	minHash := GetminHash(100)
	PrivateMap, PublicMap, need := GetMap(100, 2)

	var recvCh chan mc.ElectionEvent
	var recvSub event.Subscription
	recvCh = make(chan mc.ElectionEvent, 10)
	recvSub, _ = mc.SubscribeEvent("Random_TopoSeedRsp", recvCh)

	err = mc.PostEvent("ReElec_TopoSeedReq", mc.RandomRequest{MinHash: minHash, PrivateMap: PrivateMap, PublicMap: PublicMap})
	data := <-recvCh

	need.Add(need, minHash.Big())
	if need.Cmp(data.Seed) != 0 {
		t.Error("Need", need, "is", data.Seed)
	}
	recvSub.Unsubscribe()

}
func testSeed_4(t *testing.T) {
	seedSrv, err := NewElectionSeed()
	if err != nil {
		t.Error("新建随机种子对象失败")
	}
	log.Info("asd", seedSrv, err)

	minHash := GetminHash(100)
	PrivateMap, PublicMap, need := GetMap(100, 1)

	var recvCh chan mc.ElectionEvent
	var recvSub event.Subscription
	recvCh = make(chan mc.ElectionEvent, 10)
	recvSub, _ = mc.SubscribeEvent("Random_TopoSeedRsp", recvCh)

	err = mc.PostEvent("ReElec_TopoSeedReq", mc.RandomRequest{MinHash: minHash, PrivateMap: PrivateMap, PublicMap: PublicMap})
	data := <-recvCh

	need.Add(need, minHash.Big())
	if need.Cmp(data.Seed) != 0 {
		t.Error("Need", need, "is", data.Seed)
	}
	recvSub.Unsubscribe()

}
func TestSeed(t *testing.T) {

	/*
		t.Run("testSeed_1", testSeed_1)
		t.Run("testSeed_2", testSeed_2)
		t.Run("testSeed_3", testSeed_3)
		t.Run("testSeed_4", testSeed_4)

		ans:=big.NewInt(100)
		ans_1:=commjjkjjk
	*/
	fmt.Println(crypto.Keccak256Hash([]byte("sdfsfsd")).String())

}
