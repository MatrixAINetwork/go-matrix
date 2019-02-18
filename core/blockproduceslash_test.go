package core

import (
	"fmt"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/consensus/manash"
	"github.com/MatrixAINetwork/go-matrix/core/matrixstate"
	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/core/vm"
	"github.com/MatrixAINetwork/go-matrix/mandb"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/params"
	"github.com/MatrixAINetwork/go-matrix/params/manparams"
	"math/big"
	"testing"
)

func TestReadSlashParacheckCfg(t *testing.T) {
	diskdb := mandb.NewMemDatabase()
	blockchain, _ := NewBlockChain(diskdb, nil, &params.ChainConfig{}, manash.NewFaker(), vm.Config{})

	_, ok, err := blockchain.slashCfgProc(nil, 0)
	if err != ErrStatePtrIsNil {
		t.Errorf("state 指针为空检查错误", err)
	}
	if ok {
		t.Errorf("返回状态错误", err, "status", ok)
	}
}
func Test_shouldBlockProduceStatsStartCase0(t *testing.T) {
	diskdb := mandb.NewMemDatabase()
	blockchain, _ := NewBlockChain(diskdb, nil, &params.ChainConfig{}, manash.NewFaker(), vm.Config{})

	state, _ := state.New(common.Hash{}, state.NewDatabase(diskdb))
	matrixstate.SetVersionInfo(state, manparams.VersionAlpha)

	status, err := blockchain.shouldBlockProduceStatsStart(nil, common.Hash{}, &SlashCfg)
	if status != false || err != ErrStatePtrIsNil {
		t.Errorf("status", status, "error", err)
	}

	status, err = blockchain.shouldBlockProduceStatsStart(state, common.Hash{}, nil)
	if status != false || err != ErrSlashCfgPtrIsNil {
		t.Errorf("status", status, "error", err)
	}

	var slashCfg = mc.BlockProduceSlashCfg{Switcher: false, LowTHR: 1, ProhibitCycleNum: 2}
	status, err = blockchain.shouldBlockProduceStatsStart(state, common.Hash{}, &slashCfg)
	if status != false || err != nil {
		t.Errorf("status", status, "error", err)
	}

}
func Test_shouldBlockProduceStatsStartCase1(t *testing.T) {
	diskdb := mandb.NewMemDatabase()
	blockchain, _ := NewBlockChain(diskdb, nil, &params.ChainConfig{}, manash.NewFaker(), vm.Config{})
	var slashCfg = mc.BlockProduceSlashCfg{Switcher: true, LowTHR: 1, ProhibitCycleNum: 2}
	state, _ := state.New(common.Hash{}, state.NewDatabase(diskdb))
	matrixstate.SetVersionInfo(state, manparams.VersionAlpha)

	matrixstate.SetBlockProduceStatsStatus(state, &mc.BlockProduceSlashStatsStatus{Number: 0})
	status, err := blockchain.shouldBlockProduceStatsStart(state, common.Hash{}, &slashCfg)
	if status != true || err != nil {
		t.Errorf("status", status, "err", err)
	}
}
func Test_getSlashStatsList(t *testing.T) {
	diskdb := mandb.NewMemDatabase()
	_, _ = NewBlockChain(diskdb, nil, &params.ChainConfig{}, manash.NewFaker(), vm.Config{})

	state, _ := state.New(common.Hash{}, state.NewDatabase(diskdb))
	matrixstate.SetVersionInfo(state, manparams.VersionAlpha)
	statsList, err := getSlashStatsList(nil)

	if statsList != nil || err != ErrStatePtrIsNil {
		t.Errorf("statsList", statsList, "err", err)
	}
	var statList = mc.BlockProduceStats{}
	for i := 0; i < 10; i++ {
		statList.StatsList = append(statList.StatsList, mc.UserBlockProduceNum{Address: common.BytesToAddress([]byte{uint8(i)}), ProduceNum: uint16(100 - i)})
	}
	err = matrixstate.SetBlockProduceStats(state, &statList)
	if err != nil {
		t.Errorf("write err", err)
	}
	readStatsList, err := getSlashStatsList(state)
	if len(statList.StatsList) != len(readStatsList.StatsList) {
		t.Errorf("数据长度不一致", err)
		fmt.Println(statList)
		fmt.Println(readStatsList)
	} else {
		for i := 0; i < len(statList.StatsList); i++ {
			if !statList.StatsList[i].Address.Equal(readStatsList.StatsList[i].Address) || statList.StatsList[i].ProduceNum != readStatsList.StatsList[i].ProduceNum {
				t.Errorf("数据不一致", nil)
			}

		}
	}

}
func Test_getLatestInitStatsNum(t *testing.T) {
	diskdb := mandb.NewMemDatabase()
	_, _ = NewBlockChain(diskdb, nil, &params.ChainConfig{}, manash.NewFaker(), vm.Config{})

	_, err := getLatestInitStatsNum(nil)
	if err != ErrStatePtrIsNil {
		t.Errorf("输入指针空检查失败")
	}

	state, _ := state.New(common.Hash{}, state.NewDatabase(diskdb))
	matrixstate.SetVersionInfo(state, manparams.VersionAlpha)

	/*	_, err = getLatestInitStatsNum(state)
		if err == nil{
			t.Errorf("未检出错误")
		}*/
	var setNumber uint64 = 12345678
	matrixstate.SetBlockProduceStatsStatus(state, &mc.BlockProduceSlashStatsStatus{setNumber})
	readVal, err := getLatestInitStatsNum(state)
	if err != nil {
		t.Errorf("状态树读不出错误")
	}
	if readVal != setNumber {
		t.Errorf("数据读取错误", readVal)
	}
}
func Test_initStatsListCase0(t *testing.T) {
	diskdb := mandb.NewMemDatabase()
	NewBlockChain(diskdb, nil, &params.ChainConfig{}, manash.NewFaker(), vm.Config{})

	//空指针不处理
	initStatsList(nil)

	//没有初选列表的情况下，应该设置空的统计列表
	state, _ := state.New(common.Hash{}, state.NewDatabase(diskdb))
	matrixstate.SetVersionInfo(state, manparams.VersionAlpha)

	initStatsList(state)
	statsList, err := matrixstate.GetBlockProduceStats(state)
	if err != nil {
		t.Errorf("读取统计列表错误, %s", err)
	}
	if 0 != len(statsList.StatsList) {
		t.Errorf("读取数据错误")
	}
}
func Test_initStatsListCase1(t *testing.T) {
	diskdb := mandb.NewMemDatabase()
	NewBlockChain(diskdb, nil, &params.ChainConfig{}, manash.NewFaker(), vm.Config{})
	state, _ := state.New(common.Hash{}, state.NewDatabase(diskdb))
	matrixstate.SetVersionInfo(state, manparams.VersionAlpha)

	//创找初选列表
	var electList = mc.ElectGraph{}
	var recoderNum = 30
	for i := 0; i < recoderNum; i++ {
		node := mc.ElectNodeInfo{Account: common.BigToAddress(big.NewInt(int64(4*i + 0))), Position: 0 + uint16(i), Stock: 0, VIPLevel: 0, Type: common.RoleMiner}
		electList.ElectList = append(electList.ElectList, node)
		node = mc.ElectNodeInfo{Account: common.BigToAddress(big.NewInt(int64(4*i + 1))), Position: 0x1000 + uint16(i), Stock: 0, VIPLevel: 0, Type: common.RoleCandidateValidator}
		electList.ElectList = append(electList.ElectList, node)
		node = mc.ElectNodeInfo{Account: common.BigToAddress(big.NewInt(int64(4*i + 2))), Position: 0x2000 + uint16(i), Stock: 0, VIPLevel: 0, Type: common.RoleValidator}
		electList.ElectList = append(electList.ElectList, node)
		node = mc.ElectNodeInfo{Account: common.BigToAddress(big.NewInt(int64(4*i + 3))), Position: 0x3000 + uint16(i), Stock: 0, VIPLevel: 0, Type: common.RoleBackupValidator}
		electList.ElectList = append(electList.ElectList, node)
	}
	matrixstate.SetElectGraph(state, &electList)
	initStatsList(state)
	statsList, err := matrixstate.GetBlockProduceStats(state)
	if err != nil {
		t.Errorf("读取统计列表错误, %s", err)
	}

	if recoderNum != len(statsList.StatsList) {
		t.Errorf("读取数据错误")
	} else {
		for i := 0; i < recoderNum; i++ {
			if !statsList.StatsList[i].Address.Equal(common.BigToAddress(big.NewInt(int64(4*i+2)))) || statsList.StatsList[i].ProduceNum != 0 {
				t.Errorf("存储数据错误, %s, %d", statsList.StatsList[i].Address.String(), statsList.StatsList[i].ProduceNum)
			}
		}
	}

}
func chainInit(rootHash common.Hash) (*BlockChain, *state.StateDB) {
	diskdb := mandb.NewMemDatabase()
	blockchain, _ := NewBlockChain(diskdb, nil, &params.ChainConfig{}, manash.NewFaker(), vm.Config{})

	state, _ := state.New(common.Hash{}, state.NewDatabase(diskdb))
	matrixstate.SetVersionInfo(state, manparams.VersionAlpha)

	return blockchain, state
}
func slashstatsInit(state *state.StateDB) {
	var statList = mc.BlockProduceStats{}
	for i := 0; i < 10; i++ {
		statList.StatsList = append(statList.StatsList, mc.UserBlockProduceNum{Address: common.BytesToAddress([]byte{uint8(i)}), ProduceNum: uint16(100 - i)})
	}
	matrixstate.SetBlockProduceStats(state, &statList)
}

