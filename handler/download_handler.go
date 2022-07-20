package handler

import (
	"fmt"
	"net/http"

	"github.com/sealsurlaw/gouvre/errs"
	"github.com/sealsurlaw/gouvre/helper"
	"github.com/sealsurlaw/gouvre/request"
	"github.com/sealsurlaw/gouvre/response"
)

func (h *Handler) DownloadImage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.downloadImage(w, r)
		return
	} else {
		response.SendMethodNotFound(w)
		return
	}
}

func (h *Handler) downloadImage(w http.ResponseWriter, r *http.Request) {
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

	var properFilename string
	if resolution != nil {
		thumbnailParameters := &ThumbnailParameters{filename, *resolution, square, encryptionSecret}
		thumbnailFilename, err := h.checkOrCreateThumbnailFile(thumbnailParameters)
		if err != nil {
			// Bypass thumbnail creation for Gifs
			if err == errs.ErrGif {
				thumbnailFilename = h.getProperFilename(filename)
			} else {
				response.SendError(w, 500, "Couldn't check/create thumbnail file.", err)
				return
			}
		}
		properFilename = thumbnailFilename
	} else {
		properFilename = h.getProperFilename(filename)
	}

	// open file
	fullFilePath := h.makeFullFilePath(properFilename)
	fileData, err := helper.OpenFile(fullFilePath)
	if err != nil {
		response.SendCouldntFindImage(w, err)
		return
	}
	_ = h.tryDecryptFile(&fileData, encryptionSecret)

	// For gif thumbnails, abort and give full version
	// Gif thumbnails currently do not support animations
	contentType := http.DetectContentType(fileData)
	fmt.Println(contentType)
	if contentType == "image/gif" && resolution != nil {
		properFilename = h.getProperFilename(filename)
		fullFilePath = h.makeFullFilePath(properFilename)
		fileData, err = helper.OpenFile(fullFilePath)
		if err != nil {
			response.SendCouldntFindImage(w, err)
			return
		}
		_ = h.tryDecryptFile(&fileData, encryptionSecret)
	}

	response.SendFile(w, fileData, nil)
}
