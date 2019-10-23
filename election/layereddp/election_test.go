// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package layereddp

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"strconv"
	"testing"
	"time"

	"github.com/MatrixAINetwork/go-matrix/common/mt19937"
	"github.com/MatrixAINetwork/go-matrix/params/manversion"

	"github.com/MatrixAINetwork/go-matrix/baseinterface"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/matrixstate"
	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/core/vm"
	"github.com/MatrixAINetwork/go-matrix/election/support"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mandb"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/params/manparams"
	"github.com/pkg/errors"
)

type DepositInfo struct {
	Address     string
	SignAddress string
	Amount      uint64
}

type InInfo struct {
	Deposit           []DepositInfo
	SlashList         []SlashList
	BlackList         []string
	WhiteList         []string
	WhiteListSwitcher bool
	RandNum           uint64
	TurnsBuffer       TurnsBufferInfo
}
type TurnsBufferInfo struct {
	CandidatorList []string
	Number         uint64
	Seq            uint64
	MinerNum       uint64
}
type OutInfo struct {
	TurnsBuffer TurnsBufferInfo
	Miner       []DepositInfo
}

type SlashList struct {
	Account     string
	ProhibitNum uint16
}

type VectorInfo struct {
	IN  InInfo
	OUT OutInfo
}
type ElectInfo struct {
	PriodsNum int
	Vector    [][]VectorInfo
}

func init() {
	log.InitLog(0)
}

func genState() *state.StateDBManage {
	chaindb := mandb.NewMemDatabase()
	roots := make([]common.CoinRoot, 0)
	State, _ := state.NewStateDBManage(roots, chaindb, state.NewDatabase(chaindb))
	matrixstate.SetVersionInfo(State, manversion.VersionAIMine)
	matrixstate.SetBasePowerSlashCfg(State, &mc.BasePowerSlashCfg{Switcher: true, LowTHR: 2, ProhibitCycleNum: 10})
	return State
}

func MinerElectProcess(times int, num int, whiteList []common.Address, blackList []common.Address, bpslashblacklist *mc.BasePowerSlashBlackList) (bool, error) {

	MinerList := make([]vm.DepositDetail, 0)

	for i := 0; i < num; i++ {
		MinerList = append(MinerList, vm.DepositDetail{Address: common.HexToAddress(strconv.Itoa(i))})
	}
	data := &mc.MasterMinerReElectionReqMsg{
		RandSeed:    new(big.Int).SetUint64(uint64(time.Now().Unix())),
		ElectConfig: mc.ElectConfigInfo_All{MinerNum: 32, BlackList: blackList, WhiteList: whiteList, WhiteListSwitcher: len(whiteList) != 0},
		MinerList:   MinerList,
		SeqNum:      295,
	}
	currstate := genState()
	matrixstate.SetBasePowerBlackList(currstate, bpslashblacklist)

	for i := 0; i < times; i++ {
		data.SeqNum = uint64(i)
		preEleDpi, _ := matrixstate.GetElectDynamicPollingInfo(currstate)
		if i == 0 {
			preEleDpi.Seq = 1
		}
		if i == 1 {
			data.MinerList = data.MinerList[1:]

		}
		dpElect := support.NewDpElection(data, bpslashblacklist)
		dpElect.BpBlackList, _ = matrixstate.GetBasePowerBlackList(currstate)

		getnode := baseinterface.NewElect(manparams.ElectPlug_layerdDP).MinerTopGen(data, currstate)
		curEleDpi, wantchoosenode, start := getVerifyElect(currstate, preEleDpi, MinerList, data, dpElect)
		if status, err := compareResult(getnode, wantchoosenode, curEleDpi, start); !status {
			return status, err
		}
		log.Info("动态选举方案", "验证通过", "")
		//fmt.Println("")
	}

	/*	if status, err := MinerDataCmp(vEleRsp, cfg.Election); !status {
			return status, err
		}
		status, err := MinerSlashCmp(currstate, cfg)
	*/
	return true, nil
}

