// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package common

type CommitContext struct {
	Version   string
	Submitter string
	Commit    []string
}

var (
	PutCommit = []CommitContext{
		CommitContext{
			Version:   "Alg_0.0.1",
			Submitter: "孙春风",
			Commit: []string{
				"选举的验证者个数，备选验证者个数，矿工个数可配",
				"广播高度和选举高度从配置中读取",
				"提交记录通过debug.getCommit()和./geth commit 获取",
			},
		},
		CommitContext{
			Version:   "Alg_0.0.2",
			Submitter: "胡源凯",
			Commit: []string{
				"leader服务优化，增加低轮次向高轮次询问的流程，加快同步速度",
			},
		},
	}
)