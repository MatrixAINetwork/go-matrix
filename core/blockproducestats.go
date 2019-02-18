package core

import (
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/matrixstate"
	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/params/manparams"
	"github.com/pkg/errors"
)

var (
	SlashCfg                 = mc.BlockProduceSlashCfg{Switcher: true, LowTHR: 1, ProhibitCycleNum: 2}
	ErrWriteSlashCfgStateErr = errors.Errorf("Init Set SlashCfg Err")
	ErrGetSlashCfgStateErr   = errors.Errorf("Get SlashCfg Err")
	ErrStatePtrIsNil         = errors.Errorf("State Ptr Is Null")
	ErrHeaderPtrIsNil        = errors.Errorf("Header Ptr Is Null")
	ErrSlashCfgPtrIsNil      = errors.Errorf("Slash Cfg Ptr Is Null")
)

func (bc *BlockChain) slashCfgProc(state *state.StateDB, num uint64) (*mc.BlockProduceSlashCfg, bool, error) {
	if nil == state {
		return nil, false, ErrStatePtrIsNil
	}
	slashCfg, err := bc.GetSlashCfg(state)
	if nil != err {
		log.Crit(ModuleName, "Get BlockProduce Slash Cfg", ErrGetSlashCfgStateErr)
		return nil, false, ErrGetSlashCfgStateErr
	} else {
		return slashCfg, true, nil
	}
}
func (bc *BlockChain) ProcessBlockGProduceSlash(state *state.StateDB, header *types.Header) error {
	if nil == state {
		return ErrStatePtrIsNil
	}
	if nil == header {
		return ErrHeaderPtrIsNil
	}

	bcInterval, err := matrixstate.GetBroadcastInterval(state)
	if err != nil {
		return err
	}

	if bcInterval.IsBroadcastNumber(header.Number.Uint64()) {
		log.Debug("ProcessBlockGProduceSlash", "区块是广播区块", "跳过处理")
		return nil
	}

	slashCfg, _, err := bc.slashCfgProc(state, header.Number.Uint64())
	if nil == slashCfg {
		return err
	}

	if status, err := bc.shouldBlockProduceStatsStart(state, header.ParentHash, slashCfg); status {
		log.Trace(ModuleName, "执行初始化统计列表,高度", header.Number.Uint64(), "Err", err)
		initStatsList(state, header.Number.Uint64())
	}

	if statsList, ok := bc.shouldAddRecorder(state, slashCfg); ok {
		log.Trace(ModuleName, "增加出块统计,高度", header.Number.Uint64(), "账户", header.Leader.String(), "高度", header.Number.Uint64())
		statsListAddRecorder(state, statsList, header.Leader)
		statsListPrint(statsList)
		if ok := shouldBlockProduceSlash(state, header, slashCfg); ok {
			preBlackList := bc.GetBlackList(state)
			blackListPrint(preBlackList)

			var handleBlackList = NewBlackListMaintain(preBlackList.BlackList)
			handleBlackList.CounterMaintain()

			handleBlackList.AddBlackList(statsList.StatsList, slashCfg)
			log.Trace(ModuleName, "黑名单更新后状态，高度", header.Number.Uint64())
			blackListPrint(&mc.BlockProduceSlashBlackList{BlackList: handleBlackList.blacklist})
			if err := matrixstate.SetBlockProduceBlackList(state, &mc.BlockProduceSlashBlackList{BlackList: handleBlackList.blacklist}); err != nil {
				log.Crit("State Write Err : ", mc.MSKeyBlockProduceBlackList)
			}
		}
	}

	return nil
}
func statsListPrint(stats *mc.BlockProduceStats) {
	for _, v := range stats.StatsList {
		log.Debug(ModuleName, "Address", v.Address.String(), "Produce Block Num", v.ProduceNum)
	}
}
func blackListPrint(blackList *mc.BlockProduceSlashBlackList) {
	for _, v := range blackList.BlackList {
		log.Debug(ModuleName, "Address", v.Address.String(), "Ban", v.ProhibitCycleCounter)
	}
}

