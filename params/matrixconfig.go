// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package params

import (
	"encoding/json"
	"io/ioutil"

	"fmt"
	"github.com/matrix/go-matrix/log"
	"os"
)

const (
	VerifyNetChangeUpTime = 6
	MinerNetChangeUpTime  = 4

	VerifyTopologyGenerateUpTime = 8
	MinerTopologyGenerateUpTime  = 8

	RandomVoteTime = 5

	LRSParentMiningTime = int64(20)
	LRSPOSOutTime       = int64(20)
	LRSReelectOutTime   = int64(40)
	LRSReelectInterval  = 5

	VotePoolTimeout    = 55 * 1000
	VotePoolCountLimit = 5

	BlkPosReqSendInterval   = 5
	BlkPosReqSendTimes      = 6
	BlkVoteSendInterval     = 3
	BlkVoteSendTimes        = 8
	MinerReqSendInterval    = 3
	PosedReqSendInterval    = 10
	MinerResultSendInterval = 3
)

var (
	DifficultList = []uint64{1}
)

func Config_Init(Config_PATH string) {
	log.INFO("Config_Init 函数", "Config_PATH", Config_PATH)

	JsonParse := NewJsonStruct()
	v := Config{}
	JsonParse.Load(Config_PATH, &v)
	MainnetBootnodes = v.BootNode
	if len(MainnetBootnodes) <= 0 {
		fmt.Println("无bootnode节点")
		os.Exit(-1)
	}
	BroadCastNodes = v.BroadNode
	if len(BroadCastNodes) <= 0 {
		fmt.Println("无广播节点")
		os.Exit(-1)
	}
}

type Config struct {
	BootNode  []string
	BroadNode []BroadCastNode
}

type JsonStruct struct {
}

func NewJsonStruct() *JsonStruct {
	return &JsonStruct{}
}

func (jst *JsonStruct) Load(filename string, v interface{}) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("读取通用配置文件失败 err", err, "file", filename)
		os.Exit(-1)
		return
	}
	err = json.Unmarshal(data, v)
	if err != nil {
		fmt.Println("通用配置文件数据获取失败 err", err)
		os.Exit(-1)
		return
	}
}
