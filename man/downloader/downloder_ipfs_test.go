// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package downloader

import (
	"fmt"
	"math/big"
	"os"
	"path"
	"testing"
	"time"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/ethdb"
	"github.com/MatrixAINetwork/go-matrix/event"
)

func TestIPfsInit(t *testing.T) {
	testdb := mandb.NewMemDatabase()
	genesis := core.GenesisBlockForTesting(testdb, testAddress, big.NewInt(1000000000))

	tester := &downloadTester{
		genesis:           genesis,
		peerDb:            testdb,
		ownHashes:         []common.Hash{genesis.Hash()},
		ownHeaders:        map[common.Hash]*types.Header{genesis.Hash(): genesis.Header()},
		ownBlocks:         map[common.Hash]*types.Block{genesis.Hash(): genesis},
		ownReceipts:       map[common.Hash]types.Receipts{genesis.Hash(): nil},
		ownChainTd:        map[common.Hash]*big.Int{genesis.Hash(): genesis.Difficulty()},
		peerHashes:        make(map[string][]common.Hash),
		peerHeaders:       make(map[string]map[common.Hash]*types.Header),
		peerBlocks:        make(map[string]map[common.Hash]*types.Block),
		peerReceipts:      make(map[string]map[common.Hash]types.Receipts),
		peerChainTds:      make(map[string]map[common.Hash]*big.Int),
		peerMissingStates: make(map[string]map[common.Hash]bool),
	}
	tester.stateDb = mandb.NewMemDatabase()
	tester.stateDb.Put(genesis.Root().Bytes(), []byte{0x00})

	tester.downloader = New(FullSync, tester.stateDb, new(event.TypeMux), tester, nil, tester.dropPeer)
	//newIpfsDownload()
	//dl.ipfsBodyCh = make(chan BlockIpfs, 1)
	//go dl.IpfsDownloadInit()
	//IPfsDownloader()
	//if ipfsDownloadInit() != nil {
	//若是返回失败，则置ipfs模式为false
	//	dl.IpfsMode = false
	//}
	//	= New(FullSync, tester.stateDb, new(event.TypeMux), tester, nil, tester.dropPeer)
	//	go IpfsProcessRcvHead()
	//go WaitBlockInfoFromIpfs()
}
func TestRecv(t *testing.T) {
	testd := newTester()
	time.Sleep(5 * time.Second) //waitgroup
	//	tester := newTester()
	//var Address [20]byte
	//	var elct []common.Elect
	//Address = [20]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	/*var elct common.Elect
	// for i := 0; i < 20; i++ {
	// 	elct.Account[i] = byte(i)
	// }
	elct.Account = common.Address(Address)
	elct.Stock = 16
	elct.Type = 0x002
	*/
	var sgn [65]byte
	for i := 0; i < 65; i++ {
		sgn[i] = byte(i)
	}

	testHeader := &types.Header{
		ParentHash: common.BigToHash(big.NewInt(100)),
		UncleHash:  common.BigToHash(big.NewInt(100)),
		//		Leader:      nil,
		Coinbase:    common.BigToAddress(big.NewInt(123)),
		Root:        common.BigToHash(big.NewInt(100)),
		TxHash:      common.BigToHash(big.NewInt(100)),
		ReceiptHash: common.BigToHash(big.NewInt(100)),
		//			Bloom       Bloom          `json:"logsBloom"        gencodec:"required"`
		Difficulty: big.NewInt(int64(600)),
		Number:     big.NewInt(int64(2)),
		GasLimit:   937333,
		GasUsed:    937445,
		Time:       big.NewInt(int64(900)),
		//	Elect:      []common.Elect{elct, elct},
		//		NetTopology: nil,
		//	Signatures: []common.Signature{sgn},

		Extra: []byte{'s', 's', 's', 's', 'l', 'i', 'u'}, // //[]byte{1, 2, 3, 4, 5, 6, 7},
		//		MixDigest: nil,
		//		Nonce:     nil,
		//	Version: []byte{1, 2, 3, 4, 5, 6, 7},
	}
	// 存block
	Testblock := types.NewBlockWithHeader(testHeader)
	testd.downloader.RecvBlockToDeal(Testblock)

	time.Sleep(10000 * time.Second)
	/*
		// 下载block
		types.NewBlockWithHeader(testHeader)
		testd.downloader.SyncBlockFromIpfs(testHeader.Hash(), testHeader.Number.Uint64())

		time.Sleep(1000 * time.Second)*/

	//	tester.downloader.
}
func TestBatchRecv(t *testing.T) {
	testd := newTester()
	time.Sleep(5 * time.Second)
	testHeader := &types.Header{
		ParentHash: common.BigToHash(big.NewInt(100)),
		UncleHash:  common.BigToHash(big.NewInt(100)),
		//		Leader:      nil,
		Coinbase:    common.BigToAddress(big.NewInt(123)),
		Root:        common.BigToHash(big.NewInt(100)),
		TxHash:      common.BigToHash(big.NewInt(100)),
		ReceiptHash: common.BigToHash(big.NewInt(100)),
		//			Bloom       Bloom          `json:"logsBloom"        gencodec:"required"`
		Difficulty: big.NewInt(int64(600)),
		Number:     big.NewInt(int64(2)),
		GasLimit:   111,
		GasUsed:    222,
		Time:       big.NewInt(int64(900)),
		Extra:      []byte{'s', 's', 's', 's', 'l', 'i', 'u'}, // //[]byte{1, 2, 3, 4, 5, 6, 7},

	}
	// 存block
	/*
		for i := 5; i < 42; i++ {
			testHeader.Number = big.NewInt(int64(i))
			testHeader.GasLimit = 111
			Testblock := types.NewBlockWithHeader(testHeader)
			fmt.Println("\n---------------------%d------------------------\n", i)
			testd.downloader.RecvBlockToDeal(Testblock)
			if i == 8 || i == 15 { //改变limit 使一个number 可对应多个 hash
				testHeader.GasLimit = 158
				Testblock = types.NewBlockWithHeader(testHeader)
				testd.downloader.RecvBlockToDeal(Testblock)
			}
			//time.Sleep(5 * time.Second)
		}
		time.Sleep(50000 * time.Second)*/

	// 不断去获取
	for i := 5; i < 42; i++ {
		testHeader.Number = big.NewInt(int64(i))
		testHeader.GasLimit = 111
		fmt.Println("\n---------下载区块-------%d------------------------\n", i)
		testd.downloader.SyncBlockFromIpfs(testHeader.Hash(), testHeader.Number.Uint64())
	}
}
func TestIpfs(t *testing.T) {
	//testd := newTester()
	//testd.downloader.IPfsDirectoryUpdate()
	CheckDirAndCreate(strCacheDirectory)
	testd := newTester()
	testd.downloader.IpfsDownloadInit()
	//IpfsAddNewFile("asd.txt")
}
func TestClearIPFSAll(t *testing.T) {
	tmpFile, err := os.OpenFile(path.Join(strCacheDirectory, strLastestBlockFile), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	fmt.Println("file1", err)
	tmpFile2, err := os.OpenFile(path.Join(strCacheDirectory, strCache1BlockFile), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	fmt.Println("file1", err)

	/*len, err := tmpFile.WriteString("1")
	fmt.Println("file1", len, err)
	len, err = tmpFile2.WriteString("2")
	fmt.Println("file2", len, err)*/
	tmpFile.Close()
	tmpFile2.Close()

	testd := newTester()
	time.Sleep(5 * time.Second) //
	//发布目录
	testd.downloader.IPfsDirectoryUpdate()
}

func TestGetAll1and2IpfsInfo(t *testing.T) {

}
