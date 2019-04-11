// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package p2p

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"sync/atomic"
	"time"

	"github.com/MatrixAINetwork/go-matrix/ca"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/event"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/p2p/discover"
	"github.com/MatrixAINetwork/go-matrix/rlp"
	"github.com/pkg/errors"
)

// Msg defines the structure of a p2p message.
//
// Note that a Msg can only be sent once since the Payload reader is
// consumed during sending. It is not possible to create a Msg and
// send it any number of times. If you want to reuse an encoded
// structure, encode the payload into a byte array and create a
// separate Msg with a bytes.Reader as Payload for each send.
type Msg struct {
	Code       uint64
	Size       uint32 // size of the paylod
	Payload    io.Reader
	ReceivedAt time.Time
}

// Decode parses the RLP content of a message into
// the given value, which must be a pointer.
//
// For the decoding rules, please see package rlp.
func (msg Msg) Decode(val interface{}) error {
	s := rlp.NewStream(msg.Payload, uint64(msg.Size))
	if err := s.Decode(val); err != nil {
		return newPeerError(errInvalidMsg, "(code %x) (size %d) %v", msg.Code, msg.Size, err)
	}
	return nil
}

func (msg Msg) String() string {
	return fmt.Sprintf("msg #%v (%v bytes)", msg.Code, msg.Size)
}

// Discard reads any remaining payload data into a black hole.
func (msg Msg) Discard() error {
	_, err := io.Copy(ioutil.Discard, msg.Payload)
	return err
}

type MsgReader interface {
	ReadMsg() (Msg, error)
}

type MsgWriter interface {
	// WriteMsg sends a message. It will block until the message's
	// Payload has been consumed by the other end.
	//
	// Note that messages can be sent only once because their
	// payload reader is drained.
	WriteMsg(Msg) error
}

// MsgReadWriter provides reading and writing of encoded messages.
// Implementations should ensure that ReadMsg and WriteMsg can be
// called simultaneously from multiple goroutines.
type MsgReadWriter interface {
	MsgReader
	MsgWriter
}

var (
	ErrCanNotFindPeer = errors.New("p2p: can`t find peer")
	ErrMsgWriterIsNil = errors.New("p2p: message writer is nil")
	ErrCanNotConvert  = errors.New("p2p: can`t convert addr to id")
)

// Send writes an RLP-encoded message with the given code.
// data should encode as an RLP list.
func Send(w MsgWriter, msgcode uint64, data interface{}) error {
	if w == nil {
		return ErrMsgWriterIsNil
	}
	size, r, err := rlp.EncodeToReader(data)
	if err != nil {
		return err
	}
	return w.WriteMsg(Msg{Code: msgcode, Size: uint32(size), Payload: r})
}

// SendToSingle send message to single peer.
func SendToSingle(addr common.Address, msgCode uint64, data interface{}) error {
	if addr == ServerP2p.ManAddress {
		return nil
	}
	id := ServerP2p.ConvertAddressToId(addr)
	if id == EmptyNodeId {
		log.Error("send to single peer failed, id convert failed", "peer addr", addr)
		return ErrCanNotConvert
	}

	peers := ServerP2p.Peers()
	for _, peer := range peers {
		if id == peer.ID() {
			return Send(peer.MsgReadWriter(), msgCode, data)
		}
	}
	return ErrCanNotFindPeer
}

// SendToGroup send message to a group.
func SendToGroupWithBackup(to common.RoleType, msgCode uint64, data interface{}) error {
	address := ca.GetRolesByGroupWithNextElect(to)
	peers := ServerP2p.Peers()
	for _, addr := range address {
		if addr == ServerP2p.ManAddress {
			continue
		}
		id := ServerP2p.ConvertAddressToId(addr)
		if id == EmptyNodeId {
			log.Error("send to single peer failed, id convert failed", "peer addr", addr)
			continue
		}

		for _, peer := range peers {
			if id == peer.ID() {
				err := Send(peer.MsgReadWriter(), msgCode, data)
				if err != nil {
					log.Error("send to group with backup", "error:", err)
				}
				break
			}
		}
	}

	return nil
}