func MineGetUselist(times int, num int, whiteList []common.Address, blackList []common.Address, bpslashblacklist *mc.BasePowerSlashBlackList) (bool, error) {

	MinerList := make([]vm.DepositDetail, 0)
	candidatlist := make([]common.Address, 0)
	for i := 0; i < num; i++ {
		MinerList = append(MinerList, vm.DepositDetail{Address: common.HexToAddress(strconv.Itoa(i))})
		candidatlist = append(candidatlist, common.HexToAddress(strconv.Itoa(i)))
	}
	data := &mc.MasterMinerReElectionReqMsg{
		RandSeed:    new(big.Int).SetUint64(uint64(time.Now().Unix())),
		ElectConfig: mc.ElectConfigInfo_All{MinerNum: 32, BlackList: blackList, WhiteList: whiteList, WhiteListSwitcher: len(whiteList) != 0},
		MinerList:   MinerList,
		SeqNum:      295,
	}
	currstate := genState()
	matrixstate.SetBasePowerBlackList(currstate, bpslashblacklist)

	for i := 0; i < times; i++ {
		data.SeqNum = uint64(i)
		preEleDpi, _ := matrixstate.GetElectDynamicPollingInfo(currstate)
		if i == 0 {
			preEleDpi.Seq = 1
		}
		if i == 1 {
			data.MinerList = data.MinerList[1:]

		}
		dpElect := support.NewDpElection(data, bpslashblacklist)
		dpElect.BpBlackList, _ = matrixstate.GetBasePowerBlackList(currstate)
		s := time.Now().UnixNano()
		dpElect.OldFilterWhiteBlackList(candidatlist, nil)
		fmt.Println(time.Now().UnixNano() - s)

		s = time.Now().UnixNano()
		dpElect.NewFilterWhiteBlackList(candidatlist, nil)
		fmt.Println(time.Now().UnixNano() - s)
		/*		curEleDpi, wantchoosenode, start := getVerifyElect(currstate, preEleDpi, MinerList, data, dpElect)
				if status, err := compareResult(getnode, wantchoosenode, curEleDpi, start); !status {
					return status, err
				}*/
		log.Info("动态选举方案", "验证通过", "")
		//fmt.Println("")
	}

	/*	if status, err := MinerDataCmp(vEleRsp, cfg.Election); !status {
			return status, err
		}
		status, err := MinerSlashCmp(currstate, cfg)
	*/
	return true, nil
}

func getVerifyElect(currstate *state.StateDBManage, preEleDpi *mc.ElectDynamicPollingInfo, MinerList []vm.DepositDetail, data *mc.MasterMinerReElectionReqMsg, dpElect *support.ElectDP) (*mc.ElectDynamicPollingInfo, []common.Address, int) {
	curEleDpi, _ := matrixstate.GetElectDynamicPollingInfo(currstate)
	wantchoosenode := make([]common.Address, 0)
	if preEleDpi.MinerNum == 0 {
		preEleDpi.MinerNum = calcMinerNum(uint64(len(MinerList)), data.ElectConfig.MinerNum)
		addr := make([]common.Address, 0)
		for _, v := range data.MinerList {
			addr = append(addr, v.Address)
		}
		preEleDpi.CandidateList = addr
	}
	var start int
	if curEleDpi.Seq == 33 {
		log.Info("")
	}
	//preEleDpi.CandidateList = getAccountListInDeposit(data.MinerList, preEleDpi.CandidateList)
	if preEleDpi.Seq == curEleDpi.Seq && uint64(len(preEleDpi.CandidateList)) > preEleDpi.MinerNum {
		//没有轮次切换
		usableNodeList := dpElect.GetUsableNodeList(preEleDpi.CandidateList, nil)
		wantchoosenode = getwantnode(data, usableNodeList, preEleDpi)
	} else {
		if 0 == len(preEleDpi.CandidateList) {
			preEleDpi.MinerNum = calcMinerNum(uint64(len(MinerList)), data.ElectConfig.MinerNum)
			addr := make([]common.Address, 0)
			for _, v := range data.MinerList {
				addr = append(addr, v.Address)
			}
			preEleDpi.CandidateList = addr
		}
		usableNodeList := dpElect.GetUsableNodeList(preEleDpi.CandidateList, nil)
		if uint64(len(usableNodeList)) <= uint64(preEleDpi.MinerNum) {
			for i := 0; i < len(usableNodeList); i++ {
				wantchoosenode = append(wantchoosenode, usableNodeList[i])
			}
		} else {
			wantchoosenode = getwantnode(data, usableNodeList, preEleDpi)
		}
		start = len(wantchoosenode)
		preEleDpi.MinerNum = preEleDpi.MinerNum - uint64(len(wantchoosenode))
		addr := make([]common.Address, 0)
		//更新候选列表,上一轮当前选出来的节点暂时从抵押列表移除
		for _, v := range dpElect.DepositNode {
			if !support.FindAddress(v.Address, wantchoosenode) {
				addr = append(addr, v.Address)
			}
		}
		usableNodeList = dpElect.GetUsableNodeList(addr, nil)
		wantchoosenode = append(wantchoosenode, getwantnode(data, usableNodeList, preEleDpi)...)

	}
	return curEleDpi, wantchoosenode, start
}

