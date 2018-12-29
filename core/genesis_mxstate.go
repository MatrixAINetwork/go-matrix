package core

import (
	"encoding/binary"
	"math/big"

	"github.com/matrix/go-matrix/params/manparams"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/matrixstate"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/pkg/errors"
	"sort"
)

const (
	RewardFullRate = uint64(10000)
)

type GenesisMState struct {
	Broadcast            *common.Address         `json:"Broadcast"`
	InnerMiners          *[]common.Address       `json:"InnerMiners"`
	Foundation           *common.Address         `json:"Foundation"`
	VersionSuperAccounts *[]common.Address       `json:"VersionSuperAccounts"`
	BlockSuperAccounts   *[]common.Address       `json:"BlockSuperAccounts"`
	VIPCfg               *[]mc.VIPConfig         `json:"VIPCfg" gencodec:"required"`
	BCICfg               *mc.BCIntervalInfo      `json:"BroadcastInterval" gencodec:"required"`
	LeaderCfg            *mc.LeaderConfig        `json:"LeaderCfg" gencodec:"required"`
	BlkRewardCfg         *mc.BlkRewardCfg        `json:"BlkRewardCfg" gencodec:"required"`
	TxsRewardCfg         *mc.TxsRewardCfgStruct  `json:"TxsRewardCfg" gencodec:"required"`
	LotteryCfg           *mc.LotteryCfgStruct    `json:"LotteryCfg" gencodec:"required"`
	InterestCfg          *mc.InterestCfgStruct   `json:"InterestCfg" gencodec:"required"`
	SlashCfg             *mc.SlashCfgStruct      `json:"SlashCfg" gencodec:"required"`
	EleTimeCfg           *mc.ElectGenTimeStruct  `json:"EleTime" gencodec:"required"`
	EleInfoCfg           *mc.ElectConfigInfo     `json:"EleInfo" gencodec:"required"`
	ElectMinerNumCfg     *mc.ElectMinerNumStruct `json:"ElectMinerNum" gencodec:"required"`
	ElectBlackListCfg    *[]common.Address       `json:"ElectBlackList" gencodec:"required"`
	ElectWhiteListCfg    *[]common.Address       `json:"ElectWhiteList" gencodec:"required"`
	CurElect             *[]common.Elect         `json:"CurElect"  gencodec:"required"`
}
type GenesisMState1 struct {
	Broadcast            *string                 `json:"Broadcast,omitempty"`
	InnerMiners          *[]string               `json:"InnerMiners,omitempty"`
	Foundation           *string                 `json:"Foundation,omitempty"`
	VersionSuperAccounts *[]string               `json:"VersionSuperAccounts,omitempty"`
	BlockSuperAccounts   *[]string               `json:"BlockSuperAccounts,omitempty"`
	BCICfg               *mc.BCIntervalInfo      `json:"BroadcastInterval" gencodec:"required"`
	VIPCfg               *[]mc.VIPConfig         `json:"VIPCfg" ,omitempty"`
	LeaderCfg            *mc.LeaderConfig        `json:"LeaderCfg" ,omitempty"`
	BlkRewardCfg         *mc.BlkRewardCfg        `json:"BlkRewardCfg" ,omitempty"`
	TxsRewardCfg         *mc.TxsRewardCfgStruct  `json:"TxsRewardCfg" ,omitempty"`
	LotteryCfg           *mc.LotteryCfgStruct    `json:"LotteryCfg" ,omitempty"`
	InterestCfg          *mc.InterestCfgStruct   `json:"InterestCfg" ,omitempty"`
	SlashCfg             *mc.SlashCfgStruct      `json:"SlashCfg" ,omitempty"`
	EleTimeCfg           *mc.ElectGenTimeStruct  `json:"EleTime" ,omitempty"`
	EleInfoCfg           *mc.ElectConfigInfo     `json:"EleInfo" ,omitempty"`
	ElectMinerNumCfg     *mc.ElectMinerNumStruct `json:"ElectMinerNum" gencodec:"required"`
	ElectBlackListCfg    *[]string               `json:"ElectBlackList" gencodec:"required"`
	ElectWhiteListCfg    *[]string               `json:"ElectWhiteList" gencodec:"required"`
	CurElect             *[]common.Elect1        `json:"curElect"    gencodec:"required"`
}

