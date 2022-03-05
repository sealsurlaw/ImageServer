package handler

import (
	"net/http"

	"github.com/sealsurlaw/gouvre/errs"
	"github.com/sealsurlaw/gouvre/request"
	"github.com/sealsurlaw/gouvre/response"
)

type ThumbnailParameters struct {
	Filename         string
	Resolution       int
	Cropped          bool
	EncryptionSecret string
}

func (h *Handler) CreateThumbnailLink(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		h.createThumbnailLink(w, r)
		return
	} else {
		response.SendMethodNotFound(w)
		return
	}
}

func (h *Handler) CreateBatchThumbnailLinks(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		h.createBatchThumbnailLinks(w, r)
		return
	} else {
		response.SendMethodNotFound(w)
		return
	}
}

func (h *Handler) createThumbnailLink(w http.ResponseWriter, r *http.Request) {
	if !h.hasWhitelistedToken(r) {
		response.SendInvalidAuthToken(w)
		return
	}

	if !h.hasWhitelistedIpAddress(r) {
		response.SendError(w, 401, "Not on ip whitelist.", errs.ErrNotAuthorized)
		return
	}

	// request json
	req := request.CreateThumbnailLinkRequest{}
	err := request.ParseJson(r, &req)
	if err != nil {
		response.SendError(w, 400, "Could not parse json request.", err)
		return
	}

	// optional queries
	square := request.ParseSquare(r)
	expiresAt := request.ParseExpires(r)

	thumbnailParameters := &ThumbnailParameters{req.Filename, req.Resolution, square, req.Secret}
	thumbnailFilename, err := h.checkOrCreateThumbnailFile(thumbnailParameters)
	if err != nil {
		response.SendError(w, 500, "Couldn't check/create thumbnail file.", err)
		return
	}

	token, err := h.tokenizer.CreateToken(thumbnailFilename, expiresAt, req.Secret, nil)
	if err != nil {
		response.SendError(w, 500, "Couldn't create token.", err)
		return
	}

	response.SendJson(w, &response.GetLinkResponse{
		Url:       h.makeTokenUrl(token),
		ExpiresAt: expiresAt,
	}, 200)
}

func (h *Handler) createBatchThumbnailLinks(w http.ResponseWriter, r *http.Request) {
	if !h.hasWhitelistedToken(r) {
		response.SendInvalidAuthToken(w)
		return
	}

	if !h.hasWhitelistedIpAddress(r) {
		response.SendError(w, 401, "Not on ip whitelist.", errs.ErrNotAuthorized)
		return
	}

	// request json
	req := request.CreateBatchThumbnailLinksRequest{}
	err := request.ParseJson(r, &req)
	if err != nil {
		response.SendError(w, 400, "Could not parse json request.", err)
		return
	}

	// optional queries
	square := request.ParseSquare(r)
	expiresAt := request.ParseExpires(r)

	filenameToUrls := make(map[string]string)
	for _, filename := range req.Filenames {
		thumbnailParameters := &ThumbnailParameters{filename, req.Resolution, square, req.Secret}
		thumbnailFilename, err := h.checkOrCreateThumbnailFile(thumbnailParameters)
		if err != nil {
			continue
		}

		token, err := h.tokenizer.CreateToken(thumbnailFilename, expiresAt, req.Secret, nil)
		if err != nil {
			response.SendError(w, 500, "Couldn't create token.", err)
			return
		}

		url := h.makeTokenUrl(token)
		filenameToUrls[filename] = url
	}

	response.SendJson(w, &response.GetThumbnailLinksResponse{
		ExpiresAt:     expiresAt,
		FilenameToUrl: filenameToUrls,
	}, 200)
}
