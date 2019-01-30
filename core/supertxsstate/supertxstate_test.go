package supertxsstate

import (
	"fmt"
	"testing"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/reward/util"

	"github.com/matrix/go-matrix/mc"

	"github.com/matrix/go-matrix/log"

	"github.com/matrix/go-matrix/params/manparams"
	"math/big"
)

const (
	testAddress             = "0x8605cdbbdb6d264aa742e77020dcbc58fcdce182"
	ValidatorsTxsRewardRate = uint64(util.RewardFullRate) //验证者交易奖励比例100%
	MinerTxsRewardRate      = uint64(0)                   //矿工交易奖励比例0%
	FoundationTxsRewardRate = uint64(0)                   //基金会交易奖励比例0%

	MinerOutRewardRate     = uint64(4000) //出块矿工奖励40%
	ElectedMinerRewardRate = uint64(6000) //当选矿工奖励60%

	LeaderRewardRate            = uint64(4000) //出块验证者（leader）奖励40%
	ElectedValidatorsRewardRate = uint64(6000) //当选验证者奖励60%

	OriginElectOfflineRewardRate = uint64(5000) //初选下线验证者奖励50%
	BackupRate                   = uint64(5000) //当前替补验证者奖励50%
)

func Test_newManager(t *testing.T) {
	log.InitLog(3)
	a := GetManager(manparams.VersionAlpha)
	var slash mc.SlashCfg
	slash.SlashRate = 7500
	if a.Check(mc.MSKeySlashCfg, slash) {
		fmt.Println(a.Output(mc.MSKeySlashCfg, slash))
	}

	var electMiner mc.ElectMinerNumStruct
	electMiner.MinerNum = 1
	if a.Check(mc.MSKeyElectMinerNum, electMiner) {
		fmt.Println(a.Output(mc.MSKeyElectMinerNum, electMiner))
	}

	var accountblklist1 []common.Address
	accountblklist1 = make([]common.Address, 0)
	//accountblklist = append(accountblklist, common.HexToAddress("0x12345"))
	if a.Check(mc.MSAccountBlackList, accountblklist1) {
		fmt.Println(a.Output(mc.MSAccountBlackList, accountblklist1))
	} else {
		t.Error("执行失败")
		//return
	}

	var black []common.Address
	black = make([]common.Address, 0)
	//black = append(black, common.HexToAddress("0x01"))
	if a.Check(mc.MSKeyElectBlackList, black) {
		fmt.Println(a.Output(mc.MSKeyElectBlackList, black))
	}

	var white []common.Address
	white = append(white, common.HexToAddress("0x02"))
	if a.Check(mc.MSKeyElectWhiteList, white) {
		fmt.Println(a.Output(mc.MSKeyElectWhiteList, white))
	}

	var broadcast []common.Address
	broadcast = append(broadcast, common.HexToAddress("0x03"))
	if a.Check(mc.MSKeyAccountBroadcasts, broadcast) {
		fmt.Println(a.Output(mc.MSKeyAccountBroadcasts, broadcast))
	} else {
		t.Error("执行失败")
		return
	}

	var innerminers []common.Address
	innerminers = append(innerminers, common.HexToAddress("0x04"))
	if a.Check(mc.MSKeyAccountInnerMiners, innerminers) {
		fmt.Println(a.Output(mc.MSKeyAccountInnerMiners, innerminers))
	} else {
		t.Error("执行失败")
		return
	}

	var vip []mc.VIPConfig
	vip = append(vip, mc.VIPConfig{0, 5, 1, 1})
	if a.Check(mc.MSKeyVIPConfig, vip) {
		fmt.Println(a.Output(mc.MSKeyVIPConfig, vip))
	} else {
		t.Error("执行失败")
		return
	}

	blkcfg := mc.BlkRewardCfg{MinerMount: 6,
		MinerHalf:      50000,
		ValidatorMount: 3,
		ValidatorHalf:  50000,
		RewardRate: mc.RewardRateCfg{
			MinerOutRate:        MinerOutRewardRate,
			ElectedMinerRate:    ElectedMinerRewardRate,
			FoundationMinerRate: FoundationTxsRewardRate,

			LeaderRate:              LeaderRewardRate,
			ElectedValidatorsRate:   ElectedValidatorsRewardRate,
			FoundationValidatorRate: FoundationTxsRewardRate,

			OriginElectOfflineRate: OriginElectOfflineRewardRate,
			BackupRewardRate:       BackupRate,
		}}
	if a.Check(mc.MSKeyBlkRewardCfg, blkcfg) {
		fmt.Println(a.Output(mc.MSKeyBlkRewardCfg, blkcfg))
	} else {
		t.Error("执行失败")
		return
	}

	txscfg := mc.TxsRewardCfg{MinersRate: 0, ValidatorsRate: 10000, RewardRate: mc.RewardRateCfg{
		MinerOutRate:        MinerOutRewardRate,
		ElectedMinerRate:    ElectedMinerRewardRate,
		FoundationMinerRate: FoundationTxsRewardRate,

		LeaderRate:              LeaderRewardRate,
		ElectedValidatorsRate:   ElectedValidatorsRewardRate,
		FoundationValidatorRate: FoundationTxsRewardRate,

		OriginElectOfflineRate: OriginElectOfflineRewardRate,
		BackupRewardRate:       BackupRate,
	}}
	if a.Check(mc.MSKeyTxsRewardCfg, txscfg) {
		fmt.Println(a.Output(mc.MSKeyTxsRewardCfg, txscfg))
	} else {
		t.Error("执行失败")
		return
	}

	Interestcfg := mc.InterestCfg{100, 3600}
	if a.Check(mc.MSKeyInterestCfg, Interestcfg) {
		fmt.Println(a.Output(mc.MSKeyInterestCfg, txscfg))
	} else {
		t.Error("执行失败")
		return
	}
	var LotteryInfo []mc.LotteryInfo
	LotteryInfo = append(LotteryInfo, mc.LotteryInfo{0, 1, 6})
	LotterCfg := mc.LotteryCfg{LotteryInfo: LotteryInfo}
	if a.Check(mc.MSKeyLotteryCfg, LotterCfg) {
		fmt.Println(a.Output(mc.MSKeyLotteryCfg, LotterCfg))
	} else {
		t.Error("执行失败")
		return
	}

	blkcalc := "1"
	if a.Check(mc.MSKeyBlkCalc, blkcalc) {
		fmt.Println(a.Output(mc.MSKeyBlkCalc, blkcalc))
	} else {
		t.Error("执行失败")
		return
	}
	if a.Check(mc.MSKeyTxsCalc, blkcalc) {
		fmt.Println(a.Output(mc.MSKeyTxsCalc, blkcalc))
	} else {
		t.Error("执行失败")
		return
	}

	if a.Check(mc.MSKeyInterestCalc, blkcalc) {
		fmt.Println(a.Output(mc.MSKeyInterestCalc, blkcalc))
	} else {
		t.Error("执行失败")
		return
	}

	if a.Check(mc.MSKeyLotteryCalc, blkcalc) {
		fmt.Println(a.Output(mc.MSKeyLotteryCalc, blkcalc))
	} else {
		t.Error("执行失败")
		return
	}

	if a.Check(mc.MSKeySlashCalc, blkcalc) {
		fmt.Println(a.Output(mc.MSKeySlashCalc, blkcalc))
	} else {
		t.Error("执行失败")
		return
	}

	s1 := "DDDDAA"
	if a.Check(mc.MSCurrencyPack, s1) {
		fmt.Println(a.Output(mc.MSCurrencyPack, s1))
	} else {
		t.Error("执行失败")
		//return
	}

	gas := *big.NewInt(int64(1800000))
	if a.Check(mc.MSTxpoolGasLimitCfg, gas) {
		fmt.Println(a.Output(mc.MSTxpoolGasLimitCfg, gas))
	} else {
		t.Error("执行失败")
		//return
	}

	var accountblklist []common.Address
	accountblklist = append(accountblklist, common.HexToAddress("0x12345"))
	if a.Check(mc.MSAccountBlackList, accountblklist) {
		fmt.Println(a.Output(mc.MSAccountBlackList, accountblklist))
	} else {
		t.Error("执行失败")
		//return
	}
	return
}