func (ms *GenesisMState) setMatrixState(state *state.StateDB, netTopology common.NetTopology, nextElect []common.Elect, num uint64) error {
	if err := ms.setElectTime(state, num); err != nil {
		return err
	}

	if err := ms.setElectInfo(state, num); err != nil {
		return err
	}

	if err := ms.setElectMinerNumInfo(state, num); err != nil {
		return err
	}
	if err := ms.setElectBlackListInfo(state, num); err != nil {
		return err
	}
	if err := ms.setElectWhiteListInfo(state, num); err != nil {
		return err
	}

	if err := ms.setTopologyToState(state, netTopology, num); err != nil {
		return err
	}

	if err := ms.setElectToState(state, nextElect, num); err != nil {
		return err
	}

	if err := ms.setBroadcastAccountToState(state, num); err != nil {
		return err
	}
	if err := ms.setInnerMinerAccountsToState(state, num); err != nil {
		return err
	}
	if err := ms.setFoundationAccountToState(state, num); err != nil {
		return err
	}
	if err := ms.setVersionSuperAccountsToState(state, num); err != nil {
		return err
	}
	if err := ms.setBlockSuperAccountsToState(state, num); err != nil {
		return err
	}

	if err := ms.setBlkRewardCfgToState(state, num); err != nil {
		return err
	}
	if err := ms.setTxsRewardCfgToState(state, num); err != nil {
		return err
	}
	if err := ms.setLotteryCfgToState(state, num); err != nil {
		return err
	}
	if err := ms.setInterestCfgToState(state, num); err != nil {
		return err
	}
	if err := ms.setSlashCfgToState(state, num); err != nil {
		return err
	}
	if err := ms.setVIPCfgToState(state, num); err != nil {
		return err
	}
	if err := ms.setLeaderCfgToState(state, num); err != nil {
		return err
	}
	if err := ms.setBCIntervalToState(state, num); err != nil {
		return err
	}
	if err := ms.setPreMinHashToStat(state, num); err != nil {
		return err
	}
	if err := ms.setPreBroadcastRootToStat(state, num); err != nil {
		return err
	}
	return nil
}

func (g *GenesisMState) setElectTime(state *state.StateDB, num uint64) error {
	if g.EleTimeCfg == nil {
		if num == 0 {
			return errors.New("选举配置信息为nil")
		} else {
			log.INFO("Geneis", "没有配置选举信息", "")
			return nil
		}
	}
	if g.EleTimeCfg.ValidatorGen < g.EleTimeCfg.ValidatorNetChange {
		return errors.New("验证者切换点小于验证者生成点")
	}
	if g.EleTimeCfg.MinerGen < g.EleTimeCfg.MinerNetChange {
		return errors.New("矿工切换点小于矿工生效时间点")
	}
	log.Info("Geneis", "electime", g.EleTimeCfg)
	return matrixstate.SetDataToState(mc.MSKeyElectGenTime, g.EleTimeCfg, state)
}

func (g *GenesisMState) setElectInfo(state *state.StateDB, num uint64) error {
	if g.EleInfoCfg == nil {
		if num == 0 {
			return errors.New("electconfig配置信息为nil")
		} else {
			log.INFO("Geneis", "没有配置electconfig信息", "")
			return nil
		}
	}
	log.Info("Geneis", "electconfig", g.EleInfoCfg)
	return matrixstate.SetDataToState(mc.MSKeyElectConfigInfo, g.EleInfoCfg, state)
}

