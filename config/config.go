package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

type Config struct {
	Port             string `json:"port"`
	BaseUrl          string `json:"baseUrl"`
	BasePath         string `json:"basePath"`
	ThumbnailQuality int    `json:"thumbnailQuality"`
	CleanupDuration  string `json:"cleanupDuration"`
	PostgresqlConfig struct {
		Enabled        bool   `json:"enabled"`
		DatabaseString string `json:"databaseString"`
	} `json:"postgresqlConfig"`
	WhitelistedTokens []string `json:"whitelistedTokens"`
}

func NewConfig() *Config {
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
	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	if cfg.BaseUrl == "" {
		cfg.BaseUrl = fmt.Sprintf("http://localhost:%s", cfg.Port)
	}

	if cfg.BasePath == "" || basePathDoesNotExists(cfg.BasePath) {
		bp, err := os.MkdirTemp(os.TempDir(), "imageserver.*")
		if err != nil {
			log.Fatal("Couldn't create tmp directory")
		}
		cfg.BasePath = bp
	}

	fmt.Printf("Writing images to %s\n", cfg.BasePath)

	if cfg.ThumbnailQuality == 0 {
		cfg.ThumbnailQuality = 50
	}

	duration, err := time.ParseDuration(cfg.CleanupDuration)
	if err != nil {
		duration = time.Hour * 24
	}
	cfg.CleanupDuration = duration.String()
	fmt.Println(cfg.CleanupDuration)

	if cfg.WhitelistedTokens == nil {
		cfg.WhitelistedTokens = []string{}
	}
}

func basePathDoesNotExists(basePath string) bool {
	_, err := os.ReadDir(basePath)
	if err != nil {
		return true
	}

	return false
}