package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/beihai0xff/turl/internal/tests"
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

func TestProxyClose(t *testing.T) {
	p, err := newProxy(tests.GlobalConfig.Cache)
	require.NoError(t, err)

	require.NoError(t, p.Close())
}
