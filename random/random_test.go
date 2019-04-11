// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package random

import (
	"testing"

	"github.com/MatrixAINetwork/go-matrix/consensus/ethash"
	"github.com/MatrixAINetwork/go-matrix/core"
	"github.com/MatrixAINetwork/go-matrix/params"
	. "github.com/smartystreets/goconvey/convey"

	"fmt"
	"math/big"
	"time"

	"bou.ke/monkey"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/random/commonsupport"
	"github.com/MatrixAINetwork/go-matrix/random/electionseed"
)

func Monkey_NeedVote() *monkey.PatchGuard {
	return monkey.Patch(electionseed.NeedVote, func(uint64) bool {
		return true
	})
}
func Monkey_GetKeyTransInfo() *monkey.PatchGuard {
	return monkey.Patch(commonsupport.GetKeyTransInfo, func(uint64, string) map[common.Address][]byte {
		ans := make(map[common.Address][]byte, 0)
		return ans
	})
}

func NewBlockChain(n int) *core.BlockChain {
	_, bc, err := core.NewCanonical(ethash.NewFaker(), n, true)
	if err != nil {
		fmt.Println("生成blockchain失败")
	}
	return bc

}

type TestEth struct {
}

func (self *TestEth) BlockChain() *core.BlockChain {
	return NewBlockChain(321)
}

func TestUnit0(t *testing.T) {
	//子服务的注册
	//因为子服务的注册时放在init出，所以执行永远在test文件执行之前，所以，不能在这里该参数，会晚
}

func TestUnit1(t *testing.T) {

}

func TestUnit2(t *testing.T) {
	log.InitLog(3)
	mapConfig := make(map[string]string)
	mapConfig["electionseed"] = "Minhash&Key"
	///mapConfig["everyblockseed"]="Nonce&Address&Coinbase"
	params.RandomConfig = mapConfig
	testEth := &TestEth{}
	random, err := New(testEth)
	Convey("初始化", t, func() {
		So(err, ShouldBeNil)
	})
	random.Stop()
	//time.Sleep(100*time.Second)
}

func TestUnit3(t *testing.T) {
	log.InitLog(3)
	mapConfig := make(map[string]string)
	mapConfig["electionseed"] = "Minhash&Key"
	mapConfig["everyblockseed"] = "Nonce&Address&Coinbase"
	params.RandomConfig = mapConfig
	testEth := &TestEth{}
	random, err := New(testEth)
	Convey("初始化", t, func() {
		So(err, ShouldBeNil)
	})
	random.Stop()
	//time.Sleep(100*time.Second)
}

func TestUnit4(t *testing.T) {
	log.InitLog(3)
	mapConfig := make(map[string]string)
	mapConfig["ss"] = "sss"
	mapConfig["cc"] = "ccc"
	mapConfig["ff"] = "fff"
	params.RandomConfig = mapConfig
	testEth := &TestEth{}
	random, err := New(testEth)
	Convey("初始化", t, func() {
		So(err, ShouldBeNil)
	})
	random.Stop()
}

func TestUnit5(t *testing.T) {
	log.InitLog(3)
	mapConfig := make(map[string]string)
	mapConfig["electionseed"] = "Minhash&Key"
	///mapConfig["everyblockseed"]="Nonce&Address&Coinbase"
	params.RandomConfig = mapConfig
	testEth := &TestEth{}
	random, err := New(testEth)
	Convey("初始化", t, func() {
		So(err, ShouldBeNil)
	})
	SendNum := 0
	go func() {
		for {
			time.Sleep(2 * time.Second)
			mc.PublishEvent(mc.CA_RoleUpdated, &mc.RoleUpdatedMsg{BlockNum: uint64(0), Leader: common.BigToAddress(big.NewInt(100)), Role: common.RoleMiner})
			if SendNum == 3 {
				fmt.Println("SendNum", SendNum)
				random.Stop()
			}
			SendNum++
		}
	}()
	time.Sleep(100 * time.Second)
}

//RandomServiceName = []string{"electionseed", "everyblockseed", "everybroadcastseed"}
func TestUnit6(t *testing.T) {
	log.InitLog(3)
	mapConfig := make(map[string]string)
	mapConfig["electionseed"] = "1"
	///mapConfig["everyblockseed"]="Nonce&Address&Coinbase"
	params.RandomConfig = mapConfig
	testEth := &TestEth{}
	_, err := New(testEth)
	fmt.Println(err)
	Convey("初始化", t, func() {
		needVote := Monkey_NeedVote()
		defer needVote.Unpatch()

		for i := 95; i < 1000; i += 100 {
			time.Sleep(2 * time.Second)
			mc.PublishEvent(mc.CA_RoleUpdated, &mc.RoleUpdatedMsg{BlockNum: uint64(i), Leader: common.BigToAddress(big.NewInt(100)), Role: common.RoleMiner})
		}
	})
	time.Sleep(100 * time.Second)
}

