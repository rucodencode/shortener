package middleware

import (
	"compress/gzip"
	"io"
	"mime"
	"net/http"
	"slices"
	"strings"
)

type compressWriter struct {
	w           http.ResponseWriter
	zw          *gzip.Writer
	wroteHeader bool
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w: w,
	}
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {
	if !c.wroteHeader {
		if c.Header().Get("Content-Type") == "" {
			c.Header().Set("Content-Type", http.DetectContentType(p))
		}

		c.WriteHeader(http.StatusOK)
	}

	if c.zw != nil {
		return c.zw.Write(p)
	}

	return c.w.Write(p)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	if c.wroteHeader {
		return
	}

	c.wroteHeader = true

	needContentTypes := []string{"application/json", "text/html"}
	noNeedStatusCodes := []int{http.StatusNoContent, http.StatusNotModified}

	contentType, _, err := mime.ParseMediaType(
		c.Header().Get("Content-Type"),
	)

	shouldCompress := err == nil &&
		slices.Contains(needContentTypes, contentType) &&
		c.Header().Get("Content-Encoding") == "" &&
		!slices.Contains(noNeedStatusCodes, statusCode)

	if shouldCompress {
		c.Header().Set("Content-Encoding", "gzip")
		c.Header().Del("Content-Length")
		c.zw = gzip.NewWriter(c.w)
	}

	c.w.WriteHeader(statusCode)
}

func (c *compressWriter) Close() error {
	if c.zw != nil {
		return c.zw.Close()
	}

	return nil
}

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}

	return c.zr.Close()
}

func GzipMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")

		if sendsGzip {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				http.Error(
					w,
					"invalid gzip body",
					http.StatusBadRequest,
				)
				return
			}

			r.Body = cr
			r.ContentLength = -1

			defer cr.Close()
		}

		ow := w
		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")

		if supportsGzip {
			cw := newCompressWriter(w)
			ow = cw
			defer cw.Close()
		}

		h.ServeHTTP(ow, r)
	})
}
