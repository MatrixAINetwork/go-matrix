// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package layeredBss

import (
	"encoding/json"
	"github.com/MatrixAINetwork/go-matrix/baseinterface"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/matrixstate"
	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/core/vm"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mandb"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/params/manparams"
	"github.com/pkg/errors"
	"io/ioutil"
	"math/big"
	"testing"
)

type VipCfg struct {
	Amount string
	Number uint64
}

type StringAccount struct {
	Account string
}
type NodeList struct {
	Address     string
	SignAddress string
	Account     string
}
type SlashList struct {
	Account     string
	ProhibitNum uint16
}
type Election struct {
	Account  string
	Position uint64
	Stock    uint16
	VipRole  uint16
	Amount   uint64
}
type ElectInfo struct {
	TopNodeNum        uint16
	BackUpNodeNum     uint16
	CandidateNodeNum  uint64
	VipCfg            []VipCfg
	Random            uint64
	NodeList          []NodeList
	SlashCfgList      []SlashList
	SlashUpdateList   []SlashList
	Election          []Election
	WhiteList         []StringAccount
	BlackList         []StringAccount
	WhiteListSwitcher bool
}

func init() {
	log.InitLog(3)
}

func genState() *state.StateDBManage {
	chaindb := mandb.NewMemDatabase()
	roots := make([]common.CoinRoot, 0)
	State, _ := state.NewStateDBManage(roots, chaindb, state.NewDatabase(chaindb))
	matrixstate.SetVersionInfo(State, manparams.VersionAlpha)
	return State
}

func readTestCfg(path string, cfg *ElectInfo) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return errors.New("case not open case file")
	}
	if err := json.Unmarshal([]byte(data), cfg); err != nil {
		log.ERROR("", "", err)
		return errors.New("unmarshal case file failed")
	}
	return nil
}
func changeMatlabSlash(s []SlashList) *mc.BlockProduceSlashBlackList {
	slashList := &mc.BlockProduceSlashBlackList{make([]mc.UserBlockProduceSlash, 0)}
	for i := 0; i < len(s); i++ {
		slashNode := mc.UserBlockProduceSlash{common.HexToAddress(s[i].Account), s[i].ProhibitNum}
		slashList.BlackList = append(slashList.BlackList, slashNode)
	}
	return slashList
}
func changeMatlabAccount(list []StringAccount) []common.Address {
	var goAddress = make([]common.Address, 0)
	for _, v := range list {
		goAddress = append(goAddress, common.HexToAddress(v.Account))
	}
	return goAddress
}
func changeMatlabDeposit(list []NodeList) []vm.DepositDetail {
	goList := make([]vm.DepositDetail, 0)
	for _, v := range list {
		deposit, _ := new(big.Int).SetString(v.Account, 0)
		goList = append(goList, vm.DepositDetail{Address: common.HexToAddress(v.Address), SignAddress: common.HexToAddress(v.SignAddress), Deposit: deposit})
	}
	return goList
}
func ValidatorElectProcess(vectorPath string) (bool, error) {
	cfg := &ElectInfo{}
	if err := readTestCfg(vectorPath, cfg); err != nil {
		return false, err
	}

	ValidatorList := changeMatlabDeposit(cfg.NodeList)
	blackList := changeMatlabAccount(cfg.BlackList)
	whiteList := changeMatlabAccount(cfg.WhiteList)
	slashList := changeMatlabSlash(cfg.SlashCfgList)
	data := &mc.MasterValidatorReElectionReqMsg{
		RandSeed:                new(big.Int).SetUint64(cfg.Random),
		ValidatorList:           ValidatorList,
		FoundationValidatorList: []vm.DepositDetail{},
		ElectConfig:             mc.ElectConfigInfo_All{ValidatorNum: cfg.TopNodeNum, BackValidator: cfg.BackUpNodeNum, BlackList: blackList, WhiteList: whiteList, WhiteListSwitcher: cfg.WhiteListSwitcher},
		BlockProduceBlackList:   *slashList,
	}
	currstate := genState()
	vEleRsp := baseinterface.NewElect("layerd_BSS").ValidatorTopGen(data, currstate)

	if status, err := validatorDataCmp(vEleRsp, cfg.Election); !status {
		return status, err
	}
	status, err := validatorSlashCmp(currstate, cfg)

	return status, err
}

