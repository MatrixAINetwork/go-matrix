// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package validatorGroup

import (
	"testing"
	"github.com/MatrixAINetwork/go-matrix/common"
)

func TestAddressSliceInsert(t *testing.T){
	testSlice := AddressSlice{}
	for i:=byte(1);i<255;i++{
		addr := common.Address{i}
		index,have := testSlice.Find(addr)
		if have {
			t.Fatal("Find Error")
		}
		if index != len(testSlice){
			t.Fatal("Find Error")
		}
		testSlice.Insert(addr)
		index,have = testSlice.Find(addr)
		if !have {
			t.Fatal("Find Error")
		}
		if index != len(testSlice)-1{
			t.Fatal("Find Error")
		}
	}
	testSlice1 := AddressSlice{}
	for i:=byte(255);i>0;i--{
		addr := common.Address{i}
		index,have := testSlice1.Find(addr)
		if have {
			t.Fatal("Find Error")
		}
		if index != 0{
			t.Fatal("Find Error")
		}
		testSlice1.Insert(addr)
		index,have = testSlice1.Find(addr)
		if !have {
			t.Fatal("Find Error")
		}
		if index != 0{
			t.Fatal("Find Error")
		}
	}
	t.Log(testSlice1)
}
func TestAddressSliceRemove1(t *testing.T){
	testSlice := AddressSlice{}
	testSlice.Insert(common.HexToAddress("0xa907e81ba7d52546e2aa374a1fdc929d6723c472"))
	testSlice.Insert(common.HexToAddress("0xaea4865313eb4178fa666bea27fab7c421a6db3e"))
	t.Log(testSlice)
	testSlice.Remove(common.HexToAddress("0xaea4865313eb4178fa666bea27fab7c421a6db3e"))
	t.Log(testSlice)

}
func TestAddressSliceRemove(t *testing.T){
	testSlice := AddressSlice{}
	for i:=byte(1);i<255;i++{
		addr := common.Address{i}
		index,have := testSlice.Find(addr)
		if have {
			t.Fatal("Find Error")
		}
		if index != len(testSlice){
			t.Fatal("Find Error")
		}
		testSlice.Insert(addr)
		index,have = testSlice.Find(addr)
		if !have {
			t.Fatal("Find Error")
		}
		if index != len(testSlice)-1{
			t.Fatal("Find Error")
		}
		if !testSlice.Remove(addr) {
			t.Fatal("Remove Error")
		}
	}
	testSlice1 := AddressSlice{}
	for i:=byte(255);i>0;i--{
		addr := common.Address{i}
		index,have := testSlice1.Find(addr)
		if have {
			t.Fatal("Find Error")
		}
		if index != 0{
			t.Fatal("Find Error")
		}
		testSlice1.Insert(addr)
		index,have = testSlice1.Find(addr)
		if !have {
			t.Fatal("Find Error")
		}
		if index != 0{
			t.Fatal("Find Error")
		}
		if i%10 != 0 {
			continue
		}
		if !testSlice1.Remove(addr) {
			t.Fatal("Remove Error")
		}
		index,have = testSlice1.Find(addr)
		if have {
			t.Fatal("Find Error")
		}
		if index != 0{
			t.Fatal("Find Error")
		}
		t.Log(testSlice1)
	}
}

func TestAddressValidatorSliceInsert(t *testing.T){
	testSlice := ValidatorInfoSlice{}
	for i:=byte(1);i<255;i++{
		addr := common.Address{i}
		index,have := testSlice.Find(addr)
		if have {
			t.Fatal("Find Error")
		}
		if index != len(testSlice){
			t.Fatal("Find Error")
		}
		testSlice.Insert(ValidatorInfo{Address:addr})
		index,have = testSlice.Find(addr)
		if !have {
			t.Fatal("Find Error")
		}
		if index != len(testSlice)-1{
			t.Fatal("Find Error")
		}
	}
	testSlice1 := ValidatorInfoSlice{}
	for i:=byte(255);i>0;i--{
		addr := common.Address{i}
		index,have := testSlice1.Find(addr)
		if have {
			t.Fatal("Find Error")
		}
		if index != 0{
			t.Fatal("Find Error")
		}
		testSlice1.Insert(ValidatorInfo{Address:addr})
		index,have = testSlice1.Find(addr)
		if !have {
			t.Fatal("Find Error")
		}
		if index != 0{
			t.Fatal("Find Error")
		}
	}
	t.Log(testSlice1)
}
func TestAddressValidatorSliceRemove(t *testing.T){
	testSlice := ValidatorInfoSlice{}
	for i:=byte(1);i<255;i++{
		addr := common.Address{i}
		index,have := testSlice.Find(addr)
		if have {
			t.Fatal("Find Error")
		}
		if index != len(testSlice){
			t.Fatal("Find Error")
		}
		testSlice.Insert(ValidatorInfo{Address:addr})
		index,have = testSlice.Find(addr)
		if !have {
			t.Fatal("Find Error")
		}
		if index != len(testSlice)-1{
			t.Fatal("Find Error")
		}
		if !testSlice.Remove(addr) {
			t.Fatal("Remove Error")
		}
	}
	testSlice1 := ValidatorInfoSlice{}
	for i:=byte(255);i>0;i--{
		addr := common.Address{i}
		index,have := testSlice1.Find(addr)
		if have {
			t.Fatal("Find Error")
		}
		if index != 0{
			t.Fatal("Find Error")
		}
		testSlice1.Insert(ValidatorInfo{Address:addr})
		index,have = testSlice1.Find(addr)
		if !have {
			t.Fatal("Find Error")
		}
		if index != 0{
			t.Fatal("Find Error")
		}
		if i%10 != 0 {
			continue
		}
		if !testSlice1.Remove(addr) {
			t.Fatal("Remove Error")
		}
		index,have = testSlice1.Find(addr)
		if have {
			t.Fatal("Find Error")
		}
		if index != 0{
			t.Fatal("Find Error")
		}
		t.Log(testSlice1)
	}
}