func compareResult(getnode *mc.MasterMinerReElectionRsp, wantchoosenode []common.Address, curEleDpi *mc.ElectDynamicPollingInfo, start int) (bool, error) {
	for i, v := range getnode.MasterMiner {
		if !v.Account.Equal(wantchoosenode[i]) {
			return false, errors.Errorf("layeredDdp.getUsableNodeList() = %v, want %v", v.Account.String(), wantchoosenode[i].String())
		}
		if 0 == len(curEleDpi.CandidateList) {
			return true, nil
		}
		for _, v := range getnode.MasterMiner[start:] {
			if support.FindAddress(v.Account, curEleDpi.CandidateList) {
				return false, errors.Errorf("find choosenode = %v, in CandidateList", v.Account.String())
			}
		}

	}
	return true, nil
}
func getwantnode(data *mc.MasterMinerReElectionReqMsg, usableNodeList []common.Address, preEleDpi *mc.ElectDynamicPollingInfo) []common.Address {
	indexList := make([]uint64, 0)
	wantchoosenode := make([]common.Address, 0)
	randseed := mt19937.RandUniformInit(data.RandSeed.Int64())
	for i := 0; i < len(usableNodeList) && i < int(preEleDpi.MinerNum); i++ {

		randomData := uint64(randseed.Uniform(0, float64(^uint64(0))))
		index := randomData % (uint64(len(usableNodeList) - i))
		indexList = append(indexList, index)

	}
	for i := 0; i < len(indexList); i++ {
		wantchoosenode = append(wantchoosenode, usableNodeList[indexList[i]])
		usableNodeList = append(usableNodeList[:indexList[i]], usableNodeList[indexList[i]+1:]...)
	}
	return wantchoosenode
}
func TestMinerCase1(t *testing.T) {
	if status, err := MinerElectProcess(10, 1, nil, nil, &mc.BasePowerSlashBlackList{BlackList: []mc.BasePowerSlash{{Address: common.HexToAddress("10"), ProhibitCycleCounter: 0}}}); !status {
		t.Error(err)
	}
}
func TestMinerCase2(t *testing.T) {
	if status, err := MinerElectProcess(10, 5, nil, nil, &mc.BasePowerSlashBlackList{BlackList: []mc.BasePowerSlash{{Address: common.HexToAddress("10"), ProhibitCycleCounter: 0}}}); !status {
		t.Error(err)
	}
}
func TestMinerCase3(t *testing.T) {
	if status, err := MinerElectProcess(10, 31, nil, nil, &mc.BasePowerSlashBlackList{BlackList: []mc.BasePowerSlash{{Address: common.HexToAddress("10"), ProhibitCycleCounter: 0}}}); !status {
		t.Error(err)
	}
}
func TestMinerCase4(t *testing.T) {
	if status, err := MinerElectProcess(10, 32, nil, nil, &mc.BasePowerSlashBlackList{BlackList: []mc.BasePowerSlash{{Address: common.HexToAddress("10"), ProhibitCycleCounter: 0}}}); !status {
		t.Error(err)
	}
}
func TestMinerCase6(t *testing.T) {
	if status, err := MinerElectProcess(40, 1023, nil, nil, &mc.BasePowerSlashBlackList{BlackList: []mc.BasePowerSlash{{Address: common.HexToAddress("10"), ProhibitCycleCounter: 0}}}); !status {
		t.Error(err)
	}
}
func TestMinerCase7(t *testing.T) {
	if status, err := MinerElectProcess(10, 1024, nil, nil, &mc.BasePowerSlashBlackList{BlackList: []mc.BasePowerSlash{{Address: common.HexToAddress("10"), ProhibitCycleCounter: 0}}}); !status {
		t.Error(err)
	}
}
func TestMinerCase8(t *testing.T) {
	if status, err := MinerElectProcess(10, 1025, nil, nil, &mc.BasePowerSlashBlackList{BlackList: []mc.BasePowerSlash{{Address: common.HexToAddress("10"), ProhibitCycleCounter: 0}}}); !status {
		t.Error(err)
	}
}
func TestMinerCase9(t *testing.T) {
	if status, err := MinerElectProcess(10, 1087, nil, nil, &mc.BasePowerSlashBlackList{BlackList: []mc.BasePowerSlash{{Address: common.HexToAddress("10"), ProhibitCycleCounter: 0}}}); !status {
		t.Error(err)
	}
}
func TestMinerCase10(t *testing.T) {
	if status, err := MinerElectProcess(10, 1088, nil, nil, &mc.BasePowerSlashBlackList{BlackList: []mc.BasePowerSlash{{Address: common.HexToAddress("10"), ProhibitCycleCounter: 0}}}); !status {
		t.Error(err)
	}
}
func TestMinerCase11(t *testing.T) {
	if status, err := MinerElectProcess(10, 1089, nil, nil, &mc.BasePowerSlashBlackList{BlackList: []mc.BasePowerSlash{{Address: common.HexToAddress("10"), ProhibitCycleCounter: 0}}}); !status {
		t.Error(err)
	}
}

