// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package common

import (
	"bytes"
	"math/big"
)

//RoleType
//type RoleType uint32

/*
const (
	RoleNil             RoleType = 0x001
	RoleDefault                  = 0x002
	RoleBucket                   = 0x004
	RoleBackupMiner              = 0x008
	RoleMiner                    = 0x010
	RoleInnerMiner               = 0x020
	RoleBackupValidator          = 0x040
	RoleValidator                = 0x080
	RoleBackupBroadcast          = 0x100
	RoleBroadcast                = 0x200
)*/

type ElectRoleType uint8

const (
	ElectRoleMiner           ElectRoleType = 0x00
	ElectRoleMinerBackUp     ElectRoleType = 0x01
	ElectRoleValidator       ElectRoleType = 0x02
	ElectRoleValidatorBackUp ElectRoleType = 0x03
)

func (ert ElectRoleType) Transfer2CommonRole() RoleType {
	switch ert {
	case ElectRoleMiner:
		return RoleMiner
	case ElectRoleMinerBackUp:
		return RoleBackupMiner
	case ElectRoleValidator:
		return RoleValidator
	case ElectRoleValidatorBackUp:
		return RoleBackupValidator
	}
	return RoleNil
}

func GetRoleTypeFromPosition(position uint16) RoleType {
	return ElectRoleType(position >> 12).Transfer2CommonRole()
}

func GeneratePosition(index uint16, electRole ElectRoleType) uint16 {
	return uint16(electRole)<<12 + index
}

const (
	MasterValidatorNum = 11
	BackupValidatorNum = 3
)


type VrfMsg struct {
	VrfValue []byte
	VrfProof []byte
	Hash Hash
}
func GetHeaderVrf(account []byte,vrfvalue []byte,vrfproof []byte)[]byte{
	var buf bytes.Buffer
	buf.Write(account)
	buf.Write(vrfvalue)
	buf.Write(vrfproof)

	return buf.Bytes()

}

func GetVrfInfoFromHeader(headerVrf []byte)([]byte,[]byte,[]byte){
	var account,vrfvalue,vrfproof []byte
	if len(headerVrf)>=33{
		account=headerVrf[0:33]
	}
	if (len(headerVrf)>=33+65){
		vrfvalue=headerVrf[33:33+65]
	}
	if (len(headerVrf)>=33+65+64){
		vrfproof=headerVrf[33+65:33+65+64]
	}

	return account,vrfvalue,vrfproof
}


func GetRoleVipGrade(aim uint64)int{
	switch  {
	case aim>=10000000:
		return 1
	case aim> 1000000&&aim<10000000:
		return 2
	default:
		return 0


	}
}
type Echelon struct {
	MinMoney *big.Int
	Quota    int
	Ratio    float64
}

var (
	ManValue                = new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	vip1     = new(big.Int).Mul(big.NewInt(100000), ManValue)
	vip2     = new(big.Int).Mul(big.NewInt(40000), ManValue)
	EchelonArrary = []Echelon{
		Echelon{
			MinMoney: vip1,
			Quota:    5,
			Ratio:    2.0,
		},
		Echelon{
			MinMoney: vip2,
			Quota:    3,
			Ratio:    1.0,
		},
	}
)