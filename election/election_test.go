// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package election

import (
	"github.com/MatrixAINetwork/go-matrix/election/layered"
	"github.com/MatrixAINetwork/go-matrix/election/support"
	"testing"

	"fmt"
	"github.com/MatrixAINetwork/go-matrix/log"
	"math/big"

	"github.com/MatrixAINetwork/go-matrix/baseinterface"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/vm"
	_ "github.com/MatrixAINetwork/go-matrix/election/layered"
	_ "github.com/MatrixAINetwork/go-matrix/election/layeredmep"
	_ "github.com/MatrixAINetwork/go-matrix/election/nochoice"
	_ "github.com/MatrixAINetwork/go-matrix/election/stock"
	"github.com/MatrixAINetwork/go-matrix/mc"
)

func init() {
	log.InitLog(1)
}
func initElecCfg(usableList []bool) *support.Electoion {
	elechandle := &support.Electoion{}
	for i := 0; i < len(usableList); i++ {
		node := support.Node{Address: common.BigToAddress(big.NewInt(int64(i))), Usable: usableList[i]}
		elechandle.NodeList = append(elechandle.NodeList, node)
	}
	return elechandle
}
func filterCmp(electHandle *support.Electoion, expectedStatus []bool) bool {
	if len(electHandle.NodeList) != len(expectedStatus) {
		return false
	}
	for i := 0; i < len(electHandle.NodeList); i++ {
		if electHandle.NodeList[i].Usable != expectedStatus[i] {
			return false
		}
	}
	return true
}
func TestTryFilterBlockProduceBlackListCase0(t *testing.T) {
	//无黑名单下，不过滤
	var nodeList []bool = make([]bool, 0)
	for i := 0; i < 100; i++ {
		nodeList = append(nodeList, true)
	}
	elec := initElecCfg(nodeList)
	layered.TryFilterBlockProduceBlackList(elec, make([]mc.UserBlockProduceSlash, 0), 10)
	if !filterCmp(elec, nodeList) {
		t.Errorf("set usable err")
	}
}
func genSlashList(addressList []int) []mc.UserBlockProduceSlash {
	slashList := make([]mc.UserBlockProduceSlash, 0)
	for i := 0; i < len(addressList); i++ {
		slashNode := mc.UserBlockProduceSlash{}
		slashNode.Address = common.BigToAddress(big.NewInt(int64(addressList[i])))
		slashNode.ProhibitCycleCounter = 10
		slashList = append(slashList, slashNode)
	}
	return slashList
}

//黑名单节点已经不可用情况
func TestTryFilterBlockProduceBlackListCase1(t *testing.T) {
	//已被过滤的情况下，不再过滤
	var nodeList []bool = make([]bool, 0)
	for i := 0; i < 15; i++ {
		nodeList = append(nodeList, false)
	}
	for i := 15; i < 30; i++ {
		nodeList = append(nodeList, true)
	}
	elec := initElecCfg(nodeList)
	slashAdddrList := []int{1, 2, 3, 4, 5, 0}
	slashNodeList := genSlashList(slashAdddrList)
	layered.TryFilterBlockProduceBlackList(elec, slashNodeList, 0)
	if !filterCmp(elec, nodeList) {
		t.Errorf("set usable err")
	}
}

//可用数目限制，部分过滤case
func TestTryFilterBlockProduceBlackListCase2(t *testing.T) {
	//已被过滤的情况下，不再过滤
	var nodeList []bool = make([]bool, 0)
	for i := 0; i < 15; i++ {
		nodeList = append(nodeList, false)
	}
	for i := 15; i < 30; i++ {
		nodeList = append(nodeList, true)
	}
	elec := initElecCfg(nodeList)
	slashAdddrList := []int{15, 16, 17, 18, 19, 20, 21}
	slashNodeList := genSlashList(slashAdddrList)
	layered.TryFilterBlockProduceBlackList(elec, slashNodeList, 0)
	//预期全部修改
	for i := 15; i < 22; i++ {
		nodeList[i] = false
	}
	if !filterCmp(elec, nodeList) {
		t.Errorf("set usable err")
	}
}

