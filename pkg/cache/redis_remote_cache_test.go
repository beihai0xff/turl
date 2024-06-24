package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/beihai0xff/turl/internal/tests"
)

func TestNewRedisCache(t *testing.T) {
	got := NewRedisRemoteCache(tests.GlobalConfig.CacheConfig.RedisConfig)
	require.NotNil(t, got)
}

func Test_newRedisCache(t *testing.T) {
	got := NewRedisRemoteCache(tests.GlobalConfig.CacheConfig.RedisConfig)
	require.NotNil(t, got)
}

func Test_redisCache_Set(t *testing.T) {
	c := newRedisCache(tests.GlobalConfig.CacheConfig.RedisConfig)
	t.Cleanup(
		func() {
			c.Close()
		})

	ctx := context.Background()
	k, v, ttl := "key", []byte("value"), time.Minute
	require.NoError(t, c.Set(ctx, k, v, ttl))
}

func Test_redisCache_Get(t *testing.T) {
	c := newRedisCache(tests.GlobalConfig.CacheConfig.RedisConfig)
	t.Cleanup(
		func() {
			c.Close()
		})

	ctx := context.Background()
	k, v, ttl := "key_get", []byte("value"), time.Minute
	require.NoError(t, c.Set(ctx, k, v, ttl))
	got, err := c.Get(ctx, k)
	require.NoError(t, err)
	require.Equal(t, v, got)

	// test cache miss
	got, err = c.Get(ctx, "empty")
	require.ErrorIs(t, err, ErrCacheMiss)
	require.Nil(t, got)
}
