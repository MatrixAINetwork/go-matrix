package matrix_trie

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/core/types"
	"github.com/pkg/errors"
)

const (
	BroadcastInterval = "broad_interval" // 广播区块周期
	BroadcastTx       = "broad_txs"      // 广播交易
)

var (
	ErrRootDBNil    = errors.New("root state db is nil")
	ErrMatrixDBNil  = errors.New("matrix state db is nil")
	ErrValueNotFind = errors.New("value not find is state db")
)

type stateReader interface {
	State() (*state.StateDB, error)
	StateAt(root common.Hash) (*state.StateDB, error)
	MatrixState() (*state.StateDB, error)
	MatrixStateAt(root common.Hash) (*state.StateDB, error)
}

const (
	matrixStatePrefix = "ms_"
)

type storePos uint8

const (
	storePosRoot   storePos = 1 // 存储在root树上，强共识
	storePosMatrix          = 2 // 存储在matrix树上，弱共识
)

type storeRule struct {
	pos     storePos
	keyHash common.Hash
}

func genKeyMap() (keyMap map[string]storeRule) {
	keyMap[BroadcastInterval] = storeRule{storePosRoot, types.RlpHash(matrixStatePrefix + BroadcastInterval)}
	keyMap[BroadcastTx] = storeRule{storePosMatrix, types.RlpHash(matrixStatePrefix + BroadcastTx)}
	return
}
