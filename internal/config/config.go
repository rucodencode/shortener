package config

import (
	"flag"
	"os"
)

type config struct {
	ServerAddress string
	BaseURL       string
	LogLevel      string
}

func NewConfig() config {
	config := config{}
	flag.StringVar(&config.ServerAddress, "a", ":8080", "address and port to run server")
	flag.StringVar(&config.BaseURL, "b", "http://localhost:8080", "base url")
	flag.StringVar(&config.LogLevel, "l", "info", "log level")
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

	return config
}
