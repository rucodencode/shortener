package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/rucodencode/shortener/internal/handler/api"
)

func RegisterAPIRouters(r chi.Router, linkHandler *api.LinkHandler) {
	r.Route("/api", func(r chi.Router) {
		r.Post("/shorten", linkHandler.Create)
	})
}
