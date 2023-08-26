package helpers

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/dsoprea/go-exif/v3"
	exifcommon "github.com/dsoprea/go-exif/v3/common"
	log "github.com/dsoprea/go-logging"
)

const DateTimeOriginal = "DateTimeOriginal"

func GetImageTime(imagePath string) (time.Time, error) {
	f, err := os.Open(imagePath)
	if err != nil {
		return time.Time{}, err
	}
	defer f.Close()

	rawExif, _, err := extractExifBytes(f)
	if err != nil {
		return time.Time{}, err
	}

	rootIfd, err := getRootIfd(rawExif)
	if err != nil {
		return time.Time{}, err
	}
	exifIfd, err := rootIfd.ChildWithIfdPath(exifcommon.IfdExifStandardIfdIdentity)
	if err != nil {
		return time.Time{}, err
	}

	results, err := exifIfd.FindTagWithName(DateTimeOriginal)
	if err != nil {
		return time.Time{}, err
	}

	if len(results) == 0 {
		return time.Time{}, fmt.Errorf("image does not have tags with name %q", DateTimeOriginal)
	}

	dateStr, err := results[0].FormatFirst()
	if err != nil {
		return time.Time{}, err
	}

	return time.Parse("2006:01:02 15:04:05", dateStr)
}

func SetImageTime(imagePath string, newTime time.Time) error {
	f, err := os.OpenFile(imagePath, os.O_RDWR, 0o755)
	if err != nil {
		return err
	}
	defer f.Close()

	exifBytes, exifStartByte, err := extractExifBytes(f)
	if err != nil {
		return err
	}

	oldDate := []byte("2023:06:18 11:28:22")
	newDate := newTime.AppendFormat([]byte{}, "2006:01:02 15:04:05")
	updatedExifBytes := bytes.ReplaceAll(exifBytes, oldDate, newDate)

	offset, err := f.Seek(exifStartByte, io.SeekStart)
	if err != nil {
		return err
	}

	_, err = f.WriteAt(updatedExifBytes, offset)
	if err != nil {
		return err
	}

	return nil
}

func getRootIfd(rawExif []byte) (*exif.Ifd, error) {
	im, err := exifcommon.NewIfdMappingWithStandard()
	if err != nil {
		return nil, err
	}

	ti := exif.NewTagIndex()
	_, index, err := exif.Collect(im, ti, rawExif)
	if err != nil {
		return nil, err
	}

	return index.RootIfd, nil
}

func extractExifBytes(r io.Reader) ([]byte, int64, error) {
	// Search for the beginning of the EXIF information. The EXIF is near the
	// beginning of most JPEGs, so this likely doesn't have a high cost (at
	// least, again, with JPEGs).

	br := bufio.NewReader(r)

	var discarded int64 = 0
	for {
		window, err := br.Peek(exif.ExifSignatureLength)
		if err != nil {
			if err == io.EOF {
				return nil, 0, exif.ErrNoExif
			}

			log.Panic(err)
		}

		_, err = exif.ParseExifHeader(window)
		if err != nil {
			if log.Is(err, exif.ErrNoExif) {
				// No EXIF. Move forward by one byte.

				_, err = br.Discard(1)
				log.PanicIf(err)

				discarded++

				continue
			}

			// Some other error.
			log.Panic(err)
		}

		break
	}

	rawExif, err := io.ReadAll(br)
	log.PanicIf(err)

	return rawExif, discarded, nil
}
