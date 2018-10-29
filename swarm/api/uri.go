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
	"fmt"
	"net/url"
	"strings"
)

// URI is a reference to content stored in swarm.
type URI struct {
	// Scheme has one of the following values:
	//
	// * bzz           - an entry in a swarm manifest
	// * bzz-raw       - raw swarm content
	// * bzz-immutable - immutable URI of an entry in a swarm manifest
	//                   (address is not resolved)
	// * bzz-list      -  list of all files contained in a swarm manifest
	//
	// Deprecated Schemes:
	// * bzzr - raw swarm content
	// * bzzi - immutable URI of an entry in a swarm manifest
	//          (address is not resolved)
	// * bzz-hash - hash of swarm content
	//
	Scheme string

	// Addr is either a hexadecimal storage key or it an address which
	// resolves to a storage key
	Addr string

	// Path is the path to the content within a swarm manifest
	Path string
}

// Parse parses rawuri into a URI struct, where rawuri is expected to have one
// of the following formats:
//
// * <scheme>:/
// * <scheme>:/<addr>
// * <scheme>:/<addr>/<path>
// * <scheme>://
// * <scheme>://<addr>
// * <scheme>://<addr>/<path>
//
// with scheme one of bzz, bzz-raw, bzz-immutable, bzz-list or bzz-hash
// or deprecated ones bzzr and bzzi
func Parse(rawuri string) (*URI, error) {
	u, err := url.Parse(rawuri)
	if err != nil {
		return nil, err
	}
	uri := &URI{Scheme: u.Scheme}

	// check the scheme is valid
	switch uri.Scheme {
	case "bzz", "bzz-raw", "bzz-immutable", "bzz-list", "bzz-hash", "bzzr", "bzzi":
	default:
		return nil, fmt.Errorf("unknown scheme %q", u.Scheme)
	}

	// handle URIs like bzz://<addr>/<path> where the addr and path
	// have already been split by url.Parse
	if u.Host != "" {
		uri.Addr = u.Host
		uri.Path = strings.TrimLeft(u.Path, "/")
		return uri, nil
	}

	// URI is like bzz:/<addr>/<path> so split the addr and path from
	// the raw path (which will be /<addr>/<path>)
	parts := strings.SplitN(strings.TrimLeft(u.Path, "/"), "/", 2)
	uri.Addr = parts[0]
	if len(parts) == 2 {
		uri.Path = parts[1]
	}
	return uri, nil
}

func (u *URI) Raw() bool {
	return u.Scheme == "bzz-raw"
}

func (u *URI) Immutable() bool {
	return u.Scheme == "bzz-immutable"
}

func (u *URI) List() bool {
	return u.Scheme == "bzz-list"
}

func (u *URI) DeprecatedRaw() bool {
	return u.Scheme == "bzzr"
}

func (u *URI) DeprecatedImmutable() bool {
	return u.Scheme == "bzzi"
}

func (u *URI) Hash() bool {
	return u.Scheme == "bzz-hash"
}

func (u *URI) String() string {
	return u.Scheme + ":/" + u.Addr + "/" + u.Path
}
