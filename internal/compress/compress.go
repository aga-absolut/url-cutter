package compress

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type (
	compressWriter struct {
		w  http.ResponseWriter
		zw *gzip.Writer
	}

	compressReader struct {
		r  io.Reader
		zr *gzip.Reader
	}
)

func NewCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

func (cw *compressWriter) Write(b []byte) (int, error) {
	ans := cw.w.Header().Get("Content-type")
	if ans == "application/json" || ans == "text/html"{
		return cw.zw.Write(b)
	}
	return cw.w.Write(b)
}

func (cw *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		cw.w.Header().Set("Content-Encoding", "gzip")
	}
	cw.w.WriteHeader(statusCode)
}

func (cw *compressWriter) Header() http.Header {
	return cw.w.Header()
}

func (cw *compressWriter) Close() error {
	return cw.zw.Close()
}

func NewCompressReader(r io.Reader) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (cr *compressReader) Read(p []byte) (n int, err error) {
	return cr.zr.Read(p)
}

func (cr *compressReader) Close() error {
	if err := cr.zr.Close(); err != nil {
		return err
	}
	return cr.zr.Close()
}

func GzipMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ow := w
		ansAcceptEnc := strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")
		if ansAcceptEnc {
			cw := NewCompressWriter(w)
			ow = cw
			defer cw.Close()

		}

		ansContentEnc := strings.Contains(r.Header.Get("Content-Encoding"), "gzip")
		if ansContentEnc {
			cr, err := NewCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = cr
			defer cr.Close()
		}
		h.ServeHTTP(ow, r)
	}
}
