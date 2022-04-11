package handler

import (
	"net/http"

	"github.com/sealsurlaw/gouvre/errs"
	"github.com/sealsurlaw/gouvre/request"
	"github.com/sealsurlaw/gouvre/response"
)

func (h *Handler) UploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		h.uploadFile(w, r)
		return
	} else {
		response.SendMethodNotFound(w)
		return
	}
}

func (h *Handler) UploadFileWithLink(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		h.uploadFileWithLink(w, r)
		return
	} else {
		response.SendMethodNotFound(w)
		return
	}
}

func (h *Handler) uploadFile(w http.ResponseWriter, r *http.Request) {
	if !h.hasWhitelistedToken(r) {
		response.SendInvalidAuthToken(w)
		return
	}

	if !h.hasWhitelistedIpAddress(r) {
		response.SendError(w, 401, "Not on ip whitelist.", errs.ErrNotAuthorized)
		return
	}

	// filename
	filename, err := request.ParseFilename(r)
	if filename == "" {
		response.SendBadRequest(w, "filename")
		return
	}

	// encryption-secret - optional
	encryptionSecret := request.ParseEncryptionSecret(r)

	// file
	fileData, err := request.ParseFile(r)
	if err != nil {
		response.SendError(w, 500, "Could not parse file", err)
		return
	}

	err = h.writeFile(fileData, filename, encryptionSecret)
	if err != nil {
		response.SendError(w, 500, "Could not write file", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) uploadFileWithLink(w http.ResponseWriter, r *http.Request) {
	token, err := request.ParseTokenFromUrl(r)
	if token == "" {
		response.SendBadRequest(w, "token")
		return
	}

	if h.singleUseUploadTokens[token] != true {
		response.SendInvalidAuthToken(w)
		return
	}

	// filename
	filename, _, encryptionSecret, resolutions, err := h.tokenizer.ParseToken(token)
	if filename == "" || err != nil {
		response.SendInvalidAuthToken(w)
		return
	}

	// secret if not provided from token
	if encryptionSecret == "" {
		encryptionSecret = request.ParseEncryptionSecret(r)
	}

	// file
	fileData, err := request.ParseFile(r)
	if err != nil {
		response.SendError(w, 500, "Could not parse file", err)
		return
	}

	err = h.writeFile(fileData, filename, encryptionSecret)
	if err != nil {
		response.SendError(w, 500, "Could not write file", err)
		return
	}

	for _, resolution := range resolutions {
		// not square
		thumbnailParameters := &ThumbnailParameters{filename, resolution, false, encryptionSecret}
		_, err = h.checkOrCreateThumbnailFile(thumbnailParameters)
		if err != nil {
			continue
		}

		// square
		thumbnailParameters = &ThumbnailParameters{filename, resolution, true, encryptionSecret}
		_, err = h.checkOrCreateThumbnailFile(thumbnailParameters)
		if err != nil {
			continue
		}
	}

	delete(h.singleUseUploadTokens, token)

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusCreated)
}
