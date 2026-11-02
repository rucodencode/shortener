package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type WebHandler interface {
	Create(w http.ResponseWriter, r *http.Request)
	Get(w http.ResponseWriter, r *http.Request)
}

func RegisterWebRouters(r chi.Router, h WebHandler) {
	r.Post("/", h.Create)
	r.Get("/{code}", h.Get)
}
