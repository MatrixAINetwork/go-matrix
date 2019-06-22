// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package lessdisk

import (
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/params"
	"github.com/MatrixAINetwork/go-matrix/rlp"
	"github.com/pkg/errors"
	"math/big"
	"math/rand"
	"testing"
	"time"
)

var cfg = &params.LessDiskConfig{
	OptInterval:     10,
	HeightThreshold: 10,
	TimeThreshold:   10,
}

func sendBlkInsertMsg(hash common.Hash, number uint64, insertTime uint64) *mc.BlockInsertedMsg {
	msg := &mc.BlockInsertedMsg{
		Block: mc.BlockInfo{
			Hash:   hash,
			Number: number,
		},
		InsertTime: insertTime,
		CanonState: true,
	}

	mc.PublishEvent(mc.BlockInserted, msg)
	time.Sleep(100 * time.Millisecond)
	return msg
}

func Test_IndexUpdate_normal(t *testing.T) {
	db := newSimDB()
	svr := NewLessDiskSvr(cfg, db, &simChain1{10})
	t.Logf("创建服务成功: %s", svr.logInfo)

	hash := common.HexToHash("0x000031")
	number := uint64(31)
	insertTime := uint64(time.Now().Unix())
	sendBlkInsertMsg(hash, number, insertTime)

	checkData := make(map[uint64][]dbBlkIndex, 0)
	checkData[number] = []dbBlkIndex{
		{
			Hash:       hash,
			InsertTime: insertTime,
		},
	}
	if err := db.checkState(2, number, checkData); err != nil {
		t.Fatal(err)
	}
}

func Test_IndexUpdate_SameBlock_DiffTime(t *testing.T) {
	db := newSimDB()
	svr := NewLessDiskSvr(cfg, db, &simChain1{10})
	t.Logf("创建服务成功: %s", svr.logInfo)

	myRand := rand.New(rand.NewSource(time.Now().Unix()))
	hash := common.HexToHash("0x000031")
	number := uint64(31)
	insertTime1 := uint64(myRand.Intn(50000))
	insertTime2 := uint64(myRand.Intn(50000))
	sendBlkInsertMsg(hash, number, insertTime1)
	sendBlkInsertMsg(hash, number, insertTime2)

	checkData := make(map[uint64][]dbBlkIndex, 0)
	checkData[number] = []dbBlkIndex{
		{
			Hash:       hash,
			InsertTime: insertTime2,
		},
	}
	if err := db.checkState(2, number, checkData); err != nil {
		t.Fatal(err)
	}
}

func Test_IndexUpdate_SameHeight_RandBlock(t *testing.T) {
	db := newSimDB()
	svr := NewLessDiskSvr(cfg, db, &simChain1{10})
	t.Logf("创建服务成功: %s", svr.logInfo)

	myRand := rand.New(rand.NewSource(time.Now().Unix()))
	hashList := []common.Hash{common.HexToHash("0x000001"), common.HexToHash("0x000002"), common.HexToHash("0x000003"), common.HexToHash("0x000004")}
	number := uint64(10)
	insertBlkMap := make(map[common.Hash]dbBlkIndex)
	for i := 0; i < 10; i++ {
		hash := hashList[myRand.Intn(len(hashList))]
		insertTime := uint64(myRand.Intn(50000))
		sendBlkInsertMsg(hash, number, insertTime)
		insertBlkMap[hash] = dbBlkIndex{
			Hash:       hash,
			InsertTime: insertTime,
		}
		t.Logf("发送区块插入消息: blk[%s] number[%d] insertTime[%d]", hash.Hex(), number, insertTime)
	}

	checkData := make(map[uint64][]dbBlkIndex, 0)
	numberIndex := make([]dbBlkIndex, 0)
	for _, item := range insertBlkMap {
		numberIndex = append(numberIndex, dbBlkIndex{Hash: item.Hash, InsertTime: item.InsertTime})
	}
	checkData[number] = numberIndex
	if err := db.checkState(2, number, checkData); err != nil {
		t.Fatal(err)
	}
	t.Log("测试通过")
}

