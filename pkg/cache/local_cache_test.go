package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	c, err := NewLocalCache(1e6, 10*time.Minute)
	require.NoError(t, err)
	t.Cleanup(func() {
		c.Close()
	})
	require.NotNil(t, c)
}

func Test_newLocalCache_failed(t *testing.T) {
	// invalid cap
	c, err := newLocalCache(-1, 10*time.Minute)
	require.Error(t, err)
	require.Nil(t, c)

	c, err = newLocalCache(0, 10*time.Minute)
	require.Error(t, err)
	require.Nil(t, c)
}

func Test_newLocalCache_success(t *testing.T) {
	// new cache failed
	c, err := newLocalCache(1e6, 10*time.Minute)
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
	c, err := newLocalCache(1e6, 10*time.Minute)
	require.NoError(t, err)
	t.Cleanup(func() {
		c.Close()
	})

	k, v := "key", []byte("value")

	require.NoError(t, c.Set(context.Background(), k, v, 10*time.Minute))
}

func Test_localCache_Get(t *testing.T) {
	c, err := newLocalCache(1e6, 10*time.Minute)
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
