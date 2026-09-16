package config

import "flag"

var Options struct {
	RunAddr string
	BaseURL string
}

func InitOptions() {
	flag.StringVar(&Options.RunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&Options.BaseURL, "b", "http://localhost:8080", "base url")

	flag.Parse()
}