func Test_IndexUpdate_DiffHeight_Order(t *testing.T) {
	db := newSimDB()
	svr := NewLessDiskSvr(cfg, db, &simChain1{10})
	t.Logf("创建服务成功: %s", svr.logInfo)

	myRand := rand.New(rand.NewSource(time.Now().Unix()))
	numberList := []uint64{3, 3, 6, 7, 7}
	insertBlkMap := make(map[common.Hash]*mc.BlockInsertedMsg)
	minNumber := numberList[0]
	for i := 0; i < len(numberList); i++ {
		hash := common.BigToHash(big.NewInt(int64(100000000 + i)))
		insertTime := uint64(myRand.Intn(50000))
		msg := sendBlkInsertMsg(hash, numberList[i], insertTime)
		insertBlkMap[hash] = msg
		if minNumber > numberList[i] {
			minNumber = numberList[i]
		}
		t.Logf("发送区块插入消息: blk[%s] number[%d] insertTime[%d]", hash.Hex(), numberList[i], insertTime)
	}

	checkData := make(map[uint64][]dbBlkIndex, 0)
	for hash, blk := range insertBlkMap {
		checkData[blk.Block.Number] = append(checkData[blk.Block.Number], dbBlkIndex{Hash: hash, InsertTime: blk.InsertTime})
	}
	if err := db.checkState(len(checkData)+1, minNumber, checkData); err != nil {
		t.Fatal(err)
	}
	t.Log("测试通过")
}

func Test_IndexUpdate_DiffHeight_Rand(t *testing.T) {
	db := newSimDB()
	svr := NewLessDiskSvr(cfg, db, &simChain1{10})
	t.Logf("创建服务成功: %s", svr.logInfo)

	myRand := rand.New(rand.NewSource(time.Now().Unix()))
	min, max := uint64(2), uint64(7)
	insertBlkMap := make(map[common.Hash]*mc.BlockInsertedMsg)
	minNumber := max
	for i := 0; i < 10; i++ {
		hash := common.BigToHash(big.NewInt(int64(100000000 + i)))
		number := uint64(myRand.Int63n(int64(max-min)+1)) + min
		insertTime := uint64(myRand.Intn(50000))
		msg := sendBlkInsertMsg(hash, number, insertTime)
		insertBlkMap[hash] = msg
		if minNumber > number {
			minNumber = number
		}
		t.Logf("发送区块插入消息: blk[%s] number[%d] insertTime[%d]", hash.Hex(), number, insertTime)
	}

	checkData := make(map[uint64][]dbBlkIndex, 0)
	for hash, blk := range insertBlkMap {
		checkData[blk.Block.Number] = append(checkData[blk.Block.Number], dbBlkIndex{Hash: hash, InsertTime: blk.InsertTime})
	}
	if err := db.checkState(len(checkData)+1, minNumber, checkData); err != nil {
		t.Fatal(err)
	}
	t.Log("测试通过")
}

func Test_DelBlk_HeightThreshold(t *testing.T) {
	log.InitLog(5)
	db := newSimDB()
	curTime := time.Now().Unix()
	blkIndex := make(map[uint64][]dbBlkIndex, 0)
	for i := 5; i <= 10; i++ {
		for j := 0; j < 3; j++ {
			blkIndex[uint64(i)] = append(blkIndex[uint64(i)], dbBlkIndex{
				Hash:       common.BigToHash(big.NewInt(int64(100000000*i + j))),
				InsertTime: uint64(curTime),
			})
		}
	}
	if err := db.initDB(5, blkIndex); err != nil {
		t.Fatalf("db数据初始化失败:%v", err)
	}
	svr := NewLessDiskSvr(cfg, db, &simChain1{10})
	svr.FuncSwitch(true)

	checkTimer := time.NewTimer(5 * time.Second)

	for i := 0; i < 10; i++ {
		select {
		case <-checkTimer.C:
			log.Info("test检查状态", "次数", i, "状态", "开始")
			var err error
			switch i {
			case 0, 1, 2, 3:
				err = db.checkState(7, 5, blkIndex)
			case 4:
				delete(blkIndex, 5)
				err = db.checkState(6, 6, blkIndex)
			case 5:
				delete(blkIndex, 6)
				delete(blkIndex, 7)
				err = db.checkState(4, 8, blkIndex)
			case 6:
				delete(blkIndex, 8)
				delete(blkIndex, 9)
				err = db.checkState(2, 10, blkIndex)
			case 7:
				delete(blkIndex, 10)
				err = db.checkState(1, 12, blkIndex)
			case 8:
				err = db.checkState(1, 14, blkIndex)
			case 98:
				err = db.checkState(1, 16, blkIndex)
			}

			if err != nil {
				t.Fatalf("第%d次检查数据异常:%v", i, err)
			}
			log.Info("test检查状态", "次数", i, "状态", "完成")
			checkTimer = time.NewTimer(10 * time.Second)
		}
	}

}

