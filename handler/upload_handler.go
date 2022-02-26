package handler

import (
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"strings"

	"github.com/sealsurlaw/ImageServer/errs"
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

	// check if directories need to be created
	if strings.Contains(filename, "/") {
		filenameSplit := strings.Split(filename, "/")
		path := h.BasePath
		for _, dir := range filenameSplit[:len(filenameSplit)-1] {
			path += "/" + dir
			_, err := os.ReadDir(path)
			if err == nil {
				continue
			}
			err = os.Mkdir(path, 0700)
			if err != nil {
				response.SendError(w, 400, "Error creating directories.", err)
				return
			}
		}
	}

	// write file
	fileData, err := ioutil.ReadAll(file)
	if err != nil {
		response.SendError(w, 400, "Couldn't parse file.", err)
		return
	}

	contentType := http.DetectContentType(fileData)
	if !strings.Contains(contentType, "image/") {
		msg := fmt.Sprintf("Content type %s not supported.", contentType)
		response.SendError(w, 400, msg, errs.ErrInvalidContentType)
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