func validatorSlashCmp(state *state.StateDBManage, config *ElectInfo) (bool, error) {
	matlabResult := changeMatlabSlash(config.SlashUpdateList)
	if updateSlash, err := matrixstate.GetBlockProduceBlackList(state); err != nil {
		return false, errors.New("读取状态树不正确")
	} else {
		if len(updateSlash.BlackList) != len(matlabResult.BlackList) {
			return false, errors.New("惩罚更新长度不一致")
		}
		for i := 0; i < len(updateSlash.BlackList); i++ {
			if !updateSlash.BlackList[i].Address.Equal(matlabResult.BlackList[i].Address) {
				return false, errors.New("黑名单地址不一致")
			}
			if updateSlash.BlackList[i].ProhibitCycleCounter != matlabResult.BlackList[i].ProhibitCycleCounter {
				log.Info(updateSlash.BlackList[i].Address.String(), "Go", updateSlash.BlackList[i].ProhibitCycleCounter, "Matlab", matlabResult.BlackList[i].ProhibitCycleCounter)
				return false, errors.New("惩罚更新不一致")
			}
		}
	}
	return true, nil
}
func validatorDataCmp(goResult *mc.MasterValidatorReElectionRsq, matlabResult []Election) (bool, error) {
	var reshape = make([]mc.ElectNodeInfo, 0)

	reshape = append(reshape, goResult.MasterValidator...)
	reshape = append(reshape, goResult.BackUpValidator...)
	reshape = append(reshape, goResult.CandidateValidator...)

	log.Trace("GO输出")
	for _, v := range reshape {
		log.Trace("GO Result", "Address", v.Account.String(), "Vip", v.VIPLevel, "Stock", v.Stock, "Type", v.Type)
	}
	if len(reshape) != len(matlabResult) {
		log.ERROR("输出结果长度不一致", "MatLab", len(matlabResult), "Go", len(reshape))
		return false, errors.New("比较长度不一致")
	}

	for i := 0; i < len(reshape); i++ {
		mdata := matlabResult[i]
		gdata := reshape[i]
		if !common.HexToAddress(mdata.Account).Equal(gdata.Account) {
			log.Info("", "索引", i, "M Addr", common.HexToAddress(mdata.Account).String(), "GO Addr", gdata.Account.String())
			return false, errors.New("地址不一致")
		}
		if mdata.Stock != gdata.Stock {
			log.Info("", "索引", i, "M Addr", common.HexToAddress(mdata.Account).String(), "GO Addr", gdata.Account.String(), "M Stock", mdata.Stock, "GO Stock", gdata.Stock)
			return false, errors.New("股权不一致")
		}

		/*		if mdata.VipRole != uint16(gdata.VIPLevel) {
				return false, errors.New("VIP不一致")
			}*/
	}
	return true, nil
}

func MinerElectProcess(vectorPath string) (bool, error) {
	cfg := new(ElectInfo)
	testdata, err := ioutil.ReadFile(vectorPath)
	if err != nil {
		return false, errors.New("case not open case file")
	}
	if err := json.Unmarshal([]byte(testdata), cfg); err != nil {
		return false, errors.New("unmarshal case file failed")
	}

	minerList := make([]vm.DepositDetail, 0)
	for _, v := range cfg.NodeList {
		deposit, _ := new(big.Int).SetString(v.Account, 0)
		minerList = append(minerList, vm.DepositDetail{Address: common.HexToAddress(v.Address), SignAddress: common.HexToAddress(v.SignAddress), Deposit: deposit})
	}
	Vip := make([]mc.VIPConfig, 0)
	Vip = append(Vip, mc.VIPConfig{MinMoney: 0, ElectUserNum: 0, StockScale: 1000})
	for i := 0; i < len(cfg.VipCfg); i++ {
		v := cfg.VipCfg[len(cfg.VipCfg)-i-1]
		minmoney, status := new(big.Int).SetString(v.Amount, 10)
		if !status {
			return false, errors.New("set vip cfg minmoney failed")
		}
		Vip = append(Vip, mc.VIPConfig{MinMoney: minmoney.Uint64(), ElectUserNum: uint8(v.Number), StockScale: 1000})
	}
	var blackList = make([]common.Address, 0)
	for _, v := range cfg.BlackList {
		blackList = append(blackList, common.HexToAddress(v.Account))
	}
	var whiteList = make([]common.Address, 0)
	for _, v := range cfg.WhiteList {
		whiteList = append(whiteList, common.HexToAddress(v.Account))
	}
	data := &mc.MasterMinerReElectionReqMsg{
		SeqNum:      0,
		RandSeed:    new(big.Int).SetUint64(cfg.Random),
		MinerList:   minerList,
		ElectConfig: mc.ElectConfigInfo_All{MinerNum: cfg.TopNodeNum, BlackList: blackList, WhiteList: whiteList, WhiteListSwitcher: cfg.WhiteListSwitcher},
	}
	ans := baseinterface.NewElect("layerd").MinerTopGen(data)
	status, err := minerDataCmp(ans, cfg.Election)
	return status, err
}

