// Copyright 2018 The MATRIX Authors as well as Copyright 2014-2017 The go-ethereum Authors
// This file is consisted of the MATRIX library and part of the go-ethereum library.
//
// The MATRIX-ethereum library is free software: you can redistribute it and/or modify it under the terms of the MIT License.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, 
//and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject tothe following conditions:
//
//The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
//
//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, 
//WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISINGFROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE
//OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// Contains the metrics collected by the downloader.

package downloader

import (
	"github.com/matrix/go-matrix/metrics"
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
