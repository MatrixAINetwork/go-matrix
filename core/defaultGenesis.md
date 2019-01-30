```
	DefaultGenesisJson = `{
    "nettopology":{
    },
    "alloc":{},
    "mstate":{
		"BroadcastInterval" :{         \\广播周期间隔
			"LastBCNumber" : 0,        \\不能配置
			"LastReelectNumber" : 0,   \\不能配置
			"BCInterval" : 100,        \\广播周期
			"BackupEnableNumber" : 0,  \\超级区块使能时刻
			"BackupBCInterval" : 0     \\超级区块修改后的广播周期
		},
		"VIPCfg": [
					{
				"MinMoney": 0,         \\不能配置
				"InterestRate": 5,     \\一个广播周期默认100块，6s出块的利率，年利率=（1+5*36/10000000)^(600*24*365/3600)=0.040
				"ElectUserNum": 0,     \\不能配置
				"StockScale": 1000     \\计算定标值
			},
			{
				"MinMoney": 1000000,    \\vip2的最小额
				"InterestRate": 10,     \\一个广播周期默认100块，6s出块的利率，年利率=（1+10*36/10000000)^(600*24*365/3600)=0.082
				"ElectUserNum": 3,      \\vip2数量
				"StockScale": 1000      \\计算定标值
			},
		{
				"MinMoney": 10000000,    \\vip1的最小额
				"InterestRate": 15,      \\一个广播周期默认100块，6s出块的利率，年利率=（1+15*36/10000000)^(600*24*365/3600)=0.126
				"ElectUserNum": 5,       \\vip1数量
				"StockScale": 1000       \\计算定标值
			}
		],
        "BlkCalcCfg":"1",                \\固定区块奖励算法引擎
        "TxsCalcCfg":"1",                \\交易奖励算法引擎
        "InterestCalcCfg":"1",           \\利息奖励算法引擎
        "LotteryCalcCfg":"1",		     \\彩票奖励算法引擎
        "SlashCalcCfg":"1",              \\惩罚算法引擎
		"BlkRewardCfg": {
			"MinerMount": 3,             \\矿工奖励数额man
			"MinerHalf": 5000000,        \\矿工折半周期
			"ValidatorMount": 7,         \\验证者奖励数额man
			"ValidatorHalf": 5000000,    \\验证者折半周期
			"RewardRate": {
				"MinerOutRate": 4000,      \\出块矿工奖励10000分位
				"ElectedMinerRate": 5000,  \\选中矿工奖励10000分位
				"FoundationMinerRate": 1000,\\基金会奖励10000分位
				"LeaderRate": 4000,         \\验证者leader奖励10000分位
				"ElectedValidatorsRate": 5000, \\参与验证者奖励10000分位
				"FoundationValidatorRate": 1000, \\基金会验证者奖励10000分位
				"OriginElectOfflineRate": 5000,  \\初选下线验证者奖励10000分位
				"BackupRewardRate": 5000         \\备选验证者奖励10000分位
			}
		},
		"TxsRewardCfg": {
			"MinersRate": 0,                   \\矿工奖励
			"ValidatorsRate": 10000,           \\验证者奖励
			"RewardRate": {
				"MinerOutRate": 4000,           \\出块矿工奖励10000分位
				"ElectedMinerRate": 6000,       \\选中矿工奖励10000分位
				"FoundationMinerRate":0,        \\基金会奖励10000分位
				"LeaderRate": 4000,             \\验证者leader奖励10000分位
				"ElectedValidatorsRate": 6000,  \\参与验证者奖励10000分位
				"FoundationValidatorRate": 0,    \\基金会验证者奖励10000分位
				"OriginElectOfflineRate": 5000,   \\初选下线验证者奖励10000分位
				"BackupRewardRate": 5000          \\备选验证者奖励10000分位
			}
		},
		"LotteryCfg": {
//			"LotteryCalc": "1",            
			"LotteryInfo": [{                      
				"PrizeLevel": 0,             \\彩票的级别
				"PrizeNum": 1,               \\彩票当前级别的数目
				"PrizeMoney": 6               \\每个中奖账户的奖励
			}]
		},
		"InterestCfg": {
			"CalcInterval": 100,     \\利息奖励的计息周期
			"PayInterval": 3600      \\利息奖励的结息打款周期
		},
		"LeaderCfg": {
			"ParentMiningTime": 20,   
			"PosOutTime": 20,
			"ReelectOutTime": 40,
			"ReelectHandleInterval": 3
		},
		"SlashCfg": {
			"SlashRate": 7500       \\利息惩罚最大比例10000分位
		},
		"EleTime": {
			"MinerGen": 6,           \\矿工选举时，使用的抵押信息，相对于广播周期的差块
			"MinerNetChange": 5,     \\相对于广播周期生成矿工拓扑图的差块
			"ValidatorGen": 4,       \\验证者选举时，使用的抵押信息，相对于广播周期的差块
			"ValidatorNetChange": 3, \\相对于广播周期生成验证者拓扑图的差块
 			"VoteBeforeTime": 7      \\相对于广播周期公私钥交易的提前时间
		},
		"EleInfo": {
			"ValidatorNum": 19,       \\验证者数目
			"BackValidator": 5,       \\备份验证者数目
			"ElectPlug": "layerd"     \\选举算法插件引擎
		},
		"ElectMinerNum": {
			"MinerNum": 21            \\矿工数目
		},
		"ElectBlackList": null,       \\黑名单
		"ElectWhiteListSwitcherCfg":  { \\选举白名单使能开关
			"Switcher" : false        \\关闭
		},
		"ElectWhiteList": null        
    },
  "config": {
					"chainID": 1,           \\链id
					"byzantiumBlock": 0,     \\
					"homesteadBlock": 0,
					"eip155Block": 0,
			        "eip158Block": 0                        				             
	},
    "versionSignatures": [],                    \\版本签名
    "difficulty":"0x100",                   \\创世难度
    "timestamp":"0x5c26f140",                 \\创世时间
	"version": "1.0.0.0",            \\版本号
  
	"signatures": [	],                        \\创始块签名
    "coinbase": "MAN.1111111111111111111cs", \\创始块奖励人
    "leader":"MAN.1111111111111111111cs", \\创世块leader
    "gasLimit": "0x2FEFD8",                     \\创世块交易费限制
    "nonce": "0x00000000000000178",              \\创世块计数
    "mixHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
    "parentHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
	"extraData": "0x68656c6c6f2c77656c636f6d6520746f206d617472697820776f726c6421"   \\创世块附加数据"hello,welcome to matrix world!"
}
```