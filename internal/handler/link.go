package handler

import (
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rucodencode/shortener/internal/service"
)

type LinkHandler struct {
	service *service.LinkService
	baseURL string
}

func NewLinkHandler(service *service.LinkService, baseURL string) *LinkHandler {
	return &LinkHandler{
		service: service,
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
	link, err := h.service.Create(originalURL)

	if err != nil {
		if errors.Is(err, service.ErrInvalidURL) ||
			errors.Is(err, service.ErrCodeCollision) {
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

	link, err := h.service.GetByCode(code)

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
