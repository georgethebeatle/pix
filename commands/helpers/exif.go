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

type ImageExif struct {
	imagePath  string
	exifBytes  []byte
	exifOffset int64
	ifd        *exif.Ifd
}

func NewImageExif(imagePath string) (*ImageExif, error) {
	f, err := os.Open(imagePath)
	if err != nil {
		return &ImageExif{}, err
	}
	defer f.Close()

	rawExif, exifOffset, err := extractExifBytes(f)
	if err != nil {
		return &ImageExif{}, err
	}

	rootIfd, err := getRootIfd(rawExif)
	if err != nil {
		return &ImageExif{}, err
	}
	exifIfd, err := rootIfd.ChildWithIfdPath(exifcommon.IfdExifStandardIfdIdentity)
	if err != nil {
		return &ImageExif{}, err
	}

	return &ImageExif{
		imagePath:  imagePath,
		exifBytes:  rawExif,
		exifOffset: exifOffset,
		ifd:        exifIfd,
	}, nil
}

func (i *ImageExif) GetTagValue(tagName string) (string, error) {
	tags, err := i.ifd.FindTagWithName(tagName)
	if err != nil {
		return "", err
	}

	if len(tags) == 0 {
		return "", fmt.Errorf("image does not have tags with name %q", tagName)
	}

	tagStr, err := tags[0].FormatFirst()
	if err != nil {
		return "", err
	}

	return tagStr, nil
}

func (i *ImageExif) GetImageTime() (time.Time, error) {
	dateStr, err := i.GetTagValue(DateTimeOriginal)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get image time: %w", err)
	}

	imageTime, err := time.Parse("2006:01:02 15:04:05", dateStr)
	if err != nil {
		return time.Time{}, err
	}

	timeOffset, _ := i.GetTagValue("OffsetTimeOriginal")
	if timeOffset == "" {
		timeOffset = "00:00"
	}

	var hours, minutes int
	_, err = fmt.Sscanf(timeOffset, "%d:%d", &hours, &minutes)
	if err != nil {
		return time.Time{}, err
	}

	return imageTime.Add(-time.Duration(hours) * time.Hour).Add(-time.Duration(minutes) * time.Minute), nil
}

func (i *ImageExif) SetImageTime(newTime time.Time) error {
	f, err := os.OpenFile(i.imagePath, os.O_RDWR, 0o755)
	if err != nil {
		return err
	}
	defer f.Close()

	dateStr, err := i.GetTagValue(DateTimeOriginal)
	if err != nil {
		return err
	}
	oldDate := []byte(dateStr)
	newDate := newTime.AppendFormat([]byte{}, "2006:01:02 15:04:05")
	updatedExifBytes := bytes.ReplaceAll(i.exifBytes, oldDate, newDate)

	offset, err := f.Seek(i.exifOffset, io.SeekStart)
	if err != nil {
		return err
	}

	_, err = f.WriteAt(updatedExifBytes, offset)
	if err != nil {
		return err
	}

	rawExif, exifOffset, err := extractExifBytes(f)
	if err != nil {
		return err
	}

	rootIfd, err := getRootIfd(rawExif)
	if err != nil {
		return err
	}
	exifIfd, err := rootIfd.ChildWithIfdPath(exifcommon.IfdExifStandardIfdIdentity)
	if err != nil {
		return err
	}

	i.exifBytes = rawExif
	i.exifOffset = exifOffset
	i.ifd = exifIfd

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
