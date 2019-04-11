// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

// Contains the metrics collected by the downloader.

package downloader

import (
	"github.com/MatrixAINetwork/go-matrix/metrics"
)

var (
	headerInMeter      = metrics.NewRegisteredMeter("man/downloader/headers/in", nil)
	headerReqTimer     = metrics.NewRegisteredTimer("man/downloader/headers/req", nil)
	headerDropMeter    = metrics.NewRegisteredMeter("man/downloader/headers/drop", nil)
	headerTimeoutMeter = metrics.NewRegisteredMeter("man/downloader/headers/timeout", nil)

	bodyInMeter      = metrics.NewRegisteredMeter("man/downloader/bodies/in", nil)
	bodyReqTimer     = metrics.NewRegisteredTimer("man/downloader/bodies/req", nil)
	bodyDropMeter    = metrics.NewRegisteredMeter("man/downloader/bodies/drop", nil)
	bodyTimeoutMeter = metrics.NewRegisteredMeter("man/downloader/bodies/timeout", nil)

	receiptInMeter      = metrics.NewRegisteredMeter("man/downloader/receipts/in", nil)
	receiptReqTimer     = metrics.NewRegisteredTimer("man/downloader/receipts/req", nil)
	receiptDropMeter    = metrics.NewRegisteredMeter("man/downloader/receipts/drop", nil)
	receiptTimeoutMeter = metrics.NewRegisteredMeter("man/downloader/receipts/timeout", nil)

	stateInMeter   = metrics.NewRegisteredMeter("man/downloader/states/in", nil)
	stateDropMeter = metrics.NewRegisteredMeter("man/downloader/states/drop", nil)
)
