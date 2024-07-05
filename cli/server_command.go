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

	"github.com/beihai0xff/turl/app/turl"
	"github.com/beihai0xff/turl/configs"
	"github.com/beihai0xff/turl/pkg/log"
	"github.com/beihai0xff/turl/pkg/shutdown"
)

var (
	configPathFlag = &cli.StringFlag{
		Name:    "file",
		Aliases: []string{"f"},
		Usage:   "TURL Server Config File Path",
		Value:   "./config.yaml",
		EnvVars: []string{"TURL_CONFIG_FILE", "TURL_FILE"},
	}
	readonlyFlag = &cli.BoolFlag{
		Name:    "readonly",
		Aliases: []string{"ro"},
		Usage:   "Start Server IN Read-Only Mode",
		Value:   false,
		EnvVars: []string{"TURL_READONLY", "TURL_RO"},
	}
	debugFlag = &cli.BoolFlag{
		Name:    "debug",
		Aliases: []string{"d"},
		Usage:   "Enable Debug Mode",
		Value:   false,
		EnvVars: []string{"TURL_DEBUG"},
	}
)

type serverCLI struct{}

func (c *serverCLI) getServerStartFlags() []cli.Flag {
	return []cli.Flag{configPathFlag, readonlyFlag, debugFlag}
}

func (c *serverCLI) serverHealth(ctx *cli.Context) error {
	rsp, err := http.Get(fmt.Sprintf("http://%s%s", ctx.String("addr"), turl.HealthCheckPath))
	if err != nil {
		slog.Error("health check failed", slog.Any("error", err))
		return err
	}

	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		slog.Error("health check failed", slog.Any("status", rsp.Status))
		return fmt.Errorf("health check failed, status: %s", rsp.Status)
	}

	slog.Info("health check success")

	return nil
}

func (c *serverCLI) serverStart(ctx *cli.Context) error {
	conf, err := c.parseServerStartConfig(ctx)
	if err != nil {
		return err
	}

	if err = log.SetDefaultLogger(conf.Log); err != nil {
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
	filePath := ctx.String(configPathFlag.Name)

	var mp = map[string]interface{}{}
	if ctx.Bool(readonlyFlag.Name) {
		mp[readonlyFlag.Name] = true
	}

	if ctx.Bool(debugFlag.Name) {
		mp[debugFlag.Name] = true
	}

	conf, err := configs.ReadFile(filePath, mp)
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
