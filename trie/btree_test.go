package trie

import (
	"fmt"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/mandb"
	"math/rand"
	"os"
	"strconv"
	"testing"
)

func perm(n int) (out []Item) {
	for _, v := range rand.Perm(n) {
		out = append(out, Int(v))
	}
	return
}

func rang(n int) (out []Item) {
	for i := 0; i < n; i++ {
		out = append(out, Int(i))
	}
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
	treeRoot := BtreeSaveHash(tr1.root, triedb)

	RestoreBtree(tr2, nil, treeRoot, triedb)

	tr2.root.print(os.Stdout, 2)
}
