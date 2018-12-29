package core

import (
	"encoding/json"
	"fmt"
	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/common/readstatedb"
	"github.com/matrix/go-matrix/core/matrixstate"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/depoistInfo"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/params/manparams"
	"github.com/pkg/errors"
	"math/big"
)

func (bc *BlockChain) getUpTimeAccounts(num uint64, bcInterval *manparams.BCInterval) ([]common.Address, error) {
	originData, err := bc.GetMatrixStateDataByNumber(mc.MSKeyElectGenTime, num-1)
	if err != nil {
		log.ERROR(ModuleName, "获取选举生成点配置失败 err", err)
		return nil, err
	}
	electGenConf, Ok := originData.(*mc.ElectGenTimeStruct)
	if Ok == false {
		log.ERROR(ModuleName, "选举生成点信息失败 err", err)
		return nil, err
	}

	log.INFO(ModuleName, "获取所有参与uptime点名高度", num)

	upTimeAccounts := make([]common.Address, 0)

	minerNum := num - (num % bcInterval.GetBroadcastInterval()) - uint64(electGenConf.MinerGen)
	log.Debug(ModuleName, "参选矿工节点uptime高度", minerNum)
	ans, err := ca.GetElectedByHeightAndRole(big.NewInt(int64(minerNum)), common.RoleMiner)
	if err != nil {
		return nil, err
	}

	for _, v := range ans {
		upTimeAccounts = append(upTimeAccounts, v.Address)
		log.INFO("v.Address", "v.Address", v.Address)
	}
	validatorNum := num - (num % bcInterval.GetBroadcastInterval()) - uint64(electGenConf.ValidatorGen)
	log.Debug(ModuleName, "参选验证节点uptime高度", validatorNum)
	ans1, err := ca.GetElectedByHeightAndRole(big.NewInt(int64(validatorNum)), common.RoleValidator)
	if err != nil {
		return upTimeAccounts, err
	}
	for _, v := range ans1 {
		upTimeAccounts = append(upTimeAccounts, v.Address)
	}
	log.Debug(ModuleName, "获取所有uptime账户为", upTimeAccounts)
	return upTimeAccounts, nil
}
func (bc *BlockChain) getUpTimeData(root common.Hash, num uint64) (map[common.Address]uint32, map[common.Address][]byte, error) {

	heatBeatUnmarshallMMap, error := GetBroadcastTxMap(bc, root, mc.Heartbeat)
	if nil != error {
		log.WARN(ModuleName, "获取主动心跳交易错误", error)
	}
	//每个广播周期发一次
	calltherollUnmarshall, error := GetBroadcastTxMap(bc, root, mc.CallTheRoll)
	if nil != error {
		log.ERROR(ModuleName, "获取点名心跳交易错误", error)
		return nil, nil, error
	}
	calltherollMap := make(map[common.Address]uint32, 0)
	for _, v := range calltherollUnmarshall {
		temp := make(map[string]uint32, 0)
		error := json.Unmarshal(v, &temp)
		if nil != error {
			log.ERROR(ModuleName, "序列化点名心跳交易错误", error)
			return nil, nil, error
		}
		log.INFO(ModuleName, "点名心跳交易", temp)
		for k, v := range temp {
			calltherollMap[common.HexToAddress(k)] = v
		}
	}
	return calltherollMap, heatBeatUnmarshallMMap, nil
}
func (bc *BlockChain) handleUpTime(BeforeLastStateRoot common.Hash, state *state.StateDB, accounts []common.Address, calltherollRspAccounts map[common.Address]uint32, heatBeatAccounts map[common.Address][]byte, blockNum uint64, bcInterval *manparams.BCInterval) (map[common.Address]uint64, error) {
	HeartBeatMap := bc.getHeatBeatAccount(BeforeLastStateRoot, bcInterval, blockNum, accounts, heatBeatAccounts)

	originValidatorMap, originMinerMap, err := bc.getElectMap(blockNum, bcInterval)
	if nil != err {
		return nil, err
	}

	return bc.calcUpTime(accounts, calltherollRspAccounts, HeartBeatMap, bcInterval, state, originValidatorMap, originMinerMap), nil
}

