package router

import (
	"github.com/aga-absolut/url-cutter/internal/handler"
	"github.com/go-chi/chi/v5"
)

func NewRouter(handler handler.Handler) *chi.Mux {
	router := chi.NewRouter()
	router.Post("/", handler.ShortPostReq)
	router.Get("/{rf}", handler.ShortGetReq)
	return router
}
