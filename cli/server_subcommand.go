package cli

import (
	"fmt"
	"log"

	"github.com/urfave/cli/v2"

	"github.com/beihai0xff/turl/api"
	"github.com/beihai0xff/turl/configs"
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

func (c *serverCLI) serverStart(_ *cli.Context) error {
	// conf := c.parseServerStartConfig(ctx)
	//
	// srv := server.Start(conf)
	//
	// exitSignal := make(chan os.Signal, 1)
	// signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)
	// <-exitSignal
	//
	// quitCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// defer cancel()
	//
	// shutdown.GracefulShutdown(quitCtx,
	// 	shutdown.HTTPServerShutdown(srv),
	// )
	//
	// _, _ = fmt.Println("HTTP Server exited")
	return nil
}

func (c *serverCLI) parseServerStartConfig(ctx *cli.Context) *configs.ServerConfig {
	conf := configs.ServerConfig{
		Listen:      ctx.String("listen"),
		Port:        ctx.Int("port"),
		LogFilePath: ctx.String("logfile"),
		LogOutput:   ctx.StringSlice("log-output"),
	}
	if err := conf.Validate(); err != nil {
		log.Fatalf("invalid server config: %v \n%+v", err, conf)
	}

	return &conf
}
