// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package signhelper

import (
	"errors"
	"github.com/matrix/go-matrix/baseinterface"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/log"
)

var (
	mode     = "empty"
	realData = common.HexToAddress("0xb7efab17215a43983d766114feb69172587a4090")
)

var (
	ErrorNoEntrustTrans   = errors.New("无该签名交易")
	ErrorNoDepositAccount = errors.New("无该抵押账户")
)

var (
	EntrustValue = make(map[common.Address]string, 0) //委托交易的账户密码

)

func init() {
	baseinterface.RegEntrust("secondKey", NewSecondKey)
}

func NewSecondKey() baseinterface.EntrustInterface {
	return &SecondKey{}
}

type SecondKey struct {
}

func (self *SecondKey) GetEntrustSignInfo(height uint64) (common.Address, string, error) {
	ans := GetEntrustSignFromStatusDB(height) //郑贺提供接口
	log.Info("签名助手", "获取到的账户", ans.String(), "高度", height)
	/*for k, v := range EntrustValue {
		fmt.Println("k", k.String(), "v", v)
	}*/
	if _, ok := EntrustValue[ans]; ok {
		return ans, EntrustValue[ans], nil
	}
	return common.Address{}, "", ErrorNoEntrustTrans
}

func (self *SecondKey) TransSignAccontToDeposit(signAccount common.Address, height uint64) (common.Address, error) {
	ans := GetDepositAccont(signAccount, height)
	if ans.Equal(common.Address{}) {
		return common.Address{}, ErrorNoDepositAccount
	}
	return ans, nil

}

/////////////////////////////////////////////////////////////////////
/*
	temp
	调用郑贺接口
*/
func GetEntrustSignFromStatusDB(height uint64) common.Address {
	if mode == "empty" {
		return common.Address{}
	}
	return realData
}

func GetDepositAccont(entrust common.Address, height uint64) common.Address {
	if mode == "empty" {
		return common.Address{}
	}
	return realData

}
