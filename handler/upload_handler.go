package handler

import (
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"

	"github.com/sealsurlaw/ImageServer/response"
)

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		response.SendMethodNotFound(w)
		return
	}

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

	// file
	r.ParseMultipartForm(math.MaxInt64)
	file, _, err := r.FormFile("file")
	if err != nil {
		response.SendError(w, 400, "Error getting file.", err)
		return
	}
	defer file.Close()

	// write file
	fileData, err := ioutil.ReadAll(file)
	if err != nil {
		response.SendError(w, 400, "Couldn't parse file.", err)
		return
	}
	fullFileName := fmt.Sprintf("%s/%s", h.BasePath, filename)
	err = os.WriteFile(fullFileName, fileData, 0600)
	if err != nil {
		response.SendError(w, 500, "Couldn't write file.", err)
		return
	}

	response.SendJson(w, &response.UploadImageResponse{
		Filename: filename,
	})
}
