package cache

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/beihai0xff/turl/internal/tests"
)

func testSet(b *testing.B, cache Interface, ttl time.Duration) {
	v := []byte("https://abc.com/images/100040.jpg")

	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		var counter int
		for pb.Next() {
			_ = cache.Set(ctx, strconv.Itoa(counter), v, ttl)
			counter++
		}
	})
}

func testGet(b *testing.B, cache Interface, ttl time.Duration) {
	v := []byte("https://abc.com/images/100040.jpg")

	ctx, nums := context.Background(), 10000

	for i := range nums {
		cache.Set(ctx, strconv.Itoa(i), v, ttl)
	}

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		var counter int
		for pb.Next() {
			_, _ = cache.Get(ctx, strconv.Itoa(counter%nums))
			counter++
		}
	})
}

func Benchmark_LocalCache_Set(b *testing.B) {
	cache, _ := newLocalCache(tests.GlobalConfig.CacheConfig.LocalCacheConfig)
	defer cache.Close()

	testSet(b, cache, 10*time.Minute)
}

func Benchmark_RedisCache_Set(b *testing.B) {
	cache := NewRedisRemoteCache(tests.GlobalConfig.CacheConfig.RedisConfig)
	defer cache.Close()

	testSet(b, cache, 10*time.Minute)
}

func Benchmark_LocalCache_Get(b *testing.B) {
	cache, _ := newLocalCache(tests.GlobalConfig.CacheConfig.LocalCacheConfig)
	defer cache.Close()

	testGet(b, cache, 10*time.Minute)
}

func Benchmark_RedisCache_Get(b *testing.B) {
	cache := NewRedisRemoteCache(tests.GlobalConfig.CacheConfig.RedisConfig)
	defer cache.Close()

	testGet(b, cache, 10*time.Minute)
}
