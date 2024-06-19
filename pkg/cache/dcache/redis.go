// Package dcache implements a distributed cache Interface
// redis.go implements a distributed cache with redis
package dcache

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/beiai0xff/turl/pkg/cache"
	redis2 "github.com/beiai0xff/turl/pkg/db/redis"
)

var _ cache.Interface = (*redisCache)(nil)

type redisCache struct {
	rdb redis.UniversalClient
	// bucketSize int
}

// NewRedis returns a new redis cache
func NewRedis(addr []string) cache.Interface {
	return newRedisCache(addr)
}

func newRedisCache(addr []string) *redisCache {
	return &redisCache{
		rdb: redis2.Client(addr),
	}
}

// Set the k v pair to the cache
func (c *redisCache) Set(ctx context.Context, k string, v []byte, ttl time.Duration) error {
	return c.rdb.SetEx(ctx, k, v, ttl).Err()
}

// Get the value by key
func (c *redisCache) Get(ctx context.Context, k string) ([]byte, error) {
	value, err := c.rdb.Get(ctx, k).Bytes()
	if err != nil && errors.Is(err, redis.Nil) {
		return nil, cache.ErrCacheMiss
	}

	return value, err
}

// Close the cache
func (c *redisCache) Close() error {
	return c.rdb.Close()
}
