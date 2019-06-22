// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package blkgenor

import (
	"github.com/MatrixAINetwork/go-matrix/params/manparams"
	"math/big"
	"math/rand"
	"strconv"
	"testing"

	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"

	"github.com/MatrixAINetwork/go-matrix/common"

	. "github.com/smartystreets/goconvey/convey"
)

const (
	testAddress0 = "0x8605cdbbdb6d264aa742e77020dcbc58fcdce182"
	testAddress1 = "0x8605cdbbdb6d264aa742e77020dcbc58fcdce183"
	testAddress2 = "0x8605cdbbdb6d264aa742e77020dcbc58fcdce184"
	testAddress3 = "0x8605cdbbdb6d264aa742e77020dcbc58fcdce185"
	testAddressM = "0x8605cdbbdb6d264aa742e77020dcbc58fcdce186"
)

func TestPowPool_AddMinerResult(t *testing.T) {
	log.InitLog(3)

	Convey("矿工缓存池测试", t, func() {
		Convey("添加矿工缓存中的数据", func() {
			powPool := NewPowPool("矿工结果池(高度)" + strconv.Itoa(int(0)))
			blockhash := common.Hash{0x01}
			diff := big.NewInt(100)
			from := common.HexToAddress(testAddress)
			Number := rand.Uint64()
			Signatures := []common.Signature{common.BytesToSignature(common.BigToHash(big.NewInt(100)).Bytes()), common.BytesToSignature(common.BigToHash(big.NewInt(101)).Bytes()), common.BytesToSignature(common.BigToHash(big.NewInt(102)).Bytes())}
			minerResult := &mc.HD_MiningRspMsg{from, Number, blockhash, diff, types.EncodeNonce(0), from, common.Hash{0x01}, Signatures}
			err := powPool.AddMinerResult(minerResult.BlockHash, minerResult.Difficulty, minerResult)
			So(err, ShouldBeNil)
		})

		Convey("批量增加矿工缓存中的数据", func() {
			powPool := NewPowPool("矿工结果池(高度)" + strconv.Itoa(int(0)))
			blockhash := common.Hash{0x01}
			diff := big.NewInt(100)
			from := common.HexToAddress(testAddress)
			Number := rand.Uint64()
			Signatures := []common.Signature{common.BytesToSignature(common.BigToHash(big.NewInt(100)).Bytes()), common.BytesToSignature(common.BigToHash(big.NewInt(101)).Bytes()), common.BytesToSignature(common.BigToHash(big.NewInt(102)).Bytes())}
			minerResult := &mc.HD_MiningRspMsg{from, Number, blockhash, diff, types.EncodeNonce(0), from, common.Hash{0x01}, Signatures}
			fromlist := []common.Address{common.HexToAddress(testAddress1), common.HexToAddress(testAddress2), common.HexToAddress(testAddress3)}
			for i := 0; i < 3; i++ {
				tempResult := &mc.HD_MiningRspMsg{fromlist[i], Number, blockhash, diff, types.EncodeNonce(uint64(0)), from, common.Hash{0x01}, Signatures}
				err := powPool.AddMinerResult(minerResult.BlockHash, minerResult.Difficulty, tempResult)
				So(err, ShouldBeNil)
			}
		})

		Convey(" 返回错误超过每个账户存储的最大值 ", func() {
			powPool := NewPowPool("矿工结果池(高度)" + strconv.Itoa(int(0)))
			blockhash := common.Hash{0x01}
			from := common.HexToAddress(testAddress)
			Number := rand.Uint64()
			Signatures := []common.Signature{common.BytesToSignature(common.BigToHash(big.NewInt(100)).Bytes()), common.BytesToSignature(common.BigToHash(big.NewInt(101)).Bytes()), common.BytesToSignature(common.BigToHash(big.NewInt(102)).Bytes())}
			difflist := []*big.Int{big.NewInt(1), big.NewInt(2), big.NewInt(3), big.NewInt(4), big.NewInt(5)}
			for i := 0; i < manparams.VotePoolCountLimit; i++ {
				tempResult := &mc.HD_MiningRspMsg{common.HexToAddress(testAddressM), Number, blockhash, difflist[i], types.EncodeNonce(uint64(0)), from, common.Hash{0x01}, Signatures}
				err := powPool.AddMinerResult(tempResult.BlockHash, tempResult.Difficulty, tempResult)
				So(err, ShouldBeNil)
			}
			errDifflist := big.NewInt(5)
			tempResult := &mc.HD_MiningRspMsg{common.HexToAddress(testAddressM), Number, blockhash, errDifflist, types.EncodeNonce(uint64(0)), from, common.Hash{0x01}, Signatures}
			err := powPool.AddMinerResult(tempResult.BlockHash, tempResult.Difficulty, tempResult)
			So(err, ShouldBeError)
		})
	})

}
func TestPowPool_AddWrongMinerResult(t *testing.T) {
	log.InitLog(3)
	Convey("矿工缓存池测试", t, func() {
		Convey(" 输入参数异常测试，blockhash为初值", func() {

			blockhash := common.Hash{}
			diff := big.NewInt(0)
			from := common.HexToAddress(testAddress)
			Number := rand.Uint64()
			Signatures := []common.Signature{common.BytesToSignature(common.BigToHash(big.NewInt(100)).Bytes()), common.BytesToSignature(common.BigToHash(big.NewInt(101)).Bytes()), common.BytesToSignature(common.BigToHash(big.NewInt(102)).Bytes())}
			minerResult := &mc.HD_MiningRspMsg{from, Number, blockhash, diff, types.EncodeNonce(0), from, common.Hash{0x01}, Signatures}

			abnormalPowPool := NewPowPool("矿工结果池(高度)" + strconv.Itoa(int(0)))
			err := abnormalPowPool.AddMinerResult(minerResult.BlockHash, minerResult.Difficulty, minerResult)
			Println("error=", err)
			So(err, ShouldBeError)
		})
		Convey(" 输入参数异常测试，难度值为NULL", func() {

			blockhash := common.Hash{100}
			diff := big.NewInt(0)
			from := common.HexToAddress(testAddress)
			Number := rand.Uint64()
			Signatures := []common.Signature{common.BytesToSignature(common.BigToHash(big.NewInt(100)).Bytes()), common.BytesToSignature(common.BigToHash(big.NewInt(101)).Bytes()), common.BytesToSignature(common.BigToHash(big.NewInt(102)).Bytes())}
			minerResult := &mc.HD_MiningRspMsg{from, Number, blockhash, diff, types.EncodeNonce(0), from, common.Hash{0x01}, Signatures}

			abnormalPowPool := NewPowPool("矿工结果池(高度)" + strconv.Itoa(int(0)))
			err := abnormalPowPool.AddMinerResult(minerResult.BlockHash, nil, minerResult)
			Println("error=", err)
			So(err, ShouldBeError)
		})

		Convey(" 输入参数异常测试，难度值为0", func() {
			blockhash := common.Hash{0x01}
			diff := big.NewInt(0)
			from := common.HexToAddress(testAddress)
			Number := rand.Uint64()
			Signatures := []common.Signature{common.BytesToSignature(common.BigToHash(big.NewInt(100)).Bytes()), common.BytesToSignature(common.BigToHash(big.NewInt(101)).Bytes()), common.BytesToSignature(common.BigToHash(big.NewInt(102)).Bytes())}
			minerResult := &mc.HD_MiningRspMsg{from, Number, blockhash, diff, types.EncodeNonce(0), from, common.Hash{0x01}, Signatures}

			abnormalPowPool := NewPowPool("矿工结果池(高度)" + strconv.Itoa(int(0)))
			err := abnormalPowPool.AddMinerResult(minerResult.BlockHash, minerResult.Difficulty, minerResult)
			Println("error=", err)
			So(err, ShouldBeError)
		})
		Convey(" 输入参数异常测试，难度值为负值", func() {
			blockhash := common.Hash{0x01}
			diff := big.NewInt(-100)
			from := common.HexToAddress(testAddress)
			Number := rand.Uint64()
			Signatures := []common.Signature{common.BytesToSignature(common.BigToHash(big.NewInt(100)).Bytes()), common.BytesToSignature(common.BigToHash(big.NewInt(101)).Bytes()), common.BytesToSignature(common.BigToHash(big.NewInt(102)).Bytes())}
			minerResult := &mc.HD_MiningRspMsg{from, Number, blockhash, diff, types.EncodeNonce(0), from, common.Hash{0x01}, Signatures}

			abnormalPowPool := NewPowPool("矿工结果池(高度)" + strconv.Itoa(int(0)))
			err := abnormalPowPool.AddMinerResult(minerResult.BlockHash, minerResult.Difficulty, minerResult)
			Println("error=", err)
			So(err, ShouldBeError)
		})

		Convey(" 输入参数异常测试，挖矿结果为空", func() {
			blockhash := common.Hash{0x01}
			diff := big.NewInt(100)
			from := common.HexToAddress(testAddress)
			Number := rand.Uint64()
			Signatures := []common.Signature{common.BytesToSignature(common.BigToHash(big.NewInt(100)).Bytes()), common.BytesToSignature(common.BigToHash(big.NewInt(101)).Bytes()), common.BytesToSignature(common.BigToHash(big.NewInt(102)).Bytes())}
			minerResult := &mc.HD_MiningRspMsg{from, Number, blockhash, diff, types.EncodeNonce(0), from, common.Hash{0x01}, Signatures}

			abnormalPowPool := NewPowPool("矿工结果池(高度)" + strconv.Itoa(int(0)))
			err := abnormalPowPool.AddMinerResult(minerResult.BlockHash, minerResult.Difficulty, nil)
			Println("error=", err)
			So(err, ShouldBeError)
		})
	})
}

