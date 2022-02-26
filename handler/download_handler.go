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

	if !h.isAuthorized(r) {
		msg := "Not on the authorized IP Address list."
		res := errs.NewErrorResponse(http.StatusNotFound, msg, errs.ErrNotAuthorized)
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

func (h *Handler) isAuthorized(r *http.Request) bool {
	ip := helper.GetIP(r)
	auth := false
	for _, ipAddress := range h.AuthorizedIpAddresses {
		if ipAddress.Equal(ip) {
			auth = true
			break
		}
	}

	return auth
}
