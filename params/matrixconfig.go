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
package params

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/matrix/go-matrix/log"
)

const (
	VerifyNetChangeUpTime = 6
	MinerNetChangeUpTime  = 4

	VerifyTopologyGenerateUpTime = 8
	MinerTopologyGenerateUptime  = 8

	RandomVoteTime = 5
	HCSIM          = 1
	HCP2P          = 2
)

var (
	SignAccount         = "0xc47d9e507c1c5cb65cc7836bb668549fc8f547df"
	SignAccountPassword = "12345"
	HcMethod            = HCP2P
)

const (
	//TODO: VotePoolTimeout
	VotePoolTimeout    = 30 * 1000
	VotePoolCountLimit = 5
)

func init() {
	fmt.Println("PARAMS INIT")
	JsonParse := NewJsonStruct()
	v := Config{}
	JsonParse.Load("./man.json", &v)
	MainnetBootnodes = v.BootNode
	BroadCastNodes = v.BroadNode
}

type Config struct {
	BootNode  []string
	BroadNode []BroadCastNode
}

type JsonStruct struct {
}

func NewJsonStruct() *JsonStruct {
	return &JsonStruct{}
}

func (jst *JsonStruct) Load(filename string, v interface{}) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Error("ATTTTTTTTTTTTTTTTTTTEEEEEEEEEEEEEEBNNNNNNNNNNNNNTTTTTTTTTTTTTIIIIIIIIIIIOOOOOOOOOOOOOOONNNNNNNNNNNNNN", "filename", filename)
		return
	}
	err = json.Unmarshal(data, v)
	if err != nil {
		return
	}
}
