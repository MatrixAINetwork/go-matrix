// Copyright (c) 2018 The MATRIX Authors

// Distributed under the MIT software license, see the accompanying

// file COPYING or http://www.opensource.org/licenses/mit-license.php

package leaderelect

import (
	"github.com/MatrixAINetwork/go-matrix/params/manparams"
	"testing"
)

func Test_turnTimes_SetBeginTime(t *testing.T) {
	type args struct {
		consensusTurn uint32
		time          int64
	}
	tests := []struct {
		name string
		args []args
		want []bool
	}{
		{
			name: "各种情况下的时间设置",
			args: []args{
				{
					consensusTurn: 0,
					time:          100,
				},
				{
					consensusTurn: 0,
					time:          98,
				},
				{
					consensusTurn: 1,
					time:          98,
				},
				{
					consensusTurn: 0,
					time:          100,
				},
				{
					consensusTurn: 0,
					time:          101,
				},
			},
			want: []bool{true, false, true, false, true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tts := newTurnTimes()
			for i, ttarg := range tt.args {
				if got := tts.SetBeginTime(ttarg.consensusTurn, ttarg.time); got != tt.want[i] {
					t.Errorf("index(%d), turnTimes.SetBeginTime() = %v, want %v", i, got, tt.want[i])
				}
			}
		})
	}
}

func Test_turnTimes_GetBeginTime(t *testing.T) {
	tt := newTurnTimes()
	tt.SetBeginTime(0, 100)
	tt.SetBeginTime(5, 500)

	if got := tt.GetBeginTime(0); got != 100 {
		t.Errorf("turnTimes.GetBeginTime() = %v, want %v", got, 100)
	}

	if got := tt.GetBeginTime(3); got != 0 {
		t.Errorf("turnTimes.GetBeginTime() = %v, want %v", got, 0)
	}

	if got := tt.GetBeginTime(5); got != 500 {
		t.Errorf("turnTimes.GetBeginTime() = %v, want %v", got, 500)
	}
}

