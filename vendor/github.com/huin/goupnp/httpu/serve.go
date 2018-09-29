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
package httpu

import (
	"bufio"
	"bytes"
	"log"
	"net"
	"net/http"
	"regexp"
)

const (
	DefaultMaxMessageBytes = 2048
)

var (
	trailingWhitespaceRx = regexp.MustCompile(" +\r\n")
	crlf                 = []byte("\r\n")
)

// Handler is the interface by which received HTTPU messages are passed to
// handling code.
type Handler interface {
	// ServeMessage is called for each HTTPU message received. peerAddr contains
	// the address that the message was received from.
	ServeMessage(r *http.Request)
}

// HandlerFunc is a function-to-Handler adapter.
type HandlerFunc func(r *http.Request)

func (f HandlerFunc) ServeMessage(r *http.Request) {
	f(r)
}

// A Server defines parameters for running an HTTPU server.
type Server struct {
	Addr            string         // UDP address to listen on
	Multicast       bool           // Should listen for multicast?
	Interface       *net.Interface // Network interface to listen on for multicast, nil for default multicast interface
	Handler         Handler        // handler to invoke
	MaxMessageBytes int            // maximum number of bytes to read from a packet, DefaultMaxMessageBytes if 0
}

// ListenAndServe listens on the UDP network address srv.Addr. If srv.Multicast
// is true, then a multicast UDP listener will be used on srv.Interface (or
// default interface if nil).
func (srv *Server) ListenAndServe() error {
	var err error

	var addr *net.UDPAddr
	if addr, err = net.ResolveUDPAddr("udp", srv.Addr); err != nil {
		log.Fatal(err)
	}

	var conn net.PacketConn
	if srv.Multicast {
		if conn, err = net.ListenMulticastUDP("udp", srv.Interface, addr); err != nil {
			return err
		}
	} else {
		if conn, err = net.ListenUDP("udp", addr); err != nil {
			return err
		}
	}

	return srv.Serve(conn)
}

// Serve messages received on the given packet listener to the srv.Handler.
func (srv *Server) Serve(l net.PacketConn) error {
	maxMessageBytes := DefaultMaxMessageBytes
	if srv.MaxMessageBytes != 0 {
		maxMessageBytes = srv.MaxMessageBytes
	}
	for {
		buf := make([]byte, maxMessageBytes)
		n, peerAddr, err := l.ReadFrom(buf)
		if err != nil {
			return err
		}
		buf = buf[:n]

		go func(buf []byte, peerAddr net.Addr) {
			// At least one router's UPnP implementation has added a trailing space
			// after "HTTP/1.1" - trim it.
			buf = trailingWhitespaceRx.ReplaceAllLiteral(buf, crlf)

			req, err := http.ReadRequest(bufio.NewReader(bytes.NewBuffer(buf)))
			if err != nil {
				log.Printf("httpu: Failed to parse request: %v", err)
				return
			}
			req.RemoteAddr = peerAddr.String()
			srv.Handler.ServeMessage(req)
			// No need to call req.Body.Close - underlying reader is bytes.Buffer.
		}(buf, peerAddr)
	}
}

// Serve messages received on the given packet listener to the given handler.
func Serve(l net.PacketConn, handler Handler) error {
	srv := Server{
		Handler:         handler,
		MaxMessageBytes: DefaultMaxMessageBytes,
	}
	return srv.Serve(l)
}
