// Package dcache implements a distributed cache Interface
// redis.go implements a distributed cache with redis
package cache

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/beihai0xff/turl/configs"
	redis2 "github.com/beihai0xff/turl/pkg/db/redis"
)

var _ Interface = (*redisCache)(nil)

type redisCache struct {
	rdb redis.UniversalClient
	// bucketSize int
}

// NewRedisRemoteCache returns a new redis cache
func NewRedisRemoteCache(c *configs.RedisConfig) Interface {
	return newRedisCache(c)
}

func newRedisCache(c *configs.RedisConfig) *redisCache {
	return &redisCache{
		rdb: redis2.Client(c),
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
		return nil, ErrCacheMiss
	}

	return value, err
}

// Close the cache
func (c *redisCache) Close() error {
	return c.rdb.Close()
}