func TestPowPool_GetMinerResult(t *testing.T) {
	log.InitLog(3)
	powPool := NewPowPool("矿工结果池(高度)" + strconv.Itoa(int(0)))
	blockhash := common.Hash{0x01}
	diff := big.NewInt(10)
	from := common.HexToAddress(testAddress)
	Number := rand.Uint64()
	Signatures := []common.Signature{common.BytesToSignature(common.BigToHash(big.NewInt(100)).Bytes()), common.BytesToSignature(common.BigToHash(big.NewInt(101)).Bytes()), common.BytesToSignature(common.BigToHash(big.NewInt(102)).Bytes())}
	minerResult := &mc.HD_MiningRspMsg{from, Number, blockhash, diff, types.EncodeNonce(0), from, common.Hash{0x01}, Signatures}

	Convey("获取矿工缓存池数据测试", t, func() {
		Convey(" 获取的矿工换成含有之前添加的数据", func() {
			err := powPool.AddMinerResult(minerResult.BlockHash, minerResult.Difficulty, minerResult)
			So(err, ShouldBeNil)
			fromlist := []common.Address{common.HexToAddress(testAddress1), common.HexToAddress(testAddress2), common.HexToAddress(testAddress3)}
			for i := 0; i < 3; i++ {
				tempResult := &mc.HD_MiningRspMsg{fromlist[i], Number, blockhash, diff, types.EncodeNonce(uint64(0)), from, common.Hash{0x01}, Signatures}
				err := powPool.AddMinerResult(minerResult.BlockHash, minerResult.Difficulty, tempResult)
				So(err, ShouldBeNil)
			}
			retMinerResults, err := powPool.GetMinerResults(minerResult.BlockHash, minerResult.Difficulty)
			So(err, ShouldBeNil)
			So(retMinerResults, ShouldContain, minerResult)
			So(len(retMinerResults), ShouldEqual, 4)
		})

		Convey(" 获取的不存在blockhash的矿工数据", func() {

			_, err := powPool.GetMinerResults(common.Hash{0x10}, minerResult.Difficulty)
			Println("error=", err)
			So(err, ShouldBeError)
		})

		Convey(" 获取的不存在 难度值的矿工数据", func() {

			_, err := powPool.GetMinerResults(minerResult.BlockHash, big.NewInt(100))
			Println("error=", err)
			So(err, ShouldBeError)
		})
		Convey(" 获取的难度值为nil的数据", func() {

			_, err := powPool.GetMinerResults(minerResult.BlockHash, nil)
			Println("error=", err)
			So(err, ShouldBeError)
		})
		Convey(" 获取的难度值为0的数据", func() {

			_, err := powPool.GetMinerResults(minerResult.BlockHash, big.NewInt(0))
			Println("error=", err)
			So(err, ShouldBeError)
		})
		Convey(" 获取的难度值为负值的数据", func() {

			_, err := powPool.GetMinerResults(minerResult.BlockHash, big.NewInt(-1))
			Println("error=", err)
			So(err, ShouldBeError)
		})
	})

}

