package config

import (
	"flag"
	"os"
)

type options struct {
	ServerAddress string
	BaseURL       string
}

func NewOptions() options {
	options := options{}
	flag.StringVar(&options.ServerAddress, "a", ":8080", "address and port to run server")
	flag.StringVar(&options.BaseURL, "b", "http://localhost:8080", "base url")
	flag.Parse()

	if envServerAddress := os.Getenv("SERVER_ADDRESS"); envServerAddress != "" {
		options.ServerAddress = envServerAddress
	}

	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		options.BaseURL = envBaseURL
	}

	return options
}
