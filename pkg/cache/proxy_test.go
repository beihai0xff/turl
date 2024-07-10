package cache

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/beihai0xff/turl/internal/tests"
	"github.com/beihai0xff/turl/internal/tests/mocks"
)

func TestProxySet(t *testing.T) {
	p, err := NewProxy(tests.GlobalConfig.Cache)
	require.NoError(t, err)

	ctx := context.Background()
	k, v, ttl := "key", []byte("value"), time.Minute
	require.NoError(t, p.Set(ctx, k, v, ttl))
}

func TestProxyGet(t *testing.T) {
	p, err := newProxy(tests.GlobalConfig.Cache)
	require.NoError(t, err)

	ctx := context.Background()
	k, v, ttl := "key_get", []byte("value"), time.Minute
	require.NoError(t, p.Set(ctx, k, v, ttl))
	got, err := p.Get(ctx, k)
	require.NoError(t, err)
	require.Equal(t, v, got)

	// test cache miss
	got, err = p.Get(ctx, "empty")
	require.ErrorIs(t, err, ErrCacheMiss)
	require.Nil(t, got)

	// test remote cache exists but local cache miss
	k = "key_get2"
	require.NoError(t, p.distributedCache.Set(ctx, k, v, ttl))
	got, err = p.Get(ctx, k)
	require.NoError(t, err)
	require.Equal(t, v, got)

	got, err = p.localCache.Get(ctx, k)
	require.NoError(t, err)
	require.Equal(t, v, got)
}

func TestProxyDel(t *testing.T) {
	c, err := newProxy(tests.GlobalConfig.Cache)
	require.NoError(t, err)

	k, v := "key", []byte("value")

	ctx := context.Background()
	t.Run("del", func(t *testing.T) {
		require.NoError(t, c.Set(ctx, k, v, 10*time.Minute))

		require.NoError(t, c.Del(ctx, k))

		got, err := c.Get(ctx, k)
		require.ErrorIs(t, err, ErrCacheMiss)
		require.Nil(t, got)
	})

	t.Run("del_not_exist", func(t *testing.T) {
		require.NoError(t, c.Del(ctx, "not_exist"), ErrCacheMiss)
	})

	lc, rc := mocks.NewMockCache(t), mocks.NewMockCache(t)
	c = &proxy{
		localCache:       lc,
		distributedCache: rc,
		remoteCacheTTL:   time.Minute,
		localCacheTTL:    time.Minute,
	}

	testError := errors.New("test error")
	t.Run("del_remote_cache_error", func(t *testing.T) {
		rc.EXPECT().Del(ctx, k).Return(testError).Times(1)

		require.ErrorIs(t, c.Del(ctx, k), testError)
	})

	t.Run("del_local_cache_error", func(t *testing.T) {
		rc.EXPECT().Del(ctx, k).Return(nil).Times(1)
		lc.EXPECT().Del(ctx, k).Return(testError).Times(1)

		require.ErrorIs(t, c.Del(ctx, k), testError)
	})
}

func TestProxyClose(t *testing.T) {
	p, err := newProxy(tests.GlobalConfig.Cache)
	require.NoError(t, err)

	require.NoError(t, p.Close())
}
