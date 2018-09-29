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
package blockgenor

import (
	"sync"
	"testing"

	"github.com/matrix/go-matrix/ca"

	"github.com/matrix/go-matrix/log"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/mc"
)

const (
	testInstance = "block-verify"
	testAddress  = "0x8605cdbbdb6d264aa742e77020dcbc58fcdce182"
)

func TestProcess_handleUpTime(t *testing.T) {
	log.InitLog(3)
	type fields struct {
		mu                sync.Mutex
		curLeader         common.Address
		nextLeader        common.Address
		preBlockHash      common.Hash
		number            uint64
		role              common.RoleType
		state             State
		pm                *ProcessManage
		powPool           *PowPool
		broadcastRstCache []*mc.BlockData
		consensusBlock    *mc.BlockVerifyConsensusOK
		genBlockData      *mc.BlockVerifyConsensusOK
		insertBlockData   *mc.BlockInsertNotify
	}
	type args struct {
		accounts               []common.Address
		calltherollRspAccounts map[common.Address]uint32
		heatBeatAccounts       map[common.Address][]byte
	}
	ValidatorAccount := make([]common.Address, 0)
	calltherollRspAccounts := make(map[common.Address]uint32)
	heatbeat := make(map[common.Address][]byte)
	for i, v := range ca.Validatoraccountlist {
		ValidatorAccount = append(ValidatorAccount, common.HexToAddress(v))
		if i < 5 {
			calltherollRspAccounts[common.HexToAddress(v)] = uint32(i)
		}
		if i < 3 {
			heatbeat[common.HexToAddress(v)] = nil
		}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"test1", *new(fields), args{ValidatorAccount, calltherollRspAccounts, heatbeat}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Process{}
			if err := p.handleUpTime(tt.args.accounts, tt.args.calltherollRspAccounts, tt.args.heatBeatAccounts, 0); (err != nil) != tt.wantErr {
				t.Errorf("Process.handleUpTime() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
