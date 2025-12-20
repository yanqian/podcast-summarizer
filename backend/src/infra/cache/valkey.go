package cache

import (
	"context"
	"os"

	"github.com/redis/go-redis/v9"
)

func NewValkeyClient() *redis.Client {
	url := os.Getenv("VALKEY_URL")
	opt, err := redis.ParseURL(url)
	if err != nil {
		return redis.NewClient(&redis.Options{Addr: url})
	}
	return redis.NewClient(opt)
}

func PingValkey(ctx context.Context, client *redis.Client) error {
	return client.Ping(ctx).Err()
}
