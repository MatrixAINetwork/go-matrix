// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package x11

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"github.com/MatrixAINetwork/go-matrix/common"
	"testing"
)

func TestFunc(t *testing.T) {
	input := common.Hex2Bytes("0000002002aacd6033b9ba23494220d2af3b0ccf8b02b06a533b08bd0e00000000000000be172e11d0d4f444c1c1d5678f665e14668872cdb618f256af14e9674d703a8e3f8c355d6b451a1973f802ec")
	t.Log(input)
	hash := Hash(input)
	t.Log(hash)

	binary.Read(bytes.NewReader(hash), binary.BigEndian, hash)

	hexStr := common.BytesToHash(hash).Hex()
	t.Log(hexStr)
}

func TestHash(t *testing.T) {
	for i := range tsInfo {
		ln := len(tsInfo[i].out)
		dest := make([]byte, ln)

		out := Hash(tsInfo[i].in[:])
		if ln != hex.Encode(dest, out[:]) {
			t.Errorf("%s: invalid length", tsInfo[i])
		}
		if !bytes.Equal(dest[:], tsInfo[i].out[:]) {
			t.Errorf("%s: invalid hash", tsInfo[i].id)
		}

		hash := common.BytesToHash(out)
		t.Logf("%d hash: %s", i, hash.Hex())

	}
}

////////////////

var tsInfo = []struct {
	id  string
	in  []byte
	out []byte
}{
	{
		"Empty",
		[]byte(""),
		[]byte("51b572209083576ea221c27e62b4e22063257571ccb6cc3dc3cd17eb67584eba"),
	},
	{
		"Dash",
		[]byte("DASH"),
		[]byte("fe809ebca8753d907f6ad32cdcf8e5c4e090d7bece5df35b2147e10b88c12d26"),
	},
	{
		"Fox",
		[]byte("The quick brown fox jumps over the lazy dog"),
		[]byte("534536a4e4f16b32447f02f77200449dc2f23b532e3d9878fe111c9de666bc5c"),
	},
}
