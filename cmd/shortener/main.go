package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
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
	addr := ":8080"
	baseURL := "http://localhost:8080"

	repo := repository.NewLinkRepository()
	linkService := service.NewLinkService(repo)
	linkHandler := handler.NewLinkHandler(linkService, baseURL)

	router := NewRouter(linkHandler)

	log.Fatal(http.ListenAndServe(addr, router))
}
