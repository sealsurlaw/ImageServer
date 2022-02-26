package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/sealsurlaw/ImageServer/config"
	"github.com/sealsurlaw/ImageServer/handler"
)

func main() {
	cfg := config.NewConfig()
	h := handler.NewHandler(cfg)

	http.HandleFunc("/ping", h.Ping)
	http.HandleFunc("/link/", h.Link)
	http.HandleFunc("/upload", h.Upload)
	http.HandleFunc("/download/", h.Download)
	http.HandleFunc("/thumbnail/", h.Thumbnail)

	fmt.Printf(fmt.Sprintf("Starting server at port %s\n", cfg.Port))
	if err := http.ListenAndServe(fmt.Sprintf(":%s", cfg.Port), nil); err != nil {
		log.Fatal(err)
	}
}
