package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rucodencode/shortener/internal/repository"
	"github.com/rucodencode/shortener/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const baseURL = "http://localhost:8080"

func newTestHandler() *LinkHandler {

	repo := repository.NewLinkRepository()
	linkService := service.NewLinkService(repo)
	return NewLinkHandler(linkService, baseURL)
}

func TestLinkHandler_Create(t *testing.T) {
	type requestData struct {
		contentType string
		body        string
	}
	type want struct {
		code        int
		response    string
		contentType string
	}

	tests := []struct {
		name    string
		request requestData
		want    want
	}{
		{
			name: "valid http URL",
			request: requestData{
				contentType: "text/plain",
				body:        "http://example.com/",
			},
			want: want{
				code:        http.StatusCreated,
				response:    "a6bf1757",
				contentType: "text/plain",
			},
		},
		{
			name: "valid https URL",
			request: requestData{
				contentType: "text/plain",
				body:        "https://practicum.yandex.ru/",
			},
			want: want{
				code:        http.StatusCreated,
				response:    "0dd19817",
				contentType: "text/plain",
			},
		},
		{
			name: "invalid URL without scheme",
			request: requestData{
				contentType: "text/plain",
				body:        "practicum.yandex.ru",
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "body empty",
			request: requestData{
				contentType: "text/plain",
				body:        "",
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			linkHandler := newTestHandler()

			request, err := http.NewRequest(http.MethodPost, `/`, strings.NewReader(test.request.body))

			require.NoError(t, err)

			w := httptest.NewRecorder()

			request.Header.Set("Content-Type", test.request.contentType)
			linkHandler.Create(w, request)
			response := w.Result()

			defer response.Body.Close()
			assert.Equal(t, test.want.code, response.StatusCode)

			resBody, err := io.ReadAll(response.Body)
			require.NoError(t, err)

			if test.want.contentType != "" {
				assert.Equal(t, test.want.contentType, response.Header.Get("Content-Type"))
			}

			if test.want.response != "" {
				code := strings.TrimPrefix(string(resBody), baseURL+"/")
				assert.Equal(t, test.want.response, code)
			}

		})
	}
}

func TestLinkHandler_CreateSameURL(t *testing.T) {
	linkHandler := newTestHandler()

	firstRequest := httptest.NewRequest(
		http.MethodPost,
		"/",
		strings.NewReader("https://practicum.yandex.ru/"),
	)

	firstRequest.Header.Set("Content-Type", "text/plain")

	firstRecorder := httptest.NewRecorder()
	linkHandler.Create(firstRecorder, firstRequest)

	firstResponse := firstRecorder.Result()
	defer firstResponse.Body.Close()

	firstBody, err := io.ReadAll(firstResponse.Body)
	require.NoError(t, err)

	secondRequest := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://practicum.yandex.ru/"))

	secondRequest.Header.Set("Content-Type", "text/plain")

	secondRecorder := httptest.NewRecorder()
	linkHandler.Create(secondRecorder, secondRequest)

	secondResponse := secondRecorder.Result()
	defer secondResponse.Body.Close()

	secondBody, err := io.ReadAll(secondResponse.Body)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, firstResponse.StatusCode)
	assert.Equal(t, http.StatusCreated, secondResponse.StatusCode)
	assert.Equal(t, string(firstBody), string(secondBody))
}

func TestLinkHandler_Get(t *testing.T) {
	linkHandler := newTestHandler()
	originalURL := "https://practicum.yandex.ru/"

	firstRequest := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(originalURL))
	firstRequest.Header.Set("Content-Type", "text/plain")

	firstRecorder := httptest.NewRecorder()
	linkHandler.Create(firstRecorder, firstRequest)

	firstResponse := firstRecorder.Result()
	defer firstResponse.Body.Close()

	firstBody, err := io.ReadAll(firstResponse.Body)
	require.NoError(t, err)

	code := strings.TrimPrefix(string(firstBody), baseURL+"/")

	secondRequest := httptest.NewRequest(http.MethodGet, "/"+code, nil)
	secondRequest.Header.Set("Content-Type", "text/plain")

	secondRecorder := httptest.NewRecorder()
	linkHandler.Get(secondRecorder, secondRequest)

	secondResponse := secondRecorder.Result()
	defer secondResponse.Body.Close()

	assert.Equal(t, http.StatusTemporaryRedirect, secondResponse.StatusCode)
	assert.Equal(t, originalURL, secondResponse.Header.Get("Location"))
}

func TestLinkHandler_GetUnknownCode(t *testing.T) {
	linkHandler := newTestHandler()
	code := "unknown"
	request := httptest.NewRequest(http.MethodGet, "/"+code, nil)
	recorder := httptest.NewRecorder()
	linkHandler.Get(recorder, request)

	response := recorder.Result()
	defer response.Body.Close()

	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
}
