package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type ValkeyCache struct {
	client *redis.Client
}

func NewValkeyCache(client *redis.Client) *ValkeyCache {
	return &ValkeyCache{client: client}
}

func (c *ValkeyCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}

func (c *ValkeyCache) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return val, err
}

func (c *ValkeyCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}
