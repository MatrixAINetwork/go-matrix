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

package whisperv5

import (
	"encoding/json"
	"testing"
)

var topicStringTests = []struct {
	topic TopicType
	str   string
}{
	{topic: TopicType{0x00, 0x00, 0x00, 0x00}, str: "0x00000000"},
	{topic: TopicType{0x00, 0x7f, 0x80, 0xff}, str: "0x007f80ff"},
	{topic: TopicType{0xff, 0x80, 0x7f, 0x00}, str: "0xff807f00"},
	{topic: TopicType{0xf2, 0x6e, 0x77, 0x79}, str: "0xf26e7779"},
}

func TestTopicString(t *testing.T) {
	for i, tst := range topicStringTests {
		s := tst.topic.String()
		if s != tst.str {
			t.Fatalf("failed test %d: have %s, want %s.", i, s, tst.str)
		}
	}
}

var bytesToTopicTests = []struct {
	data  []byte
	topic TopicType
}{
	{topic: TopicType{0x8f, 0x9a, 0x2b, 0x7d}, data: []byte{0x8f, 0x9a, 0x2b, 0x7d}},
	{topic: TopicType{0x00, 0x7f, 0x80, 0xff}, data: []byte{0x00, 0x7f, 0x80, 0xff}},
	{topic: TopicType{0x00, 0x00, 0x00, 0x00}, data: []byte{0x00, 0x00, 0x00, 0x00}},
	{topic: TopicType{0x00, 0x00, 0x00, 0x00}, data: []byte{0x00, 0x00, 0x00}},
	{topic: TopicType{0x01, 0x00, 0x00, 0x00}, data: []byte{0x01}},
	{topic: TopicType{0x00, 0xfe, 0x00, 0x00}, data: []byte{0x00, 0xfe}},
	{topic: TopicType{0xea, 0x1d, 0x43, 0x00}, data: []byte{0xea, 0x1d, 0x43}},
	{topic: TopicType{0x6f, 0x3c, 0xb0, 0xdd}, data: []byte{0x6f, 0x3c, 0xb0, 0xdd, 0x0f, 0x00, 0x90}},
	{topic: TopicType{0x00, 0x00, 0x00, 0x00}, data: []byte{}},
	{topic: TopicType{0x00, 0x00, 0x00, 0x00}, data: nil},
}

var unmarshalTestsGood = []struct {
	topic TopicType
	data  []byte
}{
	{topic: TopicType{0x00, 0x00, 0x00, 0x00}, data: []byte(`"0x00000000"`)},
	{topic: TopicType{0x00, 0x7f, 0x80, 0xff}, data: []byte(`"0x007f80ff"`)},
	{topic: TopicType{0xff, 0x80, 0x7f, 0x00}, data: []byte(`"0xff807f00"`)},
	{topic: TopicType{0xf2, 0x6e, 0x77, 0x79}, data: []byte(`"0xf26e7779"`)},
}

var unmarshalTestsBad = []struct {
	topic TopicType
	data  []byte
}{
	{topic: TopicType{0x00, 0x00, 0x00, 0x00}, data: []byte(`"0x000000"`)},
	{topic: TopicType{0x00, 0x00, 0x00, 0x00}, data: []byte(`"0x0000000"`)},
	{topic: TopicType{0x00, 0x00, 0x00, 0x00}, data: []byte(`"0x000000000"`)},
	{topic: TopicType{0x00, 0x00, 0x00, 0x00}, data: []byte(`"0x0000000000"`)},
	{topic: TopicType{0x00, 0x00, 0x00, 0x00}, data: []byte(`"000000"`)},
	{topic: TopicType{0x00, 0x00, 0x00, 0x00}, data: []byte(`"0000000"`)},
	{topic: TopicType{0x00, 0x00, 0x00, 0x00}, data: []byte(`"000000000"`)},
	{topic: TopicType{0x00, 0x00, 0x00, 0x00}, data: []byte(`"0000000000"`)},
	{topic: TopicType{0x00, 0x00, 0x00, 0x00}, data: []byte(`"abcdefg0"`)},
}

var unmarshalTestsUgly = []struct {
	topic TopicType
	data  []byte
}{
	{topic: TopicType{0x01, 0x00, 0x00, 0x00}, data: []byte(`"0x00000001"`)},
}

func TestBytesToTopic(t *testing.T) {
	for i, tst := range bytesToTopicTests {
		top := BytesToTopic(tst.data)
		if top != tst.topic {
			t.Fatalf("failed test %d: have %v, want %v.", i, t, tst.topic)
		}
	}
}

func TestUnmarshalTestsGood(t *testing.T) {
	for i, tst := range unmarshalTestsGood {
		var top TopicType
		err := json.Unmarshal(tst.data, &top)
		if err != nil {
			t.Errorf("failed test %d. input: %v. err: %v", i, tst.data, err)
		} else if top != tst.topic {
			t.Errorf("failed test %d: have %v, want %v.", i, t, tst.topic)
		}
	}
}

func TestUnmarshalTestsBad(t *testing.T) {
	// in this test UnmarshalJSON() is supposed to fail
	for i, tst := range unmarshalTestsBad {
		var top TopicType
		err := json.Unmarshal(tst.data, &top)
		if err == nil {
			t.Fatalf("failed test %d. input: %v.", i, tst.data)
		}
	}
}

func TestUnmarshalTestsUgly(t *testing.T) {
	// in this test UnmarshalJSON() is NOT supposed to fail, but result should be wrong
	for i, tst := range unmarshalTestsUgly {
		var top TopicType
		err := json.Unmarshal(tst.data, &top)
		if err != nil {
			t.Errorf("failed test %d. input: %v.", i, tst.data)
		} else if top == tst.topic {
			t.Errorf("failed test %d: have %v, want %v.", i, top, tst.topic)
		}
	}
}
