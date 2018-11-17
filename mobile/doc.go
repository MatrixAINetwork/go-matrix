// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

// Package gman contains the simplified mobile APIs to go-matrix.
//
// The scope of this package is *not* to allow writing a custom Matrix client
// with pieces plucked from go-matrix, rather to allow writing native dapps on
// mobile platforms. Keep this in mind when using or extending this package!
//
// API limitations
//
// Since gomobile cannot bridge arbitrary types between Go and Android/iOS, the
// exposed APIs need to be manually wrapped into simplified types, with custom
// constructors and getters/setters to ensure that they can be meaninfully used
// from Java/ObjC too.
//
// With this in mind, please try to limit the scope of this package and only add
// essentials without which mobile support cannot work, especially since manually
// syncing the code will be unwieldy otherwise. In the long term we might consider
// writing custom library generators, but those are out of scope now.
//
// Content wise each file in this package corresponds to an entire Go package
// from the go-matrix repository. Please adhere to this scoping to prevent this
// package getting unmaintainable.
//
// Wrapping guidelines:
//
// Every type that is to be exposed should be wrapped into its own plain struct,
// which internally contains a single field: the original go-matrix version.
// This is needed because gomobile cannot expose named types for now.
//
// Whenever a method argument or a return type is a custom struct, the pointer
// variant should always be used as value types crossing over between language
// boundaries might have strange behaviors.
//
// Slices of types should be converted into a single multiplicative type wrapping
// a go slice with the methods `Size`, `Get` and `Set`. Further slice operations
// should not be provided to limit the remote code complexity. Arrays should be
// avoided as much as possible since they complicate bounds checking.
//
// If a method has multiple return values (e.g. some return + an error), those
// are generated as output arguments in ObjC. To avoid weird generated names like
// ret_0 for them, please always assign names to output variables if tuples.
//
// Note, a panic *cannot* cross over language boundaries, instead will result in
// an undebuggable SEGFAULT in the process. For error handling only ever use error
// returns, which may be the only or the second return.
package gman
