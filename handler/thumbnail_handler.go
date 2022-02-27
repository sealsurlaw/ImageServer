package handler

import (
	"net/http"

	"github.com/sealsurlaw/ImageServer/handler/request"
	"github.com/sealsurlaw/ImageServer/response"
)

func (h *Handler) Thumbnail(w http.ResponseWriter, r *http.Request) {
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

	// filename
	filename, err := request.ParseFilename(r)
	if err != nil {
		response.SendBadRequest(w, "filename")
		return
	}

	// resolution
	resolution, err := request.ParseResolution(r)
	if err != nil {
		response.SendBadRequest(w, "resolution")
		return
	}

	// optional queries
	cropped := request.ParseCropped(r)
	expiresAt := request.ParseExpires(r)

	fullFilename, err := h.checkOrCreateThumbnailFile(filename, resolution, cropped)
	if err != nil {
		response.SendError(w, 500, err.Error(), err)
		return
	}

	// create and add link to link store
	token, err := h.tryToAddLink(fullFilename, expiresAt)
	if err != nil {
		response.SendError(w, 500, err.Error(), err)
		return
	}

	response.SendJson(w, &response.GetLinkResponse{
		Url:       h.makeTokenUrl(token),
		ExpiresAt: expiresAt,
	})
}
