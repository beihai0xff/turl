// Package configs is the config of turl server
package configs

import (
	"errors"

	"golang.org/x/exp/slices"

	"github.com/beihai0xff/turl/pkg/validate"
)

// ServerConfig is the config of turl server
type ServerConfig struct {
	// Listen is the http server listen address of turl server
	Listen string `validate:"required,ip_addr|hostname" json:"listen" yaml:"listen" mapstructure:"listen"`
	// Port is the http server port of turl server
	Port int `validate:"required,min=1,max=65535" json:"port" yaml:"port" mapstructure:"port"`
	// Rate is the token bucket rate of http server request rate limiter
	Rate int `validate:"required,min=1" json:"rate" yaml:"rate" mapstructure:"rate"`
	// Burst is the token bucket burst of http server request rate limiter
	Burst int `validate:"required,min=1" json:"burst" yaml:"burst" mapstructure:"burst"`

	// LogConfig is the log config of turl server
	LogConfig *LogConfig `json:"log" yaml:"log" mapstructure:"log"`

	// TDDLConfig is the tddl config of turl server
	TDDLConfig *TDDLConfig `json:"tddl" yaml:"tddl" mapstructure:"tddl"`
	// MySQLConfig is the mysql config of turl server
	MySQLConfig *MySQLConfig `json:"mysql" yaml:"mysql" mapstructure:"mysql"`
	// CacheConfig is the cache config of turl server
	CacheConfig *CacheConfig `json:"redis" yaml:"redis" mapstructure:"redis"`
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

	for _, v := range c.LogConfig.Writers {
		if !slices.Contains([]string{OutputConsole, OutputFile}, v) {
			return errInvalidOutput
		}

		if v == OutputFile && c.LogConfig.FileConfig.Filepath == "" {
			return errNonFilePath
		}
	}

	if !slices.Contains([]string{EncoderTypeText, EncoderTypeJSON}, c.LogConfig.Format) {
		return errInvalidFormat
	}

	return nil
}
