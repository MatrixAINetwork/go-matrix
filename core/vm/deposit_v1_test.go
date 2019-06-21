// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package vm

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/mandb"
	"github.com/MatrixAINetwork/go-matrix/params"
)

var p PrecompiledContract

func init() {
	p = PrecompiledContractsByzantium[common.BytesToAddress([]byte{10})]
}

//v退选
func TestMarkMatrixValiWithDraw(t *testing.T) {
	in, env, contract := tMatrixDeposit(t, p, "valiDeposit", validatorThreshold)
	//退选前提是参选成功,并且需要设置区块高度
	env.BlockNumber = big.NewInt(1)
	withDraw(t, in, p, env, contract, "withdraw")
	//退选 是可以无限期退选的
	env.BlockNumber = big.NewInt(1)
	withDraw(t, in, p, env, contract, "withdraw")
	fmt.Println("withdraw success")
}

//m退选
func TestMarkMatrixminerWithDraw(t *testing.T) {
	p := PrecompiledContractsByzantium[common.BytesToAddress([]byte{10})]
	in, env, contract := tMatrixDeposit(t, p, "valiDeposit", validatorThreshold)
	//退选前提是参选成功,并且需要设置区块高度
	env.BlockNumber = big.NewInt(1)
	withDraw(t, in, p, env, contract, "withdraw")
	fmt.Println("withdraw success")
}

//退选之后继续参选
func TestContract_Address(t *testing.T) {
	in := make([]byte, 4)
	copy(in[:4], depositAbi.Methods["valiDeposit"].Id())

	var nodeID = make([]byte, 64)
	copy(nodeID, []byte("ceaccac640adf55b2028469bd36ba501f28b699d"))
	bytes, _ := depositAbi.Methods["valiDeposit"].Inputs.Pack(nodeID)
	in = append(in, bytes...)
	reqGas := p.RequiredGas(in)
	contract := NewContract(AccountRef(common.HexToAddress("1337")),
		AccountRef(common.BytesToAddress([]byte{10})), validatorThreshold, reqGas)
	var (
		res  []byte
		err  error
		data = make([]byte, len(in))
	)
	//测试参选多次
	copy(data, in)
	statedb, _ := state.New(common.Hash{}, state.NewDatabase(mandb.NewMemDatabase()))
	env := NewEVM(Context{}, statedb, params.TestChainConfig, Config{})
	env.CanTransfer = func(db StateDB, address common.Address, amount *big.Int) bool {
		return true
	}
	env.Transfer = func(StateDB, common.Address, common.Address, *big.Int) { return }
	for i := 0; i < 2; i++ {
		t.Run(fmt.Sprintf("%s-Gas=%d", "deposit", contract.Gas), func(t *testing.T) {
			contract.Gas = reqGas
			res, err = RunPrecompiledContract(p, data, contract, env)
			//Check if it is correct
			if err != nil {
				t.Fatal(err)
				return
			}
			//filed.output
			if common.Bytes2Hex(res) != "01" {
				t.Error(fmt.Sprintf("Expected %v, got %v", "deposit", common.Bytes2Hex(res)))
				return
			}
		})
		env.BlockNumber = big.NewInt(1)
		withDraw(t, in, p, env, contract, "withdraw")
	}
	fmt.Println("参选成功后id值", env.StateDB)
	return
}

//v退款
func TestMarkMatrixValirefund(t *testing.T) {
	in, env, contract := tMatrixDeposit(t, p, "valiDeposit", validatorThreshold)
	//退选前提是参选成功,并且需要设置区块高度
	env.BlockNumber = big.NewInt(1)
	in, env, contract = withDraw(t, in, p, env, contract, "withdraw")
	//退款必须超过600块之后/12s*600才能申请退款
	env.BlockNumber = big.NewInt(700)
	refund(t, in, p, env, contract, "refund")

	//不能二次退款
	env.BlockNumber = big.NewInt(700)
	refund(t, in, p, env, contract, "refund")
	fmt.Println("refund success")
}

//查询所有参选节点
func TestGetDepositList(t *testing.T) {
	in, env, contract := tMatrixDeposit(t, p, "valiDeposit", validatorThreshold)

	var (
		res  []byte
		err  error
		data = make([]byte, 4)
	)

	copy(in[:4], depositAbi.Methods["getDepositList"].Id())
	reqGas := p.RequiredGas(in)
	t.Run(fmt.Sprintf("%s-Gas=%d", "deposit", contract.Gas), func(t *testing.T) {
		contract.Gas = reqGas
		copy(data, in[:4])
		res, err = RunPrecompiledContract(p, data, contract, env)
		if err != nil {
			t.Fatal(err)
			return
		}
		var addrList []common.Address
		depositAbi.Methods["getDepositList"].Outputs.Unpack(&addrList, res)
		fmt.Printf("%+v\n", addrList)
	})
}

