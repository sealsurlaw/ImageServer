package config

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"
)

type Config struct {
	Port             string `json:"port"`
	BaseUrl          string `json:"baseUrl"`
	BasePath         string `json:"basePath"`
	ThumbnailQuality int    `json:"thumbnailQuality"`
	CleanupDuration  string `json:"cleanupDuration"`
	HashFilename     bool   `json:"hashFilename"`
	PostgresqlConfig struct {
		Enabled        bool   `json:"enabled"`
		DatabaseString string `json:"databaseString"`
	} `json:"postgresqlConfig"`
	WhitelistedTokens []string `json:"whitelistedTokens"`
}

func NewConfig() *Config {
	rand.Seed(time.Now().Unix())

	cfg := &Config{}
	configFile := "config.json"
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}

	file, err := os.ReadFile(configFile)
	if err != nil {
		fmt.Println(err)
	} else {
		_ = json.Unmarshal(file, cfg)
	}

	populateConfigWithDefaults(cfg)

	return cfg
}

func populateConfigWithDefaults(cfg *Config) {
	cfg.Port = configurePort(cfg.Port)
	cfg.BaseUrl = configureBaseUrl(cfg.BaseUrl, cfg.Port)
	cfg.BasePath = configureBasePath(cfg.BasePath)
	cfg.CleanupDuration = configureCleanupDuration(cfg.CleanupDuration)
	cfg.ThumbnailQuality = configureThumbnailQuality(cfg.ThumbnailQuality)
	cfg.WhitelistedTokens = configureWhitelistedTokens(cfg.WhitelistedTokens)

	fmt.Printf("Writing images to %s\n", cfg.BasePath)
}

func configurePort(port string) string {
	if port == "" {
		port = "8080"
	}
	return port
}

func configureBaseUrl(baseUrl, port string) string {
	if baseUrl == "" {
		baseUrl = fmt.Sprintf("http://localhost:%s", port)
	}
	return baseUrl
}

func configureBasePath(basePath string) string {
	if basePath == "" || basePathDoesNotExists(basePath) {
		bp, err := os.MkdirTemp(os.TempDir(), "imageserver.*")
		if err != nil {
			log.Fatal("Couldn't create tmp directory")
		}
		basePath = bp
	}
	return basePath
}

func configureCleanupDuration(cleanupDuration string) string {
	duration, err := time.ParseDuration(cleanupDuration)
	if err != nil {
		duration = time.Hour * 24
	}
	return duration.String()
}

func configureThumbnailQuality(thumbnailQuality int) int {
	if thumbnailQuality == 0 {
		thumbnailQuality = 50
	}
	return thumbnailQuality
}

func configureWhitelistedTokens(whitelistedTokens []string) []string {
	if whitelistedTokens == nil {
		whitelistedTokens = []string{}
	}
	return whitelistedTokens
}

func basePathDoesNotExists(basePath string) bool {
	_, err := os.ReadDir(basePath)
	if err != nil {
		return true
	}

	return false
}
