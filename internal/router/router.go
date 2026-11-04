package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rucodencode/shortener/internal/handler"
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
