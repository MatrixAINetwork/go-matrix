// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package reelection

import (
	"fmt"
	"math/big"
	"testing"

	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/MatrixAINetwork/go-matrix/common"
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
	ans1, ans2 := GetAllElectedByHash(big.NewInt(100), common.RoleMiner)
	fmt.Println(ans1, ans2)
}

func TestNew1(t *testing.T) {
	info_temp := Info{
		Position: 1,
		Account:  common.BigToAddress(big.NewInt(100)),
	}
	info := []Info{}
	info = append(info, info_temp)
	node := NodeSupport{}
	node.First = append(node.First, info[0:]...)
	node.Second = append(node.Second, info[0:]...)
	node.Third = append(node.Third, info[0:]...)

	marshalData, err := json.Marshal(node)
	if err != nil {
		fmt.Println("测试支持", "Marshal失败 data", node)

	}
	err = ioutil.WriteFile("./test.json", marshalData, os.ModeAppend)
	if err != nil {
		fmt.Println("测试支持", "生成test文件成功")
	}

}

type JsonConfig struct {
	First  []Info
	Second []Info
	Third  []Info
}

type JsonStruct struct {
}

func NewJsonStruct() *JsonStruct {
	return &JsonStruct{}
}

func (jst *JsonStruct) Load(filename string, v interface{}) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Println("读取通用配置文件失败 err", err, "file", filename)
		os.Exit(-1)
		return
	}
	err = json.Unmarshal(data, v)
	if err != nil {
		log.Println("通用配置文件数据获取失败 err", err)
		os.Exit(-1)
		return
	}
}

func Test111(t *testing.T) {
	f, err := os.Open()
}
