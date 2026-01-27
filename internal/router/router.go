package router

import (
	"github.com/aga-absolut/url-cutter/internal/handler"
	"github.com/aga-absolut/url-cutter/middleware/compress"
	"github.com/aga-absolut/url-cutter/middleware/jwt"
	"github.com/aga-absolut/url-cutter/middleware/logger"
	"github.com/go-chi/chi/v5"
)

func NewRouter(handler *handler.Handler) *chi.Mux {
	router := chi.NewRouter()
	router.Use(logger.WithFieldsInfo)

	router.Post("/", jwt.AuthMiddleware(compress.Decompress(handler.PostHandler)))
	router.Post("/api/shorten", jwt.AuthMiddleware(compress.Decompress(handler.JSONPostHandler)))
	router.Post("/api/shorten/batch", jwt.AuthMiddleware(handler.PostBatchHandler))
	router.Get("/api/user/urls", handler.GetUserURLs)
	router.Get("/{id}", compress.Compress(handler.GetHandler))
	router.Get("/ping", handler.CheckConnecToDB)
	return router
}
