// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package blkgenor

const (
	testInstance = "block-verify"
	testAddress  = "0x6a3217d128a76e4777403e092bde8362d4117773"
)

//func TestProcess_handleUpTime(t *testing.T) {
//	log.InitLog(3)
//	type fields struct {
//		mu                sync.Mutex
//		curLeader         common.Address
//		nextLeader        common.Address
//		preBlockHash      common.Hash
//		number            uint64
//		role              common.RoleType
//		state             State
//		pm                *ProcessManage
//		powPool           *PowPool
//		broadcastRstCache []*mc.BlockData
//		consensusBlock    *mc.BlockVerifyConsensusOK
//		genBlockData      *mc.BlockVerifyConsensusOK
//		insertBlockData   *mc.BlockInsertNotify
//	}
//	type args struct {
//		accounts               []common.Address
//		calltherollRspAccounts map[common.Address]uint32
//		heatBeatAccounts       map[common.Address][]byte
//	}
//	ValidatorAccount := make([]common.Address, 0)
//	calltherollRspAccounts := make(map[common.Address]uint32)
//	heatbeat := make(map[common.Address][]byte)
//	for i, v := range ca.Validatoraccountlist {
//		ValidatorAccount = append(ValidatorAccount, common.HexToAddress(v))
//		if i < 5 {
//			calltherollRspAccounts[common.HexToAddress(v)] = uint32(i)
//		}
//		if i < 3 {
//			heatbeat[common.HexToAddress(v)] = nil
//		}
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		wantErr bool
//	}{
//		{"test1", *new(fields), args{ValidatorAccount, calltherollRspAccounts, heatbeat}, false},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			p := &Process{}
//			if err := p.handleUpTime(tt.args.accounts, tt.args.calltherollRspAccounts, tt.args.heatBeatAccounts, 0); (err != nil) != tt.wantErr {
//				t.Errorf("Process.handleUpTime() error = %v, wantErr %v", err, tt.wantErr)
//			}
//		})
//	}
//}
