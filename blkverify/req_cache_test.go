// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package blkverify

import (
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"math/big"
	"testing"
)

func TestReqCacheSort(t *testing.T) {
	log.InitLog(3)

	cache := make([]*reqData, 0)
	cache = append(cache, &reqData{reqType: reqTypeUnknownReq, hash: common.HexToHash("0x0001"), req: &mc.HD_BlkConsensusReqMsg{Header: &types.Header{Time: big.NewInt(11)}}})
	cache = append(cache, &reqData{reqType: reqTypeUnknownReq, hash: common.HexToHash("0x0002"), req: &mc.HD_BlkConsensusReqMsg{Header: &types.Header{Time: big.NewInt(11)}}})
	cache = append(cache, &reqData{reqType: reqTypeLeaderReq, hash: common.HexToHash("0x0003"), req: &mc.HD_BlkConsensusReqMsg{Header: &types.Header{Time: big.NewInt(13)}}})
	cache = append(cache, &reqData{reqType: reqTypeUnknownReq, hash: common.HexToHash("0x0004"), req: &mc.HD_BlkConsensusReqMsg{Header: &types.Header{Time: big.NewInt(14)}}})
	cache = append(cache, &reqData{reqType: reqTypeUnknownReq, hash: common.HexToHash("0x0005"), req: &mc.HD_BlkConsensusReqMsg{Header: &types.Header{Time: big.NewInt(15)}}})
	cache = append(cache, &reqData{reqType: reqTypeLocalReq, hash: common.HexToHash("0x0006"), req: &mc.HD_BlkConsensusReqMsg{Header: &types.Header{Time: big.NewInt(11)}}})

	cache = delBadReqAndSort(cache, true)

	for i, req := range cache {
		log.Info("cache", "index", i, "req type", req.reqType, "time", req.req.Header.Time, "hash", req.hash.Big())
	}
}
