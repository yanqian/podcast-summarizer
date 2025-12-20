package lock

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type ValkeyLock struct {
	client *redis.Client
	ttl    time.Duration
}

func NewValkeyLock(client *redis.Client, ttl time.Duration) *ValkeyLock {
	return &ValkeyLock{client: client, ttl: ttl}
}

// Acquire returns true if the lock was acquired.
func (l *ValkeyLock) Acquire(ctx context.Context, key string) (bool, error) {
	ok, err := l.client.SetNX(ctx, key, "locked", l.ttl).Result()
	return ok, err
}

func (l *ValkeyLock) Release(ctx context.Context, key string) error {
	return l.client.Del(ctx, key).Err()
}
