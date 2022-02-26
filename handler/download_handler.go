package handler

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/sealsurlaw/ImageServer/errs"
	"github.com/sealsurlaw/ImageServer/helper"
)

func (h *Handler) Download(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		msg := "Method not found."
		res := errs.NewErrorResponse(http.StatusBadRequest, msg)
		helper.SendJson(w, res)
		return
	}

	pathArr := strings.Split(r.URL.Path, "/")
	filename := pathArr[len(pathArr)-1]

	fullFileName := fmt.Sprintf("%s/%s", h.BasePath, filename)
	file, err := os.Open(fullFileName)
	if err != nil {
		msg := "Couldn't find file."
		res := errs.NewErrorResponse(http.StatusNotFound, msg, err)
		helper.SendJson(w, res)
		return
	}
	defer file.Close()

	helper.SendImage(w, file)
}
