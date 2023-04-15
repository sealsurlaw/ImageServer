package handler

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multicodec"
	"github.com/multiformats/go-multihash"
	"github.com/sealsurlaw/gouvre/errs"
	"github.com/sealsurlaw/gouvre/helper"
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

func (h *Handler) UploadImage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		h.uploadImage(w, r)
		return
	} else {
		response.SendMethodNotFound(w)
		return
	}
}

func (h *Handler) UploadImageWithLink(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		h.uploadImageWithLink(w, r)
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

	// encryption-secret - optional
	encryptionSecret := request.ParseEncryptionSecret(r)

	// file
	fileData, err := request.ParseFile(r)
	if err != nil {
		response.SendError(w, 500, "Could not parse file", err)
		return
	}

	var cidStr = ""
	pref := cid.Prefix{
		Version:  1,
		Codec:    uint64(multicodec.Raw),
		MhType:   multihash.SHA2_256,
		MhLength: -1, // default length
	}
	cidData, err := pref.Sum(fileData)
	if err != nil {
		response.SendError(w, 500, "Could not create cid", err)
		return
	}
	cidStr = cidData.String()

	// filename
	filename, _ := request.ParseFilename(r)
	if filename == "" && !h.pinToIpfs {
		if cidData.ByteLen() == 0 {
			filename = uuid.NewString()
		} else {
			filename = cidStr
		}
	}

	if h.pinToIpfs {
		cidStr, err = helper.PinFile(fileData, filename)
		if err != nil {
			response.SendError(w, 500, "Could not pin file to IPFS", err)
			return
		}

		fmt.Printf("Pinned file %s with cid: %s\n", filename, cidStr)

		if filename == "" {
			filename = cidStr
		}
	}

	err = h.writeFile(fileData, filename, encryptionSecret)
	if err != nil {
		response.SendError(w, 500, "Could not write file", err)
		return
	}

	res := &response.UploadImageResponse{
		Cid: cidStr,
	}
	response.SendJson(w, res, http.StatusCreated)
}

func (h *Handler) uploadImage(w http.ResponseWriter, r *http.Request) {
	if !h.hasWhitelistedToken(r) {
		response.SendInvalidAuthToken(w)
		return
	}

	if !h.hasWhitelistedIpAddress(r) {
		response.SendError(w, 401, "Not on ip whitelist.", errs.ErrNotAuthorized)
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

	fileData, err = helper.AutoRotateImage(fileData)
	if err != nil {
		response.SendError(w, 500, "Could not auto-rotate image", err)
		return
	}

	var cidStr = ""
	pref := cid.Prefix{
		Version:  1,
		Codec:    uint64(multicodec.Raw),
		MhType:   multihash.SHA2_256,
		MhLength: -1, // default length
	}
	cidData, err := pref.Sum(fileData)
	if err != nil {
		response.SendError(w, 500, "Could not create cid", err)
		return
	}
	cidStr = cidData.String()

	// filename
	filename, _ := request.ParseFilename(r)
	if filename == "" && !h.pinToIpfs {
		if cidData.ByteLen() == 0 {
			filename = uuid.NewString()
		} else {
			filename = cidStr
		}
	}

	if h.pinToIpfs {
		cidStr, err = helper.PinFile(fileData, filename)
		if err != nil {
			response.SendError(w, 500, "Could not pin file to IPFS", err)
			return
		}

		fmt.Printf("Pinned file %s with cid: %s\n", filename, cidStr)

		if filename == "" {
			filename = cidStr
		}
	}

	err = h.writeImage(fileData, filename, encryptionSecret)
	if err != nil {
		response.SendError(w, 500, "Could not write file", err)
		return
	}

	res := &response.UploadImageResponse{
		Cid: cidStr,
	}
	response.SendJson(w, res, http.StatusCreated)
}

func (h *Handler) uploadImageWithLink(w http.ResponseWriter, r *http.Request) {
	token, _ := request.ParseTokenFromUrl(r)
	if token == "" {
		response.SendBadRequest(w, "token")
		return
	}

	if !h.singleUseUploadTokens[token] {
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

	err = h.writeImage(fileData, filename, encryptionSecret)
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
