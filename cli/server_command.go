package cli

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/beihai0xff/turl/app/turl"
	"github.com/beihai0xff/turl/configs"
	"github.com/beihai0xff/turl/pkg/log"
	"github.com/beihai0xff/turl/pkg/shutdown"
)

type serverCLI struct{}

func (c *serverCLI) getServerStartFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "file",
			Aliases: []string{"l"},
			Usage:   "turl server config file path",
			Value:   "./turl.yaml",
		},
	}
}

func (c *serverCLI) serverStart(ctx *cli.Context) error {
	conf, err := c.parseServerStartConfig(ctx)
	if err != nil {
		return err
	}

	if err = log.SetDefaultLogger(conf.LogConfig); err != nil {
		return err
	}

	handler, err := turl.NewHandler(conf)
	if err != nil {
		return err
	}

	srv, err := turl.NewServer(handler, conf)
	if err != nil {
		return err
	}

	go func() {
		if err = srv.ListenAndServe(); err != nil && !errors.Is(http.ErrServerClosed, err) {
			slog.Error("listen and serve failed", slog.Any("error", err))
		}
	}()

	exitSignal := make(chan os.Signal, 1)
	signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)
	<-exitSignal

	quitCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second) //nolint:mnd
	defer cancel()

	shutdown.GracefulShutdown(quitCtx,
		shutdown.HTTPServerShutdown(srv),
		shutdown.HandlerShutdown(handler),
	)

	slog.Info("HTTP Server exited")

	return nil
}

func (c *serverCLI) parseServerStartConfig(ctx *cli.Context) (*configs.ServerConfig, error) {
	filePath := ctx.String("file")

	conf, err := configs.ReadFile(filePath)
	if err != nil {
		slog.Error("read server config file failed", slog.Any("error", err), slog.Any("file", filePath))
		return nil, err
	}

	if err = conf.Validate(); err != nil {
		slog.Error("invalid server config", slog.Any("error", err), slog.Any("config", conf))
		return nil, err
	}

	return conf, nil
}