func Test_turnTimes_GetPosEndTime(t *testing.T) {
	type fields struct {
		beginTimes map[uint32]int64
	}
	type args struct {
		consensusTurn uint32
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int64
	}{
		{
			name:   "无开始时间时，0轮次的POS结束时间",
			fields: fields{make(map[uint32]int64)},
			args:   args{0},
			want:   manparams.LRSParentMiningTime + manparams.LRSPOSOutTime,
		},
		{
			name:   "有开始时间时，0轮次的POS结束时间",
			fields: fields{map[uint32]int64{0: 300, 3: 400}},
			args:   args{0},
			want:   300 + manparams.LRSParentMiningTime + manparams.LRSPOSOutTime,
		},
		{
			name:   "无开始时间时，2轮次的POS结束时间",
			fields: fields{map[uint32]int64{0: 300, 3: 400}},
			args:   args{2},
			want:   manparams.LRSPOSOutTime,
		},
		{
			name:   "有开始时间时，3轮次的POS结束时间",
			fields: fields{map[uint32]int64{0: 300, 3: 400}},
			args:   args{3},
			want:   400 + manparams.LRSPOSOutTime,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tts := &turnTimes{
				beginTimes: tt.fields.beginTimes,
			}
			if got := tts.GetPosEndTime(tt.args.consensusTurn); got != tt.want {
				t.Errorf("turnTimes.GetPosEndTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_turnTimes_CalState(t *testing.T) {
	type fields struct {
		beginTimes map[uint32]int64
	}
	type args struct {
		consensusTurn uint32
		time          int64
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		wantSt          state
		wantRemainTime  int64
		wantReelectTurn uint32
	}{
		{
			name:            "无开始时间时，0轮次，39秒时的状态计算",
			fields:          fields{make(map[uint32]int64)},
			args:            args{0, 39},
			wantSt:          stPos,
			wantRemainTime:  1,
			wantReelectTurn: 0,
		},
		{
			name:            "无开始时间时，0轮次，40秒时的状态计算",
			fields:          fields{make(map[uint32]int64)},
			args:            args{0, 40},
			wantSt:          stReelect,
			wantRemainTime:  40,
			wantReelectTurn: 1,
		},
		{
			name:            "无开始时间时，0轮次，77秒时的状态计算",
			fields:          fields{make(map[uint32]int64)},
			args:            args{0, 77},
			wantSt:          stReelect,
			wantRemainTime:  3,
			wantReelectTurn: 1,
		},
		{
			name:            "开始时间=100秒时，0轮次，88秒时的状态计算",
			fields:          fields{map[uint32]int64{0: 100}},
			args:            args{0, 88},
			wantSt:          stPos,
			wantRemainTime:  52,
			wantReelectTurn: 0,
		},
		{
			name:            "开始时间=100秒时，0轮次，108秒时的状态计算",
			fields:          fields{map[uint32]int64{0: 100}},
			args:            args{0, 105},
			wantSt:          stPos,
			wantRemainTime:  35,
			wantReelectTurn: 0,
		},
		{
			name:            "开始时间=300秒时，4轮次，320秒时的状态计算",
			fields:          fields{map[uint32]int64{4: 300}},
			args:            args{4, 320},
			wantSt:          stReelect,
			wantRemainTime:  40,
			wantReelectTurn: 1,
		},
		{
			name:            "开始时间=300秒时，4轮次，406秒时的状态计算",
			fields:          fields{map[uint32]int64{4: 300}},
			args:            args{4, 406},
			wantSt:          stReelect,
			wantRemainTime:  34,
			wantReelectTurn: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tts := &turnTimes{
				beginTimes: tt.fields.beginTimes,
			}
			gotSt, gotRemainTime, gotReelectTurn := tts.CalState(tt.args.consensusTurn, tt.args.time)
			if gotSt != tt.wantSt {
				t.Errorf("turnTimes.CalState() gotSt = %v, want %v", gotSt, tt.wantSt)
			}
			if gotRemainTime != tt.wantRemainTime {
				t.Errorf("turnTimes.CalState() gotRemainTime = %v, want %v", gotRemainTime, tt.wantRemainTime)
			}
			if gotReelectTurn != tt.wantReelectTurn {
				t.Errorf("turnTimes.CalState() gotReelectTurn = %v, want %v", gotReelectTurn, tt.wantReelectTurn)
			}
		})
	}
}

func Test_turnTimes_CalRemainTime(t *testing.T) {
	type fields struct {
		beginTimes map[uint32]int64
	}
	type args struct {
		consensusTurn uint32
		reelectTurn   uint32
		time          int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int64
	}{
		{
			name:   "无开始时间时，0共识轮次，0重选轮次，39秒时的剩余时间计算",
			fields: fields{make(map[uint32]int64)},
			args:   args{0, 0, 39},
			want:   1,
		},
		{
			name:   "无开始时间时，0共识轮次，0重选轮次，40秒时的剩余时间计算",
			fields: fields{make(map[uint32]int64)},
			args:   args{0, 0, 40},
			want:   0,
		},
		{
			name:   "无开始时间时，0共识轮次，0重选轮次，41秒时的剩余时间计算",
			fields: fields{make(map[uint32]int64)},
			args:   args{0, 0, 41},
			want:   -1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tts := &turnTimes{
				beginTimes: tt.fields.beginTimes,
			}
			if got := tts.CalRemainTime(tt.args.consensusTurn, tt.args.reelectTurn, tt.args.time); got != tt.want {
				t.Errorf("turnTimes.CalRemainTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_turnTimes_CalTurnTime(t *testing.T) {
	type fields struct {
		beginTimes map[uint32]int64
	}
	type args struct {
		consensusTurn uint32
		reelectTurn   uint32
	}
	tests := []struct {
		name          string
		fields        fields
		args          args
		wantBeginTime int64
		wantEndTime   int64
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tts := &turnTimes{
				beginTimes: tt.fields.beginTimes,
			}
			gotBeginTime, gotEndTime := tts.CalTurnTime(tt.args.consensusTurn, tt.args.reelectTurn)
			if gotBeginTime != tt.wantBeginTime {
				t.Errorf("turnTimes.CalTurnTime() gotBeginTime = %v, want %v", gotBeginTime, tt.wantBeginTime)
			}
			if gotEndTime != tt.wantEndTime {
				t.Errorf("turnTimes.CalTurnTime() gotEndTime = %v, want %v", gotEndTime, tt.wantEndTime)
			}
		})
	}
}

func Test_turnTimes_CheckTimeLegal(t *testing.T) {
	type fields struct {
		beginTimes map[uint32]int64
	}
	type args struct {
		consensusTurn uint32
		reelectTurn   uint32
		checkTime     int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tts := &turnTimes{
				beginTimes: tt.fields.beginTimes,
			}
			if err := tts.CheckTimeLegal(tt.args.consensusTurn, tt.args.reelectTurn, tt.args.checkTime); (err != nil) != tt.wantErr {
				t.Errorf("turnTimes.CheckTimeLegal() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
