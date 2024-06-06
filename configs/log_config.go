// Package configs provides the config management
// log_config.go provides config management
package configs

import (
	"encoding/json"
	"log/slog"
)

// Levels slog level
var Levels = map[string]slog.Level{
	"":         slog.LevelInfo,
	DebugLevel: slog.LevelDebug,
	InfoLevel:  slog.LevelInfo,
	WarnLevlel: slog.LevelWarn,
	ErrorLevel: slog.LevelError,
}

const (
	// OutputConsole console output
	OutputConsole = "console"
	// OutputFile file output
	OutputFile = "file"

	// EncoderTypeText log format console encoder
	EncoderTypeText = "text"
	// EncoderTypeJSON log format json encoder
	EncoderTypeJSON = "json"

	// DebugLevel log level debug, equal to slog.LevelDebug
	DebugLevel = "debug"
	// InfoLevel log level info, equal to slog.LevelInfo
	InfoLevel = "info"
	// WarnLevlel log level warn, equal to slog.LevelWarn
	WarnLevlel = "warn"
	// ErrorLevel log level error, equal to slog.LevelError
	ErrorLevel = "error"
)

// LogConfig log output: console file remote
type LogConfig struct {
	// Writers log output(OutputConsole, OutputFile)
	Writers []string `yaml:"writers" mapstructure:"writers" json:"writers"`
	// FileConfig log file config, if writers has file, must set file config
	FileConfig FileConfig `yaml:"file_config" mapstructure:"file_config" json:"file_config"`

	// Format log format type (console, json)
	Format string `yaml:"format" mapstructure:"format" json:"format"`
	// AddSource add source file and line
	AddSource bool `json:"add_source" yaml:"add_source" mapstructure:"add_source"`
	// Level log level debug info error...
	Level string `yaml:"level" mapstructure:"level" json:"level"`
}

func (c *LogConfig) String() string {
	j, _ := json.MarshalIndent(c, "", "    ")
	return string(j)
}

// FileConfig log file config
type FileConfig struct {
	// Filepath log file path
	Filepath string `yaml:"filepath" mapstructure:"filepath" json:"filepath"`
	// MaxAge log file max age, days
	MaxAge int `yaml:"max_age" mapstructure:"max_age" json:"max_age"`
	// MaxBackups max backup files
	MaxBackups int `yaml:"max_backups" mapstructure:"max_backups" json:"max_backups"`
	// Compress log file is compress
	Compress bool `yaml:"compress" mapstructure:"compress" json:"compress"`
	// MaxSize max file size, MB
	MaxSize int `yaml:"max_size" mapstructure:"max_size" json:"max_size"`
}
