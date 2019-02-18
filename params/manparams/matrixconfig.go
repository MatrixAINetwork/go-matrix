// Copyright (c) 2018-2019 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package manparams

import (
	"encoding/json"
	"fmt"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/params"
	"io/ioutil"
	"os"
)

const (
	VotePoolTimeout    = 55 * 1000
	VotePoolCountLimit = 5

	BlkPosReqSendInterval   = 5
	BlkPosReqSendTimes      = 6
	BlkVoteSendInterval     = 3
	BlkVoteSendTimes        = 8
	MinerReqSendInterval    = 3
	PosedReqSendInterval    = 10
	MinerResultSendInterval = 3

	MinerPickTimeout = 20

	OnlineConsensusValidityTime = 5
)

const (
	ElectionSeed                         = "electionseed"
	ElectionSeed_Plug_MinHash            = "MinHash"
	EveryBlockSeed                       = "everyblockseed"
	EveryBlockSeed_Plug_NonceAndCoinbase = "NonceAndCoinbase"
	EveryBroadcastSeed                   = "everybroadcastseed"
	EveryBroadcastSeed_Plug_MaxNonce     = "MaxNonce"

	ElectPlug_layerd = "layerd"
	ElectPlug_stock  = "stock"
	ELectPlug_direct = "direct"
)

var (
	//随机数相关
	RandomConfig              = DefaultRandomConfig //man.json配置中读的
	RandomServiceName         = []string{ElectionSeed, EveryBlockSeed, EveryBroadcastSeed}
	RandomServicePlugs        = make(map[string][]string, 0) //子服务对应的插件名
	RandomServiceDefaultPlugs = make(map[string]string, 0)
)

func init() {
	RandomServicePlugs[RandomServiceName[0]] = []string{ElectionSeed_Plug_MinHash}
	RandomServicePlugs[RandomServiceName[1]] = []string{EveryBlockSeed_Plug_NonceAndCoinbase}
	RandomServicePlugs[RandomServiceName[2]] = []string{EveryBroadcastSeed_Plug_MaxNonce}

	RandomServiceDefaultPlugs[RandomServiceName[0]] = RandomServicePlugs[RandomServiceName[0]][0]
	RandomServiceDefaultPlugs[RandomServiceName[1]] = RandomServicePlugs[RandomServiceName[1]][0]
	RandomServiceDefaultPlugs[RandomServiceName[2]] = RandomServicePlugs[RandomServiceName[2]][0]
}

func Config_Init(Config_PATH string) {
	log.INFO("Config_Init 函数", "Config_PATH", Config_PATH)

	JsonParse := NewJsonStruct()
	v := Config{}
	JsonParse.Load(Config_PATH, &v)

	params.MainnetBootnodes = v.BootNode
	if len(params.MainnetBootnodes) <= 0 {
		fmt.Println("无bootnode节点")
		os.Exit(-1)
	}
	log.INFO("MainBootNode", "data", params.MainnetBootnodes)
}

type Config struct {
	BootNode []string
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
