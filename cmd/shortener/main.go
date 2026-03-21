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
	"github.com/aga-absolut/url-cutter/internal/router"
	"github.com/aga-absolut/url-cutter/internal/storage"
	"github.com/aga-absolut/url-cutter/internal/storage/postgreSQL/database"
	"github.com/aga-absolut/url-cutter/internal/workers"
	"github.com/aga-absolut/url-cutter/middleware/logger"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	deleteChan := make(chan string, 10)
	cfg := config.NewConfig()
	logger := logger.NewLogger()
	if cfg.DBDSN != "" {
		if err := database.InitMigrations(cfg, logger); err != nil {
			logger.Fatalw("don`t create migrations", "error", err)
		}
	}
	storage := storage.NewStorage(cfg, logger)
	worker := workers.NewWorkerPool(ctx, deleteChan, storage, config.SizeWorkers, logger)
	handler := handler.NewHandler(cfg, storage, logger, deleteChan)
	router := router.NewRouter(handler)

	server := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: router,
	}

	go func() {
		logger.Infow("Starting server", "addr", cfg.ServerAddress)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Errorw("Server error", "Error", err)
		}
	}()
	time.Sleep(5 * time.Second)

	<-ctx.Done()
	logger.Info("Shutdown signal received")

	ShutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ShutdownCtx); err != nil {
		logger.Errorw("Server shutdown error", "Error", err)
	}

	worker.Stop()
	logger.Info("Application stopped successfully")
}
