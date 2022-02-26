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
	if r.Method != "POST" {
		msg := "Method not found."
		res := errs.NewErrorResponse(http.StatusBadRequest, msg)
		helper.SendJson(w, res)
		return
	}

	if !h.hasWhitelistedToken(r) {
		msg := "Invalid auth token."
		res := errs.NewErrorResponse(http.StatusNotFound, msg, errs.ErrNotAuthorized)
		helper.SendJson(w, res)
		return
	}

	filename := r.FormValue("filename")

	r.ParseMultipartForm(math.MaxInt64)

	file, _, err := r.FormFile("file")
	if err != nil {
		msg := "Error getting file."
		res := errs.NewErrorResponse(http.StatusBadRequest, msg, err)
		helper.SendJson(w, res)
		return
	}
	defer file.Close()

	fileData, err := ioutil.ReadAll(file)
	if err != nil {
		msg := "Couldn't parse file."
		res := errs.NewErrorResponse(http.StatusBadRequest, msg, err)
		helper.SendJson(w, res)
		return
	}

	fullFileName := fmt.Sprintf("%s/%s", h.BasePath, filename)
	err = os.WriteFile(fullFileName, fileData, 0666)
	if err != nil {
		msg := "Couldn't write file."
		res := errs.NewErrorResponse(http.StatusInternalServerError, msg, err)
		helper.SendJson(w, res)
		return
	}

	helper.SendJson(w, &response.UploadImageResponse{
		Filename: filename,
	})
}