func Test_DelBlk_TimeThreshold(t *testing.T) {
	log.InitLog(5)
	db := newSimDB()
	curTime := time.Now().Unix()
	blkIndex := make(map[uint64][]dbBlkIndex, 0)
	for i := 5; i <= 10; i++ {
		for j := 0; j < 3; j++ {
			data := dbBlkIndex{
				Hash:       common.BigToHash(big.NewInt(int64(100000000*i + j))),
				InsertTime: uint64(curTime),
			}
			if i == 7 && j == 1 {
				data.InsertTime += 45
			}
			blkIndex[uint64(i)] = append(blkIndex[uint64(i)], data)
		}
	}
	if err := db.initDB(5, blkIndex); err != nil {
		t.Fatalf("db数据初始化失败:%v", err)
	}
	svr := NewLessDiskSvr(cfg, db, &simChain1{16})
	svr.FuncSwitch(true)

	checkTimer := time.NewTimer(5 * time.Second)

	for i := 0; i < 9; i++ {
		select {
		case <-checkTimer.C:
			log.Info("test检查状态", "次数", i, "状态", "开始")
			var err error
			switch i {
			case 0:
				err = db.checkState(7, 5, blkIndex)
			case 1:
				delete(blkIndex, 5)
				err = db.checkState(6, 6, blkIndex)
			case 2:
				delete(blkIndex, 6)
				err = db.checkState(5, 7, blkIndex)
			case 3, 4, 5:
				err = db.checkState(5, 7, blkIndex)
			case 6:
				delete(blkIndex, 7)
				delete(blkIndex, 8)
				delete(blkIndex, 9)
				delete(blkIndex, 10)
				err = db.checkState(1, 16, blkIndex)
			case 7:
				err = db.checkState(1, 18, blkIndex)
			case 8:
				err = db.checkState(1, 20, blkIndex)
			}

			if err != nil {
				t.Fatalf("第%d次检查数据异常:%v", i, err)
			}
			log.Info("test检查状态", "次数", i, "状态", "完成")
			checkTimer = time.NewTimer(10 * time.Second)
		}
	}

}

func Test_DelBlk_SomeBlkDelFail(t *testing.T) {
	log.InitLog(5)
	chain := &simChain2{}
	chain.curNumber = 16
	chain.delFailBlk = make(map[common.Hash]int)
	db := newSimDB()
	curTime := time.Now().Unix()
	blkIndex := make(map[uint64][]dbBlkIndex, 0)
	for i := 5; i <= 10; i++ {
		for j := 0; j < 3; j++ {
			data := dbBlkIndex{
				Hash:       common.BigToHash(big.NewInt(int64(100000000*i + j))),
				InsertTime: uint64(curTime),
			}
			if i == 7 && j == 1 ||
				i == 8 && j == 0 ||
				i == 8 && j == 1 {
				chain.delFailBlk[data.Hash] = 1
			}
			blkIndex[uint64(i)] = append(blkIndex[uint64(i)], data)
		}
	}
	if err := db.initDB(5, blkIndex); err != nil {
		t.Fatalf("db数据初始化失败:%v", err)
	}
	svr := NewLessDiskSvr(cfg, db, chain)
	svr.FuncSwitch(true)

	checkTimer := time.NewTimer(5 * time.Second)

	for i := 0; i < 6; i++ {
		select {
		case <-checkTimer.C:
			log.Info("test检查状态", "次数", i, "状态", "开始")
			var err error
			switch i {
			case 0:
				err = db.checkState(7, 5, blkIndex)
			case 1:
				delete(blkIndex, 5)
				err = db.checkState(6, 6, blkIndex)
			case 2:
				delete(blkIndex, 6)
				err = db.checkState(5, 7, blkIndex)
			case 3:
				delete(blkIndex, 7)
				err = db.checkState(4, 8, blkIndex)
			case 4:
				delete(blkIndex, 8)
				delete(blkIndex, 9)
				delete(blkIndex, 10)
				err = db.checkState(1, 12, blkIndex)
			case 5:
				err = db.checkState(1, 14, blkIndex)
			}

			if err != nil {
				t.Fatalf("第%d次检查数据异常:%v", i, err)
			}
			log.Info("test检查状态", "次数", i, "状态", "完成")
			checkTimer = time.NewTimer(10 * time.Second)
		}
	}

}

