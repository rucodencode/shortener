package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type APIHandler interface {
	Create(w http.ResponseWriter, r *http.Request)
}

func RegisterAPIRouters(r chi.Router, h APIHandler) {
	r.Route("/api", func(r chi.Router) {
		r.Post("/shorten", h.Create)
	})
}
