package rlp

import (
	"testing"
)

type testInterface interface {
	test1()
	test2()
	test3()
	GetConstructorType()uint16
}
type testStruct1 struct {
	A uint64
	B uint64
	C uint64
}
func (t *testStruct1)test1(){

}
func (t *testStruct1)test2(){

}
func (t *testStruct1)test3(){

}
func (t *testStruct1)GetConstructorType()uint16{
	return 10
}
type testStruct2 struct {
	A uint64
	B uint64
	C uint64
	D uint64
}
func (t *testStruct2)test1(){

}
func (t *testStruct2)test2(){

}
func (t *testStruct2)test3(){

}
func (t *testStruct2)GetConstructorType()uint16{
	return 20
}
type testStruct struct {
	Test1 testInterface //`rlp:"interface"`
	Test2 testInterface //`rlp:"interface"`
//	Test3
}
func TestDecodeInterface1(t *testing.T) {
	testRlp := testStruct{&testStruct1{100,100,100},&testStruct2{100,100,100,100}}
	b,_ := EncodeToBytes(testRlp)
	t.Log(b)
	testRlp1 := testStruct{}
//	testSlice1 := []testInterface{}
	InterfaceConstructorMap[testRlp.Test1.GetConstructorType()] = func()interface{}{
		return &testStruct1{}
	}
	InterfaceConstructorMap[testRlp.Test2.GetConstructorType()] = func()interface{}{
		return &testStruct2{}
	}
	DecodeBytes(b,&testRlp1)
	t.Log(testRlp1.Test1,testRlp1.Test2)
}
func TestDecodeInterface(t *testing.T) {
	testSlice := []testInterface{}
	test1 := testStruct1{100,100,100}
	testSlice = append(testSlice,&testStruct1{100,100,100},&testStruct1{200,200,200})
	testSlice = append(testSlice,&testStruct2{100,100,100,100},&testStruct2{100,100,100,100})
	InterfaceConstructorMap[test1.GetConstructorType()] = func()interface{}{
		return &testStruct1{}
	}
	InterfaceConstructorMap[testSlice[2].GetConstructorType()] = func()interface{}{
		return &testStruct2{}
	}
	b1,_ := EncodeToBytes(test1)
	b,_ := EncodeToBytes(testSlice)
	testSlice1 := []testInterface{}
	DecodeBytes(b,&testSlice1)
	DecodeBytes(b1,test1)
	t.Log(testSlice1[0],testSlice1[1],testSlice1[2],testSlice1[3])
	t.Log(test1)
}
