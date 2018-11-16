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
package reelection

import (
	"github.com/matrix/go-matrix/common"

	"github.com/matrix/go-matrix/params/man"
	"github.com/matrix/go-matrix/log"
)

func (self *ReElection)boolTopStatus(height uint64,types common.RoleType)bool{
	if _,_,err:=self.readElectData(types,height);err!=nil{
		return false
	}
	return true
}
func (self *ReElection)checkTopGenStatus(height uint64)error{

	if ok:=self.boolTopStatus(common.GetNextReElectionNumber(height),common.RoleMiner);ok==false{
		log.Warn(Module,"height re-caculation needed for miner topology",height)
		if height==0{
			return nil
		}
		ReElectionHeight:=common.GetNextReElectionNumber(height)
		 if err:=self.ToGenMinerTop(ReElectionHeight - man.MinerTopologyGenerateUpTime);err!=nil{
		 	return err
		 }

	}

	if ok:=self.boolTopStatus(common.GetNextReElectionNumber(height),common.RoleValidator);ok==false{
		log.Warn(Module,"height re-caculation needed for validator topology",height)
		if height==0{
			return nil
		}
		ReElectionHeight:=common.GetNextReElectionNumber(height)
		if err:=self.ToGenValidatorTop(ReElectionHeight-man.VerifyTopologyGenerateUpTime);err!=nil{
			return err
		}
	}
	return nil
}