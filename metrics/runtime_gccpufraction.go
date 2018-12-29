// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
// +build go1.5

package metrics

import "runtime"

func gcCPUFraction(memStats *runtime.MemStats) float64 {
	return memStats.GCCPUFraction
}
