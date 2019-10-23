// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package mc

import (
	"math/big"

	"github.com/MatrixAINetwork/go-matrix/common"
)

const (
	RewardFullRate = uint64(10000)
)
const (
	MSKeyVersionInfo      = "version_info"   // 版本信息
	MSKeyBroadcastTx      = "broad_txs"      // 广播交易
	MSKeyTopologyGraph    = "topology_graph" // 拓扑图
	MSKeyElectGraph       = "elect_graph"    // 选举图
	MSKeyElectOnlineState = "elect_state"    // 选举节点在线信息

	//通用
	MSKeyBroadcastInterval       = "broad_interval"             // 广播区块周期
	MSKeyElectGenTime            = "elect_gen_time"             // 选举生成时间
	MSKeyElectMinerNum           = "elect_miner_num"            // 选举矿工数量
	MSKeyElectConfigInfo         = "elect_details_info"         // 选举配置
	MSKeyElectBlackList          = "elect_black_list"           // 选举黑名单
	MSKeyElectWhiteList          = "elect_white_list"           // 选举白名单
	MSKeyElectWhiteListSwitcher  = "elect_white_list_switcher"  // 选举白名单生效开关
	MSKeyElectDynamicPollingInfo = "elect_dynamic_polling_info" // 选举动态轮询信息
	MSKeyAccountBroadcasts       = "account_broadcasts"         // 广播账户 []common.Address
	MSKeyAccountInnerMiners      = "account_inner_miners"       // 基金会矿工 []common.Address
	MSKeyAccountFoundation       = "account_foundation"         // 基金会账户 common.Address
	MSKeyAccountVersionSupers    = "account_version_supers"     // 版本签名账户 []common.Address
	MSKeyAccountBlockSupers      = "account_block_supers"       // 超级区块签名账户 []common.Address
	MSKeyAccountMultiCoinSupers  = "account_multicoin_supers"   // 超级多币种签名账户 []common.Address
	MSKeyAccountSubChainSupers   = "account_subchain_supers"    // 子链签名账户 []common.Address
	MSKeyVIPConfig               = "vip_config"                 // VIP配置信息
	MSKeyPreBroadcastRoot        = "pre_broadcast_Root"         // 前广播区块root信息
	MSKeyLeaderConfig            = "leader_config"              // leader服务配置信息
	MSKeyMinHash                 = "pre_100_min_hash"           // 最小hash
	MSKeySuperBlockCfg           = "super_block_config"         // 超级区块配置
	MSKeyMinimumDifficulty       = "min_difficulty"             // 最小挖矿难度
	MSKeyMaximumDifficulty       = "max_difficulty"             // 最大挖矿难度
	MSKeyReelectionDifficulty    = "reelection_difficulty"      // 换届挖矿难度
	MSKeyBlockDurationStatus     = "block_durationstatus"       // 出块时间

	//奖励配置
	MSKeyBlkRewardCfg       = "blk_reward"                // 区块奖励配置
	MSKeyAIBlkRewardCfg     = "aiblk_reward"              // 区块奖励配置
	MSKeyTxsRewardCfg       = "txs_reward"                // 交易奖励配置
	MSKeyInterestCfg        = "interest_reward"           // 利息配置
	MSKeyLotteryCfg         = "lottery_reward"            // 彩票配置
	MSKeySlashCfg           = "slash_reward"              // 惩罚配置
	MSKeyPreMinerBlkReward  = "preMiner_blkreward"        // 上一矿工区块奖励金额
	MSKeyPreMinerTxsReward  = "preMiner_txsreward"        // 上一矿工交易奖励金额
	MSKeyUpTimeNum          = "upTime_num"                // upTime状态
	MSKeyLotteryNum         = "lottery_num"               // 彩票状态
	MSKeyLotteryAccount     = "lottery_from"              // 彩票候选账户
	MSKeyInterestCalcNum    = "interest_calc_num"         // 利息计算状态
	MSKeyInterestPayNum     = "interest_pay_num"          // 利息支付状态
	MSKeySlashNum           = "slash_num"                 // 惩罚状态
	MSKeySelMinerNum        = "selMiner_blkreward"        // 矿工参与奖励状态
	MSKeyBLKSelValidatorNum = "selValidator_blkrewardnum" // 验证者固定区块参与奖励状态
	MSKeyBLKSelValidator    = "selValidator_blkreward"    // 验证者固定参与奖励名单
	MSKeyTXSSelValidatorNum = "selValidator_txsrewardnum" // 验证者交易费参与奖励状态
	MSKeyTXSSelValidator    = "selValidator_txsreward"    // 验证者交易费参与奖励名单

	//奖励算法配置
	MSKeyBlkCalc      = "blk_calc"
	MSKeyTxsCalc      = "txs_calc"
	MSKeyInterestCalc = "interest_calc"
	MSKeyLotteryCalc  = "lottery_calc"
	MSKeySlashCalc    = "slash_calc"

	//未出块选举惩罚配置相关
	MSKeyBlockProduceStatsStatus = "block_produce_stats_status" //
	MSKeyBlockProduceSlashCfg    = "block_produce_slash_cfg"    //
	MSKeyBlockProduceStats       = "block_produce_stats"        //
	MSKeyBlockProduceBlackList   = "block_produce_blacklist"    //

	//算力检测惩罚配置相关
	MSKeyBasePowerStatsStatus = "base_power_stats_status" //
	MSKeyBasePowerSlashCfg    = "base_power_slash_cfg"    //
	MSKeyBasePowerStats       = "base_power_stats"        //
	MSKeyBasePowerBlackList   = "base_power_blacklist"    //
	//交易配置
	MSTxpoolGasLimitCfg = "man_TxpoolGasLimitCfg" //入池gas配置
	MSCurrencyConfig    = "man_CurrencyConfig"    //币种配置
	MSAccountBlackList  = "man_AccountBlackList"  //账户黑名单设置
	MSCurrencyHeader    = "man_CurrencyHeader"    //币种配置
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

type ElectMinerNumStruct struct {
	MinerNum uint16
}

type ElectConfigInfo_All struct {
	MinerNum          uint16
	ValidatorNum      uint16
	BackValidator     uint16
	ElectPlug         string
	WhiteList         []common.Address
	BlackList         []common.Address
	WhiteListSwitcher bool
}
type ElectConfigInfo struct {
	ValidatorNum  uint16
	BackValidator uint16
	ElectPlug     string
}

type ElectDynamicPollingInfo struct {
	Number        uint64           //高度
	Seq           uint64           //轮次序列号
	MinerNum      uint64           //矿工选取个数
	CandidateList []common.Address //候选节点列表
}

type SortVIPConfig []VIPConfig

func (self SortVIPConfig) Len() int {
	return len(self)
}
func (self SortVIPConfig) Less(i, j int) bool {
	return self[i].MinMoney < self[j].MinMoney
}
func (self SortVIPConfig) Swap(i, j int) {
	temp := self[i]
	self[i] = self[j]
	self[j] = temp
}

type VIPConfig struct {
	MinMoney     uint64
	InterestRate uint64 //(分母待定为1000w)
	ElectUserNum uint8
	StockScale   uint16 //千分比
}

type LeaderConfig struct {
	ParentMiningTime      uint64 // 预留父区块挖矿时间
	PosOutTime            uint64 // 区块POS共识超时时间
	ReelectOutTime        uint64 // 重选超时时间
	ReelectHandleInterval uint64 // 重选处理间隔时间
}

type PreBroadStateRoot struct {
	LastStateRoot       []common.CoinRoot
	BeforeLastStateRoot []common.CoinRoot
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

type AIRewardRateCfg struct {
	MinerOutRate        uint64 `json:"MinerOutRate" gencodec:"required"` //矿工奖励单位man //出块矿工奖励
	AIMinerOutRate      uint64 //AI出块矿工奖励
	ElectedMinerRate    uint64 //当选矿工奖励
	FoundationMinerRate uint64 //基金会网络奖励

	LeaderRate              uint64 //出块验证者（leader）奖励
	ElectedValidatorsRate   uint64 //当选验证者奖励
	FoundationValidatorRate uint64 //基金会网络奖励

	OriginElectOfflineRate uint64 //初选下线验证者奖励
	BackupRewardRate       uint64 //当前替补验证者奖励
}

type BlkRewardCfg struct {
	MinerMount               uint64 `json:"MinerMount" gencodec:"required"` //矿工奖励单位man
	MinerAttenuationRate     uint16 //矿工衰减比例
	MinerAttenuationNum      uint64 //矿工衰减周期
	ValidatorMount           uint64 //验证者奖励 单位man
	ValidatorAttenuationRate uint16 //验证者衰减比例
	ValidatorAttenuationNum  uint64 //验证者衰减周期
	RewardRate               RewardRateCfg
}

type AIBlkRewardCfg struct {
	MinerMount               uint64          `json:"MinerMount" gencodec:"required"` //矿工奖励单位man      //矿工奖励单位man
	MinerAttenuationRate     uint16          //矿工衰减比例
	MinerAttenuationNum      uint64          //矿工衰减周期
	ValidatorMount           uint64          //验证者奖励 单位man
	ValidatorAttenuationRate uint16          //验证者衰减比例
	ValidatorAttenuationNum  uint64          //验证者衰减周期
	RewardRate               AIRewardRateCfg `json:"RewardRate" gencodec:"required"`
}

type Genesiscurrencys struct {
	Account string
	Quant   *big.Int
}

type TxsRewardCfg struct {
	MinersRate     uint64 //矿工网络奖励
	ValidatorsRate uint64 //验证者网络奖励
	RewardRate     RewardRateCfg
}

type LotteryInfo struct {
	PrizeLevel uint8  //奖励级别
	PrizeNum   uint64 //奖励名额
	PrizeMoney uint64 //奖励金额 单位man
}

type LotteryCfg struct {
	LotteryInfo []LotteryInfo
}

type InterestCfg struct {
	RewardMount       uint64 //奖励单位man
	AttenuationRate   uint16 //衰减比例
	AttenuationPeriod uint64 //衰减周期
	PayInterval       uint64
}

type SlashCfg struct {
	SlashRate uint64
}

type SuperBlkCfg struct {
	Seq uint64
	Num uint64
}

type MinerOutReward struct {
	Reward big.Int
}

type MultiCoinMinerOutReward struct {
	CoinType string
	Reward   big.Int
}

type LotteryFrom struct {
	From []common.Address
}

type RandomInfoStruct struct {
	MinHash  common.Hash
	MaxNonce uint64
}

type ElectWhiteListSwitcher struct {
	Switcher bool
}

type BlockProduceSlashCfg struct {
	Switcher         bool
	LowTHR           uint16
	ProhibitCycleNum uint16
}

type UserBlockProduceNum struct {
	Address    common.Address
	ProduceNum uint16
}

type BlockProduceStats struct {
	StatsList []UserBlockProduceNum
}

type UserBlockProduceSlash struct {
	Address              common.Address
	ProhibitCycleCounter uint16
}

type BlockProduceSlashBlackList struct {
	BlackList []UserBlockProduceSlash
}

type BlockProduceSlashStatsStatus struct {
	Number uint64
}

type BasePowerSlashCfg struct {
	Switcher         bool
	LowTHR           uint16
	ProhibitCycleNum uint16
}

type BasePowerNum struct {
	Address    common.Address
	ProduceNum uint16
}

type BasePowerStats struct {
	StatsList []BasePowerNum
}

type BasePowerSlash struct {
	Address              common.Address
	ProhibitCycleCounter uint16
}

type BasePowerSlashBlackList struct {
	BlackList []BasePowerSlash
}

type BasePowerSlashStatsStatus struct {
	Number uint64
}
type ChainState struct {
	superSeq  uint64
	curNumber uint64
}
type CurrencyHeader struct {
	Roots    []common.CoinRoot `json:"stateRoot"        gencodec:"required"` //不包含man币
	Sharding []common.Coinbyte `json:"sharding"        gencodec:"required"`
}

type SelReward struct {
	Address common.Address
	Amount  *big.Int
}
type ValidatorSelReward struct {
	CoinType   string
	RewardList []SelReward
}

type BlockDurationStatus struct {
	Status []uint8 //0：无超时；1：低于pos时间；2：超时无矿工
}
