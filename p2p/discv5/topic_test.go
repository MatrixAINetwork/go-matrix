// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

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
