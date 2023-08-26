package main

import (
	"log"
	"os"

	"github.com/georgethebeatle/pix/commands"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "pix",
		Usage: "A tool for organising pics from multiple devices",
		Commands: []*cli.Command{
			commands.TimeDiffCommand,
			commands.OrganiseCommand,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