//黑名单节点不存在抵押列表情况
func TestTryFilterBlockProduceBlackListCase3(t *testing.T) {
	//已被过滤的情况下，不再过滤
	var nodeList []bool = make([]bool, 0)
	for i := 0; i < 15; i++ {
		nodeList = append(nodeList, false)
	}
	for i := 15; i < 30; i++ {
		nodeList = append(nodeList, true)
	}
	elec := initElecCfg(nodeList)
	slashAdddrList := []int{45, 56, 67, 108, 219, 220, 231}
	slashNodeList := genSlashList(slashAdddrList)
	layered.TryFilterBlockProduceBlackList(elec, slashNodeList, 10)
	//预期只修改前5个
	if !filterCmp(elec, nodeList) {
		t.Errorf("set usable err")
	}
}

//可用节点充分，正常设置情况
func TestTryFilterBlockProduceBlackListCase4(t *testing.T) {
	//已被过滤的情况下，不再过滤
	var nodeList []bool = make([]bool, 0)
	for i := 0; i < 15; i++ {
		nodeList = append(nodeList, false)
	}
	for i := 15; i < 30; i++ {
		nodeList = append(nodeList, true)
	}
	elec := initElecCfg(nodeList)
	slashAdddrList := []int{15, 16}
	slashNodeList := genSlashList(slashAdddrList)
	layered.TryFilterBlockProduceBlackList(elec, slashNodeList, 10)
	nodeList[15] = false
	nodeList[16] = false
	//预期只修改前5个
	if !filterCmp(elec, nodeList) {
		t.Errorf("set usable err")
	}
}

//测试矿工选举随机性能
type candidate struct {
	Address common.Address
	Value   *big.Int
}

func prepareMinerElect(nodeList []candidate, randseed *big.Int, minerNum uint16, blackList []common.Address) *mc.MasterMinerReElectionReqMsg {
	minerElectReq := &mc.MasterMinerReElectionReqMsg{}
	minerElectReq.SeqNum = 1
	minerElectReq.RandSeed = randseed
	for _, v := range nodeList {
		minerElectReq.MinerList = append(minerElectReq.MinerList, vm.DepositDetail{Address: v.Address, Deposit: v.Value, Role: big.NewInt(common.RoleMiner)})
	}
	minerElectReq.ElectConfig.BlackList = blackList
	minerElectReq.ElectConfig.MinerNum = minerNum
	minerElectReq.ElectConfig.ElectPlug = "layerd"

	return minerElectReq
}

func genCandidateList(depositVal []*big.Int) []candidate {
	candidateList := make([]candidate, 0)
	for k, v := range depositVal {
		weiVal := big.NewInt(0)
		weiVal.Mul(v, big.NewInt(1000000000000000000))
		candidateList = append(candidateList, candidate{Address: common.BigToAddress(big.NewInt(int64(k))), Value: weiVal})
	}
	return candidateList
}
func printMinerElecRsq(msg *mc.MasterMinerReElectionRsp) {
	for _, v := range msg.MasterMiner {
		log.Error("MinerElect", "Miner", v)
	}
}

type statsNode struct {
	Stock *big.Int
	Value *big.Int
}
type stockstats struct {
	statsMap map[common.Address]statsNode
}

func (stats *stockstats) stockStatsInit(candidataList []candidate) {
	stats.statsMap = make(map[common.Address]statsNode)
	for _, v := range candidataList {
		if _, ok := stats.statsMap[v.Address]; ok {
			continue
		}
		stats.statsMap[v.Address] = statsNode{Stock: big.NewInt(0), Value: v.Value}
	}
}
func (stats *stockstats) print(extra string) {
	for _, v := range stats.statsMap {
		normalStock := big.NewInt(0)
		normalStock.Mul(v.Stock, big.NewInt(1E15))
		normalStock.Mul(normalStock, big.NewInt(1E7))
		normalRadio := big.NewInt(0)
		normalRadio.Div(normalStock, v.Value)
		log.Error(extra, "deposit", v.Value, "stock", v.Stock, "stock/deposit", normalRadio)
	}
}
func (stats *stockstats) stockStats(elecedList []mc.ElectNodeInfo) {
	for _, v := range elecedList {
		if storeInfo, ok := stats.statsMap[v.Account]; !ok {
			log.Error("VIP Miner测试", "未初始化账户", v.Account)
		} else {
			storeInfo.Stock = storeInfo.Stock.Add(storeInfo.Stock, big.NewInt(int64(v.Stock)))
			stats.statsMap[v.Account] = storeInfo
		}
	}
}

