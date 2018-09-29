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
package fuse

// There is very little commonality in extended attribute errors
// across platforms.
//
// getxattr return value for "extended attribute does not exist" is
// ENOATTR on OS X, and ENODATA on Linux and apparently at least
// NetBSD. There may be a #define ENOATTR on Linux too, but the value
// is ENODATA in the actual syscalls. FreeBSD and OpenBSD have no
// ENODATA, only ENOATTR. ENOATTR is not in any of the standards,
// ENODATA exists but is only used for STREAMs.
//
// Each platform will define it a errNoXattr constant, and this file
// will enforce that it implements the right interfaces and hide the
// implementation.
//
// https://developer.apple.com/library/mac/documentation/Darwin/Reference/ManPages/man2/getxattr.2.html
// http://mail-index.netbsd.org/tech-kern/2012/04/30/msg013090.html
// http://mail-index.netbsd.org/tech-kern/2012/04/30/msg013097.html
// http://pubs.opengroup.org/onlinepubs/9699919799/basedefs/errno.h.html
// http://www.freebsd.org/cgi/man.cgi?query=extattr_get_file&sektion=2
// http://nixdoc.net/man-pages/openbsd/man2/extattr_get_file.2.html

// ErrNoXattr is a platform-independent error value meaning the
// extended attribute was not found. It can be used to respond to
// GetxattrRequest and such.
const ErrNoXattr = errNoXattr

var _ error = ErrNoXattr
var _ Errno = ErrNoXattr
var _ ErrorNumber = ErrNoXattr