/*确定是否执行统计初始化：
配置不存在，或惩罚关闭，不执行
*/
func (bc *BlockChain) shouldBlockProduceStatsStart(currentState *state.StateDB, parentHash common.Hash, slashCfg *mc.BlockProduceSlashCfg) (bool, error) {
	if nil == currentState {
		return false, ErrStatePtrIsNil
	}
	if nil == slashCfg {
		return false, ErrSlashCfgPtrIsNil
	}
	//如果惩罚关闭，不执行
	if !slashCfg.Switcher {
		return false, nil
	}

	//如果初始化高度为空，需要执行
	latestUpdateTime, err := getLatestInitStatsNum(currentState)
	if 0 == latestUpdateTime {
		return true, nil
	}
	//如果初始化高度是上一个周期，需要执行
	hasInit, err := hasStatsInit(parentHash, latestUpdateTime)
	if nil != err {
		//状态数读取错误，不执行初始化。不会照成错误的黑名单
		return false, err
	} else {
		return !hasInit, nil
	}

}

func (bc *BlockChain) GetSlashCfg(state *state.StateDB) (*mc.BlockProduceSlashCfg, error) {
	slashCfg, err := matrixstate.GetBlockProduceSlashCfg(state)
	if err != nil {
		log.Error(ModuleName, "获取区块生产惩罚配置失败", err)
		return &mc.BlockProduceSlashCfg{}, err
	}

	return slashCfg, nil
}

func getSlashStatsList(state *state.StateDB) (*mc.BlockProduceStats, error) {
	if nil == state {
		return nil, ErrStatePtrIsNil
	}

	statsInfo, err := matrixstate.GetBlockProduceStats(state)
	if nil != err {
		log.ERROR(ModuleName, "获取区块惩罚统计信息错误", err)
		return &mc.BlockProduceStats{}, err
	}

	return statsInfo, nil
}
func getLatestInitStatsNum(state *state.StateDB) (uint64, error) {
	if nil == state {
		return 0, ErrStatePtrIsNil
	}
	updateInfo, err := matrixstate.GetBlockProduceStatsStatus(state)
	if nil != err {
		log.DEBUG(ModuleName, "获取区块生产统计错误", err)
		return 0, err
	}

	return updateInfo.Number, nil
}
func hasStatsInit(parentHash common.Hash, latestUpdateTime uint64) (bool, error) {
	bcInterval, err := manparams.GetBCIntervalInfoByHash(parentHash)
	if err != nil {
		log.Error(ModuleName, "获取广播周期失败", err)
		return false, err
	}

	if latestUpdateTime < bcInterval.GetLastReElectionNumber()+1 {
		return false, nil
	} else {
		return true, nil
	}

}
func initStatsList(state *state.StateDB, updateNumber uint64) {
	if nil == state {
		log.Error(ModuleName, "Input state ptr ", nil)
		return
	}
	//获取当前的初选主节点
	currElectInfo, err := matrixstate.GetElectGraph(state)
	statsList := mc.BlockProduceStats{}

	if nil != err {
		log.ERROR(ModuleName, "获取状态树选举信息错误", err)
		matrixstate.SetBlockProduceStats(state, &statsList)
		matrixstate.SetBlockProduceStatsStatus(state, &mc.BlockProduceSlashStatsStatus{updateNumber})
		return
	}

	for _, v := range currElectInfo.ElectList {
		if v.Type == common.RoleValidator {
			statsList.StatsList = append(statsList.StatsList, mc.UserBlockProduceNum{Address: v.Account, ProduceNum: 0})
		}
	}
	matrixstate.SetBlockProduceStats(state, &statsList)
	matrixstate.SetBlockProduceStatsStatus(state, &mc.BlockProduceSlashStatsStatus{updateNumber})
}
func (bc *BlockChain) shouldAddRecorder(state *state.StateDB, slashCfg *mc.BlockProduceSlashCfg) (*mc.BlockProduceStats, bool) {
	if !slashCfg.Switcher {
		return nil, false
	}
	if statsList, err := getSlashStatsList(state); nil != err {
		return statsList, false
	} else {
		return statsList, true
	}
}

func statsListAddRecorder(state *state.StateDB, list *mc.BlockProduceStats, userAddress common.Address) {
	for k, v := range list.StatsList {
		if v.Address.Equal(userAddress) {
			list.StatsList[k].ProduceNum = list.StatsList[k].ProduceNum + 1
		}
	}
	err := matrixstate.SetBlockProduceStats(state, list)
	if err != nil {
		log.Error(ModuleName, "写区块生成错误", err)
	}
}