// SendToGroup send message to a group.
func SendToGroup(to common.RoleType, msgCode uint64, data interface{}) error {
	address := ca.GetRolesByGroup(to)
	peers := ServerP2p.Peers()
	log.Trace("message.go", "查看所有的 ServerP2P peers Count", len(peers), "目标IDS数量", len(address), "role", to.String())
	for _, addr := range address {
		if addr == ServerP2p.ManAddress {
			continue
		}
		bSend := false

		id := ServerP2p.ConvertAddressToId(addr)
		if id == EmptyNodeId {
			log.Error("send to single peer failed, id convert failed", "peer addr", addr)
			continue
		}
		for _, peer := range peers {
			if id == peer.ID() {
				err := Send(peer.MsgReadWriter(), msgCode, data)
				if err != nil {
					log.Trace("message.go", "发送消息失败, id", id, "addr", addr.Hex(), "err", err)
				} else {
					log.Trace("message.go", "发送消息成功, id", id, "addr", addr.Hex(), "IP", peer.Info().Network.RemoteAddress)
					bSend = true
				}
				break
			}
		}
		if !bSend {
			log.ERROR("message.go", "该节点未发送成功 nodeId", id)
		}
	}
	return nil
}

// SendItems writes an RLP with the given code and data elements.
// For a call such as:
//
//    SendItems(w, code, e1, e2, e3)
//
// the message payload will be an RLP list containing the items:
//
//    [e1, e2, e3]
//
func SendItems(w MsgWriter, msgcode uint64, elems ...interface{}) error {
	return Send(w, msgcode, elems)
}

// eofSignal wraps a reader with eof signaling. the eof channel is
// closed when the wrapped reader returns an error or when count bytes
// have been read.
type eofSignal struct {
	wrapped io.Reader
	count   uint32 // number of bytes left
	eof     chan<- struct{}
}

// note: when using eofSignal to detect whether a message payload
// has been read, Read might not be called for zero sized messages.
func (r *eofSignal) Read(buf []byte) (int, error) {
	if r.count == 0 {
		if r.eof != nil {
			r.eof <- struct{}{}
			r.eof = nil
		}
		return 0, io.EOF
	}

	max := len(buf)
	if int(r.count) < len(buf) {
		max = int(r.count)
	}
	n, err := r.wrapped.Read(buf[:max])
	r.count -= uint32(n)
	if (err != nil || r.count == 0) && r.eof != nil {
		r.eof <- struct{}{} // tell Peer that msg has been consumed
		r.eof = nil
	}
	return n, err
}

// MsgPipe creates a message pipe. Reads on one end are matched
// with writes on the other. The pipe is full-duplex, both ends
// implement MsgReadWriter.
func MsgPipe() (*MsgPipeRW, *MsgPipeRW) {
	var (
		c1, c2  = make(chan Msg), make(chan Msg)
		closing = make(chan struct{})
		closed  = new(int32)
		rw1     = &MsgPipeRW{c1, c2, closing, closed}
		rw2     = &MsgPipeRW{c2, c1, closing, closed}
	)
	return rw1, rw2
}

// ErrPipeClosed is returned from pipe operations after the
// pipe has been closed.
var ErrPipeClosed = errors.New("p2p: read or write on closed message pipe")

// MsgPipeRW is an endpoint of a MsgReadWriter pipe.
type MsgPipeRW struct {
	w       chan<- Msg
	r       <-chan Msg
	closing chan struct{}
	closed  *int32
}

// WriteMsg sends a messsage on the pipe.
// It blocks until the receiver has consumed the message payload.
func (p *MsgPipeRW) WriteMsg(msg Msg) error {
	if atomic.LoadInt32(p.closed) == 0 {
		consumed := make(chan struct{}, 1)
		msg.Payload = &eofSignal{msg.Payload, msg.Size, consumed}
		select {
		case p.w <- msg:
			if msg.Size > 0 {
				// wait for payload read or discard
				select {
				case <-consumed:
				case <-p.closing:
				}
			}
			return nil
		case <-p.closing:
		}
	}
	return ErrPipeClosed
}

