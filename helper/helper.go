package helper

import (
	"bytes"
	"crypto/sha512"
	"encoding/base64"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"math"
	"net"
	"net/http"

	"github.com/sealsurlaw/gouvre/errs"
	"golang.org/x/image/bmp"
	"golang.org/x/image/draw"
)

func CalculateHash(str string) string {
	s := sha512.Sum384([]byte(str))
	return base64.URLEncoding.EncodeToString(s[:])
}

func CreateThumbnail(fileData []byte, resolution int, cropped bool, quality int) ([]byte, error) {
	contentType := http.DetectContentType(fileData)

	r := bytes.NewReader(fileData)
	img, err := decodeImage(contentType, r)
	if err != nil {
		return nil, err
	}

	var thumbImgScaled *image.RGBA
	if cropped {
		thumbImgScaled = cropAndScale(img, resolution)
	} else {
		thumbImgScaled = scale(img, resolution)
	}

	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, thumbImgScaled, &jpeg.Options{
		Quality: quality,
	})
	if err != nil {
		return nil, err
	}

	thumbData, err := ioutil.ReadAll(buf)
	if err != nil {
		return nil, err
	}

	return thumbData, nil
}

func GetIpAddress(r *http.Request) string {
	var ip string
	if r.Header.Get("x-forward-for") != "" {
		return r.Header.Get("x-forward-for")
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return ""
	}

	return ip
}

func IsSupportedContentType(contentType string) bool {
	switch contentType {
	case "image/jpeg":
		fallthrough
	case "image/png":
		fallthrough
	case "image/gif":
		fallthrough
	case "image/bmp":
		return true
	}

	return false
}

func decodeImage(contentType string, r *bytes.Reader) (image.Image, error) {
	switch contentType {
	case "image/jpeg":
		return jpeg.Decode(r)
	case "image/png":
		return png.Decode(r)
	case "image/gif":
		return gif.Decode(r)
	case "image/bmp":
		return bmp.Decode(r)
	}

	return nil, errs.ErrInvalidContentType
}

func cropAndScale(img image.Image, resolution int) *image.RGBA {
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

	thumbImgScaled := image.NewRGBA(image.Rect(0, 0, resolution, resolution))
	draw.BiLinear.Scale(thumbImgScaled, thumbImgScaled.Bounds(), thumbImg, thumbImg.Bounds(), draw.Src, nil)
	return thumbImgScaled
}

func scale(img image.Image, resolution int) *image.RGBA {
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

	thumbImgScaled := image.NewRGBA(rect)
	draw.NearestNeighbor.Scale(thumbImgScaled, thumbImgScaled.Bounds(), img, img.Bounds(), draw.Src, nil)
	return thumbImgScaled
}
