package core

import (
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/matrixstate"
	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/pkg/errors"
)

// State returns a new mutable state based on the current HEAD block.
func (bc *BlockChain) State() (*state.StateDBManage, error) {
	return bc.StateAt(bc.CurrentBlock().Root())
}

// StateAt returns a new mutable state based on a particular point in time.
func (bc *BlockChain) StateAt(root []common.CoinRoot) (*state.StateDBManage, error) {
	return state.NewStateDBManage(root, bc.db, bc.stateCache)
}

func (bc *BlockChain) StateAtNumber(number uint64) (*state.StateDBManage, error) {
	block := bc.GetBlockByNumber(number)
	if block == nil {
		return nil, errors.Errorf("can't find block by number(%d)", number)
	}
	return bc.StateAt(block.Root())
}

func (bc *BlockChain) StateAtBlockHash(hash common.Hash) (*state.StateDBManage, error) {
	block := bc.GetBlockByHash(hash)
	if block == nil {
		return nil, errors.New("can't find block by hash")
	}
	return bc.StateAt(block.Root())
}

func (bc *BlockChain) RegisterMatrixStateDataProducer(key string, producer ProduceMatrixStateDataFn) {
	bc.matrixProcessor.RegisterProducer(key, producer)
}

func (bc *BlockChain) ProcessStateVersion(version []byte, st *state.StateDBManage) error {
	return bc.matrixProcessor.ProcessStateVersion(version, st)
}

func (bc *BlockChain) ProcessMatrixState(block *types.Block, preVersion string, state *state.StateDBManage) error {
	return bc.matrixProcessor.ProcessMatrixState(block, preVersion, state)
}

func (bc *BlockChain) GetGraphByHash(hash common.Hash) (*mc.TopologyGraph, *mc.ElectGraph, error) {
	topologyGraph, err := bc.topologyStore.GetTopologyGraphByHash(hash)
	if err != nil {
		return nil, nil, err
	}
	electGraph, err := bc.topologyStore.GetElectGraphByHash(hash)
	if err != nil {
		return nil, nil, err
	}
	return topologyGraph, electGraph, nil
}

func (bc *BlockChain) GetGraphByState(state matrixstate.StateDB) (*mc.TopologyGraph, *mc.ElectGraph, error) {
	topologyGraph, err := matrixstate.GetTopologyGraph(state)
	if err != nil {
		return nil, nil, err
	}
	electGraph, err := matrixstate.GetElectGraph(state)
	if err != nil {
		return nil, nil, err
	}
	return topologyGraph, electGraph, nil
}

func (bc *BlockChain) GetTopologyStore() *TopologyStore {
	return bc.topologyStore
}

func (bc *BlockChain) GetBroadcastInterval() (*mc.BCIntervalInfo, error) {
	st, err := bc.State()
	if err != nil {
		return nil, errors.Errorf("get cur state err(%v)", err)
	}
	return matrixstate.GetBroadcastInterval(st)
}

func (bc *BlockChain) GetBroadcastIntervalByHash(blockHash common.Hash) (*mc.BCIntervalInfo, error) {
	st, err := bc.StateAtBlockHash(blockHash)
	if err != nil {
		return nil, errors.Errorf("get state by hash(%s) err(%v)", blockHash.Hex(), err)
	}
	return matrixstate.GetBroadcastInterval(st)
}

func (bc *BlockChain) GetBroadcastIntervalByNumber(number uint64) (*mc.BCIntervalInfo, error) {
	st, err := bc.StateAtNumber(number)
	if err != nil {
		return nil, errors.Errorf("get state by number(%d) err(%v)", number, err)
	}
	return matrixstate.GetBroadcastInterval(st)
}

func (bc *BlockChain) GetBroadcastAccounts(blockHash common.Hash) ([]common.Address, error) {
	st, err := bc.StateAtBlockHash(blockHash)
	if err != nil {
		return nil, errors.Errorf("get state by hash(%s) err(%v)", blockHash.Hex(), err)
	}
	return matrixstate.GetBroadcastAccounts(st)
}

func (bc *BlockChain) GetInnerMinerAccounts(blockHash common.Hash) ([]common.Address, error) {
	st, err := bc.StateAtBlockHash(blockHash)
	if err != nil {
		return nil, errors.Errorf("get state by hash(%s) err(%v)", blockHash.Hex(), err)
	}
	return matrixstate.GetInnerMinerAccounts(st)
}

func (bc *BlockChain) GetVersionSuperAccounts(blockHash common.Hash) ([]common.Address, error) {
	st, err := bc.StateAtBlockHash(blockHash)
	if err != nil {
		return nil, errors.Errorf("get state by hash(%s) err(%v)", blockHash.Hex(), err)
	}
	return matrixstate.GetVersionSuperAccounts(st)
}

