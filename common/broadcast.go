//1542983923.644285
//1542983057.874398
//1542982278.1239665
//1542981528.3021789
//1542980915.4978232
//1542980237.6891313
//1542979563.9943235
//1542978729.0482628
//1542978010.262432
// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package common

var (
	BroadcastInterval  = uint64(100)
	ReelectionInterval = uint64(300)
)

func IsBroadcastNumber(number uint64) bool {
	if number%BroadcastInterval == 0 {
		return true
	}
	return false
}

func IsReElectionNumber(number uint64) bool {
	if number%ReelectionInterval == 0 {
		return true
	}
	return false
}

func GetLastBroadcastNumber(number uint64) uint64 {
	if IsBroadcastNumber(number) {
		return number
	}
	ans := (number / BroadcastInterval) * BroadcastInterval
	return ans
}

func GetLastReElectionNumber(number uint64) uint64 {
	if IsReElectionNumber(number) {
		return number
	}
	ans := (number / ReelectionInterval) * ReelectionInterval
	return ans
}

func GetNextBroadcastNumber(number uint64) uint64 {
	if IsBroadcastNumber(number) {
		return number
	}
	ans := (number/BroadcastInterval + 1) * BroadcastInterval
	return ans
}

func GetNextReElectionNumber(number uint64) uint64 {
	if IsReElectionNumber(number) {
		return number
	}
	ans := (number/ReelectionInterval + 1) * ReelectionInterval
	return ans
}

func GetBroadcastInterval() uint64 {
	return BroadcastInterval
}
func GetReElectionInterval() uint64 {
	return ReelectionInterval
}
