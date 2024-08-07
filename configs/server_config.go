// Package configs is the config of turl server
package configs

import (
	"errors"
	"time"

	"golang.org/x/exp/slices"

	"github.com/beihai0xff/turl/pkg/validate"
)

// ServerConfig is the config of turl server
type ServerConfig struct {
	// Listen is the http server listen address of turl server
	Listen string `validate:"required,ip_addr|hostname" json:"listen" yaml:"listen" mapstructure:"listen"`
	// Port is the http server port of turl server
	Port int `validate:"required,min=1,max=65535" json:"port" yaml:"port" mapstructure:"port"`
	// Debug is the debug mode of turl server
	Debug bool `json:"debug" yaml:"debug" mapstructure:"debug"`
	// Domain is the domain of redirect url
	Domain string `validate:"required" json:"domain" yaml:"domain" mapstructure:"domain"`
	// Readonly is the read-only mode of turl server
	Readonly bool `json:"readonly" yaml:"readonly" mapstructure:"readonly"`
	// RequestTimeout is the http server request timeout of turl server
	RequestTimeout time.Duration `validate:"required" json:"request_timeout" yaml:"request_timeout" mapstructure:"request_timeout"`
	// GlobalRateLimitKey is the key of global rate limiter
	GlobalRateLimitKey string `validate:"required" json:"global_rate_limit_key" yaml:"global_rate_limit_key" mapstructure:"global_rate_limit_key"`
	// GlobalWriteRate is the token bucket rate of write api rate limiter
	GlobalWriteRate int `validate:"required,gt=0" json:"global_write_rate" yaml:"global_write_rate" mapstructure:"global_write_rate"`
	// GlobalWriteBurst is the token bucket burst of write api rate limiter
	GlobalWriteBurst int `validate:"required,min=1" json:"global_write_burst" yaml:"global_write_burst" mapstructure:"global_write_burst"`
	// StandAloneReadRate is the token bucket rate of read api rate limiter
	StandAloneReadRate int `validate:"required,gt=0" json:"stand_alone_read_rate" yaml:"stand_alone_read_rate" mapstructure:"stand_alone_read_rate"`
	// StandAloneReadBurst is the token bucket burst of read api rate limiter
	StandAloneReadBurst int `validate:"required,min=1" json:"stand_alone_read_burst" yaml:"stand_alone_read_burst" mapstructure:"stand_alone_read_burst"`

	// Log is the log config of turl server
	Log *LogConfig `validate:"required" json:"log" yaml:"log" mapstructure:"log"`
	// TDDL is the tddl config of turl server
	TDDL *TDDLConfig `validate:"required" json:"tddl" yaml:"tddl" mapstructure:"tddl"`
	// MySQL is the mysql config of turl server
	MySQL *MySQLConfig `validate:"required" json:"mysql" yaml:"mysql" mapstructure:"mysql"`
	// Cache is the cache config of turl server
	Cache *CacheConfig `validate:"required" json:"cache" yaml:"cache" mapstructure:"cache"`
}

var (
	errInvalidOutput = errors.New("log output only support console and file")
	errNonFilePath   = errors.New("log file path is required when log output contains file")
	errInvalidFormat = errors.New("log format only support text and json")
)

// Validate validates the config
// if return nil, the config is valid
func (c *ServerConfig) Validate() error {
	if err := validate.Instance().Struct(c); err != nil {
		return err
	}

	if c.RequestTimeout < time.Second {
		return errors.New("request timeout should be greater than 1s")
	}

	for _, v := range c.Log.Writers {
		if !slices.Contains([]string{OutputConsole, OutputFile}, v) {
			return errInvalidOutput
		}

		if v == OutputFile && c.Log.FileConfig.Filepath == "" {
			return errNonFilePath
		}
	}

	if !slices.Contains([]string{EncoderTypeText, EncoderTypeJSON}, c.Log.Format) {
		return errInvalidFormat
	}

	return nil
}
