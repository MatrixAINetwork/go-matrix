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
// FUSE directory tree, for servers that wish to use it with the service loop.

package fs

import (
	"os"
	pathpkg "path"
	"strings"

	"golang.org/x/net/context"
)

import (
	"bazil.org/fuse"
)

// A Tree implements a basic read-only directory tree for FUSE.
// The Nodes contained in it may still be writable.
type Tree struct {
	tree
}

func (t *Tree) Root() (Node, error) {
	return &t.tree, nil
}

// Add adds the path to the tree, resolving to the given node.
// If path or a prefix of path has already been added to the tree,
// Add panics.
//
// Add is only safe to call before starting to serve requests.
func (t *Tree) Add(path string, node Node) {
	path = pathpkg.Clean("/" + path)[1:]
	elems := strings.Split(path, "/")
	dir := Node(&t.tree)
	for i, elem := range elems {
		dt, ok := dir.(*tree)
		if !ok {
			panic("fuse: Tree.Add for " + strings.Join(elems[:i], "/") + " and " + path)
		}
		n := dt.lookup(elem)
		if n != nil {
			if i+1 == len(elems) {
				panic("fuse: Tree.Add for " + path + " conflicts with " + elem)
			}
			dir = n
		} else {
			if i+1 == len(elems) {
				dt.add(elem, node)
			} else {
				dir = &tree{}
				dt.add(elem, dir)
			}
		}
	}
}

type treeDir struct {
	name string
	node Node
}

type tree struct {
	dir []treeDir
}

func (t *tree) lookup(name string) Node {
	for _, d := range t.dir {
		if d.name == name {
			return d.node
		}
	}
	return nil
}

func (t *tree) add(name string, n Node) {
	t.dir = append(t.dir, treeDir{name, n})
}

func (t *tree) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Mode = os.ModeDir | 0555
	return nil
}

func (t *tree) Lookup(ctx context.Context, name string) (Node, error) {
	n := t.lookup(name)
	if n != nil {
		return n, nil
	}
	return nil, fuse.ENOENT
}

func (t *tree) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	var out []fuse.Dirent
	for _, d := range t.dir {
		out = append(out, fuse.Dirent{Name: d.name})
	}
	return out, nil
}
