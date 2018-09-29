// Copyright 2018 The MATRIX Authors as well as Copyright 2014-2017 The go-ethereum Authors
// This file is consisted of the MATRIX library and part of the go-ethereum library.
//
// The MATRIX-ethereum library is free software: you can redistribute it and/or modify it under the terms of the MIT License.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, 
//and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject tothe following conditions:
//
//The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
//
//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, 
//WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISINGFROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE
//OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
// +build !windows,!linux,!darwin,!openbsd,!freebsd,!netbsd

package liner

import (
	"bufio"
	"errors"
	"os"
)

// State represents an open terminal
type State struct {
	commonState
}

// Prompt displays p, and then waits for user input. Prompt does not support
// line editing on this operating system.
func (s *State) Prompt(p string) (string, error) {
	return s.promptUnsupported(p)
}

// PasswordPrompt is not supported in this OS.
func (s *State) PasswordPrompt(p string) (string, error) {
	return "", errors.New("liner: function not supported in this terminal")
}

// NewLiner initializes a new *State
//
// Note that this operating system uses a fallback mode without line
// editing. Patches welcome.
func NewLiner() *State {
	var s State
	s.r = bufio.NewReader(os.Stdin)
	return &s
}

// Close returns the terminal to its previous mode
func (s *State) Close() error {
	return nil
}

// TerminalSupported returns false because line editing is not
// supported on this platform.
func TerminalSupported() bool {
	return false
}

type noopMode struct{}

func (n noopMode) ApplyMode() error {
	return nil
}

// TerminalMode returns a noop InputModeSetter on this platform.
func TerminalMode() (ModeApplier, error) {
	return noopMode{}, nil
}
