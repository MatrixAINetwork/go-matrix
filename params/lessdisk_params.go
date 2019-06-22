// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package params

type LessDiskConfig struct {
	OptInterval     int64  // 操作间隔，单位秒
	HeightThreshold uint64 // 高度阈值
	TimeThreshold   int64  // 事件阈值，单位秒
}

var DefLessDiskConfig = &LessDiskConfig{
	OptInterval:     120,
	HeightThreshold: 30000,
	TimeThreshold:   2 * 60 * 60,
}