func (g *GenesisMState) setElectMinerNumInfo(state *state.StateDB, num uint64) error {
	if g.ElectMinerNumCfg == nil {
		if num == 0 {
			return errors.New("electMinerNum为nil")
		} else {
			log.Info("Geneis", "没有配置ElectMinerNumCfg信息", "")
			return nil
		}
	}
	log.Info("Geneis", "ElectMinerNumCfg", g.ElectMinerNumCfg)
	return matrixstate.SetDataToState(mc.MSKeyElectMinerNum, g.ElectMinerNumCfg, state)
}

func (g *GenesisMState) setElectWhiteListInfo(state *state.StateDB, num uint64) error {
	var whiteList []common.Address = nil
	if g.ElectWhiteListCfg == nil || *g.ElectWhiteListCfg == nil {
		if num == 0 {
			whiteList = make([]common.Address, 0)
		} else {
			log.Info("Geneis", "没有配置ElectWhiteListCfg信息", "")
			return nil
		}
	} else {
		whiteList = *g.ElectWhiteListCfg
	}

	log.Info("Geneis", "ElectWhiteListCfg", whiteList)
	return matrixstate.SetDataToState(mc.MSKeyElectWhiteList, whiteList, state)
}

func (g *GenesisMState) setElectBlackListInfo(state *state.StateDB, num uint64) error {
	var blackList []common.Address = nil
	if g.ElectBlackListCfg == nil || *g.ElectBlackListCfg == nil {
		if num == 0 {
			blackList = make([]common.Address, 0)
		} else {
			log.Info("Geneis", "没有配置ElectBlackListCfg信息", "")
			return nil
		}
	} else {
		blackList = *g.ElectBlackListCfg
	}

	log.Info("Geneis", "ElectBlackListCfg", blackList)
	return matrixstate.SetDataToState(mc.MSKeyElectBlackList, blackList, state)
}

func (g *GenesisMState) setTopologyToState(state *state.StateDB, genesisNt common.NetTopology, num uint64) error {
	if genesisNt.Type != common.NetTopoTypeAll {
		return nil
	}
	if len(genesisNt.NetTopologyData) == 0 {
		return errors.New("genesis net topology is empty！")
	}

	var newGraph *mc.TopologyGraph = nil
	var err error
	if num == 0 {
		newGraph, err = mc.NewGenesisTopologyGraph(num, genesisNt)
		if err != nil {
			return err
		}
	} else {

		data, err := matrixstate.GetDataByState(mc.MSKeyTopologyGraph, state)
		if err != nil {
			return errors.Errorf("get pre topology graph from state err: %v", err)
		}
		preGraph, _ := data.(*mc.TopologyGraph)
		if preGraph == nil {
			return errors.New("pre topology graph is nil")
		}
		newGraph, err = preGraph.Transfer2NextGraph(num, &genesisNt)
		if err != nil {
			return err
		}
	}

	if newGraph == nil {
		return errors.New("topology graph is nil")
	}
	return matrixstate.SetDataToState(mc.MSKeyTopologyGraph, newGraph, state)
}

