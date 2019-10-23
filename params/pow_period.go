// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package params

const PowBlockPeriod = 3

func IsPowBlock(number uint64, broadcastInterval uint64) bool {
	remainder := number % broadcastInterval
	if remainder == 0 { // 广播区块
		return false
	}
	return remainder%PowBlockPeriod == 0
}

func IsAIBlock(number uint64, broadcastInterval uint64) bool {
	remainder := number % broadcastInterval
	return (remainder+2)%PowBlockPeriod == 0
}

func GetCurAIBlockNumber(number uint64, broadcastInterval uint64) uint64 {
	remainder := number % broadcastInterval
	if remainder == 0 {
		return number - PowBlockPeriod
	}
	return number - (remainder+2)%PowBlockPeriod
}

func GetNextAIBlockNumber(number uint64, broadcastInterval uint64) uint64 {
	curAINumber := GetCurAIBlockNumber(number, broadcastInterval)
	nextAINumber := curAINumber + PowBlockPeriod
	if nextAINumber%broadcastInterval == 0 {
		nextAINumber += 1
	}
	return nextAINumber
}
