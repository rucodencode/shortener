package config

import (
	"flag"
	"os"
)

type config struct {
	ServerAddress   string
	BaseURL         string
	LogLevel        string
	FileStoragePath string
}

func NewConfig() config {
	config := config{}
	flag.StringVar(&config.ServerAddress, "a", ":8080", "address and port to run server")
	flag.StringVar(&config.BaseURL, "b", "http://localhost:8080", "base url")
	flag.StringVar(&config.LogLevel, "l", "info", "log level")
	flag.StringVar(&config.FileStoragePath, "f", "storage.json", "file storage path")
	flag.Parse()

	if envServerAddress := os.Getenv("SERVER_ADDRESS"); envServerAddress != "" {
		config.ServerAddress = envServerAddress
	}

	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		config.BaseURL = envBaseURL
	}

	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		config.LogLevel = envLogLevel
	}

	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
		config.FileStoragePath = envFileStoragePath
	}

	return config
}
