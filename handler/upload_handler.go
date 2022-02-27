package handler

import (
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"

	"github.com/sealsurlaw/ImageServer/errs"
	"github.com/sealsurlaw/ImageServer/helper"
	"github.com/sealsurlaw/ImageServer/response"
)

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		h.uploadFile(w, r)
		return
	} else {
		response.SendMethodNotFound(w)
		return
	}
}

func (h *Handler) uploadFile(w http.ResponseWriter, r *http.Request) {
	if !h.hasWhitelistedToken(r) {
		response.SendInvalidAuthToken(w)
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

	// write file
	fullFilename := h.makeFullFilename(filename)
	err = os.WriteFile(fullFilename, fileData, 0600)
	if err != nil {
		response.SendError(w, 500, "Couldn't write file.", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
