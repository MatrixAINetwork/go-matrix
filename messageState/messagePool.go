// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package messageState

import (
	"errors"
	"github.com/MatrixAINetwork/go-matrix/event"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/rlp"
	"reflect"
	"time"
)

var (
	errBadChannel    = errors.New("event: Subscribe argument does not have sendable channel type")
	errMultiSubcribe = errors.New("event: MessageState must Subscribe once")
	fnv_prime        = uint64(1099511628211)
	offset_basis     = uint64(14695981039346656037)
)

//send into state channel
type MessageSend struct {
	Message []map[uint64]interface{}
	Index   int
	Round   uint64
}
type SubScribeInfo struct {
	Round   uint64
	Index   int
	Message interface{}
}

//MessgeStateInterface is an interface to gather All messages and send to state channel
/*
  	1.  SubscribeEvent all correlative Event
	2.  SetStateChan set the channel for wait
	3.  SetMessageChecker check all the event messages
	4.  RunLoop
	5.  You can selece the stateChan and it will send MessageSend struct every Round if Round state is full
*/

type MessgeStateInterface interface {
	SubscribeEvent(aim mc.EventCode, ch interface{}) error
	SetStateChan(stateChan chan MessageSend)
	SetMessageChecker(checker MessageChecker)
	Require(RequireInfo)
	RunLoop()
	Quit()
}

// fnvHash mixes in data into mix using the manash fnv method.
func fnvHash(data []byte) uint64 {
	hash := offset_basis
	if len(data) >= 8 {
		nLen := len(data) - 7
		for i := 0; i < nLen; i++ {
			b := data[i:]
			hash ^= uint64(b[7]) | uint64(b[6])<<8 | uint64(b[5])<<16 | uint64(b[4])<<24 |
				uint64(b[3])<<32 | uint64(b[2])<<40 | uint64(b[1])<<48 | uint64(b[0])<<56
			hash *= fnv_prime
		}
	} else {
		b := [8]byte{}
		copy(b[8-len(data):], data)
		hash ^= uint64(b[7]) | uint64(b[6])<<8 | uint64(b[5])<<16 | uint64(b[4])<<24 |
			uint64(b[3])<<32 | uint64(b[2])<<40 | uint64(b[1])<<48 | uint64(b[0])<<56
		hash *= fnv_prime
	}
	return hash
}
func RlpFnvHash(data interface{}) uint64 {
	val, err := rlp.EncodeToBytes(data)
	if err != nil {
		log.Error("rlpFnvHash rlp encode error", "error", err)
		return 0
	}
	return fnvHash(val)
}

//MessageChecker is an interface to verify message value and decode Round Number
type MessageChecker interface {
	checkMessage(aim mc.EventCode, value interface{}) (uint64, bool)
	getKeyBytes(value interface{}) []byte
	checkState(state []byte, round uint64) bool
}

//subscribe from messageCenter package
type RequireInfo struct {
	index       int
	key         []byte
	requireChan chan []interface{}
}
type subscribeInfo struct {
	code      mc.EventCode
	ch        interface{}
	subscribe event.Subscription
}

//gather Messages and set states
type stateInfo struct {
	message []map[uint64]interface{}
	state   []byte
}

func newstateInfo(nLen int) *stateInfo {
	state := stateInfo{make([]map[uint64]interface{}, nLen), make([]byte, nLen)}
	for i := 0; i < nLen; i++ {
		state.message[i] = make(map[uint64]interface{})
	}
	return &state
}

func (si *stateInfo) addMessage(index int, key uint64, data interface{}) {
	si.message[index][key] = data
	si.state[index] = 1
}

//an simple Message state process
type MessageStatePool struct {
	message               map[uint64]*stateInfo
	subscription          []subscribeInfo
	chanList              []reflect.SelectCase
	checker               MessageChecker
	quit                  chan struct{}
	stateChan             chan MessageSend
	require               chan RequireInfo
	capacity, msgCapacity int
}

