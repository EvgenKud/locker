package memory

import (
	"sync"
	"time"
)

type config interface {
	TTL() time.Duration
}

type locker struct {
	cfg     config
	storage *storage
}

type storage struct {
	sync.Mutex
	blocked map[string]*block
}

type block struct {
	waitList  map[string]*waited
	timerChan chan struct{}
}

type waited struct {
	waitChan  chan struct{}
	closeChan chan struct{}
}
