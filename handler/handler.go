package handler

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/sealsurlaw/ImageServer/config"
	"github.com/sealsurlaw/ImageServer/linkstore"
)

type Handler struct {
	BaseUrl               string
	BasePath              string
	LinkStore             linkstore.LinkStore
	AuthorizedIpAddresses []net.IP
}

func NewHandler(cfg *config.Config) *Handler {
	return &Handler{
		BaseUrl:               getBaseUrl(cfg),
		BasePath:              getBasePath(cfg),
		LinkStore:             getLinkStore(cfg),
		AuthorizedIpAddresses: getAuthorizedIpAddresses(cfg),
	}
}

func getBaseUrl(cfg *config.Config) string {
	baseUrl := cfg.BaseUrl
	if baseUrl == "" {
		baseUrl = fmt.Sprintf("http://localhost:%s", cfg.Port)
	}

	return baseUrl
}

func getBasePath(cfg *config.Config) string {
	basePath := cfg.BasePath
	if basePath == "" {
		bp, err := os.MkdirTemp("/tmp", "imageserver.*")
		if err != nil {
			log.Fatal("Couldn't create tmp directory")
		}
		basePath = bp
	}

	return basePath
}

func getLinkStore(cfg *config.Config) linkstore.LinkStore {
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

	return linkStore
}

func getAuthorizedIpAddresses(cfg *config.Config) []net.IP {
	authIps := []net.IP{}
	if !cfg.DirectDownload.Enabled {
		return authIps
	}

	for _, ipAddress := range cfg.DirectDownload.AuthorizedIpAddresses {
		authIps = append(authIps, net.ParseIP(ipAddress))
	}

	return authIps
}
