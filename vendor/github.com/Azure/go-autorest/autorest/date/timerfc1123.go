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
package date

import (
	"errors"
	"time"
)

const (
	rfc1123JSON = `"` + time.RFC1123 + `"`
	rfc1123     = time.RFC1123
)

// TimeRFC1123 defines a type similar to time.Time but assumes a layout of RFC1123 date-time (i.e.,
// Mon, 02 Jan 2006 15:04:05 MST).
type TimeRFC1123 struct {
	time.Time
}

// UnmarshalJSON reconstitutes the Time from a JSON string conforming to RFC1123 date-time
// (i.e., Mon, 02 Jan 2006 15:04:05 MST).
func (t *TimeRFC1123) UnmarshalJSON(data []byte) (err error) {
	t.Time, err = ParseTime(rfc1123JSON, string(data))
	if err != nil {
		return err
	}
	return nil
}

// MarshalJSON preserves the Time as a JSON string conforming to RFC1123 date-time (i.e.,
// Mon, 02 Jan 2006 15:04:05 MST).
func (t TimeRFC1123) MarshalJSON() ([]byte, error) {
	if y := t.Year(); y < 0 || y >= 10000 {
		return nil, errors.New("Time.MarshalJSON: year outside of range [0,9999]")
	}
	b := []byte(t.Format(rfc1123JSON))
	return b, nil
}

// MarshalText preserves the Time as a byte array conforming to RFC1123 date-time (i.e.,
// Mon, 02 Jan 2006 15:04:05 MST).
func (t TimeRFC1123) MarshalText() ([]byte, error) {
	if y := t.Year(); y < 0 || y >= 10000 {
		return nil, errors.New("Time.MarshalText: year outside of range [0,9999]")
	}

	b := []byte(t.Format(rfc1123))
	return b, nil
}

// UnmarshalText reconstitutes a Time saved as a byte array conforming to RFC1123 date-time
// (i.e., Mon, 02 Jan 2006 15:04:05 MST).
func (t *TimeRFC1123) UnmarshalText(data []byte) (err error) {
	t.Time, err = ParseTime(rfc1123, string(data))
	if err != nil {
		return err
	}
	return nil
}

// MarshalBinary preserves the Time as a byte array conforming to RFC1123 date-time (i.e.,
// Mon, 02 Jan 2006 15:04:05 MST).
func (t TimeRFC1123) MarshalBinary() ([]byte, error) {
	return t.MarshalText()
}

// UnmarshalBinary reconstitutes a Time saved as a byte array conforming to RFC1123 date-time
// (i.e., Mon, 02 Jan 2006 15:04:05 MST).
func (t *TimeRFC1123) UnmarshalBinary(data []byte) error {
	return t.UnmarshalText(data)
}

// ToTime returns a Time as a time.Time
func (t TimeRFC1123) ToTime() time.Time {
	return t.Time
}

// String returns the Time formatted as an RFC1123 date-time string (i.e.,
// Mon, 02 Jan 2006 15:04:05 MST).
func (t TimeRFC1123) String() string {
	// Note: time.Time.String does not return an RFC1123 compliant string, time.Time.MarshalText does.
	b, err := t.MarshalText()
	if err != nil {
		return ""
	}
	return string(b)
}
