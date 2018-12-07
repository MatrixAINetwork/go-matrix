// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package baseinterface

import (
	"github.com/matrix/go-matrix/common"
)

var (
	mapEntrust         = make(map[string]func() EntrustInterface)
	DefaultEntrustPlug = "secondKey"
)

type EntrustInterface interface {
	GetEntrustSignInfo(uint64) (common.Address, string, error)
	TransSignAccontToDeposit(common.Address, uint64) (common.Address, error)
}

func NewEntrust() EntrustInterface {
	return mapEntrust[DefaultEntrustPlug]()
}

func RegEntrust(name string, value func() EntrustInterface) {
	//fmt.Println("委托交易 注册函数", "name", name)
	mapEntrust[name] = value
}