/*节点少， 等抵押，股权统计*/
func TestMinerElectRandomnessCase0(t *testing.T) {
	candidatesNum := 20
	electNum := 100
	minerNum := 5

	candidatesDeposit := make([]*big.Int, 0)
	for i := 0; i < candidatesNum; i++ {
		candidatesDeposit = append(candidatesDeposit, big.NewInt(10000))
	}
	candidateList := genCandidateList(candidatesDeposit)

	var stats stockstats
	stats.stockStatsInit(candidateList)

	for electLoop := 0; electLoop < electNum; electLoop++ {
		engine := baseinterface.NewElect("layerd")
		minerElectReq := prepareMinerElect(candidateList, big.NewInt(int64(electLoop)), uint16(minerNum), make([]common.Address, 0))
		electedMinerMsg := engine.MinerTopGen(minerElectReq)
		stats.stockStats(electedMinerMsg.MasterMiner)
	}

	extra := fmt.Sprint(candidatesNum, "抵押节点，", minerNum, "期望选举数", electNum, "采样，", "等额抵押统计")
	stats.print(extra)
}

/*节点少， 均匀抵押，股权统计*/
func TestMinerElectRandomnessCase1(t *testing.T) {
	candidatesNum := 20
	minerNum := 5
	electNum := 10000
	candidatesDeposit := make([]*big.Int, 0)
	for i := 0; i < candidatesNum; i++ {
		candidatesDeposit = append(candidatesDeposit, big.NewInt(int64(10000*(i+1))))
	}
	candidateList := genCandidateList(candidatesDeposit)

	var stats stockstats
	stats.stockStatsInit(candidateList)

	for electLoop := 0; electLoop < electNum; electLoop++ {
		engine := baseinterface.NewElect("layerd")
		minerElectReq := prepareMinerElect(candidateList, big.NewInt(int64(electLoop)), uint16(minerNum), make([]common.Address, 0))
		electedMinerMsg := engine.MinerTopGen(minerElectReq)
		stats.stockStats(electedMinerMsg.MasterMiner)
	}

	extra := fmt.Sprint(candidatesNum, "抵押节点，", minerNum, "期望选举数", electNum, "采样，", "线性抵押统计")
	stats.print(extra)
}

/*节点少， 非均匀抵押，股权统计*/
func TestMinerElectRandomnessCase2(t *testing.T) {
	minerNum := 5
	candidatesNum := 20
	electNum := 10000
	candidatesDeposit := make([]*big.Int, 0)
	for i := 0; i < candidatesNum/2; i++ {
		candidatesDeposit = append(candidatesDeposit, big.NewInt(int64(10000*(i+1))))
	}
	depositOffset := 5000000
	for i := candidatesNum / 2; i < candidatesNum; i++ {
		candidatesDeposit = append(candidatesDeposit, big.NewInt(int64(10000*(i+1)+depositOffset)))
	}
	candidateList := genCandidateList(candidatesDeposit)

	var stats stockstats
	stats.stockStatsInit(candidateList)

	for electLoop := 0; electLoop < electNum; electLoop++ {
		engine := baseinterface.NewElect("layerd")
		minerElectReq := prepareMinerElect(candidateList, big.NewInt(int64(electLoop)), uint16(minerNum), make([]common.Address, 0))
		electedMinerMsg := engine.MinerTopGen(minerElectReq)
		stats.stockStats(electedMinerMsg.MasterMiner)
	}

	extra := fmt.Sprint(candidatesNum, "抵押节点，", minerNum, "期望选举数", electNum, "采样，", "非线性抵押统计")
	stats.print(extra)
}

/*节点多， 等抵押，股权统计*/
func TestMinerElectRandomnessCase3(t *testing.T) {
	candidatesNum := 200
	electNum := 10000
	minerNum := 21

	candidatesDeposit := make([]*big.Int, 0)
	for i := 0; i < candidatesNum; i++ {
		candidatesDeposit = append(candidatesDeposit, big.NewInt(10000))
	}
	candidateList := genCandidateList(candidatesDeposit)

	var stats stockstats
	stats.stockStatsInit(candidateList)

	for electLoop := 0; electLoop < electNum; electLoop++ {
		engine := baseinterface.NewElect("layerd")
		minerElectReq := prepareMinerElect(candidateList, big.NewInt(int64(electLoop)), uint16(minerNum), make([]common.Address, 0))
		electedMinerMsg := engine.MinerTopGen(minerElectReq)
		stats.stockStats(electedMinerMsg.MasterMiner)
	}

	extra := fmt.Sprint(candidatesNum, "抵押节点，", minerNum, "期望选举数", electNum, "采样，", "等额抵押")
	stats.print(extra)
}

