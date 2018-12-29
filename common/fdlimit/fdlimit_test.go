// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package fdlimit

import (
	"fmt"
	"testing"
)

// TestFileDescriptorLimits simply tests whether the file descriptor allowance
// per this process can be retrieved.
func TestFileDescriptorLimits(t *testing.T) {
	target := 4096
	hardlimit, err := Maximum()
	if err != nil {
		t.Fatal(err)
	}
	if hardlimit < target {
		t.Skip(fmt.Sprintf("system limit is less than desired test target: %d < %d", hardlimit, target))
	}

	if limit, err := Current(); err != nil || limit <= 0 {
		t.Fatalf("failed to retrieve file descriptor limit (%d): %v", limit, err)
	}
	if err := Raise(uint64(target)); err != nil {
		t.Fatalf("failed to raise file allowance")
	}
	if limit, err := Current(); err != nil || limit < target {
		t.Fatalf("failed to retrieve raised descriptor limit (have %v, want %v): %v", limit, target, err)
	}
}
