// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package secp256k1

import "C"
import "unsafe"

// Callbacks for converting libsecp256k1 internal faults into
// recoverable Go panics.

//export secp256k1GoPanicIllegal
func secp256k1GoPanicIllegal(msg *C.char, data unsafe.Pointer) {
	panic("illegal argument: " + C.GoString(msg))
}

//export secp256k1GoPanicError
func secp256k1GoPanicError(msg *C.char, data unsafe.Pointer) {
	panic("internal error: " + C.GoString(msg))
}
