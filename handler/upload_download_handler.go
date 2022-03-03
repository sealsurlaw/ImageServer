package handler

import (
	"fmt"
	"net/http"
	"os"

	"github.com/sealsurlaw/gouvre/errs"
	"github.com/sealsurlaw/gouvre/helper"
	"github.com/sealsurlaw/gouvre/request"
	"github.com/sealsurlaw/gouvre/response"
)

func (h *Handler) UploadDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		h.downloadFile(w, r)
		return
	} else if r.Method == "POST" {
		h.uploadFile(w, r)
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

	if !h.hasWhitelistedIpAddress(r) {
		response.SendError(w, 401, "Not on ip whitelist.", errs.ErrNotAuthorized)
		return
	}

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

	filename = h.getProperFilename(filename)
	err = h.createDirectories(filename)
	if err != nil {
		response.SendError(w, 500, "Couldn't create directories.", err)
		return
	}

	contentType := http.DetectContentType(fileData)
	if !helper.IsSupportedContentType(contentType) {
		msg := fmt.Sprintf("Content type %s not supported.", contentType)
		response.SendError(w, 400, msg, errs.ErrInvalidContentType)
		return
	}

	fullFilePath := h.makeFullFilePath(filename)

	// delete files from dep file including itself
	err = h.deleteDepFiles(fullFilePath)
	if err != nil {
		response.SendError(w, 500, "Couldn't delete dependency files.", err)
		return
	}

	if h.tryEncryptFile(&fileData, encryptionSecret) != nil {
		response.SendError(w, 500, "Couldn't encrypt file.", err)
		return
	}

	// write file
	err = os.WriteFile(fullFilePath, fileData, 0600)
	if err != nil {
		response.SendError(w, 500, "Couldn't write file.", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
