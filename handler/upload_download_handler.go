package handler

import (
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"

	"github.com/sealsurlaw/ImageServer/errs"
	"github.com/sealsurlaw/ImageServer/helper"
	"github.com/sealsurlaw/ImageServer/request"
	"github.com/sealsurlaw/ImageServer/response"
)

func (h *Handler) UploadDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		h.downloadFile(w, r)
		return
	} else if r.Method == "POST" {
		h.uploadFile(w, r)
		return
	} else {
		response.SendMethodNotFound(w)
		return
	}
}

func (h *Handler) downloadFile(w http.ResponseWriter, r *http.Request) {
	if !h.hasWhitelistedToken(r) {
		response.SendInvalidAuthToken(w)
		return
	}

	if !h.hasWhitelistedIpAddress(r) {
		response.SendError(w, 401, "Not on ip whitelist.", errs.ErrNotAuthorized)
		return
	}

	// filename
	filename, err := request.ParseFilenameFromUrl(r)
	if err != nil {
		response.SendBadRequest(w, "filename")
	}

	// open file
	filename = h.getProperFilename(filename)
	fullFileName := h.makeFullFilename(filename)
	file, err := os.Open(fullFileName)
	if err != nil {
		response.SendCouldntFindImage(w, err)
		return
	}
	defer file.Close()

	response.SendImage(w, file)
}

func (h *Handler) uploadFile(w http.ResponseWriter, r *http.Request) {
	if !h.hasWhitelistedToken(r) {
		response.SendInvalidAuthToken(w)
		return
	}

	if !h.hasWhitelistedIpAddress(r) {
		response.SendError(w, 401, "Not on ip whitelist.", errs.ErrNotAuthorized)
		return
	}

	// filename
	filename := r.FormValue("filename")
	if filename == "" {
		response.SendBadRequest(w, "filename")
		return
	}
	filename = h.getProperFilename(filename)

	// file
	r.ParseMultipartForm(math.MaxInt64)
	file, _, err := r.FormFile("file")
	if err != nil {
		response.SendError(w, 400, "Error getting file.", err)
		return
	}
	defer file.Close()

	err = h.createDirectories(filename)
	if err != nil {
		response.SendError(w, 500, "Couldn't create directories.", err)
		return
	}

	fileData, err := ioutil.ReadAll(file)
	if err != nil {
		response.SendError(w, 400, "Couldn't parse file.", err)
		return
	}

	contentType := http.DetectContentType(fileData)
	if !helper.IsSupportedContentType(contentType) {
		msg := fmt.Sprintf("Content type %s not supported.", contentType)
		response.SendError(w, 400, msg, errs.ErrInvalidContentType)
		return
	}

	fullFilename := h.makeFullFilename(filename)

	// delete files from dep file including itself
	err = h.deleteDepFiles(fullFilename)
	if err != nil {
		response.SendError(w, 500, "Couldn't delete dependency files.", err)
		return
	}

	// write file
	err = os.WriteFile(fullFilename, fileData, 0600)
	if err != nil {
		response.SendError(w, 500, "Couldn't write file.", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
