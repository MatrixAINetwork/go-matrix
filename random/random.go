// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package random

import (
	"github.com/matrix/go-matrix/mc"
)

const (
	ModuleSeed = "随机种子生成"
	ModuleVote = "随机数投票"
)

type Random struct {
	electionseed *ElectionSeed
	randomvote   *RandomVote
}

func New(msgcenter *mc.Center) (*Random, error) {
	random := &Random{}
	var err error
	random.electionseed, err = newElectionSeed(msgcenter)
	if err != nil {
		return nil, err
	}
	random.randomvote, err = newRandomVote(msgcenter)
	if err != nil {
		return nil, err
	}

	return random, nil

}
