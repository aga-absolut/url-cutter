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

	router.Post("/", compress.Decompress(handler.ShortPostReq))
	router.Post("/api/shorten", compress.Decompress(handler.HandlerJSON))
	router.Get("/{rf}", compress.Compress(handler.ShortGetReq))
	return router
}
