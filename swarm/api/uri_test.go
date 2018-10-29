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

package api

import (
	"reflect"
	"testing"
)

func TestParseURI(t *testing.T) {
	type test struct {
		uri                       string
		expectURI                 *URI
		expectErr                 bool
		expectRaw                 bool
		expectImmutable           bool
		expectList                bool
		expectHash                bool
		expectDeprecatedRaw       bool
		expectDeprecatedImmutable bool
	}
	tests := []test{
		{
			uri:       "",
			expectErr: true,
		},
		{
			uri:       "foo",
			expectErr: true,
		},
		{
			uri:       "bzz",
			expectErr: true,
		},
		{
			uri:       "bzz:",
			expectURI: &URI{Scheme: "bzz"},
		},
		{
			uri:             "bzz-immutable:",
			expectURI:       &URI{Scheme: "bzz-immutable"},
			expectImmutable: true,
		},
		{
			uri:       "bzz-raw:",
			expectURI: &URI{Scheme: "bzz-raw"},
			expectRaw: true,
		},
		{
			uri:       "bzz:/",
			expectURI: &URI{Scheme: "bzz"},
		},
		{
			uri:       "bzz:/abc123",
			expectURI: &URI{Scheme: "bzz", Addr: "abc123"},
		},
		{
			uri:       "bzz:/abc123/path/to/entry",
			expectURI: &URI{Scheme: "bzz", Addr: "abc123", Path: "path/to/entry"},
		},
		{
			uri:       "bzz-raw:/",
			expectURI: &URI{Scheme: "bzz-raw"},
			expectRaw: true,
		},
		{
			uri:       "bzz-raw:/abc123",
			expectURI: &URI{Scheme: "bzz-raw", Addr: "abc123"},
			expectRaw: true,
		},
		{
			uri:       "bzz-raw:/abc123/path/to/entry",
			expectURI: &URI{Scheme: "bzz-raw", Addr: "abc123", Path: "path/to/entry"},
			expectRaw: true,
		},
		{
			uri:       "bzz://",
			expectURI: &URI{Scheme: "bzz"},
		},
		{
			uri:       "bzz://abc123",
			expectURI: &URI{Scheme: "bzz", Addr: "abc123"},
		},
		{
			uri:       "bzz://abc123/path/to/entry",
			expectURI: &URI{Scheme: "bzz", Addr: "abc123", Path: "path/to/entry"},
		},
		{
			uri:        "bzz-hash:",
			expectURI:  &URI{Scheme: "bzz-hash"},
			expectHash: true,
		},
		{
			uri:        "bzz-hash:/",
			expectURI:  &URI{Scheme: "bzz-hash"},
			expectHash: true,
		},
		{
			uri:        "bzz-list:",
			expectURI:  &URI{Scheme: "bzz-list"},
			expectList: true,
		},
		{
			uri:        "bzz-list:/",
			expectURI:  &URI{Scheme: "bzz-list"},
			expectList: true,
		},
		{
			uri:                 "bzzr:",
			expectURI:           &URI{Scheme: "bzzr"},
			expectDeprecatedRaw: true,
		},
		{
			uri:                 "bzzr:/",
			expectURI:           &URI{Scheme: "bzzr"},
			expectDeprecatedRaw: true,
		},
		{
			uri:                       "bzzi:",
			expectURI:                 &URI{Scheme: "bzzi"},
			expectDeprecatedImmutable: true,
		},
		{
			uri:                       "bzzi:/",
			expectURI:                 &URI{Scheme: "bzzi"},
			expectDeprecatedImmutable: true,
		},
	}
	for _, x := range tests {
		actual, err := Parse(x.uri)
		if x.expectErr {
			if err == nil {
				t.Fatalf("expected %s to error", x.uri)
			}
			continue
		}
		if err != nil {
			t.Fatalf("error parsing %s: %s", x.uri, err)
		}
		if !reflect.DeepEqual(actual, x.expectURI) {
			t.Fatalf("expected %s to return %#v, got %#v", x.uri, x.expectURI, actual)
		}
		if actual.Raw() != x.expectRaw {
			t.Fatalf("expected %s raw to be %t, got %t", x.uri, x.expectRaw, actual.Raw())
		}
		if actual.Immutable() != x.expectImmutable {
			t.Fatalf("expected %s immutable to be %t, got %t", x.uri, x.expectImmutable, actual.Immutable())
		}
		if actual.List() != x.expectList {
			t.Fatalf("expected %s list to be %t, got %t", x.uri, x.expectList, actual.List())
		}
		if actual.Hash() != x.expectHash {
			t.Fatalf("expected %s hash to be %t, got %t", x.uri, x.expectHash, actual.Hash())
		}
		if actual.DeprecatedRaw() != x.expectDeprecatedRaw {
			t.Fatalf("expected %s deprecated raw to be %t, got %t", x.uri, x.expectDeprecatedRaw, actual.DeprecatedRaw())
		}
		if actual.DeprecatedImmutable() != x.expectDeprecatedImmutable {
			t.Fatalf("expected %s deprecated immutable to be %t, got %t", x.uri, x.expectDeprecatedImmutable, actual.DeprecatedImmutable())
		}
	}
}
