package slash

import (
	"math/big"

	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/params"

	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/core/types"

	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/depoistInfo"
	"github.com/matrix/go-matrix/log"

	"github.com/matrix/go-matrix/common"
)

const PackageName = "惩罚"

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

type BlockSlash struct {
	chain            ChainReader
	eleMaxOnlineTime uint64
}

func New(chain ChainReader) *BlockSlash {
	return &BlockSlash{chain: chain, eleMaxOnlineTime: 97 * common.GetReElectionInterval() / common.GetBroadcastInterval()}
}

func (bp *BlockSlash) CalcSlash(currentState *state.StateDB, num uint64) map[common.Address]*big.Int {
	var eleNum uint64
	slashMap := make(map[common.Address]*big.Int)
	//选举周期的最后时刻分配
	if !common.IsReElectionNumber(num + 1) {
		return nil
	}
	//计算选举的拓扑图的高度
	if num < common.GetReElectionInterval() {
		eleNum = 0
	} else {
		eleNum = common.GetLastReElectionNumber(num) - 1
	}

	currentElectNodes, err := ca.GetTopologyByNumber(common.RoleValidator|common.RoleBackupValidator, eleNum)
	if err != nil {
		log.Error(PackageName, "get topology by number error", err)
		return nil
	}

	if 0 == len(currentElectNodes.NodeList) {
		log.Error(PackageName, "get NodeList is Nill", "")
		return nil
	}
	header := bp.chain.GetHeaderByNumber(eleNum)
	preState, err := bp.chain.StateAt(header.Root)

	if err != nil {
		return nil
	}
	for _, v := range currentElectNodes.NodeList {

		currentAccountReward, error := depoistInfo.GetInterest(currentState, v.Account)
		if nil != error {
			log.WARN(PackageName, "获取奖励错误，账户", v.Account)
			continue
		}
		preAccountReward, error := depoistInfo.GetInterest(preState, v.Account)
		if nil != error {
			log.WARN(PackageName, "获取奖励错误，账户", v.Account)
			continue
		}
		accountReward := new(big.Int).Sub(currentAccountReward, preAccountReward)
		if 0 == accountReward.Cmp(new(big.Int).SetUint64(0)) {
			log.WARN(PackageName, "获取奖励值为0，账户", v.Account)
			continue
		}
		preOnlineTime, err := depoistInfo.GetOnlineTime(preState, v.Account)
		if nil != err {
			log.WARN(PackageName, "获取起始uptime错误，账户", v.Account)
			continue
		}

		currentOnlineTime, err := depoistInfo.GetOnlineTime(currentState, v.Account)
		if nil != err {
			log.WARN(PackageName, "获取结束uptime错误，账户", v.Account)
			continue
		}

		temp := 1 - float64(currentOnlineTime.Uint64()-preOnlineTime.Uint64())/float64(bp.eleMaxOnlineTime)

		if temp >= 0.75 {
			temp = 0.75
		}
		slash := new(big.Int).SetUint64(uint64(float64(accountReward.Uint64()) * temp))
		log.INFO(PackageName, "账户", v.Account.String(), "惩罚金额", accountReward, "slash", slash)
		depoistInfo.AddSlash(currentState, v.Account, slash)
		slashMap[v.Account] = slash
	}
	return slashMap
}
