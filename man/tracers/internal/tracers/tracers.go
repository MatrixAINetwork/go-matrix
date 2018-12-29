// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

//go:generate go-bindata -nometadata -o assets.go -pkg tracers -ignore ((tracers)|(assets)).go ./...
//go:generate gofmt -s -w assets.go

// Package tracers contains the actual JavaScript tracer assets.
package tracers
