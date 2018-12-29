package matrixstate

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/log"
	"github.com/pkg/errors"
	"sync"
)

type keyInfo struct {
	keyHash      common.Hash
	dataProducer ProduceMatrixStateDataFn
}

func genKeyMap() (keyMap map[string]*keyInfo) {
	keyMap = make(map[string]*keyInfo)
	for key, hash := range km.keys {
		keyMap[key] = &keyInfo{hash, nil}
	}
	return
}

type MatrixState struct {
	mu     sync.RWMutex
	keyMap map[string]*keyInfo
}

func NewMatrixState() *MatrixState {
	return &MatrixState{
		keyMap: genKeyMap(),
	}
}

func (ms *MatrixState) RegisterProducer(key string, producer ProduceMatrixStateDataFn) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	info, err := ms.findKeyInfo(key)
	if err != nil {
		return err
	}
	info.dataProducer = producer
	return nil
}

func (ms *MatrixState) ProcessMatrixState(block *types.Block, state StateDB) error {
	if block == nil || state == nil {
		return errors.New("param is nil")
	}
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	readFn := func(key string) (interface{}, error) {
		return GetDataByState(key, state)
	}

	dataMap := make(map[common.Hash][]byte)
	for key, info := range ms.keyMap {
		if info == nil || info.dataProducer == nil {
			continue
		}

		codec, exist := km.codecMap[key]
		if exist == false {
			log.Error("matrix state", "编解码器未找到", key)
			continue
		}

		data, err := info.dataProducer(block, readFn)
		if err != nil {
			return errors.Errorf("key(%s) produce matrix state data err(%v)", key, err)
		}
		if nil == data {
			continue
		}
		bytes, err := codec.encodeFn(data)
		if err != nil {
			return errors.Errorf("encode data of key(%s) err: %v", key, err)
		}
		if len(bytes) == 0 {
			return errors.Errorf("the encoded data of key(%s) is empty", key)
		}

		dataMap[info.keyHash] = bytes
	}

	for keyHash, data := range dataMap {
		state.SetMatrixData(keyHash, data)
	}

	return nil
}

func (ms *MatrixState) findKeyInfo(key string) (*keyInfo, error) {
	info, OK := ms.keyMap[key]
	if !OK {
		return nil, errors.Errorf("key(%s) is illegal", key)
	}
	if info == nil {
		return nil, errors.Errorf("CRITICAL the info of key(%s) is nil in map", key)
	}
	return info, nil
}
