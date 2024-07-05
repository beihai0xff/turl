// Package workqueue provides a simple queue that used to rate limit or retry processing of requests.
// This package learn from https://github.com/kubernetes/client-go/blob/master/util/workqueue,
// you can find more RateLimiter implementation in the original package.
// rate_limiters.go provides rate limiters implementation for workqueue.
package workqueue

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"
)

const (
	tokenFormat     = "{%s}.tokens" //nolint: gosec
	timestampFormat = "{%s}.ts"
	pingInterval    = time.Millisecond * 100
)

// RateLimiter is an interface that knows how to limit the rate at which something is processed
// It provides methods to decide how long an item should wait, to stop tracking an item, and to get the number of failures an item has had.
type RateLimiter[T comparable] interface {
	// Take gets an item and gets to decide whether it should run now or not,
	// use this method if you intend to drop / skip events that exceed the rate.
	Take(ctx context.Context, item T) bool
	// When gets an item and gets to decide how long that item should wait,
	// use this method if you wish to wait and slow down in accordance with the rate limit without dropping events.
	When(ctx context.Context, item T) time.Duration
	// Forget indicates that an item is finished being retried. Doesn't matter whether it's for failing
	// or for success, we'll stop tracking it
	Forget(ctx context.Context, item T)
	// Retries returns back how many failures the item has had
	Retries(ctx context.Context, item T) int
}

// BucketRateLimiter adapts a standard bucket to the RateLimiter API
type BucketRateLimiter[T comparable] struct {
	*rate.Limiter
}

var _ RateLimiter[any] = &BucketRateLimiter[any]{}

// NewBucketRateLimiter creates a new BucketRateLimiter
func NewBucketRateLimiter[T comparable](l *rate.Limiter) RateLimiter[T] {
	return &BucketRateLimiter[T]{Limiter: l}
}

// Take gets an item and gets to decide whether it should run now or not,
func (r *BucketRateLimiter[T]) Take(_ context.Context, _ T) bool {
	return r.Limiter.Allow()
}

// When returns the delay for a reservation for a token from the bucket.
func (r *BucketRateLimiter[T]) When(_ context.Context, _ T) time.Duration {
	return r.Limiter.Reserve().Delay()
}

// Retries returns 0 as the number of retries for the bucket rate limiter.
func (r *BucketRateLimiter[T]) Retries(_ context.Context, _ T) int {
	return 0
}

// Forget is a no-op for the bucket rate limiter.
func (r *BucketRateLimiter[T]) Forget(_ context.Context, _ T) {
}

// ItemExponentialFailureRateLimiter does a simple baseDelay*2^<num-failures> limit
// dealing with max failures and expiration are up to the caller
type ItemExponentialFailureRateLimiter[T comparable] struct {
	failuresLock sync.Mutex
	failures     map[T]int

	baseDelay time.Duration
	maxDelay  time.Duration
}

var _ RateLimiter[any] = &ItemExponentialFailureRateLimiter[any]{}

// NewItemExponentialFailureRateLimiter creates a new ItemExponentialFailureRateLimiter with the specified base and max delays.
func NewItemExponentialFailureRateLimiter[T comparable](baseDelay, maxDelay time.Duration) RateLimiter[T] {
	return &ItemExponentialFailureRateLimiter[T]{
		failures:  map[T]int{},
		baseDelay: baseDelay,
		maxDelay:  maxDelay,
	}
}

// Take gets an item and gets to decide whether it should run now or not,
func (r *ItemExponentialFailureRateLimiter[T]) Take(_ context.Context, item T) bool {
	r.failuresLock.Lock()
	defer r.failuresLock.Unlock()

	_, ok := r.failures[item]

	return !ok
}

// When calculates the delay for an item based on the exponential backoff algorithm.
func (r *ItemExponentialFailureRateLimiter[T]) When(_ context.Context, item T) time.Duration {
	r.failuresLock.Lock()
	defer r.failuresLock.Unlock()

	exp := r.failures[item]
	r.failures[item]++

	// The backoff is capped such that 'calculated' value never overflows.
	backoff := float64(r.baseDelay.Nanoseconds()) * math.Pow(2, float64(exp)) //nolint: mnd
	if backoff > math.MaxInt64 {
		return r.maxDelay
	}

	calculated := time.Duration(backoff)
	if calculated > r.maxDelay {
		return r.maxDelay
	}

	return calculated
}

