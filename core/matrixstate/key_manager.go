package matrixstate

import (
	"encoding/binary"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/mc"
	"github.com/pkg/errors"
)

func GetKeyHash(key string) common.Hash {
	hash, OK := km.keys[key]
	if OK {
		return hash
	} else {
		return common.Hash{}
	}
}

func GetDataByState(key string, state StateDB) (interface{}, error) {
	hash := GetKeyHash(key)
	if (hash == common.Hash{}) {
		return nil, errors.Errorf("key(%s) not find", key)
	}
	codec, exist := km.codecMap[key]
	if !exist {
		return nil, errors.Errorf("codec of key(%s) not find", key)
	}

	bytes := state.GetMatrixData(hash)
	if len(bytes) == 0 {
		return nil, errors.Errorf("no data in state of key(%s)", key)
	}

	return codec.decodeFn(bytes)
}

func GetNumByState(key string, state StateDB) (uint64, error) {
	bytes := state.GetMatrixData(common.BytesToHash([]byte(key)))
	if len(bytes) <= 0 {
		return 0, errors.Errorf("no data in state of key(%s)", key)
	}

	return uint64(binary.BigEndian.Uint64(bytes[:8])), nil
}

func SetNumByState(key string, state StateDB, num uint64) error {
	data := make([]byte, 8)
	binary.BigEndian.PutUint64(data, num)
	state.SetMatrixData(common.BytesToHash([]byte(key)), data)
	return nil
}

func SetDataToState(key string, data interface{}, state StateDB) error {
	hash := GetKeyHash(key)
	if (hash == common.Hash{}) {
		return errors.Errorf("key(%s) not find", key)
	}
	codec, exist := km.codecMap[key]
	if !exist {
		return errors.Errorf("codec of key(%s) not find", key)
	}

	bytes, err := codec.encodeFn(data)
	if err != nil {
		return errors.Errorf("encode data of key(%s) err: %v", key, err)
	}

	if len(bytes) == 0 {
		return errors.Errorf("the encoded data of key(%s) is empty", key)
	}

	state.SetMatrixData(hash, bytes)
	return nil
}

const (
	matrixStatePrefix = "ms_"
)

var km *keyManager

func init() {
	km = newKeyManager()
}

type keyManager struct {
	keys     map[string]common.Hash
	codecMap map[string]codec
}

func newKeyManager() *keyManager {
	km := &keyManager{
		keys: map[string]common.Hash{
			mc.MSKeyBroadcastTx:          types.RlpHash(matrixStatePrefix + mc.MSKeyBroadcastTx),
			mc.MSKeyTopologyGraph:        types.RlpHash(matrixStatePrefix + mc.MSKeyTopologyGraph),
			mc.MSKeyElectGraph:           types.RlpHash(matrixStatePrefix + mc.MSKeyElectGraph),
			mc.MSKeyElectOnlineState:     types.RlpHash(matrixStatePrefix + mc.MSKeyElectOnlineState),
			mc.MSKeyBroadcastInterval:    types.RlpHash(matrixStatePrefix + mc.MSKeyBroadcastInterval),
			mc.MSKeyElectGenTime:         types.RlpHash(matrixStatePrefix + mc.MSKeyElectGenTime),
			mc.MSKeyAccountBroadcast:     types.RlpHash(matrixStatePrefix + mc.MSKeyAccountBroadcast),
			mc.MSKeyAccountInnerMiners:   types.RlpHash(matrixStatePrefix + mc.MSKeyAccountInnerMiners),
			mc.MSKeyAccountFoundation:    types.RlpHash(matrixStatePrefix + mc.MSKeyAccountFoundation),
			mc.MSKeyAccountVersionSupers: types.RlpHash(matrixStatePrefix + mc.MSKeyAccountVersionSupers),
			mc.MSKeyAccountBlockSupers:   types.RlpHash(matrixStatePrefix + mc.MSKeyAccountBlockSupers),
			mc.MSKeyElectConfigInfo:      types.RlpHash(matrixStatePrefix + mc.MSKeyElectConfigInfo),
			mc.MSKeyElectMinerNum:        types.RlpHash(matrixStatePrefix + mc.MSKeyElectMinerNum),
			mc.MSKeyElectBlackList:       types.RlpHash(matrixStatePrefix + mc.MSKeyElectBlackList),
			mc.MSKeyElectWhiteList:       types.RlpHash(matrixStatePrefix + mc.MSKeyElectWhiteList),
			mc.MSKeyVIPConfig:            types.RlpHash(matrixStatePrefix + mc.MSKeyVIPConfig),

			mc.MSKeyPreBroadcastRoot:  types.RlpHash(matrixStatePrefix + mc.MSKeyPreBroadcastRoot),
			mc.MSKeyMinHash:           types.RlpHash(matrixStatePrefix + mc.MSKeyMinHash),
			mc.MSKeyPerAllTop:         types.RlpHash(matrixStatePrefix + mc.MSKeyPerAllTop),
			mc.MSKeyPreMinerBlkReward: types.RlpHash(matrixStatePrefix + mc.MSKeyPreMinerBlkReward),
			mc.MSKeyPreMinerTxsReward: types.RlpHash(matrixStatePrefix + mc.MSKeyPreMinerTxsReward),
			mc.MSKEYLotteryAccount:    types.RlpHash(matrixStatePrefix + mc.MSKEYLotteryAccount),

			mc.MSKeyLeaderConfig:  types.RlpHash(matrixStatePrefix + mc.MSKeyLeaderConfig),
			mc.MSKeyBlkRewardCfg:  types.RlpHash(matrixStatePrefix + mc.MSKeyBlkRewardCfg),
			mc.MSKeyTxsRewardCfg:  types.RlpHash(matrixStatePrefix + mc.MSKeyTxsRewardCfg),
			mc.MSKeyInterestCfg:   types.RlpHash(matrixStatePrefix + mc.MSKeyInterestCfg),
			mc.MSKeyLotteryCfg:    types.RlpHash(matrixStatePrefix + mc.MSKeyLotteryCfg),
			mc.MSKeySlashCfg:      types.RlpHash(matrixStatePrefix + mc.MSKeySlashCfg),
			mc.MSKeyMultiCoin:     types.RlpHash(matrixStatePrefix + mc.MSKeyMultiCoin),
			mc.MSKeySuperBlockCfg: types.RlpHash(matrixStatePrefix + mc.MSKeySuperBlockCfg),
		},
		codecMap: make(map[string]codec),
	}
	km.initCodec()
	return km
}
