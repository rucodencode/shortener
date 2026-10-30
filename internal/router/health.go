package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type HealthHandler interface {
	PingDB(w http.ResponseWriter, r *http.Request)
}

func RegisterHealthRouters(r chi.Router, h HealthHandler) {
	r.Get("/ping", h.PingDB)
}
