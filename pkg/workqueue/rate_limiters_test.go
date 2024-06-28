package workqueue

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"

	"github.com/beihai0xff/turl/configs"
	"github.com/beihai0xff/turl/internal/tests"
	"github.com/beihai0xff/turl/pkg/db/redis"
)

func TestBucketRateLimiter(t *testing.T) {
	limiter := NewBucketRateLimiter[any](rate.NewLimiter(rate.Limit(1), 1))
	ctx := context.Background()
	if e, a := 0*time.Second, limiter.When(ctx, "one"); e != a {
		t.Errorf("expected %v, got %v", e, a)
	}

	if limiter.Take(ctx, "one") {
		t.Errorf("expected false, got true")
	}

	if e, a := 0, limiter.Retries(ctx, "one"); e != a {
		t.Errorf("expected %v, got %v", e, a)
	}

	limiter.Forget(ctx, "one")
	if e, a := 0, limiter.Retries(ctx, "one"); e != a {
		t.Errorf("expected %v, got %v", e, a)
	}

}

func TestItemExponentialFailureRateLimiter(t *testing.T) {
	limiter := NewItemExponentialFailureRateLimiter[any](1*time.Millisecond, 1*time.Second)
	ctx := context.Background()

	if e, a := 1*time.Millisecond, limiter.When(ctx, "one"); e != a {
		t.Errorf("expected %v, got %v", e, a)
	}
	if e, a := 2*time.Millisecond, limiter.When(ctx, "one"); e != a {
		t.Errorf("expected %v, got %v", e, a)
	}
	if e, a := 4*time.Millisecond, limiter.When(ctx, "one"); e != a {
		t.Errorf("expected %v, got %v", e, a)
	}
	if e, a := 8*time.Millisecond, limiter.When(ctx, "one"); e != a {
		t.Errorf("expected %v, got %v", e, a)
	}
	if e, a := 16*time.Millisecond, limiter.When(ctx, "one"); e != a {
		t.Errorf("expected %v, got %v", e, a)
	}
	if e, a := 5, limiter.Retries(ctx, "one"); e != a {
		t.Errorf("expected %v, got %v", e, a)
	}

	if limiter.Take(ctx, "one") {
		t.Errorf("expected false, got true")
	}
	if !limiter.Take(ctx, "two") {
		t.Errorf("expected true, got false")
	}

	if e, a := 1*time.Millisecond, limiter.When(ctx, "two"); e != a {
		t.Errorf("expected %v, got %v", e, a)
	}
	if e, a := 2*time.Millisecond, limiter.When(ctx, "two"); e != a {
		t.Errorf("expected %v, got %v", e, a)
	}
	if e, a := 2, limiter.Retries(ctx, "two"); e != a {
		t.Errorf("expected %v, got %v", e, a)
	}

	limiter.Forget(ctx, "one")
	if e, a := 0, limiter.Retries(ctx, "one"); e != a {
		t.Errorf("expected %v, got %v", e, a)
	}
	if e, a := 1*time.Millisecond, limiter.When(ctx, "one"); e != a {
		t.Errorf("expected %v, got %v", e, a)
	}

}

func TestItemExponentialFailureRateLimiterOverFlow(t *testing.T) {
	limiter := NewItemExponentialFailureRateLimiter[any](1*time.Millisecond, 1000*time.Second)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		limiter.When(ctx, "one")
	}
	if e, a := 32*time.Millisecond, limiter.When(ctx, "one"); e != a {
		t.Errorf("expected %v, got %v", e, a)
	}

	for i := 0; i < 1000; i++ {
		limiter.When(ctx, "overflow1")
	}
	if e, a := 1000*time.Second, limiter.When(ctx, "overflow1"); e != a {
		t.Errorf("expected %v, got %v", e, a)
	}

	limiter = NewItemExponentialFailureRateLimiter[any](1*time.Minute, 1000*time.Hour)
	for i := 0; i < 2; i++ {
		limiter.When(ctx, "two")
	}
	if e, a := 4*time.Minute, limiter.When(ctx, "two"); e != a {
		t.Errorf("expected %v, got %v", e, a)
	}

	for i := 0; i < 1000; i++ {
		limiter.When(ctx, "overflow2")
	}
	if e, a := 1000*time.Hour, limiter.When(ctx, "overflow2"); e != a {
		t.Errorf("expected %v, got %v", e, a)
	}

}

