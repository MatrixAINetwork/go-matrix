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

package netutil

import (
	"net"
	"testing"
	"time"
)

// This test checks that isPacketTooBig correctly identifies
// errors that result from receiving a UDP packet larger
// than the supplied receive buffer.
func TestIsPacketTooBig(t *testing.T) {
	listener, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	sender, err := net.Dial("udp", listener.LocalAddr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer sender.Close()

	sendN := 1800
	recvN := 300
	for i := 0; i < 20; i++ {
		go func() {
			buf := make([]byte, sendN)
			for i := range buf {
				buf[i] = byte(i)
			}
			sender.Write(buf)
		}()

		buf := make([]byte, recvN)
		listener.SetDeadline(time.Now().Add(1 * time.Second))
		n, _, err := listener.ReadFrom(buf)
		if err != nil {
			if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
				continue
			}
			if !isPacketTooBig(err) {
				t.Fatalf("unexpected read error: %v", err)
			}
			continue
		}
		if n != recvN {
			t.Fatalf("short read: %d, want %d", n, recvN)
		}
		for i := range buf {
			if buf[i] != byte(i) {
				t.Fatalf("error in pattern")
				break
			}
		}
	}
}
