// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package les

import (
	"testing"
)

func TestExecQueue(t *testing.T) {
	var (
		N        = 10000
		q        = newExecQueue(N)
		counter  int
		execd    = make(chan int)
		testexit = make(chan struct{})
	)
	defer q.quit()
	defer close(testexit)

	check := func(state string, wantOK bool) {
		c := counter
		counter++
		qf := func() {
			select {
			case execd <- c:
			case <-testexit:
			}
		}
		if q.canQueue() != wantOK {
			t.Fatalf("canQueue() == %t for %s", !wantOK, state)
		}
		if q.queue(qf) != wantOK {
			t.Fatalf("canQueue() == %t for %s", !wantOK, state)
		}
	}

	for i := 0; i < N; i++ {
		check("queue below cap", true)
	}
	check("full queue", false)
	for i := 0; i < N; i++ {
		if c := <-execd; c != i {
			t.Fatal("execution out of order")
		}
	}
	q.quit()
	check("closed queue", false)
}
