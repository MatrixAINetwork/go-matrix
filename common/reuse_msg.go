// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

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

func (self ReuseMsgController) GetMsgList() []interface{} {
	result := make([]interface{}, 0)
	for _, value := range self.msgMap {
		result = append(result, value.msg)
	}

	return result
}

func (self ReuseMsgController) GetMsgByKey(msgKey Hash) interface{} {
	cache, exist := self.msgMap[msgKey]
	if !exist {
		return nil
	} else {
		return cache.msg
	}
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
