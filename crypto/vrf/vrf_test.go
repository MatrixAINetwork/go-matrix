// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package vrf

import (
	"crypto/ecdsa"
	"fmt"
	"testing"

	"github.com/matrix/go-matrix/baseinterface"
	"github.com/matrix/go-matrix/crypto"
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

	msg := []byte("test")
	vrf, proof, err := serveice.ComputeVrf(pri, msg)
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