func TestItemFastSlowRateLimiter(t *testing.T) {
	limiter := NewItemFastSlowRateLimiter[any](5*time.Millisecond, 10*time.Second, 3)
	ctx := context.Background()

	if e, a := 5*time.Millisecond, limiter.When(ctx, "one"); e != a {
		t.Errorf("expected %v, got %v", e, a)
	}
	if e, a := 5*time.Millisecond, limiter.When(ctx, "one"); e != a {
		t.Errorf("expected %v, got %v", e, a)
	}
	if e, a := 5*time.Millisecond, limiter.When(ctx, "one"); e != a {
		t.Errorf("expected %v, got %v", e, a)
	}
	if e, a := 10*time.Second, limiter.When(ctx, "one"); e != a {
		t.Errorf("expected %v, got %v", e, a)
	}
	if e, a := 10*time.Second, limiter.When(ctx, "one"); e != a {
		t.Errorf("expected %v, got %v", e, a)
	}
	if e, a := 5, limiter.Retries(ctx, "one"); e != a {
		t.Errorf("expected %v, got %v", e, a)
	}

	if limiter.Take(ctx, "one") {
		t.Errorf("expected false, got true")
	}
	if !limiter.Take(ctx, "two") {
		t.Errorf("expected true, got false")
	}

	if e, a := 5*time.Millisecond, limiter.When(ctx, "two"); e != a {
		t.Errorf("expected %v, got %v", e, a)
	}
	if e, a := 5*time.Millisecond, limiter.When(ctx, "two"); e != a {
		t.Errorf("expected %v, got %v", e, a)
	}
	if e, a := 2, limiter.Retries(ctx, "two"); e != a {
		t.Errorf("expected %v, got %v", e, a)
	}

	limiter.Forget(ctx, "one")
	if e, a := 0, limiter.Retries(ctx, "one"); e != a {
		t.Errorf("expected %v, got %v", e, a)
	}
	if e, a := 5*time.Millisecond, limiter.When(ctx, "one"); e != a {
		t.Errorf("expected %v, got %v", e, a)
	}

}

func TestMaxOfRateLimiter(t *testing.T) {
	limiter := NewMaxOfRateLimiter(
		NewItemFastSlowRateLimiter[any](5*time.Millisecond, 3*time.Second, 3),
		NewItemExponentialFailureRateLimiter[any](1*time.Millisecond, 1*time.Second),
	)
	ctx := context.Background()

	if e, a := 5*time.Millisecond, limiter.When(ctx, "one"); e != a {
		t.Errorf("expected %v, got %v", e, a)
	}
	if e, a := 5*time.Millisecond, limiter.When(ctx, "one"); e != a {
		t.Errorf("expected %v, got %v", e, a)
	}
	if e, a := 5*time.Millisecond, limiter.When(ctx, "one"); e != a {
		t.Errorf("expected %v, got %v", e, a)
	}
	if e, a := 3*time.Second, limiter.When(ctx, "one"); e != a {
		t.Errorf("expected %v, got %v", e, a)
	}
	if e, a := 3*time.Second, limiter.When(ctx, "one"); e != a {
		t.Errorf("expected %v, got %v", e, a)
	}
	if e, a := 5, limiter.Retries(ctx, "one"); e != a {
		t.Errorf("expected %v, got %v", e, a)
	}

	if limiter.Take(ctx, "one") {
		t.Errorf("expected false, got true")
	}
	if !limiter.Take(ctx, "two") {
		t.Errorf("expected true, got false")
	}

	if e, a := 5*time.Millisecond, limiter.When(ctx, "two"); e != a {
		t.Errorf("expected %v, got %v", e, a)
	}
	if e, a := 5*time.Millisecond, limiter.When(ctx, "two"); e != a {
		t.Errorf("expected %v, got %v", e, a)
	}
	if e, a := 2, limiter.Retries(ctx, "two"); e != a {
		t.Errorf("expected %v, got %v", e, a)
	}

	limiter.Forget(ctx, "one")
	if e, a := 0, limiter.Retries(ctx, "one"); e != a {
		t.Errorf("expected %v, got %v", e, a)
	}
	if e, a := 5*time.Millisecond, limiter.When(ctx, "one"); e != a {
		t.Errorf("expected %v, got %v", e, a)
	}
}

func TestItemRedisTokenRateLimiter(t *testing.T) {
	rdb := redis.Client(tests.GlobalConfig.Cache.Redis)

	ctx := context.Background()

	t.Run("Take", func(t *testing.T) {
		r := NewItemRedisTokenRateLimiter[any](rdb, "test_Take", 1, 1, time.Second)
		require.True(t, r.Take(ctx, "one"))
		require.False(t, r.Take(ctx, "one"))
		require.False(t, r.Take(ctx, "one"))
		// create a new one to test the rate limit
		r = NewItemRedisTokenRateLimiter[any](rdb, "test_Take", 1, 1, time.Second)
		require.False(t, r.Take(ctx, "one"))
	})

	t.Run("When", func(t *testing.T) {
		r := NewItemRedisTokenRateLimiter[any](rdb, "test_When", 1, 1, time.Second)
		require.Zero(t, r.When(ctx, "one"))
		require.Equal(t, time.Second, r.When(ctx, "one"))
	})

	t.Run("Forget", func(t *testing.T) {
		r := NewItemRedisTokenRateLimiter[any](rdb, "test_Forget", 1, 1, time.Second)
		require.True(t, r.Take(ctx, "one"))
		r.Forget(ctx, "one")
		require.ErrorIs(t, rdb.Get(ctx, "test_Forget.tokens").Err(), redis.Nil)
	})

	t.Run("reserveN", func(t *testing.T) {
		r := NewItemRedisTokenRateLimiter[any](rdb, "test_reserveN", 1, 1, time.Second)
		require.True(t, r.reserveN(ctx, "one"))
		require.False(t, r.reserveN(ctx, "one"))
	})

	t.Run("reserveN_rdb_disconnect", func(t *testing.T) {
		rdb = redis.Client(&configs.RedisConfig{Addr: make([]string, 0)})
		r := NewItemRedisTokenRateLimiter[any](rdb, "test_reserveN_rdb_disconnect", 1, 1, time.Second)
		require.True(t, r.reserveN(ctx, "one"))
		require.False(t, r.reserveN(ctx, "one"))
	})
}
