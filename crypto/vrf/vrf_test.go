// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package vrf

import (
	"crypto/ecdsa"
	"fmt"
	"testing"

	"github.com/MatrixAINetwork/go-matrix/accounts/keystore"
	"github.com/MatrixAINetwork/go-matrix/baseinterface"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/crypto"
)

func TestCompute_1(t *testing.T) {
	//指针正常，msg正常
	serveice := baseinterface.NewVrf()
	pri, _ := crypto.GenerateKey()

	msg := []byte("test")
	vrf, proof, err := serveice.ComputeVrf(pri, msg)
	fmt.Println("init", vrf, proof, err)
	err = serveice.VerifyVrf(&pri.PublicKey, msg, vrf, proof)
	fmt.Println("init", "verify err", err)

	for index := 1; index <= 10; index++ {
		vrf_temp, proof_temp, err_temp := serveice.ComputeVrf(pri, msg)
		fmt.Println("index", vrf_temp, proof_temp, err_temp)

		err = serveice.VerifyVrf(&pri.PublicKey, msg, vrf_temp, proof_temp)
		fmt.Println("index", "verify err", err)

		/*		if string(vrf_temp)!=string(vrf)||string(proof_temp)!=string(proof)||err!=err_temp{
				fmt.Println("index",index,"vrf",string(vrf),"vrf_temp",string(vrf_temp))
				t.Fatalf("生成结果不一致")
			}*/
	}

}

func TestCompute_2(t *testing.T) {
	//指针正常，msg为空
	serveice := baseinterface.NewVrf()
	pri, _ := crypto.GenerateKey()

	msg := []byte("")
	vrf, proof, err := serveice.ComputeVrf(pri, msg)
	fmt.Println("init", vrf, proof, err)
	err = serveice.VerifyVrf(&pri.PublicKey, msg, vrf, proof)
	fmt.Println("init", "verify err", err)

	for index := 1; index <= 10; index++ {
		vrf_temp, proof_temp, err_temp := serveice.ComputeVrf(pri, msg)
		fmt.Println("index", vrf_temp, proof_temp, err_temp)

		err = serveice.VerifyVrf(&pri.PublicKey, msg, vrf_temp, proof_temp)
		fmt.Println("index", "verify err", err)

		/*		if string(vrf_temp)!=string(vrf)||string(proof_temp)!=string(proof)||err!=err_temp{
				fmt.Println("index",index,"vrf",string(vrf),"vrf_temp",string(vrf_temp))
				t.Fatalf("生成结果不一致")
			}*/
	}

}

func TestCompute_3(t *testing.T) {
	//生成阶段-公钥为空内容
	serveice := baseinterface.NewVrf()
	//	pri, _ := crypto.GenerateKey()

	pri := &ecdsa.PrivateKey{}
	msg := []byte("")
	vrf, proof, err := serveice.ComputeVrf(pri, msg)
	fmt.Println(vrf, proof, err)

}
func TestCompute_3_1(t *testing.T) {
	//生成阶段-公钥为空指针
	serveice := baseinterface.NewVrf()
	//	pri, _ := crypto.GenerateKey()

	msg := []byte("")
	vrf, proof, err := serveice.ComputeVrf(nil, msg)
	fmt.Println(vrf, proof, err)

}

func TestCompute_4(t *testing.T) {
	//验证阶段-公钥为空内容
	serveice := baseinterface.NewVrf()
	pri, _ := crypto.GenerateKey()

	msg := []byte("")
	vrf, proof, err := serveice.ComputeVrf(pri, msg)
	fmt.Println("init", vrf, proof, err)
	err = serveice.VerifyVrf(&ecdsa.PublicKey{}, msg, vrf, proof)
	fmt.Println("init", "verify err", err)

}
func TestCompute_5(t *testing.T) {
	//验证阶段-公钥为空指针
	serveice := baseinterface.NewVrf()
	pri, _ := crypto.GenerateKey()

	msg := []byte("")
	vrf, proof, err := serveice.ComputeVrf(pri, msg)
	fmt.Println("init", vrf, proof, err)
	err = serveice.VerifyVrf(nil, msg, vrf, proof)
	fmt.Println("init", "verify err", err)

}