func (g *GenesisMState) setElectToState(state *state.StateDB, nextElect []common.Elect, num uint64) error {
	var curElect []common.Elect = nil
	if g.CurElect != nil {
		curElect = *g.CurElect
	}

	if len(nextElect) == 0 && len(curElect) == 0 {
		return nil
	}

	elect := &mc.ElectGraph{
		Number:             num,
		ElectList:          make([]mc.ElectNodeInfo, 0),
		NextMinerElect:     make([]mc.ElectNodeInfo, 0),
		NextValidatorElect: make([]mc.ElectNodeInfo, 0),
	}

	minerIndex, validatorIndex, backUpValidatorIndex := uint16(0), uint16(0), uint16(0)
	for _, item := range nextElect {
		nodeInfo := mc.ElectNodeInfo{
			Account: item.Account,
			Stock:   item.Stock,
			Type:    item.Type.Transfer2CommonRole(),
		}
		switch item.Type {
		case common.ElectRoleMiner:
			nodeInfo.Position = common.GeneratePosition(minerIndex, item.Type)
			minerIndex++
			elect.NextMinerElect = append(elect.NextMinerElect, nodeInfo)
		case common.ElectRoleValidator:
			nodeInfo.Position = common.GeneratePosition(validatorIndex, item.Type)
			validatorIndex++
			elect.NextValidatorElect = append(elect.NextValidatorElect, nodeInfo)
		case common.ElectRoleValidatorBackUp:
			nodeInfo.Position = common.GeneratePosition(backUpValidatorIndex, item.Type)
			backUpValidatorIndex++
			elect.NextValidatorElect = append(elect.NextValidatorElect, nodeInfo)
		default:
			nodeInfo.Position = 0
		}
	}

	for _, item := range curElect {
		nodeInfo := mc.ElectNodeInfo{
			Account: item.Account,
			Stock:   item.Stock,
			Type:    item.Type.Transfer2CommonRole(),
		}
		switch item.Type {
		case common.ElectRoleMiner:
			nodeInfo.Position = common.GeneratePosition(minerIndex, item.Type)
			minerIndex++
		case common.ElectRoleValidator:
			nodeInfo.Position = common.GeneratePosition(validatorIndex, item.Type)
			validatorIndex++
		case common.ElectRoleValidatorBackUp:
			nodeInfo.Position = common.GeneratePosition(backUpValidatorIndex, item.Type)
			backUpValidatorIndex++
		default:
			nodeInfo.Position = 0
		}
		elect.ElectList = append(elect.ElectList, nodeInfo)
	}

	err := matrixstate.SetDataToState(mc.MSKeyElectGraph, elect, state)
	if err != nil {
		return err
	}

	electOnlineData := &mc.ElectOnlineStatus{
		Number: elect.Number,
	}
	for _, v := range elect.ElectList {
		tt := v
		tt.Position = common.PosOnline
		electOnlineData.ElectOnline = append(electOnlineData.ElectOnline, tt)
	}

	return matrixstate.SetDataToState(mc.MSKeyElectOnlineState, electOnlineData, state)
}

func (g *GenesisMState) setBroadcastAccountToState(state *state.StateDB, num uint64) error {
	if g.Broadcast == nil || *g.Broadcast == (common.Address{}) {
		if num == 0 {
			return errors.Errorf("the `broadcast` of genesis is empty")
		} else {
			return nil
		}
	}
	matrixstate.SetDataToState(mc.MSKeyAccountBroadcast, *g.Broadcast, state)
	return nil
}

func (g *GenesisMState) setInnerMinerAccountsToState(state *state.StateDB, num uint64) error {
	var innerMiners []common.Address = nil
	if g.InnerMiners == nil || *g.InnerMiners == nil {
		if num == 0 {
			innerMiners = make([]common.Address, 0)
		} else {
			return nil
		}
	} else {
		innerMiners = *g.InnerMiners
	}

	matrixstate.SetDataToState(mc.MSKeyAccountInnerMiners, innerMiners, state)
	return nil
}

func (g *GenesisMState) setFoundationAccountToState(state *state.StateDB, num uint64) error {
	var foundation common.Address
	if g.Foundation == nil || *g.Foundation == (common.Address{}) {
		if num == 0 {
			foundation = common.Address{}
		} else {
			return nil
		}
	} else {
		foundation = *g.Foundation
	}
	matrixstate.SetDataToState(mc.MSKeyAccountFoundation, foundation, state)
	return nil
}

func (g *GenesisMState) setVersionSuperAccountsToState(state *state.StateDB, num uint64) error {
	if g.VersionSuperAccounts == nil || len(*g.VersionSuperAccounts) == 0 {
		if num == 0 {
			return errors.Errorf("the version superAccounts of genesis is empty")
		} else {
			return nil
		}
	}
	matrixstate.SetDataToState(mc.MSKeyAccountVersionSupers, *g.VersionSuperAccounts, state)
	return nil
}

