package handler

import (
	"net/http"

	"github.com/sealsurlaw/ImageServer/errs"
	"github.com/sealsurlaw/ImageServer/helper"
)

func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		w.WriteHeader(http.StatusOK)
		return
	} else {
		msg := "Method not found."
		res := errs.NewErrorResponse(http.StatusBadRequest, msg)
		helper.SendJson(w, res)
		return
	}
}