func (bc *BlockChain) GetMultiCoinSuperAccounts(blockHash common.Hash) ([]common.Address, error) {
	st, err := bc.StateAtBlockHash(blockHash)
	if err != nil {
		return nil, errors.Errorf("get state by hash(%s) err(%v)", blockHash.Hex(), err)
	}
	return matrixstate.GetMultiCoinSuperAccounts(st)
}

func (bc *BlockChain) GetSubChainSuperAccounts(blockHash common.Hash) ([]common.Address, error) {
	st, err := bc.StateAtBlockHash(blockHash)
	if err != nil {
		return nil, errors.Errorf("get state by hash(%s) err(%v)", blockHash.Hex(), err)
	}
	return matrixstate.GetSubChainSuperAccounts(st)
}

func (bc *BlockChain) GetBlockSuperAccounts(blockHash common.Hash) ([]common.Address, error) {
	st, err := bc.StateAtBlockHash(blockHash)
	if err != nil {
		return nil, errors.Errorf("get state err by hash(%s) err(%v)", blockHash.Hex(), err)
	}
	return matrixstate.GetBlockSuperAccounts(st)
}

func (bc *BlockChain) GetSuperBlockSeq() (uint64, error) {
	st, err := bc.State()
	if err != nil {
		return 0, errors.Errorf("get cur state err (%v)", err)
	}
	superBlkCfg, err := matrixstate.GetSuperBlockCfg(st)
	if err != nil {
		return 0, err
	}
	log.INFO("blockChain", "超级区块序号", superBlkCfg.Seq)
	return superBlkCfg.Seq, nil
}

func (bc *BlockChain) GetSuperBlockNum() (uint64, error) {
	st, err := bc.State()
	if err != nil {
		return 0, errors.Errorf("get cur state err (%v)", err)
	}
	superBlkCfg, err := matrixstate.GetSuperBlockCfg(st)
	if err != nil {
		return 0, err
	}
	log.INFO("blockChain", "超级区块高度", superBlkCfg.Num)
	return superBlkCfg.Num, nil
}

func (bc *BlockChain) GetSuperBlockInfo() (*mc.SuperBlkCfg, error) {
	st, err := bc.State()
	if err != nil {
		return nil, errors.Errorf("get cur state err (%v)", err)
	}
	superBlkCfg, err := matrixstate.GetSuperBlockCfg(st)
	if err != nil {
		return nil, err
	}
	log.Trace("blockChain", "超级区块高度", superBlkCfg.Num, "超级区块序号", superBlkCfg.Seq)
	return superBlkCfg, nil
}

/*func (bc *BlockChain) GetVersionByHash(blockHash common.Hash) (string, error) {
	st, err := bc.StateAtBlockHash(blockHash)
	if err != nil {
		return "", errors.Errorf("get state by hash(%s) err(%v)", blockHash.Hex(), err)
	}
	return matrixstate.GetVersionInfo(st), nil
}*/

func ProduceBroadcastIntervalData(block *types.Block, readFn PreStateReadFn) (interface{}, error) {
	bciData, err := readFn(mc.MSKeyBroadcastInterval)
	if err != nil {
		log.Error("ProduceBroadcastIntervalData", "read pre broadcast interval err", err)
		return nil, err
	}

	bcInterval, OK := bciData.(*mc.BCIntervalInfo)
	if OK == false {
		return nil, errors.New("pre broadcast interval reflect failed")
	}

	modify := false
	number := block.NumberU64()
	backupEnableNumber := bcInterval.GetBackupEnableNumber()
	if number == backupEnableNumber {
		// 备选生效时间点
		if bcInterval.IsReElectionNumber(number) == false || bcInterval.IsBroadcastNumber(number) == false {
			// 生效时间点不是原周期的选举点，数据错误
			log.Crit("ProduceBroadcastIntervalData", "backup enable number illegal", backupEnableNumber,
				"old interval", bcInterval.GetBroadcastInterval(), "last broadcast number", bcInterval.GetLastBroadcastNumber(), "last reelect number", bcInterval.GetLastReElectionNumber())
		}

		oldInterval := bcInterval.GetBroadcastInterval()

		// 设置最后的广播区块和选举区块
		bcInterval.SetLastBCNumber(backupEnableNumber)
		bcInterval.SetLastReelectNumber(backupEnableNumber)
		// 启动备选周期
		bcInterval.UsingBackupInterval()
		log.INFO("ProduceBroadcastIntervalData", "old interval", oldInterval, "new interval", bcInterval.GetBroadcastInterval())
		modify = true
	} else {
		if bcInterval.IsBroadcastNumber(number) {
			bcInterval.SetLastBCNumber(number)
			modify = true
		}

		if bcInterval.IsReElectionNumber(number) {
			bcInterval.SetLastReelectNumber(number)
			modify = true
		}
	}

	if modify {
		log.INFO("ProduceBroadcastIntervalData", "生成广播区块内容", "成功", "block number", number, "data", bcInterval)
		return bcInterval, nil
	} else {
		return nil, nil
	}
}
