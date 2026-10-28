package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rucodencode/shortener/internal/repository"
	"github.com/rucodencode/shortener/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const baseURL = "http://localhost:8080"

func newTestHandler(t *testing.T) *LinkHandler {
	storagePath := filepath.Join(t.TempDir(), "storage.json")

	repo, err := repository.NewLinkRepository(storagePath)
	if err != nil {
		t.Fatal(err)
	}

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
				contentType: "application/json",
				body:        `{"url":"http://example.com/"}`,
			},
			want: want{
				code:        http.StatusCreated,
				response:    `{"result":"http://localhost:8080/a6bf1757"}`,
				contentType: "application/json",
			},
		},
		{
			name: "valid https URL",
			request: requestData{
				contentType: "application/json",
				body:        `{"url":"https://practicum.yandex.ru/"}`,
			},
			want: want{
				code:        http.StatusCreated,
				response:    `{"result":"http://localhost:8080/0dd19817"}`,
				contentType: "application/json",
			},
		},
		{
			name: "invalid URL without scheme",
			request: requestData{
				contentType: "application/json",
				body:        `{"url":"practicum.yandex.ru"}`,
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "empty URL",
			request: requestData{
				contentType: "application/json",
				body:        `{"url":""}`,
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "missing URL field",
			request: requestData{
				contentType: "application/json",
				body:        `{}`,
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "empty body",
			request: requestData{
				contentType: "application/json",
				body:        "",
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "invalid JSON",
			request: requestData{
				contentType: "application/json",
				body:        `{"url":"https://example.com/"`,
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "wrong URL type",
			request: requestData{
				contentType: "application/json",
				body:        `{"url":123}`,
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			linkHandler := newTestHandler(t)

			request, err := http.NewRequest(
				http.MethodPost,
				"api/shorten",
				strings.NewReader(test.request.body),
			)

			require.NoError(t, err)
			request.Header.Set("Content-Type", test.request.contentType)

			w := httptest.NewRecorder()
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
				assert.JSONEq(t, test.want.response, string(resBody))
			}
		})
	}
}

func TestLinkHandler_CreateSameURL(t *testing.T) {
	linkHandler := newTestHandler(t)

	firstRequest := httptest.NewRequest(
		http.MethodPost,
		"/",
		strings.NewReader(`{"url":"http://example.com/"}`),
	)

	firstRequest.Header.Set("Content-Type", "application/json")

	firstRecorder := httptest.NewRecorder()
	linkHandler.Create(firstRecorder, firstRequest)

	firstResponse := firstRecorder.Result()
	defer firstResponse.Body.Close()

	firstBody, err := io.ReadAll(firstResponse.Body)
	require.NoError(t, err)

	secondRequest := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"url":"http://example.com/"}`))

	secondRequest.Header.Set("Content-Type", "text/plain")

	secondRecorder := httptest.NewRecorder()
	linkHandler.Create(secondRecorder, secondRequest)

	secondResponse := secondRecorder.Result()
	defer secondResponse.Body.Close()

	secondBody, err := io.ReadAll(secondResponse.Body)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, firstResponse.StatusCode)
	assert.Equal(t, http.StatusCreated, secondResponse.StatusCode)
	require.JSONEq(t, string(firstBody), string(secondBody))
}
