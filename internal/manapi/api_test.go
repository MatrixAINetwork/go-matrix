// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package manapi

import (
	"encoding/json"
	"github.com/MatrixAINetwork/go-matrix/common/hexutil"
	"math/big"
	"strings"
	"testing"
)

func TestSendArgsMarshal(t *testing.T) {
	aaa := SendTxArgs1{From: "mmamamamam", To: new(string)}
	*aaa.To = "bbbb"
	aaa.GasPrice = (*hexutil.Big)(big.NewInt(1000000000000))
	buff, err := json.Marshal(aaa)
	if err != nil {
		t.Error(err)
	}
	str := string(buff)
	str1 := strings.Replace(str, "0xe8d4a51000", "0x00e8d4a51000", -1)
	t.Log(str1)
	bbb := SendTxArgs1{}
	err = json.Unmarshal(buff, &bbb)
	if err != nil {
		t.Error(err)
	}
	t.Log(bbb)
}
