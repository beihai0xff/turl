package tests

import (
	"time"

	"github.com/beihai0xff/turl/configs"
)

// DSN is the data source name of the MySQL database
const DSN = "root:test123@tcp(127.0.0.1:3306)/tiny-url?charset=utf8mb4&parseTime=True&loc=Local"

// RedisAddr is the address of the redis server
var RedisAddr = []string{"127.0.0.1:6379"}

var GlobalConfig = &configs.ServerConfig{
	Listen: "localhost",
	Port:   8080,
	TDDLConfig: &configs.TDDLConfig{
		Step:     100,
		StartNum: 10000,
		SeqName:  "tiny_url",
	},
	MySQLConfig: &configs.MySQLConfig{
		DSN: DSN,
	},
	CacheConfig: &configs.CacheConfig{
		RedisConfig: &configs.RedisConfig{
			Addr:        RedisAddr,
			DialTimeout: time.Second,
		},
		RemoteCacheTTL: 10 * time.Minute,
		LocalCacheConfig: &configs.LocalCacheConfig{
			TTL:       10 * time.Minute,
			Capacity:  1e8,
			MaxMemory: 512,
		},
	},
}
