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

package discv5

import (
	"encoding/binary"
	"testing"
	"time"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/common/mclock"
)

func TestTopicRadius(t *testing.T) {
	now := mclock.Now()
	topic := Topic("qwerty")
	rad := newTopicRadius(topic)
	targetRad := (^uint64(0)) / 100

	waitFn := func(addr common.Hash) time.Duration {
		prefix := binary.BigEndian.Uint64(addr[0:8])
		dist := prefix ^ rad.topicHashPrefix
		relDist := float64(dist) / float64(targetRad)
		relTime := (1 - relDist/2) * 2
		if relTime < 0 {
			relTime = 0
		}
		return time.Duration(float64(targetWaitTime) * relTime)
	}

	bcnt := 0
	cnt := 0
	var sum float64
	for cnt < 100 {
		addr := rad.nextTarget(false).target
		wait := waitFn(addr)
		ticket := &ticket{
			topics:  []Topic{topic},
			regTime: []mclock.AbsTime{mclock.AbsTime(wait)},
			node:    &Node{nodeNetGuts: nodeNetGuts{sha: addr}},
		}
		rad.adjustWithTicket(now, addr, ticketRef{ticket, 0})
		if rad.radius != maxRadius {
			cnt++
			sum += float64(rad.radius)
		} else {
			bcnt++
			if bcnt > 500 {
				t.Errorf("Radius did not converge in 500 iterations")
			}
		}
	}
	avgRel := sum / float64(cnt) / float64(targetRad)
	if avgRel > 1.05 || avgRel < 0.95 {
		t.Errorf("Average/target ratio is too far from 1 (%v)", avgRel)
	}
}
