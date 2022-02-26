package handler

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/sealsurlaw/ImageServer/helper"
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

	// resolution
	resolutionStr := r.FormValue("resolution")
	resolution, err := strconv.Atoi(resolutionStr)
	if resolutionStr == "" || err != nil {
		response.SendBadRequest(w, "resolution")
		return
	}

	// filename
	filename := r.FormValue("filename")
	if filename == "" {
		response.SendBadRequest(w, "filename")
		return
	}

	// expires - optional
	expiresIn := r.URL.Query().Get("expires")
	expiresDuration, err := time.ParseDuration(expiresIn)
	if err != nil {
		expiresDuration = 24 * time.Hour
	}

	// cropped - optional
	croppedStr := r.URL.Query().Get("cropped")
	cropped, err := strconv.ParseBool(croppedStr)
	if croppedStr == "" || err != nil {
		cropped = false
	}

	// open file to make sure it exists
	thumbnailFilename := h.getThumbnailFilename(filename, resolution, cropped)
	fullFileName := fmt.Sprintf("%s/%s", h.BasePath, thumbnailFilename)
	file, err := os.Open(fullFileName)
	if err != nil {
		// if not found, attempt to make it
		err = h.createThumbnail(filename, resolution, cropped)
		if err != nil {
			response.SendError(w, 500, "Couldn't create thumbnail.", err)
			return
		}
		file, err = os.Open(fullFileName)
		if err != nil {
			response.SendCouldntFindImage(w, err)
			return
		}
	}
	defer file.Close()

	// create and add link to link store
	expiresAt, token, err := h.tryToAddLink(fullFileName, expiresDuration)
	if err != nil {
		response.SendError(w, 500, err.Error(), err)
		return
	}

	url := fmt.Sprintf("%s/link/%d", h.BaseUrl, token)

	response.SendJson(w, &response.GetLinkResponse{
		Url:       url,
		ExpiresAt: expiresAt,
	})
}

func (h *Handler) createThumbnail(filename string, resolution int, cropped bool) error {
	// open file
	if h.hashFilename {
		filename = helper.CalculateHash(filename)
	}
	fullFileName := fmt.Sprintf("%s/%s", h.BasePath, filename)
	file, err := os.Open(fullFileName)
	if err != nil {
		return err
	}
	defer file.Close()

	// create thumbnail
	thumbFile, err := helper.CreateThumbnail(file, resolution, cropped, h.thumbnailQuality)
	if err != nil {
		return err
	}

	// write file
	fileData, err := ioutil.ReadAll(thumbFile)
	if err != nil {
		return err
	}

	thumbnailFilename := h.getThumbnailFilename(filename, resolution, cropped)
	fullFileName = fmt.Sprintf("%s/%s", h.BasePath, thumbnailFilename)
	err = os.WriteFile(fullFileName, fileData, 0600)
	if err != nil {
		return err
	}

	return nil
}

func (h *Handler) getThumbnailFilename(filename string, resolution int, cropped bool) string {
	var thumbnailFilename string
	if h.hashFilename {
		filename += strconv.Itoa(resolution)
		if cropped {
			filename += "crop"
		}
		thumbnailFilename = helper.CalculateHash(filename)
	} else {
		thumbnailFilename = fmt.Sprintf("%s_%d", filename, resolution)
		if cropped {
			thumbnailFilename += "_crop"
		}
	}

	return thumbnailFilename
}
