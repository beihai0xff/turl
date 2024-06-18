// Package lcahce provides the local cache
package lcahce

import (
	"context"
	"errors"
	"time"

	"github.com/allegro/bigcache/v3"

	"github.com/beiai0xff/turl/pkg/cache"
)

var (
	_             cache.Interface = (*localCache)(nil)
	errInvalidCap                 = errors.New("cache: invalid capacity")
)

type localCache struct {
	cache *bigcache.BigCache
}

// New create a local cache
// cap is the cache capacity
// mem is the cache Max memory, unit is byte e.g. 1 << 30 is 1GB
func New(cap int) (cache.Interface, error) {
	return newLocalCache(cap)
}

func newLocalCache(cap int) (*localCache, error) {
	if cap <= 0 {
		return nil, errInvalidCap
	}
	config := bigcache.DefaultConfig(10 * time.Minute)
	config.MaxEntriesInWindow = cap

	c, err := bigcache.New(context.Background(), config)
	if err != nil {
		return nil, err
	}

	return &localCache{cache: c}, err

}

func (l *localCache) Set(_ context.Context, k string, v []byte, _ time.Duration) error {
	return l.cache.Set(k, v)
}

func (l *localCache) Get(_ context.Context, k string) ([]byte, error) {
	return l.cache.Get(k)
}

func (l *localCache) Close() error {
	return l.cache.Close()
}
