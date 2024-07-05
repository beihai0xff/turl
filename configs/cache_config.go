package configs

import "time"

// RedisConfig Redis config
type RedisConfig struct {
	// Addr is the redis address
	Addr []string `validate:"required" json:"addr" yaml:"addr" mapstructure:"addr"`
	// DialTimeout Dial timeout for establishing new connections.
	// Default is 5 seconds.
	DialTimeout time.Duration `validate:"required" json:"dial_timeout" yaml:"dial_timeout" mapstructure:"dial_timeout"`
	// MaxIdleConn is the max open connections
	MaxConn int `validate:"required,min=1" json:"max_conn" yaml:"max_conn" mapstructure:"max_conn"`
	// TTL is the redis cache ttl
	TTL time.Duration `validate:"required" json:"ttl" yaml:"ttl" mapstructure:"ttl"`
}

// LocalCacheConfig is the local cache config of turl server
type LocalCacheConfig struct {
	// TTL is the local cache ttl
	TTL time.Duration `validate:"required" json:"ttl" yaml:"ttl" mapstructure:"ttl"`
	// Capacity is the local cache capacity
	Capacity int `validate:"required" json:"capacity" yaml:"capacity" mapstructure:"capacity"`
	// MaxMemory max memory for value size in MB
	MaxMemory int `validate:"required" json:"max_memory" yaml:"max_memory" mapstructure:"max_memory"`
}

// CacheConfig is the cache config of turl server
type CacheConfig struct {
	// Redis is the redis config of turl server
	Redis *RedisConfig `json:"redis" yaml:"redis" mapstructure:"redis"`
	// LocalCache is the local cache config
	LocalCache *LocalCacheConfig `validate:"required" json:"local_cache" yaml:"local_cache" mapstructure:"local_cache"`
}
