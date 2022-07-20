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
	handle("/files/upload", h.UploadFile)
	handle("/images/links/thumbnails/batch", h.CreateBatchImageThumbnailLinks)
	handle("/images/links/thumbnails", h.CreateImageThumbnailLink)
	handle("/images/links/upload", h.CreateImageUploadLink)
	handle("/images/links/", h.GetImageFromTokenLink)
	handle("/images/links", h.CreateImageLink)
	handle("/images/uploads/", h.UploadImageWithLink)
	handle("/images/uploads", h.UploadImage)
	handle("/images/", h.DownloadImage)
	handle("/ipfs/", h.GetIpfsFile)
	handle("/meta", h.DownloadImage)

	handle("/", func(w http.ResponseWriter, r *http.Request) {
		response.SendMethodNotFound(w)
	})

	fmt.Printf("Starting server at port %s\n", cfg.Port)
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
