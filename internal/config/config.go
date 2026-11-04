package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddress   string
	BaseURL         string
	LogLevel        string
	FileStoragePath string
}

func NewConfig() Config {
	cfg := Config{}
	flag.StringVar(&cfg.ServerAddress, "a", ":8080", "address and port to run server")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "base url")
	flag.StringVar(&cfg.LogLevel, "l", "info", "log level")
	flag.StringVar(&cfg.FileStoragePath, "f", "storage.json", "file storage path")
	flag.Parse()

	if envServerAddress := os.Getenv("SERVER_ADDRESS"); envServerAddress != "" {
		cfg.ServerAddress = envServerAddress
	}

	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		cfg.BaseURL = envBaseURL
	}

	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		cfg.LogLevel = envLogLevel
	}

	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
		cfg.FileStoragePath = envFileStoragePath
	}

	return cfg
}
