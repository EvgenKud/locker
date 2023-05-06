package memory

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

func New(cfg config) *locker {
	l := &locker{
		cfg: cfg,
		storage: &storage{
			blocked: make(map[string]*block),
		},
	}
	return l
}

func (l *locker) Lock(ctx context.Context, id string) (func(), error) {
	l.storage.Lock()
	if found, ok := l.storage.blocked[id]; ok {
		wChan := make(chan struct{})
		cChan := make(chan struct{})
		waitID := fmt.Sprintf("%s:%s", id, uuid.NewString())
		found.waitList[waitID] = &waited{
			waitChan:  wChan,
			closeChan: cChan,
		}
		l.storage.blocked[id] = found
		l.storage.Unlock()
		select {
		case <-ctx.Done():
			cChan <- struct{}{}
			return nil, ctx.Err()
		case <-wChan:
			return func() {
				l.unlock(id)
			}, nil
		}
	}

	l.storage.blocked[id] = &block{
		waitList:  map[string]*waited{},
		timerChan: l.startTTLTimer(id),
	}
	l.storage.Unlock()

	return func() {
		l.unlock(id)
	}, nil
}

func (l *locker) unlock(id string) {
	l.storage.Lock()
	defer l.storage.Unlock()
	if found, ok := l.storage.blocked[id]; ok {
		found.timerChan <- struct{}{} // close ttl timer
		for waitID, wat := range found.waitList {
			select {
			case <-wat.closeChan:
				deleteWaitItem(wat, found, waitID)
				continue
			default:
				wat.waitChan <- struct{}{}
				deleteWaitItem(wat, found, waitID)
				found.timerChan = l.startTTLTimer(id)
				l.storage.blocked[id] = found
				return
			}
		}
		if len(found.waitList) == 0 {
			delete(l.storage.blocked, id)
		}
	}
}

func (l *locker) startTTLTimer(id string) chan struct{} {
	t := time.NewTicker(l.cfg.TTL())
	defer t.Stop()
	tChan := make(chan struct{})
	go func() {
		select {
		case <-t.C:
			l.unlock(id)
			return
		case <-tChan:
			return
		}
	}()
	return tChan
}

func deleteWaitItem(wat *waited, found *block, waitID string) {
	close(wat.closeChan)
	close(wat.waitChan)
	delete(found.waitList, waitID)
}
