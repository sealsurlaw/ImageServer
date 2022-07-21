package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/sealsurlaw/gouvre/errs"
	"github.com/sealsurlaw/gouvre/helper"
	"github.com/sealsurlaw/gouvre/request"
	"github.com/sealsurlaw/gouvre/response"
)

func (h *Handler) AddJsonToIpfs(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		h.addJsonToIpfs(w, r)
		return
	} else {
		response.SendMethodNotFound(w)
		return
	}
}

func (h *Handler) GetIpfsFile(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.getIpfsFile(w, r)
		return
	} else {
		response.SendMethodNotFound(w)
		return
	}
}

func (h *Handler) addJsonToIpfs(w http.ResponseWriter, r *http.Request) {
	if !h.pinToIpfs {
		response.SendError(w, 400, "Ipfs is not enabled.", nil)
	}

	if !h.hasWhitelistedToken(r) {
		response.SendInvalidAuthToken(w)
		return
	}

	if !h.hasWhitelistedIpAddress(r) {
		response.SendError(w, 401, "Not on ip whitelist.", errs.ErrNotAuthorized)
		return
	}

	var req interface{}
	err := request.ParseJson(r, &req)
	if err != nil {
		response.SendError(w, 400, "Could not parse json request.", err)
		return
	}

	fileData, err := json.MarshalIndent(req, "", "   ")
	if err != nil {
		response.SendError(w, 400, "Could marshal json.", err)
		return
	}

	cid, err := helper.PinFile(fileData, "")
	if err != nil {
		response.SendError(w, 500, "Could not pin file to IPFS", err)
		return
	}

	fmt.Printf("Pinned file with cid: %s\n", cid)

	res := &response.UploadImageResponse{
		Cid: cid,
	}
	response.SendJson(w, res, http.StatusCreated)
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
