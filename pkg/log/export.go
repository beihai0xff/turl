package log

import (
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/beiai0xff/turl/configs"
)

// SetDefaultLogger new a slog log, default callerSkip is 1
func SetDefaultLogger(c *configs.LogConfig) error {
	l, err := NewLogger(c)
	if err != nil {
		return err
	}

	slog.SetDefault(l)

	return nil
}

// func GetLoggerByName(name string, c configs.LogConfig) (*slog.Logger, error) {
//
// }

func NewLogger(c *configs.LogConfig) (*slog.Logger, error) {
	w := getWriters(c)

	h, err := getLogHandler(w, c)
	if err != nil {
		return nil, err
	}

	return slog.New(h), nil
}

func getLogHandler(w io.Writer, c *configs.LogConfig) (slog.Handler, error) {
	opts := &slog.HandlerOptions{
		AddSource: c.AddSource,
		Level:     configs.Levels[c.Level],
	}

	switch c.Format {
	case configs.EncoderTypeText:
		return slog.NewTextHandler(w, opts), nil
	case configs.EncoderTypeJSON:
		return slog.NewJSONHandler(w, opts), nil
	default:
		return nil, errors.New(fmt.Sprintf("unknown log format %s", c.Format))
	}
}