// Retries returns the number of times an item has been requeued.
func (r *ItemExponentialFailureRateLimiter[T]) Retries(_ context.Context, item T) int {
	r.failuresLock.Lock()
	defer r.failuresLock.Unlock()

	return r.failures[item]
}

// Forget removes an item from the failure map.
func (r *ItemExponentialFailureRateLimiter[T]) Forget(_ context.Context, item T) {
	r.failuresLock.Lock()
	defer r.failuresLock.Unlock()

	delete(r.failures, item)
}

// ItemFastSlowRateLimiter does a quick retry for a certain number of attempts, then a slow retry after that
type ItemFastSlowRateLimiter[T comparable] struct {
	failuresLock sync.Mutex
	failures     map[T]int

	maxFastAttempts int
	fastDelay       time.Duration
	slowDelay       time.Duration
}

var _ RateLimiter[any] = &ItemFastSlowRateLimiter[any]{}

// NewItemFastSlowRateLimiter creates a new ItemFastSlowRateLimiter with the specified fast and slow delays and the maximum number of fast attempts.
func NewItemFastSlowRateLimiter[T comparable](fastDelay, slowDelay time.Duration, maxFastAttempts int) RateLimiter[T] {
	return &ItemFastSlowRateLimiter[T]{
		failures:        map[T]int{},
		fastDelay:       fastDelay,
		slowDelay:       slowDelay,
		maxFastAttempts: maxFastAttempts,
	}
}

// Take gets an item and gets to decide whether it should run now or not,
func (r *ItemFastSlowRateLimiter[T]) Take(_ context.Context, item T) bool {
	r.failuresLock.Lock()
	defer r.failuresLock.Unlock()

	_, ok := r.failures[item]

	return !ok
}

// When calculates the delay for an item based on whether it has exceeded the maximum number of fast attempts.
func (r *ItemFastSlowRateLimiter[T]) When(_ context.Context, item T) time.Duration {
	r.failuresLock.Lock()
	defer r.failuresLock.Unlock()

	r.failures[item]++

	if r.failures[item] <= r.maxFastAttempts {
		return r.fastDelay
	}

	return r.slowDelay
}

// Retries returns the number of times an item has been requeued.
func (r *ItemFastSlowRateLimiter[T]) Retries(_ context.Context, item T) int {
	r.failuresLock.Lock()
	defer r.failuresLock.Unlock()

	return r.failures[item]
}

// Forget removes an item from the failure map.
func (r *ItemFastSlowRateLimiter[T]) Forget(_ context.Context, item T) {
	r.failuresLock.Lock()
	defer r.failuresLock.Unlock()

	delete(r.failures, item)
}

// MaxOfRateLimiter calls every RateLimiter and returns the worst case response
// When used with a token bucket limiter, the capacity could be apparently exceeded in cases where particular items
// were separately delayed a longer time.
type MaxOfRateLimiter[T comparable] struct {
	limiters []RateLimiter[T]
}

var _ RateLimiter[any] = &MaxOfRateLimiter[any]{}

// NewMaxOfRateLimiter creates a new MaxOfRateLimiter with the specified limiters.
func NewMaxOfRateLimiter[T comparable](limiters ...RateLimiter[T]) RateLimiter[T] {
	return &MaxOfRateLimiter[T]{limiters: limiters}
}

// Take gets an item and gets to decide whether it should run now or not,
func (r *MaxOfRateLimiter[T]) Take(ctx context.Context, item T) bool {
	for _, limiter := range r.limiters {
		if !limiter.Take(ctx, item) {
			return false
		}
	}

	return true
}

// When calculates the maximum delay among all the limiters for an item.
func (r *MaxOfRateLimiter[T]) When(ctx context.Context, item T) time.Duration {
	ret := time.Duration(0)

	for _, limiter := range r.limiters {
		curr := limiter.When(ctx, item)
		if curr > ret {
			ret = curr
		}
	}

	return ret
}

