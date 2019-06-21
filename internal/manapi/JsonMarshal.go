// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package manapi

import (
	"encoding"
	"encoding/json"
	"github.com/MatrixAINetwork/go-matrix/base58"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/params"
	"math/big"
	"reflect"
)

var (
	addrType           = reflect.TypeOf(common.Address{})
	bigInt             = reflect.TypeOf(big.Int{})
	marshalerInterface = reflect.TypeOf(new(json.Marshaler)).Elem()
	textMarshaler      = reflect.TypeOf(new(encoding.TextMarshaler)).Elem()
	rawValueType       = reflect.TypeOf([]byte{})
)

func MakeJsonInferface(value interface{}) interface{} {
	val := reflect.ValueOf(value)
	return makeJsonInferface(val)
}
func makeJsonInferface(val reflect.Value) interface{} {
	valtype := val.Type()
	kind := valtype.Kind()
	switch {
	case valtype.AssignableTo(reflect.PtrTo(bigInt)):
		info := val.Interface().(*big.Int)
		if info == nil {
			return 0
		}
		return info.String()
	case valtype == addrType:
		return base58.Base58EncodeToString(params.MAN_COIN, val.Interface().(common.Address))
	case valtype == rawValueType || kind == reflect.String:
		if val.CanInterface() {
			return val.Interface()
		}
		return ""
	case valtype.Implements(marshalerInterface) || valtype.Implements(textMarshaler):
		if val.CanInterface() {
			return val.Interface()
		}
		return ""
	case kind != reflect.Ptr && reflect.PtrTo(valtype).Implements(marshalerInterface):
		if val.CanInterface() {
			return val.Interface()
		}
		return ""
	case kind == reflect.Slice || kind == reflect.Array:
		return makeJsonSlice(val)
	case kind == reflect.Map:
		return makeJsonMap(val)
	case kind == reflect.Struct:
		return makeJsonStruct(val)
	case kind == reflect.Ptr:
		return makeJsonInferface(val.Elem())
	default:
		if val.CanInterface() {
			return val.Interface()
		}
		return ""
	}

}
func makeJsonSlice(val reflect.Value) []interface{} {
	vlen := val.Len()
	info := make([]interface{}, 0, vlen)
	for i := 0; i < vlen; i++ {
		info = append(info, makeJsonInferface(val.Index(i)))
	}
	return info
}
func makeJsonMap(val reflect.Value) interface{} {
	info := make(map[string]interface{})
	for _, key := range val.MapKeys() {
		name := makeJsonInferface(key)
		nametype := reflect.TypeOf(name)
		if nametype == rawValueType {
			info[string(name.([]byte))] = makeJsonInferface(val.MapIndex(key))
		} else if nametype.Kind() == reflect.String {
			info[name.(string)] = makeJsonInferface(val.MapIndex(key))
		}
	}
	return info
}

func makeJsonStruct(val reflect.Value) map[string]interface{} {
	info := make(map[string]interface{})
	valtype := val.Type()
	for i := 0; i < valtype.NumField(); i++ {
		fieldi := valtype.Field(i)
		fieldV := val.Field(i)
		if !fieldV.CanInterface() {
			continue
		}
		name := fieldi.Tag.Get("json")
		if len(name) == 0 {
			name = fieldi.Name
		}
		info[name] = makeJsonInferface(fieldV)
	}
	return info
}
func MarshalInterface(value interface{}) (text []byte, err error) {
	info := MakeJsonInferface(value)
	return json.Marshal(info)
}
