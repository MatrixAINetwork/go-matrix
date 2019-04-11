// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

// Package filter implements event filters.
package filter

import "reflect"

type Filter interface {
	Compare(Filter) bool
	Trigger(data interface{})
}

type FilterEvent struct {
	filter Filter
	data   interface{}
}

type Filters struct {
	id       int
	watchers map[int]Filter
	ch       chan FilterEvent

	quit chan struct{}
}

func New() *Filters {
	return &Filters{
		ch:       make(chan FilterEvent),
		watchers: make(map[int]Filter),
		quit:     make(chan struct{}),
	}
}

func (f *Filters) Start() {
	go f.loop()
}

func (f *Filters) Stop() {
	close(f.quit)
}

func (f *Filters) Notify(filter Filter, data interface{}) {
	f.ch <- FilterEvent{filter, data}
}

func (f *Filters) Install(watcher Filter) int {
	f.watchers[f.id] = watcher
	f.id++

	return f.id - 1
}

func (f *Filters) Uninstall(id int) {
	delete(f.watchers, id)
}

func (f *Filters) loop() {
out:
	for {
		select {
		case <-f.quit:
			break out
		case event := <-f.ch:
			for _, watcher := range f.watchers {
				if reflect.TypeOf(watcher) == reflect.TypeOf(event.filter) {
					if watcher.Compare(event.filter) {
						watcher.Trigger(event.data)
					}
				}
			}
		}
	}
}

func (f *Filters) Match(a, b Filter) bool {
	return reflect.TypeOf(a) == reflect.TypeOf(b) && a.Compare(b)
}

func (f *Filters) Get(i int) Filter {
	return f.watchers[i]
}
