package response

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sealsurlaw/gouvre/errs"
	"github.com/sealsurlaw/gouvre/helper"
)

type UploadImageResponse struct {
	Cid string `json:"cid"`
}

type GetLinkResponse struct {
	ExpiresAt *time.Time `json:"expiresAt"`
	Url       string     `json:"url"`
}

type GetThumbnailLinkResponse struct {
	ExpiresAt *time.Time `json:"expiresAt"`
	Url       string     `json:"url"`
}

type GetThumbnailLinksResponse struct {
	ExpiresAt     *time.Time        `json:"expiresAt"`
	FilenameToUrl map[string]string `json:"filenameToUrl"`
}

func SendJson(w http.ResponseWriter, obj interface{}, statusCode int) {
	j, err := json.Marshal(obj)
	if err != nil {
		fmt.Print(err.Error())
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(j)
}

func SendFile(w http.ResponseWriter, fileData []byte, expiresAt *time.Time) {
	cacheControl := "public, max-age=86400"
	if expiresAt != nil {
		cacheControl = fmt.Sprintf("public, max-age=%d", int(time.Until(*expiresAt).Seconds()))
	}
	w.Header().Add("Cache-Control", cacheControl)
	contentType := http.DetectContentType(fileData)
	if strings.Contains(contentType, "text/plain") && helper.IsJson(string(fileData)) {
		contentType = "application/json"
	}
	w.Header().Add("Content-Type", contentType)
	w.Header().Add("Content-Length", strconv.Itoa(len(fileData)))
	w.Write(fileData)
}

func SendError(w http.ResponseWriter, code int, msg string, err ...error) {
	res := errs.NewErrorResponse(code, msg, err...)
	SendJson(w, res, code)
}

func SendBadRequest(w http.ResponseWriter, missingField string) {
	msg := fmt.Sprintf("No %s provided.", missingField)
	SendError(w, http.StatusBadRequest, msg)
}

func SendMethodNotFound(w http.ResponseWriter) {
	SendError(w, http.StatusBadRequest, "Method not found.")
}

func SendInvalidAuthToken(w http.ResponseWriter) {
	SendError(w, http.StatusNotFound, "Invalid auth token.", errs.ErrNotAuthorized)
}

func SendCouldntFindImage(w http.ResponseWriter, err error) {
	SendError(w, http.StatusNotFound, "Couldn't find image.", err)
}
