// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package filter

import (
	"testing"
	"time"
)

// Simple test to check if baseline matching/mismatching filtering works.
func TestFilters(t *testing.T) {
	fm := New()
	fm.Start()

	// Register two filters to catch posted data
	first := make(chan struct{})
	fm.Install(Generic{
		Str1: "hello",
		Fn: func(data interface{}) {
			first <- struct{}{}
		},
	})
	second := make(chan struct{})
	fm.Install(Generic{
		Str1: "hello1",
		Str2: "hello",
		Fn: func(data interface{}) {
			second <- struct{}{}
		},
	})
	// Post an event that should only match the first filter
	fm.Notify(Generic{Str1: "hello"}, true)
	fm.Stop()

	// Ensure only the mathcing filters fire
	select {
	case <-first:
	case <-time.After(100 * time.Millisecond):
		t.Error("matching filter timed out")
	}
	select {
	case <-second:
		t.Error("mismatching filter fired")
	case <-time.After(100 * time.Millisecond):
	}
}
