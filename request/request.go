package request

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sealsurlaw/gouvre/errs"
)

type GetImageFromTokenLinkRequest struct {
	Secret string `json:"secret"`
}

type CreateLinkRequest struct {
	Filename string `json:"filename"`
	Secret   string `json:"secret"`
}

type CreateUploadLinkRequest struct {
	Filename    string `json:"filename"`
	Secret      string `json:"secret"`
	Resolutions []int  `json:"resolutions"`
}

type CreateThumbnailLinkRequest struct {
	Resolution int    `json:"resolution"`
	Filename   string `json:"filename"`
	Secret     string `json:"secret"`
}

type CreateBatchThumbnailLinksRequest struct {
	Resolution int      `json:"resolution"`
	Filenames  []string `json:"filenames"`
	Secret     string   `json:"secret"`
}

func ParseJson(r *http.Request, obj interface{}) error {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(b, obj)
	if err != nil {
		return err
	}

	return nil
}

func ParseSquare(r *http.Request) bool {
	squareStr := r.URL.Query().Get("square")
	square, err := strconv.ParseBool(squareStr)
	if squareStr == "" || err != nil {
		return false
	}

	return square
}

func ParseExpires(r *http.Request) *time.Time {
	expiresIn := r.URL.Query().Get("expires")
	expiresDuration, err := time.ParseDuration(expiresIn)
	if err != nil {
		expiresDuration = 24 * time.Hour
	}

	expiresAt := time.Now().Add(expiresDuration).UTC()
	return &expiresAt
}

func ParseFile(r *http.Request) ([]byte, error) {
	file, _, err := r.FormFile("file")
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

func ParseFilename(r *http.Request) (string, error) {
	filename := r.FormValue("filename")
	if filename == "" {
		return "", errs.ErrBadRequest
	}

	return filename, nil
}

func ParseEncryptionSecret(r *http.Request) string {
	encryptionSecret := r.FormValue("secret")
	return encryptionSecret
}

func ParseEncryptionSecretFromQuery(r *http.Request) string {
	encryptionSecret := r.URL.Query().Get("secret")
	return encryptionSecret
}

func ParseFilenameFromUrl(r *http.Request) (string, error) {
	pathArr := strings.Split(r.URL.Path, "images/")
	filename := pathArr[len(pathArr)-1]
	if filename == "" {
		return "", errs.ErrBadRequest
	}

	return filename, nil
}

func ParseResolution(r *http.Request) (int, error) {
	resolutionStr := r.FormValue("resolution")
	resolution, err := strconv.Atoi(resolutionStr)
	if resolutionStr == "" || err != nil {
		return 0, errs.ErrBadRequest
	}

	return resolution, nil
}

func ParseResolutionFromQuery(r *http.Request) *int {
	resolution := r.URL.Query().Get("resolution")
	resolutionInt64, err := strconv.ParseInt(resolution, 10, 32)
	if err != nil {
		return nil
	}

	resolutionInt := int(resolutionInt64)
	return &resolutionInt
}

func ParseTokenFromUrl(r *http.Request) (string, error) {
	pathArr := strings.Split(r.URL.Path, "/")
	token := pathArr[len(pathArr)-1]
	if token == "" {
		return "", errs.ErrBadRequest
	}

	return token, nil
}
