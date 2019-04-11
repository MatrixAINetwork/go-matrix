// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package mtxdpos

import (
	"bou.ke/monkey"
	"crypto/ecdsa"
	"github.com/MatrixAINetwork/go-matrix/ca"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/crypto"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"math/big"
	"testing"
)

/*func TestDPOS_01(t *testing.T) {
	validators, keys := generateTestValidators(3)
	ca.SetTestValidatorStocks(validators)

	hash1 := types.RlpHash(validators)
	hash2 := types.RlpHash(keys)

	signs := make([]common.Signature, 0)
	for _, key := range keys {
		sign, err := crypto.SignWithValidate(hash1.Bytes(), true, key)
		if err != nil {
			t.Fatalf("sign err(%s)", err)
		}
		signs = append(signs, common.BytesToSignature(sign))
	}

	dpos := NewMtxDPOS(nil)
	rightSigns, err := dpos.VerifyHashWithNumber(hash1, signs, 0)
	if err != nil {
		t.Fatalf("dpos err(%s)!", err)
	}
	t.Logf("right Signs Count %d", len(rightSigns))

	_, err = dpos.VerifyHashWithNumber(hash2, signs, 0)
	t.Logf("dpos return: %s", err)
	if err == nil {
		t.Fatalf("err")
	}
}

func TestDPOS_02(t *testing.T) {
	validators, keys := generateTestValidators(2)
	ca.SetTestValidatorStocks(validators)

	hash := types.RlpHash(validators)

	signs := make([]common.Signature, 0)
	for _, key := range keys {
		sign, err := crypto.SignWithValidate(hash.Bytes(), true, key)
		if err != nil {
			t.Fatalf("sign err(%s)", err)
		}
		signs = append(signs, common.BytesToSignature(sign))
	}

	dpos := NewMtxDPOS(nil)
	_, err := dpos.VerifyHashWithNumber(hash, signs, 0)
	t.Logf("dpos return: %s", err)
	if err == nil {
		t.Fatalf("dpos err: 2 validator but pass dpos!")
	}
}

func TestDPOS_03(t *testing.T) {
	validators, keys := generateTestValidators(3)
	ca.SetTestValidatorStocks(validators)

	hash := types.RlpHash(validators)

	signs := make([]common.Signature, 0)
	i := 0
	for _, key := range keys {
		valid := true
		if i == 1 {
			valid = false
		}
		sign, err := crypto.SignWithValidate(hash.Bytes(), valid, key)
		if err != nil {
			t.Fatalf("sign err(%s)", err)
		}
		signs = append(signs, common.BytesToSignature(sign))
		i++
	}

	dpos := NewMtxDPOS(nil)
	_, err := dpos.VerifyHashWithNumber(hash, signs, 0)
	t.Logf("dpos return: %s", err)
	if err == nil {
		t.Fatalf("dpos err: 2 validator but pass dpos!")
	}
}

func TestDPOS_04(t *testing.T) {
	validators, keys := generateTestValidators(7)
	guard := monkey.Patch(ca.GetTopologyByNumber, func(reqTypes common.RoleType, number uint64) (*mc.TopologyGraph, error) {
		graph := &mc.TopologyGraph{
			Number:   big.NewInt(int64(number)),
			NodeList: validators,
		}
		return graph, nil
	})
	defer guard.Unpatch()

	hash := types.RlpHash(validators)

	signs := make([]common.Signature, 0)
	i := 0
	for _, key := range keys {
		valid := true
		if i == 6 {
			break
		}
		sign, err := crypto.SignWithValidate(hash.Bytes(), valid, key)
		if err != nil {
			t.Fatalf("sign err(%s)", err)
		}
		signs = append(signs, common.BytesToSignature(sign))
		i++
	}

	dpos := NewMtxDPOS(nil)
	_, err := dpos.VerifyHashWithNumber(hash, signs, 0)
	t.Logf("dpos return: %s", err)
	if err == nil {
		t.Fatalf("dpos err: 2 validator but pass dpos!")
	}
}

func TestDPOS_05(t *testing.T) {
	validators, keys := generateTestValidators(7)
	ca.SetTestValidatorStocks(validators)

	hash := types.RlpHash(validators)

	signs := make([]common.Signature, 0)
	i := 0
	for _, key := range keys {
		valid := true
		if i == 6 {
			signs = append(signs, signs[3])
			break
		}
		sign, err := crypto.SignWithValidate(hash.Bytes(), valid, key)
		if err != nil {
			t.Fatalf("sign err(%s)", err)
		}
		signs = append(signs, common.BytesToSignature(sign))
		i++
	}

	dpos := NewMtxDPOS(nil)
	_, err := dpos.VerifyHashWithNumber(hash, signs, 0)
	t.Logf("dpos return: %s", err)
	if err == nil {
		t.Fatalf("dpos err: 2 validator but pass dpos!")
	}
}*/