// Retries returns the maximum number of retries among all the limiters for an item.
func (r *MaxOfRateLimiter[T]) Retries(ctx context.Context, item T) int {
	ret := 0

	for _, limiter := range r.limiters {
		curr := limiter.Retries(ctx, item)
		if curr > ret {
			ret = curr
		}
	}

	return ret
}

// Forget calls the Forget method on all the limiters for an item.
func (r *MaxOfRateLimiter[T]) Forget(ctx context.Context, item T) {
	for _, limiter := range r.limiters {
		limiter.Forget(ctx, item)
	}
}

// ItemRedisTokenRateLimiter is a rate limiter that uses a token bucket in redis to rate limit items
type ItemRedisTokenRateLimiter[T comparable] struct {
	rate     int
	capacity int
	tokenKey string
	tsKey    string

	rdb        redis.UniversalClient
	redisAlive *atomic.Bool

	maxDelay      time.Duration
	rescueLock    sync.Mutex
	rescueLimiter RateLimiter[any]
}

var _ RateLimiter[any] = &ItemRedisTokenRateLimiter[any]{}

// NewItemRedisTokenRateLimiter creates a new ItemRedisTokenRateLimiter
func NewItemRedisTokenRateLimiter[T comparable](rdb redis.UniversalClient, key string, r, b int, maxDelay time.Duration) *ItemRedisTokenRateLimiter[T] {
	alive := atomic.Bool{}
	alive.Store(true)

	return &ItemRedisTokenRateLimiter[T]{
		rate:     r,
		capacity: b,
		tokenKey: fmt.Sprintf(tokenFormat, key),
		tsKey:    fmt.Sprintf(timestampFormat, key),

		rdb:           rdb,
		redisAlive:    &alive,
		maxDelay:      maxDelay,
		rescueLimiter: NewBucketRateLimiter[any](rate.NewLimiter(rate.Limit(r), b)),
	}
}

// Take gets an item and gets to decide whether it should run now or not
func (r *ItemRedisTokenRateLimiter[T]) Take(ctx context.Context, item T) bool {
	return r.reserveN(ctx, item)
}

// When calculates the delay for an item based on the token bucket in redis
func (r *ItemRedisTokenRateLimiter[T]) When(ctx context.Context, item T) time.Duration {
	if r.reserveN(ctx, item) {
		return 0
	}

	return r.maxDelay
}

// Forget removes an item from the token bucket in redis
func (r *ItemRedisTokenRateLimiter[T]) Forget(ctx context.Context, _ T) {
	r.rdb.Del(ctx, r.tokenKey, r.tsKey)
}

// Retries returns 0 as the number of retries for the token bucket in redis
func (r *ItemRedisTokenRateLimiter[T]) Retries(_ context.Context, _ T) int {
	return 0
}

func (r *ItemRedisTokenRateLimiter[T]) reserveN(ctx context.Context, item T) bool {
	if !r.redisAlive.Load() {
		slog.Warn("redis is not alive, use in-process limiter for rescue")
		return r.rescueLimiter.Take(ctx, item)
	}

	ok, err := allowN.Run(ctx, r.rdb,
		[]string{r.tokenKey, r.tsKey},
		[]string{
			strconv.Itoa(r.rate),
			strconv.Itoa(r.capacity),
			strconv.FormatInt(time.Now().UnixMilli(), 10),
			"1",
		}).Bool()

	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		slog.Error("fail to use rate limiter", slog.Any("error", err))
		return false
	}

	if err != nil {
		if !errors.Is(err, redis.Nil) { // redis.Nil is a normal error, it means the key is not exist
			slog.Error("fail to use rate limiter, use in-process limiter for rescue", slog.Any("error", err))

			go r.waitForRedis()

			return r.rescueLimiter.Take(ctx, item)
		}
	}

	// redis allowed == true
	return ok
}

// waitForRedis start a goroutine to check redis connection status until it's alive
func (r *ItemRedisTokenRateLimiter[T]) waitForRedis() {
	if !r.rescueLock.TryLock() {
		return
	}

	r.redisAlive.Store(false)

	ticker := time.NewTicker(pingInterval)

	defer func() {
		ticker.Stop()
		r.rescueLock.Unlock()
	}()

	for range ticker.C {
		if r.rdb.Ping(context.Background()).String() == "pong" {
			r.redisAlive.Store(true)
			return
		}
	}
}
