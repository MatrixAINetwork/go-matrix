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
package models

import (
	"errors"
	"strings"
)

// ConsistencyLevel represent a required replication criteria before a write can
// be returned as successful.
//
// The consistency level is handled in open-source InfluxDB but only applicable to clusters.
type ConsistencyLevel int

const (
	// ConsistencyLevelAny allows for hinted handoff, potentially no write happened yet.
	ConsistencyLevelAny ConsistencyLevel = iota

	// ConsistencyLevelOne requires at least one data node acknowledged a write.
	ConsistencyLevelOne

	// ConsistencyLevelQuorum requires a quorum of data nodes to acknowledge a write.
	ConsistencyLevelQuorum

	// ConsistencyLevelAll requires all data nodes to acknowledge a write.
	ConsistencyLevelAll
)

var (
	// ErrInvalidConsistencyLevel is returned when parsing the string version
	// of a consistency level.
	ErrInvalidConsistencyLevel = errors.New("invalid consistency level")
)

// ParseConsistencyLevel converts a consistency level string to the corresponding ConsistencyLevel const.
func ParseConsistencyLevel(level string) (ConsistencyLevel, error) {
	switch strings.ToLower(level) {
	case "any":
		return ConsistencyLevelAny, nil
	case "one":
		return ConsistencyLevelOne, nil
	case "quorum":
		return ConsistencyLevelQuorum, nil
	case "all":
		return ConsistencyLevelAll, nil
	default:
		return 0, ErrInvalidConsistencyLevel
	}
}
