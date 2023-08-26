package commands

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/dsoprea/go-exif/v3"
	"github.com/georgethebeatle/pix/commands/helpers"
	"github.com/urfave/cli/v2"
)

var OrganiseCommand = &cli.Command{
	Name:  "organise",
	Usage: "Stores images in a destination dir with timestamped filenames. An optional time shift can be supplied.",
	Action: func(ctx *cli.Context) error {
		helpers.AssertArgsNumber(ctx, 2)

		shift := ctx.Duration("shift")
		srcDir := ctx.Args().Get(0)
		dstDir := ctx.Args().Get(1)

		return organise(srcDir, dstDir, shift)
	},
	Flags: []cli.Flag{
		&cli.DurationFlag{
			Name: "shift",
		},
	},
}

func organise(srcDir, dstDir string, shift time.Duration) error {
	srcDir, err := filepath.Abs(srcDir)
	if err != nil {
		return err
	}

	dstDir, err = filepath.Abs(dstDir)
	if err != nil {
		return err
	}

	err = os.MkdirAll(dstDir, 0o755)
	if err != nil {
		return err
	}

	files, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		imagePath := filepath.Join(srcDir, file.Name())

		originalTime, err := helpers.GetImageTime(imagePath)
		if err != nil {
			if errors.As(err, &exif.ErrNoExif) {
				fmt.Printf("Skipping file with no exif data: %q", imagePath)
				continue
			}
			return err
		}

		newTime := originalTime.Add(shift)

		dstPath := generateUniquePath(dstDir, newTime)
		fmt.Printf("Copying %q to %q\n", imagePath, dstPath)
		err = copyFile(imagePath, dstPath)
		if err != nil {
			return err
		}

		err = helpers.SetImageTime(dstPath, newTime)
		if err != nil {
			return err
		}

	}
	return nil
}

func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}

func generateUniquePath(dstDir string, dateTaken time.Time) string {
	disambiguation := ""

	for i := 1; i < 100; i++ {
		if _, err := os.Stat(generatePath(dstDir, dateTaken, disambiguation)); err != nil {
			break
		}
		disambiguation = fmt.Sprintf("(%d)", i)
	}

	return generatePath(dstDir, dateTaken, disambiguation)
}

func generatePath(dstDir string, dateTaken time.Time, disambiguation string) string {
	return filepath.Join(dstDir, dateTaken.Format("2006-01-02H15-04-05")+disambiguation+".jpg")
}
