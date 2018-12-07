// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package support

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/mc"
)

const (
	MaxSample = 1000 //配置参数,采样最多发生1000次,是一个离P+M较远的值
	J         = 0    //基金会验证节点个数tps_weight
	M         = 11   //验证主节点个数
	P         = 5    //备份主节点个数
	N         = 21   //矿工主节点个数

	DefaultDeposit    = 50000
	DefaultWithdrawH  = 0
	DefaultOnlineTime = 300
)

type AllNative struct {
	Master    []mc.TopologyNodeInfo //验证者主节点
	BackUp    []mc.TopologyNodeInfo //验证者备份
	Candidate []mc.TopologyNodeInfo //验证者候选

	MasterQ    []common.Address //第一梯队候选
	BackUpQ    []common.Address //第二梯队候选
	CandidateQ []common.Address //第三梯队候选

}