func TestDPOS_06(t *testing.T) {
	validators, keys := generateTestValidators(11)
	guard := monkey.Patch(ca.GetTopologyByNumber, func(reqTypes common.RoleType, number uint64) (*mc.TopologyGraph, error) {
		graph := &mc.TopologyGraph{
			Number:   big.NewInt(int64(number)),
			NodeList: validators,
		}
		return graph, nil
	})
	defer guard.Unpatch()

	hash := types.RlpHash(validators)

	signs := make([]common.Signature, 0)
	i := 0
	for _, key := range keys {
		valid := true
		if i == 8 {
			break
		}
		sign, err := crypto.SignWithValidate(hash.Bytes(), valid, key)
		if err != nil {
			t.Fatalf("sign err(%s)", err)
		}
		signs = append(signs, common.BytesToSignature(sign))
		i++
	}

	dpos := NewMtxDPOS(nil)
	_, err := dpos.VerifyHashWithNumber(hash, signs, 0)
	t.Logf("dpos return: %v", err)
	if err == nil {
		t.Fatalf("dpos err: 2 validator but pass dpos!")
	}
}

func TestDPOS_07(t *testing.T) {
	validators, keys := generateTestValidators(11)
	hash := types.RlpHash(validators)
	signs := make([]common.Signature, 0)
	i := 0
	for addr, key := range keys {
		valid := true
		i++
		if i == 2 {
			for v := 0; v < len(validators); v++ {
				if validators[v].Account == addr {
					validators[v].Stock = 10
					break
				}
			}
		}
		if i == 8 || i == 2 || i == 5 || i == 6 {
			continue
		}
		sign, err := crypto.SignWithValidate(hash.Bytes(), valid, key)
		if err != nil {
			t.Fatalf("sign err(%s)", err)
		}
		signs = append(signs, common.BytesToSignature(sign))
	}

	guard := monkey.Patch(ca.GetTopologyByNumber, func(reqTypes common.RoleType, number uint64) (*mc.TopologyGraph, error) {
		graph := &mc.TopologyGraph{
			Number:   big.NewInt(int64(number)),
			NodeList: validators,
		}
		return graph, nil
	})
	defer guard.Unpatch()

	dpos := NewMtxDPOS(nil)
	rightSigns, err := dpos.VerifyHashWithNumber(hash, signs, 0)
	if err != nil {
		t.Fatalf("dpos err: right signs but didn`t pass dpos, %v", err)
	}

	t.Logf("input signs count=%d,right signs count=%d", len(signs), len(rightSigns))
}

/*func TestDPOS_08(t *testing.T) {
	validators, keys := generateTestValidators(11)
	hash := types.RlpHash(validators)
	signs := make([]common.Signature, 0)
	i := 0
	for addr, key := range keys {
		valid := true
		if i == 0 {
			for _, v := range validators {
				if v.Account == addr {
					v.Stock = 20
					break
				}
			}
			valid = false
		}

		sign, err := crypto.SignWithValidate(hash.Bytes(), valid, key)
		if err != nil {
			t.Fatalf("sign err(%s)", err)
		}
		signs = append(signs, common.BytesToSignature(sign))
		i++
	}

	ca.SetTestValidatorStocks(validators)

	dpos := NewMtxDPOS(nil)
	_, err := dpos.VerifyHashWithNumber(hash, signs, 0)
	t.Logf("dpos return: %s", err)
	if err == nil {
		t.Fatalf("dpos err: 2 validator but pass dpos!")
	}
}

func TestDPOS_09(t *testing.T) {
	validators, keys := generateTestValidators(11)
	hash := types.RlpHash(validators)
	signs := make([]common.Signature, 0)
	i := 0
	for addr, key := range keys {
		valid := true
		if i == 0 {
			for _, v := range validators {
				if v.Account == addr {
					v.Stock = 20
					break
				}
			}
			valid = false
		}

		sign, err := crypto.SignWithValidate(hash.Bytes(), valid, key)
		if err != nil {
			t.Fatalf("sign err(%s)", err)
		}
		signs = append(signs, common.BytesToSignature(sign))
		i++
	}

	ca.SetTestValidatorStocks(validators)

	dpos := NewMtxDPOS(nil)
	_, err := dpos.VerifyHashWithNumber(hash, signs, 0)
	t.Logf("dpos return: %s", err)
	if err == nil {
		t.Fatalf("dpos err: 2 validator but pass dpos!")
	}
}*/

func generateTestValidators(count int) ([]mc.TopologyNodeInfo, map[common.Address]*ecdsa.PrivateKey) {
	validators := make([]mc.TopologyNodeInfo, 0)
	keys := make(map[common.Address]*ecdsa.PrivateKey)

	for len(validators) < count {
		key, err := crypto.GenerateKey()
		if err != nil {
			continue
		}

		info := mc.TopologyNodeInfo{
			Account:  crypto.PubkeyToAddress(key.PublicKey),
			Position: 0,
			Type:     common.RoleValidator,
			Stock:    3,
		}
		keys[info.Account] = key
		validators = append(validators, info)
	}

	return validators, keys
}