//测试抵押1000次 可以任意抵押
func TestMarkMatrixminerDeposit2(t *testing.T) {
	tMatrixMinerDeposit(t, p, "minerDeposit")
}

func tMatrixMinerDeposit(t *testing.T, p PrecompiledContract, deposit string) (in []byte, env *EVM, contract *Contract) {
	in = make([]byte, 4)
	copy(in[:4], depositAbi.Methods[deposit].Id())

	var nodeID = make([]byte, 64)
	copy(nodeID, []byte("ceaccac640adf55b2028469bd36ba501f28b699d"))
	bytes, _ := depositAbi.Methods["valiDeposit"].Inputs.Pack(nodeID)
	in = append(in, bytes...)
	reqGas := p.RequiredGas(in)
	contract = NewContract(AccountRef(common.HexToAddress("1337")),
		AccountRef(common.BytesToAddress([]byte{10})), validatorThreshold, reqGas)
	var (
		res  []byte
		err  error
		data = make([]byte, len(in))
	)
	//测试参选多次
	copy(data, in)
	statedb, _ := state.New(common.Hash{}, state.NewDatabase(mandb.NewMemDatabase()))
	env = NewEVM(Context{}, statedb, params.TestChainConfig, Config{})
	env.CanTransfer = func(db StateDB, address common.Address, amount *big.Int) bool {
		return true
	}
	env.Transfer = func(StateDB, common.Address, common.Address, *big.Int) { return }
	for i := 0; i < 1000; i++ {
		t.Run(fmt.Sprintf("%s-Gas=%d", "deposit", contract.Gas), func(t *testing.T) {
			contract.Gas = reqGas
			res, err = RunPrecompiledContract(p, data, contract, env)
			//Check if it is correct
			if err != nil {
				t.Fatal(err)
				return
			}
			//filed.output
			if common.Bytes2Hex(res) != "01" {
				t.Error(fmt.Sprintf("Expected %v, got %v", "deposit", common.Bytes2Hex(res)))
				return
			}
		})
	}
	fmt.Println("参选成功后id值", env.StateDB)
	return
}

//查询info信息
func TestGetDepositInfo(t *testing.T) {
	in, env, contract := tMatrixDeposit(t, p, "valiDeposit", validatorThreshold)

	//查询参选成功节点集合
	var (
		res      []byte
		err      error
		data     = make([]byte, len(in))
		addrList []common.Address
	)
	copy(in[:4], depositAbi.Methods["getDepositList"].Id())
	reqGas := p.RequiredGas(in)
	t.Run(fmt.Sprintf("%s-Gas=%d", "deposit", contract.Gas), func(t *testing.T) {
		contract.Gas = reqGas
		copy(data, in)
		res, err = RunPrecompiledContract(p, data, contract, env)
		if err != nil {
			t.Fatal(err)
			return
		}

		depositAbi.Methods["getDepositList"].Outputs.Unpack(&addrList, res)
		fmt.Printf("%+v\n", addrList)
	})

	//查询详细信息
	copy(in[:4], depositAbi.Methods["getDepositInfo"].Id())
	reqGas = p.RequiredGas(in)
	t.Run(fmt.Sprintf("%s-Gas=%d", "deposit", contract.Gas), func(t *testing.T) {
		contract.Gas = reqGas

		copy(in[4:], addrList[0][:])
		bytes, _ := depositAbi.Methods["getDepositInfo"].Inputs.Pack(addrList[0])
		input := make([]byte, len(bytes)+4)
		copy(input[:4], in[:4])
		copy(input[4:], bytes)
		res, err = RunPrecompiledContract(p, input, contract, env)
		if err != nil {
			t.Fatal(err)
			return
		}
		Deposit := big.NewInt(0)
		NodeID := make([]byte, 0)
		Withdraw := big.NewInt(0)

		fmt.Println("res", res)
		err := depositAbi.Methods["getDepositInfo"].Outputs.Unpack(&[]interface{}{&Deposit, &NodeID, &Withdraw}, res)
		if err != nil {
			fmt.Println("error:", err)
		}
		fmt.Printf("Deposit:%+v NodeID:%s Withdraw:%+v\n", Deposit, NodeID, Withdraw)
	})
}

