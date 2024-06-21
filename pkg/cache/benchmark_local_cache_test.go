package cache

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/allegro/bigcache/v3"
)

func Benchmark_bigCache_Set(b *testing.B) {
	config := bigcache.DefaultConfig(10 * time.Minute)
	config.MaxEntriesInWindow = 1e8

	c, _ := bigcache.New(context.Background(), config)

	v := []byte("https://www.abc.com/images/100040.jpg")

	keys := make([]string, 0, b.N)
	for i := range b.N {
		keys = append(keys, strconv.Itoa(i+10000))
	}

	b.ReportAllocs()

	b.ResetTimer()
	for i := range b.N {
		_ = c.Set(keys[i], v)
	}
}
