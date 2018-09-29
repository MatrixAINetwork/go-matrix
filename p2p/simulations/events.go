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

package simulations

import (
	"fmt"
	"time"
)

// EventType is the type of event emitted by a simulation network
type EventType string

const (
	// EventTypeNode is the type of event emitted when a node is either
	// created, started or stopped
	EventTypeNode EventType = "node"

	// EventTypeConn is the type of event emitted when a connection is
	// is either established or dropped between two nodes
	EventTypeConn EventType = "conn"

	// EventTypeMsg is the type of event emitted when a p2p message it
	// sent between two nodes
	EventTypeMsg EventType = "msg"
)

// Event is an event emitted by a simulation network
type Event struct {
	// Type is the type of the event
	Type EventType `json:"type"`

	// Time is the time the event happened
	Time time.Time `json:"time"`

	// Control indicates whether the event is the result of a controlled
	// action in the network
	Control bool `json:"control"`

	// Node is set if the type is EventTypeNode
	Node *Node `json:"node,omitempty"`

	// Conn is set if the type is EventTypeConn
	Conn *Conn `json:"conn,omitempty"`

	// Msg is set if the type is EventTypeMsg
	Msg *Msg `json:"msg,omitempty"`
}

// NewEvent creates a new event for the given object which should be either a
// Node, Conn or Msg.
//
// The object is copied so that the event represents the state of the object
// when NewEvent is called.
func NewEvent(v interface{}) *Event {
	event := &Event{Time: time.Now()}
	switch v := v.(type) {
	case *Node:
		event.Type = EventTypeNode
		node := *v
		event.Node = &node
	case *Conn:
		event.Type = EventTypeConn
		conn := *v
		event.Conn = &conn
	case *Msg:
		event.Type = EventTypeMsg
		msg := *v
		event.Msg = &msg
	default:
		panic(fmt.Sprintf("invalid event type: %T", v))
	}
	return event
}

// ControlEvent creates a new control event
func ControlEvent(v interface{}) *Event {
	event := NewEvent(v)
	event.Control = true
	return event
}

// String returns the string representation of the event
func (e *Event) String() string {
	switch e.Type {
	case EventTypeNode:
		return fmt.Sprintf("<node-event> id: %s up: %t", e.Node.ID().TerminalString(), e.Node.Up)
	case EventTypeConn:
		return fmt.Sprintf("<conn-event> nodes: %s->%s up: %t", e.Conn.One.TerminalString(), e.Conn.Other.TerminalString(), e.Conn.Up)
	case EventTypeMsg:
		return fmt.Sprintf("<msg-event> nodes: %s->%s proto: %s, code: %d, received: %t", e.Msg.One.TerminalString(), e.Msg.Other.TerminalString(), e.Msg.Protocol, e.Msg.Code, e.Msg.Received)
	default:
		return ""
	}
}
