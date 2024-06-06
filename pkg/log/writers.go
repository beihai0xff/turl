package log

import (
	"io"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/beiai0xff/turl/configs"
)

func getWriters(c *configs.LogConfig) io.Writer {
	var writers []io.Writer

	for _, writer := range c.Writers {
		if writer == configs.OutputConsole {
			writers = append(writers, getConsoleWriter())
		}
		if writer == configs.OutputFile {
			writer, _ := getFileWriter(&c.FileConfig)
			// TODO: add cleanFunc to cleanFuncs
			// cleanFuncs = append(cleanFuncs, cleanFunc)
			writers = append(writers, writer)
		}
	}

	return io.MultiWriter(writers...)
}

// getConsoleWriter write log to console
func getConsoleWriter() io.Writer {
	return os.Stdout
}

// getFileWriter write log to file
func getFileWriter(c *configs.FileConfig) (io.Writer, func()) {
	writer := lumberjack.Logger{
		// It uses <processname>-lumberjack.log in os.TempDir() if empty.
		Filename:   c.Filepath,
		MaxSize:    c.MaxSize,
		MaxAge:     c.MaxAge,
		MaxBackups: c.MaxBackups,
		Compress:   c.Compress,
	}

	return &writer, func() {
		writer.Close()
	}
}
