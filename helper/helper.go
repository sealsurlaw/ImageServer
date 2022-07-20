package helper

import (
	"bytes"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io/ioutil"
	"math"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/disintegration/imaging"
)

type IpfsResponse struct {
	Name string `json:"Name"`
	Hash string `json:"Hash"`
	Size string `json:"Size"`
}

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

func PinFile(
	fileData []byte,
	filename string,
) (cid string, err error) {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", err
	}

	_, err = part.Write(fileData)
	if err != nil {
		return "", err
	}

	err = writer.Close()
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "http://localhost:5001/api/v0/add?cid-version=1&pin=true", body)
	if err != nil {
		return "", err
	}

	req.Header.Add("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("status code: %d", resp.StatusCode)
	}

	p, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	j := &IpfsResponse{}
	err = json.Unmarshal(p, j)
	if err != nil {
		return "", err
	}

	return j.Hash, nil
}

func IsIpfsFilePinned(
	cid string,
) error {
	url := "http://localhost:5001/api/v0/pin/ls"
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("status code: %d", resp.StatusCode)
	}

	resBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if !strings.Contains(string(resBody), cid) {
		return fmt.Errorf("ipfs link not pinned locally")
	}

	return nil
}

func GetIpfsFile(
	cid string,
) (fileData []byte, err error) {
	url := fmt.Sprintf("http://localhost:5001/api/v0/cat?arg=%s", cid)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	fileData, err = ioutil.ReadAll(resp.Body)
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

func IsJson(s string) bool {
	var js interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}
