package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/rucodencode/shortener/internal/logger"
	"github.com/rucodencode/shortener/internal/model"
	"github.com/rucodencode/shortener/internal/service"
	"go.uber.org/zap"
)

type CreateLinkRequest struct {
	URL string `json:"url"`
}

type CreateLinkResponse struct {
	Result string `json:"result"`
}

type LinkService interface {
	Create(ctx context.Context, originalURL string) (*model.Link, error)
}

type LinkHandler struct {
	service LinkService
	baseURL string
}

func NewLinkHandler(src LinkService, baseURL string) *LinkHandler {
	return &LinkHandler{
		service: src,
		baseURL: baseURL,
	}
}

func (h *LinkHandler) Create(w http.ResponseWriter, r *http.Request) {
	logger.Log.Debug("decoding request")
	var req CreateLinkRequest

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&req); err != nil {
		logger.Log.Debug("cannot decode request JSON body", zap.Error(err))
		http.Error(
			w, "bad request",
			http.StatusBadRequest,
		)
		return
	}

	link, err := h.service.Create(r.Context(), req.URL)

	if err != nil {
		if errors.Is(err, service.ErrInvalidURL) {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
		logger.Log.Debug("bad URL", zap.String("URL", req.URL))
		return
	}

	resp := CreateLinkResponse{
		Result: h.baseURL + "/" + link.Code,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	enc := json.NewEncoder(w)
	if err := enc.Encode(resp); err != nil {
		logger.Log.Debug("error encoding response", zap.Error(err))
		return
	}

	logger.Log.Debug("sending HTTP response", zap.Int("status", http.StatusCreated))
}
