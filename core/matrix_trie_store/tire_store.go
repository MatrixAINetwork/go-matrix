package matrix_trie

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/state"
	"github.com/pkg/errors"
)

type TrieStore struct {
	keyMap map[string]storeRule
	reader stateReader
}

func NewTrieStore(stReader stateReader) *TrieStore {
	return &TrieStore{
		keyMap: genKeyMap(),
		reader: stReader,
	}
}

func (ts *TrieStore) StoreState(key string, value []byte, rootDB *state.StateDB, matrixDB *state.StateDB) error {
	rule, err := ts.getStoreRule(key)
	if err != nil {
		return err
	}
	return ts.writeToDB(rule.pos, rule.keyHash, value, rootDB, matrixDB)
}

func (ts *TrieStore) GetState(key string) ([]byte, error) {
	rule, err := ts.getStoreRule(key)
	if err != nil {
		return nil, err
	}

	var value []byte
	switch rule.pos {
	case storePosRoot: // root树上
		rootDB, err := ts.reader.State()
		if err != nil || nil == rootDB {
			return nil, errors.Errorf("get state DB err(%v)", err)
		}
		value = rootDB.GetMatrixData(rule.keyHash)
	case storePosMatrix: // matrix树上
		matrixDB, err := ts.reader.MatrixState()
		if err != nil || nil == matrixDB {
			return nil, errors.Errorf("get matrix state DB err(%v)", err)
		}
		value = matrixDB.GetMatrixData(rule.keyHash)
	default:
		return nil, errors.Errorf("storePos(%d) is unknown", rule.pos)
	}

	if nil == value {
		return nil, ErrValueNotFind
	}
	return value, nil
}

func (ts *TrieStore) GetStateWithRootHash(key string, root common.Hash) ([]byte, error) {
	rule, err := ts.getStoreRule(key)
	if err != nil {
		return nil, err
	}
	var value []byte
	switch rule.pos {
	case storePosRoot: // root树上
		rootDB, err := ts.reader.StateAt(root)
		if err != nil || nil == rootDB {
			return nil, errors.Errorf("root(%s) get state DB err(%v)", root.Hex(), err)
		}
		value = rootDB.GetMatrixData(rule.keyHash)
	case storePosMatrix: // matrix树上
		matrixDB, err := ts.reader.MatrixStateAt(root)
		if err != nil || nil == matrixDB {
			return nil, errors.Errorf("root(%s)  get matrix state DB err(%v)", root.Hex(), err)
		}
		value = matrixDB.GetMatrixData(rule.keyHash)
	default:
		return nil, errors.Errorf("storePos(%d) is unknown", rule.pos)
	}

	if nil == value {
		return nil, ErrValueNotFind
	}
	return value, nil
}

func (ts *TrieStore) GetStateWithStateDB(key string, rootDB *state.StateDB, matrixDB *state.StateDB) ([]byte, error) {
	rule, err := ts.getStoreRule(key)
	if err != nil {
		return nil, err
	}
	return ts.readFromDB(rule.pos, rule.keyHash, rootDB, matrixDB)
}

func (ts *TrieStore) getStoreRule(key string) (*storeRule, error) {
	rule, OK := ts.keyMap[key]
	if !OK {
		return nil, errors.Errorf("key(%s) is illegal", key)
	}
	return &rule, nil
}

func (ts *TrieStore) writeToDB(pos storePos, keyHash common.Hash, value []byte, rootDB *state.StateDB, matrixDB *state.StateDB) error {
	switch pos {
	case storePosRoot: // 存储在root树上，强共识
		if nil == rootDB {
			return ErrRootDBNil
		}
		rootDB.SetMatrixData(keyHash, value)
	case storePosMatrix: // 存储在matrix树上，弱共识
		if nil == matrixDB {
			return ErrMatrixDBNil
		}
		matrixDB.SetMatrixData(keyHash, value)
	default:
		return errors.Errorf("storePos(%d) is unknown", pos)
	}
	return nil
}

func (ts *TrieStore) readFromDB(pos storePos, keyHash common.Hash, rootDB *state.StateDB, matrixDB *state.StateDB) (value []byte, err error) {
	switch pos {
	case storePosRoot: // root树上

		if nil == rootDB {
			return nil, ErrRootDBNil
		}
		value = rootDB.GetMatrixData(keyHash)
	case storePosMatrix: // matrix树上
		if nil == matrixDB {
			return nil, ErrMatrixDBNil
		}
		value = matrixDB.GetMatrixData(keyHash)
	default:
		return nil, errors.Errorf("storePos(%d) is unknown", pos)
	}

	if nil == value {
		return nil, ErrValueNotFind
	}
	return value, nil
}
