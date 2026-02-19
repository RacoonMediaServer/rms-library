package lock

import (
	"context"
	"sync"
	"time"

	"github.com/RacoonMediaServer/rms-library/v3/internal/model"
)

type Locker interface {
	Lock(id model.ID) Unlocker
	ContextLock(ctx context.Context, id model.ID) (Unlocker, error)
}

type Unlocker interface {
	Unlock()
}

type lock struct {
	mu     sync.Mutex
	ref    uint64
	locker *locker
	id     *model.ID
}

// Unlock implements Unlocker.
func (lck *lock) Unlock() {
	lck.locker.release(lck)
	lck.mu.Unlock()
}

type locker struct {
	mu sync.Mutex
	l  map[model.ID]*lock
}

func (l *locker) getOrCreate(id model.ID) *lock {
	l.mu.Lock()
	defer l.mu.Unlock()

	result, ok := l.l[id]
	if !ok {
		result = &lock{locker: l, id: &id}
		l.l[id] = result
	}
	result.ref++
	return result
}

// ContextLock implements Locker.
func (l *locker) ContextLock(ctx context.Context, id model.ID) (Unlocker, error) {
	itemLock := l.getOrCreate(id)
	if itemLock.mu.TryLock() {
		return itemLock, nil
	}

	for {
		select {
		case <-ctx.Done():
			l.release(itemLock)
			return nil, ctx.Err()
		case <-time.After(100 * time.Millisecond):
			if itemLock.mu.TryLock() {
				return itemLock, nil
			}
		}
	}
}

// Lock implements Locker.
func (l *locker) Lock(id model.ID) Unlocker {
	itemLock := l.getOrCreate(id)
	itemLock.mu.Lock()
	return itemLock
}

func (l *locker) release(lck *lock) {
	l.mu.Lock()
	defer l.mu.Unlock()

	lck.ref--
	if lck.ref == 0 {
		delete(l.l, *lck.id)
	}
}

func NewLocker() Locker {
	return &locker{
		l: map[model.ID]*lock{},
	}
}
