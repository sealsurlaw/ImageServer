package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/sealsurlaw/gouvre/config"
	"github.com/sealsurlaw/gouvre/handler"
	"github.com/sealsurlaw/gouvre/middle"
	"github.com/sealsurlaw/gouvre/response"
)

func main() {
	cfg := config.NewConfig()
	h := handler.NewHandler(cfg)

	handle("/ping", h.Ping)
	handle("/links/thumbnails/batch", h.CreateBatchThumbnailLinks)
	handle("/links/thumbnails", h.CreateThumbnailLink)
	handle("/links/upload", h.CreateUploadLink)
	handle("/uploads/", h.UploadFileWithLink)
	handle("/uploads", h.UploadFile)
	handle("/images/", h.DownloadFile)
	handle("/links/", h.GetImageFromTokenLink)
	handle("/links", h.CreateLink)

	handle("/", func(w http.ResponseWriter, r *http.Request) {
		response.SendMethodNotFound(w)
		return
	})

	fmt.Printf(fmt.Sprintf("Starting server at port %s\n", cfg.Port))
	if err := http.ListenAndServe(fmt.Sprintf(":%s", cfg.Port), nil); err != nil {
		log.Fatal(err)
	}
}

func handle(pattern string, handler http.HandlerFunc) {
	http.Handle(pattern, middleware(handler))
}

func middleware(next http.HandlerFunc) http.Handler {
	m := middle.LogRoutes(next)
	return m
}
