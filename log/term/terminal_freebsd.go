// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package term

import (
	"syscall"
)

const ioctlReadTermios = syscall.TIOCGETA

// Go 1.2 doesn't include Termios for FreeBSD. This should be added in 1.3 and this could be merged with terminal_darwin.
type Termios struct {
	Iflag  uint32
	Oflag  uint32
	Cflag  uint32
	Lflag  uint32
	Cc     [20]uint8
	Ispeed uint32
	Ospeed uint32
}
