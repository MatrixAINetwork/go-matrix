//1543944015.210298
//1543943140.8189604
//1543942536.0560176
//1543941654.2538803
//1543940796.5100539
//1543940093.7242396
//1543939373.4027364
//1543938701.4985857
//1543937889.8427432
// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package manparams

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/p2p/discover"
	"github.com/matrix/go-matrix/params"
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

	MinerPickTimeout = 20
)

var (
	//随机数相关
	RandomConfig              = make(map[string]string, 0)   //man.json配置中读的
	RandomServiceName         = []string{}                   //子服务的名字
	RandomServicePlugs        = make(map[string][]string, 0) //子服务对应的插件名
	RandomServiceDefaultPlugs = make(map[string]string, 0)

	//选举相关
	ElectPlugs string
)

func init() {
	RandomServiceName = []string{"electionseed", "everyblockseed", "everybroadcastseed"}
	//RandomServiceName = []string{"electionseed", "everyblockseed"}
	RandomServicePlugs[RandomServiceName[0]] = []string{"Minhash&Key", "plug2"}
	RandomServicePlugs[RandomServiceName[1]] = []string{"Nonce&Address&Coinbase", "plug2"}
	RandomServicePlugs[RandomServiceName[2]] = []string{"MaxNonce&Key", "plug2"}

	RandomServiceDefaultPlugs[RandomServiceName[0]] = RandomServicePlugs[RandomServiceName[0]][0]
	RandomServiceDefaultPlugs[RandomServiceName[1]] = RandomServicePlugs[RandomServiceName[1]][0]
	RandomServiceDefaultPlugs[RandomServiceName[2]] = RandomServicePlugs[RandomServiceName[2]][0]
}

type NodeInfo struct {
	NodeID  discover.NodeID
	Address common.Address
}

var BroadCastNodes = []NodeInfo{}
var InnerMinerNodes = []NodeInfo{}
var FoundationNodes = []NodeInfo{}

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

	BroadCastNodes = v.BroadNode
	if len(BroadCastNodes) <= 0 {
		fmt.Println("无广播节点")
		os.Exit(-1)
	}
	log.INFO("BroadCastNode", "data", BroadCastNodes)

	InnerMinerNodes = v.InnerMinerNode
	if len(InnerMinerNodes) == 0 {
		log.Error("内部矿工节点个数为0", "读取man.json失败", "内部矿工节点个数为0")
	}
	log.INFO("InnerMinerNode:", "data", InnerMinerNodes)
	FoundationNodes = v.FoundationNode
	if len(FoundationNodes) == 0 {
		log.Error("基金会节点个数为0", "读取man.json失败", "基金会节点个数为0")
	}

	RandomConfig = v.RandomConfig
	log.INFO("RandomConfig", "data", RandomConfig)
	ElectPlugs = v.ElectPlugs
	log.INFO("ElectPlugs", "data", ElectPlugs)
	if v.BroadcastInterval <= 0 || v.ReelectionInterval <= 0 || v.BroadcastInterval >= v.ReelectionInterval {
		log.Error("广播区块高度和选举区块高度不正确或者尚未配置，将使用默认值 100 300")
		//os.Exit(-1)
	} else {
		common.SetBroadcastInterval(uint64(v.BroadcastInterval))
		common.SetReElectionInterval(uint64(v.ReelectionInterval))
		log.INFO("BroadcastInterval", "BroadcastInterval", common.GetBroadcastInterval())
		log.INFO("ReelectionInterval", "ReelectionInterval", common.GetReElectionInterval())
	}
	//fmt.Println("echeloc",v.Echelon)
	if len(v.Echelon)>0{

		common.EchelonArrary=v.Echelon
	}
	log.INFO("EchelonArrary","EchelonArrary",common.EchelonArrary)
}

type Config struct {
	BootNode           []string
	BroadNode          []NodeInfo
	InnerMinerNode     []NodeInfo
	FoundationNode     []NodeInfo
	RandomConfig       map[string]string
	ElectPlugs         string
	ReelectionInterval int
	BroadcastInterval int
	Echelon []common.Echelon
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
