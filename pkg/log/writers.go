package log

import (
	"io"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/beihai0xff/turl/configs"
)

func getWriters(c *configs.LogConfig) io.Writer {
	var writers []io.Writer

	for _, writer := range c.Writers {
		switch writer {
		case configs.OutputFile:
			writer, _ := getFileWriter(&c.FileConfig)
			// TODO: add cleanFunc to close file writer
			// cleanFuncs = append(cleanFuncs, cleanFunc)
			writers = append(writers, writer)
		case configs.OutputConsole:
			writers = append(writers, getConsoleWriter())
		default:
			writers = append(writers, getConsoleWriter())
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
		// It uses <processname>-lumberjack.log in os.TempDir() if Filename is empty.
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
