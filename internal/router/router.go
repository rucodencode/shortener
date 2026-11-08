package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rucodencode/shortener/internal/auth"
	"github.com/rucodencode/shortener/internal/config"
	"github.com/rucodencode/shortener/internal/logger"
	"github.com/rucodencode/shortener/internal/middleware"
)

func NewRouter(cfg config.Config) *chi.Mux {
	jwtManager := auth.NewJWTManager(auth.ConfigJWT{
		SingingKey: cfg.TokenSigningKey,
	})
	manager := middleware.NewAuthMiddleware(jwtManager)

	r := chi.NewRouter()
	r.Use(logger.RequestLogger)
	r.Use(middleware.GzipMiddleware)
	r.Use(manager.Handler)

	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	})

	return r
}
