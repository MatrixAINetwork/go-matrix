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

import (
	"errors"
	"time"
)

var (
	ErrMsgExist       = errors.New("msg already exist")
	ErrMsgNotExist    = errors.New("msg not exist")
	ErrUseMsgTooOften = errors.New("use msg too often, please try later")
)

type msgCache struct {
	msg     interface{}
	useTime int64
}

type ReuseMsgController struct {
	msgMap      map[Hash]*msgCache
	useInterval int64
}

func NewReuseMsgController(useInterval int64) *ReuseMsgController {
	return &ReuseMsgController{
		msgMap:      make(map[Hash]*msgCache),
		useInterval: useInterval,
	}
}

func (self ReuseMsgController) IsExistMsg(msgKey Hash) bool {
	_, exist := self.msgMap[msgKey]
	return exist
}

func (self ReuseMsgController) AddMsg(msgKey Hash, msg interface{}, lastUseTime int64) error {
	if self.IsExistMsg(msgKey) {
		return ErrMsgExist
	}
	self.msgMap[msgKey] = &msgCache{msg: msg, useTime: lastUseTime}
	return nil
}

func (self ReuseMsgController) ReUseMsg(msgKey Hash) (interface{}, error) {
	cache, exist := self.msgMap[msgKey]
	if !exist {
		return nil, ErrMsgNotExist
	}

	curTime := time.Now().Unix()
	if curTime-cache.useTime < self.useInterval {
		return nil, ErrUseMsgTooOften
	}

	cache.useTime = curTime
	return cache.msg, nil
}
