// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package reward

import (
	"math/big"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/reward/cfg"
)

type Reward interface {
	CalcNodesRewards(blockReward *big.Int, Leader common.Address, num uint64, parentHash common.Hash, coinType string) map[common.Address]*big.Int
	CalcValidatorRewards(Leader common.Address, num uint64) map[common.Address]*big.Int
	CalcMinerRewards(num uint64, parentHash common.Hash) map[common.Address]*big.Int
	CalcMinerRateMount(blockReward *big.Int) (*big.Int, *big.Int, *big.Int)
	CalcValidatorRateMount(blockReward *big.Int) (*big.Int, *big.Int, *big.Int)
	GetRewardCfg() *cfg.RewardCfg
}