func minerDataCmp(goResult *mc.MasterMinerReElectionRsp, matlabResult []Election) (bool, error) {
	var reshape = make([]mc.ElectNodeInfo, 0)
	reshape = append(reshape, goResult.MasterMiner...)
	log.INFO("GO输出")
	for _, v := range reshape {
		log.INFO("GO Result", "Address", v.Account.String(), "Vip", v.VIPLevel, "Stock", v.Stock, "Type", v.Type)
	}
	if len(reshape) != len(matlabResult) {
		return false, errors.New("比较长度不一致")
	}

	for i := 0; i < len(reshape); i++ {
		mdata := matlabResult[i]
		gdata := reshape[i]
		if !common.HexToAddress(mdata.Account).Equal(gdata.Account) {
			log.INFO("", "索引", i, "M Addr", common.HexToAddress(mdata.Account).String(), "GO Addr", gdata.Account.String())
			return false, errors.New("地址不一致")
		}
		if mdata.Stock != gdata.Stock {
			log.INFO("", "索引", i, "M Addr", common.HexToAddress(mdata.Account).String(), "GO Addr", gdata.Account.String(), "M Stock", mdata.Stock, "GO Stock", gdata.Stock)
			return false, errors.New("股权不一致")
		}
	}
	return true, nil
}

func TestValidatorCase1(t *testing.T) {
	if status, err := ValidatorElectProcess(".\\testdata\\testvectorV\\case1.json"); !status {
		t.Error(err)
	}
}

func TestValidatorCase2(t *testing.T) {
	if status, err := ValidatorElectProcess(".\\testdata\\testvectorV\\case2.json"); !status {
		t.Error(err)
	}
}

func TestValidatorCase3(t *testing.T) {
	if status, err := ValidatorElectProcess(".\\testdata\\testvectorV\\case3.json"); !status {
		t.Error(err)
	}
}

func TestValidatorCase4(t *testing.T) {
	if status, err := ValidatorElectProcess(".\\testdata\\testvectorV\\case4.json"); !status {
		t.Error(err)
	}
}

func TestValidatorCase5(t *testing.T) {
	if status, err := ValidatorElectProcess(".\\testdata\\testvectorV\\case5.json"); !status {
		t.Error(err)
	}
}

func TestValidatorCase6(t *testing.T) {
	if status, err := ValidatorElectProcess(".\\testdata\\testvectorV\\case6.json"); !status {
		t.Error(err)
	}
}
func TestValidatorCase7(t *testing.T) {
	if status, err := ValidatorElectProcess(".\\testdata\\testvectorV\\case7.json"); !status {
		t.Error(err)
	}
}

func TestValidatorCase8(t *testing.T) {
	if status, err := ValidatorElectProcess(".\\testdata\\testvectorV\\case8.json"); !status {
		t.Error(err)
	}
}

func TestValidatorCase9(t *testing.T) {
	if status, err := ValidatorElectProcess(".\\testdata\\testvectorV\\case9.json"); !status {
		t.Error(err)
	}
}
func TestValidatorCase10(t *testing.T) {
	if status, err := ValidatorElectProcess(".\\testdata\\testvectorV\\case10.json"); !status {
		t.Error(err)
	}
}

func TestValidatorCase11(t *testing.T) {
	if status, err := ValidatorElectProcess(".\\testdata\\testvectorV\\case11.json"); !status {
		t.Error(err)
	}
}

func TestValidatorCase12(t *testing.T) {
	if status, err := ValidatorElectProcess(".\\testdata\\testvectorV\\case12.json"); !status {
		t.Error(err)
	}
}

func TestValidatorCase13(t *testing.T) {
	if status, err := ValidatorElectProcess(".\\testdata\\testvectorV\\case13.json"); !status {
		t.Error(err)
	}
}
func TestValidatorCase14(t *testing.T) {
	if status, err := ValidatorElectProcess(".\\testdata\\testvectorV\\case14.json"); !status {
		t.Error(err)
	}
}
func TestValidatorCase15(t *testing.T) {
	if status, err := ValidatorElectProcess(".\\testdata\\testvectorV\\case15.json"); !status {
		t.Error(err)
	}
}
func TestValidatorCase16(t *testing.T) {
	if status, err := ValidatorElectProcess(".\\testdata\\testvectorV\\case16.json"); !status {
		t.Error(err)
	}
}

