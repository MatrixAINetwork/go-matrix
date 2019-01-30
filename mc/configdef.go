package mc

import (
	"math/big"
	"sort"

	"github.com/matrix/go-matrix/base58"

	"github.com/matrix/go-matrix/log"

	"encoding/json"
	"reflect"

	"github.com/matrix/go-matrix/common"
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
	MSKeyBroadcastInterval      = "broad_interval"           // 广播区块周期
	MSKeyElectGenTime           = "elect_gen_time"           // 选举生成时间
	MSKeyElectMinerNum          = "elect_miner_num"          // 选举矿工数量
	MSKeyElectConfigInfo        = "elect_details_info"       // 选举配置
	MSKeyElectBlackList         = "elect_black_list"         // 选举黑名单
	MSKeyElectWhiteList         = "elect_white_list"         // 选举白名单
	MSKeyElectWhiteListSwitcher = "elect_white_list_switcher"         // 选举白名单生效开关
	MSKeyAccountBroadcasts      = "account_broadcasts"       // 广播账户 []common.Address
	MSKeyAccountInnerMiners     = "account_inner_miners"     // 基金会矿工 []common.Address
	MSKeyAccountFoundation      = "account_foundation"       // 基金会账户 common.Address
	MSKeyAccountVersionSupers   = "account_version_supers"   // 版本签名账户 []common.Address
	MSKeyAccountBlockSupers     = "account_block_supers"     // 超级区块签名账户 []common.Address
	MSKeyAccountTxsSupers       = "account_txs_supers"       // 超级交易签名账户 []common.Address
	MSKeyAccountMultiCoinSupers = "account_multicoin_supers" // 超级多币种签名账户 []common.Address
	MSKeyAccountSubChainSupers  = "account_subchain_supers"  // 子链签名账户 []common.Address
	MSKeyVIPConfig              = "vip_config"               // VIP配置信息
	MSKeyPreBroadcastRoot       = "pre_broadcast_Root"       // 前广播区块root信息
	MSKeyLeaderConfig           = "leader_config"            // leader服务配置信息
	MSKeyMinHash                = "pre_100_min_hash"         // 最小hash
	MSKeySuperBlockCfg          = "super_block_config"       // 超级区块配置

	//奖励配置
	MSKeyBlkRewardCfg      = "blk_reward"         // 区块奖励配置
	MSKeyTxsRewardCfg      = "txs_reward"         // 交易奖励配置
	MSKeyInterestCfg       = "interest_reward"    // 利息配置
	MSKeyLotteryCfg        = "lottery_reward"     // 彩票配置
	MSKeySlashCfg          = "slash_reward"       // 惩罚配置
	MSKeyPreMinerBlkReward = "preMiner_blkreward" // 上一矿工区块奖励金额
	MSKeyPreMinerTxsReward = "preMiner_txsreward" // 上一矿工交易奖励金额
	MSKeyUpTimeNum         = "upTime_num"         // upTime状态
	MSKeyLotteryNum        = "lottery_num"        // 彩票状态
	MSKeyLotteryAccount    = "lottery_from"       // 彩票候选账户
	MSKeyInterestCalcNum   = "interest_calc_num"  // 利息计算状态
	MSKeyInterestPayNum    = "interest_pay_num"   // 利息支付状态
	MSKeySlashNum          = "slash_num"          // 惩罚状态
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

	//交易配置
	MSTxpoolGasLimitCfg = "man_TxpoolGasLimitCfg" //入池gas配置
	MSCurrencyPack      = "man_CurrencyPack"      //币种打包限制
	MSAccountBlackList  = "man_AccountBlackList"  //账户黑名单设置
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

func (b *ElectMinerNumStruct) Check(k, v interface{}) (interface{}, bool) {
	if v == nil || k == nil {
		log.ERROR("超级交易配置", "随机选举矿工数目配置为空", "")
		return nil, false
	}
	key, ok := k.(string)
	if !ok {
		log.ERROR("超级交易配置", "key值反射失败", "")
		return nil, false
	}
	if key != MSKeyElectMinerNum {
		log.ERROR("超级交易配置", "key值非法，非法值为", key)
		return nil, false
	}

	codedata, err := json.Marshal(v)
	if err != nil {
		return nil, false
	}
	var value ElectMinerNumStruct
	err = json.Unmarshal(codedata, &value)
	if err != nil {
		return nil, false
	}

	if value.MinerNum == 0 {
		log.ERROR("超级交易配置", "矿工数目配置为0", "")
		return nil, false
	}
	log.Info("参选矿工超级交易配置", "ElectMinerNumStruct", value)
	return value, true

}

func (b *ElectMinerNumStruct) Output(k, v interface{}) (interface{}, interface{}) {

	return k, v
}

type ElectConfigInfo_All struct {
	MinerNum      uint16
	ValidatorNum  uint16
	BackValidator uint16
	ElectPlug     string
	WhiteList     []common.Address
	BlackList     []common.Address
	WhiteListSwitcher bool
}
type ElectConfigInfo struct {
	ValidatorNum  uint16
	BackValidator uint16
	ElectPlug     string
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

func (b *VIPConfig) Check(k, v interface{}) (interface{}, bool) {
	if v == nil || k == nil {
		log.ERROR("VIP超级交易配置", "配置为空", "")
		return nil, false
	}
	key, ok := k.(string)
	if !ok {
		log.ERROR("VIP超级交易配置", "配置key值反射失败", "")
		return nil, false
	}
	if key != MSKeyVIPConfig {
		log.ERROR("VIP超级交易配置", "key值非法，非法值为", key)
		return nil, false
	}

	codedata, err := json.Marshal(v)
	if err != nil {
		return nil, false
	}
	value := make([]VIPConfig, 0)
	err = json.Unmarshal(codedata, &value)
	if err != nil {
		return nil, false
	}
	if 0 == len(value) {
		log.ERROR("VIP超级交易配置", "vip配置为空", "")
		return nil, false
	}
	sort.Sort(SortVIPConfig(value))
	if (value)[0].MinMoney != uint64(0) {
		log.ERROR("VIP超级交易配置", "vip配置中需包含最小值为0的配置", "")
		return nil, false
	}
	for index := 0; index < len(value)-1; index++ {
		if (value)[index].MinMoney == (value)[index+1].MinMoney {
			log.ERROR("VIP超级交易配置", "vip配置中不能包含最小值相同的配置", "")
			return nil, false
		}
	}

	log.Info("VIP超级交易配置", "VIPCfg", value)
	return value, true

}

func (b *VIPConfig) Output(k, v interface{}) (interface{}, interface{}) {

	return k, v
}

type LeaderConfig struct {
	ParentMiningTime      uint64 // 预留父区块挖矿时间
	PosOutTime            uint64 // 区块POS共识超时时间
	ReelectOutTime        uint64 // 重选超时时间
	ReelectHandleInterval uint64 // 重选处理间隔时间
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
	MinerMount     uint64 //矿工奖励单位man
	MinerHalf      uint64 //矿工折半周期
	ValidatorMount uint64 //验证者奖励 单位man
	ValidatorHalf  uint64 //验证者折半周期
	RewardRate     RewardRateCfg
}

func (b *BlkRewardCfg) Check(k, v interface{}) (interface{}, bool) {
	if v == nil || k == nil {
		log.ERROR("超级交易固定区块奖励配置", "奖励配置为空", "")
		return nil, false
	}
	key, ok := k.(string)
	if !ok {
		log.ERROR("超级交易固定区块奖励配置", "key值反射失败", "")
		return nil, false
	}
	if key != MSKeyBlkRewardCfg {
		log.ERROR("超级交易固定区块奖励配置", "key值非法，非法值为", key)
		return nil, false
	}

	codedata, err := json.Marshal(v)
	if err != nil {
		return nil, false
	}
	var value BlkRewardCfg
	err = json.Unmarshal(codedata, &value)
	if err != nil {
		return nil, false
	}
	rateCfg := value.RewardRate

	if RewardFullRate != rateCfg.MinerOutRate+rateCfg.ElectedMinerRate+rateCfg.FoundationMinerRate {

		log.ERROR("超级交易固定区块奖励配置", "矿工固定区块奖励比例配置错误", "")
		return nil, false
	}
	if RewardFullRate != rateCfg.LeaderRate+rateCfg.ElectedValidatorsRate+rateCfg.FoundationValidatorRate {

		log.ERROR("超级交易固定区块奖励配置", "验证者固定区块奖励比例配置错误", "")
		return nil, false
	}

	if RewardFullRate != rateCfg.OriginElectOfflineRate+rateCfg.BackupRewardRate {

		log.ERROR("超级交易固定区块奖励配置", "替补固定区块奖励比例配置错误", "")
		return nil, false
	}
	log.Info("超级交易配置", "BlkRewardCfg", rateCfg)

	return value, true

}

func (b *BlkRewardCfg) Output(k, v interface{}) (interface{}, interface{}) {

	return k, v
}

type TxsRewardCfg struct {
	MinersRate     uint64 //矿工网络奖励
	ValidatorsRate uint64 //验证者网络奖励
	RewardRate     RewardRateCfg
}

func (b *TxsRewardCfg) Check(k, v interface{}) (interface{}, bool) {
	if v == nil || k == nil {
		log.ERROR("超级交易交易费奖励配置", "奖励配置为空", "")
		return nil, false
	}
	key, ok := k.(string)
	if !ok {
		log.ERROR("超级交易交易费奖励配置", "key值反射失败", "")
		return nil, false
	}
	if key != MSKeyTxsRewardCfg {
		log.ERROR("超级交易交易费奖励配置", "key值非法，非法值为", key)
		return nil, false
	}

	codedata, err := json.Marshal(v)
	if err != nil {
		return nil, false
	}
	var value TxsRewardCfg
	err = json.Unmarshal(codedata, &value)
	if err != nil {
		return nil, false
	}
	rateCfg := value.RewardRate

	if RewardFullRate != value.ValidatorsRate+value.MinersRate {

		log.ERROR("超级交易交易费奖励配置", "交易矿工验证者奖励比例配置错误", "")
		return nil, false
	}
	if RewardFullRate != rateCfg.MinerOutRate+rateCfg.ElectedMinerRate+rateCfg.FoundationMinerRate {

		log.ERROR("超级交易交易费奖励配置", "矿工交易费奖励比例配置错误", "")
		return nil, false
	}
	if RewardFullRate != rateCfg.LeaderRate+rateCfg.ElectedValidatorsRate+rateCfg.FoundationValidatorRate {

		log.ERROR("超级交易交易费奖励配置", "验证者交易费奖励比例配置错误", "")
		return nil, false
	}

	if RewardFullRate != rateCfg.OriginElectOfflineRate+rateCfg.BackupRewardRate {

		log.ERROR("超级交易交易费奖励配置", "替补交易费奖励比例配置错误", "")
		return nil, false
	}
	log.Info("超级交易配置", "TxsRewardCfg", rateCfg)

	return value, true

}

func (b *TxsRewardCfg) Output(k, v interface{}) (interface{}, interface{}) {

	return k, v
}

type BlkRewardCalc struct {
}

func (b *BlkRewardCalc) Check(k, v interface{}) (interface{}, bool) {
	if v == nil || k == nil {
		log.ERROR("超级交易区块算法配置", "奖励配置为空", "")
		return nil, false
	}
	key, ok := k.(string)
	if !ok {
		log.ERROR("超级交易区块算法配置", "key值反射失败", "")
		return nil, false
	}
	if key != MSKeyBlkCalc {
		log.ERROR("超级交易区块算法配置", "key值非法，非法值为", key)
		return nil, false
	}

	codedata, err := json.Marshal(v)
	if err != nil {
		return nil, false
	}
	var values string
	err = json.Unmarshal(codedata, &values)
	if err != nil {
		return nil, false
	}
	log.Info("超级交易区块算法配置", "BlkRewardCalc", v)
	return values, true
}
func (b *BlkRewardCalc) Output(k, v interface{}) (interface{}, interface{}) {

	return k, v
}

type TxsRewardCalc struct {
}

func (b *TxsRewardCalc) Check(k, v interface{}) (interface{}, bool) {
	if v == nil || k == nil {
		log.ERROR("超级交易交易费算法配置", "奖励配置为空", "")
		return nil, false
	}
	key, ok := k.(string)
	if !ok {
		log.ERROR("超级交易交易费算法配置", "key值反射失败", "")
		return nil, false
	}
	if key != MSKeyTxsCalc {
		log.ERROR("超级交易交易费算法配置", "key值非法，非法值为", key)
		return nil, false
	}

	codedata, err := json.Marshal(v)
	if err != nil {
		return nil, false
	}
	var values string
	err = json.Unmarshal(codedata, &values)
	if err != nil {
		return nil, false
	}
	log.Info("超级交易交易费算法配置", "TxsRewardCalc", v)
	return values, true
}
func (b *TxsRewardCalc) Output(k, v interface{}) (interface{}, interface{}) {

	return k, v
}

type InterestRewardCalc struct {
}

func (b *InterestRewardCalc) Check(k, v interface{}) (interface{}, bool) {
	if v == nil || k == nil {
		log.ERROR("超级交易利息算法配置", "奖励配置为空", "")
		return nil, false
	}
	key, ok := k.(string)
	if !ok {
		log.ERROR("超级交易利息算法配置", "key值反射失败", "")
		return nil, false
	}
	if key != MSKeyInterestCalc {
		log.ERROR("超级交易利息算法配置", "key值非法，非法值为", key)
		return nil, false
	}

	codedata, err := json.Marshal(v)
	if err != nil {
		return nil, false
	}
	var values string
	err = json.Unmarshal(codedata, &values)
	if err != nil {
		return nil, false
	}
	log.Info("超级交易利息算法配置", "InterestRewardCalc", v)
	return values, true
}
func (b *InterestRewardCalc) Output(k, v interface{}) (interface{}, interface{}) {

	return k, v
}

type LotteryRewardCalc struct {
}

func (b *LotteryRewardCalc) Check(k, v interface{}) (interface{}, bool) {
	if v == nil || k == nil {
		log.ERROR("超级交易彩票算法配置", "奖励配置为空", "")
		return nil, false
	}
	key, ok := k.(string)
	if !ok {
		log.ERROR("超级交易彩票算法配置", "key值反射失败", "")
		return nil, false
	}
	if key != MSKeyLotteryCalc {
		log.ERROR("超级交易彩票算法配置", "key值非法，非法值为", key)
		return nil, false
	}

	codedata, err := json.Marshal(v)
	if err != nil {
		return nil, false
	}
	var values string
	err = json.Unmarshal(codedata, &values)
	if err != nil {
		return nil, false
	}
	log.Info("超级交易彩票算法配置", "LotteryRewardCalc", v)
	return values, true
}
func (b *LotteryRewardCalc) Output(k, v interface{}) (interface{}, interface{}) {

	return k, v
}

type SlashCalc struct {
}

func (b *SlashCalc) Check(k, v interface{}) (interface{}, bool) {
	if v == nil || k == nil {
		log.ERROR("超级交易惩罚算法配置", "奖励配置为空", "")
		return nil, false
	}
	key, ok := k.(string)
	if !ok {
		log.ERROR("超级交易惩罚算法配置", "key值反射失败", "")
		return nil, false
	}
	if key != MSKeySlashCalc {
		log.ERROR("超级交易惩罚算法配置", "key值非法，非法值为", key)
		return nil, false
	}

	codedata, err := json.Marshal(v)
	if err != nil {
		return nil, false
	}
	var values string
	err = json.Unmarshal(codedata, &values)
	if err != nil {
		return nil, false
	}
	log.Info("超级交易惩罚算法配置", "SlashCalc", v)
	return values, true
}
func (b *SlashCalc) Output(k, v interface{}) (interface{}, interface{}) {

	return k, v
}

type LotteryInfo struct {
	PrizeLevel uint8  //奖励级别
	PrizeNum   uint64 //奖励名额
	PrizeMoney uint64 //奖励金额 单位man
}

type LotteryCfg struct {
	LotteryInfo []LotteryInfo
}

func (b *LotteryCfg) Check(k, v interface{}) (interface{}, bool) {

	if v == nil || k == nil {
		log.ERROR("超级交易彩票奖励配置", "奖励配置为空", "")
		return nil, false
	}
	key, ok := k.(string)
	if !ok {
		log.ERROR("超级交易彩票奖励配置", "key值反射失败", "")
		return nil, false
	}
	if key != MSKeyLotteryCfg {
		log.ERROR("超级交易彩票奖励配置", "key值非法，非法值为", key)
		return nil, false
	}

	codedata, err := json.Marshal(v)
	if err != nil {
		return nil, false
	}
	values := LotteryCfg{}
	err = json.Unmarshal(codedata, &values)
	if err != nil {
		return nil, false
	}

	log.Info("超级交易彩票奖励配置", "LotteryCfg", values)
	return values, true
}

func (b *LotteryCfg) Output(k, v interface{}) (interface{}, interface{}) {

	return k, v
}

type InterestCfg struct {
	CalcInterval uint64
	PayInterval  uint64
}

func (b *InterestCfg) Check(k, v interface{}) (interface{}, bool) {
	if v == nil || k == nil {
		log.ERROR("超级交易利息奖励配置", "奖励配置为空", "")
		return nil, false
	}
	key, ok := k.(string)
	if !ok {
		log.ERROR("超级交易利息奖励配置", "key值反射失败", "")
		return nil, false
	}
	if key != MSKeyInterestCfg {
		log.ERROR("超级交易利息奖励配置", "key值非法，非法值为", key)
		return nil, false
	}

	codedata, err := json.Marshal(v)
	if err != nil {
		return nil, false
	}
	values := InterestCfg{}
	err = json.Unmarshal(codedata, &values)
	if err != nil {
		return nil, false
	}
	log.Info("超级交易利息奖励配置", "InterestCfg", values)
	return values, true
}

func (b *InterestCfg) Output(k, v interface{}) (interface{}, interface{}) {

	return k, v
}

type SlashCfg struct {
	SlashRate uint64
}

func (b *SlashCfg) Check(k, v interface{}) (interface{}, bool) {
	if v == nil || k == nil {
		log.ERROR("超级交易惩罚奖励配置", "奖励配置为空", "")
		return nil, false
	}
	key, ok := k.(string)
	if !ok {
		log.ERROR("超级交易惩罚奖励配置", "key值反射失败", "")
		return nil, false
	}
	if key != MSKeySlashCfg {
		log.ERROR("超级交易惩罚奖励配置", "key值非法，非法值为", key)
		return nil, false
	}

	codedata, err := json.Marshal(v)
	if err != nil {
		return nil, false
	}
	values := SlashCfg{}
	err = json.Unmarshal(codedata, &values)
	if err != nil {
		return nil, false
	}

	if values.SlashRate > RewardFullRate {

		log.ERROR("超级交易惩罚奖励配置", "配置的最大惩罚比例系数大于10000", values.SlashRate)
		return nil, false
	}
	log.Info("超级交易惩罚奖励配置", "SlashCfg", values)
	return values, true
}

func (b *SlashCfg) Output(k, v interface{}) (interface{}, interface{}) {

	return k, v
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

type BroadcastAccounts struct {
}

func (b *BroadcastAccounts) Check(k, v interface{}) (interface{}, bool) {
	if v == nil || k == nil {
		log.ERROR("超级交易配置", "广播节点为空", "")
		return nil, false
	}
	key, ok := k.(string)
	if !ok {
		log.ERROR("超级交易配置", "key值反射失败", "")
		return nil, false
	}
	if key != MSKeyAccountBroadcasts {
		log.ERROR("超级交易配置", "key值非法，非法值为", key)
		return nil, false
	}

	codedata, err := json.Marshal(v)
	if err != nil {
		return nil, false
	}
	values := make([]common.Address, 0)
	err = json.Unmarshal(codedata, &values)
	if err != nil {
		return nil, false
	}
	if 0 == len(values) {
		log.ERROR("超级交易广播节点配置", "为空", "")
		return nil, false
	}
	log.Info("超级交易广播节点配置", "BroadcastAccounts", v)
	return values, true

}

func (b *BroadcastAccounts) Output(k, v interface{}) (interface{}, interface{}) {
	value, ok := v.([]common.Address)
	if !ok {
		log.ERROR("超级交易配置", "value值反射失败", "")
		return k, v
	}
	if len(value) == 0 {
		log.ERROR("超级交易配置", "设置的广播节点个数为0", value)
		return k, v
	}
	base58Accounts := make([]string, 0)
	for _, v := range value {
		base58Accounts = append(base58Accounts, base58.Base58EncodeToString("MAN", v))
	}
	return k, base58Accounts
}

type InnerMinersAccounts struct {
}

func (b *InnerMinersAccounts) Check(k, v interface{}) (interface{}, bool) {
	if v == nil || k == nil {
		log.ERROR("超级交易配置", "广播节点为空", "")
		return nil, false
	}
	key, ok := k.(string)
	if !ok {
		log.ERROR("超级交易配置", "key值反射失败", "")
		return nil, false
	}
	if key != MSKeyAccountInnerMiners {
		log.ERROR("超级交易配置", "key值非法，非法值为", key)
		return nil, false
	}
	codedata, err := json.Marshal(v)
	if err != nil {
		return nil, false
	}
	values := make([]common.Address, 0)
	err = json.Unmarshal(codedata, &values)
	if err != nil {
		return nil, false
	}
	log.Info("超级交易内部矿工节点配置", "InnerMinersAccounts", v)
	return values, true

}

func (b *InnerMinersAccounts) Output(k, v interface{}) (interface{}, interface{}) {
	value, ok := v.([]common.Address)
	if !ok {
		log.ERROR("超级交易配置", "value值反射失败", "")
		return k, v
	}
	if len(value) == 0 {
		log.Info("超级交易配置", "设置的内部矿工节点个数为0", value)
		return k, v
	}
	base58Accounts := make([]string, 0)
	for _, v := range value {
		base58Accounts = append(base58Accounts, base58.Base58EncodeToString("MAN", v))
	}
	return k, base58Accounts
}

type ElectBlackList struct {
}

func (b *ElectBlackList) Check(k, v interface{}) (interface{}, bool) {
	if v == nil || k == nil {
		log.ERROR("超级交易选举黑名单配置", "广播节点为空", "")
		return nil, false
	}
	key, ok := k.(string)
	if !ok {
		log.ERROR("超级交易选举黑名单配置", "key值反射失败", "")
		return nil, false
	}
	if key != MSKeyElectBlackList {
		log.ERROR("超级交易选举黑名单配置", "key值非法，非法值为", key)
		return nil, false
	}
	codedata, err := json.Marshal(v)
	if err != nil {
		return nil, false
	}
	values := make([]common.Address, 0)
	err = json.Unmarshal(codedata, &values)
	if err != nil {
		return nil, false
	}
	log.Info("超级交易选举黑名单配置", "ElectBlackList", v)
	return values, true

}

func (b *ElectBlackList) Output(k, v interface{}) (interface{}, interface{}) {
	value, ok := v.([]common.Address)
	if !ok {
		log.ERROR("超级交易选举黑名单配置", "value值反射失败", "")
		return k, v
	}
	if len(value) == 0 {
		log.INFO("超级交易选举黑名单配置", "设置的选举黑名单个数为0", value)
		return k, v
	}
	base58Accounts := make([]string, 0)
	for _, v := range value {
		base58Accounts = append(base58Accounts, base58.Base58EncodeToString("MAN", v))
	}
	return k, base58Accounts
}

type ElectWhiteList struct {
}

func (b *ElectWhiteList) Check(k, v interface{}) (interface{}, bool) {
	if v == nil || k == nil {
		log.ERROR("超级交易选举白名单配置", "广播节点为空", "")
		return nil, false
	}
	key, ok := k.(string)
	if !ok {
		log.ERROR("超级交易选举白名单配置", "key值反射失败", "")
		return nil, false
	}
	if key != MSKeyElectWhiteList {
		log.ERROR("超级交易选举白名单配置", "key值非法，非法值为", key)
		return nil, false
	}
	codedata, err := json.Marshal(v)
	if err != nil {
		return nil, false
	}
	values := make([]common.Address, 0)
	err = json.Unmarshal(codedata, &values)
	if err != nil {
		return nil, false
	}

	log.Info("超级交易选举白名单配置", "ElectBlackList", v)
	return values, true

}

func (b *ElectWhiteList) Output(k, v interface{}) (interface{}, interface{}) {
	value, ok := v.([]common.Address)
	if !ok {
		log.ERROR("超级交易选举白名单配置", "value值反射失败", "")
		return k, v
	}
	if len(value) == 0 {
		log.INFO("超级交易选举白名单配置", "设置的白名单为nil", value)
		return k, v
	}
	base58Accounts := make([]string, 0)
	for _, v := range value {
		base58Accounts = append(base58Accounts, base58.Base58EncodeToString("MAN", v))
	}
	return k, base58Accounts
}

type ElectWhiteListSwitcher struct {
	Switcher bool
}

func (b *ElectWhiteListSwitcher) Check(k, v interface{}) (interface{}, bool) {
	if v == nil || k == nil {
		log.ERROR("超级交易白名单开关配置", "k v为空", "")
		return nil, false
	}
	key, ok := k.(string)
	if !ok {
		log.ERROR("超级交易白名单开关配置", "key值反射失败", "")
		return nil, false
	}
	if key != MSKeyElectWhiteListSwitcher{
		log.ERROR("超级交易白名单开关配置", "key值非法，非法值为", key)
		return nil, false
	}

	codedata, err := json.Marshal(v)
	if err != nil {
		return nil, false
	}
	value:= ElectWhiteListSwitcher{}
	err = json.Unmarshal(codedata, &value)
	if err != nil {
		return nil, false
	}
	log.Info("超级交易白名单开关配置", "ElectWhiteListSwitcher", value)
	return value, true
}

func (b *ElectWhiteListSwitcher) Output(k, v interface{}) (interface{}, interface{}) {

	return k, v
}
type BlockProduceSlashCfg struct {
	Switcher         bool
	LowTHR           uint16
	ProhibitCycleNum uint16
}

func (b *BlockProduceSlashCfg) Check(k, v interface{}) (interface{}, bool) {
	if v == nil || k == nil {
		log.ERROR("超级交易出块惩罚配置", "k v为空", "")
		return nil, false
	}
	key, ok := k.(string)
	if !ok {
		log.ERROR("超级交易出块惩罚配置", "key值反射失败", "")
		return nil, false
	}
	if key != MSKeyBlockProduceSlashCfg {
		log.ERROR("超级交易出块惩罚配置", "key值非法，非法值为", key)
		return nil, false
	}

	codedata, err := json.Marshal(v)
	if err != nil {
		return nil, false
	}
	values := BlockProduceSlashCfg{}
	err = json.Unmarshal(codedata, &values)
	if err != nil {
		return nil, false
	}
	log.Info("超级交易出块惩罚配置", "BlockProduceSlashCfg", v)
	return values, true
}

func (b *BlockProduceSlashCfg) Output(k, v interface{}) (interface{}, interface{}) {

	return k, v
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

type TxpoolGasLimit struct {
}

func (b *TxpoolGasLimit) Check(k, v interface{}) (interface{}, bool) {
	if v == nil || k == nil {
		log.ERROR("超级交易入池gas配置为空")
		return nil, false
	}
	key, ok := k.(string)
	if !ok {
		log.ERROR("超级交易入池gas配置", "key值反射失败", "")
		return nil, false
	}
	if key != MSTxpoolGasLimitCfg {
		log.ERROR("超级交易区块算法配置", "key值非法，非法值为", key)
		return nil, false
	}

	codedata, err := json.Marshal(v)
	if err != nil {
		return nil, false
	}
	value := new(big.Int)
	err = json.Unmarshal(codedata, &value)
	if err != nil {
		return nil, false
	}

	return value, true
}
func (b *TxpoolGasLimit) Output(k, v interface{}) (interface{}, interface{}) {
	return k, v
}

type CurrencyPackLimt struct {
}

func (b *CurrencyPackLimt) Check(k, v interface{}) (interface{}, bool) {
	if v == nil || k == nil {
		log.ERROR("超级交易区块配置", "币种打包限制输入为空", "")
		return nil, false
	}
	key, ok := k.(string)
	if !ok {
		log.ERROR("超级交易币种打包限制配置", "key值反射失败", "")
		return nil, false
	}
	if key != MSCurrencyPack {
		log.ERROR("超级交易币种打包限制配置", "key值非法，非法值为", key)
		return nil, false
	}

	v1 := reflect.ValueOf(v)
	if v1.Kind() == reflect.Slice && v1.Len() == 0 {
		log.INFO("超级交易币种打包限制配置", "设置的币种限制个数为0", "")
		return make([]string, 0), true
	}
	v2, ok := v.([]interface{})
	if !ok {
		log.ERROR("超级交易币种打包限制配置", "v值反射失败", "")
		return nil, false
	}

	if reflect.ValueOf(v2[0]).Kind() == reflect.String && v2[0] == "" {
		log.INFO("超级交易币种打包限制配置", "设置的币种限制个数为0", "")
		return make([]string, 0), true
	}

	codedata, err := json.Marshal(v)
	if err != nil {
		return nil, false
	}
	values := make([]string, 0)
	err = json.Unmarshal(codedata, &values)
	if err != nil {
		return nil, false
	}

	for _, currency := range values {
		if !common.IsValidityCurrency(currency) {
			log.ERROR("超级交易币种打包限制配置", "币种格式不正确", "")
			return nil, false
		}
	}
	return values, true
}
func (b *CurrencyPackLimt) Output(k, v interface{}) (interface{}, interface{}) {
	return k, v
}

type AccountBlackList struct {
}

func (b *AccountBlackList) Check(k, v interface{}) (interface{}, bool) {
	if v == nil || k == nil {
		log.ERROR("超级交易配置", "账户黑名单为空", "")
		return nil, false
	}
	key, ok := k.(string)
	if !ok {
		log.ERROR("超级交易配置账户黑名单", "key值反射失败", "")
		return nil, false
	}
	if key != MSAccountBlackList {
		log.ERROR("超级交易配置账户黑名单", "key值非法，非法值为", key)
		return nil, false
	}
	v1 := reflect.ValueOf(v)
	if v1.Kind() == reflect.Slice && v1.Len() == 0 {
		log.INFO("超级交易配置账户黑名单", "设置的账户黑名单个数为0", "")
		return make([]common.Address, 0), true
	}

	codedata, err := json.Marshal(v)
	if err != nil {
		return nil, false
	}
	values := make([]common.Address, 0)
	err = json.Unmarshal(codedata, &values)
	if err != nil {
		return nil, false
	}
	return values, true
}

func (b *AccountBlackList) Output(k, v interface{}) (interface{}, interface{}) {
	value, ok := v.([]common.Address)
	if !ok {
		log.ERROR("超级交易配置账户黑名单", "value值反射失败", "")
		return nil, nil
	}
	if len(value) == 0 {
		log.INFO("超级交易配置账户黑名单", "设置的账户黑名单个数为0", value)
		return k, v
	}
	base58Accounts := make([]string, 0)
	for _, v := range value {
		base58Accounts = append(base58Accounts, base58.Base58EncodeToString("MAN", v))
	}
	return k, base58Accounts
}
