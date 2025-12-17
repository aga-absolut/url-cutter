package main

import (
	"log"
	"net/http"

	"github.com/aga-absolut/url-cutter/internal/handler"
	"github.com/go-chi/chi/v5"
)

func main() {
	r := chi.NewRouter()
	r.Post("/", handler.ShortPostReq)
	r.Get(`/{rf}`, handler.ShortGetReq)
	log.Fatal(http.ListenAndServe(":8080", r))
}