func TestMinerCase13(t *testing.T) {
	blacklist := make([]common.Address, 0)
	blacklist = append(blacklist, common.HexToAddress("01"))
	blacklist = append(blacklist, common.HexToAddress("08"))
	blacklist = append(blacklist, common.HexToAddress("11"))
	blacklist = append(blacklist, common.HexToAddress("23"))
	if status, err := MinerElectProcess(10, 1089, nil, blacklist, &mc.BasePowerSlashBlackList{BlackList: []mc.BasePowerSlash{{Address: common.HexToAddress("10"), ProhibitCycleCounter: 0}}}); !status {
		t.Error(err)
	}
}

func TestMinerCase16(t *testing.T) {
	blacklist := make([]common.Address, 0)
	blacklist = append(blacklist, common.HexToAddress("01"))
	blacklist = append(blacklist, common.HexToAddress("08"))
	blacklist = append(blacklist, common.HexToAddress("11"))
	blacklist = append(blacklist, common.HexToAddress("23"))
	if status, err := MinerElectProcess(10, 1089, nil, blacklist, &mc.BasePowerSlashBlackList{BlackList: []mc.BasePowerSlash{{Address: common.HexToAddress("01"), ProhibitCycleCounter: 10}}}); !status {
		t.Error(err)
	}
}
func TestMinerCase17(t *testing.T) {
	blacklist := make([]common.Address, 0)
	blacklist = append(blacklist, common.HexToAddress("01"))
	blacklist = append(blacklist, common.HexToAddress("08"))
	blacklist = append(blacklist, common.HexToAddress("11"))
	blacklist = append(blacklist, common.HexToAddress("23"))
	if status, err := MinerElectProcess(10, 1089, nil, blacklist, &mc.BasePowerSlashBlackList{BlackList: []mc.BasePowerSlash{{Address: common.HexToAddress("03"), ProhibitCycleCounter: 10}, {Address: common.HexToAddress("25"), ProhibitCycleCounter: 8}}}); !status {
		t.Error(err)
	}
}

