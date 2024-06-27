package configs

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
		LogConfig: &LogConfig{
			Writers: []string{OutputConsole},
			Format:  EncoderTypeText,
			Level:   InfoLevel,
		},
		TDDLConfig:  nil,
		MySQLConfig: nil,
		CacheConfig: nil,
	}
	assert.NoError(t, c.Validate())

	c.Listen = "localhost"
	assert.NoError(t, c.Validate())
	c.Listen = "github.com"
	assert.NoError(t, c.Validate())
	c.Listen = "0.0.0.0"
	assert.NoError(t, c.Validate())
	c.Listen = "192.168.1.1"
	assert.NoError(t, c.Validate())

	c.Listen = "127.0.0.1"

	c.Port = 0
	assert.Equal(t, "Key: 'ServerConfig.Port' Error:Field validation for 'Port' failed on the 'required' tag", c.Validate().Error())
	c.Port = -1
	assert.Equal(t, "Key: 'ServerConfig.Port' Error:Field validation for 'Port' failed on the 'min' tag", c.Validate().Error())
	c.Port = 65536
	assert.Equal(t, "Key: 'ServerConfig.Port' Error:Field validation for 'Port' failed on the 'max' tag", c.Validate().Error())
	c.Port = 65535
	assert.NoError(t, c.Validate())

	c.RequestTimeout = time.Millisecond
	assert.Error(t, c.Validate())
}
