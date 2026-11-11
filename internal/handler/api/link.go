package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/rucodencode/shortener/internal/logger"
	"github.com/rucodencode/shortener/internal/middleware"
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

type BatchCreateLinkRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type BatchCreateLinkResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type GetListByUserIDResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type LinkService interface {
	Create(ctx context.Context, userID string, originalURL string) (*model.Link, error)
	CreateBatch(ctx context.Context, userID string, originalURLs []string) ([]*model.Link, error)
	GetListByUserID(ctx context.Context, userID string) ([]*model.Link, error)
	DeleteLinksByCodes(ctx context.Context, userID string, codes []string) error
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

	userID, _ := middleware.UserIDFromContext(r.Context())

	link, err := h.service.Create(r.Context(), userID, req.URL)

	sendResponse := func(link *model.Link, statusCode int) {
		resp := CreateLinkResponse{
			Result: h.baseURL + "/" + link.Code,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)

		enc := json.NewEncoder(w)
		if err := enc.Encode(resp); err != nil {
			logger.Log.Debug("error encoding response", zap.Error(err))
			return
		}

		logger.Log.Debug("sending HTTP response", zap.Int("status", statusCode))
	}

	if err != nil {
		if errors.Is(err, service.ErrInvalidURL) {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		if errors.Is(err, service.ErrAlreadyExists) {
			sendResponse(link, http.StatusConflict)
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
		logger.Log.Error("failed to create link", zap.Error(err))
		return
	}

	sendResponse(link, http.StatusCreated)
}

func (h *LinkHandler) CreateBatch(w http.ResponseWriter, r *http.Request) {
	logger.Log.Debug("batch request started")
	var items []BatchCreateLinkRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&items); err != nil {
		logger.Log.Debug("cannot decode batch request", zap.Error(err))
		http.Error(w, "bad request: invalid JSON", http.StatusBadRequest)
		return
	}

	originalURLs := make([]string, 0, len(items))
	for _, v := range items {
		originalURLs = append(originalURLs, v.OriginalURL)
	}

	userID, _ := middleware.UserIDFromContext(r.Context())
	links, err := h.service.CreateBatch(r.Context(), userID, originalURLs)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidURL):
			http.Error(w, "bad request", http.StatusBadRequest)
		case errors.Is(err, service.ErrAlreadyExists):
			http.Error(w, "duplicate original URL", http.StatusConflict)
		default:
			logger.Log.Error("failed to create batch links", zap.Error(err))
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	results := make([]BatchCreateLinkResponse, 0, len(items))
	for key, v := range links {
		results = append(results, BatchCreateLinkResponse{
			CorrelationID: items[key].CorrelationID,
			ShortURL:      h.baseURL + "/" + v.Code,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	enc := json.NewEncoder(w)
	if err := enc.Encode(results); err != nil {
		logger.Log.Debug("error encoding response", zap.Error(err))
		return
	}

	logger.Log.Debug("sending HTTP response", zap.Int("status", http.StatusCreated))
}

func (h *LinkHandler) GetListByUser(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	links, err := h.service.GetListByUserID(r.Context(), userID)
	if err != nil {
		logger.Log.Error("failed to create batch links", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if len(links) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	results := make([]GetListByUserIDResponse, 0, len(links))
	for _, link := range links {
		results = append(results, GetListByUserIDResponse{
			ShortURL:    h.baseURL + "/" + link.Code,
			OriginalURL: link.OriginalURL,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	enc := json.NewEncoder(w)
	if err := enc.Encode(results); err != nil {
		logger.Log.Debug("error encoding response", zap.Error(err))
		return
	}

	logger.Log.Debug("sending HTTP response", zap.Int("status", http.StatusCreated))
}

func (h *LinkHandler) DeleteLinks(w http.ResponseWriter, r *http.Request) {
	logger.Log.Debug("delete links started")
	userID, _ := middleware.UserIDFromContext(r.Context())

	var codes []string
	dec := json.NewDecoder(r.Body)

	if err := dec.Decode(&codes); err != nil {
		logger.Log.Debug("cannot decode batch request", zap.Error(err))
		http.Error(w, "bad request: invalid JSON", http.StatusBadRequest)
		return
	}

	err := h.service.DeleteLinksByCodes(r.Context(), userID, codes)
	if err != nil {
		logger.Log.Error("failed to delete links", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
