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

package les

import (
	"github.com/matrix/go-matrix/metrics"
	"github.com/matrix/go-matrix/p2p"
)

var (
	/*	propTxnInPacketsMeter     = metrics.NewMeter("man/prop/txns/in/packets")
		propTxnInTrafficMeter     = metrics.NewMeter("man/prop/txns/in/traffic")
		propTxnOutPacketsMeter    = metrics.NewMeter("man/prop/txns/out/packets")
		propTxnOutTrafficMeter    = metrics.NewMeter("man/prop/txns/out/traffic")
		propHashInPacketsMeter    = metrics.NewMeter("man/prop/hashes/in/packets")
		propHashInTrafficMeter    = metrics.NewMeter("man/prop/hashes/in/traffic")
		propHashOutPacketsMeter   = metrics.NewMeter("man/prop/hashes/out/packets")
		propHashOutTrafficMeter   = metrics.NewMeter("man/prop/hashes/out/traffic")
		propBlockInPacketsMeter   = metrics.NewMeter("man/prop/blocks/in/packets")
		propBlockInTrafficMeter   = metrics.NewMeter("man/prop/blocks/in/traffic")
		propBlockOutPacketsMeter  = metrics.NewMeter("man/prop/blocks/out/packets")
		propBlockOutTrafficMeter  = metrics.NewMeter("man/prop/blocks/out/traffic")
		reqHashInPacketsMeter     = metrics.NewMeter("man/req/hashes/in/packets")
		reqHashInTrafficMeter     = metrics.NewMeter("man/req/hashes/in/traffic")
		reqHashOutPacketsMeter    = metrics.NewMeter("man/req/hashes/out/packets")
		reqHashOutTrafficMeter    = metrics.NewMeter("man/req/hashes/out/traffic")
		reqBlockInPacketsMeter    = metrics.NewMeter("man/req/blocks/in/packets")
		reqBlockInTrafficMeter    = metrics.NewMeter("man/req/blocks/in/traffic")
		reqBlockOutPacketsMeter   = metrics.NewMeter("man/req/blocks/out/packets")
		reqBlockOutTrafficMeter   = metrics.NewMeter("man/req/blocks/out/traffic")
		reqHeaderInPacketsMeter   = metrics.NewMeter("man/req/headers/in/packets")
		reqHeaderInTrafficMeter   = metrics.NewMeter("man/req/headers/in/traffic")
		reqHeaderOutPacketsMeter  = metrics.NewMeter("man/req/headers/out/packets")
		reqHeaderOutTrafficMeter  = metrics.NewMeter("man/req/headers/out/traffic")
		reqBodyInPacketsMeter     = metrics.NewMeter("man/req/bodies/in/packets")
		reqBodyInTrafficMeter     = metrics.NewMeter("man/req/bodies/in/traffic")
		reqBodyOutPacketsMeter    = metrics.NewMeter("man/req/bodies/out/packets")
		reqBodyOutTrafficMeter    = metrics.NewMeter("man/req/bodies/out/traffic")
		reqStateInPacketsMeter    = metrics.NewMeter("man/req/states/in/packets")
		reqStateInTrafficMeter    = metrics.NewMeter("man/req/states/in/traffic")
		reqStateOutPacketsMeter   = metrics.NewMeter("man/req/states/out/packets")
		reqStateOutTrafficMeter   = metrics.NewMeter("man/req/states/out/traffic")
		reqReceiptInPacketsMeter  = metrics.NewMeter("man/req/receipts/in/packets")
		reqReceiptInTrafficMeter  = metrics.NewMeter("man/req/receipts/in/traffic")
		reqReceiptOutPacketsMeter = metrics.NewMeter("man/req/receipts/out/packets")
		reqReceiptOutTrafficMeter = metrics.NewMeter("man/req/receipts/out/traffic")*/
	miscInPacketsMeter  = metrics.NewRegisteredMeter("les/misc/in/packets", nil)
	miscInTrafficMeter  = metrics.NewRegisteredMeter("les/misc/in/traffic", nil)
	miscOutPacketsMeter = metrics.NewRegisteredMeter("les/misc/out/packets", nil)
	miscOutTrafficMeter = metrics.NewRegisteredMeter("les/misc/out/traffic", nil)
)

// meteredMsgReadWriter is a wrapper around a p2p.MsgReadWriter, capable of
// accumulating the above defined metrics based on the data stream contents.
type meteredMsgReadWriter struct {
	p2p.MsgReadWriter     // Wrapped message stream to meter
	version           int // Protocol version to select correct meters
}

// newMeteredMsgWriter wraps a p2p MsgReadWriter with metering support. If the
// metrics system is disabled, this function returns the original object.
func newMeteredMsgWriter(rw p2p.MsgReadWriter) p2p.MsgReadWriter {
	if !metrics.Enabled {
		return rw
	}
	return &meteredMsgReadWriter{MsgReadWriter: rw}
}

// Init sets the protocol version used by the stream to know which meters to
// increment in case of overlapping message ids between protocol versions.
func (rw *meteredMsgReadWriter) Init(version int) {
	rw.version = version
}

func (rw *meteredMsgReadWriter) ReadMsg() (p2p.Msg, error) {
	// Read the message and short circuit in case of an error
	msg, err := rw.MsgReadWriter.ReadMsg()
	if err != nil {
		return msg, err
	}
	// Account for the data traffic
	packets, traffic := miscInPacketsMeter, miscInTrafficMeter
	packets.Mark(1)
	traffic.Mark(int64(msg.Size))

	return msg, err
}

func (rw *meteredMsgReadWriter) WriteMsg(msg p2p.Msg) error {
	// Account for the data traffic
	packets, traffic := miscOutPacketsMeter, miscOutTrafficMeter
	packets.Mark(1)
	traffic.Mark(int64(msg.Size))

	// Send the packet to the p2p layer
	return rw.MsgReadWriter.WriteMsg(msg)
}
