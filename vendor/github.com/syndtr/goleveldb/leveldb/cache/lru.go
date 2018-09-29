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
// Copyright (c) 2012, Suryandaru Triandana <syndtr@gmail.com>
// All rights reserved.
//
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cache

import (
	"sync"
	"unsafe"
)

type lruNode struct {
	n   *Node
	h   *Handle
	ban bool

	next, prev *lruNode
}

func (n *lruNode) insert(at *lruNode) {
	x := at.next
	at.next = n
	n.prev = at
	n.next = x
	x.prev = n
}

func (n *lruNode) remove() {
	if n.prev != nil {
		n.prev.next = n.next
		n.next.prev = n.prev
		n.prev = nil
		n.next = nil
	} else {
		panic("BUG: removing removed node")
	}
}

type lru struct {
	mu       sync.Mutex
	capacity int
	used     int
	recent   lruNode
}

func (r *lru) reset() {
	r.recent.next = &r.recent
	r.recent.prev = &r.recent
	r.used = 0
}

func (r *lru) Capacity() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.capacity
}

func (r *lru) SetCapacity(capacity int) {
	var evicted []*lruNode

	r.mu.Lock()
	r.capacity = capacity
	for r.used > r.capacity {
		rn := r.recent.prev
		if rn == nil {
			panic("BUG: invalid LRU used or capacity counter")
		}
		rn.remove()
		rn.n.CacheData = nil
		r.used -= rn.n.Size()
		evicted = append(evicted, rn)
	}
	r.mu.Unlock()

	for _, rn := range evicted {
		rn.h.Release()
	}
}

func (r *lru) Promote(n *Node) {
	var evicted []*lruNode

	r.mu.Lock()
	if n.CacheData == nil {
		if n.Size() <= r.capacity {
			rn := &lruNode{n: n, h: n.GetHandle()}
			rn.insert(&r.recent)
			n.CacheData = unsafe.Pointer(rn)
			r.used += n.Size()

			for r.used > r.capacity {
				rn := r.recent.prev
				if rn == nil {
					panic("BUG: invalid LRU used or capacity counter")
				}
				rn.remove()
				rn.n.CacheData = nil
				r.used -= rn.n.Size()
				evicted = append(evicted, rn)
			}
		}
	} else {
		rn := (*lruNode)(n.CacheData)
		if !rn.ban {
			rn.remove()
			rn.insert(&r.recent)
		}
	}
	r.mu.Unlock()

	for _, rn := range evicted {
		rn.h.Release()
	}
}

func (r *lru) Ban(n *Node) {
	r.mu.Lock()
	if n.CacheData == nil {
		n.CacheData = unsafe.Pointer(&lruNode{n: n, ban: true})
	} else {
		rn := (*lruNode)(n.CacheData)
		if !rn.ban {
			rn.remove()
			rn.ban = true
			r.used -= rn.n.Size()
			r.mu.Unlock()

			rn.h.Release()
			rn.h = nil
			return
		}
	}
	r.mu.Unlock()
}

func (r *lru) Evict(n *Node) {
	r.mu.Lock()
	rn := (*lruNode)(n.CacheData)
	if rn == nil || rn.ban {
		r.mu.Unlock()
		return
	}
	n.CacheData = nil
	r.mu.Unlock()

	rn.h.Release()
}

func (r *lru) EvictNS(ns uint64) {
	var evicted []*lruNode

	r.mu.Lock()
	for e := r.recent.prev; e != &r.recent; {
		rn := e
		e = e.prev
		if rn.n.NS() == ns {
			rn.remove()
			rn.n.CacheData = nil
			r.used -= rn.n.Size()
			evicted = append(evicted, rn)
		}
	}
	r.mu.Unlock()

	for _, rn := range evicted {
		rn.h.Release()
	}
}

func (r *lru) EvictAll() {
	r.mu.Lock()
	back := r.recent.prev
	for rn := back; rn != &r.recent; rn = rn.prev {
		rn.n.CacheData = nil
	}
	r.reset()
	r.mu.Unlock()

	for rn := back; rn != &r.recent; rn = rn.prev {
		rn.h.Release()
	}
}

func (r *lru) Close() error {
	return nil
}

// NewLRU create a new LRU-cache.
func NewLRU(capacity int) Cacher {
	r := &lru{capacity: capacity}
	r.reset()
	return r
}
