package lcahce

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_newLocalCache_failed(t *testing.T) {
	// invalid cap
	c, err := newLocalCache(-1)
	require.Error(t, err)
	require.Nil(t, c)

	c, err = newLocalCache(0)
	require.Error(t, err)
	require.Nil(t, c)
}

func Test_newLocalCache_success(t *testing.T) {
	// new cache failed
	c, err := newLocalCache(1e6)
	require.NoError(t, err)
	t.Cleanup(func() {
		c.Close()
	})

	k, v := "key", []byte("value")
	require.NoError(t, c.Set(context.Background(), k, v, 10*time.Minute))
	got, err := c.Get(context.Background(), k)
	require.NoError(t, err)
	require.Equal(t, v, got)
}
