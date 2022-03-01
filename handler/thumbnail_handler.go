package handler

import (
	"net/http"

	"github.com/sealsurlaw/gouvre/errs"
	"github.com/sealsurlaw/gouvre/request"
	"github.com/sealsurlaw/gouvre/response"
)

type ThumbnailParameters struct {
	Filename   string
	Resolution int
	Cropped    bool
}

func (h *Handler) Thumbnails(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		h.getThumbnailLink(w, r)
		return
	} else {
		response.SendMethodNotFound(w)
		return
	}
}

func (h *Handler) getThumbnailLink(w http.ResponseWriter, r *http.Request) {
	if !h.hasWhitelistedToken(r) {
		response.SendInvalidAuthToken(w)
		return
	}

	if !h.hasWhitelistedIpAddress(r) {
		response.SendError(w, 401, "Not on ip whitelist.", errs.ErrNotAuthorized)
		return
	}

	// request json
	req := request.ThumbnailRequest{}
	err := request.ParseJson(r, &req)
	if err != nil {
		response.SendError(w, 400, "Could not parse json request.", err)
		return
	}

	// optional queries
	square := request.ParseSquare(r)
	expiresAt := request.ParseExpires(r)

	thumbnailParameters := &ThumbnailParameters{req.Filename, req.Resolution, square}
	thumbnailFilename, err := h.checkOrCreateThumbnailFile(thumbnailParameters)
	if err != nil {
		response.SendError(w, 500, "Couldn't check/create thumbnail file.", err)
		return
	}

	token, err := h.tokenizer.CreateToken(thumbnailFilename, expiresAt)
	if err != nil {
		response.SendError(w, 500, "Couldn't create token.", err)
		return
	}

	response.SendJson(w, &response.GetLinkResponse{
		Url:       h.makeTokenUrl(token),
		ExpiresAt: expiresAt,
	}, 200)
}