func shouldBlockProduceSlash(state *state.StateDB, header *types.Header, slashCfg *mc.BlockProduceSlashCfg) bool {
	if !slashCfg.Switcher {
		log.Debug(ModuleName, "不执行惩罚，原因", "配置关闭")
		return false
	}
	if isNextBlockValidatorGenTimming(state, header.Number.Uint64(), header.ParentHash) {
		log.Trace(ModuleName, "执行惩罚，原因高度", header.Number.Uint64())
		return true
	}
	return false
}

func getElectTimingCfg(state *state.StateDB) (*mc.ElectGenTimeStruct, error) {
	electTiming, err := matrixstate.GetElectGenTime(state)
	if nil != err {
		log.ERROR(ModuleName, "读取选举时序错误", err)
		return nil, err
	}

	return electTiming, nil
}
func isNextBlockValidatorGenTimming(state *state.StateDB, currNum uint64, parentHash common.Hash) bool {

	bcInterval, err := manparams.GetBCIntervalInfoByHash(parentHash)
	if nil != err {
		log.ERROR(ModuleName, "获取广播配置 err", err)
		return false
	}

	electTiming, err := getElectTimingCfg(state)
	if nil != err {
		log.ERROR(ModuleName, "获取选举时序错误", err)
		return false
	}

	return bcInterval.IsReElectionNumber(currNum + 1 + uint64(electTiming.ValidatorGen))
}

func (bc *BlockChain) GetBlackList(state *state.StateDB) *mc.BlockProduceSlashBlackList {
	blackList, err := matrixstate.GetBlockProduceBlackList(state)
	if nil != err {
		log.Error(ModuleName, "Get Block Produce BlackList State Err", err)
		return &mc.BlockProduceSlashBlackList{}
	}
	return blackList
}

type blacklistMaintain struct {
	blacklist []mc.UserBlockProduceSlash
}

func NewBlackListMaintain(list []mc.UserBlockProduceSlash) *blacklistMaintain {
	var bl = new(blacklistMaintain)

	for _, v := range list {
		if v.ProhibitCycleCounter > 0 {
			bl.blacklist = append(bl.blacklist, v)
		}
	}
	return bl
}

func (bl *blacklistMaintain) CounterMaintain() {
	for i := 0; i < len(bl.blacklist); i++ {
		if 0 != bl.blacklist[i].ProhibitCycleCounter {
			bl.blacklist[i].ProhibitCycleCounter--
		}
	}
}

func searchExistAddress(statsList []mc.UserBlockProduceSlash, target common.Address) (int, bool) {

	for k, v := range statsList {
		if v.Address.Equal(target) {
			return k, true
		}
	}

	return 0, false
}
func (bl *blacklistMaintain) AddBlackList(statsList []mc.UserBlockProduceNum, slashCfg *mc.BlockProduceSlashCfg) {
	if nil == bl.blacklist {
		log.Warn(ModuleName, "blacklist Err", "Is Nil")
		bl.blacklist = make([]mc.UserBlockProduceSlash, 0)
	}

	if nil == slashCfg {
		log.Warn(ModuleName, "惩罚配置错误", "未配置")
		return
	}

	if false == slashCfg.Switcher {
		bl.blacklist = make([]mc.UserBlockProduceSlash, 0)
		return
	}

	if nil == statsList {
		log.Error(ModuleName, "统计列表错误", "未配置")
		return
	}

	if 0 == slashCfg.ProhibitCycleNum {
		log.Warn(ModuleName, "禁止周期为", "0", "不加入黑名单")
		return
	}
	for _, v := range statsList {
		if v.ProduceNum >= slashCfg.LowTHR {
			continue
		}
		if position, exist := searchExistAddress(bl.blacklist, v.Address); exist {
			bl.blacklist[position].ProhibitCycleCounter = slashCfg.ProhibitCycleNum - 1
		} else {
			bl.blacklist = append(bl.blacklist, mc.UserBlockProduceSlash{Address: v.Address, ProhibitCycleCounter: slashCfg.ProhibitCycleNum - 1})
		}
	}
}