// ReadMsg returns a message sent on the other end of the pipe.
func (p *MsgPipeRW) ReadMsg() (Msg, error) {
	if atomic.LoadInt32(p.closed) == 0 {
		select {
		case msg := <-p.r:
			return msg, nil
		case <-p.closing:
		}
	}
	return Msg{}, ErrPipeClosed
}

// Close unblocks any pending ReadMsg and WriteMsg calls on both ends
// of the pipe. They will return ErrPipeClosed. Close also
// interrupts any reads from a message payload.
func (p *MsgPipeRW) Close() error {
	if atomic.AddInt32(p.closed, 1) != 1 {
		// someone else is already closing
		atomic.StoreInt32(p.closed, 1) // avoid overflow
		return nil
	}
	close(p.closing)
	return nil
}

// ExpectMsg reads a message from r and verifies that its
// code and encoded RLP content match the provided values.
// If content is nil, the payload is discarded and not verified.
func ExpectMsg(r MsgReader, code uint64, content interface{}) error {
	msg, err := r.ReadMsg()
	if err != nil {
		return err
	}
	if msg.Code != code {
		return fmt.Errorf("message code mismatch: got %d, expected %d", msg.Code, code)
	}
	if content == nil {
		return msg.Discard()
	}
	contentEnc, err := rlp.EncodeToBytes(content)
	if err != nil {
		panic("content encode error: " + err.Error())
	}
	if int(msg.Size) != len(contentEnc) {
		return fmt.Errorf("message size mismatch: got %d, want %d", msg.Size, len(contentEnc))
	}
	actualContent, err := ioutil.ReadAll(msg.Payload)
	if err != nil {
		return err
	}
	if !bytes.Equal(actualContent, contentEnc) {
		return fmt.Errorf("message payload mismatch:\ngot:  %x\nwant: %x", actualContent, contentEnc)
	}
	return nil
}

// msgEventer wraps a MsgReadWriter and sends events whenever a message is sent
// or received
type msgEventer struct {
	MsgReadWriter

	feed     *event.Feed
	peerID   discover.NodeID
	Protocol string
}

// newMsgEventer returns a msgEventer which sends message events to the given
// feed
func newMsgEventer(rw MsgReadWriter, feed *event.Feed, peerID discover.NodeID, proto string) *msgEventer {
	return &msgEventer{
		MsgReadWriter: rw,
		feed:          feed,
		peerID:        peerID,
		Protocol:      proto,
	}
}

// ReadMsg reads a message from the underlying MsgReadWriter and emits a
// "message received" event
func (ev *msgEventer) ReadMsg() (Msg, error) {
	msg, err := ev.MsgReadWriter.ReadMsg()
	if err != nil {
		return msg, err
	}
	ev.feed.Send(&PeerEvent{
		Type:     PeerEventTypeMsgRecv,
		Peer:     ev.peerID,
		Protocol: ev.Protocol,
		MsgCode:  &msg.Code,
		MsgSize:  &msg.Size,
	})
	return msg, nil
}

// WriteMsg writes a message to the underlying MsgReadWriter and emits a
// "message sent" event
func (ev *msgEventer) WriteMsg(msg Msg) error {
	err := ev.MsgReadWriter.WriteMsg(msg)
	if err != nil {
		return err
	}
	ev.feed.Send(&PeerEvent{
		Type:     PeerEventTypeMsgSend,
		Peer:     ev.peerID,
		Protocol: ev.Protocol,
		MsgCode:  &msg.Code,
		MsgSize:  &msg.Size,
	})
	return nil
}

// Close closes the underlying MsgReadWriter if it implements the io.Closer
// interface
func (ev *msgEventer) Close() error {
	if v, ok := ev.MsgReadWriter.(io.Closer); ok {
		return v.Close()
	}
	return nil
}
