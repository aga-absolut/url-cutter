package compress

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type gzipWriter struct {
	http.ResponseWriter
	zw *gzip.Writer
}

func (c gzipWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

func Compress(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			h.ServeHTTP(w, r)
		}

		zw := gzip.NewWriter(w)
		defer zw.Close()

		w.Header().Set("Content-Encoding", "gzip")
		h.ServeHTTP(gzipWriter{ResponseWriter: w, zw: zw}, r)
	}
}

func Decompress(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Comtent-Encoding"), "gzip") {
			h.ServeHTTP(w, r)
		}

		zw, err := gzip.NewReader(r.Body)
		if err != nil {
			panic(err)
		}
		defer zw.Close()

		body, err := io.ReadAll(zw)
		if err != nil {
			panic(err)
		}

		r.Body = io.NopCloser(bytes.NewBuffer(body))
		r.ContentLength = int64(len(body))

		h.ServeHTTP(w, r)
	}
}
