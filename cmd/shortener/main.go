package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aga-absolut/url-cutter/internal/config"
	"github.com/aga-absolut/url-cutter/internal/handler"
	"github.com/aga-absolut/url-cutter/internal/repository"
	"github.com/aga-absolut/url-cutter/internal/router"
	"github.com/aga-absolut/url-cutter/internal/workerpool"
	"github.com/aga-absolut/url-cutter/middleware/logger"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	deleteChan := make(chan string, 10)

	config := config.NewConfig()
	logger := logger.NewLogger()
	storage := repository.NewStorage(config, logger)
	workerpool := workerpool.NewWorkerPool(ctx, deleteChan, storage, 10)
	handler := handler.NewHandler(config, storage, logger, deleteChan)
	router := router.NewRouter(handler)

	server := &http.Server{
		Addr:    config.ServerAddress,
		Handler: router,
	}
	go func() {
		logger.Infow("Starting server", "addr", config.ServerAddress)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Errorw("Server error", "Error", err)
		}
	}()

	<-ctx.Done()
	logger.Info("Shutdown signal received")

	ShutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ShutdownCtx); err != nil {
		logger.Errorw("Server shutdown error", "Error", err)
	}

	workerpool.Stop()
	logger.Info("Application stopped successfully")
}
