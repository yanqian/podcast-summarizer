package cache

import (
	"context"
	"sync"
	"time"
)

type memoryCache struct {
	mu    sync.Mutex
	store map[string][]byte
}

func NewInMemoryCache() *memoryCache {
	return &memoryCache{store: make(map[string][]byte)}
}

func (c *memoryCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store[key] = value
	return nil
}

func (c *memoryCache) Get(ctx context.Context, key string) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.store[key], nil
}

func (c *memoryCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.store, key)
	return nil
}
