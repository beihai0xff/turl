// Package cli provides CLI app
package cli

import (
	"fmt"
	"io"
	"runtime"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/urfave/cli/v2"

	"github.com/beiai0xff/turl/api"
)

const (
	exitCodeOK    = 0
	exitCodeError = 1
)

var (
	gitHash   string
	buildTime string
	version   = "v0.0.0"
)

// New returns a new cli app
func New() *cli.App {
	c := serverCLI{}

	app := cli.App{
		Name:                 "turl",
		EnableBashCompletion: true,
		HideVersion:          false,
		Version:              version,
		ExitErrHandler: func(c *cli.Context, err error) {
			if err != nil {
				err = cli.Exit(fmt.Sprintf("run [%s] command failed: %s\n", c.Command.FullName(), err), exitCodeError)
			}
			cli.HandleExitCoder(err)
		},
		Commands: []*cli.Command{
			{
				Name:   "start",
				Usage:  "Start turl HTTP server",
				Action: c.serverStart,
				Flags:  c.getServerStartFlags(),
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "address",
				Usage:   "turl server address, format as: ip:port",
				Value:   api.DefaultServerAddr,
				Aliases: []string{"addr"},
			},
		},
	}

	cli.VersionPrinter = versionPrinter

	return &app
}

func versionPrinter(c *cli.Context) {
	t := table.NewWriter()

	t.AppendHeader(table.Row{"# TURL Server INFO", ""})
	t.AppendRows([]table.Row{
		{"Version:", c.App.Version},
		{"Go Version:", runtime.Version()},
		{"Git Commit:", gitHash},
		{"OS/Arch:", fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)},
		{"Built Time:", buildTime},
	})

	t.SetStyle(table.Style{
		Name:    "StyleLightNoBordersAndSeparators",
		Box:     table.StyleBoxLight,
		Color:   table.ColorOptionsDefault,
		Format:  table.FormatOptionsDefault,
		HTML:    table.DefaultHTMLOptions,
		Options: table.OptionsNoBordersAndSeparators,
		Title:   table.TitleOptionsDefault,
	})
	t.SetOutputMirror(c.App.Writer)
	t.Render()
}

func renderTable(writer io.Writer, t table.Writer) {
	t.SetStyle(table.StyleLight)
	t.SetAutoIndex(true)
	t.SetOutputMirror(writer)
	t.Render()
}
