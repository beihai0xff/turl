// Package dcache implements a distributed cache Interface
// redis.go implements a distributed cache with redis
package dcache

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/beiai0xff/turl/pkg/cache"
)

var _ cache.Interface = (*redisCache)(nil)

func newRedisClient(addr []string) redis.UniversalClient {
	return redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs: addr,
	})
}

type redisCache struct {
	rdb redis.UniversalClient
	// bucketSize int
}

func NewRedisCache(addr []string) cache.Interface {
	return newRedisCache(addr)
}

func newRedisCache(addr []string) *redisCache {
	return &redisCache{
		rdb: newRedisClient(addr),
	}
}

func (c *redisCache) Set(ctx context.Context, k string, v []byte, ttl time.Duration) error {
	return c.rdb.SetEx(ctx, k, v, ttl).Err()
}

func (c *redisCache) Get(ctx context.Context, k string) ([]byte, error) {
	value, err := c.rdb.Get(ctx, k).Bytes()
	if err != nil && errors.Is(err, redis.Nil) {
		return nil, cache.ErrCacheMiss
	}

	return value, err
}

func (c *redisCache) Close() error {
	return c.rdb.Close()
}
