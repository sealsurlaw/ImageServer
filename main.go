package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/sealsurlaw/ImageServer/config"
	"github.com/sealsurlaw/ImageServer/handler"
	"github.com/sealsurlaw/ImageServer/helper"
)

func main() {
	cfg := config.NewConfig()
	h := handler.NewHandler(cfg)

	go helper.CleanExpiredTokens(h.LinkStore, cfg.CleanupDuration)

	http.HandleFunc("/ping", h.Ping)
	http.HandleFunc("/links/", h.Links)
	http.HandleFunc("/links", h.Links)
	http.HandleFunc("/images/", h.UploadDownload)
	http.HandleFunc("/images", h.UploadDownload)
	http.HandleFunc("/thumbnails/batch", h.ThumbnailsBatch)
	http.HandleFunc("/thumbnails", h.Thumbnails)

	fmt.Printf(fmt.Sprintf("Starting server at port %s\n", cfg.Port))
	if err := http.ListenAndServe(fmt.Sprintf(":%s", cfg.Port), nil); err != nil {
		log.Fatal(err)
	}
}
