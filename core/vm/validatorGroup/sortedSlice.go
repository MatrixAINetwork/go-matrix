// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package validatorGroup

import (
	"github.com/MatrixAINetwork/go-matrix/common"
	"bytes"
	"reflect"
	"math/big"
)
type SortedSliceInterface interface {
	Greater(i int,b interface{})bool
	Less(i int,b interface{})bool
}
func Find(slice SortedSliceInterface,data interface{})(int,bool){
	sliceValue := reflect.ValueOf(slice)
	if sliceValue.Kind() != reflect.Ptr ||
		sliceValue.Elem().Kind() != reflect.Slice{
		return -1,false
	}
	sliceEle := sliceValue.Elem()
	length := sliceEle.Len()
	left, right, mid := 0, length-1, 0
	if right < 0 {
		return 0, false
	}
	for {
		mid = (left + right) / 2
		if slice.Greater(mid, data) {
			right = mid - 1
		} else if slice.Less(mid, data) {
			left = mid + 1
		} else {
			return mid, true
		}
		if left > right {
			return left, false
		}
	}
	return mid, false
}
func EnlargeSlice(val reflect.Value,newLen int){
	if newLen > val.Cap() {
		newcap := val.Cap() + val.Cap()/2
		if newcap < 4 {
			newcap = 4
		}
		newv := reflect.MakeSlice(val.Type(), val.Len(), newcap)
		reflect.Copy(newv, val)
		val.Set(newv)
	}
	if newLen > val.Len() {
		val.SetLen(newLen)
	}
}
func Insert(slice SortedSliceInterface,data interface{}){
	index,_ := Find(slice,data)
	if index<0{
		return
	}
	sliceValue := reflect.ValueOf(slice).Elem()
	length := sliceValue.Len()+1
	EnlargeSlice(sliceValue,length)
	for i := length-1;i > index; i--{
		sliceValue.Index(i).Set(sliceValue.Index(i-1))
	}
	sliceValue.Index(index).Set(reflect.ValueOf(data))
}
func Remove(slice SortedSliceInterface,data interface{})bool{
	index,exist := Find(slice,data)
	if !exist{
		return false
	}
	sliceValue := reflect.ValueOf(slice).Elem()
	length := sliceValue.Len()
	for i := index;i <length-1; i++{
		sliceValue.Index(i).Set(sliceValue.Index(i+1))
	}
	sliceValue.SetLen(length-1)
	return true
}
type AddressSlice []common.Address
func (ad AddressSlice)Greater(i int,b interface{})bool {
	addr := (b.(common.Address))
	return bytes.Compare(ad[i][:],addr[:])>0
}
func (ad AddressSlice)Less(i int,b interface{})bool {
	addr := (b.(common.Address))
	return bytes.Compare(ad[i][:],addr[:])<0
}
func (ad *AddressSlice)Find(address common.Address)(int,bool){
	return Find(ad,address)
}
func (ad *AddressSlice)Insert(address common.Address){
	Insert(ad,address)
}
func (ad *AddressSlice)Remove(address common.Address)bool{
	return Remove(ad,address)
}
//Participants Deposit info
type DepositPos struct {
	DType uint64
	Position uint64
	Amount* big.Int			`rlp:"-"`
	EndTime uint64			`rlp:"-"`
}
type CurrentData struct {
	Amount *big.Int
	PreAmount *big.Int
	Interest *big.Int
	WithdrawList []common.WithDrawInfo
}
type ValidatorInfo struct {
	Address common.Address
	Reward *big.Int
	AllAmount *big.Int 	`rlp:"-"`
	Current CurrentData
	Positions []DepositPos
}
func NewValidatorInfo(address common.Address)*ValidatorInfo{
	return &ValidatorInfo{
		Address:address,
		Reward:big.NewInt(0),
		AllAmount:big.NewInt(0),
		Positions:[]DepositPos{},
		Current:CurrentData{big.NewInt(0),big.NewInt(0),big.NewInt(0),[]common.WithDrawInfo{}},
	}
}
type ValidatorInfoSlice []ValidatorInfo
func (ad ValidatorInfoSlice)Greater(i int,b interface{})bool {
	addr := (b.(ValidatorInfo))
	return bytes.Compare(ad[i].Address[:],addr.Address[:])>0
}
func (ad ValidatorInfoSlice)Less(i int,b interface{})bool {
	addr := (b.(ValidatorInfo))
	return bytes.Compare(ad[i].Address[:],addr.Address[:])<0
}
func (ad *ValidatorInfoSlice)Find(address common.Address)(int,bool){
	return Find(ad,ValidatorInfo{Address:address})
}
func (ad *ValidatorInfoSlice)Insert(info ValidatorInfo){
	Insert(ad,info)
}
func (ad *ValidatorInfoSlice)Remove(address common.Address)bool{
	return Remove(ad,ValidatorInfo{Address:address})
}
/*
func (ad *AddressSlice)Find(address common.Address)(int,bool){
	left, right, mid := 0, len(*ad)-1, 0
	if right < 0 {
		return 0, false
	}
	for {
		mid = (left + right) / 2
		if greater((*ad)[mid], address) {
			right = mid - 1
		} else if less((*ad)[mid], address) {
			left = mid + 1
		} else {
			return mid, true
		}
		if left > right {
			return left, false
		}
	}
	return mid, false

}
func (ad *AddressSlice)Insert(address common.Address){
	index,_ := ad.Find(address)
	*ad = append(*ad,address)
	end := len(*ad) - 1
	for i := end; i > index; i-- {
		(*ad)[i], (*ad)[i-1] = (*ad)[i-1], (*ad)[i]
	}
}
func (ad *AddressSlice)Remove(address common.Address)bool{
	index,exist := ad.Find(address)
	if !exist{
		return false
	}
	*ad = append((*ad)[:index],(*ad)[index+1:]...)
	return true
}*/