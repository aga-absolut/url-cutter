package main

import (
	"log"
	"net/http"

	"github.com/aga-absolut/url-cutter/internal/config"
	"github.com/aga-absolut/url-cutter/internal/file"
	"github.com/aga-absolut/url-cutter/internal/handler"
	"github.com/aga-absolut/url-cutter/internal/middleware/logger"
	"github.com/aga-absolut/url-cutter/internal/router"
	"github.com/aga-absolut/url-cutter/internal/storage"
)

func main() {
	config := config.NewConfig()
	logger := logger.NewLogger()
	storage := storage.NewStorage()
	file := file.NewFile(config, logger)
	handler := handler.NewHandler(storage, config, file)
	router := router.NewRouter(handler)

	logger.Infow("Starting server", "addr", config.ServerAddress)
	log.Fatal(http.ListenAndServe(config.ServerAddress, router))
}
