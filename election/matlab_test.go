// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package election

import (
	"encoding/json"
	"github.com/MatrixAINetwork/go-matrix/baseinterface"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/vm"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/pkg/errors"
	"io/ioutil"
	"math/big"
	"testing"
)

type VipCfg struct {
	Amount string
	Number uint64
}
type NodeList struct {
	Address     string
	SignAddress string
	Account     string
}

type BlackList struct {
	Account string
}
type WhiteList struct {
	Account string
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
	BlackList         []BlackList
	Election          []Election
	WhiteList         []WhiteList
	WhiteListSwitcher bool
}

func init() {
	log.InitLog(3)
}
func ValidatorElectProcess(vectorPath string) (bool, error) {
	cfg := new(ElectInfo)
	testdata, err := ioutil.ReadFile(vectorPath)
	if err != nil {
		return false, errors.New("case not open case file")
	}
	if err := json.Unmarshal([]byte(testdata), cfg); err != nil {
		return false, errors.New("unmarshal case file failed")
	}

	ValidatorList := make([]vm.DepositDetail, 0)
	for _, v := range cfg.NodeList {
		deposit, _ := new(big.Int).SetString(v.Account, 0)
		ValidatorList = append(ValidatorList, vm.DepositDetail{Address: common.HexToAddress(v.Address), SignAddress: common.HexToAddress(v.SignAddress), Deposit: deposit})
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

	data := &mc.MasterValidatorReElectionReqMsg{
		SeqNum:                  0,
		RandSeed:                new(big.Int).SetUint64(cfg.Random),
		ValidatorList:           ValidatorList,
		FoundationValidatorList: []vm.DepositDetail{},
		ElectConfig:             mc.ElectConfigInfo_All{ValidatorNum: cfg.TopNodeNum, BackValidator: cfg.BackUpNodeNum, BlackList: blackList, WhiteList: whiteList, WhiteListSwitcher: cfg.WhiteListSwitcher},
		VIPList:                 Vip,
	}
	ans := baseinterface.NewElect("layerd").ValidatorTopGen(data, nil)
	status, err := validatorDataCmp(ans, cfg.Election)
	return status, err
}

func validatorDataCmp(goResult *mc.MasterValidatorReElectionRsq, matlabResult []Election) (bool, error) {
	var reshape = make([]mc.ElectNodeInfo, 0)

	reshape = append(reshape, goResult.MasterValidator...)
	reshape = append(reshape, goResult.BackUpValidator...)
	reshape = append(reshape, goResult.CandidateValidator...)

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

		if mdata.VipRole != uint16(gdata.VIPLevel) {
			return false, errors.New("VIP不一致")
		}
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
func TestMinerCase1(t *testing.T) {
	if status, err := MinerElectProcess(".\\testdata\\testvectorM\\case1.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase2(t *testing.T) {
	if status, err := MinerElectProcess(".\\testdata\\testvectorM\\case2.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase3(t *testing.T) {
	if status, err := MinerElectProcess(".\\testdata\\testvectorM\\case3.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase4(t *testing.T) {
	if status, err := MinerElectProcess(".\\testdata\\testvectorM\\case4.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase5(t *testing.T) {
	if status, err := MinerElectProcess(".\\testdata\\testvectorM\\case5.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase6(t *testing.T) {
	if status, err := MinerElectProcess(".\\testdata\\testvectorM\\case6.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase7(t *testing.T) {
	if status, err := MinerElectProcess(".\\testdata\\testvectorM\\case7.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase8(t *testing.T) {
	if status, err := MinerElectProcess(".\\testdata\\testvectorM\\case8.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase9(t *testing.T) {
	if status, err := MinerElectProcess(".\\testdata\\testvectorM\\case9.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase10(t *testing.T) {
	if status, err := MinerElectProcess(".\\testdata\\testvectorM\\case10.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase11(t *testing.T) {
	if status, err := MinerElectProcess(".\\testdata\\testvectorM\\case11.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase12(t *testing.T) {
	if status, err := MinerElectProcess(".\\testdata\\testvectorM\\case12.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase13(t *testing.T) {
	if status, err := MinerElectProcess(".\\testdata\\testvectorM\\case13.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase14(t *testing.T) {
	if status, err := MinerElectProcess(".\\testdata\\testvectorM\\case14.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase15(t *testing.T) {
	if status, err := MinerElectProcess(".\\testdata\\testvectorM\\case15.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase16(t *testing.T) {
	if status, err := MinerElectProcess(".\\testdata\\testvectorM\\case16.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase17(t *testing.T) {
	if status, err := MinerElectProcess(".\\testdata\\testvectorM\\case17.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase18(t *testing.T) {
	if status, err := MinerElectProcess(".\\testdata\\testvectorM\\case18.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase19(t *testing.T) {
	if status, err := MinerElectProcess(".\\testdata\\testvectorM\\case19.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase20(t *testing.T) {
	if status, err := MinerElectProcess(".\\testdata\\testvectorM\\case20.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase21(t *testing.T) {
	if status, err := MinerElectProcess(".\\testdata\\testvectorM\\case21.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase22(t *testing.T) {
	if status, err := MinerElectProcess(".\\testdata\\testvectorM\\case22.json"); !status {
		t.Error(err)
	}
}
func TestMinerCase23(t *testing.T) {
	if status, err := MinerElectProcess(".\\testdata\\testvectorM\\case23.json"); !status {
		t.Error(err)
	}
}
