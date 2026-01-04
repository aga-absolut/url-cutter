package router

import (
	"github.com/aga-absolut/url-cutter/internal/handler"
	"github.com/aga-absolut/url-cutter/internal/middleware/compress"
	"github.com/aga-absolut/url-cutter/internal/middleware/logger"
	"github.com/go-chi/chi/v5"
)

func NewRouter(handler *handler.Handler) *chi.Mux {
	router := chi.NewRouter()
	router.Use(logger.WithFieldsInfo)

	router.With(compress.Compress).Post("/", handler.ShortPostReq)
	router.With(compress.Compress).Post("/api/shorten", handler.HandlerJSON)
	router.With(compress.Decompress).Get("/{rf}",handler.ShortGetReq)
	return router
}
