package router

import (
	"github.com/aga-absolut/url-cutter/internal/handler"
	"github.com/aga-absolut/url-cutter/internal/logger"
	"github.com/go-chi/chi/v5"
)

func NewRouter(handler *handler.Handler) *chi.Mux {
	router := chi.NewRouter()
	router.Post("/", logger.WithFieldsInfo(handler.ShortPostReq))
	router.Post("/api/shorten/", logger.WithFieldsInfo(handler.HandlerJSON))
	router.Get("/{rf}", logger.WithFieldsInfo(handler.ShortGetReq))
	return router
}
