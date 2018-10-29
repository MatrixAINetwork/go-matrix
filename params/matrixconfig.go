// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package params

import (
	"encoding/json"
	"io/ioutil"

	"fmt"
	"github.com/matrix/go-matrix/log"
	"github.com/pkg/errors"
	"os"
)

const (
	VerifyNetChangeUpTime = 6
	MinerNetChangeUpTime  = 4

	VerifyTopologyGenerateUpTime = 8
	MinerTopologyGenerateUptime  = 8

	RandomVoteTime = 5
	HCSIM          = 1
	HCP2P          = 2
)

var (
	SignAccount         = "0xc47d9e507c1c5cb65cc7836bb668549fc8f547df"
	SignAccountPassword = "12345"
	HcMethod            = HCP2P

	NoBootNode      = errors.New("无boot节点")
	NoBroadCastNode = errors.New("无广播节点")
)

const (
	//TODO: VotePoolTimeout
	VotePoolTimeout    = 37 * 1000
	VotePoolCountLimit = 5
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
