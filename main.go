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
	handle("/links/", h.Links)
	handle("/links", h.Links)
	handle("/images/", h.UploadDownload)
	handle("/images", h.UploadDownload)
	handle("/thumbnails/batch", h.ThumbnailsBatch)
	handle("/thumbnails", h.Thumbnails)

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
	return middle.LogRoutes(next)
}
