package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGzipMiddleware_Compress(t *testing.T) {
	original := `{"url":"https://example.com"}`
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(original))
	})

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	rec := httptest.NewRecorder()

	GzipMiddleware(handler).ServeHTTP(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "gzip", resp.Header.Get("Content-Encoding"))

	zr, err := gzip.NewReader(resp.Body)
	require.NoError(t, err)
	defer zr.Close()

	body, err := io.ReadAll(zr)
	require.NoError(t, err)

	assert.Equal(t, original, string(body))
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
}

func TestGzipMiddleware_Decompress(t *testing.T) {
	original := `{"url":"https://example.com"}`

	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, err := zw.Write([]byte(original))
	require.NoError(t, err)
	require.NoError(t, zw.Close())

	var received []byte
	var readErr error

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received, readErr = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(buf.Bytes()))
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	GzipMiddleware(handler).ServeHTTP(rec, req)

	require.NoError(t, readErr)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, original, string(received))
}

func TestGzipMiddleware_InvalidGzip(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	})

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("invalid gzip data"))
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	GzipMiddleware(handler).ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Empty(t, rec.Header().Get("Content-Encoding"))
}
