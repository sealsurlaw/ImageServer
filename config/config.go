package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type Config struct {
	Port             string `json:"port"`
	BaseUrl          string `json:"baseUrl"`
	BasePath         string `json:"basePath"`
	PostgresqlConfig struct {
		Enabled        bool   `json:"enabled"`
		DatabaseString string `json:"databaseString"`
	} `json:"postgresqlConfig"`
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
}

func basePathDoesNotExists(basePath string) bool {
	_, err := os.ReadDir(basePath)
	if err != nil {
		return true
	}

	return false
}
