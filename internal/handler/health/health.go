package health

import (
	"context"
	"net/http"
	"time"
)

type HealthService interface {
	PingDB(ctx context.Context) error
}

type HealthHandler struct {
	service HealthService
}

func NewHealthHandler(svc HealthService) *HealthHandler {
	return &HealthHandler{
		service: svc,
	}
}

func (h *HealthHandler) PingDB(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second)
	defer cancel()

	if err := h.service.PingDB(ctx); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
