package compress

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// gzipWriter структура.
type gzipWriter struct {
	http.ResponseWriter
	zipWriter io.Writer
}

// Write изменяет метод структуры ResponseWriter.
func (c gzipWriter) Write(p []byte) (int, error) {
	return c.zipWriter.Write(p)
}

// Compress сжимает данные.
func Compress(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			h.ServeHTTP(w, r)
			return
		}

		zipWriter := gzip.NewWriter(w)
		defer zipWriter.Close()

		w.Header().Set("Content-Encoding", "gzip")
		h.ServeHTTP(gzipWriter{ResponseWriter: w, zipWriter: zipWriter}, r)
	})
}

// Decompress распаковывает данные.
func Decompress(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	})
}
