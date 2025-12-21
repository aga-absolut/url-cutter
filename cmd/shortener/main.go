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
	dt := make(map[string]string)
	config  := config.NewConfig()
	storage := storage.NewStorage(dt)
	handler := handler.NewHandler(*storage,*config)
	log.Fatal(http.ListenAndServe(":8080", router.NewRouter(*handler)))
}
