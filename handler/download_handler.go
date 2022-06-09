package handler

import (
	"net/http"

	"github.com/sealsurlaw/gouvre/helper"
	"github.com/sealsurlaw/gouvre/request"
	"github.com/sealsurlaw/gouvre/response"
)

func (h *Handler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.downloadFile(w, r)
		return
	} else {
		response.SendMethodNotFound(w)
		return
	}
}

func (h *Handler) downloadFile(w http.ResponseWriter, r *http.Request) {
	// if !h.hasWhitelistedToken(r) {
	// 	response.SendInvalidAuthToken(w)
	// 	return
	// }

	// if !h.hasWhitelistedIpAddress(r) {
	// 	response.SendError(w, 401, "Not on ip whitelist.", errs.ErrNotAuthorized)
	// 	return
	// }

	// filename
	filename, err := request.ParseFilenameFromUrl(r)
	if err != nil {
		response.SendBadRequest(w, "filename")
		return
	}

	// optional queries
	square := request.ParseSquare(r)
	resolution := request.ParseResolutionFromQuery(r)
	encryptionSecret := request.ParseEncryptionSecretFromQuery(r)

	if resolution != nil {
		thumbnailParameters := &ThumbnailParameters{filename, *resolution, square, encryptionSecret}
		thumbnailFilename, err := h.checkOrCreateThumbnailFile(thumbnailParameters)
		if err != nil {
			response.SendError(w, 500, "Couldn't check/create thumbnail file.", err)
			return
		}
		filename = thumbnailFilename
	} else {
		filename = h.getProperFilename(filename)
	}

	// open file
	fullFilePath := h.makeFullFilePath(filename)
	fileData, err := helper.OpenFile(fullFilePath)
	if err != nil {
		response.SendCouldntFindImage(w, err)
		return
	}

	_ = h.tryDecryptFile(&fileData, encryptionSecret)

	response.SendImage(w, fileData, nil)
}