func TestVrf_1(t *testing.T) {
	serveice := baseinterface.NewVrf()
	//正常测试
	pri, err := crypto.GenerateKey()
	pub := pri.PublicKey

	msg := []byte("")
	vrf, proof, err := serveice.ComputeVrf(pri, msg)
	fmt.Println("len vrf", len(vrf), "len proof", len(proof))
	if err != nil {
		t.Fatalf("compute vrf: %v", err)
	}
	err = serveice.VerifyVrf(&pub, msg, vrf, proof)
	if err != nil {
		t.Fatalf("verify vrf: %v", err)
	}

}

func TestVrf_2(t *testing.T) {
	serveice := baseinterface.NewVrf()
	//公钥不对
	pri, err := crypto.GenerateKey()

	//pub:=pri.PublicKey

	msg := []byte("test")
	vrf, proof, err := serveice.ComputeVrf(pri, msg)
	if err != nil {
		t.Fatalf("compute vrf: %v", err)
	}

	pri_back, _ := crypto.GenerateKey()
	pub_back := pri_back.PublicKey

	err = serveice.VerifyVrf(&pub_back, msg, vrf, proof)
	if err != nil {
		t.Fatalf("verify vrf: %v", err)
	}

}
func TestVrf_3(t *testing.T) {
	serveice := baseinterface.NewVrf()
	//msg不对
	pri, err := crypto.GenerateKey()
	pub := pri.PublicKey

	msg := []byte("test")
	vrf, proof, err := serveice.ComputeVrf(pri, msg)
	if err != nil {
		t.Fatalf("compute vrf: %v", err)
	}

	msg = []byte("asd")
	err = serveice.VerifyVrf(&pub, msg, vrf, proof)
	if err != nil {
		t.Fatalf("verify vrf: %v", err)
	}

}
func TestVrf_4(t *testing.T) {
	serveice := baseinterface.NewVrf()
	//vrf不对
	pri, err := crypto.GenerateKey()
	pub := pri.PublicKey

	msg := []byte("test")
	vrf, proof, err := serveice.ComputeVrf(pri, msg)
	if err != nil {
		t.Fatalf("compute vrf: %v", err)
	}
	vrf = []byte("asd")
	err = serveice.VerifyVrf(&pub, msg, vrf, proof)
	if err != nil {
		t.Fatalf("verify vrf: %v", err)
	}

}
func TestVrf_5(t *testing.T) {
	serveice := baseinterface.NewVrf()
	//proof不对
	pri, err := crypto.GenerateKey()
	pub := pri.PublicKey

	msg := []byte("test")
	vrf, proof, err := serveice.ComputeVrf(pri, msg)
	if err != nil {
		t.Fatalf("compute vrf: %v", err)
	}
	proof = []byte("asd")
	err = serveice.VerifyVrf(&pub, msg, vrf, proof)
	if err != nil {
		t.Fatalf("verify vrf: %v", err)
	}

}

func TestTTT(t *testing.T) {

	tt := "0x02dd0147a1232ea49865c069ac839de414f2ae006167639cb25647411329e58a7d04a4019877373ff38361ac9aa0313feab46fcf9a24a75b32e9b171bc915f6afa92e35d3fc0c6deaf12b54f728032d914131f1c1ede14c59f0364373b30f28cc1695a3e6da7d44432b0481612ec17e8513433cc0137f821de566ce79519318929085fce63618f440de581e350c654390955ada1c3d698420e5d9e4e6097ba0b709c"
	fmt.Println("len", len(tt))
	str := "vrfVAlue1.1.1"
	ansa := []byte(str)
	fmt.Println("str", str, "len str", len(ansa), ansa)

	serveice := baseinterface.NewVrf()
	//正常测试
	pri, err := crypto.GenerateKey()
	pub := pri.PublicKey

	account := keystore.ECDSAPKCompression(&pub)

	msg := []byte("test")
	vrf, proof, err := serveice.ComputeVrf(pri, msg)
	fmt.Println("len vrf", len(vrf), "len proof", len(proof))

	ans := common.GetHeaderVrf(account, vrf, proof)
	fmt.Println("ans", len(ans))

	//TRANS

	ac, vv, vp, _ := common.GetVrfInfoFromHeader(ans)

	fmt.Println("ac", len(ac), ac)
	fmt.Println("vv", len(vv), vv)
	fmt.Println("vp", len(vp), vp)

	if err != nil {
		t.Fatalf("compute vrf: %v", err)
	}
	err = serveice.VerifyVrf(&pub, msg, vrf, proof)
	if err != nil {
		t.Fatalf("verify vrf: %v", err)
	}
}
