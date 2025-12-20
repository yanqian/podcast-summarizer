package lock

import (
	"context"
	"sync"
	"time"
)

type memoryLock struct {
	mu    sync.Mutex
	locks map[string]time.Time
}

func NewInMemoryLock() *memoryLock {
	return &memoryLock{locks: make(map[string]time.Time)}
}

func (l *memoryLock) Acquire(ctx context.Context, key string) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if _, ok := l.locks[key]; ok {
		return false, nil
	}
	l.locks[key] = time.Now().Add(5 * time.Minute)
	return true, nil
}

func (l *memoryLock) Release(ctx context.Context, key string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.locks, key)
	return nil
}
