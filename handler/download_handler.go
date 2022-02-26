package handler

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/sealsurlaw/ImageServer/response"
)

func (h *Handler) Download(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		response.SendMethodNotFound(w)
		return
	}

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

	// open file
	fullFileName := fmt.Sprintf("%s/%s", h.BasePath, filename)
	file, err := os.Open(fullFileName)
	if err != nil {
		response.SendCouldntFindImage(w, err)
		return
	}
	defer file.Close()

	response.SendImage(w, file)
}
