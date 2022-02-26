package handler

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sealsurlaw/ImageServer/helper"
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
		response.SendMethodNotFound(w)
		return
	}
}

func (h *Handler) createLink(w http.ResponseWriter, r *http.Request) {
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
	if h.hashFilename {
		filename = helper.CalculateHash(filename)
	}

	// expires - optional
	expiresIn := r.URL.Query().Get("expires")
	var expiresDuration time.Duration
	expiresDuration, err := time.ParseDuration(expiresIn)
	if err != nil {
		expiresDuration = 24 * time.Hour
	}

	// open file to make sure it exists
	fullFileName := fmt.Sprintf("%s/%s", h.BasePath, filename)
	file, err := os.Open(fullFileName)
	if err != nil {
		response.SendCouldntFindImage(w, err)
		return
	}
	defer file.Close()

	// create and add link to link store
	expiresAt, token, err := h.tryToAddLink(fullFileName, expiresDuration)
	if err != nil {
		response.SendError(w, 500, err.Error(), err)
	}

	url := fmt.Sprintf("%s/link/%d", h.BaseUrl, token)

	response.SendJson(w, &response.GetLinkResponse{
		Url:       url,
		ExpiresAt: expiresAt,
	})
}

func (h *Handler) getImageFromToken(w http.ResponseWriter, r *http.Request) {
	pathArr := strings.Split(r.URL.Path, "/")

	// token
	tokenStr := pathArr[len(pathArr)-1]
	token, err := strconv.ParseInt(tokenStr, 10, 64)
	if tokenStr == "" || err != nil {
		response.SendBadRequest(w, "token")
		return
	}

	// get link from link store
	link, err := h.LinkStore.GetLink(token)
	if err != nil {
		response.SendError(w, 400, err.Error(), err)
		return
	}

	// open file
	file, err := os.Open(link.FullFilename)
	if err != nil {
		response.SendCouldntFindImage(w, err)
		return
	}
	defer file.Close()

	response.SendImage(w, file)
}