func TestGetUseList(t *testing.T) {
	blacklist := make([]common.Address, 0)
	for i := 3000; i < 6000; i++ {
		blacklist = append(blacklist, common.HexToAddress(strconv.Itoa(i)))
	}
	if status, err := MineGetUselist(2, 3000, nil, blacklist, &mc.BasePowerSlashBlackList{BlackList: []mc.BasePowerSlash{{Address: common.HexToAddress("03"), ProhibitCycleCounter: 10}, {Address: common.HexToAddress("25"), ProhibitCycleCounter: 8}}}); !status {
		t.Error(err)
	}
}

func readTestCfg(path string, cfg *ElectInfo) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return errors.New("case not open case file")
	}
	if err := json.Unmarshal([]byte(data), cfg); err != nil {
		log.Error("", "", err)
		return errors.New("unmarshal case file failed")
	}
	return nil
}

func changeMatlabSlash(s []SlashList) *mc.BasePowerSlashBlackList {
	slashList := &mc.BasePowerSlashBlackList{make([]mc.BasePowerSlash, 0)}
	for i := 0; i < len(s); i++ {
		slashNode := mc.BasePowerSlash{common.HexToAddress(s[i].Account), s[i].ProhibitNum}
		slashList.BlackList = append(slashList.BlackList, slashNode)
	}
	return slashList
}

func changeMatlabAccount(list []string) []common.Address {
	var goAddress = make([]common.Address, 0)
	for _, v := range list {
		goAddress = append(goAddress, common.HexToAddress(v))
	}
	return goAddress
}
func changeMatlabDeposit(list []DepositInfo) []vm.DepositDetail {
	goList := make([]vm.DepositDetail, 0)
	for _, v := range list {
		deposit := new(big.Int).SetUint64(v.Amount)
		goList = append(goList, vm.DepositDetail{Address: common.HexToAddress(v.Address), SignAddress: common.HexToAddress(v.SignAddress), Deposit: deposit})
	}
	return goList
}

func minerDataCmp(goResult *mc.MasterMinerReElectionRsp, matlabResult []DepositInfo) (bool, error) {

	log.Trace("GO输出")
	for _, v := range goResult.MasterMiner {
		log.Trace("GO Result", "Address", v.Account.String(), "Vip", v.VIPLevel, "Stock", v.Stock, "Type", v.Type)
	}
	if len(goResult.MasterMiner) != len(matlabResult) {
		log.Error("输出结果长度不一致", "MatLab", len(matlabResult), "Go", len(goResult.MasterMiner))
		return false, errors.New("比较长度不一致")
	}

	for i := 0; i < len(goResult.MasterMiner); i++ {
		mdata := matlabResult[i]
		gdata := goResult.MasterMiner[i]
		if !common.HexToAddress(mdata.Address).Equal(gdata.Account) {
			log.Info("", "索引", i, "M Addr", common.HexToAddress(mdata.Address).String(), "GO Addr", gdata.Account.String())
			return false, errors.New("地址不一致")
		}
		/*		if mdata.VipRole != uint16(gdata.VIPLevel) {
				return false, errors.New("VIP不一致")
			}*/
	}
	return true, nil
}

