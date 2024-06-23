// Package configs is the config of turl server
package configs

import (
	"errors"

	"github.com/samber/lo"

	"github.com/beihai0xff/turl/pkg/validate"
)

// ServerConfig is the config of turl server
type ServerConfig struct {
	// Listen is the http server listen address of turl server
	Listen string `validate:"required,ip_addr|hostname" json:"listen" yaml:"listen" mapstructure:"listen"`
	// Port is the http server port of turl server
	Port int `validate:"required,min=1,max=65535" json:"port" yaml:"port" mapstructure:"port"`
	// LogFilePath is the log file path of turl server
	LogFilePath string `json:"log_file_path" yaml:"log_file_path" mapstructure:"log_file_path"`
	// LogOutput is the log output of turl server
	// should be one of [console, file], can not be nil or empty
	LogOutput []string `validate:"required,min=1" json:"log_output" yaml:"log_output" mapstructure:"log_output"`

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
)

// Validate validates the config
// if return nil, the config is valid
func (c *ServerConfig) Validate() error {
	if err := validate.Instance().Struct(c); err != nil {
		return err
	}

	for _, v := range c.LogOutput {
		if !lo.Contains([]string{OutputConsole, OutputFile}, v) {
			return errInvalidOutput
		}

		if v == OutputFile && c.LogFilePath == "" {
			return errNonFilePath
		}
	}

	return nil
}
