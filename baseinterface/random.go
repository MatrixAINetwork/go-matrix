// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package baseinterface

import (
	"math/big"

	"github.com/MatrixAINetwork/go-matrix/core/matrixstate"
	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/params"

	"fmt"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/event"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/params/manparams"
)

const (
	ModuleRandom = "随机数接口服务"
)

var (
	mapReg = make(map[string]func(string, RandomChainSupport) (RandomSubService, error), 0)
)

func RegRandom(name string, fun func(string, RandomChainSupport) (RandomSubService, error)) {
	mapReg[name] = fun
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
	GetAncestorHash(sonHash common.Hash, ancestorNumber uint64) (common.Hash, error)
	// GetBlock retrieves a block sfrom the database by hash and number.
	GetBlock(hash common.Hash, number uint64) *types.Block
	StateAt(root []common.CoinRoot) (*state.StateDBManage, error)
	State() (*state.StateDBManage, error)
	StateAtNumber(number uint64) (*state.StateDBManage, error)
	StateAtBlockHash(hash common.Hash) (*state.StateDBManage, error)
	GetSuperBlockNum() (uint64, error)
	GetGraphByState(state matrixstate.StateDB) (*mc.TopologyGraph, *mc.ElectGraph, error)
}

type Random struct {
	roleUpdateCh  chan *mc.RoleUpdatedMsg
	roleUpdateSub event.Subscription
	quitChan      chan struct{}
	mapSubService map[string]RandomSubService
}

type RandomChainSupport interface {
	BlockChain() ChainReader
}
type RandomChain struct {
	bc ChainReader
}

func (self *RandomChain) BlockChain() ChainReader {
	return self.bc
}

type RandomSubService interface {
	Prepare(uint64, common.Hash) error
	CalcData(data common.Hash) (*big.Int, error)
}

func checkDataValidity(support interface{}) bool {
	return common.IsNil(support)
}
func NewRandom(bc ChainReader) (*Random, error) {
	//if checkDataValidity(support)==false{
	//	log.Error(ModuleRandom,"创建随机数服务阶段,输入不合法","输入为空接口")
	//	return nil,errors.New("创建随机数服务阶段,输入不合法")
	//}
	random := &Random{
		roleUpdateCh:  make(chan *mc.RoleUpdatedMsg, 1),
		quitChan:      make(chan struct{}, 1),
		mapSubService: make(map[string]RandomSubService, 0),
	}
	support := &RandomChain{bc: bc}
	for _, name := range manparams.RandomServiceName {
		Plug, needNewSubService := getSubServicePlug(name)
		if needNewSubService == false {
			log.Warn(ModuleRandom, "新建子服务阶段,子服务不需要被创建 名称", name)
			continue
		}
		if err := random.newSubServer(name, Plug, support); err != nil {
			log.Error(ModuleRandom, "新建子服务阶段,子服务创建失败 名称", name)
			return nil, err
		}
		log.Info(ModuleRandom, "新建子服务阶段,子服务创建成功 名称", name)
	}

	var err error
	random.roleUpdateSub, err = mc.SubscribeEvent(mc.CA_RoleUpdated, random.roleUpdateCh)
	if err != nil {
		log.Error(ModuleRandom, "订阅CA消息阶段,CA消息订阅失败 err", err)
		return nil, err
	}
	go random.update()
	return random, nil
}

func (self *Random) update() {
	defer self.roleUpdateSub.Unsubscribe()
	for {
		select {
		case RoleUpdateData := <-self.roleUpdateCh:
			go self.processRoleUpdateData(RoleUpdateData)
		case <-self.quitChan:
			return
		}
	}
}

func (self *Random) Stop() {
	close(self.quitChan)
}
func (self *Random) processRoleUpdateData(data *mc.RoleUpdatedMsg) {
	for _, v := range self.mapSubService {
		go v.Prepare(data.BlockNum, data.BlockHash)
	}
}

func (self *Random) newSubServer(name string, plugConfig string, support RandomChainSupport) error {
	var err error
	if _, ok := mapReg[name]; ok == false {
		log.Error(ModuleRandom, "新建子服务阶段,该子服务未注册", name)
		return fmt.Errorf("该子服务未注册 %v", name)
	}
	if self.mapSubService[name], err = mapReg[name](plugConfig, support); err != nil {
		log.Error(ModuleRandom, "新建子服务阶段,该子服务新建失败", name, "err", err)
	}
	log.Info(ModuleRandom, "新建子服务阶段,该子服务创建成功 index", name)
	return nil
}

func (self *Random) GetRandom(hash common.Hash, Type string) (*big.Int, error) {
	return self.mapSubService[Type].CalcData(hash)
}

func getSubServicePlug(name string) (string, bool) {
	plug, ok := manparams.RandomConfig[name]
	if ok == false {
		log.Warn(ModuleRandom, "获取插件阶段,配置中无该子服务,不需要开启", name)
		return "", false
	}

	plugs, ok := manparams.RandomServicePlugs[name]
	if ok == false {
		log.Error(ModuleRandom, "获取插件阶段 无该子服务 服务名称", name)
		return "", false
	}
	for _, v := range plugs {
		if v == plug {
			log.Info(ModuleRandom, "获取插件阶段", "", "插件列表中有该插件", plug)
			return v, true
		}
	}
	log.Warn(ModuleRandom, "获取插件阶段,配置中的插件不合法，使用默认插件 名称", manparams.RandomServiceDefaultPlugs[name])
	return manparams.RandomServiceDefaultPlugs[name], true
}
