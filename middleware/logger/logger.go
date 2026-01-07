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
		start := time.Now()

		responseData := &responseData{}
		lw := LoggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}

		h.ServeHTTP(&lw, r)

		duration := time.Since(start)

		sugar.Infoln(
			"\n",
			"-----REQUEST-----\n",
			"URI:", r.RequestURI, "\n",
			"Method:", r.Method, "\n",
			"Duration:", duration, "\n",
			"-----RESPONSE-----\n",
			"Status:", responseData.status, "\n",
			"Size:", responseData.size, "\n",
		)
	})
}
