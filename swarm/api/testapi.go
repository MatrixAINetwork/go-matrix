// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package api

import (
	"github.com/matrix/go-matrix/swarm/network"
)

type Control struct {
	api  *Api
	hive *network.Hive
}

func NewControl(api *Api, hive *network.Hive) *Control {
	return &Control{api, hive}
}

func (self *Control) BlockNetworkRead(on bool) {
	self.hive.BlockNetworkRead(on)
}

func (self *Control) SyncEnabled(on bool) {
	self.hive.SyncEnabled(on)
}

func (self *Control) SwapEnabled(on bool) {
	self.hive.SwapEnabled(on)
}

func (self *Control) Hive() string {
	return self.hive.String()
}
