package config

import (
	"flag"
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

	return options
}
