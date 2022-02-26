package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/sealsurlaw/ImageServer/config"
	"github.com/sealsurlaw/ImageServer/handler"
	"github.com/sealsurlaw/ImageServer/helper"
)

func main() {
	rand.Seed(time.Now().Unix())

	cfg := config.NewConfig()
	h := handler.NewHandler(cfg)

	// error handled in NewConfig
	duration, _ := time.ParseDuration(cfg.CleanupDuration)
	go helper.CleanExpiredTokens(h.LinkStore, duration)

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
