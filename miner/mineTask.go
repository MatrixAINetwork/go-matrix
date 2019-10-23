// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package miner

import (
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"math/big"
)

type mineTaskType int

const (
	mineTaskTypePow mineTaskType = 0
	mineTaskTypeAI               = 1
)

func (self mineTaskType) String() string {
	switch self {
	case mineTaskTypePow:
		return "pow"
	case mineTaskTypeAI:
		return "ai"
	default:
		return "未知类型"
	}
}

type mineTask interface {
	TaskType() mineTaskType
	MineHeaderTime() *big.Int
	MineNumber() *big.Int
	MineHash() common.Hash
	MineDifficulty() *big.Int
	VrfValue() []byte
}

type powMineTask struct {
	mineHash            common.Hash
	mineHeader          *types.Header
	bcInterval          *mc.BCIntervalInfo
	minedBasePow        bool
	minedPow            bool
	powMiningNumber     uint64
	powMiningDifficulty *big.Int
	powMiner            common.Address
	mixDigest           common.Hash
	nonce               types.BlockNonce
	sm3Nonce            types.BlockNonce
}

func newPowMineTask(mineHash common.Hash, mineHeader *types.Header, powMiningNumber uint64, bcInterval *mc.BCIntervalInfo, difficulty *big.Int) *powMineTask {
	return &powMineTask{
		mineHash:            mineHash,
		mineHeader:          mineHeader,
		bcInterval:          bcInterval,
		minedBasePow:        false,
		minedPow:            false,
		powMiningNumber:     powMiningNumber,
		powMiningDifficulty: difficulty,
		powMiner:            common.Address{},
		mixDigest:           common.Hash{},
		nonce:               types.BlockNonce{},
		sm3Nonce:            types.BlockNonce{},
	}
}

func (task *powMineTask) TaskType() mineTaskType   { return mineTaskTypePow }
func (task *powMineTask) MineHeaderTime() *big.Int { return task.mineHeader.Time }
func (task *powMineTask) MineNumber() *big.Int     { return big.NewInt(int64(task.powMiningNumber)) }
func (task *powMineTask) MineHash() common.Hash    { return task.mineHash }
func (task *powMineTask) MineDifficulty() *big.Int { return task.powMiningDifficulty }
func (task *powMineTask) VrfValue() []byte         { return task.mineHeader.VrfValue }

type aiMineTask struct {
	mineHash       common.Hash
	mineHeader     *types.Header
	bcInterval     *mc.BCIntervalInfo
	minedAI        bool
	aiMiningNumber uint64
	aiMiner        common.Address
	aiHash         common.Hash
}

func newAIMineTask(mineHash common.Hash, mineHeader *types.Header, aiMiningNumber uint64, bcInterval *mc.BCIntervalInfo) *aiMineTask {
	return &aiMineTask{
		mineHash:       mineHash,
		mineHeader:     mineHeader,
		bcInterval:     bcInterval,
		minedAI:        false,
		aiMiningNumber: aiMiningNumber,
		aiMiner:        common.Address{},
		aiHash:         common.Hash{},
	}
}

func (task *aiMineTask) TaskType() mineTaskType   { return mineTaskTypeAI }
func (task *aiMineTask) MineHeaderTime() *big.Int { return task.mineHeader.Time }
func (task *aiMineTask) MineNumber() *big.Int     { return big.NewInt(int64(task.aiMiningNumber)) }
func (task *aiMineTask) MineHash() common.Hash    { return task.mineHash }
func (task *aiMineTask) MineDifficulty() *big.Int { return big.NewInt(1) }
func (task *aiMineTask) VrfValue() []byte         { return task.mineHeader.VrfValue }
