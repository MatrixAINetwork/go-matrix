package supertxsstate

import (
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/params/manparams"
)

const logInfo = "super txs state"

var mangerAlpha *SuperTxsStateManager
var mangerBeta *SuperTxsStateManager

func init() {
	mangerAlpha = newManager(manparams.VersionAlpha)
}

type SuperTxsStateChecker interface {
	Check(k, v interface{}) (interface{}, bool)
	Output(k, v interface{}) (interface{}, interface{})
}

type SuperTxsStateManager struct {
	version   string
	operators map[string]SuperTxsStateChecker
}

func GetManager(version string) *SuperTxsStateManager {
	switch version {
	case manparams.VersionAlpha:
		return mangerAlpha
	default:
		log.Error(logInfo, "get Manger err", "version not exist", "version", version)
		return nil
	}
}

//todo：将SuperTxsStateChecker写到opt处理对象里
func newManager(version string) *SuperTxsStateManager {
	switch version {
	case manparams.VersionAlpha:
		return &SuperTxsStateManager{
			version: version,
			operators: map[string]SuperTxsStateChecker{
				mc.MSKeyElectMinerNum:        new(mc.ElectMinerNumStruct),
				mc.MSKeyElectBlackList:       new(mc.ElectBlackList),
				mc.MSKeyElectWhiteList:       new(mc.ElectWhiteList),
				mc.MSKeyElectWhiteListSwitcher:new(mc.ElectWhiteListSwitcher),
				mc.MSKeyAccountBroadcasts:    new(mc.BroadcastAccounts),
				mc.MSKeyAccountInnerMiners:   new(mc.InnerMinersAccounts),
				mc.MSKeyVIPConfig:            new(mc.VIPConfig),
				mc.MSKeyBlkRewardCfg:         new(mc.BlkRewardCfg),
				mc.MSKeyTxsRewardCfg:         new(mc.TxsRewardCfg),
				mc.MSKeyInterestCfg:          new(mc.InterestCfg),
				mc.MSKeyLotteryCfg:           new(mc.LotteryCfg),
				mc.MSKeySlashCfg:             new(mc.SlashCfg),
				mc.MSKeyBlkCalc:              new(mc.BlkRewardCalc),
				mc.MSKeyTxsCalc:              new(mc.TxsRewardCalc),
				mc.MSKeyInterestCalc:         new(mc.InterestRewardCalc),
				mc.MSKeyLotteryCalc:          new(mc.LotteryRewardCalc),
				mc.MSKeySlashCalc:            new(mc.SlashCalc),
				mc.MSKeyBlockProduceSlashCfg: new(mc.BlockProduceSlashCfg),

				mc.MSTxpoolGasLimitCfg: new(mc.TxpoolGasLimit),
				mc.MSCurrencyPack:      new(mc.CurrencyPackLimt),
				mc.MSAccountBlackList:  new(mc.AccountBlackList),
			},
		}
	default:
		log.Error(logInfo, "创建管理类", "失败", "版本", version)
		return nil
	}
}

func (s *SuperTxsStateManager) Check(k string, v interface{}) (interface{}, bool) {
	if opt, ok := s.operators[k]; ok {
		return opt.Check(k, v)
	}
	return nil, false
}

func (s *SuperTxsStateManager) Output(k string, v interface{}) (interface{}, interface{}) {
	if opt, ok := s.operators[k]; ok {
		return opt.Output(k, v)
	}
	return k, v
}
