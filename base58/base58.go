// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package base58

import (
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/crc8"
	"github.com/pkg/errors"
	"math/big"
	"strings"
)

const (
	// alphabet is the modified base58 alphabet used by matrix.
	alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

	alphabetIdx0 = '1'
)

var b58 = [256]byte{
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 0, 1, 2, 3, 4, 5, 6,
	7, 8, 255, 255, 255, 255, 255, 255,
	255, 9, 10, 11, 12, 13, 14, 15,
	16, 255, 17, 18, 19, 20, 21, 255,
	22, 23, 24, 25, 26, 27, 28, 29,
	30, 31, 32, 255, 255, 255, 255, 255,
	255, 33, 34, 35, 36, 37, 38, 39,
	40, 41, 42, 43, 255, 44, 45, 46,
	47, 48, 49, 50, 51, 52, 53, 54,
	55, 56, 57, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
}

//go:generate go run genalphabet.go

var bigRadix = big.NewInt(58)
var bigZero = big.NewInt(0)

// Decode decodes a modified base58 string to a byte slice.
func Decode(b string) []byte {
	answer := big.NewInt(0)
	j := big.NewInt(1)

	scratch := new(big.Int)
	for i := len(b) - 1; i >= 0; i-- {
		//字符，ascii码表的简版-->得到字符代表的值(0，1,2，..57)
		tmp := b58[b[i]]
		//出现不该出现的字符
		if tmp == 255 {
			return []byte("")
		}

		scratch.SetInt64(int64(tmp))

		//scratch = j*scratch
		scratch.Mul(j, scratch)

		answer.Add(answer, scratch)
		//每次进位都要乘上58
		j.Mul(j, bigRadix)
	}

	//得到大端的字节序
	tmpval := answer.Bytes()

	var numZeros int
	for numZeros = 0; numZeros < len(b); numZeros++ {
		//得到高位0的位数
		if b[numZeros] != alphabetIdx0 {
			break
		}
	}
	//得到原来数字的长度
	flen := numZeros + len(tmpval)

	val := make([]byte, flen, flen)
	copy(val[numZeros:], tmpval)

	return val
}

// Encode encodes a byte slice to a modified base58 string.
func Encode(b []byte) string {
	x := new(big.Int)
	//将b解释为大端存储
	x.SetBytes(b)

	//Base58编码可以表示的比特位数为Log258 {\displaystyle \approx } \approx5.858bit。经过Base58编码的数据为原始的数据长度的1.37倍
	answer := make([]byte, 0, len(b)*136/100)

	for x.Cmp(bigZero) > 0 {
		mod := new(big.Int)
		//x除于58的余数mod，并将商赋值给x
		x.DivMod(x, bigRadix, mod)
		answer = append(answer, alphabet[mod.Int64()])
	}

	// leading zero bytes
	//因为如果高位为0，0除任何数为0，可以直接设置为‘1’
	for _, i := range b {
		if i != 0 {
			break
		}
		answer = append(answer, alphabetIdx0)
	}

	// reverse
	//因为之前先附加低位的，后附加高位的，所以需要翻转
	alen := len(answer)
	for i := 0; i < alen/2; i++ {
		answer[i], answer[alen-1-i] = answer[alen-1-i], answer[i]
	}

	return string(answer)
}

func EncodeInt(data uint8) string {
	if len(alphabet)-1 < int(data) {
		return ""
	}
	return string(alphabet[data])
}

func Base58EncodeToString(currency string, b common.Address) string {
	str := Encode(b[:])
	strAddr := currency + "." + str
	crc := crc8.CalCRC8([]byte(strAddr))
	strCrc := EncodeInt(crc % 58)
	return strAddr + strCrc
}

func Base58DecodeToAddress(strData string) (common.Address, error) {
	strData = strings.TrimSpace(strData)
	if strData == "" {
		return common.Address{}, errors.New("input address invalid")
	}
	if !strings.Contains(strData, ".") {
		return common.Address{}, errors.New("input address invalid")
	}
	currency := strings.Split(strData, ".")[0]
	if !common.IsValidityManCurrency(currency) {
		return common.Address{}, errors.New("input address invalid")
	}

	crc := strData[len(strData)-1]
	crc1 := crc8.CalCRC8([]byte(strData[0 : len(strData)-1]))
	strCrc := EncodeInt(crc1 % 58)
	if strCrc != string(crc) {
		return common.Address{}, errors.New("input address invalid")
	}

	tmpaddres := strings.Split(strData, ".")[1]
	addres := Decode(tmpaddres[0 : len(tmpaddres)-1]) //最后一位为crc%58
	return common.BytesToAddress(addres), nil
}
