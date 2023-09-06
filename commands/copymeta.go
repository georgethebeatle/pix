package commands

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/georgethebeatle/pix/commands/helpers"
	"github.com/urfave/cli/v2"
)

var CopyMetaCommand = &cli.Command{
	Name:  "copymeta",
	Usage: `Batch copies exif metadata for each photo in one directory to the file with the same name in another. You need to have the "exiftool" program on the path in order to use this command`,
	Action: func(ctx *cli.Context) error {
		helpers.AssertArgsNumber(ctx, 2)

		srcDir := ctx.Args().Get(0)
		dstDir := ctx.Args().Get(1)

		return copyMetadata(srcDir, dstDir)
	},
}

func copyMetadata(srcDir, dstDir string) error {
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
		if _, err := os.Stat(filepath.Join(dstDir, file.Name())); err != nil {
			return fmt.Errorf("destination file does not exist: %w", err)
		}
	}

	for _, file := range files {
		srcImagePath := filepath.Join(srcDir, file.Name())
		dstImagePath := filepath.Join(dstDir, file.Name())

		cmdErr := bytes.NewBuffer([]byte{})
		cmd := exec.Command("exiftool", "-overwrite_original", "-tagsFromFile", srcImagePath, dstImagePath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = cmdErr
		if err := cmd.Run(); err != nil {
			fmt.Printf("err.Error() = %+v\n", err.Error())
			if strings.Contains(cmdErr.String(), "Unknown file type") {
				fmt.Printf("Skipping unknown file: %q\n", srcImagePath)
				continue
			}
			return fmt.Errorf("failed to copy exif metadata: %w", err)
		}
	}
	return nil
}
