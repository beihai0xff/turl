package cli

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/beihai0xff/turl/api"
	"github.com/beihai0xff/turl/app/turl"
	"github.com/beihai0xff/turl/configs"
	"github.com/beihai0xff/turl/pkg/log"
	"github.com/beihai0xff/turl/pkg/shutdown"
)

type serverCLI struct{}

func (c *serverCLI) getServerStartFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  "listen",
			Usage: "curl HTTP server listen address",
			Value: "0.0.0.0",
		},
		&cli.IntFlag{
			Name:    "port",
			Aliases: []string{"p"},
			Usage:   "turl HTTP server port",
			Value:   api.DefaultPort,
			Action: func(_ *cli.Context, v int) error {
				if v >= 65536 || v < 0 {
					return fmt.Errorf("flag port value %v out of range[0-65535]", v)
				}
				return nil
			},
		},
		&cli.StringSliceFlag{
			Name:  "log-output",
			Usage: "Set log output console or file",
			Value: cli.NewStringSlice(configs.OutputConsole),
		},
		&cli.StringFlag{
			Name:    "logfile",
			Aliases: []string{"l"},
			Usage:   "Set log file path",
			Value:   "./log/server.log",
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
	conf := configs.ServerConfig{
		Listen: ctx.String("listen"),
		Port:   ctx.Int("port"),
	}
	if err := conf.Validate(); err != nil {
		slog.Error("invalid server config", slog.Any("error", err), slog.Any("config", conf))
		return nil, err
	}

	return &conf, nil
}
