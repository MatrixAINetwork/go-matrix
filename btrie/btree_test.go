// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package btrie

import (
	"fmt"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/mandb"
	_ "math/rand"
	"os"
	"strconv"
	"testing"
)

func perm(n int) (out []Item) {
	//for _, _ := range rand.Perm(n) {
	//	//out = append(out, Int(v))
	//}
	return
}

func rang(n int) (out []Item) {
	//for i := 0; i < n; i++ {
	//	out = append(out, Int(i))
	//}
	return
}

var x []map[common.Hash][]byte = make([]map[common.Hash][]byte, 0)

func mapMake() []map[common.Hash][]byte {

	for j := 0; j < 100; j++ {
		y := make(map[common.Hash][]byte)
		for i := 0; i < 10; i++ {
			y[common.HexToHash(strconv.Itoa(i))] = common.Hex2Bytes("0e3d")
		}
		x = append(x, y)
	}
	return x
}
func allrev(t *BTree) (out []Item) {
	t.Descend(func(a Item) bool {
		out = append(out, a)
		return true
	})
	return
}

func rangSpcial(n int, m []map[common.Hash][]byte) (out []Item) {

	for i := 0; i < n; i++ {
		tmpData := SpcialTxData{Key_Time: uint32(i), Value_Tx: m[i]}
		out = append(out, SpcialTxData(tmpData))
	}
	return

}

//var btreeDegree = flag.Int("degree", 32, "B-Tree degree")

func TestBTree(t *testing.T) {
	triedb := NewDatabase(mandb.NewMemDatabase())

	tr1 := NewBtree(2, triedb)
	tr2 := NewBtree(2, triedb)

	m := mapMake()

	for _, v := range rangSpcial(30, m) {
		tr1.ReplaceOrInsert(v)
	}

	tr1.root.print(os.Stdout, 2)
	fmt.Println("============================================================================================")
	treeRoot := BtreeSaveHash(tr1.root, triedb, 0)

	RestoreBtree(tr2, nil, treeRoot, triedb, 0)

	tr2.root.print(os.Stdout, 2)
	tr1.Delete(SpcialTxData{Key_Time: 5})
	tr1.root.print(os.Stdout, 2)
}
