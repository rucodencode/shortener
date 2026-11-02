package web

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rucodencode/shortener/internal/model"
	"github.com/rucodencode/shortener/internal/service"
)

type LinkService interface {
	Create(ctx context.Context, originalURL string) (*model.Link, error)
	GetByCode(ctx context.Context, code string) (*model.Link, error)
}

type LinkHandler struct {
	service LinkService
	baseURL string
}

func NewLinkHandler(svc LinkService, baseURL string) *LinkHandler {
	return &LinkHandler{
		service: svc,
		baseURL: baseURL,
	}
}

func (h *LinkHandler) Create(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	originalURL := string(body)
	link, err := h.service.Create(r.Context(), originalURL)

	if err != nil {
		if errors.Is(err, service.ErrInvalidURL) {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(h.baseURL + "/" + link.Code))
}

func (h *LinkHandler) Get(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	link, err := h.service.GetByCode(r.Context(), code)

	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, link.OriginalURL, http.StatusTemporaryRedirect)
}
