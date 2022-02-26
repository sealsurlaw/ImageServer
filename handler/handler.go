package handler

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/sealsurlaw/ImageServer/config"
	"github.com/sealsurlaw/ImageServer/linkstore"
)

type Handler struct {
	BaseUrl           string
	BasePath          string
	LinkStore         linkstore.LinkStore
	WhitelistedTokens []string
}

func NewHandler(cfg *config.Config) *Handler {
	return &Handler{
		BaseUrl:           getBaseUrl(cfg),
		BasePath:          getBasePath(cfg),
		LinkStore:         getLinkStore(cfg),
		WhitelistedTokens: cfg.WhitelistedTokens,
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

func (h *Handler) hasWhitelistedToken(r *http.Request) bool {
	// If nothing configured, allow all
	if len(h.WhitelistedTokens) == 0 {
		return true
	}

	authentication := r.Header.Get("Authorization")
	if authentication == "" {
		return false
	}

	authSplit := strings.Split(authentication, " ")
	if len(authSplit) != 2 {
		return false
	}
	if authSplit[0] != "Bearer" {
		return false
	}

	auth := false
	for _, token := range h.WhitelistedTokens {
		if token == authSplit[1] {
			auth = true
			break
		}
	}

	return auth
}
