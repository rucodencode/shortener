package health_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/rucodencode/shortener/internal/handler/health"
	"github.com/rucodencode/shortener/internal/mock"
	"github.com/stretchr/testify/assert"
)

func TestHealthHandler_PingDB_Healthy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mock.NewMockHealthService(ctrl)
	mockService.EXPECT().PingDB(gomock.Any()).Return(nil).Times(1)

	handler := health.NewHealthHandler(mockService)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	recorder := httptest.NewRecorder()

	handler.PingDB(recorder, req)

	assert.Equal(t, recorder.Code, http.StatusOK)
}

func TestHealthHandler_PingDB_Unhealthy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mock.NewMockHealthService(ctrl)
	mockService.EXPECT().PingDB(gomock.Any()).Return(errors.New("database connection failed")).Times(1)

	handler := health.NewHealthHandler(mockService)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	recorder := httptest.NewRecorder()

	handler.PingDB(recorder, req)

	assert.Equal(t, recorder.Code, http.StatusInternalServerError)
}
