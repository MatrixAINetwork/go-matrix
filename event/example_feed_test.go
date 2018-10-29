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

package event_test

import (
	"fmt"

	"github.com/matrix/go-matrix/event"
)

func ExampleFeed_acknowledgedEvents() {
	// This example shows how the return value of Send can be used for request/reply
	// interaction between event consumers and producers.
	var feed event.Feed
	type ackedEvent struct {
		i   int
		ack chan<- struct{}
	}

	// Consumers wait for events on the feed and acknowledge processing.
	done := make(chan struct{})
	defer close(done)
	for i := 0; i < 3; i++ {
		ch := make(chan ackedEvent, 100)
		sub := feed.Subscribe(ch)
		go func() {
			defer sub.Unsubscribe()
			for {
				select {
				case ev := <-ch:
					fmt.Println(ev.i) // "process" the event
					ev.ack <- struct{}{}
				case <-done:
					return
				}
			}
		}()
	}

	// The producer sends values of type ackedEvent with increasing values of i.
	// It waits for all consumers to acknowledge before sending the next event.
	for i := 0; i < 3; i++ {
		acksignal := make(chan struct{})
		n := feed.Send(ackedEvent{i, acksignal})
		for ack := 0; ack < n; ack++ {
			<-acksignal
		}
	}
	// Output:
	// 0
	// 0
	// 0
	// 1
	// 1
	// 1
	// 2
	// 2
	// 2
}
