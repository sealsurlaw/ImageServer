package helper

import (
	"bytes"
	"crypto/sha512"
	"encoding/base64"
	"image"
	"image/jpeg"
	"io/ioutil"
	"math"
	"net"
	"net/http"
	"os"

	"github.com/disintegration/imaging"
)

func CalculateHash(str string) string {
	s := sha512.Sum384([]byte(str))
	return base64.URLEncoding.EncodeToString(s[:])
}

func CreateThumbnail(fileData []byte, resolution int, cropped bool, quality int) ([]byte, error) {
	r := bytes.NewReader(fileData)
	img, err := imaging.Decode(r, imaging.AutoOrientation(true))
	if err != nil {
		return nil, err
	}

	var thumbImgScaled *image.NRGBA
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

func OpenFile(fullFilePath string) ([]byte, error) {
	file, err := os.Open(fullFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileData, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return fileData, nil
}

func cropAndScale(img image.Image, resolution int) *image.NRGBA {
	return imaging.Fill(img, resolution, resolution, imaging.Center, imaging.Lanczos)
}

func scale(img image.Image, resolution int) *image.NRGBA {
	newWidth := resolution

	// Height is larger than width
	if img.Bounds().Dy() > img.Bounds().Dx() {
		width := int(math.Min(float64(img.Bounds().Dx()), float64(img.Bounds().Dy())))
		height := int(math.Max(float64(img.Bounds().Dx()), float64(img.Bounds().Dy())))
		ratio := float64(width) / float64(height)
		newWidth = int(float64(resolution) * ratio)
	}

	return imaging.Resize(img, newWidth, 0, imaging.Lanczos)
}
