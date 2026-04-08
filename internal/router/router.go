package router

import (
	"github.com/aga-absolut/url-cutter/internal/handler"
	"github.com/aga-absolut/url-cutter/middleware/compress"
	"github.com/aga-absolut/url-cutter/middleware/jwt"
	"github.com/aga-absolut/url-cutter/middleware/logger"
	"github.com/go-chi/chi/v5"
)

// NewRouter создает новый маршрутизатор.
func NewRouter(handler *handler.Handler) *chi.Mux {
	router := chi.NewRouter()
	router.Use(logger.WithFieldsInfo)

	router.With(jwt.AuthMiddleware, compress.Decompress).Post("/", handler.PostHandler)
	router.With(jwt.AuthMiddleware, compress.Decompress).Post("/api/shorten", handler.JSONPostHandler)
	router.With(jwt.AuthMiddleware, compress.Decompress).Post("/api/shorten/batch", handler.PostBatchHandler)
	router.With(compress.Compress).Get("/api/internal/stats", handler.GetStatsHandler)
	router.With(compress.Compress).Get("/api/user/urls", handler.GetUserURLs)
	router.With(compress.Compress).Get("/{id}", handler.GetHandler)
	router.Delete("/api/user/urls", handler.DeleteUserURLs)
	router.Get("/ping", handler.CheckConnecToDB)
	return router
}
