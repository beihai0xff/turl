// Package redis provides a redis client
package redis

import "github.com/redis/go-redis/v9"

// Client returns a redis client
func Client(addr []string) redis.UniversalClient {
	return redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs: addr,
	})
}
