package selectedreward

import (
	"errors"
	"github.com/matrix/go-matrix/params/manparams"
	"math/big"

	"github.com/matrix/go-matrix/core/vm"

	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/params"

	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/reward/util"

	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/log"
)

const (
	PackageName = "参与奖励"
)

type SelectedReward struct {
}
type ChainReader interface {
	// Config retrieves the blockchain's chain configuration.
	Config() *params.ChainConfig

	// CurrentHeader retrieves the current header from the local chain.
	CurrentHeader() *types.Header

	// GetHeader retrieves a block header from the database by hash and number.
	GetHeader(hash common.Hash, number uint64) *types.Header

	// GetHeaderByNumber retrieves a block header from the database by number.
	GetHeaderByNumber(number uint64) *types.Header

	// GetHeaderByHash retrieves a block header from the database by its hash.
	GetHeaderByHash(hash common.Hash) *types.Header

	GetBlockByNumber(number uint64) *types.Block

	// GetBlock retrieves a block sfrom the database by hash and number.
	GetBlock(hash common.Hash, number uint64) *types.Block
	StateAt(root common.Hash) (*state.StateDB, error)
	State() (*state.StateDB, error)
	NewTopologyGraph(header *types.Header) (*mc.TopologyGraph, error)
}

func (sr *SelectedReward) getTopAndDeposit(chain util.ChainReader, state util.StateDB, currentNum uint64, roleType common.RoleType) ([]common.Address, map[common.Address]uint16, []vm.DepositDetail, error) {

	currentTop, originElectNodes, err := chain.GetGraphByState(state)
	if err != nil {
		log.Error(PackageName, "获取拓扑图错误", err)
		return nil, nil, nil, errors.New("获取拓扑图错误")
	}

	if originElectNodes == nil || 0 == len(originElectNodes.ElectList) {
		log.Error(PackageName, "get获取初选列表为空", "")
		return nil, nil, nil, errors.New("get获取初选列表为空")
	}

	if currentTop == nil || 0 == len(currentTop.NodeList) {
		log.Error(PackageName, "当前拓扑图是 空", "")
		return nil, nil, nil, errors.New("当前拓扑图是 空")
	}

	topNodes := make([]common.Address, 0)
	for _, node := range currentTop.NodeList {
		if node.Type == node.Type&roleType {
			topNodes = append(topNodes, node.Account)
		}
	}

	electNodes := make(map[common.Address]uint16, 0)
	for _, node := range originElectNodes.ElectList {
		if node.Type == node.Type&roleType {
			electNodes[node.Account] = node.Stock
		}
	}
	var depositNum uint64
	originInfo, err := chain.GetMatrixStateDataByNumber(mc.MSKeyElectGenTime, currentNum-1)

	if nil != err {
		return nil, nil, nil, errors.New("获取选举信息的出错")
	}

	bcInterval, err := manparams.NewBCIntervalByNumber(currentNum - 1)
	if err != nil {
		log.Error(PackageName, "获取广播周期失败", err)
		return nil, nil, nil, errors.New("获取广播周期失败")
	}

	if currentNum < bcInterval.GetReElectionInterval() {
		depositNum = 0
	} else {
		if common.RoleValidator == common.RoleValidator&roleType {
			depositNum = bcInterval.GetLastReElectionNumber() - uint64(originInfo.(*mc.ElectGenTimeStruct).ValidatorGen)
		} else {
			depositNum = bcInterval.GetLastReElectionNumber() - uint64(originInfo.(*mc.ElectGenTimeStruct).MinerGen)
		}
	}

	depositNodes, err := ca.GetElectedByHeightAndRole(new(big.Int).SetUint64(depositNum), roleType)
	if nil != err {
		log.ERROR(PackageName, "获取抵押列表错误", err)
		return nil, nil, nil, errors.New("获取抵押列表错误 ")
	}
	if 0 == len(depositNodes) {
		log.ERROR(PackageName, "获取抵押列表为空", "")
		return nil, nil, nil, errors.New("获取抵押列表为空 ")
	}
	return topNodes, electNodes, depositNodes, nil
}

func (sr *SelectedReward) GetSelectedRewards(reward *big.Int, state util.StateDB, chain util.ChainReader, roleType common.RoleType, currentNum uint64, rate uint64) map[common.Address]*big.Int {

	//计算选举的拓扑图的高度
	if reward.Cmp(big.NewInt(0)) <= 0 {
		log.WARN(PackageName, "奖励金额不合法", reward)
		return nil
	}
	log.Debug(PackageName, "参与奖励大家共发放", reward)

	currentTop, originElectNodes, depositNodes, err := sr.getTopAndDeposit(chain, state, currentNum, roleType)
	if nil != err {
		return nil
	}

	selectedNodesDeposit := sr.caclSelectedDeposit(currentTop, originElectNodes, depositNodes, rate)
	if 0 == len(selectedNodesDeposit) {
		log.Error(PackageName, "获取参与的抵押列表错误", "")
		return nil
	}

	return util.CalcStockRate(reward, selectedNodesDeposit)

}

func (sr *SelectedReward) caclSelectedDeposit(newGraph []common.Address, originElectNodes map[common.Address]uint16, depositNodes []vm.DepositDetail, rewardRate uint64) map[common.Address]util.DepositInfo {
	NodesRewardMap := make(map[common.Address]uint64, 0)
	for _, nodelist := range newGraph {
		NodesRewardMap[nodelist] = rewardRate
		log.Debug(PackageName, "当前节点", nodelist.Hex())
	}
	for electList := range originElectNodes {
		if _, ok := NodesRewardMap[electList]; ok {
			NodesRewardMap[electList] = util.RewardFullRate
		} else {
			NodesRewardMap[electList] = util.RewardFullRate - rewardRate
		}
		log.Debug(PackageName, "初选节点", electList.Hex(), "比例", NodesRewardMap[electList])
	}

	selectedNodesDeposit := make(map[common.Address]util.DepositInfo, 0)

	for _, v := range depositNodes {

		if depositRate, ok := NodesRewardMap[v.Address]; ok {
			if v.Deposit.Cmp(big.NewInt(0)) < 0 {
				log.ERROR(PackageName, "获取抵押值错误，抵押", v.Deposit, "账户", v.Address.Hex())
				return nil
			}
			deposit := util.CalcRateReward(v.Deposit, depositRate)
			var finalStock uint64
			if stock, ok := originElectNodes[v.Address]; ok {
				finalStock = uint64(stock) * depositRate
			} else {
				finalStock = depositRate
			}
			selectedNodesDeposit[v.Address] = util.DepositInfo{Deposit: deposit, FixStock: finalStock}
			log.Debug(PackageName, "计算抵押总额,账户", v.Address.Hex(), "抵押", deposit, "股权", finalStock)
		}
	}
	return selectedNodesDeposit
}
