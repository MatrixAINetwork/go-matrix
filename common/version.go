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
		CommitContext{
			Version:   "Alg_1207_0.0.1",
			Submitter: "胡源凯",
			Commit: []string{
				"顶点服务优化: 支持消息乱序",
				"顶点共识上头方案优化: 共识结果放入区块共识请求中，并增加共识结果有效期限制",
				"顶点服务优化: 顶点在线共识简化共识流程",
				"区块生成、区块验证服务由于顶点方案改动，响应修改变化拓扑生成及验证流程",
				"为支持顶点服务，修改elect节点在线信息缓存方式",
				"修改共识投票消息结构，取消轮次，增加高度",
				"修改顶点在线共识请求消息结构",
			},
		},
	}
)
