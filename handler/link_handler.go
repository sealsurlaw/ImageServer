package handler

import (
	"net/http"
	"os"

	"github.com/sealsurlaw/ImageServer/errs"
	"github.com/sealsurlaw/ImageServer/request"
	"github.com/sealsurlaw/ImageServer/response"
)

func (h *Handler) Links(w http.ResponseWriter, r *http.Request) {
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

	if !h.hasWhitelistedIpAddress(r) {
		response.SendError(w, 401, "Not on ip whitelist.", errs.ErrNotAuthorized)
		return
	}

	// filename
	req := request.LinkRequest{}
	err := request.ParseJson(r, &req)
	if err != nil {
		response.SendError(w, 400, "Could not parse json request.", err)
		return
	}

	// optional queries
	expiresAt := request.ParseExpires(r)

	fullFilename, err := h.checkFileExists(req.Filename)
	if err != nil {
		response.SendCouldntFindImage(w, err)
		return
	}

	// create and add link to link store
	token, err := h.tryToAddLink(fullFilename, expiresAt)
	if err != nil {
		response.SendError(w, 500, err.Error(), err)
	}

	response.SendJson(w, &response.GetLinkResponse{
		Url:       h.makeTokenUrl(token),
		ExpiresAt: expiresAt,
	}, 200)
}

func (h *Handler) getImageFromToken(w http.ResponseWriter, r *http.Request) {
	// token
	token, err := request.ParseTokenFromUrl(r)
	if err != nil {
		response.SendBadRequest(w, "token")
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
