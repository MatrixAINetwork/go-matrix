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
// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package html

// Section 12.2.3.2 of the HTML5 specification says "The following elements
// have varying levels of special parsing rules".
// https://html.spec.whatwg.org/multipage/syntax.html#the-stack-of-open-elements
var isSpecialElementMap = map[string]bool{
	"address":    true,
	"applet":     true,
	"area":       true,
	"article":    true,
	"aside":      true,
	"base":       true,
	"basefont":   true,
	"bgsound":    true,
	"blockquote": true,
	"body":       true,
	"br":         true,
	"button":     true,
	"caption":    true,
	"center":     true,
	"col":        true,
	"colgroup":   true,
	"dd":         true,
	"details":    true,
	"dir":        true,
	"div":        true,
	"dl":         true,
	"dt":         true,
	"embed":      true,
	"fieldset":   true,
	"figcaption": true,
	"figure":     true,
	"footer":     true,
	"form":       true,
	"frame":      true,
	"frameset":   true,
	"h1":         true,
	"h2":         true,
	"h3":         true,
	"h4":         true,
	"h5":         true,
	"h6":         true,
	"head":       true,
	"header":     true,
	"hgroup":     true,
	"hr":         true,
	"html":       true,
	"iframe":     true,
	"img":        true,
	"input":      true,
	"isindex":    true,
	"li":         true,
	"link":       true,
	"listing":    true,
	"marquee":    true,
	"menu":       true,
	"meta":       true,
	"nav":        true,
	"noembed":    true,
	"noframes":   true,
	"noscript":   true,
	"object":     true,
	"ol":         true,
	"p":          true,
	"param":      true,
	"plaintext":  true,
	"pre":        true,
	"script":     true,
	"section":    true,
	"select":     true,
	"source":     true,
	"style":      true,
	"summary":    true,
	"table":      true,
	"tbody":      true,
	"td":         true,
	"template":   true,
	"textarea":   true,
	"tfoot":      true,
	"th":         true,
	"thead":      true,
	"title":      true,
	"tr":         true,
	"track":      true,
	"ul":         true,
	"wbr":        true,
	"xmp":        true,
}

func isSpecialElement(element *Node) bool {
	switch element.Namespace {
	case "", "html":
		return isSpecialElementMap[element.Data]
	case "svg":
		return element.Data == "foreignObject"
	}
	return false
}
