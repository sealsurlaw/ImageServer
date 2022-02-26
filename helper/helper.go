package helper

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/sealsurlaw/ImageServer/errs"
	"github.com/sealsurlaw/ImageServer/linkstore"
	"golang.org/x/image/draw"
)

func CleanExpiredTokens(ls linkstore.LinkStore, duration time.Duration) {
	if duration == 0 {
		fmt.Println("Cleanup turned off.")
		return
	}

	timer := time.NewTicker(duration)

	for {
		select {
		case <-timer.C:
			fmt.Println("Cleaning up expired tokens...")
			ls.Cleanup()
			timer = time.NewTicker(duration)
		}
	}
}

func CreateThumbnail(file *os.File, resolution int, cropped bool) (io.Reader, error) {
	fileData, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	contentType := http.DetectContentType(fileData)

	r := bytes.NewReader(fileData)

	var img image.Image
	switch contentType {
	case "image/jpeg":
		img, err = jpeg.Decode(r)
	case "image/png":
		img, err = png.Decode(r)
	default:
		err = errs.ErrInvalidContentType
	}
	if err != nil {
		return nil, err
	}

	var thumbImgScaled *image.RGBA
	if cropped {
		smallestSide := int(math.Min(float64(img.Bounds().Dx()), float64(img.Bounds().Dy())))
		rect := image.Rect(0, 0, smallestSide, smallestSide)
		thumbImg := image.NewRGBA(rect)

		var sp image.Point
		if img.Bounds().Dx() > img.Bounds().Dy() {
			x := int((img.Bounds().Dx() - img.Bounds().Dy()) / 2)
			sp = image.Pt(x, 0)
		} else {
			y := int((img.Bounds().Dy() - img.Bounds().Dx()) / 2)
			sp = image.Pt(0, y)
		}

		draw.Draw(thumbImg, rect, img, sp, draw.Src)

		thumbImgScaled = image.NewRGBA(image.Rect(0, 0, resolution, resolution))
		draw.NearestNeighbor.Scale(thumbImgScaled, thumbImgScaled.Bounds(), thumbImg, thumbImg.Bounds(), draw.Src, nil)
	} else {
		smallestSide := int(math.Min(float64(img.Bounds().Dx()), float64(img.Bounds().Dy())))
		largestSide := int(math.Max(float64(img.Bounds().Dx()), float64(img.Bounds().Dy())))
		ratio := float64(smallestSide) / float64(largestSide)

		largestAfter := resolution
		smallestAfter := int(float64(largestAfter) * ratio)

		var rect image.Rectangle
		if img.Bounds().Dy() == largestSide {
			rect = image.Rect(0, 0, smallestAfter, largestAfter)
		} else {
			rect = image.Rect(0, 0, largestAfter, smallestAfter)
		}

		thumbImgScaled = image.NewRGBA(rect)
		draw.NearestNeighbor.Scale(thumbImgScaled, thumbImgScaled.Bounds(), img, img.Bounds(), draw.Src, nil)
	}

	buf := new(bytes.Buffer)
	err = png.Encode(buf, thumbImgScaled)
	if err != nil {
		return nil, err
	}

	return buf, nil
}