func refund(t *testing.T, in []byte, p PrecompiledContract, env *EVM, contract *Contract, refund string) {
	var (
		res  []byte
		err  error
		data = make([]byte, len(in))
	)

	copy(in[:4], depositAbi.Methods[refund].Id())
	reqGas := p.RequiredGas(in)
	t.Run(fmt.Sprintf("%s-Gas=%d", "deposit", contract.Gas), func(t *testing.T) {
		contract.Gas = reqGas
		copy(data, in)
		res, err = RunPrecompiledContract(p, data, contract, env)
		if err != nil {
			t.Fatal(err)
			return
		}
		if common.Bytes2Hex(res) != "01" {
			fmt.Println("tt", common.Bytes2Hex(res))
			t.Error(fmt.Sprintf("expected %v, got %v", "deposit", common.Bytes2Hex(res)))
			return
		}
	})
	fmt.Println("refund id", env.StateDB)
}

func withDraw(t *testing.T, in []byte, p PrecompiledContract, env *EVM, contract *Contract, withdraw string) ([]byte, *EVM, *Contract) {
	var (
		res  []byte
		err  error
		data = make([]byte, len(in))
	)

	copy(in[:4], depositAbi.Methods[withdraw].Id())
	reqGas := p.RequiredGas(in)
	t.Run(fmt.Sprintf("%s-Gas=%d", "deposit", contract.Gas), func(t *testing.T) {
		contract.Gas = reqGas
		copy(data, in)
		res, err = RunPrecompiledContract(p, data, contract, env)
		if err != nil {
			t.Fatal(err)
			return
		}
		if common.Bytes2Hex(res) != "01" {
			t.Error(fmt.Sprintf("Expected %v, got %v", "deposit", common.Bytes2Hex(res)))
			return
		}
	})
	fmt.Println("withDraw id", env.StateDB)
	return in, env, contract
}

func TestManConst(t *testing.T) {
	man = new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	minerThreshold := new(big.Int).Mul(big.NewInt(1000), man)
	validatorThreshold := new(big.Int).Mul(big.NewInt(10000), man)
	fmt.Println(man)
	fmt.Println(minerThreshold)
	fmt.Println(validatorThreshold)
}

func TestCMP(t *testing.T) {
	//1  2  -1
	//1  1   0
	//1  0   1
	newInt := big.NewInt(1)
	newInt2 := big.NewInt(-1)
	//前面大 返回1
	//相等   返回0
	//后面大 返回-1
	fmt.Println(newInt.Cmp(newInt2))
	//bigInt相加
	newInt.Add(newInt, big.NewInt(600))
	fmt.Println(newInt)
}

func TestNewByzantiumInstructionSet(t *testing.T) {
	deposit := big.NewInt(0)
	if deposit.Sign() == 0 {
		fmt.Println("hello")
	}
}

func TestGetMinerDepositList(t *testing.T) {
	zj := big.NewFloat(100000000000000000000000)
	v := big.NewFloat(10000000000000000000000)
	if zj.Cmp(v) >= 0 {
		fmt.Println("zj 大")
	} else {
		fmt.Println("zj 小")
	}
}

func TestInt(t *testing.T) {
	var cc Ccc
	cc.ca = big.NewInt(11)
	demo(cc.ca)
}

func demo(c *big.Int) {
	i := c.Int64()
	fmt.Println(i)
	fmt.Printf("%T", c)
	fmt.Printf("%T", i)
}

type Ccc struct {
	ca *big.Int
}

//vali测试
func TestMarkMatrixvaliDeposit(t *testing.T) {
	tMatrixDeposit(t, p, "valiDeposit", validatorThreshold)
}

//miner测试
func TestMarkMatrixminerDeposit(t *testing.T) {
	tMatrixDeposit(t, p, "minerDeposit", minerThreshold)
}

//0x8aee9b014f13024d7ff307fe5583d31c9b79ec1c
//0x74ac366ca20182623a0b482665b08f7476c4ff66aabf3efb8b57cb0b6b94c69fdc3b8918af484a9434d2c2b3dff03a7d8b8d35496f583d118181f2633035d5ef

//0x13a370af4c4f0af90c00bbf7d270721ee4826be8
//0xc1587d8a2efcdcf78d291d77f6bef801bbe1cb7d8823c3c7159116a5d858662139974d899f06b1aa82041cda13e3d8e1d7d299dffed2e637af87c0f4b334aa31

//郑贺 v
//0xfabff5c20c795aa698c23a3e2a02570c9e0bb020
//0xf26fa4112f2cc603a114d3eec20d5a4605debe1c3cecc36c347982aaf3c30e5790b9f936c2c4e6862e615255cb1bce05a1578f4e9766ab43991c87864d3ff1fe

