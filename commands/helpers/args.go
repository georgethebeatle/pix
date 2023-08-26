package helpers

import (
	"log"

	"github.com/urfave/cli/v2"
)

func AssertArgsNumber(ctx *cli.Context, numRequired int) {
	if ctx.Args().Len() != numRequired {
		log.Fatalf("%d args required, %d args supplied", numRequired, ctx.Args().Len())
	}
}