func (g *GenesisMState) setBlockSuperAccountsToState(state *state.StateDB, num uint64) error {
	if num != 0 {
		return errors.New("the block superAccounts can't modify")
	}

	if g.BlockSuperAccounts == nil || len(*g.BlockSuperAccounts) == 0 {
		return errors.Errorf("the block superAccounts of genesis is empty")
	}

	matrixstate.SetDataToState(mc.MSKeyAccountBlockSupers, *g.BlockSuperAccounts, state)
	return nil
}

func (g *GenesisMState) setBlkRewardCfgToState(state *state.StateDB, num uint64) error {

	if g.BlkRewardCfg == nil {
		if num == 0 {
			return errors.New("固定区块配置信息为nil")
		} else {
			log.INFO("Geneis", "没有配置固定区块配置信息", "")
			return nil
		}
	}

	rateCfg := g.BlkRewardCfg.RewardRate

	if RewardFullRate != rateCfg.MinerOutRate+rateCfg.ElectedMinerRate+rateCfg.FoundationMinerRate {

		return errors.Errorf("矿工固定区块奖励比例配置错误")
	}
	if RewardFullRate != rateCfg.LeaderRate+rateCfg.ElectedValidatorsRate+rateCfg.FoundationValidatorRate {

		return errors.Errorf("验证者固定区块奖励比例配置错误")
	}

	if RewardFullRate != rateCfg.OriginElectOfflineRate+rateCfg.BackupRewardRate {

		return errors.Errorf("替补固定区块奖励比例配置错误")
	}
	log.Info("Geneis", "BlkRewardCfg", g.BlkRewardCfg)
	minerOutReward := &mc.MinerOutReward{Reward: *big.NewInt(0)}
	matrixstate.SetDataToState(mc.MSKeyPreMinerBlkReward, minerOutReward, state)
	return matrixstate.SetDataToState(mc.MSKeyBlkRewardCfg, g.BlkRewardCfg, state)
}

func (g *GenesisMState) setTxsRewardCfgToState(state *state.StateDB, num uint64) error {
	if g.TxsRewardCfg == nil {
		if num == 0 {
			return errors.New("交易费区块配置信息为nil")
		} else {
			log.INFO("Geneis", "没有配置交易费区块配置信息", "")
			return nil
		}
	}
	rateCfg := g.TxsRewardCfg.RewardRate

	if RewardFullRate != g.TxsRewardCfg.ValidatorsRate+g.TxsRewardCfg.MinersRate {

		return errors.Errorf("交易奖励比例配置错误")
	}

	if RewardFullRate != rateCfg.MinerOutRate+rateCfg.ElectedMinerRate+rateCfg.FoundationMinerRate {

		return errors.Errorf("矿工固定区块奖励比例配置错误")
	}
	if RewardFullRate != rateCfg.LeaderRate+rateCfg.ElectedValidatorsRate+rateCfg.FoundationValidatorRate {

		return errors.Errorf("验证者固定区块奖励比例配置错误")
	}

	if RewardFullRate != rateCfg.OriginElectOfflineRate+rateCfg.BackupRewardRate {

		return errors.Errorf("替补固定区块奖励比例配置错误")
	}
	log.Info("Geneis", "TxsRewardCfg", g.TxsRewardCfg)
	minerOutReward := &mc.MinerOutReward{Reward: *big.NewInt(0)}
	matrixstate.SetDataToState(mc.MSKeyPreMinerTxsReward, minerOutReward, state)
	return matrixstate.SetDataToState(mc.MSKeyTxsRewardCfg, g.TxsRewardCfg, state)
}

func (g *GenesisMState) setLotteryCfgToState(state *state.StateDB, num uint64) error {
	if num == 0 {
		if g.LotteryCfg == nil {
			return errors.New("利息配置信息为nil")
		}
		account := &mc.LotteryFrom{From: make([]common.Address, 0)}
		matrixstate.SetDataToState(mc.MSKEYLotteryAccount, account, state)
		matrixstate.SetNumByState(mc.MSKEYLotteryNum, state, 1)
	} else {
		if g.LotteryCfg == nil {
			log.INFO("Geneis", "没有配置利息配置信息", "")
			return nil
		}
	}
	log.Info("Geneis", "LotteryCfg", g.LotteryCfg)
	return matrixstate.SetDataToState(mc.MSKeyLotteryCfg, g.LotteryCfg, state)
}

