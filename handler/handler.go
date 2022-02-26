package handler

import (
	"fmt"
	"log"
	"os"

	"github.com/sealsurlaw/ImageServer/config"
	"github.com/sealsurlaw/ImageServer/linkstore"
)

type Handler struct {
	BaseUrl   string
	BasePath  string
	LinkStore linkstore.LinkStore
}

func NewHandler(cfg *config.Config) *Handler {
	baseUrl := cfg.BaseUrl
	if baseUrl == "" {
		baseUrl = fmt.Sprintf("http://localhost:%s", cfg.Port)
	}

	basePath := cfg.BasePath
	if basePath == "" {
		bp, err := os.MkdirTemp("/tmp", "imageserver.*")
		if err != nil {
			log.Fatal("Couldn't create tmp directory")
		}
		basePath = bp
	}

	var linkStore linkstore.LinkStore
	linkStore = linkstore.NewMemoryLinkStore()
	if cfg.PostgresqlConfig.Enabled {
		ls, err := linkstore.NewPostgresqlLinkStore(cfg)
		if err != nil {
			fmt.Println("PostgreSQL connection failed. Falling back to memory link store.")
		} else {
			linkStore = ls
			fmt.Println("Connected to PostgreSQL link store.")
		}
	} else {
		fmt.Println("Connected to Memory link store.")
	}

	return &Handler{
		BaseUrl:   baseUrl,
		BasePath:  basePath,
		LinkStore: linkStore,
	}
}
