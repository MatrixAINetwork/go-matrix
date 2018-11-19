//1542632059.875263
//1542631453.3738391
//1542630778.6039674
// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package reelection

import (
	"github.com/matrix/go-matrix/common"

	"github.com/matrix/go-matrix/log"
)

func (self *ReElection)boolTopStatus(hash common.Hash,types common.RoleType)bool{
	if _,_,err:=self.readElectData(types,hash);err!=nil{
		return false
	}
	return true
}
func (self *ReElection)checkTopGenStatus(hash common.Hash)error{
	height,err:=self.GetNumberByHash(hash)
	if err!=nil{
		return err
	}

	if ok:=self.boolTopStatus(hash,common.RoleMiner);ok==false{
		log.Warn(Module,"矿工拓扑图需要重新算 hash",hash.String())
		if height==0{
			return nil
		}
		 if err:=self.ToGenMinerTop(hash);err!=nil{
		 	return err
		 }

	}

	if ok:=self.boolTopStatus(hash,common.RoleValidator);ok==false{
		log.Warn(Module,"验证者拓扑图需要重新算 高度",height)
		if height==0{
			return nil
		}
		if err:=self.ToGenValidatorTop(hash);err!=nil{
			return err
		}
	}
	return nil
}