func TestValidatorCase17(t *testing.T) {
	if status, err := ValidatorElectProcess(".\\testdata\\testvectorV\\case17.json"); !status {
		t.Error(err)
	}
}

func TestValidatorCase18(t *testing.T) {
	if status, err := ValidatorElectProcess(".\\testdata\\testvectorV\\case18.json"); !status {
		t.Error(err)
	}
}

func TestValidatorCase19(t *testing.T) {
	if status, err := ValidatorElectProcess(".\\testdata\\testvectorV\\case19.json"); !status {
		t.Error(err)
	}
}

func TestValidatorCase20(t *testing.T) {
	if status, err := ValidatorElectProcess(".\\testdata\\testvectorV\\case20.json"); !status {
		t.Error(err)
	}
}

func TestValidatorCase21(t *testing.T) {
	if status, err := ValidatorElectProcess(".\\testdata\\testvectorV\\case21.json"); !status {
		t.Error(err)
	}
}

func TestValidatorCase22(t *testing.T) {
	if status, err := ValidatorElectProcess(".\\testdata\\testvectorV\\case22.json"); !status {
		t.Error(err)
	}
}

func TestValidatorCase23(t *testing.T) {
	if status, err := ValidatorElectProcess(".\\testdata\\testvectorV\\case23.json"); !status {
		t.Error(err)
	}
}

func TestValidatorCase24(t *testing.T) {
	if status, err := ValidatorElectProcess(".\\testdata\\testvectorV\\case24.json"); !status {
		t.Error(err)
	}
}

func TestValidatorCase25(t *testing.T) {
	if status, err := ValidatorElectProcess(".\\testdata\\testvectorV\\case25.json"); !status {
		t.Error(err)
	}
}

func TestValidatorCase26(t *testing.T) {
	if status, err := ValidatorElectProcess(".\\testdata\\testvectorV\\case26.json"); !status {
		t.Error(err)
	}
}

func TestValidatorCase27(t *testing.T) {
	if status, err := ValidatorElectProcess(".\\testdata\\testvectorV\\case27.json"); !status {
		t.Error(err)
	}
}

func TestValidatorCase28(t *testing.T) {
	if status, err := ValidatorElectProcess(".\\testdata\\testvectorV\\case28.json"); !status {
		t.Error(err)
	}
}

func TestValidatorCase29(t *testing.T) {
	if status, err := ValidatorElectProcess(".\\testdata\\testvectorV\\case29.json"); !status {
		t.Error(err)
	}
}

func TestValidatorCase30(t *testing.T) {
	if status, err := ValidatorElectProcess(".\\testdata\\testvectorV\\case30.json"); !status {
		t.Error(err)
	}
}

func TestValidatorCase31(t *testing.T) {
	if status, err := ValidatorElectProcess(".\\testdata\\testvectorV\\case31.json"); !status {
		t.Error(err)
	}
}

func TestValidatorCase32(t *testing.T) {
	if status, err := ValidatorElectProcess(".\\testdata\\testvectorV\\case32.json"); !status {
		t.Error(err)
	}
}

func TestValidatorCase33(t *testing.T) {
	if status, err := ValidatorElectProcess(".\\testdata\\testvectorV\\case33.json"); !status {
		t.Error(err)
	}
}

func TestValidatorCase34(t *testing.T) {
	if status, err := ValidatorElectProcess(".\\testdata\\testvectorV\\case34.json"); !status {
		t.Error(err)
	}
}

func TestValidatorCase35(t *testing.T) {
	if status, err := ValidatorElectProcess(".\\testdata\\testvectorV\\case35.json"); !status {
		t.Error(err)
	}
}

func TestValidatorCase36(t *testing.T) {
	if status, err := ValidatorElectProcess(".\\testdata\\testvectorV\\case36.json"); !status {
		t.Error(err)
	}
}

func TestValidatorCase37(t *testing.T) {
	if status, err := ValidatorElectProcess(".\\testdata\\testvectorV\\case37.json"); !status {
		t.Error(err)
	}
}

func TestValidatorCase38(t *testing.T) {
	if status, err := ValidatorElectProcess(".\\testdata\\testvectorV\\case38.json"); !status {
		t.Error(err)
	}
}

func TestValidatorCase39(t *testing.T) {
	if status, err := ValidatorElectProcess(".\\testdata\\testvectorV\\case39.json"); !status {
		t.Error(err)
	}
}
