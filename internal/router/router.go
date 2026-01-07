package router

import (
	"github.com/aga-absolut/url-cutter/internal/handler"
	"github.com/aga-absolut/url-cutter/middleware/compress"
	"github.com/aga-absolut/url-cutter/middleware/logger"
	"github.com/go-chi/chi/v5"
)

func NewRouter(handler *handler.Handler) *chi.Mux {
	router := chi.NewRouter()
	router.Use(logger.WithFieldsInfo)

	router.Post("/", compress.Decompress(handler.PostHandler))
	router.Post("/api/shorten", compress.Decompress(handler.JSONPostHandler))
	router.Get("/{id}", compress.Compress(handler.GetHandler))
	return router
}
