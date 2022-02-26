package handler

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sealsurlaw/ImageServer/errs"
	"github.com/sealsurlaw/ImageServer/helper"
	"github.com/sealsurlaw/ImageServer/linkstore"
	"github.com/sealsurlaw/ImageServer/response"
)

func (h *Handler) Link(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		h.createLink(w, r)
		return
	} else if r.Method == "GET" {
		h.getImageFromToken(w, r)
		return
	} else {
		msg := "Method not found."
		res := errs.NewErrorResponse(http.StatusBadRequest, msg)
		helper.SendJson(w, res)
		return
	}
}

func (h *Handler) createLink(w http.ResponseWriter, r *http.Request) {
	if !h.hasWhitelistedToken(r) {
		msg := "Invalid auth token."
		res := errs.NewErrorResponse(http.StatusNotFound, msg, errs.ErrNotAuthorized)
		helper.SendJson(w, res)
		return
	}

	pathArr := strings.Split(r.URL.Path, "/")
	filename := pathArr[len(pathArr)-1]

	expiresIn := r.URL.Query().Get("expires")

	if filename == "" {
		msg := "No filename provided"
		res := errs.NewErrorResponse(http.StatusBadRequest, msg)
		helper.SendJson(w, res)
		return
	}

	fullFileName := fmt.Sprintf("%s/%s", h.BasePath, filename)
	file, err := os.Open(fullFileName)
	if err != nil {
		msg := "Couldn't find image."
		res := errs.NewErrorResponse(http.StatusNotFound, msg, err)
		helper.SendJson(w, res)
		return
	}
	defer file.Close()

	var expiresDuration time.Duration
	expiresDuration, err = time.ParseDuration(expiresIn)
	if err != nil {
		expiresDuration = 24 * time.Hour
	}

	expiresAt := time.Now().Add(expiresDuration)
	token := rand.Int63()

	h.LinkStore.AddLink(token, &linkstore.Link{
		FullFilename: fullFileName,
		ExpiresAt:    &expiresAt,
	})

	url := fmt.Sprintf("%s/link/%d", h.BaseUrl, token)

	helper.SendJson(w, &response.GetLinkResponse{
		Url:       url,
		ExpiresAt: &expiresAt,
	})
}

func (h *Handler) getImageFromToken(w http.ResponseWriter, r *http.Request) {
	pathArr := strings.Split(r.URL.Path, "/")
	tokenStr := pathArr[len(pathArr)-1]

	if tokenStr == "" {
		msg := "No token provided"
		res := errs.NewErrorResponse(http.StatusBadRequest, msg)
		helper.SendJson(w, res)
		return
	}

	token, _ := strconv.ParseInt(tokenStr, 10, 64)

	link, err := h.LinkStore.GetLink(token)
	if err != nil {
		msg := err.Error()
		res := errs.NewErrorResponse(http.StatusBadRequest, msg, err)
		helper.SendJson(w, res)
		return
	}

	file, err := os.Open(link.FullFilename)
	if err != nil {
		msg := "Couldn't find image."
		res := errs.NewErrorResponse(http.StatusNotFound, msg, err)
		helper.SendJson(w, res)
		return
	}
	defer file.Close()

	helper.SendImage(w, file)
}
