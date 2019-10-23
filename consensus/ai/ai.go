// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package ai

/*
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/crypto/sha3"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
)

var instance *Picture

var picHashList = []common.Hash{
	common.HexToHash("0x130d6675a1ed8d31e7ce8900e572a051770ac3bb0cdb7b7bbf94dbdb4f321d93"),
	common.HexToHash("0x484be396fd2dab5eaff13df071aa238e05b78c4e4c5caa2b3524cf52346f55ec"),
	common.HexToHash("0x6bb1ffeda6cf38139452eda2f23b5bb2034aa1e5829781b9a5ed21f9cf104a9c"),
	common.HexToHash("0xd00caf04f906ef66ed47bd07de253f1d2555ac1443474ffd92ff3bc8d31d0c5c"),
	common.HexToHash("0xefbbc3ba84b265fe775f90887d3106ae14101cefcfa0cd2421587a85cdd1a2dd"),
	common.HexToHash("0xb33a2b4015ee7264c1de382258d0e82ee887671c1e7bb496a4af93dbe2ae40c6"),
	common.HexToHash("0x4c7edef337e9c20016bc09b7a3fceb2e5f211e54e517ff9748dbc1e178bb944e"),
	common.HexToHash("0xcaef02f6e6f6c1bbe7deee6c4115a20c07641c5f62ae1eee923c01e30db430de"),
	common.HexToHash("0x6c15aa4dbd0d72a720dca043980a008b7fd3c210a9082727664173de5b04e47e"),
	common.HexToHash("0x7028844036870d1c79c8b95f30199624c8e21b71ca2d38945dbcb8b8aa33477d"),
	common.HexToHash("0x5a1d3c2a23eae610ab4d72723e388c372c7f9cf41c885ccdbd964924b2198906"),
	common.HexToHash("0x59309bd833a7e72591c7f3209759785bcb052a38e9c96f1da109f243b0a5dbc6"),
	common.HexToHash("0x980584165d1d4cd709a9663967330ad8dbaa45a8fdf7b6db5bfe5ffb8e5cc5fd"),
	common.HexToHash("0x248fe2b2c9fe246d6960cb3adfd541e07659f43575a6adebb47585e6116eeff8"),
	common.HexToHash("0xe7f4c74c1ee620c19b9c9282ad27972442006997b3302aef8fa3d967321803a1"),
	common.HexToHash("0x6862799d569399f465cb776f099167bc595236b8101e381baef82f5903f0eddd"),
}

func Init(picStorePath string) {
	if instance == nil {
		pictureList := make([]string, 0)
		for i := 0; i < 16; i++ {
			pictureList = append(pictureList, filepath.Join(picStorePath, "matrix_"+strconv.Itoa(i)+".jpg"))
		}
		for index, path := range pictureList {
			if err := checkPictures(index, path); err != nil {
				fatalf("ai picture store err: %v", err)
			}
		}

		var err error
		instance, err = New(pictureList)
		if err != nil {
			fatalf("init ai instance err: %v", err)
		}
	}
}

func Mining(seed int64, stopCh chan struct{}, resultCh chan []byte, errCh chan error) {
	found := make(chan []byte, 1)
	go aiMining(seed, found, errCh)

	select {
	case <-stopCh:
		log.Info("ai mining", "receive stop signal", "stop ai mining")
		return

	case rst := <-found:
		log.Info("ai mining", "get ai digging result", rst)
		resultCh <- rst
		return
	}
}

func aiMining(seed int64, resultCh chan []byte, errCh chan error) {
	rst := instance.AIMine(seed)
	resultCh <- rst
}

func checkPictures(index int, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return errors.Errorf("open No.%d picture(%s) err(%v)", index+1, path, err)
	}
	defer file.Close()

	// check picture file info
	body, err := ioutil.ReadAll(file)
	if err != nil {
		return errors.Errorf("read No.%d picture err(%v)", index+1, err)
	}

	hash := sha3.Sum256(body)
	fileHash := common.BytesToHash(hash[:])

	if fileHash != picHashList[index] {
		return errors.Errorf("No.%d picture(%s) hash not match", index+1, path)
	}

	return nil
}

func fatalf(format string, args ...interface{}) {
	w := io.MultiWriter(os.Stdout, os.Stderr)
	if runtime.GOOS == "windows" {
		// The SameFile check below doesn't work on Windows.
		// stdout is unlikely to get redirected though, so just print there.
		w = os.Stdout
	} else {
		outf, _ := os.Stdout.Stat()
		errf, _ := os.Stderr.Stat()
		if outf != nil && errf != nil && os.SameFile(outf, errf) {
			w = os.Stderr
		}
	}
	fmt.Fprintf(w, "Fatal: "+format+"\n", args...)
	os.Exit(1)
}
