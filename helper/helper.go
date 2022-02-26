package helper

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"

	"github.com/sealsurlaw/ImageServer/errs"
	"golang.org/x/image/draw"
)

func CreateThumbnail(file *os.File, resolution int) (io.Reader, error) {
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

	smallestSide := int(math.Min(float64(img.Bounds().Dx()), float64(img.Bounds().Dy())))
	rect := image.Rect(0, 0, smallestSide, smallestSide)

	var sp image.Point
	if img.Bounds().Dx() > img.Bounds().Dy() {
		x := int((img.Bounds().Dx() - img.Bounds().Dy()) / 2)
		sp = image.Pt(x, 0)
	} else {
		y := int((img.Bounds().Dy() - img.Bounds().Dx()) / 2)
		sp = image.Pt(0, y)
	}

	thumbImg := image.NewRGBA(rect)
	draw.Draw(thumbImg, rect, img, sp, draw.Src)

	thumbImgScaled := image.NewRGBA(image.Rect(0, 0, resolution, resolution))
	draw.NearestNeighbor.Scale(thumbImgScaled, thumbImgScaled.Bounds(), thumbImg, thumbImg.Bounds(), draw.Src, nil)

	buf := new(bytes.Buffer)
	err = png.Encode(buf, thumbImgScaled)
	if err != nil {
		return nil, err
	}

	return buf, nil
}
