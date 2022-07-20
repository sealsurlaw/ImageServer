package handler

import (
	"net/http"
	"strings"

	"github.com/sealsurlaw/gouvre/helper"
	"github.com/sealsurlaw/gouvre/response"
)

func (h *Handler) GetIpfsFile(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.getIpfsFile(w, r)
		return
	} else {
		response.SendMethodNotFound(w)
		return
	}
}

func (h *Handler) getIpfsFile(w http.ResponseWriter, r *http.Request) {
	// filename
	pathArr := strings.Split(r.URL.Path, "ipfs/")
	cid := pathArr[len(pathArr)-1]
	if cid == "" {
		response.SendBadRequest(w, "cid")
		return
	}

	err := helper.IsIpfsFilePinned(cid)
	if err != nil {
		response.SendError(w, 500, "Ipfs link is not locally pinned.", err)
		return
	}

	fileData, err := helper.GetIpfsFile(cid)
	if err != nil {
		response.SendError(w, 500, "Couldn't get ipfs file.", err)
		return
	}

	response.SendFile(w, fileData, nil)
}
