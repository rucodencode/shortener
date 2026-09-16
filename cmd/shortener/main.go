package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rucodencode/shortener/internal/config"
	"github.com/rucodencode/shortener/internal/handler"
	"github.com/rucodencode/shortener/internal/repository"
	"github.com/rucodencode/shortener/internal/service"
)

func NewRouter(linkHandler *handler.LinkHandler) http.Handler {
	router := chi.NewRouter()

	router.Post("/", linkHandler.Create)

	router.Get("/{code}", linkHandler.Get)

	router.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	})

	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	})

	return router
}

func main() {
	config.InitOptions()

	repo := repository.NewLinkRepository()
	linkService := service.NewLinkService(repo)
	linkHandler := handler.NewLinkHandler(linkService, config.Options.BaseURL)

	router := NewRouter(linkHandler)

	fmt.Println("Running server on", config.Options.RunAddr)
	log.Fatal(http.ListenAndServe(config.Options.RunAddr, router))
}