func (g *GenesisMState) setInterestCfgToState(state *state.StateDB, num uint64) error {
	if num == 0 {
		if g.InterestCfg == nil {
			return errors.New("利息配置信息为nil")
		}
		matrixstate.SetNumByState(mc.MSInterestCalcNum, state, 1)
		matrixstate.SetNumByState(mc.MSInterestPayNum, state, 1)
	} else {
		if g.InterestCfg == nil {
			log.INFO("Geneis", "没有配置利息配置信息", "")
			return nil
		}
	}
	StateCfg := g.InterestCfg

	if StateCfg.PayInterval < StateCfg.CalcInterval {

		return errors.Errorf("配置的发放周期小于计息周期")
	}

	log.Info("Geneis", "InterestCfg", g.InterestCfg)
	return matrixstate.SetDataToState(mc.MSKeyInterestCfg, g.InterestCfg, state)
}

func (g *GenesisMState) setSlashCfgToState(state *state.StateDB, num uint64) error {
	if num == 0 {
		if g.SlashCfg == nil {
			return errors.New("惩罚配置信息为nil")
		}
		matrixstate.SetNumByState(mc.MSKeySlashNum, state, 1)
		matrixstate.SetNumByState(mc.MSKeyUpTimeNum, state, 1)
	} else {
		if g.SlashCfg == nil {
			log.INFO("Geneis", "没有配置惩罚配置信息", "")
			return nil
		}

	}

	log.Info("Geneis", "SlashCfg", g.SlashCfg)
	return matrixstate.SetDataToState(mc.MSKeySlashCfg, g.SlashCfg, state)
}

type SortVIPConfig []mc.VIPConfig

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

func (g *GenesisMState) setVIPCfgToState(state *state.StateDB, number uint64) error {
	if g.VIPCfg == nil {
		if number == 0 {
			return errors.New("VIP配置信息为nil")
		} else {
			log.INFO("Geneis", "没有配置惩VIP配置信息", "")
			return nil
		}
	}
	if nil == g.VIPCfg || 0 == len(*g.VIPCfg) {

		return errors.Errorf("vip 配置为nil")
	}
	sort.Sort(SortVIPConfig(*g.VIPCfg))
	if (*g.VIPCfg)[0].MinMoney != uint64(0) {
		return errors.New("vip配置中需包含最小值为0的配置")
	}
	for index := 0; index < len(*g.VIPCfg)-1; index++ {
		if (*g.VIPCfg)[index].MinMoney == (*g.VIPCfg)[index+1].MinMoney {
			return errors.New("vip配置中不能包含最小值相同的配置")
		}
	}

	log.Info("Geneis", "VIPCfg", *g.VIPCfg)
	return matrixstate.SetDataToState(mc.MSKeyVIPConfig, *g.VIPCfg, state)
}

func (g *GenesisMState) setLeaderCfgToState(state *state.StateDB, num uint64) error {
	if g.LeaderCfg == nil {
		if num == 0 {
			return errors.New("leader配置信息为nil")
		} else {
			log.INFO("Geneis", "没有配置leader配置信息", "")
			return nil
		}
	}
	cfg := g.LeaderCfg
	if cfg.ParentMiningTime <= 0 {
		return errors.Errorf("`ParentMiningTime`(%d) of leader config illegal", cfg.ParentMiningTime)
	}
	if cfg.PosOutTime <= 0 {
		return errors.Errorf("`PosOutTime`(%d) of leader config illegal", cfg.PosOutTime)
	}
	if cfg.ReelectOutTime <= 0 {
		return errors.Errorf("`ReelectOutTime`(%d) of leader config illegal", cfg.ReelectOutTime)
	}
	if cfg.ReelectHandleInterval <= 0 {
		return errors.Errorf("`ReelectHandleInterval`(%d) of leader config illegal", cfg.ReelectHandleInterval)
	}

	log.Info("Geneis", "LeaderCfg", g.LeaderCfg)
	return matrixstate.SetDataToState(mc.MSKeyLeaderConfig, g.LeaderCfg, state)
}