func Test_shouldAddRecorder(t *testing.T) {
	bc, state := chainInit(common.Hash{})
	slashCfg := &mc.BlockProduceSlashCfg{Switcher: false, LowTHR: 1, ProhibitCycleNum: 2}
	if list, status := bc.shouldAddRecorder(state, slashCfg); status != false || list != nil {
		t.Errorf("惩罚关闭下处理错误")
	}
	slashCfg = &mc.BlockProduceSlashCfg{Switcher: true, LowTHR: 1, ProhibitCycleNum: 2}
	/*	if list, status := bc.shouldAddRecorder(state, slashCfg); status!= false||len(list.StatsList) != 0{
		t.Errorf("未初始化统计列表下，处理错误")
		fmt.Println(list)
		fmt.Println(status)
	}*/
	slashstatsInit(state)
	slashCfg = &mc.BlockProduceSlashCfg{Switcher: true, LowTHR: 1, ProhibitCycleNum: 2}
	if _, status := bc.shouldAddRecorder(state, slashCfg); status != true {
		t.Errorf("未初始化统计列表下，处理错误")
	}
}
func statsListCompare(list0 *mc.BlockProduceStats, list1 *mc.BlockProduceStats) bool {
	if len(list0.StatsList) != len(list1.StatsList) {
		return false
	}

	for i := 0; i < len(list0.StatsList); i++ {
		if !list0.StatsList[i].Address.Equal(list1.StatsList[i].Address) {
			return false
		}
		if list0.StatsList[i].ProduceNum != list1.StatsList[i].ProduceNum {
			return false
		}
	}
	return true
}

