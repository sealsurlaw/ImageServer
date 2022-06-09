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
	if !h.hasWhitelistedToken(r) {
		response.SendInvalidAuthToken(w)
		return
	}

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

	// encryption-secret
	encryptionSecret := request.ParseEncryptionSecretFromQuery(r)

	// open file
	filename = h.getProperFilename(filename)
	fullFilePath := h.makeFullFilePath(filename)
	fileData, err := helper.OpenFile(fullFilePath)
	if err != nil {
		response.SendCouldntFindImage(w, err)
		return
	}

	_ = h.tryDecryptFile(&fileData, encryptionSecret)

	response.SendImage(w, fileData, nil)
}