func minerTurnsBufferCmp(statedb *state.StateDBManage, matlabResult TurnsBufferInfo) (bool, error) {
	dynmaicPolling, err := matrixstate.GetElectDynamicPollingInfo(statedb)
	if err != nil {
		return false, errors.New("读取状态树不正确")
	}
	if dynmaicPolling.Seq != matlabResult.Seq {
		return false, errors.New("比较序号不一致")
	}
	if dynmaicPolling.MinerNum != matlabResult.MinerNum {
		return false, errors.New("比较矿工选举数目不一致")
	}
	if len(dynmaicPolling.CandidateList) != len(matlabResult.CandidatorList) {
		return false, errors.New("比较矿工数目不一致")
	}
	for i, v := range dynmaicPolling.CandidateList {
		if !v.Equal(common.HexToAddress(matlabResult.CandidatorList[i])) {
			return false, errors.New("比较矿工账户不一致")
		}
	}
	return true, nil
}
func MinerElectFileProcess(vectorPath string) (bool, error) {
	cfg := &ElectInfo{}
	if err := readTestCfg(vectorPath, cfg); err != nil {
		return false, err
	}
	currstate := genState()
	matrixstate.SetElectDynamicPollingInfo(currstate, &mc.ElectDynamicPollingInfo{Seq: 0, MinerNum: 32})
	for i := 0; i < cfg.PriodsNum; i++ {
		MinerList := changeMatlabDeposit(cfg.Vector[0][i].IN.Deposit)
		blackList := changeMatlabAccount(cfg.Vector[0][i].IN.BlackList)
		whiteList := changeMatlabAccount(cfg.Vector[0][i].IN.WhiteList)
		slashList := changeMatlabSlash(cfg.Vector[0][i].IN.SlashList)

		matrixstate.SetBasePowerBlackList(currstate, slashList)

		data := &mc.MasterMinerReElectionReqMsg{
			RandSeed:    new(big.Int).SetUint64(cfg.Vector[0][i].IN.RandNum),
			ElectConfig: mc.ElectConfigInfo_All{MinerNum: 32, BlackList: blackList, WhiteList: whiteList, WhiteListSwitcher: cfg.Vector[0][i].IN.WhiteListSwitcher},
			MinerList:   MinerList,
			SeqNum:      uint64(i),
		}
		getnode := baseinterface.NewElect(manparams.ElectPlug_layerdDP).MinerTopGen(data, currstate)

		if status, err := minerDataCmp(getnode, cfg.Vector[0][i].OUT.Miner); !status {
			return status, err
		}
		if status, err := minerTurnsBufferCmp(currstate, cfg.Vector[0][i].OUT.TurnsBuffer); !status {
			return status, err
		}
	}

	return true, nil
}

func TestMinerCase18(t *testing.T) {
	if status, err := MinerElectFileProcess(".\\testdata\\case1.json"); !status {
		t.Error(err)
	}
}

func TestMinerCase19(t *testing.T) {
	if status, err := MinerElectFileProcess(".\\testdata\\case2.json"); !status {
		t.Error(err)
	}
}