func (bc *BlockChain) getElectMap(blockNum uint64, bcInterval *manparams.BCInterval) (map[common.Address]uint32, map[common.Address]uint32, error) {
	var eleNum uint64
	if blockNum < bcInterval.GetReElectionInterval()+2 {
		eleNum = 1
	} else {
		// 下一个选举+1
		eleNum = blockNum - bcInterval.GetBroadcastInterval()
	}
	electGraph, err := bc.GetMatrixStateDataByNumber(mc.MSKeyElectGraph, eleNum)
	if err != nil {
		log.Error(ModuleName, "获取拓扑图错误", err)
		return nil, nil, errors.New("获取拓扑图错误")
	}
	if electGraph == nil {
		log.Error(ModuleName, "获取拓扑图反射错误")
		return nil, nil, errors.New("获取拓扑图反射错误")
	}
	originElectNodes := electGraph.(*mc.ElectGraph)
	if 0 == len(originElectNodes.ElectList) {
		log.Error(ModuleName, "get获取初选列表为空", "")
		return nil, nil, errors.New("get获取初选列表为空")
	}
	log.Debug(ModuleName, "获取原始拓扑图所有的验证者和矿工，高度为", eleNum)
	originValidatorMap := make(map[common.Address]uint32, 0)
	originMinerMap := make(map[common.Address]uint32, 0)
	for _, v := range originElectNodes.ElectList {
		if v.Type == common.RoleValidator || v.Type == common.RoleBackupValidator {
			originValidatorMap[v.Account] = 0
		} else if v.Type == common.RoleMiner || v.Type == common.RoleBackupMiner {
			originMinerMap[v.Account] = 0
		}
	}
	return originValidatorMap, originMinerMap, nil
}

func (bc *BlockChain) getHeatBeatAccount(beforeLastStateRoot common.Hash, bcInterval *manparams.BCInterval, blockNum uint64, accounts []common.Address, heatBeatAccounts map[common.Address][]byte) map[common.Address]bool {
	HeatBeatReqAccounts := make([]common.Address, 0)
	HeartBeatMap := make(map[common.Address]bool, 0)
	//subVal就是最新的广播区块，例如当前区块高度是198或者是101，那么subVal就是100

	broadcastBlock := beforeLastStateRoot.Big()
	val := new(big.Int).Rem(broadcastBlock, big.NewInt(int64(bcInterval.GetBroadcastInterval())-1))
	for _, v := range accounts {
		currentAcc := v.Big()
		ret := new(big.Int).Rem(currentAcc, big.NewInt(int64(bcInterval.GetBroadcastInterval())-1))
		if ret.Cmp(val) == 0 {
			HeatBeatReqAccounts = append(HeatBeatReqAccounts, v)
			if _, ok := heatBeatAccounts[v]; ok {
				HeartBeatMap[v] = true
			} else {
				HeartBeatMap[v] = false

			}
			log.Debug(ModuleName, "计算主动心跳的账户", v, "心跳状态", HeartBeatMap[v])
		}
	}
	return HeartBeatMap
}

func (bc *BlockChain) calcUpTime(accounts []common.Address, calltherollRspAccounts map[common.Address]uint32, HeartBeatMap map[common.Address]bool, bcInterval *manparams.BCInterval, state *state.StateDB, originValidatorMap map[common.Address]uint32, originMinerMap map[common.Address]uint32) map[common.Address]uint64 {
	var upTime uint64
	maxUptime := bcInterval.GetBroadcastInterval() - 3
	upTimeMap := make(map[common.Address]uint64, 0)
	for _, account := range accounts {
		onlineBlockNum, ok := calltherollRspAccounts[account]
		if ok { //被点名,使用点名的uptime
			upTime = uint64(onlineBlockNum)
			log.INFO(ModuleName, "点名账号", account, "uptime", upTime)

		} else { //没被点名，没有主动上报，则为最大值，

			if v, ok := HeartBeatMap[account]; ok { //有主动上报
				if v {
					upTime = maxUptime
					log.Debug(ModuleName, "没被点名，有主动上报有响应", account, "uptime", upTime)
				} else {
					upTime = 0
					log.Debug(ModuleName, "没被点名，有主动上报无响应", account, "uptime", upTime)
				}
			} else { //没被点名和主动上报
				upTime = maxUptime
				log.Debug(ModuleName, "没被点名，没要求主动上报", account, "uptime", upTime)

			}
		}
		upTimeMap[account] = upTime
		// todo: add
		bc.saveUptime(account, upTime, state, originValidatorMap, originMinerMap)
	}
	return upTimeMap
}

