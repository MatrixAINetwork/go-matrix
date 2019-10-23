// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package x11

import (
	"github.com/MatrixAINetwork/go-matrix/consensus/x11/blake"
	"github.com/MatrixAINetwork/go-matrix/consensus/x11/bmw"
	"github.com/MatrixAINetwork/go-matrix/consensus/x11/cubed"
	"github.com/MatrixAINetwork/go-matrix/consensus/x11/echo"
	"github.com/MatrixAINetwork/go-matrix/consensus/x11/groest"
	"github.com/MatrixAINetwork/go-matrix/consensus/x11/hash"
	"github.com/MatrixAINetwork/go-matrix/consensus/x11/jhash"
	"github.com/MatrixAINetwork/go-matrix/consensus/x11/keccak"
	"github.com/MatrixAINetwork/go-matrix/consensus/x11/luffa"
	"github.com/MatrixAINetwork/go-matrix/consensus/x11/shavite"
	"github.com/MatrixAINetwork/go-matrix/consensus/x11/simd"
	"github.com/MatrixAINetwork/go-matrix/consensus/x11/skein"
)

type hashObject struct {
	tha [64]byte
	thb [64]byte

	blake   hash.Digest
	bmw     hash.Digest
	cubed   hash.Digest
	echo    hash.Digest
	groest  hash.Digest
	jhash   hash.Digest
	keccak  hash.Digest
	luffa   hash.Digest
	shavite hash.Digest
	simd    hash.Digest
	skein   hash.Digest
}

func newObj() *hashObject {
	ref := &hashObject{}
	ref.blake = blake.New()
	ref.bmw = bmw.New()
	ref.cubed = cubed.New()
	ref.echo = echo.New()
	ref.groest = groest.New()
	ref.jhash = jhash.New()
	ref.keccak = keccak.New()
	ref.luffa = luffa.New()
	ref.shavite = shavite.New()
	ref.simd = simd.New()
	ref.skein = skein.New()
	return ref
}

// Hash computes the hash from the src bytes and stores the result in dst.
func (ref *hashObject) hash(src []byte, dst []byte) {
	ta := ref.tha[:]
	tb := ref.thb[:]

	ref.blake.Write(src)
	ref.blake.Close(tb, 0, 0)

	ref.bmw.Write(tb)
	ref.bmw.Close(ta, 0, 0)

	ref.groest.Write(ta)
	ref.groest.Close(tb, 0, 0)

	ref.skein.Write(tb)
	ref.skein.Close(ta, 0, 0)

	ref.jhash.Write(ta)
	ref.jhash.Close(tb, 0, 0)

	ref.keccak.Write(tb)
	ref.keccak.Close(ta, 0, 0)

	ref.luffa.Write(ta)
	ref.luffa.Close(tb, 0, 0)

	ref.cubed.Write(tb)
	ref.cubed.Close(ta, 0, 0)

	ref.shavite.Write(ta)
	ref.shavite.Close(tb, 0, 0)

	ref.simd.Write(tb)
	ref.simd.Close(ta, 0, 0)

	ref.echo.Write(ta)
	ref.echo.Close(tb, 0, 0)

	copy(dst, tb)
}

func Hash(src []byte) []byte {
	result := make([]byte, 32)
	obj := newObj()
	obj.hash(src, result)
	return result
}
