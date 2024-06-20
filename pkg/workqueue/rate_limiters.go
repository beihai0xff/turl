// Package workqueue provides a simple queue
// rate_limiters.go provides rate limiters for workqueue
// this package learn from https://github.com/kubernetes/client-go/blob/master/util/workqueue
package workqueue

import (
	"math"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter is an interface that knows how to limit the rate at which something is processed
// It provides methods to decide how long an item should wait, to stop tracking an item, and to get the number of failures an item has had.
type RateLimiter[T comparable] interface {
	// When gets an item and gets to decide how long that item should wait
	When(item T) time.Duration
	// Forget indicates that an item is finished being retried.  Doesn't matter whether it's for failing
	// or for success, we'll stop tracking it
	Forget(item T)
	// Retries returns back how many failures the item has had
	Retries(item T) int
}

// BucketRateLimiter adapts a standard bucket to the workqueue ratelimiter API
type BucketRateLimiter[T comparable] struct {
	*rate.Limiter
}

var _ RateLimiter[any] = &BucketRateLimiter[any]{}

// NewBucketRateLimiter creates a new BucketRateLimiter
func NewBucketRateLimiter[T comparable](l *rate.Limiter) RateLimiter[T] {
	return &BucketRateLimiter[T]{Limiter: l}
}

// When returns the delay for a reservation for a token from the bucket.
func (r *BucketRateLimiter[T]) When(_ T) time.Duration {
	return r.Limiter.Reserve().Delay()
}

// Retries returns 0 as the number of retries for the bucket rate limiter.
func (r *BucketRateLimiter[T]) Retries(_ T) int {
	return 0
}

// Forget is a no-op for the bucket rate limiter.
func (r *BucketRateLimiter[T]) Forget(_ T) {
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
func NewItemExponentialFailureRateLimiter[T comparable](baseDelay time.Duration, maxDelay time.Duration) RateLimiter[T] {
	return &ItemExponentialFailureRateLimiter[T]{
		failures:  map[T]int{},
		baseDelay: baseDelay,
		maxDelay:  maxDelay,
	}
}

// When calculates the delay for an item based on the exponential backoff algorithm.
func (r *ItemExponentialFailureRateLimiter[T]) When(item T) time.Duration {
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
func (r *ItemExponentialFailureRateLimiter[T]) Retries(item T) int {
	r.failuresLock.Lock()
	defer r.failuresLock.Unlock()

	return r.failures[item]
}

// Forget removes an item from the failure map.
func (r *ItemExponentialFailureRateLimiter[T]) Forget(item T) {
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

// When calculates the delay for an item based on whether it has exceeded the maximum number of fast attempts.
func (r *ItemFastSlowRateLimiter[T]) When(item T) time.Duration {
	r.failuresLock.Lock()
	defer r.failuresLock.Unlock()

	r.failures[item]++

	if r.failures[item] <= r.maxFastAttempts {
		return r.fastDelay
	}

	return r.slowDelay
}

// Retries returns the number of times an item has been requeued.
func (r *ItemFastSlowRateLimiter[T]) Retries(item T) int {
	r.failuresLock.Lock()
	defer r.failuresLock.Unlock()

	return r.failures[item]
}

// Forget removes an item from the failure map.
func (r *ItemFastSlowRateLimiter[T]) Forget(item T) {
	r.failuresLock.Lock()
	defer r.failuresLock.Unlock()

	delete(r.failures, item)
}

// MaxOfRateLimiter calls every RateLimiter and returns the worst case response
// When used with a token bucket limiter, the burst could be apparently exceeded in cases where particular items
// were separately delayed a longer time.
type MaxOfRateLimiter[T comparable] struct {
	limiters []RateLimiter[T]
}

var _ RateLimiter[any] = &MaxOfRateLimiter[any]{}

// NewMaxOfRateLimiter creates a new MaxOfRateLimiter with the specified limiters.
func NewMaxOfRateLimiter[T comparable](limiters ...RateLimiter[T]) RateLimiter[T] {
	return &MaxOfRateLimiter[T]{limiters: limiters}
}

// When calculates the maximum delay among all the limiters for an item.
func (r *MaxOfRateLimiter[T]) When(item T) time.Duration {
	ret := time.Duration(0)

	for _, limiter := range r.limiters {
		curr := limiter.When(item)
		if curr > ret {
			ret = curr
		}
	}

	return ret
}

// Retries returns the maximum number of retries among all the limiters for an item.
func (r *MaxOfRateLimiter[T]) Retries(item T) int {
	ret := 0

	for _, limiter := range r.limiters {
		curr := limiter.Retries(item)
		if curr > ret {
			ret = curr
		}
	}

	return ret
}

// Forget calls the Forget method on all the limiters for an item.
func (r *MaxOfRateLimiter[T]) Forget(item T) {
	for _, limiter := range r.limiters {
		limiter.Forget(item)
	}
}
