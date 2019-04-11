// Copyright (c) 2018 The MATRIX Authors
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
			Version:   "Gman_Alg_0.0.1",
			Submitter: "孙春风,胡源凯",
			Commit: []string{
				"修改委托交易下的vrf失败问题",
				"pos参数配置有误",
			},
		},
		CommitContext{
			Version:   "Gman_Alg_0.0.2",
			Submitter: "孙春风",
			Commit: []string{
				"出块趋向时间由1改为6",
			},
		},
		CommitContext{
			Version:   "Gman_Alg_0.0.3",
			Submitter: "孙春风",
			Commit: []string{
				"删除开发者模式 删除测试网模式 删除rinkeby模式",
				"禁用默认创世文件",
				"委托交易账户外部可见改为man账户",
			},
		},
		CommitContext{
			Version:   "Gman_Alg_0.0.4",
			Submitter: "孙春风",
			Commit: []string{
				"换届服务漏合并的代码",
				"顶点在线修改可能panic的问题",
			},
		},
		{
			Version:   "Gman_Alg_0.0.5",
			Submitter: "Ryan",
			Commit: []string{
				"merge nodeId fixed version, modify bucket limit from two to four and modify broadcast block sender",
			},
		},
		CommitContext{
			Version:   "Gman_Alg_0.0.6",
			Submitter: "孙春风",
			Commit: []string{
				"提供创世文件默认配置,(用户可选择性的填写创世文件,也可不填)",
			},
		},
		CommitContext{
			Version:   "Gman_Alg_0.0.7",
			Submitter: "yeying",
			Commit: []string{
				"修复发送定时交易或者24小时可撤销交易后重启节点导致区块root不一致的问题",
				"修复24小时可撤销交易正常执行完毕后在撤销该笔交易出现崩溃的问题",
				"修复同时发送定时交易和24小时可撤销交易，撤销其中的一笔交易后，转账金额没有减少的问题",
				"修复dump崩溃问题",
				"修改log",
				"deposit bug fixed",
			},
		},
		CommitContext{
			Version:   "Gman_Alg_0.0.8",
			Submitter: "zhangwen",
			Commit: []string{
				"参与奖励使用股权收放系数",
				"彩票奖励修改算法",
				"利息奖励使用初选列表获取vip等级",
				"超级区块签名不允许修改，使用本地状态树账户",
				"出块矿工奖励金额在下一块发放",
			},
		},
		CommitContext{
			Version:   "Gman_Alg_0.0.9",
			Submitter: "huyuankai",
			Commit: []string{
				"特殊账户状态树key值拆分为独立key值",
			},
		},
		CommitContext{
			Version:   "Gman_Alg_0.0.10",
			Submitter: "sunchunfeng",
			Commit: []string{
				"选举算法修改",
			},
		},
		CommitContext{
			Version:   "Gman_Alg_0.0.11",
			Submitter: "yeying",
			Commit: []string{
				"uniform gas price (18000000000)",
			},
		},
		CommitContext{
			Version:   "Gman_Alg_0.0.12",
			Submitter: "张文",
			Commit: []string{
				"矿工出块奖励使用parenthash取前一块的coinbase，解决选取的矿工不一致的问题",
				"修改二级备份节点会多选一个问题",
			},
		},
		CommitContext{
			Version:   "Gman_Alg_0.0.13",
			Submitter: "张文",
			Commit: []string{
				"修改默认创世文件配置",
			},
		},
		CommitContext{
			Version:   "Gman_Alg_0.0.14",
			Submitter: "liubo",
			Commit: []string{
				"区块同步fetch增加log部分打印，便于定位问题",
				"去掉ipfs 相关printf打印",
				"ipfs同步频繁启动协程异常判断改为管道方式",
				"去掉高层使用fetch请求区块",
			},
		},
		{
			Version:   "Gman_Alg_0.0.15",
			Submitter: "Ryan",
			Commit: []string{
				"modify the way to create signature file.",
			},
		},
		CommitContext{
			Version:   "Gman_Alg_0.0.16",
			Submitter: "sunyang",
			Commit: []string{
				"增加setVoted方法，标识节点已对某共识请求投过票",
				"在相关投票操作完成之后调用该方法标识该共识请求投票完成",
			},
		},
		{
			Version:   "Gman_Alg_0.0.17",
			Submitter: "Ryan",
			Commit: []string{
				"modify deposit: delete old signature address, reset account after withdraw and refund",
			},
		},
		{
			Version:   "Gman_Alg_0.0.18",
			Submitter: "Ryan",
			Commit: []string{
				"fix peer bug after change identity, fix change nodeId bug",
			},
		},
		{
			Version:   "Gman_Alg_0.0.19",
			Submitter: "liubo",
			Commit: []string{
				"更新linux下终端无法显示log问题",
				"ipfs下载请求修正及偶然的崩溃问题",
			},
		},
		{
			Version:   "Gman_Alg_0.0.20",
			Submitter: "zhangwen",
			Commit: []string{
				"uptime一直累加",
			},
		},
		{
			Version:   "Gman_Alg_0.0.21",
			Submitter: "huyuankai",
			Commit: []string{
				"增加广播节点热备",
				"状态树版本兼容调整",
				"交易码判断错误bug修复",
			},
		},
		{
			Version:   "Gman_Alg_0.0.22",
			Submitter: "zhenghe",
			Commit: []string{
				"修改抵押竞选方式为按角色竞选(勇哥修改)",
			},
		},
		{
			Version:   "Gman_Alg_0.0.23",
			Submitter: "huyuankai",
			Commit: []string{
				"二次秘钥功调整，拓扑图全部使用A0账户",
				"web3接口 获取区块签名列表，返回A0账户",
				"增加拓扑图状态web3接口",
			},
		},
		{
			Version:   "Gman_Alg_0.0.24",
			Submitter: "hemao",
			Commit: []string{
				"未出块惩罚代功能添加",
			},
		},
		{
			Version:   "Gman_Alg_0.0.25",
			Submitter: "zhenghe,liubo",
			Commit: []string{
				"快照功能合入",
			},
		},
		{
			Version:   "Gman_Alg_0.0.26",
			Submitter: "mehao,huuyankai,zhangwen",
			Commit: []string{
				"创世文件配置及默认创世文件优化",
				"matrix state 改为RLP编码",
				"entrust检查流程代码",
				"共识流程版本兼容",
			},
		},
		{
			Version:   "Gman_Alg_0.0.27",
			Submitter: "zhenghe",
			Commit: []string{
				"download过程按时间委托交易和区块头时间比较",
			},
		},
		{
			Version:   "Gman_Alg_0.0.28",
			Submitter: "zhangwen",
			Commit: []string{
				"解决通过收据获取不到交易的问题",
			},
		},
		{
			Version:   "Gman_Alg_0.0.29",
			Submitter: "zhenghe",
			Commit: []string{
				"修改sendrawtransaction发送方式",
			},
		},
		{
			Version:   "Gman_Alg_0.0.30",
			Submitter: "zhenghe",
			Commit: []string{
				"修改交易相关的日志打印格式",
				"用协程生成快照(yeting)",
			},
		},
		{
			Version:   "Gman_Alg_0.0.31",
			Submitter: "zhangwen",
			Commit: []string{
				"公私钥增加打印",
				"矿工和验证者生成时间点修改支持合约及时生效",
			},
		},
		{
			Version:   "Gman_Alg_0.0.32",
			Submitter: "liubo",
			Commit: []string{
				"1. 区块同步只批量300块，剩余走原来同步方式代码",
				"2. 增加区块压缩存储ipfs功能",
				"3. 增加区块大小统计",
				"4. panic增加时间 ",
			},
		},
		{
			Version:   "Gman_Alg_0.0.33",
			Submitter: "lb",
			Commit: []string{
				"fetch增加loop周期例行检查功能，防止fetch阻塞区块入库",
				"panic增加换行",
			},
		},
		{
			Version:   "Gman_Alg_0.0.34",
			Submitter: "zhenghe",
			Commit: []string{
				"修改多币种奖励",
				"增加黑名单机制",
				"委托交易增加按次数委托",
			},
		},
		{
			Version:   "Gman_Alg_0.0.35",
			Submitter: "zhenghe",
			Commit: []string{
				"创建币种增加to地址限制，销毁金额衰减机制",
			},
		},
	}
)
