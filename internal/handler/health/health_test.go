package health_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/rucodencode/shortener/internal/handler/health"
	"github.com/rucodencode/shortener/internal/mock"
	"github.com/rucodencode/shortener/internal/service"
	"github.com/stretchr/testify/assert"
)

func TestHealthHandler_PingDB_Healthy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockHealthRepository(ctrl)
	mockRepo.EXPECT().Ping(gomock.Any()).Return(nil).Times(1)

	svc := service.NewHealthService(mockRepo)

	handler := health.NewHealthHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	recorder := httptest.NewRecorder()

	handler.PingDB(recorder, req)

	assert.Equal(t, recorder.Code, http.StatusOK)
}

func TestHealthHandler_PingDB_Unhealthy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockHealthRepository(ctrl)
	mockRepo.EXPECT().Ping(gomock.Any()).Return(errors.New("database connection failed")).Times(1)

	svc := service.NewHealthService(mockRepo)

	handler := health.NewHealthHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	recorder := httptest.NewRecorder()

	handler.PingDB(recorder, req)

	assert.Equal(t, recorder.Code, http.StatusInternalServerError)
}
