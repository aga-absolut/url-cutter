package logger

import (
	"log"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type (
	responseData struct {
		size   int
		status int
	}

	LoggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *LoggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *LoggingResponseWriter) WriteHeader(statuscode int) {
	r.ResponseWriter.WriteHeader(statuscode)
	r.responseData.status = statuscode
}

func WithFieldsInfo(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger, err := zap.NewDevelopment()
		if err != nil {
			log.Fatal(err.Error())
		}

		switch r.Method {
		case http.MethodPost:
			start := time.Now()
			uri := r.RequestURI
			method := r.Method

			h.ServeHTTP(w, r)

			Duration := time.Since(start)

			logger.Info("Request data",
				zap.String("URI", uri),
				zap.String("method", method),
				zap.Duration("time", Duration),
			)

		case http.MethodGet:
			responseData := &responseData{}
			lw := LoggingResponseWriter{
				ResponseWriter: w,
				responseData:   responseData,
			}

			h.ServeHTTP(&lw, r)
			logger.Info("Request data",
				zap.Int("size", responseData.size),
				zap.Int("status", responseData.status),
			)
		default:
			logger.Error("Bad request")
		}
	}
}
