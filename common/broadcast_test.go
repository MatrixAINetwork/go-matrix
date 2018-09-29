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
package common

import "testing"

type Testcal struct {
	have_broadcast               uint64
	have_reelection              uint64
	want_IsBroadcastNumber       bool
	want_IsReElectionNumber      bool
	want_GetLastBroadcastNumber  uint64
	want_GetLastReElectionNumber uint64
	want_GetNextBroadcastNumber  uint64
	want_GetNextReElectionNumber uint64
	want_GetBroadcastInterval    uint64
	want_GetReElectionInterval   uint64
}

/*
	test1 := Testcal{
			have_broadcast:               100,
			have_reelection:              300,
			want_IsBroadcastNumber:       true,
			want_IsReElectionNumber:      true,
			want_GetLastBroadcastNumber:  100,
			want_GetLastReElectionNumber: 300,
			want_GetNextBroadcastNumber:  100,
			want_GetNextReElectionNumber: 300,
			want_GetBroadcastInterval:    100,
			want_GetReElectionInterval:   300,
		}
*/
func NewTestca1(bro uint64, reel uint64, want_IsBro bool, want_IsReE bool, want_GetLastBro uint64, want_GetLastReE uint64, want_GetNextBro uint64, want_GetNextReE uint64, want_GetBro uint64, want_GetEle uint64) Testcal {
	test1 := Testcal{
		have_broadcast:               bro,
		have_reelection:              reel,
		want_IsBroadcastNumber:       want_IsBro,
		want_IsReElectionNumber:      want_IsReE,
		want_GetLastBroadcastNumber:  want_GetLastBro,
		want_GetLastReElectionNumber: want_GetLastReE,
		want_GetNextBroadcastNumber:  want_GetNextBro,
		want_GetNextReElectionNumber: want_GetNextReE,
		want_GetBroadcastInterval:    want_GetBro,
		want_GetReElectionInterval:   want_GetEle,
	}
	return test1
}
func Compare(t *testing.T, aim Testcal) {
	if aim.want_GetBroadcastInterval != GetBroadcastInterval() {
		t.Errorf("want:%d  have:%d", aim.want_GetBroadcastInterval, GetBroadcastInterval())
	}
	if aim.want_GetReElectionInterval != GetReElectionInterval() {
		t.Errorf("want:%d have:%d", aim.want_GetReElectionInterval, GetReElectionInterval())
	}
	if aim.want_GetLastBroadcastNumber != GetLastBroadcastNumber(aim.have_broadcast) {
		t.Errorf("want:%d have %d", aim.want_GetLastBroadcastNumber, GetLastBroadcastNumber(aim.have_broadcast))
	}
	if aim.want_GetLastReElectionNumber != GetLastReElectionNumber(aim.have_reelection) {
		t.Errorf("want %d have %d", aim.want_GetLastReElectionNumber, GetLastReElectionNumber(aim.have_reelection))
	}
	if aim.want_GetNextBroadcastNumber != GetNextBroadcastNumber(aim.have_broadcast) {
		t.Errorf("want %d have %d", aim.want_GetNextBroadcastNumber, GetNextBroadcastNumber(aim.have_broadcast))
	}
	if aim.want_GetNextReElectionNumber != GetNextReElectionNumber(aim.have_reelection) {
		t.Errorf("want %d have %d", aim.want_GetNextReElectionNumber, GetNextReElectionNumber(aim.have_reelection))
	}
	if aim.want_IsBroadcastNumber != IsBroadcastNumber(aim.have_broadcast) {
		t.Errorf("want %d have %d", aim.want_IsBroadcastNumber, IsBroadcastNumber(aim.have_broadcast))
	}
	if aim.want_IsReElectionNumber != IsReElectionNumber(aim.have_reelection) {
		t.Errorf("want %d have %d", aim.want_IsReElectionNumber, IsReElectionNumber(aim.have_reelection))
	}

}
func Test1(t *testing.T) {
	{

		test := NewTestca1(100, 300, true, true, 100, 300, 100, 300, 100, 300)

		Compare(t, test)
	}

}
func Test2(t *testing.T) {
	{
		test := NewTestca1(99, 299, false, false, 0, 0, 100, 300, 100, 300)
		Compare(t, test)
	}
}

func Test3(t *testing.T) {
	{
		test := NewTestca1(101, 301, false, false, 100, 300, 200, 600, 100, 300)
		Compare(t, test)
	}
}
func TestAll(t *testing.T) {
	t.Run("test1", Test1)
	t.Run("test2", Test2)
	t.Run("test3", Test3)
}
