package main

import (
	"log"
	"net/http"

	"github.com/aga-absolut/url-cutter/internal/config"
	"github.com/aga-absolut/url-cutter/internal/handler"
	"github.com/aga-absolut/url-cutter/internal/router"
	"github.com/aga-absolut/url-cutter/internal/storage"
)

func main() {
	cfg := config.NewConfig()
	storage := storage.NewStorage()
	handler := handler.NewHandler(storage, cfg)
	router := router.NewRouter(handler)
 	log.Fatal(http.ListenAndServe(cfg.Host, router))

}
