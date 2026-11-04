package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/rucodencode/shortener/internal/handler/web"
)

func RegisterWebRouters(r chi.Router, linkHandler *web.LinkHandler) {
	r.Post("/", linkHandler.Create)
	r.Get("/{code}", linkHandler.Get)
}