func TestMinerCase20(t *testing.T) {
	if status, err := MinerElectFileProcess(".\\testdata\\case3.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase21(t *testing.T) {
	if status, err := MinerElectFileProcess(".\\testdata\\case4.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase22(t *testing.T) {
	if status, err := MinerElectFileProcess(".\\testdata\\case5.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase23(t *testing.T) {
	if status, err := MinerElectFileProcess(".\\testdata\\case6.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase24(t *testing.T) {
	if status, err := MinerElectFileProcess(".\\testdata\\case7.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase25(t *testing.T) {
	if status, err := MinerElectFileProcess(".\\testdata\\case8.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase26(t *testing.T) {
	if status, err := MinerElectFileProcess(".\\testdata\\case9.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase27(t *testing.T) {
	if status, err := MinerElectFileProcess(".\\testdata\\case10.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase28(t *testing.T) {
	if status, err := MinerElectFileProcess(".\\testdata\\case11.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase29(t *testing.T) {
	if status, err := MinerElectFileProcess(".\\testdata\\case12.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase30(t *testing.T) {
	if status, err := MinerElectFileProcess(".\\testdata\\case13.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase31(t *testing.T) {
	if status, err := MinerElectFileProcess(".\\testdata\\case14.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase32(t *testing.T) {
	if status, err := MinerElectFileProcess(".\\testdata\\case15.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase33(t *testing.T) {
	if status, err := MinerElectFileProcess(".\\testdata\\case16.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase34(t *testing.T) {
	if status, err := MinerElectFileProcess(".\\testdata\\case17.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase35(t *testing.T) {
	if status, err := MinerElectFileProcess(".\\testdata\\case18.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase36(t *testing.T) {
	if status, err := MinerElectFileProcess(".\\testdata\\case19.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase37(t *testing.T) {
	if status, err := MinerElectFileProcess(".\\testdata\\case20.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase38(t *testing.T) {
	if status, err := MinerElectFileProcess(".\\testdata\\case21.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase39(t *testing.T) {
	if status, err := MinerElectFileProcess(".\\testdata\\case22.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase40(t *testing.T) {
	if status, err := MinerElectFileProcess(".\\testdata\\case23.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase41(t *testing.T) {
	if status, err := MinerElectFileProcess(".\\testdata\\case24.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase42(t *testing.T) {
	if status, err := MinerElectFileProcess(".\\testdata\\case25.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase43(t *testing.T) {
	if status, err := MinerElectFileProcess(".\\testdata\\case26.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase44(t *testing.T) {
	if status, err := MinerElectFileProcess(".\\testdata\\case27.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase45(t *testing.T) {
	if status, err := MinerElectFileProcess(".\\testdata\\case28.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase46(t *testing.T) {
	if status, err := MinerElectFileProcess(".\\testdata\\case29.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase47(t *testing.T) {
	if status, err := MinerElectFileProcess(".\\testdata\\case30.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase48(t *testing.T) {
	if status, err := MinerElectFileProcess(".\\testdata\\case31.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase49(t *testing.T) {
	if status, err := MinerElectFileProcess(".\\testdata\\case32.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase50(t *testing.T) {
	if status, err := MinerElectFileProcess(".\\testdata\\case33.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase51(t *testing.T) {
	if status, err := MinerElectFileProcess(".\\testdata\\case34.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase52(t *testing.T) {
	if status, err := MinerElectFileProcess(".\\testdata\\case35.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase53(t *testing.T) {
	if status, err := MinerElectFileProcess(".\\testdata\\case36.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase54(t *testing.T) {
	if status, err := MinerElectFileProcess(".\\testdata\\case37.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase55(t *testing.T) {
	if status, err := MinerElectFileProcess(".\\testdata\\case38.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase56(t *testing.T) {
	if status, err := MinerElectFileProcess(".\\testdata\\case39.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase57(t *testing.T) {
	if status, err := MinerElectFileProcess(".\\testdata\\case40.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase58(t *testing.T) {
	if status, err := MinerElectFileProcess(".\\testdata\\case41.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase59(t *testing.T) {
	if status, err := MinerElectFileProcess(".\\testdata\\case42.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase60(t *testing.T) {
	if status, err := MinerElectFileProcess(".\\testdata\\case43.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase61(t *testing.T) {
	if status, err := MinerElectFileProcess(".\\testdata\\case44.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase62(t *testing.T) {
	if status, err := MinerElectFileProcess(".\\testdata\\case45.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase63(t *testing.T) {
	if status, err := MinerElectFileProcess(".\\testdata\\case46.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase64(t *testing.T) {
	if status, err := MinerElectFileProcess(".\\testdata\\case47.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase65(t *testing.T) {
	if status, err := MinerElectFileProcess(".\\testdata\\case48.json"); !status {
		t.Error(err)
	}
}

func TestMinerCase66(t *testing.T) {
	if status, err := MinerElectFileProcess(".\\testdata\\case49.json"); !status {
		t.Error(err)
	}
}
