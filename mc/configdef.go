package mc

import (
	"github.com/matrix/go-matrix/common"
	"math/big"
)

const (
	MSKeyBroadcastTx      = "broad_txs"      // 广播交易
	MSKeyTopologyGraph    = "topology_graph" // 拓扑图
	MSKeyElectGraph       = "elect_graph"    // 选举图
	MSKeyElectOnlineState = "elect_state"    // 选举节点在线消息

	//通用
	MSKeyBroadcastInterval    = "broad_interval" // 广播区块周期
	MSKeyElectGenTime         = "elect_gen_time"
	MSKeyAccountBroadcast     = "account_broadcast"      //广播账户  			common.Address
	MSKeyAccountInnerMiners   = "account_inner_miners"   //基金会矿工 		[]common.Address
	MSKeyAccountFoundation    = "account_foundation"     //基金会账户			common.Address
	MSKeyAccountVersionSupers = "account_version_supers" //版本签名账户		[]common.Address
	MSKeyAccountBlockSupers   = "account_block_supers"   //超级区块签名账户	[]common.Address
	MSKeyElectConfigInfo      = "elect_details_info"
	MSKeyElectMinerNum        = "elect_miner_num"
	MSKeyElectBlackList       = "elect_black_list"
	MSKeyElectWhiteList       = "elect_white_list"
	MSKeyVIPConfig            = "vip_config"
	MSKeyPreBroadcastRoot     = "pre_broadcast_Root"
	MSKeyLeaderConfig         = "leader_config"
	MSKeyMinHash              = "pre_100_min_hash"
	MSKeyPerAllTop            = "pre_all_top_timing"
	MSKeyPreMiner             = "pre_miner"
	MSKeySuperBlockCfg        = "super_block_config"
	//奖励配置
	MSKeyBlkRewardCfg = "blk_reward"
	MSKeyTxsRewardCfg = "txs_reward"
	MSKeyInterestCfg  = "interest_reward" //利息状态
	MSKeyLotteryCfg   = "lottery_reward"
	MSKeySlashCfg     = "slash_reward"
	MSKeyMultiCoin    = "coin_reward"
	//上一矿工奖励金额
	MSKeyPreMinerBlkReward = "preMiner_blkreward"
	//上一矿工交易奖励金额
	MSKeyPreMinerTxsReward = "preMiner_txsreward"
	//upTime状态
	MSKeyUpTimeNum = "upTime_num"
	//彩票状态
	MSKEYLotteryNum     = "lottery_num"
	MSKEYLotteryAccount = "lottery_from"
	//利息状态
	MSInterestCalcNum = "interest_calc_num"
	MSInterestPayNum  = "interest_pay_num"
	//惩罚状态
	MSKeySlashNum = "slash_num"
)

type BCIntervalInfo struct {
	LastBCNumber       uint64 // 最后的广播区块高度
	LastReelectNumber  uint64 // 最后的选举区块高度
	BCInterval         uint64 // 广播周期
	BackupEnableNumber uint64 // 预备广播周期启用高度
	BackupBCInterval   uint64 // 预备广播周期
}

type ElectGenTimeStruct struct {
	MinerGen           uint16
	MinerNetChange     uint16
	ValidatorGen       uint16
	ValidatorNetChange uint16
	VoteBeforeTime     uint16
}

type ElectConfigInfo_All struct {
	MinerNum      uint16
	ValidatorNum  uint16
	BackValidator uint16
	ElectPlug     string
	WhiteList     []common.Address
	BlackList     []common.Address
}
type ElectConfigInfo struct {
	ValidatorNum  uint16
	BackValidator uint16
	ElectPlug     string
}

type VIPConfig struct {
	MinMoney     uint64
	InterestRate uint64 //(分母待定为1000w)
	ElectUserNum uint8
	StockScale   uint16 //千分比
}

type LeaderConfig struct {
	ParentMiningTime      int64 // 预留父区块挖矿时间
	PosOutTime            int64 // 区块POS共识超时时间
	ReelectOutTime        int64 // 重选超时时间
	ReelectHandleInterval int64 // 重选处理间隔时间
}

type PreBroadStateRoot struct {
	LastStateRoot       common.Hash
	BeforeLastStateRoot common.Hash
}

type RewardRateCfg struct {
	MinerOutRate        uint64 //出块矿工奖励
	ElectedMinerRate    uint64 //当选矿工奖励
	FoundationMinerRate uint64 //基金会网络奖励

	LeaderRate              uint64 //出块验证者（leader）奖励
	ElectedValidatorsRate   uint64 //当选验证者奖励
	FoundationValidatorRate uint64 //基金会网络奖励

	OriginElectOfflineRate uint64 //初选下线验证者奖励
	BackupRewardRate       uint64 //当前替补验证者奖励
}

type BlkRewardCfg struct {
	BlkRewardCalc  string
	MinerMount     uint64 //矿工奖励单位man
	MinerHalf      uint64 //矿工折半周期
	ValidatorMount uint64 //验证者奖励 单位man
	ValidatorHalf  uint64 //验证者折半周期
	RewardRate     RewardRateCfg
}

type TxsRewardCfgStruct struct {
	TxsRewardCalc  string
	MinersRate     uint64 //矿工网络奖励
	ValidatorsRate uint64 //验证者网络奖励
	RewardRate     RewardRateCfg
}

type LotteryInfo struct {
	PrizeLevel uint8  //奖励级别
	PrizeNum   uint64 //奖励名额
	PrizeMoney uint64 //奖励金额 单位man
}

type LotteryCfgStruct struct {
	LotteryCalc string
	LotteryInfo []LotteryInfo
}

type InterestCfgStruct struct {
	InterestCalc string
	CalcInterval uint64
	PayInterval  uint64
}

type SlashCfgStruct struct {
	SlashCalc string
	SlashRate uint64
}

type SuperBlkCfg struct {
	Seq uint64
	Num uint64
}

type MinerOutReward struct {
	Reward big.Int
}

type LotteryFrom struct {
	From []common.Address
}

type RandomInfoStruct struct {
	MinHash  common.Hash
	MaxNonce uint64
}
type PreAllTopStruct struct {
	PreAllTopRoot common.Hash
}
type ElectMinerNumStruct struct {
	MinerNum uint16
}
