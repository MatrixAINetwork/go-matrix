// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package metrics

import (
	"sort"
	"testing"
)

func TestMetricsSorting(t *testing.T) {
	var namedMetrics = namedMetricSlice{
		{name: "zzz"},
		{name: "bbb"},
		{name: "fff"},
		{name: "ggg"},
	}

	sort.Sort(namedMetrics)
	for i, name := range []string{"bbb", "fff", "ggg", "zzz"} {
		if namedMetrics[i].name != name {
			t.Fail()
		}
	}
}
