package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

var sugar *zap.SugaredLogger

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

func NewLogger() zap.SugaredLogger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	sugar = logger.Sugar()
	return *sugar
}

func WithFieldsInfo(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			start := time.Now()
			uri := r.RequestURI
			method := r.Method

			h.ServeHTTP(w, r)

			duration := time.Since(start)

			sugar.Info("Request data",
				zap.String("URI", uri),
				zap.String("method", method),
				zap.Duration("time", duration),
			)

		case http.MethodGet:
			responseData := &responseData{}
			lw := LoggingResponseWriter{
				ResponseWriter: w,
				responseData:   responseData,
			}

			h.ServeHTTP(&lw, r)
			sugar.Info("Request data",
				zap.Int("size", responseData.size),
				zap.Int("status", responseData.status),
			)
		default:
			sugar.Error("Bad request")
		}
	})
}