//叶营 m
//0x6b4701e32477232d50b8110fd13ba5fb9abe937a
//0x88e3f601edac6f553ec5d65ed5e543e858da1590ae3ad922f3e077189823b4d1b24e5028de6424ed59ffe76e88e52811ac879e7f5b3b1bc45e71090811653168
func tMatrixDeposit(t *testing.T, p PrecompiledContract, deposit string, threshold *big.Int) (in []byte, env *EVM, contract *Contract) {
	in = make([]byte, 4)
	copy(in[:4], depositAbi.Methods[deposit].Id())

	var addr = make([]byte, 20)
	var temp = make([]byte, 12)

	copy(addr, []byte("05e3c16931c6e578f948231dca609d754c18fc09"))
	//bytes, _ := depositAbi.Methods[deposit].Inputs.Pack(addr)
	in = append(in, temp...)
	in = append(in, addr...)
	reqGas := p.RequiredGas(in)
	contract = NewContract(AccountRef(common.HexToAddress("0xfabff5c20c795aa698c23a3e2a02570c9e0bb020")),
		AccountRef(common.BytesToAddress([]byte{10})), threshold, reqGas)
	var (
		res  []byte
		err  error
		data = make([]byte, len(in))
	)

	//contract.Gas = reqGas
	copy(data, in)
	statedb, _ := state.New(common.Hash{}, state.NewDatabase(mandb.NewMemDatabase()))
	env = NewEVM(Context{}, statedb, params.TestChainConfig, Config{})
	env.CanTransfer = func(db StateDB, address common.Address, amount *big.Int) bool {
		return true
	}
	env.Transfer = func(StateDB, common.Address, common.Address, *big.Int) { return }
	res, err = RunPrecompiledContract(p, data, contract, env)
	//Check if it is correct
	if err != nil {
		t.Fatal(err)
		return
	}
	//filed.output
	if common.Bytes2Hex(res) != "01" {
		t.Error(fmt.Sprintf("Expected %v, got %v", "deposit", common.Bytes2Hex(res)))
		return
	}

	fmt.Println("参选成功后id值", env.StateDB)
	return
}

var m = make(map[common.Hash]common.Hash)

//批量设置v和miner节点到创始区块
func TestSetGenesis(t *testing.T) {
	//nodeid:字符串
	//账户信息:0x开头
	tests := []struct {
		nodeid  string
		account string
		role    string
	}{

		{"dbf8dcc4c82eb2ea2e1350b0ea94c7e29f5be609736b91f0faf334851d18f8de1a518def870c774649db443fbce5f72246e1c6bc4a901ef33429fdc3244a93b3",
			"0x6a3217d128a76e4777403e092bde8362d4117773", "vali"},
		{"b624a3fb585a48b4c96e4e6327752b1ba82a90a948f258be380ba17ead7c01f6d4ad43d665bb11c50475c058d3aad1ba9a35c0e0c4aa118503bf3ce79609bef6",
			"0x0ead6cdb8d214389909a535d4ccc21a393dddba9", "vali"},
		{"80606b6c1eecb8ce91ca8a49a5a183aa42f335eb0d8628824e715571c1f9d1d757911b80ebc3afab06647da228f36ecf1c39cb561ef7684467c882212ce55cdb",
			"0x8c3d1a9504a36d49003f1652fadb9f06c32a4408", "vali"},
		{"43b553fae2184b25e76b69a2386bfc9a014486db7da3df75bba9fa2e3eed8aaf063a5f1aab68488a8645fd6a230a27bfe4e8d3393232fe107ba0f68a9bf541ad",
			"0x05e3c16931c6e578f948231dca609d754c18fc09", "vali"},
		{"8ce7defe2dde8297f7b55dd9ba8c5e13e0274371b716250ea0dd725974fa076ca379fc7226789a91678f4e38f8f60f8e6405ec9539cab77d4822614e80f743cf",
			"0x55fbba0496ef137be57d4c179a3a74c4d0c643be", "vali"},
		{"9f237f9842f70b0417d2c25ce987248c991310b2bd4034e300a6eec46b517bd8c4f7f31f157128d0732786181a481bcf725c41a655bdcce282a4bc95638d9aae",
			"0x915b5295dde0cebb11c6cb25828b546a9b2c9369", "vali"},
		{"68315573b123b44367f9fefcce38c4d5c4d5d2caf04158a9068de2060380b81f26b220543de7402745160141f932012a792722fd4dd2a7a2751771097eeef5f2",
			"0x92e0fea9aba517398c2f0dd628f8cfc7e32ba984", "vali"},
		{"bc5e761c9d0ba42f22433be14973b399662456763f033a4cdbb8ec37b80266526e6c56f92d0591825c7d644e487fcee828d537c58ce583a72578309ec6ebbd39",
			"0x7eb0bcd103810a6bf463d6d230ebcacc85157d19", "vali"},
		{"25ea3bca7679192612aed14d5e83a4f2a30824ff2af705d2d7c6795470f9cbbc258d9b102a726c3982cda6c4732ba3715551b6fbf9c0ae4ddca4a6c80bc4bbe9",
			"0xcded44bd41476a69e8e68ba8286952c414d28af7", "vali"},
		{"14f62dfd8826734fe75120849e11614b0763bc584fba4135c2f32b19501525d55d217742893801ecc871023fc42ed7e80196357fb5b1f762d181e827e626637d",
			"0x9cde10b889fca53c0a560b90b3cb21c2fc965d2b", "vali"},
		{"df57387d6505d0f71d7000da9642cf16d44feb7fcaa5f3a8a7d9fa58b6cbb6d33d145746d4fb544c049d3ff9b534bf9245a5b8052231c51695fd298032bd4a79",
			"0x7823a1bea7aca2f902b87effdd4da9a7ef1fc5fb", "vali"},

		{"a9f94b62067e993f3f02ada1a448c70ae90bdbe4c6b281f8ff16c6f4832e0e9aba1827531c260b380c776938b9975ac7170a7e822f567660333622ea3e529313",
			"0x0a3f28de9682df49f9f393931062c5204c2bc404", "mainer"},
	}

	var miner = "21e19e0c9bab2400000"
	var vali = "152d02c7e14af6800000"
	for k, test := range tests {
		switch test.role {
		case "miner":
			getJsonStr(test.nodeid, test.account, miner, k)
		case "vali":
			getJsonStr(test.nodeid, test.account, vali, k)
		}
	}
	bytes, _ := json.Marshal(m)
	fmt.Println(string(bytes))
}

