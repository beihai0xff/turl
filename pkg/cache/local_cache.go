// Package lcahce provides the local cache
package cache

import (
	"context"
	"errors"
	"time"

	"github.com/allegro/bigcache/v3"
)

var (
	_             Interface = (*localCache)(nil)
	errInvalidCap           = errors.New("cache: invalid capacity")
)

type localCache struct {
	cache *bigcache.BigCache
}

// NewLocalCache create a local cache
// capacity is the cache capacity
// ttl is the time to live
func NewLocalCache(capacity int, ttl time.Duration) (Interface, error) {
	return newLocalCache(capacity, ttl)
}

func newLocalCache(capacity int, ttl time.Duration) (*localCache, error) {
	if capacity <= 0 {
		return nil, errInvalidCap
	}

	config := bigcache.DefaultConfig(ttl)
	config.MaxEntriesInWindow = capacity

	c, err := bigcache.New(context.Background(), config)
	if err != nil {
		return nil, err
	}

	return &localCache{cache: c}, err
}

// Set the k v pair to the cache
// Note that the duration is not used
func (l *localCache) Set(_ context.Context, k string, v []byte, _ time.Duration) error {
	return l.cache.Set(k, v)
}

// Get the value by key
func (l *localCache) Get(_ context.Context, k string) ([]byte, error) {
	v, err := l.cache.Get(k)
	if err != nil && errors.Is(err, bigcache.ErrEntryNotFound) {
		return nil, ErrCacheMiss
	}

	return v, err
}

// Close the cache
func (l *localCache) Close() error {
	return l.cache.Close()
}
