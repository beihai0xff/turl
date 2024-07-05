// Package redis provides a redis client
package redis

import (
	"github.com/redis/go-redis/v9"

	"github.com/beihai0xff/turl/configs"
)

// Nil is the redis.Nil, used to check if a key exists
var Nil = redis.Nil

// Client returns a redis client
func Client(c *configs.RedisConfig) redis.UniversalClient {
	return redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:          c.Addr,
		DialTimeout:    c.DialTimeout,
		MaxIdleConns:   c.MaxConn,
		MaxActiveConns: c.MaxConn,
	})
}
