// Package main provides the entry of turl.
package main

import (
	"log/slog"
	"os"

	"github.com/beihai0xff/turl/cli"
)

func main() {
	app := cli.New()
	if err := app.Run(os.Args); err != nil {
		slog.Error("app run failed", slog.Any("error", err))
		os.Exit(1)
	}
}
