// Package main provides the entry of turl.
package main

import (
	"log"
	"os"

	"github.com/beihai0xff/turl/cli"
)

func main() {
	app := cli.New()
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
