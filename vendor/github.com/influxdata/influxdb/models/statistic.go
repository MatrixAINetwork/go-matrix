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
package models

// Statistic is the representation of a statistic used by the monitoring service.
type Statistic struct {
	Name   string                 `json:"name"`
	Tags   map[string]string      `json:"tags"`
	Values map[string]interface{} `json:"values"`
}

// NewStatistic returns an initialized Statistic.
func NewStatistic(name string) Statistic {
	return Statistic{
		Name:   name,
		Tags:   make(map[string]string),
		Values: make(map[string]interface{}),
	}
}

// StatisticTags is a map that can be merged with others without causing
// mutations to either map.
type StatisticTags map[string]string

// Merge creates a new map containing the merged contents of tags and t.
// If both tags and the receiver map contain the same key, the value in tags
// is used in the resulting map.
//
// Merge always returns a usable map.
func (t StatisticTags) Merge(tags map[string]string) map[string]string {
	// Add everything in tags to the result.
	out := make(map[string]string, len(tags))
	for k, v := range tags {
		out[k] = v
	}

	// Only add values from t that don't appear in tags.
	for k, v := range t {
		if _, ok := tags[k]; !ok {
			out[k] = v
		}
	}
	return out
}
