package handler

import (
	"net/http"

	"github.com/sealsurlaw/gouvre/errs"
	"github.com/sealsurlaw/gouvre/helper"
	"github.com/sealsurlaw/gouvre/request"
	"github.com/sealsurlaw/gouvre/response"
)

func (h *Handler) CreateLink(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		h.createLink(w, r)
		return
	} else {
		response.SendMethodNotFound(w)
		return
	}
}

func (h *Handler) CreateUploadLink(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		h.createUploadLink(w, r)
		return
	} else {
		response.SendMethodNotFound(w)
		return
	}
}

func (h *Handler) GetImageFromTokenLink(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.getImageFromTokenLink(w, r)
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
	req := request.CreateLinkRequest{}
	err := request.ParseJson(r, &req)
	if err != nil {
		response.SendError(w, 400, "Could not parse json request.", err)
		return
	}

	// optional queries
	expiresAt := request.ParseExpires(r)

	filename, err := h.checkFileExists(req.Filename, req.EncryptionSecret)
	if err != nil {
		response.SendCouldntFindImage(w, err)
		return
	}

	token, err := h.tokenizer.CreateToken(filename, expiresAt, req.EncryptionSecret, nil)
	if err != nil {
		response.SendError(w, 500, "Couldn't create token.", err)
		return
	}

	response.SendJson(w, &response.GetLinkResponse{
		Url:       h.makeTokenUrl(token),
		ExpiresAt: expiresAt,
	}, 200)
}

func (h *Handler) createUploadLink(w http.ResponseWriter, r *http.Request) {
	if !h.hasWhitelistedToken(r) {
		response.SendInvalidAuthToken(w)
		return
	}

	if !h.hasWhitelistedIpAddress(r) {
		response.SendError(w, 401, "Not on ip whitelist.", errs.ErrNotAuthorized)
		return
	}

	// filename
	req := request.CreateUploadLinkRequest{}
	err := request.ParseJson(r, &req)
	if err != nil {
		response.SendError(w, 400, "Could not parse json request.", err)
		return
	}

	// optional queries
	expiresAt := request.ParseExpires(r)

	token, err := h.tokenizer.CreateToken(req.Filename, expiresAt, req.EncryptionSecret, req.Resolutions)
	if err != nil {
		response.SendError(w, 500, "Couldn't create token.", err)
		return
	}

	h.singleUseUploadTokens[token] = true

	response.SendJson(w, &response.GetLinkResponse{
		Url:       h.makeUploadTokenUrl(token),
		ExpiresAt: expiresAt,
	}, 200)
}

func (h *Handler) getImageFromTokenLink(w http.ResponseWriter, r *http.Request) {
	// token
	token, err := request.ParseTokenFromUrl(r)
	if err != nil {
		response.SendBadRequest(w, "token")
	}

	filename, expiresAt, encryptionSecret, _, err := h.tokenizer.ParseToken(token)
	if err != nil {
		response.SendError(w, 500, "Couldn't parse token.", err)
		return
	}

	if encryptionSecret == "" {
		encryptionSecret = request.ParseEncryptionSecretFromQuery(r)
	}

	// open file
	fullFilePath := h.makeFullFilePath(filename)
	fileData, err := helper.OpenFile(fullFilePath)
	if err != nil {
		response.SendError(w, 500, "Couldn't read file data.", err)
		return
	}

	// if secret is wrong, just return encrypted file
	_ = h.tryDecryptFile(&fileData, encryptionSecret)

	response.SendImage(w, fileData, expiresAt)
}
