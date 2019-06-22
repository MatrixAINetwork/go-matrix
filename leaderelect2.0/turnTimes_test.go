// Copyright (c) 2018 The MATRIX Authors

// Distributed under the MIT software license, see the accompanying

// file COPYING or http://www.opensource.org/licenses/mit-license.php

package leaderelect2

import (
	"github.com/MatrixAINetwork/go-matrix/mc"
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
				if got := tts.SetBeginTime(ttarg.time); got != tt.want[i] {
					t.Errorf("index(%d), turnTimes.SetBeginTime() = %v, want %v", i, got, tt.want[i])
				}
			}
		})
	}
}

func Test_turnTimes_GetBeginTime(t *testing.T) {
	tt := newTurnTimes()
	tt.SetBeginTime(100)

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
		beginTimes       int64
		parentMiningTime int64 // 预留父区块挖矿时间
		turnOutTime      int64 // 轮次超时时间
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
			fields: fields{0, 20, 40},
			args:   args{0},
			want:   60,
		},
		{
			name:   "有开始时间时，0轮次的POS结束时间",
			fields: fields{0, 20, 40},
			args:   args{0},
			want:   360,
		},
		{
			name:   "无开始时间时，2轮次的POS结束时间",
			fields: fields{0, 20, 40},
			args:   args{2},
			want:   140,
		},
		{
			name:   "有开始时间时，3轮次的POS结束时间",
			fields: fields{400, 20, 40},
			args:   args{3},
			want:   400 + 180,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tts := &turnTimes{
				beginTime: tt.fields.beginTimes,
			}
			if got := tts.GetPosEndTime(tt.args.consensusTurn); got != tt.want {
				t.Errorf("turnTimes.GetPosEndTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_turnTimes_CalState(t *testing.T) {
	tts := newTurnTimes()
	tts.SetBeginTime(5)

	tts.SetTimeConfig(&mc.LeaderConfig{
		ParentMiningTime:      5,
		PosOutTime:            5,
		ReelectOutTime:        30,
		ReelectHandleInterval: 5,
	})

}
