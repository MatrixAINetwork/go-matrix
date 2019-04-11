// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package common

import (
	"math/big"
	"reflect"
)

type ElectRoleType uint8

const (
	ElectRoleMiner              ElectRoleType = 0x00
	ElectRoleMinerBackUp        ElectRoleType = 0x01
	ElectRoleValidator          ElectRoleType = 0x02
	ElectRoleValidatorBackUp    ElectRoleType = 0x03
	ElectRoleCandidateValidator ElectRoleType = 0x04
	ElectRoleNil                ElectRoleType = 0xff
)

var (
	SignLog = "SignLog"
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
	case ElectRoleCandidateValidator:
		return RoleCandidateValidator
	}
	return RoleNil
}

func GetRoleTypeFromPosition(position uint16) RoleType {
	return ElectRoleType(position >> 12).Transfer2CommonRole()
}

func GeneratePosition(index uint16, electRole ElectRoleType) uint16 {
	if electRole >= ElectRoleNil {
		return 0xffff
	}
	return uint16(electRole)<<12 + index
}

type Echelon struct {
	MinMoney *big.Int
	MaxNum   int
	Ratio    uint16
}

var (
	ManValue = new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
)

func IsNil(i interface{}) bool {
	vi := reflect.ValueOf(i)
	if vi.Kind() == reflect.Ptr {
		flag := vi.IsNil()
		return flag
	}
	return false
}

type VIPRoleType uint16

const (
	VIP_Nil VIPRoleType = 0
	VIP_1   VIPRoleType = 1
	VIP_2   VIPRoleType = 2
	VIP_3   VIPRoleType = 3
)

func GetVIPLevel(index int) VIPRoleType {
	switch index {
	case 1:
		return VIP_1
	case 2:
		return VIP_2
	case 3:
		return VIP_3
	default:
		return VIP_Nil
	}
}
