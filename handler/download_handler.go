package handler

import (
	"net/http"
	"os"

	"github.com/sealsurlaw/ImageServer/request"
	"github.com/sealsurlaw/ImageServer/response"
)

func (h *Handler) Download(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		h.downloadFile(w, r)
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

	// filename
	filename, err := request.ParseFilenameFromUrl(r)
	if err != nil {
		response.SendBadRequest(w, "filename")
	}

	// open file
	fullFileName := h.makeFullFilename(filename)
	file, err := os.Open(fullFileName)
	if err != nil {
		response.SendCouldntFindImage(w, err)
		return
	}
	defer file.Close()

	response.SendImage(w, file)
}