func Test_statsListAddRecorderCase0(t *testing.T) {
	bc, state := chainInit(common.Hash{})
	//初始空列表
	preList := mc.BlockProduceStats{}
	matrixstate.SetBlockProduceStats(state, &preList)
	statsListAddRecorder(state, &preList, common.BigToAddress(big.NewInt(0)))
	slashCfg := mc.BlockProduceSlashCfg{true, 1, 2}
	if newList, status := bc.shouldAddRecorder(state, &slashCfg); status != false {
		if !statsListCompare(newList, &preList) {
			t.Errorf("Block Stats List Err")
		}
	} else {
		t.Errorf("Read Block Stats List Err")
	}
}
func Test_statsListAddRecorderCase1(t *testing.T) {
	bc, state := chainInit(common.Hash{})
	//初始空列表
	preList := mc.BlockProduceStats{}
	for i := 0; i < 100; i++ {
		preList.StatsList = append(preList.StatsList, mc.UserBlockProduceNum{common.BigToAddress(big.NewInt(int64(i))), 0})
	}

	matrixstate.SetBlockProduceStats(state, &preList)
	statsListAddRecorder(state, &preList, common.BigToAddress(big.NewInt(101)))
	slashCfg := mc.BlockProduceSlashCfg{true, 1, 2}
	if newList, status := bc.shouldAddRecorder(state, &slashCfg); status != false {
		if !statsListCompare(newList, &preList) {
			t.Errorf("Block Stats List Err")
		}
	} else {
		t.Errorf("Read Block Stats List Err")
	}
}
func Test_statsListAddRecorderCase2(t *testing.T) {
	bc, state := chainInit(common.Hash{})
	//初始空列表
	preList := mc.BlockProduceStats{}
	for i := 0; i < 100; i++ {
		preList.StatsList = append(preList.StatsList, mc.UserBlockProduceNum{common.BigToAddress(big.NewInt(int64(i))), 0})
	}

	matrixstate.SetBlockProduceStats(state, &preList)
	statsListAddRecorder(state, &preList, common.BigToAddress(big.NewInt(11)))
	slashCfg := mc.BlockProduceSlashCfg{true, 1, 2}
	preList.StatsList[11].ProduceNum = 1
	if newList, status := bc.shouldAddRecorder(state, &slashCfg); status != false {
		if !statsListCompare(newList, &preList) {
			t.Errorf("Block Stats List Err")
		}
	} else {
		t.Errorf("Read Block Stats List Err")
	}
}
func Test_getElectTimingCfg(t *testing.T) {
	_, state := chainInit(common.Hash{})
	//无配置下返回错误
	if _, err := getElectTimingCfg(state); err == nil {
		t.Errorf("未检查到错误")
	}
	var testTimmingCfg = mc.ElectGenTimeStruct{10, 11, 12, 13, 14}
	matrixstate.SetElectGenTime(state, &testTimmingCfg)
	if readCfg, err := getElectTimingCfg(state); err != nil {
		t.Errorf("未读取数据")
	} else {
		if *readCfg != testTimmingCfg {
			t.Errorf("读数据不符合预期")
		}
	}

}
func genTestBlackList(num int) *mc.BlockProduceSlashBlackList {
	var blackList = mc.BlockProduceSlashBlackList{}
	for i := 0; i < num; i++ {
		blackList.BlackList = append(blackList.BlackList, mc.UserBlockProduceSlash{common.BigToAddress(big.NewInt(int64(i))), uint16(i % 3)})
	}
	return &blackList
}
func blackListCompare(list0 *mc.BlockProduceSlashBlackList, list1 *mc.BlockProduceSlashBlackList) bool {
	if len(list0.BlackList) != len(list1.BlackList) {
		return false
	}

	for i := 0; i < len(list0.BlackList); i++ {
		if !list0.BlackList[i].Address.Equal(list1.BlackList[i].Address) {
			return false
		}
		if list0.BlackList[i].ProhibitCycleCounter != list1.BlackList[i].ProhibitCycleCounter {
			return false
		}
	}
	return true
}
func Test_GetBlackList(t *testing.T) {
	bc, state := chainInit(common.Hash{})
	if blacklist := bc.GetBlackList(state); blacklist == nil {
		t.Errorf("未输出列表")
	} else {
		if !blackListCompare(blacklist, &mc.BlockProduceSlashBlackList{}) {
			t.Errorf("黑名单不正确")
		}
	}

	expectedList := genTestBlackList(100)
	matrixstate.SetBlockProduceBlackList(state, expectedList)
	readBlackList := bc.GetBlackList(state)
	if !blackListCompare(readBlackList, expectedList) {
		t.Errorf("黑名单不正确")
	}
}
func Test_BlackListMaitainCase0(t *testing.T) {
	//lift ban Test
	blklist := genTestBlackList(100)
	handle := NewBlackListMaintain(blklist.BlackList)
	if len(handle.blacklist) == 0 {
		t.Errorf("add list wrong")
	}
	for _, v := range handle.blacklist {
		if 0 == v.ProhibitCycleCounter {
			t.Errorf("lift the ban wrong")
		}
	}
}
func Test_BlackListMaitainCase1(t *testing.T) {
	//test decrement
	var blackList = mc.BlockProduceSlashBlackList{}
	for i := 0; i < 100; i++ {
		blackList.BlackList = append(blackList.BlackList, mc.UserBlockProduceSlash{common.BigToAddress(big.NewInt(int64(i))), uint16(i%3) + 1})
	}
	handle := NewBlackListMaintain(blackList.BlackList)
	handle.CounterMaintain()
	for i := 0; i < 100; i++ {
		if handle.blacklist[i].ProhibitCycleCounter != uint16(i%3) {
			t.Errorf("self decrement err")
			break
		}
	}
}

