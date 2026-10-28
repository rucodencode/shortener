package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/rucodencode/shortener/internal/config"
	"github.com/rucodencode/shortener/internal/handler/api"
	"github.com/rucodencode/shortener/internal/handler/web"
	"github.com/rucodencode/shortener/internal/logger"
	"github.com/rucodencode/shortener/internal/repository"
	"github.com/rucodencode/shortener/internal/router"
	"github.com/rucodencode/shortener/internal/service"
)

func main() {
	cfg := config.NewConfig()

	if err := logger.Initialize(cfg.LogLevel); err != nil {
		log.Fatal(err)
	}

	repo, err := repository.NewLinkRepository(cfg.FileStoragePath)
	if err != nil {
		log.Fatal("Rep didn't loaded")
	}

	linkService := service.NewLinkService(repo)
	webLinkHandler := web.NewLinkHandler(linkService, cfg.BaseURL)
	apiLinkHandler := api.NewLinkHandler(linkService, cfg.BaseURL)

	router := router.NewRouter(webLinkHandler, apiLinkHandler)

	fmt.Println("Running server on", cfg.ServerAddress)
	log.Fatal(http.ListenAndServe(cfg.ServerAddress, router))
}
