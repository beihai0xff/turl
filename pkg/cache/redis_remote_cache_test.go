package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/beihai0xff/turl/internal/tests"
)

func TestNewRedisCache(t *testing.T) {
	got := NewRedisRemoteCache(tests.GlobalConfig.Cache.Redis)
	require.NotNil(t, got)
}

func Test_newRedisCache(t *testing.T) {
	got := NewRedisRemoteCache(tests.GlobalConfig.Cache.Redis)
	require.NotNil(t, got)
}

func Test_redisCache_Set(t *testing.T) {
	c := newRedisCache(tests.GlobalConfig.Cache.Redis)
	t.Cleanup(
		func() {
			c.Close()
		})

	ctx := context.Background()
	k, v, ttl := "key", []byte("value"), time.Minute
	require.NoError(t, c.Set(ctx, k, v, ttl))
}

func Test_redisCache_Get(t *testing.T) {
	c := newRedisCache(tests.GlobalConfig.Cache.Redis)
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

func Test_redisCache_Del(t *testing.T) {
	c := newRedisCache(tests.GlobalConfig.Cache.Redis)
	t.Cleanup(
		func() {
			c.Close()
		})

	k, v := "key", []byte("value")

	t.Run("del", func(t *testing.T) {
		require.NoError(t, c.Set(context.Background(), k, v, 10*time.Minute))

		require.NoError(t, c.Del(context.Background(), k))

		got, err := c.Get(context.Background(), k)
		require.ErrorIs(t, err, ErrCacheMiss)
		require.Nil(t, got)
	})

	t.Run("del_not_exist", func(t *testing.T) {
		require.NoError(t, c.Del(context.Background(), "not_exist"), ErrCacheMiss)
	})
}
