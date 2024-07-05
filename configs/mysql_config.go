// Package configs provides config management
package configs

// MySQLConfig MySQLConfig Config
type MySQLConfig struct {
	// DSN is the data source name
	DSN string `json:"dsn" yaml:"dsn" mapstructure:"dsn"`
	// MaxIdleConn is the max open connections
	MaxConn int `validate:"required,min=1" json:"max_conn" yaml:"max_conn" mapstructure:"max_conn"`
}
