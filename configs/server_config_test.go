package configs

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestServerConfig_Validate(t *testing.T) {
	c := &ServerConfig{
		Listen:              "127.0.0.1",
		Port:                1231,
		RequestTimeout:      time.Second,
		GlobalRateLimitKey:  "test_rate",
		GlobalWriteRate:     1,
		GlobalWriteBurst:    1,
		StandAloneReadRate:  1,
		StandAloneReadBurst: 1,
		Log: &LogConfig{
			Writers: []string{OutputConsole},
			Format:  EncoderTypeText,
			Level:   InfoLevel,
		},
		TDDL: &TDDLConfig{
			Step:     1000,
			SeqName:  "test",
			StartNum: 10,
		},
		MySQL: &MySQLConfig{
			DSN: "test",
		},
		Cache: &CacheConfig{
			Redis:          nil,
			RemoteCacheTTL: time.Second,
			LocalCache: &LocalCacheConfig{
				TTL:       time.Second,
				Capacity:  100000,
				MaxMemory: 512,
			},
		},
	}
	require.NoError(t, c.Validate())

	c.Listen = "localhost"
	require.NoError(t, c.Validate())
	c.Listen = "github.com"
	require.NoError(t, c.Validate())
	c.Listen = "0.0.0.0"
	require.NoError(t, c.Validate())
	c.Listen = "192.168.1.1"
	require.NoError(t, c.Validate())

	c.Listen = "127.0.0.1"

	c.Port = 0
	require.Equal(t, "Key: 'ServerConfig.Port' Error:Field validation for 'Port' failed on the 'required' tag", c.Validate().Error())
	c.Port = -1
	require.Equal(t, "Key: 'ServerConfig.Port' Error:Field validation for 'Port' failed on the 'min' tag", c.Validate().Error())
	c.Port = 65536
	require.Equal(t, "Key: 'ServerConfig.Port' Error:Field validation for 'Port' failed on the 'max' tag", c.Validate().Error())
	c.Port = 65535
	require.NoError(t, c.Validate())

	c.RequestTimeout = time.Millisecond
	require.Error(t, c.Validate())
}
