// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php

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
				"支持获取版本号和修改内容功能",
				"广播高度和选举高度从配置中读取",
			},
		},
		CommitContext{
			Version:   "Alg_0.0.2",
			Submitter: "孙春风",
			Commit: []string{
				"支持从debug.getCommit()获取",
			},
		},
		CommitContext{
			Version:   "Alg_0.0.3",
			Submitter: "孙春风",
			Commit: []string{
				"提交换届模块未提交的代码",
			},
		},
		CommitContext{
			Version:   "Alg_0.0.4",
			Submitter: "胡源凯",
			Commit: []string{
				"leader服务优化，增加低轮次向高轮次询问流程",
			},
		},
		CommitContext{
			Version:   "Alg_0.0.5",
			Submitter: "胡源凯",
			Commit: []string{
				"修改通过完全区块恢复状态处理时的bug2处",
			},
		},
	}
)
