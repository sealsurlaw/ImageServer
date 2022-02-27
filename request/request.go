package request

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sealsurlaw/ImageServer/errs"
)

func ParseCropped(r *http.Request) bool {
	croppedStr := r.URL.Query().Get("cropped")
	cropped, err := strconv.ParseBool(croppedStr)
	if croppedStr == "" || err != nil {
		return false
	}

	return cropped
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

func ParseFilename(r *http.Request) (string, error) {
	filename := r.FormValue("filename")
	if filename == "" {
		return "", errs.ErrBadRequest
	}

	return filename, nil
}

func ParseFilenameFromUrl(r *http.Request) (string, error) {
	pathArr := strings.Split(r.URL.Path, "download/")
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

func ParseTokenFromUrl(r *http.Request) (int64, error) {
	pathArr := strings.Split(r.URL.Path, "/")
	tokenStr := pathArr[len(pathArr)-1]
	token, err := strconv.ParseInt(tokenStr, 10, 64)
	if tokenStr == "" || err != nil {
		return 0, errs.ErrBadRequest
	}

	return token, nil
}