func getJsonStr(nodeID, account, miner string, num int) {
	nodex := fmt.Sprintf("0x%s", nodeID[:len(nodeID)/2])
	nodey := fmt.Sprintf("0x%s", nodeID[len(nodeID)/2:])
	//statedb, _ := state.New(common.Hash{}, state.NewDatabase(mandb.NewMemDatabase()))
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
	//statedb.SetState(common.BytesToAddress([]byte{10}),common.BytesToHash(depositDI),common.BytesToHash(deposit[:]))
	//set nx
	//statedb.SetState(common.BytesToAddress([]byte{10}),common.BytesToHash(depositD[:]),common.HexToHash("21e19e0c9bab2400000"))
	//statedb.SetState(common.BytesToAddress([]byte{10}),common.BytesToHash(depositDNUM),common.HexToHash("1"))//index+1
	//m[common.BytesToHash(depositD[:])]=common.HexToHash("21e19e0c9bab2400000")
	depositNY := append(deposit[:], 'N', 'Y')
	if _, ok := m[common.BytesToHash(depositDI)]; ok {
		m[common.HexToHash("depositDI")] = common.BytesToHash(deposit[:])
	} else {
		m[common.BytesToHash(depositDI)] = common.BytesToHash(deposit[:])
	}

	m[common.BytesToHash(depositDNUM)] = common.HexToHash(fmt.Sprintf("%d", num+1))
	m[common.BytesToHash(depositNX)] = common.BytesToHash([]byte("nodex"))
	m[common.BytesToHash(depositNY)] = common.BytesToHash([]byte("nodey"))
	m[common.BytesToHash(depositD[:])] = common.HexToHash(miner)
	bytes, _ := json.Marshal(m)
	res := string(bytes)
	str := strings.Replace(res, "0x0000000000000000000000000000000000000000000000000000006e6f646578", nodex, 1)
	str = strings.Replace(str, "0x0000000000000000000000000000000000000000000000000000006e6f646579", nodey, 1)
	key1 := getNum(fmt.Sprintf("%d", num))
	rev := fmt.Sprintf("0x0000000000000000000000000000000000000000000a4449%s", key1)
	str = strings.Replace(str, "0x000000000000000000000000000000000000000000000000000000000000000d", rev, 1)
	delete(m, common.BytesToHash(depositNX))
	delete(m, common.BytesToHash(depositNY))
	delete(m, common.HexToHash("depositDI"))
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
