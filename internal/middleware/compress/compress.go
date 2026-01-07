package compress

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type gzipWriter struct {
	http.ResponseWriter
	zipWriter io.Writer
}

func (c gzipWriter) Write(p []byte) (int, error) {
	return c.zipWriter.Write(p)
}

func Compress(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			h.ServeHTTP(w, r)
			return
		}

		zipWriter := gzip.NewWriter(w)
		defer zipWriter.Close()

		w.Header().Set("Content-Encoding", "gzip")
		h.ServeHTTP(gzipWriter{ResponseWriter: w, zipWriter: zipWriter}, r)
	}
}

func Decompress(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			h.ServeHTTP(w, r)
			return
		}

		zipReader, err := gzip.NewReader(r.Body)
		if err != nil {
			http.Error(w, "Bad Request: invalid gzip data", http.StatusBadRequest)
			return
		}

		r.Body = zipReader
		defer zipReader.Close()
		h.ServeHTTP(w, r)
	}
}