func Test_BlackListAddBlackListCase0(t *testing.T) {
	slashCfg := mc.BlockProduceSlashCfg{Switcher: false, LowTHR: 2, ProhibitCycleNum: 10}
	var blackList = mc.BlockProduceSlashBlackList{}
	for i := 0; i < 100; i++ {
		blackList.BlackList = append(blackList.BlackList, mc.UserBlockProduceSlash{common.BigToAddress(big.NewInt(int64(i))), uint16(i%3) + 1})
	}
	handle := NewBlackListMaintain(blackList.BlackList)

	var statsList []mc.UserBlockProduceNum
	statsList = append(statsList, mc.UserBlockProduceNum{common.BigToAddress(big.NewInt(1)), 1})
	handle.AddBlackList(statsList, &slashCfg)
	//预期：关闭情况下，初始化黑名单
	if 0 != len(handle.blacklist) {
		t.Errorf("配置关闭情况下，未初始化黑名单")
	}
}

func Test_BlackListAddBlackListCase1(t *testing.T) {
	slashCfg := mc.BlockProduceSlashCfg{Switcher: true, LowTHR: 2, ProhibitCycleNum: 10}
	var blackList = mc.BlockProduceSlashBlackList{}
	for i := 0; i < 100; i++ {
		blackList.BlackList = append(blackList.BlackList, mc.UserBlockProduceSlash{common.BigToAddress(big.NewInt(int64(i))), uint16(i%3) + 1})
	}
	handle := NewBlackListMaintain(blackList.BlackList)

	var statsList []mc.UserBlockProduceNum
	statsList = append(statsList, mc.UserBlockProduceNum{common.BigToAddress(big.NewInt(1)), 1})
	handle.AddBlackList(statsList, &slashCfg)
	//预期：已存在黑名单中的节点，重置禁止值
	blackList.BlackList[1].ProhibitCycleCounter = 10

	for i := 0; i < 100; i++ {
		if handle.blacklist[i].ProhibitCycleCounter != blackList.BlackList[i].ProhibitCycleCounter {
			t.Errorf("self decrement err")
			break
		}
		if !handle.blacklist[i].Address.Equal(blackList.BlackList[i].Address) {
			t.Errorf("address decrement err")
			break
		}
	}
}
func Test_BlackListAddBlackListCase2(t *testing.T) {
	slashCfg := mc.BlockProduceSlashCfg{Switcher: true, LowTHR: 2, ProhibitCycleNum: 10}
	var blackList = mc.BlockProduceSlashBlackList{}
	for i := 0; i < 100; i++ {
		blackList.BlackList = append(blackList.BlackList, mc.UserBlockProduceSlash{common.BigToAddress(big.NewInt(int64(i))), uint16(i%3) + 1})
	}
	handle := NewBlackListMaintain(blackList.BlackList)

	var statsList []mc.UserBlockProduceNum
	statsList = append(statsList, mc.UserBlockProduceNum{common.BigToAddress(big.NewInt(100)), 1})
	handle.AddBlackList(statsList, &slashCfg)
	//预期：已存在黑名单中的节点，重置禁止值
	blackList.BlackList = append(blackList.BlackList, mc.UserBlockProduceSlash{common.BigToAddress(big.NewInt(100)), 10})

	for i := 0; i < 100; i++ {
		if handle.blacklist[i].ProhibitCycleCounter != blackList.BlackList[i].ProhibitCycleCounter {
			t.Errorf("self decrement err")
			break
		}
		if !handle.blacklist[i].Address.Equal(blackList.BlackList[i].Address) {
			t.Errorf("address decrement err")
			break
		}
	}
}