/*节点多， 均匀抵押，股权统计*/
func TestMinerElectRandomnessCase4(t *testing.T) {
	candidatesNum := 200
	electNum := 10000
	minerNum := 21
	candidatesDeposit := make([]*big.Int, 0)
	for i := 0; i < candidatesNum; i++ {
		candidatesDeposit = append(candidatesDeposit, big.NewInt(int64(10000*(i+1))))
	}
	candidateList := genCandidateList(candidatesDeposit)

	var stats stockstats
	stats.stockStatsInit(candidateList)

	for electLoop := 0; electLoop < electNum; electLoop++ {
		engine := baseinterface.NewElect("layerd")
		minerElectReq := prepareMinerElect(candidateList, big.NewInt(int64(electLoop)), uint16(minerNum), make([]common.Address, 0))
		electedMinerMsg := engine.MinerTopGen(minerElectReq)
		stats.stockStats(electedMinerMsg.MasterMiner)
	}

	extra := fmt.Sprint(candidatesNum, "抵押节点，", minerNum, "期望选举数", electNum, "采样，", "线性抵押统计")
	stats.print(extra)
}

/*节点多， 非均匀抵押，股权统计*/
func TestMinerElectRandomnessCase5(t *testing.T) {
	candidatesNum := 200
	electNum := 10000
	minerNum := 21
	candidatesDeposit := make([]*big.Int, 0)
	for i := 0; i < candidatesNum/2; i++ {
		candidatesDeposit = append(candidatesDeposit, big.NewInt(int64(10000*(i+1))))
	}
	depositOffset := 5000000
	for i := candidatesNum / 2; i < candidatesNum; i++ {
		candidatesDeposit = append(candidatesDeposit, big.NewInt(int64(10000*(i+1)+depositOffset)))
	}
	candidateList := genCandidateList(candidatesDeposit)

	var stats stockstats
	stats.stockStatsInit(candidateList)

	for electLoop := 0; electLoop < electNum; electLoop++ {
		engine := baseinterface.NewElect("layerd")
		minerElectReq := prepareMinerElect(candidateList, big.NewInt(int64(electLoop)), uint16(minerNum), make([]common.Address, 0))
		electedMinerMsg := engine.MinerTopGen(minerElectReq)
		stats.stockStats(electedMinerMsg.MasterMiner)
	}

	extra := fmt.Sprint(candidatesNum, "抵押节点，", minerNum, "期望选举数", electNum, "采样，", "非线性抵押统计")
	stats.print(extra)
}

func TestMEPStats(t *testing.T) {
	cfg := new(ElectInfo)
	var elecstats = make(map[string]int)
	minerList := make([]vm.DepositDetail, 0)
	for i := 0; i < 100; i++ {
		address := fmt.Sprint(i)
		elecstats[common.HexToAddress(address).String()] = 0
		deposit := big.NewInt(0).Mul(big.NewInt(1000000000000000000), big.NewInt(int64(i)))
		minerList = append(minerList, vm.DepositDetail{Address: common.HexToAddress(address), SignAddress: common.HexToAddress(address), Deposit: deposit})
	}
	Vip := make([]mc.VIPConfig, 0)
	Vip = append(Vip, mc.VIPConfig{MinMoney: 0, ElectUserNum: 0, StockScale: 1000})
	var blackList = make([]common.Address, 0)
	var whiteList = make([]common.Address, 0)
	for i := 0; i < 1000000; i++ {
		data := &mc.MasterMinerReElectionReqMsg{
			SeqNum:      0,
			RandSeed:    new(big.Int).SetUint64(uint64(i)),
			MinerList:   minerList,
			ElectConfig: mc.ElectConfigInfo_All{MinerNum: 21, BlackList: blackList, WhiteList: whiteList, WhiteListSwitcher: cfg.WhiteListSwitcher},
		}
		ans := baseinterface.NewElect("layerd_MEP").MinerTopGen(data)
		if len(ans.MasterMiner) != 21 {
			t.Error("MinerNum Elect To less: %d", len(ans.MasterMiner))
			t.FailNow()
		}
		for _, v := range ans.MasterMiner {
			if v.Stock != 1 {
				t.Error("Elec Stock Err: %d", v.Stock)
				t.FailNow()
			}
			if _, ok := elecstats[v.Account.String()]; ok {
				elecstats[v.Account.String()] = elecstats[v.Account.String()] + 1
			} else {
				t.Error("Invalid Account: ", v.Account.String())
				t.FailNow()
			}
		}
	}
	for k, v := range elecstats {
		fmt.Println("Account", k, "HitNum", v, "Prob", float64(v)/10000)
	}
}
