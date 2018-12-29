// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package manparams

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
