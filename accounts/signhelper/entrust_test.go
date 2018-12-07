// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package signhelper

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/matrix/go-matrix/baseinterface"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/log"
)

func init() {
	EntrustValue[common.HexToAddress("0xb7efab17215a43983d766114feb69172587a4090")] = "111"
	EntrustValue[common.HexToAddress("0xf9e18acc86179925353713a4a5d0e9bf381fbc17")] = "222"
	EntrustValue[common.HexToAddress("0x992fcd5f39a298e58776a87441f5ee3319a101a0")] = "333"

}

func TestUnit1(t *testing.T) {
	//存在签名账户
	log.InitLog(3)
	ans := baseinterface.NewEntrust()
	mode = ""
	addr, pass, err := ans.GetEntrustSignInfo(uint64(100))
	fmt.Println("addr", addr, "pass", pass, "err", err)
}

func TestUnit2(t *testing.T) {
	//无签名账户
	log.InitLog(3)
	ans := baseinterface.NewEntrust()
	addr, pass, err := ans.GetEntrustSignInfo(uint64(100))
	fmt.Println("addr", addr, "pass", pass, "err", err)
}

func TestUnit3(t *testing.T) {
	//存在抵押账户
	log.InitLog(3)
	ans := baseinterface.NewEntrust()
	mode = ""
	addr, err := ans.TransSignAccontToDeposit(common.BigToAddress(big.NewInt(int64(100))), uint64(100))
	fmt.Println("addr", addr, "err", err)
}
func TestUnit4(t *testing.T) {

	//无抵押账户
	log.InitLog(3)
	ans := baseinterface.NewEntrust()
	addr, err := ans.TransSignAccontToDeposit(common.BigToAddress(big.NewInt(int64(100))), uint64(100))
	fmt.Println("addr", addr, "err", err)
}
