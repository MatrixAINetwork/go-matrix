package manparams

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/log"
	"sync"
	"time"
)

type EntrustValue struct {
	mu           sync.RWMutex
	entrustValue map[common.Address]string
}

func newEntrustValue() *EntrustValue {
	return &EntrustValue{
		entrustValue: make(map[common.Address]string, 0),
	}
}

var (
	EntrustAccountValue = newEntrustValue()
)

func (self *EntrustValue) SetEntrustValue(data map[common.Address]string) {
	self.mu.RLock()
	defer self.mu.RUnlock()
	self.entrustValue = data
}
func (self *EntrustValue) GetEntrustValue() map[common.Address]string {
	self.mu.RLock()
	defer self.mu.RUnlock()
	return self.entrustValue
}
func SetTimer(times int64) {
	if times <= 0 {
		return
	}
	d := time.Duration(times) * time.Second
	t := time.NewTimer(d)
	defer t.Stop()
	select {

	case <-t.C:
		EntrustAccountValue.entrustValue = make(map[common.Address]string, 0)
	}
	log.Warn("修改委托签名账户", "数据已失效 ,上次设置的有效期", times)
}