func (g *GenesisMState) SetSuperBlkToState(state *state.StateDB, extra []byte, num uint64) error {
	var superBlkCfg *mc.SuperBlkCfg
	if num == 0 {
		superBlkCfg = &mc.SuperBlkCfg{Seq: 0, Num: 0}
	} else {
		if len(extra) < 8 {
			return errors.New("没有配置超级区块配置信息")
		}

		seq := uint64(binary.BigEndian.Uint64(extra[:8]))

		superBlkCfg = &mc.SuperBlkCfg{Seq: seq, Num: num}
	}
	log.INFO("Geneis", "超级区块配置", superBlkCfg)
	return matrixstate.SetDataToState(mc.MSKeySuperBlockCfg, superBlkCfg, state)
}

func (g *GenesisMState) setBCIntervalToState(state *state.StateDB, num uint64) error {
	var interval *mc.BCIntervalInfo = nil
	if num == 0 {
		if nil == g.BCICfg {
			return errors.New("广播周期配置信息为nil")
		}
		if g.BCICfg.BCInterval < 20 {
			return errors.Errorf("`BCInterval`(%d) of broadcast interval config illegal", g.BCICfg.BCInterval)
		}

		interval = &mc.BCIntervalInfo{
			LastBCNumber:       0,
			LastReelectNumber:  0,
			BCInterval:         g.BCICfg.BCInterval,
			BackupEnableNumber: 0,
			BackupBCInterval:   0,
		}
	} else {
		if nil == g.BCICfg {
			log.INFO("Geneis", "没有配置广播周期配置信息", "")
			return nil
		}
		if g.BCICfg.BackupBCInterval < 20 {
			return errors.Errorf("`BackupBCInterval`(%d) of broadcast interval config illegal", g.BCICfg.BackupBCInterval)
		}
		if g.BCICfg.BackupEnableNumber < num {
			return errors.Errorf("广播周期生效高度(%d)非法, < 当前高度(%d)", g.BCICfg.BackupEnableNumber, num)
		}

		preData, err := matrixstate.GetDataByState(mc.MSKeyBroadcastInterval, state)
		if err != nil {
			return errors.Errorf("获取前广播周期数据失败(%v)", err)
		}

		bcInterval, err := manparams.NewBCIntervalWithInterval(preData)
		if err != nil {
			return errors.Errorf("前广播周期数据异常(%v)", err)
		}

		if bcInterval.GetBroadcastInterval() == g.BCICfg.BackupBCInterval {
			log.INFO("GenesisMState", "广播周期一致，不配置", g.BCICfg.BackupBCInterval)
			return nil
		}

		if bcInterval.IsReElectionNumber(g.BCICfg.BackupBCInterval) {
			return errors.Errorf("生效高度(%d)必须是选举周期, 上个选举高度(%d), 原广播周期(%d)",
				g.BCICfg.BackupBCInterval, bcInterval.GetLastReElectionNumber(), bcInterval.GetBroadcastInterval())
		}

		bcInterval.SetBackupBCInterval(g.BCICfg.BackupBCInterval, g.BCICfg.BackupEnableNumber)
		interval = bcInterval.ToInfoStu()
	}

	if interval != nil {
		return matrixstate.SetDataToState(mc.MSKeyBroadcastInterval, interval, state)
	}
	return nil
}

func (g *GenesisMState) setPreMinHashToStat(state *state.StateDB, num uint64) error {
	return matrixstate.SetDataToState(mc.MSKeyMinHash, &mc.RandomInfoStruct{}, state)
}
func (g *GenesisMState) setPreBroadcastRootToStat(state *state.StateDB, num uint64) error {
	return matrixstate.SetDataToState(mc.MSKeyPreBroadcastRoot, &mc.PreBroadStateRoot{}, state)
}