func Test_DelBlk_FunSwitch(t *testing.T) {
	log.InitLog(5)
	chain := &simChain3{16}
	db := newSimDB()
	curTime := time.Now().Unix()
	blkIndex := make(map[uint64][]dbBlkIndex, 0)
	for i := 5; i <= 10; i++ {
		for j := 0; j < 3; j++ {
			data := dbBlkIndex{
				Hash:       common.BigToHash(big.NewInt(int64(100000000*i + j))),
				InsertTime: uint64(curTime),
			}
			blkIndex[uint64(i)] = append(blkIndex[uint64(i)], data)
		}
	}
	if err := db.initDB(5, blkIndex); err != nil {
		t.Fatalf("db数据初始化失败:%v", err)
	}
	svr := NewLessDiskSvr(cfg, db, chain)
	svr.FuncSwitch(true)

	checkTimer := time.NewTimer(5 * time.Second)

	for i := 0; i < 7; i++ {
		select {
		case <-checkTimer.C:
			log.Info("test检查状态", "次数", i, "状态", "开始")
			chain.curNumber = 16 + uint64(i*2)
			var err error
			switch i {
			case 0:
				err = db.checkState(7, 5, blkIndex)
			case 1:
				svr.FuncSwitch(false)
				delete(blkIndex, 5)
				err = db.checkState(6, 6, blkIndex)
			case 2:
				err = db.checkState(6, 6, blkIndex)
			case 3:
				svr.FuncSwitch(true)
				err = db.checkState(6, 6, blkIndex)
			case 4:
				delete(blkIndex, 6)
				delete(blkIndex, 7)
				delete(blkIndex, 8)
				delete(blkIndex, 9)
				delete(blkIndex, 10)
				err = db.checkState(1, 12, blkIndex)
			case 5, 6:
				err = db.checkState(1, chain.curNumber-cfg.HeightThreshold-2, blkIndex)
			}

			if err != nil {
				t.Fatalf("第%d次检查数据异常:%v", i, err)
			}
			log.Info("test检查状态", "次数", i, "状态", "完成")
			checkTimer = time.NewTimer(10 * time.Second)
		}
	}

}

type simDB struct {
	cache map[string][]byte
}

func newSimDB() *simDB {
	return &simDB{
		cache: make(map[string][]byte, 0),
	}
}

func (db *simDB) Has(key []byte) (bool, error) {
	keyStr := string(key)
	_, ok := db.cache[keyStr]
	return ok, nil
}

func (db *simDB) Get(key []byte) ([]byte, error) {
	keyStr := string(key)
	data, ok := db.cache[keyStr]
	if ok == false {
		return nil, errors.New("数据不存在")
	}
	return data, nil
}

func (db *simDB) Put(key []byte, value []byte) error {
	keyStr := string(key)
	db.cache[keyStr] = value
	return nil
}

func (db *simDB) Delete(key []byte) error {
	keyStr := string(key)
	delete(db.cache, keyStr)
	return nil
}

