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
// +build !windows,!plan9

package log

import (
	"log/syslog"
	"strings"
)

// SyslogHandler opens a connection to the system syslog daemon by calling
// syslog.New and writes all records to it.
func SyslogHandler(priority syslog.Priority, tag string, fmtr Format) (Handler, error) {
	wr, err := syslog.New(priority, tag)
	return sharedSyslog(fmtr, wr, err)
}

// SyslogNetHandler opens a connection to a log daemon over the network and writes
// all log records to it.
func SyslogNetHandler(net, addr string, priority syslog.Priority, tag string, fmtr Format) (Handler, error) {
	wr, err := syslog.Dial(net, addr, priority, tag)
	return sharedSyslog(fmtr, wr, err)
}

func sharedSyslog(fmtr Format, sysWr *syslog.Writer, err error) (Handler, error) {
	if err != nil {
		return nil, err
	}
	h := FuncHandler(func(r *Record) error {
		var syslogFn = sysWr.Info
		switch r.Lvl {
		case LvlCrit:
			syslogFn = sysWr.Crit
		case LvlError:
			syslogFn = sysWr.Err
		case LvlWarn:
			syslogFn = sysWr.Warning
		case LvlInfo:
			syslogFn = sysWr.Info
		case LvlDebug:
			syslogFn = sysWr.Debug
		case LvlTrace:
			syslogFn = func(m string) error { return nil } // There's no syslog level for trace
		}

		s := strings.TrimSpace(string(fmtr.Format(r)))
		return syslogFn(s)
	})
	return LazyHandler(&closingHandler{sysWr, h}), nil
}

func (m muster) SyslogHandler(priority syslog.Priority, tag string, fmtr Format) Handler {
	return must(SyslogHandler(priority, tag, fmtr))
}

func (m muster) SyslogNetHandler(net, addr string, priority syslog.Priority, tag string, fmtr Format) Handler {
	return must(SyslogNetHandler(net, addr, priority, tag, fmtr))
}
