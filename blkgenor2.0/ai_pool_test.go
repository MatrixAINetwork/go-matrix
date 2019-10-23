// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package blkgenorV2

import (
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"testing"
)

func TestAIResultPool_GetAIResults(t *testing.T) {
	pool := NewAIResultPool("test pool")
	hash := common.HexToHash("0x123456789a")
	resultMap := make(map[common.Address]*aiResultInfo)
	pool.aiMap[hash] = resultMap

	resultMap[common.HexToAddress("0x0001")] = &aiResultInfo{aiMsg: &mc.HD_V2_AIMiningRspMsg{From: common.HexToAddress("0x0001")}, verified: false, legal: false, localTime: 5}
	resultMap[common.HexToAddress("0x0002")] = &aiResultInfo{aiMsg: &mc.HD_V2_AIMiningRspMsg{From: common.HexToAddress("0x0002")}, verified: false, legal: false, localTime: 7}
	resultMap[common.HexToAddress("0x0003")] = &aiResultInfo{aiMsg: &mc.HD_V2_AIMiningRspMsg{From: common.HexToAddress("0x0003")}, verified: false, legal: false, localTime: 3}
	resultMap[common.HexToAddress("0x0004")] = &aiResultInfo{aiMsg: &mc.HD_V2_AIMiningRspMsg{From: common.HexToAddress("0x0004")}, verified: false, legal: false, localTime: 11}
	resultMap[common.HexToAddress("0x0005")] = &aiResultInfo{aiMsg: &mc.HD_V2_AIMiningRspMsg{From: common.HexToAddress("0x0005")}, verified: false, legal: false, localTime: 7}

	list, _ := pool.GetAIResults(hash)
	for i, item := range list {
		t.Logf("%d/%d address=%s time=%d", i, len(list), item.aiMsg.From.Hex(), item.localTime)
	}
}
