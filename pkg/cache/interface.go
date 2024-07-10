// Package cache provides the cache management interface define
package cache

import (
	"context"
	"errors"
	"time"
)

// ErrCacheMiss cache miss error
// if Get() method return this error, means key is not exist
var ErrCacheMiss = errors.New("cache: key is missing")

// Interface cache interface
type Interface interface {
	// Set the key value to cache
	Set(ctx context.Context, k string, v []byte, ttl time.Duration) error
	// Get the key value from cache
	Get(ctx context.Context, k string) ([]byte, error)
	// Del delete the key from cache
	Del(ctx context.Context, k string) error
	// Close the cache
	Close() error
}
