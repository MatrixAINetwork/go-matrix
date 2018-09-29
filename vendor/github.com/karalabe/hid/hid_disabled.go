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
// hid - Gopher Interface Devices (USB HID)
// Copyright (c) 2017 Péter Szilágyi. All rights reserved.
//
// This file is released under the 3-clause BSD license. Note however that Linux
// support depends on libusb, released under GNU LGPL 2.1 or later.

// +build !linux,!darwin,!windows ios !cgo

package hid

// Supported returns whether this platform is supported by the HID library or not.
// The goal of this method is to allow programatically handling platforms that do
// not support USB HID and not having to fall back to build constraints.
func Supported() bool {
	return false
}

// Enumerate returns a list of all the HID devices attached to the system which
// match the vendor and product id. On platforms that this file implements the
// function is a noop and returns an empty list always.
func Enumerate(vendorID uint16, productID uint16) []DeviceInfo {
	return nil
}

// Device is a live HID USB connected device handle. On platforms that this file
// implements the type lacks the actual HID device and all methods are noop.
type Device struct {
	DeviceInfo // Embed the infos for easier access
}

// Open connects to an HID device by its path name. On platforms that this file
// implements the method just returns an error.
func (info DeviceInfo) Open() (*Device, error) {
	return nil, ErrUnsupportedPlatform
}

// Close releases the HID USB device handle. On platforms that this file implements
// the method is just a noop.
func (dev *Device) Close() {}

// Write sends an output report to a HID device. On platforms that this file
// implements the method just returns an error.
func (dev *Device) Write(b []byte) (int, error) {
	return 0, ErrUnsupportedPlatform
}

// Read retrieves an input report from a HID device. On platforms that this file
// implements the method just returns an error.
func (dev *Device) Read(b []byte) (int, error) {
	return 0, ErrUnsupportedPlatform
}
