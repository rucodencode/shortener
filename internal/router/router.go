package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rucodencode/shortener/internal/handler/api"
	"github.com/rucodencode/shortener/internal/handler/web"
	"github.com/rucodencode/shortener/internal/logger"
)

func NewRouter(webHandler *web.LinkHandler, apiHandler *api.LinkHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(logger.RequestLogger)
	RegisterWebRouters(r, webHandler)
	RegisterAPIRouters(r, apiHandler)

	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	})

	return r
}
