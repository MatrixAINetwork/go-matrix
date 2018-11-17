// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
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