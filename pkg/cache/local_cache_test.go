package cache

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/beihai0xff/turl/configs"
	"github.com/beihai0xff/turl/internal/tests"
)

func TestNew(t *testing.T) {
	c, err := NewLocalCache(tests.GlobalConfig.CacheConfig.LocalCacheConfig)
	require.NoError(t, err)
	t.Cleanup(func() {
		c.Close()
	})
	require.NotNil(t, c)
}

func Test_newLocalCache_failed(t *testing.T) {
	// invalid cap
	c, err := newLocalCache(&configs.LocalCacheConfig{Capacity: 0})
	require.Error(t, err)
	require.Nil(t, c)

	c, err = newLocalCache(&configs.LocalCacheConfig{Capacity: -1})
	require.Error(t, err)
	require.Nil(t, c)
}

func Test_newLocalCache_success(t *testing.T) {
	// new cache failed
	c, err := newLocalCache(tests.GlobalConfig.CacheConfig.LocalCacheConfig)
	require.NoError(t, err)
	t.Cleanup(func() {
		c.Close()
	})

	k, v := "key", []byte("value")
	require.NoError(t, c.Set(context.Background(), k, v, 0))
	got, err := c.Get(context.Background(), k)
	require.NoError(t, err)
	require.Equal(t, v, got)
}

func Test_localCache_Set(t *testing.T) {
	c, err := newLocalCache(tests.GlobalConfig.CacheConfig.LocalCacheConfig)
	require.NoError(t, err)
	t.Cleanup(func() {
		c.Close()
	})

	k, v := "key", []byte("value")

	require.NoError(t, c.Set(context.Background(), k, v, 10*time.Minute))
}

func Test_localCache_Get(t *testing.T) {
	c, err := newLocalCache(tests.GlobalConfig.CacheConfig.LocalCacheConfig)
	require.NoError(t, err)
	t.Cleanup(func() {
		c.Close()
	})

	k, v := "key", []byte("value")

	require.NoError(t, c.Set(context.Background(), k, v, 10*time.Minute))
	got, err := c.Get(context.Background(), k)
	require.NoError(t, err)
	require.Equal(t, v, got)

	got, err = c.Get(context.Background(), "empty_get")
	require.ErrorIs(t, err, ErrCacheMiss)
	require.Nil(t, got)
}

func Test_localCache_Get_Large(t *testing.T) {
	c, err := newLocalCache(tests.GlobalConfig.CacheConfig.LocalCacheConfig)
	require.NoError(t, err)
	t.Cleanup(func() {
		c.Close()
	})

	v := []byte("value")
	var nums int = 1e6
	for i := range nums {
		require.NoError(t, c.Set(context.Background(), strconv.Itoa(i), v, 10*time.Minute))
	}

	for i := range nums {
		got, err := c.Get(context.Background(), strconv.Itoa(i))
		require.NoError(t, err)
		require.Equal(t, v, got)
	}
}