func NewMessageStatePool(roundCapacity, messageCapacity int, checker MessageChecker) *MessageStatePool {
	return &MessageStatePool{
		message:     make(map[uint64]*stateInfo),
		quit:        make(chan struct{}, 2),
		require:     make(chan RequireInfo, 2),
		capacity:    roundCapacity,
		msgCapacity: messageCapacity,
		checker:     checker,
	}
}
func (ms *MessageStatePool) SetStateChan(stateChan chan MessageSend) {
	ms.stateChan = stateChan
}

func (ms *MessageStatePool) SetMessageChecker(checker MessageChecker) {
	ms.checker = checker
}
func (ms *MessageStatePool) Quit() {
	ms.quit <- struct{}{}
}
func (ms *MessageStatePool) Require(key RequireInfo) {
	ms.require <- key
}
func (ms *MessageStatePool) getSubscribeInfo(aim mc.EventCode) (subscribeInfo, int) {
	for i, subInfo := range ms.subscription {
		if subInfo.code == aim {
			return subInfo, i
		}
	}
	return subscribeInfo{}, -1
}
func (ms *MessageStatePool) SubscribeEvent(aim mc.EventCode, ch interface{}) error {
	_, index := ms.getSubscribeInfo(aim)
	if index >= 0 {
		log.Error("MessageStateProcess SubscribeEvent", "error", errMultiSubcribe)
		return errMultiSubcribe
	}
	chanval := reflect.ValueOf(ch)
	chantyp := chanval.Type()
	if chantyp.Kind() != reflect.Chan {
		panic(errBadChannel)
	}
	sub, err := mc.SubscribeEvent(aim, ch)
	if err == nil {
		ms.subscription = append(ms.subscription, subscribeInfo{aim, ch, sub})
	} else {
		log.Error("MessageStateProcess SubscribeEvent", "error", err)
		return err
	}
	cas := reflect.SelectCase{Dir: reflect.SelectRecv, Chan: chanval}
	ms.chanList = append(ms.chanList, cas)
	return nil
}
func (ms *MessageStatePool) unSubscribeEvents() {
	for _, subInfo := range ms.subscription {
		subInfo.subscribe.Unsubscribe()
	}
	ms.subscription = make([]subscribeInfo, 0)
	ms.chanList = make([]reflect.SelectCase, 0)
	ms.message = make(map[uint64]*stateInfo)
}
func (ms *MessageStatePool) RunLoop() {
	defer ms.unSubscribeEvents()
	caseList := append(ms.chanList, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ms.quit)})
	caseList = append(caseList, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ms.require)})
	tm := time.NewTicker(time.Second)
	caseList = append(caseList, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(tm.C)})
	nLen := len(ms.chanList)
	for {
		chosen, recv, _ := reflect.Select(caseList)
		if chosen < nLen {
			aim := ms.subscription[chosen].code
			if num, check := ms.checker.checkMessage(aim, recv.Interface()); check {
				ms.setMessage(num, chosen, recv.Interface())
			}
		} else if chosen == nLen {
			break
		} else if chosen == nLen+1 {
			ms.requireMessage(recv.Interface().(RequireInfo))
		} else {
			ms.deleteRound()
		}
	}
}
func (ms *MessageStatePool) requireMessage(key RequireInfo) {
	var req []interface{}
	find := fnvHash(key.key)
	for _, item := range ms.message {
		msg := item.message[key.index]
		if value, exist := msg[find]; exist {
			req = append(req, value)
		}
	}
	key.requireChan <- req
}
func (ms *MessageStatePool) setMessage(round uint64, index int, value interface{}) {
	if _, exist := ms.message[round]; !exist {
		nLen := len(ms.subscription)
		ms.message[round] = newstateInfo(nLen)
	}
	msg := ms.message[round]
	msg.addMessage(index, fnvHash(ms.checker.getKeyBytes(value)), value)
	if ms.checker.checkState(msg.state, round) {
		ms.stateChan <- MessageSend{msg.message, index, round}
	}
}

func (ms *MessageStatePool) deleteRound() {
	//todo
	/*if len(ms.message) > ms.capacity {
		var keys []uint64
		for round, _ := range ms.message {
			keys = append(keys, round)
		}
		sortkeys.Uint64s(keys)
		keys = keys[:len(keys)-ms.capacity]
		for _, key := range keys {
			delete(ms.message, key)
		}
		log.Info("deleteRound", "keys", keys)
	}*/
}