//RandomServiceName = []string{"electionseed", "everyblockseed", "everybroadcastseed"}
func TestUnit7(t *testing.T) {
	log.InitLog(3)
	mapConfig := make(map[string]string)
	mapConfig["everyblockseed"] = "2"
	///mapConfig["everyblockseed"]="Nonce&Address&Coinbase"
	params.RandomConfig = mapConfig
	testEth := &TestEth{}
	_, err := New(testEth)
	fmt.Println(err)
	Convey("初始化", t, func() {
		needVote := Monkey_NeedVote()
		defer needVote.Unpatch()

		for i := 90; i < 101; i++ {
			time.Sleep(2 * time.Second)
			mc.PublishEvent(mc.CA_RoleUpdated, &mc.RoleUpdatedMsg{BlockNum: uint64(i), Leader: common.BigToAddress(big.NewInt(100)), Role: common.RoleMiner})
		}
	})
	time.Sleep(100 * time.Second)
}

func TestUnit8(t *testing.T) {
	log.InitLog(3)
	mapConfig := make(map[string]string)
	mapConfig["everybroadcastseed"] = "2"

	params.RandomConfig = mapConfig
	testEth := &TestEth{}
	_, err := New(testEth)
	fmt.Println(err)

	for i := 90; i <= 101; i++ {
		time.Sleep(2 * time.Second)
		mc.PublishEvent(mc.CA_RoleUpdated, &mc.RoleUpdatedMsg{BlockNum: uint64(i), Leader: common.BigToAddress(big.NewInt(100)), Role: common.RoleMiner})
	}

	time.Sleep(100 * time.Second)
}

func TestUnit9(t *testing.T) {
	log.InitLog(3)
	mapConfig := make(map[string]string)
	mapConfig["electionseed"] = "Minhash&Key"
	mapConfig["everybroadcastseed"] = "2"
	mapConfig["everyblockseed"] = "2"

	params.RandomConfig = mapConfig
	testEth := &TestEth{}
	_, err := New(testEth)
	fmt.Println(err)

	Convey("初始化", t, func() {
		needVote := Monkey_NeedVote()
		defer needVote.Unpatch()
		for i := 90; i <= 101; i++ {
			time.Sleep(2 * time.Second)
			mc.PublishEvent(mc.CA_RoleUpdated, &mc.RoleUpdatedMsg{BlockNum: uint64(i), Leader: common.BigToAddress(big.NewInt(100)), Role: common.RoleMiner})
		}
	})

	time.Sleep(100 * time.Second)

}

func TestUnit10(t *testing.T) {
	//log.InitLog(3)
	mapConfig := make(map[string]string)
	mapConfig["electionseed"] = "Minhash&Key"
	mapConfig["everybroadcastseed"] = "2"
	mapConfig["everyblockseed"] = "2"

	params.RandomConfig = mapConfig
	testEth := &TestEth{}
	_, err := New(testEth)
	fmt.Println(err)
	fmt.Println(testEth.BlockChain().GetBlockByNumber(0).Header().Leader)
	fmt.Println(time.Now())
	ans := testEth.BlockChain().GetBlockByNumber(1).Hash().Big()
	fmt.Println(time.Now())
	ansi := 1

	fmt.Println(time.Now())
	for i := 1; i <= 100; i++ {
		temp := testEth.BlockChain().GetBlockByNumber(uint64(i)).Hash().Big()
		if ans.Cmp(temp) >= 0 {
			ans = temp
			ansi = i
		}
	}
	fmt.Println(time.Now())
	fmt.Println("---ans,", ans, "ansi", ansi)

	Convey("获取选举种子数据", t, func() {
		Getkey := Monkey_GetKeyTransInfo()
		defer Getkey.Unpatch()
		ans, err := GetRandom(100, "electionseed")
		fmt.Println("electionseed-100-------", ans, "err", err, "uint64", ans.Uint64())
	})

	fmt.Println("nonce", testEth.BlockChain().GetBlockByNumber(34).Header().Nonce, "leader", testEth.BlockChain().GetBlockByNumber(34).Header().Leader)
	Convey("获取每个块一个的随机数", t, func() {
		ans, err := GetRandom(34, "everyblockseed")
		fmt.Println("everyblockseed-34------", ans, "err", err)
	})

	Convey("每个广播区块一个的随机数", t, func() {
		Getkey := Monkey_GetKeyTransInfo()
		defer Getkey.Unpatch()

		fmt.Println("nonce", testEth.BlockChain().GetBlockByNumber(100).Header().Nonce.Uint64(), "leader", testEth.BlockChain().GetBlockByNumber(100).Header().Leader)
		ans, err := GetRandom(100, "everybroadcastseed")
		fmt.Println("everybroadcastseed-100---", ans, "err", err)
	})

}
