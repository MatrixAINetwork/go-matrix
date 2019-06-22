// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package depositcfg

type DepositCfgInterface interface {
	GetDepositPositionCfg(depositType uint64) DepositCfger
}

const (
	VersionA = "A"
)

var depositCfgMap map[string]DepositCfgInterface

func init() {
	depositCfgMap = make(map[string]DepositCfgInterface)
	depositCfgMap[VersionA] = newDepositCfgA()
}

func GetDepositCfg(version string) DepositCfgInterface {
	return depositCfgMap[version]
}