func (db *simDB) initDB(minNumber uint64, blkIndex map[uint64][]dbBlkIndex) error {
	db.Put(minNumberIndex, encodeUint64(minNumber))
	for number, index := range blkIndex {
		data, err := rlp.EncodeToBytes(index)
		if err != nil {
			return err
		}
		db.Put(append(blkIndexPrefix, encodeUint64(number)...), data)
	}
	return nil
}

func (db *simDB) checkState(cacheSize int, min uint64, indexTargets map[uint64][]dbBlkIndex) error {
	if len(db.cache) != cacheSize {
		return errors.Errorf("db数据: 数量错误, %d != %d", len(db.cache), cacheSize)
	}

	if data, err := db.Get(minNumberIndex); err != nil {
		return errors.Errorf("db数据: 获取最低高度索引失败: %v", err)
	} else {
		if minNumber, err := decodeUint64(data); err != nil {
			return errors.Errorf("db数据: 最低高度索引解码失败 err=%v", minNumber, err)
		} else {
			if minNumber != min {
				return errors.Errorf("db数据: 最低高度索引不匹配 db=%d target=%d", minNumber, min)
			}
		}
	}

	for number, checkData := range indexTargets {
		if err := db.checkIndex(number, checkData); err != nil {
			return errors.Errorf("db数据: 检查区块索引数据失败 高度[%d] err[%v]", number, err)
		}
	}
	return nil
}

func (db *simDB) checkIndex(number uint64, target []dbBlkIndex) error {
	data, _ := db.Get(append(blkIndexPrefix, encodeUint64(number)...))
	if len(data) == 0 && len(target) != 0 {
		return errors.Errorf("db size = 0 ; target size = %d", len(target))
	}

	index := make([]dbBlkIndex, 0)
	err := rlp.DecodeBytes(data, &index)
	if err != nil {
		return errors.Errorf("db 数据解码错误: %v", err)
	}

	if len(index) != len(target) {
		return errors.Errorf("db size = %d ; target size = %d", len(index), len(target))
	}

	for i := 0; i < len(target); i++ {
		find := false
		for j := 0; j < len(index); j++ {
			if target[i].Hash == index[j].Hash {
				find = true
				if target[i].InsertTime != index[j].InsertTime {
					return errors.Errorf("block[%s],  db time = %d ; target time = %d", target[i].Hash.Hex(), index[j].InsertTime, target[i].InsertTime)
				}
				break
			}
		}
		if find == false {
			return errors.Errorf("block[%s] can`t find in db", target[i].Hash.Hex())
		}
	}
	return nil
}

type simChain1 struct {
	curNumber uint64
}

func (chain *simChain1) CurrentHeader() *types.Header {
	header := &types.Header{
		Number: big.NewInt(int64(chain.curNumber)),
	}
	chain.curNumber += 2
	return header
}

func (chain *simChain1) DelLocalBlocks(blocks []*mc.BlockInfo) (fails []*mc.BlockInfo, err error) {
	return nil, nil
}

type simChain2 struct {
	curNumber  uint64
	delFailBlk map[common.Hash]int
}

func (chain *simChain2) CurrentHeader() *types.Header {
	header := &types.Header{
		Number: big.NewInt(int64(chain.curNumber)),
	}
	chain.curNumber += 2
	return header
}

func (chain *simChain2) DelLocalBlocks(blocks []*mc.BlockInfo) (fails []*mc.BlockInfo, err error) {
	for _, blk := range blocks {
		times, OK := chain.delFailBlk[blk.Hash]
		if OK && times > 0 {
			chain.delFailBlk[blk.Hash] = times - 1
			fails = append(fails, blk)
		}
	}
	if len(fails) != 0 {
		return fails, errors.New("删除部分区块失败!")
	}
	return nil, nil
}

type simChain3 struct {
	curNumber uint64
}

func (chain *simChain3) CurrentHeader() *types.Header {
	header := &types.Header{
		Number: big.NewInt(int64(chain.curNumber)),
	}
	return header
}

func (chain *simChain3) DelLocalBlocks(blocks []*mc.BlockInfo) (fails []*mc.BlockInfo, err error) {
	return nil, nil
}
