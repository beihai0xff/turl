package configs

import "time"

// RedisConfig Redis config
type RedisConfig struct {
	// Addr is the redis address
	Addr []string `json:"addr" yaml:"addr" mapstructure:"addr"`
	// DialTimeout Dial timeout for establishing new connections.
	// Default is 5 seconds.
	DialTimeout time.Duration `json:"dial_timeout" yaml:"dial_timeout" mapstructure:"dial_timeout"`
}

// CacheConfig is the cache config of turl server
type CacheConfig struct {
	// RemoteCache    bool         `json:"remote_cache" yaml:"remote_cache" mapstructure:"remote_cache"`
	// RedisConfig is the redis config of turl server
	RedisConfig *RedisConfig `json:"redis" yaml:"redis" mapstructure:"redis"`
	// RemoteCacheTTL is the remote cache ttl
	RemoteCacheTTL time.Duration
	// LocalCacheTTL is the local cache ttl
	LocalCacheTTL time.Duration
	// LocalCacheSize is the local cache size
	LocalCacheSize int
}
