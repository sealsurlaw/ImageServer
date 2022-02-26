package handler

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sealsurlaw/ImageServer/helper"
	"github.com/sealsurlaw/ImageServer/response"
)

func (h *Handler) Thumbnail(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		h.createThumbnail(w, r)
		return
	} else if r.Method == "GET" {
		h.getThumbnailLink(w, r)
		return
	} else {
		response.SendMethodNotFound(w)
		return
	}
}

func (h *Handler) createThumbnail(w http.ResponseWriter, r *http.Request) {
	if !h.hasWhitelistedToken(r) {
		response.SendInvalidAuthToken(w)
		return
	}

	// filename
	pathArr := strings.Split(r.URL.Path, "/")
	filename := pathArr[len(pathArr)-1]
	if filename == "" {
		response.SendBadRequest(w, "filename")
		return
	}

	// resolution
	resolutionStr := r.FormValue("resolution")
	resolution, err := strconv.Atoi(resolutionStr)
	if resolutionStr == "" || err != nil {
		response.SendBadRequest(w, "resolution")
		return
	}

	// open file
	fullFileName := fmt.Sprintf("%s/%s", h.BasePath, filename)
	file, err := os.Open(fullFileName)
	if err != nil {
		response.SendCouldntFindImage(w, err)
		return
	}
	defer file.Close()

	// create thumbnail
	thumbFile, err := helper.CreateThumbnail(file, resolution)
	if err != nil {
		response.SendError(w, 500, "Couldn't create thumbnail.", err)
		return
	}

	// write file
	fileData, err := ioutil.ReadAll(thumbFile)
	if err != nil {
		response.SendError(w, 400, "Couldn't parse file.", err)
		return
	}
	fullFileName = fmt.Sprintf("%s/%s_%d", h.BasePath, filename, resolution)
	err = os.WriteFile(fullFileName, fileData, 0600)
	if err != nil {
		response.SendError(w, 500, "Couldn't write file.", err)
		return
	}
}

func (h *Handler) getThumbnailLink(w http.ResponseWriter, r *http.Request) {
	if !h.hasWhitelistedToken(r) {
		response.SendInvalidAuthToken(w)
		return
	}

	// filename
	pathArr := strings.Split(r.URL.Path, "/")
	filename := pathArr[len(pathArr)-1]
	if filename == "" {
		response.SendBadRequest(w, "filename")
		return
	}

	// resolution
	resolution := r.FormValue("resolution")
	if resolution == "" {
		response.SendBadRequest(w, "resolution")
		return
	}

	// expires - optional
	expiresIn := r.URL.Query().Get("expires")
	var expiresDuration time.Duration
	expiresDuration, err := time.ParseDuration(expiresIn)
	if err != nil {
		expiresDuration = 24 * time.Hour
	}

	// open file to make sure it exists
	fullFileName := fmt.Sprintf("%s/%s_%s", h.BasePath, filename, resolution)
	file, err := os.Open(fullFileName)
	if err != nil {
		response.SendCouldntFindImage(w, err)
		return
	}
	defer file.Close()

	// create and add link to link store
	expiresAt, token, err := h.tryToAddLink(fullFileName, expiresDuration)
	if err != nil {
		response.SendError(w, 500, err.Error(), err)
	}

	url := fmt.Sprintf("%s/link/%d", h.BaseUrl, token)

	response.SendJson(w, &response.GetLinkResponse{
		Url:       url,
		ExpiresAt: expiresAt,
	})
}
