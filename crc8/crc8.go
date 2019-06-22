// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
// Package crc8 implements the 8-bit cyclic redundancy check, or CRC-8, checksum.
//
// It provides parameters for the majority of well-known CRC-8 algorithms.
package crc8

//import "github.com/sigurn/utils"

// Params represents parameters of a CRC-8 algorithm including polynomial and initial value.
// More information about algorithms parametrization and parameter descriptions
// can be found here - http://www.zlib.net/crc_v3.txt
type Params struct {
	Poly   uint8
	Init   uint8
	RefIn  bool
	RefOut bool
	XorOut uint8
	Check  uint8
	Name   string
}

// Predefined CRC-8 algorithms.
// List of algorithms with their parameters borrowed from here - http://reveng.sourceforge.net/crc-catalogue/1-15.htm#crc.cat-bits.8
//
// The variables can be used to create Table for the selected algorithm.
var (
	CRC8          = Params{0x07, 0x00, false, false, 0x00, 0xF4, "CRC-8"}
	CRC8_CDMA2000 = Params{0x9B, 0xFF, false, false, 0x00, 0xDA, "CRC-8/CDMA2000"}
	CRC8_DARC     = Params{0x39, 0x00, true, true, 0x00, 0x15, "CRC-8/DARC"}
	CRC8_DVB_S2   = Params{0xD5, 0x00, false, false, 0x00, 0xBC, "CRC-8/DVB-S2"}
	CRC8_EBU      = Params{0x1D, 0xFF, true, true, 0x00, 0x97, "CRC-8/EBU"}
	CRC8_I_CODE   = Params{0x1D, 0xFD, false, false, 0x00, 0x7E, "CRC-8/I-CODE"}
	CRC8_ITU      = Params{0x07, 0x00, false, false, 0x55, 0xA1, "CRC-8/ITU"}
	CRC8_MAXIM    = Params{0x31, 0x00, true, true, 0x00, 0xA1, "CRC-8/MAXIM"}
	CRC8_ROHC     = Params{0x07, 0xFF, true, true, 0x00, 0xD0, "CRC-8/ROHC"}
	CRC8_WCDMA    = Params{0x9B, 0x00, true, true, 0x00, 0x25, "CRC-8/WCDMA"}
)

// Table is a 256-byte table representing polinomial and algorithm settings for efficient processing.
type Table struct {
	params Params
	data   [256]uint8
}

func ReverseByte(val byte) byte {
	var rval byte = 0
	for i := uint(0); i < 8; i++ {
		if val&(1<<i) != 0 {
			rval |= 0x80 >> i
		}
	}
	return rval
}

func ReverseUint8(val uint8) uint8 {
	return ReverseByte(val)
}

func ReverseUint16(val uint16) uint16 {
	var rval uint16 = 0
	for i := uint(0); i < 16; i++ {
		if val&(uint16(1)<<i) != 0 {
			rval |= uint16(0x8000) >> i
		}
	}
	return rval
}

// MakeTable returns the Table constructed from the specified algorithm.
func MakeTable(params Params) *Table {
	table := new(Table)
	table.params = params
	for n := 0; n < 256; n++ {
		crc := uint8(n)
		for i := 0; i < 8; i++ {
			bit := (crc & 0x80) != 0
			crc <<= 1
			if bit {
				crc ^= params.Poly
			}
		}
		table.data[n] = crc
	}
	return table
}

// Init returns the initial value for CRC register corresponding to the specified algorithm.
func Init(table *Table) uint8 {
	return table.params.Init
}

// Update returns the result of adding the bytes in data to the crc.
func Update(crc uint8, data []byte, table *Table) uint8 {
	for _, d := range data {
		if table.params.RefIn {
			d = ReverseByte(d)
		}
		crc = table.data[crc^d]
	}

	return crc
}

// Complete returns the result of CRC calculation and post-calculation processing of the crc.
func Complete(crc uint8, table *Table) uint8 {
	if table.params.RefOut {
		crc = ReverseUint8(crc)
	}

	return crc ^ table.params.XorOut
}

// Checksum returns CRC checksum of data usign scpecified algorithm represented by the Table.
func Checksum(data []byte, table *Table) uint8 {
	crc := Init(table)
	crc = Update(crc, data, table)
	return Complete(crc, table)
}

func CalCRC8(data []byte) uint8 {
	table := MakeTable(CRC8)
	return Checksum(data, table)
}
