// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/MatrixAINetwork/go-matrix/common"
)

type Gen struct {
	NodeID, Account, Role string
}

var (
	gens         = make([]*Gen, 0)
	miner        = "21e19e0c9bab2400000"
	validator    = "152d02c7e14af6800000"
	m            = make(map[common.Hash]common.Hash)
	readFileName = "./genesis.txt"
	saveFileName = "./saveGenesis.txt"
)

func init() {
	file, err := os.Open(readFileName)
	if err != nil {
		os.Exit(1)
	}
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		os.Exit(1)
	}

	dadaSclice := strings.Split(string(data), "\n")
	for _, v := range dadaSclice {
		if len(v) == 0 {
			continue
		}
		sc := strings.Split(v, ",")
		if len(sc) != 3 {
			continue
		}
		if strings.HasPrefix(v, "0x") || strings.HasPrefix(v, "miner") || strings.HasPrefix(v, "validator") {
			continue
		}
		var gen = &Gen{}
		gen.NodeID = strings.Trim(sc[0], " ")
		gen.Account = strings.Trim(sc[1], " ")
		gen.Role = strings.Trim(strings.Trim(strings.Trim(sc[2], " "), "\r"), "\n")
		gens = append(gens, gen)
	}
	return
}

func TestCreateGenesisData(t *testing.T) {
	var str string
	for k, v := range gens {
		switch v.Role {
		case "miner":
			getJsonStr(v.NodeID, v.Account, miner, k)
		case "validator":
			getJsonStr(v.NodeID, v.Account, validator, k)
		}
	}
	bytes, _ := json.Marshal(m)
	str = string(bytes)
	save(t, str)
}

func save(t *testing.T, data string) {
	file, err := os.OpenFile(saveFileName, os.O_CREATE, 0644)
	if err != nil {
		t.Fatal("save data failed")
		os.Exit(1)
	}
	file.WriteString(data)
}
func getJsonStr(nodeID, account, miner string, num int) {
	nodex := fmt.Sprintf("0x%s", nodeID[:len(nodeID)/2])
	nodey := fmt.Sprintf("0x%s", nodeID[len(nodeID)/2:])
	//SET D
	deposit := common.HexToAddress(account)
	depositAddr := common.BytesToAddress([]byte{10})
	depositD := append(deposit[:], 'D')
	//set dnum
	depositDNUM := append(depositAddr[:], 'D', 'N', 'U', 'M')
	//set nx
	depositNX := append(deposit[:], 'N', 'X')
	//SET DI
	key := make([]byte, 8)
	depositDI := append(depositAddr[:], 'D', 'I')
	depositDI = append(depositDI, key...)
	depositNY := append(deposit[:], 'N', 'Y')

	if _, ok := m[common.BytesToHash(depositDI)]; ok {
		m[common.BytesToHash([]byte("depositDI"))] = common.BytesToHash(deposit[:])
	} else {
		m[common.BytesToHash(depositDI)] = common.BytesToHash(deposit[:])
	}
	m[common.BytesToHash(depositDNUM)] = common.HexToHash(fmt.Sprintf("%x", num+1))

	m[common.BytesToHash(depositNX)] = common.BytesToHash([]byte("nodex"))
	m[common.BytesToHash(depositNY)] = common.BytesToHash([]byte("nodey"))
	m[common.BytesToHash(depositD[:])] = common.HexToHash(miner)
	bytes, _ := json.Marshal(m)
	res := string(bytes)
	str := strings.Replace(res, "0x0000000000000000000000000000000000000000000000000000006e6f646578", nodex, 1)
	str = strings.Replace(str, "0x0000000000000000000000000000000000000000000000000000006e6f646579", nodey, 1)
	key1 := getNum(fmt.Sprintf("%x", num))
	rev := fmt.Sprintf("0x0000000000000000000000000000000000000000000a4449%s", key1)
	str = strings.Replace(str, "0x00000000000000000000000000000000000000000000006465706f7369744449", rev, 1)
	delete(m, common.BytesToHash(depositNX))
	delete(m, common.BytesToHash(depositNY))
	delete(m, common.BytesToHash([]byte("depositDI")))
	json.Unmarshal([]byte(str), &m)
}
func getNum(num string) string {
	if len(num) < 16 {
		num = fmt.Sprintf("%s%s", "0", num)
		return getNum(num)
	} else {
		return num
	}
}

func TestApplyMessage(t *testing.T) {
	fmt.Println(common.HexToHash("depositDI"))
}
