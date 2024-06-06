// Package main provides the entry of turl.
package main

import (
	"log"
	"os"

	"github.com/beiai0xff/turl/cli"
)

func main() {
	app := cli.New()
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
