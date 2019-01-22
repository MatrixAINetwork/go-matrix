// Copyright (c) 2018-2019 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package core

import (
	"github.com/matrix/go-matrix/core/matrixstate"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/pkg/errors"
	"sync"
)

type PreStateReadFn func(key string) (interface{}, error)
type ProduceMatrixStateDataFn func(block *types.Block, readFn PreStateReadFn) (interface{}, error)

type MatrixProcessor struct {
	mu          sync.RWMutex
	producerMap map[string]ProduceMatrixStateDataFn
}

func NewMatrixProcessor() *MatrixProcessor {
	return &MatrixProcessor{
		producerMap: make(map[string]ProduceMatrixStateDataFn),
	}
}

func (mp *MatrixProcessor) RegisterProducer(key string, producer ProduceMatrixStateDataFn) {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	if _, exist := mp.producerMap[key]; exist {
		log.Warn("MatrixProcessor", "已存在的key重复注册Producer", key)
	}
	mp.producerMap[key] = producer
}

func (mp *MatrixProcessor) ProcessMatrixState(block *types.Block, state *state.StateDB) error {
	if block == nil || state == nil {
		return errors.New("param is nil")
	}
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	version := matrixstate.ReaderVersionInfo(state)
	mgr := matrixstate.GetManager(version)
	if mgr == nil {
		return matrixstate.ErrFindManager
	}

	readFn := func(key string) (interface{}, error) {
		if key == mc.MSKeyVersionInfo {
			return version, nil
		}
		opt, err := mgr.FindOperator(key)
		if err != nil {
			return nil, err
		}
		return opt.GetValue(state)
	}

	dataMap := make(map[string]interface{})
	for key := range mp.producerMap {
		data, err := mp.producerMap[key](block, readFn)
		if err != nil {
			return errors.Errorf("key(%s) produce matrix state data err(%v)", key, err)
		}
		if nil == data {
			continue
		}

		dataMap[key] = data
	}

	for key := range dataMap {
		opt, err := mgr.FindOperator(key)
		if err != nil {
			return errors.Errorf("key(%s) find operator err: %v", err)
		}
		if err := opt.SetValue(state, dataMap[key]); err != nil {
			return errors.Errorf("key(%s) set value err: %v", err)
		}
	}

	return nil
}