func TestPowPool_DelMinerResult(t *testing.T) {
	log.InitLog(3)
	powPool := NewPowPool("矿工结果池(高度)" + strconv.Itoa(int(0)))
	blockhash := common.Hash{0x01}
	diff := big.NewInt(10)
	from := common.HexToAddress(testAddress)
	Number := rand.Uint64()
	Signatures := []common.Signature{common.BytesToSignature(common.BigToHash(big.NewInt(100)).Bytes()), common.BytesToSignature(common.BigToHash(big.NewInt(101)).Bytes()), common.BytesToSignature(common.BigToHash(big.NewInt(102)).Bytes())}
	minerResult := &mc.HD_MiningRspMsg{from, Number, blockhash, diff, types.EncodeNonce(0), from, common.Hash{0x01}, Signatures}

	Convey("删除矿工缓存池数据测试", t, func() {
		Convey(" 删除之前添加的数据", func() {
			err := powPool.AddMinerResult(minerResult.BlockHash, minerResult.Difficulty, minerResult)
			So(err, ShouldBeNil)
			fromlist := []common.Address{common.HexToAddress(testAddress1), common.HexToAddress(testAddress2), common.HexToAddress(testAddress3)}
			for i := 0; i < 3; i++ {
				tempResult := &mc.HD_MiningRspMsg{fromlist[i], Number, blockhash, diff, types.EncodeNonce(uint64(0)), from, common.Hash{0x01}, Signatures}
				err := powPool.AddMinerResult(minerResult.BlockHash, minerResult.Difficulty, tempResult)
				So(err, ShouldBeNil)
			}
			retMinerResults, err := powPool.GetMinerResults(minerResult.BlockHash, minerResult.Difficulty)
			So(err, ShouldBeNil)
			So(retMinerResults, ShouldContain, minerResult)
			So(len(retMinerResults), ShouldEqual, 4)
			powPool.DelOneResult(minerResult.BlockHash, minerResult.Difficulty, minerResult.From)

			retMinerResults2, err2 := powPool.GetMinerResults(minerResult.BlockHash, minerResult.Difficulty)
			So(err2, ShouldBeNil)
			So(retMinerResults2, ShouldNotContain, minerResult)
			So(len(retMinerResults2), ShouldEqual, 3)
		})
		Convey(" 删除之前没有添加的数据", func() {
			blockhash := common.Hash{0x30}
			err := powPool.DelOneResult(blockhash, minerResult.Difficulty, minerResult.From)
			Println("error=", err)
			So(err, ShouldBeError)
			retMinerResults2, err2 := powPool.GetMinerResults(minerResult.BlockHash, minerResult.Difficulty)
			So(err2, ShouldBeNil)
			So(len(retMinerResults2), ShouldEqual, 3)
		})
		Convey(" 删除难度值为nil的数据", func() {
			blockhash := common.Hash{0x30}
			err := powPool.DelOneResult(blockhash, nil, minerResult.From)
			Println("error=", err)
			So(err, ShouldBeError)

		})
		Convey(" 删除难度值为0的数据", func() {
			blockhash := common.Hash{0x30}
			err := powPool.DelOneResult(blockhash, big.NewInt(0), minerResult.From)
			Println("error=", err)
			So(err, ShouldBeError)

		})
		Convey(" 删除难度值负数的数据", func() {
			blockhash := common.Hash{0x30}
			err := powPool.DelOneResult(blockhash, big.NewInt(-3), minerResult.From)
			Println("error=", err)
			So(err, ShouldBeError)

		})
		Convey(" 空hash", func() {
			blockhash := common.Hash{}
			err := powPool.DelOneResult(blockhash, minerResult.Difficulty, minerResult.From)
			Println("error=", err)
			So(err, ShouldBeError)

		})
	})
}