//f(x)=ax+b
func (bc *BlockChain) upTimesReset(oldUpTime *big.Int, a float64, b int64) *big.Int {

	return big.NewInt(int64(a*float64(oldUpTime.Int64())) + b)

}
func (bc *BlockChain) saveUptime(account common.Address, upTime uint64, state *state.StateDB, originValidatorMap map[common.Address]uint32, originMinerMap map[common.Address]uint32) {
	old, err := depoistInfo.GetOnlineTime(state, account)
	if nil != err {
		return
	}
	log.Debug(ModuleName, "读取状态树", account, "upTime处理前", old)
	var newTime *big.Int
	if _, ok := originValidatorMap[account]; ok {

		newTime = bc.upTimesReset(old, 0.5, int64(upTime/2))
		log.Debug(ModuleName, "是原始验证节点，upTime减半", account, "upTime", newTime.Uint64())

	} else if _, ok := originMinerMap[account]; ok {
		newTime = bc.upTimesReset(old, 0.5, int64(upTime/2))
		log.Debug(ModuleName, "是原始矿工节点，upTime减半", account, "upTime", newTime.Uint64())

	} else {
		newTime = bc.upTimesReset(old, 1, int64(upTime))
		log.Debug(ModuleName, "其它节点，upTime累加", account, "upTime", newTime.Uint64())
	}

	depoistInfo.SetOnlineTime(state, account, newTime)

	depoistInfo.GetOnlineTime(state, account)
	log.Debug(ModuleName, "读取存入upTime账户", account, "upTime处理后", newTime.Uint64())
}
func (bc *BlockChain) HandleUpTimeWithSuperBlock(state *state.StateDB, accounts []common.Address, blockNum uint64, bcInterval *manparams.BCInterval) (map[common.Address]uint64, error) {
	broadcastInterval := bcInterval.GetBroadcastInterval()
	originTopologyNum := blockNum - blockNum%broadcastInterval - 1
	originTopology, err := ca.GetTopologyByNumber(common.RoleValidator|common.RoleBackupValidator|common.RoleMiner|common.RoleBackupMiner, originTopologyNum)
	if err != nil {
		return nil, err
	}
	originTopologyMap := make(map[common.Address]uint32, 0)
	for _, v := range originTopology.NodeList {
		originTopologyMap[v.Account] = 0
	}
	upTimeMap := make(map[common.Address]uint64, 0)
	for _, account := range accounts {

		upTime := broadcastInterval - 3
		log.Debug(ModuleName, "没被点名，没要求主动上报", account, "uptime", upTime)

		// todo: add
		depoistInfo.AddOnlineTime(state, account, new(big.Int).SetUint64(upTime))
		read, err := depoistInfo.GetOnlineTime(state, account)
		upTimeMap[account] = upTime
		if nil == err {
			log.Debug(ModuleName, "读取状态树", account, "upTime减半", read)
			if _, ok := originTopologyMap[account]; ok {
				updateData := new(big.Int).SetUint64(read.Uint64() / 2)
				log.INFO(ModuleName, "是原始拓扑图节点，upTime减半", account, "upTime", updateData.Uint64())
				depoistInfo.AddOnlineTime(state, account, updateData)
			}
		}

	}
	return upTimeMap, nil

}
func (bc *BlockChain) ProcessUpTime(state *state.StateDB, header *types.Header) (map[common.Address]uint64, error) {

	latestNum, err := matrixstate.GetNumByState(mc.MSKeyUpTimeNum, state)
	if nil != err {
		return nil, err
	}

	bcInterval, err := manparams.NewBCIntervalByHash(header.ParentHash)
	if err != nil {
		log.Error(ModuleName, "获取广播周期失败", err)
		return nil, err
	}

	if header.Number.Uint64() < bcInterval.GetBroadcastInterval() {
		return nil, err
	}
	sbh, err := bc.GetSuperBlockNum()
	if nil != err {
		return nil, errors.Errorf("get super seq error")
	}
	if latestNum < bcInterval.GetLastBroadcastNumber()+1 {
		log.Debug(ModuleName, "区块插入验证", "完成创建work, 开始执行uptime", "高度", header.Number.Uint64())
		matrixstate.SetNumByState(mc.MSKeyUpTimeNum, state, header.Number.Uint64())
		upTimeAccounts, err := bc.getUpTimeAccounts(header.Number.Uint64(), bcInterval)
		if err != nil {
			log.ERROR("core", "获取所有抵押账户错误!", err, "高度", header.Number.Uint64())
			return nil, err
		}
		if sbh < bcInterval.GetLastBroadcastNumber() &&
			sbh >= bcInterval.GetLastBroadcastNumber()-bcInterval.GetBroadcastInterval() {
			upTimeMap, err := bc.HandleUpTimeWithSuperBlock(state, upTimeAccounts, header.Number.Uint64(), bcInterval)
			if nil != err {
				log.ERROR("core", "处理uptime错误", err)
				return nil, err
			}
			return upTimeMap, nil
		} else {
			log.Debug(ModuleName, "获取所有心跳交易", "")
			preBroadcastRoot, err := readstatedb.GetPreBroadcastRoot(bc, header.Number.Uint64()-1)
			if err != nil {
				log.Error(ModuleName, "获取之前广播区块的root值失败 err", err)
				return nil, fmt.Errorf("从状态树获取前2个广播区块root失败")
			}
			log.Debug(ModuleName, "获取最新的root", preBroadcastRoot.LastStateRoot.Hex(), "上一个root", preBroadcastRoot.BeforeLastStateRoot)

			calltherollMap, heatBeatUnmarshallMMap, err := bc.getUpTimeData(preBroadcastRoot.LastStateRoot, header.Number.Uint64())
			if err != nil {
				log.WARN("core", "获取心跳交易错误!", err, "高度", header.Number.Uint64())
			}
			upTimeMap, err := bc.handleUpTime(preBroadcastRoot.BeforeLastStateRoot, state, upTimeAccounts, calltherollMap, heatBeatUnmarshallMMap, header.Number.Uint64(), bcInterval)
			if nil != err {
				log.ERROR("core", "处理uptime错误", err)
				return nil, err
			}
			return upTimeMap, nil
		}

	}

	return nil, nil
}
