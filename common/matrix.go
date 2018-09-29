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
package common

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
