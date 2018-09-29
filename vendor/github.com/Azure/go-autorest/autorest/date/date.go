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
/*
Package date provides time.Time derivatives that conform to the Swagger.io (https://swagger.io/)
defined date   formats: Date and DateTime. Both types may, in most cases, be used in lieu of
time.Time types. And both convert to time.Time through a ToTime method.
*/
package date

import (
	"fmt"
	"time"
)

const (
	fullDate     = "2006-01-02"
	fullDateJSON = `"2006-01-02"`
	dateFormat   = "%04d-%02d-%02d"
	jsonFormat   = `"%04d-%02d-%02d"`
)

// Date defines a type similar to time.Time but assumes a layout of RFC3339 full-date (i.e.,
// 2006-01-02).
type Date struct {
	time.Time
}

// ParseDate create a new Date from the passed string.
func ParseDate(date string) (d Date, err error) {
	return parseDate(date, fullDate)
}

func parseDate(date string, format string) (Date, error) {
	d, err := time.Parse(format, date)
	return Date{Time: d}, err
}

// MarshalBinary preserves the Date as a byte array conforming to RFC3339 full-date (i.e.,
// 2006-01-02).
func (d Date) MarshalBinary() ([]byte, error) {
	return d.MarshalText()
}

// UnmarshalBinary reconstitutes a Date saved as a byte array conforming to RFC3339 full-date (i.e.,
// 2006-01-02).
func (d *Date) UnmarshalBinary(data []byte) error {
	return d.UnmarshalText(data)
}

// MarshalJSON preserves the Date as a JSON string conforming to RFC3339 full-date (i.e.,
// 2006-01-02).
func (d Date) MarshalJSON() (json []byte, err error) {
	return []byte(fmt.Sprintf(jsonFormat, d.Year(), d.Month(), d.Day())), nil
}

// UnmarshalJSON reconstitutes the Date from a JSON string conforming to RFC3339 full-date (i.e.,
// 2006-01-02).
func (d *Date) UnmarshalJSON(data []byte) (err error) {
	d.Time, err = time.Parse(fullDateJSON, string(data))
	return err
}

// MarshalText preserves the Date as a byte array conforming to RFC3339 full-date (i.e.,
// 2006-01-02).
func (d Date) MarshalText() (text []byte, err error) {
	return []byte(fmt.Sprintf(dateFormat, d.Year(), d.Month(), d.Day())), nil
}

// UnmarshalText reconstitutes a Date saved as a byte array conforming to RFC3339 full-date (i.e.,
// 2006-01-02).
func (d *Date) UnmarshalText(data []byte) (err error) {
	d.Time, err = time.Parse(fullDate, string(data))
	return err
}

// String returns the Date formatted as an RFC3339 full-date string (i.e., 2006-01-02).
func (d Date) String() string {
	return fmt.Sprintf(dateFormat, d.Year(), d.Month(), d.Day())
}

// ToTime returns a Date as a time.Time
func (d Date) ToTime() time.Time {
	return d.Time
}
