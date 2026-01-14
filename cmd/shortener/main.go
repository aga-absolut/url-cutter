package main

import (
	"net/http"

	"github.com/aga-absolut/url-cutter/internal/config"
	"github.com/aga-absolut/url-cutter/internal/handler"
	"github.com/aga-absolut/url-cutter/internal/router"
	"github.com/aga-absolut/url-cutter/internal/storage/database"
	"github.com/aga-absolut/url-cutter/internal/storage/file"
	"github.com/aga-absolut/url-cutter/internal/storage/memory"
	"github.com/aga-absolut/url-cutter/middleware/logger"
)

func main() {
	config := config.NewConfig()
	logger := logger.NewLogger()
	storage := memory.NewMemoryStorage()
	db := database.NewDBPostgreSQL(config)
	file := file.NewFile(config, logger)
	handler := handler.NewHandler(storage, config, file, db)
	router := router.NewRouter(handler)

	logger.Infow("Starting server", "addr", config.ServerAddress)
	err := http.ListenAndServe(config.ServerAddress, router)
	if err != nil {
		logger.Fatalw("Server did't start", "Error", err)
	}
}
