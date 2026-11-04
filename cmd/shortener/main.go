package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/rucodencode/shortener/internal/config"
	"github.com/rucodencode/shortener/internal/handler"
	"github.com/rucodencode/shortener/internal/repository"
	"github.com/rucodencode/shortener/internal/router"
	"github.com/rucodencode/shortener/internal/service"
)

func main() {
	cfg := config.NewOptions()

	repo := repository.NewLinkRepository()
	linkService := service.NewLinkService(repo)
	linkHandler := handler.NewLinkHandler(linkService, cfg.BaseURL)

	router := router.NewRouter(linkHandler)

	fmt.Println("Running server on", cfg.ServerAddress)
	log.Fatal(http.ListenAndServe(cfg.ServerAddress, router))
}
