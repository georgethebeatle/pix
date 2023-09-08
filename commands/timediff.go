package commands

import (
	"fmt"
	"time"

	"github.com/georgethebeatle/pix/commands/helpers"
	"github.com/urfave/cli/v2"
)

var TimeDiffCommand = &cli.Command{
	Name:  "timediff",
	Usage: "Calculates the time diff between two images, based on a datetime tag",
	Action: func(ctx *cli.Context) error {
		helpers.AssertArgsNumber(ctx, 2)

		image1 := ctx.Args().Get(0)
		image2 := ctx.Args().Get(1)

		timediff, err := timediff(image1, image2)
		if err != nil {
			return err
		}

		fmt.Println(timediff)
		return nil
	},
}

func timediff(image1, image2 string) (time.Duration, error) {
	image1Exif, err := helpers.NewImageExif(image1)
	if err != nil {
		return 0, err
	}

	image2Exif, err := helpers.NewImageExif(image2)
	if err != nil {
		return 0, err
	}

	time1, err := image1Exif.GetImageTime()
	if err != nil {
		return 0, err
	}
	time2, err := image2Exif.GetImageTime()
	if err != nil {
		return 0, err
	}

	return time1.Sub(time2), nil
}
